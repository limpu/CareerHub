package linkedin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type StoryBankFixtureSuite struct {
	Scenarios []struct {
		ID                        string          `json:"id"`
		Description               string          `json:"description"`
		WorkspaceID               string          `json:"workspace_id,omitempty"`
		WorkspaceAlpha            string          `json:"workspace_alpha,omitempty"`
		WorkspaceBeta             string          `json:"workspace_beta,omitempty"`
		StoryAlphaID              string          `json:"story_alpha_id,omitempty"`
		Entry                     *StoryEntry     `json:"entry,omitempty"`
		ViolatingEntry            *StoryEntry     `json:"violating_entry,omitempty"`
		VerifiedFacts             []string        `json:"verified_facts,omitempty"`
		CandidateGroundtruthFacts []string        `json:"candidate_groundtruth_facts,omitempty"`
		ExpectedAuditIssues       []struct {
			Field    string `json:"field"`
			Metric   string `json:"metric"`
			Severity string `json:"severity"`
			Message  string `json:"message"`
		} `json:"expected_audit_issues,omitempty"`
		EditedNarrativeHook    string `json:"edited_narrative_hook,omitempty"`
		ExpectedPostEditStatus string `json:"expected_post_edit_status,omitempty"`
		ExpectedPostEditToken  string `json:"expected_post_edit_token,omitempty"`
	} `json:"scenarios"`
}

func loadStoryBankFixtures(t *testing.T) StoryBankFixtureSuite {
	t.Helper()
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_story_bank.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read linkedin_story_bank.json fixture: %v", err)
	}

	var suite StoryBankFixtureSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("Failed to unmarshal linkedin_story_bank.json fixture: %v", err)
	}
	return suite
}

func TestStoryBank_FixtureGroundTruth(t *testing.T) {
	suite := loadStoryBankFixtures(t)
	if len(suite.Scenarios) != 4 {
		t.Fatalf("Expected 4 scenarios, got %d", len(suite.Scenarios))
	}

	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()
	secretKey := "test_story_secret_key"

	for _, s := range suite.Scenarios {
		t.Run(s.ID, func(t *testing.T) {
			switch s.ID {
			case "distributed_systems_outage_story":
				if s.Entry == nil {
					t.Fatalf("Expected non-nil entry")
				}
				s.Entry.TenantID = "ws-alpha"

				// Create story entry
				created, err := svc.CreateStoryEntry(ctx, s.Entry, s.VerifiedFacts)
				if err != nil {
					t.Fatalf("CreateStoryEntry failed: %v", err)
				}
				if created.ApprovalStatus != StoryStatusDraft {
					t.Errorf("Expected draft status, got %s", created.ApprovalStatus)
				}
				if created.AuditReport.HasBlockingIssues {
					t.Errorf("Expected zero blocking issues, got %v", created.AuditReport.Issues)
				}

				// Approve story entry
				approved, err := svc.ApproveStoryEntry(ctx, created.ID, "ws-alpha", secretKey, s.VerifiedFacts)
				if err != nil {
					t.Fatalf("ApproveStoryEntry failed: %v", err)
				}
				if approved.ApprovalStatus != StoryStatusApproved {
					t.Errorf("Expected approved status, got %s", approved.ApprovalStatus)
				}
				if !strings.HasPrefix(approved.ApprovalToken, "story_hmac_") {
					t.Errorf("Expected story_hmac_ prefix, got %s", approved.ApprovalToken)
				}

				// Copy story entry
				copied, md, err := svc.MarkStoryEntryCopied(ctx, created.ID, "ws-alpha", secretKey)
				if err != nil {
					t.Fatalf("MarkStoryEntryCopied failed: %v", err)
				}
				if copied.ApprovalStatus != StoryStatusApproved {
					t.Errorf("Expected approved status, got %s", copied.ApprovalStatus)
				}
				if !strings.Contains(md, s.Entry.Title) {
					t.Errorf("Markdown missing title: %s", md)
				}
				if !strings.Contains(md, "Redis") {
					t.Errorf("Markdown missing 'Redis': %s", md)
				}

			case "career_turning_point_interview":
				// Start interview session
				session, err := svc.StartInterviewSession(ctx, "ws-alpha", "ws-alpha", StoryCategoryTurningPoint)
				if err != nil {
					t.Fatalf("StartInterviewSession failed: %v", err)
				}
				if session.Status != "in_progress" {
					t.Errorf("Expected in_progress status, got %s", session.Status)
				}

				// Record answer
				session, err = svc.RecordInterviewAnswer(ctx, session.ID, "ws-alpha", "prompt-tp-01", "We were rewriting our billing service for the third time in two years.")
				if err != nil {
					t.Fatalf("RecordInterviewAnswer failed: %v", err)
				}
				if len(session.QuestionsAndAnswers) != 1 {
					t.Fatalf("Expected 1 QA item, got %d", len(session.QuestionsAndAnswers))
				}

				// Record follow-up answer
				session, err = svc.RecordInterviewFollowUpAnswer(ctx, session.ID, "ws-alpha", "prompt-tp-01", "I drafted the domain contract RFC for our billing events.")
				if err != nil {
					t.Fatalf("RecordInterviewFollowUpAnswer failed: %v", err)
				}
				if session.QuestionsAndAnswers[0].CandidateFollowUpAnswer != "I drafted the domain contract RFC for our billing events." {
					t.Errorf("Follow-up answer mismatch: %s", session.QuestionsAndAnswers[0].CandidateFollowUpAnswer)
				}

				// Synthesize into story
				synthesized, err := svc.SynthesizeInterviewStory(ctx, session.ID, "ws-alpha", "David Miller", []string{"billing events"})
				if err != nil {
					t.Fatalf("SynthesizeInterviewStory failed: %v", err)
				}
				if synthesized.Category != StoryCategoryTurningPoint {
					t.Errorf("Expected category %s, got %s", StoryCategoryTurningPoint, synthesized.Category)
				}
				if synthesized.ApprovalStatus != StoryStatusDraft {
					t.Errorf("Expected draft status, got %s", synthesized.ApprovalStatus)
				}
				if synthesized.Provenance.SourceType != StoryProvenanceInterviewSession {
					t.Errorf("Expected source %s, got %s", StoryProvenanceInterviewSession, synthesized.Provenance.SourceType)
				}

			case "unverified_metric_hallucination_guard":
				if s.ViolatingEntry == nil {
					t.Fatalf("Expected non-nil violating entry")
				}
				s.ViolatingEntry.TenantID = "ws-alpha"

				// Audit should flag blocking issues because $12M and 500% are not in candidate facts
				audit := AuditStoryProvenanceAndMetrics(s.ViolatingEntry.Narrative, s.CandidateGroundtruthFacts)
				if !audit.HasBlockingIssues {
					t.Errorf("Expected blocking issues for unverified metrics")
				}
				if len(audit.Issues) < 2 {
					t.Errorf("Expected at least 2 issues, got %d", len(audit.Issues))
				}

				// Approval MUST be rejected under AT-003
				_, err := svc.CreateStoryEntry(ctx, s.ViolatingEntry, s.CandidateGroundtruthFacts)
				if err != nil {
					t.Fatalf("CreateStoryEntry failed: %v", err)
				}

				_, err = svc.ApproveStoryEntry(ctx, s.ViolatingEntry.ID, "ws-alpha", secretKey, s.CandidateGroundtruthFacts)
				if err == nil {
					t.Fatalf("Expected ApproveStoryEntry to fail for unverified metrics under AT-003")
				}
				if !strings.Contains(err.Error(), "unverified metric issue") {
					t.Errorf("Unexpected error message: %v", err)
				}

			case "tamper_invalidation_and_multi_tenant":
				entry := &StoryEntry{
					ID:          s.StoryAlphaID,
					WorkspaceID: s.WorkspaceAlpha,
					TenantID:    s.WorkspaceAlpha,
					Title:       "Original Story Title",
					Category:    StoryCategoryScarOrFailure,
					Provenance: StoryProvenance{
						SourceType: StoryProvenanceManualEntry,
						AuthorName: "Elena Rostova",
					},
					Narrative: StoryNarrative{
						HookSummary:       "Original hook statement.",
						ContextBackground: "Context info.",
						ChallengeConflict: "Challenge info.",
						ActionTaken:       "Action info.",
						QuantifiedOutcome: "Outcome info.",
						LessonLearned:     "Lesson info.",
					},
					Tags:           []string{"infra"},
					ApprovalStatus: StoryStatusDraft,
				}
				_, err := svc.CreateStoryEntry(ctx, entry, nil)
				if err != nil {
					t.Fatalf("CreateStoryEntry failed: %v", err)
				}

				// Approve
				approved, err := svc.ApproveStoryEntry(ctx, entry.ID, s.WorkspaceAlpha, secretKey, nil)
				if err != nil {
					t.Fatalf("ApproveStoryEntry failed: %v", err)
				}
				if approved.ApprovalStatus != StoryStatusApproved {
					t.Errorf("Expected approved status, got %s", approved.ApprovalStatus)
				}
				if approved.ApprovalToken == "" {
					t.Errorf("Expected non-empty approval token")
				}

				// Tamper Invalidation (AT-007): Edit narrative hook
				editedNarrative := approved.Narrative
				editedNarrative.HookSummary = s.EditedNarrativeHook
				updated, err := svc.UpdateStoryEntry(ctx, entry.ID, s.WorkspaceAlpha, editedNarrative, nil)
				if err != nil {
					t.Fatalf("UpdateStoryEntry failed: %v", err)
				}

				if updated.ApprovalStatus != StoryStatusDraft {
					t.Errorf("Expected draft status after tamper, got %s", updated.ApprovalStatus)
				}
				if updated.ApprovalToken != "" {
					t.Errorf("Expected empty token after tamper, got %s", updated.ApprovalToken)
				}
				if updated.LastApprovedAt != nil {
					t.Errorf("Expected nil LastApprovedAt after tamper")
				}

				// Cross-tenant isolation (AT-011, AT-012)
				_, err = svc.GetStoryEntry(ctx, entry.ID, s.WorkspaceBeta)
				if err == nil {
					t.Fatalf("Expected cross-tenant access error")
				}
				if err != ErrCrossTenantAccessDenied {
					t.Errorf("Expected ErrCrossTenantAccessDenied, got %v", err)
				}
			}
		})
	}
}

func TestStoryBank_ApprovalAndTamperInvalidation(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()
	secretKey := "test_secret"

	entry := &StoryEntry{
		ID:          "story-test-01",
		WorkspaceID: "ws-test",
		TenantID:    "ws-test",
		Title:       "Testing Incident Postmortem",
		Category:    StoryCategoryScarOrFailure,
		Provenance: StoryProvenance{
			SourceType: StoryProvenanceManualEntry,
			AuthorName: "Test Engineer",
		},
		Narrative: StoryNarrative{
			HookSummary:       "A simple test hook.",
			ContextBackground: "Background information.",
			ChallengeConflict: "Conflict description.",
			ActionTaken:       "Engineered solution.",
			QuantifiedOutcome: "Zero downtime.",
			LessonLearned:     "Always write unit tests.",
		},
	}

	created, err := svc.CreateStoryEntry(ctx, entry, nil)
	if err != nil {
		t.Fatalf("CreateStoryEntry failed: %v", err)
	}

	// Unapproved copying should fail
	_, _, err = svc.MarkStoryEntryCopied(ctx, created.ID, "ws-test", secretKey)
	if err == nil {
		t.Fatalf("Expected error when copying unapproved story")
	}
	if !strings.Contains(err.Error(), "must be approved") {
		t.Errorf("Unexpected error: %v", err)
	}

	// Approve
	approved, err := svc.ApproveStoryEntry(ctx, created.ID, "ws-test", secretKey, nil)
	if err != nil {
		t.Fatalf("ApproveStoryEntry failed: %v", err)
	}
	if approved.ApprovalStatus != StoryStatusApproved {
		t.Errorf("Expected approved status, got %s", approved.ApprovalStatus)
	}

	// Modify text -> resets to draft
	newNarrative := approved.Narrative
	newNarrative.LessonLearned = "Updated lesson learned."
	updated, err := svc.UpdateStoryEntry(ctx, created.ID, "ws-test", newNarrative, nil)
	if err != nil {
		t.Fatalf("UpdateStoryEntry failed: %v", err)
	}
	if updated.ApprovalStatus != StoryStatusDraft {
		t.Errorf("Expected draft status, got %s", updated.ApprovalStatus)
	}
	if updated.ApprovalToken != "" {
		t.Errorf("Expected empty approval token")
	}

	// Archive
	archived, err := svc.ArchiveStoryEntry(ctx, created.ID, "ws-test")
	if err != nil {
		t.Fatalf("ArchiveStoryEntry failed: %v", err)
	}
	if archived.ApprovalStatus != StoryStatusArchived {
		t.Errorf("Expected archived status, got %s", archived.ApprovalStatus)
	}
}

func TestStoryBank_CrossTenantIsolation(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	entry := &StoryEntry{
		ID:          "story-tenant-alpha",
		WorkspaceID: "ws-alpha",
		TenantID:    "ws-alpha",
		Title:       "Alpha Confession",
		Category:    StoryCategoryContrarianBelief,
		Provenance: StoryProvenance{
			SourceType: StoryProvenanceManualEntry,
			AuthorName: "Alpha Candidate",
		},
		Narrative: StoryNarrative{
			HookSummary:       "Hook for alpha.",
			ContextBackground: "Context.",
			ChallengeConflict: "Challenge.",
			ActionTaken:       "Action.",
			QuantifiedOutcome: "Outcome.",
			LessonLearned:     "Lesson.",
		},
	}

	_, err := svc.CreateStoryEntry(ctx, entry, nil)
	if err != nil {
		t.Fatalf("CreateStoryEntry failed: %v", err)
	}

	// Tenant beta cannot view
	_, err = svc.GetStoryEntry(ctx, "story-tenant-alpha", "ws-beta")
	if err != ErrCrossTenantAccessDenied {
		t.Errorf("Expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// Tenant beta cannot update
	_, err = svc.UpdateStoryEntry(ctx, "story-tenant-alpha", "ws-beta", entry.Narrative, nil)
	if err != ErrCrossTenantAccessDenied {
		t.Errorf("Expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// Tenant beta cannot delete
	err = svc.DeleteStoryEntry(ctx, "story-tenant-alpha", "ws-beta")
	if err != ErrCrossTenantAccessDenied {
		t.Errorf("Expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// List filter checks
	alphaList, err := svc.ListStoryEntries(ctx, StoryFilter{
		WorkspaceID:        "ws-alpha",
		RequestingTenantID: "ws-alpha",
	})
	if err != nil {
		t.Fatalf("ListStoryEntries failed: %v", err)
	}
	if len(alphaList) != 1 {
		t.Errorf("Expected 1 item for alpha, got %d", len(alphaList))
	}

	betaList, err := svc.ListStoryEntries(ctx, StoryFilter{
		WorkspaceID:        "ws-alpha",
		RequestingTenantID: "ws-beta",
	})
	if err != nil {
		t.Fatalf("ListStoryEntries failed: %v", err)
	}
	if len(betaList) != 0 {
		t.Errorf("Expected 0 items for beta, got %d", len(betaList))
	}
}

func TestStoryBank_CuratedPrompts(t *testing.T) {
	prompts := GetCuratedInterviewPrompts()
	if len(prompts) < 5 {
		t.Errorf("Expected at least 5 prompts, got %d", len(prompts))
	}

	tpPrompt := GetCuratedPromptByID("prompt-tp-01")
	if tpPrompt == nil {
		t.Fatalf("Expected prompt-tp-01 to be found")
	}
	if tpPrompt.Category != StoryCategoryTurningPoint {
		t.Errorf("Expected turning_point category, got %s", tpPrompt.Category)
	}
	if len(tpPrompt.FollowUpProbes) == 0 {
		t.Errorf("Expected non-empty follow up probes")
	}
}

func TestStoryBank_FormatMarkdown(t *testing.T) {
	entry := &StoryEntry{
		ID:          "story-md-01",
		WorkspaceID: "ws-alpha",
		TenantID:    "ws-alpha",
		Title:       "Markdown Formatting Test",
		Category:    StoryCategoryBreakthroughWin,
		Provenance: StoryProvenance{
			SourceType:      StoryProvenanceManualEntry,
			AuthorName:      "Author Test",
			GroundedFactIDs: []string{"fact-01", "fact-02"},
		},
		Narrative: StoryNarrative{
			HookSummary:       "Crisp hook.",
			ContextBackground: "Rich context.",
			ChallengeConflict: "Great obstacle.",
			ActionTaken:       "Clever architecture.",
			QuantifiedOutcome: "Massive speedup.",
			LessonLearned:     "Humility in design.",
		},
		Tags:           []string{"arch", "speed"},
		ApprovalStatus: StoryStatusApproved,
	}

	md := FormatStoryMarkdown(entry)
	if !strings.Contains(md, "# Markdown Formatting Test") {
		t.Errorf("Markdown missing title")
	}
	if !strings.Contains(md, "💡 Takeaway") {
		t.Errorf("Markdown missing takeaway section")
	}
	if !strings.Contains(md, "Crisp hook.") {
		t.Errorf("Markdown missing hook")
	}
	if !strings.Contains(md, "fact-01, fact-02") {
		t.Errorf("Markdown missing grounded facts")
	}
	if !strings.Contains(md, "#arch #speed") {
		t.Errorf("Markdown missing tags")
	}
}
