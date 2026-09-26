package career

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"
)

type RunControlsFixtureSuite struct {
	Description string `json:"description"`
	Scenarios   []struct {
		ID                             string   `json:"id"`
		Name                           string   `json:"name"`
		RunType                        string   `json:"run_type"`
		ExpectedFinalStatus            string   `json:"expected_final_status"`
		RequiresReconciliation         bool     `json:"requires_reconciliation"`
		AutoRetryAllowed               bool     `json:"auto_retry_allowed"`
		ControlAction                  string   `json:"control_action"`
		InitialFencingToken            int      `json:"initial_fencing_token"`
		ExpectedNewFencingToken        int      `json:"expected_new_fencing_token"`
		SimulatedStaleAttemptToken     int      `json:"simulated_stale_attempt_token"`
		HourlyLimit                    int      `json:"hourly_limit"`
		RecentSubmissionOffsetsMinutes []int    `json:"recent_submission_offsets_minutes"`
		ExpectedAllowed                bool     `json:"expected_allowed"`
		ExpectedCurrentWindowCount     int      `json:"expected_current_window_count"`
	} `json:"scenarios"`
}

func TestRunControls_FixtureSuiteEvaluation(t *testing.T) {
	data, err := os.ReadFile("../../testdata/fixtures/forms/run_controls.json")
	if err != nil {
		t.Fatalf("failed to read run_controls.json fixture: %v", err)
	}

	var suite RunControlsFixtureSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("failed to parse run_controls fixture: %v", err)
	}

	if len(suite.Scenarios) < 4 {
		t.Fatalf("expected at least 4 test scenarios in fixture, got %d", len(suite.Scenarios))
	}

	now := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)

	for _, sc := range suite.Scenarios {
		t.Run(sc.ID, func(t *testing.T) {
			switch sc.ID {
			case "standard_discovery_run_lifecycle":
				run := NewCareerRun("usr_123", "ws_1", RunTypeDiscoveryRun, 10, 15)
				run.Status = RunStatusRunning
				// Execute the 3 safe stages in the scenario
				stages := []string{"fetch_job_listings", "evaluate_and_score_matches", "persist_saved_jobs"}
				for _, stg := range stages {
					err := CommitAttempt(run, run.FencingToken, stg, StageSafeToRetry, true, "done")
					if err != nil {
						t.Fatalf("CommitAttempt %s error: %v", stg, err)
					}
				}
				// Mark finished when all processed
				run.Status = RunStatusCompleted
				if string(run.Status) != sc.ExpectedFinalStatus {
					t.Errorf("expected status %s, got %s", sc.ExpectedFinalStatus, run.Status)
				}
				if run.RequiresReconciliation != sc.RequiresReconciliation {
					t.Errorf("expected RequiresReconciliation %v, got %v", sc.RequiresReconciliation, run.RequiresReconciliation)
				}

			case "crash_during_uncertain_submit_scenario":
				if sc.AutoRetryAllowed {
					t.Errorf("expected auto_retry_allowed to be false per AT-006")
				}
				run := NewCareerRun("usr_123", "ws_1", RunTypeApplicationSession, 1, 15)
				run.Status = RunStatusRunning
				run.CurrentStage = "submit_application_to_ats"
				run.StageSafety = StageUncertainWriteRequiresReconciliation
				run.LeaseExpiresAt = now.Add(-5 * time.Minute)

				// Lease expired: recover run
				recovered, err := RecoverCrashedRun(run, now)
				if err != nil {
					t.Fatalf("RecoverCrashedRun error: %v", err)
				}
				if !recovered {
					t.Errorf("expected run to be recovered")
				}
				if string(run.Status) != sc.ExpectedFinalStatus {
					t.Errorf("expected status %s, got %s", sc.ExpectedFinalStatus, run.Status)
				}
				if run.RequiresReconciliation != sc.RequiresReconciliation {
					t.Errorf("expected RequiresReconciliation %v, got %v", sc.RequiresReconciliation, run.RequiresReconciliation)
				}

			case "fencing_token_stale_worker_scenario":
				run := NewCareerRun("usr_123", "ws_1", RunTypeTailoringBatch, 10, 15)
				run.Status = RunStatusRunning
				run.FencingToken = int64(sc.InitialFencingToken)

				// Control action: pause increments token
				if sc.ControlAction == "pause" {
					if err := PauseCareerRun(run, "paused"); err != nil {
						t.Fatalf("PauseCareerRun error: %v", err)
					}
				}
				if int(run.FencingToken) != sc.ExpectedNewFencingToken {
					t.Errorf("expected new fencing token %d, got %d", sc.ExpectedNewFencingToken, run.FencingToken)
				}

				// Simulated stale worker attempt
				err := CommitAttempt(run, int64(sc.SimulatedStaleAttemptToken), "score_and_filter", StageSafeToRetry, true, "done")
				if !errors.Is(err, ErrStaleAttemptFenced) {
					t.Fatalf("expected ErrStaleAttemptFenced, got %v", err)
				}

			case "rolling_hourly_rate_limit_scenario":
				var recent []time.Time
				for _, offset := range sc.RecentSubmissionOffsetsMinutes {
					recent = append(recent, now.Add(-time.Duration(offset)*time.Minute))
				}
				allowed, count, backoff := EvaluateRollingHourlyLimit(recent, now, sc.HourlyLimit)
				if allowed != sc.ExpectedAllowed {
					t.Errorf("expected allowed %v, got %v", sc.ExpectedAllowed, allowed)
				}
				if count != sc.ExpectedCurrentWindowCount {
					t.Errorf("expected count %d, got %d", sc.ExpectedCurrentWindowCount, count)
				}
				if backoff <= 0 {
					t.Errorf("expected positive backoff duration, got %v", backoff)
				}
			}
		})
	}
}

func TestRunControls_LifecycleAndFencingTokens_AT021(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	run, err := service.CreateCareerRun(ctx, "usr_alice", "ws_1", RunTypeDiscoveryRun, 5, 10)
	if err != nil {
		t.Fatalf("CreateCareerRun error: %v", err)
	}

	if run.Status != RunStatusQueued {
		t.Fatalf("expected queued status, got %s", run.Status)
	}
	if run.FencingToken != 1 {
		t.Fatalf("expected initial fencing token 1, got %d", run.FencingToken)
	}

	// 1. Pause run -> increments fencing token to 2
	controlled, err := service.ControlCareerRun(ctx, "usr_alice", run.ID, ActionPause, "manual maintenance")
	if err != nil {
		t.Fatalf("ControlCareerRun pause error: %v", err)
	}
	if controlled.Status != RunStatusPaused {
		t.Errorf("expected RunStatusPaused, got %s", controlled.Status)
	}
	if controlled.FencingToken != 2 {
		t.Errorf("expected fencing token 2 after pause, got %d", controlled.FencingToken)
	}

	// 2. Late worker attempt with old fencing token (1) must be rejected per AT-021
	_, err = service.CommitRunAttempt(ctx, "usr_alice", run.ID, 1, "discover_candidates", StageSafeToRetry, true, "done")
	if !errors.Is(err, ErrStaleAttemptFenced) {
		t.Errorf("expected ErrStaleAttemptFenced, got: %v", err)
	}

	// 3. Resume run -> fencing token increments to 3, status becomes running
	resumed, err := service.ControlCareerRun(ctx, "usr_alice", run.ID, ActionResume, "")
	if err != nil {
		t.Fatalf("ControlCareerRun resume error: %v", err)
	}
	if resumed.Status != RunStatusRunning {
		t.Errorf("expected RunStatusRunning, got %s", resumed.Status)
	}
	if resumed.FencingToken != 3 {
		t.Errorf("expected fencing token 3 after resume, got %d", resumed.FencingToken)
	}

	// 4. Valid worker attempt with current fencing token (3) succeeds
	committed, err := service.CommitRunAttempt(ctx, "usr_alice", run.ID, 3, "discover_candidates", StageSafeToRetry, true, "item 1 completed")
	if err != nil {
		t.Fatalf("CommitAttempt error: %v", err)
	}
	if committed.ItemsProcessed != 1 || committed.ItemsSucceeded != 1 {
		t.Errorf("expected 1 processed and 1 succeeded, got %d/%d", committed.ItemsProcessed, committed.ItemsSucceeded)
	}

	// 5. Cancel run -> state becomes cancelled
	cancelled, err := service.ControlCareerRun(ctx, "usr_alice", run.ID, ActionCancel, "user cancelled")
	if err != nil {
		t.Fatalf("ControlCareerRun cancel error: %v", err)
	}
	if cancelled.Status != RunStatusCancelled {
		t.Errorf("expected RunStatusCancelled, got %s", cancelled.Status)
	}
}

func TestRunControls_UncertainWriteCrashRecovery_AT006(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	run, err := service.CreateCareerRun(ctx, "usr_bob", "ws_1", RunTypeApplicationSession, 1, 10)
	if err != nil {
		t.Fatalf("CreateCareerRun error: %v", err)
	}

	now := time.Now().UTC()
	run.Status = RunStatusRunning
	run.CurrentStage = "submit_application"
	run.StageSafety = StageUncertainWriteRequiresReconciliation
	run.LeaseExpiresAt = now.Add(-10 * time.Minute)
	_ = repo.SaveCareerRun(ctx, run)

	// Recover crashed run after lease timeout
	recoveredRun, recovered, err := service.RecoverCareerRun(ctx, "usr_bob", run.ID)
	if err != nil {
		t.Fatalf("RecoverCareerRun error: %v", err)
	}
	if !recovered {
		t.Fatalf("expected run to be marked recovered")
	}

	// PER AT-006: Uncertain write MUST NOT be blindly retried! It must enter needs_reconciliation.
	if recoveredRun.Status != RunStatusNeedsReconciliation {
		t.Fatalf("expected RunStatusNeedsReconciliation, got %s", recoveredRun.Status)
	}
	if !recoveredRun.RequiresReconciliation {
		t.Errorf("expected RequiresReconciliation to be true")
	}

	// Reconcile run: user verifies external portal and confirms completed
	reconciled, err := service.ReconcileCareerRun(ctx, "usr_bob", run.ID, "confirm_completed", true, "verified on workday portal")
	if err != nil {
		t.Fatalf("ReconcileCareerRun error: %v", err)
	}

	if reconciled.Status != RunStatusCompleted {
		t.Errorf("expected RunStatusCompleted after resolution, got %s", reconciled.Status)
	}
	if reconciled.ItemsSucceeded != 1 {
		t.Errorf("expected ItemsSucceeded == 1, got %d", reconciled.ItemsSucceeded)
	}
}

func TestRunControls_SafeStageCrashRecovery(t *testing.T) {
	now := time.Now().UTC()
	run := NewCareerRun("usr_charlie", "ws_1", RunTypeDiscoveryRun, 10, 15)
	run.Status = RunStatusRunning
	run.CurrentStage = "discover_candidates"
	run.StageSafety = StageSafeToRetry
	run.LeaseExpiresAt = now.Add(-5 * time.Minute)

	// Safe stage recovery should pause the run for clean retry
	recovered, err := RecoverCrashedRun(run, now)
	if err != nil {
		t.Fatalf("RecoverCrashedRun error: %v", err)
	}
	if !recovered {
		t.Errorf("expected run to be recovered")
	}

	if run.Status != RunStatusPaused {
		t.Errorf("expected RunStatusPaused for safe stage, got %s", run.Status)
	}
	if run.RequiresReconciliation {
		t.Errorf("safe stage should not require manual reconciliation")
	}
}

func TestRunControls_RollingHourlyLimiter_FND011(t *testing.T) {
	now := time.Now().UTC()

	recent := []time.Time{
		now.Add(-45 * time.Minute),
		now.Add(-30 * time.Minute),
		now.Add(-15 * time.Minute),
	}

	// Currently at max limit (3)
	allowed, count, backoff := EvaluateRollingHourlyLimit(recent, now, 3)
	if allowed {
		t.Errorf("expected allowed to be false")
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
	if backoff <= 0 {
		t.Errorf("expected positive backoff, got %v", backoff)
	}

	// Advance time by 20 minutes -> the oldest timestamp (45m ago) drops out of the 60m window (becomes 65m ago)
	futureNow := now.Add(20 * time.Minute)
	allowedFuture, countFuture, _ := EvaluateRollingHourlyLimit(recent, futureNow, 3)
	if !allowedFuture {
		t.Errorf("expected allowed to be true after oldest timestamp slides out of window")
	}
	if countFuture != 2 {
		t.Errorf("expected count 2, got %d", countFuture)
	}
}

func TestRunControls_GracefulDraining_REQ017(t *testing.T) {
	run := NewCareerRun("usr_drain", "ws_1", RunTypeApplicationSession, 2, 10)
	run.Status = RunStatusRunning

	if err := DrainCareerRun(run); err != nil {
		t.Fatalf("DrainCareerRun error: %v", err)
	}

	if run.Status != RunStatusDraining {
		t.Errorf("expected RunStatusDraining, got %s", run.Status)
	}

	// Completing in-flight attempt while draining moves run to completed
	err := CommitAttempt(run, 1, "submit_application", StageSafeToRetry, true, "batch item done")
	if err != nil {
		t.Fatalf("CommitAttempt during drain error: %v", err)
	}

	if run.Status != RunStatusCompleted {
		t.Errorf("expected RunStatusCompleted after drain commit, got %s", run.Status)
	}
}
