package search_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/social-platform/services/core/internal/search"
)

// AT-012: Search results, counts, facets, autocomplete and downloads cannot leak another tenant's data.
func TestSearch_AT012_StrictTenantIsolation(t *testing.T) {
	ctx := context.Background()
	backend := search.NewMemorySearchBackend()
	facade := search.NewSearchFacade(backend)

	wsAlpha := "ws-alpha-100"
	wsBeta := "ws-beta-200"

	// 1. Index documents for Workspace Alpha (Acme Corp)
	err := facade.ProjectDocument(ctx, search.IndexJobs, map[string]interface{}{
		"id":              "job-alpha-1",
		"workspace_id":    wsAlpha,
		"title":           "Senior Go Distributed Systems Engineer",
		"company":         "Acme Corporation",
		"tags":            []string{"golang", "redis", "distributed"},
		"employment_type": "full_time",
	})
	if err != nil {
		t.Fatalf("failed to index Alpha job 1: %v", err)
	}

	err = facade.ProjectDocument(ctx, search.IndexJobs, map[string]interface{}{
		"id":              "job-alpha-2",
		"workspace_id":    wsAlpha,
		"title":           "Principal Cloud Infrastructure Architect",
		"company":         "Acme Corporation",
		"tags":            []string{"kubernetes", "aws", "terraform"},
		"employment_type": "full_time",
	})
	if err != nil {
		t.Fatalf("failed to index Alpha job 2: %v", err)
	}

	// 2. Index documents for Workspace Beta (Beta Global)
	err = facade.ProjectDocument(ctx, search.IndexJobs, map[string]interface{}{
		"id":              "job-beta-1",
		"workspace_id":    wsBeta,
		"title":           "Confidential Senior Executive Recruiter",
		"company":         "Beta Global Stealth",
		"tags":            []string{"hr", "executive", "stealth"},
		"employment_type": "contract",
	})
	if err != nil {
		t.Fatalf("failed to index Beta job 1: %v", err)
	}

	// 3. Alpha searches with generic query: MUST NOT see Beta's jobs
	resAlpha, err := facade.Search(ctx, search.SearchRequest{
		WorkspaceID: wsAlpha,
		Index:       search.IndexJobs,
		Query:       "",
		Facets:      []string{"company", "employment_type"},
	})
	if err != nil {
		t.Fatalf("Alpha search failed: %v", err)
	}

	if resAlpha.TotalHits != 2 {
		t.Fatalf("expected Alpha to have exactly 2 hits, got %d", resAlpha.TotalHits)
	}
	for _, hit := range resAlpha.Hits {
		if hit["workspace_id"] != wsAlpha {
			t.Fatalf("AT-012 LEAK: Alpha query returned document belonging to workspace %v", hit["workspace_id"])
		}
		if hit["company"] == "Beta Global Stealth" {
			t.Fatalf("AT-012 LEAK: Alpha saw Beta company name!")
		}
	}

	// Verify Facets: Beta's company must NOT appear in Alpha's facet distribution
	companyFacets := resAlpha.FacetDistribution["company"]
	if companyFacets["Beta Global Stealth"] > 0 {
		t.Fatalf("AT-012 LEAK: Beta company leaked in Alpha's facet counts: %v", companyFacets)
	}
	if companyFacets["Acme Corporation"] != 2 {
		t.Fatalf("expected Acme Corporation facet count to be 2, got %d", companyFacets["Acme Corporation"])
	}

	// 4. Beta searches: MUST NOT see Alpha's jobs
	resBeta, err := facade.Search(ctx, search.SearchRequest{
		WorkspaceID: wsBeta,
		Index:       search.IndexJobs,
		Query:       "Engineer", // Only Alpha has "Engineer" in title
	})
	if err != nil {
		t.Fatalf("Beta search failed: %v", err)
	}
	if resBeta.TotalHits != 0 {
		t.Fatalf("AT-012 LEAK: Beta search for 'Engineer' found %d hits from Alpha!", resBeta.TotalHits)
	}

	// 5. Autocomplete Isolation: Alpha searches prefix "Beta" -> must return 0 suggestions
	autoAlpha, err := facade.Autocomplete(ctx, search.AutocompleteRequest{
		WorkspaceID: wsAlpha,
		Index:       search.IndexJobs,
		Prefix:      "Beta",
		Field:       "company",
	})
	if err != nil {
		t.Fatalf("Autocomplete failed: %v", err)
	}
	if len(autoAlpha.Suggestions) != 0 {
		t.Fatalf("AT-012 LEAK: Alpha autocomplete returned Beta suggestions: %v", autoAlpha.Suggestions)
	}

	// Beta searches prefix "Beta" -> returns "Beta Global Stealth"
	autoBeta, err := facade.Autocomplete(ctx, search.AutocompleteRequest{
		WorkspaceID: wsBeta,
		Index:       search.IndexJobs,
		Prefix:      "Beta",
		Field:       "company",
	})
	if err != nil {
		t.Fatalf("Beta autocomplete failed: %v", err)
	}
	if len(autoBeta.Suggestions) != 1 || autoBeta.Suggestions[0] != "Beta Global Stealth" {
		t.Fatalf("expected Beta autocomplete to find its own company, got %v", autoBeta.Suggestions)
	}
}

// AT-012 & REQ-017: Delete tombstones immediately hide documents even if search index has lag.
func TestSearch_DeleteTombstoneImmediateFilter(t *testing.T) {
	ctx := context.Background()
	backend := search.NewMemorySearchBackend()
	facade := search.NewSearchFacade(backend)

	wsID := "ws-test-300"
	docID := "content-item-999"

	// 1. Index a post
	err := facade.ProjectDocument(ctx, search.IndexContent, map[string]interface{}{
		"id":           docID,
		"workspace_id": wsID,
		"title":        "Exclusive Company Announcement",
		"status":       "published",
	})
	if err != nil {
		t.Fatalf("failed to project content item: %v", err)
	}

	// Verify it can be found
	res, err := facade.Search(ctx, search.SearchRequest{
		WorkspaceID: wsID,
		Index:       search.IndexContent,
		Query:       "Exclusive",
	})
	if err != nil || res.TotalHits != 1 {
		t.Fatalf("expected 1 hit before deletion, got %d (err: %v)", res.TotalHits, err)
	}

	// 2. Delete document
	err = facade.DeleteDocument(ctx, search.IndexContent, wsID, docID)
	if err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}

	// 3. Search immediately: Tombstone MUST hide it even if query runs
	resAfterDelete, err := facade.Search(ctx, search.SearchRequest{
		WorkspaceID: wsID,
		Index:       search.IndexContent,
		Query:       "Exclusive",
	})
	if err != nil {
		t.Fatalf("Search after delete failed: %v", err)
	}
	if resAfterDelete.TotalHits != 0 || len(resAfterDelete.Hits) != 0 {
		t.Fatalf("expected 0 hits after deletion via tombstone, got %d", resAfterDelete.TotalHits)
	}
}

// Mandatory workspace scoping: Missing workspace scope fails closed.
func TestSearch_WorkspaceScopeEnforcement(t *testing.T) {
	ctx := context.Background()
	backend := search.NewMemorySearchBackend()
	facade := search.NewSearchFacade(backend)

	// Missing workspace_id in Search
	_, err := facade.Search(ctx, search.SearchRequest{
		WorkspaceID: "", // Missing
		Index:       search.IndexJobs,
		Query:       "Software",
	})
	if !errors.Is(err, search.ErrWorkspaceScopeRequired) {
		t.Fatalf("expected ErrWorkspaceScopeRequired when workspace_id is empty, got %v", err)
	}

	// Missing workspace_id in Autocomplete
	_, err = facade.Autocomplete(ctx, search.AutocompleteRequest{
		WorkspaceID: "", // Missing
		Index:       search.IndexJobs,
		Prefix:      "Acme",
		Field:       "company",
	})
	if !errors.Is(err, search.ErrWorkspaceScopeRequired) {
		t.Fatalf("expected ErrWorkspaceScopeRequired on autocomplete without workspace_id, got %v", err)
	}

	// Missing workspace_id in ProjectDocument
	err = facade.ProjectDocument(ctx, search.IndexJobs, map[string]interface{}{
		"id":    "job-missing-ws",
		"title": "Engineer",
	})
	if !errors.Is(err, search.ErrWorkspaceScopeRequired) {
		t.Fatalf("expected ErrWorkspaceScopeRequired on document without workspace_id, got %v", err)
	}
}

// REQ-017 & REQ-023: Raw resumes must never be indexed into Meilisearch.
func TestSearch_RawResumeIndexProhibited(t *testing.T) {
	ctx := context.Background()
	backend := search.NewMemorySearchBackend()
	facade := search.NewSearchFacade(backend)

	err := facade.ProjectDocument(ctx, search.IndexJobs, map[string]interface{}{
		"id":             "doc-bad-1",
		"workspace_id":   "ws-1",
		"title":          "Resume match",
		"raw_resume_text": "John Doe, SSN: 000-00-0000, 10 years experience...",
	})
	if !errors.Is(err, search.ErrRawDocumentIndexProhibited) {
		t.Fatalf("expected ErrRawDocumentIndexProhibited for raw_resume_text, got %v", err)
	}
}

// REQ-017 & REQ-019: Rebuild index from primary records and check diagnostics.
func TestSearch_RebuildAndLagDiagnostics(t *testing.T) {
	ctx := context.Background()
	backend := search.NewMemorySearchBackend()
	facade := search.NewSearchFacade(backend)

	wsID := "ws-rebuild"

	// Mock primary database fetcher function
	primaryDBDocs := func() ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{
				"id":           "app-1",
				"workspace_id": wsID,
				"company":      "Stripe",
				"role_title":   "Staff Engineer",
				"status":       "applied",
			},
			{
				"id":           "app-2",
				"workspace_id": wsID,
				"company":      "GitHub",
				"role_title":   "Senior Developer Advocate",
				"status":       "interviewing",
			},
		}, nil
	}

	// Rebuild index
	err := facade.RebuildIndex(ctx, search.IndexApplications, primaryDBDocs)
	if err != nil {
		t.Fatalf("RebuildIndex failed: %v", err)
	}

	// Verify search works from rebuilt projection
	res, err := facade.Search(ctx, search.SearchRequest{
		WorkspaceID: wsID,
		Index:       search.IndexApplications,
		Query:       "Staff",
	})
	if err != nil || res.TotalHits != 1 {
		t.Fatalf("expected 1 hit after rebuild, got %d (err: %v)", res.TotalHits, err)
	}

	// Verify Diagnostics
	diag, err := facade.GetDiagnostics(ctx, search.IndexApplications)
	if err != nil {
		t.Fatalf("GetDiagnostics failed: %v", err)
	}
	if diag.TotalDocuments != 2 {
		t.Errorf("expected 2 total documents, got %d", diag.TotalDocuments)
	}
	if diag.LastRebuildAt == nil || time.Since(*diag.LastRebuildAt) > time.Minute {
		t.Errorf("expected recent LastRebuildAt timestamp, got %v", diag.LastRebuildAt)
	}
}
