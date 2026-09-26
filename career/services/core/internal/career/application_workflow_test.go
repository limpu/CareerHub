package career

import (
	"context"
	"testing"
	"time"
)

// Helper to create basic test setup
func setupWorkflowTest(t *testing.T) (*CareerService, string, DiscoveredJob) {
	repo := NewMemoryCareerRepository()
	svc := NewCareerService(repo)
	userID := "usr_workflow_test"
	ctx := context.Background()

	// Initialize user profile
	profile := NewMasterCareerProfile(userID)
	profile.Contact.FullName = "Alex Morgan"
	profile.Contact.Email = "alex.morgan@example.com"
	profile.Contact.Phone = "+1 (555) 019-2834"
	profile.Contact.Location = "San Francisco, CA"
	profile.Experiences = []ExperienceItem{
		{
			ID:          "exp_1",
			Title:       "Senior Software Engineer",
			Company:     "TechCorp",
			Description: "Engineered scalable Go microservices and Next.js applications.",
			IsCurrent:   true,
		},
		{
			ID:          "exp_2",
			Title:       "Software Engineer",
			Company:     "InnovateLab",
			Description: "Built distributed backend systems.",
			IsCurrent:   false,
		},
	}
	profile.Skills = []SkillItem{
		{Name: "Go"},
		{Name: "TypeScript"},
		{Name: "PostgreSQL"},
	}
	_ = repo.SaveProfile(ctx, profile)

	job := DiscoveredJob{
		ID:             "job_workflow_target_1",
		Source:         BoardSourceGreenhouse,
		SourceJobID:    "gh_12345",
		CanonicalURL:   "https://boards.greenhouse.io/techcorp/jobs/12345",
		DirectApplyURL: "https://boards.greenhouse.io/techcorp/jobs/12345#apply",
		Title:          "Lead Backend Engineer",
		Company:        "TechCorp International",
		Location: JobLocation{
			City:     "Berlin",
			Country:  "DE",
			IsRemote: true,
		},
		JobType:        "full_time",
		RequiredSkills: []string{"Go", "PostgreSQL", "Docker"},
	}

	return svc, userID, job
}

func TestWorkflow_ZeroFabrication_NeedsInput_AT003(t *testing.T) {
	svc, userID, job := setupWorkflowTest(t)
	ctx := context.Background()

	// 1. Prepare workflow session
	session, err := svc.PrepareApplicationWorkflow(ctx, userID, "ws_1", job, "", "", ExecutionModeNativeManual)
	if err != nil {
		t.Fatalf("unexpected error preparing session: %v", err)
	}

	// 2. Invariant AT-003: Unconfirmed legal/sponsorship questions MUST have needs_input = true
	// and session must initially be in draft or ready_for_review with unresolved flags
	var authQ, sponsorQ *ApplicationQuestionAnswer
	for i := range session.Bundle.Questions {
		q := &session.Bundle.Questions[i]
		if q.QuestionID == "work_authorization" {
			authQ = q
		}
		if q.QuestionID == "visa_sponsorship" {
			sponsorQ = q
		}
	}

	if authQ == nil || !authQ.NeedsInput || authQ.AnswerValue != "" {
		t.Errorf("expected work_authorization to require candidate input without guessing, got %+v", authQ)
	}
	if sponsorQ == nil || !sponsorQ.NeedsInput || sponsorQ.AnswerValue != "" {
		t.Errorf("expected visa_sponsorship to require candidate input without guessing, got %+v", sponsorQ)
	}

	// 3. Attempting to approve while required questions need input MUST fail
	err = svc.workflowEngine.ApproveSession(session, userID)
	if err == nil {
		t.Errorf("expected error approving session with unresolved questions (AT-003), but got nil")
	}

	// 4. Resolve questions explicitly
	authQ.AnswerValue = "Yes"
	authQ.IsConfirmed = true
	authQ.NeedsInput = false

	sponsorQ.AnswerValue = "No"
	sponsorQ.IsConfirmed = true
	sponsorQ.NeedsInput = false

	updatedSession, _, err := svc.UpdateApplicationWorkflowAnswers(ctx, userID, session.ID, session.Bundle.Questions)
	if err != nil {
		t.Fatalf("unexpected error updating answers: %v", err)
	}

	// 5. Now approval must succeed
	approvedSession, err := svc.ApproveApplicationWorkflow(ctx, userID, updatedSession.ID, userID)
	if err != nil {
		t.Fatalf("expected approval to succeed once all questions resolved: %v", err)
	}
	if approvedSession.Status != WorkflowStatusApproved || approvedSession.ApprovalToken == "" {
		t.Errorf("expected status approved with non-empty approval token, got status=%s, token=%s", approvedSession.Status, approvedSession.ApprovalToken)
	}
}

func TestWorkflow_EditInvalidatesApproval_AT007(t *testing.T) {
	svc, userID, job := setupWorkflowTest(t)
	ctx := context.Background()

	session, err := svc.PrepareApplicationWorkflow(ctx, userID, "ws_1", job, "", "", ExecutionModeAssistantPreFill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Resolve all required fields
	for i := range session.Bundle.Questions {
		q := &session.Bundle.Questions[i]
		if q.NeedsInput {
			q.AnswerValue = "Confirmed Answer"
			q.IsConfirmed = true
			q.NeedsInput = false
		}
	}
	_, _, err = svc.UpdateApplicationWorkflowAnswers(ctx, userID, session.ID, session.Bundle.Questions)
	if err != nil {
		t.Fatalf("unexpected error updating answers: %v", err)
	}

	// Approve session
	approvedSession, err := svc.ApproveApplicationWorkflow(ctx, userID, session.ID, userID)
	if err != nil {
		t.Fatalf("failed to approve: %v", err)
	}
	originalToken := approvedSession.ApprovalToken

	// Now modify an answer (AT-007)
	modifiedQuestions := make([]ApplicationQuestionAnswer, len(approvedSession.Bundle.Questions))
	copy(modifiedQuestions, approvedSession.Bundle.Questions)
	modifiedQuestions[0].AnswerValue = "Tampered Name"

	afterEditSession, invalidated, err := svc.UpdateApplicationWorkflowAnswers(ctx, userID, session.ID, modifiedQuestions)
	if err != nil {
		t.Fatalf("failed to update modified answers: %v", err)
	}

	if !invalidated {
		t.Errorf("expected approval to be flagged as invalidated upon answer modification (AT-007)")
	}
	if afterEditSession.Status != WorkflowStatusReadyForReview {
		t.Errorf("expected status to reset to ready_for_review, got %s", afterEditSession.Status)
	}
	if afterEditSession.ApprovalToken != "" {
		t.Errorf("expected approval token to be cleared, got %s", afterEditSession.ApprovalToken)
	}

	// Trying to dispatch with invalidated state MUST be rejected
	err = svc.workflowEngine.DispatchSession(afterEditSession, ExecutionModeAssistantPreFill)
	if err == nil {
		t.Errorf("expected dispatch to fail without valid approval token")
	}

	// Re-validating old token must fail
	if svc.workflowEngine.ValidateApprovalToken(afterEditSession.ID, afterEditSession.Bundle.BundleChecksumSHA256, originalToken) {
		t.Errorf("expected old approval token to be cryptographically invalid for modified bundle (AT-007)")
	}
}

func TestWorkflow_ClickApplyDoesNotMarkApplied_AT005(t *testing.T) {
	svc, userID, job := setupWorkflowTest(t)
	ctx := context.Background()

	session, _ := svc.PrepareApplicationWorkflow(ctx, userID, "ws_1", job, "", "", ExecutionModeNativeManual)

	for i := range session.Bundle.Questions {
		session.Bundle.Questions[i].AnswerValue = "Yes"
		session.Bundle.Questions[i].IsConfirmed = true
		session.Bundle.Questions[i].NeedsInput = false
	}
	_, _, _ = svc.UpdateApplicationWorkflowAnswers(ctx, userID, session.ID, session.Bundle.Questions)
	_, _ = svc.ApproveApplicationWorkflow(ctx, userID, session.ID, userID)

	// Invariant AT-005: Dispatching / clicking Apply must NEVER mark status as 'applied'
	dispatchedSession, err := svc.DispatchApplicationWorkflow(ctx, userID, session.ID, ExecutionModeNativeManual)
	if err != nil {
		t.Fatalf("failed to dispatch application: %v", err)
	}

	if dispatchedSession.Status == WorkflowStatusApplied {
		t.Fatalf("VIOLATION AT-005: dispatching or clicking Apply marked job as applied directly!")
	}
	if dispatchedSession.Status != WorkflowStatusDispatched {
		t.Errorf("expected status 'dispatched', got '%s'", dispatchedSession.Status)
	}
	if dispatchedSession.DispatchedAt == nil {
		t.Errorf("expected DispatchedAt timestamp to be recorded")
	}
}

func TestWorkflow_TimeoutEntersNeedsConfirmation_AT005(t *testing.T) {
	svc, userID, job := setupWorkflowTest(t)
	ctx := context.Background()

	session, _ := svc.PrepareApplicationWorkflow(ctx, userID, "ws_1", job, "", "", ExecutionModeAssistantPreFill)
	for i := range session.Bundle.Questions {
		session.Bundle.Questions[i].AnswerValue = "Confirmed"
		session.Bundle.Questions[i].IsConfirmed = true
		session.Bundle.Questions[i].NeedsInput = false
	}
	_, _, _ = svc.UpdateApplicationWorkflowAnswers(ctx, userID, session.ID, session.Bundle.Questions)
	_, _ = svc.ApproveApplicationWorkflow(ctx, userID, session.ID, userID)
	dispatchedSession, _ := svc.DispatchApplicationWorkflow(ctx, userID, session.ID, ExecutionModeAssistantPreFill)

	// Simulate elapsed timeout
	pastTime := time.Now().UTC().Add(-15 * time.Minute)
	dispatchedSession.DispatchedAt = &pastTime
	dispatchedSession.TimeoutDurationSecs = 600 // 10 minutes timeout
	_ = svc.repo.UpdateReviewSession(ctx, dispatchedSession)

	timedOutSession, timedOut, err := svc.CheckApplicationWorkflowTimeout(ctx, userID, session.ID)
	if err != nil {
		t.Fatalf("failed checking timeout: %v", err)
	}

	if !timedOut {
		t.Errorf("expected session to be marked as timed out after 15 minutes")
	}
	if timedOutSession.Status != WorkflowStatusNeedsConfirmation {
		t.Errorf("expected status to transition to 'needs_confirmation' (AT-005), got '%s'", timedOutSession.Status)
	}
}

func TestWorkflow_ExplicitConfirmationMarksApplied_REQ005_AT005(t *testing.T) {
	svc, userID, job := setupWorkflowTest(t)
	ctx := context.Background()

	// Shortlist job first in SavedJobs ledger
	_, _ = svc.SaveJob(ctx, userID, SaveJobInput{
		Job:    job,
		Status: SavedJobStatusReadyToApply,
		Notes:  "Ready to apply for this target",
	})

	session, _ := svc.PrepareApplicationWorkflow(ctx, userID, "ws_1", job, "", "", ExecutionModeNativeManual)
	for i := range session.Bundle.Questions {
		session.Bundle.Questions[i].AnswerValue = "Yes"
		session.Bundle.Questions[i].IsConfirmed = true
		session.Bundle.Questions[i].NeedsInput = false
	}
	_, _, _ = svc.UpdateApplicationWorkflowAnswers(ctx, userID, session.ID, session.Bundle.Questions)
	_, _ = svc.ApproveApplicationWorkflow(ctx, userID, session.ID, userID)
	_, _ = svc.DispatchApplicationWorkflow(ctx, userID, session.ID, ExecutionModeNativeManual)

	// Explicit user confirmation with receipt (REQ-005, AT-005)
	receipt := SubmissionReceipt{
		ReceiptID:         "rcpt_greenhouse_987654",
		ProviderReference: "app_gh_submission_proof",
		SubmissionURL:     job.DirectApplyURL,
		ConfirmedAt:       time.Now().UTC(),
		ConfirmedByUser:   true,
		Notes:             "Application submitted successfully on employer Greenhouse portal.",
	}

	confirmedSession, appRecord, err := svc.ConfirmApplicationWorkflowSubmission(ctx, userID, session.ID, receipt)
	if err != nil {
		t.Fatalf("failed confirming submission: %v", err)
	}

	if confirmedSession.Status != WorkflowStatusApplied {
		t.Errorf("expected status to be 'applied', got '%s'", confirmedSession.Status)
	}
	if appRecord == nil || appRecord.ID == "" {
		t.Errorf("expected durable ApplicationRecord to be created in ledger")
	}

	// Verify application is now in the durable application ledger (CAR-10)
	apps, err := svc.ListApplications(ctx, userID)
	if err != nil || len(apps) == 0 {
		t.Errorf("expected application to be recorded in ledger, found %d", len(apps))
	}

	// Verify saved job status updated to applied
	savedJob, err := svc.repo.GetSavedJobByCanonicalURL(ctx, userID, job.CanonicalURL)
	if err != nil || savedJob.Status != SavedJobStatusApplied {
		t.Errorf("expected saved job status to update to applied, got %+v", savedJob)
	}
}

func TestWorkflow_NativeManualFallback_CAR11(t *testing.T) {
	svc, userID, job := setupWorkflowTest(t)
	ctx := context.Background()

	// Verify native manual mode is supported without requiring automated browser agents (CAR-11)
	session, err := svc.PrepareApplicationWorkflow(ctx, userID, "ws_1", job, "", "", ExecutionModeNativeManual)
	if err != nil {
		t.Fatalf("failed to prepare native manual session: %v", err)
	}

	if session.ExecutionMode != ExecutionModeNativeManual {
		t.Errorf("expected execution mode native_manual, got %s", session.ExecutionMode)
	}
	if session.ApplyURL == "" {
		t.Errorf("expected valid ApplyURL for candidate manual navigation")
	}
}
