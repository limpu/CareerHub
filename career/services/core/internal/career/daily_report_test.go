package career

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDailyReport_Generation_AllComponents(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	userID := "usr_test_daily_01"
	workspaceID := "ws_test_daily_01"
	now := time.Now().UTC()

	// Setup profile
	profile := &MasterCareerProfile{
		UserID: userID,
		Contact: ContactInfo{
			Email:    "test@example.com",
			FullName: "Alex Rivera",
			Summary:  "Experienced backend engineer specializing in distributed systems and Go.",
		},
		Experiences: []ExperienceItem{
			{
				Company: "TechCorp",
				Title:   "Senior Engineer",
				Highlights: []string{
					"Scaled backend services to 50k RPS with 99.99% uptime",
				},
			},
		},
		Skills: []SkillItem{
			{Name: "Go"}, {Name: "PostgreSQL"}, {Name: "Redis"}, {Name: "Docker"}, {Name: "Kubernetes"},
		},
	}
	_ = repo.SaveProfile(ctx, profile)

	// Setup applications
	app1 := &ApplicationRecord{
		ID:        "app_001",
		UserID:    userID,
		Company:   "Datadog",
		Title:     "Staff Engineer",
		Stage:     StageApplied,
		AppliedAt: now.Add(-10 * 24 * time.Hour),
		UpdatedAt: now.Add(-10 * 24 * time.Hour), // Stalled 10 days
	}
	app2 := &ApplicationRecord{
		ID:        "app_002",
		UserID:    userID,
		Company:   "Uber",
		Title:     "Senior Systems Engineer",
		Stage:     StageInterviewing,
		AppliedAt: now.Add(-2 * 24 * time.Hour),
		UpdatedAt: now,
	}
	_ = repo.SaveApplicationRecord(ctx, app1)
	_ = repo.SaveApplicationRecord(ctx, app2)

	// Setup saved jobs
	job1 := &SavedJob{
		ID:       "job_001",
		UserID:   userID,
		Title:    "Principal Go Engineer",
		Company:  "HashiCorp",
		Location: JobLocation{City: "Remote", Country: "US"},
	}
	_ = repo.SaveSavedJob(ctx, job1)

	// Generate report
	report, err := service.GetDailyReport(ctx, userID, workspaceID, now.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("failed to generate daily report: %v", err)
	}

	if report.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, report.UserID)
	}

	// Verify follow-ups (app1 stalled 10 days)
	if len(report.FollowUps) != 1 {
		t.Fatalf("expected 1 follow-up recommendation, got %d", len(report.FollowUps))
	}
	if report.FollowUps[0].Company != "Datadog" {
		t.Errorf("expected follow-up for Datadog, got %s", report.FollowUps[0].Company)
	}
	if report.FollowUps[0].DaysSinceApplied < 9 {
		t.Errorf("expected at least 9 days since applied, got %d", report.FollowUps[0].DaysSinceApplied)
	}

	// Verify interviews
	if len(report.Interviews) != 1 {
		t.Fatalf("expected 1 interview scheduled, got %d", len(report.Interviews))
	}
	if report.Interviews[0].Company != "Uber" {
		t.Errorf("expected interview with Uber, got %s", report.Interviews[0].Company)
	}

	// Verify activity summary
	if report.Activity.ActiveApplicationsTotal != 2 {
		t.Errorf("expected 2 active applications, got %d", report.Activity.ActiveApplicationsTotal)
	}
	if report.Activity.MomentumScore <= 0 || report.Activity.MomentumScore > 100 {
		t.Errorf("expected momentum score between 1-100, got %d", report.Activity.MomentumScore)
	}

	// Verify job highlights
	if len(report.JobHighlights) != 1 {
		t.Errorf("expected 1 job highlight, got %d", len(report.JobHighlights))
	}
}

func TestDailyReport_FollowUpDetection_StalledApplications(t *testing.T) {
	now := time.Now().UTC()

	apps := []ApplicationRecord{
		{
			ID:        "app_fresh",
			Company:   "FreshCo",
			Title:     "Dev",
			Stage:     StageApplied,
			AppliedAt: now.Add(-2 * 24 * time.Hour), // 2 days ago
			UpdatedAt: now.Add(-2 * 24 * time.Hour),
		},
		{
			ID:        "app_stalled_1",
			Company:   "StalledCo",
			Title:     "Dev",
			Stage:     StageApplied,
			AppliedAt: now.Add(-8 * 24 * time.Hour), // 8 days ago
			UpdatedAt: now.Add(-8 * 24 * time.Hour),
		},
		{
			ID:        "app_stalled_2",
			Company:   "NeedsConfirmCo",
			Title:     "Dev",
			Stage:     StageNeedsConfirmation,
			AppliedAt: now.Add(-14 * 24 * time.Hour), // 14 days ago
			UpdatedAt: now.Add(-14 * 24 * time.Hour),
		},
		{
			ID:        "app_rejected",
			Company:   "RejectedCo",
			Title:     "Dev",
			Stage:     StageRejected, // Should be ignored
			AppliedAt: now.Add(-20 * 24 * time.Hour),
			UpdatedAt: now.Add(-20 * 24 * time.Hour),
		},
	}

	followUps := DetectApplicationFollowUps(apps, 7, now)
	if len(followUps) != 2 {
		t.Fatalf("expected 2 stalled applications detected, got %d", len(followUps))
	}

	foundStalled := false
	foundNeedsConfirm := false
	for _, f := range followUps {
		if f.Company == "StalledCo" {
			foundStalled = true
			if f.DaysSinceApplied != 8 {
				t.Errorf("expected 8 days, got %d", f.DaysSinceApplied)
			}
		}
		if f.Company == "NeedsConfirmCo" {
			foundNeedsConfirm = true
			if f.DaysSinceApplied != 14 {
				t.Errorf("expected 14 days, got %d", f.DaysSinceApplied)
			}
		}
	}

	if !foundStalled || !foundNeedsConfirm {
		t.Errorf("missing expected stalled companies in follow-up recommendations")
	}
}

func TestDailyReport_ProfileCompletenessAudit(t *testing.T) {
	// Profile with missing summary, no metrics in experience, low skills, missing email
	incompleteProfile := &MasterCareerProfile{
		UserID: "usr_incomplete",
		Contact: ContactInfo{
			Email:    "", // Missing email
			FullName: "Jane Doe",
			Summary:  "", // Blank summary
		},
		Experiences: []ExperienceItem{
			{
				Company:    "OldCo",
				Title:      "Developer",
				Highlights: []string{"Wrote some code", "Fixed bugs"}, // No metrics
			},
		},
		Skills: []SkillItem{
			{Name: "JavaScript"},
			{Name: "HTML"},
		}, // Only 2 skills (< 5)
	}

	alerts := AuditProfileCompleteness(incompleteProfile)
	if len(alerts) < 4 {
		t.Fatalf("expected at least 4 profile gap alerts, got %d", len(alerts))
	}

	sections := make(map[string]bool)
	for _, a := range alerts {
		sections[a.Section] = true
	}

	if !sections["summary"] {
		t.Errorf("expected alert for missing summary")
	}
	if !sections["experience"] {
		t.Errorf("expected alert for experience lacking metrics")
	}
	if !sections["skills"] {
		t.Errorf("expected alert for low skills count")
	}
	if !sections["contact"] {
		t.Errorf("expected alert for missing contact email")
	}
}

func TestDailyReport_TimezoneAndDST_AT018(t *testing.T) {
	// Test AT-018: Daylight Saving Time transition in America/New_York
	// In 2026, DST starts Sunday, March 8, 2026 (2:00 AM becomes 3:00 AM, EST UTC-5 -> EDT UTC-4)

	// Standard Time: March 7, 2026 at 12:00 UTC (07:00 EST). Target 09:00 EST -> 14:00 UTC
	t1 := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	next1, err := CalculateNextReportDelivery(t1, "America/New_York", 9, 0, "daily")
	if err != nil {
		t.Fatalf("unexpected error for America/New_York: %v", err)
	}

	// 09:00 EST on March 7 is 14:00 UTC. Since t1 is 12:00 UTC (07:00 EST), next run is March 7 14:00 UTC
	expected1 := time.Date(2026, 3, 7, 14, 0, 0, 0, time.UTC)
	if !next1.Equal(expected1) {
		t.Errorf("expected %v, got %v", expected1, next1)
	}

	// After DST transition: March 9, 2026 at 10:00 UTC (06:00 EDT). Target 09:00 EDT -> 13:00 UTC
	t2 := time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC)
	next2, err := CalculateNextReportDelivery(t2, "America/New_York", 9, 0, "daily")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected2 := time.Date(2026, 3, 9, 13, 0, 0, 0, time.UTC)
	if !next2.Equal(expected2) {
		t.Errorf("expected %v during EDT, got %v", expected2, next2)
	}

	// Test weekday filtering: Friday 15:00 UTC -> next delivery Monday
	// Friday April 10, 2026 15:00 UTC (already past 09:00 EDT / 13:00 UTC)
	fri := time.Date(2026, 4, 10, 15, 0, 0, 0, time.UTC)
	nextMon, err := CalculateNextReportDelivery(fri, "America/New_York", 9, 0, "weekdays")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Expected next is Monday, April 13, 2026 at 13:00 UTC (09:00 EDT)
	expectedMon := time.Date(2026, 4, 13, 13, 0, 0, 0, time.UTC)
	if !nextMon.Equal(expectedMon) {
		t.Errorf("expected Monday delivery %v, got %v", expectedMon, nextMon)
	}
}

func TestDailyReport_ReminderLifecycle_SnoozeAndDismiss(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	userID := "usr_reminder_test"
	workspaceID := "ws_reminder_test"

	// Create reminder
	rem, err := service.CreateCustomReminder(ctx, userID, workspaceID, &CareerReminder{
		Title:       "Follow up with Stripe recruiter",
		Description: "Check status of system design interview outcome",
		Priority:    ReminderPriorityHigh,
	})
	if err != nil {
		t.Fatalf("failed to create reminder: %v", err)
	}

	if rem.Status != ReminderStatusPending {
		t.Errorf("expected pending status, got %s", rem.Status)
	}

	// Snooze for 48 hours
	err = service.UpdateReminderStatus(ctx, userID, rem.ID, ReminderStatusSnoozed, 48)
	if err != nil {
		t.Fatalf("failed to snooze reminder: %v", err)
	}

	updated, err := repo.GetReminder(ctx, userID, rem.ID)
	if err != nil {
		t.Fatalf("failed to get reminder: %v", err)
	}
	if updated.Status != ReminderStatusSnoozed {
		t.Errorf("expected status snoozed, got %s", updated.Status)
	}
	if updated.SnoozedUntil == nil || updated.SnoozedUntil.Before(time.Now().Add(47*time.Hour)) {
		t.Errorf("expected SnoozedUntil around 48h in future, got %v", updated.SnoozedUntil)
	}

	// Complete reminder
	err = service.UpdateReminderStatus(ctx, userID, rem.ID, ReminderStatusCompleted, 0)
	if err != nil {
		t.Fatalf("failed to complete reminder: %v", err)
	}

	completed, _ := repo.GetReminder(ctx, userID, rem.ID)
	if completed.Status != ReminderStatusCompleted {
		t.Errorf("expected status completed, got %s", completed.Status)
	}
}

func TestDailyReport_FixtureEvaluation(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "daily_reports_and_reminders.json")
	bytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture file: %v", err)
	}

	var root struct {
		Scenarios []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"scenarios"`
	}

	if err := json.Unmarshal(bytes, &root); err != nil {
		t.Fatalf("failed to parse fixture json: %v", err)
	}

	if len(root.Scenarios) != 4 {
		t.Fatalf("expected 4 scenarios in fixture, found %d", len(root.Scenarios))
	}

	scenarioIDs := make(map[string]bool)
	for _, sc := range root.Scenarios {
		scenarioIDs[sc.ID] = true
	}

	expectedIDs := []string{
		"standard_daily_digest_scenario",
		"dst_timezone_transition_scenario",
		"incomplete_profile_alerts_scenario",
		"reminder_lifecycle_scenario",
	}

	for _, expID := range expectedIDs {
		if !scenarioIDs[expID] {
			t.Errorf("missing expected scenario %s in fixture evaluation", expID)
		}
	}
}
