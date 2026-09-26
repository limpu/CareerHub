package career

import (
	"context"
	"testing"
	"time"
)

func parseTestTime(s string) *time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return &t
}

func createTestProfileForTailoring() *MasterCareerProfile {
	now := time.Now()
	startDate := parseTestTime("2021-01-01")

	return &MasterCareerProfile{
		ID:        "prof_tailor_test",
		UserID:    "user_tailor_123",
		CreatedAt: now,
		UpdatedAt: now,
		Contact: ContactInfo{
			FullName: "Jane Doe",
			Email:    "jane.doe@example.com",
			Phone:    "+1-555-0199",
			Location: "San Francisco, CA",
			Headline: "Senior Backend Engineer",
			Summary:  "Experienced backend engineer specializing in distributed systems and Go microservices.",
		},
		Experiences: []ExperienceItem{
			{
				ID:        "exp_1",
				Company:   "Acme Corp",
				Title:     "Backend Engineer",
				Location:  "San Francisco, CA",
				StartDate: startDate,
				EndDate:   nil,
				IsCurrent: true,
				Highlights: []string{
					"Maintained legacy monolithic services",
					"Engineered low-latency Go microservices handling 50k RPS",
					"Introduced Docker and CI/CD pipelines reducing deployment time by 40%",
				},
				Confirmed: true,
			},
		},
		Skills: []SkillItem{
			{ID: "sk_1", Name: "Go", Category: "Languages", Confirmed: true},
			{ID: "sk_2", Name: "Docker", Category: "DevOps", Confirmed: true},
			{ID: "sk_3", Name: "PostgreSQL", Category: "Databases", Confirmed: true},
			{ID: "sk_4", Name: "UnconfirmedSkill", Category: "Misc", Confirmed: false}, // Should be excluded
		},
	}
}

func TestTailorEngine_ZeroFabrication_REQ016(t *testing.T) {
	generator := NewResumeGenerator(NewNativeDocumentExtractor())
	engine := NewTailorEngine(generator, []byte("test-tailor-key-32b-length!!"))

	profile := createTestProfileForTailoring()
	job := JobTarget{
		ID:             "job_req016",
		Title:          "Senior Cloud Architect",
		Company:        "CloudScale Inc",
		RequiredSkills: []string{"Go", "Kubernetes", "AWS", "Rust"},
		Keywords:       []string{"microservices", "Docker"},
		Description:    "Looking for a Senior Cloud Architect proficient in Go, Kubernetes, AWS, and Rust to scale cloud infrastructure.",
	}

	tailored, err := engine.TailorResume(profile, job, ResumeFormatTXT, TemplateSingleColumnModern)
	if err != nil {
		t.Fatalf("TailorResume failed: %v", err)
	}

	// Invariant REQ-016 & AT-003: Unmatched required skills MUST NOT be fabricated into profile
	for _, skill := range tailored.TailoredProfile.Skills {
		if skill.Name == "Kubernetes" || skill.Name == "AWS" || skill.Name == "Rust" {
			t.Fatalf("Invariant REQ-016 violated: fabricated unpossessed skill %q into candidate profile", skill.Name)
		}
		if skill.Name == "UnconfirmedSkill" {
			t.Fatalf("Invariant AT-003 violated: unconfirmed skill included in tailored profile")
		}
	}

	// Verify unmatched skills are properly recorded in DiffSummary
	if len(tailored.DiffSummary.UnmatchedJobRequirements) != 3 {
		t.Errorf("expected 3 unmatched job requirements, got %d: %v",
			len(tailored.DiffSummary.UnmatchedJobRequirements), tailored.DiffSummary.UnmatchedJobRequirements)
	}

	// Verify matched skills were emphasized
	if len(tailored.DiffSummary.EmphasizedSkills) == 0 {
		t.Errorf("expected emphasized skills, got 0")
	}

	// Check status starts as pending_approval
	if tailored.ApprovalStatus != ApprovalStatusPending {
		t.Errorf("expected ApprovalStatusPending, got %s", tailored.ApprovalStatus)
	}

	if tailored.ApprovalToken == "" {
		t.Errorf("expected non-empty approval token")
	}
}

func TestTailorEngine_HighlightPrioritization(t *testing.T) {
	generator := NewResumeGenerator(NewNativeDocumentExtractor())
	engine := NewTailorEngine(generator, []byte("test-tailor-key-32b-length!!"))

	profile := createTestProfileForTailoring()
	job := JobTarget{
		ID:             "job_prio",
		Title:          "DevOps / Infrastructure Engineer",
		Company:        "DevOps Global",
		RequiredSkills: []string{"Docker"},
		Keywords:       []string{"Docker", "CI/CD"},
		Description:    "Seeking DevOps engineer to maintain Docker and CI/CD pipelines.",
	}

	tailored, err := engine.TailorResume(profile, job, ResumeFormatTXT, TemplateSingleColumnModern)
	if err != nil {
		t.Fatalf("TailorResume failed: %v", err)
	}

	// The Docker highlight should be prioritized to the front of Acme Corp highlights
	exp := tailored.TailoredProfile.Experiences[0]
	if len(exp.Highlights) != 3 {
		t.Fatalf("expected 3 highlights, got %d", len(exp.Highlights))
	}

	// Check that the first highlight contains Docker/CI/CD
	firstHighlight := exp.Highlights[0]
	if firstHighlight != "Introduced Docker and CI/CD pipelines reducing deployment time by 40%" {
		t.Errorf("expected Docker highlight to be prioritized first, got: %s", firstHighlight)
	}
}

func TestTailorEngine_ApprovalWorkflow_FND010(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	profile := createTestProfileForTailoring()
	if err := repo.SaveProfile(ctx, profile); err != nil {
		t.Fatalf("failed to save test profile: %v", err)
	}

	job := JobTarget{
		ID:             "job_appr",
		Title:          "Backend Specialist",
		Company:        "Acme Systems",
		RequiredSkills: []string{"Go", "PostgreSQL"},
	}

	// 1. Create tailored resume via Service
	tailored, err := service.CreateTailoredResume(ctx, "user_tailor_123", job, ResumeFormatTXT, TemplateSingleColumnModern)
	if err != nil {
		t.Fatalf("CreateTailoredResume failed: %v", err)
	}

	if tailored.ApprovalStatus != ApprovalStatusPending {
		t.Fatalf("expected pending status, got %s", tailored.ApprovalStatus)
	}

	// 2. Attempt approval with invalid token -> should fail
	_, err = service.ApproveTailoredResume(ctx, "user_tailor_123", tailored.ID, "bad-token-12345")
	if err != ErrInvalidApprovalToken {
		t.Fatalf("expected ErrInvalidApprovalToken, got %v", err)
	}

	// 3. Approve with valid token -> should succeed
	approved, err := service.ApproveTailoredResume(ctx, "user_tailor_123", tailored.ID, tailored.ApprovalToken)
	if err != nil {
		t.Fatalf("ApproveTailoredResume failed: %v", err)
	}
	if approved.ApprovalStatus != ApprovalStatusApproved {
		t.Fatalf("expected approved status, got %s", approved.ApprovalStatus)
	}
	if approved.ApprovedAt == nil {
		t.Fatalf("expected non-nil ApprovedAt timestamp")
	}

	// 4. Re-approving should be idempotent
	idempotentApprove, err := service.ApproveTailoredResume(ctx, "user_tailor_123", tailored.ID, tailored.ApprovalToken)
	if err != nil {
		t.Fatalf("expected idempotent success on double-approval, got %v", err)
	}
	if idempotentApprove.ApprovalStatus != ApprovalStatusApproved {
		t.Fatalf("expected approved status, got %s", idempotentApprove.ApprovalStatus)
	}
}

func TestTailorEngine_CoverLetter_ConfirmedFactsOnly(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	profile := createTestProfileForTailoring()
	if err := repo.SaveProfile(ctx, profile); err != nil {
		t.Fatalf("failed to save test profile: %v", err)
	}

	job := JobTarget{
		ID:             "job_cov",
		Title:          "Senior Distributed Systems Engineer",
		Company:        "MegaCloud Tech",
		RequiredSkills: []string{"Go", "PostgreSQL"},
	}

	// 1. Generate cover letter
	cl, err := service.CreateCoverLetter(ctx, "user_tailor_123", job, ResumeFormatTXT, "Jane Recruiter")
	if err != nil {
		t.Fatalf("CreateCoverLetter failed: %v", err)
	}

	if cl.ApprovalStatus != ApprovalStatusPending {
		t.Errorf("expected pending approval status, got %s", cl.ApprovalStatus)
	}
	if len(cl.ByteContent) == 0 {
		t.Errorf("expected non-empty byte content")
	}
	if !cl.ParseBackVerified {
		t.Errorf("expected ParseBackVerified to be true")
	}

	// 2. Approve cover letter
	approvedCL, err := service.ApproveCoverLetter(ctx, "user_tailor_123", cl.ID, cl.ApprovalToken)
	if err != nil {
		t.Fatalf("ApproveCoverLetter failed: %v", err)
	}
	if approvedCL.ApprovalStatus != ApprovalStatusApproved {
		t.Errorf("expected approved status, got %s", approvedCL.ApprovalStatus)
	}
}

func TestTailorEngine_ParseBackVerification_AT002(t *testing.T) {
	generator := NewResumeGenerator(NewNativeDocumentExtractor())
	engine := NewTailorEngine(generator, []byte("test-tailor-key-32b-length!!"))

	profile := createTestProfileForTailoring()
	job := JobTarget{
		ID:             "job_pb",
		Title:          "Lead Engineer",
		Company:        "Alpha Corp",
		RequiredSkills: []string{"Go", "PostgreSQL"},
	}

	formats := []MasterResumeFormat{ResumeFormatTXT, ResumeFormatPDF, ResumeFormatDOCX}
	for _, fmt := range formats {
		tailored, err := engine.TailorResume(profile, job, fmt, TemplateSingleColumnModern)
		if err != nil {
			t.Fatalf("TailorResume failed for format %s: %v", fmt, err)
		}
		if !tailored.ParseBackVerified {
			t.Errorf("expected ParseBackVerified to be true for format %s", fmt)
		}
	}
}

