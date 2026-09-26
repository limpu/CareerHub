package career

import (
	"time"
)

// FeedbackSeverity denotes the urgency level of a detected resume issue.
type FeedbackSeverity string

const (
	SeverityCritical FeedbackSeverity = "critical"
	SeverityWarning  FeedbackSeverity = "warning"
	SeverityInfo     FeedbackSeverity = "info"
)

// FeedbackIssue represents a specific flaw or improvement area in a resume section.
type FeedbackIssue struct {
	Section        string           `json:"section"`
	Severity       FeedbackSeverity `json:"severity"`
	Message        string           `json:"message"`
	Recommendation string           `json:"recommendation"`
	AffectedItem   string           `json:"affected_item,omitempty"`
}

// FeedbackSuggestion provides actionable Before/After phrasing for impact metrics.
type FeedbackSuggestion struct {
	Category             string `json:"category"` // "impact", "action_verbs", "brevity", "keywords"
	CurrentText          string `json:"current_text"`
	SuggestedImprovement string `json:"suggested_improvement"`
	Rationale            string `json:"rationale"`
}

// ResumeFeedbackReport is the comprehensive assessment artifact for a master career profile.
type ResumeFeedbackReport struct {
	ID                     string                `json:"id"`
	UserID                 string                `json:"user_id"`
	ProfileID              string                `json:"profile_id"`
	OverallScore           int                   `json:"overall_score"` // 0-100
	ATSReadabilityScore    int                   `json:"ats_readability_score"` // 0-100
	ImpactScore            int                   `json:"impact_score"` // 0-100
	QuantificationRate     float64               `json:"quantification_rate"` // 0.0 - 1.0 (% bullets with measurable metrics)
	ActionVerbDensity      float64               `json:"action_verb_density"` // 0.0 - 1.0 (% bullets starting with strong verbs)
	SectionScores          map[string]int        `json:"section_scores"`
	Strengths              []string              `json:"strengths"`
	CriticalIssues         []FeedbackIssue       `json:"critical_issues"`
	ImprovementSuggestions []FeedbackSuggestion  `json:"improvement_suggestions"`
	AnalyzedAt             time.Time             `json:"analyzed_at"`
}

// MarketSkillDemandItem represents market intelligence on a technical skill (CAR-06, AT-028).
type MarketSkillDemandItem struct {
	SkillName        string    `json:"skill_name"`
	Category         string    `json:"category"` // "Languages", "Cloud & Infra", "Databases", "Frameworks", "Architecture"
	DemandLevel      string    `json:"demand_level"` // "critical", "high", "moderate", "niche"
	DemandPercentile float64   `json:"demand_percentile"` // e.g. 94.5% of analyzed postings
	GrowthYoY        float64   `json:"growth_yoy"` // e.g. +28.4%
	EvidenceDate     time.Time `json:"evidence_date"`
	SampleSize       int       `json:"sample_size"` // Number of job postings analyzed
	SourceCitation   string    `json:"source_citation"`
	IsPossessed      bool      `json:"is_possessed"` // Candidate confirmed fact
	Status           string    `json:"status"` // "possessed_and_in_demand", "market_demand_gap"
}

// SkillsDemandAnalysis delivers evidence-based market requirements while strictly separating
// suggestions from actual possessed candidate facts (CAR-06, AT-003, AT-028).
type SkillsDemandAnalysis struct {
	RoleCategory               string                  `json:"role_category"`
	TotalMarketSkillsAnalyzed  int                     `json:"total_market_skills_analyzed"`
	CandidatePossessedCount    int                     `json:"candidate_possessed_count"`
	CandidateGapCount          int                     `json:"candidate_gap_count"`
	PossessedSkills            []MarketSkillDemandItem `json:"possessed_skills"` // Proven candidate strengths
	MarketGaps                 []MarketSkillDemandItem `json:"market_gaps"` // Market demand suggestions (NOT possessed facts)
	AnalysisDate               time.Time               `json:"analysis_date"`
	DataAvailabilityStatus     string                  `json:"data_availability_status"` // "live_market_sample" vs "unavailable" (AT-028)
}
