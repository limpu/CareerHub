package linkedin

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// LeadStatus represents the relationship lifecycle state for a recruiter or hiring lead (LI-06).
type LeadStatus string

const (
	LeadStatusNew                LeadStatus = "new"
	LeadStatusContacted          LeadStatus = "contacted"
	LeadStatusInDialogue         LeadStatus = "in_dialogue"
	LeadStatusInterviewScheduled LeadStatus = "interview_scheduled"
	LeadStatusOfferPending       LeadStatus = "offer_pending"
	LeadStatusClosed             LeadStatus = "closed"
	LeadStatusArchived           LeadStatus = "archived"
)

// LeadOutreachStage tracks communication delivery state (AT-010: no fake live send).
type LeadOutreachStage string

const (
	OutreachDraft             LeadOutreachStage = "draft"
	OutreachReadyToSend       LeadOutreachStage = "ready_to_send"
	OutreachCopiedToClipboard LeadOutreachStage = "copied_to_clipboard"
	OutreachContacted         LeadOutreachStage = "contacted"
	OutreachReplied           LeadOutreachStage = "replied"
	OutreachArchived          LeadOutreachStage = "archived"
)

// ReminderStatus represents the follow-up reminder lifecycle.
type ReminderStatus string

const (
	ReminderPending   ReminderStatus = "pending"
	ReminderCompleted ReminderStatus = "completed"
	ReminderSnoozed   ReminderStatus = "snoozed"
)

var (
	ErrLeadNotFound         = errors.New("recruiter lead not found")
	ErrEmptyWorkspaceID     = errors.New("workspace_id cannot be empty")
	ErrEmptyTenantID        = errors.New("tenant_id cannot be empty")
	ErrEmptyRecruiterName   = errors.New("recruiter name cannot be empty")
	ErrEmptyCompany         = errors.New("company cannot be empty")
	ErrEmptyNoteContent     = errors.New("note content cannot be empty")
	ErrEmptyReminderMessage = errors.New("reminder message cannot be empty")
	ErrPastReminderDate     = errors.New("reminder due date cannot be in the distant past")
)

// LeadNote represents a timestamped observation or communication log entry.
type LeadNote struct {
	ID        string    `json:"id"`
	LeadID    string    `json:"lead_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// FollowUpReminder defines a scheduled action alert for the candidate.
type FollowUpReminder struct {
	DueDate     time.Time      `json:"due_date"`
	Message     string         `json:"message"`
	Status      ReminderStatus `json:"status"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

// RecruiterLead represents a tracked relationship with a recruiter or hiring manager (LI-06).
type RecruiterLead struct {
	ID                   string            `json:"id"`
	WorkspaceID          string            `json:"workspace_id"`
	TenantID             string            `json:"tenant_id"`
	RecruiterName        string            `json:"recruiter_name"`
	RecruiterTitle       string            `json:"recruiter_title"`
	Company              string            `json:"company"`
	LinkedInURL          string            `json:"linkedin_url"`
	SourceEntityType     string            `json:"source_entity_type,omitempty"` // "person_record", "post_record", "manual"
	SourceEntityID       string            `json:"source_entity_id,omitempty"`
	Status               LeadStatus        `json:"status"`
	OutreachStage        LeadOutreachStage `json:"outreach_stage"`
	RelatedJobID         string            `json:"related_job_id,omitempty"`
	RelatedApplicationID string            `json:"related_application_id,omitempty"`
	Notes                []LeadNote        `json:"notes"`
	Reminder             *FollowUpReminder `json:"reminder,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

// RecruiterLeadFilter provides query parameters for listing workspace leads.
type RecruiterLeadFilter struct {
	WorkspaceID string     `json:"workspace_id"`
	TenantID    string     `json:"tenant_id"`
	Company     string     `json:"company,omitempty"`
	Status      LeadStatus `json:"status,omitempty"`
	SearchQuery string     `json:"search_query,omitempty"`
}

// ValidateRecruiterLead validates invariant properties and mandatory tenant bindings (AT-011).
func ValidateRecruiterLead(lead *RecruiterLead) error {
	if lead == nil {
		return errors.New("recruiter lead cannot be nil")
	}
	if strings.TrimSpace(lead.WorkspaceID) == "" {
		return ErrEmptyWorkspaceID
	}
	if strings.TrimSpace(lead.TenantID) == "" {
		return ErrEmptyTenantID
	}
	if strings.TrimSpace(lead.RecruiterName) == "" {
		return ErrEmptyRecruiterName
	}
	if strings.TrimSpace(lead.Company) == "" {
		return ErrEmptyCompany
	}
	if strings.TrimSpace(lead.LinkedInURL) == "" || !strings.Contains(lead.LinkedInURL, "linkedin.com") {
		return ErrInvalidLinkedInURL
	}
	if lead.Status == "" {
		lead.Status = LeadStatusNew
	}
	if lead.OutreachStage == "" {
		lead.OutreachStage = OutreachDraft
	}
	return nil
}

// AddNote appends an interaction note to the lead.
func (l *RecruiterLead) AddNote(author, content string) (*LeadNote, error) {
	if strings.TrimSpace(content) == "" {
		return nil, ErrEmptyNoteContent
	}
	if strings.TrimSpace(author) == "" {
		author = "operator"
	}
	now := time.Now().UTC()
	note := LeadNote{
		ID:        fmt.Sprintf("note-%d", time.Now().UnixNano()),
		LeadID:    l.ID,
		Author:    strings.TrimSpace(author),
		Content:   strings.TrimSpace(content),
		CreatedAt: now,
	}
	l.Notes = append(l.Notes, note)
	l.UpdatedAt = now
	return &note, nil
}

// SetReminder schedules or updates a follow-up reminder.
func (l *RecruiterLead) SetReminder(dueDate time.Time, message string) (*FollowUpReminder, error) {
	if strings.TrimSpace(message) == "" {
		return nil, ErrEmptyReminderMessage
	}
	// Allow reminders up to 24h in past in case recording retroactive milestones, but bar further back
	if dueDate.Before(time.Now().Add(-24 * time.Hour)) {
		return nil, ErrPastReminderDate
	}
	reminder := &FollowUpReminder{
		DueDate: dueDate,
		Message: strings.TrimSpace(message),
		Status:  ReminderPending,
	}
	l.Reminder = reminder
	l.UpdatedAt = time.Now().UTC()
	return reminder, nil
}

// CompleteReminder marks the existing reminder as completed.
func (l *RecruiterLead) CompleteReminder() error {
	if l.Reminder == nil {
		return errors.New("no reminder scheduled on this lead")
	}
	now := time.Now().UTC()
	l.Reminder.Status = ReminderCompleted
	l.Reminder.CompletedAt = &now
	l.UpdatedAt = now
	return nil
}

// LinkApplication links the recruiter lead to a specific job or application record (IMP-CAR-20, LI-06).
func (l *RecruiterLead) LinkApplication(jobID, applicationID string) {
	l.RelatedJobID = strings.TrimSpace(jobID)
	l.RelatedApplicationID = strings.TrimSpace(applicationID)
	l.UpdatedAt = time.Now().UTC()
}

// GenerateOutreachDraft constructs a personalized, compliance-aligned message draft (AT-010).
// In accordance with truth-in-advertising, no automated messaging occurs; clipboard copy and native link are provided.
type OutreachDraftPayload struct {
	Subject       string `json:"subject"`
	Body          string `json:"body"`
	CharacterCount int   `json:"character_count"`
	WithinLimit   bool   `json:"within_limit"`
	DirectChatURL string `json:"direct_chat_url"`
}

func (l *RecruiterLead) GenerateOutreachDraft(candidateName, targetRole string) OutreachDraftPayload {
	if candidateName == "" {
		candidateName = "Candidate"
	}
	if targetRole == "" {
		targetRole = "open engineering opportunities"
	}

	body := fmt.Sprintf("Hi %s, noticed your work at %s and wanted to connect regarding %s. With my background in high-scale systems, I would welcome the opportunity to connect and stay in touch!",
		l.RecruiterName, l.Company, targetRole)

	charCount := len(body)
	withinLimit := charCount <= 300 // Standard LinkedIn connection note limit

	chatURL := l.LinkedInURL
	if !strings.HasSuffix(chatURL, "/") {
		chatURL += "/"
	}

	return OutreachDraftPayload{
		Subject:       fmt.Sprintf("Introduction - %s via LinkedIn", candidateName),
		Body:          body,
		CharacterCount: charCount,
		WithinLimit:   withinLimit,
		DirectChatURL: chatURL,
	}
}
