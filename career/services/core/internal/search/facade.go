package search

import (
	"context"
	"strings"
	"sync"
	"time"
)

// SearchFacade coordinates tenant validation, raw-document sanitization,
// projection writes, deletion tombstones, and scoped querying (ADR-005, REQ-017, AT-012).
type SearchFacade struct {
	backend    SearchBackend
	mu         sync.RWMutex
	tombstones map[string]DeleteTombstone // key: "index:doc_id"
}

// NewSearchFacade initializes the search facade with a chosen backend.
func NewSearchFacade(backend SearchBackend) *SearchFacade {
	return &SearchFacade{
		backend:    backend,
		tombstones: make(map[string]DeleteTombstone),
	}
}

// ProjectDocument validates multi-tenant metadata, ensures no prohibited raw data is indexed,
// and sends projection documents to the search backend.
func (f *SearchFacade) ProjectDocument(ctx context.Context, index IndexName, doc map[string]interface{}) error {
	// 1. Mandatory workspace scope validation
	wsID, ok := doc["workspace_id"].(string)
	if !ok || strings.TrimSpace(wsID) == "" {
		return ErrWorkspaceScopeRequired
	}

	docID, ok := doc["id"].(string)
	if !ok || strings.TrimSpace(docID) == "" {
		return ErrWorkspaceScopeRequired
	}

	// 2. Prohibit unscrubbed raw private resumes or deep inboxes (REQ-017, REQ-023)
	for key := range doc {
		lowerKey := strings.ToLower(key)
		if strings.Contains(lowerKey, "raw_resume") ||
			strings.Contains(lowerKey, "resume_blob") ||
			strings.Contains(lowerKey, "full_inbox") ||
			strings.Contains(lowerKey, "unscrubbed") {
			return ErrRawDocumentIndexProhibited
		}
	}

	// 3. Clear any active deletion tombstone for this document
	tombstoneKey := makeTombstoneKey(index, docID)
	f.mu.Lock()
	delete(f.tombstones, tombstoneKey)
	f.mu.Unlock()

	// 4. Index in backend
	return f.backend.IndexDocuments(ctx, index, []map[string]interface{}{doc})
}

// DeleteDocument marks a deletion tombstone immediately and asynchronously removes from backend.
// Tombstone guarantees immediate unsearchability even during indexing lag (AT-012).
func (f *SearchFacade) DeleteDocument(ctx context.Context, index IndexName, workspaceID, docID string) error {
	if strings.TrimSpace(workspaceID) == "" {
		return ErrWorkspaceScopeRequired
	}

	tombstone := DeleteTombstone{
		Index:       index,
		DocumentID:  docID,
		WorkspaceID: workspaceID,
		DeletedAt:   time.Now().UTC(),
	}

	tombstoneKey := makeTombstoneKey(index, docID)
	f.mu.Lock()
	f.tombstones[tombstoneKey] = tombstone
	f.mu.Unlock()

	return f.backend.DeleteDocuments(ctx, index, []string{docID})
}

// Search enforces tenant scoping, executes backend search, and hydrates through tombstone filter (AT-012).
func (f *SearchFacade) Search(ctx context.Context, req SearchRequest) (*SearchResult, error) {
	// Mandatory tenant scope check
	if strings.TrimSpace(req.WorkspaceID) == "" {
		return nil, ErrWorkspaceScopeRequired
	}

	rawResult, err := f.backend.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	// Hydration Filter: Strip any results that match an active tombstone
	var cleanHits []map[string]interface{}
	for _, hit := range rawResult.Hits {
		docID, ok := hit["id"].(string)
		if !ok {
			continue
		}
		tombstoneKey := makeTombstoneKey(req.Index, docID)
		if _, isTombstoned := f.tombstones[tombstoneKey]; isTombstoned {
			continue // Drop immediately due to deletion tombstone
		}
		cleanHits = append(cleanHits, hit)
	}

	rawResult.Hits = cleanHits
	rawResult.TotalHits = len(cleanHits)
	return rawResult, nil
}

// Autocomplete provides tenant-isolated search suggestions (AT-012).
func (f *SearchFacade) Autocomplete(ctx context.Context, req AutocompleteRequest) (*AutocompleteResult, error) {
	if strings.TrimSpace(req.WorkspaceID) == "" {
		return nil, ErrWorkspaceScopeRequired
	}

	return f.backend.Autocomplete(ctx, req)
}

// RebuildIndex completely recreates a projection index from primary PostgreSQL storage (REQ-017).
func (f *SearchFacade) RebuildIndex(ctx context.Context, index IndexName, fetchPrimaryDocs func() ([]map[string]interface{}, error)) error {
	// 1. Wipe and recreate index in search backend
	if err := f.backend.RecreateIndex(ctx, index); err != nil {
		return err
	}

	// 2. Clear tombstones for this index
	f.mu.Lock()
	for k, t := range f.tombstones {
		if t.Index == index {
			delete(f.tombstones, k)
		}
	}
	f.mu.Unlock()

	// 3. Fetch primary records from authoritative PostgreSQL storage
	docs, err := fetchPrimaryDocs()
	if err != nil {
		return err
	}

	if len(docs) == 0 {
		return nil
	}

	// 4. Validate and re-project all documents
	for _, doc := range docs {
		if err := f.ProjectDocument(ctx, index, doc); err != nil {
			return err
		}
	}

	return nil
}

// GetDiagnostics provides visibility into index volume, active tombstones, and indexing lag (REQ-019).
func (f *SearchFacade) GetDiagnostics(ctx context.Context, index IndexName) (*SearchDiagnostics, error) {
	diag, err := f.backend.GetStats(ctx, index)
	if err != nil {
		return nil, err
	}

	f.mu.RLock()
	activeTombstones := 0
	for _, t := range f.tombstones {
		if t.Index == index {
			activeTombstones++
		}
	}
	f.mu.RUnlock()

	diag.ActiveTombstones = activeTombstones
	return diag, nil
}

func makeTombstoneKey(index IndexName, docID string) string {
	return string(index) + ":" + docID
}
