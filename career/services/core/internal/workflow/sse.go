package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RunEvent represents a durable, sequenced event emitted during task/run execution (REQ-019).
type RunEvent struct {
	ID        string          `json:"id"`
	RunID     string          `json:"run_id"`
	Sequence  int64           `json:"sequence"`
	EventType string          `json:"event_type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

// RunEventStore stores ordered run events for SSE reconnection replay.
type RunEventStore interface {
	AppendEvent(ctx context.Context, runID, eventType string, data json.RawMessage) (*RunEvent, error)
	GetEventsAfter(ctx context.Context, runID, lastEventID string) ([]*RunEvent, error)
	GetAllEvents(ctx context.Context, runID string) ([]*RunEvent, error)
}

// MemoryRunEventStore implements RunEventStore in memory.
type MemoryRunEventStore struct {
	mu     sync.RWMutex
	events map[string][]*RunEvent // runID -> list of events
}

func NewMemoryRunEventStore() *MemoryRunEventStore {
	return &MemoryRunEventStore{
		events: make(map[string][]*RunEvent),
	}
}

func (s *MemoryRunEventStore) AppendEvent(ctx context.Context, runID, eventType string, data json.RawMessage) (*RunEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := s.events[runID]
	seq := int64(len(current) + 1)
	eventID := fmt.Sprintf("%s:%d", runID, seq)

	evt := &RunEvent{
		ID:        eventID,
		RunID:     runID,
		Sequence:  seq,
		EventType: eventType,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	s.events[runID] = append(current, evt)
	return evt, nil
}

func (s *MemoryRunEventStore) GetEventsAfter(ctx context.Context, runID, lastEventID string) ([]*RunEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := s.events[runID]
	if lastEventID == "" {
		out := make([]*RunEvent, len(all))
		copy(out, all)
		return out, nil
	}

	// Parse sequence from lastEventID (format: "runID:seq" or pure integer)
	var afterSeq int64 = 0
	if parts := strings.Split(lastEventID, ":"); len(parts) == 2 {
		if val, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
			afterSeq = val
		}
	} else if val, err := strconv.ParseInt(lastEventID, 10, 64); err == nil {
		afterSeq = val
	}

	var result []*RunEvent
	for _, evt := range all {
		if evt.Sequence > afterSeq {
			result = append(result, evt)
		}
	}
	return result, nil
}

func (s *MemoryRunEventStore) GetAllEvents(ctx context.Context, runID string) ([]*RunEvent, error) {
	return s.GetEventsAfter(ctx, runID, "")
}

// SSEBroadcaster coordinates live event distribution and catch-up on reconnection.
type SSEBroadcaster struct {
	store       RunEventStore
	mu          sync.RWMutex
	subscribers map[string]map[chan *RunEvent]struct{} // runID -> set of subscriber channels
}

func NewSSEBroadcaster(store RunEventStore) *SSEBroadcaster {
	return &SSEBroadcaster{
		store:       store,
		subscribers: make(map[string]map[chan *RunEvent]struct{}),
	}
}

// Subscribe returns a channel of events and replayed history since lastEventID.
func (b *SSEBroadcaster) Subscribe(ctx context.Context, runID, lastEventID string, bufferSize int) (<-chan *RunEvent, []*RunEvent, func(), error) {
	b.mu.Lock()
	if b.subscribers[runID] == nil {
		b.subscribers[runID] = make(map[chan *RunEvent]struct{})
	}
	ch := make(chan *RunEvent, bufferSize)
	b.subscribers[runID][ch] = struct{}{}
	b.mu.Unlock()

	// Get history for catch-up (REQ-019 reconnection guarantee)
	history, err := b.store.GetEventsAfter(ctx, runID, lastEventID)
	if err != nil {
		history = nil
	}

	cleanup := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if subs, ok := b.subscribers[runID]; ok {
			delete(subs, ch)
			close(ch)
			if len(subs) == 0 {
				delete(b.subscribers, runID)
			}
		}
	}

	return ch, history, cleanup, nil
}

// Broadcast records an event and delivers it to all current subscribers.
func (b *SSEBroadcaster) Broadcast(ctx context.Context, runID, eventType string, data json.RawMessage) (*RunEvent, error) {
	evt, err := b.store.AppendEvent(ctx, runID, eventType, data)
	if err != nil {
		return nil, err
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, ok := b.subscribers[runID]; ok {
		for ch := range subs {
			select {
			case ch <- evt:
			default:
				// If client buffer is full, do not block other subscribers
			}
		}
	}
	return evt, nil
}

// FormatSSE formats a RunEvent according to the W3C Server-Sent Events wire standard.
func FormatSSE(evt *RunEvent) []byte {
	return []byte(fmt.Sprintf("id: %s\nevent: %s\ndata: %s\n\n", evt.ID, evt.EventType, string(evt.Data)))
}

// FormatKeepalive returns a standard SSE comment keepalive line.
func FormatKeepalive() []byte {
	return []byte(":keepalive\n\n")
}
