package workflow

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestScheduler_DueJobLeaseAndHeartbeat(t *testing.T) {
	ctx := context.Background()
	repo := NewMemorySchedulerRepository()
	service := NewSchedulerService(repo)

	// 1. Register a schedule due immediately
	now := time.Now().UTC()
	job, err := service.RegisterSchedule(ctx, &ScheduleJob{
		WorkspaceID: "ws-1",
		TaskType:    "social.publish",
		Payload:     json.RawMessage(`{"post_id":"post-123"}`),
		Interval:    10 * time.Minute,
		NextRunAt:   now.Add(-1 * time.Minute), // already due
	})
	if err != nil {
		t.Fatalf("RegisterSchedule failed: %v", err)
	}

	// 2. Worker 1 acquires due job with 5-second lease
	leased, err := service.AcquireDueJobs(ctx, "worker-1", now, 5*time.Second, 10)
	if err != nil {
		t.Fatalf("AcquireDueJobs failed: %v", err)
	}
	if len(leased) != 1 || leased[0].ID != job.ID {
		t.Fatalf("Expected 1 leased job for worker-1, got %d", len(leased))
	}
	if leased[0].LeaseHolder != "worker-1" {
		t.Fatalf("Expected LeaseHolder to be worker-1, got %s", leased[0].LeaseHolder)
	}

	// 3. Worker 2 attempts to acquire the same job while lease is valid -> should fail / get 0 jobs
	worker2Leased, err := service.AcquireDueJobs(ctx, "worker-2", now, 5*time.Second, 10)
	if err != nil {
		t.Fatalf("Worker 2 AcquireDueJobs error: %v", err)
	}
	if len(worker2Leased) != 0 {
		t.Fatalf("Expected worker 2 to acquire 0 jobs while lease active, got %d", len(worker2Leased))
	}

	// 4. Worker 1 renews heartbeat
	err = service.HeartbeatLease(ctx, job.ID, "worker-1", 10*time.Second)
	if err != nil {
		t.Fatalf("HeartbeatLease failed: %v", err)
	}

	// Worker 2 tries heartbeat on someone else's lease -> must fail
	err = service.HeartbeatLease(ctx, job.ID, "worker-2", 10*time.Second)
	if err == nil {
		t.Fatalf("Expected error for non-holder heartbeat, got nil")
	}

	// 5. Worker 1 completes job and releases lease with 10m interval
	err = service.ReleaseLease(ctx, job.ID, "worker-1", 10*time.Minute)
	if err != nil {
		t.Fatalf("ReleaseLease failed: %v", err)
	}

	// 6. Verify job state after release
	updatedJob, err := service.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if updatedJob.LeaseHolder != "" {
		t.Fatalf("Expected empty LeaseHolder after release, got %s", updatedJob.LeaseHolder)
	}
	if updatedJob.LastRunAt == nil {
		t.Fatalf("Expected LastRunAt to be set after release")
	}
}

func TestWorkflow_PauseResumeCancel_AT018_AT021(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRunLedgerRepository()
	service := NewRunLedgerService(repo)

	// 1. Create a new run
	run, isNew, err := service.RegisterOrGetRun(ctx, &ActionRun{
		WorkspaceID:    "ws-test",
		ActorID:        "user-test",
		TaskType:       "career.apply",
		InputPayload:   json.RawMessage(`{"jobId":"job-999"}`),
		IdempotencyKey: "idemp-lifecycle-1",
	})
	if err != nil || !isNew {
		t.Fatalf("RegisterOrGetRun failed: isNew=%v, err=%v", isNew, err)
	}

	// 2. Pause the run
	pausedRun, err := service.PauseRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("PauseRun failed: %v", err)
	}
	if pausedRun.State != StatePaused {
		t.Fatalf("Expected state to be paused, got %s", pausedRun.State)
	}

	// 3. Starting an attempt on a paused run must be blocked (REQ-019, AT-018)
	_, err = service.StartAttempt(ctx, run.ID, "worker-1")
	if err != ErrRunPaused {
		t.Fatalf("Expected ErrRunPaused, got %v", err)
	}

	// 4. Resume the run
	resumedRun, err := service.ResumeRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("ResumeRun failed: %v", err)
	}
	if resumedRun.State != StateQueued {
		t.Fatalf("Expected state to be queued, got %s", resumedRun.State)
	}

	// 5. Worker starts attempt after resume
	attemptNum, err := service.StartAttempt(ctx, run.ID, "worker-1")
	if err != nil {
		t.Fatalf("StartAttempt after resume failed: %v", err)
	}
	if attemptNum != 1 {
		t.Fatalf("Expected attempt 1, got %d", attemptNum)
	}

	// 6. User cancels run while worker is executing
	cancelledRun, err := service.CancelRun(ctx, run.ID, "user aborted application")
	if err != nil {
		t.Fatalf("CancelRun failed: %v", err)
	}
	if cancelledRun.State != StateCancelled {
		t.Fatalf("Expected state cancelled, got %s", cancelledRun.State)
	}

	// 7. Late worker completion must be fenced off and rejected (AT-021)
	err = service.CompleteRun(ctx, run.ID, attemptNum, json.RawMessage(`{"submitted":true}`))
	if err != ErrRunCancelled {
		t.Fatalf("Expected ErrRunCancelled for late completion on cancelled run, got %v", err)
	}

	// 8. Late worker failure must also be rejected
	err = service.FailRun(ctx, run.ID, attemptNum, json.RawMessage(`{"error":"failed"}`))
	if err != ErrRunCancelled {
		t.Fatalf("Expected ErrRunCancelled for late fail on cancelled run, got %v", err)
	}
}

func TestSSE_ReconnectionAndHistoryReplay(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryRunEventStore()
	broadcaster := NewSSEBroadcaster(store)

	runID := "run-sse-1"

	// 1. Publish 3 initial events
	_, err := broadcaster.Broadcast(ctx, runID, "status", json.RawMessage(`{"state":"queued"}`))
	if err != nil {
		t.Fatalf("Broadcast event 1 failed: %v", err)
	}
	_, err = broadcaster.Broadcast(ctx, runID, "attempt_started", json.RawMessage(`{"attempt":1}`))
	if err != nil {
		t.Fatalf("Broadcast event 2 failed: %v", err)
	}
	evt3, err := broadcaster.Broadcast(ctx, runID, "log", json.RawMessage(`{"msg":"processing PDF"}`))
	if err != nil {
		t.Fatalf("Broadcast event 3 failed: %v", err)
	}

	// 2. Client connects with Last-Event-ID = evt2 (or sequence 2)
	// Last-Event-ID is evt3.ID = "run-sse-1:3"
	// If reconnecting with Last-Event-ID = "run-sse-1:1", should receive events 2 and 3 in history.
	ch, history, cleanup, err := broadcaster.Subscribe(ctx, runID, "run-sse-1:1", 10)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer cleanup()

	if len(history) != 2 {
		t.Fatalf("Expected 2 replayed history events, got %d", len(history))
	}
	if history[0].EventType != "attempt_started" || history[1].EventType != "log" {
		t.Fatalf("Unexpected history events: %v", history)
	}

	// 3. Publish live event while subscribed
	go func() {
		time.Sleep(50 * time.Millisecond)
		_, _ = broadcaster.Broadcast(ctx, runID, "completed", json.RawMessage(`{"result":"ok"}`))
	}()

	select {
	case liveEvt := <-ch:
		if liveEvt.EventType != "completed" {
			t.Fatalf("Expected live event 'completed', got %s", liveEvt.EventType)
		}
		// Verify W3C wire format
		formatted := FormatSSE(liveEvt)
		if len(formatted) == 0 {
			t.Fatalf("Formatted SSE message is empty")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Timed out waiting for live broadcast event")
	}

	// 4. Verify keepalive formatting
	keepalive := FormatKeepalive()
	if string(keepalive) != ":keepalive\n\n" {
		t.Fatalf("Unexpected keepalive format: %q", string(keepalive))
	}

	_ = evt3
}
