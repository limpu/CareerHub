package linkedin

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Standard domain errors for session lifecycle and optional local tool leasing (AT-016, FND-009, SRC-L3)
var (
	ErrPlaintextCredentialProhibited = errors.New("plaintext credential or cookie paste is strictly prohibited as default connection; use cryptographic vault reference (AT-016, REQ-021)")
	ErrSessionNotFound               = errors.New("linkedin session record not found or inaccessible for tenant")
	ErrSessionNotActive              = errors.New("session is not active; operations cannot proceed")
	ErrSessionRevoked                = errors.New("session has been explicitly revoked by user (FND-009)")
	ErrViewerLeaseExpired            = errors.New("short-lived viewer lease has expired; lock engaged (SRC-L3 profile_lease.py)")
	ErrViewerLeaseNotActive          = errors.New("viewer lease is not active")
	ErrViewerDurationTooLong         = errors.New("viewer lease duration exceeds maximum allowable bound of 900 seconds (15 minutes)")
	ErrCrossTenantSessionAccess      = errors.New("cross-tenant session access denied (AT-011, AT-012)")
)

// SessionStatus represents the operational lifecycle state of an owner's LinkedIn connection.
type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusIdle    SessionStatus = "idle"
	SessionStatusLeasing SessionStatus = "leasing"
	SessionStatusRevoked SessionStatus = "revoked"
	SessionStatusExpired SessionStatus = "expired"
)

// ViewerLeaseStatus defines the lifecycle status of a temporary supervised viewer lock.
type ViewerLeaseStatus string

const (
	ViewerLeaseStatusActive   ViewerLeaseStatus = "active"
	ViewerLeaseStatusReleased ViewerLeaseStatus = "released"
	ViewerLeaseStatusExpired  ViewerLeaseStatus = "expired"
)

// LinkedInSessionRecord holds per-owner isolated storage records for external sessions (LI-19, AT-016).
type LinkedInSessionRecord struct {
	SessionID            string        `json:"session_id"`
	TenantID             string        `json:"tenant_id"`
	OwnerID              string        `json:"owner_id"`
	AccountID            string        `json:"account_id"`
	VaultCredentialRef   string        `json:"vault_credential_ref"` // E.g. "vault://credentials/tenant-01/linkedin/token"
	Status               SessionStatus `json:"status"`
	IsLocalOnly          bool          `json:"is_local_only"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
	ExpiresAt            time.Time     `json:"expires_at"`
	RevokedAt            *time.Time    `json:"revoked_at,omitempty"`
	RevokedReason        string        `json:"revoked_reason,omitempty"`
}

// ShortLivedViewerLease encapsulates a time-bounded supervised browser viewer lock (SRC-L3 profile_lease.py, daemon_lock.py).
type ShortLivedViewerLease struct {
	LeaseID                 string            `json:"lease_id"`
	SessionID               string            `json:"session_id"`
	TenantID                string            `json:"tenant_id"`
	OwnerID                 string            `json:"owner_id"`
	Purpose                 string            `json:"purpose"`
	DurationSeconds         int               `json:"duration_seconds"`
	RequiresUserSupervision bool              `json:"requires_user_supervision"`
	StartedAt               time.Time         `json:"started_at"`
	ExpiresAt               time.Time         `json:"expires_at"`
	Status                  ViewerLeaseStatus `json:"status"`
}

// CreateSessionRequest defines incoming initialization parameters.
type CreateSessionRequest struct {
	TenantID           string `json:"tenant_id"`
	OwnerID            string `json:"owner_id"`
	AccountID          string `json:"account_id"`
	VaultCredentialRef string `json:"vault_credential_ref"`
	RawCookiePaste     string `json:"raw_cookie_paste"`
	RawPasswordPaste   string `json:"raw_password_paste"`
	DurationHours      int    `json:"duration_hours"`
	IsLocalOnly        bool   `json:"is_local_only"`
}

// AcquireViewerLeaseRequest defines temporary viewer lock acquisition.
type AcquireViewerLeaseRequest struct {
	SessionID               string `json:"session_id"`
	Purpose                 string `json:"purpose"`
	DurationSeconds         int    `json:"duration_seconds"`
	RequiresUserSupervision bool   `json:"requires_user_supervision"`
}

// RevokeSessionRequest defines explicit user cancellation.
type RevokeSessionRequest struct {
	SessionID string `json:"session_id"`
	Reason    string `json:"reason"`
}

// CreateIsolatedSession validates storage security constraints and constructs a new per-owner session.
// In compliance with AT-016 and REQ-021, raw plaintext cookie/password pastes are permanently rejected.
func CreateIsolatedSession(req CreateSessionRequest, now time.Time) (*LinkedInSessionRecord, error) {
	// Zero plaintext credential / cookie paste rule (AT-016, REQ-021)
	if strings.TrimSpace(req.RawCookiePaste) != "" || strings.TrimSpace(req.RawPasswordPaste) != "" {
		return nil, ErrPlaintextCredentialProhibited
	}

	if strings.TrimSpace(req.VaultCredentialRef) == "" {
		return nil, errors.New("vault_credential_ref is required; unvaulted credentials cannot be stored")
	}

	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.OwnerID) == "" {
		return nil, errors.New("tenant_id and owner_id are required for isolated per-owner storage")
	}

	if strings.TrimSpace(req.AccountID) == "" {
		req.AccountID = "li:member_default"
	}

	duration := 24 * time.Hour * 30 // Default 30-day credential validity window
	if req.DurationHours > 0 {
		duration = time.Duration(req.DurationHours) * time.Hour
	}

	sessionID := fmt.Sprintf("sess-%s-%d", req.OwnerID, now.UnixNano())

	return &LinkedInSessionRecord{
		SessionID:          sessionID,
		TenantID:           req.TenantID,
		OwnerID:            req.OwnerID,
		AccountID:          req.AccountID,
		VaultCredentialRef: req.VaultCredentialRef,
		Status:             SessionStatusActive,
		IsLocalOnly:        req.IsLocalOnly,
		CreatedAt:          now,
		UpdatedAt:          now,
		ExpiresAt:          now.Add(duration),
	}, nil
}

// RevokeSession processes an immediate, irreversible user revocation (FND-009, LI-19).
func RevokeSession(session *LinkedInSessionRecord, reason string, now time.Time) error {
	if session == nil {
		return ErrSessionNotFound
	}
	session.Status = SessionStatusRevoked
	session.RevokedAt = &now
	session.RevokedReason = reason
	session.UpdatedAt = now
	return nil
}

// AcquireShortLivedViewerLease creates a time-bounded supervised viewer lock (SRC-L3 daemon_lock.py, profile_lease.py).
func AcquireShortLivedViewerLease(
	session *LinkedInSessionRecord,
	req AcquireViewerLeaseRequest,
	now time.Time,
) (*ShortLivedViewerLease, error) {
	if session == nil {
		return nil, ErrSessionNotFound
	}
	if session.Status == SessionStatusRevoked {
		return nil, ErrSessionRevoked
	}
	if session.Status != SessionStatusActive && session.Status != SessionStatusLeasing {
		return nil, ErrSessionNotActive
	}

	duration := req.DurationSeconds
	if duration <= 0 {
		duration = 300 // Standard 5-minute conservative default
	}
	if duration > 900 {
		return nil, ErrViewerDurationTooLong // Capped at 15 minutes
	}

	leaseID := fmt.Sprintf("lease-%s-%d", session.SessionID, now.UnixNano())
	expiresAt := now.Add(time.Duration(duration) * time.Second)

	session.Status = SessionStatusLeasing
	session.UpdatedAt = now

	return &ShortLivedViewerLease{
		LeaseID:                 leaseID,
		SessionID:               session.SessionID,
		TenantID:                session.TenantID,
		OwnerID:                 session.OwnerID,
		Purpose:                 req.Purpose,
		DurationSeconds:         duration,
		RequiresUserSupervision: req.RequiresUserSupervision,
		StartedAt:               now,
		ExpiresAt:               expiresAt,
		Status:                  ViewerLeaseStatusActive,
	}, nil
}

// ValidateViewerLease inspects lease validity and ensures automatic locking upon expiry.
func ValidateViewerLease(lease *ShortLivedViewerLease, now time.Time) error {
	if lease == nil {
		return errors.New("lease record not found")
	}

	if lease.Status == ViewerLeaseStatusReleased {
		return errors.New("viewer lease has already been released")
	}

	if now.After(lease.ExpiresAt) {
		lease.Status = ViewerLeaseStatusExpired
		return ErrViewerLeaseExpired
	}

	if lease.Status != ViewerLeaseStatusActive {
		return ErrViewerLeaseNotActive
	}

	return nil
}

// ReleaseViewerLease finishes a viewer session and unlocks the underlying account session.
func ReleaseViewerLease(lease *ShortLivedViewerLease, session *LinkedInSessionRecord, now time.Time) error {
	if lease == nil {
		return errors.New("lease record not found")
	}
	lease.Status = ViewerLeaseStatusReleased
	if session != nil && session.Status == SessionStatusLeasing {
		session.Status = SessionStatusActive
		session.UpdatedAt = now
	}
	return nil
}
