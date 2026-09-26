package linkedin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Fixture JSON matching forms/linkedin_limits_reauth.json
type limitsReauthScenario struct {
	ScenarioID            string `json:"scenario_id"`
	Description           string `json:"description"`
	TenantID              string `json:"tenant_id"`
	WorkspaceID           string `json:"workspace_id"`
	AccountID             string `json:"account_id"`
	IncomingStatusCode    int    `json:"incoming_status_code"`
	IncomingErrorBody     string `json:"incoming_error_body"`
	ExpectedLimitType     string `json:"expected_limit_type"`
	ExpectedChallengeType string `json:"expected_challenge_type"`
	ExpectedAccountStatus string `json:"expected_account_status"`
	ExpectedCooldownHours int    `json:"expected_cooldown_hours"`
	ExpectedFailClosed    bool   `json:"expected_fail_closed"`
	ExpectedActionBlocked string `json:"expected_action_blocked"`

	// Scenario 3
	WorkspaceA     string `json:"workspace_a"`
	WorkspaceB     string `json:"workspace_b"`
	ActivityEvents []struct {
		WorkspaceID string `json:"workspace_id"`
		Module      string `json:"module"`
		Action      string `json:"action"`
		Amount      int    `json:"amount"`
		Timestamp   string `json:"timestamp"`
	} `json:"activity_events"`
	ExpectedCombinedUsedInvitations int `json:"expected_combined_used_invitations"`
	ExpectedRemainingDailyInvites   int `json:"expected_remaining_daily_invitations"`
	ExpectedCombinedUsedRecruiterDM int `json:"expected_combined_used_recruiter_dm"`
	ExpectedRemainingRecruiterDM    int `json:"expected_remaining_recruiter_dm"`

	// Scenario 4
	InitialState struct {
		Status          string `json:"status"`
		ActiveChallenge string `json:"active_challenge"`
		ChallengeReason string `json:"challenge_reason"`
	} `json:"initial_state"`
	ResolutionRequest struct {
		ResolutionMethod string `json:"resolution_method"`
		VerifiedByUser   bool   `json:"verified_by_user"`
		Notes            string `json:"notes"`
	} `json:"resolution_request"`
	ExpectedResultingStatus  string `json:"expected_resulting_status"`
	ExpectedChallengeCleared bool   `json:"expected_challenge_cleared"`
	ExpectedAuditLogged      bool   `json:"expected_audit_logged"`
}

type limitsReauthFixtureDoc struct {
	Scenarios []limitsReauthScenario `json:"scenarios"`
}

func loadLimitsReauthFixture(t *testing.T) map[string]limitsReauthScenario {
	t.Helper()
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_limits_reauth.json")
	bytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture file: %v", err)
	}
	var doc limitsReauthFixtureDoc
	if err := json.Unmarshal(bytes, &doc); err != nil {
		t.Fatalf("Failed to unmarshal fixture JSON: %v", err)
	}
	scenarioMap := make(map[string]limitsReauthScenario)
	for _, sc := range doc.Scenarios {
		scenarioMap[sc.ScenarioID] = sc
	}
	return scenarioMap
}

func TestLimitsReauth_FixtureGroundTruth(t *testing.T) {
	scenarios := loadLimitsReauthFixture(t)
	ctx := context.Background()
	store := NewMemoryRepository()
	service := NewService(store)

	// 1. Weekly invitation limit detection
	sc1, ok := scenarios["weekly_invitation_limit_detection"]
	if !ok {
		t.Fatalf("Scenario weekly_invitation_limit_detection not found")
	}
	inc1, state1, err := service.DetectAndRecordRestriction(ctx, sc1.AccountID, sc1.TenantID, sc1.WorkspaceID, "connection_request", sc1.IncomingStatusCode, sc1.IncomingErrorBody)
	if err != nil {
		t.Fatalf("DetectAndRecordRestriction failed: %v", err)
	}
	if string(inc1.LimitType) != sc1.ExpectedLimitType {
		t.Errorf("LimitType mismatch: got %v, want %v", inc1.LimitType, sc1.ExpectedLimitType)
	}
	if string(state1.Status) != sc1.ExpectedAccountStatus {
		t.Errorf("SafetyStatus mismatch: got %v, want %v", state1.Status, sc1.ExpectedAccountStatus)
	}
	if inc1.CooldownHours != sc1.ExpectedCooldownHours {
		t.Errorf("CooldownHours mismatch: got %v, want %v", inc1.CooldownHours, sc1.ExpectedCooldownHours)
	}

	gate1, err := service.CheckSafetyGate(ctx, sc1.AccountID, sc1.TenantID, sc1.ExpectedActionBlocked)
	if err != nil {
		t.Fatalf("CheckSafetyGate failed: %v", err)
	}
	if gate1.Allowed == sc1.ExpectedFailClosed {
		t.Errorf("SafetyGate allowed expected fail-closed, got allowed=%v", gate1.Allowed)
	}

	// 2. CAPTCHA checkpoint fail-closed (AT-010, SRC-S3)
	sc2, ok := scenarios["captcha_checkpoint_fail_closed"]
	if !ok {
		t.Fatalf("Scenario captcha_checkpoint_fail_closed not found")
	}
	inc2, state2, err := service.DetectAndRecordRestriction(ctx, sc2.AccountID, sc2.TenantID, sc2.WorkspaceID, "message", sc2.IncomingStatusCode, sc2.IncomingErrorBody)
	if err != nil {
		t.Fatalf("DetectAndRecordRestriction for checkpoint failed: %v", err)
	}
	if string(inc2.LimitType) != sc2.ExpectedLimitType {
		t.Errorf("LimitType mismatch: got %v, want %v", inc2.LimitType, sc2.ExpectedLimitType)
	}
	if string(state2.Status) != sc2.ExpectedAccountStatus {
		t.Errorf("SafetyStatus mismatch: got %v, want %v", state2.Status, sc2.ExpectedAccountStatus)
	}
	if inc2.ChallengeType == nil || string(*inc2.ChallengeType) != sc2.ExpectedChallengeType {
		t.Errorf("ChallengeType mismatch: got %v, want %v", inc2.ChallengeType, sc2.ExpectedChallengeType)
	}

	gate2, err := service.CheckSafetyGate(ctx, sc2.AccountID, sc2.TenantID, "message")
	if err != nil {
		t.Fatalf("CheckSafetyGate failed: %v", err)
	}
	if gate2.Allowed {
		t.Errorf("Security checkpoint must be fail-closed, but dispatch was allowed!")
	}
	if !strings.Contains(gate2.Reason, "legitimate user-facing re-authentication required") {
		t.Errorf("Expected reason to mention legitimate re-authentication required, got %q", gate2.Reason)
	}

	// 3. Cross-workspace budget sharing (AT-008, AT-009)
	sc3, ok := scenarios["cross_workspace_account_budget_sharing_at009"]
	if !ok {
		t.Fatalf("Scenario cross_workspace_account_budget_sharing_at009 not found")
	}
	now := time.Now().UTC()
	for i, evt := range sc3.ActivityEvents {
		_ = service.RecordAccountActivity(ctx, &ActivityLogEntry{
			EntryID:     fmt.Sprintf("act-event-%d", i),
			AccountID:   sc3.AccountID,
			TenantID:    sc3.TenantID,
			WorkspaceID: evt.WorkspaceID,
			Module:      evt.Module,
			Action:      evt.Action,
			Amount:      evt.Amount,
			Status:      "dispatched",
			Timestamp:   now, // Using current time so it falls within rolling window
		})
	}

	budgetReport, err := service.GetCombinedBudgetReport(ctx, sc3.AccountID, sc3.TenantID)
	if err != nil {
		t.Fatalf("GetCombinedBudgetReport failed: %v", err)
	}
	if budgetReport.DailyInvitationsUsed != sc3.ExpectedCombinedUsedInvitations {
		t.Errorf("DailyInvitationsUsed mismatch: got %d, want %d", budgetReport.DailyInvitationsUsed, sc3.ExpectedCombinedUsedInvitations)
	}
	if budgetReport.DailyInvitationsRemaining != sc3.ExpectedRemainingDailyInvites {
		t.Errorf("DailyInvitationsRemaining mismatch: got %d, want %d", budgetReport.DailyInvitationsRemaining, sc3.ExpectedRemainingDailyInvites)
	}
	if budgetReport.DailyRecruiterDMUsed != sc3.ExpectedCombinedUsedRecruiterDM {
		t.Errorf("DailyRecruiterDMUsed mismatch: got %d, want %d", budgetReport.DailyRecruiterDMUsed, sc3.ExpectedCombinedUsedRecruiterDM)
	}
	if budgetReport.DailyRecruiterDMRemaining != sc3.ExpectedRemainingRecruiterDM {
		t.Errorf("DailyRecruiterDMRemaining mismatch: got %d, want %d", budgetReport.DailyRecruiterDMRemaining, sc3.ExpectedRemainingRecruiterDM)
	}

	// 4. Legitimate reauth resolution
	sc4, ok := scenarios["legitimate_reauth_resolution"]
	if !ok {
		t.Fatalf("Scenario legitimate_reauth_resolution not found")
	}
	// Initial state setup
	initState := &AccountSafetyState{
		AccountID:       sc4.AccountID,
		TenantID:        sc4.TenantID,
		WorkspaceID:     sc4.WorkspaceID,
		Status:          SafetyStatusChallengeRequired,
		ChallengeReason: sc4.InitialState.ChallengeReason,
		UpdatedAt:       now,
	}
	chalType := ChallengeCaptcha
	initState.ActiveChallenge = &chalType
	if err := store.SaveAccountSafetyState(ctx, initState); err != nil {
		t.Fatalf("Failed to save initial state: %v", err)
	}

	stateBefore, err := service.GetAccountSafetyState(ctx, sc4.AccountID, sc4.TenantID)
	if err != nil {
		t.Fatalf("GetAccountSafetyState failed: %v", err)
	}
	if string(stateBefore.Status) != sc4.InitialState.Status {
		t.Errorf("Initial status mismatch: got %v, want %v", stateBefore.Status, sc4.InitialState.Status)
	}

	resolvedState, err := service.ResolveChallenge(ctx, sc4.AccountID, sc4.TenantID, ResolveChallengeRequest{
		ResolutionMethod: sc4.ResolutionRequest.ResolutionMethod,
		VerifiedByUser:   sc4.ResolutionRequest.VerifiedByUser,
		Notes:            sc4.ResolutionRequest.Notes,
	})
	if err != nil {
		t.Fatalf("ResolveChallenge failed: %v", err)
	}
	if string(resolvedState.Status) != sc4.ExpectedResultingStatus {
		t.Errorf("Resulting status mismatch: got %v, want %v", resolvedState.Status, sc4.ExpectedResultingStatus)
	}
	if sc4.ExpectedChallengeCleared && resolvedState.ActiveChallenge != nil {
		t.Errorf("Expected active challenge to be cleared, got: %v", resolvedState.ActiveChallenge)
	}

	gate4, err := service.CheckSafetyGate(ctx, sc4.AccountID, sc4.TenantID, "profile_view")
	if err != nil {
		t.Fatalf("CheckSafetyGate failed: %v", err)
	}
	if !gate4.Allowed {
		t.Errorf("Post-resolution safety gate should allow operations, got false: %s", gate4.Reason)
	}

	if sc4.ExpectedAuditLogged {
		activities, err := service.ListAccountActivity(ctx, sc4.AccountID, sc4.TenantID, 10)
		if err != nil || len(activities) == 0 {
			t.Errorf("Expected resolution to be audit logged in activity ledger")
		}
	}
}

func TestLimitsReauth_ErrorInspectionAndDetection(t *testing.T) {
	testCases := []struct {
		name                 string
		rawError             string
		statusCode           int
		expectedLimitType    PlatformLimitType
		expectedChallenge    ChallengeType
		expectedSafetyStatus AccountSafetyStatus
		minCooldownHours     int
	}{
		{
			name:                 "Weekly Invite Limit reached",
			rawError:             "You've reached the weekly invitation limit. Connections help you stay in touch.",
			statusCode:           429,
			expectedLimitType:    LimitWeeklyInvitations,
			expectedChallenge:    ChallengeWeeklyInvitationLimit,
			expectedSafetyStatus: SafetyStatusRateLimited,
			minCooldownHours:     72,
		},
		{
			name:                 "Commercial search limit reached",
			rawError:             "You've reached the commercial use limit on search.",
			statusCode:           429,
			expectedLimitType:    LimitProfileSearches,
			expectedChallenge:    ChallengeCommercialUseCap,
			expectedSafetyStatus: SafetyStatusRateLimited,
			minCooldownHours:     24,
		},
		{
			name:                 "Standard HTTP 429 Too Many Requests",
			rawError:             "HTTP 429 Too Many Requests: Rate limit exceeded. Try again in 60 minutes.",
			statusCode:           429,
			expectedLimitType:    LimitRateLimit429,
			expectedSafetyStatus: SafetyStatusRateLimited,
			minCooldownHours:     1,
		},
		{
			name:                 "Security Checkpoint 403",
			rawError:             "Security checkpoint required. Please complete verification on your device.",
			statusCode:           403,
			expectedLimitType:    LimitSecurityCheckpoint,
			expectedChallenge:    ChallengeCaptcha,
			expectedSafetyStatus: SafetyStatusChallengeRequired,
			minCooldownHours:     0,
		},
		{
			name:                 "CAPTCHA challenge",
			rawError:             "Automated access prohibited. Challenge/captcha required to continue.",
			statusCode:           403,
			expectedLimitType:    LimitSecurityCheckpoint,
			expectedChallenge:    ChallengeCaptcha,
			expectedSafetyStatus: SafetyStatusChallengeRequired,
			minCooldownHours:     0,
		},
		{
			name:                 "Session Expired Re-auth needed",
			rawError:             "Session expired. Reauthentication required.",
			statusCode:           401,
			expectedLimitType:    LimitSessionExpired,
			expectedChallenge:    ChallengeCredentialReauth,
			expectedSafetyStatus: SafetyStatusChallengeRequired,
			minCooldownHours:     0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now().UTC()
			ins := InspectPlatformErrorForRestrictions("li:acc1", "tenant-1", "ws-1", "action", tc.statusCode, tc.rawError, now)
			if ins == nil {
				t.Fatalf("InspectPlatformErrorForRestrictions returned nil")
			}
			if ins.LimitType != tc.expectedLimitType {
				t.Errorf("LimitType: got %v, want %v", ins.LimitType, tc.expectedLimitType)
			}
			if tc.expectedChallenge != "" {
				if ins.ChallengeType == nil || *ins.ChallengeType != tc.expectedChallenge {
					t.Errorf("ChallengeType: got %v, want %v", ins.ChallengeType, tc.expectedChallenge)
				}
			}
			state := &AccountSafetyState{
				AccountID: "li:acc1",
				TenantID:  "tenant-1",
				Status:    SafetyStatusActive,
			}
			ApplyRestrictionToAccount(state, ins, now)
			if state.Status != tc.expectedSafetyStatus {
				t.Errorf("SafetyStatus: got %v, want %v", state.Status, tc.expectedSafetyStatus)
			}
			if tc.minCooldownHours > 0 && ins.CooldownHours < tc.minCooldownHours {
				t.Errorf("CooldownHours too short: got %v, want >= %v hours", ins.CooldownHours, tc.minCooldownHours)
			}
		})
	}
}

func TestLimitsReauth_FailClosedSafetyGate_AT010(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryRepository()
	service := NewService(store)
	tenantID := "tenant-checkpoint"
	providerAcc := "li:member_checkpoint_test"

	// 1. Initial state should be active and allow dispatch
	gate, err := service.CheckSafetyGate(ctx, providerAcc, tenantID, "connection_request")
	if err != nil {
		t.Fatalf("CheckSafetyGate initial failed: %v", err)
	}
	if !gate.Allowed {
		t.Errorf("Initial state should allow dispatch, got false")
	}

	// 2. Platform returns CAPTCHA error
	_, _, err = service.DetectAndRecordRestriction(ctx, providerAcc, tenantID, "ws-1", "message", 403, "A checkpoint check was encountered")
	if err != nil {
		t.Fatalf("DetectAndRecordRestriction failed: %v", err)
	}

	// 3. Fail-closed safety gate should strictly block ALL dispatch operations
	actionsToTest := []string{"connection_request", "message", "profile_view", "post_publish", "withdraw"}
	for _, action := range actionsToTest {
		g, err := service.CheckSafetyGate(ctx, providerAcc, tenantID, action)
		if err != nil {
			t.Fatalf("CheckSafetyGate failed for %s: %v", action, err)
		}
		if g.Allowed {
			t.Fatalf("Safety gate MUST FAIL CLOSED during CAPTCHA challenge for action %s, but got allowed=true!", action)
		}
		if g.Status != SafetyStatusChallengeRequired {
			t.Errorf("Expected status challenge_required, got %v", g.Status)
		}
	}

	// 4. Attempting to resolve without user confirmation must fail
	_, err = service.ResolveChallenge(ctx, providerAcc, tenantID, ResolveChallengeRequest{
		ResolutionMethod: "oauth_reauth",
		VerifiedByUser:   false, // Unconfirmed
		Notes:            "Bypass attempt",
	})
	if err == nil {
		t.Errorf("Resolving challenge without user confirmation should fail")
	}

	// 5. Resolving with legitimate confirmation succeeds and reopens gate
	_, err = service.ResolveChallenge(ctx, providerAcc, tenantID, ResolveChallengeRequest{
		ResolutionMethod: "oauth_reauth",
		VerifiedByUser:   true,
		Notes:            "Legitimate OAuth token refreshed by user",
	})
	if err != nil {
		t.Fatalf("Legitimate resolve failed: %v", err)
	}

	gateAfter, err := service.CheckSafetyGate(ctx, providerAcc, tenantID, "connection_request")
	if err != nil {
		t.Fatalf("CheckSafetyGate after resolve failed: %v", err)
	}
	if !gateAfter.Allowed {
		t.Errorf("Post-resolve gate should be allowed, got false: %s", gateAfter.Reason)
	}
}

func TestLimitsReauth_CombinedBudgetPooling_AT008_AT009(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryRepository()
	service := NewService(store)
	tenantID := "tenant-pool"
	accountID := "li:member_pool_audit_99"
	now := time.Now().UTC()

	// 10 invites in ws-alpha
	for i := 0; i < 10; i++ {
		_ = service.RecordAccountActivity(ctx, &ActivityLogEntry{
			EntryID:     fmt.Sprintf("act-alpha-inv-%d", i),
			AccountID:   accountID,
			TenantID:    tenantID,
			WorkspaceID: "ws-alpha",
			Module:      "career",
			Action:      "connection_request",
			Amount:      1,
			Status:      "dispatched",
			Timestamp:   now,
		})
	}

	// 15 invites in ws-beta
	for i := 0; i < 15; i++ {
		_ = service.RecordAccountActivity(ctx, &ActivityLogEntry{
			EntryID:     fmt.Sprintf("act-beta-inv-%d", i),
			AccountID:   accountID,
			TenantID:    tenantID,
			WorkspaceID: "ws-beta",
			Module:      "linkedin",
			Action:      "connection_request",
			Amount:      1,
			Status:      "dispatched",
			Timestamp:   now,
		})
	}

	// 5 recruiter DMs in ws-alpha
	for i := 0; i < 5; i++ {
		_ = service.RecordAccountActivity(ctx, &ActivityLogEntry{
			EntryID:     fmt.Sprintf("act-alpha-rec-%d", i),
			AccountID:   accountID,
			TenantID:    tenantID,
			WorkspaceID: "ws-alpha",
			Module:      "career",
			Action:      "recruiter_dm",
			Amount:      1,
			Status:      "dispatched",
			Timestamp:   now,
		})
	}

	// 3 inmail in ws-beta
	for i := 0; i < 3; i++ {
		_ = service.RecordAccountActivity(ctx, &ActivityLogEntry{
			EntryID:     fmt.Sprintf("act-beta-inmail-%d", i),
			AccountID:   accountID,
			TenantID:    tenantID,
			WorkspaceID: "ws-beta",
			Module:      "linkedin",
			Action:      "inmail",
			Amount:      1,
			Status:      "dispatched",
			Timestamp:   now,
		})
	}

	report, err := service.GetCombinedBudgetReport(ctx, accountID, tenantID)
	if err != nil {
		t.Fatalf("GetCombinedBudgetReport failed: %v", err)
	}

	// Total weekly invites: 10 + 15 = 25
	if report.WeeklyInvitationsUsed != 25 {
		t.Errorf("WeeklyInvitationsUsed: got %d, want 25", report.WeeklyInvitationsUsed)
	}
	if report.WeeklyInvitationsRemaining != 55 { // Default cap is 80 -> 80 - 25 = 55
		t.Errorf("WeeklyInvitationsRemaining: got %d, want 55", report.WeeklyInvitationsRemaining)
	}

	// Total Daily Recruiter DM: 5
	if report.DailyRecruiterDMUsed != 5 {
		t.Errorf("DailyRecruiterDMUsed: got %d, want 5", report.DailyRecruiterDMUsed)
	}

	// Total Daily InMail: 5 (recruiter DMs count as inmail) + 3 = 8
	if report.DailyInMailUsed != 8 {
		t.Errorf("DailyInMailUsed: got %d, want 8", report.DailyInMailUsed)
	}
	if report.DailyInMailRemaining != 7 { // Default cap is 15 -> 15 - 8 = 7
		t.Errorf("DailyInMailRemaining: got %d, want 7", report.DailyInMailRemaining)
	}

	// Check activity list
	logs, err := service.ListAccountActivity(ctx, accountID, tenantID, 100)
	if err != nil {
		t.Fatalf("ListAccountActivity failed: %v", err)
	}
	if len(logs) != 33 { // 10 + 15 + 5 + 3 = 33
		t.Errorf("Total activity logs: got %d, want 33", len(logs))
	}
}
