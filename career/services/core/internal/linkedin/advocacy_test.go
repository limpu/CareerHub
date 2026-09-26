package linkedin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type advocacyFixtureFile struct {
	Scenarios []struct {
		ScenarioID          string            `json:"scenario_id"`
		Description         string            `json:"description"`
		TenantID            string            `json:"tenant_id"`
		WorkspaceID         string            `json:"workspace_id"`
		Campaign            *AdvocacyCampaign `json:"campaign,omitempty"`
		ExpectedApprovalVal bool              `json:"expected_approval_valid"`
		ExpectedForbidden   bool              `json:"expected_forbidden_word_detected"`
		PodRequest          *struct {
			CampaignID               string   `json:"campaign_id"`
			TargetPostURN            string   `json:"target_post_urn"`
			Action                   string   `json:"action"`
			ParticipatingEmployeeIDs []string `json:"participating_employee_ids"`
			SyntheticComments        []string `json:"synthetic_comments"`
		} `json:"pod_request,omitempty"`
		ExpectedError        string `json:"expected_error,omitempty"`
		OriginalCampaignID   string `json:"original_campaign_id,omitempty"`
		TamperedModification *struct {
			VariantID    string `json:"variant_id"`
			ModifiedText string `json:"modified_text"`
		} `json:"tampered_modification,omitempty"`
		ExpectedForbiddenKWs     []string `json:"expected_forbidden_keywords,omitempty"`
		ExpectedStatusAfter      string   `json:"expected_status_after_tamper,omitempty"`
		ExpectedTokenCleared     bool     `json:"expected_approval_token_cleared,omitempty"`
		PrimaryWorkspace         string   `json:"primary_workspace,omitempty"`
		IntruderWorkspace        string   `json:"intruder_workspace,omitempty"`
		ExpectedAccessDenied     bool     `json:"expected_access_denied,omitempty"`
	} `json:"scenarios"`
}

func loadAdvocacyFixture(t *testing.T) advocacyFixtureFile {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_employee_advocacy.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read advocacy fixture: %v", err)
	}
	var f advocacyFixtureFile
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("failed to unmarshal advocacy fixture: %v", err)
	}
	return f
}

func TestAdvocacy_FixtureGroundTruth(t *testing.T) {
	fixture := loadAdvocacyFixture(t)
	if len(fixture.Scenarios) != 4 {
		t.Fatalf("expected 4 scenarios, got %d", len(fixture.Scenarios))
	}

	secret := "test-advocacy-secret"

	// Scenario 1: Approved campaign with brand governance & variants
	sc1 := fixture.Scenarios[0]
	if sc1.ScenarioID != "engineering_launch_advocacy_campaign" {
		t.Errorf("expected sc1 ID 'engineering_launch_advocacy_campaign', got %s", sc1.ScenarioID)
	}
	camp := sc1.Campaign
	if camp == nil || len(camp.Variants) != 3 {
		t.Fatalf("expected 3 variants in sc1, got %v", camp)
	}
	// Verify no forbidden keywords in approved copy
	for _, v := range camp.Variants {
		violations := CheckBrandCompliance(v.SuggestedText, camp.Governance)
		if len(violations) > 0 {
			t.Errorf("unexpected forbidden keyword in variant %s: %v", v.VariantID, violations)
		}
	}
	token := ComputeCampaignApprovalToken(*camp, secret)
	camp.ApprovalToken = token
	if !VerifyCampaignApprovalToken(*camp, secret) {
		t.Errorf("expected approval token to be valid for sc1")
	}

	// Scenario 2: Coordinated engagement pod rejection
	sc2 := fixture.Scenarios[1]
	podReq := CoordinatedEngagementRequest{
		CampaignID:                sc2.PodRequest.CampaignID,
		TargetPostURN:             sc2.PodRequest.TargetPostURN,
		Action:                    sc2.PodRequest.Action,
		ParticipatingEmployeeIDs: sc2.PodRequest.ParticipatingEmployeeIDs,
		SyntheticComments:         sc2.PodRequest.SyntheticComments,
	}
	err := RejectCoordinatedEngagementPod(podReq)
	if err == nil || err != ErrCoordinatedEngagementProhibited {
		t.Errorf("expected ErrCoordinatedEngagementProhibited, got %v", err)
	}

	// Scenario 3: Tamper invalidation on post-approval edit
	sc3 := fixture.Scenarios[2]
	tamperedViolations := CheckBrandCompliance(sc3.TamperedModification.ModifiedText, camp.Governance)
	if len(tamperedViolations) != len(sc3.ExpectedForbiddenKWs) {
		t.Errorf("expected %d forbidden violations, got %d (%v)", len(sc3.ExpectedForbiddenKWs), len(tamperedViolations), tamperedViolations)
	}

	// Scenario 4: Multi-tenant workspace isolation
	sc4 := fixture.Scenarios[3]
	if sc4.PrimaryWorkspace == sc4.IntruderWorkspace {
		t.Errorf("sc4 expected different workspaces")
	}
}

func TestAdvocacy_BrandComplianceAndForbiddenWords(t *testing.T) {
	gov := BrandGovernance{
		Guidelines: "Be honest and authentic.",
		ForbiddenKeywords: []string{
			"guaranteed 100% returns",
			"secret trick",
			"competitor x sucks",
		},
	}

	cleanText := "We just released our latest benchmark results for our streaming pipeline."
	if violations := CheckBrandCompliance(cleanText, gov); len(violations) > 0 {
		t.Errorf("expected 0 violations for clean text, got %v", violations)
	}

	dirtyText := "Our new software is a SECRET TRICK that offers GUARANTEED 100% RETURNS!"
	violations := CheckBrandCompliance(dirtyText, gov)
	if len(violations) != 2 {
		t.Errorf("expected 2 violations for dirty text, got %d (%v)", len(violations), violations)
	}
}

func TestAdvocacy_RejectCoordinatedFakeEngagement(t *testing.T) {
	req := CoordinatedEngagementRequest{
		CampaignID:    "camp-01",
		TargetPostURN: "urn:li:share:123",
		Action:        "auto_like_all_employees",
		ParticipatingEmployeeIDs: []string{
			"emp-1", "emp-2", "emp-3",
		},
	}

	repo := NewMemoryRepository()
	svc := NewService(repo)
	err := svc.TriggerCoordinatedPod(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for coordinated engagement pod, got nil")
	}
	if err != ErrCoordinatedEngagementProhibited {
		t.Errorf("expected ErrCoordinatedEngagementProhibited, got %v", err)
	}
}

func TestAdvocacy_ApprovalAndTamperInvalidation_AT007(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)
	secret := "test-signing-secret"

	camp := &AdvocacyCampaign{
		CampaignID:  "camp-test-01",
		TenantID:    "tenant-1",
		WorkspaceID: "ws-1",
		Title:       "Q3 Engineering Milestone",
		Description: "Celebrating zero-incident migration",
		Governance: BrandGovernance{
			Guidelines: "Professional tone.",
			ForbiddenKeywords: []string{
				"hacked the system",
				"foolproof",
			},
		},
		Variants: []AdvocacyCopyVariant{
			{
				VariantID:     "var-1",
				Persona:       PersonaEngineering,
				Headline:      "Zero-Incident Migration",
				SuggestedText: "Excited to share that our team completed our DB migration without customer interruption.",
			},
		},
	}

	created, err := svc.CreateAdvocacyCampaign(ctx, camp)
	if err != nil {
		t.Fatalf("failed to create campaign: %v", err)
	}
	if created.Status != AdvocacyStatusDraft {
		t.Errorf("expected initial status draft, got %s", created.Status)
	}

	// Approve campaign
	approved, err := svc.ApproveAdvocacyCampaign(ctx, created.CampaignID, "lead-user", "tenant-1", secret)
	if err != nil {
		t.Fatalf("failed to approve campaign: %v", err)
	}
	if approved.Status != AdvocacyStatusApproved {
		t.Errorf("expected status approved, got %s", approved.Status)
	}
	if approved.ApprovalToken == "" {
		t.Errorf("expected non-empty approval token")
	}

	// Verify token
	if !VerifyCampaignApprovalToken(*approved, secret) {
		t.Errorf("expected valid approval token")
	}

	// Tamper/modify variant text
	approved.Variants[0].SuggestedText = "Modified text after approval!"
	updated, err := svc.UpdateAdvocacyCampaign(ctx, approved, "tenant-1")
	if err != nil {
		t.Fatalf("failed to update campaign: %v", err)
	}

	// Invariant AT-007: edits invalidate approval and reset status to draft
	if updated.Status != AdvocacyStatusDraft {
		t.Errorf("expected status to reset to draft upon edit, got %s", updated.Status)
	}
	if updated.ApprovalToken != "" {
		t.Errorf("expected approval token to be cleared upon edit, got %s", updated.ApprovalToken)
	}
	if VerifyCampaignApprovalToken(*updated, secret) {
		t.Errorf("expected VerifyCampaignApprovalToken to return false after edit")
	}
}

func TestAdvocacy_EmployeeShareWorkflow(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)
	secret := "test-signing-secret"

	camp := &AdvocacyCampaign{
		CampaignID:  "camp-share-01",
		TenantID:    "tenant-1",
		WorkspaceID: "ws-1",
		Title:       "Hiring Senior Engineers",
		Description: "Call for senior systems engineers",
		Governance: BrandGovernance{
			Guidelines: "Highlight career growth.",
			ForbiddenKeywords: []string{
				"10x rockstar",
			},
		},
		Variants: []AdvocacyCopyVariant{
			{
				VariantID:     "var-hire",
				Persona:       PersonaTalentCulture,
				Headline:      "We're Hiring!",
				SuggestedText: "We are growing our distributed systems team! Reach out if you're interested in building scalable infra.",
			},
		},
	}

	_, err := svc.CreateAdvocacyCampaign(ctx, camp)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Sharing before approval must fail
	_, err = svc.ShareAdvocacyContent(ctx, "camp-share-01", "var-hire", "emp-101", "", "ws-1", "tenant-1", secret)
	if err == nil || err != ErrCampaignNotApproved {
		t.Fatalf("expected ErrCampaignNotApproved, got %v", err)
	}

	// Approve
	_, err = svc.ApproveAdvocacyCampaign(ctx, "camp-share-01", "lead-recruiter", "tenant-1", secret)
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	// Sharing with customized text containing forbidden keyword must fail
	_, err = svc.ShareAdvocacyContent(ctx, "camp-share-01", "var-hire", "emp-101", "Looking for a 10x rockstar to join us!", "ws-1", "tenant-1", secret)
	if err == nil {
		t.Fatal("expected error for forbidden keyword in customized share, got nil")
	}

	// Valid share
	res, err := svc.ShareAdvocacyContent(ctx, "camp-share-01", "var-hire", "emp-101", "", "ws-1", "tenant-1", secret)
	if err != nil {
		t.Fatalf("valid share failed: %v", err)
	}
	if !res.Success {
		t.Errorf("expected res.Success = true")
	}
	if res.ClipboardText != camp.Variants[0].SuggestedText {
		t.Errorf("expected clipboard text to match variant")
	}
	if res.ComposeDeepLink == "" {
		t.Errorf("expected non-empty compose deep link")
	}

	// Verify share event recorded and ShareCount incremented
	shares, err := svc.ListEmployeeShares(ctx, "camp-share-01", "ws-1", "tenant-1")
	if err != nil {
		t.Fatalf("list shares failed: %v", err)
	}
	if len(shares) != 1 {
		t.Errorf("expected 1 share event, got %d", len(shares))
	}

	fetched, _ := svc.GetAdvocacyCampaign(ctx, "camp-share-01", "tenant-1")
	if fetched.ShareCount != 1 {
		t.Errorf("expected ShareCount = 1, got %d", fetched.ShareCount)
	}
}

func TestAdvocacy_CrossTenantIsolation_AT011(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)
	secret := "test-signing-secret"

	camp := &AdvocacyCampaign{
		CampaignID:  "camp-iso-01",
		TenantID:    "tenant-alpha",
		WorkspaceID: "ws-alpha",
		Title:       "Confidential Internal Milestone",
		Governance: BrandGovernance{
			Guidelines: "Internal only.",
		},
		Variants: []AdvocacyCopyVariant{
			{
				VariantID:     "var-iso-01",
				Persona:       PersonaGeneral,
				Headline:      "Milestone",
				SuggestedText: "Proud of the milestone.",
			},
		},
	}

	_, err := svc.CreateAdvocacyCampaign(ctx, camp)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Intruder tenant-beta attempt to get campaign
	_, err = svc.GetAdvocacyCampaign(ctx, "camp-iso-01", "tenant-beta")
	if err == nil {
		t.Errorf("expected access denied error for foreign tenant, got nil")
	}

	// Intruder attempt to share
	_, err = svc.ShareAdvocacyContent(ctx, "camp-iso-01", "var-iso-01", "emp-intruder", "", "ws-beta", "tenant-beta", secret)
	if err == nil {
		t.Errorf("expected share access denied for foreign tenant, got nil")
	}

	// Intruder listing campaigns
	list, err := svc.ListAdvocacyCampaigns(ctx, "ws-beta", "tenant-beta")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 campaigns for tenant-beta, got %d", len(list))
	}
}
