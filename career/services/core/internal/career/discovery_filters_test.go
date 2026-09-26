package career

import (
	"context"
	"testing"
	"time"
)

func createSampleJobsForFilterTests(baseTime time.Time) []DiscoveredJob {
	return []DiscoveredJob{
		{
			ID:           "job-1-fresh-remote",
			Title:        "Senior Go Architect",
			Company:      "TechCorp",
			CanonicalURL: "https://boards.greenhouse.io/techcorp/jobs/1",
			Location: JobLocation{
				RawLocation: "Remote",
				IsRemote:    true,
			},
			JobType: "Full-Time",
			Compensation: NormalizedCompensation{
				IsDisclosed: true,
				MinAmount:   190000,
				MaxAmount:   230000,
				Currency:    "USD",
				Period:      PeriodYearly,
			},
			DatePosted: baseTime.AddDate(0, 0, -2), // 2 days old
		},
		{
			ID:           "job-2-old-onsite",
			Title:        "Staff Infrastructure Engineer",
			Company:      "MegaBank",
			CanonicalURL: "https://boards.greenhouse.io/megabank/jobs/2",
			Location: JobLocation{
				RawLocation: "New York, NY",
				City:        "New York",
				IsRemote:    false,
			},
			JobType: "Vollzeit", // German for full-time
			Compensation: NormalizedCompensation{
				IsDisclosed: true,
				MinAmount:   150000,
				MaxAmount:   170000,
				Currency:    "USD",
				Period:      PeriodYearly,
			},
			DatePosted: baseTime.AddDate(0, 0, -18), // 18 days old
		},
		{
			ID:           "job-3-contract-hybrid",
			Title:        "Kubernetes Cloud Consultant",
			Company:      "CloudScale Partners",
			CanonicalURL: "https://jobs.lever.co/cloudscale/3",
			Location: JobLocation{
				RawLocation: "Berlin, Germany (Hybrid)",
				IsHybrid:    true,
			},
			JobType: "Contractor",
			Compensation: NormalizedCompensation{
				IsDisclosed: false, // Undisclosed salary per AT-003
			},
			DatePosted: baseTime.AddDate(0, 0, -5), // 5 days old
		},
		{
			ID:           "job-4-blacklisted-agency",
			Title:        "Junior Software Developer",
			Company:      "Spam Staffing Agency Inc.",
			CanonicalURL: "https://jobs.ashbyhq.com/spam/4",
			Location: JobLocation{
				RawLocation: "Remote",
				IsRemote:    true,
			},
			JobType: "CDI", // French for permanent full-time
			Compensation: NormalizedCompensation{
				IsDisclosed: true,
				MinAmount:   80000,
				MaxAmount:   90000,
				Currency:    "USD",
				Period:      PeriodYearly,
			},
			DatePosted: baseTime.AddDate(0, 0, -1), // 1 day old
		},
	}
}

func TestJobFilter_AgeFiltering(t *testing.T) {
	now := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
	jobs := createSampleJobsForFilterTests(now)
	evaluator := NewFilterEvaluator()

	// Filter for max 7 days old
	filter := AdvancedJobFilter{
		MaxAgeDays:               7,
		IncludeUndisclosedSalary: true,
	}

	filtered, report := evaluator.Evaluate(jobs, filter, now)
	// job-2 (18 days old) must be excluded
	if len(filtered) != 3 {
		t.Fatalf("expected 3 jobs within 7 days, got %d", len(filtered))
	}
	if report.DisqualifiedByAge != 1 {
		t.Errorf("expected 1 disqualified by age, got %d", report.DisqualifiedByAge)
	}

	for _, j := range filtered {
		if j.ID == "job-2-old-onsite" {
			t.Errorf("job-2 should have been filtered out due to age")
		}
	}
}

func TestJobFilter_MultiLanguageJobTypeParser(t *testing.T) {
	testCases := []struct {
		raw      string
		expected JobType
	}{
		{"Full-Time", JobTypeFullTime},
		{"Permanent", JobTypeFullTime},
		{"Vollzeit", JobTypeFullTime},
		{"Festanstellung", JobTypeFullTime},
		{"Temps plein", JobTypeFullTime},
		{"CDI", JobTypeFullTime},
		{"Tiempo Completo", JobTypeFullTime},
		{"Part Time", JobTypePartTime},
		{"Teilzeit", JobTypePartTime},
		{"Temps partiel", JobTypePartTime},
		{"Contract", JobTypeContract},
		{"Contractor", JobTypeContract},
		{"Befristet", JobTypeContract},
		{"CDD", JobTypeContract},
		{"Internship", JobTypeInternship},
		{"Praktikum", JobTypeInternship},
		{"Stage", JobTypeInternship},
		{"Pasantia", JobTypeInternship},
		{"Unknown Garbage Type", JobTypeOther},
	}

	for _, tc := range testCases {
		got := ParseStandardJobType(tc.raw)
		if got != tc.expected {
			t.Errorf("ParseStandardJobType(%q) = %s; want %s", tc.raw, got, tc.expected)
		}
	}
}

func TestJobFilter_CompanyInclusionsExclusions_CAR08(t *testing.T) {
	now := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
	jobs := createSampleJobsForFilterTests(now)
	evaluator := NewFilterEvaluator()

	// Exclude agency
	filter := AdvancedJobFilter{
		CompanyExclusions:        []string{"Spam Staffing Agency", "OtherBlacklisted"},
		IncludeUndisclosedSalary: true,
	}

	filtered, report := evaluator.Evaluate(jobs, filter, now)
	if len(filtered) != 3 {
		t.Fatalf("expected 3 jobs after company exclusion, got %d", len(filtered))
	}
	if report.DisqualifiedByCompany != 1 {
		t.Errorf("expected 1 disqualified by company, got %d", report.DisqualifiedByCompany)
	}

	for _, j := range filtered {
		if j.Company == "Spam Staffing Agency Inc." {
			t.Errorf("blacklisted company was not filtered")
		}
	}
}

func TestJobFilter_SalaryThreshold_AT003_AT028(t *testing.T) {
	now := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
	jobs := createSampleJobsForFilterTests(now)
	evaluator := NewFilterEvaluator()

	// 1. High threshold with undisclosed salary PRESERVED (IncludeUndisclosedSalary=true per AT-003)
	filterPreserve := AdvancedJobFilter{
		MinimumAnnualSalary:      180000,
		IncludeUndisclosedSalary: true,
	}

	filteredPreserve, _ := evaluator.Evaluate(jobs, filterPreserve, now)
	// job-1 ($190k-$230k) and job-3 (undisclosed salary) should remain
	// job-2 ($150k-$170k) and job-4 ($80k-$90k) should be excluded
	if len(filteredPreserve) != 2 {
		t.Fatalf("expected 2 jobs with undisclosed preserved, got %d", len(filteredPreserve))
	}

	var foundUndisclosed bool
	for _, j := range filteredPreserve {
		if !j.Compensation.IsDisclosed {
			foundUndisclosed = true
		}
	}
	if !foundUndisclosed {
		t.Error("job with undisclosed salary should be preserved when IncludeUndisclosedSalary=true")
	}

	// 2. High threshold with undisclosed salary EXCLUDED (IncludeUndisclosedSalary=false)
	filterExclude := AdvancedJobFilter{
		MinimumAnnualSalary:      180000,
		IncludeUndisclosedSalary: false,
	}

	filteredExclude, reportExclude := evaluator.Evaluate(jobs, filterExclude, now)
	if len(filteredExclude) != 1 { // Only job-1 survives
		t.Fatalf("expected 1 job with undisclosed excluded, got %d", len(filteredExclude))
	}
	if filteredExclude[0].ID != "job-1-fresh-remote" {
		t.Errorf("expected job-1 to survive salary threshold, got %s", filteredExclude[0].ID)
	}
	if reportExclude.DisqualifiedBySalary != 3 {
		t.Errorf("expected 3 disqualified by salary, got %d", reportExclude.DisqualifiedBySalary)
	}
}

func TestJobFilter_FilterAuditTrail_CAR08_AT010(t *testing.T) {
	now := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
	jobs := createSampleJobsForFilterTests(now)
	evaluator := NewFilterEvaluator()

	filter := AdvancedJobFilter{
		MaxAgeDays:          7,
		WorkModes:           []string{"remote"},
		CompanyExclusions:   []string{"Spam Staffing Agency"},
		MinimumAnnualSalary: 100000,
	}

	filtered, report := evaluator.Evaluate(jobs, filter, now)
	if len(filtered) != 1 { // Only job-1 matches remote + under 7 days + not blacklisted + above 100k
		t.Fatalf("expected 1 retained job, got %d", len(filtered))
	}

	if report.TotalCandidatesInput != 4 {
		t.Errorf("expected TotalCandidatesInput 4, got %d", report.TotalCandidatesInput)
	}
	if report.TotalResultsRetained != 1 {
		t.Errorf("expected TotalResultsRetained 1, got %d", report.TotalResultsRetained)
	}

	// Verify transparent filter evaluation modes (AT-010)
	if report.FilterEvaluations["remote_filter"] != FilterModeUpstreamNative {
		t.Error("expected remote_filter to be evaluated as FilterModeUpstreamNative")
	}
	if report.FilterEvaluations["age_filter"] != FilterModePostFiltered {
		t.Error("expected age_filter to be evaluated as FilterModePostFiltered")
	}
	if report.FilterEvaluations["company_exclusion"] != FilterModePostFiltered {
		t.Error("expected company_exclusion to be evaluated as FilterModePostFiltered")
	}
}

func TestCareerService_FilterIntegration(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	now := time.Now().UTC()
	jobs := createSampleJobsForFilterTests(now)

	filtered, report := service.FilterDiscoveredJobs(ctx, jobs, AdvancedJobFilter{
		WorkModes:                []string{"remote"},
		IncludeUndisclosedSalary: true,
	})

	if len(filtered) != 2 { // job-1 and job-4 are remote
		t.Errorf("expected 2 remote jobs, got %d", len(filtered))
	}
	if report.TotalResultsRetained != len(filtered) {
		t.Errorf("report count mismatch: %d vs %d", report.TotalResultsRetained, len(filtered))
	}
}
