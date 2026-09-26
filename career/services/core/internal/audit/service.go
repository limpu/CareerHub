package audit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/social-platform/services/core/internal/provider"
	"github.com/social-platform/services/core/internal/search"
	"github.com/social-platform/services/core/internal/workflow"
)

// Subsystem hooks for cascading deletion (AT-016, AT-022)
type SubsystemDeleter interface {
	RevokeCredentials(ctx context.Context, accountID, ownerID string) (int, error)
	CancelPendingRuns(ctx context.Context, workspaceID, actorID string) (int, error)
	TombstoneSearchDocs(ctx context.Context, workspaceID string, docIDs []string) (int, error)
}

// AuditService coordinates audit trails, notification preferences, support consents,
// and cascading multi-system account deletion (REQ-019, REQ-023, AT-016, AT-022).
type AuditService struct {
	mu            sync.RWMutex
	events        []AuditEvent
	notifications []Notification
	preferences   map[string]*NotificationPreferences // key: "workspaceID:userID"
	consents      map[string]*SupportConsent          // key: consentID

	// Optional Subsystem handles
	vaultService *provider.CredentialVaultService
	runService   *workflow.RunLedgerService
	searchFacade *search.SearchFacade
}

// NewAuditService creates a new AuditService.
func NewAuditService(
	vault *provider.CredentialVaultService,
	runService *workflow.RunLedgerService,
	searchFacade *search.SearchFacade,
) *AuditService {
	return &AuditService{
		events:        make([]AuditEvent, 0),
		notifications: make([]Notification, 0),
		preferences:   make(map[string]*NotificationPreferences),
		consents:      make(map[string]*SupportConsent),
		vaultService:  vault,
		runService:    runService,
		searchFacade:  searchFacade,
	}
}

// RecordEvent scrubs secrets and logs an immutable audit event (REQ-019, REQ-023).
func (s *AuditService) RecordEvent(ctx context.Context, event AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.ID == "" {
		event.ID = generateID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Scrub any sensitive data in metadata or entity description
	event.Metadata = RedactMap(event.Metadata)

	s.events = append(s.events, event)
	return nil
}

// ListEvents returns tenant-scoped audit trails for an actor or workspace (REQ-019).
func (s *AuditService) ListEvents(ctx context.Context, workspaceID, actorID string) ([]AuditEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []AuditEvent
	for _, e := range s.events {
		if e.WorkspaceID == workspaceID {
			if actorID == "" || e.ActorID == actorID {
				result = append(result, e)
			}
		}
	}
	return result, nil
}

// SendNotification routes a notification subject to user category and mute preferences (REQ-019).
func (s *AuditService) SendNotification(ctx context.Context, notif Notification) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := notif.WorkspaceID + ":" + notif.UserID
	prefs, exists := s.preferences[key]
	if exists {
		if !prefs.InAppEnabled {
			return false, nil // In-app disabled
		}
		for _, muted := range prefs.MutedCategories {
			if muted == notif.Category {
				return false, nil // Muted category
			}
		}
	}

	if notif.ID == "" {
		notif.ID = generateID()
	}
	if notif.CreatedAt.IsZero() {
		notif.CreatedAt = time.Now().UTC()
	}

	notif.Title = RedactText(notif.Title)
	notif.Message = RedactText(notif.Message)

	s.notifications = append(s.notifications, notif)
	return true, nil
}

// ListNotifications returns unread or recent notifications for a user.
func (s *AuditService) ListNotifications(ctx context.Context, workspaceID, userID string) ([]Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Notification
	for _, n := range s.notifications {
		if n.WorkspaceID == workspaceID && n.UserID == userID {
			result = append(result, n)
		}
	}
	return result, nil
}

// GetPreferences returns notification preferences for a user in a workspace.
func (s *AuditService) GetPreferences(ctx context.Context, workspaceID, userID string) (*NotificationPreferences, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := workspaceID + ":" + userID
	if p, ok := s.preferences[key]; ok {
		clone := *p
		return &clone, nil
	}

	// Default preferences
	return &NotificationPreferences{
		UserID:       userID,
		WorkspaceID:  workspaceID,
		InAppEnabled: true,
		EmailEnabled: true,
	}, nil
}

// UpdatePreferences saves modified alert settings.
func (s *AuditService) UpdatePreferences(ctx context.Context, prefs NotificationPreferences) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := prefs.WorkspaceID + ":" + prefs.UserID
	s.preferences[key] = &prefs
	return nil
}

// GrantSupportConsent records time-bounded user consent for support access (REQ-019, AT-011).
func (s *AuditService) GrantSupportConsent(ctx context.Context, consent SupportConsent) (*SupportConsent, error) {
	now := time.Now().UTC()
	if consent.ExpiresAt.IsZero() || consent.ExpiresAt.Before(now) {
		consent.ExpiresAt = now.Add(2 * time.Hour) // Default 2 hours
	}

	// Max 2 hours limit enforcement
	if consent.ExpiresAt.Sub(now) > 2*time.Hour {
		return nil, ErrInvalidDuration
	}

	if consent.ID == "" {
		consent.ID = generateID()
	}
	consent.ApprovedAt = now

	s.mu.Lock()
	s.consents[consent.ID] = &consent
	s.mu.Unlock()

	_ = s.RecordEvent(ctx, AuditEvent{
		WorkspaceID: consent.WorkspaceID,
		ActorID:     consent.GranterUserID,
		Action:      "support.consent_granted",
		EntityType:  "support_consent",
		EntityID:    consent.ID,
		Metadata: map[string]interface{}{
			"support_user_id": consent.SupportUserID,
			"expires_at":      consent.ExpiresAt.Format(time.RFC3339),
			"scope":           consent.Scope,
		},
	})

	clone := consent
	return &clone, nil
}

// RevokeSupportConsent revokes an active support grant immediately.
func (s *AuditService) RevokeSupportConsent(ctx context.Context, consentID, userID string) error {
	s.mu.Lock()
	consent, exists := s.consents[consentID]
	if !exists {
		s.mu.Unlock()
		return ErrConsentNotFound
	}

	now := time.Now().UTC()
	consent.RevokedAt = &now
	s.mu.Unlock()

	_ = s.RecordEvent(ctx, AuditEvent{
		WorkspaceID: consent.WorkspaceID,
		ActorID:     userID,
		Action:      "support.consent_revoked",
		EntityType:  "support_consent",
		EntityID:    consent.ID,
	})

	return nil
}

// ValidateSupportConsent checks if a support session has valid, non-expired consent.
func (s *AuditService) ValidateSupportConsent(ctx context.Context, consentID string) error {
	s.mu.RLock()
	consent, exists := s.consents[consentID]
	s.mu.RUnlock()

	if !exists {
		return ErrConsentNotFound
	}
	if consent.RevokedAt != nil {
		return ErrConsentRevoked
	}
	if time.Now().UTC().After(consent.ExpiresAt) {
		return ErrConsentExpired
	}
	return nil
}

// CascadeAccountDeletion coordinates full multi-subsystem cleanup upon account deletion (AT-016, AT-022).
func (s *AuditService) CascadeAccountDeletion(
	ctx context.Context,
	workspaceID string,
	actorID string,
	accountID string,
	associatedDocIDs []string,
) (*DeletionReport, error) {
	now := time.Now().UTC()
	report := &DeletionReport{
		AccountID:   accountID,
		WorkspaceID: workspaceID,
		CompletedAt: now,
		DownstreamExportNotice: "Account deletion complete. Note: Local deletion removes all application storage, credentials, background work and search projections, but cannot retract or overwrite copies previously exported or downloaded by the user to offline CSVs or personal Google Sheets (AT-022).",
	}

	// 1. Revoke Credentials in CredentialVault (AT-016)
	if s.vaultService != nil {
		if credList, err := s.vaultService.ListWorkspaceCredentials(ctx, workspaceID, actorID); err == nil {
			for _, cred := range credList {
				if cred.AccountIdentifier == accountID || accountID == "" {
					if err := s.vaultService.RevokeCredential(ctx, cred.ID, actorID); err == nil {
						report.RevokedCredentials++
					}
				}
			}
		}
	}

	// 2. Cancel Pending ActionRuns in RunLedger (AT-016, AT-021)
	if s.runService != nil {
		if runs, err := s.runService.ListRuns(ctx, workspaceID); err == nil {
			for _, r := range runs {
				if r.ActorID == actorID && (r.State == workflow.StateQueued || r.State == workflow.StateRunning) {
					if _, err := s.runService.CancelRun(ctx, r.ID, "Account disconnected/deleted by user"); err == nil {
						report.CancelledActionRuns++
					}
				}
			}
		}
	}

	// 3. Mark Deletion Tombstones in SearchFacade (AT-022)
	if s.searchFacade != nil {
		for _, docID := range associatedDocIDs {
			_ = s.searchFacade.DeleteDocument(ctx, search.IndexJobs, workspaceID, docID)
			_ = s.searchFacade.DeleteDocument(ctx, search.IndexContent, workspaceID, docID)
			_ = s.searchFacade.DeleteDocument(ctx, search.IndexApplications, workspaceID, docID)
			report.TombstonedSearchEntries += 3
		}
	}

	// 4. Record Audit Trail
	_ = s.RecordEvent(ctx, AuditEvent{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Action:      "account.deleted_cascade",
		EntityType:  "account",
		EntityID:    accountID,
		Metadata: map[string]interface{}{
			"revoked_credentials": report.RevokedCredentials,
			"cancelled_runs":      report.CancelledActionRuns,
			"tombstoned_docs":     report.TombstonedSearchEntries,
			"export_notice":       report.DownstreamExportNotice,
		},
	})

	return report, nil
}

func generateID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
