package linkedin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type sessionLifecycleScenario struct {
	ScenarioID            string `json:"scenario_id"`
	Description           string `json:"description"`
	TenantID              string `json:"tenant_id"`
	OwnerID               string `json:"owner_id"`
	AccountID             string `json:"account_id"`
	IncomingPayload       struct {
		VaultCredentialRef string `json:"vault_credential_ref"`
		RawCookiePaste     string `json:"raw_cookie_paste"`
		RawPasswordPaste   string `json:"raw_password_paste"`
		StorageMode        string `json:"storage_mode"`
	} `json:"incoming_payload"`
	ExpectedSessionStatus     string `json:"expected_session_status"`
	ExpectedPlaintextRejected bool   `json:"expected_plaintext_rejected"`
	ExpectedVaultLinked       bool   `json:"expected_vault_linked"`
	ExpectedErrorCode         string `json:"expected_error_code"`
	ExpectedRemediation       string `json:"expected_remediation"`

	// Scenario 3
	OwnerA struct {
		TenantID  string `json:"tenant_id"`
		OwnerID   string `json:"owner_id"`
		SessionID string `json:"session_id"`
	} `json:"owner_a"`
	OwnerB struct {
		TenantID  string `json:"tenant_id"`
		OwnerID   string `json:"owner_id"`
		SessionID string `json:"session_id"`
	} `json:"owner_b"`
	ExpectedCrossTenantAccessAllowed bool `json:"expected_cross_tenant_access_allowed"`
	ExpectedIsolationEnforced        bool `json:"expected_isolation_barrier_enforced"`

	// Scenario 4
	SessionID                        string `json:"session_id"`
	RevocationTrigger                string `json:"revocation_trigger"`
	ExpectedPostRevocationStatus     string `json:"expected_post_revocation_status"`
	ExpectedActiveLeaseCleared       bool   `json:"expected_active_lease_cleared"`
	ExpectedDispatchAllowedAfterRev  bool   `json:"expected_dispatch_allowed_after_revoke"`

	// Scenario 5
	LeaseRequest struct {
		Purpose                 string `json:"purpose"`
		DurationSeconds         int    `json:"duration_seconds"`
		RequiresUserSupervision bool   `json:"requires_user_supervision"`
	} `json:"lease_request"`
	ExpectedInitialLeaseStatus    string `json:"expected_initial_lease_status"`
	SimulatedElapsedSeconds       int    `json:"simulated_elapsed_seconds"`
	ExpectedExpiredLeaseStatus    string `json:"expected_expired_lease_status"`
	ExpectedSubsequentBlock       bool   `json:"expected_subsequent_access_blocked"`
}

type sessionLifecycleFixtureDoc struct {
	Scenarios []sessionLifecycleScenario `json:"scenarios"`
}

func loadSessionLifecycleFixture(t *testing.T) map[string]sessionLifecycleScenario {
	t.Helper()
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_session_lifecycle.json")
	bytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture file: %v", err)
	}
	var doc sessionLifecycleFixtureDoc
	if err := json.Unmarshal(bytes, &doc); err != nil {
		t.Fatalf("Failed to unmarshal fixture JSON: %v", err)
	}
	scenarioMap := make(map[string]sessionLifecycleScenario)
	for _, sc := range doc.Scenarios {
		scenarioMap[sc.ScenarioID] = sc
	}
	return scenarioMap
}

func TestSessionLifecycle_FixtureGroundTruth(t *testing.T) {
	scenarios := loadSessionLifecycleFixture(t)
	ctx := context.Background()
	store := NewMemoryRepository()
	service := NewService(store)

	// 1. Secure session initialization (AT-016, REQ-021)
	sc1, ok := scenarios["secure_session_initialization"]
	if !ok {
		t.Fatalf("Scenario secure_session_initialization missing")
	}
	sess1, err := service.CreateSession(ctx, CreateSessionRequest{
		TenantID:           sc1.TenantID,
		OwnerID:            sc1.OwnerID,
		AccountID:          sc1.AccountID,
		VaultCredentialRef: sc1.IncomingPayload.VaultCredentialRef,
		RawCookiePaste:     sc1.IncomingPayload.RawCookiePaste,
		RawPasswordPaste:   sc1.IncomingPayload.RawPasswordPaste,
		DurationHours:      24 * 30,
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if string(sess1.Status) != sc1.ExpectedSessionStatus {
		t.Errorf("Session status mismatch: got %v, want %v", sess1.Status, sc1.ExpectedSessionStatus)
	}
	if sc1.ExpectedVaultLinked && sess1.VaultCredentialRef != sc1.IncomingPayload.VaultCredentialRef {
		t.Errorf("Vault credential ref mismatch: got %v", sess1.VaultCredentialRef)
	}

	// 2. Plaintext cookie paste rejected (AT-016, REQ-021)
	sc2, ok := scenarios["plaintext_cookie_paste_rejected"]
	if !ok {
		t.Fatalf("Scenario plaintext_cookie_paste_rejected missing")
	}
	_, err2 := service.CreateSession(ctx, CreateSessionRequest{
		TenantID:           sc2.TenantID,
		OwnerID:            sc2.OwnerID,
		AccountID:          sc2.AccountID,
		VaultCredentialRef: sc2.IncomingPayload.VaultCredentialRef,
		RawCookiePaste:     sc2.IncomingPayload.RawCookiePaste,
		RawPasswordPaste:   sc2.IncomingPayload.RawPasswordPaste,
	})
	if err2 == nil {
		t.Errorf("Expected raw plaintext cookie/password upload to fail, but succeeded")
	}
	if err2 != ErrPlaintextCredentialProhibited {
		t.Errorf("Expected ErrPlaintextCredentialProhibited, got %v", err2)
	}

	// 3. Cross-owner storage isolation (AT-011, AT-012)
	sc3, ok := scenarios["cross_owner_storage_isolation"]
	if !ok {
		t.Fatalf("Scenario cross_owner_storage_isolation missing")
	}
	// Create Owner A session in Tenant Corp 01
	sessA, err := service.CreateSession(ctx, CreateSessionRequest{
		TenantID:           sc3.OwnerA.TenantID,
		OwnerID:            sc3.OwnerA.OwnerID,
		AccountID:          "li:member_alex_101",
		VaultCredentialRef: "vault://credentials/tenant-corp-01/alex",
	})
	if err != nil {
		t.Fatalf("CreateSession A failed: %v", err)
	}

	// Attempt to access Owner A's session using Tenant Corp 02
	_, errCross := service.GetSession(ctx, sessA.SessionID, sc3.OwnerB.TenantID)
	if errCross == nil {
		t.Errorf("Cross-tenant session lookup MUST be denied, but succeeded!")
	}
	if errCross != ErrCrossTenantSessionAccess {
		t.Errorf("Expected ErrCrossTenantSessionAccess, got: %v", errCross)
	}

	// 4. Explicit session revocation (FND-009, LI-19)
	sc4, ok := scenarios["explicit_session_revocation"]
	if !ok {
		t.Fatalf("Scenario explicit_session_revocation missing")
	}
	revokedSess, err := service.RevokeSession(ctx, sess1.SessionID, sc4.TenantID, sc4.RevocationTrigger)
	if err != nil {
		t.Fatalf("RevokeSession failed: %v", err)
	}
	if string(revokedSess.Status) != sc4.ExpectedPostRevocationStatus {
		t.Errorf("Expected status %v, got %v", sc4.ExpectedPostRevocationStatus, revokedSess.Status)
	}
	if revokedSess.RevokedAt == nil {
		t.Errorf("Expected RevokedAt to be populated")
	}

	// Subsequent lease attempt on revoked session must fail
	_, errLeaseRevoked := service.AcquireViewerLease(ctx, sess1.SessionID, sc4.TenantID, AcquireViewerLeaseRequest{
		Purpose:         "test",
		DurationSeconds: 300,
	})
	if errLeaseRevoked == nil {
		t.Errorf("Lease acquisition on revoked session should fail")
	}
	if errLeaseRevoked != ErrSessionRevoked {
		t.Errorf("Expected ErrSessionRevoked, got: %v", errLeaseRevoked)
	}

	// 5. Short-lived viewer lease expiry (SRC-L3 profile_lease.py, daemon_lock.py)
	sc5, ok := scenarios["short_lived_viewer_lease_expiry"]
	if !ok {
		t.Fatalf("Scenario short_lived_viewer_lease_expiry missing")
	}
	// Create fresh active session for lease testing
	sessLeaseTest, err := service.CreateSession(ctx, CreateSessionRequest{
		TenantID:           sc5.TenantID,
		OwnerID:            sc5.OwnerID,
		AccountID:          sc5.AccountID,
		VaultCredentialRef: "vault://credentials/tenant-corp-01/lease_test",
	})
	if err != nil {
		t.Fatalf("Failed to create lease test session: %v", err)
	}

	lease, err := service.AcquireViewerLease(ctx, sessLeaseTest.SessionID, sc5.TenantID, AcquireViewerLeaseRequest{
		Purpose:                 sc5.LeaseRequest.Purpose,
		DurationSeconds:         sc5.LeaseRequest.DurationSeconds,
		RequiresUserSupervision: sc5.LeaseRequest.RequiresUserSupervision,
	})
	if err != nil {
		t.Fatalf("AcquireViewerLease failed: %v", err)
	}
	if string(lease.Status) != sc5.ExpectedInitialLeaseStatus {
		t.Errorf("Initial lease status mismatch: got %v, want %v", lease.Status, sc5.ExpectedInitialLeaseStatus)
	}

	// Fast-forward time past expiration
	futureTime := lease.StartedAt.Add(time.Duration(sc5.SimulatedElapsedSeconds) * time.Second)
	valErr := ValidateViewerLease(lease, futureTime)
	if valErr == nil {
		t.Errorf("Expired lease must fail validation, but returned nil error")
	}
	if valErr != ErrViewerLeaseExpired {
		t.Errorf("Expected ErrViewerLeaseExpired, got: %v", valErr)
	}
	if string(lease.Status) != sc5.ExpectedExpiredLeaseStatus {
		t.Errorf("Expected lease status to transition to expired, got %v", lease.Status)
	}
}

func TestSessionLifecycle_ZeroPlaintextCookiePaste(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryRepository()
	service := NewService(store)

	badInputs := []struct {
		name     string
		cookie   string
		password string
	}{
		{
			name:   "Raw li_at cookie string",
			cookie: "li_at=AQEDAT43521abcdef9876",
		},
		{
			name:     "Raw plaintext password",
			password: "P@ssw0rd12345!",
		},
		{
			name:     "Both cookie and password",
			cookie:   "JSESSIONID=\"ajax:12345\"",
			password: "MySecretPassword",
		},
	}

	for _, tc := range badInputs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.CreateSession(ctx, CreateSessionRequest{
				TenantID:           "tenant-1",
				OwnerID:            "user-1",
				AccountID:          "li:member_test",
				VaultCredentialRef: "vault://creds/1",
				RawCookiePaste:     tc.cookie,
				RawPasswordPaste:   tc.password,
			})
			if err == nil {
				t.Fatalf("Expected error for %s, got success", tc.name)
			}
			if err != ErrPlaintextCredentialProhibited {
				t.Errorf("Expected ErrPlaintextCredentialProhibited, got %v", err)
			}
		})
	}
}

func TestSessionLifecycle_PerOwnerStorageIsolation(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryRepository()
	service := NewService(store)

	// User 1 in Tenant Alpha
	sess1, err := service.CreateSession(ctx, CreateSessionRequest{
		TenantID:           "tenant-alpha",
		OwnerID:            "user-alpha",
		AccountID:          "li:member_1",
		VaultCredentialRef: "vault://alpha/token",
	})
	if err != nil {
		t.Fatalf("CreateSession user 1 failed: %v", err)
	}

	// User 2 in Tenant Beta
	sess2, err := service.CreateSession(ctx, CreateSessionRequest{
		TenantID:           "tenant-beta",
		OwnerID:            "user-beta",
		AccountID:          "li:member_2",
		VaultCredentialRef: "vault://beta/token",
	})
	if err != nil {
		t.Fatalf("CreateSession user 2 failed: %v", err)
	}

	// 1. Tenant Alpha listing should only return Tenant Alpha sessions
	listAlpha, err := service.ListSessions(ctx, "", "tenant-alpha")
	if err != nil {
		t.Fatalf("ListSessions alpha failed: %v", err)
	}
	if len(listAlpha) != 1 || listAlpha[0].SessionID != sess1.SessionID {
		t.Errorf("Tenant Alpha list leaked foreign sessions: got %v", listAlpha)
	}

	// 2. Cross lookup of sess2 by tenant-alpha must be rejected
	_, errCross := service.GetSession(ctx, sess2.SessionID, "tenant-alpha")
	if errCross == nil {
		t.Fatalf("Cross tenant access of sess2 by tenant-alpha should fail")
	}
	if errCross != ErrCrossTenantSessionAccess {
		t.Errorf("Expected ErrCrossTenantSessionAccess, got %v", errCross)
	}
}

func TestSessionLifecycle_ShortLivedViewerLease(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryRepository()
	service := NewService(store)

	sess, err := service.CreateSession(ctx, CreateSessionRequest{
		TenantID:           "tenant-viewer",
		OwnerID:            "user-viewer",
		AccountID:          "li:member_viewer",
		VaultCredentialRef: "vault://viewer/token",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// 1. Cannot request lease longer than 900 seconds (15 mins)
	_, errOver := service.AcquireViewerLease(ctx, sess.SessionID, "tenant-viewer", AcquireViewerLeaseRequest{
		Purpose:         "long_job",
		DurationSeconds: 1200, // 20 mins
	})
	if errOver == nil {
		t.Fatalf("Expected error for duration > 900 seconds")
	}
	if errOver != ErrViewerDurationTooLong {
		t.Errorf("Expected ErrViewerDurationTooLong, got %v", errOver)
	}

	// 2. Normal 5-minute lease succeeds
	lease, err := service.AcquireViewerLease(ctx, sess.SessionID, "tenant-viewer", AcquireViewerLeaseRequest{
		Purpose:                 "profile_verification",
		DurationSeconds:         300,
		RequiresUserSupervision: true,
	})
	if err != nil {
		t.Fatalf("AcquireViewerLease failed: %v", err)
	}

	// Validate within window
	_, validErr := service.ValidateViewerLease(ctx, lease.LeaseID)
	if validErr != nil {
		t.Fatalf("ValidateViewerLease within window failed: %v", validErr)
	}

	// Release lease early
	relErr := service.ReleaseViewerLease(ctx, lease.LeaseID, "tenant-viewer")
	if relErr != nil {
		t.Fatalf("ReleaseViewerLease failed: %v", relErr)
	}

	// After release, session returns to active
	sessAfter, err := service.GetSession(ctx, sess.SessionID, "tenant-viewer")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if sessAfter.Status != SessionStatusActive {
		t.Errorf("Expected session to return to active status, got: %v", sessAfter.Status)
	}
}
