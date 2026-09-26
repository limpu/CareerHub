package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEnvelope = errors.New("task envelope validation failed")
)

// TaskEnvelope represents a versioned event/job envelope conforming to monorepo contracts.
type TaskEnvelope struct {
	ID             string          `json:"id"`
	Version        string          `json:"version"`
	Section        string          `json:"section"`   // "career", "linkedin", "social"
	TaskType       string          `json:"taskType"`  // e.g. "document.parse", "job.match"
	ActorID        string          `json:"actorId"`
	WorkspaceID    string          `json:"workspaceId"`
	Payload        json.RawMessage `json:"payload"`
	IdempotencyKey string          `json:"idempotencyKey"`
	CorrelationID  string          `json:"correlationId"`
	AttemptNumber  int             `json:"attemptNumber"`
	CreatedAt      time.Time       `json:"createdAt"`
}

// NewTaskEnvelope constructs a valid, versioned task envelope.
func NewTaskEnvelope(section, taskType, actorID, workspaceID, idempotencyKey string, payload json.RawMessage) (*TaskEnvelope, error) {
	if section == "" || taskType == "" || actorID == "" || workspaceID == "" {
		return nil, fmt.Errorf("%w: missing required envelope metadata", ErrInvalidEnvelope)
	}

	if idempotencyKey == "" {
		idempotencyKey = uuid.New().String()
	}

	envID := uuid.New().String()
	return &TaskEnvelope{
		ID:             envID,
		Version:        "v1",
		Section:        section,
		TaskType:       taskType,
		ActorID:        actorID,
		WorkspaceID:    workspaceID,
		Payload:        payload,
		IdempotencyKey: idempotencyKey,
		CorrelationID:  uuid.New().String(),
		AttemptNumber:  1,
		CreatedAt:      time.Now().UTC(),
	}, nil
}

// ToJSON serializes the task envelope to JSON bytes.
func (e *TaskEnvelope) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializes JSON bytes into a TaskEnvelope.
func FromJSON(data []byte) (*TaskEnvelope, error) {
	var env TaskEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidEnvelope, err)
	}
	if env.ID == "" || env.TaskType == "" || env.Version == "" {
		return nil, fmt.Errorf("%w: missing critical envelope fields", ErrInvalidEnvelope)
	}
	return &env, nil
}
