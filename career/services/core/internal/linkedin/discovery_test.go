package linkedin

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type discoveryFixtureSuite struct {
	Description string `json:"description"`
	Scenarios   []struct {
		ScenarioID           string            `json:"scenario_id"`
		Description          string            `json:"description"`
		Criteria             DiscoveryCriteria `json:"criteria"`
		ExpectedBooleanTerms []string          `json:"expected_boolean_terms"`
		ShouldReject         bool              `json:"should_reject"`
	} `json:"scenarios"`
	URLNormalizationCases []struct {
		RawURL             string `json:"raw_url"`
		ExpectedNormalized string `json:"expected_normalized"`
		EntityType         string `json:"entity_type"`
		Slug               string `json:"slug"`
	} `json:"url_normalization_cases"`
}

func loadDiscoveryFixture(t *testing.T) *discoveryFixtureSuite {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_discovery.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read linkedin_discovery.json fixture: %v", err)
	}

	var fix discoveryFixtureSuite
	if err := json.Unmarshal(bytes, &fix); err != nil {
		t.Fatalf("failed to parse linkedin_discovery.json fixture: %v", err)
	}
	return &fix
}

func TestLinkedInDiscovery_FixtureScenarios(t *testing.T) {
	fix := loadDiscoveryFixture(t)

	for _, sc := range fix.Scenarios {
		t.Run(sc.ScenarioID, func(t *testing.T) {
			if sc.ShouldReject {
				err := ValidateDiscoveryCriteria(sc.Criteria)
				if !errors.Is(err, ErrProtectedAttributeProhibited) && !errors.Is(err, ErrInvalidDiscoveryCriteria) {
					t.Fatalf("expected criteria to be rejected, got: %v", err)
				}
				return
			}

			bq, err := BuildBooleanQuery(sc.Criteria)
			if err != nil {
				t.Fatalf("failed to build boolean query: %v", err)
			}

			if !strings.Contains(bq.LinkedInSearchURL, "linkedin.com/search/results/people") {
				t.Errorf("expected search URL to point to linkedin people search, got: %s", bq.LinkedInSearchURL)
			}

			for _, term := range sc.ExpectedBooleanTerms {
				if !strings.Contains(bq.RawQuery, term) {
					t.Errorf("expected query to contain term %s, got query: %s", term, bq.RawQuery)
				}
			}
		})
	}
}

func TestLinkedInDiscovery_ProtectedAttributeRejection_LI05(t *testing.T) {
	cases := []struct {
		name     string
		criteria DiscoveryCriteria
	}{
		{
			name: "Age discrimination filter",
			criteria: DiscoveryCriteria{
				TargetRole: "Software Engineer",
				CustomFilters: map[string]string{
					"age_group": "under_30",
				},
				UserGoal: GoalRecruiting,
			},
		},
		{
			name: "Gender filter in role alternatives",
			criteria: DiscoveryCriteria{
				TargetRole:        "Software Engineer",
				RoleAlternatives:  []string{"Female Developer"},
				UserGoal:          GoalRecruiting,
			},
		},
		{
			name: "Race or ethnicity keyword",
			criteria: DiscoveryCriteria{
				TargetRole: "Account Executive",
				CustomFilters: map[string]string{
					"ethnicity": "asian",
				},
				UserGoal: GoalNetworking,
			},
		},
		{
			name: "Religion exclusion",
			criteria: DiscoveryCriteria{
				TargetRole: "Product Lead",
				Exclusions: []string{"Christian"},
				UserGoal:   GoalRecruiting,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateDiscoveryCriteria(tc.criteria)
			if !errors.Is(err, ErrProtectedAttributeProhibited) {
				t.Fatalf("expected ErrProtectedAttributeProhibited for %s, got: %v", tc.name, err)
			}
		})
	}
}

func TestLinkedInDiscovery_CanonicalURLNormalization(t *testing.T) {
	fix := loadDiscoveryFixture(t)

	for _, c := range fix.URLNormalizationCases {
		t.Run(c.Slug, func(t *testing.T) {
			norm, err := NormalizeLinkedInURL(c.RawURL)
			if err != nil {
				t.Fatalf("failed to normalize URL '%s': %v", c.RawURL, err)
			}
			if norm.NormalizedURL != c.ExpectedNormalized {
				t.Errorf("expected normalized '%s', got '%s'", c.ExpectedNormalized, norm.NormalizedURL)
			}
			if norm.EntityType != c.EntityType {
				t.Errorf("expected entity_type '%s', got '%s'", c.EntityType, norm.EntityType)
			}
			if norm.Slug != c.Slug {
				t.Errorf("expected slug '%s', got '%s'", c.Slug, norm.Slug)
			}
		})
	}

	// Invalid URL check
	_, err := NormalizeLinkedInURL("https://twitter.com/johndoe")
	if !errors.Is(err, ErrInvalidLinkedInURL) {
		t.Errorf("expected ErrInvalidLinkedInURL for non-linkedin URL, got: %v", err)
	}
}

func TestLinkedInDiscovery_ServiceExecution(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	// Seed vault records in ws-alpha
	_ = svc.UpsertPersonRecord(ctx, &PersonRecord{
		RecordID:       "rec-p-01",
		WorkspaceID:    "ws-alpha",
		TenantID:       "tenant-alpha",
		FullName:       "Sarah Chen",
		Headline:       "Lead Systems Architect @ CloudScale",
		CurrentCompany: "CloudScale Systems",
		CurrentRole:    "Lead Systems Architect",
		Location:       "San Francisco Bay Area",
		ObservedAt:     time.Now().UTC(),
	})
	_ = svc.UpsertCompanyRecord(ctx, &CompanyRecord{
		RecordID:    "rec-c-01",
		WorkspaceID: "ws-alpha",
		TenantID:    "tenant-alpha",
		CompanyName: "CloudScale Systems",
		Domain:      "cloudscale.tech",
		Industry:    "Software Development",
		ObservedAt:  time.Now().UTC(),
	})

	criteria := DiscoveryCriteria{
		TargetRole:       "Systems Architect",
		CurrentCompanies: []string{"CloudScale Systems"},
		UserGoal:         GoalRecruiting,
	}

	res, err := svc.DiscoverEntities(ctx, criteria, "tenant-alpha", "ws-alpha")
	if err != nil {
		t.Fatalf("failed to discover entities: %v", err)
	}

	if len(res.MatchedPersons) != 1 || res.MatchedPersons[0].FullName != "Sarah Chen" {
		t.Errorf("expected to match Sarah Chen in vault, got: %+v", res.MatchedPersons)
	}
	if len(res.MatchedCompanies) != 1 || res.MatchedCompanies[0].CompanyName != "CloudScale Systems" {
		t.Errorf("expected to match CloudScale Systems in vault, got: %+v", res.MatchedCompanies)
	}
	if res.TotalVaultMatches != 2 {
		t.Errorf("expected total vault matches 2, got %d", res.TotalVaultMatches)
	}
	if !strings.Contains(res.PlatformNotice, "AT-010") {
		t.Errorf("expected platform notice to reference AT-010, got: %s", res.PlatformNotice)
	}
}
