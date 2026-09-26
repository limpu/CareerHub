package career

import (
	"strings"
	"time"
)

// Multi-language employment type mapping dictionary (SRC-C7 JobSpy pattern).
var jobTypeDictionary = map[string]JobType{
	// English
	"full time":   JobTypeFullTime,
	"full-time":   JobTypeFullTime,
	"full_time":   JobTypeFullTime,
	"permanent":   JobTypeFullTime,
	"part time":   JobTypePartTime,
	"part-time":   JobTypePartTime,
	"part_time":   JobTypePartTime,
	"contract":    JobTypeContract,
	"contractor":  JobTypeContract,
	"freelance":   JobTypeContract,
	"temp":        JobTypeTemporary,
	"temporary":   JobTypeTemporary,
	"intern":      JobTypeInternship,
	"internship":  JobTypeInternship,
	"co-op":       JobTypeInternship,

	// German
	"vollzeit":    JobTypeFullTime,
	"festanstellung": JobTypeFullTime,
	"teilzeit":    JobTypePartTime,
	"befristet":   JobTypeContract,
	"praktikum":   JobTypeInternship,
	"werkstudent": JobTypeInternship,

	// French
	"temps plein": JobTypeFullTime,
	"cdi":         JobTypeFullTime,
	"temps partiel": JobTypePartTime,
	"cdd":         JobTypeContract,
	"stage":       JobTypeInternship,
	"alternance":  JobTypeInternship,

	// Spanish
	"tiempo completo": JobTypeFullTime,
	"jornada completa": JobTypeFullTime,
	"medio tiempo":    JobTypePartTime,
	"tiempo parcial":  JobTypePartTime,
	"contrato":        JobTypeContract,
	"practicas":       JobTypeInternship,
	"pasantia":        JobTypeInternship,
}

// ParseStandardJobType normalizes regional and multi-lingual job type strings into canonical types.
func ParseStandardJobType(raw string) JobType {
	clean := strings.ToLower(strings.TrimSpace(raw))
	if clean == "" {
		return JobTypeOther
	}

	if jt, ok := jobTypeDictionary[clean]; ok {
		return jt
	}

	// Substring / fuzzy checks
	for phrase, jt := range jobTypeDictionary {
		if strings.Contains(clean, phrase) {
			return jt
		}
	}
	return JobTypeOther
}

// AdvancedJobFilter captures all fine-grained filtering parameters for job discovery (IMP-CAR-08, CAR-08).
type AdvancedJobFilter struct {
	MaxAgeDays               int        `json:"max_age_days,omitempty"` // 1, 3, 7, 14, 30 days
	WorkModes                []string   `json:"work_modes,omitempty"` // "remote", "hybrid", "on_site"
	JobTypes                 []JobType  `json:"job_types,omitempty"` // "full_time", "contract", etc.
	CompanyInclusions        []string          `json:"company_inclusions,omitempty"` // Whitelist
	CompanyExclusions        []string          `json:"company_exclusions,omitempty"` // Blacklist (IMP-CAR-03)
	TitleKeywordsMustInclude []string          `json:"title_keywords_must_include,omitempty"`
	TitleKeywordsMustExclude []string          `json:"title_keywords_must_exclude,omitempty"`
	MinimumAnnualSalary      float64           `json:"minimum_annual_salary,omitempty"`
	SalaryCurrency           string            `json:"salary_currency,omitempty"`
	IncludeUndisclosedSalary bool              `json:"include_undisclosed_salary"` // default true (AT-003, AT-028)
}

// FilterExecutionMode records how a specific criterion was evaluated for transparency (CAR-08).
type FilterExecutionMode string

const (
	FilterModeUpstreamNative FilterExecutionMode = "upstream_native" // Supported by board query
	FilterModePostFiltered   FilterExecutionMode = "post_filtered"   // Applied in Go service post-fetch
	FilterModeUnsupported    FilterExecutionMode = "unsupported"     // Feature unavailable on board
)

// FilterAuditReport provides complete visibility into how filters were applied (CAR-08, AT-010).
type FilterAuditReport struct {
	TotalCandidatesInput   int                            `json:"total_candidates_input"`
	TotalResultsRetained   int                            `json:"total_results_retained"`
	DisqualifiedByAge      int                            `json:"disqualified_by_age"`
	DisqualifiedByMode     int                            `json:"disqualified_by_mode"`
	DisqualifiedByType     int                            `json:"disqualified_by_type"`
	DisqualifiedByCompany  int                            `json:"disqualified_by_company"`
	DisqualifiedBySalary   int                            `json:"disqualified_by_salary"`
	FilterEvaluations      map[string]FilterExecutionMode `json:"filter_evaluations"`
}

// FilterEvaluator applies fine-grained filtering rules while honoring board capabilities.
type FilterEvaluator struct{}

func NewFilterEvaluator() *FilterEvaluator {
	return &FilterEvaluator{}
}

// Evaluate applies AdvancedJobFilter to a list of DiscoveredJobs and returns filtered results + audit.
func (fe *FilterEvaluator) Evaluate(jobs []DiscoveredJob, filter AdvancedJobFilter, now time.Time) ([]DiscoveredJob, FilterAuditReport) {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	report := FilterAuditReport{
		TotalCandidatesInput: len(jobs),
		FilterEvaluations:    make(map[string]FilterExecutionMode),
	}

	// Document execution modes
	report.FilterEvaluations["remote_filter"] = FilterModeUpstreamNative
	report.FilterEvaluations["job_type_filter"] = FilterModePostFiltered
	report.FilterEvaluations["company_exclusion"] = FilterModePostFiltered
	report.FilterEvaluations["age_filter"] = FilterModePostFiltered
	report.FilterEvaluations["salary_threshold"] = FilterModePostFiltered

	var filtered []DiscoveredJob

	for _, job := range jobs {
		// 1. Age Filter
		if filter.MaxAgeDays > 0 && !job.DatePosted.IsZero() {
			cutoff := now.AddDate(0, 0, -filter.MaxAgeDays)
			if job.DatePosted.Before(cutoff) {
				report.DisqualifiedByAge++
				continue
			}
		}

		// 2. Work Mode Filter (Remote / Hybrid / On-site)
		if len(filter.WorkModes) > 0 {
			modeMatch := false
			for _, m := range filter.WorkModes {
				cleanMode := strings.ToLower(strings.TrimSpace(m))
				if cleanMode == "remote" && job.Location.IsRemote {
					modeMatch = true
					break
				}
				if cleanMode == "hybrid" && job.Location.IsHybrid {
					modeMatch = true
					break
				}
				if (cleanMode == "on_site" || cleanMode == "onsite") && !job.Location.IsRemote && !job.Location.IsHybrid {
					modeMatch = true
					break
				}
			}
			if !modeMatch {
				report.DisqualifiedByMode++
				continue
			}
		}

		// 3. Job Type Filter (using multi-language parser)
		if len(filter.JobTypes) > 0 {
			parsedJobType := ParseStandardJobType(job.JobType)
			typeMatch := false
			for _, targetType := range filter.JobTypes {
				if parsedJobType == targetType {
					typeMatch = true
					break
				}
			}
			if !typeMatch {
				report.DisqualifiedByType++
				continue
			}
		}

		// 4. Company Inclusions & Exclusions (IMP-CAR-03)
		if len(filter.CompanyInclusions) > 0 {
			companyMatch := false
			for _, inc := range filter.CompanyInclusions {
				if strings.EqualFold(strings.TrimSpace(job.Company), strings.TrimSpace(inc)) {
					companyMatch = true
					break
				}
			}
			if !companyMatch {
				report.DisqualifiedByCompany++
				continue
			}
		}

		if len(filter.CompanyExclusions) > 0 {
			excluded := false
			for _, exc := range filter.CompanyExclusions {
				excClean := strings.ToLower(strings.TrimSpace(exc))
				if excClean != "" && strings.Contains(strings.ToLower(job.Company), excClean) {
					excluded = true
					break
				}
			}
			if excluded {
				report.DisqualifiedByCompany++
				continue
			}
		}

		// 5. Title Keywords Inclusions / Exclusions
		if len(filter.TitleKeywordsMustInclude) > 0 {
			titleLower := strings.ToLower(job.Title)
			allMatch := true
			for _, kw := range filter.TitleKeywordsMustInclude {
				if !strings.Contains(titleLower, strings.ToLower(strings.TrimSpace(kw))) {
					allMatch = false
					break
				}
			}
			if !allMatch {
				continue
			}
		}

		if len(filter.TitleKeywordsMustExclude) > 0 {
			titleLower := strings.ToLower(job.Title)
			hasExcluded := false
			for _, kw := range filter.TitleKeywordsMustExclude {
				if kw != "" && strings.Contains(titleLower, strings.ToLower(strings.TrimSpace(kw))) {
					hasExcluded = true
					break
				}
			}
			if hasExcluded {
				continue
			}
		}

		// 6. Salary Threshold Filter (AT-003, AT-028)
		if filter.MinimumAnnualSalary > 0 {
			if !job.Compensation.IsDisclosed {
				// Undisclosed salary
				if !filter.IncludeUndisclosedSalary {
					report.DisqualifiedBySalary++
					continue
				}
			} else {
				// Normalize to annual if hourly or monthly
				annualAmount := job.Compensation.MaxAmount
				if annualAmount == 0 {
					annualAmount = job.Compensation.MinAmount
				}

				if job.Compensation.Period == PeriodHourly {
					annualAmount = annualAmount * 2080 // 40 hrs * 52 weeks
				} else if job.Compensation.Period == PeriodMonthly {
					annualAmount = annualAmount * 12
				}

				if annualAmount > 0 && annualAmount < filter.MinimumAnnualSalary {
					report.DisqualifiedBySalary++
					continue
				}
			}
		}

		filtered = append(filtered, job)
	}

	report.TotalResultsRetained = len(filtered)
	return filtered, report
}
