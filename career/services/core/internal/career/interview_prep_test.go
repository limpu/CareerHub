package career

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type interviewFixtureSuite struct {
	Version   string `json:"version"`
	Domain    string `json:"domain"`
	Scenarios []struct {
		ScenarioID   string `json:"scenario_id"`
		Description  string `json:"description"`
		CompanyInput struct {
			CompanyName      string   `json:"company_name"`
			Domain           string   `json:"domain"`
			Industry         string   `json:"industry"`
			KnownSize        string   `json:"known_size"`
			KnownTechStack   []string `json:"known_tech_stack"`
			Mission          string   `json:"mission"`
			UnverifiedFields []string `json:"unverified_fields"`
		} `json:"company_input,omitempty"`
		JobTarget struct {
			JobID          string   `json:"job_id"`
			Title          string   `json:"title"`
			Department     string   `json:"department"`
			RequiredSkills []string `json:"required_skills"`
		} `json:"job_target,omitempty"`
		CandidateProfile struct {
			CandidateID     string   `json:"candidate_id"`
			ConfirmedSkills []string `json:"confirmed_skills"`
			ConfirmedExperiences []struct {
				Company      string   `json:"company"`
				Role         string   `json:"role"`
				Duration     string   `json:"duration"`
				BulletPoints []string `json:"bullet_points"`
			} `json:"confirmed_experiences"`
		} `json:"candidate_profile,omitempty"`
		ExpectedSkillGaps []string `json:"expected_skill_gaps,omitempty"`
		ApplicationRecordID string `json:"application_record_id,omitempty"`
		RoundSchedule struct {
			RoundID         string `json:"round_id"`
			StageName       string `json:"stage_name"`
			ScheduledAt     string `json:"scheduled_at"`
			Timezone        string `json:"timezone"`
			Format          string `json:"format"`
			MeetingLink     string `json:"meeting_link"`
			InterviewerName string `json:"interviewer_name"`
		} `json:"round_schedule,omitempty"`
		CandidateNotes struct {
			PreInterviewNotes         string   `json:"pre_interview_notes"`
			QuestionsAskedByCandidate []string `json:"questions_asked_by_candidate"`
			PostInterviewReflections  string   `json:"post_interview_reflections"`
			SelfRating                int      `json:"self_rating"`
			FollowUpActions           string   `json:"follow_up_actions"`
		} `json:"candidate_notes,omitempty"`
		SampleApplications []struct {
			AppID       string `json:"app_id"`
			Stage       string `json:"stage"`
			Outcome     string `json:"outcome"`
			DaysInStage int    `json:"days_in_stage"`
		} `json:"sample_applications,omitempty"`
	} `json:"scenarios"`
}

func loadInterviewFixture(t *testing.T) interviewFixtureSuite {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "interview_prep_and_outcomes.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture file: %v", err)
	}

	var suite interviewFixtureSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("failed to unmarshal fixture JSON: %v", err)
	}
	return suite
}

func TestInterviewPrep_FixtureSuiteEvaluation(t *testing.T) {
	suite := loadInterviewFixture(t)
	if len(suite.Scenarios) != 4 {
		t.Fatalf("expected 4 fixture scenarios, got %d", len(suite.Scenarios))
	}

	for _, s := range suite.Scenarios {
		t.Run(s.ScenarioID, func(t *testing.T) {
			switch s.ScenarioID {
			case "full_interview_prep_and_brief":
				brief := GenerateCompanyBrief(
					s.CompanyInput.CompanyName,
					s.CompanyInput.Domain,
					s.CompanyInput.Industry,
					s.CompanyInput.KnownSize,
					s.CompanyInput.Mission,
					s.CompanyInput.KnownTechStack,
					s.CompanyInput.UnverifiedFields,
				)
				if brief.CompanyName != s.CompanyInput.CompanyName {
					t.Errorf("company name mismatch: got %s, want %s", brief.CompanyName, s.CompanyInput.CompanyName)
				}
				if len(brief.UnverifiedFields) != 2 {
					t.Errorf("expected 2 unverified fields, got %d", len(brief.UnverifiedFields))
				}

				profile := &MasterCareerProfile{
					Skills: []SkillItem{
						{Name: "Go"},
						{Name: "PostgreSQL"},
						{Name: "Kafka"},
					},
					Experiences: []ExperienceItem{
						{
							ID:      "exp-01",
							Company: "FastFintech Corp",
							Title:   "Staff Backend Engineer",
						},
					},
				}

				questions := GenerateInterviewQuestions(s.JobTarget.Title, s.JobTarget.RequiredSkills, profile, brief)
				if len(questions) != 4 {
					t.Fatalf("expected 4 generated questions, got %d", len(questions))
				}

				// Check STAR guidance grounded in experience
				behQ := questions[0]
				if behQ.Category != CategoryBehavioralSTAR {
					t.Errorf("expected first question to be STAR, got %s", behQ.Category)
				}
				if behQ.StarGuidance == nil {
					t.Fatal("expected non-nil STAR guidance")
				}
				if len(behQ.GroundedFactIDs) == 0 {
					t.Errorf("expected STAR question to link grounded fact ID")
				}

			case "grounded_questions_with_unmatched_gaps":
				brief := GenerateCompanyBrief(
					s.CompanyInput.CompanyName,
					s.CompanyInput.Domain,
					s.CompanyInput.Industry,
					s.CompanyInput.KnownSize,
					s.CompanyInput.Mission,
					s.CompanyInput.KnownTechStack,
					s.CompanyInput.UnverifiedFields,
				)

				profile := &MasterCareerProfile{
					Skills: []SkillItem{
						{Name: "Go"},
						{Name: "Linux"},
					},
				}

				questions := GenerateInterviewQuestions(s.JobTarget.Title, s.JobTarget.RequiredSkills, profile, brief)
				var techQ *InterviewQuestion
				for i := range questions {
					if questions[i].Category == CategoryTechnical {
						techQ = &questions[i]
						break
					}
				}
				if techQ == nil {
					t.Fatal("expected technical question in pack")
				}

				if len(techQ.IdentifiedSkillGaps) == 0 {
					t.Fatalf("expected identified skill gaps for missing skills per AT-003")
				}
				gapMap := make(map[string]bool)
				for _, g := range techQ.IdentifiedSkillGaps {
					gapMap[g] = true
				}
				for _, expectedGap := range s.ExpectedSkillGaps {
					if !gapMap[expectedGap] {
						t.Errorf("missing expected skill gap: %s", expectedGap)
					}
				}

			case "interview_round_transition_and_notes":
				schedTime, err := time.Parse(time.RFC3339, s.RoundSchedule.ScheduledAt)
				if err != nil {
					t.Fatalf("invalid schedule time: %v", err)
				}

				sched, err := CreateInterviewSchedule(
					s.ApplicationRecordID,
					s.RoundSchedule.RoundID,
					s.RoundSchedule.StageName,
					schedTime,
					s.RoundSchedule.Timezone,
					s.RoundSchedule.Format,
					s.RoundSchedule.MeetingLink,
					s.RoundSchedule.InterviewerName,
				)
				if err != nil {
					t.Fatalf("failed to create schedule: %v", err)
				}
				if sched.RoundID != s.RoundSchedule.RoundID {
					t.Errorf("round ID mismatch: got %s, want %s", sched.RoundID, s.RoundSchedule.RoundID)
				}
				if sched.PrepReminderAt.After(sched.ScheduledAt) {
					t.Errorf("prep reminder must be before scheduled time")
				}

				note, err := RecordSessionNote(
					s.ApplicationRecordID,
					s.RoundSchedule.RoundID,
					"cand-01",
					s.CandidateNotes.PreInterviewNotes,
					s.CandidateNotes.QuestionsAskedByCandidate,
					s.CandidateNotes.PostInterviewReflections,
					s.RoundSchedule.InterviewerName,
					"Principal Architect",
					s.CandidateNotes.SelfRating,
					s.CandidateNotes.FollowUpActions,
				)
				if err != nil {
					t.Fatalf("failed to record session note: %v", err)
				}
				if note.SelfRating != 4 {
					t.Errorf("expected self rating 4, got %d", note.SelfRating)
				}

			case "outcome_funnel_and_conversion_rates":
				var apps []ApplicationRecord
				for _, sample := range s.SampleApplications {
					apps = append(apps, ApplicationRecord{
						ID:        sample.AppID,
						Stage:     ApplicationStage(sample.Stage),
						AppliedAt: time.Now().Add(-time.Duration(sample.DaysInStage*24) * time.Hour),
					})
				}
				funnel := CalculateOutcomeFunnel(apps)
				if funnel.TotalApplications != 8 {
					t.Errorf("expected 8 total applications, got %d", funnel.TotalApplications)
				}
				if funnel.StageCounts[string(StageOffered)] != 1 {
					t.Errorf("expected 1 offer, got %d", funnel.StageCounts[string(StageOffered)])
				}
				if funnel.StageCounts[string(StageRejected)] != 1 {
					t.Errorf("expected 1 rejection, got %d", funnel.StageCounts[string(StageRejected)])
				}
				if funnel.OverallOfferRate <= 0.0 {
					t.Errorf("expected positive overall offer rate, got %f", funnel.OverallOfferRate)
				}
			}
		})
	}
}

func TestInterviewPrep_GroundedQuestionsAndSkillGaps_AT003(t *testing.T) {
	brief := GenerateCompanyBrief("Stripe Inc", "stripe.com", "Fintech", "5000+", "Global payments", []string{"Ruby", "Go"}, []string{"Internal Compensation Bands"})
	profile := &MasterCareerProfile{
		Skills: []SkillItem{
			{Name: "Go"},
		},
		Experiences: []ExperienceItem{
			{
				ID:      "exp-stripe-01",
				Company: "PaymentsCo",
				Title:   "Software Engineer",
			},
		},
	}

	requiredSkills := []string{"Go", "Kubernetes", "gRPC", "Distributed Locking"}
	pack := CreateInterviewPreparationPack("app-123", "cand-001", brief, "Senior Infrastructure Engineer", requiredSkills, profile)

	if pack.PackID == "" {
		t.Fatal("expected valid pack ID")
	}
	if len(pack.PreparationChecklist) == 0 {
		t.Errorf("expected preparation checklist items")
	}

	var techQ *InterviewQuestion
	for i := range pack.Questions {
		if pack.Questions[i].Category == CategoryTechnical {
			techQ = &pack.Questions[i]
			break
		}
	}
	if techQ == nil {
		t.Fatal("expected technical question in pack")
	}

	// Kubernetes, gRPC, and Distributed Locking should be marked as identified gaps per AT-003
	if len(techQ.IdentifiedSkillGaps) != 3 {
		t.Errorf("expected 3 identified skill gaps, got %d (%v)", len(techQ.IdentifiedSkillGaps), techQ.IdentifiedSkillGaps)
	}
}

func TestInterviewPrep_ScheduleAndReminders(t *testing.T) {
	now := time.Now().Add(48 * time.Hour)
	sched, err := CreateInterviewSchedule("app-456", "round-recruiter", "Recruiter Screen", now, "America/New_York", "video", "https://zoom.us/j/12345", "Alice Recruiter")
	if err != nil {
		t.Fatalf("unexpected schedule error: %v", err)
	}

	if sched.Outcome != OutcomePending {
		t.Errorf("initial outcome must be pending, got %s", sched.Outcome)
	}

	// Validation checks
	_, err = CreateInterviewSchedule("app-456", "", "Recruiter Screen", now, "UTC", "video", "", "")
	if err == nil {
		t.Errorf("expected error on missing round_id")
	}
}

func TestInterviewPrep_NotesAndDebriefReflections(t *testing.T) {
	// Valid rating
	note, err := RecordSessionNote("app-456", "round-01", "cand-01", "Review STAR points", []string{"What is the team size?"}, "Good conversation with team lead", "Bob", "Engineering Manager", 5, "Follow up next Tuesday")
	if err != nil {
		t.Fatalf("unexpected session note error: %v", err)
	}
	if note.SelfRating != 5 {
		t.Errorf("expected self rating 5, got %d", note.SelfRating)
	}

	// Invalid rating 0
	_, err = RecordSessionNote("app-456", "round-01", "cand-01", "", nil, "", "", "", 0, "")
	if err == nil {
		t.Errorf("expected error for self rating 0")
	}

	// Invalid rating 6
	_, err = RecordSessionNote("app-456", "round-01", "cand-01", "", nil, "", "", "", 6, "")
	if err == nil {
		t.Errorf("expected error for self rating 6")
	}
}

func TestInterviewPrep_ServiceIntegration(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	brief := GenerateCompanyBrief("Acme Corp", "acme.com", "Tech", "100", "Innovate", []string{"Go"}, nil)
	pack, err := service.GenerateInterviewPrep(ctx, "app-test-1", "user-test-1", brief, "Backend Lead", []string{"Go", "PostgreSQL"})
	if err != nil {
		t.Fatalf("failed to generate interview prep: %v", err)
	}

	retrieved, err := service.GetInterviewPrep(ctx, pack.PackID)
	if err != nil {
		t.Fatalf("failed to retrieve interview prep: %v", err)
	}
	if retrieved.PackID != pack.PackID {
		t.Errorf("pack ID mismatch: got %s, want %s", retrieved.PackID, pack.PackID)
	}

	// Schedule round
	sched, err := service.ScheduleInterviewRound(ctx, "app-test-1", "round-01", "Technical Assessment", time.Now().Add(24*time.Hour), "UTC", "video", "https://meet.google.com/xyz", "Interviewer 1")
	if err != nil {
		t.Fatalf("failed to schedule round: %v", err)
	}

	// List schedules
	schedules, err := service.ListInterviewSchedules(ctx, "app-test-1")
	if err != nil || len(schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d (err: %v)", len(schedules), err)
	}

	// Update round outcome
	if err := service.UpdateRoundOutcome(ctx, sched.ScheduleID, OutcomePassed); err != nil {
		t.Fatalf("failed to update outcome: %v", err)
	}

	// Note recording
	note, err := service.RecordInterviewSessionNote(ctx, "app-test-1", "round-01", "user-test-1", "Pre notes", []string{"Q1?"}, "Post notes", "Interviewer 1", "Staff Eng", 4, "Send thank you")
	if err != nil {
		t.Fatalf("failed to record note: %v", err)
	}

	retrievedNote, err := service.GetInterviewSessionNote(ctx, "round-01")
	if err != nil || retrievedNote.NoteID != note.NoteID {
		t.Fatalf("failed to retrieve session note: %v", err)
	}

	// Outcome funnel
	funnel, err := service.GetOutcomeFunnel(ctx, "user-test-1")
	if err != nil {
		t.Fatalf("failed to get outcome funnel: %v", err)
	}
	if funnel.TotalApplications < 0 {
		t.Errorf("invalid total applications in funnel")
	}
}
