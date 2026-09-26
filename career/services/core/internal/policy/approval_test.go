package policy

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestApproval_HashIntegrityAndTamperRejection_AT007(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryApprovalRepository()
	service := NewApprovalService(repo)

	originalPayload := json.RawMessage(`{"job_id":"job-456","answers":{"authorized_to_work":true,"years_experience":5}}`)
	resumeBytes := []byte("%PDF-1.4 mock resume content")
	docHash := ComputeDocumentHash(resumeBytes)

	// 1. Create approval request
	req, err := service.CreateApprovalRequest(ctx, RequestApprovalInput{
		WorkspaceID:     "ws-alpha",
		ActorID:         "user-alice",
		ActionType:      "career.apply",
		TargetRecipient: "https://careers.example.com/job/456",
		Payload:         originalPayload,
		DocumentHash:    docHash,
		TTL:             2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("CreateApprovalRequest failed: %v", err)
	}

	// 2. Gate check before approval must fail (StatusPending)
	err = service.ValidateAndConsumeGate(
		ctx,
		req.ID,
		"ws-alpha",
		"user-alice",
		"career.apply",
		"https://careers.example.com/job/456",
		originalPayload,
		docHash,
	)
	if err == nil || !strings.Contains(err.Error(), "current status is pending") {
		t.Fatalf("Expected error for unapproved request, got: %v", err)
	}

	// 3. User reviews and approves the request
	approved, err := service.Approve(ctx, req.ID, "user-alice")
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if approved.Status != StatusApproved {
		t.Fatalf("Expected status approved, got %s", approved.Status)
	}

	// 4. Case A: Tampered Payload (e.g. user or bot modified answers behind the back)
	tamperedPayload := json.RawMessage(`{"job_id":"job-456","answers":{"authorized_to_work":true,"years_experience":8}}`)
	err = service.ValidateAndConsumeGate(
		ctx,
		req.ID,
		"ws-alpha",
		"user-alice",
		"career.apply",
		"https://careers.example.com/job/456",
		tamperedPayload,
		docHash,
	)
	if err == nil || !strings.Contains(err.Error(), "payload hash changed since approval") {
		t.Fatalf("Expected payload hash mismatch error, got: %v", err)
	}

	// 5. Case B: Tampered Document (e.g. attached another resume after approval)
	tamperedDocHash := ComputeDocumentHash([]byte("%PDF-1.4 modified resume content"))
	err = service.ValidateAndConsumeGate(
		ctx,
		req.ID,
		"ws-alpha",
		"user-alice",
		"career.apply",
		"https://careers.example.com/job/456",
		originalPayload,
		tamperedDocHash,
	)
	if err == nil || !strings.Contains(err.Error(), "document hash changed since approval") {
		t.Fatalf("Expected document hash mismatch error, got: %v", err)
	}

	// 6. Case C: Target recipient changed (e.g. redirecting application)
	err = service.ValidateAndConsumeGate(
		ctx,
		req.ID,
		"ws-alpha",
		"user-alice",
		"career.apply",
		"https://malicious.example.com/target",
		originalPayload,
		docHash,
	)
	if err == nil || !strings.Contains(err.Error(), "does not match approval") {
		t.Fatalf("Expected recipient mismatch error, got: %v", err)
	}

	// 7. Case D: Untampered, exact original payload and document -> Success!
	err = service.ValidateAndConsumeGate(
		ctx,
		req.ID,
		"ws-alpha",
		"user-alice",
		"career.apply",
		"https://careers.example.com/job/456",
		originalPayload,
		docHash,
	)
	if err != nil {
		t.Fatalf("Expected valid gate execution, got: %v", err)
	}

	// Verify status is now consumed
	consumed, err := service.GetApproval(ctx, req.ID)
	if err != nil {
		t.Fatalf("GetApproval failed: %v", err)
	}
	if consumed.Status != StatusConsumed || consumed.ConsumedAt == nil {
		t.Fatalf("Expected status consumed with ConsumedAt set, got %s", consumed.Status)
	}
}

func TestApproval_SingleUseReplayProtection_AT007(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryApprovalRepository()
	service := NewApprovalService(repo)

	payload := json.RawMessage(`{"post_text":"Hello LinkedIn!"}`)
	req, err := service.CreateApprovalRequest(ctx, RequestApprovalInput{
		WorkspaceID: "ws-1",
		ActorID:     "user-1",
		ActionType:  "linkedin.post.create",
		Payload:     payload,
		TTL:         1 * time.Hour,
	})
	if err != nil {
		t.Fatalf("CreateApprovalRequest failed: %v", err)
	}

	_, err = service.Approve(ctx, req.ID, "user-1")
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}

	// First execution -> success
	err = service.ValidateAndConsumeGate(ctx, req.ID, "ws-1", "user-1", "linkedin.post.create", "", payload, "")
	if err != nil {
		t.Fatalf("First gate consumption failed: %v", err)
	}

	// Second execution (Replay attack / double side effect attempt) -> strictly blocked (AT-007)
	err = service.ValidateAndConsumeGate(ctx, req.ID, "ws-1", "user-1", "linkedin.post.create", "", payload, "")
	if err != ErrApprovalAlreadyConsumed {
		t.Fatalf("Expected ErrApprovalAlreadyConsumed on replay, got: %v", err)
	}
}

func TestApproval_ExpirationAndRejection(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryApprovalRepository()
	service := NewApprovalService(repo)

	payload := json.RawMessage(`{"recipient":"lead-1"}`)

	// 1. Expired approval
	reqExpired, _ := service.CreateApprovalRequest(ctx, RequestApprovalInput{
		WorkspaceID: "ws-1",
		ActorID:     "user-1",
		ActionType:  "linkedin.connect",
		Payload:     payload,
		TTL:         -1 * time.Second, // already expired
	})

	_, err := service.Approve(ctx, reqExpired.ID, "user-1")
	if err != ErrApprovalExpired {
		t.Fatalf("Expected ErrApprovalExpired on approve, got: %v", err)
	}

	// 2. Rejected approval
	reqReject, _ := service.CreateApprovalRequest(ctx, RequestApprovalInput{
		WorkspaceID: "ws-1",
		ActorID:     "user-1",
		ActionType:  "linkedin.connect",
		Payload:     payload,
		TTL:         1 * time.Hour,
	})

	rejected, err := service.Reject(ctx, reqReject.ID, "reviewer-1", "content inappropriate")
	if err != nil {
		t.Fatalf("Reject failed: %v", err)
	}
	if rejected.Status != StatusRejected || rejected.RejectionReason != "content inappropriate" {
		t.Fatalf("Unexpected reject state: %v", rejected)
	}

	err = service.ValidateAndConsumeGate(ctx, reqReject.ID, "ws-1", "user-1", "linkedin.connect", "", payload, "")
	if err != ErrApprovalRejected {
		t.Fatalf("Expected ErrApprovalRejected at gate, got: %v", err)
	}
}
