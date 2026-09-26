package career

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// BoardSource represents supported job distribution channels.
type BoardSource string

const (
	BoardSourceGreenhouse BoardSource = "greenhouse"
	BoardSourceLever      BoardSource = "lever"
	BoardSourceAshby      BoardSource = "ashby"
	BoardSourceLinkedIn   BoardSource = "linkedin"
	BoardSourceIndeed     BoardSource = "indeed"
	BoardSourceGoogle     BoardSource = "google"
)

// CompensationPeriod defines standardized salary intervals.
type CompensationPeriod string

const (
	PeriodHourly  CompensationPeriod = "hourly"
	PeriodMonthly CompensationPeriod = "monthly"
	PeriodYearly  CompensationPeriod = "yearly"
)

// NormalizedCompensation captures job compensation while strictly preserving unknown
// salaries (never defaulting to 0 or inventing ranges per AT-003, AT-028).
type NormalizedCompensation struct {
	IsDisclosed bool               `json:"is_disclosed"` // false if salary is unlisted
	MinAmount   float64            `json:"min_amount,omitempty"`
	MaxAmount   float64            `json:"max_amount,omitempty"`
	Currency    string             `json:"currency,omitempty"` // 3-letter ISO code e.g. "USD", "EUR"
	Period      CompensationPeriod `json:"period,omitempty"`
	IsEstimated bool               `json:"is_estimated"` // true if calculated by board heuristic rather than employer
}

// JobLocation standardizes physical and remote work arrangements.
type JobLocation struct {
	RawLocation string `json:"raw_location"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	Country     string `json:"country,omitempty"` // 2-letter ISO code or name
	IsRemote    bool   `json:"is_remote"`
	IsHybrid    bool   `json:"is_hybrid"`
}

// DiscoveredJob is the canonical unified multi-board job record (REQ-002, REQ-005, CAR-07).
type DiscoveredJob struct {
	ID               string                 `json:"id"`
	Source           BoardSource            `json:"source"`
	SourceJobID      string                 `json:"source_job_id"`
	CanonicalURL     string                 `json:"canonical_url"` // Striped of tracking params (AT-004)
	DirectApplyURL   string                 `json:"direct_apply_url,omitempty"`
	Title            string                 `json:"title"`
	Company          string                 `json:"company"`
	CompanyLogoURL   string                 `json:"company_logo_url,omitempty"`
	Location         JobLocation            `json:"location"`
	JobType          string                 `json:"job_type"` // "full_time", "contract", "part_time", "internship"
	Description      string                 `json:"description"`
	RequiredSkills   []string               `json:"required_skills"`
	Compensation     NormalizedCompensation `json:"compensation"`
	DatePosted       time.Time              `json:"date_posted"`
	DiscoveredAt     time.Time              `json:"discovered_at"`
	IsDirectEmployer bool                   `json:"is_direct_employer"` // True for Greenhouse/Lever/Ashby ATS
}

// BoardCapabilities documents exact programmatic limits and supported parameters per board (AT-010).
type BoardCapabilities struct {
	BoardName             BoardSource `json:"board_name"`
	DisplayName           string      `json:"display_name"`
	IsDirectATS           bool        `json:"is_direct_ats"`
	SupportsKeywordSearch bool        `json:"supports_keyword_search"`
	SupportsLocationFilter bool       `json:"supports_location_filter"`
	SupportsRemoteFilter  bool        `json:"supports_remote_filter"`
	SupportsSalaryFilter  bool        `json:"supports_salary_filter"`
	RequiresAuthentication bool       `json:"requires_authentication"`
	RateLimitPerMinute    int         `json:"rate_limit_per_minute"`
	Notes                 string      `json:"notes"`
}

// JobSearchQuery represents unified search criteria across multiple job boards.
type JobSearchQuery struct {
	Keywords     string   `json:"keywords"`
	Location     string   `json:"location,omitempty"`
	RemoteOnly   bool     `json:"remote_only"`
	JobTypes     []string `json:"job_types,omitempty"`
	Sources      []BoardSource `json:"sources,omitempty"` // If empty, search all enabled boards
	LimitPerBoard int     `json:"limit_per_board,omitempty"`
}

// JobBoardAdapter defines the contract for multi-board discovery integrations.
type JobBoardAdapter interface {
	Source() BoardSource
	Capabilities() BoardCapabilities
	DiscoverJobs(ctx context.Context, query JobSearchQuery) ([]DiscoveredJob, error)
}

var trackingParamRegex = regexp.MustCompile(`(?i)^(utm_[a-z0-9_]+|refid|ref_id|trackingid|tracking_id|trk|fbclid|gclid|mc_cid|mc_eid|gh_src|[a-z0-9_-]*source.*|[a-z0-9_-]*origin.*)$`)

// NormalizeCanonicalURL strips marketing/tracking query parameters to produce deterministic canonical URLs (AT-004).
func NormalizeCanonicalURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Lowercase host and strip default ports
	u.Host = strings.ToLower(u.Host)
	u.Host = strings.TrimSuffix(u.Host, ":80")
	u.Host = strings.TrimSuffix(u.Host, ":443")

	// Filter query params
	if u.RawQuery != "" {
		q := u.Query()
		for key := range q {
			if trackingParamRegex.MatchString(key) {
				q.Del(key)
			}
		}
		u.RawQuery = q.Encode()
	}

	// Remove trailing slash on path if root isn't empty
	if len(u.Path) > 1 && strings.HasSuffix(u.Path, "/") {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}

	u.Fragment = "" // Strip hash fragments for indexing
	return u.String()
}
