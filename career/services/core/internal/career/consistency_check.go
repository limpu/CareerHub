package career

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrMismatchNotFound        = errors.New("consistency mismatch not found in report")
	ErrInvalidResolutionAction = errors.New("invalid resolution action; must be retain_resume, retain_linkedin, retain_profile, or custom_override (AT-007)")
	ErrMissingChosenValue      = errors.New("chosen value cannot be empty upon resolution")
	ErrReportNotFound          = errors.New("consistency audit report not found")
	ErrLinkedInSnapshotNotFound = errors.New("linkedin profile snapshot not found")
)

// FactSource identifies the origin of a career fact (CAR-24, REQ-004).
type FactSource string

const (
	SourceMasterResume  FactSource = "master_resume"
	SourceMasterProfile FactSource = "master_profile"
	SourceLinkedIn      FactSource = "linkedin_profile"
)

// MismatchedFieldType defines the specific career attribute exhibiting a discrepancy.
type MismatchedFieldType string

const (
	FieldJobTitle       MismatchedFieldType = "job_title"
	FieldEmploymentDate MismatchedFieldType = "employment_date"
	FieldCompanyName    MismatchedFieldType = "company_name"
	FieldSkillCoverage  MismatchedFieldType = "skill_coverage"
	FieldEducation      MismatchedFieldType = "education"
)

// MismatchSeverity indicates the gravity of the fact discrepancy.
type MismatchSeverity string

const (
	SeverityHigh   MismatchSeverity = "high"   // Date overlap, conflicting company names, significantly divergent titles
	SeverityMedium MismatchSeverity = "medium" // Title seniority nuance, minor date discrepancies (1-2 months)
	SeverityLow    MismatchSeverity = "low"    // Skill coverage gaps (present on LinkedIn but omitted on resume)
)

// FactSourceValue captures a concrete fact observed from a specific source.
type FactSourceValue struct {
	Source     FactSource `json:"source"`
	Value      string     `json:"value"`
	Context    string     `json:"context,omitempty"`
	ObservedAt time.Time  `json:"observed_at"`
}

// ResolutionAction defines the candidate's chosen reconciliation decision (AT-007).
type ResolutionAction string

const (
	ResolveRetainResume   ResolutionAction = "retain_resume"
	ResolveRetainLinkedIn ResolutionAction = "retain_linkedin"
	ResolveRetainProfile  ResolutionAction = "retain_profile"
	ResolveCustomOverride ResolutionAction = "custom_override"
)

// ResolutionDecision records the candidate's explicit human resolution gate decision (AT-007, CAR-24).
type ResolutionDecision struct {
	Action            ResolutionAction `json:"action"`
	ChosenValue       string           `json:"chosen_value"`
	ResolvedAt        time.Time        `json:"resolved_at"`
	UserJustification string           `json:"user_justification"`
}

// ConsistencyMismatch describes an individual detected fact discrepancy.
type ConsistencyMismatch struct {
	MismatchID   string              `json:"mismatch_id"`
	FieldType    MismatchedFieldType `json:"field_type"`
	Severity     MismatchSeverity    `json:"severity"`
	EntityKey    string              `json:"entity_key"` // e.g. "experience:FastFintech", "skill:Rust"
	Description  string              `json:"description"`
	SourceValues []FactSourceValue   `json:"source_values"`
	Resolution   *ResolutionDecision `json:"resolution,omitempty"`
}

// LinkedInProfileSnapshot stores an imported LinkedIn profile snapshot (REQ-004).
type LinkedInProfileSnapshot struct {
	SnapshotID   string           `json:"snapshot_id"`
	UserID       string           `json:"user_id"`
	Headline     string           `json:"headline"`
	Summary      string           `json:"summary,omitempty"`
	Experiences  []ExperienceItem `json:"experiences"`
	Skills       []string         `json:"skills"`
	Education    []EducationItem  `json:"education,omitempty"`
	ImportedAt   time.Time        `json:"imported_at"`
	SourceMode   string           `json:"source_mode"` // "user_export", "approved_api"
}

// ConsistencyAuditReport compiles all cross-source discrepancies and health scores.
type ConsistencyAuditReport struct {
	ReportID                 string                `json:"report_id"`
	UserID                   string                `json:"user_id"`
	Mismatches               []ConsistencyMismatch `json:"mismatches"`
	TotalMismatches          int                   `json:"total_mismatches"`
	HighSeverityCount        int                   `json:"high_severity_count"`
	MediumSeverityCount      int                   `json:"medium_severity_count"`
	LowSeverityCount         int                   `json:"low_severity_count"`
	ResolvedCount            int                   `json:"resolved_count"`
	OverallConsistencyScore  float64               `json:"overall_consistency_score"` // 0.0 to 100.0
	GeneratedAt              time.Time             `json:"generated_at"`
}

// CompareCareerRecords cross-references facts across profile, resume, and LinkedIn snapshot (CAR-24).
func CompareCareerRecords(userID string, profile *MasterCareerProfile, resume *GeneratedResume, linkedIn *LinkedInProfileSnapshot) *ConsistencyAuditReport {
	var mismatches []ConsistencyMismatch
	now := time.Now().UTC()

	// Map company experiences across sources
	type sourceExp struct {
		title     string
		startDate string
		endDate   string
		current   bool
	}

	profileExps := make(map[string]sourceExp)
	if profile != nil {
		for _, e := range profile.Experiences {
			key := strings.ToLower(strings.TrimSpace(e.Company))
			startStr := ""
			if e.StartDate != nil {
				startStr = e.StartDate.Format("2006-01")
			}
			endStr := "Present"
			if !e.IsCurrent && e.EndDate != nil {
				endStr = e.EndDate.Format("2006-01")
			}
			profileExps[key] = sourceExp{
				title:     e.Title,
				startDate: startStr,
				endDate:   endStr,
				current:   e.IsCurrent,
			}
		}
	}

	linkedInExps := make(map[string]sourceExp)
	if linkedIn != nil {
		for _, e := range linkedIn.Experiences {
			key := strings.ToLower(strings.TrimSpace(e.Company))
			startStr := ""
			if e.StartDate != nil {
				startStr = e.StartDate.Format("2006-01")
			}
			endStr := "Present"
			if !e.IsCurrent && e.EndDate != nil {
				endStr = e.EndDate.Format("2006-01")
			}
			linkedInExps[key] = sourceExp{
				title:     e.Title,
				startDate: startStr,
				endDate:   endStr,
				current:   e.IsCurrent,
			}
		}
	}

	// 1. Compare Job Titles & Dates per Company
	allCompanies := make(map[string]bool)
	for c := range profileExps {
		allCompanies[c] = true
	}
	for c := range linkedInExps {
		allCompanies[c] = true
	}

	for comp := range allCompanies {
		pExp, hasP := profileExps[comp]
		liExp, hasLI := linkedInExps[comp]

		if hasP && hasLI {
			// Title Check
			if strings.TrimSpace(strings.ToLower(pExp.title)) != strings.TrimSpace(strings.ToLower(liExp.title)) {
				mismatches = append(mismatches, ConsistencyMismatch{
					MismatchID:  fmt.Sprintf("mis-title-%s", comp),
					FieldType:   FieldJobTitle,
					Severity:    SeverityHigh,
					EntityKey:   fmt.Sprintf("experience:%s", comp),
					Description: fmt.Sprintf("Mismatched job title at %s between Master Profile and LinkedIn", comp),
					SourceValues: []FactSourceValue{
						{Source: SourceMasterProfile, Value: pExp.title, Context: "Profile Experience", ObservedAt: now},
						{Source: SourceLinkedIn, Value: liExp.title, Context: "LinkedIn Snapshot Experience", ObservedAt: now},
					},
				})
			}

			// Dates Check
			if pExp.startDate != liExp.startDate || pExp.endDate != liExp.endDate || pExp.current != liExp.current {
				pDates := fmt.Sprintf("%s - %s", pExp.startDate, pExp.endDate)
				liDates := fmt.Sprintf("%s - %s", liExp.startDate, liExp.endDate)
				mismatches = append(mismatches, ConsistencyMismatch{
					MismatchID:  fmt.Sprintf("mis-date-%s", comp),
					FieldType:   FieldEmploymentDate,
					Severity:    SeverityHigh,
					EntityKey:   fmt.Sprintf("experience:%s", comp),
					Description: fmt.Sprintf("Employment date discrepancy at %s between Master Profile and LinkedIn", comp),
					SourceValues: []FactSourceValue{
						{Source: SourceMasterProfile, Value: pDates, Context: "Profile Start/End Dates", ObservedAt: now},
						{Source: SourceLinkedIn, Value: liDates, Context: "LinkedIn Start/End Dates", ObservedAt: now},
					},
				})
			}
		}
	}

	// 2. Compare Skills Coverage
	profileSkills := make(map[string]bool)
	if profile != nil {
		for _, s := range profile.Skills {
			profileSkills[strings.ToLower(strings.TrimSpace(s.Name))] = true
		}
	}

	if linkedIn != nil {
		for _, liSkill := range linkedIn.Skills {
			cleanSkill := strings.TrimSpace(liSkill)
			if cleanSkill == "" {
				continue
			}
			if !profileSkills[strings.ToLower(cleanSkill)] {
				mismatches = append(mismatches, ConsistencyMismatch{
					MismatchID:  fmt.Sprintf("mis-skill-%s", strings.ToLower(cleanSkill)),
					FieldType:   FieldSkillCoverage,
					Severity:    SeverityLow,
					EntityKey:   fmt.Sprintf("skill:%s", cleanSkill),
					Description: fmt.Sprintf("Skill '%s' listed on LinkedIn but omitted from Master Profile/Resume", cleanSkill),
					SourceValues: []FactSourceValue{
						{Source: SourceLinkedIn, Value: cleanSkill, Context: "LinkedIn Skills Section", ObservedAt: now},
						{Source: SourceMasterProfile, Value: "[Not Listed]", Context: "Master Profile Skills", ObservedAt: now},
					},
				})
			}
		}
	}

	// Calculate counts and overall consistency score
	highCount := 0
	medCount := 0
	lowCount := 0

	for _, m := range mismatches {
		switch m.Severity {
		case SeverityHigh:
			highCount++
		case SeverityMedium:
			medCount++
		case SeverityLow:
			lowCount++
		}
	}

	// Score formula: starts at 100%, penalties: High -20%, Med -10%, Low -5%
	penalty := float64(highCount)*20.0 + float64(medCount)*10.0 + float64(lowCount)*5.0
	score := 100.0 - penalty
	if score < 0.0 {
		score = 0.0
	}

	return &ConsistencyAuditReport{
		ReportID:                fmt.Sprintf("rep-%d", time.Now().UnixNano()),
		UserID:                  userID,
		Mismatches:              mismatches,
		TotalMismatches:         len(mismatches),
		HighSeverityCount:       highCount,
		MediumSeverityCount:     medCount,
		LowSeverityCount:        lowCount,
		ResolvedCount:           0,
		OverallConsistencyScore: score,
		GeneratedAt:             now,
	}
}

// ApplyConsistencyResolution applies candidate decision to the mismatch and updates canonical MasterCareerProfile (AT-007).
func ApplyConsistencyResolution(report *ConsistencyAuditReport, mismatchID string, decision ResolutionDecision, profile *MasterCareerProfile) (*ConsistencyAuditReport, error) {
	if report == nil {
		return nil, ErrReportNotFound
	}
	if strings.TrimSpace(decision.ChosenValue) == "" {
		return nil, ErrMissingChosenValue
	}

	switch decision.Action {
	case ResolveRetainResume, ResolveRetainLinkedIn, ResolveRetainProfile, ResolveCustomOverride:
		// Valid
	default:
		return nil, ErrInvalidResolutionAction
	}

	found := false
	for i := range report.Mismatches {
		m := &report.Mismatches[i]
		if m.MismatchID == mismatchID {
			decision.ResolvedAt = time.Now().UTC()
			m.Resolution = &decision
			found = true

			// Reconcile Canonical MasterCareerProfile (AT-007)
			if profile != nil {
				now := time.Now().UTC()
				if m.FieldType == FieldJobTitle && strings.HasPrefix(m.EntityKey, "experience:") {
					compName := strings.TrimPrefix(m.EntityKey, "experience:")
					for j := range profile.Experiences {
						if strings.EqualFold(profile.Experiences[j].Company, compName) {
							oldTitle := profile.Experiences[j].Title
							profile.Experiences[j].Title = decision.ChosenValue
							profile.AuditTrail = append(profile.AuditTrail, ProfileFieldAudit{
								ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
								Field:     fmt.Sprintf("experiences[%d].title", j),
								OldValue:  oldTitle,
								NewValue:  decision.ChosenValue,
								ChangedBy: report.UserID,
								ChangedAt: now,
								Reason:    decision.UserJustification,
							})
							break
						}
					}
				} else if m.FieldType == FieldSkillCoverage && strings.HasPrefix(m.EntityKey, "skill:") {
					skillName := strings.TrimPrefix(m.EntityKey, "skill:")
					profile.Skills = append(profile.Skills, SkillItem{
						ID:          fmt.Sprintf("skill-%d", time.Now().UnixNano()),
						Name:        decision.ChosenValue,
						Category:    "technical",
						Confirmed:   true,
					})
					profile.AuditTrail = append(profile.AuditTrail, ProfileFieldAudit{
						ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
						Field:     "skills",
						OldValue:  "",
						NewValue:  decision.ChosenValue,
						ChangedBy: report.UserID,
						ChangedAt: now,
						Reason:    fmt.Sprintf("Added skill via consistency check: %s", skillName),
					})
				}
			}
			break
		}
	}

	if !found {
		return nil, ErrMismatchNotFound
	}

	// Recalculate resolved count and updated consistency score
	resolved := 0
	unresolvedHigh := 0
	unresolvedMed := 0
	unresolvedLow := 0

	for _, m := range report.Mismatches {
		if m.Resolution != nil {
			resolved++
		} else {
			switch m.Severity {
			case SeverityHigh:
				unresolvedHigh++
			case SeverityMedium:
				unresolvedMed++
			case SeverityLow:
				unresolvedLow++
			}
		}
	}

	report.ResolvedCount = resolved
	penalty := float64(unresolvedHigh)*20.0 + float64(unresolvedMed)*10.0 + float64(unresolvedLow)*5.0
	score := 100.0 - penalty
	if score < 0.0 {
		score = 0.0
	}
	report.OverallConsistencyScore = score

	return report, nil
}
