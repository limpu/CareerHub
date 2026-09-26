package career

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

var (
	ErrSavedJobNotFound     = errors.New("saved job not found")
	ErrExclusionNotFound    = errors.New("job exclusion not found")
	ErrDuplicateSavedJob    = errors.New("job is already in saved list")
	ErrDuplicateApplication = errors.New("application already recorded for this job")
	ErrInvalidExclusionType = errors.New("invalid exclusion type")
	ErrEmptyExclusionValue  = errors.New("exclusion value cannot be empty")
)

// SavedJobStatus defines the candidate's personal progression state for an opportunity (CAR-10).
type SavedJobStatus string

const (
	SavedJobStatusSaved        SavedJobStatus = "saved"
	SavedJobStatusShortlisted  SavedJobStatus = "shortlisted"
	SavedJobStatusReadyToApply SavedJobStatus = "ready_to_apply"
	SavedJobStatusApplied      SavedJobStatus = "applied"
	SavedJobStatusArchived     SavedJobStatus = "archived"
)

// SavedJob represents a candidate's saved or shortlisted job record (CAR-10).
type SavedJob struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	JobID          string                 `json:"job_id"`
	CanonicalURL   string                 `json:"canonical_url"`
	Fingerprint    string                 `json:"fingerprint"`
	Title          string                 `json:"title"`
	Company        string                 `json:"company"`
	Location       JobLocation            `json:"location"`
	JobType        string                 `json:"job_type"`
	Compensation   NormalizedCompensation `json:"compensation"`
	RequiredSkills []string               `json:"required_skills"`
	DirectApplyURL string                 `json:"direct_apply_url"`
	Status         SavedJobStatus         `json:"status"`
	Notes          string                 `json:"notes"`
	Priority       int                    `json:"priority"` // 1 (lowest) to 5 (highest)
	Tags           []string               `json:"tags"`
	SavedAt        time.Time              `json:"saved_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	AppliedAt      *time.Time             `json:"applied_at,omitempty"`
}

// ExclusionType defines the target dimension for persistent exclusion (CAR-10, C8).
type ExclusionType string

const (
	ExclusionTypeJobCanonicalURL ExclusionType = "job_canonical_url"
	ExclusionTypeJobFingerprint  ExclusionType = "job_fingerprint"
	ExclusionTypeCompanyName     ExclusionType = "company_name"
)

// JobExclusion represents a permanent "never see this job/company again" record (CAR-10, C8).
type JobExclusion struct {
	ID          string        `json:"id"`
	UserID      string        `json:"user_id"`
	Type        ExclusionType `json:"type"`
	Value       string        `json:"value"` // canonical URL, normalized fingerprint, or company name
	Reason      string        `json:"reason"`
	JobTitle    string        `json:"job_title,omitempty"`
	CompanyName string        `json:"company_name,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

// ApplicationRecord tracks historical job applications separately from discoveries (CAR-10, CAR-15, AT-004, AT-005, AT-006).
type ApplicationRecord struct {
	ID                  string                      `json:"id"`
	UserID              string                      `json:"user_id"`
	JobID               string                      `json:"job_id"`
	CanonicalURL        string                      `json:"canonical_url"`
	Fingerprint         string                      `json:"fingerprint"`
	Title               string                      `json:"title"`
	Company             string                      `json:"company"`
	Location            string                      `json:"location,omitempty"`
	PortalType          string                      `json:"portal_type,omitempty"`
	Status              string                      `json:"status"`          // "submitted", "in_review", "interviewing", "rejected", "offered"
	Stage               ApplicationStage            `json:"stage"`           // Detailed stage from ApplicationStage
	SubmissionMode      string                      `json:"submission_mode"` // "manual_browser", "assisted", "direct_ats"
	IsVerifiedApplied   bool                        `json:"is_verified_applied"` // REQ-005, REQ-016
	AppliedAt           time.Time                   `json:"applied_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
	Notes               string                      `json:"notes"`
	VerificationDetails *AppliedVerificationDetails `json:"verification_details,omitempty"`
	InterviewRounds     []InterviewRound            `json:"interview_rounds,omitempty"`
	Timeline            []ApplicationTimelineEvent  `json:"timeline,omitempty"`
}

// DedupeMatchLevel classifies the degree of match found by the deduplication engine.
type DedupeMatchLevel string

const (
	DedupeMatchExact     DedupeMatchLevel = "exact"
	DedupeMatchAmbiguous DedupeMatchLevel = "ambiguous" // Reviewable match per AT-004
	DedupeMatchNone      DedupeMatchLevel = "none"
)

// DedupeMatch represents an identified duplicate or ambiguous match.
type DedupeMatch struct {
	MatchLevel      DedupeMatchLevel `json:"match_level"`
	TargetLedger    string           `json:"target_ledger"` // "saved_jobs", "applications", "exclusions", "discovery"
	ExistingID      string           `json:"existing_id"`
	ExistingTitle   string           `json:"existing_title"`
	ExistingCompany string           `json:"existing_company"`
	SimilarityScore float64          `json:"similarity_score"` // 0.0 to 1.0
	Reason          string           `json:"reason"`
	IsReviewable    bool             `json:"is_reviewable"` // True if ambiguous match requires human review (AT-004)
}

// JobDedupeResult is the comprehensive evaluation of a discovered job against historical ledgers.
type JobDedupeResult struct {
	JobID        string        `json:"job_id"`
	CanonicalURL string        `json:"canonical_url"`
	Fingerprint  string        `json:"fingerprint"`
	Matches      []DedupeMatch `json:"matches"`
	IsDuplicate  bool          `json:"is_duplicate"`
	IsAmbiguous  bool          `json:"is_ambiguous"` // AT-004
	IsExcluded   bool          `json:"is_excluded"`
	ActionTaken  string        `json:"action_taken"` // "retained", "linked", "flagged_for_review", "filtered_out"
}

// SaveJobInput contains fields to create or update a saved job.
type SaveJobInput struct {
	Job      DiscoveredJob  `json:"job"`
	Status   SavedJobStatus `json:"status"`
	Notes    string         `json:"notes"`
	Priority int            `json:"priority"`
	Tags     []string       `json:"tags"`
}

// UpdateSavedJobInput contains fields to update status/notes/priority on a saved job.
type UpdateSavedJobInput struct {
	Status    *SavedJobStatus `json:"status,omitempty"`
	Notes     *string         `json:"notes,omitempty"`
	Priority  *int            `json:"priority,omitempty"`
	Tags      []string        `json:"tags,omitempty"`
	AppliedAt *time.Time      `json:"applied_at,omitempty"`
}

// AddExclusionInput contains parameters to permanently exclude a job or company.
type AddExclusionInput struct {
	Type        ExclusionType `json:"type"`
	Value       string        `json:"value"`
	Reason      string        `json:"reason"`
	JobTitle    string        `json:"job_title,omitempty"`
	CompanyName string        `json:"company_name,omitempty"`
}

// RecordApplicationInput records an application in the separate application ledger.
type RecordApplicationInput struct {
	Job            DiscoveredJob `json:"job"`
	Status         string        `json:"status"`
	SubmissionMode string        `json:"submission_mode"`
	Notes          string        `json:"notes"`
	AppliedAt      *time.Time    `json:"applied_at,omitempty"`
}

// DedupeEngine executes multi-ledger deduplication and text similarity thresholding (CAR-10, AT-004).
type DedupeEngine struct {
	// Similarity threshold for ambiguous title/description matching (default: 0.75 per C8/JobFunnel)
	ambiguousThreshold float64
}

// NewDedupeEngine constructs a deduplication engine.
func NewDedupeEngine() *DedupeEngine {
	return &DedupeEngine{
		ambiguousThreshold: 0.75,
	}
}

// ComputeFingerprint generates a stable cryptographic hash based on normalized company + title + location.
func (e *DedupeEngine) ComputeFingerprint(company, title, location string) string {
	normCompany := strings.ToLower(strings.TrimSpace(company))
	normTitle := strings.ToLower(strings.TrimSpace(title))
	normLoc := strings.ToLower(strings.TrimSpace(location))

	// Remove common punctuation and extraneous spaces
	clean := func(s string) string {
		var b strings.Builder
		for _, r := range s {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
			}
		}
		return b.String()
	}

	payload := fmt.Sprintf("%s|%s|%s", clean(normCompany), clean(normTitle), clean(normLoc))
	h := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(h[:16]) // 32-char hex
}

// Tokenize converts text into a set of lowercased word tokens.
func (e *DedupeEngine) Tokenize(text string) map[string]struct{} {
	tokens := make(map[string]struct{})
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	for _, w := range words {
		if len(w) > 1 {
			tokens[w] = struct{}{}
		}
	}
	return tokens
}

// ComputeSimilarity calculates Jaccard token overlap similarity between two text strings (0.0 to 1.0).
func (e *DedupeEngine) ComputeSimilarity(text1, text2 string) float64 {
	tokens1 := e.Tokenize(text1)
	tokens2 := e.Tokenize(text2)

	if len(tokens1) == 0 && len(tokens2) == 0 {
		return 1.0
	}
	if len(tokens1) == 0 || len(tokens2) == 0 {
		return 0.0
	}

	intersection := 0
	for t := range tokens1 {
		if _, ok := tokens2[t]; ok {
			intersection++
		}
	}

	union := len(tokens1) + len(tokens2) - intersection
	if union == 0 {
		return 0.0
	}

	score := float64(intersection) / float64(union)
	return math.Round(score*100) / 100
}

// EvaluateJobDeduplication checks a discovered job against saved jobs, applications, and exclusions (CAR-10, AT-004).
func (e *DedupeEngine) EvaluateJobDeduplication(
	job DiscoveredJob,
	savedJobs []SavedJob,
	applications []ApplicationRecord,
	exclusions []JobExclusion,
) JobDedupeResult {
	canonicalURL := NormalizeCanonicalURL(job.CanonicalURL)
	fingerprint := e.ComputeFingerprint(job.Company, job.Title, job.Location.RawLocation)

	result := JobDedupeResult{
		JobID:        job.ID,
		CanonicalURL: canonicalURL,
		Fingerprint:  fingerprint,
		Matches:      make([]DedupeMatch, 0),
		ActionTaken:  "retained",
	}

	// 1. Check Persistent Exclusions Ledger (CAR-10)
	for _, ex := range exclusions {
		switch ex.Type {
		case ExclusionTypeCompanyName:
			if strings.EqualFold(job.Company, ex.Value) || strings.Contains(strings.ToLower(job.Company), strings.ToLower(ex.Value)) {
				result.IsExcluded = true
				result.Matches = append(result.Matches, DedupeMatch{
					MatchLevel:      DedupeMatchExact,
					TargetLedger:    "exclusions",
					ExistingID:      ex.ID,
					ExistingCompany: ex.CompanyName,
					SimilarityScore: 1.0,
					Reason:          fmt.Sprintf("Excluded: Company matches blocked entity '%s' (%s)", ex.Value, ex.Reason),
					IsReviewable:    false,
				})
			}
		case ExclusionTypeJobCanonicalURL:
			if canonicalURL != "" && strings.EqualFold(canonicalURL, NormalizeCanonicalURL(ex.Value)) {
				result.IsExcluded = true
				result.Matches = append(result.Matches, DedupeMatch{
					MatchLevel:      DedupeMatchExact,
					TargetLedger:    "exclusions",
					ExistingID:      ex.ID,
					ExistingTitle:   ex.JobTitle,
					ExistingCompany: ex.CompanyName,
					SimilarityScore: 1.0,
					Reason:          fmt.Sprintf("Excluded: Job canonical URL matches blocked posting (%s)", ex.Reason),
					IsReviewable:    false,
				})
			}
		case ExclusionTypeJobFingerprint:
			if fingerprint != "" && strings.EqualFold(fingerprint, ex.Value) {
				result.IsExcluded = true
				result.Matches = append(result.Matches, DedupeMatch{
					MatchLevel:      DedupeMatchExact,
					TargetLedger:    "exclusions",
					ExistingID:      ex.ID,
					ExistingTitle:   ex.JobTitle,
					ExistingCompany: ex.CompanyName,
					SimilarityScore: 1.0,
					Reason:          fmt.Sprintf("Excluded: Job fingerprint matches blocked posting (%s)", ex.Reason),
					IsReviewable:    false,
				})
			}
		}
	}

	// If excluded, mark action and return early
	if result.IsExcluded {
		result.ActionTaken = "filtered_out"
		return result
	}

	// 2. Check Applications Ledger (CAR-10: separate from discovery)
	for _, app := range applications {
		// Exact URL match
		if canonicalURL != "" && strings.EqualFold(canonicalURL, NormalizeCanonicalURL(app.CanonicalURL)) {
			result.IsDuplicate = true
			result.Matches = append(result.Matches, DedupeMatch{
				MatchLevel:      DedupeMatchExact,
				TargetLedger:    "applications",
				ExistingID:      app.ID,
				ExistingTitle:   app.Title,
				ExistingCompany: app.Company,
				SimilarityScore: 1.0,
				Reason:          fmt.Sprintf("Already Applied: Exact canonical URL matches previous application submitted on %s", app.AppliedAt.Format("2006-01-02")),
				IsReviewable:    false,
			})
			break
		}

		// Exact Fingerprint match
		if fingerprint == app.Fingerprint {
			result.IsDuplicate = true
			result.Matches = append(result.Matches, DedupeMatch{
				MatchLevel:      DedupeMatchExact,
				TargetLedger:    "applications",
				ExistingID:      app.ID,
				ExistingTitle:   app.Title,
				ExistingCompany: app.Company,
				SimilarityScore: 1.0,
				Reason:          fmt.Sprintf("Already Applied: Job fingerprint matches previous application at '%s'", app.Company),
				IsReviewable:    false,
			})
			break
		}

		// Fuzzy / Ambiguous Match with previous application (AT-004)
		if strings.EqualFold(job.Company, app.Company) {
			sim := e.ComputeSimilarity(job.Title, app.Title)
			if sim >= e.ambiguousThreshold {
				result.IsAmbiguous = true
				result.Matches = append(result.Matches, DedupeMatch{
					MatchLevel:      DedupeMatchAmbiguous,
					TargetLedger:    "applications",
					ExistingID:      app.ID,
					ExistingTitle:   app.Title,
					ExistingCompany: app.Company,
					SimilarityScore: sim,
					Reason:          fmt.Sprintf("Ambiguous Application Match: Candidate previously applied to '%s' at '%s' (similarity: %.0f%%). Review to prevent redundant submission.", app.Title, app.Company, sim*100),
					IsReviewable:    true,
				})
			}
		}
	}

	// 3. Check Saved Jobs Ledger (CAR-10)
	for _, saved := range savedJobs {
		// Exact URL match
		if canonicalURL != "" && strings.EqualFold(canonicalURL, NormalizeCanonicalURL(saved.CanonicalURL)) {
			result.IsDuplicate = true
			result.Matches = append(result.Matches, DedupeMatch{
				MatchLevel:      DedupeMatchExact,
				TargetLedger:    "saved_jobs",
				ExistingID:      saved.ID,
				ExistingTitle:   saved.Title,
				ExistingCompany: saved.Company,
				SimilarityScore: 1.0,
				Reason:          fmt.Sprintf("Already Saved: Exact canonical URL matches saved shortlist item (status: %s)", saved.Status),
				IsReviewable:    false,
			})
			break
		}

		// Exact Fingerprint match
		if fingerprint == saved.Fingerprint {
			result.IsDuplicate = true
			result.Matches = append(result.Matches, DedupeMatch{
				MatchLevel:      DedupeMatchExact,
				TargetLedger:    "saved_jobs",
				ExistingID:      saved.ID,
				ExistingTitle:   saved.Title,
				ExistingCompany: saved.Company,
				SimilarityScore: 1.0,
				Reason:          fmt.Sprintf("Already Saved: Fingerprint matches existing saved job at '%s'", saved.Company),
				IsReviewable:    false,
			})
			break
		}

		// Fuzzy / Ambiguous Match (AT-004)
		if strings.EqualFold(job.Company, saved.Company) {
			sim := e.ComputeSimilarity(job.Title, saved.Title)
			if sim >= e.ambiguousThreshold {
				result.IsAmbiguous = true
				result.Matches = append(result.Matches, DedupeMatch{
					MatchLevel:      DedupeMatchAmbiguous,
					TargetLedger:    "saved_jobs",
					ExistingID:      saved.ID,
					ExistingTitle:   saved.Title,
					ExistingCompany: saved.Company,
					SimilarityScore: sim,
					Reason:          fmt.Sprintf("Ambiguous Saved Match: Similar job '%s' at '%s' is already in saved shortlist (similarity: %.0f%%). Linkable or reviewable.", saved.Title, saved.Company, sim*100),
					IsReviewable:    true,
				})
			}
		}
	}

	// Determine final action taken
	if result.IsDuplicate {
		result.ActionTaken = "linked"
	} else if result.IsAmbiguous {
		result.ActionTaken = "flagged_for_review"
	} else {
		result.ActionTaken = "retained"
	}

	return result
}
