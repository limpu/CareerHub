package linkedin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type connectionQueueFixtureSuite struct {
	Description string `json:"description"`
	Scenarios   []struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Description    string `json:"description"`
		WorkspaceID    string `json:"workspace_id"`
		TenantID       string `json:"tenant_id"`
		ExpectedError  string `json:"expected_error,omitempty"`
		ApprovedItemID string `json:"approved_item_id,omitempty"`
		OriginalText   string `json:"original_text,omitempty"`
		TamperedText   string `json:"tampered_text,omitempty"`
	} `json:"scenarios"`
}

func loadConnectionQueueFixtures(t *testing.T) *connectionQueueFixtureSuite {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_connection_queue.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read linkedin_connection_queue.json: %v", err)
	}

	var suite connectionQueueFixtureSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("failed to unmarshal linkedin_connection_queue.json: %v", err)
	}
	return &suite
}

func TestConnectionQueue_DraftGeneration(t *testing.T) {
	note, factors := GenerateConnectionNoteDraft(
		"Alex Chen",
		"Principal Infrastructure Architect",
		"Sarah Chen",
		"CloudScale Systems",
		"Lead Systems Architect",
		"Distributed Systems, Go, Kubernetes",
	)

	if note == "" {
		t.Fatal("expected non-empty connection note draft")
	}
	if len([]rune(note)) > LinkedInMaxNoteCharacters {
		t.Fatalf("note length %d exceeds max LinkedIn characters %d", len([]rune(note)), LinkedInMaxNoteCharacters)
	}
	if !strings.Contains(note, "Sarah") {
		t.Errorf("expected note to contain recipient name 'Sarah', got: %s", note)
	}
	if !strings.Contains(note, "CloudScale Systems") {
		t.Errorf("expected note to contain company 'CloudScale Systems', got: %s", note)
	}
	if len(factors) == 0 {
		t.Error("expected personalization factors to be present")
	}
}

func TestConnectionQueue_StandardApprovalAndExecutionLifecycle(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	workspaceID := "ws-alpha"
	tenantID := "tenant-alpha"

	// 1. Initial budget setup
	budget := NewConnectionBudget(workspaceID, tenantID)
	budget.DailyUsed = 2
	budget.WeeklyUsed = 14
	if err := svc.SaveConnectionBudget(ctx, budget); err != nil {
		t.Fatalf("failed to save budget: %v", err)
	}

	// 2. Draft note and enqueue item
	note, factors := svc.DraftConnectionNote(
		"Alex Chen",
		"Principal Infrastructure Architect",
		"Sarah Chen",
		"CloudScale Systems",
		"Lead Systems Architect",
		"Distributed Systems, Go, Kubernetes",
	)

	item := &ConnectionQueueItem{
		ID:                   "cq-test-001",
		WorkspaceID:          workspaceID,
		TenantID:             tenantID,
		RecipientID:          "rec-p-01",
		RecipientName:        "Sarah Chen",
		RecipientTitle:       "Lead Systems Architect",
		RecipientCompany:     "CloudScale Systems",
		RecipientLinkedInURL: "https://www.linkedin.com/in/sarah-chen-arch",
		NoteText:             note,
		ContextFactors:       factors,
	}

	enqueued, err := svc.EnqueueConnectionNote(ctx, item)
	if err != nil {
		t.Fatalf("failed to enqueue connection note: %v", err)
	}
	if enqueued.Status != QueueItemPendingApproval {
		t.Fatalf("expected status %s, got %s", QueueItemPendingApproval, enqueued.Status)
	}
	if !enqueued.WithinLimit {
		t.Fatal("expected item to be within 300 character limit")
	}
	if enqueued.ApprovalToken != "" {
		t.Fatal("approval token should be empty before explicit approval")
	}

	// Trying to mark copied without approval fails closed (AT-007)
	_, err = svc.MarkConnectionNoteCopied(ctx, enqueued.ID, tenantID)
	if err != ErrNotApproved {
		t.Fatalf("expected ErrNotApproved, got %v", err)
	}

	// 3. Human-in-the-loop explicit approval (FND-010, AT-007)
	approved, err := svc.ApproveConnectionQueueItem(ctx, enqueued.ID, tenantID, "Alex Chen (operator)")
	if err != nil {
		t.Fatalf("failed to approve item: %v", err)
	}
	if approved.Status != QueueItemApproved {
		t.Fatalf("expected status %s, got %s", QueueItemApproved, approved.Status)
	}
	if approved.ApprovalToken == "" {
		t.Fatal("expected non-empty approval token")
	}
	if approved.ApprovedAt == nil {
		t.Fatal("expected ApprovedAt timestamp to be set")
	}
	if approved.ApprovedBy != "Alex Chen (operator)" {
		t.Fatalf("expected approver 'Alex Chen (operator)', got %s", approved.ApprovedBy)
	}

	// 4. Mark copied to clipboard (AT-010 manual fallback)
	copied, err := svc.MarkConnectionNoteCopied(ctx, enqueued.ID, tenantID)
	if err != nil {
		t.Fatalf("failed to mark copied: %v", err)
	}
	if copied.Status != QueueItemCopied {
		t.Fatalf("expected status %s, got %s", QueueItemCopied, copied.Status)
	}

	// 5. Confirm sent manually (updates rolling budget counters)
	sent, err := svc.ConfirmConnectionSent(ctx, enqueued.ID, tenantID)
	if err != nil {
		t.Fatalf("failed to confirm sent: %v", err)
	}
	if sent.Status != QueueItemCompleted {
		t.Fatalf("expected status %s, got %s", QueueItemCompleted, sent.Status)
	}
	if sent.CompletedAt == nil {
		t.Fatal("expected CompletedAt to be set")
	}

	// Verify budget was consumed
	updatedBudget, err := svc.GetConnectionBudget(ctx, workspaceID, tenantID)
	if err != nil {
		t.Fatalf("failed to get budget: %v", err)
	}
	if updatedBudget.DailyUsed != 3 {
		t.Fatalf("expected daily used 3, got %d", updatedBudget.DailyUsed)
	}
	if updatedBudget.WeeklyUsed != 15 {
		t.Fatalf("expected weekly used 15, got %d", updatedBudget.WeeklyUsed)
	}
	if updatedBudget.RemainingDaily() != 17 {
		t.Fatalf("expected remaining daily 17, got %d", updatedBudget.RemainingDaily())
	}
	if updatedBudget.RemainingWeekly() != 65 {
		t.Fatalf("expected remaining weekly 65, got %d", updatedBudget.RemainingWeekly())
	}
}

func TestConnectionQueue_RecipientDeduplication(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	workspaceID := "ws-alpha"
	tenantID := "tenant-alpha"

	item1 := &ConnectionQueueItem{
		ID:                   "cq-dup-01",
		WorkspaceID:          workspaceID,
		TenantID:             tenantID,
		RecipientName:        "Marcus Vance",
		RecipientLinkedInURL: "https://www.linkedin.com/in/marcus-vance-talent",
		NoteText:             "Hi Marcus, saw your talent posts at Stripe and wanted to connect.",
	}
	_, err := svc.EnqueueConnectionNote(ctx, item1)
	if err != nil {
		t.Fatalf("failed to enqueue item 1: %v", err)
	}

	// Attempting duplicate recipient in same workspace must fail closed (AT-007, AT-004)
	item2 := &ConnectionQueueItem{
		ID:                   "cq-dup-02",
		WorkspaceID:          workspaceID,
		TenantID:             tenantID,
		RecipientName:        "Marcus Vance",
		RecipientLinkedInURL: "https://www.linkedin.com/in/marcus-vance-talent/", // with trailing slash
		NoteText:             "Hi Marcus, connecting again regarding roles.",
	}
	_, err = svc.EnqueueConnectionNote(ctx, item2)
	if err != ErrDuplicateRecipient {
		t.Fatalf("expected ErrDuplicateRecipient, got %v", err)
	}
}

func TestConnectionQueue_BudgetExhaustionThrottling(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	workspaceID := "ws-throttled"
	tenantID := "tenant-throttled"

	// 1. Daily limit exhausted
	exhaustedDailyBudget := &ConnectionBudget{
		WorkspaceID:   workspaceID,
		TenantID:      tenantID,
		DailyLimit:    20,
		DailyUsed:     20,
		WeeklyLimit:   80,
		WeeklyUsed:    40,
		LastResetDate: time.Now().UTC(),
	}
	if err := svc.SaveConnectionBudget(ctx, exhaustedDailyBudget); err != nil {
		t.Fatalf("failed to save daily exhausted budget: %v", err)
	}

	item := &ConnectionQueueItem{
		ID:                   "cq-limit-01",
		WorkspaceID:          workspaceID,
		TenantID:             tenantID,
		RecipientName:        "Sophia Lin",
		RecipientLinkedInURL: "https://www.linkedin.com/in/sophia-lin-recruiter",
		NoteText:             "Hi Sophia, wanted to connect.",
	}
	enqueued, err := svc.EnqueueConnectionNote(ctx, item)
	if err != nil {
		t.Fatalf("failed to enqueue item: %v", err)
	}

	// Approval should be blocked due to daily budget exhaustion (FND-011, AT-008)
	_, err = svc.ApproveConnectionQueueItem(ctx, enqueued.ID, tenantID, "operator")
	if err != ErrDailyBudgetExceeded {
		t.Fatalf("expected ErrDailyBudgetExceeded, got %v", err)
	}

	// 2. Weekly limit exhausted
	exhaustedWeeklyBudget := &ConnectionBudget{
		WorkspaceID:   workspaceID,
		TenantID:      tenantID,
		DailyLimit:    20,
		DailyUsed:     5,
		WeeklyLimit:   80,
		WeeklyUsed:    80,
		LastResetDate: time.Now().UTC(),
	}
	if err := svc.SaveConnectionBudget(ctx, exhaustedWeeklyBudget); err != nil {
		t.Fatalf("failed to save weekly exhausted budget: %v", err)
	}

	_, err = svc.ApproveConnectionQueueItem(ctx, enqueued.ID, tenantID, "operator")
	if err != ErrWeeklyBudgetExceeded {
		t.Fatalf("expected ErrWeeklyBudgetExceeded, got %v", err)
	}
}

func TestConnectionQueue_TamperInvalidation(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	workspaceID := "ws-alpha"
	tenantID := "tenant-alpha"

	item := &ConnectionQueueItem{
		ID:                   "cq-tamper-01",
		WorkspaceID:          workspaceID,
		TenantID:             tenantID,
		RecipientName:        "David Kim",
		RecipientLinkedInURL: "https://www.linkedin.com/in/david-kim-eng",
		NoteText:             "Hi David, noticed your engineering post at CloudScale Systems and would love to connect.",
	}
	enqueued, err := svc.EnqueueConnectionNote(ctx, item)
	if err != nil {
		t.Fatalf("failed to enqueue item: %v", err)
	}

	// Approve original payload
	approved, err := svc.ApproveConnectionQueueItem(ctx, enqueued.ID, tenantID, "Alex Chen")
	if err != nil {
		t.Fatalf("failed to approve item: %v", err)
	}
	if approved.Status != QueueItemApproved {
		t.Fatalf("expected status %s, got %s", QueueItemApproved, approved.Status)
	}
	if approved.ApprovalToken == "" {
		t.Fatal("expected non-empty approval token")
	}

	// Tamper/edit note text after approval (FND-010, AT-007)
	tampered, err := svc.UpdateConnectionNoteText(ctx, enqueued.ID, tenantID, "Hi David, buy our product immediately!")
	if err != nil {
		t.Fatalf("failed to update note text: %v", err)
	}
	if tampered.Status != QueueItemPendingApproval {
		t.Fatalf("expected status %s after tampering, got %s", QueueItemPendingApproval, tampered.Status)
	}
	if tampered.ApprovalToken != "" {
		t.Fatal("tampering must clear approval token")
	}
	if tampered.ApprovedAt != nil {
		t.Fatal("tampering must clear ApprovedAt")
	}

	// Copying unapproved tampered note fails closed
	_, err = svc.MarkConnectionNoteCopied(ctx, enqueued.ID, tenantID)
	if err != ErrNotApproved {
		t.Fatalf("expected ErrNotApproved, got %v", err)
	}
}

func TestConnectionQueue_FixturesIntegrity(t *testing.T) {
	suite := loadConnectionQueueFixtures(t)
	if len(suite.Scenarios) != 4 {
		t.Fatalf("expected 4 scenarios in fixture, got %d", len(suite.Scenarios))
	}
}
