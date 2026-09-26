package career

import (
	"context"
	"testing"
	"time"
)

func createTestProfileForFeedback() *MasterCareerProfile {
	now := time.Now()
	startDate := parseTestTime("2022-01-01")
	startDate2 := parseTestTime("2019-03-01")
	endDate2 := parseTestTime("2021-12-31")
	eduStart := parseTestTime("2014-01-01")
	eduEnd := parseTestTime("2018-12-31")

	return &MasterCareerProfile{
		ID:        "prof_feedback_test",
		UserID:    "user_feedback_100",
		CreatedAt: now,
		UpdatedAt: now,
		Contact: ContactInfo{
			FullName: "Alex Morgan",
			Email:    "alex.morgan@example.com",
			Phone:    "+1 (555) 019-2834",
			Location: "Remote (Global)",
			Headline: "Staff Backend Engineer | Distributed Systems & High-Throughput Microservices",
			Summary:  "Experienced architect with 8+ years building enterprise payment infrastructure, achieving 99.99% uptime at scale.",
		},
		Experiences: []ExperienceItem{
			{
				ID:          "exp_1",
				Company:     "Fintech Global",
				Title:       "Lead Backend Engineer",
				Location:    "Dhaka",
				StartDate:   startDate,
				IsCurrent:   true,
				Description: "Architected microservices processing 45,000 RPS.",
				Highlights: []string{
					"Engineered low-latency transaction processing engine in Go, reducing latency by 42% across 10M daily transactions.",
					"Scaled Kafka event broker pipeline handling 15,000 messages/sec with 99.99% SLA.",
					"Responsible for team standups and bug fixes.", // Weak, unquantified, passive
				},
				SkillsUsed: []string{"Go", "Kafka", "PostgreSQL", "Docker", "Kubernetes"},
				Confirmed:  true,
			},
			{
				ID:          "exp_2",
				Company:     "Cloud Scale Ltd",
				Title:       "Senior Software Engineer",
				Location:    "Remote",
				StartDate:   startDate2,
				EndDate:     endDate2,
				IsCurrent:   false,
				Description: "Maintained cloud infrastructure.",
				Highlights: []string{
					"Automated CI/CD deployment pipelines cut release turnaround time by 65%.",
					"Worked on various database optimizations and helped developers.", // Weak phrase
				},
				SkillsUsed: []string{"Go", "AWS", "Docker"},
				Confirmed:  true,
			},
		},
		Education: []EducationItem{
			{
				ID:           "edu_1",
				Institution:  "Bangladesh University of Engineering and Technology (BUET)",
				Degree:       "Bachelor of Science",
				FieldOfStudy: "Computer Science and Engineering",
				StartDate:    eduStart,
				EndDate:      eduEnd,
				Confirmed:    true,
			},
		},
		Skills: []SkillItem{
			{ID: "sk_1", Name: "Go", Category: "Backend", Proficiency: ProficiencyExpert, Confirmed: true},
			{ID: "sk_2", Name: "Kafka", Category: "Data", Proficiency: ProficiencyAdvanced, Confirmed: true},
			{ID: "sk_3", Name: "PostgreSQL", Category: "Database", Proficiency: ProficiencyAdvanced, Confirmed: true},
			{ID: "sk_4", Name: "Kubernetes", Category: "DevOps", Proficiency: ProficiencyIntermediate, Confirmed: true},
		},
	}
}

func TestFeedbackAnalyzer_AnalyzeProfile(t *testing.T) {
	analyzer := NewFeedbackAnalyzer()
	profile := createTestProfileForFeedback()

	report, err := analyzer.AnalyzeProfile(profile)
	if err != nil {
		t.Fatalf("AnalyzeProfile failed: %v", err)
	}

	if report.ProfileID != profile.ID {
		t.Errorf("expected ProfileID %s, got %s", profile.ID, report.ProfileID)
	}
	if report.ID == "" {
		t.Error("expected non-empty report ID")
	}
	if report.OverallScore <= 70 {
		t.Errorf("expected OverallScore > 70, got %d", report.OverallScore)
	}
	if report.ATSReadabilityScore <= 80 {
		t.Errorf("expected ATSReadabilityScore > 80, got %d", report.ATSReadabilityScore)
	}

	// Metrics & Quantification rate check (3 out of 5 bullets quantified: 0.6)
	if report.QuantificationRate < 0.5 {
		t.Errorf("expected QuantificationRate >= 0.5, got %f", report.QuantificationRate)
	}

	// Action verb density check
	if report.ActionVerbDensity < 0.5 {
		t.Errorf("expected ActionVerbDensity >= 0.5, got %f", report.ActionVerbDensity)
	}

	// Section breakdown check
	if len(report.SectionScores) == 0 {
		t.Error("expected non-empty SectionScores")
	}

	// Verify weak phrase or issue detection in CriticalIssues
	var foundIssue bool
	for _, issue := range report.CriticalIssues {
		if issue.Section == "experience" || issue.Section == "Experience" {
			foundIssue = true
			break
		}
	}
	if !foundIssue {
		t.Error("expected experience section issues or recommendations to be flagged")
	}

	// Verify improvement suggestions exist with Before & After text
	if len(report.ImprovementSuggestions) == 0 {
		t.Fatal("expected ImprovementSuggestions to be generated")
	}
	var hasRewrites bool
	for _, s := range report.ImprovementSuggestions {
		if s.CurrentText != "" && s.SuggestedImprovement != "" {
			hasRewrites = true
			break
		}
	}
	if !hasRewrites {
		t.Error("expected suggestions to provide concrete CurrentText and SuggestedImprovement")
	}
}

func TestFeedbackAnalyzer_SkillsDemandAnalysis_CAR06_AT028(t *testing.T) {
	analyzer := NewFeedbackAnalyzer()

	profile := &MasterCareerProfile{
		ID:        "prof_200",
		UserID:    "user_200",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Contact: ContactInfo{
			FullName: "Sara Ahmed",
			Email:    "sara@example.com",
			Headline: "Backend Developer",
		},
		Skills: []SkillItem{
			{ID: "sk_1", Name: "Go", Category: "Backend", Proficiency: ProficiencyAdvanced, Confirmed: true},
			{ID: "sk_2", Name: "PostgreSQL", Category: "Database", Proficiency: ProficiencyAdvanced, Confirmed: true},
			{ID: "sk_3", Name: "Docker", Category: "DevOps", Proficiency: ProficiencyIntermediate, Confirmed: true},
		},
	}

	analysis, err := analyzer.AnalyzeSkillsDemand(profile, "Backend")
	if err != nil {
		t.Fatalf("AnalyzeSkillsDemand failed: %v", err)
	}

	// Verify role category and baseline counts
	if analysis.RoleCategory != "backend" {
		t.Errorf("expected RoleCategory 'backend', got '%s'", analysis.RoleCategory)
	}
	if analysis.TotalMarketSkillsAnalyzed == 0 {
		t.Error("expected positive TotalMarketSkillsAnalyzed")
	}

	// CAR-06 & AT-003: Strict Separation check
	// Possessed skills must NOT be contaminated by market gap suggestions
	if analysis.CandidatePossessedCount != len(analysis.PossessedSkills) {
		t.Errorf("mismatch CandidatePossessedCount %d vs len(PossessedSkills) %d",
			analysis.CandidatePossessedCount, len(analysis.PossessedSkills))
	}

	// All items in PossessedSkills must have IsPossessed = true
	for _, item := range analysis.PossessedSkills {
		if !item.IsPossessed {
			t.Errorf("possessed skill %s must have IsPossessed=true", item.SkillName)
		}
		if item.Status != "possessed_and_in_demand" {
			t.Errorf("possessed skill %s must have Status 'possessed_and_in_demand'", item.SkillName)
		}
		// AT-028 Evidence requirement check
		if item.EvidenceDate.IsZero() {
			t.Errorf("market skill demand %s missing EvidenceDate", item.SkillName)
		}
		if item.SampleSize <= 0 {
			t.Errorf("market skill demand %s invalid SampleSize %d", item.SkillName, item.SampleSize)
		}
	}

	// All items in MarketGaps must have IsPossessed = false (Market suggestions, NOT possessed facts)
	if len(analysis.MarketGaps) == 0 {
		t.Error("expected market gaps to be identified")
	}
	for _, item := range analysis.MarketGaps {
		if item.IsPossessed {
			t.Errorf("gap skill %s cannot have IsPossessed=true (CAR-06 violation)", item.SkillName)
		}
		if item.Status != "market_demand_gap" {
			t.Errorf("gap skill %s must have Status 'market_demand_gap'", item.SkillName)
		}
		// AT-028 Evidence requirement check
		if item.EvidenceDate.IsZero() {
			t.Errorf("gap skill %s missing EvidenceDate", item.SkillName)
		}
		if item.SampleSize <= 0 {
			t.Errorf("gap skill %s invalid SampleSize %d", item.SkillName, item.SampleSize)
		}
	}

	// Profile skills must remain unaltered (zero hallucination / zero silent injection)
	if len(profile.Skills) != 3 {
		t.Errorf("profile skills count modified! expected 3, got %d", len(profile.Skills))
	}
}

func TestCareerService_FeedbackIntegration(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)

	profile := createTestProfileForFeedback()
	err := repo.SaveProfile(ctx, profile)
	if err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	// Generate feedback report
	report, err := service.GenerateResumeFeedback(ctx, profile.UserID)
	if err != nil {
		t.Fatalf("GenerateResumeFeedback failed: %v", err)
	}
	if report.ID == "" {
		t.Error("expected non-empty report ID")
	}
	if report.ProfileID != profile.ID {
		t.Errorf("expected ProfileID %s, got %s", profile.ID, report.ProfileID)
	}
	if report.OverallScore <= 0 {
		t.Errorf("expected positive OverallScore, got %d", report.OverallScore)
	}

	// Fetch latest feedback report
	latest, err := service.GetLatestResumeFeedback(ctx, profile.UserID)
	if err != nil {
		t.Fatalf("GetLatestResumeFeedback failed: %v", err)
	}
	if latest.ID != report.ID {
		t.Errorf("expected report ID %s, got %s", report.ID, latest.ID)
	}

	// Get skills demand analysis
	demand, err := service.GetSkillsDemandAnalysis(ctx, profile.UserID, "Backend")
	if err != nil {
		t.Fatalf("GetSkillsDemandAnalysis failed: %v", err)
	}
	if demand.RoleCategory != "backend" {
		t.Errorf("expected RoleCategory backend, got %s", demand.RoleCategory)
	}
	if demand.TotalMarketSkillsAnalyzed == 0 {
		t.Error("expected non-empty TotalMarketSkillsAnalyzed")
	}
}
