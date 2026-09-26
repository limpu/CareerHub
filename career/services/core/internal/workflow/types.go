package workflow

import (
	"encoding/json"
	"errors"
	"time"
)

type RunState string

const (
	StateQueued        RunState = "queued"
	StateRunning       RunState = "running"
	StateNeedsApproval RunState = "needs_approval"
	StatePaused        RunState = "paused"
	StateCompleted     RunState = "completed"
	StateFailed        RunState = "failed"
	StateCancelled     RunState = "cancelled"
)

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "pending"
	OutboxStatusDispatched OutboxStatus = "dispatched"
	OutboxStatusFailed     OutboxStatus = "failed"
)

var (
	ErrRunNotFound             = errors.New("action run not found")
	ErrRunAlreadyCompleted     = errors.New("action run is already completed")
	ErrRunPaused               = errors.New("action run is paused")
	ErrRunCancelled            = errors.New("action run is cancelled")
	ErrStaleAttemptFenced      = errors.New("attempt completion rejected: stale fencing token")
	ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")
	ErrInvalidStateTransition  = errors.New("invalid run state transition")
	ErrLeaseExpiredOrHeld      = errors.New("lease expired or held by another worker")
	ErrJobNotFound             = errors.New("scheduled job not found")
	ErrInvalidLease            = errors.New("invalid lease holder or expired lease")
)

// ActionRun represents durable task execution state in PostgreSQL.
type ActionRun struct {
	ID             string          `json:"id"`
	WorkspaceID    string          `json:"workspace_id"`
	ActorID        string          `json:"actor_id"`
	TaskType       string          `json:"task_type"`
	State          RunState        `json:"state"`
	InputPayload   json.RawMessage `json:"input_payload"`
	ResultData     json.RawMessage `json:"result_data,omitempty"`
	ErrorDetails   json.RawMessage `json:"error_details,omitempty"`
	IdempotencyKey string          `json:"idempotency_key"`
	AttemptCount   int             `json:"attempt_count"`
	CurrentAttempt int             `json:"current_attempt"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// ActionAttempt tracks individual execution attempts of an ActionRun.
type ActionAttempt struct {
	ID            string    `json:"id"`
	ActionRunID   string    `json:"action_run_id"`
	AttemptNumber int       `json:"attempt_number"`
	WorkerID      string    `json:"worker_id"`
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   *time.Time`json:"completed_at,omitempty"`
	Error         string    `json:"error,omitempty"`
	Status        string    `json:"status"` // "running", "success", "failed"
}

// OutboxEvent represents a transactional outbox event waiting to be relayed to Redis Streams.
type OutboxEvent struct {
	ID            string          `json:"id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	Status        OutboxStatus    `json:"status"`
	RetryCount    int             `json:"retry_count"`
	CreatedAt     time.Time       `json:"created_at"`
	ProcessedAt   *time.Time      `json:"processed_at,omitempty"`
}
