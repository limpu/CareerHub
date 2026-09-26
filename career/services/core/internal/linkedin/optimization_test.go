package linkedin

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type optimizationScenarioFixture struct {
	ScenarioID     string `json:"scenario_id"`
	CandidateName  string `json:"candidate_name"`
	TargetRole     string `json:"target_role"`
	CurrentProfile struct {
		Headline    string `json:"headline"`
		About       string `json:"about"`
		Experiences []struct {
			CompanyName string   `json:"company_name"`
			Title       string   `json:"title"`
			Description string   `json:"description"`
			SkillsUsed  []string `json:"skills_used"`
		} `json:"experiences"`
		Featured string   `json:"featured"`
		Skills   []string `json:"skills"`
	} `json:"current_profile"`
	VerifiedSourceFacts  []string `json:"verified_source_facts"`
	ExpectedOptimizations []struct {
		Section         string   `json:"section"`
		Before          string   `json:"before"`
		After           string   `json:"after"`
		Rationale       string   `json:"rationale"`
		SourceFactsUsed []string `json:"source_facts_used"`
		ImpactScore     int      `json:"impact_score"`
	} `json:"expected_optimizations"`
}

type optimizationFixtureFile struct {
	Version   string                        `json:"version"`
	Domain    string                        `json:"domain"`
	Scenarios []optimizationScenarioFixture `json:"scenarios"`
}

func loadOptimizationFixture(t *testing.T) *optimizationFixtureFile {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_profile_optimization.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read optimization fixture: %v", err)
	}

	var file optimizationFixtureFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("failed to unmarshal optimization fixture: %v", err)
	}
	return &file
}

func TestLinkedInProfileOptimization_AllScenarios(t *testing.T) {
	fixture := loadOptimizationFixture(t)
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	for _, sc := range fixture.Scenarios {
		sc := sc
		t.Run(sc.ScenarioID, func(t *testing.T) {
			var exps []LinkedInImportedExperience
			for _, e := range sc.CurrentProfile.Experiences {
				exps = append(exps, LinkedInImportedExperience{
					CompanyName: e.CompanyName,
					Title:       e.Title,
					Description: e.Description,
				})
			}

			input := CandidateOptimizationInput{
				UserID:              "usr-" + sc.ScenarioID,
				CandidateName:       sc.CandidateName,
				TargetRole:          sc.TargetRole,
				CurrentHeadline:     sc.CurrentProfile.Headline,
				CurrentAbout:        sc.CurrentProfile.About,
				Experiences:         exps,
				Skills:              sc.CurrentProfile.Skills,
				Featured:            sc.CurrentProfile.Featured,
				VerifiedSourceFacts: sc.VerifiedSourceFacts,
			}

			report, err := svc.GenerateOptimizationReport(ctx, input)
			if err != nil {
				t.Fatalf("unexpected error generating report: %v", err)
			}
			if report.ReportID == "" {
				t.Errorf("expected non-empty ReportID")
			}
			if report.OverallProfileScore < 60 || report.OverallProfileScore > 100 {
				t.Errorf("expected reasonable profile score, got %d", report.OverallProfileScore)
			}
			if len(report.Suggestions) == 0 {
				t.Fatalf("expected suggestions, got 0")
			}

			// Validate expected optimizations
			for _, expected := range sc.ExpectedOptimizations {
				var match *SectionSuggestion
				for i := range report.Suggestions {
					if string(report.Suggestions[i].Section) == expected.Section {
						match = &report.Suggestions[i]
						break
					}
				}

				if match == nil {
					t.Errorf("section %q not found in generated suggestions", expected.Section)
					continue
				}

				if match.Before != expected.Before {
					t.Errorf("section %q: Before = %q, want %q", expected.Section, match.Before, expected.Before)
				}
				if match.After != expected.After {
					t.Errorf("section %q: After = %q, want %q", expected.Section, match.After, expected.After)
				}
				if match.ApprovalState != StatePending {
					t.Errorf("initial state should be pending, got %s", match.ApprovalState)
				}

				// Check fact grounding (AT-003): all source facts used must be from verified list
				for _, usedFact := range match.SourceFactsUsed {
					found := false
					for _, vf := range sc.VerifiedSourceFacts {
						if strings.EqualFold(usedFact, vf) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("AT-003 violation: used unverified fact %q", usedFact)
					}
				}
			}

			// Test persistence retrieval
			latest, getErr := svc.GetLatestOptimizationReport(ctx, input.UserID)
			if getErr != nil {
				t.Fatalf("failed to retrieve saved report: %v", getErr)
			}
			if latest.ReportID != report.ReportID {
				t.Errorf("retrieved report ID = %q, want %q", latest.ReportID, report.ReportID)
			}
		})
	}
}

func TestLinkedInProfileOptimization_ApprovalWorkflow_AT007(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	input := CandidateOptimizationInput{
		UserID:              "usr-candidate-99",
		CandidateName:       "Sarah Connor",
		TargetRole:          "Staff Infrastructure Engineer",
		CurrentHeadline:     "Software Engineer at Nexus Cloud",
		CurrentAbout:        "Experienced backend developer.",
		VerifiedSourceFacts: []string{"10+ years experience in distributed cloud infrastructure"},
	}

	report, err := svc.GenerateOptimizationReport(ctx, input)
	if err != nil {
		t.Fatalf("failed to generate report: %v", err)
	}

	if len(report.Suggestions) == 0 {
		t.Fatalf("expected suggestions, got 0")
	}

	sugID := report.Suggestions[0].SuggestionID

	// 1. Approve recommendation
	approved, err := svc.ApproveOptimizationSuggestion(ctx, report.ReportID, sugID)
	if err != nil {
		t.Fatalf("failed to approve suggestion: %v", err)
	}
	if approved.ApprovalState != StateApproved {
		t.Errorf("state = %s, want %s", approved.ApprovalState, StateApproved)
	}
	if approved.ApprovedAt == nil {
		t.Errorf("expected non-nil ApprovedAt timestamp")
	}

	// 2. Custom edit invalidates prior approval (AT-007)
	customText := "Staff Infrastructure Engineer | Distributed Systems & High-Scale Cloud"
	edited, err := svc.CustomEditOptimizationSuggestion(ctx, report.ReportID, sugID, customText)
	if err != nil {
		t.Fatalf("failed to custom edit suggestion: %v", err)
	}
	if edited.ApprovalState != StateCustomEdited {
		t.Errorf("state = %s, want %s", edited.ApprovalState, StateCustomEdited)
	}
	if edited.CustomOverride != customText {
		t.Errorf("customOverride = %q, want %q", edited.CustomOverride, customText)
	}

	// 3. Reject recommendation
	rejected, err := svc.RejectOptimizationSuggestion(ctx, report.ReportID, sugID)
	if err != nil {
		t.Fatalf("failed to reject suggestion: %v", err)
	}
	if rejected.ApprovalState != StateRejected {
		t.Errorf("state = %s, want %s", rejected.ApprovalState, StateRejected)
	}

	// 4. Invalid suggestion ID
	_, err = svc.ApproveOptimizationSuggestion(ctx, report.ReportID, "invalid-sug-id")
	if !errors.Is(err, ErrSuggestionNotFound) {
		t.Errorf("expected ErrSuggestionNotFound, got %v", err)
	}

	// 5. Empty custom edit
	_, err = svc.CustomEditOptimizationSuggestion(ctx, report.ReportID, sugID, "   ")
	if !errors.Is(err, ErrEmptyCustomOverride) {
		t.Errorf("expected ErrEmptyCustomOverride, got %v", err)
	}
}
