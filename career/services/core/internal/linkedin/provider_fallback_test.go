package linkedin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type providerFallbackFixture struct {
	Scenarios []struct {
		ScenarioID                 string            `json:"scenario_id"`
		Description                string            `json:"description"`
		TenantID                   string            `json:"tenant_id"`
		WorkspaceID                string            `json:"workspace_id"`
		Action                     string            `json:"action"`
		RequiredPermission         string            `json:"required_permission"`
		Providers                  []ProviderAdapter `json:"providers"`
		ExpectedDispatchedProvider string            `json:"expected_dispatched_provider"`
		ExpectedFallbackInvoked    bool              `json:"expected_fallback_invoked"`
		ExpectedFallbackReason     string            `json:"expected_fallback_reason"`
		ExpectedSuccess            bool              `json:"expected_success"`
		ExpectedError              string            `json:"expected_error"`
		RawDiagnosticError         string            `json:"raw_diagnostic_error"`
		ExpectedRedactedText       string            `json:"expected_redacted_text"`
	} `json:"scenarios"`
}

func loadProviderFallbackFixture(t *testing.T) *providerFallbackFixture {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_provider_fallback.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read provider fallback fixture at %s: %v", path, err)
	}
	var f providerFallbackFixture
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("failed to parse provider fallback fixture: %v", err)
	}
	return &f
}

func TestProviderFallback_FixtureGroundTruth(t *testing.T) {
	fixture := loadProviderFallbackFixture(t)
	ctx := context.Background()

	for _, sc := range fixture.Scenarios {
		sc := sc
		t.Run(sc.ScenarioID, func(t *testing.T) {
			if sc.RawDiagnosticError != "" {
				// Scrubber test scenario
				scrubbed := ScrubDiagnosticSecrets(sc.RawDiagnosticError)
				if scrubbed != sc.ExpectedRedactedText {
					t.Fatalf("expected scrubbed text %q, got %q", sc.ExpectedRedactedText, scrubbed)
				}
				return
			}

			// Setup in-memory service
			repo := NewMemoryRepository()
			svc := NewService(repo)

			for _, p := range sc.Providers {
				pCopy := p
				pCopy.TenantID = sc.TenantID
				pCopy.WorkspaceID = sc.WorkspaceID
				pCopy.CreatedAt = time.Now().UTC()
				pCopy.UpdatedAt = time.Now().UTC()
				if err := svc.RegisterProviderAdapter(ctx, &pCopy); err != nil {
					t.Fatalf("failed to register provider %s: %v", pCopy.ProviderID, err)
				}
			}

			req := DispatchActionRequest{
				TenantID:           sc.TenantID,
				WorkspaceID:        sc.WorkspaceID,
				Action:             sc.Action,
				RequiredPermission: sc.RequiredPermission,
			}

			res, err := svc.DispatchWithFallback(ctx, req)
			if sc.ExpectedSuccess {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				if res == nil {
					t.Fatal("expected non-nil dispatch result")
				}
				if res.DispatchedProviderID != sc.ExpectedDispatchedProvider {
					t.Errorf("expected dispatched provider %s, got %s", sc.ExpectedDispatchedProvider, res.DispatchedProviderID)
				}
				if res.FallbackInvoked != sc.ExpectedFallbackInvoked {
					t.Errorf("expected fallback invoked = %v, got %v", sc.ExpectedFallbackInvoked, res.FallbackInvoked)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error containing %s, got nil error", sc.ExpectedError)
				}
			}
		})
	}
}

func TestProviderFallback_DiagnosticDoctor_Probes(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	// Register 1 healthy provider with complete scopes
	p1 := ProviderAdapter{
		ProviderID:       "prov-alpha-ent",
		Name:             "Alpha Enterprise",
		AdapterType:      AdapterOfficialEnterpriseAPI,
		Tier:             1,
		SupportedActions: []string{"identity_read", "post_authoring_assisted"},
		ExpectedScopes:   []string{"openid", "profile", "email"},
		GrantedScopes:    []string{"openid", "profile", "email"},
		HealthStatus:     HealthStatusHealthy,
		LatencyMs:        120,
		TenantID:         "tenant-alpha",
		WorkspaceID:      "ws-alpha",
	}
	if err := svc.RegisterProviderAdapter(ctx, &p1); err != nil {
		t.Fatalf("failed to register p1: %v", err)
	}

	// Register 1 degraded provider with missing scope
	p2 := ProviderAdapter{
		ProviderID:       "prov-alpha-oidc",
		Name:             "Alpha OIDC",
		AdapterType:      AdapterConsumerOIDCApp,
		Tier:             2,
		SupportedActions: []string{"identity_read"},
		ExpectedScopes:   []string{"openid", "profile", "email"},
		GrantedScopes:    []string{"openid", "profile"}, // missing email
		HealthStatus:     HealthStatusHealthy,
		LatencyMs:        150,
		TenantID:         "tenant-alpha",
		WorkspaceID:      "ws-alpha",
	}
	if err := svc.RegisterProviderAdapter(ctx, &p2); err != nil {
		t.Fatalf("failed to register p2: %v", err)
	}

	report, err := svc.RunDiagnosticDoctor(ctx, "ws-alpha", "tenant-alpha")
	if err != nil {
		t.Fatalf("RunDiagnosticDoctor failed: %v", err)
	}

	if report.TotalProviders != 2 {
		t.Errorf("expected 2 total providers, got %d", report.TotalProviders)
	}
	if report.HealthyProviders != 1 {
		t.Errorf("expected 1 healthy provider, got %d", report.HealthyProviders)
	}
	if report.DegradedProviders != 1 {
		t.Errorf("expected 1 degraded provider, got %d", report.DegradedProviders)
	}
	if report.OverallHealth != "degraded" {
		t.Errorf("expected overall health 'degraded', got %s", report.OverallHealth)
	}
}

func TestProviderFallback_FailClosedOnChainExhaustion(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	// Register a rate-limited adapter
	p := ProviderAdapter{
		ProviderID:       "prov-degraded",
		Name:             "Degraded Adapter",
		AdapterType:      AdapterOfficialEnterpriseAPI,
		Tier:             1,
		SupportedActions: []string{"identity_read"},
		ExpectedScopes:   []string{"openid"},
		GrantedScopes:    []string{"openid"},
		HealthStatus:     HealthStatusRateLimited,
		LatencyMs:        2500,
		LastErrorMessage: "429 Too Many Requests",
		TenantID:         "tenant-alpha",
		WorkspaceID:      "ws-alpha",
	}
	_ = svc.RegisterProviderAdapter(ctx, &p)

	req := DispatchActionRequest{
		TenantID:           "tenant-alpha",
		WorkspaceID:        "ws-alpha",
		Action:             "identity_read",
		RequiredPermission: "read_identity",
	}

	res, err := svc.DispatchWithFallback(ctx, req)
	if err == nil {
		t.Fatal("expected ErrProviderChainExhausted when all candidates degraded, got nil")
	}
	if res != nil {
		t.Fatal("expected nil result on exhaustion")
	}
}

func TestProviderFallback_CrossTenantIsolation_AT011(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	pAlpha := ProviderAdapter{
		ProviderID:       "prov-alpha",
		Name:             "Alpha Provider",
		AdapterType:      AdapterOfficialEnterpriseAPI,
		Tier:             1,
		SupportedActions: []string{"identity_read"},
		ExpectedScopes:   []string{"openid"},
		GrantedScopes:    []string{"openid"},
		HealthStatus:     HealthStatusHealthy,
		TenantID:         "tenant-alpha",
		WorkspaceID:      "ws-alpha",
	}
	_ = svc.RegisterProviderAdapter(ctx, &pAlpha)

	// Attempt access from foreign tenant
	_, err := svc.GetProviderAdapter(ctx, "prov-alpha", "tenant-beta")
	if err == nil {
		t.Fatal("expected cross-tenant error, got nil")
	}

	// Foreign workspace list should be empty
	foreignList, err := svc.ListProviderAdapters(ctx, "ws-beta", "tenant-beta")
	if err != nil {
		t.Fatalf("ListProviderAdapters failed: %v", err)
	}
	if len(foreignList) != 0 {
		t.Errorf("expected 0 foreign providers, got %d", len(foreignList))
	}
}
