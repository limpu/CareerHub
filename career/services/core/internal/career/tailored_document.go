package career

import (
	"errors"
	"time"
)

var (
	ErrJobTargetRequired       = errors.New("job target title, company, and description are required")
	ErrTailoredResumeNotFound  = errors.New("tailored resume not found")
	ErrCoverLetterNotFound     = errors.New("cover letter not found")
	ErrAlreadyApproved         = errors.New("document has already been approved and is immutable")
	ErrInvalidApprovalToken    = errors.New("invalid or tampered approval token")
)

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending_approval"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

// JobTarget defines the target role and employer requirements for tailoring.
type JobTarget struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Company        string   `json:"company"`
	Location       string   `json:"location,omitempty"`
	Description    string   `json:"description"`
	RequiredSkills []string `json:"required_skills,omitempty"`
	Keywords       []string `json:"keywords,omitempty"`
	SourceURL      string   `json:"source_url,omitempty"`
}

// ResumeDiffSummary details exact adaptations made relative to the Master Profile.
// Invariant REQ-016 & AT-003: Unmatched requirements are transparently listed and NEVER fabricated.
type ResumeDiffSummary struct {
	EmphasizedSkills         []string `json:"emphasized_skills"`
	PrioritizedHighlights    []string `json:"prioritized_highlights"`
	UnmatchedJobRequirements []string `json:"unmatched_job_requirements"`
	ExcludedIrrelevantPoints int      `json:"excluded_irrelevant_points"`
	TotalConfirmedFactsUsed  int      `json:"total_confirmed_facts_used"`
}

// TailoredResume represents a job-specific, human-reviewed, ATS-compliant resume.
type TailoredResume struct {
	ID                  string             `json:"id"`
	UserID              string             `json:"user_id"`
	JobID               string             `json:"job_id"`
	JobTarget           JobTarget          `json:"job_target"`
	Format              MasterResumeFormat `json:"format"`
	Template            ResumeTemplateType `json:"template"`
	OriginalProfileID   string             `json:"original_profile_id"`
	TailoredProfile     MasterCareerProfile`json:"tailored_profile"`
	ByteContent         []byte             `json:"-"`
	ContentLength       int                `json:"content_length"`
	ChecksumSHA256      string             `json:"checksum_sha256"`
	FileName            string             `json:"file_name"`
	MimeType            string             `json:"mime_type"`
	DiffSummary         ResumeDiffSummary  `json:"diff_summary"`
	ApprovalStatus      ApprovalStatus     `json:"approval_status"`
	ApprovalToken       string             `json:"approval_token"`
	ParseBackVerified   bool               `json:"parse_back_verified"`
	CreatedAt           time.Time          `json:"created_at"`
	ApprovedAt          *time.Time         `json:"approved_at,omitempty"`
}

// CoverLetter represents a targeted, fact-grounded cover letter document.
type CoverLetter struct {
	ID                string             `json:"id"`
	UserID            string             `json:"user_id"`
	JobID             string             `json:"job_id"`
	JobTarget         JobTarget          `json:"job_target"`
	RecipientName     string             `json:"recipient_name,omitempty"`
	Format            MasterResumeFormat `json:"format"`
	Salutation        string             `json:"salutation"`
	OpeningParagraph  string             `json:"opening_paragraph"`
	BodyParagraphs    []string           `json:"body_paragraphs"`
	ClosingParagraph  string             `json:"closing_paragraph"`
	Signoff           string             `json:"signoff"`
	FullText          string             `json:"full_text"`
	ByteContent       []byte             `json:"-"`
	ContentLength     int                `json:"content_length"`
	ChecksumSHA256    string             `json:"checksum_sha256"`
	FileName          string             `json:"file_name"`
	MimeType          string             `json:"mime_type"`
	ApprovalStatus    ApprovalStatus     `json:"approval_status"`
	ApprovalToken     string             `json:"approval_token"`
	ParseBackVerified bool               `json:"parse_back_verified"`
	CreatedAt         time.Time          `json:"created_at"`
	ApprovedAt        *time.Time         `json:"approved_at,omitempty"`
}
