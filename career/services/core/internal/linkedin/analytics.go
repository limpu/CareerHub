package linkedin

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Errors for Engagement Monitoring and Analytics (IMP-LI-15, AT-028, AT-010).
var (
	ErrPostAnalyticsNotFound     = errors.New("analytics: post analytics snapshot not found")
	ErrInvalidAnalyticsData      = errors.New("analytics: invalid post analytics payload")
	ErrUnpermittedMetricFaking   = errors.New("analytics: cannot report unavailable metrics as confirmed zeros or fabricate simulated metrics (AT-028, AT-010)")
	ErrCrossTenantAnalyticsDeny = errors.New("analytics: cross-tenant access strictly denied (AT-012)")
)

// MetricAvailabilityStatus captures the transparency state of a metric (AT-028).
type MetricAvailabilityStatus string

const (
	MetricStatusAvailable       MetricAvailabilityStatus = "available"
	MetricStatusUnavailable     MetricAvailabilityStatus = "unavailable"
	MetricStatusModeledEstimate MetricAvailabilityStatus = "modeled_estimate"
)

// EngagerSegment classifies interacting professionals by seniority and role (SRC-L2).
type EngagerSegment string

const (
	SegmentDecisionMaker    EngagerSegment = "decision_maker"
	SegmentPeerPractitioner EngagerSegment = "peer_practitioner"
	SegmentTalentPartner    EngagerSegment = "talent_partner"
	SegmentOtherNetwork     EngagerSegment = "other_network"
)

// InteractionType represents the nature of engagement.
type InteractionType string

const (
	InteractionReaction InteractionType = "reaction"
	InteractionComment  InteractionType = "comment"
	InteractionRepost   InteractionType = "repost"
)

// MetricItem represents an individual performance metric with audit metadata (AT-028).
type MetricItem struct {
	Value             float64                  `json:"value"`
	Status            MetricAvailabilityStatus `json:"status"`
	UnavailableReason string                   `json:"unavailable_reason,omitempty"`
	IsModeled         bool                     `json:"is_modeled"`
	ModelBasis        string                   `json:"model_basis,omitempty"`
}

// EngagerProfile captures a verified professional who engaged with content (SRC-L2).
type EngagerProfile struct {
	AuthorURN       string          `json:"author_urn"`
	Name            string          `json:"name"`
	Headline        string          `json:"headline"`
	Company         string          `json:"company"`
	InteractionType InteractionType `json:"interaction_type"`
	CommentText     string          `json:"comment_text,omitempty"`
	Segment         EngagerSegment  `json:"segment"`
	ObservedAt      time.Time       `json:"observed_at"`
}

// ICPBreakdown summarizes high-value engager demographics.
type ICPBreakdown struct {
	DecisionMakers         int `json:"decision_makers"`
	PeerPractitioners      int `json:"peer_practitioners"`
	TalentPartners         int `json:"talent_partners"`
	OtherNetwork           int `json:"other_network"`
	HighValueEngagersCount int `json:"high_value_engagers_count"`
}

// PostAnalyticsSnapshot captures an immutable point-in-time performance audit for a post.
type PostAnalyticsSnapshot struct {
	SnapshotID        string                `json:"snapshot_id"`
	WorkspaceID       string                `json:"workspace_id"`
	TenantID          string                `json:"tenant_id"`
	PostURN           string                `json:"post_urn"`
	PostTitle         string                `json:"post_title"`
	PublishedAt       time.Time             `json:"published_at"`
	Metrics           map[string]MetricItem `json:"metrics"`
	Engagers          []EngagerProfile      `json:"engagers"`
	ICPBreakdown      ICPBreakdown          `json:"icp_breakdown"`
	DisclosureNotice  string                `json:"disclosure_notice,omitempty"`
	ObservedAt        time.Time             `json:"observed_at"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

// PerformanceBenchmark captures baseline performance across a category (AT-028).
type PerformanceBenchmark struct {
	Category           string  `json:"category"`
	SampleSize         int     `json:"sample_size"`
	AvgReactions       float64 `json:"avg_reactions"`
	AvgComments        float64 `json:"avg_comments"`
	AvgReposts         float64 `json:"avg_reposts"`
	AvgEngagementRate  float64 `json:"avg_engagement_rate"`
	BenchmarkSource    string  `json:"benchmark_source"`
}

// CategorizeEngager inspects a professional's headline and company to assign an ICP segment (SRC-L2).
func CategorizeEngager(headline, company string) EngagerSegment {
	lowerH := strings.ToLower(headline)

	// 1. Talent Partner / Recruiter checks
	talentKeywords := []string{"recruiter", "talent acquisition", "headhunter", "talent partner", "staffing", "recruiting", "talent scout"}
	for _, kw := range talentKeywords {
		if strings.Contains(lowerH, kw) {
			return SegmentTalentPartner
		}
	}

	// 2. Decision Maker checks (VP, Director, Head of, CTO, Founder, Chief)
	decisionKeywords := []string{
		"vp ", "vp of", "vice president", "director", "head of", "chief", "cto",
		"founder", "co-founder", "partner", "engineering manager", "general manager",
	}
	for _, kw := range decisionKeywords {
		if strings.Contains(lowerH, kw) {
			return SegmentDecisionMaker
		}
	}

	// 3. Peer Practitioner checks (Architect, Staff, Principal, Lead, Senior)
	peerKeywords := []string{
		"architect", "staff", "principal", "lead", "senior", "sre", "devops",
		"backend", "frontend", "infrastructure", "platform engineer", "systems engineer",
	}
	for _, kw := range peerKeywords {
		if strings.Contains(lowerH, kw) {
			return SegmentPeerPractitioner
		}
	}

	return SegmentOtherNetwork
}

// ComputeICPBreakdown calculates engager segmentation tallies.
func ComputeICPBreakdown(engagers []EngagerProfile) ICPBreakdown {
	var breakdown ICPBreakdown
	for _, engager := range engagers {
		switch engager.Segment {
		case SegmentDecisionMaker:
			breakdown.DecisionMakers++
		case SegmentPeerPractitioner:
			breakdown.PeerPractitioners++
		case SegmentTalentPartner:
			breakdown.TalentPartners++
		case SegmentOtherNetwork:
			breakdown.OtherNetwork++
		}
	}
	breakdown.HighValueEngagersCount = breakdown.DecisionMakers + breakdown.PeerPractitioners + breakdown.TalentPartners
	return breakdown
}

// ValidateAnalyticsSnapshot ensures compliance with AT-028 & AT-010 invariants.
func ValidateAnalyticsSnapshot(snapshot *PostAnalyticsSnapshot) error {
	if snapshot == nil {
		return ErrInvalidAnalyticsData
	}
	if strings.TrimSpace(snapshot.WorkspaceID) == "" || strings.TrimSpace(snapshot.TenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	if strings.TrimSpace(snapshot.PostURN) == "" {
		return fmt.Errorf("%w: post_urn is required", ErrInvalidAnalyticsData)
	}

	// AT-028 & AT-010 validation:
	// If a metric is marked as "unavailable", it must have an explicit reason and cannot have a non-zero value.
	for metricName, item := range snapshot.Metrics {
		if item.Status == MetricStatusUnavailable {
			if item.Value != 0 {
				return fmt.Errorf("%w: metric '%s' marked unavailable cannot have non-zero value %v", ErrUnpermittedMetricFaking, metricName, item.Value)
			}
			if strings.TrimSpace(item.UnavailableReason) == "" {
				return fmt.Errorf("%w: metric '%s' marked unavailable must specify an explanatory unavailable_reason", ErrInvalidAnalyticsData, metricName)
			}
		}
		if item.Status == MetricStatusModeledEstimate {
			if !item.IsModeled {
				return fmt.Errorf("%w: metric '%s' has status modeled_estimate but is_modeled is false", ErrInvalidAnalyticsData, metricName)
			}
			if strings.TrimSpace(item.ModelBasis) == "" {
				return fmt.Errorf("%w: metric '%s' marked modeled_estimate must provide transparent model_basis explanation (AT-028)", ErrInvalidAnalyticsData, metricName)
			}
		}
	}

	return nil
}

// CuratedCreatorBenchmarks returns verified baseline statistics for comparison (AT-028).
func CuratedCreatorBenchmarks() []PerformanceBenchmark {
	return []PerformanceBenchmark{
		{
			Category:          "engineering_leadership",
			SampleSize:        25,
			AvgReactions:      85.4,
			AvgComments:       22.1,
			AvgReposts:        8.6,
			AvgEngagementRate: 1.45,
			BenchmarkSource:   "Verified historical user post archive",
		},
		{
			Category:          "technical_practitioner",
			SampleSize:        40,
			AvgReactions:      54.2,
			AvgComments:       14.8,
			AvgReposts:        4.1,
			AvgEngagementRate: 1.12,
			BenchmarkSource:   "Verified historical user post archive",
		},
	}
}
