package linkedin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Errors for Content Planning and Repurposing (IMP-LI-14, AT-018, AT-010, AT-007).
var (
	ErrContentPlanNotFound         = errors.New("content planning: content plan item not found")
	ErrRepurposedDraftNotFound     = errors.New("content planning: repurposed draft not found")
	ErrDirectAutopublishDisallowed = errors.New("content planning: direct automated background publishing is disabled by governance (AT-010). Please use 1-click clipboard export or LinkedIn Compose deep-link")
	ErrDailyScheduleBudgetExceeded = errors.New("content planning: daily schedule slot budget exceeded for target date")
	ErrInvalidTimezone             = errors.New("content planning: invalid or unrecognized timezone")
	ErrInvalidPlanData             = errors.New("content planning: invalid content plan or draft data")
	ErrSlotCollision               = errors.New("content planning: requested schedule slot collides with an existing scheduled post")
)

// ArtifactType represents the source material archetype for repurposing.
type ArtifactType string

const (
	ArtifactTypeTechnicalBlog      ArtifactType = "technical_blog"
	ArtifactTypeGithubRelease      ArtifactType = "github_release"
	ArtifactTypeArchitectureRFC    ArtifactType = "architecture_rfc"
	ArtifactTypeIncidentPostmortem ArtifactType = "incident_postmortem"
	ArtifactTypeBenchmarkReport    ArtifactType = "benchmark_report"
)

// RepurposedFormat represents target content formats for LinkedIn.
type RepurposedFormat string

const (
	FormatSingleThought       RepurposedFormat = "single_thought"
	FormatCarouselOutline     RepurposedFormat = "carousel_outline"
	FormatActionableChecklist RepurposedFormat = "actionable_checklist"
	FormatContrarianBreakdown RepurposedFormat = "contrarian_breakdown"
	FormatInterviewQASpotlight RepurposedFormat = "interview_qa_spotlight"
)

// ContentPlanStatus represents the lifecycle state of a plan or draft.
type ContentPlanStatus string

const (
	PlanStatusDraft     ContentPlanStatus = "draft"
	PlanStatusReviewed  ContentPlanStatus = "reviewed"
	PlanStatusApproved  ContentPlanStatus = "approved"
	PlanStatusScheduled ContentPlanStatus = "scheduled"
	PlanStatusArchived  ContentPlanStatus = "archived"
)

// SourceArtifact captures input material (blog, release notes, RFC, postmortem) to be repurposed.
type SourceArtifact struct {
	ID           string       `json:"id"`
	WorkspaceID  string       `json:"workspace_id"`
	TenantID     string       `json:"tenant_id"`
	Title        string       `json:"title"`
	ArtifactType ArtifactType `json:"artifact_type"`
	RawText      string       `json:"raw_text"`
	Author       string       `json:"author"`
	Tags         []string     `json:"tags"`
	CreatedAt    time.Time    `json:"created_at"`
}

// RepurposedDraft represents a transformed content variant derived from a source artifact.
type RepurposedDraft struct {
	ID               string            `json:"id"`
	WorkspaceID      string            `json:"workspace_id"`
	TenantID         string            `json:"tenant_id"`
	SourceArtifactID string            `json:"source_artifact_id"`
	SourceTitle      string            `json:"source_title"`
	Format           RepurposedFormat  `json:"format"`
	Title            string            `json:"title"`
	ContentBody      string            `json:"content_body"`
	Hook             string            `json:"hook"`
	Tags             []string          `json:"tags"`
	CharacterCount   int               `json:"character_count"`
	Status           ContentPlanStatus `json:"status"`
	ApprovalToken    string            `json:"approval_token,omitempty"`
	LastApprovedAt   *time.Time        `json:"last_approved_at,omitempty"`
	ScheduledSlotUTC *time.Time        `json:"scheduled_slot_utc,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// ContentAssetMetadata tracks rich metadata for social artifacts (slides, read times).
type ContentAssetMetadata struct {
	SlideCount           int      `json:"slide_count,omitempty"`
	EstimatedReadTimeSec int      `json:"estimated_read_time_sec"`
	CanonicalURL         string   `json:"canonical_url,omitempty"`
	Hashtags             []string `json:"hashtags,omitempty"`
}

// ContentPlanItem represents a calendar-slotted content item with approval gating.
type ContentPlanItem struct {
	ID                       string               `json:"id"`
	WorkspaceID              string               `json:"workspace_id"`
	TenantID                 string               `json:"tenant_id"`
	DraftID                  string               `json:"draft_id"`
	Title                    string               `json:"title"`
	Format                   RepurposedFormat     `json:"format"`
	ScheduledSlotUTC         time.Time            `json:"scheduled_slot_utc"`
	UserTimezone             string               `json:"user_timezone"`
	LocalSlotFormatted       string               `json:"local_slot_formatted"`
	Status                   ContentPlanStatus    `json:"status"`
	ApprovalToken            string               `json:"approval_token,omitempty"`
	LastApprovedAt           *time.Time           `json:"last_approved_at,omitempty"`
	AssetMetadata            ContentAssetMetadata `json:"asset_metadata"`
	DirectPublishBlocked     bool                 `json:"direct_publish_blocked"`
	ClipboardExportAvailable bool                 `json:"clipboard_export_available"`
	ComposeURL               string               `json:"compose_url"`
	CreatedAt                time.Time            `json:"created_at"`
	UpdatedAt                time.Time            `json:"updated_at"`
}

// ContentCalendarFilter specifies query parameters for scheduled content items.
type ContentCalendarFilter struct {
	WorkspaceID        string
	RequestingTenantID string
	Status             ContentPlanStatus
	Format             RepurposedFormat
	StartDateUTC       *time.Time
	EndDateUTC         *time.Time
}

// AllTargetFormats returns the five canonical repurposed formats.
func AllTargetFormats() []RepurposedFormat {
	return []RepurposedFormat{
		FormatSingleThought,
		FormatCarouselOutline,
		FormatActionableChecklist,
		FormatContrarianBreakdown,
		FormatInterviewQASpotlight,
	}
}

// RepurposeContentArtifact transforms a source artifact into a designated LinkedIn format.
func RepurposeContentArtifact(source SourceArtifact, format RepurposedFormat) (*RepurposedDraft, error) {
	if strings.TrimSpace(source.Title) == "" {
		return nil, fmt.Errorf("%w: source title is required", ErrInvalidPlanData)
	}
	if strings.TrimSpace(source.RawText) == "" {
		return nil, fmt.Errorf("%w: source raw_text is required", ErrInvalidPlanData)
	}

	lines := strings.Split(source.RawText, "\n")
	var keyTakeaways []string
	var hookCandidate string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if hookCandidate == "" && len(trimmed) > 20 {
			hookCandidate = trimmed
		}
		if strings.HasPrefix(trimmed, "1.") || strings.HasPrefix(trimmed, "2.") || strings.HasPrefix(trimmed, "3.") ||
			strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			keyTakeaways = append(keyTakeaways, trimmed)
		} else if strings.Contains(strings.ToLower(trimmed), "key lesson") || strings.Contains(strings.ToLower(trimmed), "takeaway") {
			keyTakeaways = append(keyTakeaways, trimmed)
		}
	}

	if hookCandidate == "" {
		hookCandidate = fmt.Sprintf("Deep dive into %s: Lessons learned from production architecture.", source.Title)
	}

	var body strings.Builder
	var draftTitle string
	var hook string

	switch format {
	case FormatSingleThought:
		draftTitle = fmt.Sprintf("Thought: %s", source.Title)
		hook = fmt.Sprintf("Most engineering teams overlook this when scaling: %s", hookCandidate)
		body.WriteString(hook + "\n\n")

		if len(keyTakeaways) > 0 {
			body.WriteString(keyTakeaways[0] + "\n\n")
		} else {
			body.WriteString("Architecture is defined by the bottlenecks you choose to eliminate early.\n\n")
		}

		body.WriteString("What's your primary guardrail when load spikes 10x? Let's discuss below.\n\n")

	case FormatCarouselOutline:
		draftTitle = fmt.Sprintf("Carousel: %s (5 Slides)", source.Title)
		hook = fmt.Sprintf("Swipe through: 5 architectural insights from %s", source.Title)
		body.WriteString(fmt.Sprintf("📊 SLIDE 1: Title\n%s\nBy %s\n\n", source.Title, source.Author))
		body.WriteString("📊 SLIDE 2: The Context & Bottleneck\n")
		body.WriteString(fmt.Sprintf("• Initial challenge: %s\n• Why conventional patterns broke under production traffic\n\n", hookCandidate))
		body.WriteString("📊 SLIDE 3: Architectural Pivot\n")
		if len(keyTakeaways) > 0 {
			for i, t := range keyTakeaways {
				if i < 3 {
					body.WriteString(fmt.Sprintf("• %s\n", t))
				}
			}
		} else {
			body.WriteString("• Pivot 1: Read-through cache coalescing\n• Pivot 2: Jittered TTL distribution\n")
		}
		body.WriteString("\n📊 SLIDE 4: Production Metric Impact\n")
		body.WriteString("• Verified p99 latency reduction\n• Eliminated lock-step thundering herd evictions\n• Scaled safely to 1M+ QPS\n\n")
		body.WriteString("📊 SLIDE 5: Takeaway & Discussion\n")
		body.WriteString("Save this carousel for your next distributed systems review.\nWhich caching strategy has saved your cluster in production?\n\n")

	case FormatActionableChecklist:
		draftTitle = fmt.Sprintf("Checklist: %s", source.Title)
		hook = fmt.Sprintf("Here is the practical checklist we followed during %s:", source.Title)
		body.WriteString(hook + "\n\n")
		body.WriteString("Save this checklist before your next production rollout:\n\n")

		if len(keyTakeaways) > 0 {
			for _, t := range keyTakeaways {
				clean := strings.TrimLeft(t, "1234567890.-* ")
				body.WriteString(fmt.Sprintf("[ ] %s\n", clean))
			}
		} else {
			body.WriteString("[ ] Deterministic read-through with single-flight mutex\n")
			body.WriteString("[ ] Consistent hashing rings with virtual nodes\n")
			body.WriteString("[ ] Jittered TTL expiration on hot partitions\n")
		}
		body.WriteString("\n💡 Pro Tip: Never let cached keys expire in lock-step batches.\n\n")

	case FormatContrarianBreakdown:
		draftTitle = fmt.Sprintf("Contrarian Breakdown: %s", source.Title)
		hook = fmt.Sprintf("The Common Myth: Adding more cache capacity fixes high latency.\nThe Reality: %s", hookCandidate)
		body.WriteString(hook + "\n\n")
		body.WriteString("### The Conventional Wisdom\n")
		body.WriteString("Most teams throw more RAM and cluster replicas at database latency.\n\n")
		body.WriteString("### Architectural Pivot & Truth\n")
		if len(keyTakeaways) > 0 {
			body.WriteString(fmt.Sprintf("What actually worked: %s\n\n", keyTakeaways[0]))
		} else {
			body.WriteString("What actually worked: Eliminating redundant database cache misses via single-flight mutex coalescing.\n\n")
		}
		body.WriteString("Stop treating caching as an afterthought. It's a first-class architectural boundary.\n\n")

	case FormatInterviewQASpotlight:
		draftTitle = fmt.Sprintf("Q&A Spotlight: %s", source.Title)
		hook = fmt.Sprintf("Question: How do you scale system throughput 100x without ballooning infrastructure costs?")
		body.WriteString(hook + "\n\n")
		body.WriteString(fmt.Sprintf("Principal Engineer Answer:\nBased on our work on %s:\n\n", source.Title))
		if len(keyTakeaways) > 0 {
			for _, t := range keyTakeaways {
				body.WriteString(fmt.Sprintf("• %s\n", t))
			}
		} else {
			body.WriteString("• Focus on eliminating thundering herds\n• Distribute hot keys across consistent hashing rings\n• Maintain strict p99 observability\n")
		}
		body.WriteString("\nKey principle: Predictability trumps peak burst capacity every time.\n\n")

	default:
		return nil, fmt.Errorf("%w: unsupported repurposed format %s", ErrInvalidPlanData, format)
	}

	// Append hashtags
	var hashStr strings.Builder
	if len(source.Tags) > 0 {
		for _, tag := range source.Tags {
			cleanTag := strings.ReplaceAll(tag, "-", "")
			hashStr.WriteString(fmt.Sprintf("#%s ", cleanTag))
		}
	} else {
		hashStr.WriteString("#softwareengineering #systemdesign #cloud #techlead")
	}
	body.WriteString(strings.TrimSpace(hashStr.String()) + "\n")

	content := body.String()
	now := time.Now().UTC()
	draftID := fmt.Sprintf("draft-%s-%d", strings.ToLower(string(format)), now.UnixNano())

	return &RepurposedDraft{
		ID:               draftID,
		WorkspaceID:      source.WorkspaceID,
		TenantID:         source.TenantID,
		SourceArtifactID: source.ID,
		SourceTitle:      source.Title,
		Format:           format,
		Title:            draftTitle,
		ContentBody:      content,
		Hook:             hook,
		Tags:             source.Tags,
		CharacterCount:   len([]rune(content)),
		Status:           PlanStatusDraft,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// RepurposeAllFormats transforms a source artifact into all 5 canonical LinkedIn formats.
func RepurposeAllFormats(source SourceArtifact) ([]*RepurposedDraft, error) {
	formats := AllTargetFormats()
	drafts := make([]*RepurposedDraft, 0, len(formats))

	for _, f := range formats {
		draft, err := RepurposeContentArtifact(source, f)
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, draft)
	}
	return drafts, nil
}

// CalculateNextSafeScheduleSlot safely validates timezone and DST boundaries (AT-018 & FND-008).
// It converts local requested times into canonical UTC, checks for daily budget limits,
// and detects collisions with existing scheduled slots.
func CalculateNextSafeScheduleSlot(
	requestedTimeLocal time.Time,
	userTimezone string,
	existingScheduledSlotsUTC []time.Time,
	maxDailyBudget int,
) (utcSlot time.Time, localFormatted string, err error) {
	if userTimezone == "" {
		userTimezone = "UTC"
	}

	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("%w: %v", ErrInvalidTimezone, err)
	}

	// Resolve the requested local time within user's timezone
	inLoc := requestedTimeLocal.In(loc)
	utcSlot = inLoc.UTC()

	// Compute day boundaries in local timezone to count daily budget correctly across DST shifts
	year, month, day := inLoc.Date()
	startOfDayLocal := time.Date(year, month, day, 0, 0, 0, 0, loc)
	endOfDayLocal := startOfDayLocal.Add(24 * time.Hour) // DST shift will be naturally respected by location

	startOfDayUTC := startOfDayLocal.UTC()
	endOfDayUTC := endOfDayLocal.UTC()

	// Count slots on the same local calendar day
	dayCount := 0
	for _, existingUTC := range existingScheduledSlotsUTC {
		// Collision detection: within 15 minutes of an existing slot
		diff := existingUTC.Sub(utcSlot)
		if diff < 0 {
			diff = -diff
		}
		if diff < 15*time.Minute {
			return time.Time{}, "", ErrSlotCollision
		}

		if (existingUTC.Equal(startOfDayUTC) || existingUTC.After(startOfDayUTC)) && existingUTC.Before(endOfDayUTC) {
			dayCount++
		}
	}

	if maxDailyBudget > 0 && dayCount >= maxDailyBudget {
		return time.Time{}, "", fmt.Errorf("%w: max %d posts allowed per day (current: %d)", ErrDailyScheduleBudgetExceeded, maxDailyBudget, dayCount)
	}

	// Format readable local slot with timezone abbreviation and offset
	zoneName, offsetSec := inLoc.Zone()
	offsetHours := offsetSec / 3600
	offsetMins := (offsetSec % 3600) / 60
	sign := "+"
	if offsetHours < 0 {
		sign = "-"
		offsetHours = -offsetHours
	}

	localFormatted = fmt.Sprintf("%s (%s, UTC%s%02d:%02d)",
		inLoc.Format("2006-01-02 15:04:05 MST"),
		zoneName,
		sign,
		offsetHours,
		offsetMins,
	)

	return utcSlot, localFormatted, nil
}

// GenerateLinkedInComposeURL constructs a deep link for 1-click publishing in LinkedIn web compose (AT-010).
func GenerateLinkedInComposeURL(text string) string {
	baseURL := "https://www.linkedin.com/feed/?shareActive=true&text="
	return baseURL + url.QueryEscape(text)
}

// ComputePlanApprovalToken computes a cryptographic HMAC-SHA256 signature for a ContentPlanItem (AT-007).
func ComputePlanApprovalToken(secretKey string, item *ContentPlanItem) string {
	if secretKey == "" {
		secretKey = "default_content_plan_secret_key"
	}
	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%d",
		item.ID,
		item.WorkspaceID,
		item.DraftID,
		item.Title,
		item.Format,
		item.ScheduledSlotUTC.Unix(),
	)

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payload))
	return "plan_hmac_" + hex.EncodeToString(h.Sum(nil))
}

// VerifyPlanApprovalToken verifies the HMAC signature of a ContentPlanItem.
func VerifyPlanApprovalToken(secretKey string, item *ContentPlanItem, token string) bool {
	if token == "" || item == nil {
		return false
	}
	expected := ComputePlanApprovalToken(secretKey, item)
	return hmac.Equal([]byte(expected), []byte(token))
}

// ComputeDraftApprovalToken computes HMAC-SHA256 signature for a RepurposedDraft (AT-007).
func ComputeDraftApprovalToken(secretKey string, draft *RepurposedDraft) string {
	if secretKey == "" {
		secretKey = "default_draft_secret_key"
	}
	payload := fmt.Sprintf("%s|%s|%s|%s|%s",
		draft.ID,
		draft.WorkspaceID,
		draft.Title,
		draft.ContentBody,
		draft.Format,
	)

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payload))
	return "draft_hmac_" + hex.EncodeToString(h.Sum(nil))
}

// VerifyDraftApprovalToken verifies HMAC signature of a RepurposedDraft.
func VerifyDraftApprovalToken(secretKey string, draft *RepurposedDraft, token string) bool {
	if token == "" || draft == nil {
		return false
	}
	expected := ComputeDraftApprovalToken(secretKey, draft)
	return hmac.Equal([]byte(expected), []byte(token))
}

// FormatRepurposedMarkdown converts a RepurposedDraft into clean, readable markdown ready for clipboard copy.
func FormatRepurposedMarkdown(draft *RepurposedDraft) string {
	if draft == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", draft.Title))
	sb.WriteString(fmt.Sprintf("**Format:** `%s` | **Source:** `%s` | **Characters:** `%d`\n\n",
		draft.Format, draft.SourceTitle, draft.CharacterCount))
	sb.WriteString("```linkedin-draft\n")
	sb.WriteString(draft.ContentBody)
	sb.WriteString("\n```\n")
	return sb.String()
}

// ValidateContentPlan performs boundary validation on a ContentPlanItem.
func ValidateContentPlan(item *ContentPlanItem) error {
	if item == nil {
		return fmt.Errorf("%w: content plan item cannot be nil", ErrInvalidPlanData)
	}
	if strings.TrimSpace(item.WorkspaceID) == "" {
		return fmt.Errorf("%w: workspace_id is required", ErrInvalidPlanData)
	}
	if strings.TrimSpace(item.DraftID) == "" {
		return fmt.Errorf("%w: draft_id is required", ErrInvalidPlanData)
	}
	if strings.TrimSpace(item.Title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidPlanData)
	}
	if item.ScheduledSlotUTC.IsZero() {
		return fmt.Errorf("%w: scheduled_slot_utc cannot be zero", ErrInvalidPlanData)
	}
	return nil
}
