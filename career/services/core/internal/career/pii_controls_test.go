package career

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

type PIIFixtureSuite struct {
	Description string `json:"description"`
	Scenarios   []struct {
		ID                            string   `json:"id"`
		Name                          string   `json:"name"`
		InputText                     string   `json:"input_text"`
		ExpectedTokensPresent         []string `json:"expected_tokens_present"`
		ForbiddenRawSubstrings        []string `json:"forbidden_raw_substrings"`
		PreservedContentSubstrings    []string `json:"preserved_content_substrings"`
		ExpectedDisclaimerContains   string   `json:"expected_disclaimer_contains"`
		ProviderID                    string   `json:"provider_id"`
		ProviderName                  string   `json:"provider_name"`
		UserConsentGranted            bool     `json:"user_consent_granted"`
		ExpectedBlocked               bool     `json:"expected_blocked"`
		ExpectedInjectionDetected     bool     `json:"expected_injection_detected"`
		ExpectedSanitized             bool     `json:"expected_sanitized"`
		InitialConsent                bool     `json:"initial_consent"`
		RevokeAction                  bool     `json:"revoke_action"`
		ExpectedStatusAfterRevocation string   `json:"expected_status_after_revocation"`
		SubsequentDispatchBlocked     bool     `json:"subsequent_dispatch_blocked"`
	} `json:"scenarios"`
}

func TestPIIControls_FixtureSuiteEvaluation(t *testing.T) {
	data, err := os.ReadFile("../../testdata/fixtures/forms/pii_controls.json")
	if err != nil {
		t.Fatalf("failed to read pii_controls.json fixture: %v", err)
	}

	var suite PIIFixtureSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("failed to parse pii_controls fixture: %v", err)
	}

	if len(suite.Scenarios) < 4 {
		t.Fatalf("expected at least 4 test scenarios in fixture, got %d", len(suite.Scenarios))
	}

	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	for _, sc := range suite.Scenarios {
		t.Run(sc.ID, func(t *testing.T) {
			switch sc.ID {
			case "full_pii_redaction_and_rehydration":
				identities := CandidateIdentities{
					FullName:     "Jane Doe",
					Emails:       []string{"jane.doe@example.com"},
					Phones:       []string{"+1-555-0199"},
					Locations:    []string{"San Francisco, CA"},
					Links:        []string{"https://linkedin.com/in/janedoe-cloud"},
					Compensation: []string{"$175,000/yr"},
				}

				res, err := RedactPII(sc.InputText, identities, DefaultPIIRedactionPolicy())
				if err != nil {
					t.Fatalf("RedactPII error: %v", err)
				}

				for _, token := range sc.ExpectedTokensPresent {
					if !strings.Contains(res.RedactedText, token) {
						t.Errorf("expected token %s in redacted text: %s", token, res.RedactedText)
					}
				}

				for _, raw := range sc.ForbiddenRawSubstrings {
					if strings.Contains(res.RedactedText, raw) {
						t.Errorf("forbidden raw string %s leaked in redacted text: %s", raw, res.RedactedText)
					}
				}

				for _, pres := range sc.PreservedContentSubstrings {
					if !strings.Contains(res.RedactedText, pres) {
						t.Errorf("expected preserved text %s missing from redacted text", pres)
					}
				}

				if !strings.Contains(res.Disclaimer, sc.ExpectedDisclaimerContains) {
					t.Errorf("disclaimer must contain %q, got: %s", sc.ExpectedDisclaimerContains, res.Disclaimer)
				}

				// Test roundtrip rehydration
				rehydrated := RehydrateText(res.RedactedText, res.TokenMap)
				if !strings.Contains(rehydrated, "Jane Doe") || !strings.Contains(rehydrated, "jane.doe@example.com") {
					t.Errorf("rehydration failed to restore original values: %s", rehydrated)
				}

			case "unconsented_provider_dispatch_block":
				err := service.VerifyProviderConsentForTask(ctx, "usr_unconsented", sc.ProviderID, "resume_tailor")
				if !errors.Is(err, ErrProviderConsentRequired) {
					t.Errorf("expected ErrProviderConsentRequired when no consent granted, got: %v", err)
				}

			case "adversarial_injection_via_pii_field":
				identities := CandidateIdentities{
					FullName: "John Smith",
					Emails:   []string{"john@gmail.com"},
				}
				res, err := RedactPII(sc.InputText, identities, DefaultPIIRedactionPolicy())
				if err != nil {
					t.Fatalf("RedactPII error: %v", err)
				}

				if sc.ExpectedInjectionDetected && len(res.SecurityAlerts) == 0 {
					t.Errorf("expected prompt injection security alert")
				}

				for _, raw := range sc.ForbiddenRawSubstrings {
					if strings.Contains(res.RedactedText, raw) {
						t.Errorf("forbidden raw payload %s leaked in redacted text: %s", raw, res.RedactedText)
					}
				}

			case "consent_revocation_and_key_purge":
				// 1. Grant initial consent
				consent, err := service.GrantProviderConsent(ctx, "usr_revoke_test", sc.ProviderID, sc.ProviderName, []string{"*"}, true, 30)
				if err != nil {
					t.Fatalf("GrantProviderConsent error: %v", err)
				}
				if consent.Status != "active" {
					t.Errorf("expected active consent status, got %s", consent.Status)
				}

				// Verify dispatch allowed
				if err := service.VerifyProviderConsentForTask(ctx, "usr_revoke_test", sc.ProviderID, "resume_tailor"); err != nil {
					t.Fatalf("expected dispatch authorized with active consent, got: %v", err)
				}

				// 2. Revoke consent
				revoked, err := service.RevokeProviderConsent(ctx, "usr_revoke_test", sc.ProviderID)
				if err != nil {
					t.Fatalf("RevokeProviderConsent error: %v", err)
				}
				if revoked.Status != sc.ExpectedStatusAfterRevocation {
					t.Errorf("expected status %s, got %s", sc.ExpectedStatusAfterRevocation, revoked.Status)
				}

				// 3. Verify subsequent dispatch blocked
				err = service.VerifyProviderConsentForTask(ctx, "usr_revoke_test", sc.ProviderID, "resume_tailor")
				if !errors.Is(err, ErrProviderConsentRequired) {
					t.Errorf("expected ErrProviderConsentRequired after revocation, got: %v", err)
				}
			}
		})
	}
}

func TestPIIControls_AES256GCMEncryption(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	originalMap := map[string]string{
		"[CANDIDATE_NAME]": "Alice Wonder",
		"[EMAIL_1]":        "alice@wonderland.io",
		"[PHONE_1]":        "+1-415-555-0144",
		"[LOCATION_1]":     "New York, NY",
	}

	encrypted, err := EncryptTokenMap(originalMap, key)
	if err != nil {
		t.Fatalf("EncryptTokenMap error: %v", err)
	}

	if strings.Contains(encrypted, "Alice Wonder") || strings.Contains(encrypted, "alice@wonderland.io") {
		t.Fatalf("plaintext leaked in encrypted token map ciphertext: %s", encrypted)
	}

	decrypted, err := DecryptTokenMap(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptTokenMap error: %v", err)
	}

	for k, v := range originalMap {
		if decrypted[k] != v {
			t.Errorf("decrypted map mismatch for key %s: expected %s, got %s", k, v, decrypted[k])
		}
	}

	// Invalid key length check
	_, err = EncryptTokenMap(originalMap, []byte("too-short"))
	if !errors.Is(err, ErrInvalidEncryptionKey) {
		t.Errorf("expected ErrInvalidEncryptionKey for short key, got: %v", err)
	}
}

func TestPIIControls_CustomTermsAndTruthInAdvertising(t *testing.T) {
	policy := DefaultPIIRedactionPolicy()
	policy.CustomPIITerms = []string{"Project Titan", "Secret Internal API"}

	rawText := "I was lead architect on Project Titan, integrating Secret Internal API for payments."
	res, err := RedactPII(rawText, CandidateIdentities{}, policy)
	if err != nil {
		t.Fatalf("RedactPII error: %v", err)
	}

	if strings.Contains(res.RedactedText, "Project Titan") {
		t.Errorf("custom term 'Project Titan' was not redacted: %s", res.RedactedText)
	}
	if strings.Contains(res.RedactedText, "Secret Internal API") {
		t.Errorf("custom term 'Secret Internal API' was not redacted: %s", res.RedactedText)
	}

	// Truth-in-advertising disclaimer guarantee (CAR-22)
	if !strings.Contains(res.Disclaimer, "does NOT guarantee complete anonymity") {
		t.Errorf("truth-in-advertising disclaimer missing required no-anonymity guarantee clause")
	}
}

func TestPIIControls_ProfileAutoExtractionIntegration(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	// Setup profile
	profile := &MasterCareerProfile{
		UserID: "usr_auto_pii",
		Contact: ContactInfo{
			FullName: "Bob Builder",
			Email:    "bob@builder.com",
			Phone:    "+1-800-555-1234",
			Location: "London, UK",
		},
	}
	_ = repo.SaveProfile(ctx, profile)

	rawText := "Hello, I am Bob Builder. Reach me at bob@builder.com or +1-800-555-1234 in London, UK."
	res, err := service.RedactCandidatePII(ctx, "usr_auto_pii", rawText, DefaultPIIRedactionPolicy())
	if err != nil {
		t.Fatalf("RedactCandidatePII error: %v", err)
	}

	if strings.Contains(res.RedactedText, "Bob Builder") {
		t.Errorf("candidate name not redacted: %s", res.RedactedText)
	}
	if strings.Contains(res.RedactedText, "bob@builder.com") {
		t.Errorf("candidate email not redacted: %s", res.RedactedText)
	}
	if !strings.Contains(res.RedactedText, "[CANDIDATE_NAME]") {
		t.Errorf("missing [CANDIDATE_NAME] token")
	}
}
