package fakes

import (
	"errors"
	"time"
)

// ErrorMode defines simulated failure scenarios for deterministic testing.
type ErrorMode string

const (
	ErrorModeNone                ErrorMode = "none"
	ErrorModeTimeout             ErrorMode = "timeout"
	ErrorModeRateLimit429        ErrorMode = "rate_limit_429"
	ErrorModeServer500           ErrorMode = "server_500"
	ErrorModeServer503           ErrorMode = "server_503"
	ErrorModeCrashAfterDispatch  ErrorMode = "crash_after_dispatch"
	ErrorModeAmbiguousDrop       ErrorMode = "ambiguous_network_drop"
)

var (
	ErrSimulatedTimeout            = errors.New("fakes: simulated network timeout")
	ErrSimulatedRateLimit          = errors.New("fakes: HTTP 429 Too Many Requests (Retry-After: 60s)")
	ErrSimulatedServer500          = errors.New("fakes: HTTP 500 Internal Server Error")
	ErrSimulatedServer503          = errors.New("fakes: HTTP 503 Service Unavailable")
	ErrSimulatedCrashAfterDispatch = errors.New("fakes: worker process crashed immediately after remote dispatch")
	ErrApprovalMissingOrTampered   = errors.New("fakes: side-effect rejected: approval missing, expired, or payload hash mismatch")
	ErrAmbiguousSubmission         = errors.New("fakes: remote outcome unknown; entered needs_confirmation state")
	ErrDuplicateSubmissionBlocked  = errors.New("fakes: safety invariant: unreviewed duplicate submission blocked after ambiguous dispatch")
)

// SubmissionReceipt represents an authenticatable external receipt (REQ-005, AT-005).
type SubmissionReceipt struct {
	ReceiptID        string    `json:"receipt_id"`
	Provider         string    `json:"provider"`
	ExternalTargetID string    `json:"external_target_id"`
	DispatchedAt     time.Time `json:"dispatched_at"`
	PayloadSHA256    string    `json:"payload_sha256"`
	ConfirmationCode string    `json:"confirmation_code"`
	VerifiedStatus   string    `json:"verified_status"` // "verified_applied", "needs_confirmation", "pending_review"
}

// ApplicationSubmission represents a candidate job application payload.
type ApplicationSubmission struct {
	JobID             string            `json:"job_id"`
	CandidateID       string            `json:"candidate_id"`
	PayloadHash       string            `json:"payload_hash"`
	ApprovalID        string            `json:"approval_id"`
	SubmittedAnswers  map[string]string `json:"submitted_answers"`
	HasUserConsent    bool              `json:"has_user_consent"`
}

// PublishingChannel represents a targeted social destination.
type PublishingChannel struct {
	ChannelID string `json:"channel_id"`
	Platform  string `json:"platform"` // "linkedin", "x", "threads"
}

// SocialPostSubmission represents content to be published to one or more channels.
type SocialPostSubmission struct {
	PostID       string              `json:"post_id"`
	WorkspaceID  string              `json:"workspace_id"`
	Channels     []PublishingChannel `json:"channels"`
	ContentText  string              `json:"content_text"`
	PayloadHash  string              `json:"payload_hash"`
	ApprovalID   string              `json:"approval_id"`
	IsApproved   bool                `json:"is_approved"`
}

// ChannelPublishResult represents the isolated dispatch result per channel (AT-015).
type ChannelPublishResult struct {
	ChannelID    string    `json:"channel_id"`
	Platform     string    `json:"platform"`
	Success      bool      `json:"success"`
	ExternalID   string    `json:"external_id,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	DispatchedAt time.Time `json:"dispatched_at"`
}

// SocialPublishReport aggregates channel execution results (AT-015).
type SocialPublishReport struct {
	PostID        string                 `json:"post_id"`
	OverallStatus string                 `json:"overall_status"` // "published", "partially_failed", "failed"
	Results       []ChannelPublishResult `json:"results"`
}
