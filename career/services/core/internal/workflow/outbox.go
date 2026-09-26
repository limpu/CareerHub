package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEventNotFound = errors.New("outbox event not found")
)

type OutboxRepository interface {
	SaveEvent(ctx context.Context, evt *OutboxEvent) error
	GetPendingEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)
	MarkDispatched(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, reason string) error
}

type MemoryOutboxRepository struct {
	mu     sync.RWMutex
	events map[string]*OutboxEvent
}

func NewMemoryOutboxRepository() *MemoryOutboxRepository {
	return &MemoryOutboxRepository{
		events: make(map[string]*OutboxEvent),
	}
}

func (r *MemoryOutboxRepository) SaveEvent(ctx context.Context, evt *OutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[evt.ID] = evt
	return nil
}

func (r *MemoryOutboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]*OutboxEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var pending []*OutboxEvent
	for _, evt := range r.events {
		if evt.Status == OutboxStatusPending {
			pending = append(pending, evt)
			if len(pending) >= limit {
				break
			}
		}
	}
	return pending, nil
}

func (r *MemoryOutboxRepository) MarkDispatched(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	evt, ok := r.events[id]
	if !ok {
		return ErrEventNotFound
	}
	now := time.Now().UTC()
	evt.Status = OutboxStatusDispatched
	evt.ProcessedAt = &now
	return nil
}

func (r *MemoryOutboxRepository) MarkFailed(ctx context.Context, id string, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	evt, ok := r.events[id]
	if !ok {
		return ErrEventNotFound
	}
	evt.RetryCount++
	if evt.RetryCount >= 5 {
		evt.Status = OutboxStatusFailed
	}
	return nil
}

// OutboxService creates transactional events in PostgreSQL.
type OutboxService struct {
	repo OutboxRepository
}

func NewOutboxService(repo OutboxRepository) *OutboxService {
	return &OutboxService{repo: repo}
}

func (s *OutboxService) RecordEvent(ctx context.Context, aggregateType, aggregateID, eventType string, payload json.RawMessage) (*OutboxEvent, error) {
	evt := &OutboxEvent{
		ID:            uuid.New().String(),
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       payload,
		Status:        OutboxStatusPending,
		RetryCount:    0,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.repo.SaveEvent(ctx, evt); err != nil {
		return nil, fmt.Errorf("failed to save outbox event: %w", err)
	}
	return evt, nil
}

// OutboxRelay polls pending outbox events and pushes them to Redis Streams with idempotency.
type OutboxRelay struct {
	repo   OutboxRepository
	stream StreamEngine
}

func NewOutboxRelay(repo OutboxRepository, stream StreamEngine) *OutboxRelay {
	return &OutboxRelay{
		repo:   repo,
		stream: stream,
	}
}

// ProcessBatch reads up to limit pending outbox events and publishes to the corresponding stream.
func (r *OutboxRelay) ProcessBatch(ctx context.Context, limit int) (int, error) {
	events, err := r.repo.GetPendingEvents(ctx, limit)
	if err != nil {
		return 0, err
	}

	dispatched := 0
	for _, evt := range events {
		// Map event to envelope
		var payloadMap map[string]interface{}
		_ = json.Unmarshal(evt.Payload, &payloadMap)

		section := "career"
		if s, ok := payloadMap["section"].(string); ok && s != "" {
			section = s
		}
		actorID := "system"
		if a, ok := payloadMap["actor_id"].(string); ok && a != "" {
			actorID = a
		}
		workspaceID := "system"
		if w, ok := payloadMap["workspace_id"].(string); ok && w != "" {
			workspaceID = w
		}

		env, err := NewTaskEnvelope(section, evt.EventType, actorID, workspaceID, evt.ID, evt.Payload)
		if err != nil {
			_ = r.repo.MarkFailed(ctx, evt.ID, err.Error())
			continue
		}

		streamName := fmt.Sprintf("stream:%s:%s", section, evt.EventType)
		_, pubErr := r.stream.Publish(ctx, streamName, env)
		if pubErr != nil {
			_ = r.repo.MarkFailed(ctx, evt.ID, pubErr.Error())
			continue
		}

		if err := r.repo.MarkDispatched(ctx, evt.ID); err == nil {
			dispatched++
		}
	}

	return dispatched, nil
}

// Start begins a background polling loop for pending outbox events.
func (r *OutboxRelay) Start(ctx context.Context, pollInterval time.Duration, batchSize int) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = r.ProcessBatch(ctx, batchSize)
		}
	}
}
