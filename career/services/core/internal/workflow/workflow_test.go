package workflow

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestWorkflow_EnvelopeSerialization(t *testing.T) {
	rawPayload := json.RawMessage(`{"document_id":"doc-123","format":"pdf"}`)
	env, err := NewTaskEnvelope("career", "document.parse", "user-1", "ws-1", "idemp-abc", rawPayload)
	if err != nil {
		t.Fatalf("NewTaskEnvelope failed: %v", err)
	}

	if env.Version != "v1" || env.Section != "career" || env.TaskType != "document.parse" {
		t.Errorf("unexpected envelope fields: %+v", env)
	}

	data, err := env.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	parsed, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if parsed.ID != env.ID || parsed.IdempotencyKey != "idemp-abc" {
		t.Errorf("parsed envelope mismatch: %+v", parsed)
	}
}

func TestWorkflow_IdempotencyAndDoubleSideEffectProtection_AT021(t *testing.T) {
	repo := NewMemoryRunLedgerRepository()
	service := NewRunLedgerService(repo)
	ctx := context.Background()

	run1 := &ActionRun{
		WorkspaceID:    "ws-100",
		ActorID:        "user-100",
		TaskType:       "career.apply",
		InputPayload:   json.RawMessage(`{"job_id":"job-xyz"}`),
		IdempotencyKey: "apply-req-unique-999",
	}

	created, isNew, err := service.RegisterOrGetRun(ctx, run1)
	if err != nil || !isNew {
		t.Fatalf("first registration should be new, err: %v", err)
	}

	// Re-attempt registration with same idempotency key (simulating network retry / Redis restart)
	run2 := &ActionRun{
		WorkspaceID:    "ws-100",
		ActorID:        "user-100",
		TaskType:       "career.apply",
		InputPayload:   json.RawMessage(`{"job_id":"job-xyz"}`),
		IdempotencyKey: "apply-req-unique-999",
	}

	existing, isNew2, err := service.RegisterOrGetRun(ctx, run2)
	if err != nil || isNew2 {
		t.Fatalf("second registration should return existing run without new insertion, err: %v", err)
	}
	if existing.ID != created.ID {
		t.Fatalf("expected same run ID %s, got %s", created.ID, existing.ID)
	}

	// Complete the run
	attemptToken, err := service.StartAttempt(ctx, created.ID, "worker-alpha")
	if err != nil {
		t.Fatalf("StartAttempt failed: %v", err)
	}

	result := json.RawMessage(`{"status":"applied","receipt":"rcpt-123"}`)
	err = service.CompleteRun(ctx, created.ID, attemptToken, result)
	if err != nil {
		t.Fatalf("CompleteRun failed: %v", err)
	}

	// Duplicate delivery of completion: must not produce double side effects or errors
	err = service.CompleteRun(ctx, created.ID, attemptToken, result)
	if err != nil {
		t.Fatalf("idempotent replay of completed run must succeed silently, got: %v", err)
	}

	// Attempting to start a new attempt on a completed run must fail
	_, err = service.StartAttempt(ctx, created.ID, "worker-beta")
	if err != ErrRunAlreadyCompleted {
		t.Fatalf("expected ErrRunAlreadyCompleted, got: %v", err)
	}
}

func TestWorkflow_FencingTokenRejection(t *testing.T) {
	repo := NewMemoryRunLedgerRepository()
	service := NewRunLedgerService(repo)
	ctx := context.Background()

	run := &ActionRun{
		WorkspaceID:  "ws-1",
		ActorID:      "user-1",
		TaskType:     "linkedin.post",
		InputPayload: json.RawMessage(`{"content":"Hello World"}`),
	}

	registered, _, err := service.RegisterOrGetRun(ctx, run)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Worker 1 takes lease (attempt 1)
	attempt1, err := service.StartAttempt(ctx, registered.ID, "worker-1")
	if err != nil || attempt1 != 1 {
		t.Fatalf("attempt 1 start failed: %v", err)
	}

	// Worker 1 times out / freezes; Worker 2 takes lease (attempt 2)
	attempt2, err := service.StartAttempt(ctx, registered.ID, "worker-2")
	if err != nil || attempt2 != 2 {
		t.Fatalf("attempt 2 start failed: %v", err)
	}

	// Worker 1 unfreezes and attempts to complete with stale attempt 1 fencing token
	staleResult := json.RawMessage(`{"post_id":"stale-1"}`)
	err = service.CompleteRun(ctx, registered.ID, attempt1, staleResult)
	if err != ErrStaleAttemptFenced {
		t.Fatalf("expected ErrStaleAttemptFenced for stale token, got: %v", err)
	}

	// Worker 2 completes with valid attempt 2 fencing token
	validResult := json.RawMessage(`{"post_id":"valid-2"}`)
	err = service.CompleteRun(ctx, registered.ID, attempt2, validResult)
	if err != nil {
		t.Fatalf("valid attempt completion failed: %v", err)
	}
}

func TestWorkflow_OutboxRelayBatch(t *testing.T) {
	outboxRepo := NewMemoryOutboxRepository()
	outboxService := NewOutboxService(outboxRepo)
	streamEngine := NewMemoryStreamEngine()
	relay := NewOutboxRelay(outboxRepo, streamEngine)
	ctx := context.Background()

	// Record outbox event atomically
	payload := json.RawMessage(`{"section":"career","job_id":"job-456","actor_id":"user-1","workspace_id":"ws-1"}`)
	evt, err := outboxService.RecordEvent(ctx, "application", "app-100", "application.submitted", payload)
	if err != nil {
		t.Fatalf("RecordEvent failed: %v", err)
	}

	if evt.Status != OutboxStatusPending {
		t.Errorf("initial event status should be pending, got: %s", evt.Status)
	}

	// Process batch via OutboxRelay
	dispatched, err := relay.ProcessBatch(ctx, 10)
	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}
	if dispatched != 1 {
		t.Fatalf("expected 1 event dispatched, got: %d", dispatched)
	}

	// Verify status updated to dispatched
	pending, err := outboxRepo.GetPendingEvents(ctx, 10)
	if err != nil || len(pending) != 0 {
		t.Fatalf("expected 0 pending events after relay, got: %d", len(pending))
	}
}

func TestWorkflow_StreamConsumerGroupAndReclaim_AT021(t *testing.T) {
	streamEngine := NewMemoryStreamEngine()
	ctx := context.Background()
	stream := "stream:career:job.match"
	group := "worker-group"

	err := streamEngine.CreateConsumerGroup(ctx, stream, group, "0")
	if err != nil {
		t.Fatalf("CreateConsumerGroup failed: %v", err)
	}

	rawPayload := json.RawMessage(`{"criteria":"remote-go"}`)
	env, _ := NewTaskEnvelope("career", "job.match", "user-1", "ws-1", "idemp-match-1", rawPayload)

	msgID, err := streamEngine.Publish(ctx, stream, env)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Consumer 1 reads the message but crashes before Ack
	msgs, err := streamEngine.ReadGroup(ctx, stream, group, "consumer-1", 1, 0)
	if err != nil || len(msgs) != 1 {
		t.Fatalf("ReadGroup failed for consumer-1: %v", err)
	}
	if msgs[0].ID != msgID {
		t.Fatalf("message ID mismatch: expected %s, got %s", msgID, msgs[0].ID)
	}

	// Stale check with 0 idle duration simulates lease expiration/crash reclaim
	claimed, err := streamEngine.ClaimStale(ctx, stream, group, "consumer-2", 0*time.Second, 1)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("ClaimStale failed to recover unacked message: %v", err)
	}
	if claimed[0].ID != msgID {
		t.Fatalf("claimed message ID mismatch: expected %s, got %s", msgID, claimed[0].ID)
	}

	// Consumer 2 completes and ACKs
	err = streamEngine.Ack(ctx, stream, group, claimed[0].ID)
	if err != nil {
		t.Fatalf("Ack failed: %v", err)
	}

	// Further claims yield no pending messages
	reclaimed, err := streamEngine.ClaimStale(ctx, stream, group, "consumer-3", 0*time.Second, 1)
	if err != nil || len(reclaimed) != 0 {
		t.Fatalf("expected 0 pending messages after Ack, got: %d", len(reclaimed))
	}
}
