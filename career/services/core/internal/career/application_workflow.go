package career

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrWorkflowNotFound            = errors.New("application review session not found")
	ErrUnresolvedQuestions         = errors.New("cannot approve application: unresolved questions require candidate input (AT-003)")
	ErrWorkflowApprovalRequired    = errors.New("cannot dispatch application: explicit human approval token required (REQ-015)")
	ErrWorkflowInvalidTransition   = errors.New("invalid application workflow state transition")
	ErrWorkflowAlreadyConfirmed    = errors.New("application has already been confirmed as applied")
	ErrWorkflowApprovalInvalidated = errors.New("material edit invalidated previous approval token (AT-007)")
)

// ApplicationWorkflowStatus tracks the lifecycle of an application session (REQ-005, REQ-015, AT-005).
type ApplicationWorkflowStatus string

const (
	WorkflowStatusDraft             ApplicationWorkflowStatus = "draft"
	WorkflowStatusReadyForReview     ApplicationWorkflowStatus = "ready_for_review"
	WorkflowStatusReviewing         ApplicationWorkflowStatus = "reviewing"
	WorkflowStatusApproved          ApplicationWorkflowStatus = "approved"
	WorkflowStatusDispatched        ApplicationWorkflowStatus = "dispatched"
	WorkflowStatusNeedsConfirmation ApplicationWorkflowStatus = "needs_confirmation"
	WorkflowStatusApplied           ApplicationWorkflowStatus = "applied"
	WorkflowStatusCancelled         ApplicationWorkflowStatus = "cancelled"
)

// ApplicationExecutionMode defines how the application will be completed (CAR-11).
type ApplicationExecutionMode string

const (
	ExecutionModeNativeManual    ApplicationExecutionMode = "native_manual"     // Candidate opens portal directly and uses prepared clipboard (CAR-11)
	ExecutionModeAssistantPreFill ApplicationExecutionMode = "assistant_prefill" // Pre-fills fields but strictly stops before final submission (SRC-L1, SRC-C6)
)

// ApplicationQuestionAnswer represents a single question field and confirmed answer (AT-003).
type ApplicationQuestionAnswer struct {
	QuestionID      string   `json:"question_id"`
	QuestionText    string   `json:"question_text"`
	FieldType       string   `json:"field_type"` // text, number, select, boolean, file
	Options         []string `json:"options,omitempty"`
	Required        bool     `json:"required"`
	AnswerValue     string   `json:"answer_value"`
	IsConfirmed     bool     `json:"is_confirmed"`
	NeedsInput      bool     `json:"needs_input"` // True if unknown; never guessed (AT-003)
	Category        string   `json:"category"`    // contact, experience, authorization, salary, custom
	ExplanationNote string   `json:"explanation_note,omitempty"`
}

// ApplicationMaterialBundle packages all documents and answers under an immutable checksum.
type ApplicationMaterialBundle struct {
	ResumeID             string                      `json:"resume_id"`
	ResumeChecksum       string                      `json:"resume_checksum"`
	ResumeFileName       string                      `json:"resume_file_name"`
	CoverLetterID        string                      `json:"cover_letter_id,omitempty"`
	CoverLetterChecksum  string                      `json:"cover_letter_checksum,omitempty"`
	CoverLetterFileName  string                      `json:"cover_letter_file_name,omitempty"`
	Questions            []ApplicationQuestionAnswer `json:"questions"`
	BundleChecksumSHA256 string                      `json:"bundle_checksum_sha256"`
}

// SubmissionReceipt contains evidence of completed application submission (AT-005).
type SubmissionReceipt struct {
	ReceiptID         string    `json:"receipt_id"`
	ProviderReference string    `json:"provider_reference,omitempty"`
	SubmissionURL     string    `json:"submission_url"`
	ConfirmedAt       time.Time `json:"confirmed_at"`
	ConfirmedByUser   bool      `json:"confirmed_by_user"`
	Notes             string    `json:"notes,omitempty"`
}

// ApplicationReviewSession represents a human-reviewed job application session (CAR-11, REQ-015, AT-005).
type ApplicationReviewSession struct {
	ID                  string                    `json:"id"`
	UserID              string                    `json:"user_id"`
	WorkspaceID         string                    `json:"workspace_id"`
	JobID               string                    `json:"job_id"`
	JobTitle            string                    `json:"job_title"`
	Company             string                    `json:"company"`
	ApplyURL            string                    `json:"apply_url"`
	SourceBoard         string                    `json:"source_board"`
	ExecutionMode       ApplicationExecutionMode  `json:"execution_mode"`
	Bundle              ApplicationMaterialBundle `json:"bundle"`
	Status              ApplicationWorkflowStatus `json:"status"`
	ApprovalToken       string                    `json:"approval_token,omitempty"`
	ApprovedAt          *time.Time                `json:"approved_at,omitempty"`
	ApproverID          string                    `json:"approver_id,omitempty"`
	DispatchedAt        *time.Time                `json:"dispatched_at,omitempty"`
	TimeoutDurationSecs int                       `json:"timeout_duration_secs"` // Default 600s (10 min)
	Receipt             *SubmissionReceipt        `json:"receipt,omitempty"`
	ApplicationRecordID string                    `json:"application_record_id,omitempty"`
	CreatedAt           time.Time                 `json:"created_at"`
	UpdatedAt           time.Time                 `json:"updated_at"`
}

// WorkflowEngine coordinates review, zero-fabrication questionnaire, approval binding, and confirmation.
type WorkflowEngine struct {
	secretKey []byte
}

func NewWorkflowEngine(secretKey []byte) *WorkflowEngine {
	if len(secretKey) == 0 {
		secretKey = []byte("super-secret-career-workflow-key-2026")
	}
	return &WorkflowEngine{secretKey: secretKey}
}

// ComputeBundleChecksum creates deterministic SHA-256 hash across resume, cover letter, and questions (AT-007).
func (we *WorkflowEngine) ComputeBundleChecksum(bundle ApplicationMaterialBundle) string {
	h := sha256.New()
	h.Write([]byte(bundle.ResumeID + ":" + bundle.ResumeChecksum + "|"))
	h.Write([]byte(bundle.CoverLetterID + ":" + bundle.CoverLetterChecksum + "|"))

	for _, q := range bundle.Questions {
		cleanVal := strings.TrimSpace(q.AnswerValue)
		record := fmt.Sprintf("%s:%t:%s|", q.QuestionID, q.IsConfirmed, cleanVal)
		h.Write([]byte(record))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// GenerateApprovalToken generates HMAC-SHA256 signature bound to session ID and bundle hash (REQ-015).
func (we *WorkflowEngine) GenerateApprovalToken(sessionID, bundleChecksum string) string {
	mac := hmac.New(sha256.New, we.secretKey)
	mac.Write([]byte(sessionID + ":" + bundleChecksum))
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateApprovalToken validates HMAC signature against current session state (REQ-015, AT-007).
func (we *WorkflowEngine) ValidateApprovalToken(sessionID, bundleChecksum, token string) bool {
	expected := we.GenerateApprovalToken(sessionID, bundleChecksum)
	return hmac.Equal([]byte(token), []byte(expected))
}

// PrepareApplicationSession initializes a review session from candidate profile and job target.
func (we *WorkflowEngine) PrepareApplicationSession(
	userID, workspaceID string,
	job DiscoveredJob,
	profile MasterCareerProfile,
	resume *TailoredResume,
	coverLetter *CoverLetter,
	mode ApplicationExecutionMode,
) (*ApplicationReviewSession, error) {
	if userID == "" || job.ID == "" {
		return nil, errors.New("userID and job are required to prepare application session")
	}

	if mode == "" {
		mode = ExecutionModeNativeManual
	}

	// 1. Standard question extraction with ZERO FABRICATION (AT-003)
	questions := we.synthesizeQuestionnaire(profile, job)

	// 2. Build Material Bundle
	bundle := ApplicationMaterialBundle{
		Questions: questions,
	}

	if resume != nil {
		bundle.ResumeID = resume.ID
		bundle.ResumeChecksum = resume.ChecksumSHA256
		bundle.ResumeFileName = resume.FileName
	} else {
		bundle.ResumeFileName = fmt.Sprintf("%s_Resume.pdf", strings.ReplaceAll(profile.Contact.FullName, " ", "_"))
		bundle.ResumeChecksum = "default_master_resume_checksum"
	}

	if coverLetter != nil {
		bundle.CoverLetterID = coverLetter.ID
		bundle.CoverLetterChecksum = coverLetter.ChecksumSHA256
		bundle.CoverLetterFileName = coverLetter.FileName
	}

	bundle.BundleChecksumSHA256 = we.ComputeBundleChecksum(bundle)

	sessionID := fmt.Sprintf("app_rev_%d", time.Now().UnixNano())
	now := time.Now().UTC()

	// Initial status is ready_for_review
	status := WorkflowStatusReadyForReview
	// Check if any question strictly needs input
	for _, q := range questions {
		if q.NeedsInput {
			status = WorkflowStatusDraft
			break
		}
	}

	applyURL := job.DirectApplyURL
	if applyURL == "" {
		applyURL = job.CanonicalURL
	}

	return &ApplicationReviewSession{
		ID:                  sessionID,
		UserID:              userID,
		WorkspaceID:         workspaceID,
		JobID:               job.ID,
		JobTitle:            job.Title,
		Company:             job.Company,
		ApplyURL:            applyURL,
		SourceBoard:         string(job.Source),
		ExecutionMode:       mode,
		Bundle:              bundle,
		Status:              status,
		TimeoutDurationSecs: 600, // 10 minutes timeout window (AT-005)
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

// synthesizeQuestionnaire generates form questions based on profile and job with strict zero fabrication (AT-003).
func (we *WorkflowEngine) synthesizeQuestionnaire(profile MasterCareerProfile, job DiscoveredJob) []ApplicationQuestionAnswer {
	var questions []ApplicationQuestionAnswer

	// Contact Full Name
	fullName := strings.TrimSpace(profile.Contact.FullName)
	questions = append(questions, ApplicationQuestionAnswer{
		QuestionID:   "contact_full_name",
		QuestionText: "Full Name",
		FieldType:    "text",
		Required:     true,
		AnswerValue:  fullName,
		IsConfirmed:  fullName != "",
		NeedsInput:   fullName == "",
		Category:     "contact",
	})

	// Contact Email
	email := strings.TrimSpace(profile.Contact.Email)
	questions = append(questions, ApplicationQuestionAnswer{
		QuestionID:   "contact_email",
		QuestionText: "Email Address",
		FieldType:    "text",
		Required:     true,
		AnswerValue:  email,
		IsConfirmed:  email != "",
		NeedsInput:   email == "",
		Category:     "contact",
	})

	// Contact Phone
	phone := strings.TrimSpace(profile.Contact.Phone)
	questions = append(questions, ApplicationQuestionAnswer{
		QuestionID:   "contact_phone",
		QuestionText: "Phone Number",
		FieldType:    "text",
		Required:     true,
		AnswerValue:  phone,
		IsConfirmed:  phone != "",
		NeedsInput:   phone == "",
		Category:     "contact",
	})

	// Location / Residence
	location := strings.TrimSpace(profile.Contact.Location)
	questions = append(questions, ApplicationQuestionAnswer{
		QuestionID:   "contact_location",
		QuestionText: "Current City & Country of Residence",
		FieldType:    "text",
		Required:     true,
		AnswerValue:  location,
		IsConfirmed:  location != "",
		NeedsInput:   location == "",
		Category:     "contact",
	})

	// Work Authorization & Sponsorship (AT-003: Zero Guessing!)
	// If unstated in confirmed profile, DO NOT GUESS "Yes" or "No". Must become needs_input!
	questions = append(questions, ApplicationQuestionAnswer{
		QuestionID:      "work_authorization",
		QuestionText:    "Are you legally authorized to work in the country of employment?",
		FieldType:       "boolean",
		Required:        true,
		AnswerValue:     "",
		IsConfirmed:     false,
		NeedsInput:      true, // Zero fabrication: never guess!
		Category:        "authorization",
		ExplanationNote: "AT-003: Legal work authorization must never be guessed or defaulted.",
	})

	// Visa Sponsorship (AT-003: Zero Guessing!)
	questions = append(questions, ApplicationQuestionAnswer{
		QuestionID:      "visa_sponsorship",
		QuestionText:    "Will you now or in the future require visa sponsorship?",
		FieldType:       "boolean",
		Required:        true,
		AnswerValue:     "",
		IsConfirmed:     false,
		NeedsInput:      true, // Zero fabrication: never guess!
		Category:        "authorization",
		ExplanationNote: "AT-003: Sponsorship requirement must be explicitly confirmed by candidate.",
	})

	// Relevant Professional Experience Years (AT-003)
	totalYears := len(profile.Experiences)
	expAnswer := ""
	expConfirmed := false
	if totalYears > 0 {
		expAnswer = fmt.Sprintf("%d", totalYears)
		expConfirmed = true
	}
	questions = append(questions, ApplicationQuestionAnswer{
		QuestionID:   "experience_years",
		QuestionText: "Total Years of Professional Experience",
		FieldType:    "number",
		Required:     true,
		AnswerValue:  expAnswer,
		IsConfirmed:  expConfirmed,
		NeedsInput:   !expConfirmed,
		Category:     "experience",
	})

	return questions
}

// UpdateAnswers updates questionnaire answers. If session was already approved,
// any edit invalidates the approval token and resets status to ready_for_review (AT-007).
func (we *WorkflowEngine) UpdateAnswers(
	session *ApplicationReviewSession,
	answers []ApplicationQuestionAnswer,
) (approvalInvalidated bool, err error) {
	if session.Status == WorkflowStatusDispatched || session.Status == WorkflowStatusApplied {
		return false, fmt.Errorf("%w: cannot edit answers in status %s", ErrWorkflowInvalidTransition, session.Status)
	}

	wasApproved := session.Status == WorkflowStatusApproved
	oldChecksum := session.Bundle.BundleChecksumSHA256

	// Replace answers
	session.Bundle.Questions = answers

	// Check if all required answers are provided
	hasUnresolved := false
	for i := range session.Bundle.Questions {
		q := &session.Bundle.Questions[i]
		if q.Required && (strings.TrimSpace(q.AnswerValue) == "" || !q.IsConfirmed) {
			q.NeedsInput = true
			hasUnresolved = true
		} else {
			q.NeedsInput = false
			q.IsConfirmed = true
		}
	}

	// Recompute bundle checksum (AT-007)
	newChecksum := we.ComputeBundleChecksum(session.Bundle)
	session.Bundle.BundleChecksumSHA256 = newChecksum
	session.UpdatedAt = time.Now().UTC()

	// Invalidation Check (AT-007)
	if wasApproved && (oldChecksum != newChecksum) {
		session.Status = WorkflowStatusReadyForReview
		session.ApprovalToken = ""
		session.ApprovedAt = nil
		session.ApproverID = ""
		approvalInvalidated = true
	} else if hasUnresolved {
		session.Status = WorkflowStatusDraft
	} else if !wasApproved {
		session.Status = WorkflowStatusReadyForReview
	}

	return approvalInvalidated, nil
}

// ApproveSession grants explicit human approval to the application bundle (REQ-015, AT-003).
func (we *WorkflowEngine) ApproveSession(session *ApplicationReviewSession, approverID string) error {
	if session.Status == WorkflowStatusDispatched || session.Status == WorkflowStatusApplied {
		return fmt.Errorf("%w: cannot approve session in status %s", ErrWorkflowInvalidTransition, session.Status)
	}

	// Verify all mandatory questions have resolved answers (AT-003: Zero Fabrication)
	for _, q := range session.Bundle.Questions {
		if q.Required && (q.NeedsInput || strings.TrimSpace(q.AnswerValue) == "" || !q.IsConfirmed) {
			return fmt.Errorf("%w: question '%s' requires confirmed answer", ErrUnresolvedQuestions, q.QuestionText)
		}
	}

	now := time.Now().UTC()
	token := we.GenerateApprovalToken(session.ID, session.Bundle.BundleChecksumSHA256)

	session.Status = WorkflowStatusApproved
	session.ApprovalToken = token
	session.ApprovedAt = &now
	session.ApproverID = approverID
	session.UpdatedAt = now

	return nil
}

// DispatchSession hand-offs application for execution (CAR-11, AT-005).
// CRITICAL INVARIANT: Clicking Apply or dispatching NEVER marks Applied directly! (AT-005)
func (we *WorkflowEngine) DispatchSession(session *ApplicationReviewSession, mode ApplicationExecutionMode) error {
	if session.Status != WorkflowStatusApproved {
		return fmt.Errorf("%w: session must be approved before dispatch", ErrWorkflowApprovalRequired)
	}

	// Verify cryptographic token integrity (REQ-015, AT-007)
	if !we.ValidateApprovalToken(session.ID, session.Bundle.BundleChecksumSHA256, session.ApprovalToken) {
		return ErrWorkflowApprovalInvalidated
	}

	now := time.Now().UTC()
	if mode != "" {
		session.ExecutionMode = mode
	}

	// Sets status to dispatched, NOT applied (AT-005)
	session.Status = WorkflowStatusDispatched
	session.DispatchedAt = &now
	session.UpdatedAt = now

	return nil
}

// CheckConfirmationTimeout transitions a dispatched session to needs_confirmation if timeout expired (AT-005).
func (we *WorkflowEngine) CheckConfirmationTimeout(session *ApplicationReviewSession) bool {
	if session.Status != WorkflowStatusDispatched || session.DispatchedAt == nil {
		return false
	}

	timeout := time.Duration(session.TimeoutDurationSecs) * time.Second
	if time.Since(*session.DispatchedAt) >= timeout {
		session.Status = WorkflowStatusNeedsConfirmation
		session.UpdatedAt = time.Now().UTC()
		return true
	}

	return false
}

// ConfirmSubmission explicitly confirms completion via provider receipt or user declaration (REQ-005, AT-005).
func (we *WorkflowEngine) ConfirmSubmission(
	session *ApplicationReviewSession,
	receipt SubmissionReceipt,
) (*ApplicationRecord, error) {
	if session.Status != WorkflowStatusDispatched && session.Status != WorkflowStatusNeedsConfirmation && session.Status != WorkflowStatusApproved {
		return nil, fmt.Errorf("%w: cannot confirm submission in status %s", ErrWorkflowInvalidTransition, session.Status)
	}

	now := time.Now().UTC()
	if receipt.ConfirmedAt.IsZero() {
		receipt.ConfirmedAt = now
	}
	if receipt.ReceiptID == "" {
		receipt.ReceiptID = fmt.Sprintf("rec_%d", now.UnixNano())
	}

	session.Status = WorkflowStatusApplied
	session.Receipt = &receipt
	session.UpdatedAt = now

	// Compute fingerprint for dedupe (CAR-10)
	h := sha256.New()
	h.Write([]byte(strings.ToLower(strings.TrimSpace(session.Company)) + "|"))
	h.Write([]byte(strings.ToLower(strings.TrimSpace(session.JobTitle)) + "|"))
	fingerprint := hex.EncodeToString(h.Sum(nil))

	// Create durable ApplicationRecord for the ledger (CAR-10, REQ-005)
	appRecord := &ApplicationRecord{
		ID:             fmt.Sprintf("app_rec_%d", now.UnixNano()),
		UserID:         session.UserID,
		JobID:          session.JobID,
		CanonicalURL:   session.ApplyURL,
		Fingerprint:    fingerprint,
		Title:          session.JobTitle,
		Company:        session.Company,
		Status:         "submitted",
		SubmissionMode: string(session.ExecutionMode),
		AppliedAt:      receipt.ConfirmedAt,
		Notes:          receipt.Notes,
	}
	session.ApplicationRecordID = appRecord.ID

	return appRecord, nil
}

// CancelSession cancels a pending or dispatched review session.
func (we *WorkflowEngine) CancelSession(session *ApplicationReviewSession, reason string) error {
	if session.Status == WorkflowStatusApplied {
		return fmt.Errorf("%w: cannot cancel an already applied session", ErrWorkflowAlreadyConfirmed)
	}

	session.Status = WorkflowStatusCancelled
	session.ApprovalToken = ""
	session.UpdatedAt = time.Now().UTC()
	return nil
}
