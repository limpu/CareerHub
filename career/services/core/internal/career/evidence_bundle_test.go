package career

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEvidenceBundle_CreationAndChecksum(t *testing.T) {
	now := time.Now().UTC()
	job := JobSnapshot{
		Title:              "Senior Distributed Systems Engineer (Go)",
		Company:            "Stripe",
		Location:           "Remote",
		JobType:            "full_time",
		RequiredSkills:     []string{"Go", "Distributed Systems", "Kubernetes"},
		DirectApplyURL:     "https://boards.greenhouse.io/stripe/jobs/5412987#apply",
		DescriptionSnippet: "Building world-class financial infrastructure in Go.",
		CapturedAt:         now,
	}

	resume := ResumeArtifactSnapshot{
		ResumeID:        "res-v2",
		Version:         "2.0",
		Format:          "pdf",
		ContentChecksum: "abc123hashsum456",
		DocumentPath:    "storage/resumes/user-1/res-v2.pdf",
		ContentSnippet:  "Senior Go engineer with 6 years experience.",
	}

	coverLetter := &CoverLetterSnapshot{
		CoverLetterID: "cl-01",
		Content:       "I am excited to apply for the Go Distributed Systems position...",
		Checksum:      "clhash789",
	}

	answers := []SubmittedAnswer{
		{
			QuestionID:    "q-1",
			QuestionText:  "Years of Go experience",
			AnswerValue:   "6",
			FieldCategory: "experience",
		},
		{
			QuestionID:    "q-2",
			QuestionText:  "Work authorization status",
			AnswerValue:   "US Citizen",
			FieldCategory: "legal",
		},
	}

	rawArtifacts := []ScrubbedArtifactItem{
		{
			ArtifactID:      "art-1",
			Name:            "api_receipt.json",
			MimeType:        "application/json",
			ScrubbedContent: "{\"status\":\"ok\",\"token\":\"Bearer secret-api-key-99\"}",
		},
	}

	bundle, err := AssembleEvidenceBundle(
		"bundle-101",
		"app-rec-102",
		"user-candidate-1",
		job,
		resume,
		coverLetter,
		answers,
		now,
		VerificationProviderReceipt,
		"STRIPE-GH-98412",
		rawArtifacts,
	)
	if err != nil {
		t.Fatalf("unexpected error assembling bundle: %v", err)
	}

	if bundle.BundleID != "bundle-101" {
		t.Errorf("expected bundle ID 'bundle-101', got %s", bundle.BundleID)
	}
	if bundle.IntegrityChecksum == "" {
		t.Fatalf("expected non-empty integrity checksum")
	}
	if len(bundle.IntegrityChecksum) != 64 {
		t.Errorf("expected 64-char SHA-256 hex checksum, got length %d (%s)", len(bundle.IntegrityChecksum), bundle.IntegrityChecksum)
	}

	// Verify integrity check succeeds on pristine bundle
	if !VerifyEvidenceIntegrity(bundle) {
		t.Errorf("expected bundle integrity verification to pass on pristine bundle")
	}
}

func TestEvidenceBundle_ExactArtifactPreservation(t *testing.T) {
	now := time.Now().UTC()
	job := JobSnapshot{
		Title:              "Staff Platform Engineer",
		Company:            "Acme Cloud",
		Location:           "San Francisco, CA",
		JobType:            "full_time",
		RequiredSkills:     []string{"Go", "PostgreSQL", "Kafka"},
		DirectApplyURL:     "https://acme.careers/job/101",
		DescriptionSnippet: "Platform engineer snippet",
		CapturedAt:         now,
	}

	resume := ResumeArtifactSnapshot{
		ResumeID:        "res-acme-v1",
		Version:         "1.0",
		Format:          "docx",
		ContentChecksum: "resume-sha256-exact-hash",
		DocumentPath:    "storage/resumes/user-1/res-acme-v1.docx",
		ContentSnippet:  "Full resume text",
	}

	answers := []SubmittedAnswer{
		{
			QuestionID:    "salary_expectation",
			QuestionText:  "What is your annual salary expectation?",
			AnswerValue:   "$195,000",
			FieldCategory: "compensation",
		},
	}

	bundle, err := AssembleEvidenceBundle(
		"bundle-202",
		"app-rec-202",
		"user-2",
		job,
		resume,
		nil,
		answers,
		now,
		VerificationUserAttestation,
		"MANUAL-ATTEST-123",
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bundle.JobSnapshot.Title != "Staff Platform Engineer" || bundle.JobSnapshot.Company != "Acme Cloud" {
		t.Errorf("job snapshot fields were not preserved exactly")
	}
	if bundle.ResumeArtifact.ContentChecksum != "resume-sha256-exact-hash" {
		t.Errorf("resume checksum was not preserved exactly")
	}
	if len(bundle.QuestionAnswers) != 1 || bundle.QuestionAnswers[0].AnswerValue != "$195,000" {
		t.Errorf("question answers were not preserved exactly")
	}
	if bundle.ConfirmationType != VerificationUserAttestation {
		t.Errorf("expected verification type %s, got %s", VerificationUserAttestation, bundle.ConfirmationType)
	}
}

func TestEvidenceBundle_PIIScrubbing_AT011(t *testing.T) {
	rawText := `
		{
			"status": "success",
			"auth": "Bearer secret_token_xyz_998811",
			"session": "session=\"cookie_session_abc123\"",
			"ssn": "000-12-3456",
			"cc": "4111-2222-3333-4444",
			"password": "secret_candidate_password_123"
		}
	`

	scrubbed := ScrubRawArtifactContent(rawText)

	if strings.Contains(scrubbed, "secret_token_xyz_998811") {
		t.Errorf("bearer token was not scrubbed: %s", scrubbed)
	}
	if !strings.Contains(scrubbed, "Bearer [REDACTED_BEARER_TOKEN]") {
		t.Errorf("expected [REDACTED_BEARER_TOKEN] replacement: %s", scrubbed)
	}
	if strings.Contains(scrubbed, "cookie_session_abc123") {
		t.Errorf("session cookie was not scrubbed: %s", scrubbed)
	}
	if strings.Contains(scrubbed, "000-12-3456") {
		t.Errorf("SSN was not scrubbed: %s", scrubbed)
	}
	if !strings.Contains(scrubbed, "[REDACTED_SSN]") {
		t.Errorf("expected [REDACTED_SSN] replacement: %s", scrubbed)
	}
	if strings.Contains(scrubbed, "4111-2222-3333-4444") {
		t.Errorf("credit card was not scrubbed: %s", scrubbed)
	}
	if !strings.Contains(scrubbed, "[REDACTED_CREDIT_CARD]") {
		t.Errorf("expected [REDACTED_CREDIT_CARD] replacement: %s", scrubbed)
	}
	if strings.Contains(scrubbed, "secret_candidate_password_123") {
		t.Errorf("password was not scrubbed: %s", scrubbed)
	}
}

func TestEvidenceBundle_AdminAccessDenied_WithoutGrant_AT011(t *testing.T) {
	bundle := &ApplicationEvidenceBundle{
		BundleID: "b-01",
		UserID:   "candidate-user-1",
	}

	// 1. Owner should succeed unconditionally
	err := AuthorizeEvidenceAccess(bundle, "candidate-user-1", false, false)
	if err != nil {
		t.Errorf("owner should have access, got error: %v", err)
	}

	// 2. Workspace Admin without explicit grant must be DENIED (AT-011, REQ-018)
	err = AuthorizeEvidenceAccess(bundle, "workspace-admin-user", true, false)
	if err == nil {
		t.Fatalf("expected ErrWorkspaceAdminAccessDenied for admin without grant, got nil")
	}
	if err != ErrWorkspaceAdminAccessDenied {
		t.Errorf("expected ErrWorkspaceAdminAccessDenied, got: %v", err)
	}

	// 3. Other member without permission must be DENIED
	err = AuthorizeEvidenceAccess(bundle, "stranger-user", false, false)
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied for unauthorized stranger, got: %v", err)
	}
}

func TestEvidenceBundle_AdminAccessAllowed_WithActiveGrant_AT011(t *testing.T) {
	bundle := &ApplicationEvidenceBundle{
		BundleID: "b-02",
		UserID:   "candidate-user-2",
	}

	// Workspace Admin WITH explicit resource grant is ALLOWED (AT-011)
	err := AuthorizeEvidenceAccess(bundle, "workspace-admin-user", true, true)
	if err != nil {
		t.Errorf("workspace admin with explicit grant should be allowed, got: %v", err)
	}
}

func TestEvidenceBundle_TamperDetection(t *testing.T) {
	now := time.Now().UTC()
	job := JobSnapshot{
		Title:   "Staff Security Engineer",
		Company: "Vault Corp",
	}
	resume := ResumeArtifactSnapshot{
		ResumeID:        "res-1",
		ContentChecksum: "hash-res-1",
	}
	answers := []SubmittedAnswer{
		{
			QuestionID:   "years_sec",
			QuestionText: "Years of security experience?",
			AnswerValue:  "5",
		},
	}

	bundle, err := AssembleEvidenceBundle(
		"bundle-tamper-test",
		"app-tamper-1",
		"user-tamper",
		job,
		resume,
		nil,
		answers,
		now,
		VerificationProviderReceipt,
		"RCPT-100",
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !VerifyEvidenceIntegrity(bundle) {
		t.Fatalf("bundle should be valid before tampering")
	}

	// Case 1: Maliciously alter question answer
	bundle.QuestionAnswers[0].AnswerValue = "12 (tampered)"
	if VerifyEvidenceIntegrity(bundle) {
		t.Errorf("tamper detection failed: modified answer value was not caught")
	}

	// Revert answer
	bundle.QuestionAnswers[0].AnswerValue = "5"
	if !VerifyEvidenceIntegrity(bundle) {
		t.Errorf("reverted bundle should verify successfully")
	}

	// Case 2: Maliciously alter job title
	bundle.JobSnapshot.Title = "Chief Security Officer"
	if VerifyEvidenceIntegrity(bundle) {
		t.Errorf("tamper detection failed: modified job title was not caught")
	}
}

func TestEvidenceBundle_ServiceIntegration(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	// 1. Create Evidence Bundle
	createReq := CreateEvidenceBundleRequest{
		ApplicationRecordID: "app-service-test-1",
		JobSnapshot: JobSnapshot{
			Title:          "Lead Go Developer",
			Company:        "Fintech Corp",
			DirectApplyURL: "https://fintech.example.com/apply",
			CapturedAt:     time.Now().UTC(),
		},
		ResumeArtifact: ResumeArtifactSnapshot{
			ResumeID:        "res-svc-1",
			Version:         "1.0",
			Format:          "pdf",
			ContentChecksum: "res-checksum-9912",
			DocumentPath:    "storage/resumes/u1/res-svc-1.pdf",
		},
		QuestionAnswers: []SubmittedAnswer{
			{
				QuestionID:   "auth",
				QuestionText: "Authorized to work?",
				AnswerValue:  "Yes",
			},
		},
		SubmissionTimestamp: time.Now().UTC(),
		ConfirmationType:    VerificationProviderReceipt,
		ProviderReference:   "FINTECH-APP-4401",
	}

	createResp, err := service.CreateEvidenceBundle(ctx, "user-svc-owner", createReq)
	if err != nil {
		t.Fatalf("failed to create evidence bundle via service: %v", err)
	}

	bundleID := createResp.Bundle.BundleID
	if bundleID == "" {
		t.Fatalf("expected non-empty bundle ID")
	}

	// 2. Retrieve as Owner
	getResp, err := service.GetEvidenceBundle(ctx, GetEvidenceBundleRequest{
		BundleID:         bundleID,
		RequestingUserID: "user-svc-owner",
	})
	if err != nil {
		t.Fatalf("owner failed to get evidence bundle: %v", err)
	}
	if !getResp.IsIntegrityValid {
		t.Errorf("expected bundle integrity to be valid")
	}

	// 3. Retrieve as Workspace Admin without Grant (must fail with AT-011)
	_, err = service.GetEvidenceBundle(ctx, GetEvidenceBundleRequest{
		BundleID:         bundleID,
		RequestingUserID: "admin-svc-user",
		IsWorkspaceAdmin: true,
		HasExplicitGrant: false,
	})
	if err != ErrWorkspaceAdminAccessDenied {
		t.Errorf("expected ErrWorkspaceAdminAccessDenied, got: %v", err)
	}

	// 4. Retrieve as Workspace Admin WITH Grant (must succeed)
	adminResp, err := service.GetEvidenceBundle(ctx, GetEvidenceBundleRequest{
		BundleID:         bundleID,
		RequestingUserID: "admin-svc-user",
		IsWorkspaceAdmin: true,
		HasExplicitGrant: true,
	})
	if err != nil {
		t.Fatalf("admin with explicit grant should succeed, got: %v", err)
	}
	if adminResp.Bundle.BundleID != bundleID {
		t.Errorf("mismatched bundle ID retrieved by admin")
	}

	// 5. Verify Integrity API
	verifyResp, err := service.VerifyEvidenceBundleIntegrity(ctx, bundleID)
	if err != nil {
		t.Fatalf("verify integrity failed: %v", err)
	}
	if !verifyResp.IsValid {
		t.Errorf("expected verify integrity response to be valid")
	}

	// 6. Retrieve by Application ID
	appResp, err := service.GetEvidenceBundleByApplicationID(ctx, "app-service-test-1", "user-svc-owner", false, false)
	if err != nil {
		t.Fatalf("failed to get bundle by application ID: %v", err)
	}
	if appResp.Bundle.ApplicationRecordID != "app-service-test-1" {
		t.Errorf("expected app ID 'app-service-test-1', got %s", appResp.Bundle.ApplicationRecordID)
	}
}

func TestEvidenceBundle_FixtureEvaluation(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "application_evidence_bundle.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture file %s: %v", fixturePath, err)
	}

	var fixture struct {
		Scenarios []struct {
			ScenarioID          string                    `json:"scenario_id"`
			Description         string                    `json:"description"`
			Bundle              ApplicationEvidenceBundle `json:"bundle,omitempty"`
			TamperedField       string                    `json:"tampered_field,omitempty"`
			TamperedValue       string                    `json:"tampered_value,omitempty"`
			RequestingActorRole string                    `json:"requesting_actor_role,omitempty"`
			HasExplicitGrant    bool                      `json:"has_explicit_grant,omitempty"`
			ExpectedError       string                    `json:"expected_error,omitempty"`
		} `json:"scenarios"`
	}

	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("failed to parse fixture json: %v", err)
	}

	if len(fixture.Scenarios) < 4 {
		t.Fatalf("expected at least 4 scenarios in fixture, found %d", len(fixture.Scenarios))
	}

	for _, sc := range fixture.Scenarios {
		t.Run(sc.ScenarioID, func(t *testing.T) {
			switch sc.ScenarioID {
			case "greenhouse_direct_verified_bundle", "workday_external_attested_bundle":
				b := sc.Bundle
				checksum, err := ComputeEvidenceChecksum(&b)
				if err != nil {
					t.Fatalf("failed to compute checksum for scenario %s: %v", sc.ScenarioID, err)
				}
				b.IntegrityChecksum = checksum
				if !VerifyEvidenceIntegrity(&b) {
					t.Errorf("scenario %s failed integrity check", sc.ScenarioID)
				}
				if len(b.ScrubbedArtifacts) == 0 {
					t.Errorf("scenario %s expected scrubbed artifacts", sc.ScenarioID)
				}

			case "tamper_detection_bundle":
				// Build bundle and alter field
				bundle := &ApplicationEvidenceBundle{
					BundleID:            "b-fixture-tamper",
					ApplicationRecordID: "app-1",
					UserID:              "u-1",
					JobSnapshot:         JobSnapshot{Title: "Title", Company: "Company"},
					ResumeArtifact:      ResumeArtifactSnapshot{ResumeID: "r1", ContentChecksum: "hash1"},
					QuestionAnswers:     []SubmittedAnswer{{QuestionID: "q1", AnswerValue: "original"}},
					SubmissionTimestamp: time.Now().UTC(),
				}
				c, _ := ComputeEvidenceChecksum(bundle)
				bundle.IntegrityChecksum = c

				// Modify answer value
				bundle.QuestionAnswers[0].AnswerValue = sc.TamperedValue
				if VerifyEvidenceIntegrity(bundle) {
					t.Errorf("tamper detection scenario failed to flag tampered value")
				}

			case "admin_access_denial_bundle":
				b := &ApplicationEvidenceBundle{
					BundleID: "b-fixture-admin",
					UserID:   "candidate-owner",
				}
				err := AuthorizeEvidenceAccess(b, "admin-actor", true, sc.HasExplicitGrant)
				if err == nil || err != ErrWorkspaceAdminAccessDenied {
					t.Errorf("expected ErrWorkspaceAdminAccessDenied, got: %v", err)
				}
			}
		})
	}
}
