package career

import (
	"context"
	"testing"
)

func TestCareerPreferences_AT003_ZeroFabricationDefaults(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-pref-zero-fab"

	// 1. Initial preferences must NOT fabricate salary or guess visa eligibility (AT-003)
	pref, err := service.GetCareerPreferences(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pref.Sponsorship != SponsorshipUnspecified {
		t.Errorf("AT-003 violation: expected SponsorshipUnspecified, got %q", pref.Sponsorship)
	}
	if pref.Salary.IsExplicitlySet {
		t.Errorf("AT-003 violation: expected Salary.IsExplicitlySet false, got true")
	}
	if len(pref.TargetRoles) != 0 {
		t.Errorf("expected empty target roles, got %v", pref.TargetRoles)
	}
	if len(pref.Exclusions.ExcludedCompanies) != 0 {
		t.Errorf("expected empty excluded companies, got %v", pref.Exclusions.ExcludedCompanies)
	}
}

func TestCareerPreferences_Validation(t *testing.T) {
	pref := NewCareerPreferences("user-val-001")

	// Invalid salary: Min > Max
	pref.Salary.IsExplicitlySet = true
	pref.Salary.MinimumAmount = 150000
	pref.Salary.MaximumAmount = 120000
	pref.Salary.Currency = "USD"
	if err := pref.Validate(); err != ErrInvalidSalaryRange {
		t.Errorf("expected ErrInvalidSalaryRange, got %v", err)
	}

	// Invalid currency: not 3 characters
	pref.Salary.MaximumAmount = 180000
	pref.Salary.Currency = "US"
	if err := pref.Validate(); err != ErrInvalidCurrency {
		t.Errorf("expected ErrInvalidCurrency, got %v", err)
	}

	// Valid preferences
	pref.Salary.Currency = "USD"
	if err := pref.Validate(); err != nil {
		t.Errorf("expected valid preferences, got: %v", err)
	}
}

func TestCareerPreferences_CRUD(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-pref-crud"

	pref, err := service.GetCareerPreferences(ctx, userID)
	if err != nil {
		t.Fatalf("failed to get initial preferences: %v", err)
	}

	// Configure preferences
	pref.TargetRoles = []string{"Lead Go Engineer", "Distributed Systems Architect"}
	pref.TargetLocations = []string{"United States", "Germany", "Remote"}
	pref.WorkModes = []WorkMode{WorkModeRemote, WorkModeHybrid}
	pref.JobTypes = []JobType{JobTypeFullTime, JobTypeContract}
	pref.Salary = SalaryPreference{
		MinimumAmount:   130000,
		TargetAmount:    155000,
		MaximumAmount:   180000,
		Currency:        "USD",
		Interval:        SalaryIntervalAnnual,
		IsExplicitlySet: true,
	}
	pref.Sponsorship = SponsorshipRequiresSponsorship
	pref.OpenToRelocation = true
	pref.NoticePeriodDays = 30
	pref.Exclusions = ExclusionRules{
		ExcludedCompanies:     []string{"Previous Toxic Corp", "BadTech Ltd"},
		ExcludedKeywords:      []string{"crypto", "gambling", "junior"},
		ExcludedIndustries:    []string{"defense"},
		BlockStaffingAgencies: true,
	}

	saved, err := service.SaveCareerPreferences(ctx, userID, pref)
	if err != nil {
		t.Fatalf("failed to save preferences: %v", err)
	}

	if len(saved.TargetRoles) != 2 || saved.TargetRoles[0] != "Lead Go Engineer" {
		t.Errorf("unexpected target roles: %v", saved.TargetRoles)
	}
	if saved.Sponsorship != SponsorshipRequiresSponsorship {
		t.Errorf("expected sponsorship requires_sponsorship, got %q", saved.Sponsorship)
	}
	if !saved.Salary.IsExplicitlySet || saved.Salary.MinimumAmount != 130000 {
		t.Errorf("unexpected salary preference: %+v", saved.Salary)
	}

	// Verify persistence in repo
	fetched, err := service.GetCareerPreferences(ctx, userID)
	if err != nil {
		t.Fatalf("failed to fetch saved preferences: %v", err)
	}
	if fetched.NoticePeriodDays != 30 {
		t.Errorf("expected notice period 30, got %d", fetched.NoticePeriodDays)
	}
}

func TestCareerPreferences_MatchesExclusion(t *testing.T) {
	pref := NewCareerPreferences("user-exclusion-test")
	pref.Exclusions = ExclusionRules{
		ExcludedCompanies:  []string{"Current Employer Corp", "Spammy Recruiter Inc"},
		ExcludedKeywords:   []string{"blockchain", "casino", "intern"},
		ExcludedIndustries: []string{"gambling"},
	}

	tests := []struct {
		company     string
		title       string
		description string
		expectedEx  bool
		reasonSub   string
	}{
		{
			company:     "current employer corp",
			title:       "Staff Engineer",
			description: "Great backend role",
			expectedEx:  true,
			reasonSub:   "company excluded",
		},
		{
			company:     "Acme Health",
			title:       "Blockchain Protocol Engineer",
			description: "Decentralized consensus",
			expectedEx:  true,
			reasonSub:   "title contains excluded keyword",
		},
		{
			company:     "FinTech Global",
			title:       "Senior Software Engineer",
			description: "Join our online casino gaming platform",
			expectedEx:  true,
			reasonSub:   "description contains excluded keyword",
		},
		{
			company:     "OpenAI Innovators",
			title:       "Principal Go Architect",
			description: "Building scalable distributed AI serving pipelines in Go and Kubernetes",
			expectedEx:  false,
			reasonSub:   "",
		},
	}

	for _, tt := range tests {
		excluded, reason := pref.MatchesExclusion(tt.company, tt.title, tt.description)
		if excluded != tt.expectedEx {
			t.Errorf("job (%s, %s): expected excluded=%v, got=%v (reason: %s)", tt.company, tt.title, tt.expectedEx, excluded, reason)
		}
	}
}
