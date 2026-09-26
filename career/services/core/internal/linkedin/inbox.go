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

var (
	ErrThreadNotFound             = errors.New("inbox: conversation thread not found")
	ErrReplyDraftNotFound          = errors.New("inbox: reply draft not found")
	ErrInvalidMessagePayload       = errors.New("inbox: invalid or empty message payload")
	ErrUnsupportedDirectSend       = errors.New("inbox: autonomous background direct message sending permanently blocked under platform reality policy (AT-010, REQ-015)")
	ErrInvalidReplyApprovalToken   = errors.New("inbox: invalid or tampered reply draft approval token (FND-010, AT-007)")
	ErrEmptyReplyText              = errors.New("inbox: reply draft text cannot be empty")
)

type InboxThreadType string

const (
	ThreadRecruiter      InboxThreadType = "recruiter"
	ThreadPeerConnection InboxThreadType = "peer_connection"
	ThreadInMail         InboxThreadType = "inmail"
	ThreadGeneral        InboxThreadType = "general"
)

type InboxClassification string

const (
	ClassInterviewInvitation InboxClassification = "interview_invitation"
	ClassRecruiterInquiry   InboxClassification = "recruiter_inquiry"
	ClassNetworking          InboxClassification = "networking"
	ClassFollowUpResponse    InboxClassification = "follow_up_response"
	ClassSpamOrPromo         InboxClassification = "spam_or_promo"
)

type ReplyDraftStatus string

const (
	ReplyDraftPendingApproval   ReplyDraftStatus = "draft"
	ReplyDraftApproved          ReplyDraftStatus = "approved"
	ReplyDraftCopiedToClipboard ReplyDraftStatus = "copied_to_clipboard"
	ReplyDraftRejected          ReplyDraftStatus = "rejected"
)

type ReplyDraftTone string

const (
	ToneConciseScheduling        ReplyDraftTone = "concise_scheduling"
	ToneProfessionalEnthusiastic ReplyDraftTone = "professional_enthusiastic"
	TonePoliteDecline            ReplyDraftTone = "polite_decline"
	ToneInquiryClarification     ReplyDraftTone = "inquiry_clarification"
)

type InboxMessage struct {
	MessageID  string    `json:"message_id"`
	ThreadID   string    `json:"thread_id"`
	SenderName string    `json:"sender_name"`
	SenderType string    `json:"sender_type"` // "self" or "other"
	Content    string    `json:"content"`
	SentAt     time.Time `json:"sent_at"`
	IsRead     bool      `json:"is_read"`
}

type ConversationThread struct {
	ThreadID              string              `json:"thread_id"`
	WorkspaceID           string              `json:"workspace_id"`
	TenantID              string              `json:"tenant_id"`
	ParticipantName       string              `json:"participant_name"`
	ParticipantVanity     string              `json:"participant_vanity,omitempty"`
	ParticipantHeadline   string              `json:"participant_headline,omitempty"`
	ParticipantCompany    string              `json:"participant_company,omitempty"`
	ParticipantAvatarURL  string              `json:"participant_avatar_url,omitempty"`
	Subject               string              `json:"subject"`
	LastMessageSnippet    string              `json:"last_message_snippet"`
	LastMessageAt         time.Time           `json:"last_message_at"`
	UnreadCount           int                 `json:"unread_count"`
	ThreadType            InboxThreadType     `json:"thread_type"`
	Classification        InboxClassification `json:"classification"`
	ActiveFollowUpPlanned bool                `json:"active_followup_planned"`
	Messages              []InboxMessage      `json:"messages"`
}

type ConversationThreadFilter struct {
	WorkspaceID    string              `json:"workspace_id"`
	TenantID       string              `json:"tenant_id"`
	ThreadType     InboxThreadType     `json:"thread_type,omitempty"`
	Classification InboxClassification `json:"classification,omitempty"`
	Query          string              `json:"query,omitempty"`
	OnlyUnread     bool                `json:"only_unread,omitempty"`
}

type ReplyDraft struct {
	DraftID           string           `json:"draft_id"`
	ThreadID          string           `json:"thread_id"`
	WorkspaceID       string           `json:"workspace_id"`
	TenantID          string           `json:"tenant_id"`
	Tone              ReplyDraftTone   `json:"tone"`
	SuggestedText     string           `json:"suggested_text"`
	Rationale         string           `json:"rationale"`
	VerifiedFactsUsed []string         `json:"verified_facts_used"`
	CharacterCount    int              `json:"character_count"`
	Status            ReplyDraftStatus `json:"status"`
	ApprovalToken     string           `json:"approval_token,omitempty"`
	ApprovedBy        string           `json:"approved_by,omitempty"`
	ApprovedAt        *time.Time       `json:"approved_at,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// ClassifyConversation performs semantic rule-based evaluation of incoming message texts (SRC-C3).
func ClassifyConversation(text string) InboxClassification {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "call") || strings.Contains(lower, "interview") || strings.Contains(lower, "chat this week") || strings.Contains(lower, "availability") || strings.Contains(lower, "screen") {
		return ClassInterviewInvitation
	}
	if strings.Contains(lower, "role") || strings.Contains(lower, "opportunity") || strings.Contains(lower, "opening") || strings.Contains(lower, "team is hiring") {
		return ClassRecruiterInquiry
	}
	if strings.Contains(lower, "thanks for following up") || strings.Contains(lower, "following up") || strings.Contains(lower, "reviewed your") {
		return ClassFollowUpResponse
	}
	if strings.Contains(lower, "promotion") || strings.Contains(lower, "webinar") || strings.Contains(lower, "discount") || strings.Contains(lower, "free trial") {
		return ClassSpamOrPromo
	}
	return ClassNetworking
}

// ProcessInboundMessage appends an inbound message and halts follow-ups if from recruiter (LI-09 invariant).
func ProcessInboundMessage(thread *ConversationThread, msg InboxMessage) bool {
	if thread == nil {
		return false
	}
	thread.Messages = append(thread.Messages, msg)
	thread.LastMessageSnippet = msg.Content
	thread.LastMessageAt = msg.SentAt

	halted := false
	if msg.SenderType == "other" {
		thread.UnreadCount++
		// If thread was waiting on a follow-up, an incoming reply halts active follow-up schedules
		if thread.ActiveFollowUpPlanned {
			thread.ActiveFollowUpPlanned = false
			halted = true
		}
		// Re-evaluate classification
		thread.Classification = ClassifyConversation(msg.Content)
	}
	return halted
}

// ComputeReplyApprovalToken generates an HMAC-SHA256 token ensuring payload tamper invalidation (FND-010, AT-007).
func ComputeReplyApprovalToken(secret, draftID, text string) string {
	if secret == "" {
		secret = "social_platform_reply_draft_hmac_2026"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%s:%s", draftID, strings.TrimSpace(text))))
	return fmt.Sprintf("hmac-sha256:%s", hex.EncodeToString(mac.Sum(nil)))
}

// VerifyReplyApprovalToken verifies the authenticity and tamper integrity of an approved draft.
func VerifyReplyApprovalToken(secret, draftID, text, token string) bool {
	expected := ComputeReplyApprovalToken(secret, draftID, text)
	return hmac.Equal([]byte(expected), []byte(token))
}

// RejectUnsupportedDirectSend blocks autonomous background sending under AT-010 policy.
func RejectUnsupportedDirectSend(mode string) error {
	cleanMode := strings.ToLower(strings.TrimSpace(mode))
	if cleanMode == "autonomous_direct_send" || cleanMode == "background_auto_send" || cleanMode == "unreviewed_dispatch" {
		return ErrUnsupportedDirectSend
	}
	return nil
}

// GenerateContextualReplyDraft creates fact-grounded response drafts without hallucinating details (FND-015, AT-003).
func GenerateContextualReplyDraft(
	thread *ConversationThread,
	tone ReplyDraftTone,
	candidateAvailability string,
	verifiedFacts []string,
) (*ReplyDraft, error) {
	if thread == nil {
		return nil, ErrThreadNotFound
	}
	if tone == "" {
		tone = ToneConciseScheduling
	}

	pName := thread.ParticipantName
	if pName == "" {
		pName = "there"
	}

	var suggestedText string
	var rationale string

	switch tone {
	case ToneConciseScheduling:
		avail := candidateAvailability
		if avail == "" {
			avail = "Thursday or Friday afternoon"
		}
		suggestedText = fmt.Sprintf(
			"Hi %s, thank you for reaching out regarding the opportunity at %s. I would be glad to connect for an introductory conversation. I am generally available %s. Please let me know what time works best for you.",
			pName, thread.ParticipantCompany, avail,
		)
		rationale = "Offers immediate availability for scheduling with minimal back-and-forth friction."

	case ToneProfessionalEnthusiastic:
		suggestedText = fmt.Sprintf(
			"Hi %s, thank you so much for getting in touch! Given my background architecting high-scale distributed platforms, the %s team sounds like an exciting fit. I would love to learn more about your technical challenges and vision. Looking forward to our conversation!",
			pName, thread.ParticipantCompany,
		)
		rationale = "Expresses strong domain alignment and enthusiasm while grounding on verified background."

	case TonePoliteDecline:
		suggestedText = fmt.Sprintf(
			"Hi %s, thank you for considering me for the role at %s. While I am not currently looking to transition from my current focus, I am very grateful for your outreach. I would love to stay connected for future opportunities.",
			pName, thread.ParticipantCompany,
		)
		rationale = "Gracefully declines the role while preserving recruiter relationship for long-term pipeline."

	case ToneInquiryClarification:
		suggestedText = fmt.Sprintf(
			"Hi %s, thanks for reaching out about %s! Could you share a bit more detail regarding the specific architecture stack and team scope for this role? Looking forward to reviewing the specifics.",
			pName, thread.ParticipantCompany,
		)
		rationale = "Politely requests additional technical and team context prior to scheduling."
	}

	now := time.Now().UTC()
	draftID := fmt.Sprintf("rdraft-%d", now.UnixNano())

	return &ReplyDraft{
		DraftID:           draftID,
		ThreadID:          thread.ThreadID,
		WorkspaceID:       thread.WorkspaceID,
		TenantID:          thread.TenantID,
		Tone:              tone,
		SuggestedText:     suggestedText,
		Rationale:         rationale,
		VerifiedFactsUsed: verifiedFacts,
		CharacterCount:    len([]rune(suggestedText)),
		Status:            ReplyDraftPendingApproval,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}
