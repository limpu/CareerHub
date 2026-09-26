package linkedin

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type ContentPlanningFixtureSuite struct {
	Version   string `json:"version"`
	Task      string `json:"task"`
	Scenarios []struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		WorkspaceID    string `json:"workspace_id,omitempty"`
		SourceArtifact struct {
			ArtifactID   string   `json:"artifact_id"`
			Title        string   `json:"title"`
			ArtifactType string   `json:"artifact_type"`
			RawText      string   `json:"raw_text"`
			Author       string   `json:"author"`
			Tags         []string `json:"tags"`
		} `json:"source_artifact,omitempty"`
		TargetFormats []string `json:"target_formats,omitempty"`
	} `json:"scenarios"`
}

func loadContentPlanningFixture(t *testing.T) ContentPlanningFixtureSuite {
	t.Helper()
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_content_planning.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture %s: %v", fixturePath, err)
	}
	var suite ContentPlanningFixtureSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("Failed to unmarshal fixture JSON: %v", err)
	}
	return suite
}

func TestContentPlanning_FixtureValidation(t *testing.T) {
	suite := loadContentPlanningFixture(t)
	if suite.Task != "IMP-LI-14" {
		t.Errorf("Expected task IMP-LI-14, got %s", suite.Task)
	}
	if len(suite.Scenarios) != 4 {
		t.Errorf("Expected 4 scenarios, got %d", len(suite.Scenarios))
	}
}

func TestRepurposeContentArtifact_AllFormats(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	source := SourceArtifact{
		ID:           "art-redis-001",
		WorkspaceID:  "ws-career-prod",
		TenantID:     "ws-career-prod",
		Title:        "Scaling Distributed Caches: From 10k to 1M QPS with Redis & Go",
		ArtifactType: ArtifactTypeTechnicalBlog,
		RawText: `Scaling our distributed caching layer from 10,000 to 1,000,000 queries per second required three critical architectural pivots:
1. Moving from naive cache-aside to deterministic read-through caching with single-flight mutex coalescing.
2. Partitioning hot keys across consistent hashing rings with virtual nodes to eliminate cache cluster hot-spots.
3. Implementing jittered TTL expiration to prevent catastrophic thundering herd effects during batch cache evictions.

Key lesson: 90% of cache degradation at peak is self-inflicted by lock-step expiration and un-throttled cache stampedes. By instrumenting distributed tracing with OpenTelemetry and pinning p99 latencies below 4ms, we cut Redis cluster CPU utilization by 42% while handling 100x traffic.`,
		Author:    "SRE Lead",
		Tags:      []string{"distributed-systems", "redis", "golang", "scalability"},
		CreatedAt: time.Now().UTC(),
	}

	drafts, err := svc.RepurposeArtifactAllFormats(ctx, source)
	if err != nil {
		t.Fatalf("RepurposeArtifactAllFormats failed: %v", err)
	}
	if len(drafts) != 5 {
		t.Fatalf("Expected 5 drafts, got %d", len(drafts))
	}

	formatMap := make(map[RepurposedFormat]*RepurposedDraft)
	for _, d := range drafts {
		formatMap[d.Format] = d
		if d.WorkspaceID != "ws-career-prod" {
			t.Errorf("Expected workspace ws-career-prod, got %s", d.WorkspaceID)
		}
		if d.Status != PlanStatusDraft {
			t.Errorf("Expected draft status, got %s", d.Status)
		}
		if d.CharacterCount <= 50 {
			t.Errorf("Expected character count > 50, got %d", d.CharacterCount)
		}
		if !strings.Contains(d.ContentBody, "#distributedsystems") {
			t.Errorf("Expected hashtags in body, got: %s", d.ContentBody)
		}
	}

	// 1. Single Thought
	st := formatMap[FormatSingleThought]
	if st == nil {
		t.Fatal("FormatSingleThought draft is nil")
	}
	if !strings.Contains(st.Title, "Thought:") {
		t.Errorf("Expected title with 'Thought:', got %s", st.Title)
	}
	if !strings.Contains(st.ContentBody, "Most engineering teams overlook this") {
		t.Errorf("Expected single thought hook in body, got %s", st.ContentBody)
	}

	// 2. Carousel Outline
	co := formatMap[FormatCarouselOutline]
	if co == nil {
		t.Fatal("FormatCarouselOutline draft is nil")
	}
	if !strings.Contains(co.Title, "Carousel:") {
		t.Errorf("Expected title with 'Carousel:', got %s", co.Title)
	}
	if !strings.Contains(co.ContentBody, "SLIDE 1: Title") || !strings.Contains(co.ContentBody, "SLIDE 5: Takeaway") {
		t.Errorf("Expected slide markers in body, got %s", co.ContentBody)
	}

	// 3. Actionable Checklist
	ac := formatMap[FormatActionableChecklist]
	if ac == nil {
		t.Fatal("FormatActionableChecklist draft is nil")
	}
	if !strings.Contains(ac.Title, "Checklist:") {
		t.Errorf("Expected title with 'Checklist:', got %s", ac.Title)
	}
	if !strings.Contains(ac.ContentBody, "[ ]") {
		t.Errorf("Expected checklist items with '[ ]', got %s", ac.ContentBody)
	}

	// 4. Contrarian Breakdown
	cb := formatMap[FormatContrarianBreakdown]
	if cb == nil {
		t.Fatal("FormatContrarianBreakdown draft is nil")
	}
	if !strings.Contains(cb.Title, "Contrarian Breakdown:") {
		t.Errorf("Expected title with 'Contrarian Breakdown:', got %s", cb.Title)
	}
	if !strings.Contains(cb.ContentBody, "The Conventional Wisdom") {
		t.Errorf("Expected 'The Conventional Wisdom' in body, got %s", cb.ContentBody)
	}

	// 5. Interview QA Spotlight
	qa := formatMap[FormatInterviewQASpotlight]
	if qa == nil {
		t.Fatal("FormatInterviewQASpotlight draft is nil")
	}
	if !strings.Contains(qa.Title, "Q&A Spotlight:") {
		t.Errorf("Expected title with 'Q&A Spotlight:', got %s", qa.Title)
	}
	if !strings.Contains(qa.ContentBody, "Question:") || !strings.Contains(qa.ContentBody, "Principal Engineer Answer:") {
		t.Errorf("Expected Q&A spotlight markers, got %s", qa.ContentBody)
	}
}

func TestCalculateNextSafeScheduleSlot_DSTTransitions(t *testing.T) {
	// Fall Back transition day in America/New_York: 2026-11-01
	userTZ := "America/New_York"
	loc, err := time.LoadLocation(userTZ)
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	// 9:00 AM EDT on Nov 1, 2026 (UTC-5 in standard time)
	reqTime := time.Date(2026, 11, 1, 9, 0, 0, 0, loc)

	existingSlots := []time.Time{}
	maxDailyBudget := 2

	// First slot assignment
	utcSlot, localFormatted, err := CalculateNextSafeScheduleSlot(reqTime, userTZ, existingSlots, maxDailyBudget)
	if err != nil {
		t.Fatalf("CalculateNextSafeScheduleSlot failed: %v", err)
	}
	if utcSlot.Year() != 2026 || utcSlot.Month() != time.November || utcSlot.Day() != 1 || utcSlot.Hour() != 14 {
		t.Errorf("Expected 2026-11-01 14:00:00 UTC, got %v", utcSlot)
	}
	if !strings.Contains(localFormatted, "EST") {
		t.Errorf("Expected EST timezone formatted string, got %s", localFormatted)
	}

	// Collision detection (< 15 mins apart)
	existingSlots = append(existingSlots, utcSlot)
	collisionTime := reqTime.Add(5 * time.Minute)
	_, _, colErr := CalculateNextSafeScheduleSlot(collisionTime, userTZ, existingSlots, maxDailyBudget)
	if !errors.Is(colErr, ErrSlotCollision) {
		t.Errorf("Expected ErrSlotCollision, got %v", colErr)
	}

	// Safe second slot (2 hours later)
	secondTime := reqTime.Add(2 * time.Hour)
	utcSlot2, _, err2 := CalculateNextSafeScheduleSlot(secondTime, userTZ, existingSlots, maxDailyBudget)
	if err2 != nil {
		t.Fatalf("Second slot failed: %v", err2)
	}
	existingSlots = append(existingSlots, utcSlot2)

	// Exceeding daily budget (max 2)
	thirdTime := reqTime.Add(4 * time.Hour)
	_, _, err3 := CalculateNextSafeScheduleSlot(thirdTime, userTZ, existingSlots, maxDailyBudget)
	if !errors.Is(err3, ErrDailyScheduleBudgetExceeded) {
		t.Errorf("Expected ErrDailyScheduleBudgetExceeded, got %v", err3)
	}

	// Invalid timezone error
	_, _, tzErr := CalculateNextSafeScheduleSlot(reqTime, "Invalid/TimeZone_Bogus", existingSlots, maxDailyBudget)
	if !errors.Is(tzErr, ErrInvalidTimezone) {
		t.Errorf("Expected ErrInvalidTimezone, got %v", tzErr)
	}
}

func TestTamperInvalidationAndApproval(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	source := SourceArtifact{
		ID:           "art-incident-01",
		WorkspaceID:  "ws-alpha",
		TenantID:     "ws-alpha",
		Title:        "Postmortem: Redis Cluster Outage 2026",
		ArtifactType: ArtifactTypeIncidentPostmortem,
		RawText:      "Incident summary: DNS partition caused Redis cluster split brain for 14 minutes.",
		Author:       "Infra Lead",
		CreatedAt:    time.Now().UTC(),
	}

	draft, err := svc.RepurposeArtifact(ctx, source, FormatActionableChecklist)
	if err != nil {
		t.Fatalf("RepurposeArtifact failed: %v", err)
	}

	secretKey := "test-secret-at-007-key"

	// 1. Approve Draft
	approvedDraft, err := svc.ApproveRepurposedDraft(ctx, draft.ID, "ws-alpha", secretKey)
	if err != nil {
		t.Fatalf("ApproveRepurposedDraft failed: %v", err)
	}
	if approvedDraft.Status != PlanStatusApproved {
		t.Errorf("Expected approved status, got %s", approvedDraft.Status)
	}
	if approvedDraft.ApprovalToken == "" {
		t.Errorf("Expected non-empty approval token")
	}
	if !VerifyDraftApprovalToken(secretKey, approvedDraft, approvedDraft.ApprovalToken) {
		t.Errorf("Draft approval token verification failed")
	}

	// 2. Mutate Draft Body -> Post-approval tamper invalidation (AT-007)
	mutatedDraft, err := svc.UpdateRepurposedDraft(ctx, draft.ID, "ws-alpha", "", "Tampered body injected without authorization")
	if err != nil {
		t.Fatalf("UpdateRepurposedDraft failed: %v", err)
	}
	if mutatedDraft.Status != PlanStatusDraft {
		t.Errorf("Expected draft status after mutation, got %s", mutatedDraft.Status)
	}
	if mutatedDraft.ApprovalToken != "" {
		t.Errorf("Expected cleared approval token after mutation, got %s", mutatedDraft.ApprovalToken)
	}
	if mutatedDraft.LastApprovedAt != nil {
		t.Errorf("Expected cleared last approved at timestamp")
	}

	// 3. Re-approve Draft and Schedule Plan Item
	approvedAgain, err := svc.ApproveRepurposedDraft(ctx, draft.ID, "ws-alpha", secretKey)
	if err != nil {
		t.Fatalf("Re-approve failed: %v", err)
	}

	loc, _ := time.LoadLocation("UTC")
	schedTime := time.Date(2026, 10, 15, 12, 0, 0, 0, loc)

	planItem, err := svc.ScheduleContentPlan(ctx, "ws-alpha", "ws-alpha", approvedAgain.ID, schedTime, "UTC", 3)
	if err != nil {
		t.Fatalf("ScheduleContentPlan failed: %v", err)
	}
	if planItem.Status != PlanStatusScheduled {
		t.Errorf("Expected scheduled status, got %s", planItem.Status)
	}
	if !planItem.DirectPublishBlocked {
		t.Errorf("Expected DirectPublishBlocked to be true under AT-010")
	}

	// 4. Approve Content Plan Item
	approvedPlan, err := svc.ApproveContentPlanItem(ctx, planItem.ID, "ws-alpha", secretKey)
	if err != nil {
		t.Fatalf("ApproveContentPlanItem failed: %v", err)
	}
	if approvedPlan.Status != PlanStatusApproved {
		t.Errorf("Expected approved status, got %s", approvedPlan.Status)
	}
	if approvedPlan.ApprovalToken == "" {
		t.Errorf("Expected non-empty plan approval token")
	}
	if !VerifyPlanApprovalToken(secretKey, approvedPlan, approvedPlan.ApprovalToken) {
		t.Errorf("Plan approval token verification failed")
	}

	// 5. Mutate Schedule -> Plan Item Tamper Invalidation
	newSchedTime := schedTime.Add(1 * time.Hour)
	mutatedPlan, err := svc.UpdateContentPlanItem(ctx, planItem.ID, "ws-alpha", "Updated Title", &newSchedTime, "UTC", 3)
	if err != nil {
		t.Fatalf("UpdateContentPlanItem failed: %v", err)
	}
	if mutatedPlan.Status != PlanStatusDraft {
		t.Errorf("Expected draft status after mutation, got %s", mutatedPlan.Status)
	}
	if mutatedPlan.ApprovalToken != "" {
		t.Errorf("Expected cleared plan approval token, got %s", mutatedPlan.ApprovalToken)
	}
}

func TestDirectAutopublishRejection_AT010(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	source := SourceArtifact{
		ID:           "art-rfc-02",
		WorkspaceID:  "ws-career-prod",
		TenantID:     "ws-career-prod",
		Title:        "RFC: Zero-Downtime Database Migration Engine",
		ArtifactType: ArtifactTypeArchitectureRFC,
		RawText:      "We designed an asynchronous replication pipeline to support live schema migrations with zero client timeouts.",
		Author:       "Staff Engineer",
		CreatedAt:    time.Now().UTC(),
	}

	draft, err := svc.RepurposeArtifact(ctx, source, FormatSingleThought)
	if err != nil {
		t.Fatalf("RepurposeArtifact failed: %v", err)
	}

	loc, _ := time.LoadLocation("UTC")
	schedTime := time.Date(2026, 12, 1, 15, 0, 0, 0, loc)

	planItem, err := svc.ScheduleContentPlan(ctx, "ws-career-prod", "ws-career-prod", draft.ID, schedTime, "UTC", 5)
	if err != nil {
		t.Fatalf("ScheduleContentPlan failed: %v", err)
	}

	// Verify direct background publish is rejected under AT-010
	pubErr := svc.AttemptDirectPublish(ctx, planItem.ID, "ws-career-prod")
	if !errors.Is(pubErr, ErrDirectAutopublishDisallowed) {
		t.Errorf("Expected ErrDirectAutopublishDisallowed, got %v", pubErr)
	}

	// Verify 1-click clipboard export and deep link works
	item, md, composeURL, err := svc.ExportPlanForClipboard(ctx, planItem.ID, "ws-career-prod", "key")
	if err != nil {
		t.Fatalf("ExportPlanForClipboard failed: %v", err)
	}
	if item == nil {
		t.Fatal("Expected non-nil plan item")
	}
	if !strings.HasPrefix(composeURL, "https://www.linkedin.com/feed/?shareActive=true&text=") {
		t.Errorf("Expected LinkedIn Compose URL, got %s", composeURL)
	}
	if !strings.Contains(md, "# Thought: RFC: Zero-Downtime Database Migration Engine") {
		t.Errorf("Expected markdown title, got %s", md)
	}
	if !strings.Contains(md, "```linkedin-draft") {
		t.Errorf("Expected code block fence, got %s", md)
	}
}

func TestContentPlanning_MultiTenantIsolation(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)

	sourceAlpha := SourceArtifact{
		ID:           "art-alpha-01",
		WorkspaceID:  "ws-alpha",
		TenantID:     "ws-alpha",
		Title:        "Alpha Confidential Tech Note",
		ArtifactType: ArtifactTypeTechnicalBlog,
		RawText:      "Alpha team architecture notes for internal eyes only.",
		Author:       "Lead Alpha",
		CreatedAt:    time.Now().UTC(),
	}

	draftAlpha, err := svc.RepurposeArtifact(ctx, sourceAlpha, FormatSingleThought)
	if err != nil {
		t.Fatalf("RepurposeArtifact failed: %v", err)
	}

	// Beta tenant attempts to query Alpha draft
	_, err = svc.GetRepurposedDraft(ctx, draftAlpha.ID, "ws-beta")
	if !errors.Is(err, ErrCrossTenantAccessDenied) {
		t.Errorf("Expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// Beta tenant attempts to schedule Alpha draft
	loc, _ := time.LoadLocation("UTC")
	schedTime := time.Date(2026, 10, 10, 10, 0, 0, 0, loc)
	_, err = svc.ScheduleContentPlan(ctx, "ws-beta", "ws-beta", draftAlpha.ID, schedTime, "UTC", 2)
	if !errors.Is(err, ErrCrossTenantAccessDenied) {
		t.Errorf("Expected ErrCrossTenantAccessDenied, got %v", err)
	}
}
