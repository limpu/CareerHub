package career

import (
	"time"
)

type ApplicationStatus string

const (
	StatusDraft               ApplicationStatus = "draft"
	StatusPreparing           ApplicationStatus = "preparing"
	StatusNeedsInput          ApplicationStatus = "needs_input"
	StatusReadyForReview      ApplicationStatus = "ready_for_review"
	StatusApproved            ApplicationStatus = "approved"
	StatusSubmitting          ApplicationStatus = "submitting"
	StatusApplied             ApplicationStatus = "applied"
	StatusRecruiterResponse   ApplicationStatus = "recruiter_response"
	StatusInterview           ApplicationStatus = "interviewing"
	StatusOffer               ApplicationStatus = "offer"
	StatusRejected            ApplicationStatus = "rejected"
	StatusWithdrawn           ApplicationStatus = "withdrawn"
	StatusNeedsConfirmation   ApplicationStatus = "needs_confirmation"
	StatusFailedBeforeSubmit  ApplicationStatus = "failed_before_submit"
)

type ConfirmationType string

const (
	ConfirmationProviderConfirmed ConfirmationType = "provider_confirmed"
	ConfirmationUserConfirmed     ConfirmationType = "user_confirmed"
	ConfirmationUnconfirmed       ConfirmationType = "unconfirmed"
)

type SourceProvenance string

const (
	ProvenanceUploadPDF      SourceProvenance = "upload_pdf"
	ProvenanceUploadDOCX     SourceProvenance = "upload_docx"
	ProvenanceManualEntry    SourceProvenance = "manual_entry"
	ProvenanceLinkedInImport SourceProvenance = "linkedin_import"
)

type PrivacyTier string

const (
	PrivacyPublic       PrivacyTier = "public"
	PrivacyDelegated    PrivacyTier = "delegated"
	PrivacyOwnerPrivate PrivacyTier = "owner_private"
)

type ExtractionStatus string

const (
	ExtractionPending   ExtractionStatus = "pending"
	ExtractionCompleted ExtractionStatus = "completed"
	ExtractionConfirmed ExtractionStatus = "confirmed"
	ExtractionRejected  ExtractionStatus = "rejected"
)

// ProfileFactItem represents a granular, verified or draft career fact.
// Invariant AT-001: PDF upload, DOCX upload, and manual entry map into this exact schema.
// Invariant AT-003: Unknown facts remain empty/unconfirmed rather than fabricated.
type ProfileFactItem struct {
	ID               string           `json:"id"`
	UserID           string           `json:"user_id"`
	Category         string           `json:"category"` // 'contact', 'summary', 'experience', 'education', 'skill', 'project', 'certification'
	Title            string           `json:"title"`
	Organization     string           `json:"organization,omitempty"`
	StartDate        *time.Time       `json:"start_date,omitempty"`
	EndDate          *time.Time       `json:"end_date,omitempty"`
	IsCurrent        bool             `json:"is_current"`
	Details          []string         `json:"details"`
	Confidence       float64          `json:"confidence"` // 0.0 to 1.0
	Confirmed        bool             `json:"confirmed"`  // false during draft review, true once user confirms
	SourceProvenance SourceProvenance `json:"source_provenance"`
	SourceLocation   string           `json:"source_location,omitempty"` // page/section pointer
	PrivacyTier      PrivacyTier      `json:"privacy_tier"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// ResumeExtractionDraft represents the staged extraction review artifact.
// The user reviews the draft, adjusts values, and confirms it (REQ-002, AT-001).
type ResumeExtractionDraft struct {
	ID              string            `json:"id"`
	UserID          string            `json:"user_id"`
	UploadID        string            `json:"upload_id"`
	FileName        string            `json:"file_name"`
	MimeType        string            `json:"mime_type"`
	Status          ExtractionStatus  `json:"status"`
	RawTextPreview  string            `json:"raw_text_preview"`
	WordCount       int               `json:"word_count"`
	ConfidenceScore float64           `json:"confidence_score"`
	ExtractedFacts  []ProfileFactItem `json:"extracted_facts"`
	CreatedAt       time.Time         `json:"created_at"`
	ConfirmedAt     *time.Time        `json:"confirmed_at,omitempty"`
}

// JobApplication tracks a job application lifecycle (REQ-005, AT-005).
type JobApplication struct {
	ID               string            `json:"id"`
	UserID           string            `json:"user_id"`
	JobTitle         string            `json:"job_title"`
	CompanyName      string            `json:"company_name"`
	JobURL           string            `json:"job_url"`
	Status           ApplicationStatus `json:"status"`
	ConfirmationType ConfirmationType  `json:"confirmation_type"`
	AppliedDate      *time.Time        `json:"applied_date,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}
