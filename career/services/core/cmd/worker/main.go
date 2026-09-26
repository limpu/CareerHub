package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/social-platform/services/core/internal/workflow"
)

func main() {
	log.Println("[core-worker] Starting Go durable background worker...")

	streamEngine := workflow.NewMemoryStreamEngine()
	runLedgerRepo := workflow.NewMemoryRunLedgerRepository()
	runLedgerService := workflow.NewRunLedgerService(runLedgerRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Consumer group for durable task execution
	consumerGroup := "core-workers"
	consumerName := "worker-instance-1"
	streamName := "stream:career:task.execute"

	_ = streamEngine.CreateConsumerGroup(ctx, streamName, consumerGroup, "0")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Worker loop polling stream
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				messages, err := streamEngine.ReadGroup(ctx, streamName, consumerGroup, consumerName, 5, 2*time.Second)
				if err != nil {
					time.Sleep(500 * time.Millisecond)
					continue
				}

				for _, msg := range messages {
					log.Printf("[core-worker] Received message %s on stream %s", msg.ID, streamName)

					env := msg.Envelope
					if env != nil && env.IdempotencyKey != "" {
						runID := env.IdempotencyKey
						// Attempt lease and execution
						attemptNumber, err := runLedgerService.StartAttempt(ctx, runID, consumerName)
						if err != nil {
							log.Printf("[core-worker] StartAttempt failed for run %s: %v", runID, err)
						} else {
							log.Printf("[core-worker] Acquired attempt %d for run %s", attemptNumber, runID)
							// Simulate processing and complete
							resPayload := json.RawMessage(`{"status":"success"}`)
							_ = runLedgerService.CompleteRun(ctx, runID, attemptNumber, resPayload)
						}
					}

					// Durable ACK
					_ = streamEngine.Ack(ctx, streamName, consumerGroup, msg.ID)
				}
			}
		}
	}()

	schedulerRepo := workflow.NewMemorySchedulerRepository()
	schedulerService := workflow.NewSchedulerService(schedulerRepo)

	// Scheduler loop polling for due jobs (FND-008)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				now := time.Now().UTC()
				dueJobs, err := schedulerService.AcquireDueJobs(ctx, consumerName, now, 30*time.Second, 5)
				if err == nil && len(dueJobs) > 0 {
					for _, job := range dueJobs {
						log.Printf("[core-worker] Acquired lease on scheduled job %s (%s)", job.ID, job.TaskType)
						// Heartbeat demonstration
						_ = schedulerService.HeartbeatLease(ctx, job.ID, consumerName, 30*time.Second)
						// Release and schedule next interval
						interval := job.Interval
						if interval <= 0 {
							interval = 1 * time.Hour
						}
						_ = schedulerService.ReleaseLease(ctx, job.ID, consumerName, interval)
					}
				}
			}
		}
	}()

	log.Println("[core-worker] Durable worker active with stream consumer and scheduler...")
	<-stopChan
	log.Println("[core-worker] Shutting down durable worker cleanly...")
}
