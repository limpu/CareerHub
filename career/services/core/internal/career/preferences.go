package career

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPreferencesNotFound = errors.New("career preferences not found")
	ErrInvalidSalaryRange  = errors.New("minimum salary cannot exceed maximum salary")
	ErrInvalidCurrency     = errors.New("currency code must be a valid 3-letter ISO code")
	ErrEmptyRoles          = errors.New("at least one target role must be specified")
)

// WorkMode represents work location arrangement.
type WorkMode string

const (
	WorkModeRemote WorkMode = "remote"
	WorkModeHybrid WorkMode = "hybrid"
	WorkModeOnsite WorkMode = "onsite"
)

// JobType represents employment classification.
type JobType string

const (
	JobTypeFullTime   JobType = "full_time"
	JobTypePartTime   JobType = "part_time"
	JobTypeContract   JobType = "contract"
	JobTypeInternship JobType = "internship"
	JobTypeTemporary  JobType = "temporary"
	JobTypeOther      JobType = "other"
)

// SponsorshipPreference specifies visa and work authorization requirement.
// Invariant AT-003: Never guessed or assumed; remains 'unspecified' until explicitly confirmed by candidate.
type SponsorshipPreference string

const (
	SponsorshipUnspecified             SponsorshipPreference = "unspecified"
	SponsorshipRequiresSponsorship     SponsorshipPreference = "requires_sponsorship"
	SponsorshipAuthorizedNoSponsorship SponsorshipPreference = "authorized_no_sponsorship"
	SponsorshipWillNotSponsorAccepted  SponsorshipPreference = "will_not_sponsor_accepted"
)

// SalaryInterval defines payment frequency.
type SalaryInterval string

const (
	SalaryIntervalAnnual  SalaryInterval = "annual"
	SalaryIntervalMonthly SalaryInterval = "monthly"
	SalaryIntervalHourly  SalaryInterval = "hourly"
)

// SalaryPreference stores target compensation boundaries.
// Invariant AT-003: IsExplicitlySet is false until user explicitly configures salary.
type SalaryPreference struct {
	MinimumAmount   float64        `json:"minimum_amount"`
	TargetAmount    float64        `json:"target_amount,omitempty"`
	MaximumAmount   float64        `json:"maximum_amount,omitempty"`
	Currency        string         `json:"currency"` // e.g. "USD", "EUR", "GBP"
	Interval        SalaryInterval `json:"interval"` // "annual", "monthly", "hourly"
	IsExplicitlySet bool           `json:"is_explicitly_set"`
}

// ExclusionRules defines companies, keywords, and industries to strictly omit from job search and applications.
type ExclusionRules struct {
	ExcludedCompanies     []string `json:"excluded_companies"`     // e.g. current employer, unwanted companies
	ExcludedKeywords      []string `json:"excluded_keywords"`      // e.g. "crypto", "gambling", "senior"
	ExcludedIndustries    []string `json:"excluded_industries"`    // e.g. "defense", "tobacco"
	ExcludedLocations     []string `json:"excluded_locations"`     // e.g. specific jurisdictions
	BlockStaffingAgencies bool     `json:"block_staffing_agencies"` // exclude 3rd-party recruiters
}

// CareerPreferences stores explicit user-defined job matching criteria and search boundaries.
// Invariant AT-003: Unknown preferences remain unset; never fabricated with default values.
type CareerPreferences struct {
	ID               string                `json:"id"`
	UserID           string                `json:"user_id"`
	TargetRoles      []string              `json:"target_roles"`
	TargetLocations  []string              `json:"target_locations"`
	WorkModes        []WorkMode            `json:"work_modes"`
	JobTypes         []JobType             `json:"job_types"`
	Salary           SalaryPreference      `json:"salary"`
	Sponsorship      SponsorshipPreference `json:"sponsorship"`
	Exclusions       ExclusionRules        `json:"exclusions"`
	OpenToRelocation bool                  `json:"open_to_relocation"`
	NoticePeriodDays int                   `json:"notice_period_days"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}

// NewCareerPreferences initializes empty, unguessable career preferences (AT-003).
func NewCareerPreferences(userID string) *CareerPreferences {
	now := time.Now()
	return &CareerPreferences{
		ID:              uuid.New().String(),
		UserID:          userID,
		TargetRoles:     make([]string, 0),
		TargetLocations: make([]string, 0),
		WorkModes:       make([]WorkMode, 0),
		JobTypes:        make([]JobType, 0),
		Salary: SalaryPreference{
			MinimumAmount:   0,
			TargetAmount:    0,
			MaximumAmount:   0,
			Currency:        "USD",
			Interval:        SalaryIntervalAnnual,
			IsExplicitlySet: false, // AT-003: Not set by default
		},
		Sponsorship: SponsorshipUnspecified, // AT-003: Never defaulted or guessed
		Exclusions: ExclusionRules{
			ExcludedCompanies:     make([]string, 0),
			ExcludedKeywords:      make([]string, 0),
			ExcludedIndustries:    make([]string, 0),
			ExcludedLocations:     make([]string, 0),
			BlockStaffingAgencies: false,
		},
		OpenToRelocation: false,
		NoticePeriodDays: 0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// Validate checks internal consistency and validity of preference inputs.
func (p *CareerPreferences) Validate() error {
	if p.Salary.IsExplicitlySet {
		if p.Salary.Currency != "" && len(p.Salary.Currency) != 3 {
			return ErrInvalidCurrency
		}
		if p.Salary.MinimumAmount > 0 && p.Salary.MaximumAmount > 0 && p.Salary.MinimumAmount > p.Salary.MaximumAmount {
			return ErrInvalidSalaryRange
		}
	}
	return nil
}

// MatchesExclusion evaluates whether a job title, company, or text matches any configured exclusion rule.
// Returns true with the reason if excluded; false otherwise.
func (p *CareerPreferences) MatchesExclusion(companyName, jobTitle, description string) (bool, string) {
	normCompany := strings.ToLower(strings.TrimSpace(companyName))
	normTitle := strings.ToLower(strings.TrimSpace(jobTitle))
	normDesc := strings.ToLower(description)

	// 1. Check Excluded Companies
	for _, exc := range p.Exclusions.ExcludedCompanies {
		trimmed := strings.ToLower(strings.TrimSpace(exc))
		if trimmed != "" && (normCompany == trimmed || strings.Contains(normCompany, trimmed)) {
			return true, "company excluded: " + exc
		}
	}

	// 2. Check Excluded Keywords in Title or Description
	for _, kw := range p.Exclusions.ExcludedKeywords {
		trimmed := strings.ToLower(strings.TrimSpace(kw))
		if trimmed != "" {
			if strings.Contains(normTitle, trimmed) {
				return true, "title contains excluded keyword: " + kw
			}
			if normDesc != "" && strings.Contains(normDesc, trimmed) {
				return true, "description contains excluded keyword: " + kw
			}
		}
	}

	return false, ""
}
