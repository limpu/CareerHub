package career

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

var (
	ErrInvalidScoringWeights = errors.New("scoring weights must be non-negative and sum exactly to 100")
	ErrJobNotFound           = errors.New("job not found for matching")
)

// MatchStatus represents the three-state evaluation for any job requirement (AT-003, AT-028).
// Invariant AT-028: Unknown data is strictly distinguished from negative mismatch.
type MatchStatus string

const (
	MatchStatusMatched  MatchStatus = "matched"
	MatchStatusMismatch MatchStatus = "mismatch"
	MatchStatusUnknown  MatchStatus = "unknown"
)

// HardGateType identifies the non-negotiable eligibility criteria (CAR-09).
type HardGateType string

const (
	HardGateCompanyExclusion   HardGateType = "company_exclusion"
	HardGateWorkAuthorization  HardGateType = "work_authorization"
	HardGateWorkModeRelocation HardGateType = "work_mode_relocation"
	HardGateHardMinExperience  HardGateType = "minimum_experience"
)

// HardGateResult describes the outcome of an eligibility gate.
type HardGateResult struct {
	GateType  HardGateType `json:"gate_type"`
	GateName  string       `json:"gate_name"`
	Passed    bool         `json:"passed"`
	IsUnknown bool         `json:"is_unknown"` // True if decision was neutral due to missing employer/candidate disclosure
	Reason    string       `json:"reason"`
}

// ScoringWeights defines the user-configurable or default percentage weights.
// Invariant CAR-09: Sum of all weights must equal 100%.
type ScoringWeights struct {
	SkillsWeight           int `json:"skills_weight"`             // Default: 40
	TitleExperienceWeight  int `json:"title_experience_weight"`   // Default: 30
	LocationWorkModeWeight int `json:"location_work_mode_weight"` // Default: 15
	CompensationWeight     int `json:"compensation_weight"`       // Default: 15
}

// DefaultScoringWeights returns the calibrated baseline weights summing to 100.
func DefaultScoringWeights() ScoringWeights {
	return ScoringWeights{
		SkillsWeight:           40,
		TitleExperienceWeight:  30,
		LocationWorkModeWeight: 15,
		CompensationWeight:     15,
	}
}

// Validate ensures weights sum to exactly 100 and none are negative.
func (w ScoringWeights) Validate() error {
	if w.SkillsWeight < 0 || w.TitleExperienceWeight < 0 || w.LocationWorkModeWeight < 0 || w.CompensationWeight < 0 {
		return ErrInvalidScoringWeights
	}
	sum := w.SkillsWeight + w.TitleExperienceWeight + w.LocationWorkModeWeight + w.CompensationWeight
	if sum != 100 {
		return fmt.Errorf("%w: current sum is %d", ErrInvalidScoringWeights, sum)
	}
	return nil
}

// RequirementMatch represents a granular evaluated requirement with explicit evidence.
type RequirementMatch struct {
	Category          string      `json:"category"` // "skill", "role", "location", "compensation", "gate"
	Requirement       string      `json:"requirement"`
	Status            MatchStatus `json:"status"`   // "matched", "mismatch", "unknown"
	CandidateEvidence string      `json:"candidate_evidence"`
	Confidence        float64     `json:"confidence"` // 1.0 = verified, 0.0 = unknown
}

// MatchScoreBreakdown provides dimensional transparency into how the total score was constructed.
type MatchScoreBreakdown struct {
	SkillsScore                   float64        `json:"skills_score"`                     // 0.0 - 100.0
	SkillsWeightedScore           float64        `json:"skills_weighted_score"`            // SkillsScore * (SkillsWeight / 100)
	TitleExperienceScore          float64        `json:"title_experience_score"`           // 0.0 - 100.0
	TitleExperienceWeightedScore  float64        `json:"title_experience_weighted_score"`  // TitleScore * (TitleWeight / 100)
	LocationWorkModeScore         float64        `json:"location_work_mode_score"`          // 0.0 - 100.0
	LocationWorkModeWeightedScore float64        `json:"location_work_mode_weighted_score"`
	CompensationScore             float64        `json:"compensation_score"`               // 0.0 - 100.0
	CompensationWeightedScore     float64        `json:"compensation_weighted_score"`
	WeightsUsed                   ScoringWeights `json:"weights_used"`
}

// JobMatchResult is the complete explainable matching record (CAR-09, AT-003, AT-028).
type JobMatchResult struct {
	JobID               string              `json:"job_id"`
	AlgorithmVersion    string              `json:"algorithm_version"` // e.g. "v1.0-explainable"
	EvaluatedAt         time.Time           `json:"evaluated_at"`
	IsEligible          bool                `json:"is_eligible"`       // False if any hard gate fails
	OverallScore        float64             `json:"overall_score"`     // 0.0 to 100.0
	FitTier             string              `json:"fit_tier"`          // "strong_match", "moderate_match", "low_match", "ineligible"
	HardGates           []HardGateResult    `json:"hard_gates"`
	Breakdown           MatchScoreBreakdown `json:"breakdown"`
	MatchedSkills       []string            `json:"matched_skills"`
	MissingSkills       []string            `json:"missing_skills"`        // Evidence-based gap list (CAR-06)
	UnknownRequirements []string            `json:"unknown_requirements"` // Requirements with unavailable data (AT-028)
	Explanations        []string            `json:"explanations"`          // Auditable human-readable rationale
	RequirementMatches  []RequirementMatch  `json:"requirement_matches"`
}

// MatchEngine executes transparent multi-factor job matching.
type MatchEngine struct{}

// NewMatchEngine constructs an explainable matching engine.
func NewMatchEngine() *MatchEngine {
	return &MatchEngine{}
}

// EvaluateJobMatch compares candidate profile & preferences against a discovered job.
func (e *MatchEngine) EvaluateJobMatch(
	profile *MasterCareerProfile,
	preferences *CareerPreferences,
	job *DiscoveredJob,
	customWeights *ScoringWeights,
) (*JobMatchResult, error) {
	weights := DefaultScoringWeights()
	if customWeights != nil {
		if err := customWeights.Validate(); err != nil {
			return nil, err
		}
		weights = *customWeights
	}

	result := &JobMatchResult{
		JobID:               job.ID,
		AlgorithmVersion:    "v1.0-explainable",
		EvaluatedAt:         time.Now(),
		IsEligible:          true,
		HardGates:           make([]HardGateResult, 0),
		MatchedSkills:       make([]string, 0),
		MissingSkills:       make([]string, 0),
		UnknownRequirements: make([]string, 0),
		Explanations:        make([]string, 0),
		RequirementMatches:  make([]RequirementMatch, 0),
	}

	// Step 1: Evaluate Hard Eligibility Gates (CAR-09)
	e.evaluateHardGates(profile, preferences, job, result)

	// Step 2: Evaluate Skills Match (0 - 100)
	skillsScore := e.evaluateSkills(profile, job, result)

	// Step 3: Evaluate Title & Experience Match (0 - 100)
	titleExpScore := e.evaluateTitleAndExperience(profile, preferences, job, result)

	// Step 4: Evaluate Location & Work Mode Match (0 - 100)
	locModeScore := e.evaluateLocationAndWorkMode(preferences, job, result)

	// Step 5: Evaluate Compensation Match (0 - 100) with Undisclosed Neutrality (AT-003, AT-028)
	compScore := e.evaluateCompensation(preferences, job, result)

	// Step 6: Compute Weighted Breakdown
	skillsWeighted := (skillsScore * float64(weights.SkillsWeight)) / 100.0
	titleExpWeighted := (titleExpScore * float64(weights.TitleExperienceWeight)) / 100.0
	locModeWeighted := (locModeScore * float64(weights.LocationWorkModeWeight)) / 100.0
	compWeighted := (compScore * float64(weights.CompensationWeight)) / 100.0

	overall := skillsWeighted + titleExpWeighted + locModeWeighted + compWeighted
	overall = math.Round(overall*10) / 10 // round to 1 decimal place

	// If candidate failed hard gate, overall score cannot qualify as eligible
	if !result.IsEligible {
		result.FitTier = "ineligible"
		result.Explanations = append([]string{"INELIGIBLE: Hard eligibility barrier failed. Application is disqualified."}, result.Explanations...)
	} else {
		if overall >= 80.0 {
			result.FitTier = "strong_match"
		} else if overall >= 60.0 {
			result.FitTier = "moderate_match"
		} else {
			result.FitTier = "low_match"
		}
	}

	result.OverallScore = overall
	result.Breakdown = MatchScoreBreakdown{
		SkillsScore:                   math.Round(skillsScore*10) / 10,
		SkillsWeightedScore:           math.Round(skillsWeighted*10) / 10,
		TitleExperienceScore:          math.Round(titleExpScore*10) / 10,
		TitleExperienceWeightedScore:  math.Round(titleExpWeighted*10) / 10,
		LocationWorkModeScore:         math.Round(locModeScore*10) / 10,
		LocationWorkModeWeightedScore: math.Round(locModeWeighted*10) / 10,
		CompensationScore:             math.Round(compScore*10) / 10,
		CompensationWeightedScore:     math.Round(compWeighted*10) / 10,
		WeightsUsed:                   weights,
	}

	return result, nil
}

// evaluateHardGates evaluates absolute eligibility criteria.
func (e *MatchEngine) evaluateHardGates(
	profile *MasterCareerProfile,
	preferences *CareerPreferences,
	job *DiscoveredJob,
	result *JobMatchResult,
) {
	// Gate 1: Company Exclusion Blacklist (CAR-08)
	if preferences != nil && len(preferences.Exclusions.ExcludedCompanies) > 0 {
		compLower := strings.ToLower(job.Company)
		for _, exc := range preferences.Exclusions.ExcludedCompanies {
			if strings.TrimSpace(exc) == "" {
				continue
			}
			if strings.Contains(compLower, strings.ToLower(strings.TrimSpace(exc))) {
				result.IsEligible = false
				result.HardGates = append(result.HardGates, HardGateResult{
					GateType:  HardGateCompanyExclusion,
					GateName:  "Company Exclusion Blacklist",
					Passed:    false,
					IsUnknown: false,
					Reason:    fmt.Sprintf("Employer %q matches excluded company %q in career preferences", job.Company, exc),
				})
				result.RequirementMatches = append(result.RequirementMatches, RequirementMatch{
					Category:          "gate",
					Requirement:       "Company Exclusion",
					Status:            MatchStatusMismatch,
					CandidateEvidence: fmt.Sprintf("Candidate blacklisted %q", exc),
					Confidence:        1.0,
				})
				return
			}
		}
	}
	result.HardGates = append(result.HardGates, HardGateResult{
		GateType:  HardGateCompanyExclusion,
		GateName:  "Company Exclusion Blacklist",
		Passed:    true,
		IsUnknown: false,
		Reason:    "Employer is not on candidate's exclusion blacklist",
	})

	// Gate 2: Work Authorization / Visa Sponsorship
	descLower := strings.ToLower(job.Description)
	requiresSponsorship := preferences != nil && preferences.Sponsorship == SponsorshipRequiresSponsorship
	jobDisallowsSponsorship := strings.Contains(descLower, "no sponsorship") ||
		strings.Contains(descLower, "must be a us citizen") ||
		strings.Contains(descLower, "unable to sponsor") ||
		strings.Contains(descLower, "without sponsorship")

	if requiresSponsorship && jobDisallowsSponsorship {
		result.IsEligible = false
		result.HardGates = append(result.HardGates, HardGateResult{
			GateType:  HardGateWorkAuthorization,
			GateName:  "Visa Sponsorship Eligibility",
			Passed:    false,
			IsUnknown: false,
			Reason:    "Candidate requires visa sponsorship, but job posting explicitly states no sponsorship is provided",
		})
		result.RequirementMatches = append(result.RequirementMatches, RequirementMatch{
			Category:          "gate",
			Requirement:       "Work Authorization Sponsorship",
			Status:            MatchStatusMismatch,
			CandidateEvidence: "Requires visa sponsorship; employer posting prohibits sponsorship",
			Confidence:        1.0,
		})
	} else if requiresSponsorship && !strings.Contains(descLower, "sponsor") {
		// Job does not state whether it sponsors -> treated as UNKNOWN per AT-028, not failed closed
		result.HardGates = append(result.HardGates, HardGateResult{
			GateType:  HardGateWorkAuthorization,
			GateName:  "Visa Sponsorship Eligibility",
			Passed:    true,
			IsUnknown: true,
			Reason:    "Job posting does not explicitly disclose sponsorship policy; treated as pending verification (AT-028)",
		})
		result.UnknownRequirements = append(result.UnknownRequirements, "Employer Visa Sponsorship Policy")
		result.RequirementMatches = append(result.RequirementMatches, RequirementMatch{
			Category:          "gate",
			Requirement:       "Work Authorization Sponsorship",
			Status:            MatchStatusUnknown,
			CandidateEvidence: "Candidate requires sponsorship; posting lacks explicit sponsorship policy",
			Confidence:        0.0,
		})
	} else {
		result.HardGates = append(result.HardGates, HardGateResult{
			GateType:  HardGateWorkAuthorization,
			GateName:  "Visa Sponsorship Eligibility",
			Passed:    true,
			IsUnknown: false,
			Reason:    "Work authorization criteria satisfied or no sponsorship required",
		})
	}

	// Gate 3: Work Mode / Relocation constraint
	if !job.Location.IsRemote && preferences != nil && len(preferences.WorkModes) == 1 && preferences.WorkModes[0] == WorkModeRemote {
		if !preferences.OpenToRelocation {
			// Check if candidate lives in same city
			candidateCity := ""
			if profile != nil {
				candidateCity = strings.ToLower(profile.Contact.Location)
			}
			jobCity := strings.ToLower(job.Location.RawLocation)
			if candidateCity == "" || !strings.Contains(jobCity, candidateCity) {
				result.IsEligible = false
				result.HardGates = append(result.HardGates, HardGateResult{
					GateType:  HardGateWorkModeRelocation,
					GateName:  "Strict Remote Preference",
					Passed:    false,
					IsUnknown: false,
					Reason:    fmt.Sprintf("Job is strictly on-site in %q, while candidate prefers 100%% remote and is not open to relocation", job.Location.RawLocation),
				})
				result.RequirementMatches = append(result.RequirementMatches, RequirementMatch{
					Category:          "gate",
					Requirement:       "Work Mode Constraint",
					Status:            MatchStatusMismatch,
					CandidateEvidence: "Candidate is strictly remote and not open to relocation",
					Confidence:        1.0,
				})
				return
			}
		}
	}

	result.HardGates = append(result.HardGates, HardGateResult{
		GateType:  HardGateWorkModeRelocation,
		GateName:  "Work Mode & Relocation",
		Passed:    true,
		IsUnknown: false,
		Reason:    "Work arrangement matches candidate preferences or location is compatible",
	})
}

// evaluateSkills computes the skill overlap percentage strictly based on confirmed skills (AT-003).
func (e *MatchEngine) evaluateSkills(
	profile *MasterCareerProfile,
	job *DiscoveredJob,
	result *JobMatchResult,
) float64 {
	if len(job.RequiredSkills) == 0 {
		result.Explanations = append(result.Explanations, "Skills Match: Job lists no explicit required skills (awarded neutral full score).")
		return 100.0
	}

	confirmedSkillsMap := make(map[string]bool)
	if profile != nil {
		for _, s := range profile.Skills {
			confirmedSkillsMap[strings.ToLower(strings.TrimSpace(s.Name))] = true
		}
	}

	matchedCount := 0
	for _, reqSkill := range job.RequiredSkills {
		cleanReq := strings.ToLower(strings.TrimSpace(reqSkill))
		if confirmedSkillsMap[cleanReq] {
			matchedCount++
			result.MatchedSkills = append(result.MatchedSkills, reqSkill)
			result.RequirementMatches = append(result.RequirementMatches, RequirementMatch{
				Category:          "skill",
				Requirement:       reqSkill,
				Status:            MatchStatusMatched,
				CandidateEvidence: fmt.Sprintf("Confirmed skill in candidate profile: %s", reqSkill),
				Confidence:        1.0,
			})
		} else {
			// Candidate has not confirmed this skill -> gap analysis (CAR-06, AT-003)
			result.MissingSkills = append(result.MissingSkills, reqSkill)
			result.RequirementMatches = append(result.RequirementMatches, RequirementMatch{
				Category:          "skill",
				Requirement:       reqSkill,
				Status:            MatchStatusMismatch,
				CandidateEvidence: "Not listed in candidate confirmed skills (skill gap)",
				Confidence:        1.0,
			})
		}
	}

	score := (float64(matchedCount) / float64(len(job.RequiredSkills))) * 100.0
	result.Explanations = append(result.Explanations, fmt.Sprintf(
		"Skills Match: %d/%d confirmed skills matched (%.1f%%) — Matched: %s. Missing gaps: %s.",
		matchedCount,
		len(job.RequiredSkills),
		score,
		strings.Join(result.MatchedSkills, ", "),
		strings.Join(result.MissingSkills, ", "),
	))
	return score
}

// evaluateTitleAndExperience calculates role alignment and seniority fit.
func (e *MatchEngine) evaluateTitleAndExperience(
	profile *MasterCareerProfile,
	preferences *CareerPreferences,
	job *DiscoveredJob,
	result *JobMatchResult,
) float64 {
	score := 50.0 // baseline neutral

	jobTitleLower := strings.ToLower(job.Title)

	// Check against preferences TargetRoles
	titleMatched := false
	if preferences != nil && len(preferences.TargetRoles) > 0 {
		for _, role := range preferences.TargetRoles {
			roleLower := strings.ToLower(strings.TrimSpace(role))
			if roleLower != "" && strings.Contains(jobTitleLower, roleLower) {
				titleMatched = true
				score += 30.0
				result.RequirementMatches = append(result.RequirementMatches, RequirementMatch{
					Category:          "role",
					Requirement:       fmt.Sprintf("Target Role: %s", role),
					Status:            MatchStatusMatched,
					CandidateEvidence: fmt.Sprintf("Job title %q matches target role %q", job.Title, role),
					Confidence:        1.0,
				})
				break
			}
		}
	}

	// Check against candidate previous titles in experiences
	if !titleMatched && profile != nil {
		for _, exp := range profile.Experiences {
			expTitleLower := strings.ToLower(exp.Title)
			if expTitleLower != "" && (strings.Contains(jobTitleLower, expTitleLower) || strings.Contains(expTitleLower, jobTitleLower)) {
				titleMatched = true
				score += 25.0
				result.RequirementMatches = append(result.RequirementMatches, RequirementMatch{
					Category:          "role",
					Requirement:       "Experience Title Alignment",
					Status:            MatchStatusMatched,
					CandidateEvidence: fmt.Sprintf("Job title %q aligns with previous title %q at %s", job.Title, exp.Title, exp.Company),
					Confidence:        1.0,
				})
				break
			}
		}
	}

	// Seniority check
	totalYearsExp := 0
	if profile != nil {
		totalYearsExp = calculateTotalYearsExperience(profile.Experiences)
	}

	isSeniorJob := strings.Contains(jobTitleLower, "senior") || strings.Contains(jobTitleLower, "lead") || strings.Contains(jobTitleLower, "staff") || strings.Contains(jobTitleLower, "principal")
	isJuniorJob := strings.Contains(jobTitleLower, "junior") || strings.Contains(jobTitleLower, "intern") || strings.Contains(jobTitleLower, "associate")

	if isSeniorJob {
		if totalYearsExp >= 5 {
			score += 20.0
		} else if totalYearsExp >= 3 {
			score += 10.0
		} else {
			score -= 15.0
			result.Explanations = append(result.Explanations, fmt.Sprintf("Seniority note: Role requires senior experience, candidate has %d confirmed years.", totalYearsExp))
		}
	} else if isJuniorJob {
		if totalYearsExp <= 3 {
			score += 20.0
		} else {
			score += 10.0 // slight overqualification, not harsh penalty
		}
	} else {
		// Mid-level
		if totalYearsExp >= 2 {
			score += 20.0
		} else {
			score += 10.0
		}
	}

	if score > 100.0 {
		score = 100.0
	}
	if score < 0.0 {
		score = 0.0
	}

	result.Explanations = append(result.Explanations, fmt.Sprintf(
		"Title & Experience Match: Evaluated at %.1f%% (Candidate has %d confirmed years of professional experience).",
		score,
		totalYearsExp,
	))
	return score
}

// evaluateLocationAndWorkMode computes location suitability.
func (e *MatchEngine) evaluateLocationAndWorkMode(
	preferences *CareerPreferences,
	job *DiscoveredJob,
	result *JobMatchResult,
) float64 {
	if preferences == nil {
		return 75.0 // neutral
	}

	score := 50.0

	// Work Mode matching
	if len(preferences.WorkModes) > 0 {
		jobIsRemote := job.Location.IsRemote
		jobIsHybrid := job.Location.IsHybrid || strings.Contains(strings.ToLower(job.Location.RawLocation), "hybrid")

		modeMatched := false
		for _, m := range preferences.WorkModes {
			if m == WorkModeRemote && jobIsRemote {
				modeMatched = true
				break
			}
			if m == WorkModeHybrid && jobIsHybrid {
				modeMatched = true
				break
			}
			if m == WorkModeOnsite && !jobIsRemote && !jobIsHybrid {
				modeMatched = true
				break
			}
		}

		if modeMatched {
			score += 30.0
		} else if jobIsRemote {
			// Job is remote: candidate didn't strictly request remote, but remote is generally accommodating
			score += 20.0
		} else {
			score -= 10.0
		}
	} else {
		score += 25.0
	}

	// Target Locations matching
	if len(preferences.TargetLocations) > 0 {
		jobLocLower := strings.ToLower(job.Location.RawLocation)
		locMatched := false
		for _, loc := range preferences.TargetLocations {
			if loc != "" && strings.Contains(jobLocLower, strings.ToLower(loc)) {
				locMatched = true
				break
			}
		}
		if locMatched || job.Location.IsRemote {
			score += 20.0
		}
	} else {
		score += 20.0
	}

	if score > 100.0 {
		score = 100.0
	}
	if score < 0.0 {
		score = 0.0
	}

	result.Explanations = append(result.Explanations, fmt.Sprintf(
		"Location & Work Mode Match: Evaluated at %.1f%% (Job Location: %s, Remote: %t).",
		score,
		job.Location.RawLocation,
		job.Location.IsRemote,
	))
	return score
}

// evaluateCompensation evaluates salary expectations with strict preservation of undisclosed salaries (AT-003, AT-028).
func (e *MatchEngine) evaluateCompensation(
	preferences *CareerPreferences,
	job *DiscoveredJob,
	result *JobMatchResult,
) float64 {
	// Case 1: Employer did not disclose salary range (AT-003, AT-028)
	if !job.Compensation.IsDisclosed || job.Compensation.MaxAmount == 0 {
		result.UnknownRequirements = append(result.UnknownRequirements, "Employer Compensation Disclosure")
		result.RequirementMatches = append(result.RequirementMatches, RequirementMatch{
			Category:          "compensation",
			Requirement:       "Salary Disclosure",
			Status:            MatchStatusUnknown,
			CandidateEvidence: "Employer has undisclosed salary; evaluated neutrally without penalty (AT-003, AT-028)",
			Confidence:        0.0,
		})
		result.Explanations = append(result.Explanations, "Compensation Match: Employer salary is undisclosed; awarded neutral score (80.0%) per AT-028 zero-fabrication rules.")
		return 80.0
	}

	// Case 2: Candidate has not configured a minimum salary
	if preferences == nil || !preferences.Salary.IsExplicitlySet || preferences.Salary.MinimumAmount == 0 {
		result.Explanations = append(result.Explanations, fmt.Sprintf(
			"Compensation Match: Job discloses range %s %.0f - %.0f; candidate salary preference unset (awarded neutral full score).",
			job.Compensation.Currency,
			job.Compensation.MinAmount,
			job.Compensation.MaxAmount,
		))
		return 100.0
	}

	jobMax := float64(job.Compensation.MaxAmount)
	candidateMin := preferences.Salary.MinimumAmount

	if jobMax >= candidateMin {
		// Meets or exceeds minimum expectation
		ratio := jobMax / candidateMin
		score := 90.0 + math.Min(10.0, (ratio-1.0)*20.0)
		result.RequirementMatches = append(result.RequirementMatches, RequirementMatch{
			Category:          "compensation",
			Requirement:       fmt.Sprintf("Minimum Base Salary: %.0f %s", candidateMin, preferences.Salary.Currency),
			Status:            MatchStatusMatched,
			CandidateEvidence: fmt.Sprintf("Job max budget %.0f meets or exceeds candidate minimum %.0f", jobMax, candidateMin),
			Confidence:        1.0,
		})
		result.Explanations = append(result.Explanations, fmt.Sprintf(
			"Compensation Match: Disclosed budget (up to %s %.0f) satisfies candidate minimum of %s %.0f (Score: %.1f%%).",
			job.Compensation.Currency,
			job.Compensation.MaxAmount,
			preferences.Salary.Currency,
			candidateMin,
			score,
		))
		return score
	}

	// Below candidate minimum
	shortfallPercent := ((candidateMin - jobMax) / candidateMin) * 100.0
	var score float64
	if shortfallPercent <= 15.0 {
		score = 65.0 // minor gap
	} else if shortfallPercent <= 30.0 {
		score = 40.0 // notable gap
	} else {
		score = 20.0 // major discrepancy
	}

	result.RequirementMatches = append(result.RequirementMatches, RequirementMatch{
		Category:          "compensation",
		Requirement:       fmt.Sprintf("Minimum Base Salary: %.0f %s", candidateMin, preferences.Salary.Currency),
		Status:            MatchStatusMismatch,
		CandidateEvidence: fmt.Sprintf("Job max budget %.0f is below candidate minimum %.0f (%.1f%% shortfall)", jobMax, candidateMin, shortfallPercent),
		Confidence:        1.0,
	})
	result.Explanations = append(result.Explanations, fmt.Sprintf(
		"Compensation Match: Disclosed budget (%s %.0f) is %.1f%% below candidate minimum expectation of %s %.0f (Score: %.1f%%).",
		job.Compensation.Currency,
		job.Compensation.MaxAmount,
		shortfallPercent,
		preferences.Salary.Currency,
		candidateMin,
		score,
	))
	return score
}

// calculateTotalYearsExperience sums years of professional experience across items.
func calculateTotalYearsExperience(experiences []ExperienceItem) int {
	if len(experiences) == 0 {
		return 0
	}
	totalMonths := 0
	for _, exp := range experiences {
		if exp.StartDate == nil {
			continue
		}
		endTime := time.Now()
		if !exp.IsCurrent && exp.EndDate != nil {
			endTime = *exp.EndDate
		}
		months := int(endTime.Sub(*exp.StartDate).Hours() / (24 * 30.4))
		if months > 0 {
			totalMonths += months
		}
	}
	years := totalMonths / 12
	if years == 0 && totalMonths > 0 {
		return 1
	}
	return years
}

// Request and Response HTTP DTOs
type EvaluateJobMatchRequest struct {
	Job           DiscoveredJob   `json:"job"`
	CustomWeights *ScoringWeights `json:"custom_weights,omitempty"`
}

type EvaluateBatchJobMatchesRequest struct {
	Jobs          []DiscoveredJob `json:"jobs"`
	CustomWeights *ScoringWeights `json:"custom_weights,omitempty"`
}

type EvaluateBatchJobMatchesResponse struct {
	Matches     []JobMatchResult `json:"matches"`
	Algorithm   string           `json:"algorithm"`
	WeightsUsed ScoringWeights   `json:"weights_used"`
}
