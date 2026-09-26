package audit_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/social-platform/services/core/internal/audit"
	"github.com/social-platform/services/core/internal/provider"
	"github.com/social-platform/services/core/internal/search"
	"github.com/social-platform/services/core/internal/workflow"
)

// REQ-023: No raw resumes, passwords, or bearer tokens in logs or diagnostics.
func TestAudit_RedactionOfSecretsAndRawResumes(t *testing.T) {
	rawLog := "Worker failed with Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.secretpayload and SSN 123-45-6789. raw_resume_text: John Doe Senior Full Stack Engineer with 10 years experience"

	sanitized := audit.RedactText(rawLog)

	if strings.Contains(sanitized, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Fatalf("RedactText leaked bearer token: %s", sanitized)
	}
	if !strings.Contains(sanitized, "Bearer [REDACTED_TOKEN]") {
		t.Fatalf("Expected token to be masked with [REDACTED_TOKEN]: %s", sanitized)
	}

	if strings.Contains(sanitized, "123-45-6789") {
		t.Fatalf("RedactText leaked SSN: %s", sanitized)
	}
	if !strings.Contains(sanitized, "[REDACTED_SSN]") {
		t.Fatalf("Expected SSN to be masked: %s", sanitized)
	}

	if strings.Contains(sanitized, "John Doe Senior Full Stack Engineer") {
		t.Fatalf("RedactText leaked raw resume content: %s", sanitized)
	}

	// Test map scrubbing
	meta := map[string]interface{}{
		"api_key":      "super_secret_api_key_12345",
		"user_email":   "alice@example.com",
		"raw_resume":   "Confidential Career History...",
		"error_detail": "Database connection failed with password=db_secret_pass",
	}

	scrubbed := audit.RedactMap(meta)
	if scrubbed["api_key"] != "[REDACTED_SECRET]" {
		t.Errorf("expected api_key to be redacted, got: %v", scrubbed["api_key"])
	}
	if scrubbed["raw_resume"] != "[REDACTED_RESUME_DATA]" {
		t.Errorf("expected raw_resume to be redacted, got: %v", scrubbed["raw_resume"])
	}
	if strings.Contains(scrubbed["error_detail"].(string), "db_secret_pass") {
		t.Errorf("error_detail leaked database password: %v", scrubbed["error_detail"])
	}
}

// AT-016 & AT-022: Account deletion cascades to vault, run ledger, search tombstones, and downstream disclosure.
func TestAudit_AccountDeletionCascade_AT016_AT022(t *testing.T) {
	ctx := context.Background()
	wsID := "ws-cascade-1"
	actorID := "user-alice"
	accountID := "li:alice_work"

	// 1. Setup Credential Vault
	vaultKey := []byte("12345678901234567890123456789012")
	vaultRepo := provider.NewMemoryCredentialVaultRepository()
	vaultService, err := provider.NewCredentialVaultService(vaultRepo, vaultKey)
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}
	credMeta, err := vaultService.StoreCredential(ctx, wsID, actorID, "linkedin", accountID, "secret_token_123", []string{"profile"}, nil)
	if err != nil {
		t.Fatalf("failed to store credential: %v", err)
	}

	// 2. Setup Run Ledger
	runRepo := workflow.NewMemoryRunLedgerRepository()
	runService := workflow.NewRunLedgerService(runRepo)
	run, _, err := runService.RegisterOrGetRun(ctx, &workflow.ActionRun{
		WorkspaceID: wsID,
		ActorID:     actorID,
		TaskType:    "linkedin.publish",
		InputPayload: []byte(`{"msg":"hello"}`),
	})
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	// 3. Setup Search Facade
	searchBackend := search.NewMemorySearchBackend()
	searchFacade := search.NewSearchFacade(searchBackend)
	docID := "post-100"
	_ = searchFacade.ProjectDocument(ctx, search.IndexContent, map[string]interface{}{
		"id":           docID,
		"workspace_id": wsID,
		"title":        "Alice LinkedIn Post",
		"status":       "published",
	})

	// 4. Execute Cascading Account Deletion
	auditService := audit.NewAuditService(vaultService, runService, searchFacade)
	report, err := auditService.CascadeAccountDeletion(ctx, wsID, actorID, accountID, []string{docID})
	if err != nil {
		t.Fatalf("CascadeAccountDeletion failed: %v", err)
	}

	// Verify report metrics
	if report.RevokedCredentials != 1 {
		t.Errorf("expected 1 revoked credential, got %d", report.RevokedCredentials)
	}
	if report.CancelledActionRuns != 1 {
		t.Errorf("expected 1 cancelled action run, got %d", report.CancelledActionRuns)
	}
	if report.TombstonedSearchEntries != 3 { // 3 indices (jobs, content, applications)
		t.Errorf("expected 3 tombstoned search entries, got %d", report.TombstonedSearchEntries)
	}

	// Verify Downstream Export Notice (AT-022)
	if !strings.Contains(report.DownstreamExportNotice, "cannot retract or overwrite copies previously exported") {
		t.Fatalf("expected downstream export limitation disclosure, got: %s", report.DownstreamExportNotice)
	}

	// Verify Vault revocation (AT-016)
	_, _, err = vaultService.GetDecryptedSecret(ctx, credMeta.ID, actorID)
	if err == nil {
		t.Fatalf("expected error decrypting revoked credential after deletion cascade, got nil")
	}

	// Verify Run cancellation (AT-016)
	fetchedRun, err := runService.GetRun(ctx, run.ID)
	if err != nil || fetchedRun.State != workflow.StateCancelled {
		t.Fatalf("expected action run to be cancelled, got state: %v", fetchedRun.State)
	}

	// Verify Search Tombstone: Document is no longer returned in search
	searchRes, err := searchFacade.Search(ctx, search.SearchRequest{
		WorkspaceID: wsID,
		Index:       search.IndexContent,
		Query:       "Alice",
	})
	if err != nil || searchRes.TotalHits != 0 {
		t.Fatalf("expected 0 search hits for tombstoned post, got %d", searchRes.TotalHits)
	}
}

// REQ-019: Notification preferences and category muting.
func TestAudit_NotificationPreferences(t *testing.T) {
	ctx := context.Background()
	svc := audit.NewAuditService(nil, nil, nil)
	wsID := "ws-prefs-1"
	userID := "user-bob"

	// 1. Set preferences: Mute RunCompleted notifications
	err := svc.UpdatePreferences(ctx, audit.NotificationPreferences{
		UserID:          userID,
		WorkspaceID:     wsID,
		InAppEnabled:    true,
		EmailEnabled:    false,
		MutedCategories: []audit.NotificationCategory{audit.CategoryRunCompleted},
	})
	if err != nil {
		t.Fatalf("UpdatePreferences failed: %v", err)
	}

	// 2. Send RunCompleted notification -> Must be suppressed (returns false)
	delivered, err := svc.SendNotification(ctx, audit.Notification{
		WorkspaceID: wsID,
		UserID:      userID,
		Category:    audit.CategoryRunCompleted,
		Title:       "Run #1 Completed",
		Message:     "All tasks finished successfully",
	})
	if err != nil || delivered {
		t.Fatalf("expected muted notification to be suppressed, delivered=%v, err=%v", delivered, err)
	}

	// 3. Send SecurityAlert notification -> Must be delivered (returns true)
	delivered, err = svc.SendNotification(ctx, audit.Notification{
		WorkspaceID: wsID,
		UserID:      userID,
		Category:    audit.CategorySecurityAlert,
		Title:       "New Login from Chrome",
		Message:     "Unusual IP detected",
	})
	if err != nil || !delivered {
		t.Fatalf("expected security alert to be delivered, delivered=%v, err=%v", delivered, err)
	}

	// Verify inbox contains only the security alert
	list, err := svc.ListNotifications(ctx, wsID, userID)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 notification in inbox, got %d (err: %v)", len(list), err)
	}
	if list[0].Category != audit.CategorySecurityAlert {
		t.Errorf("expected CategorySecurityAlert, got %v", list[0].Category)
	}
}

// REQ-019 & AT-011: Time-limited support consent lifecycle.
func TestAudit_SupportConsentLifecycle(t *testing.T) {
	ctx := context.Background()
	svc := audit.NewAuditService(nil, nil, nil)
	wsID := "ws-consent-1"
	userID := "user-carol"
	agentID := "support-agent-dave"

	// 1. Grant consent for 1 hour -> Success
	now := time.Now().UTC()
	consent, err := svc.GrantSupportConsent(ctx, audit.SupportConsent{
		WorkspaceID:   wsID,
		GranterUserID: userID,
		SupportUserID: agentID,
		Scope:         "diagnostics.read",
		ExpiresAt:     now.Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("GrantSupportConsent failed: %v", err)
	}

	// Validate consent
	if err := svc.ValidateSupportConsent(ctx, consent.ID); err != nil {
		t.Fatalf("expected consent to be valid, got: %v", err)
	}

	// 2. Reject grant exceeding 2 hours
	_, err = svc.GrantSupportConsent(ctx, audit.SupportConsent{
		WorkspaceID:   wsID,
		GranterUserID: userID,
		SupportUserID: agentID,
		ExpiresAt:     now.Add(3 * time.Hour), // Exceeds 2 hours max
	})
	if !errors.Is(err, audit.ErrInvalidDuration) {
		t.Fatalf("expected ErrInvalidDuration for >2 hours grant, got: %v", err)
	}

	// 3. Revoke consent
	err = svc.RevokeSupportConsent(ctx, consent.ID, userID)
	if err != nil {
		t.Fatalf("RevokeSupportConsent failed: %v", err)
	}

	// Verification after revocation must fail
	err = svc.ValidateSupportConsent(ctx, consent.ID)
	if !errors.Is(err, audit.ErrConsentRevoked) {
		t.Fatalf("expected ErrConsentRevoked after revocation, got: %v", err)
	}
}
