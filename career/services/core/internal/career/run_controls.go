package career

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// CareerRunType specifies the type of asynchronous automated or assisted run.
type CareerRunType string

const (
	RunTypeDiscoveryRun       CareerRunType = "discovery_run"
	RunTypeTailoringBatch     CareerRunType = "tailoring_batch"
	RunTypeApplicationSession CareerRunType = "application_session"
	RunTypeFeedScan           CareerRunType = "feed_scan"
)

// CareerRunStatus specifies the lifecycle phase of a run (CAR-21, REQ-017).
type CareerRunStatus string

const (
	RunStatusQueued              CareerRunStatus = "queued"
	RunStatusRunning             CareerRunStatus = "running"
	RunStatusPaused              CareerRunStatus = "paused"
	RunStatusDraining            CareerRunStatus = "draining"
	RunStatusCompleted           CareerRunStatus = "completed"
	RunStatusFailed              CareerRunStatus = "failed"
	RunStatusCancelled           CareerRunStatus = "cancelled"
	RunStatusNeedsReconciliation CareerRunStatus = "needs_reconciliation"
	RunStatusCrashed             CareerRunStatus = "crashed"
)

// StageSafetyLevel defines whether a stage can be retried safely or requires reconciliation (AT-006).
type StageSafetyLevel string

const (
	StageSafeToRetry                         StageSafetyLevel = "safe_to_retry"
	StageUncertainWriteRequiresReconciliation StageSafetyLevel = "uncertain_write_requires_reconciliation"
)

// RunControlAction represents user actions to control run execution.
type RunControlAction string

const (
	ActionPause  RunControlAction = "pause"
	ActionResume RunControlAction = "resume"
	ActionCancel RunControlAction = "cancel"
	ActionDrain  RunControlAction = "drain"
)

var (
	ErrRunNotFound               = errors.New("career run not found")
	ErrRunAlreadyFinished        = errors.New("cannot modify run: run is already completed, failed, or cancelled")
	ErrStaleAttemptFenced        = errors.New("stale attempt fenced: worker token is less than current run fencing token (AT-021)")
	ErrUncertainWriteBlindRetry  = errors.New("cannot blindly retry uncertain write stage: reconciliation required to prevent duplicates (AT-006)")
	ErrHourlyRateLimitExceeded   = errors.New("rolling hourly submission limit exceeded; backoff required (FND-011)")
	ErrInvalidRunAction          = errors.New("invalid run control action")
	ErrRunNotPaused              = errors.New("run is not paused; cannot resume")
)

// RunStageCheckpoint represents a recorded execution checkpoint for a specific phase.
type RunStageCheckpoint struct {
	StageName       string           `json:"stage_name"`
	Safety          StageSafetyLevel `json:"safety"`
	StartedAt       time.Time        `json:"started_at"`
	CompletedAt     *time.Time       `json:"completed_at,omitempty"`
	Status          string           `json:"status"` // "success", "failed", "interrupted"
	ItemsProcessed  int              `json:"items_processed"`
	Diagnostics     string           `json:"diagnostics,omitempty"`
}

// RunDiagnosticLog represents an immutable timestamped log event with fencing token.
type RunDiagnosticLog struct {
	Timestamp    time.Time `json:"timestamp"`
	Level        string    `json:"level"` // "info", "warn", "error", "security"
	Stage        string    `json:"stage"`
	Message      string    `json:"message"`
	FencingToken int64     `json:"fencing_token"`
}

// CareerRun represents a durable, crash-resilient task execution unit (CAR-21, AT-006, AT-021).
type CareerRun struct {
	ID                      string               `json:"id"`
	UserID                  string               `json:"user_id"`
	WorkspaceID             string               `json:"workspace_id"`
	RunType                 CareerRunType        `json:"run_type"`
	Status                  CareerRunStatus      `json:"status"`
	CurrentStage            string               `json:"current_stage"`
	StageSafety             StageSafetyLevel     `json:"stage_safety"`
	FencingToken            int64                `json:"fencing_token"` // Monotonically increasing (AT-021)
	ItemsTotal              int                  `json:"items_total"`
	ItemsProcessed          int                  `json:"items_processed"`
	ItemsSucceeded          int                  `json:"items_succeeded"`
	ItemsFailed             int                  `json:"items_failed"`
	HourlyLimit             int                  `json:"hourly_limit"` // e.g. 15 submissions/hr (FND-011)
	RecentSubmissions       []time.Time          `json:"recent_submissions"`
	Checkpoints             []RunStageCheckpoint `json:"checkpoints"`
	AuditLog                []RunDiagnosticLog   `json:"audit_log"`
	HeartbeatAt             time.Time            `json:"heartbeat_at"`
	LeaseExpiresAt          time.Time            `json:"lease_expires_at"`
	RequiresReconciliation  bool                 `json:"requires_reconciliation"`
	ReconciliationNotes     string               `json:"reconciliation_notes,omitempty"`
	CreatedAt               time.Time            `json:"created_at"`
	UpdatedAt               time.Time            `json:"updated_at"`
}

// NewCareerRun initializes a durable run with fencing token 1.
func NewCareerRun(userID, workspaceID string, runType CareerRunType, totalItems, hourlyLimit int) *CareerRun {
	now := time.Now().UTC()
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s:%s:%s:%d", userID, workspaceID, runType, now.UnixNano())))
	runID := "run_" + hex.EncodeToString(hasher.Sum(nil))[:16]

	if hourlyLimit <= 0 {
		hourlyLimit = 15 // Default 15 submissions/hr (FND-011)
	}

	run := &CareerRun{
		ID:                     runID,
		UserID:                 userID,
		WorkspaceID:            workspaceID,
		RunType:                runType,
		Status:                 RunStatusQueued,
		CurrentStage:           "initialization",
		StageSafety:            StageSafeToRetry,
		FencingToken:           1, // Initial fencing token (AT-021)
		ItemsTotal:             totalItems,
		ItemsProcessed:         0,
		ItemsSucceeded:         0,
		ItemsFailed:            0,
		HourlyLimit:            hourlyLimit,
		RecentSubmissions:      make([]time.Time, 0),
		Checkpoints:            make([]RunStageCheckpoint, 0),
		AuditLog:               make([]RunDiagnosticLog, 0),
		HeartbeatAt:            now,
		LeaseExpiresAt:         now.Add(2 * time.Minute),
		RequiresReconciliation: false,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	run.appendLog("info", "initialization", "Run initialized and queued for execution")
	return run
}

func (r *CareerRun) appendLog(level, stage, msg string) {
	r.AuditLog = append(r.AuditLog, RunDiagnosticLog{
		Timestamp:    time.Now().UTC(),
		Level:        level,
		Stage:        stage,
		Message:      msg,
		FencingToken: r.FencingToken,
	})
}

// PauseCareerRun gracefully pauses an active or queued run and increments the fencing token (AT-021).
func PauseCareerRun(run *CareerRun, reason string) error {
	if run == nil {
		return ErrRunNotFound
	}
	if run.Status == RunStatusCompleted || run.Status == RunStatusFailed || run.Status == RunStatusCancelled {
		return ErrRunAlreadyFinished
	}

	run.FencingToken++ // Invalidate active workers (AT-021)
	run.Status = RunStatusPaused
	run.UpdatedAt = time.Now().UTC()
	msg := "Run paused by user"
	if reason != "" {
		msg += ": " + reason
	}
	run.appendLog("warn", run.CurrentStage, msg)
	return nil
}

// ResumeCareerRun restores a paused run after validating the rolling rate limit (FND-011).
func ResumeCareerRun(run *CareerRun, now time.Time) error {
	if run == nil {
		return ErrRunNotFound
	}
	if run.Status != RunStatusPaused {
		return ErrRunNotPaused
	}

	// Validate rate limit before resuming
	allowed, count, backoff := EvaluateRollingHourlyLimit(run.RecentSubmissions, now, run.HourlyLimit)
	if !allowed {
		run.appendLog("warn", run.CurrentStage, fmt.Sprintf("Resume throttled by rolling rate limit: %d/%d used, backoff %v", count, run.HourlyLimit, backoff))
		return fmt.Errorf("%w: %d/%d submissions in current window; wait %v", ErrHourlyRateLimitExceeded, count, run.HourlyLimit, backoff)
	}

	run.FencingToken++ // New lease generation (AT-021)
	run.Status = RunStatusRunning
	run.HeartbeatAt = now
	run.LeaseExpiresAt = now.Add(2 * time.Minute)
	run.UpdatedAt = now
	run.appendLog("info", run.CurrentStage, "Run resumed from checkpoint")
	return nil
}

// CancelCareerRun permanently terminates the run and fences all in-flight workers (AT-021).
func CancelCareerRun(run *CareerRun, reason string) error {
	if run == nil {
		return ErrRunNotFound
	}
	if run.Status == RunStatusCompleted || run.Status == RunStatusCancelled {
		return ErrRunAlreadyFinished
	}

	run.FencingToken++ // Stale token fencing (AT-021)
	run.Status = RunStatusCancelled
	run.UpdatedAt = time.Now().UTC()
	msg := "Run cancelled by user"
	if reason != "" {
		msg += ": " + reason
	}
	run.appendLog("error", run.CurrentStage, msg)
	return nil
}

// DrainCareerRun enters graceful draining mode, allowing currently in-flight work to finish (SRC-C3, REQ-017).
func DrainCareerRun(run *CareerRun) error {
	if run == nil {
		return ErrRunNotFound
	}
	if run.Status != RunStatusRunning {
		return errors.New("only running runs can be drained")
	}

	run.Status = RunStatusDraining
	run.UpdatedAt = time.Now().UTC()
	run.appendLog("info", run.CurrentStage, "Graceful draining initiated: completing current batch item and shutting down")
	return nil
}

// CommitAttempt records a completed worker attempt, strictly validating the fencing token (AT-021).
func CommitAttempt(run *CareerRun, token int64, stage string, safety StageSafetyLevel, succeeded bool, diagnostics string) error {
	if run == nil {
		return ErrRunNotFound
	}

	// Stale Attempt Fencing Check (AT-021)
	if token < run.FencingToken {
		run.appendLog("error", stage, fmt.Sprintf("Rejected stale attempt with token %d (current fencing token: %d)", token, run.FencingToken))
		return fmt.Errorf("%w: attempt token %d < run token %d", ErrStaleAttemptFenced, token, run.FencingToken)
	}

	now := time.Now().UTC()
	run.CurrentStage = stage
	run.StageSafety = safety
	run.HeartbeatAt = now
	run.LeaseExpiresAt = now.Add(2 * time.Minute)
	run.ItemsProcessed++

	if succeeded {
		run.ItemsSucceeded++
		run.RecentSubmissions = append(run.RecentSubmissions, now)
		run.Checkpoints = append(run.Checkpoints, RunStageCheckpoint{
			StageName:      stage,
			Safety:         safety,
			StartedAt:      now.Add(-time.Second),
			CompletedAt:    &now,
			Status:         "success",
			ItemsProcessed: run.ItemsProcessed,
			Diagnostics:    diagnostics,
		})
		run.appendLog("info", stage, fmt.Sprintf("Stage completed successfully: %s", diagnostics))
	} else {
		run.ItemsFailed++
		run.Checkpoints = append(run.Checkpoints, RunStageCheckpoint{
			StageName:      stage,
			Safety:         safety,
			StartedAt:      now.Add(-time.Second),
			CompletedAt:    &now,
			Status:         "failed",
			ItemsProcessed: run.ItemsProcessed,
			Diagnostics:    diagnostics,
		})
		run.appendLog("error", stage, fmt.Sprintf("Stage failed: %s", diagnostics))

		// If failure occurs on an uncertain write stage, flag for reconciliation (AT-006)
		if safety == StageUncertainWriteRequiresReconciliation {
			run.Status = RunStatusNeedsReconciliation
			run.RequiresReconciliation = true
			run.ReconciliationNotes = fmt.Sprintf("Uncertain write failure in stage '%s': %s", stage, diagnostics)
			run.appendLog("warn", stage, "Stage is an uncertain write; automatic retry blocked to prevent duplicates (AT-006)")
			return nil
		}
	}

	// Check if draining is complete
	if run.Status == RunStatusDraining {
		run.Status = RunStatusCompleted
		run.appendLog("info", stage, "Graceful draining complete; run stopped cleanly")
	} else if run.ItemsProcessed >= run.ItemsTotal {
		run.Status = RunStatusCompleted
		run.appendLog("info", stage, "All items processed; run marked completed")
	}

	run.UpdatedAt = now
	return nil
}

// RecoverCrashedRun checks for expired leases and handles recovery according to stage safety (AT-006, REQ-017).
func RecoverCrashedRun(run *CareerRun, now time.Time) (bool, error) {
	if run == nil {
		return false, ErrRunNotFound
	}

	// Only recover runs stuck in running or draining whose lease has expired
	if (run.Status != RunStatusRunning && run.Status != RunStatusDraining) || now.Before(run.LeaseExpiresAt) {
		return false, nil
	}

	run.FencingToken++ // Fence out the crashed worker

	if run.StageSafety == StageUncertainWriteRequiresReconciliation {
		// Zero-duplicate invariant: uncertain writes must NEVER blindly retry (AT-006)
		run.Status = RunStatusNeedsReconciliation
		run.RequiresReconciliation = true
		run.ReconciliationNotes = fmt.Sprintf("Worker crashed during uncertain write stage '%s' at %s. Blind retry blocked (AT-006).", run.CurrentStage, now.Format(time.RFC3339))
		run.appendLog("error", run.CurrentStage, "CRASH DETECTED during uncertain write: manual reconciliation required (AT-006)")
		run.UpdatedAt = now
		return true, nil
	}

	// Safe idempotent stages can be paused for user retry from checkpoint
	run.Status = RunStatusPaused
	run.appendLog("warn", run.CurrentStage, fmt.Sprintf("Worker crashed during safe stage '%s'. Checkpointed state preserved for safe resumption.", run.CurrentStage))
	run.UpdatedAt = now
	return true, nil
}

// ReconcileRun applies a human decision to a run in needs_reconciliation (AT-006).
func ReconcileRun(run *CareerRun, resolution string, markSucceeded bool, notes string) error {
	if run == nil {
		return ErrRunNotFound
	}
	if run.Status != RunStatusNeedsReconciliation {
		return errors.New("run does not require reconciliation")
	}

	now := time.Now().UTC()
	run.RequiresReconciliation = false
	run.ReconciliationNotes = notes
	run.UpdatedAt = now

	switch resolution {
	case "confirm_completed":
		run.Status = RunStatusCompleted
		if markSucceeded {
			run.ItemsSucceeded++
		}
		run.appendLog("info", run.CurrentStage, fmt.Sprintf("Reconciled as completed: %s", notes))

	case "abandon_attempt":
		run.Status = RunStatusFailed
		run.ItemsFailed++
		run.appendLog("warn", run.CurrentStage, fmt.Sprintf("Reconciled as abandoned: %s", notes))

	case "force_retry":
		run.Status = RunStatusRunning
		run.FencingToken++
		run.HeartbeatAt = now
		run.LeaseExpiresAt = now.Add(2 * time.Minute)
		run.appendLog("warn", run.CurrentStage, fmt.Sprintf("User explicitly forced retry of uncertain stage: %s", notes))

	default:
		return fmt.Errorf("invalid reconciliation resolution: %s", resolution)
	}

	return nil
}

// EvaluateRollingHourlyLimit calculates the submissions in the sliding 60-minute window (FND-011).
func EvaluateRollingHourlyLimit(submissions []time.Time, now time.Time, hourlyLimit int) (bool, int, time.Duration) {
	oneHourAgo := now.Add(-1 * time.Hour)
	count := 0
	var oldestInWindow time.Time

	for _, t := range submissions {
		if t.After(oneHourAgo) {
			count++
			if oldestInWindow.IsZero() || t.Before(oldestInWindow) {
				oldestInWindow = t
			}
		}
	}

	if count >= hourlyLimit {
		// Backoff until the oldest submission in the window falls outside 1 hour
		backoff := oldestInWindow.Add(1 * time.Hour).Sub(now)
		if backoff < 0 {
			backoff = 0
		}
		return false, count, backoff
	}

	return true, count, 0
}
