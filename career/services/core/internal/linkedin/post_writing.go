package linkedin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

var (
	ErrPostDraftNotFound         = errors.New("posts: post draft not found")
	ErrUnsupportedAutoPublish    = errors.New("posts: autonomous unreviewed post publishing permanently barred under platform integrity policy (AT-010, REQ-015, LI-11)")
	ErrEmptyTopic                = errors.New("posts: topic cannot be empty")
	ErrEmptyVerifiedFacts        = errors.New("posts: at least one verified career fact required for fact-grounded drafting (AT-003)")
	ErrInvalidPostApprovalToken  = errors.New("posts: invalid or tampered post approval token (FND-010, AT-007)")
	ErrPostDraftTextEmpty        = errors.New("posts: post draft text cannot be empty")
	ErrPostExceedsCharacterLimit = errors.New("posts: post draft exceeds maximum LinkedIn character limit of 3000 characters")
)

type PostAngle string

const (
	AngleContrarianInsight      PostAngle = "contrarian_insight"
	AngleLessonLearnedBreakdown PostAngle = "lesson_learned_breakdown"
	AngleTechnicalDeepDive      PostAngle = "technical_deep_dive"
	AngleMilestoneCelebration   PostAngle = "milestone_celebration"
	AngleActionableGuide        PostAngle = "actionable_guide"
)

type HookType string

const (
	HookQuestion   HookType = "question"
	HookContrarian HookType = "contrarian"
	HookDataScale  HookType = "data_scale"
	HookStory      HookType = "story"
	HookPunchy     HookType = "punchy"
)

type PostDraftStatus string

const (
	PostDraftPendingApproval   PostDraftStatus = "draft"
	PostDraftApproved          PostDraftStatus = "approved"
	PostDraftCopiedToClipboard PostDraftStatus = "copied_to_clipboard"
	PostDraftRejected          PostDraftStatus = "rejected"
)

type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityWarning  AuditSeverity = "warning"
	AuditSeverityBlocking AuditSeverity = "blocking"
)

type HookVariant struct {
	HookType  HookType `json:"hook_type"`
	HookText  string   `json:"hook_text"`
	Rationale string   `json:"rationale"`
}

type AuditIssue struct {
	Code             string        `json:"code"`
	Message          string        `json:"message"`
	Severity         AuditSeverity `json:"severity"`
	OffendingSnippet string        `json:"offending_snippet,omitempty"`
	SuggestedFix     string        `json:"suggested_fix,omitempty"`
}

type EditorialAuditReport struct {
	Passed                 bool         `json:"passed"`
	ReadabilityScore       int          `json:"readability_score"`
	CharacterCount         int          `json:"character_count"`
	LineBreakDensity       float64      `json:"line_break_density"`
	ParagraphCount         int          `json:"paragraph_count"`
	EstimatedReadTimeSec   int          `json:"estimated_read_time_sec"`
	BuzzwordsCount         int          `json:"buzzwords_count"`
	ViralityClaimsCount    int          `json:"virality_claims_count"`
	AiBypassClaimsCount    int          `json:"ai_bypass_claims_count"`
	UnverifiedMetricsCount int          `json:"unverified_metrics_count"`
	Issues                 []AuditIssue `json:"issues"`
	AuditedAt              time.Time    `json:"audited_at"`
	Disclaimer             string       `json:"disclaimer"`
}

type PostDraft struct {
	DraftID           string                `json:"draft_id"`
	WorkspaceID       string                `json:"workspace_id"`
	TenantID          string                `json:"tenant_id"`
	Topic             string                `json:"topic"`
	Angle             PostAngle             `json:"angle"`
	SelectedHookType  HookType              `json:"selected_hook_type"`
	SelectedHookText  string                `json:"selected_hook_text"`
	BodyText          string                `json:"body_text"`
	CallToAction      string                `json:"call_to_action"`
	FullPostText      string                `json:"full_post_text"`
	VerifiedFactsUsed []string              `json:"verified_facts_used"`
	AvailableHooks    []HookVariant         `json:"available_hooks"`
	LatestAudit       *EditorialAuditReport `json:"latest_audit,omitempty"`
	Status            PostDraftStatus       `json:"status"`
	ApprovalToken     string                `json:"approval_token,omitempty"`
	ApprovedBy        string                `json:"approved_by,omitempty"`
	ApprovedAt        *time.Time            `json:"approved_at,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

type PostDraftFilter struct {
	WorkspaceID string          `json:"workspace_id"`
	TenantID    string          `json:"tenant_id"`
	Angle       PostAngle       `json:"angle,omitempty"`
	Status      PostDraftStatus `json:"status,omitempty"`
	Query       string          `json:"query,omitempty"`
}

type PostWritingRequest struct {
	WorkspaceID       string    `json:"workspace_id"`
	TenantID          string    `json:"tenant_id"`
	Topic             string    `json:"topic"`
	Angle             PostAngle `json:"angle"`
	VerifiedFacts     []string  `json:"verified_facts"`
	TargetAudience    string    `json:"target_audience,omitempty"`
	PreferredHookType HookType  `json:"preferred_hook_type,omitempty"`
}

var forbiddenBuzzwords = []string{
	"10x engineer",
	"10x rockstar",
	"rockstar",
	"ninja",
	"synergy",
	"paradigm shift",
	"game-changer",
	"disrupt the industry",
	"bleeding-edge",
	"guru",
}

var forbiddenViralityPhrases = []string{
	"guaranteed 100k impressions",
	"guaranteed impressions",
	"viral secret formula",
	"viral reach guaranteed",
	"guaranteed to go viral",
	"hack the linkedin algorithm",
}

var forbiddenAiBypassPhrases = []string{
	"bypass ai detectors",
	"ai detector bypass",
	"100% bypass ai",
	"undetectable ai",
	"fool ai detectors",
}

// GenerateHookVariants creates a palette of 5 distinct hooks grounded in facts (SRC-L2).
func GenerateHookVariants(topic string, angle PostAngle, facts []string) []HookVariant {
	factSummary := ""
	if len(facts) > 0 {
		factSummary = facts[0]
	}

	variants := []HookVariant{
		{
			HookType:  HookContrarian,
			HookText:  fmt.Sprintf("Most teams think %s is about following trends, but our experience showed the exact opposite.", strings.ToLower(topic)),
			Rationale: "Directly confronts conventional assumptions with real engineering experience.",
		},
		{
			HookType:  HookQuestion,
			HookText:  fmt.Sprintf("When was the last time your approach to %s actually delivered the velocity your team needed?", strings.ToLower(topic)),
			Rationale: "Invites technical reflection and problem awareness.",
		},
		{
			HookType:  HookStory,
			HookText:  fmt.Sprintf("When we tackled %s, things didn't go according to the textbook.", strings.ToLower(topic)),
			Rationale: "Narrative opener highlighting genuine project learning.",
		},
		{
			HookType:  HookPunchy,
			HookText:  fmt.Sprintf("Clear architecture and verified telemetry beat hype in %s every single time.", strings.ToLower(topic)),
			Rationale: "High-signal, assertive engineering takeaway.",
		},
	}

	// If there's a quantified metric or fact, synthesize a data_scale hook
	if factSummary != "" {
		dataHook := HookVariant{
			HookType:  HookDataScale,
			HookText:  fmt.Sprintf("Here is what happened when we focused on real metrics: %s.", factSummary),
			Rationale: "Anchored directly to verified candidate metrics without speculation (AT-003).",
		}
		variants = append([]HookVariant{dataHook}, variants...)
	} else {
		dataHook := HookVariant{
			HookType:  HookDataScale,
			HookText:  fmt.Sprintf("Quantified results in %s require disciplined measurement, not guesses.", strings.ToLower(topic)),
			Rationale: "Anchored in empirical measurement principles.",
		}
		variants = append(variants, dataHook)
	}

	return variants
}

// AssembleFullPostText joins hook, body, and CTA cleanly with appropriate line breaks.
func AssembleFullPostText(hook, body, cta string) string {
	parts := []string{}
	if strings.TrimSpace(hook) != "" {
		parts = append(parts, strings.TrimSpace(hook))
	}
	if strings.TrimSpace(body) != "" {
		parts = append(parts, strings.TrimSpace(body))
	}
	if strings.TrimSpace(cta) != "" {
		parts = append(parts, strings.TrimSpace(cta))
	}
	return strings.Join(parts, "\n\n")
}

var postDraftCounter int64

// GeneratePostDraft constructs a fully grounded post draft with hooks and initial audit report.
func GeneratePostDraft(req PostWritingRequest) (*PostDraft, error) {
	if strings.TrimSpace(req.Topic) == "" {
		return nil, ErrEmptyTopic
	}
	if len(req.VerifiedFacts) == 0 {
		return nil, ErrEmptyVerifiedFacts
	}
	if req.Angle == "" {
		req.Angle = AngleTechnicalDeepDive
	}

	hooks := GenerateHookVariants(req.Topic, req.Angle, req.VerifiedFacts)

	selectedHook := hooks[0]
	if req.PreferredHookType != "" {
		for _, h := range hooks {
			if h.HookType == req.PreferredHookType {
				selectedHook = h
				break
			}
		}
	}

	// Generate structured body text according to chosen angle
	bodyParts := []string{}
	switch req.Angle {
	case AngleContrarianInsight:
		bodyParts = append(bodyParts, "In software engineering, prevailing consensus is often optimized for yesterday's constraints.")
		bodyParts = append(bodyParts, "Here is what we observed from firsthand implementation:")
		for _, f := range req.VerifiedFacts {
			bodyParts = append(bodyParts, fmt.Sprintf("• %s", f))
		}
		bodyParts = append(bodyParts, "The takeaway: before introducing additional abstractions, verify that your telemetry and operational complexity justify the cost.")

	case AngleLessonLearnedBreakdown:
		bodyParts = append(bodyParts, "Every major system evolution teaches lessons you cannot find in documentation.")
		bodyParts = append(bodyParts, "Key milestones grounded in our architecture:")
		for _, f := range req.VerifiedFacts {
			bodyParts = append(bodyParts, fmt.Sprintf("• %s", f))
		}
		bodyParts = append(bodyParts, "Engineering maturity is recognizing that simplicity is a hard-won outcome, not a default starting point.")

	case AngleTechnicalDeepDive:
		bodyParts = append(bodyParts, fmt.Sprintf("A technical breakdown of our approach to %s:", req.Topic))
		for _, f := range req.VerifiedFacts {
			bodyParts = append(bodyParts, fmt.Sprintf("• %s", f))
		}
		bodyParts = append(bodyParts, "By grounding our design decisions in benchmarked telemetry rather than dogma, we achieved consistent reliability across workloads.")

	case AngleMilestoneCelebration:
		bodyParts = append(bodyParts, "Proud to share an important milestone reached by the team:")
		for _, f := range req.VerifiedFacts {
			bodyParts = append(bodyParts, fmt.Sprintf("• %s", f))
		}
		bodyParts = append(bodyParts, "Grateful to everyone who collaborated, reviewed code, and contributed to making this happen.")

	case AngleActionableGuide:
		bodyParts = append(bodyParts, fmt.Sprintf("Three pragmatic rules we apply when engineering %s:", req.Topic))
		for i, f := range req.VerifiedFacts {
			bodyParts = append(bodyParts, fmt.Sprintf("%d. %s", i+1, f))
		}
		bodyParts = append(bodyParts, "Focusing on measurable trade-offs always beats premature optimization.")
	}

	bodyText := strings.Join(bodyParts, "\n\n")
	ctaText := "What patterns has your team found most resilient in production? Would welcome hearing your perspectives below."
	fullText := AssembleFullPostText(selectedHook.HookText, bodyText, ctaText)

	now := time.Now().UTC()
	seq := atomic.AddInt64(&postDraftCounter, 1)
	draft := &PostDraft{
		DraftID:           fmt.Sprintf("post-draft-%d-%d", now.UnixNano(), seq),
		WorkspaceID:       req.WorkspaceID,
		TenantID:          req.TenantID,
		Topic:             req.Topic,
		Angle:             req.Angle,
		SelectedHookType:  selectedHook.HookType,
		SelectedHookText:  selectedHook.HookText,
		BodyText:          bodyText,
		CallToAction:      ctaText,
		FullPostText:      fullText,
		VerifiedFactsUsed: req.VerifiedFacts,
		AvailableHooks:    hooks,
		Status:            PostDraftPendingApproval,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	audit := AuditPostDraft(*draft, req.VerifiedFacts)
	draft.LatestAudit = &audit

	return draft, nil
}

var metricPattern = regexp.MustCompile(`(?i)\b\d+([.,]\d+)?(%|k|m|x|rps|qps|ms|s|gb|tb|pb|usd|\$)\b`)

// AuditPostDraft inspects a draft for buzzword slop, false guarantees, formatting, and unverified figures (LI-11, FND-015, AT-003, AT-019).
func AuditPostDraft(draft PostDraft, verifiedFacts []string) EditorialAuditReport {
	now := time.Now().UTC()
	text := draft.FullPostText
	lowerText := strings.ToLower(text)
	charCount := len([]rune(text))

	report := EditorialAuditReport{
		Passed:               true,
		CharacterCount:       charCount,
		AuditedAt:            now,
		Disclaimer:           "Organic LinkedIn reach is non-deterministic and governed by reader engagement and network graph dynamics. No algorithmic virality or AI detection guarantees are provided (LI-11, AT-019).",
		EstimatedReadTimeSec: (charCount / 20) + 5, // ~250 words/min
	}

	// 1. Calculate line breaks and paragraphs
	paragraphs := strings.Split(text, "\n\n")
	cleanParagraphs := []string{}
	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			cleanParagraphs = append(cleanParagraphs, strings.TrimSpace(p))
		}
	}
	report.ParagraphCount = len(cleanParagraphs)
	totalLines := len(strings.Split(text, "\n"))
	if totalLines > 0 {
		report.LineBreakDensity = float64(totalLines) / float64(charCount+1)
	}

	// Readability baseline scoring (0-100)
	score := 85
	if charCount < 150 {
		score -= 20
		report.Issues = append(report.Issues, AuditIssue{
			Code:         "short_post",
			Message:      "Draft is relatively short; consider expanding context with concrete technical details.",
			Severity:     AuditSeverityInfo,
			SuggestedFix: "Add 1-2 sentences clarifying the production problem or architecture rationale.",
		})
	} else if charCount > 2500 {
		score -= 15
		report.Issues = append(report.Issues, AuditIssue{
			Code:         "long_post",
			Message:      "Post is approaching the 3,000 character limit and may cause reader drop-off.",
			Severity:     AuditSeverityWarning,
			SuggestedFix: "Condense secondary points to keep reader attention focused on the core engineering insight.",
		})
	}

	if charCount > 3000 {
		report.Passed = false
		report.Issues = append(report.Issues, AuditIssue{
			Code:             "char_limit_exceeded",
			Message:          "Post exceeds the hard LinkedIn 3,000 character boundary.",
			Severity:         AuditSeverityBlocking,
			OffendingSnippet: fmt.Sprintf("Character count: %d", charCount),
			SuggestedFix:     "Reduce text to under 3,000 characters.",
		})
	}

	// 2. Forbidden buzzwords audit
	for _, bw := range forbiddenBuzzwords {
		if strings.Contains(lowerText, bw) {
			report.BuzzwordsCount++
			report.Passed = false
			score -= 25
			report.Issues = append(report.Issues, AuditIssue{
				Code:             "hollow_buzzwords",
				Message:          fmt.Sprintf("Contains generic buzzword '%s' which degrades professional credibility.", bw),
				Severity:         AuditSeverityBlocking,
				OffendingSnippet: bw,
				SuggestedFix:     "Replace buzzword with substantive technical terms describing actual work done.",
			})
		}
	}

	// 3. Forbidden virality guarantee audit (LI-11, AT-019)
	for _, phrase := range forbiddenViralityPhrases {
		if strings.Contains(lowerText, phrase) {
			report.ViralityClaimsCount++
			report.Passed = false
			score -= 30
			report.Issues = append(report.Issues, AuditIssue{
				Code:             "forbidden_virality_claim",
				Message:          fmt.Sprintf("Contains forbidden virality/reach claim '%s'. Platform policy forbids fake reach guarantees.", phrase),
				Severity:         AuditSeverityBlocking,
				OffendingSnippet: phrase,
				SuggestedFix:     "Remove claims of guaranteed reach or algorithm manipulation.",
			})
		}
	}

	// 4. Forbidden AI detector bypass audit (LI-11, AT-019)
	for _, phrase := range forbiddenAiBypassPhrases {
		if strings.Contains(lowerText, phrase) {
			report.AiBypassClaimsCount++
			report.Passed = false
			score -= 30
			report.Issues = append(report.Issues, AuditIssue{
				Code:             "forbidden_ai_bypass_claim",
				Message:          fmt.Sprintf("Contains forbidden claim '%s'. AI detector evasion promises are prohibited.", phrase),
				Severity:         AuditSeverityBlocking,
				OffendingSnippet: phrase,
				SuggestedFix:     "Remove AI evasion guarantees and maintain authentic factual voice.",
			})
		}
	}

	// 5. Unverified metrics audit (AT-003)
	factsCombined := strings.ToLower(strings.Join(verifiedFacts, " "))
	matches := metricPattern.FindAllString(text, -1)
	for _, m := range matches {
		cleanMetric := strings.ToLower(strings.TrimSpace(m))
		// Ignore common time/cardinal units like 1st, 2nd, 3rd, 1, 2, 3 in numbered lists
		if cleanMetric == "1" || cleanMetric == "2" || cleanMetric == "3" || cleanMetric == "4" || cleanMetric == "5" {
			continue
		}
		if !strings.Contains(factsCombined, cleanMetric) {
			report.UnverifiedMetricsCount++
			report.Passed = false
			score -= 20
			report.Issues = append(report.Issues, AuditIssue{
				Code:             "unverified_metric",
				Message:          fmt.Sprintf("Metric '%s' was not found in verified career facts (AT-003 zero-hallucination invariant).", m),
				Severity:         AuditSeverityBlocking,
				OffendingSnippet: m,
				SuggestedFix:     "Verify that this quantitative assertion exists in your master career facts or remove it.",
			})
		}
	}

	if score < 0 {
		score = 0
	}
	report.ReadabilityScore = score

	return report
}

// ValidatePostText verifies text length and presence of characters.
func ValidatePostText(text string) error {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return ErrPostDraftTextEmpty
	}
	if len([]rune(clean)) > 3000 {
		return ErrPostExceedsCharacterLimit
	}
	return nil
}

// ComputePostApprovalToken generates an HMAC-SHA256 signature binding the draft content and metadata (FND-010, AT-007).
func ComputePostApprovalToken(draft PostDraft, secret string) string {
	if secret == "" {
		secret = "linkedin-post-approval-secret-v1"
	}
	payload := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		draft.DraftID,
		draft.WorkspaceID,
		draft.TenantID,
		draft.Angle,
		draft.SelectedHookText,
		draft.FullPostText,
	)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyPostApprovalToken validates that the token matches the draft's canonical content (FND-010, AT-007).
func VerifyPostApprovalToken(draft PostDraft, token, secret string) bool {
	if token == "" {
		return false
	}
	if secret == "" {
		secret = "linkedin-post-approval-secret-v1"
	}
	expected := ComputePostApprovalToken(draft, secret)
	return hmac.Equal([]byte(expected), []byte(token))
}

// RejectUnsupportedAutoPublish enforces fail-closed policy against autonomous unreviewed publishing (AT-010, REQ-015, LI-11).
func RejectUnsupportedAutoPublish() error {
	return ErrUnsupportedAutoPublish
}
