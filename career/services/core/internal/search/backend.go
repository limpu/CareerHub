package search

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// SearchBackend defines the low-level search engine driver interface.
type SearchBackend interface {
	IndexDocuments(ctx context.Context, index IndexName, docs []map[string]interface{}) error
	DeleteDocuments(ctx context.Context, index IndexName, ids []string) error
	Search(ctx context.Context, req SearchRequest) (*SearchResult, error)
	Autocomplete(ctx context.Context, req AutocompleteRequest) (*AutocompleteResult, error)
	GetStats(ctx context.Context, index IndexName) (*SearchDiagnostics, error)
	RecreateIndex(ctx context.Context, index IndexName) error
}

// MemorySearchBackend provides an in-memory, thread-safe search backend for deterministic tests and lightweight environments.
type MemorySearchBackend struct {
	mu            sync.RWMutex
	indices       map[IndexName]map[string]map[string]interface{}
	lastRebuildAt map[IndexName]time.Time
}

// NewMemorySearchBackend creates a new MemorySearchBackend.
func NewMemorySearchBackend() *MemorySearchBackend {
	return &MemorySearchBackend{
		indices: map[IndexName]map[string]map[string]interface{}{
			IndexJobs:         make(map[string]map[string]interface{}),
			IndexContent:      make(map[string]map[string]interface{}),
			IndexApplications: make(map[string]map[string]interface{}),
		},
		lastRebuildAt: make(map[IndexName]time.Time),
	}
}

func (b *MemorySearchBackend) IndexDocuments(ctx context.Context, index IndexName, docs []map[string]interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	idx, exists := b.indices[index]
	if !exists {
		idx = make(map[string]map[string]interface{})
		b.indices[index] = idx
	}

	for _, doc := range docs {
		id, ok := doc["id"].(string)
		if !ok || id == "" {
			return fmt.Errorf("document missing string 'id' field")
		}
		// Deep copy document to prevent mutation
		clone := make(map[string]interface{}, len(doc))
		for k, v := range doc {
			clone[k] = v
		}
		idx[id] = clone
	}
	return nil
}

func (b *MemorySearchBackend) DeleteDocuments(ctx context.Context, index IndexName, ids []string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	idx, exists := b.indices[index]
	if !exists {
		return nil
	}

	for _, id := range ids {
		delete(idx, id)
	}
	return nil
}

func (b *MemorySearchBackend) Search(ctx context.Context, req SearchRequest) (*SearchResult, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	idx, exists := b.indices[req.Index]
	if !exists {
		return &SearchResult{Hits: []map[string]interface{}{}, TotalHits: 0}, nil
	}

	qLower := strings.ToLower(req.Query)
	var matchedHits []map[string]interface{}
	facetCounts := make(map[string]map[string]int)
	for _, f := range req.Facets {
		facetCounts[f] = make(map[string]int)
	}

	for _, doc := range idx {
		// 1. Mandatory Tenant Isolation Check (AT-012)
		docWs, _ := doc["workspace_id"].(string)
		if docWs != req.WorkspaceID {
			continue // Strictly exclude other tenants' data
		}

		// 2. Owner-level filter if specified
		if req.UserID != "" {
			docUser, hasUser := doc["user_id"].(string)
			if hasUser && docUser != "" && docUser != req.UserID {
				continue
			}
		}

		// 3. Attribute filters
		if !matchesFilters(doc, req.Filter) {
			continue
		}

		// 4. Query text match
		if qLower != "" && !matchesQuery(doc, qLower) {
			continue
		}

		// Document passed all criteria
		matchedHits = append(matchedHits, doc)

		// Accumulate facet distribution strictly within matching documents
		for _, f := range req.Facets {
			if val, ok := doc[f]; ok {
				switch v := val.(type) {
				case string:
					facetCounts[f][v]++
				case []string:
					for _, item := range v {
						facetCounts[f][item]++
					}
				}
			}
		}
	}

	totalHits := len(matchedHits)

	// Pagination
	offset := req.Offset
	if offset > totalHits {
		offset = totalHits
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	end := offset + limit
	if end > totalHits {
		end = totalHits
	}

	paginatedHits := matchedHits[offset:end]

	return &SearchResult{
		Hits:              paginatedHits,
		TotalHits:         totalHits,
		FacetDistribution: facetCounts,
		ProcessingTimeMs:  1,
	}, nil
}

func (b *MemorySearchBackend) Autocomplete(ctx context.Context, req AutocompleteRequest) (*AutocompleteResult, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	idx, exists := b.indices[req.Index]
	if !exists {
		return &AutocompleteResult{Suggestions: []string{}}, nil
	}

	prefixLower := strings.ToLower(req.Prefix)
	candidateSet := make(map[string]bool)

	for _, doc := range idx {
		// Mandatory tenant scope check (AT-012)
		docWs, _ := doc["workspace_id"].(string)
		if docWs != req.WorkspaceID {
			continue
		}

		val, ok := doc[req.Field]
		if !ok {
			continue
		}

		switch v := val.(type) {
		case string:
			if strings.HasPrefix(strings.ToLower(v), prefixLower) {
				candidateSet[v] = true
			}
		case []string:
			for _, item := range v {
				if strings.HasPrefix(strings.ToLower(item), prefixLower) {
					candidateSet[item] = true
				}
			}
		}
	}

	var suggestions []string
	for s := range candidateSet {
		suggestions = append(suggestions, s)
	}
	sort.Strings(suggestions)

	limit := req.Limit
	if limit <= 0 {
		limit = 5
	}
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return &AutocompleteResult{Suggestions: suggestions}, nil
}

func (b *MemorySearchBackend) GetStats(ctx context.Context, index IndexName) (*SearchDiagnostics, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	idx := b.indices[index]
	var lastRebuild *time.Time
	if t, ok := b.lastRebuildAt[index]; ok {
		lastRebuild = &t
	}

	return &SearchDiagnostics{
		IndexName:         index,
		TotalDocuments:    len(idx),
		ActiveTombstones:  0,
		PendingQueueDepth: 0,
		EstimatedLagMs:    0,
		LastRebuildAt:     lastRebuild,
	}, nil
}

func (b *MemorySearchBackend) RecreateIndex(ctx context.Context, index IndexName) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.indices[index] = make(map[string]map[string]interface{})
	b.lastRebuildAt[index] = time.Now().UTC()
	return nil
}

func matchesFilters(doc map[string]interface{}, filters map[string]string) bool {
	for k, expected := range filters {
		actualVal, exists := doc[k]
		if !exists {
			return false
		}
		if fmt.Sprintf("%v", actualVal) != expected {
			return false
		}
	}
	return true
}

func matchesQuery(doc map[string]interface{}, q string) bool {
	for _, v := range doc {
		switch str := v.(type) {
		case string:
			if strings.Contains(strings.ToLower(str), q) {
				return true
			}
		case []string:
			for _, item := range str {
				if strings.Contains(strings.ToLower(item), q) {
					return true
				}
			}
		}
	}
	return false
}
