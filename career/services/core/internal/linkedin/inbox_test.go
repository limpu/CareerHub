package linkedin_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/social-platform/services/core/internal/linkedin"
)

type InboxFixture struct {
	Scenarios struct {
		RecruiterInterviewInvitation struct {
			WorkspaceID string `json:"workspace_id"`
			TenantID    string `json:"tenant_id"`
			Thread      struct {
				ThreadID              string                   `json:"thread_id"`
				ParticipantName       string                   `json:"participant_name"`
				ParticipantVanity     string                   `json:"participant_vanity"`
				ParticipantHeadline   string                   `json:"participant_headline"`
				ParticipantCompany    string                   `json:"participant_company"`
				Subject               string                   `json:"subject"`
				ThreadType            linkedin.InboxThreadType `json:"thread_type"`
				Classification        string                   `json:"classification"`
				ActiveFollowUpPlanned bool                     `json:"active_followup_planned"`
				Messages              []struct {
					MessageID  string `json:"message_id"`
					SenderName string `json:"sender_name"`
					SenderType string `json:"sender_type"`
					Content    string `json:"content"`
					SentAt     string `json:"sent_at"`
					IsRead     bool   `json:"is_read"`
				} `json:"messages"`
			} `json:"thread"`
			ExpectedClassification string `json:"expected_classification"`
			ReplyGeneration        struct {
				Tone                  string   `json:"tone"`
				CandidateAvailability string   `json:"candidate_availability"`
				ExpectedFactsUsed     []string `json:"expected_facts_used"`
			} `json:"reply_generation"`
		} `json:"recruiter_interview_invitation_thread"`

		StopFollowUpOnReply struct {
			WorkspaceID                  string `json:"workspace_id"`
			TenantID                     string `json:"tenant_id"`
			ThreadID                     string `json:"thread_id"`
			InitialActiveFollowUpPlanned bool   `json:"initial_active_followup_planned"`
			InboundMessage               struct {
				MessageID  string `json:"message_id"`
				SenderName string `json:"sender_name"`
				SenderType string `json:"sender_type"`
				Content    string `json:"content"`
				SentAt     string `json:"sent_at"`
			} `json:"inbound_message"`
			ExpectedFollowUpHalted bool `json:"expected_followup_halted"`
		} `json:"stop_followup_on_reply_invariant"`

		CrossTenantInboxIsolation struct {
			AlphaThread struct {
				ThreadID        string `json:"thread_id"`
				WorkspaceID     string `json:"workspace_id"`
				TenantID        string `json:"tenant_id"`
				ParticipantName string `json:"participant_name"`
			} `json:"alpha_thread"`
			BetaQuery struct {
				WorkspaceID string `json:"workspace_id"`
				TenantID    string `json:"tenant_id"`
			} `json:"beta_query"`
			ExpectedIsolationError string `json:"expected_isolation_error"`
		} `json:"cross_tenant_inbox_isolation"`

		UnsupportedDirectSend struct {
			WorkspaceID   string `json:"workspace_id"`
			TenantID      string `json:"tenant_id"`
			AttemptedMode string `json:"attempted_mode"`
			ExpectedError string `json:"expected_error"`
		} `json:"unsupported_direct_send_fail_closed"`
	} `json:"scenarios"`
}

func loadInboxFixture(t *testing.T) *InboxFixture {
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_inbox_reply.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read linkedin_inbox_reply fixture: %v", err)
	}

	var fix InboxFixture
	if err := json.Unmarshal(data, &fix); err != nil {
		t.Fatalf("failed to unmarshal linkedin_inbox_reply fixture: %v", err)
	}
	return &fix
}

func TestInbox_Fixtures_Scenarios(t *testing.T) {
	fix := loadInboxFixture(t)
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	// Scenario 1: recruiter_interview_invitation_thread
	t.Run("Scenario 1: Interview Invitation & Reply Draft Approval", func(t *testing.T) {
		sc := fix.Scenarios.RecruiterInterviewInvitation
		msgs := make([]linkedin.InboxMessage, len(sc.Thread.Messages))
		for i, m := range sc.Thread.Messages {
			sentAt, _ := time.Parse(time.RFC3339, m.SentAt)
			msgs[i] = linkedin.InboxMessage{
				MessageID:  m.MessageID,
				ThreadID:   sc.Thread.ThreadID,
				SenderName: m.SenderName,
				SenderType: m.SenderType,
				Content:    m.Content,
				SentAt:     sentAt,
				IsRead:     m.IsRead,
			}
		}

		thread := &linkedin.ConversationThread{
			ThreadID:              sc.Thread.ThreadID,
			WorkspaceID:           sc.WorkspaceID,
			TenantID:              sc.TenantID,
			ParticipantName:       sc.Thread.ParticipantName,
			ParticipantVanity:     sc.Thread.ParticipantVanity,
			ParticipantHeadline:   sc.Thread.ParticipantHeadline,
			ParticipantCompany:    sc.Thread.ParticipantCompany,
			Subject:               sc.Thread.Subject,
			ThreadType:            sc.Thread.ThreadType,
			Classification:        linkedin.InboxClassification(sc.ExpectedClassification),
			ActiveFollowUpPlanned: sc.Thread.ActiveFollowUpPlanned,
			Messages:              msgs,
		}

		savedThread, err := svc.SaveConversationThread(ctx, thread)
		if err != nil {
			t.Fatalf("SaveConversationThread failed: %v", err)
		}
		if string(savedThread.Classification) != sc.ExpectedClassification {
			t.Errorf("expected classification %s, got %s", sc.ExpectedClassification, savedThread.Classification)
		}

		// Generate reply draft
		draft, err := svc.GenerateReplyDraft(
			ctx,
			savedThread.ThreadID,
			savedThread.TenantID,
			linkedin.ReplyDraftTone(sc.ReplyGeneration.Tone),
			sc.ReplyGeneration.CandidateAvailability,
			sc.ReplyGeneration.ExpectedFactsUsed,
		)
		if err != nil {
			t.Fatalf("GenerateReplyDraft failed: %v", err)
		}
		if draft.Status != linkedin.ReplyDraftPendingApproval {
			t.Errorf("expected draft status %s, got %s", linkedin.ReplyDraftPendingApproval, draft.Status)
		}
		if !strings.Contains(draft.SuggestedText, sc.Thread.ParticipantCompany) {
			t.Errorf("expected suggested text to mention %q, got %q", sc.Thread.ParticipantCompany, draft.SuggestedText)
		}

		// Approve draft (FND-010, AT-007)
		approved, err := svc.ApproveReplyDraft(ctx, draft.DraftID, sc.TenantID, "user-lead-01")
		if err != nil {
			t.Fatalf("ApproveReplyDraft failed: %v", err)
		}
		if approved.Status != linkedin.ReplyDraftApproved {
			t.Errorf("expected approved status %s, got %s", linkedin.ReplyDraftApproved, approved.Status)
		}
		if approved.ApprovalToken == "" {
			t.Errorf("expected approval token, got empty")
		}

		// Mark copied to clipboard (1-click workflow under AT-010)
		copied, err := svc.MarkReplyDraftCopied(ctx, approved.DraftID, sc.TenantID)
		if err != nil {
			t.Fatalf("MarkReplyDraftCopied failed: %v", err)
		}
		if copied.Status != linkedin.ReplyDraftCopiedToClipboard {
			t.Errorf("expected copied status %s, got %s", linkedin.ReplyDraftCopiedToClipboard, copied.Status)
		}
	})

	// Scenario 2: stop_followup_on_reply_invariant
	t.Run("Scenario 2: Active Follow-Up Schedule Halted by Inbound Message", func(t *testing.T) {
		sc := fix.Scenarios.StopFollowUpOnReply
		thread := &linkedin.ConversationThread{
			ThreadID:              sc.ThreadID,
			WorkspaceID:           sc.WorkspaceID,
			TenantID:              sc.TenantID,
			ParticipantName:       "Sophia Lin",
			ParticipantCompany:    "ScaleAI",
			Subject:               "Technical Follow Up",
			ThreadType:            linkedin.ThreadRecruiter,
			ActiveFollowUpPlanned: sc.InitialActiveFollowUpPlanned,
		}

		_, err := svc.SaveConversationThread(ctx, thread)
		if err != nil {
			t.Fatalf("SaveConversationThread failed: %v", err)
		}

		// Inbound message arrives from recruiter
		sentAt, _ := time.Parse(time.RFC3339, sc.InboundMessage.SentAt)
		inboundMsg := linkedin.InboxMessage{
			MessageID:  sc.InboundMessage.MessageID,
			SenderName: sc.InboundMessage.SenderName,
			SenderType: sc.InboundMessage.SenderType,
			Content:    sc.InboundMessage.Content,
			SentAt:     sentAt,
		}

		updatedThread, halted, err := svc.AddMessageToThread(ctx, thread.ThreadID, sc.TenantID, inboundMsg)
		if err != nil {
			t.Fatalf("AddMessageToThread failed: %v", err)
		}
		if halted != sc.ExpectedFollowUpHalted {
			t.Errorf("expected halted %v, got %v", sc.ExpectedFollowUpHalted, halted)
		}
		if updatedThread.ActiveFollowUpPlanned {
			t.Errorf("Thread active_followup_planned must now be false")
		}
		if updatedThread.Classification != linkedin.ClassFollowUpResponse {
			t.Errorf("expected ClassFollowUpResponse, got %s", updatedThread.Classification)
		}
	})

	// Scenario 3: cross_tenant_inbox_isolation (AT-011, AT-012)
	t.Run("Scenario 3: Cross-Tenant Isolation Barrier", func(t *testing.T) {
		sc := fix.Scenarios.CrossTenantInboxIsolation
		thread := &linkedin.ConversationThread{
			ThreadID:              sc.AlphaThread.ThreadID,
			WorkspaceID:           sc.AlphaThread.WorkspaceID,
			TenantID:              sc.AlphaThread.TenantID,
			ParticipantName:       sc.AlphaThread.ParticipantName,
			ThreadType:            linkedin.ThreadRecruiter,
			ActiveFollowUpPlanned: false,
		}

		_, err := svc.SaveConversationThread(ctx, thread)
		if err != nil {
			t.Fatalf("SaveConversationThread failed: %v", err)
		}

		// Attempting access with beta tenant must be rejected
		_, err = svc.GetConversationThread(ctx, thread.ThreadID, sc.BetaQuery.TenantID)
		if !errors.Is(err, linkedin.ErrCrossTenantAccessDenied) {
			t.Errorf("expected ErrCrossTenantAccessDenied, got %v", err)
		}

		// Draft access with wrong tenant must also fail
		draft, err := svc.GenerateReplyDraft(ctx, thread.ThreadID, sc.AlphaThread.TenantID, linkedin.ToneConciseScheduling, "", nil)
		if err != nil {
			t.Fatalf("GenerateReplyDraft failed: %v", err)
		}

		_, err = svc.GetReplyDraft(ctx, draft.DraftID, sc.BetaQuery.TenantID)
		if !errors.Is(err, linkedin.ErrCrossTenantAccessDenied) {
			t.Errorf("expected ErrCrossTenantAccessDenied for draft, got %v", err)
		}
	})

	// Scenario 4: unsupported_direct_send_fail_closed (AT-010, REQ-015)
	t.Run("Scenario 4: Headless Direct Background Sending Rejection", func(t *testing.T) {
		sc := fix.Scenarios.UnsupportedDirectSend
		err := linkedin.RejectUnsupportedDirectSend(sc.AttemptedMode)
		if !errors.Is(err, linkedin.ErrUnsupportedDirectSend) {
			t.Errorf("expected ErrUnsupportedDirectSend, got %v", err)
		}

		err = linkedin.RejectUnsupportedDirectSend("background_auto_send")
		if !errors.Is(err, linkedin.ErrUnsupportedDirectSend) {
			t.Errorf("expected ErrUnsupportedDirectSend, got %v", err)
		}

		err = linkedin.RejectUnsupportedDirectSend("manual_reviewed_dispatch")
		if err != nil {
			t.Errorf("expected nil error for manual_reviewed_dispatch, got %v", err)
		}
	})
}

func TestInbox_TamperInvalidation(t *testing.T) {
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	thread := &linkedin.ConversationThread{
		ThreadID:        "thread-tamper-01",
		WorkspaceID:     "ws-tamper",
		TenantID:        "tenant-tamper",
		ParticipantName: "Alex Rivera",
		ThreadType:      linkedin.ThreadRecruiter,
	}
	_, err := svc.SaveConversationThread(ctx, thread)
	if err != nil {
		t.Fatalf("SaveConversationThread failed: %v", err)
	}

	draft, err := svc.GenerateReplyDraft(ctx, thread.ThreadID, thread.TenantID, linkedin.ToneConciseScheduling, "Friday 2 PM", nil)
	if err != nil {
		t.Fatalf("GenerateReplyDraft failed: %v", err)
	}

	approved, err := svc.ApproveReplyDraft(ctx, draft.DraftID, thread.TenantID, "approver-01")
	if err != nil {
		t.Fatalf("ApproveReplyDraft failed: %v", err)
	}
	if approved.Status != linkedin.ReplyDraftApproved {
		t.Errorf("expected status %s, got %s", linkedin.ReplyDraftApproved, approved.Status)
	}
	if approved.ApprovalToken == "" {
		t.Errorf("expected non-empty approval token")
	}

	// User modifies approved draft text -> should tamper-invalidate token and reset to draft
	edited, err := svc.UpdateReplyDraftText(ctx, draft.DraftID, thread.TenantID, "Hi Alex, actually let's connect next Monday at 10 AM.")
	if err != nil {
		t.Fatalf("UpdateReplyDraftText failed: %v", err)
	}
	if edited.Status != linkedin.ReplyDraftPendingApproval {
		t.Errorf("editing text must reset status to draft, got %s", edited.Status)
	}
	if edited.ApprovalToken != "" {
		t.Errorf("editing text must invalidate approval token, got %q", edited.ApprovalToken)
	}

	// Attempting to copy unapproved edited draft should fail
	_, err = svc.MarkReplyDraftCopied(ctx, edited.DraftID, thread.TenantID)
	if err == nil {
		t.Errorf("expected error copying unapproved draft, got nil")
	}

	// Re-approving produces valid token and copy succeeds
	reApproved, err := svc.ApproveReplyDraft(ctx, edited.DraftID, thread.TenantID, "approver-01")
	if err != nil {
		t.Fatalf("ApproveReplyDraft failed: %v", err)
	}
	copied, err := svc.MarkReplyDraftCopied(ctx, reApproved.DraftID, thread.TenantID)
	if err != nil {
		t.Fatalf("MarkReplyDraftCopied failed: %v", err)
	}
	if copied.Status != linkedin.ReplyDraftCopiedToClipboard {
		t.Errorf("expected copied status, got %s", copied.Status)
	}
}

func TestInbox_ClassificationAndSearch(t *testing.T) {
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	// Semantic classification tests
	if linkedin.ClassifyConversation("Would you have 15 mins for a screen call this week?") != linkedin.ClassInterviewInvitation {
		t.Errorf("expected ClassInterviewInvitation")
	}
	if linkedin.ClassifyConversation("Our engineering team has an open opportunity for a Staff Engineer.") != linkedin.ClassRecruiterInquiry {
		t.Errorf("expected ClassRecruiterInquiry")
	}
	if linkedin.ClassifyConversation("Thanks for following up! We reviewed your profile.") != linkedin.ClassFollowUpResponse {
		t.Errorf("expected ClassFollowUpResponse")
	}
	if linkedin.ClassifyConversation("Join our upcoming webinar for a 50% discount!") != linkedin.ClassSpamOrPromo {
		t.Errorf("expected ClassSpamOrPromo")
	}
	if linkedin.ClassifyConversation("Great connecting with you at the conference.") != linkedin.ClassNetworking {
		t.Errorf("expected ClassNetworking")
	}

	// Seed multiple threads
	t1 := &linkedin.ConversationThread{
		ThreadID:           "th-search-1",
		WorkspaceID:        "ws-search",
		TenantID:           "ten-search",
		ParticipantName:    "Sarah Connor",
		ParticipantCompany: "Cyberdyne",
		Subject:            "Staff Platform Architect",
		LastMessageSnippet: "Let's set up an interview call.",
		UnreadCount:        1,
		ThreadType:         linkedin.ThreadRecruiter,
		Classification:     linkedin.ClassInterviewInvitation,
	}
	t2 := &linkedin.ConversationThread{
		ThreadID:           "th-search-2",
		WorkspaceID:        "ws-search",
		TenantID:           "ten-search",
		ParticipantName:    "John Doe",
		ParticipantCompany: "Acme Corp",
		Subject:            "Coffee chat",
		LastMessageSnippet: "Great connecting with you.",
		UnreadCount:        0,
		ThreadType:         linkedin.ThreadPeerConnection,
		Classification:     linkedin.ClassNetworking,
	}
	_, err := svc.SaveConversationThread(ctx, t1)
	if err != nil {
		t.Fatalf("SaveConversationThread failed: %v", err)
	}
	_, err = svc.SaveConversationThread(ctx, t2)
	if err != nil {
		t.Fatalf("SaveConversationThread failed: %v", err)
	}

	// Filter by unread
	threads, err := svc.ListConversationThreads(ctx, linkedin.ConversationThreadFilter{
		WorkspaceID: "ws-search",
		TenantID:    "ten-search",
		OnlyUnread:  true,
	})
	if err != nil {
		t.Fatalf("ListConversationThreads failed: %v", err)
	}
	if len(threads) != 1 || threads[0].ThreadID != "th-search-1" {
		t.Errorf("expected 1 unread thread 'th-search-1', got %v", threads)
	}

	// Search query by participant company
	threads, err = svc.ListConversationThreads(ctx, linkedin.ConversationThreadFilter{
		WorkspaceID: "ws-search",
		TenantID:    "ten-search",
		Query:       "cyberdyne",
	})
	if err != nil {
		t.Fatalf("ListConversationThreads failed: %v", err)
	}
	if len(threads) != 1 || threads[0].ThreadID != "th-search-1" {
		t.Errorf("expected 1 thread matching query 'cyberdyne', got %v", threads)
	}

	// Reject draft
	draft, err := svc.GenerateReplyDraft(ctx, t1.ThreadID, t1.TenantID, linkedin.TonePoliteDecline, "", nil)
	if err != nil {
		t.Fatalf("GenerateReplyDraft failed: %v", err)
	}
	rejected, err := svc.RejectReplyDraft(ctx, draft.DraftID, t1.TenantID, "Not interested in defense tech")
	if err != nil {
		t.Fatalf("RejectReplyDraft failed: %v", err)
	}
	if rejected.Status != linkedin.ReplyDraftRejected {
		t.Errorf("expected status %s, got %s", linkedin.ReplyDraftRejected, rejected.Status)
	}
	if !strings.Contains(rejected.Rationale, "Not interested in defense tech") {
		t.Errorf("expected rationale to contain rejection reason, got %q", rejected.Rationale)
	}
}
