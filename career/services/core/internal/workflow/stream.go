package workflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrConsumerGroupExists = errors.New("consumer group already exists")
	ErrStreamNotFound      = errors.New("stream not found")
	ErrGroupNotFound       = errors.New("consumer group not found")
)

// StreamMessage represents a consumed Redis Stream entry containing a TaskEnvelope.
type StreamMessage struct {
	ID       string        `json:"id"`
	Stream   string        `json:"stream"`
	Envelope *TaskEnvelope `json:"envelope"`
}

// StreamEngine defines the Redis Streams abstraction contract.
type StreamEngine interface {
	Publish(ctx context.Context, stream string, env *TaskEnvelope) (string, error)
	CreateConsumerGroup(ctx context.Context, stream, group, startID string) error
	ReadGroup(ctx context.Context, stream, group, consumer string, count int64, block time.Duration) ([]StreamMessage, error)
	Ack(ctx context.Context, stream, group string, ids ...string) error
	ClaimStale(ctx context.Context, stream, group, consumer string, minIdle time.Duration, count int64) ([]StreamMessage, error)
}

// PendingEntry tracks in-flight leased messages for Redis Streams consumer groups.
type PendingEntry struct {
	MessageID   string
	Consumer    string
	DeliveredAt time.Time
	Envelope    *TaskEnvelope
}

// MemoryStreamEngine provides an in-memory, deterministic simulation of Redis Streams (PEL, Ack, Claim).
type MemoryStreamEngine struct {
	mu      sync.RWMutex
	streams map[string][]StreamMessage
	groups  map[string]map[string]int64 // stream -> groupName -> lastDeliveredIndex
	pels    map[string]map[string]map[string]*PendingEntry // stream -> groupName -> messageID -> entry
	seq     int64
}

func NewMemoryStreamEngine() *MemoryStreamEngine {
	return &MemoryStreamEngine{
		streams: make(map[string][]StreamMessage),
		groups:  make(map[string]map[string]int64),
		pels:    make(map[string]map[string]map[string]*PendingEntry),
	}
}

func (e *MemoryStreamEngine) Publish(ctx context.Context, stream string, env *TaskEnvelope) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.seq++
	msgID := fmt.Sprintf("%d-%d", time.Now().UnixMilli(), e.seq)
	msg := StreamMessage{
		ID:       msgID,
		Stream:   stream,
		Envelope: env,
	}

	e.streams[stream] = append(e.streams[stream], msg)
	return msgID, nil
}

func (e *MemoryStreamEngine) CreateConsumerGroup(ctx context.Context, stream, group, startID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.groups[stream] == nil {
		e.groups[stream] = make(map[string]int64)
	}
	if _, ok := e.groups[stream][group]; ok {
		return ErrConsumerGroupExists
	}

	e.groups[stream][group] = 0

	if e.pels[stream] == nil {
		e.pels[stream] = make(map[string]map[string]*PendingEntry)
	}
	if e.pels[stream][group] == nil {
		e.pels[stream][group] = make(map[string]*PendingEntry)
	}

	return nil
}

func (e *MemoryStreamEngine) ReadGroup(ctx context.Context, stream, group, consumer string, count int64, block time.Duration) ([]StreamMessage, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	groupMap, ok := e.groups[stream]
	if !ok {
		return nil, ErrStreamNotFound
	}
	lastIdx, ok := groupMap[group]
	if !ok {
		return nil, ErrGroupNotFound
	}

	streamMsgs := e.streams[stream]
	total := int64(len(streamMsgs))
	if lastIdx >= total {
		return []StreamMessage{}, nil
	}

	var batch []StreamMessage
	for i := lastIdx; i < total && int64(len(batch)) < count; i++ {
		msg := streamMsgs[i]
		batch = append(batch, msg)

		// Record in PEL (Pending Entries List)
		e.pels[stream][group][msg.ID] = &PendingEntry{
			MessageID:   msg.ID,
			Consumer:    consumer,
			DeliveredAt: time.Now(),
			Envelope:    msg.Envelope,
		}
	}

	e.groups[stream][group] = lastIdx + int64(len(batch))
	return batch, nil
}

func (e *MemoryStreamEngine) Ack(ctx context.Context, stream, group string, ids ...string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	streamPel, ok := e.pels[stream]
	if !ok {
		return nil
	}
	groupPel, ok := streamPel[group]
	if !ok {
		return nil
	}

	for _, id := range ids {
		delete(groupPel, id)
	}
	return nil
}

func (e *MemoryStreamEngine) ClaimStale(ctx context.Context, stream, group, consumer string, minIdle time.Duration, count int64) ([]StreamMessage, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	streamPel, ok := e.pels[stream]
	if !ok {
		return nil, ErrStreamNotFound
	}
	groupPel, ok := streamPel[group]
	if !ok {
		return nil, ErrGroupNotFound
	}

	now := time.Now()
	var claimed []StreamMessage

	for msgID, entry := range groupPel {
		if int64(len(claimed)) >= count {
			break
		}
		if now.Sub(entry.DeliveredAt) >= minIdle {
			// Transfer ownership to the new consumer
			entry.Consumer = consumer
			entry.DeliveredAt = now

			claimed = append(claimed, StreamMessage{
				ID:       msgID,
				Stream:   stream,
				Envelope: entry.Envelope,
			})
		}
	}

	return claimed, nil
}
