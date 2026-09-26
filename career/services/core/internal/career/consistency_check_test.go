package career

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type consistencyFixtureSuite struct {
	Version   string `json:"version"`
	Domain    string `json:"domain"`
	Scenarios []struct {
		ScenarioID        string `json:"scenario_id"`
		Description       string `json:"description"`
		MasterProfileData struct {
			UserID      string `json:"user_id"`
			Experiences []struct {
				ID        string  `json:"id"`
				Company   string  `json:"company"`
				Title     string  `json:"title"`
				StartDate string  `json:"start_date"`
				EndDate   *string `json:"end_date"`
				Current   bool    `json:"current"`
			} `json:"experiences"`
			Skills []string `json:"skills"`
		} `json:"master_profile_data"`
		MasterResumeData struct {
			Title     string   `json:"title"`
			Company   string   `json:"company"`
			StartDate string   `json:"start_date"`
			EndDate   *string  `json:"end_date"`
			IsCurrent bool     `json:"is_current"`
			Skills    []string `json:"skills"`
		} `json:"master_resume_data"`
		LinkedInSnapshotData struct {
			Headline    string `json:"headline"`
			Experiences []struct {
				ID        string  `json:"id"`
				Company   string  `json:"company"`
				Title     string  `json:"title"`
				StartDate string  `json:"start_date"`
				EndDate   *string `json:"end_date"`
				Current   bool    `json:"current"`
			} `json:"experiences"`
			Skills []string `json:"skills"`
		} `json:"linkedin_snapshot_data"`
		ExpectedMismatchesCount   int       `json:"expected_mismatches_count,omitempty"`
		ExpectedHighSeverityCount int       `json:"expected_high_severity_count,omitempty"`
		ExpectedScoreRange        []float64 `json:"expected_score_range,omitempty"`
		ExpectedConsistencyScore  float64   `json:"expected_consistency_score,omitempty"`
		MismatchID                string    `json:"mismatch_id,omitempty"`
		ResolutionDecision        struct {
			Action            string `json:"action"`
			ChosenValue       string `json:"chosen_value"`
			UserJustification string `json:"user_justification"`
		} `json:"resolution_decision,omitempty"`
		ExpectedUpdatedTitle       string `json:"expected_updated_title,omitempty"`
		ExpectedAuditTrailRecorded bool   `json:"expected_audit_trail_recorded,omitempty"`
	} `json:"scenarios"`
}

func loadConsistencyFixture(t *testing.T) consistencyFixtureSuite {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "career_consistency_check.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture file: %v", err)
	}

	var suite consistencyFixtureSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("failed to unmarshal fixture JSON: %v", err)
	}
	return suite
}

func TestConsistency_FixtureSuiteEvaluation(t *testing.T) {
	suite := loadConsistencyFixture(t)
	if len(suite.Scenarios) != 4 {
		t.Fatalf("expected 4 fixture scenarios, got %d", len(suite.Scenarios))
	}

	for _, s := range suite.Scenarios {
		t.Run(s.ScenarioID, func(t *testing.T) {
			switch s.ScenarioID {
			case "title_and_date_discrepancies":
				profile := &MasterCareerProfile{
					UserID: s.MasterProfileData.UserID,
					Skills: make([]SkillItem, 0),
					Experiences: make([]ExperienceItem, 0),
				}
				for _, sk := range s.MasterProfileData.Skills {
					profile.Skills = append(profile.Skills, SkillItem{Name: sk, Confirmed: true})
				}
				for _, exp := range s.MasterProfileData.Experiences {
					st, _ := time.Parse(time.RFC3339, exp.StartDate)
					profile.Experiences = append(profile.Experiences, ExperienceItem{
						ID:        exp.ID,
						Company:   exp.Company,
						Title:     exp.Title,
						StartDate: &st,
						IsCurrent: exp.Current,
					})
				}

				linkedIn := &LinkedInProfileSnapshot{
					SnapshotID:  "snap-1",
					UserID:      s.MasterProfileData.UserID,
					Headline:    s.LinkedInSnapshotData.Headline,
					Skills:      s.LinkedInSnapshotData.Skills,
					Experiences: make([]ExperienceItem, 0),
				}
				for _, exp := range s.LinkedInSnapshotData.Experiences {
					st, _ := time.Parse(time.RFC3339, exp.StartDate)
					var et *time.Time
					if exp.EndDate != nil {
						tVal, _ := time.Parse(time.RFC3339, *exp.EndDate)
						et = &tVal
					}
					linkedIn.Experiences = append(linkedIn.Experiences, ExperienceItem{
						ID:        exp.ID,
						Company:   exp.Company,
						Title:     exp.Title,
						StartDate: &st,
						EndDate:   et,
						IsCurrent: exp.Current,
					})
				}

				report := CompareCareerRecords(s.MasterProfileData.UserID, profile, nil, linkedIn)
				if report.TotalMismatches < 3 {
					t.Errorf("expected at least 3 mismatches, got %d", report.TotalMismatches)
				}
				if report.HighSeverityCount != s.ExpectedHighSeverityCount {
					t.Errorf("expected %d high severity mismatches, got %d", s.ExpectedHighSeverityCount, report.HighSeverityCount)
				}
				if report.OverallConsistencyScore < s.ExpectedScoreRange[0] || report.OverallConsistencyScore > s.ExpectedScoreRange[1] {
					t.Errorf("score %f out of expected range [%f, %f]", report.OverallConsistencyScore, s.ExpectedScoreRange[0], s.ExpectedScoreRange[1])
				}

			case "skill_coverage_differences":
				profile := &MasterCareerProfile{
					UserID: s.MasterProfileData.UserID,
					Skills: make([]SkillItem, 0),
					Experiences: make([]ExperienceItem, 0),
				}
				for _, sk := range s.MasterProfileData.Skills {
					profile.Skills = append(profile.Skills, SkillItem{Name: sk, Confirmed: true})
				}
				for _, exp := range s.MasterProfileData.Experiences {
					st, _ := time.Parse(time.RFC3339, exp.StartDate)
					profile.Experiences = append(profile.Experiences, ExperienceItem{
						ID:        exp.ID,
						Company:   exp.Company,
						Title:     exp.Title,
						StartDate: &st,
						IsCurrent: exp.Current,
					})
				}

				linkedIn := &LinkedInProfileSnapshot{
					SnapshotID:  "snap-2",
					UserID:      s.MasterProfileData.UserID,
					Headline:    s.LinkedInSnapshotData.Headline,
					Skills:      s.LinkedInSnapshotData.Skills,
					Experiences: profile.Experiences,
				}

				report := CompareCareerRecords(s.MasterProfileData.UserID, profile, nil, linkedIn)
				if report.HighSeverityCount != 0 {
					t.Errorf("expected 0 high severity mismatches, got %d", report.HighSeverityCount)
				}
				if report.LowSeverityCount != 2 {
					t.Errorf("expected 2 low severity skill mismatches, got %d", report.LowSeverityCount)
				}
				if report.OverallConsistencyScore < 85.0 {
					t.Errorf("expected score >= 85.0, got %f", report.OverallConsistencyScore)
				}

			case "human_resolution_and_audit_propagation":
				profile := &MasterCareerProfile{
					UserID: "user-cand-101",
					Experiences: []ExperienceItem{
						{
							ID:      "exp-p1",
							Company: "FastFintech",
							Title:   "Senior Backend Engineer",
						},
					},
					AuditTrail: make([]ProfileFieldAudit, 0),
				}

				report := &ConsistencyAuditReport{
					ReportID: "rep-res-test",
					UserID:   "user-cand-101",
					Mismatches: []ConsistencyMismatch{
						{
							MismatchID: "mis-title-fastfintech",
							FieldType:  FieldJobTitle,
							Severity:   SeverityHigh,
							EntityKey:  "experience:FastFintech",
						},
					},
					TotalMismatches:   1,
					HighSeverityCount: 1,
				}

				decision := ResolutionDecision{
					Action:            ResolutionAction(s.ResolutionDecision.Action),
					ChosenValue:       s.ResolutionDecision.ChosenValue,
					UserJustification: s.ResolutionDecision.UserJustification,
				}

				updatedReport, err := ApplyConsistencyResolution(report, s.MismatchID, decision, profile)
				if err != nil {
					t.Fatalf("failed to apply resolution: %v", err)
				}
				if updatedReport.ResolvedCount != 1 {
					t.Errorf("expected 1 resolved mismatch, got %d", updatedReport.ResolvedCount)
				}
				if profile.Experiences[0].Title != s.ExpectedUpdatedTitle {
					t.Errorf("profile title not updated: got %s, want %s", profile.Experiences[0].Title, s.ExpectedUpdatedTitle)
				}
				if len(profile.AuditTrail) == 0 {
					t.Fatalf("expected audit trail record to be added")
				}
				if profile.AuditTrail[0].Reason != s.ResolutionDecision.UserJustification {
					t.Errorf("audit notes mismatch: got %s", profile.AuditTrail[0].Reason)
				}

			case "perfect_consistency_aligned_profiles":
				profile := &MasterCareerProfile{
					UserID: s.MasterProfileData.UserID,
					Experiences: []ExperienceItem{
						{
							Company: "Acme Corp",
							Title:   "Principal Systems Engineer",
						},
					},
					Skills: []SkillItem{
						{Name: "Go"},
						{Name: "Kubernetes"},
						{Name: "gRPC"},
					},
				}
				linkedIn := &LinkedInProfileSnapshot{
					Experiences: []ExperienceItem{
						{
							Company: "Acme Corp",
							Title:   "Principal Systems Engineer",
						},
					},
					Skills: []string{"Go", "Kubernetes", "gRPC"},
				}

				report := CompareCareerRecords(s.MasterProfileData.UserID, profile, nil, linkedIn)
				if report.TotalMismatches != 0 {
					t.Errorf("expected 0 mismatches for aligned profiles, got %d", report.TotalMismatches)
				}
				if report.OverallConsistencyScore != 100.0 {
					t.Errorf("expected 100.0 score, got %f", report.OverallConsistencyScore)
				}
			}
		})
	}
}

func TestConsistency_ResolutionValidation_AT007(t *testing.T) {
	report := &ConsistencyAuditReport{
		ReportID: "rep-val",
		UserID:   "user-1",
		Mismatches: []ConsistencyMismatch{
			{MismatchID: "mis-1", FieldType: FieldJobTitle},
		},
	}
	profile := &MasterCareerProfile{UserID: "user-1"}

	// Invalid action
	badDecision := ResolutionDecision{
		Action:      "invalid_action",
		ChosenValue: "Val",
	}
	_, err := ApplyConsistencyResolution(report, "mis-1", badDecision, profile)
	if err == nil {
		t.Errorf("expected error on invalid resolution action")
	}

	// Empty chosen value
	emptyDecision := ResolutionDecision{
		Action:      ResolveRetainProfile,
		ChosenValue: "   ",
	}
	_, err = ApplyConsistencyResolution(report, "mis-1", emptyDecision, profile)
	if err == nil {
		t.Errorf("expected error on empty chosen value")
	}

	// Non-existent mismatch
	validDecision := ResolutionDecision{
		Action:      ResolveRetainProfile,
		ChosenValue: "Staff Engineer",
	}
	_, err = ApplyConsistencyResolution(report, "non-existent", validDecision, profile)
	if err == nil {
		t.Errorf("expected error on non-existent mismatch ID")
	}
}

func TestConsistency_ServiceIntegration(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-service-test"

	// Create profile
	profile := &MasterCareerProfile{
		ID:     "prof-1",
		UserID: userID,
		Experiences: []ExperienceItem{
			{
				ID:      "exp-1",
				Company: "TechCo",
				Title:   "Backend Engineer",
			},
		},
		Skills: []SkillItem{
			{Name: "Go"},
		},
	}
	_ = repo.SaveProfile(ctx, profile)

	// Save LinkedIn snapshot
	snapshot := &LinkedInProfileSnapshot{
		SnapshotID: "snap-1",
		UserID:     userID,
		Headline:   "Senior Backend Engineer at TechCo",
		Experiences: []ExperienceItem{
			{
				ID:      "exp-li-1",
				Company: "TechCo",
				Title:   "Senior Backend Engineer",
			},
		},
		Skills:     []string{"Go", "Python"},
		ImportedAt: time.Now().UTC(),
		SourceMode: "user_export",
	}
	if err := service.SaveLinkedInSnapshot(ctx, snapshot); err != nil {
		t.Fatalf("failed to save snapshot: %v", err)
	}

	// Run consistency audit
	report, err := service.RunConsistencyAudit(ctx, userID)
	if err != nil {
		t.Fatalf("failed to run audit: %v", err)
	}
	if report.TotalMismatches == 0 {
		t.Fatalf("expected mismatches between TechCo titles and skills")
	}

	// Fetch latest report
	latest, err := service.GetLatestConsistencyReport(ctx, userID)
	if err != nil || latest.ReportID != report.ReportID {
		t.Fatalf("failed to get latest report: %v", err)
	}

	// Resolve mismatch
	mismatchID := report.Mismatches[0].MismatchID
	decision := ResolutionDecision{
		Action:            ResolveRetainLinkedIn,
		ChosenValue:       "Senior Backend Engineer",
		UserJustification: "LinkedIn reflects current promotion",
	}

	resolvedReport, err := service.ResolveConsistencyMismatch(ctx, userID, report.ReportID, mismatchID, decision)
	if err != nil {
		t.Fatalf("failed to resolve mismatch: %v", err)
	}
	if resolvedReport.ResolvedCount != 1 {
		t.Errorf("expected 1 resolved mismatch, got %d", resolvedReport.ResolvedCount)
	}

	// Verify profile updated
	updatedProfile, err := repo.GetProfile(ctx, userID)
	if err != nil {
		t.Fatalf("failed to retrieve profile: %v", err)
	}
	if updatedProfile.Experiences[0].Title != "Senior Backend Engineer" {
		t.Errorf("expected updated title 'Senior Backend Engineer', got '%s'", updatedProfile.Experiences[0].Title)
	}
}
