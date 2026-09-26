package linkedin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Errors for Story Bank and Interviewer.
var (
	ErrStoryEntryNotFound       = errors.New("story bank: story entry not found")
	ErrInterviewSessionNotFound = errors.New("story bank: interview session not found")
	ErrInvalidStoryData         = errors.New("story bank: invalid story data")
)

// StoryCategory identifies the narrative archetype of a Story Bank entry.
type StoryCategory string

const (
	StoryCategoryTurningPoint     StoryCategory = "turning_point"
	StoryCategoryScarOrFailure    StoryCategory = "scar_or_failure"
	StoryCategoryBreakthroughWin  StoryCategory = "breakthrough_win"
	StoryCategoryContrarianBelief StoryCategory = "contrarian_belief"
	StoryCategoryMentorshipCulture StoryCategory = "mentorship_culture"
)

// StoryProvenanceType identifies the origin of a Story Bank entry.
type StoryProvenanceType string

const (
	StoryProvenanceInterviewSession  StoryProvenanceType = "interview_session"
	StoryProvenanceManualEntry       StoryProvenanceType = "manual_entry"
	StoryProvenanceImportedMilestone StoryProvenanceType = "imported_milestone"
)

// StoryStatus indicates the review and cryptographic approval state.
type StoryStatus string

const (
	StoryStatusDraft    StoryStatus = "draft"
	StoryStatusReviewed StoryStatus = "reviewed"
	StoryStatusApproved StoryStatus = "approved"
	StoryStatusArchived StoryStatus = "archived"
)

// StoryProvenance captures complete origin, author attribution, and fact grounding.
type StoryProvenance struct {
	SourceType         StoryProvenanceType `json:"source_type"`
	InterviewSessionID string              `json:"interview_session_id,omitempty"`
	AuthorName         string              `json:"author_name"`
	Timestamp          time.Time           `json:"timestamp"`
	GroundedFactIDs    []string            `json:"grounded_fact_ids,omitempty"`
}

// StoryNarrative contains the structured STAR / retrospective story sections.
type StoryNarrative struct {
	HookSummary       string `json:"hook_summary"`
	ContextBackground string `json:"context_background"`
	ChallengeConflict string `json:"challenge_conflict"`
	ActionTaken       string `json:"action_taken"`
	QuantifiedOutcome string `json:"quantified_outcome"`
	LessonLearned     string `json:"lesson_learned"`
}

// StoryAuditIssue reports an issue found during narrative provenance or metric audit.
type StoryAuditIssue struct {
	Field    string `json:"field"`
	Metric   string `json:"metric"`
	Severity string `json:"severity"` // "blocking" or "advisory"
	Message  string `json:"message"`
}

// StoryAuditReport summarizes metric grounding and readability.
type StoryAuditReport struct {
	HasBlockingIssues bool              `json:"has_blocking_issues"`
	Issues            []StoryAuditIssue `json:"issues"`
	ReadabilityScore  int               `json:"readability_score"`
	GroundedFactCount int               `json:"grounded_fact_count"`
}

// StoryEntry represents a durable, user-approved story with explicit provenance.
type StoryEntry struct {
	ID             string            `json:"id"`
	WorkspaceID    string            `json:"workspace_id"`
	TenantID       string            `json:"tenant_id"`
	Title          string            `json:"title"`
	Category       StoryCategory     `json:"category"`
	Provenance     StoryProvenance   `json:"provenance"`
	Narrative      StoryNarrative    `json:"narrative"`
	Tags           []string          `json:"tags"`
	ApprovalStatus StoryStatus       `json:"approval_status"`
	ApprovalToken  string            `json:"approval_token,omitempty"`
	LastApprovedAt *time.Time        `json:"last_approved_at,omitempty"`
	AuditReport    *StoryAuditReport `json:"audit_report,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// InterviewPrompt defines a curated prompt for the guided interviewer.
type InterviewPrompt struct {
	ID                 string        `json:"id"`
	Category           StoryCategory `json:"category"`
	Title              string        `json:"title"`
	Question           string        `json:"question"`
	ContextPlaceholder string        `json:"context_placeholder"`
	FollowUpProbes     []string      `json:"follow_up_probes"`
}

// InterviewQA represents a single question and answer exchange within a session.
type InterviewQA struct {
	PromptID                string `json:"prompt_id"`
	Question                string `json:"question"`
	CandidateAnswer         string `json:"candidate_answer"`
	FollowUpProbe           string `json:"follow_up_probe,omitempty"`
	CandidateFollowUpAnswer string `json:"candidate_follow_up_answer,omitempty"`
}

// InterviewSession tracks an active or completed guided interview.
type InterviewSession struct {
	ID                  string        `json:"id"`
	WorkspaceID         string        `json:"workspace_id"`
	TenantID            string        `json:"tenant_id"`
	Category            StoryCategory `json:"category"`
	Status              string        `json:"status"` // "in_progress", "completed", "abandoned"
	QuestionsAndAnswers []InterviewQA `json:"questions_and_answers"`
	SynthesizedStoryID  string        `json:"synthesized_story_id,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

// Curated interview prompts library.
var curatedPrompts = []InterviewPrompt{
	{
		ID:                 "prompt-tp-01",
		Category:           StoryCategoryTurningPoint,
		Title:              "Defining Career Fork in the Road",
		Question:           "What was the pivotal moment that convinced you to transition from deep individual coding to systems architecture and technical leadership?",
		ContextPlaceholder: "e.g., We were rewriting our core billing service for the third time in two years...",
		FollowUpProbes: []string{
			"What concrete boundary did you establish first, and what was the cross-team result?",
			"What trade-off did you have to accept by stepping back from daily pull requests?",
		},
	},
	{
		ID:                 "prompt-scar-01",
		Category:           StoryCategoryScarOrFailure,
		Title:              "Production Incident & Hard Lessons",
		Question:           "Tell us about a major production incident, outage, or technical failure, and the hard lessons learned.",
		ContextPlaceholder: "e.g., At 2:14 AM, our primary cache cluster disintegrated because...",
		FollowUpProbes: []string{
			"What was the exact root cause that standard integration tests failed to catch?",
			"What architectural circuit breaker or guard did you implement to ensure it never happens again?",
		},
	},
	{
		ID:                 "prompt-win-01",
		Category:           StoryCategoryBreakthroughWin,
		Title:              "Major Architectural Breakthrough",
		Question:           "Describe a significant architectural breakthrough or performance optimization and its real quantified impact.",
		ContextPlaceholder: "e.g., We migrated our hot-path data ingestion from batch polling to event streaming...",
		FollowUpProbes: []string{
			"What was the baseline metric before your intervention, and what was the verified outcome?",
			"How did you gain organizational buy-in from skeptical stakeholders?",
		},
	},
	{
		ID:                 "prompt-cb-01",
		Category:           StoryCategoryContrarianBelief,
		Title:              "Contrarian Technical Perspective",
		Question:           "What is a strongly held industry opinion or common practice that you completely disagree with?",
		ContextPlaceholder: "e.g., Microservices-by-default has ruined more early-stage startups than bad product market fit...",
		FollowUpProbes: []string{
			"What real production evidence or scars shaped this conviction?",
			"Under what specific edge case would you concede that the conventional approach is actually justified?",
		},
	},
	{
		ID:                 "prompt-mc-01",
		Category:           StoryCategoryMentorshipCulture,
		Title:              "Engineering Culture & Team Growth",
		Question:           "Describe a moment where you helped an engineer level up, or overhauled an engineering team process that was causing burnout.",
		ContextPlaceholder: "e.g., Our on-call rotation was generating 120 pages a week until we instituted a strict noisy-alert policy...",
		FollowUpProbes: []string{
			"How did you measure the cultural or operational improvement over the following quarter?",
			"What friction did you encounter when changing the team's working habits?",
		},
	},
}

// GetCuratedInterviewPrompts returns all available interview frameworks.
func GetCuratedInterviewPrompts() []InterviewPrompt {
	return curatedPrompts
}

// GetCuratedPromptByID finds a prompt by ID.
func GetCuratedPromptByID(id string) *InterviewPrompt {
	for _, p := range curatedPrompts {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// Regex to capture quantitative metrics in story text: percentages, currency, time, throughput, storage.
var storyMetricRegex = regexp.MustCompile(`(?i)(\b\d+([.,]\d+)?%|\b\d+([.,]\d+)?(k|m|x|rps|qps|ms|s|gb|tb|pb|usd)\b|\$\d+([.,]\d+)?(k|m)?)`)

// AuditStoryProvenanceAndMetrics audits quantitative metrics in story narrative against verified facts.
// Prevents fabricated numbers from entering approved story banks (AT-003).
func AuditStoryProvenanceAndMetrics(narrative StoryNarrative, verifiedFacts []string) StoryAuditReport {
	report := StoryAuditReport{
		Issues:            []StoryAuditIssue{},
		GroundedFactCount: len(verifiedFacts),
	}

	// Combine all narrative text for readability computation
	fullText := strings.Join([]string{
		narrative.HookSummary,
		narrative.ContextBackground,
		narrative.ChallengeConflict,
		narrative.ActionTaken,
		narrative.QuantifiedOutcome,
		narrative.LessonLearned,
	}, " ")

	report.ReadabilityScore = computeStoryReadability(fullText)

	// Check fields for unverified metrics
	fields := []struct {
		name string
		text string
	}{
		{"hook_summary", narrative.HookSummary},
		{"challenge_conflict", narrative.ChallengeConflict},
		{"action_taken", narrative.ActionTaken},
		{"quantified_outcome", narrative.QuantifiedOutcome},
	}

	for _, f := range fields {
		matches := storyMetricRegex.FindAllString(f.text, -1)
		for _, m := range matches {
			if isCommonNonMetricWord(m) {
				continue
			}
			foundInFacts := false
			for _, fact := range verifiedFacts {
				if strings.Contains(strings.ToLower(fact), strings.ToLower(m)) {
					foundInFacts = true
					break
				}
			}
			if !foundInFacts {
				report.Issues = append(report.Issues, StoryAuditIssue{
					Field:    f.name,
					Metric:   m,
					Severity: "blocking",
					Message:  fmt.Sprintf("Metric '%s' is not found in candidate verified career facts.", m),
				})
			}
		}
	}

	report.HasBlockingIssues = len(report.Issues) > 0
	return report
}

func isCommonNonMetricWord(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	switch lower {
	case "a", "is", "as", "us", "in":
		return true
	default:
		return false
	}
}

func computeStoryReadability(text string) int {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}
	sentences := strings.Split(text, ".")
	validSentences := 0
	for _, s := range sentences {
		if len(strings.TrimSpace(s)) > 0 {
			validSentences++
		}
	}
	if validSentences == 0 {
		validSentences = 1
	}

	asl := float64(len(words)) / float64(validSentences)
	score := 100.0 - (asl * 1.3)
	if score < 20 {
		score = 20
	}
	if score > 98 {
		score = 98
	}
	return int(score)
}

// SynthesizeStoryFromInterview converts completed Q&A into a structured StoryEntry draft.
func SynthesizeStoryFromInterview(session *InterviewSession, authorName string, factIDs []string) (*StoryEntry, error) {
	if session == nil {
		return nil, fmt.Errorf("interview session cannot be nil")
	}
	if len(session.QuestionsAndAnswers) == 0 {
		return nil, fmt.Errorf("cannot synthesize story from empty interview session")
	}

	firstQA := session.QuestionsAndAnswers[0]
	title := fmt.Sprintf("Retrospective: %s Insights", strings.ReplaceAll(string(session.Category), "_", " "))
	hook := firstQA.CandidateAnswer
	if len(hook) > 160 {
		hook = hook[:160] + "..."
	}

	entry := &StoryEntry{
		ID:          fmt.Sprintf("story-%d", time.Now().UnixNano()),
		WorkspaceID: session.WorkspaceID,
		TenantID:    session.TenantID,
		Title:       title,
		Category:    session.Category,
		Provenance: StoryProvenance{
			SourceType:         StoryProvenanceInterviewSession,
			InterviewSessionID: session.ID,
			AuthorName:         authorName,
			Timestamp:          time.Now().UTC(),
			GroundedFactIDs:    factIDs,
		},
		Narrative: StoryNarrative{
			HookSummary:       hook,
			ContextBackground: firstQA.CandidateAnswer,
			ChallengeConflict: "Identified organizational and architectural friction during execution.",
			ActionTaken:       firstQA.CandidateFollowUpAnswer,
			QuantifiedOutcome: "Achieved measurable team alignment and operational resilience.",
			LessonLearned:     "Systemic interface boundaries matter more than isolated repository code.",
		},
		Tags: []string{
			string(session.Category),
			"guided-interview",
		},
		ApprovalStatus: StoryStatusDraft,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	session.SynthesizedStoryID = entry.ID
	session.Status = "completed"
	session.UpdatedAt = time.Now().UTC()

	return entry, nil
}

// FormatStoryMarkdown formats a StoryEntry into structured readable markdown.
func FormatStoryMarkdown(entry *StoryEntry) string {
	if entry == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", entry.Title))
	sb.WriteString(fmt.Sprintf("**Category:** `%s` | **Provenance:** `%s` | **Status:** `%s`\n\n", entry.Category, entry.Provenance.SourceType, entry.ApprovalStatus))

	if len(entry.Provenance.GroundedFactIDs) > 0 {
		sb.WriteString(fmt.Sprintf("**Grounded Facts:** %s\n\n", strings.Join(entry.Provenance.GroundedFactIDs, ", ")))
	}

	sb.WriteString("### 💡 Takeaway\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", entry.Narrative.HookSummary))

	sb.WriteString("### 🔍 Context & Situation\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", entry.Narrative.ContextBackground))

	sb.WriteString("### ⚡ The Conflict / Challenge\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", entry.Narrative.ChallengeConflict))

	sb.WriteString("### 🛠 Engineering Actions Taken\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", entry.Narrative.ActionTaken))

	sb.WriteString("### 📈 Quantified Outcome\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", entry.Narrative.QuantifiedOutcome))

	sb.WriteString("### 🧠 Hard-Won Lesson\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", entry.Narrative.LessonLearned))

	if len(entry.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags:** #%s\n", strings.Join(entry.Tags, " #")))
	}

	return sb.String()
}

// ComputeStoryApprovalToken generates an HMAC-SHA256 signature for a StoryEntry.
func ComputeStoryApprovalToken(secretKey string, entry *StoryEntry) string {
	if secretKey == "" {
		secretKey = "default_story_secret_key"
	}
	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s",
		entry.ID,
		entry.WorkspaceID,
		entry.Title,
		entry.Narrative.HookSummary,
		entry.Narrative.ContextBackground,
		entry.Narrative.ChallengeConflict,
		entry.Narrative.ActionTaken,
		entry.Narrative.QuantifiedOutcome,
		entry.Narrative.LessonLearned,
	)

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payload))
	return "story_hmac_" + hex.EncodeToString(h.Sum(nil))
}

// VerifyStoryApprovalToken verifies the HMAC-SHA256 signature for a StoryEntry.
func VerifyStoryApprovalToken(secretKey string, entry *StoryEntry, token string) bool {
	if token == "" {
		return false
	}
	expected := ComputeStoryApprovalToken(secretKey, entry)
	return hmac.Equal([]byte(expected), []byte(token))
}

// ValidateStoryEntry validates boundaries and required narrative fields.
func ValidateStoryEntry(entry *StoryEntry) error {
	if entry == nil {
		return fmt.Errorf("story entry cannot be nil")
	}
	if strings.TrimSpace(entry.WorkspaceID) == "" {
		return fmt.Errorf("workspace_id is required")
	}
	if strings.TrimSpace(entry.Title) == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if strings.TrimSpace(entry.Narrative.HookSummary) == "" {
		return fmt.Errorf("hook_summary cannot be empty")
	}
	if strings.TrimSpace(entry.Narrative.ContextBackground) == "" {
		return fmt.Errorf("context_background cannot be empty")
	}
	if strings.TrimSpace(entry.Narrative.ActionTaken) == "" {
		return fmt.Errorf("action_taken cannot be empty")
	}
	if strings.TrimSpace(entry.Narrative.LessonLearned) == "" {
		return fmt.Errorf("lesson_learned cannot be empty")
	}
	return nil
}
