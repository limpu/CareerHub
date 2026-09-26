package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type RunLedgerRepository interface {
	CreateRun(ctx context.Context, run *ActionRun) error
	GetRun(ctx context.Context, id string) (*ActionRun, error)
	GetRunByIdempotencyKey(ctx context.Context, key string) (*ActionRun, error)
	UpdateRun(ctx context.Context, run *ActionRun) error
	CreateAttempt(ctx context.Context, attempt *ActionAttempt) error
	UpdateAttempt(ctx context.Context, attempt *ActionAttempt) error
	GetAttempts(ctx context.Context, runID string) ([]*ActionAttempt, error)
	ListRuns(ctx context.Context, workspaceID string) ([]*ActionRun, error)
}

type MemoryRunLedgerRepository struct {
	mu          sync.RWMutex
	runs        map[string]*ActionRun
	idempMap    map[string]string // idempotencyKey -> runID
	attempts    map[string]*ActionAttempt
}

func NewMemoryRunLedgerRepository() *MemoryRunLedgerRepository {
	return &MemoryRunLedgerRepository{
		runs:     make(map[string]*ActionRun),
		idempMap: make(map[string]string),
		attempts: make(map[string]*ActionAttempt),
	}
}

func (r *MemoryRunLedgerRepository) CreateRun(ctx context.Context, run *ActionRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if run.IdempotencyKey != "" {
		if existingID, ok := r.idempMap[run.IdempotencyKey]; ok {
			return fmt.Errorf("%w: key=%s (run_id=%s)", ErrDuplicateIdempotencyKey, run.IdempotencyKey, existingID)
		}
		r.idempMap[run.IdempotencyKey] = run.ID
	}

	r.runs[run.ID] = run
	return nil
}

func (r *MemoryRunLedgerRepository) GetRun(ctx context.Context, id string) (*ActionRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	run, ok := r.runs[id]
	if !ok {
		return nil, ErrRunNotFound
	}
	return run, nil
}

func (r *MemoryRunLedgerRepository) GetRunByIdempotencyKey(ctx context.Context, key string) (*ActionRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	runID, ok := r.idempMap[key]
	if !ok {
		return nil, ErrRunNotFound
	}
	return r.runs[runID], nil
}

func (r *MemoryRunLedgerRepository) UpdateRun(ctx context.Context, run *ActionRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.runs[run.ID]; !ok {
		return ErrRunNotFound
	}
	run.UpdatedAt = time.Now().UTC()
	r.runs[run.ID] = run
	return nil
}

func (r *MemoryRunLedgerRepository) CreateAttempt(ctx context.Context, attempt *ActionAttempt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.attempts[attempt.ID] = attempt
	return nil
}

func (r *MemoryRunLedgerRepository) UpdateAttempt(ctx context.Context, attempt *ActionAttempt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.attempts[attempt.ID] = attempt
	return nil
}

func (r *MemoryRunLedgerRepository) GetAttempts(ctx context.Context, runID string) ([]*ActionAttempt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ActionAttempt
	for _, a := range r.attempts {
		if a.ActionRunID == runID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (r *MemoryRunLedgerRepository) ListRuns(ctx context.Context, workspaceID string) ([]*ActionRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ActionRun
	for _, run := range r.runs {
		if workspaceID == "" || run.WorkspaceID == workspaceID {
			clone := *run
			result = append(result, &clone)
		}
	}
	return result, nil
}

// RunLedgerService coordinates stateful runs, attempts, and fencing token verification.
type RunLedgerService struct {
	repo RunLedgerRepository
}

func NewRunLedgerService(repo RunLedgerRepository) *RunLedgerService {
	return &RunLedgerService{repo: repo}
}

// ListRuns returns all action runs in a workspace.
func (s *RunLedgerService) ListRuns(ctx context.Context, workspaceID string) ([]*ActionRun, error) {
	return s.repo.ListRuns(ctx, workspaceID)
}

// GetRun retrieves a run by its ID.
func (s *RunLedgerService) GetRun(ctx context.Context, id string) (*ActionRun, error) {
	return s.repo.GetRun(ctx, id)
}

// GetAttempts retrieves all attempts for a given run ID.
func (s *RunLedgerService) GetAttempts(ctx context.Context, runID string) ([]*ActionAttempt, error) {
	return s.repo.GetAttempts(ctx, runID)
}

// RegisterOrGetRun idempotently creates or retrieves an existing ActionRun (REQ-017, AT-021).
func (s *RunLedgerService) RegisterOrGetRun(ctx context.Context, run *ActionRun) (*ActionRun, bool, error) {
	if run.IdempotencyKey != "" {
		existing, err := s.repo.GetRunByIdempotencyKey(ctx, run.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, false, nil // already exists (idempotent replay)
		}
	}

	if run.ID == "" {
		run.ID = uuid.New().String()
	}
	run.State = StateQueued
	run.AttemptCount = 0
	run.CurrentAttempt = 0
	run.CreatedAt = time.Now().UTC()
	run.UpdatedAt = run.CreatedAt

	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, false, err
	}
	return run, true, nil
}

// StartAttempt begins a leased attempt and returns a fencing token (attemptNumber).
func (s *RunLedgerService) StartAttempt(ctx context.Context, runID, workerID string) (int, error) {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return 0, err
	}

	if run.State == StateCompleted {
		return 0, ErrRunAlreadyCompleted
	}
	if run.State == StatePaused {
		return 0, ErrRunPaused
	}
	if run.State == StateCancelled {
		return 0, ErrRunCancelled
	}

	run.AttemptCount++
	run.CurrentAttempt = run.AttemptCount
	run.State = StateRunning
	run.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateRun(ctx, run); err != nil {
		return 0, err
	}

	attempt := &ActionAttempt{
		ID:            uuid.New().String(),
		ActionRunID:   run.ID,
		AttemptNumber: run.CurrentAttempt,
		WorkerID:      workerID,
		StartedAt:     time.Now().UTC(),
		Status:        "running",
	}

	if err := s.repo.CreateAttempt(ctx, attempt); err != nil {
		return 0, err
	}

	return run.CurrentAttempt, nil
}

// CompleteRun marks a run as completed if the fencing token matches.
func (s *RunLedgerService) CompleteRun(ctx context.Context, runID string, attemptNumber int, resultData json.RawMessage) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}

	// Idempotent safeguard: if already completed, do not produce duplicate side effects
	if run.State == StateCompleted {
		return nil
	}

	// Cancelled runs must reject late worker completions (AT-021)
	if run.State == StateCancelled {
		return ErrRunCancelled
	}

	// Fencing check: reject stale/superseded attempts
	if run.CurrentAttempt != attemptNumber {
		return ErrStaleAttemptFenced
	}

	run.State = StateCompleted
	run.ResultData = resultData
	run.UpdatedAt = time.Now().UTC()

	return s.repo.UpdateRun(ctx, run)
}

// FailRun marks a run as failed with error details if the fencing token matches.
func (s *RunLedgerService) FailRun(ctx context.Context, runID string, attemptNumber int, errorDetails json.RawMessage) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}

	if run.State == StateCompleted {
		return nil
	}
	if run.State == StateCancelled {
		return ErrRunCancelled
	}

	if run.CurrentAttempt != attemptNumber {
		return ErrStaleAttemptFenced
	}

	run.State = StateFailed
	run.ErrorDetails = errorDetails
	run.UpdatedAt = time.Now().UTC()

	return s.repo.UpdateRun(ctx, run)
}

// PauseRun pauses an in-flight or queued run (REQ-019, AT-018).
func (s *RunLedgerService) PauseRun(ctx context.Context, runID string) (*ActionRun, error) {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}

	if run.State == StateCompleted || run.State == StateCancelled || run.State == StateFailed {
		return nil, fmt.Errorf("%w: cannot pause run in state %s", ErrInvalidStateTransition, run.State)
	}

	run.State = StatePaused
	run.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateRun(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

// ResumeRun transitions a paused run back to queued state (REQ-019, AT-018).
func (s *RunLedgerService) ResumeRun(ctx context.Context, runID string) (*ActionRun, error) {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}

	if run.State != StatePaused {
		return nil, fmt.Errorf("%w: cannot resume run in state %s", ErrInvalidStateTransition, run.State)
	}

	run.State = StateQueued
	run.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateRun(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

// CancelRun terminates a run permanently and fences off future completions (REQ-019, AT-021).
func (s *RunLedgerService) CancelRun(ctx context.Context, runID, reason string) (*ActionRun, error) {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}

	if run.State == StateCompleted {
		return nil, fmt.Errorf("%w: cannot cancel already completed run", ErrInvalidStateTransition)
	}

	run.State = StateCancelled
	details, _ := json.Marshal(map[string]string{
		"reason":      reason,
		"cancelledAt": time.Now().UTC().Format(time.RFC3339),
	})
	run.ErrorDetails = details
	run.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateRun(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}
