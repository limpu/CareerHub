package career

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	// ErrEvidenceBundleNotFound indicates the requested bundle does not exist.
	ErrEvidenceBundleNotFound = errors.New("evidence bundle not found")

	// ErrWorkspaceAdminAccessDenied is returned when a workspace admin attempts to access private evidence without a grant (AT-011).
	ErrWorkspaceAdminAccessDenied = errors.New("access denied: workspace admin cannot read member private application evidence without explicit grant (AT-011)")

	// ErrAccessDenied indicates unauthorized access by a non-owner actor.
	ErrAccessDenied = errors.New("access denied: unauthorized evidence bundle access")

	// ErrEvidenceIntegrityViolation indicates the bundle checksum does not match its contents.
	ErrEvidenceIntegrityViolation = errors.New("evidence integrity violation: bundle content checksum mismatch")

	// ErrInvalidEvidenceBundle is returned when required evidence fields are missing.
	ErrInvalidEvidenceBundle = errors.New("invalid evidence bundle: missing required fields")
)

// PII and Secret Scrubbing Regex Patterns (CAR-16, AT-011, FND-013)
var (
	bundleBearerTokenRegex = regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9\-\._~\+\/]+=*`)
	bundleCookieTokenRegex = regexp.MustCompile(`(?i)(session|token|auth|cookie)\s*=\s*\\?['"]?[A-Za-z0-9\-\._~\+\/=]+\\?['"]?`)
	bundleSSNRegex         = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	bundleCreditCardRegex  = regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`)
	bundlePasswordRegex    = regexp.MustCompile(`(?i)"password"\s*:\s*"[^"]*"`)
	bundleEmailRegex       = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
)

// JobSnapshot captures the exact job post state at the moment of submission (CAR-16).
type JobSnapshot struct {
	Title              string    `json:"title"`
	Company            string    `json:"company"`
	Location           string    `json:"location"`
	JobType            string    `json:"job_type"`
	RequiredSkills     []string  `json:"required_skills"`
	DirectApplyURL     string    `json:"direct_apply_url"`
	DescriptionSnippet string    `json:"description_snippet"`
	CapturedAt         time.Time `json:"captured_at"`
}

// ResumeArtifactSnapshot preserves the exact resume version and checksum used (CAR-16).
type ResumeArtifactSnapshot struct {
	ResumeID        string `json:"resume_id"`
	Version         string `json:"version"`
	Format          string `json:"format"`
	ContentChecksum string `json:"content_checksum"`
	DocumentPath    string `json:"document_path"`
	ContentSnippet  string `json:"content_snippet,omitempty"`
}

// CoverLetterSnapshot preserves the exact cover letter content and checksum (CAR-16).
type CoverLetterSnapshot struct {
	CoverLetterID string `json:"cover_letter_id"`
	Content       string `json:"content"`
	Checksum      string `json:"checksum"`
}

// SubmittedAnswer represents an individual form question-answer pair (CAR-16).
type SubmittedAnswer struct {
	QuestionID    string `json:"question_id"`
	QuestionText  string `json:"question_text"`
	AnswerValue   string `json:"answer_value"`
	FieldCategory string `json:"field_category"`
}

// ScrubbedArtifactItem holds sanitized supporting artifacts like HTTP payloads or DOM captures (AT-011).
type ScrubbedArtifactItem struct {
	ArtifactID       string `json:"artifact_id"`
	Name             string `json:"name"`
	MimeType         string `json:"mime_type"`
	ScrubbedContent  string `json:"scrubbed_content"`
	OriginalChecksum string `json:"original_checksum"`
}

// ApplicationEvidenceBundle is the complete, immutable evidence bundle for an application (CAR-16, AT-005, AT-011).
type ApplicationEvidenceBundle struct {
	BundleID            string                  `json:"bundle_id"`
	ApplicationRecordID string                  `json:"application_record_id"`
	UserID              string                  `json:"user_id"`
	JobSnapshot         JobSnapshot             `json:"job_snapshot"`
	ResumeArtifact      ResumeArtifactSnapshot  `json:"resume_artifact"`
	CoverLetterArtifact *CoverLetterSnapshot    `json:"cover_letter_artifact,omitempty"`
	QuestionAnswers     []SubmittedAnswer       `json:"question_answers"`
	SubmissionTimestamp time.Time               `json:"submission_timestamp"`
	ConfirmationType    AppliedVerificationType `json:"confirmation_type"`
	ProviderReference   string                  `json:"provider_reference"`
	ScrubbedArtifacts   []ScrubbedArtifactItem  `json:"scrubbed_artifacts,omitempty"`
	PrivacyTier         string                  `json:"privacy_tier"` // "owner_private"
	IntegrityChecksum   string                  `json:"integrity_checksum"`
	CreatedAt           time.Time               `json:"created_at"`
}

// CanonicalBundlePayload is used for deterministic SHA-256 hashing.
type CanonicalBundlePayload struct {
	BundleID            string                  `json:"bundle_id"`
	ApplicationRecordID string                  `json:"application_record_id"`
	UserID              string                  `json:"user_id"`
	JobSnapshot         JobSnapshot             `json:"job_snapshot"`
	ResumeArtifact      ResumeArtifactSnapshot  `json:"resume_artifact"`
	CoverLetterArtifact *CoverLetterSnapshot    `json:"cover_letter_artifact,omitempty"`
	QuestionAnswers     []SubmittedAnswer       `json:"question_answers"`
	SubmissionTimestamp string                  `json:"submission_timestamp"`
	ConfirmationType    AppliedVerificationType `json:"confirmation_type"`
	ProviderReference   string                  `json:"provider_reference"`
	ScrubbedArtifacts   []ScrubbedArtifactItem  `json:"scrubbed_artifacts,omitempty"`
	PrivacyTier         string                  `json:"privacy_tier"`
}

// ScrubRawArtifactContent redacts PII, tokens, and credentials from raw evidence strings (AT-011).
func ScrubRawArtifactContent(raw string) string {
	if raw == "" {
		return ""
	}

	sanitized := raw
	sanitized = bundleBearerTokenRegex.ReplaceAllString(sanitized, "Bearer [REDACTED_BEARER_TOKEN]")
	sanitized = bundleCookieTokenRegex.ReplaceAllString(sanitized, "$1=\"[REDACTED_COOKIE]\"")
	sanitized = bundlePasswordRegex.ReplaceAllString(sanitized, `"password":"[REDACTED_PASSWORD]"`)
	sanitized = bundleSSNRegex.ReplaceAllString(sanitized, "[REDACTED_SSN]")
	sanitized = bundleCreditCardRegex.ReplaceAllString(sanitized, "[REDACTED_CREDIT_CARD]")

	return sanitized
}

// ComputeEvidenceChecksum calculates the deterministic SHA-256 checksum over the canonical payload (CAR-16).
func ComputeEvidenceChecksum(bundle *ApplicationEvidenceBundle) (string, error) {
	if bundle == nil {
		return "", errors.New("bundle cannot be nil")
	}

	canonical := CanonicalBundlePayload{
		BundleID:            bundle.BundleID,
		ApplicationRecordID: bundle.ApplicationRecordID,
		UserID:              bundle.UserID,
		JobSnapshot:         bundle.JobSnapshot,
		ResumeArtifact:      bundle.ResumeArtifact,
		CoverLetterArtifact: bundle.CoverLetterArtifact,
		QuestionAnswers:     bundle.QuestionAnswers,
		SubmissionTimestamp: bundle.SubmissionTimestamp.UTC().Format(time.RFC3339),
		ConfirmationType:    bundle.ConfirmationType,
		ProviderReference:   bundle.ProviderReference,
		ScrubbedArtifacts:   bundle.ScrubbedArtifacts,
		PrivacyTier:         bundle.PrivacyTier,
	}

	data, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("failed to marshal canonical bundle payload: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// VerifyEvidenceIntegrity checks whether the bundle has been altered since checksum calculation.
func VerifyEvidenceIntegrity(bundle *ApplicationEvidenceBundle) bool {
	if bundle == nil || bundle.IntegrityChecksum == "" {
		return false
	}

	calculated, err := ComputeEvidenceChecksum(bundle)
	if err != nil {
		return false
	}

	return strings.EqualFold(calculated, bundle.IntegrityChecksum)
}

// AuthorizeEvidenceAccess evaluates access control for an evidence bundle adhering to AT-011 and REQ-018.
// Rule: A workspace admin cannot read member private application evidence without an explicit grant.
func AuthorizeEvidenceAccess(bundle *ApplicationEvidenceBundle, requestingUserID string, isWorkspaceAdmin bool, hasExplicitGrant bool) error {
	if bundle == nil {
		return ErrEvidenceBundleNotFound
	}

	// Case 1: Owner of the bundle has unconditional full read access
	if requestingUserID == bundle.UserID {
		return nil
	}

	// Case 2: Workspace Admin requesting access (AT-011, REQ-018)
	if isWorkspaceAdmin {
		if hasExplicitGrant {
			return nil // Explicit permission grant provided
		}
		return ErrWorkspaceAdminAccessDenied
	}

	// Case 3: Other third-party/workspace member without permission
	return ErrAccessDenied
}

// AssembleEvidenceBundle creates a validated and cryptographically sealed ApplicationEvidenceBundle.
func AssembleEvidenceBundle(
	bundleID string,
	appID string,
	userID string,
	jobSnapshot JobSnapshot,
	resumeArtifact ResumeArtifactSnapshot,
	coverLetter *CoverLetterSnapshot,
	answers []SubmittedAnswer,
	submissionTime time.Time,
	confirmType AppliedVerificationType,
	providerRef string,
	rawArtifacts []ScrubbedArtifactItem,
) (*ApplicationEvidenceBundle, error) {
	if bundleID == "" || appID == "" || userID == "" {
		return nil, ErrInvalidEvidenceBundle
	}
	if jobSnapshot.Title == "" || jobSnapshot.Company == "" {
		return nil, fmt.Errorf("%w: job snapshot requires title and company", ErrInvalidEvidenceBundle)
	}
	if resumeArtifact.ContentChecksum == "" {
		return nil, fmt.Errorf("%w: resume artifact requires content checksum", ErrInvalidEvidenceBundle)
	}

	// Scrub raw artifacts (AT-011)
	sanitizedArtifacts := make([]ScrubbedArtifactItem, len(rawArtifacts))
	for i, art := range rawArtifacts {
		origHash := art.OriginalChecksum
		if origHash == "" {
			h := sha256.Sum256([]byte(art.ScrubbedContent))
			origHash = hex.EncodeToString(h[:])
		}
		sanitizedArtifacts[i] = ScrubbedArtifactItem{
			ArtifactID:       art.ArtifactID,
			Name:             art.Name,
			MimeType:         art.MimeType,
			ScrubbedContent:  ScrubRawArtifactContent(art.ScrubbedContent),
			OriginalChecksum: origHash,
		}
	}

	bundle := &ApplicationEvidenceBundle{
		BundleID:            bundleID,
		ApplicationRecordID: appID,
		UserID:              userID,
		JobSnapshot:         jobSnapshot,
		ResumeArtifact:      resumeArtifact,
		CoverLetterArtifact: coverLetter,
		QuestionAnswers:     answers,
		SubmissionTimestamp: submissionTime,
		ConfirmationType:    confirmType,
		ProviderReference:   providerRef,
		ScrubbedArtifacts:   sanitizedArtifacts,
		PrivacyTier:         "owner_private",
		CreatedAt:           time.Now().UTC(),
	}

	checksum, err := ComputeEvidenceChecksum(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to seal bundle integrity: %w", err)
	}
	bundle.IntegrityChecksum = checksum

	return bundle, nil
}
