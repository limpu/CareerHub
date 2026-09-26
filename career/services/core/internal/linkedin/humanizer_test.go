package linkedin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type HumanizerFixtureSuite struct {
	Scenarios []struct {
		ScenarioID             string   `json:"scenario_id"`
		WorkspaceID            string   `json:"workspace_id"`
		IsolatedWorkspaceID    string   `json:"isolated_workspace_id"`
		TenantID               string   `json:"tenant_id"`
		IsolatedTenantID       string   `json:"isolated_tenant_id"`
		Profile                *VoiceProfile `json:"profile"`
		VoiceProfileID         string   `json:"voice_profile_id"`
		CandidateVerifiedFacts []string `json:"candidate_verified_facts"`
		InputText              string   `json:"input_text"`
		ExpectedCleanedText    string   `json:"expected_cleaned_text"`
		ExpectedSlopDetected   []string `json:"expected_slop_detected"`
		ExpectedSlopScoreBefore int     `json:"expected_slop_score_before"`
		ExpectedSlopScoreAfter  int     `json:"expected_slop_score_after"`
		ExpectedPassed         bool     `json:"expected_passed"`
		ExpectedBlockingIssues []struct {
			Code     string `json:"code"`
			Severity string `json:"severity"`
			Snippet  string `json:"snippet"`
		} `json:"expected_blocking_issues"`
		OriginalApprovedText   string `json:"original_approved_text"`
		OriginalApprovalToken  string `json:"original_approval_token"`
		TamperedText           string `json:"tampered_text"`
		ExpectedTamperInvalidated bool `json:"expected_tamper_invalidated"`
	} `json:"scenarios"`
}

func loadHumanizerFixtures(t *testing.T) *HumanizerFixtureSuite {
	t.Helper()
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_humanizer.json")
	bytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read linkedin_humanizer.json fixture: %v", err)
	}
	var suite HumanizerFixtureSuite
	if err := json.Unmarshal(bytes, &suite); err != nil {
		t.Fatalf("Failed to unmarshal linkedin_humanizer.json: %v", err)
	}
	return &suite
}

func TestHumanizer_FixtureGroundTruth(t *testing.T) {
	suite := loadHumanizerFixtures(t)
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	for _, sc := range suite.Scenarios {
		t.Run(sc.ScenarioID, func(t *testing.T) {
			switch sc.ScenarioID {
			case "staff_architect_voice_profile":
				if sc.Profile == nil {
					t.Fatalf("Expected profile in scenario 1")
				}
				saved, err := svc.CreateVoiceProfile(ctx, sc.Profile)
				if err != nil {
					t.Fatalf("Failed to save voice profile: %v", err)
				}
				if saved.ProfileID != sc.Profile.ProfileID {
					t.Errorf("Expected profile ID %s, got %s", sc.Profile.ProfileID, saved.ProfileID)
				}
				approved, err := svc.ApproveVoiceProfile(ctx, saved.ProfileID, "candidate-user-01", sc.TenantID)
				if err != nil {
					t.Fatalf("Failed to approve voice profile: %v", err)
				}
				if approved.Status != VoiceProfileApproved {
					t.Errorf("Expected approved status, got %s", approved.Status)
				}
				if !strings.HasPrefix(approved.ApprovalToken, "hmac-sha256:") {
					t.Errorf("Expected hmac-sha256 token, got %s", approved.ApprovalToken)
				}
				if !VerifyVoiceApprovalToken(approved, approved.ApprovalToken, "") {
					t.Errorf("Voice approval token verification failed")
				}

			case "multi_tier_deslop_transformation":
				profile := &VoiceProfile{
					ProfileID:        sc.VoiceProfileID,
					WorkspaceID:      sc.WorkspaceID,
					TenantID:         sc.TenantID,
					ProfileName:      "Test Architect Profile",
					CadenceStyle:     CadenceBalancedRhythm,
					BlacklistedTerms: []string{"paradigm shift"},
				}
				_ = repo.SaveVoiceProfile(ctx, profile)

				res, err := svc.HumanizeDraftContent(ctx, sc.InputText, sc.VoiceProfileID, sc.WorkspaceID, sc.TenantID, nil)
				if err != nil {
					t.Fatalf("Failed to humanize draft: %v", err)
				}

				if res.SlopScoreBefore < 70 {
					t.Errorf("Expected heavy slop before (>=70), got %d", res.SlopScoreBefore)
				}
				if res.SlopScoreAfter > 10 {
					t.Errorf("Expected cleaned slop score (<=10), got %d", res.SlopScoreAfter)
				}
				if !res.Passed {
					t.Errorf("Expected passed = true, got false")
				}

				// Verify expected slop keywords were detected
				detected := make(map[string]bool)
				for _, repl := range res.Tier1SlopReplacements {
					detected[strings.ToLower(repl.OffendingPhrase)] = true
				}
				for _, exp := range sc.ExpectedSlopDetected {
					if !detected[strings.ToLower(exp)] {
						t.Errorf("Expected slop detection for '%s', but not found", exp)
					}
				}

			case "invented_anecdote_violation":
				res, err := svc.HumanizeDraftContent(ctx, sc.InputText, sc.VoiceProfileID, sc.WorkspaceID, sc.TenantID, sc.CandidateVerifiedFacts)
				if err != nil {
					t.Fatalf("Failed to humanize draft: %v", err)
				}
				if res.Passed != sc.ExpectedPassed {
					t.Errorf("Expected passed = %v, got %v", sc.ExpectedPassed, res.Passed)
				}
				if len(res.Tier3Issues) < 2 {
					t.Fatalf("Expected at least 2 blocking issues, got %d", len(res.Tier3Issues))
				}

				hasAnecdote := false
				hasMetric := false
				for _, iss := range res.Tier3Issues {
					if iss.Code == "invented_personal_anecdote" {
						hasAnecdote = true
					}
					if iss.Code == "unverified_metric" {
						hasMetric = true
					}
				}
				if !hasAnecdote {
					t.Errorf("Expected invented_personal_anecdote issue flag")
				}
				if !hasMetric {
					t.Errorf("Expected unverified_metric issue flag")
				}

				// Verify approving blocked draft fails
				_, err = svc.ApproveHumanizedDraft(ctx, res.ResultID, "candidate-user-01", sc.TenantID)
				if err == nil {
					t.Errorf("Expected error approving draft with blocking issues under AT-003, but got nil")
				}

			case "tamper_invalidation_and_multi_tenant":
				res, err := svc.HumanizeDraftContent(ctx, sc.OriginalApprovedText, sc.VoiceProfileID, sc.WorkspaceID, sc.TenantID, nil)
				if err != nil {
					t.Fatalf("Failed to humanize draft: %v", err)
				}

				approved, err := svc.ApproveHumanizedDraft(ctx, res.ResultID, "candidate-user-01", sc.TenantID)
				if err != nil {
					t.Fatalf("Failed to approve draft: %v", err)
				}
				if approved.Status != "approved" || approved.ApprovalToken == "" {
					t.Fatalf("Expected approved status and token")
				}

				// Cross-tenant access denial
				_, err = svc.ApproveHumanizedDraft(ctx, res.ResultID, "intruder", sc.IsolatedTenantID)
				if err != ErrCrossTenantAccessDenied {
					t.Errorf("Expected ErrCrossTenantAccessDenied for isolated tenant, got %v", err)
				}

				// Tamper with text
				tampered, err := svc.UpdateHumanizedText(ctx, res.ResultID, sc.TamperedText, sc.TenantID, nil)
				if err != nil {
					t.Fatalf("Failed to update humanized text: %v", err)
				}
				if tampered.Status != "draft" {
					t.Errorf("Expected status reset to draft after tampering, got %s", tampered.Status)
				}
				if tampered.ApprovalToken != "" {
					t.Errorf("Expected approval token wiped upon tamper, got %s", tampered.ApprovalToken)
				}

				// Attempting to mark copied without re-approval should fail
				_, err = svc.MarkHumanizedDraftCopied(ctx, res.ResultID, sc.TenantID)
				if err == nil {
					t.Errorf("Expected error copying unapproved tampered draft, got nil")
				}
			}
		})
	}
}

func TestHumanizer_VoiceProfileCRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	profile := &VoiceProfile{
		ProfileID:              "vp-test-001",
		WorkspaceID:            "ws-alpha",
		TenantID:               "tenant-alpha",
		ProfileName:            "Test Lead Persona",
		Formality:              3,
		TechnicalDepth:         4,
		CadenceStyle:           CadencePunchyStaccato,
		Perspective:            PerspectiveCollectiveTeam,
		PreferredTerms:         []string{"resilience", "kubernetes"},
		BlacklistedTerms:       []string{"guru", "ninja"},
		ApprovedWritingSamples: []string{"Sample text 1"},
	}

	// 1. Create
	created, err := svc.CreateVoiceProfile(ctx, profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}
	if created.Status != VoiceProfileDraft {
		t.Errorf("Expected draft status, got %s", created.Status)
	}

	// 2. Cross-tenant isolation
	_, err = svc.GetVoiceProfile(ctx, profile.ProfileID, "tenant-beta")
	if err != ErrCrossTenantAccessDenied {
		t.Errorf("Expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// 3. Approve
	approved, err := svc.ApproveVoiceProfile(ctx, profile.ProfileID, "candidate-user", "tenant-alpha")
	if err != nil {
		t.Fatalf("Failed to approve profile: %v", err)
	}
	if approved.Status != VoiceProfileApproved || approved.ApprovalToken == "" {
		t.Errorf("Expected approved status and token")
	}

	// 4. Update -> Tamper Invalidation
	approved.Formality = 5
	updated, err := svc.UpdateVoiceProfile(ctx, approved, "tenant-alpha")
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}
	if updated.Status != VoiceProfileDraft || updated.ApprovalToken != "" {
		t.Errorf("Expected status reset to draft and token wiped on update")
	}

	// 5. List
	list, err := svc.ListVoiceProfiles(ctx, "ws-alpha", "tenant-alpha")
	if err != nil || len(list) != 1 {
		t.Errorf("Expected 1 profile in list, got %d, err: %v", len(list), err)
	}

	// 6. Delete
	if err := svc.DeleteVoiceProfile(ctx, profile.ProfileID, "tenant-alpha"); err != nil {
		t.Fatalf("Failed to delete profile: %v", err)
	}
	_, err = svc.GetVoiceProfile(ctx, profile.ProfileID, "tenant-alpha")
	if err != ErrVoiceProfileNotFound {
		t.Errorf("Expected ErrVoiceProfileNotFound, got %v", err)
	}
}

func TestHumanizer_CadenceStyles(t *testing.T) {
	shortPunchyText := "We built the consensus engine. Latency dropped sharply. Uptime stayed at four nines."
	notesPunchy := AnalyzeCadence(shortPunchyText, CadencePunchyStaccato)
	if len(notesPunchy) == 0 {
		t.Errorf("Expected cadence notes for punchy staccato")
	}

	longAnalyticalText := "In our architectural retrospective across multiple cloud regions, we thoroughly evaluated whether raft consensus leases adequately guarded against asymmetric partition drift under bursty query traffic."
	notesAnalytical := AnalyzeCadence(longAnalyticalText, CadenceAnalyticalDeep)
	if len(notesAnalytical) == 0 {
		t.Errorf("Expected cadence notes for analytical deep")
	}
}
