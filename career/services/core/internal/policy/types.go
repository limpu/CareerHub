package policy

import (
	"errors"
	"time"
)

var (
	ErrApprovalRequired         = errors.New("action execution blocked: explicit human approval required")
	ErrApprovalNotFound         = errors.New("approval record not found")
	ErrApprovalExpired          = errors.New("approval record has expired")
	ErrApprovalPayloadMismatch  = errors.New("execution payload or document hash does not match immutable approval binding")
	ErrApprovalAlreadyConsumed  = errors.New("approval record has already been consumed (replay prohibited)")
	ErrApprovalRejected         = errors.New("approval request was rejected by reviewer")
	ErrUnauthorizedApprover     = errors.New("unauthorized reviewer for approval")
	ErrInvalidStatusTransition  = errors.New("invalid approval status transition")
)

// ApprovalStatus represents the state of an outbound execution gate.
type ApprovalStatus string

const (
	StatusPending  ApprovalStatus = "pending"
	StatusApproved ApprovalStatus = "approved"
	StatusRejected ApprovalStatus = "rejected"
	StatusConsumed ApprovalStatus = "consumed"
	StatusExpired  ApprovalStatus = "expired"
)

// ApprovalRequest represents an immutable binding of execution parameters to a human approval (REQ-015, AT-007).
type ApprovalRequest struct {
	ID              string         `json:"id"`
	WorkspaceID     string         `json:"workspace_id"`
	ActorID         string         `json:"actor_id"`
	ActionType      string         `json:"action_type"`
	TargetRecipient string         `json:"target_recipient"` // Target URL, email, or profile handle
	PayloadHash     string         `json:"payload_hash"`     // SHA-256 of canonical action parameters
	DocumentHash    string         `json:"document_hash"`    // SHA-256 of attached document blob (e.g. resume)
	PolicyVersion   string         `json:"policy_version"`
	Status          ApprovalStatus `json:"status"`
	ApproverID      string         `json:"approver_id,omitempty"`
	RejectionReason string         `json:"rejection_reason,omitempty"`
	ExpiresAt       time.Time      `json:"expires_at"`
	ApprovedAt      *time.Time     `json:"approved_at,omitempty"`
	ConsumedAt      *time.Time     `json:"consumed_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
