package search

import (
	"errors"
	"time"
)

// IndexName identifies a specialized projection index.
type IndexName string

const (
	IndexJobs         IndexName = "jobs"
	IndexContent      IndexName = "content"
	IndexApplications IndexName = "applications"
)

var (
	ErrWorkspaceScopeRequired       = errors.New("workspace scope is mandatory for search and autocomplete (AT-012)")
	ErrUnauthorizedTenantAccess     = errors.New("unauthorized cross-tenant access attempt detected (AT-012)")
	ErrRawDocumentIndexProhibited   = errors.New("indexing unscrubbed raw resumes or private documents into search is prohibited (REQ-017, REQ-023)")
	ErrIndexNotFound                = errors.New("search index not found")
	ErrDocumentNotFound             = errors.New("document not found in projection index")
)

// BaseProjection contains standard multi-tenant metadata required on all searchable projections.
type BaseProjection struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	UserID      string    `json:"user_id,omitempty"` // For owner-scoped records
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// JobPostingProjection represents a searchable scrubbed job record.
type JobPostingProjection struct {
	BaseProjection
	Source          string   `json:"source"`
	Title           string   `json:"title"`
	Company         string   `json:"company"`
	Location        string   `json:"location"`
	EmploymentType  string   `json:"employment_type"`
	Tags            []string `json:"tags"`
	SalaryMin       float64  `json:"salary_min,omitempty"`
	SalaryMax       float64  `json:"salary_max,omitempty"`
	Currency        string   `json:"currency,omitempty"`
}

// ContentItemProjection represents a searchable social content item.
type ContentItemProjection struct {
	BaseProjection
	Title           string   `json:"title"`
	ContentSnippet  string   `json:"content_snippet"`
	TargetPlatforms []string `json:"target_platforms"`
	Status          string   `json:"status"`
}

// ApplicationProjection represents an authorized opportunity/application record.
type ApplicationProjection struct {
	BaseProjection
	Company   string `json:"company"`
	RoleTitle string `json:"role_title"`
	Status    string `json:"status"`
	Stage     string `json:"stage"`
}

// DeleteTombstone ensures immediate unsearchability even during indexing queue lag (REQ-017, AT-012).
type DeleteTombstone struct {
	Index       IndexName `json:"index"`
	DocumentID  string    `json:"document_id"`
	WorkspaceID string    `json:"workspace_id"`
	DeletedAt   time.Time `json:"deleted_at"`
}

// SearchRequest defines an authorized, tenant-scoped search query.
type SearchRequest struct {
	WorkspaceID string            `json:"workspace_id"` // MANDATORY
	UserID      string            `json:"user_id,omitempty"`
	Index       IndexName         `json:"index"`
	Query       string            `json:"query"`
	Filter      map[string]string `json:"filter,omitempty"` // e.g. status -> "published"
	Facets      []string          `json:"facets,omitempty"` // e.g. ["company", "status"]
	Limit       int               `json:"limit,omitempty"`
	Offset      int               `json:"offset,omitempty"`
}

// SearchResult returns tenant-isolated hits, counts, and facets.
type SearchResult struct {
	Hits              []map[string]interface{}        `json:"hits"`
	TotalHits         int                             `json:"total_hits"`
	FacetDistribution map[string]map[string]int       `json:"facet_distribution,omitempty"`
	ProcessingTimeMs  int64                           `json:"processing_time_ms"`
}

// AutocompleteRequest defines an authorized, tenant-scoped suggestion query.
type AutocompleteRequest struct {
	WorkspaceID string    `json:"workspace_id"` // MANDATORY
	Index       IndexName `json:"index"`
	Prefix      string    `json:"prefix"`
	Field       string    `json:"field"` // e.g. "company", "title", "tags"
	Limit       int       `json:"limit,omitempty"`
}

// AutocompleteResult contains scoped string suggestions.
type AutocompleteResult struct {
	Suggestions []string `json:"suggestions"`
}

// SearchDiagnostics provides visibility into index health, queue depth, and lag.
type SearchDiagnostics struct {
	IndexName         IndexName  `json:"index_name"`
	TotalDocuments    int        `json:"total_documents"`
	ActiveTombstones  int        `json:"active_tombstones"`
	PendingQueueDepth int        `json:"pending_queue_depth"`
	EstimatedLagMs    int64      `json:"estimated_lag_ms"`
	LastRebuildAt     *time.Time `json:"last_rebuild_at,omitempty"`
}
