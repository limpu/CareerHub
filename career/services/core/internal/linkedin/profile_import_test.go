package linkedin

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type profileImportScenario struct {
	ScenarioID         string                 `json:"scenario_id"`
	Description        string                 `json:"description"`
	InputType          string                 `json:"input_type"`
	ArchiveFiles       map[string]string      `json:"archive_files,omitempty"`
	RawText            string                 `json:"raw_text,omitempty"`
	APIResponsePayload map[string]interface{} `json:"api_response_payload,omitempty"`
	OIDCClaims         map[string]interface{} `json:"oidc_claims,omitempty"`
	ExpectedExtraction struct {
		DisplayName            string `json:"display_name"`
		Email                  string `json:"email,omitempty"`
		Headline               string `json:"headline,omitempty"`
		Summary                string `json:"summary,omitempty"`
		Location               string `json:"location,omitempty"`
		Industry               string `json:"industry,omitempty"`
		ExperiencesCount       int    `json:"experiences_count"`
		EducationCount         int    `json:"education_count"`
		SkillsCount            int    `json:"skills_count"`
		CertificationsCount    int    `json:"certifications_count,omitempty"`
		SourceMode             string `json:"source_mode"`
		RequiresFallback       bool   `json:"requires_fallback,omitempty"`
		RecommendedFallback    string `json:"recommended_fallback,omitempty"`
		TruthInAdvertisingNote string `json:"truth_in_advertising_note,omitempty"`
	} `json:"expected_extraction"`
}

type profileImportFixture struct {
	Version   string                  `json:"version"`
	Domain    string                  `json:"domain"`
	Scenarios []profileImportScenario `json:"scenarios"`
}

func loadProfileImportFixture(t *testing.T) *profileImportFixture {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_profile_import.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("fixture file must exist: %v", err)
	}

	var f profileImportFixture
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("fixture JSON must parse: %v", err)
	}
	return &f
}

func TestLinkedInProfileImport_AllScenarios(t *testing.T) {
	fixture := loadProfileImportFixture(t)
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	for _, sc := range fixture.Scenarios {
		sc := sc
		t.Run(sc.ScenarioID, func(t *testing.T) {
			var imported *LinkedInImportedProfile
			var err error

			switch sc.InputType {
			case "archive_zip":
				imported, err = svc.ImportProfileArchiveFiles(ctx, "usr-sarah-101", sc.ArchiveFiles)
				if err != nil {
					t.Fatalf("unexpected error importing zip archive files: %v", err)
				}
				if imported.DisplayName != sc.ExpectedExtraction.DisplayName {
					t.Errorf("DisplayName = %q, want %q", imported.DisplayName, sc.ExpectedExtraction.DisplayName)
				}
				if imported.Headline != sc.ExpectedExtraction.Headline {
					t.Errorf("Headline = %q, want %q", imported.Headline, sc.ExpectedExtraction.Headline)
				}
				if imported.Summary != sc.ExpectedExtraction.Summary {
					t.Errorf("Summary = %q, want %q", imported.Summary, sc.ExpectedExtraction.Summary)
				}
				if imported.Location != sc.ExpectedExtraction.Location {
					t.Errorf("Location = %q, want %q", imported.Location, sc.ExpectedExtraction.Location)
				}
				if imported.Industry != sc.ExpectedExtraction.Industry {
					t.Errorf("Industry = %q, want %q", imported.Industry, sc.ExpectedExtraction.Industry)
				}
				if len(imported.Experiences) != sc.ExpectedExtraction.ExperiencesCount {
					t.Errorf("Experiences count = %d, want %d", len(imported.Experiences), sc.ExpectedExtraction.ExperiencesCount)
				}
				if len(imported.Education) != sc.ExpectedExtraction.EducationCount {
					t.Errorf("Education count = %d, want %d", len(imported.Education), sc.ExpectedExtraction.EducationCount)
				}
				if len(imported.Skills) != sc.ExpectedExtraction.SkillsCount {
					t.Errorf("Skills count = %d, want %d", len(imported.Skills), sc.ExpectedExtraction.SkillsCount)
				}
				if len(imported.Certifications) != sc.ExpectedExtraction.CertificationsCount {
					t.Errorf("Certifications count = %d, want %d", len(imported.Certifications), sc.ExpectedExtraction.CertificationsCount)
				}
				if imported.SourceMode != SourceArchiveZip {
					t.Errorf("SourceMode = %q, want %q", imported.SourceMode, SourceArchiveZip)
				}

			case "manual_paste":
				imported, err = svc.ImportPastedProfile(ctx, "usr-david-202", sc.RawText)
				if err != nil {
					t.Fatalf("unexpected error importing pasted text: %v", err)
				}
				if imported.DisplayName != sc.ExpectedExtraction.DisplayName {
					t.Errorf("DisplayName = %q, want %q", imported.DisplayName, sc.ExpectedExtraction.DisplayName)
				}
				if imported.Headline != sc.ExpectedExtraction.Headline {
					t.Errorf("Headline = %q, want %q", imported.Headline, sc.ExpectedExtraction.Headline)
				}
				if imported.Location != sc.ExpectedExtraction.Location {
					t.Errorf("Location = %q, want %q", imported.Location, sc.ExpectedExtraction.Location)
				}
				if len(imported.Experiences) != sc.ExpectedExtraction.ExperiencesCount {
					t.Errorf("Experiences count = %d, want %d", len(imported.Experiences), sc.ExpectedExtraction.ExperiencesCount)
				}
				if len(imported.Education) != sc.ExpectedExtraction.EducationCount {
					t.Errorf("Education count = %d, want %d", len(imported.Education), sc.ExpectedExtraction.EducationCount)
				}
				if len(imported.Skills) != sc.ExpectedExtraction.SkillsCount {
					t.Errorf("Skills count = %d, want %d", len(imported.Skills), sc.ExpectedExtraction.SkillsCount)
				}
				if imported.SourceMode != SourceManualPaste {
					t.Errorf("SourceMode = %q, want %q", imported.SourceMode, SourceManualPaste)
				}

			case "enterprise_api":
				payloadBytes, errMarshal := json.Marshal(sc.APIResponsePayload)
				if errMarshal != nil {
					t.Fatalf("failed to marshal fixture api payload: %v", errMarshal)
				}

				imported, err = svc.ImportEnterpriseProfile(ctx, "usr-elena-303", payloadBytes)
				if err != nil {
					t.Fatalf("unexpected error importing enterprise profile: %v", err)
				}
				if imported.DisplayName != sc.ExpectedExtraction.DisplayName {
					t.Errorf("DisplayName = %q, want %q", imported.DisplayName, sc.ExpectedExtraction.DisplayName)
				}
				if imported.Headline != sc.ExpectedExtraction.Headline {
					t.Errorf("Headline = %q, want %q", imported.Headline, sc.ExpectedExtraction.Headline)
				}
				if len(imported.Experiences) != sc.ExpectedExtraction.ExperiencesCount {
					t.Errorf("Experiences count = %d, want %d", len(imported.Experiences), sc.ExpectedExtraction.ExperiencesCount)
				}
				if len(imported.Education) != sc.ExpectedExtraction.EducationCount {
					t.Errorf("Education count = %d, want %d", len(imported.Education), sc.ExpectedExtraction.EducationCount)
				}
				if len(imported.Skills) != sc.ExpectedExtraction.SkillsCount {
					t.Errorf("Skills count = %d, want %d", len(imported.Skills), sc.ExpectedExtraction.SkillsCount)
				}
				if imported.SourceMode != SourceEnterpriseAPI {
					t.Errorf("SourceMode = %q, want %q", imported.SourceMode, SourceEnterpriseAPI)
				}

			case "oidc_basic_only":
				imported, err = svc.ImportOIDCProfileFallback(ctx, "usr-marcus-404", sc.OIDCClaims)
				if err != nil {
					t.Fatalf("unexpected error in oidc fallback: %v", err)
				}
				if imported.DisplayName != sc.ExpectedExtraction.DisplayName {
					t.Errorf("DisplayName = %q, want %q", imported.DisplayName, sc.ExpectedExtraction.DisplayName)
				}
				if imported.Email != sc.ExpectedExtraction.Email {
					t.Errorf("Email = %q, want %q", imported.Email, sc.ExpectedExtraction.Email)
				}
				if !imported.RequiresFallback {
					t.Errorf("RequiresFallback = false, want true")
				}
				if imported.RecommendedFallback != sc.ExpectedExtraction.RecommendedFallback {
					t.Errorf("RecommendedFallback = %q, want %q", imported.RecommendedFallback, sc.ExpectedExtraction.RecommendedFallback)
				}
				if imported.TruthInAdvertisingNote != sc.ExpectedExtraction.TruthInAdvertisingNote {
					t.Errorf("TruthInAdvertisingNote = %q, want %q", imported.TruthInAdvertisingNote, sc.ExpectedExtraction.TruthInAdvertisingNote)
				}
				if len(imported.Experiences) != 0 {
					t.Errorf("Experiences count = %d, want 0", len(imported.Experiences))
				}
				if len(imported.Education) != 0 {
					t.Errorf("Education count = %d, want 0", len(imported.Education))
				}
				if len(imported.Skills) != 0 {
					t.Errorf("Skills count = %d, want 0", len(imported.Skills))
				}
				if imported.SourceMode != SourceOIDCBasicOnly {
					t.Errorf("SourceMode = %q, want %q", imported.SourceMode, SourceOIDCBasicOnly)
				}

			default:
				t.Fatalf("unhandled scenario input_type: %s", sc.InputType)
			}

			// Verify storage retrieval
			latest, getErr := svc.GetLatestProfileImport(ctx, imported.UserID)
			if getErr != nil {
				t.Fatalf("failed to retrieve latest profile import: %v", getErr)
			}
			if latest.ImportID != imported.ImportID {
				t.Errorf("latest ImportID = %q, want %q", latest.ImportID, imported.ImportID)
			}

			// Verify status update to merged
			merged, mergeErr := svc.MarkProfileImportMerged(ctx, imported.ImportID)
			if mergeErr != nil {
				t.Fatalf("failed to mark profile import merged: %v", mergeErr)
			}
			if merged.Status != "merged" {
				t.Errorf("merged Status = %q, want %q", merged.Status, "merged")
			}
		})
	}
}

func TestLinkedInProfileImport_ValidationErrors(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	// Empty zip archive
	if _, err := svc.ImportProfileArchive(ctx, "usr-1", []byte{}); !errors.Is(err, ErrEmptyArchive) {
		t.Errorf("expected ErrEmptyArchive, got %v", err)
	}

	// Empty pasted text
	if _, err := svc.ImportPastedProfile(ctx, "usr-1", "   \n\t  "); !errors.Is(err, ErrEmptyText) {
		t.Errorf("expected ErrEmptyText, got %v", err)
	}

	// Malformed enterprise payload
	if _, err := svc.ImportEnterpriseProfile(ctx, "usr-1", []byte("invalid-json")); err == nil {
		t.Errorf("expected error for malformed json, got nil")
	}

	// User with no import
	if _, err := svc.GetLatestProfileImport(ctx, "nonexistent-usr"); !errors.Is(err, ErrProfileDataNotFound) {
		t.Errorf("expected ErrProfileDataNotFound, got %v", err)
	}
}
