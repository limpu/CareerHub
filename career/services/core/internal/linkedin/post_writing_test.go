package linkedin_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/social-platform/services/core/internal/linkedin"
)

type postWritingFixture struct {
	Scenarios struct {
		DistributedSystemsContrarianPost struct {
			WorkspaceID   string   `json:"workspace_id"`
			TenantID      string   `json:"tenant_id"`
			Topic         string   `json:"topic"`
			Angle         string   `json:"angle"`
			VerifiedFacts []string `json:"verified_facts"`
			ExpectedHooks []struct {
				HookType  string `json:"hook_type"`
				HookText  string `json:"hook_text"`
				Rationale string `json:"rationale"`
			} `json:"expected_hooks"`
			ExpectedAudit struct {
				Passed                        bool `json:"passed"`
				ReadabilityScoreMin           int  `json:"readability_score_min"`
				CharacterCountMax             int  `json:"character_count_max"`
				BuzzwordViolationsCount       int  `json:"buzzword_violations_count"`
				FalseGuaranteeViolationsCount int  `json:"false_guarantee_violations_count"`
			} `json:"expected_audit"`
		} `json:"distributed_systems_contrarian_post"`

		CareerTransitionLessonsPost struct {
			WorkspaceID   string   `json:"workspace_id"`
			TenantID      string   `json:"tenant_id"`
			Topic         string   `json:"topic"`
			Angle         string   `json:"angle"`
			VerifiedFacts []string `json:"verified_facts"`
			ExpectedAudit struct {
				Passed            bool `json:"passed"`
				CharacterCountMin int  `json:"character_count_min"`
				CharacterCountMax int  `json:"character_count_max"`
			} `json:"expected_audit"`
		} `json:"career_transition_lessons_post"`

		BuzzwordAndFalseGuaranteeViolation struct {
			WorkspaceID       string   `json:"workspace_id"`
			TenantID          string   `json:"tenant_id"`
			Topic             string   `json:"topic"`
			Angle             string   `json:"angle"`
			VerifiedFacts     []string `json:"verified_facts"`
			ViolatingPostText string   `json:"violating_post_text"`
			ExpectedAudit     struct {
				Passed             bool     `json:"passed"`
				ExpectedViolations []string `json:"expected_violations"`
			} `json:"expected_audit"`
		} `json:"buzzword_and_false_guarantee_violation"`

		TamperInvalidationAndMultiTenant struct {
			WorkspaceID              string `json:"workspace_id"`
			TenantID                 string `json:"tenant_id"`
			InitialStatus            string `json:"initial_status"`
			ApprovedStatus           string `json:"approved_status"`
			TamperedStatus           string `json:"tampered_status"`
			BetaWorkspaceID          string `json:"beta_workspace_id"`
			BetaTenantID             string `json:"beta_tenant_id"`
			ExpectedAutoPublishError string `json:"expected_auto_publish_error"`
		} `json:"tamper_invalidation_and_multi_tenant"`
	} `json:"scenarios"`
}

func loadPostWritingFixture(t *testing.T) postWritingFixture {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_post_writing.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read linkedin_post_writing.json fixture: %v", err)
	}

	var fix postWritingFixture
	err = json.Unmarshal(data, &fix)
	if err != nil {
		t.Fatalf("failed to unmarshal post writing fixture: %v", err)
	}
	return fix
}

func TestPostWriting_FixtureGroundTruth(t *testing.T) {
	fix := loadPostWritingFixture(t)
	sc1 := fix.Scenarios.DistributedSystemsContrarianPost

	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	// 1. Contrarian Distributed Systems Post Scenario
	draft, err := svc.GeneratePostDraft(ctx, linkedin.PostWritingRequest{
		WorkspaceID:   sc1.WorkspaceID,
		TenantID:      sc1.TenantID,
		Topic:         sc1.Topic,
		Angle:         linkedin.PostAngle(sc1.Angle),
		VerifiedFacts: sc1.VerifiedFacts,
	})
	if err != nil {
		t.Fatalf("unexpected error generating draft: %v", err)
	}
	if draft == nil {
		t.Fatal("expected draft to not be nil")
	}

	if draft.WorkspaceID != sc1.WorkspaceID {
		t.Errorf("expected workspace %s, got %s", sc1.WorkspaceID, draft.WorkspaceID)
	}
	if draft.TenantID != sc1.TenantID {
		t.Errorf("expected tenant %s, got %s", sc1.TenantID, draft.TenantID)
	}
	if draft.Topic != sc1.Topic {
		t.Errorf("expected topic %s, got %s", sc1.Topic, draft.Topic)
	}
	if draft.Angle != linkedin.PostAngle(sc1.Angle) {
		t.Errorf("expected angle %s, got %s", sc1.Angle, draft.Angle)
	}
	if draft.Status != linkedin.PostDraftPendingApproval {
		t.Errorf("expected status %s, got %s", linkedin.PostDraftPendingApproval, draft.Status)
	}
	if len(draft.AvailableHooks) < 4 {
		t.Errorf("expected at least 4 available hooks, got %d", len(draft.AvailableHooks))
	}

	// Ensure all verified facts appear in body
	for _, f := range sc1.VerifiedFacts {
		if !strings.Contains(draft.FullPostText, f) {
			t.Errorf("expected post to contain verified fact '%s'", f)
		}
	}

	// Verify initial audit passed
	if draft.LatestAudit == nil {
		t.Fatal("expected LatestAudit to be non-nil")
	}
	if !draft.LatestAudit.Passed {
		t.Errorf("expected audit to pass, got failed with %d issues", len(draft.LatestAudit.Issues))
	}
	if draft.LatestAudit.ReadabilityScore < sc1.ExpectedAudit.ReadabilityScoreMin {
		t.Errorf("expected readability >= %d, got %d", sc1.ExpectedAudit.ReadabilityScoreMin, draft.LatestAudit.ReadabilityScore)
	}
	if draft.LatestAudit.CharacterCount > sc1.ExpectedAudit.CharacterCountMax {
		t.Errorf("expected character count <= %d, got %d", sc1.ExpectedAudit.CharacterCountMax, draft.LatestAudit.CharacterCount)
	}
	if draft.LatestAudit.BuzzwordsCount != sc1.ExpectedAudit.BuzzwordViolationsCount {
		t.Errorf("expected buzzword violations %d, got %d", sc1.ExpectedAudit.BuzzwordViolationsCount, draft.LatestAudit.BuzzwordsCount)
	}
	if draft.LatestAudit.ViralityClaimsCount != sc1.ExpectedAudit.FalseGuaranteeViolationsCount {
		t.Errorf("expected false guarantee violations %d, got %d", sc1.ExpectedAudit.FalseGuaranteeViolationsCount, draft.LatestAudit.ViralityClaimsCount)
	}
	if draft.LatestAudit.Disclaimer == "" {
		t.Error("expected non-empty disclaimer")
	}

	// 2. Career Transition Lessons Scenario
	sc2 := fix.Scenarios.CareerTransitionLessonsPost
	draft2, err := svc.GeneratePostDraft(ctx, linkedin.PostWritingRequest{
		WorkspaceID:   sc2.WorkspaceID,
		TenantID:      sc2.TenantID,
		Topic:         sc2.Topic,
		Angle:         linkedin.PostAngle(sc2.Angle),
		VerifiedFacts: sc2.VerifiedFacts,
	})
	if err != nil {
		t.Fatalf("unexpected error generating draft2: %v", err)
	}
	if !draft2.LatestAudit.Passed {
		t.Error("expected draft2 audit to pass")
	}
	if draft2.LatestAudit.CharacterCount < sc2.ExpectedAudit.CharacterCountMin {
		t.Errorf("expected char count >= %d, got %d", sc2.ExpectedAudit.CharacterCountMin, draft2.LatestAudit.CharacterCount)
	}
	if draft2.LatestAudit.CharacterCount > sc2.ExpectedAudit.CharacterCountMax {
		t.Errorf("expected char count <= %d, got %d", sc2.ExpectedAudit.CharacterCountMax, draft2.LatestAudit.CharacterCount)
	}

	// 3. Buzzwords, Fake Virality and Unverified Metric Violation Scenario
	sc3 := fix.Scenarios.BuzzwordAndFalseGuaranteeViolation
	violatingDraft := linkedin.PostDraft{
		DraftID:           "violating-001",
		WorkspaceID:       sc3.WorkspaceID,
		TenantID:          sc3.TenantID,
		Topic:             sc3.Topic,
		Angle:             linkedin.PostAngle(sc3.Angle),
		FullPostText:      sc3.ViolatingPostText,
		VerifiedFactsUsed: sc3.VerifiedFacts,
	}

	auditReport := linkedin.AuditPostDraft(violatingDraft, sc3.VerifiedFacts)
	if auditReport.Passed {
		t.Error("violating draft must fail editorial audit")
	}
	if auditReport.BuzzwordsCount < 1 {
		t.Error("expected at least 1 buzzword violation")
	}
	if auditReport.ViralityClaimsCount < 1 {
		t.Error("expected at least 1 virality claim violation")
	}
	if auditReport.AiBypassClaimsCount < 1 {
		t.Error("expected at least 1 AI bypass claim violation")
	}
	if auditReport.UnverifiedMetricsCount < 1 {
		t.Error("expected at least 1 unverified metric violation")
	}

	issueCodes := make(map[string]bool)
	for _, iss := range auditReport.Issues {
		issueCodes[iss.Code] = true
	}
	for _, expectedCode := range sc3.ExpectedAudit.ExpectedViolations {
		if !issueCodes[expectedCode] {
			t.Errorf("audit must contain violation code: %s", expectedCode)
		}
	}
}

func TestPostWriting_ApprovalAndTamperInvalidation(t *testing.T) {
	fix := loadPostWritingFixture(t)
	sc := fix.Scenarios.TamperInvalidationAndMultiTenant

	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	// Create valid draft
	draft, err := svc.GeneratePostDraft(ctx, linkedin.PostWritingRequest{
		WorkspaceID:   sc.WorkspaceID,
		TenantID:      sc.TenantID,
		Topic:         "Clean Architecture in Go",
		Angle:         linkedin.AngleTechnicalDeepDive,
		VerifiedFacts: []string{"Designed domain-driven microservices handling 50k QPS"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draft.Status != linkedin.PostDraftStatus(sc.InitialStatus) {
		t.Errorf("expected initial status %s, got %s", sc.InitialStatus, draft.Status)
	}
	if draft.ApprovalToken != "" {
		t.Error("approval token must be empty initially")
	}

	// Approve draft
	approved, err := svc.ApprovePostDraft(ctx, draft.DraftID, sc.TenantID, "user-alice")
	if err != nil {
		t.Fatalf("unexpected approval error: %v", err)
	}
	if approved.Status != linkedin.PostDraftStatus(sc.ApprovedStatus) {
		t.Errorf("expected approved status %s, got %s", sc.ApprovedStatus, approved.Status)
	}
	if approved.ApprovalToken == "" {
		t.Error("expected non-empty approval token")
	}
	if approved.ApprovedBy != "user-alice" {
		t.Errorf("expected approved by user-alice, got %s", approved.ApprovedBy)
	}
	if approved.ApprovedAt == nil {
		t.Error("approved timestamp must not be nil")
	}

	// Verify token validity
	tokenValid := linkedin.VerifyPostApprovalToken(*approved, approved.ApprovalToken, "")
	if !tokenValid {
		t.Error("token must verify against approved draft")
	}

	// Tamper with draft text -> must invalidate approval (FND-010, AT-007)
	tampered, err := svc.UpdatePostDraftText(ctx, draft.DraftID, sc.TenantID, "story", "New Hook", "Updated full post text with 50k QPS.")
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if tampered.Status != linkedin.PostDraftStatus(sc.TamperedStatus) {
		t.Errorf("expected status %s after tamper, got %s", sc.TamperedStatus, tampered.Status)
	}
	if tampered.ApprovalToken != "" {
		t.Error("approval token must be cleared upon text modification")
	}
	if tampered.ApprovedAt != nil {
		t.Error("approved timestamp must be cleared")
	}

	// Attempting to copy unapproved draft must fail
	_, err = svc.MarkPostDraftCopied(ctx, draft.DraftID, sc.TenantID)
	if err == nil || !strings.Contains(err.Error(), "must be approved prior to clipboard copy") {
		t.Errorf("expected error copying unapproved draft, got: %v", err)
	}

	// Re-approve after tamper
	reapproved, err := svc.ApprovePostDraft(ctx, draft.DraftID, sc.TenantID, "user-alice")
	if err != nil {
		t.Fatalf("unexpected re-approval error: %v", err)
	}
	if reapproved.Status != linkedin.PostDraftApproved {
		t.Errorf("expected approved status, got %s", reapproved.Status)
	}
	if reapproved.ApprovalToken == "" {
		t.Error("expected non-empty approval token after re-approval")
	}

	// Copy to clipboard now succeeds
	copied, err := svc.MarkPostDraftCopied(ctx, draft.DraftID, sc.TenantID)
	if err != nil {
		t.Fatalf("unexpected copy error: %v", err)
	}
	if copied.Status != linkedin.PostDraftCopiedToClipboard {
		t.Errorf("expected status copied_to_clipboard, got %s", copied.Status)
	}
}

func TestPostWriting_RejectDraft(t *testing.T) {
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	draft, err := svc.GeneratePostDraft(ctx, linkedin.PostWritingRequest{
		WorkspaceID:   "ws-alpha",
		TenantID:      "tenant-alpha",
		Topic:         "Event Driven Systems",
		Angle:         linkedin.AngleActionableGuide,
		VerifiedFacts: []string{"Built event bus processing 100M events/day"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rejected, err := svc.RejectPostDraft(ctx, draft.DraftID, "tenant-alpha", "Not aligned with current sprint theme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rejected.Status != linkedin.PostDraftRejected {
		t.Errorf("expected status rejected, got %s", rejected.Status)
	}
}

func TestPostWriting_CrossTenantIsolation(t *testing.T) {
	fix := loadPostWritingFixture(t)
	sc := fix.Scenarios.TamperInvalidationAndMultiTenant

	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	draft, err := svc.GeneratePostDraft(ctx, linkedin.PostWritingRequest{
		WorkspaceID:   sc.WorkspaceID,
		TenantID:      sc.TenantID,
		Topic:         "Confidential Architecture Migration",
		Angle:         linkedin.AngleMilestoneCelebration,
		VerifiedFacts: []string{"Completed zero-downtime DB migration"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Tenant beta attempts to get draft from tenant alpha -> denied (AT-011, AT-012)
	_, err = svc.GetPostDraft(ctx, draft.DraftID, sc.BetaTenantID)
	if err != linkedin.ErrCrossTenantAccessDenied {
		t.Errorf("expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// Tenant beta attempts to update draft -> denied
	_, err = svc.UpdatePostDraftText(ctx, draft.DraftID, sc.BetaTenantID, "punchy", "Hacked Hook", "Hacked text")
	if err != linkedin.ErrCrossTenantAccessDenied {
		t.Errorf("expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// Tenant beta attempts to delete draft -> denied
	err = svc.DeletePostDraft(ctx, draft.DraftID, sc.BetaTenantID)
	if err != linkedin.ErrCrossTenantAccessDenied {
		t.Errorf("expected ErrCrossTenantAccessDenied, got %v", err)
	}
}

func TestPostWriting_FailClosedAutoPublish(t *testing.T) {
	err := linkedin.RejectUnsupportedAutoPublish()
	if err != linkedin.ErrUnsupportedAutoPublish {
		t.Errorf("expected ErrUnsupportedAutoPublish, got %v", err)
	}
}

func TestPostWriting_ValidatePostText(t *testing.T) {
	err := linkedin.ValidatePostText("")
	if err != linkedin.ErrPostDraftTextEmpty {
		t.Errorf("expected ErrPostDraftTextEmpty, got %v", err)
	}

	err = linkedin.ValidatePostText("   ")
	if err != linkedin.ErrPostDraftTextEmpty {
		t.Errorf("expected ErrPostDraftTextEmpty, got %v", err)
	}

	tooLong := strings.Repeat("A", 3001)
	err = linkedin.ValidatePostText(tooLong)
	if err != linkedin.ErrPostExceedsCharacterLimit {
		t.Errorf("expected ErrPostExceedsCharacterLimit, got %v", err)
	}

	valid := "This is a clean, compliant LinkedIn post about engineering."
	err = linkedin.ValidatePostText(valid)
	if err != nil {
		t.Errorf("expected nil error for valid text, got %v", err)
	}
}

func TestPostWriting_AllAngles(t *testing.T) {
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	angles := []linkedin.PostAngle{
		linkedin.AngleContrarianInsight,
		linkedin.AngleLessonLearnedBreakdown,
		linkedin.AngleTechnicalDeepDive,
		linkedin.AngleMilestoneCelebration,
		linkedin.AngleActionableGuide,
	}

	for _, angle := range angles {
		draft, err := svc.GeneratePostDraft(ctx, linkedin.PostWritingRequest{
			WorkspaceID:   "ws-alpha",
			TenantID:      "tenant-alpha",
			Topic:         "Reliability Engineering",
			Angle:         angle,
			VerifiedFacts: []string{"Reduced incident MTTR by 65%"},
		})
		if err != nil {
			t.Fatalf("failed to generate draft for angle %s: %v", angle, err)
		}
		if draft.Angle != angle {
			t.Errorf("expected angle %s, got %s", angle, draft.Angle)
		}
		if !strings.Contains(draft.FullPostText, "Reduced incident MTTR by 65%") {
			t.Errorf("expected fact in post text for angle %s", angle)
		}
		if !draft.LatestAudit.Passed {
			t.Errorf("expected audit to pass for angle %s", angle)
		}
	}
}

func TestPostWriting_ListFilters(t *testing.T) {
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	// Create 2 drafts
	draft1, err := svc.GeneratePostDraft(ctx, linkedin.PostWritingRequest{
		WorkspaceID:   "ws-alpha",
		TenantID:      "tenant-alpha",
		Topic:         "Kubernetes Optimization",
		Angle:         linkedin.AngleTechnicalDeepDive,
		VerifiedFacts: []string{"Cut cluster cost by 30%"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	draft2, err := svc.GeneratePostDraft(ctx, linkedin.PostWritingRequest{
		WorkspaceID:   "ws-alpha",
		TenantID:      "tenant-alpha",
		Topic:         "Leadership Lessons",
		Angle:         linkedin.AngleLessonLearnedBreakdown,
		VerifiedFacts: []string{"Mentored 8 senior engineers"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// List all
	all, err := svc.ListPostDrafts(ctx, linkedin.PostDraftFilter{
		WorkspaceID: "ws-alpha",
		TenantID:    "tenant-alpha",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 drafts, got %d", len(all))
	}

	// Filter by angle
	onlyDeepDive, err := svc.ListPostDrafts(ctx, linkedin.PostDraftFilter{
		WorkspaceID: "ws-alpha",
		TenantID:    "tenant-alpha",
		Angle:       linkedin.AngleTechnicalDeepDive,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(onlyDeepDive) != 1 {
		t.Errorf("expected 1 deep dive draft, got %d", len(onlyDeepDive))
	}
	if len(onlyDeepDive) > 0 && onlyDeepDive[0].DraftID != draft1.DraftID {
		t.Errorf("expected draft1 ID %s, got %s", draft1.DraftID, onlyDeepDive[0].DraftID)
	}

	// Filter by query
	queryRes, err := svc.ListPostDrafts(ctx, linkedin.PostDraftFilter{
		WorkspaceID: "ws-alpha",
		TenantID:    "tenant-alpha",
		Query:       "leadership",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queryRes) != 1 {
		t.Errorf("expected 1 query result, got %d", len(queryRes))
	}
	if len(queryRes) > 0 && queryRes[0].DraftID != draft2.DraftID {
		t.Errorf("expected draft2 ID %s, got %s", draft2.DraftID, queryRes[0].DraftID)
	}
}
