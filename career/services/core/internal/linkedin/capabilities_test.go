package linkedin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type capabilitiesFixtureSuite struct {
	Version   string `json:"version"`
	Domain    string `json:"domain"`
	Scenarios []struct {
		ScenarioID     string `json:"scenario_id"`
		Description    string `json:"description"`
		ConnectionData struct {
			UserID         string   `json:"user_id"`
			MemberID       string   `json:"member_id"`
			DisplayName    string   `json:"display_name"`
			Email          string   `json:"email"`
			GrantedScopes  []string `json:"granted_scopes"`
			TokenStatus    string   `json:"token_status"`
			ExpiresInHours int      `json:"expires_in_hours"`
		} `json:"connection_data"`
		ExpectedConnectionStatus         string `json:"expected_connection_status"`
		ExpectedGrantedCapabilitiesCount int    `json:"expected_granted_capabilities_count"`
		ExpectedDeniedOrUnavailableCount int    `json:"expected_denied_or_unavailable_count"`
		ExpectedCapabilities             []struct {
			CapabilityID string `json:"capability_id"`
			IsGranted    bool   `json:"is_granted"`
			Status       string `json:"status"`
			FallbackMode string `json:"fallback_mode,omitempty"`
		} `json:"expected_capabilities,omitempty"`
		ActionsToTest []struct {
			Action          string `json:"action"`
			ExpectedAllowed bool   `json:"expected_allowed"`
			ExpectedError   string `json:"expected_error,omitempty"`
		} `json:"actions_to_test"`
	} `json:"scenarios"`
}

func loadCapabilitiesFixture(t *testing.T) capabilitiesFixtureSuite {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_capabilities_inspection.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture file: %v", err)
	}

	var suite capabilitiesFixtureSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("failed to unmarshal fixture JSON: %v", err)
	}
	return suite
}

func TestLinkedIn_FixtureSuiteEvaluation(t *testing.T) {
	suite := loadCapabilitiesFixture(t)
	if len(suite.Scenarios) != 4 {
		t.Fatalf("expected 4 fixture scenarios, got %d", len(suite.Scenarios))
	}

	for _, s := range suite.Scenarios {
		t.Run(s.ScenarioID, func(t *testing.T) {
			now := time.Now().UTC()
			var expiresAt *time.Time
			if s.ConnectionData.ExpiresInHours != 0 {
				exp := now.Add(time.Duration(s.ConnectionData.ExpiresInHours) * time.Hour)
				expiresAt = &exp
			}

			conn := &LinkedInConnection{
				ConnectionID:  "conn-" + s.ScenarioID,
				UserID:        s.ConnectionData.UserID,
				MemberID:      s.ConnectionData.MemberID,
				DisplayName:   s.ConnectionData.DisplayName,
				Email:         s.ConnectionData.Email,
				GrantedScopes: s.ConnectionData.GrantedScopes,
				TokenStatus:   s.ConnectionData.TokenStatus,
				ExpiresAt:     expiresAt,
				Status:        ConnectionStatus(s.ExpectedConnectionStatus),
			}

			capabilities := InspectCapabilities(conn)
			conn.Capabilities = capabilities

			grantedCount := 0
			deniedOrUnavailableCount := 0
			for _, c := range capabilities {
				if c.IsGranted {
					grantedCount++
				} else {
					deniedOrUnavailableCount++
				}
			}

			if grantedCount != s.ExpectedGrantedCapabilitiesCount {
				t.Errorf("expected %d granted capabilities, got %d", s.ExpectedGrantedCapabilitiesCount, grantedCount)
			}

			if s.ExpectedDeniedOrUnavailableCount > 0 && deniedOrUnavailableCount != s.ExpectedDeniedOrUnavailableCount {
				t.Errorf("expected %d denied/unavailable capabilities, got %d", s.ExpectedDeniedOrUnavailableCount, deniedOrUnavailableCount)
			}

			// Validate specific expected capabilities
			for _, expCap := range s.ExpectedCapabilities {
				found := false
				for _, c := range capabilities {
					if c.CapabilityID == expCap.CapabilityID {
						found = true
						if c.IsGranted != expCap.IsGranted {
							t.Errorf("cap %s is_granted mismatch: got %v, want %v", c.CapabilityID, c.IsGranted, expCap.IsGranted)
						}
						if string(c.Status) != expCap.Status {
							t.Errorf("cap %s status mismatch: got %s, want %s", c.CapabilityID, c.Status, expCap.Status)
						}
						if expCap.FallbackMode != "" && string(c.FallbackMode) != expCap.FallbackMode {
							t.Errorf("cap %s fallback mismatch: got %s, want %s", c.CapabilityID, c.FallbackMode, expCap.FallbackMode)
						}
						break
					}
				}
				if !found {
					t.Errorf("expected capability %s not found in inspected list", expCap.CapabilityID)
				}
			}

			// Validate actions
			for _, act := range s.ActionsToTest {
				allowed, _, err := CanExecuteAction(conn, act.Action)
				if allowed != act.ExpectedAllowed {
					t.Errorf("action %s allowed mismatch: got %v, want %v (err: %v)", act.Action, allowed, act.ExpectedAllowed, err)
				}
				if act.ExpectedError != "" {
					if err == nil || err.Error() != act.ExpectedError {
						t.Errorf("action %s error mismatch: got %v, want %s", act.Action, err, act.ExpectedError)
					}
				}
			}
		})
	}
}

func TestLinkedIn_FailClosedActionGating_AT010(t *testing.T) {
	conn := &LinkedInConnection{
		ConnectionID:  "conn-test-gate",
		UserID:        "user-1",
		GrantedScopes: []string{"openid", "profile", "email"},
		TokenStatus:   "active",
	}

	// 1. Direct messaging must fail closed with partner requirement error
	allowed, _, err := CanExecuteAction(conn, "send_direct_message")
	if allowed || err != ErrDirectMessagingPartnerRequired {
		t.Errorf("expected direct messaging to fail closed with partner required error, got allowed=%v, err=%v", allowed, err)
	}

	// 2. Automated easy apply must fail closed with unsupported platform error
	allowed, _, err = CanExecuteAction(conn, "automated_easy_apply")
	if allowed || err != ErrEasyApplyLiveUnsupported {
		t.Errorf("expected automated easy apply to fail closed with unsupported error, got allowed=%v, err=%v", allowed, err)
	}

	// 3. Unrecognized action must fail closed
	allowed, _, err = CanExecuteAction(conn, "unknown_wildcard_action")
	if allowed || err != ErrCapabilityNotGranted {
		t.Errorf("expected unknown action to fail closed with capability not granted, got allowed=%v, err=%v", allowed, err)
	}
}

func TestLinkedIn_ExpiredTokenReauth_AT016(t *testing.T) {
	past := time.Now().UTC().Add(-2 * time.Hour)
	conn := &LinkedInConnection{
		ConnectionID:  "conn-expired",
		UserID:        "user-2",
		GrantedScopes: []string{"openid", "profile", "email", "w_member_social"},
		TokenStatus:   "expired",
		ExpiresAt:     &past,
	}

	allowed, _, err := CanExecuteAction(conn, "publish_post")
	if allowed || err != ErrTokenExpiredOrRevoked {
		t.Errorf("expected expired token to fail closed with reauth error, got allowed=%v, err=%v", allowed, err)
	}

	// Even sign-in checks fail when token is expired
	allowed, _, err = CanExecuteAction(conn, "sign_in")
	if allowed || err != ErrTokenExpiredOrRevoked {
		t.Errorf("expected sign_in to fail on expired token, got allowed=%v, err=%v", allowed, err)
	}
}

func TestLinkedIn_ServiceIntegration(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()
	userID := "usr-svc-test"

	// 1. Connect Account
	conn, err := service.ConnectAccount(
		ctx,
		userID,
		"urn:li:person:svc_123",
		"Test User",
		"test.user@example.com",
		[]string{"openid", "profile", "email", "w_member_social"},
		"access_token_super_secret_12345678",
		24,
	)
	if err != nil {
		t.Fatalf("failed to connect account: %v", err)
	}
	if conn.MaskedToken != "tok_li_****_5678" {
		t.Errorf("expected masked token 'tok_li_****_5678', got '%s'", conn.MaskedToken)
	}
	if conn.Status != ConnPartiallyAuthorized {
		t.Errorf("expected partial authorization, got %s", conn.Status)
	}

	// 2. Inspect Capabilities
	inspected, err := service.InspectCapabilities(ctx, userID)
	if err != nil {
		t.Fatalf("failed to inspect capabilities: %v", err)
	}
	if len(inspected.Capabilities) != 6 {
		t.Errorf("expected 6 capabilities, got %d", len(inspected.Capabilities))
	}

	// 3. Check Action Permission
	allowed, capObj, err := service.CheckActionPermission(ctx, userID, "publish_post")
	if err != nil || !allowed || capObj == nil || !capObj.IsGranted {
		t.Errorf("expected publish_post to be allowed, got allowed=%v, err=%v", allowed, err)
	}

	allowed, _, err = service.CheckActionPermission(ctx, userID, "send_direct_message")
	if allowed || err != ErrDirectMessagingPartnerRequired {
		t.Errorf("expected send_direct_message to fail closed, got allowed=%v, err=%v", allowed, err)
	}

	// 4. Disconnect Account
	if err := service.DisconnectAccount(ctx, userID); err != nil {
		t.Fatalf("failed to disconnect account: %v", err)
	}

	_, err = service.GetConnection(ctx, userID)
	if err != ErrConnectionNotFound {
		t.Errorf("expected ErrConnectionNotFound after disconnect, got %v", err)
	}
}
