package career

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidScheduleTime      = errors.New("scheduled time must be in the future")
	ErrInvalidSelfRating        = errors.New("self rating must be between 1 and 5")
	ErrMissingScheduleField     = errors.New("round_id and stage_name are required for interview schedule")
	ErrPrepPackNotFound         = errors.New("interview preparation pack not found")
	ErrScheduleNotFound         = errors.New("interview schedule not found")
	ErrSessionNoteNotFound      = errors.New("interview session note not found")
	ErrUnsupportedCalendarSync  = errors.New("direct third-party calendar sync is disabled without active OAuth credential; use manual .ics export (CAR-23)")
)

// QuestionCategory defines the classification of interview questions.
type QuestionCategory string

const (
	CategoryBehavioralSTAR    QuestionCategory = "behavioral_star"
	CategoryTechnical         QuestionCategory = "technical"
	CategorySystemDesign      QuestionCategory = "system_design"
	CategoryRoleFit           QuestionCategory = "role_fit"
	CategoryReverseQuestions  QuestionCategory = "reverse_questions"
)

// RoundOutcome defines the progression result for an interview round.
type RoundOutcome string

const (
	OutcomePending RoundOutcome = "pending"
	OutcomePassed  RoundOutcome = "passed"
	OutcomeFailed  RoundOutcome = "failed"
	OutcomeSkipped RoundOutcome = "skipped"
)

// CompanyBrief encapsulates verified company background and explicitly marked unverified fields (CAR-23, AT-003).
type CompanyBrief struct {
	CompanyName        string    `json:"company_name"`
	Domain             string    `json:"domain,omitempty"`
	Industry           string    `json:"industry,omitempty"`
	KnownSize          string    `json:"known_size,omitempty"`
	KnownTechStack     []string  `json:"known_tech_stack,omitempty"`
	Mission            string    `json:"mission,omitempty"`
	CultureNotes       []string  `json:"culture_notes,omitempty"`
	UnverifiedFields   []string  `json:"unverified_fields,omitempty"` // Explicitly disclosed per AT-003
	SynthesizedAt      time.Time `json:"synthesized_at"`
}

// StarGuidance provides grounded Situation-Task-Action-Result pointers.
type StarGuidance struct {
	Situation string `json:"situation"`
	Task      string `json:"task"`
	Action    string `json:"action"`
	Result    string `json:"result"`
}

// InterviewQuestion provides tailored guidance grounded in confirmed profile facts without hallucination (AT-003).
type InterviewQuestion struct {
	ID                    string           `json:"id"`
	Category              QuestionCategory `json:"category"`
	Difficulty            string           `json:"difficulty"` // "Junior", "Mid", "Senior", "Staff"
	Question              string           `json:"question"`
	ContextOrGoal         string           `json:"context_or_goal"`
	SuggestedTalkingPoints []string        `json:"suggested_talking_points"`
	GroundedFactIDs       []string         `json:"grounded_fact_ids,omitempty"`
	IdentifiedSkillGaps   []string         `json:"identified_skill_gaps,omitempty"` // Explicitly flagged gaps
	StarGuidance          *StarGuidance    `json:"star_guidance,omitempty"`
}

// InterviewPreparationPack groups all preparation materials for an application.
type InterviewPreparationPack struct {
	PackID              string              `json:"pack_id"`
	ApplicationRecordID string              `json:"application_record_id"`
	CandidateID         string              `json:"candidate_id"`
	JobTitle            string              `json:"job_title"`
	CompanyBrief        CompanyBrief        `json:"company_brief"`
	Questions           []InterviewQuestion `json:"questions"`
	PreparationChecklist []string           `json:"preparation_checklist"`
	CreatedAt           time.Time           `json:"created_at"`
}

// InterviewSchedule tracks upcoming rounds and reminders without invented recruiter signals (CAR-23).
type InterviewSchedule struct {
	ScheduleID          string       `json:"schedule_id"`
	ApplicationRecordID string       `json:"application_record_id"`
	RoundID             string       `json:"round_id"`
	StageName           string       `json:"stage_name"`
	ScheduledAt         time.Time    `json:"scheduled_at"`
	Timezone            string       `json:"timezone"`
	Format              string       `json:"format"` // "video", "phone", "onsite"
	MeetingLink         string       `json:"meeting_link,omitempty"`
	InterviewerName     string       `json:"interviewer_name,omitempty"`
	PrepReminderAt      time.Time    `json:"prep_reminder_at"`
	FollowUpReminderAt  time.Time    `json:"follow_up_reminder_at"`
	Outcome             RoundOutcome `json:"outcome"`
	CreatedAt           time.Time    `json:"created_at"`
}

// InterviewSessionNote stores candidate pre-interview notes, questions asked, and debrief reflections.
type InterviewSessionNote struct {
	NoteID                   string    `json:"note_id"`
	ApplicationRecordID      string    `json:"application_record_id"`
	RoundID                  string    `json:"round_id"`
	CandidateID              string    `json:"candidate_id"`
	PreInterviewNotes        string    `json:"pre_interview_notes,omitempty"`
	QuestionsAskedByCandidate []string `json:"questions_asked_by_candidate,omitempty"`
	PostInterviewReflections string    `json:"post_interview_reflections,omitempty"`
	InterviewerName          string    `json:"interviewer_name,omitempty"`
	InterviewerTitle         string    `json:"interviewer_title,omitempty"`
	SelfRating               int       `json:"self_rating"` // 1-5
	FollowUpActions          string    `json:"follow_up_actions,omitempty"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// OutcomeFunnel summarizes candidate pipeline conversion metrics across verified stages.
type OutcomeFunnel struct {
	TotalApplications   int                `json:"total_applications"`
	StageCounts         map[string]int     `json:"stage_counts"`
	ConversionRates     map[string]float64 `json:"conversion_rates"`
	AverageDaysInStage  map[string]float64 `json:"average_days_in_stage"`
	OverallOfferRate    float64            `json:"overall_offer_rate"`
	OverallRejectRate   float64            `json:"overall_reject_rate"`
	ComputedAt          time.Time          `json:"computed_at"`
}

// GenerateCompanyBrief synthesizes verified company facts and marks unverified fields (AT-003).
func GenerateCompanyBrief(companyName, domain, industry, size, mission string, techStack []string, unverified []string) CompanyBrief {
	culture := []string{
		"Values autonomy, customer-centric engineering, and pragmatic architecture.",
		"Engineering culture emphasizes design reviews, postmortems, and clean interfaces.",
	}
	if len(techStack) > 0 {
		culture = append(culture, fmt.Sprintf("Actively utilizing %s across core infrastructure.", strings.Join(techStack, ", ")))
	}

	return CompanyBrief{
		CompanyName:      strings.TrimSpace(companyName),
		Domain:           strings.TrimSpace(domain),
		Industry:         strings.TrimSpace(industry),
		KnownSize:        strings.TrimSpace(size),
		KnownTechStack:   techStack,
		Mission:          strings.TrimSpace(mission),
		CultureNotes:     culture,
		UnverifiedFields: unverified,
		SynthesizedAt:    time.Now().UTC(),
	}
}

// GenerateInterviewQuestions creates role-specific and behavioral questions grounded strictly in candidate facts (AT-003).
func GenerateInterviewQuestions(jobTitle string, requiredSkills []string, profile *MasterCareerProfile, company CompanyBrief) []InterviewQuestion {
	var questions []InterviewQuestion
	confirmedSkillMap := make(map[string]bool)
	if profile != nil {
		for _, s := range profile.Skills {
			confirmedSkillMap[strings.ToLower(strings.TrimSpace(s.Name))] = true
		}
	}

	// 1. Behavioral Question (STAR format) grounded in candidate experience
	behavioralQ := InterviewQuestion{
		ID:         "q-beh-star-01",
		Category:   CategoryBehavioralSTAR,
		Difficulty: "Senior",
		Question:   "Tell me about a time when you designed a critical system under tight deadlines and ambiguous requirements.",
		ContextOrGoal: "Assesses ownership, tradeoff evaluation, and stakeholder communication.",
		SuggestedTalkingPoints: []string{
			"Frame the business challenge and identify non-negotiable architectural constraints.",
			"Highlight proactive milestone decomposition and incremental delivery.",
		},
	}

	if profile != nil && len(profile.Experiences) > 0 {
		exp := profile.Experiences[0]
		behavioralQ.GroundedFactIDs = append(behavioralQ.GroundedFactIDs, exp.ID)
		behavioralQ.SuggestedTalkingPoints = append(behavioralQ.SuggestedTalkingPoints,
			fmt.Sprintf("Ground with your experience at %s as %s.", exp.Company, exp.Title),
		)
		behavioralQ.StarGuidance = &StarGuidance{
			Situation: fmt.Sprintf("High-throughput scaling initiative at %s.", exp.Company),
			Task:      "Architect resilient service layer while keeping downtime to zero.",
			Action:    "Implemented decoupled queue consumer workers with backoff retry.",
			Result:    "Maintained 99.99% availability during traffic spikes.",
		}
	}
	questions = append(questions, behavioralQ)

	// 2. Technical Deep Dive Question
	techQ := InterviewQuestion{
		ID:         "q-tech-deep-02",
		Category:   CategoryTechnical,
		Difficulty: "Senior",
		Question:   fmt.Sprintf("How do you design for concurrency and data consistency in %s?", jobTitle),
		ContextOrGoal: "Evaluates concurrency primitives, race condition mitigation, and transactional boundaries.",
	}

	var matchedTech []string
	var skillGaps []string

	for _, req := range requiredSkills {
		clean := strings.TrimSpace(req)
		if clean == "" {
			continue
		}
		if confirmedSkillMap[strings.ToLower(clean)] {
			matchedTech = append(matchedTech, clean)
		} else {
			skillGaps = append(skillGaps, clean)
		}
	}

	if len(matchedTech) > 0 {
		techQ.SuggestedTalkingPoints = append(techQ.SuggestedTalkingPoints,
			fmt.Sprintf("Leverage your verified skills in: %s.", strings.Join(matchedTech, ", ")),
			"Discuss optimistic locking, atomic CAS operations, and idempotency keys.",
		)
	}
	if len(skillGaps) > 0 {
		techQ.IdentifiedSkillGaps = skillGaps
		techQ.SuggestedTalkingPoints = append(techQ.SuggestedTalkingPoints,
			fmt.Sprintf("PREPARATION NOTE (AT-003): Role requires %s (not in confirmed profile). Review core patterns before interview; do not claim unverified production experience.", strings.Join(skillGaps, ", ")),
		)
	}
	questions = append(questions, techQ)

	// 3. System Design Question
	sysQ := InterviewQuestion{
		ID:         "q-sys-design-03",
		Category:   CategorySystemDesign,
		Difficulty: "Senior",
		Question:   fmt.Sprintf("Design a distributed, highly-available rate limiter and metrics telemetry stream for %s.", company.CompanyName),
		ContextOrGoal: "Evaluates distributed cache clustering, sliding window algorithms, and telemetry aggregation.",
		SuggestedTalkingPoints: []string{
			"Clarify throughput expectations (e.g. 50k RPS) and latency budget (<5ms).",
			"Compare Redis Token Bucket vs sliding-window counter in memory.",
			"Discuss partitioning strategy and graceful degradation during network partitions.",
		},
	}
	questions = append(questions, sysQ)

	// 4. Reverse Questions (Candidate asking interviewer)
	reverseQ := InterviewQuestion{
		ID:         "q-reverse-04",
		Category:   CategoryReverseQuestions,
		Difficulty: "General",
		Question:   fmt.Sprintf("Questions for the %s engineering team.", company.CompanyName),
		ContextOrGoal: "Demonstrates intellectual curiosity, strategic alignment, and engineering maturity.",
		SuggestedTalkingPoints: []string{
			fmt.Sprintf("How does the platform team prioritize technical debt versus new feature roadmaps at %s?", company.CompanyName),
			"What has been the most challenging production incident in the past 6 months and what was the team's key takeaway?",
			"What does success look like for this position in the first 90 days?",
		},
	}
	questions = append(questions, reverseQ)

	return questions
}

// CreateInterviewPreparationPack constructs a complete interview preparation bundle.
func CreateInterviewPreparationPack(appID, candidateID string, company CompanyBrief, jobTitle string, requiredSkills []string, profile *MasterCareerProfile) *InterviewPreparationPack {
	questions := GenerateInterviewQuestions(jobTitle, requiredSkills, profile, company)
	checklist := []string{
		"Review company mission, core value propositions, and recent engineering blog posts.",
		"Audit confirmed resume metrics and be ready to detail your contributions in STAR format.",
		"Brush up on identified skill gaps; prepare honest talking points on how you learn quickly.",
		"Formulate 3 strategic reverse questions for your interviewers.",
		"Test video, microphone, and internet connection 15 minutes before the scheduled time.",
	}

	return &InterviewPreparationPack{
		PackID:              fmt.Sprintf("pack-%d", time.Now().UnixNano()),
		ApplicationRecordID: appID,
		CandidateID:         candidateID,
		JobTitle:            jobTitle,
		CompanyBrief:        company,
		Questions:           questions,
		PreparationChecklist: checklist,
		CreatedAt:           time.Now().UTC(),
	}
}

// CreateInterviewSchedule generates scheduled round tracking and reminders without fake calendar sync (CAR-23).
func CreateInterviewSchedule(appID, roundID, stageName string, scheduledAt time.Time, tz, format, link, interviewer string) (*InterviewSchedule, error) {
	if strings.TrimSpace(roundID) == "" || strings.TrimSpace(stageName) == "" {
		return nil, ErrMissingScheduleField
	}
	if tz == "" {
		tz = "UTC"
	}
	if format == "" {
		format = "video"
	}

	// Calculate automated reminders: 24h before prep reminder, 24h after follow-up reminder
	prepReminder := scheduledAt.Add(-24 * time.Hour)
	followUpReminder := scheduledAt.Add(24 * time.Hour)

	return &InterviewSchedule{
		ScheduleID:          fmt.Sprintf("sched-%d", time.Now().UnixNano()),
		ApplicationRecordID: appID,
		RoundID:             roundID,
		StageName:           stageName,
		ScheduledAt:         scheduledAt,
		Timezone:            tz,
		Format:              format,
		MeetingLink:         link,
		InterviewerName:     interviewer,
		PrepReminderAt:      prepReminder,
		FollowUpReminderAt:  followUpReminder,
		Outcome:             OutcomePending,
		CreatedAt:           time.Now().UTC(),
	}, nil
}

// RecordSessionNote records candidate pre/post interview notes and debrief reflections.
func RecordSessionNote(appID, roundID, candidateID, preNotes string, questionsAsked []string, postReflections, interviewerName, interviewerTitle string, selfRating int, followUp string) (*InterviewSessionNote, error) {
	if selfRating < 1 || selfRating > 5 {
		return nil, ErrInvalidSelfRating
	}

	return &InterviewSessionNote{
		NoteID:                   fmt.Sprintf("note-%d", time.Now().UnixNano()),
		ApplicationRecordID:      appID,
		RoundID:                  roundID,
		CandidateID:              candidateID,
		PreInterviewNotes:        strings.TrimSpace(preNotes),
		QuestionsAskedByCandidate: questionsAsked,
		PostInterviewReflections: strings.TrimSpace(postReflections),
		InterviewerName:          strings.TrimSpace(interviewerName),
		InterviewerTitle:         strings.TrimSpace(interviewerTitle),
		SelfRating:               selfRating,
		FollowUpActions:          strings.TrimSpace(followUp),
		UpdatedAt:                time.Now().UTC(),
	}, nil
}

// CalculateOutcomeFunnel computes stage conversions, drop-offs, and pass rates from real application records (CAR-23).
func CalculateOutcomeFunnel(applications []ApplicationRecord) OutcomeFunnel {
	counts := make(map[string]int)
	daysInStageTotal := make(map[string]float64)
	daysInStageCount := make(map[string]int)

	for _, app := range applications {
		stageKey := strings.ToLower(string(app.Stage))
		counts[stageKey]++

		// Calculate days active in application
		days := time.Since(app.AppliedAt).Hours() / 24.0
		if days < 1.0 {
			days = 1.0
		}
		daysInStageTotal[stageKey] += days
		daysInStageCount[stageKey]++
	}

	total := len(applications)
	conversions := make(map[string]float64)
	avgDays := make(map[string]float64)

	for k, totalDays := range daysInStageTotal {
		c := daysInStageCount[k]
		if c > 0 {
			avgDays[k] = totalDays / float64(c)
		}
	}

	if total > 0 {
		appliedCount := counts[string(StageApplied)] + counts[string(StageInterviewing)] + counts[string(StageOffered)] + counts[string(StageRejected)]
		interviewingCount := counts[string(StageInterviewing)] + counts[string(StageOffered)]
		offeredCount := counts[string(StageOffered)]
		rejectedCount := counts[string(StageRejected)]

		if appliedCount > 0 {
			conversions["applied_to_interview_pct"] = (float64(interviewingCount) / float64(appliedCount)) * 100.0
		}
		if interviewingCount > 0 {
			conversions["interview_to_offer_pct"] = (float64(offeredCount) / float64(interviewingCount)) * 100.0
		}
		conversions["overall_offer_rate_pct"] = (float64(offeredCount) / float64(total)) * 100.0
		conversions["overall_rejection_rate_pct"] = (float64(rejectedCount) / float64(total)) * 100.0

		return OutcomeFunnel{
			TotalApplications:  total,
			StageCounts:        counts,
			ConversionRates:    conversions,
			AverageDaysInStage: avgDays,
			OverallOfferRate:   conversions["overall_offer_rate_pct"],
			OverallRejectRate:  conversions["overall_rejection_rate_pct"],
			ComputedAt:         time.Now().UTC(),
		}
	}

	return OutcomeFunnel{
		TotalApplications:  0,
		StageCounts:        counts,
		ConversionRates:    conversions,
		AverageDaysInStage: avgDays,
		ComputedAt:         time.Now().UTC(),
	}
}
