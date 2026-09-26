package career

import (
	"context"
	"testing"
	"time"
)

func sampleCandidateProfile() *MasterCareerProfile {
	now := time.Now()
	startDate1 := now.AddDate(-5, 0, 0)
	endDate1 := now.AddDate(-2, 0, 0)
	startDate2 := now.AddDate(-2, 0, 0)

	return &MasterCareerProfile{
		ID:     "prof-test-1",
		UserID: "user-test-1",
		Contact: ContactInfo{
			FullName: "Jane Doe",
			Email:    "jane@example.com",
			Location: "San Francisco, CA",
		},
		Experiences: []ExperienceItem{
			{
				ID:        "exp-1",
				Title:     "Backend Engineer",
				Company:   "TechCorp",
				StartDate: &startDate1,
				EndDate:   &endDate1,
				Confirmed: true,
			},
			{
				ID:        "exp-2",
				Title:     "Senior Distributed Systems Engineer",
				Company:   "CloudCorp",
				StartDate: &startDate2,
				IsCurrent: true,
				Confirmed: true,
			},
		},
		Skills: []SkillItem{
			{ID: "sk-1", Name: "Go", Proficiency: ProficiencyExpert, Confirmed: true},
			{ID: "sk-2", Name: "Kubernetes", Proficiency: ProficiencyAdvanced, Confirmed: true},
			{ID: "sk-3", Name: "PostgreSQL", Proficiency: ProficiencyAdvanced, Confirmed: true},
			{ID: "sk-4", Name: "Kafka", Proficiency: ProficiencyIntermediate, Confirmed: true},
		},
	}
}

func sampleCandidatePreferences() *CareerPreferences {
	return &CareerPreferences{
		ID:              "pref-test-1",
		UserID:          "user-test-1",
		TargetRoles:     []string{"Senior Go Engineer", "Distributed Systems Engineer", "Backend Platform Engineer"},
		TargetLocations: []string{"San Francisco, CA", "Remote"},
		WorkModes:       []WorkMode{WorkModeRemote, WorkModeHybrid},
		JobTypes:        []JobType{JobTypeFullTime},
		Salary: SalaryPreference{
			MinimumAmount:   160000,
			TargetAmount:    190000,
			Currency:        "USD",
			Interval:        SalaryIntervalAnnual,
			IsExplicitlySet: true,
		},
		Sponsorship:      SponsorshipAuthorizedNoSponsorship,
		OpenToRelocation: false,
		Exclusions: ExclusionRules{
			ExcludedCompanies: []string{"EvilCorp", "SpamStaffing"},
		},
	}
}

func sampleDiscoveredJob() *DiscoveredJob {
	return &DiscoveredJob{
		ID:            "job-test-1",
		Source:        "greenhouse",
		SourceJobID:   "gh-12345",
		CanonicalURL:  "https://boards.greenhouse.io/stripe/jobs/12345",
		Title:         "Senior Go Distributed Systems Engineer",
		Company:       "Stripe",
		Location: JobLocation{
			RawLocation: "San Francisco, CA / Remote",
			IsRemote:    true,
			IsHybrid:    false,
		},
		JobType:     "full_time",
		Description: "Build financial infrastructure in Go. Kubernetes and Kafka experience required.",
		RequiredSkills: []string{
			"Go",
			"Kubernetes",
			"Kafka",
			"Rust", // Missing from profile
		},
		Compensation: NormalizedCompensation{
			MinAmount:   180000,
			MaxAmount:   220000,
			Currency:    "USD",
			Period:      "yearly",
			IsDisclosed: true,
		},
		DatePosted:        time.Now().AddDate(0, 0, -2),
		DiscoveredAt:      time.Now(),
		IsDirectEmployer: true,
	}
}

func TestJobMatch_WeightsValidation(t *testing.T) {
	// Valid weights: sum = 100
	valid := ScoringWeights{SkillsWeight: 50, TitleExperienceWeight: 20, LocationWorkModeWeight: 15, CompensationWeight: 15}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid weights to pass, got: %v", err)
	}

	// Invalid: sum != 100
	invalidSum := ScoringWeights{SkillsWeight: 50, TitleExperienceWeight: 20, LocationWorkModeWeight: 15, CompensationWeight: 20}
	if err := invalidSum.Validate(); err == nil {
		t.Fatalf("expected error for sum 105, got nil")
	}

	// Invalid: negative weight
	invalidNegative := ScoringWeights{SkillsWeight: -10, TitleExperienceWeight: 50, LocationWorkModeWeight: 30, CompensationWeight: 30}
	if err := invalidNegative.Validate(); err == nil {
		t.Fatalf("expected error for negative weight, got nil")
	}
}

func TestJobMatch_HardGate_CompanyExclusion(t *testing.T) {
	engine := NewMatchEngine()
	profile := sampleCandidateProfile()
	prefs := sampleCandidatePreferences()
	job := sampleDiscoveredJob()

	// Normal case: Stripe is not blacklisted
	res, err := engine.EvaluateJobMatch(profile, prefs, job, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsEligible {
		t.Errorf("expected candidate to be eligible for Stripe, got ineligible: %v", res.HardGates)
	}

	// Blacklisted employer
	job.Company = "EvilCorp Inc"
	resBlocked, err := engine.EvaluateJobMatch(profile, prefs, job, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resBlocked.IsEligible {
		t.Errorf("expected candidate to be INELIGIBLE for EvilCorp, got eligible")
	}
	if resBlocked.FitTier != "ineligible" {
		t.Errorf("expected fit_tier ineligible, got %s", resBlocked.FitTier)
	}
}

func TestJobMatch_HardGate_WorkAuthorization_And_Unknown(t *testing.T) {
	engine := NewMatchEngine()
	profile := sampleCandidateProfile()
	prefs := sampleCandidatePreferences()
	job := sampleDiscoveredJob()

	// Candidate requires sponsorship
	prefs.Sponsorship = SponsorshipRequiresSponsorship

	// Subcase A: Job explicitly disallows sponsorship -> Hard Fail
	job.Description = "Senior Engineer. Note: No sponsorship provided. US Citizens only."
	resA, err := engine.EvaluateJobMatch(profile, prefs, job, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resA.IsEligible {
		t.Errorf("expected hard fail when employer disallows sponsorship and candidate needs it")
	}

	// Subcase B: Job does not mention sponsorship -> Treated as UNKNOWN per AT-028, NOT failed closed!
	job.Description = "Senior Engineer building Go backend services. Great team culture."
	resB, err := engine.EvaluateJobMatch(profile, prefs, job, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resB.IsEligible {
		t.Errorf("expected candidate to remain eligible when sponsorship info is unknown (AT-028), got ineligible")
	}

	foundUnknown := false
	for _, unk := range resB.UnknownRequirements {
		if unk == "Employer Visa Sponsorship Policy" {
			foundUnknown = true
			break
		}
	}
	if !foundUnknown {
		t.Errorf("expected 'Employer Visa Sponsorship Policy' to be tracked in UnknownRequirements (AT-028)")
	}
}

func TestJobMatch_HardGate_RemoteRelocation(t *testing.T) {
	engine := NewMatchEngine()
	profile := sampleCandidateProfile()
	prefs := sampleCandidatePreferences()
	job := sampleDiscoveredJob()

	// Candidate prefers 100% remote and will not relocate
	prefs.WorkModes = []WorkMode{WorkModeRemote}
	prefs.OpenToRelocation = false
	profile.Contact.Location = "Miami, FL"

	// Job is strictly on-site in Seattle, WA
	job.Location.IsRemote = false
	job.Location.IsHybrid = false
	job.Location.RawLocation = "Seattle, WA"

	res, err := engine.EvaluateJobMatch(profile, prefs, job, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsEligible {
		t.Errorf("expected hard fail when candidate is strictly remote in Miami and job is onsite in Seattle")
	}
}

func TestJobMatch_SkillsScoring_And_GapAnalysis(t *testing.T) {
	engine := NewMatchEngine()
	profile := sampleCandidateProfile() // has Go, Kubernetes, PostgreSQL, Kafka
	prefs := sampleCandidatePreferences()
	job := sampleDiscoveredJob() // requires Go, Kubernetes, Kafka, Rust (3/4 = 75%)

	res, err := engine.EvaluateJobMatch(profile, prefs, job, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Breakdown.SkillsScore != 75.0 {
		t.Errorf("expected 75.0%% skills score, got %.1f", res.Breakdown.SkillsScore)
	}

	if len(res.MatchedSkills) != 3 {
		t.Errorf("expected 3 matched skills, got %d", len(res.MatchedSkills))
	}
	if len(res.MissingSkills) != 1 || res.MissingSkills[0] != "Rust" {
		t.Errorf("expected Rust in missing skills, got %v", res.MissingSkills)
	}
}

func TestJobMatch_UndisclosedSalary_NeutralScore_AT003_AT028(t *testing.T) {
	engine := NewMatchEngine()
	profile := sampleCandidateProfile()
	prefs := sampleCandidatePreferences()
	job := sampleDiscoveredJob()

	// Salary undisclosed by employer
	job.Compensation.IsDisclosed = false
	job.Compensation.MinAmount = 0
	job.Compensation.MaxAmount = 0

	res, err := engine.EvaluateJobMatch(profile, prefs, job, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must be scored neutrally (80%) without penalizing the candidate (AT-003, AT-028)
	if res.Breakdown.CompensationScore != 80.0 {
		t.Errorf("expected neutral compensation score 80.0 for undisclosed salary, got %.1f", res.Breakdown.CompensationScore)
	}

	foundUnknown := false
	for _, req := range res.RequirementMatches {
		if req.Category == "compensation" && req.Status == MatchStatusUnknown {
			foundUnknown = true
			break
		}
	}
	if !foundUnknown {
		t.Errorf("expected compensation requirement match to have status 'unknown' per AT-028")
	}
}

func TestJobMatch_ConfigurableWeights(t *testing.T) {
	engine := NewMatchEngine()
	profile := sampleCandidateProfile()
	prefs := sampleCandidatePreferences()
	job := sampleDiscoveredJob()

	// Default weights (40 skills, 30 title, 15 location, 15 comp)
	resDef, err := engine.EvaluateJobMatch(profile, prefs, job, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Custom weights: 80% skills, 10% title, 5% location, 5% comp = 100
	custom := &ScoringWeights{
		SkillsWeight:           80,
		TitleExperienceWeight:  10,
		LocationWorkModeWeight: 5,
		CompensationWeight:     5,
	}
	resCustom, err := engine.EvaluateJobMatch(profile, prefs, job, custom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resCustom.Breakdown.WeightsUsed.SkillsWeight != 80 {
		t.Errorf("expected custom skills weight 80, got %d", resCustom.Breakdown.WeightsUsed.SkillsWeight)
	}

	// Since skills match is 75%, skills weighted score should be 75 * 0.8 = 60.0
	if resCustom.Breakdown.SkillsWeightedScore != 60.0 {
		t.Errorf("expected skills weighted score 60.0, got %.1f", resCustom.Breakdown.SkillsWeightedScore)
	}

	if resDef.OverallScore == resCustom.OverallScore {
		t.Errorf("custom weights should alter the overall score mathematically")
	}
}

func TestCareerService_JobMatchIntegration(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryCareerRepository()
	svc := NewCareerService(repo)

	userID := "user-match-svc-001"
	profile := sampleCandidateProfile()
	profile.UserID = userID
	if err := repo.SaveProfile(ctx, profile); err != nil {
		t.Fatalf("failed to save profile: %v", err)
	}

	prefs := sampleCandidatePreferences()
	prefs.UserID = userID
	if err := repo.SavePreferences(ctx, prefs); err != nil {
		t.Fatalf("failed to save preferences: %v", err)
	}

	job1 := sampleDiscoveredJob()
	job1.ID = "job-svc-01"
	job1.Title = "Senior Go Developer"

	job2 := sampleDiscoveredJob()
	job2.ID = "job-svc-02"
	job2.Company = "EvilCorp" // Trigger hard gate

	// Single job match
	singleRes, err := svc.EvaluateJobMatch(ctx, userID, *job1, nil)
	if err != nil {
		t.Fatalf("EvaluateJobMatch failed: %v", err)
	}
	if !singleRes.IsEligible {
		t.Errorf("expected job1 to be eligible")
	}
	if singleRes.OverallScore <= 0 {
		t.Errorf("expected positive overall score, got %f", singleRes.OverallScore)
	}

	// Batch match
	batchRes, err := svc.EvaluateBatchJobMatches(ctx, userID, []DiscoveredJob{*job1, *job2}, nil)
	if err != nil {
		t.Fatalf("EvaluateBatchJobMatches failed: %v", err)
	}
	if len(batchRes) != 2 {
		t.Fatalf("expected 2 batch results, got %d", len(batchRes))
	}

	// Because of sorting, eligible job1 should be index 0 and ineligible job2 should be index 1
	if !batchRes[0].IsEligible {
		t.Errorf("first sorted result should be eligible")
	}
	if batchRes[1].IsEligible {
		t.Errorf("second sorted result should be ineligible due to company exclusion")
	}

	// Weights validation error handling
	badWeights := &ScoringWeights{SkillsWeight: 50, TitleExperienceWeight: 50, CompensationWeight: 10} // sum 110
	_, err = svc.EvaluateJobMatch(ctx, userID, *job1, badWeights)
	if err == nil {
		t.Errorf("expected error for invalid weights sum")
	}
}

