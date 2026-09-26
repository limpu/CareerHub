package career

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

var (
	ErrProviderConsentRequired = errors.New("explicit model consent is required before dispatching private data to external AI provider (REQ-016, REQ-023)")
	ErrConsentAlreadyRevoked   = errors.New("consent for provider is already revoked")
	ErrInvalidEncryptionKey   = errors.New("encryption key must be exactly 32 bytes for AES-256")
)

// Mandatory Truth-in-Advertising Notice per CAR-22
const NoCompleteAnonymityDisclaimer = "Tested synthetic token redaction minimizes PII exposure, but does NOT guarantee complete anonymity. Stylistic fingerprints, highly unique project descriptions, or niche employer combinations may allow inference. Complete anonymity is impossible when evaluating career history."

// PIIMinimizationLevel defines how aggressively candidate data is redacted.
type PIIMinimizationLevel string

const (
	MinimizationFullRedaction    PIIMinimizationLevel = "full_redaction"
	MinimizationPartialMask      PIIMinimizationLevel = "partial_mask"
	MinimizationStrictSkillsOnly PIIMinimizationLevel = "strict_skills_only"
)

// PIIRedactionPolicy specifies granular field-level redaction controls.
type PIIRedactionPolicy struct {
	MinimizationLevel  PIIMinimizationLevel `json:"minimization_level"`
	RedactNames        bool                 `json:"redact_names"`
	RedactEmails       bool                 `json:"redact_emails"`
	RedactPhones       bool                 `json:"redact_phones"`
	RedactLinks        bool                 `json:"redact_links"`
	RedactLocations    bool                 `json:"redact_locations"`
	RedactCompensation bool                 `json:"redact_compensation"`
	CustomPIITerms     []string             `json:"custom_pii_terms,omitempty"`
}

// DefaultPIIRedactionPolicy provides a safe-by-default redaction profile.
func DefaultPIIRedactionPolicy() PIIRedactionPolicy {
	return PIIRedactionPolicy{
		MinimizationLevel:  MinimizationFullRedaction,
		RedactNames:        true,
		RedactEmails:       true,
		RedactPhones:       true,
		RedactLinks:        true,
		RedactLocations:    true,
		RedactCompensation: true,
		CustomPIITerms:     make([]string, 0),
	}
}

// CandidateIdentities encapsulates verified personal identifiers to be anonymized.
type CandidateIdentities struct {
	FullName     string   `json:"full_name"`
	Emails       []string `json:"emails"`
	Phones       []string `json:"phones"`
	Locations    []string `json:"locations"`
	Links        []string `json:"links"`
	Compensation []string `json:"compensation"`
}

// PIIRedactionResult captures the anonymized text and the reversible token vault.
type PIIRedactionResult struct {
	OriginalLength     int               `json:"original_length"`
	RedactedLength     int               `json:"redacted_length"`
	RedactedText       string            `json:"redacted_text"`
	TokenMap           map[string]string `json:"token_map"` // token -> original plaintext
	DetectedItemsCount int               `json:"detected_items_count"`
	DetectedCategories map[string]int    `json:"detected_categories"`
	SecurityAlerts     []string          `json:"security_alerts,omitempty"`
	Disclaimer         string            `json:"disclaimer"`
}

// UserProviderConsent records explicit user choice per LLM provider (REQ-023, CAR-22).
type UserProviderConsent struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	ProviderID           string     `json:"provider_id"` // "openai", "anthropic", "google_gemini", "local_ollama"
	ProviderName         string     `json:"provider_name"`
	Status               string     `json:"status"` // "active", "revoked"
	AllowedTasks         []string   `json:"allowed_tasks"`
	ZeroTrainingAffirmed bool       `json:"zero_training_affirmed"`
	ConsentedAt          time.Time  `json:"consented_at"`
	RevokedAt            *time.Time `json:"revoked_at,omitempty"`
	RetentionDays        int        `json:"retention_days"`
}

// Regular expressions for common PII patterns
var (
	piiEmailRegex        = regexp.MustCompile(`(?i)[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	piiPhoneRegex        = regexp.MustCompile(`(?:\+?\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`)
	urlRegex          = regexp.MustCompile(`https?://[^\s)]+`)
	compensationRegex = regexp.MustCompile(`\$\s?\d{1,3}(?:,\d{3})*(?:\.\d{2})?(?:\s*(?:k|K|yr|/yr|year|annual|annually|month|/mo|hr|/hr))?`)
	injectionRegex    = regexp.MustCompile(`(?i)<system_override>[\s\S]*?</system_override>|SYSTEM:\s*|ignore previous instructions`)
)

// RedactPII replaces sensitive personal information with synthetic tokens (SRC-C3, AT-019).
func RedactPII(text string, identities CandidateIdentities, policy PIIRedactionPolicy) (*PIIRedactionResult, error) {
	originalLen := len(text)
	redacted := text
	tokenMap := make(map[string]string)
	categories := make(map[string]int)
	securityAlerts := make([]string, 0)

	// 1. Adversarial Prompt Injection Defense (AT-019)
	if injectionRegex.MatchString(redacted) {
		securityAlerts = append(securityAlerts, "Prompt injection attempt detected and neutralized in candidate text (AT-019)")
		redacted = injectionRegex.ReplaceAllString(redacted, "[SANITIZED_INJECTION_PAYLOAD]")
	}

	// 2. Redact Explicit Full Name
	if policy.RedactNames && strings.TrimSpace(identities.FullName) != "" {
		token := "[CANDIDATE_NAME]"
		if strings.Contains(redacted, identities.FullName) {
			tokenMap[token] = identities.FullName
			redacted = strings.ReplaceAll(redacted, identities.FullName, token)
			categories["name"]++
		}
		// Also scan case-insensitively if not found directly
		nameLower := strings.ToLower(identities.FullName)
		if strings.Contains(strings.ToLower(redacted), nameLower) && !strings.Contains(redacted, token) {
			tokenMap[token] = identities.FullName
			re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(identities.FullName))
			redacted = re.ReplaceAllString(redacted, token)
			categories["name"]++
		}
	}

	// 3. Redact Emails (Identities + Pattern)
	if policy.RedactEmails {
		emailIdx := 1
		// Check confirmed candidate emails
		for _, em := range identities.Emails {
			if strings.TrimSpace(em) == "" {
				continue
			}
			token := fmt.Sprintf("[EMAIL_%d]", emailIdx)
			if strings.Contains(redacted, em) {
				tokenMap[token] = em
				redacted = strings.ReplaceAll(redacted, em, token)
				categories["email"]++
				emailIdx++
			}
		}
		// Pattern search for any remaining emails
		matches := piiEmailRegex.FindAllString(redacted, -1)
		for _, m := range matches {
			if strings.HasPrefix(m, "[") && strings.HasSuffix(m, "]") {
				continue
			}
			token := fmt.Sprintf("[EMAIL_%d]", emailIdx)
			tokenMap[token] = m
			redacted = strings.ReplaceAll(redacted, m, token)
			categories["email"]++
			emailIdx++
		}
	}

	// 4. Redact Phones (Identities + Pattern)
	if policy.RedactPhones {
		phoneIdx := 1
		for _, ph := range identities.Phones {
			if strings.TrimSpace(ph) == "" {
				continue
			}
			token := fmt.Sprintf("[PHONE_%d]", phoneIdx)
			if strings.Contains(redacted, ph) {
				tokenMap[token] = ph
				redacted = strings.ReplaceAll(redacted, ph, token)
				categories["phone"]++
				phoneIdx++
			}
		}
		matches := piiPhoneRegex.FindAllString(redacted, -1)
		for _, m := range matches {
			if strings.HasPrefix(m, "[") && strings.HasSuffix(m, "]") {
				continue
			}
			token := fmt.Sprintf("[PHONE_%d]", phoneIdx)
			tokenMap[token] = m
			redacted = strings.ReplaceAll(redacted, m, token)
			categories["phone"]++
			phoneIdx++
		}
	}

	// 5. Redact Social & Profile Links
	if policy.RedactLinks {
		linkIdx := 1
		for _, lk := range identities.Links {
			if strings.TrimSpace(lk) == "" {
				continue
			}
			token := fmt.Sprintf("[LINK_%d]", linkIdx)
			if strings.Contains(redacted, lk) {
				tokenMap[token] = lk
				redacted = strings.ReplaceAll(redacted, lk, token)
				categories["link"]++
				linkIdx++
			}
		}
		matches := urlRegex.FindAllString(redacted, -1)
		for _, m := range matches {
			if strings.HasPrefix(m, "[") && strings.HasSuffix(m, "]") {
				continue
			}
			token := fmt.Sprintf("[LINK_%d]", linkIdx)
			tokenMap[token] = m
			redacted = strings.ReplaceAll(redacted, m, token)
			categories["link"]++
			linkIdx++
		}
	}

	// 6. Redact Locations
	if policy.RedactLocations {
		locIdx := 1
		for _, loc := range identities.Locations {
			if strings.TrimSpace(loc) == "" {
				continue
			}
			token := fmt.Sprintf("[LOCATION_%d]", locIdx)
			if strings.Contains(redacted, loc) {
				tokenMap[token] = loc
				redacted = strings.ReplaceAll(redacted, loc, token)
				categories["location"]++
				locIdx++
			}
		}
	}

	// 7. Redact Compensation Figures
	if policy.RedactCompensation {
		compIdx := 1
		for _, comp := range identities.Compensation {
			if strings.TrimSpace(comp) == "" {
				continue
			}
			token := fmt.Sprintf("[COMPENSATION_%d]", compIdx)
			if strings.Contains(redacted, comp) {
				tokenMap[token] = comp
				redacted = strings.ReplaceAll(redacted, comp, token)
				categories["compensation"]++
				compIdx++
			}
		}
		matches := compensationRegex.FindAllString(redacted, -1)
		for _, m := range matches {
			if strings.HasPrefix(m, "[") && strings.HasSuffix(m, "]") {
				continue
			}
			token := fmt.Sprintf("[COMPENSATION_%d]", compIdx)
			tokenMap[token] = m
			redacted = strings.ReplaceAll(redacted, m, token)
			categories["compensation"]++
			compIdx++
		}
	}

	// 8. Custom PII Terms
	for _, term := range policy.CustomPIITerms {
		trimmed := strings.TrimSpace(term)
		if trimmed != "" && strings.Contains(redacted, trimmed) {
			token := fmt.Sprintf("[CUSTOM_SECRET_%d]", len(tokenMap)+1)
			tokenMap[token] = trimmed
			redacted = strings.ReplaceAll(redacted, trimmed, token)
			categories["custom"]++
		}
	}

	totalCount := 0
	for _, cnt := range categories {
		totalCount += cnt
	}

	return &PIIRedactionResult{
		OriginalLength:     originalLen,
		RedactedLength:     len(redacted),
		RedactedText:       redacted,
		TokenMap:           tokenMap,
		DetectedItemsCount: totalCount,
		DetectedCategories: categories,
		SecurityAlerts:     securityAlerts,
		Disclaimer:         NoCompleteAnonymityDisclaimer,
	}, nil
}

// RehydrateText restores synthetic tokens to original values after LLM processing (SRC-C3).
func RehydrateText(anonymizedText string, tokenMap map[string]string) string {
	result := anonymizedText
	for token, original := range tokenMap {
		result = strings.ReplaceAll(result, token, original)
	}
	return result
}

// EncryptTokenMap encrypts the token map using AES-256-GCM for secure storage (FND-013).
func EncryptTokenMap(tokenMap map[string]string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidEncryptionKey
	}

	plaintext, err := json.Marshal(tokenMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token map: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptTokenMap decrypts a base64-encoded AES-256-GCM token map (FND-013).
func DecryptTokenMap(encodedCiphertext string, key []byte) (map[string]string, error) {
	if len(key) != 32 {
		return nil, ErrInvalidEncryptionKey
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	var tokenMap map[string]string
	if err := json.Unmarshal(plaintext, &tokenMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token map: %w", err)
	}

	return tokenMap, nil
}
