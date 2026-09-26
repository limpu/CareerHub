package audit

import (
	"errors"
	"time"
)

var (
	ErrConsentExpired   = errors.New("support consent has expired")
	ErrConsentRevoked   = errors.New("support consent was revoked by user")
	ErrConsentNotFound  = errors.New("support consent grant not found")
	ErrInvalidDuration  = errors.New("support consent duration exceeds allowed maximum of 2 hours")
)

// AuditEvent records a tamper-evident system or user action (REQ-019).
type AuditEvent struct {
	ID          string                 `json:"id"`
	WorkspaceID string                 `json:"workspace_id"`
	ActorID     string                 `json:"actor_id"`
	Action      string                 `json:"action"` // e.g., "credential.create", "run.cancel", "account.delete"
	EntityType  string                 `json:"entity_type"`
	EntityID    string                 `json:"entity_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // Strictly redacted
	IPAddress   string                 `json:"ip_address,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NotificationCategory classifies notifications for user preference management.
type NotificationCategory string

const (
	CategoryApprovalNeeded NotificationCategory = "approval_needed"
	CategoryRunCompleted   NotificationCategory = "run_completed"
	CategoryRunFailed      NotificationCategory = "run_failed"
	CategorySecurityAlert  NotificationCategory = "security_alert"
	CategoryQuotaWarning   NotificationCategory = "quota_warning"
)

// Notification represents a user-facing event message.
type Notification struct {
	ID          string               `json:"id"`
	WorkspaceID string               `json:"workspace_id"`
	UserID      string               `json:"user_id"`
	Category    NotificationCategory `json:"category"`
	Title       string               `json:"title"`
	Message     string               `json:"message"`
	Read        bool                 `json:"read"`
	CreatedAt   time.Time            `json:"created_at"`
}

// NotificationPreferences allows users to control alerts and mute categories (REQ-019).
type NotificationPreferences struct {
	UserID          string                 `json:"user_id"`
	WorkspaceID     string                 `json:"workspace_id"`
	InAppEnabled    bool                   `json:"in_app_enabled"`
	EmailEnabled    bool                   `json:"email_enabled"`
	MutedCategories []NotificationCategory `json:"muted_categories"`
}

// SupportConsent represents an explicit, time-bounded support grant (REQ-019, AT-011).
type SupportConsent struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	GranterUserID string    `json:"granter_user_id"`
	SupportUserID string    `json:"support_user_id"`
	Scope         string    `json:"scope"` // e.g. "diagnostics.read"
	ApprovedAt    time.Time `json:"approved_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
}

// DeletionReport details the comprehensive cascade cleanups across all subsystems (AT-016, AT-022).
type DeletionReport struct {
	AccountID               string    `json:"account_id"`
	WorkspaceID             string    `json:"workspace_id"`
	RevokedCredentials      int       `json:"revoked_credentials"`
	CancelledActionRuns     int       `json:"cancelled_action_runs"`
	PurgedStorageBlobs      int       `json:"purged_storage_blobs"`
	TombstonedSearchEntries int       `json:"tombstoned_search_entries"`
	DownstreamExportNotice  string    `json:"downstream_export_notice"`
	CompletedAt             time.Time `json:"completed_at"`
}
