package quota

import (
	"errors"
	"time"
)

// ActionMetric represents the identifier of an action or resource subject to quotas.
type ActionMetric string

const (
	// LinkedIn action metrics
	MetricLinkedInConnectionRequest ActionMetric = "linkedin.connection_request"
	MetricLinkedInFollow            ActionMetric = "linkedin.follow"
	MetricLinkedInDMSend            ActionMetric = "linkedin.dm.send"
	MetricLinkedInDMRecruiter       ActionMetric = "linkedin.dm.recruiter" // Nested inside MetricLinkedInDMSend

	// Career action metrics
	MetricCareerJobApplyManual    ActionMetric = "career.job_apply.manual_prep"
	MetricCareerJobApplyAutomated ActionMetric = "career.job_apply.automated"

	// Social action metrics
	MetricSocialPublishPost  ActionMetric = "social.publish.post"
	MetricSocialFollow       ActionMetric = "social.follow"
	MetricSocialCommentDraft ActionMetric = "social.comment.draft"
	MetricSocialMessageOptIn ActionMetric = "social.message.opt_in"
	MetricSocialResearchRead ActionMetric = "social.research.read"
)

// AccountStatus represents the operational safety status of an external provider account.
type AccountStatus string

const (
	AccountStatusActive            AccountStatus = "active"
	AccountStatusPaused            AccountStatus = "paused"
	AccountStatusChallengeRequired AccountStatus = "challenge_required" // CAPTCHA / 2FA / checkpoint
	AccountStatusRateLimited       AccountStatus = "rate_limited"        // 429 received, bounded by RetryAfter
	AccountStatusRevoked           AccountStatus = "revoked"
)

// ReservationStatus represents the lifecycle of a quota reservation.
type ReservationStatus string

const (
	ReservationPending  ReservationStatus = "pending"
	ReservationSettled  ReservationStatus = "settled"
	ReservationReleased ReservationStatus = "released"
	ReservationRetained ReservationStatus = "retained" // Retained on ambiguous outcome/timeout
)

// Standard domain errors
var (
	ErrQuotaExceeded                 = errors.New("quota exceeded for action metric")
	ErrNestedQuotaExceeded           = errors.New("nested parent quota exceeded for action metric")
	ErrAccountPaused                 = errors.New("provider account is paused for safety or manual review")
	ErrChallengeRequired             = errors.New("provider account requires challenge resolution (CAPTCHA/2FA)")
	ErrRateLimited                   = errors.New("provider account is rate limited (429); retry after backoff")
	ErrActionCapabilityUnauthorized  = errors.New("action capability is not authorized or approved for this provider")
	ErrReservationNotFound           = errors.New("reservation not found")
	ErrReservationAlreadyFinalized   = errors.New("reservation already settled, released or retained")
	ErrStoreUnavailable             = errors.New("quota store is unavailable (fail-closed)")
	ErrAmbiguousWriteRetained        = errors.New("reservation retained due to ambiguous execution outcome")
)

// AccountState encapsulates the operational circuit-breaking and rate-limiting state of an external account.
type AccountState struct {
	AccountID       string        `json:"account_id"` // Unified external account UID (e.g., "li:user_123")
	Provider        string        `json:"provider"`
	Status          AccountStatus `json:"status"`
	ChallengeReason string        `json:"challenge_reason,omitempty"`
	RetryAfter      *time.Time    `json:"retry_after,omitempty"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// QuotaDefinition defines the limits and rolling window for an action metric.
type QuotaDefinition struct {
	Metric                      ActionMetric  `json:"metric"`
	ParentMetric                ActionMetric  `json:"parent_metric,omitempty"` // If set, reserving this metric also consumes parent
	Limit                       int           `json:"limit"`
	Window                      time.Duration `json:"window"` // Rolling window duration, e.g. 24h
	RequiresApprovedCapability  bool          `json:"requires_approved_capability"`
	AutoExecutionDefault        int           `json:"auto_execution_default"` // Default for unapproved automation (0)
}

// Reservation represents an atomic slot lock prior to external dispatch.
type Reservation struct {
	ID           string            `json:"id"`
	AccountID    string            `json:"account_id"`
	WorkspaceID  string            `json:"workspace_id"`
	Module       string            `json:"module"` // career, linkedin, social
	Metric       ActionMetric      `json:"metric"`
	ParentMetric ActionMetric      `json:"parent_metric,omitempty"`
	Amount       int               `json:"amount"`
	Status       ReservationStatus `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	SettledAt    *time.Time        `json:"settled_at,omitempty"`
	ReleasedAt   *time.Time        `json:"released_at,omitempty"`
	RetainedAt   *time.Time        `json:"retained_at,omitempty"`
	RetainReason string            `json:"retain_reason,omitempty"`
}

// ReservationRequest contains parameters needed to atomically request quota slots.
type ReservationRequest struct {
	AccountID   string       `json:"account_id"`
	Provider    string       `json:"provider"`
	WorkspaceID string       `json:"workspace_id"`
	Module      string       `json:"module"`
	Metric      ActionMetric `json:"metric"`
	Amount      int          `json:"amount"`
	TTL         time.Duration `json:"ttl"` // How long reservation remains valid before expiration
}

// UsageSummary gives visibility into consumed and currently reserved units for an account and metric.
type UsageSummary struct {
	AccountID    string       `json:"account_id"`
	Metric       ActionMetric `json:"metric"`
	Limit        int          `json:"limit"`
	Consumed     int          `json:"consumed"`
	Reserved     int          `json:"reserved"`
	Available    int          `json:"available"`
	Window       time.Duration `json:"window"`
	ParentMetric ActionMetric `json:"parent_metric,omitempty"`
	ParentUsage  *UsageSummary `json:"parent_usage,omitempty"`
}
