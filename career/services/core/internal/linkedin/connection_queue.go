package linkedin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ConnectionQueueStatus defines the lifecycle of an outbound connection invite note (LI-07).
type ConnectionQueueStatus string

const (
	QueueItemPendingApproval ConnectionQueueStatus = "pending_approval"
	QueueItemApproved        ConnectionQueueStatus = "approved"
	QueueItemCopied          ConnectionQueueStatus = "copied_to_clipboard"
	QueueItemCompleted       ConnectionQueueStatus = "completed_manually"
	QueueItemRejected        ConnectionQueueStatus = "rejected"
	QueueItemSkipped         ConnectionQueueStatus = "skipped"
)

const (
	LinkedInMaxNoteCharacters = 300
	DefaultDailyInviteLimit   = 20
	DefaultWeeklyInviteLimit  = 80
	MaxSafeDailyInviteLimit   = 25
	MaxSafeWeeklyInviteLimit  = 100
)

var (
	ErrQueueItemNotFound    = errors.New("connection queue item not found")
	ErrDuplicateRecipient   = errors.New("recipient already queued or contacted; duplicate outreach rejected")
	ErrDailyBudgetExceeded  = errors.New("daily connection invite budget exhausted; throttling active for account safety")
	ErrWeeklyBudgetExceeded = errors.New("rolling-week connection invite budget exhausted; throttling active for account safety")
	ErrNoteExceedsLimit     = errors.New("connection note exceeds LinkedIn 300 character boundary")
	ErrEmptyNoteText        = errors.New("connection note text cannot be empty")
	ErrNotApproved          = errors.New("connection item must be explicitly approved before execution")
	ErrInvalidApprovalToken = errors.New("approval token is invalid or does not match note payload")
)

// ConnectionBudget tracks daily and rolling-week limits to guarantee account safety (FND-011, AT-008, SRC-L1).
type ConnectionBudget struct {
	WorkspaceID   string    `json:"workspace_id"`
	TenantID      string    `json:"tenant_id"`
	DailyLimit    int       `json:"daily_limit"`
	DailyUsed     int       `json:"daily_used"`
	WeeklyLimit   int       `json:"weekly_limit"`
	WeeklyUsed    int       `json:"weekly_used"`
	LastResetDate time.Time `json:"last_reset_date"`
}

// NewConnectionBudget initializes standard safe conservative budget limits.
func NewConnectionBudget(workspaceID, tenantID string) *ConnectionBudget {
	return &ConnectionBudget{
		WorkspaceID:   workspaceID,
		TenantID:      tenantID,
		DailyLimit:    DefaultDailyInviteLimit,
		DailyUsed:     0,
		WeeklyLimit:   DefaultWeeklyInviteLimit,
		WeeklyUsed:    0,
		LastResetDate: time.Now().UTC(),
	}
}

// RemainingDaily returns the remaining safe connection invites for the 24h window.
func (b *ConnectionBudget) RemainingDaily() int {
	rem := b.DailyLimit - b.DailyUsed
	if rem < 0 {
		return 0
	}
	return rem
}

// RemainingWeekly returns the remaining safe connection invites for the rolling week.
func (b *ConnectionBudget) RemainingWeekly() int {
	rem := b.WeeklyLimit - b.WeeklyUsed
	if rem < 0 {
		return 0
	}
	return rem
}

// CanConsume checks if safe invite budget remains.
func (b *ConnectionBudget) CanConsume() error {
	if b.DailyUsed >= b.DailyLimit {
		return ErrDailyBudgetExceeded
	}
	if b.WeeklyUsed >= b.WeeklyLimit {
		return ErrWeeklyBudgetExceeded
	}
	return nil
}

// Consume decrements the budget counters upon user confirmation.
func (b *ConnectionBudget) Consume() error {
	if err := b.CanConsume(); err != nil {
		return err
	}
	b.DailyUsed++
	b.WeeklyUsed++
	return nil
}

// ConnectionQueueItem represents an individualized connection note draft in the queue (LI-07).
type ConnectionQueueItem struct {
	ID                   string                `json:"id"`
	WorkspaceID          string                `json:"workspace_id"`
	TenantID             string                `json:"tenant_id"`
	RecipientID          string                `json:"recipient_id,omitempty"`
	RecipientName        string                `json:"recipient_name"`
	RecipientTitle       string                `json:"recipient_title"`
	RecipientCompany     string                `json:"recipient_company"`
	RecipientLinkedInURL string                `json:"recipient_linkedin_url"`
	NoteText             string                `json:"note_text"`
	CharacterCount       int                   `json:"character_count"`
	WithinLimit          bool                  `json:"within_limit"`
	Status               ConnectionQueueStatus `json:"status"`
	ApprovalToken        string                `json:"approval_token,omitempty"`
	ApprovedBy           string                `json:"approved_by,omitempty"`
	ApprovedAt           *time.Time            `json:"approved_at,omitempty"`
	ContextFactors       []string              `json:"context_factors,omitempty"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
	CompletedAt          *time.Time            `json:"completed_at,omitempty"`
}

// ConnectionQueueFilter provides search and status filtering for queue items.
type ConnectionQueueFilter struct {
	WorkspaceID string                `json:"workspace_id"`
	TenantID    string                `json:"tenant_id"`
	Status      ConnectionQueueStatus `json:"status,omitempty"`
	SearchQuery string                `json:"search_query,omitempty"`
}

// ComputeApprovalToken creates a cryptographic HMAC-SHA256 token binding the item and note payload (FND-010, AT-007).
func ComputeApprovalToken(secret, itemID, noteText string) string {
	if secret == "" {
		secret = "antigravity-default-approval-secret-key-2026"
	}
	payload := fmt.Sprintf("%s:%s", strings.TrimSpace(itemID), strings.TrimSpace(noteText))
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyApprovalToken checks if the token matches the item and note text.
func VerifyApprovalToken(secret, itemID, noteText, token string) bool {
	if strings.TrimSpace(token) == "" {
		return false
	}
	expected := ComputeApprovalToken(secret, itemID, noteText)
	return hmac.Equal([]byte(expected), []byte(token))
}

// ValidateConnectionQueueItem asserts domain boundaries and limits.
func ValidateConnectionQueueItem(item *ConnectionQueueItem) error {
	if item == nil {
		return errors.New("connection queue item cannot be nil")
	}
	if strings.TrimSpace(item.WorkspaceID) == "" {
		return ErrEmptyWorkspaceID
	}
	if strings.TrimSpace(item.TenantID) == "" {
		return ErrEmptyTenantID
	}
	if strings.TrimSpace(item.RecipientName) == "" {
		return errors.New("recipient name cannot be empty")
	}
	if strings.TrimSpace(item.RecipientLinkedInURL) == "" || !strings.Contains(item.RecipientLinkedInURL, "linkedin.com") {
		return ErrInvalidLinkedInURL
	}
	if strings.TrimSpace(item.NoteText) == "" {
		return ErrEmptyNoteText
	}
	charCount := len([]rune(strings.TrimSpace(item.NoteText)))
	item.CharacterCount = charCount
	item.WithinLimit = charCount <= LinkedInMaxNoteCharacters
	if !item.WithinLimit {
		return ErrNoteExceedsLimit
	}
	if item.Status == "" {
		item.Status = QueueItemPendingApproval
	}
	return nil
}

// GenerateConnectionNoteDraft constructs an AI-assisted, fact-grounded personalized note under 300 characters (SRC-L1, REQ-016).
func GenerateConnectionNoteDraft(candidateName, targetRole, recipientName, recipientCompany, recipientTitle, keyOverlap string) (string, []string) {
	if candidateName == "" {
		candidateName = "Candidate"
	}
	if targetRole == "" {
		targetRole = "engineering opportunities"
	}
	if recipientName == "" {
		recipientName = "there"
	}

	var factors []string
	if recipientCompany != "" {
		factors = append(factors, fmt.Sprintf("Company match: %s", recipientCompany))
	}
	if keyOverlap != "" {
		factors = append(factors, fmt.Sprintf("Domain overlap: %s", keyOverlap))
	}
	factors = append(factors, fmt.Sprintf("Target role: %s", targetRole))

	var note string
	if recipientCompany != "" && keyOverlap != "" {
		note = fmt.Sprintf("Hi %s, noticed your work at %s. With my background in %s, I would love to connect and follow your team's engineering work!",
			recipientName, recipientCompany, keyOverlap)
	} else if recipientCompany != "" {
		note = fmt.Sprintf("Hi %s, saw your updates at %s and wanted to connect regarding %s. Looking forward to staying in touch!",
			recipientName, recipientCompany, targetRole)
	} else {
		note = fmt.Sprintf("Hi %s, wanted to connect regarding %s and follow your professional journey on LinkedIn. Looking forward to staying in touch!",
			recipientName, targetRole)
	}

	// Double-check character count and condense if needed to ensure strictly <= 300 characters
	if len([]rune(note)) > LinkedInMaxNoteCharacters {
		runes := []rune(note)
		note = string(runes[:297]) + "..."
	}

	return note, factors
}
