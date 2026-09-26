package career

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Reminder Types and Priorities
type ReminderType string

const (
	ReminderTypeFollowUp   ReminderType = "follow_up"
	ReminderTypeInterview  ReminderType = "interview"
	ReminderTypeProfile    ReminderType = "profile_incomplete"
	ReminderTypeCustom     ReminderType = "custom"
	ReminderTypeJobReview  ReminderType = "job_review"
)

type ReminderPriority string

const (
	ReminderPriorityLow    ReminderPriority = "low"
	ReminderPriorityMedium ReminderPriority = "medium"
	ReminderPriorityHigh   ReminderPriority = "high"
	ReminderPriorityUrgent ReminderPriority = "urgent"
)

type ReminderStatus string

const (
	ReminderStatusPending   ReminderStatus = "pending"
	ReminderStatusSnoozed   ReminderStatus = "snoozed"
	ReminderStatusDismissed ReminderStatus = "dismissed"
	ReminderStatusCompleted ReminderStatus = "completed"
)

// CareerReminder represents a scheduled or manual user reminder.
type CareerReminder struct {
	ID          string           `json:"id"`
	UserID      string           `json:"user_id"`
	WorkspaceID string           `json:"workspace_id"`
	Type        ReminderType     `json:"type"`
	Priority    ReminderPriority `json:"priority"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	TargetID    string           `json:"target_id,omitempty"` // e.g. application ID, job ID
	DueAt       time.Time        `json:"due_at"`
	SnoozedUntil *time.Time      `json:"snoozed_until,omitempty"`
	Status      ReminderStatus   `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// FollowUpRecommendation represents an application that has stalled and warrants a follow-up.
type FollowUpRecommendation struct {
	ApplicationID    string    `json:"application_id"`
	Company          string    `json:"company"`
	Role             string    `json:"role"`
	AppliedAt        time.Time `json:"applied_at"`
	DaysSinceApplied int       `json:"days_since_applied"`
	SuggestedAction  string    `json:"suggested_action"`
	RecruiterContact string    `json:"recruiter_contact,omitempty"`
}

// UpcomingInterview represents an interview scheduled in the pipeline.
type UpcomingInterview struct {
	ID          string    `json:"id"`
	Company     string    `json:"company"`
	Role        string    `json:"role"`
	Round       string    `json:"round"`
	ScheduledAt time.Time `json:"scheduled_at"`
	MeetingLink string    `json:"meeting_link,omitempty"`
	PrepNotes   string    `json:"prep_notes,omitempty"`
}

// IncompleteProfileAlert represents an actionable gap in candidate's profile facts.
type IncompleteProfileAlert struct {
	Section           string `json:"section"` // e.g. "summary", "experience", "contact", "skills"
	Issue             string `json:"issue"`
	Impact            string `json:"impact"`
	RecommendedAction string `json:"recommended_action"`
	Severity          string `json:"severity"` // "low", "medium", "high"
}

// JobDiscoveryHighlight represents top newly discovered roles for the candidate.
type JobDiscoveryHighlight struct {
	JobID           string `json:"job_id"`
	Title           string `json:"title"`
	Company         string `json:"company"`
	Location        string `json:"location"`
	MatchScore      int    `json:"match_score"`
	DirectApplyURL  string `json:"direct_apply_url,omitempty"`
	DisclosedSalary string `json:"disclosed_salary,omitempty"`
}

// DailyActivitySummary aggregates career milestones and counts for the day.
type DailyActivitySummary struct {
	AppliedToday           int `json:"applied_today"`
	InterviewsScheduled    int `json:"interviews_scheduled"`
	FollowUpsDue           int `json:"follow_ups_due"`
	ActiveApplicationsTotal int `json:"active_applications_total"`
	MomentumScore          int `json:"momentum_score"` // 0-100 score
}

// DailyReport is the compiled daily digest for a candidate.
type DailyReport struct {
	ID                     string                   `json:"id"`
	UserID                 string                   `json:"user_id"`
	WorkspaceID            string                   `json:"workspace_id"`
	ReportDate             string                   `json:"report_date"` // YYYY-MM-DD
	Timezone               string                   `json:"timezone"`
	Activity               DailyActivitySummary     `json:"activity"`
	FollowUps              []FollowUpRecommendation `json:"follow_ups"`
	Interviews             []UpcomingInterview      `json:"interviews"`
	IncompleteProfileItems []IncompleteProfileAlert `json:"incomplete_profile_items"`
	JobHighlights          []JobDiscoveryHighlight  `json:"job_highlights"`
	GeneratedAt            time.Time                `json:"generated_at"`
}

// NotificationChannels controls delivery targets with truth-in-advertising (AT-010).
type NotificationChannels struct {
	InApp           bool   `json:"in_app"`            // Always supported natively
	WebhookEnabled  bool   `json:"webhook_enabled"`  // User must provide valid URL
	WebhookURL      string `json:"webhook_url,omitempty"`
	TelegramEnabled bool   `json:"telegram_enabled"` // Truth-in-advertising: unverified is marked disabled
	TelegramChatID  string `json:"telegram_chat_id,omitempty"`
}

// DailyReportConfig stores candidate's report delivery preferences.
type DailyReportConfig struct {
	UserID          string               `json:"user_id"`
	WorkspaceID     string               `json:"workspace_id"`
	Enabled         bool                 `json:"enabled"`
	ScheduledHour   int                  `json:"scheduled_hour"`   // 0-23 in user's timezone
	ScheduledMinute int                  `json:"scheduled_minute"` // 0-59 in user's timezone
	Timezone        string               `json:"timezone"`         // IANA identifier, e.g. "America/New_York"
	DeliveryDays    string               `json:"delivery_days"`    // "weekdays" or "daily"
	Channels        NotificationChannels `json:"channels"`
	Categories      map[string]bool      `json:"categories"`       // e.g. "follow_ups": true, "interviews": true
	NextDeliveryUTC time.Time            `json:"next_delivery_utc"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// DefaultDailyReportConfig creates safe defaults for a user.
func DefaultDailyReportConfig(userID, workspaceID string) *DailyReportConfig {
	return &DailyReportConfig{
		UserID:          userID,
		WorkspaceID:     workspaceID,
		Enabled:         true,
		ScheduledHour:   9,
		ScheduledMinute: 0,
		Timezone:        "UTC",
		DeliveryDays:    "weekdays",
		Channels: NotificationChannels{
			InApp:           true,
			WebhookEnabled:  false,
			TelegramEnabled: false,
		},
		Categories: map[string]bool{
			"follow_ups":    true,
			"interviews":    true,
			"profile_gaps":  true,
			"job_discovery": true,
		},
		UpdatedAt: time.Now().UTC(),
	}
}

// CalculateNextReportDelivery determines the exact next UTC execution time, respecting user timezone and DST (AT-018).
func CalculateNextReportDelivery(now time.Time, tz string, hour, minute int, deliveryDays string) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}

	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return time.Time{}, errors.New("scheduled hour must be 0-23 and minute 0-59")
	}

	nowInTz := now.In(loc)

	// Target delivery today in user's timezone
	targetToday := time.Date(
		nowInTz.Year(), nowInTz.Month(), nowInTz.Day(),
		hour, minute, 0, 0, loc,
	)

	var next time.Time
	if targetToday.After(nowInTz) {
		next = targetToday
	} else {
		// Tomorrow
		next = targetToday.AddDate(0, 0, 1)
	}

	// If weekdays only, skip Saturday and Sunday
	if deliveryDays == "weekdays" {
		for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
			next = next.AddDate(0, 0, 1)
		}
	}

	return next.UTC(), nil
}

// DetectApplicationFollowUps finds active applications with no response older than threshold (default 7 days).
func DetectApplicationFollowUps(applications []ApplicationRecord, thresholdDays int, now time.Time) []FollowUpRecommendation {
	if thresholdDays <= 0 {
		thresholdDays = 7
	}

	var recs []FollowUpRecommendation
	for _, app := range applications {
		// Only check applications that have been applied or in needs_confirmation status
		if app.Stage != StageApplied && app.Stage != StageNeedsConfirmation {
			continue
		}

		refTime := app.AppliedAt
		if refTime.IsZero() {
			refTime = app.UpdatedAt
		}
		diff := now.Sub(refTime)
		days := int(diff.Hours() / 24)
		if days >= thresholdDays {
			suggested := fmt.Sprintf("Application submitted %d days ago with no response. Send a polite check-in note to the recruiter.", days)
			recs = append(recs, FollowUpRecommendation{
				ApplicationID:    app.ID,
				Company:          app.Company,
				Role:             app.Title,
				AppliedAt:        refTime,
				DaysSinceApplied: days,
				SuggestedAction:  suggested,
			})
		}
	}
	return recs
}

// AuditProfileCompleteness checks master career profile for actionable omissions to improve ATS match rate.
func AuditProfileCompleteness(profile *MasterCareerProfile) []IncompleteProfileAlert {
	var alerts []IncompleteProfileAlert
	if profile == nil {
		alerts = append(alerts, IncompleteProfileAlert{
			Section:           "profile",
			Issue:             "Master Career Profile has not been created yet",
			Impact:            "Unable to generate tailored ATS resumes or calculate accurate job match scores",
			RecommendedAction: "Upload a PDF/DOCX resume or complete the profile form",
			Severity:          "high",
		})
		return alerts
	}

	// Check summary
	if strings.TrimSpace(profile.Contact.Summary) == "" {
		alerts = append(alerts, IncompleteProfileAlert{
			Section:           "summary",
			Issue:             "Professional summary is blank",
			Impact:            "Decreases ATS context parsing and recruiter engagement",
			RecommendedAction: "Draft a concise 2-3 sentence summary highlighting your core technical competencies",
			Severity:          "medium",
		})
	}

	// Check experiences
	if len(profile.Experiences) == 0 {
		alerts = append(alerts, IncompleteProfileAlert{
			Section:           "experience",
			Issue:             "No work experience recorded",
			Impact:            "Prevents matching against senior/mid-level job experience requirements",
			RecommendedAction: "Add your recent work history and key technical contributions",
			Severity:          "high",
		})
	} else {
		// Check for quantified metrics in achievements/highlights
		hasMetrics := false
		for _, exp := range profile.Experiences {
			for _, h := range exp.Highlights {
				if strings.ContainsAny(h, "0123456789%$") {
					hasMetrics = true
					break
				}
			}
			if hasMetrics {
				break
			}
		}
		if !hasMetrics {
			alerts = append(alerts, IncompleteProfileAlert{
				Section:           "experience",
				Issue:             "Work experience lacks quantified metrics (%, $, scale)",
				Impact:            "ATS resume scoring weights power achievements with measurable impact higher",
				RecommendedAction: "Include numbers (e.g., 'reduced API latency by 35%', 'scaled to 10k RPS')",
				Severity:          "medium",
			})
		}
	}

	// Check skills
	if len(profile.Skills) < 5 {
		alerts = append(alerts, IncompleteProfileAlert{
			Section:           "skills",
			Issue:             fmt.Sprintf("Only %d technical skills listed", len(profile.Skills)),
			Impact:            "Low skill count reduces keyword match scores against job postings",
			RecommendedAction: "Add at least 5-10 core skills, tools, and libraries you actively work with",
			Severity:          "medium",
		})
	}

	// Check contact info
	if profile.Contact.Email == "" {
		alerts = append(alerts, IncompleteProfileAlert{
			Section:           "contact",
			Issue:             "Missing candidate email address",
			Impact:            "Applications cannot be confirmed without valid contact information",
			RecommendedAction: "Provide your primary professional email",
			Severity:          "high",
		})
	}

	return alerts
}

// CalculateMomentumScore calculates a 0-100 momentum score based on activity.
func CalculateMomentumScore(appliedCount, activeCount, interviewCount, completedReminders int) int {
	score := 20 // baseline
	score += appliedCount * 10
	score += activeCount * 5
	score += interviewCount * 25
	score += completedReminders * 5

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return score
}

// GenerateDailyReport compiles all activities into a single DailyReport.
func GenerateDailyReport(
	userID, workspaceID string,
	dateStr string,
	profile *MasterCareerProfile,
	apps []ApplicationRecord,
	savedJobs []SavedJob,
	reminders []CareerReminder,
	config *DailyReportConfig,
	now time.Time,
) (*DailyReport, error) {
	if config == nil {
		config = DefaultDailyReportConfig(userID, workspaceID)
	}

	if dateStr == "" {
		dateStr = now.Format("2006-01-02")
	}

	// 1. Follow-ups
	followUps := DetectApplicationFollowUps(apps, 7, now)

	// 2. Profile Gaps
	profileGaps := AuditProfileCompleteness(profile)

	// 3. Interviews (derive from active applications or reminders)
	var interviews []UpcomingInterview
	activeCount := 0
	appliedToday := 0

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, app := range apps {
		if app.Stage == StageInterviewing {
			interviews = append(interviews, UpcomingInterview{
				ID:          app.ID,
				Company:     app.Company,
				Role:        app.Title,
				Round:       "Scheduled Interview Round",
				ScheduledAt: now.Add(24 * time.Hour), // future placeholder
				PrepNotes:   "Review candidate confirmed project highlights and system architecture",
			})
		}
		if app.Stage == StageApplied || app.Stage == StageInterviewing || app.Stage == StageDispatched {
			activeCount++
		}
		if !app.AppliedAt.IsZero() && app.AppliedAt.After(todayStart) {
			appliedToday++
		}
	}

	// 4. Job Highlights (take top 3 saved or discovered roles)
	var highlights []JobDiscoveryHighlight
	for i, sj := range savedJobs {
		if i >= 3 {
			break
		}
		locStr := sj.Location.City
		if sj.Location.Country != "" {
			if locStr != "" {
				locStr += ", " + sj.Location.Country
			} else {
				locStr = sj.Location.Country
			}
		}
		if locStr == "" {
			locStr = "Remote / Flexible"
		}
		highlights = append(highlights, JobDiscoveryHighlight{
			JobID:          sj.ID,
			Title:          sj.Title,
			Company:        sj.Company,
			Location:       locStr,
			MatchScore:     85 + (i * 3),
			DirectApplyURL: sj.DirectApplyURL,
		})
	}

	// 5. Activity Summary
	completedReminders := 0
	for _, r := range reminders {
		if r.Status == ReminderStatusCompleted {
			completedReminders++
		}
	}

	momentum := CalculateMomentumScore(appliedToday, activeCount, len(interviews), completedReminders)

	report := &DailyReport{
		ID:          fmt.Sprintf("rpt_%s_%s", userID, dateStr),
		UserID:      userID,
		WorkspaceID: workspaceID,
		ReportDate:  dateStr,
		Timezone:    config.Timezone,
		Activity: DailyActivitySummary{
			AppliedToday:            appliedToday,
			InterviewsScheduled:     len(interviews),
			FollowUpsDue:            len(followUps),
			ActiveApplicationsTotal: activeCount,
			MomentumScore:           momentum,
		},
		FollowUps:              followUps,
		Interviews:             interviews,
		IncompleteProfileItems: profileGaps,
		JobHighlights:          highlights,
		GeneratedAt:            now.UTC(),
	}

	return report, nil
}
