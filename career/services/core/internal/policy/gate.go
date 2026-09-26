package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ApprovalRepository defines persistent storage for human review approvals.
type ApprovalRepository interface {
	Save(ctx context.Context, req *ApprovalRequest) error
	GetByID(ctx context.Context, id string) (*ApprovalRequest, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*ApprovalRequest, error)
	Update(ctx context.Context, req *ApprovalRequest) error
}

// MemoryApprovalRepository provides thread-safe in-memory storage for approvals.
type MemoryApprovalRepository struct {
	mu        sync.RWMutex
	approvals map[string]*ApprovalRequest
}

func NewMemoryApprovalRepository() *MemoryApprovalRepository {
	return &MemoryApprovalRepository{
		approvals: make(map[string]*ApprovalRequest),
	}
}

func (r *MemoryApprovalRepository) Save(ctx context.Context, req *ApprovalRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clone := *req
	r.approvals[req.ID] = &clone
	return nil
}

func (r *MemoryApprovalRepository) GetByID(ctx context.Context, id string) (*ApprovalRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	req, ok := r.approvals[id]
	if !ok {
		return nil, ErrApprovalNotFound
	}
	clone := *req
	return &clone, nil
}

func (r *MemoryApprovalRepository) ListByWorkspace(ctx context.Context, workspaceID string) ([]*ApprovalRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*ApprovalRequest
	for _, req := range r.approvals {
		if req.WorkspaceID == workspaceID {
			clone := *req
			list = append(list, &clone)
		}
	}
	return list, nil
}

func (r *MemoryApprovalRepository) Update(ctx context.Context, req *ApprovalRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.approvals[req.ID]; !ok {
		return ErrApprovalNotFound
	}
	req.UpdatedAt = time.Now().UTC()
	clone := *req
	r.approvals[req.ID] = &clone
	return nil
}

// RequestApprovalInput parameters for binding an outbound action to an approval record.
type RequestApprovalInput struct {
	WorkspaceID     string
	ActorID         string
	ActionType      string
	TargetRecipient string
	Payload         json.RawMessage
	DocumentBytes   []byte
	DocumentHash    string
	PolicyVersion   string
	TTL             time.Duration
}

// ApprovalService manages the outbound execution gate and immutable hash validation (REQ-015, AT-007).
type ApprovalService struct {
	repo ApprovalRepository
}

func NewApprovalService(repo ApprovalRepository) *ApprovalService {
	return &ApprovalService{repo: repo}
}

// CreateApprovalRequest creates a pending approval record bound to the exact payload hash.
func (s *ApprovalService) CreateApprovalRequest(ctx context.Context, in RequestApprovalInput) (*ApprovalRequest, error) {
	if in.WorkspaceID == "" || in.ActorID == "" || in.ActionType == "" {
		return nil, fmt.Errorf("workspaceID, actorID, and actionType are required")
	}

	payloadHash, err := ComputePayloadHash(in.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to hash payload: %w", err)
	}

	docHash := in.DocumentHash
	if docHash == "" && len(in.DocumentBytes) > 0 {
		docHash = ComputeDocumentHash(in.DocumentBytes)
	}

	ttl := in.TTL
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	policyVer := in.PolicyVersion
	if policyVer == "" {
		policyVer = "1.0.0"
	}

	now := time.Now().UTC()
	req := &ApprovalRequest{
		ID:              uuid.New().String(),
		WorkspaceID:     in.WorkspaceID,
		ActorID:         in.ActorID,
		ActionType:      in.ActionType,
		TargetRecipient: strings.TrimSpace(in.TargetRecipient),
		PayloadHash:     payloadHash,
		DocumentHash:    docHash,
		PolicyVersion:   policyVer,
		Status:          StatusPending,
		ExpiresAt:       now.Add(ttl),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.Save(ctx, req); err != nil {
		return nil, err
	}

	return req, nil
}

// Approve transitions a pending approval to approved.
func (s *ApprovalService) Approve(ctx context.Context, approvalID, approverID string) (*ApprovalRequest, error) {
	req, err := s.repo.GetByID(ctx, approvalID)
	if err != nil {
		return nil, err
	}

	if req.Status != StatusPending {
		return nil, fmt.Errorf("%w: cannot approve record in status %s", ErrInvalidStatusTransition, req.Status)
	}

	now := time.Now().UTC()
	if req.ExpiresAt.Before(now) {
		req.Status = StatusExpired
		_ = s.repo.Update(ctx, req)
		return nil, ErrApprovalExpired
	}

	req.Status = StatusApproved
	req.ApproverID = approverID
	req.ApprovedAt = &now

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return req, nil
}

// Reject transitions a pending approval to rejected.
func (s *ApprovalService) Reject(ctx context.Context, approvalID, approverID, reason string) (*ApprovalRequest, error) {
	req, err := s.repo.GetByID(ctx, approvalID)
	if err != nil {
		return nil, err
	}

	if req.Status != StatusPending {
		return nil, fmt.Errorf("%w: cannot reject record in status %s", ErrInvalidStatusTransition, req.Status)
	}

	now := time.Now().UTC()
	req.Status = StatusRejected
	req.ApproverID = approverID
	req.RejectionReason = reason
	req.UpdatedAt = now

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return req, nil
}

// GetApproval retrieves an approval by ID.
func (s *ApprovalService) GetApproval(ctx context.Context, id string) (*ApprovalRequest, error) {
	return s.repo.GetByID(ctx, id)
}

// ListApprovals retrieves all approval records for a workspace.
func (s *ApprovalService) ListApprovals(ctx context.Context, workspaceID string) ([]*ApprovalRequest, error) {
	return s.repo.ListByWorkspace(ctx, workspaceID)
}

// ValidateAndConsumeGate executes the immutable verification and single-use consumption (REQ-015, AT-007).
func (s *ApprovalService) ValidateAndConsumeGate(
	ctx context.Context,
	approvalID, workspaceID, actorID, actionType, recipient string,
	payload json.RawMessage,
	docHash string,
) error {
	if approvalID == "" {
		return ErrApprovalRequired
	}

	req, err := s.repo.GetByID(ctx, approvalID)
	if err != nil {
		return err
	}

	// 1. Context matching
	if req.WorkspaceID != workspaceID || req.ActionType != actionType {
		return fmt.Errorf("%w: workspace or action mismatch", ErrApprovalPayloadMismatch)
	}
	if req.TargetRecipient != "" && req.TargetRecipient != strings.TrimSpace(recipient) {
		return fmt.Errorf("%w: target recipient '%s' does not match approval '%s'", ErrApprovalPayloadMismatch, recipient, req.TargetRecipient)
	}

	// 2. Lifecycle status check
	if req.Status == StatusConsumed {
		return ErrApprovalAlreadyConsumed
	}
	if req.Status == StatusRejected {
		return ErrApprovalRejected
	}
	if req.Status != StatusApproved {
		return fmt.Errorf("%w: current status is %s", ErrApprovalRequired, req.Status)
	}

	// 3. Expiration check
	now := time.Now().UTC()
	if req.ExpiresAt.Before(now) {
		req.Status = StatusExpired
		_ = s.repo.Update(ctx, req)
		return ErrApprovalExpired
	}

	// 4. Payload Hash verification (AT-007: Tamper Invalidation)
	currentPayloadHash, err := ComputePayloadHash(payload)
	if err != nil {
		return fmt.Errorf("failed to hash current payload: %w", err)
	}
	if req.PayloadHash != currentPayloadHash {
		return fmt.Errorf("%w: payload hash changed since approval", ErrApprovalPayloadMismatch)
	}

	// 5. Document Hash verification (AT-007: Attached Document Invalidation)
	if req.DocumentHash != "" && req.DocumentHash != docHash {
		return fmt.Errorf("%w: document hash changed since approval", ErrApprovalPayloadMismatch)
	}

	// 6. Atomically transition to consumed
	req.Status = StatusConsumed
	req.ConsumedAt = &now
	return s.repo.Update(ctx, req)
}
