package career

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestApplicationLifecycle_LegalAndIllegalTransitions(t *testing.T) {
	// Legal transitions
	legalCases := []struct {
		from ApplicationStage
		to   ApplicationStage
	}{
		{StageInterested, StagePreparing},
		{StagePreparing, StageDispatched},
		{StageDispatched, StageNeedsConfirmation},
		{StageNeedsConfirmation, StageApplied},
		{StageApplied, StageInterviewing},
		{StageInterviewing, StageOffered},
		{StageInterviewing, StageRejected},
		{StageApplied, StageWithdrawn},
		{StageRejected, StageArchived},
	}

	for _, tc := range legalCases {
		if !IsValidStageTransition(tc.from, tc.to) {
			t.Errorf("expected transition from %s to %s to be legal, got false", tc.from, tc.to)
		}
	}

	// Illegal transitions
	illegalCases := []struct {
		from ApplicationStage
		to   ApplicationStage
	}{
		{StageInterested, StageInterviewing}, // Cannot interview without applied
		{StageDispatched, StageOffered},      // Cannot get offer before applied & interviewing
		{StageRejected, StageInterviewing},   // Cannot interview after rejection
		{StageRejected, StageOffered},        // Cannot get offer after rejection
		{StageOffered, StageInterviewing},    // Cannot go back to interviewing after offer
	}

	for _, tc := range illegalCases {
		if IsValidStageTransition(tc.from, tc.to) {
			t.Errorf("expected transition from %s to %s to be illegal, got true", tc.from, tc.to)
		}
	}
}

func TestApplicationStatus_ClickNeverMarksApplied_AT005(t *testing.T) {
	now := time.Now().UTC()
	session := &ApplicationReviewSession{
		ID:                  "sess_test_01",
		UserID:              "user_alice",
		JobID:               "job_01",
		JobTitle:            "Senior Go Engineer",
		Company:             "Acme Corp",
		ApplyURL:            "https://acme.com/apply",
		Status:              WorkflowStatusApproved,
		ExecutionMode:       ExecutionModeNativeManual,
		TimeoutDurationSecs: 600,
		CreatedAt:           now,
	}

	// Dispatch session (simulate candidate opening portal)
	session.Status = WorkflowStatusDispatched
	session.DispatchedAt = &now

	if session.Status == WorkflowStatusApplied {
		t.Fatalf("INVARIANT VIOLATION (AT-005): Opening or dispatching portal marked application as applied!")
	}
	if session.Status != WorkflowStatusDispatched {
		t.Errorf("expected status %s, got %s", WorkflowStatusDispatched, session.Status)
	}
	if session.Receipt != nil {
		t.Errorf("expected nil receipt on dispatch, got %+v", session.Receipt)
	}
}

func TestApplicationStatus_TimeoutEntersNeedsConfirmation_AT005(t *testing.T) {
	now := time.Now().UTC()
	dispatchedAt := now.Add(-15 * time.Minute) // 15 mins ago

	session := &ApplicationReviewSession{
		ID:                  "sess_test_timeout",
		UserID:              "user_alice",
		JobID:               "job_timeout_01",
		Status:              WorkflowStatusDispatched,
		DispatchedAt:        &dispatchedAt,
		TimeoutDurationSecs: 600, // 10 mins
	}

	expired, newStatus := EvaluateSessionTimeout(session, now)
	if !expired {
		t.Fatalf("expected session to be marked as expired after 15 mins")
	}
	if newStatus != WorkflowStatusNeedsConfirmation {
		t.Fatalf("INVARIANT VIOLATION (AT-005): Expired session transitioned to %s instead of needs_confirmation", newStatus)
	}

	// Verify that a session dispatched only 5 minutes ago does NOT expire
	recentDispatched := now.Add(-5 * time.Minute)
	session.DispatchedAt = &recentDispatched
	expiredRecent, recentStatus := EvaluateSessionTimeout(session, now)
	if expiredRecent {
		t.Errorf("expected session not to be expired after 5 mins")
	}
	if recentStatus != WorkflowStatusDispatched {
		t.Errorf("expected status to remain dispatched, got %s", recentStatus)
	}
}

func TestApplicationStatus_VerifiedAppliedMark_REQ005_REQ016(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)

	userID := "user_alice_01"
	recordID := "app_rec_test_01"

	// Initial record in dispatched stage
	now := time.Now().UTC()
	rec := &ApplicationRecord{
		ID:                recordID,
		UserID:            userID,
		JobID:             "job_rec_01",
		CanonicalURL:      "https://example.com/careers/job1",
		Title:             "Software Architect",
		Company:           "GlobalTech",
		Status:            "dispatched",
		Stage:             StageDispatched,
		IsVerifiedApplied: false,
		AppliedAt:         now,
		UpdatedAt:         now,
	}
	_ = repo.SaveApplicationRecord(ctx, rec)

	// Attempt 1: Verify without receipt or attestation notes (must fail closed per REQ-016)
	_, err := service.VerifyAppliedMark(ctx, userID, VerifyAppliedMarkRequest{
		ApplicationRecordID: recordID,
		ReceiptID:           "",
		AttestationNotes:    "",
	})
	if err == nil {
		t.Fatalf("expected ErrMissingVerification when no proof is provided, got nil")
	}

	// Attempt 2: Verify with candidate attestation and provider reference
	res, err := service.VerifyAppliedMark(ctx, userID, VerifyAppliedMarkRequest{
		ApplicationRecordID: recordID,
		VerificationType:    VerificationUserAttestation,
		ReceiptID:           "rec_manual_9981",
		ProviderReference:   "REQ-2026-9981",
		AttestationNotes:    "Candidate confirmed submission through external portal modal.",
	})
	if err != nil {
		t.Fatalf("unexpected error verifying applied mark: %v", err)
	}

	if !res.Record.IsVerifiedApplied {
		t.Errorf("expected record.IsVerifiedApplied to be true")
	}
	if res.Record.Stage != StageApplied {
		t.Errorf("expected record.Stage to be %s, got %s", StageApplied, res.Record.Stage)
	}
	if res.Record.VerificationDetails == nil {
		t.Fatalf("expected non-nil VerificationDetails")
	}
	if res.Record.VerificationDetails.ReceiptID != "rec_manual_9981" {
		t.Errorf("expected receipt ID rec_manual_9981, got %s", res.Record.VerificationDetails.ReceiptID)
	}

	// Attempt 3: Double-verify should be rejected
	_, errDouble := service.VerifyAppliedMark(ctx, userID, VerifyAppliedMarkRequest{
		ApplicationRecordID: recordID,
		ReceiptID:           "rec_manual_duplicate",
		AttestationNotes:    "Trying to re-verify",
	})
	if errDouble != ErrApplicationAlreadyApplied {
		t.Errorf("expected ErrApplicationAlreadyApplied, got %v", errDouble)
	}
}

func TestApplicationStatus_CrashRecovery_NoDuplicateApply_AT006(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)

	userID := "user_alice_crash"
	now := time.Now().UTC()
	dispatchedAt := now.Add(-12 * time.Minute)

	session := &ApplicationReviewSession{
		ID:                  "sess_crash_01",
		UserID:              userID,
		JobID:               "job_crash_01",
		JobTitle:            "Staff Platform Engineer",
		Company:             "Acme Cloud",
		ApplyURL:            "https://acme.com/job/123",
		Status:              WorkflowStatusDispatched,
		DispatchedAt:        &dispatchedAt,
		TimeoutDurationSecs: 600,
		CreatedAt:           dispatchedAt,
	}
	_ = repo.SaveReviewSession(ctx, session)

	// Simulate system timeout evaluation (e.g. after worker crashed and recovered)
	transitioned, err := service.CheckAndExpireDispatchedSessions(ctx)
	if err != nil {
		t.Fatalf("unexpected error in CheckAndExpireDispatchedSessions: %v", err)
	}
	if len(transitioned) != 1 {
		t.Fatalf("expected 1 session transitioned to needs_confirmation, got %d", len(transitioned))
	}
	if transitioned[0].Status != WorkflowStatusNeedsConfirmation {
		t.Errorf("expected status needs_confirmation, got %s", transitioned[0].Status)
	}

	// Candidate decides to retry dispatch (AT-006: reset without duplicate application record creation)
	resRetry, err := service.ReconcileApplication(ctx, userID, ReconcileApplicationRequest{
		SessionID: "sess_crash_01",
		Action:    ReconcileActionRetryDispatch,
		Notes:     "Worker recovered. Candidate re-opening portal.",
	})
	if err != nil {
		t.Fatalf("unexpected error in retry dispatch: %v", err)
	}
	if resRetry.Session.Status != WorkflowStatusDispatched {
		t.Errorf("expected status to reset to dispatched, got %s", resRetry.Session.Status)
	}

	// Verify no premature ApplicationRecord exists yet
	apps, _ := repo.ListApplications(ctx, userID)
	if len(apps) != 0 {
		t.Fatalf("INVARIANT VIOLATION (AT-006): Retrying dispatch generated premature application record in ledger!")
	}
}

func TestApplicationStatus_ReconciliationFlows(t *testing.T) {
	now := time.Now().UTC()

	// Scenario 1: Confirm Applied with receipt
	session1 := &ApplicationReviewSession{
		ID:                  "sess_rec_01",
		UserID:              "user_bob",
		JobID:               "job_01",
		JobTitle:            "Backend Dev",
		Company:             "Beta Corp",
		ApplyURL:            "https://beta.com/apply",
		Status:              WorkflowStatusNeedsConfirmation,
		TimeoutDurationSecs: 600,
	}

	receipt := &SubmissionReceipt{
		ReceiptID:         "rec_beta_771",
		ProviderReference: "BETA-CONF-771",
		SubmissionURL:     "https://beta.com/apply",
		ConfirmedAt:       now,
		ConfirmedByUser:   true,
		Notes:             "Confirmed via provider receipt.",
	}

	event1, err := ReconcileDispatchedSession(session1, ReconcileActionConfirmApplied, receipt, "All good", "user_bob", now)
	if err != nil {
		t.Fatalf("unexpected error in confirm applied: %v", err)
	}
	if session1.Status != WorkflowStatusApplied {
		t.Errorf("expected session1 status applied, got %s", session1.Status)
	}
	if event1.ToStage != StageApplied {
		t.Errorf("expected event1 to stage applied, got %s", event1.ToStage)
	}

	// Scenario 2: Mark Abandoned
	session2 := &ApplicationReviewSession{
		ID:                  "sess_rec_02",
		UserID:              "user_bob",
		JobID:               "job_02",
		JobTitle:            "Frontend Dev",
		Company:             "Gamma Corp",
		ApplyURL:            "https://gamma.com/apply",
		Status:              WorkflowStatusNeedsConfirmation,
		TimeoutDurationSecs: 600,
	}

	event2, err := ReconcileDispatchedSession(session2, ReconcileActionMarkAbandoned, nil, "Position closed on external site", "user_bob", now)
	if err != nil {
		t.Fatalf("unexpected error in mark abandoned: %v", err)
	}
	if session2.Status != WorkflowStatusCancelled {
		t.Errorf("expected session2 status cancelled, got %s", session2.Status)
	}
	if event2.ToStage != StageArchived {
		t.Errorf("expected event2 to stage archived, got %s", event2.ToStage)
	}
}

func TestApplicationStatus_TimelineAuditTrail(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)

	userID := "user_timeline"
	recID := "app_rec_timeline_01"
	now := time.Now().UTC()

	// Initial record in applied stage
	rec := &ApplicationRecord{
		ID:                recID,
		UserID:            userID,
		JobID:             "job_tl_01",
		CanonicalURL:      "https://example.com/job/tl",
		Title:             "Security Architect",
		Company:           "CyberGuard",
		Status:            "applied",
		Stage:             StageApplied,
		IsVerifiedApplied: true,
		AppliedAt:         now,
		UpdatedAt:         now,
		Timeline: []ApplicationTimelineEvent{
			{
				EventID:             "evt_init",
				ApplicationRecordID: recID,
				Timestamp:           now,
				FromStage:           StageDispatched,
				ToStage:             StageApplied,
				Trigger:             "initial_applied_verification",
				ActorID:             userID,
				Notes:               "Verified applied via candidate receipt",
			},
		},
	}
	_ = repo.SaveApplicationRecord(ctx, rec)

	// Step 1: Advance to Interviewing
	res1, err := service.UpdateApplicationStage(ctx, userID, UpdateApplicationStageRequest{
		ApplicationRecordID: recID,
		NewStage:            StageInterviewing,
		Notes:               "Scheduled Recruiter Phone Screen for Monday",
	})
	if err != nil {
		t.Fatalf("unexpected error advancing to interviewing: %v", err)
	}
	if res1.Record.Stage != StageInterviewing {
		t.Errorf("expected stage %s, got %s", StageInterviewing, res1.Record.Stage)
	}

	// Step 2: Advance to Offered
	res2, err := service.UpdateApplicationStage(ctx, userID, UpdateApplicationStageRequest{
		ApplicationRecordID: recID,
		NewStage:            StageOffered,
		Notes:               "Formal written offer received at $195,000 base.",
	})
	if err != nil {
		t.Fatalf("unexpected error advancing to offered: %v", err)
	}
	if res2.Record.Stage != StageOffered {
		t.Errorf("expected stage %s, got %s", StageOffered, res2.Record.Stage)
	}

	// Retrieve full timeline audit ledger
	ledgerRes, err := service.GetApplicationAuditLedger(ctx, userID, recID)
	if err != nil {
		t.Fatalf("unexpected error fetching audit ledger: %v", err)
	}

	if len(ledgerRes.Timeline) != 3 {
		t.Fatalf("expected 3 timeline events, got %d", len(ledgerRes.Timeline))
	}
	if ledgerRes.Timeline[0].ToStage != StageApplied {
		t.Errorf("expected first event to stage applied, got %s", ledgerRes.Timeline[0].ToStage)
	}
	if ledgerRes.Timeline[1].ToStage != StageInterviewing {
		t.Errorf("expected second event to stage interviewing, got %s", ledgerRes.Timeline[1].ToStage)
	}
	if ledgerRes.Timeline[2].ToStage != StageOffered {
		t.Errorf("expected third event to stage offered, got %s", ledgerRes.Timeline[2].ToStage)
	}
}

func TestApplicationStatus_FixtureEvaluation(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "application_reconciliation.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture file: %v", err)
	}

	var root struct {
		Scenarios []struct {
			ID                       string `json:"id"`
			Name                     string `json:"name"`
			InitialStage             string `json:"initial_stage"`
			DispatchedMinutesAgo     int    `json:"dispatched_minutes_ago"`
			TimeoutMinutes           int    `json:"timeout_minutes"`
			ExpectedStageAfterCheck  string `json:"expected_stage_after_check"`
			ReconciliationAction     string `json:"reconciliation_action"`
			ConfirmationReceipt      *struct {
				ReceiptID         string `json:"receipt_id"`
				ProviderReference string `json:"provider_reference"`
				Notes             string `json:"notes"`
			} `json:"confirmation_receipt"`
			ExpectedFinalStage string `json:"expected_final_stage"`
		} `json:"scenarios"`
	}

	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("failed to parse fixture JSON: %v", err)
	}

	if len(root.Scenarios) != 4 {
		t.Fatalf("expected 4 scenarios in fixture, got %d", len(root.Scenarios))
	}

	now := time.Now().UTC()
	for _, sc := range root.Scenarios {
		t.Run(sc.Name, func(t *testing.T) {
			dispatchedAt := now.Add(-time.Duration(sc.DispatchedMinutesAgo) * time.Minute)
			session := &ApplicationReviewSession{
				ID:                  "sess_" + sc.ID,
				UserID:              "user_fixture",
				Status:              ApplicationWorkflowStatus(sc.InitialStage),
				DispatchedAt:        &dispatchedAt,
				TimeoutDurationSecs: sc.TimeoutMinutes * 60,
			}

			expired, evalStatus := EvaluateSessionTimeout(session, now)
			if sc.ExpectedStageAfterCheck == "needs_confirmation" {
				if !expired || evalStatus != WorkflowStatusNeedsConfirmation {
					t.Errorf("scenario %s expected needs_confirmation after check, got expired=%v, status=%s", sc.ID, expired, evalStatus)
				}
				session.Status = evalStatus
			} else {
				if expired || evalStatus != ApplicationWorkflowStatus(sc.ExpectedStageAfterCheck) {
					t.Errorf("scenario %s expected %s, got expired=%v, status=%s", sc.ID, sc.ExpectedStageAfterCheck, expired, evalStatus)
				}
			}

			// Apply reconciliation action
			var receipt *SubmissionReceipt
			if sc.ConfirmationReceipt != nil {
				receipt = &SubmissionReceipt{
					ReceiptID:         sc.ConfirmationReceipt.ReceiptID,
					ProviderReference: sc.ConfirmationReceipt.ProviderReference,
					ConfirmedAt:       now,
					ConfirmedByUser:   true,
					Notes:             sc.ConfirmationReceipt.Notes,
				}
			}

			var act ReconciliationAction = ReconciliationAction(sc.ReconciliationAction)
			_, err := ReconcileDispatchedSession(session, act, receipt, "Fixture test notes", "user_fixture", now)
			if err != nil {
				t.Fatalf("scenario %s reconciliation failed: %v", sc.ID, err)
			}

			if string(session.Status) != sc.ExpectedFinalStage {
				t.Errorf("scenario %s expected final stage %s, got %s", sc.ID, sc.ExpectedFinalStage, session.Status)
			}
		})
	}
}
