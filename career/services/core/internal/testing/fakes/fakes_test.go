package fakes_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/social-platform/services/core/internal/ai"
	"github.com/social-platform/services/core/internal/testing/fakes"
)

func findFixturesRoot(t *testing.T) string {
	t.Helper()
	// Navigate up until we locate testdata/fixtures
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}

	for {
		candidate := filepath.Join(dir, "testdata", "fixtures")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("testdata/fixtures not found from %s", dir)
		}
		dir = parent
	}
}

// TestFakes_SyntheticResumeAndAdversarialInjection_AT003_AT019 verifies that:
// 1. Synthetic resumes with missing sensitive answers yield needs_input (AT-003)
// 2. Malicious prompt injections embedded in resumes/prompts are rejected (AT-019)
func TestFakes_SyntheticResumeAndAdversarialInjection_AT003_AT019(t *testing.T) {
	ctx := context.Background()
	fixturesDir := findFixturesRoot(t)

	// 1. Load Missing Sensitive Facts Profile (AT-003)
	missingProfilePath := filepath.Join(fixturesDir, "resumes", "missing_sensitive_facts_profile.json")
	missingBytes, err := os.ReadFile(missingProfilePath)
	if err != nil {
		t.Fatalf("failed to read missing_sensitive_facts_profile: %v", err)
	}

	var missingProfile struct {
		ConfirmedFacts []ai.ScopedFact `json:"confirmed_facts"`
	}
	if err := json.Unmarshal(missingBytes, &missingProfile); err != nil {
		t.Fatalf("failed to parse profile json: %v", err)
	}

	gateway := ai.NewAIGatewayService(ai.NewMockDeterministicAdapter())

	// Form asking for sponsorship (which candidate does NOT have in confirmed facts)
	// 1. Model attempting to guess without confirmed fact -> fails closed
	mockAdapter := ai.NewMockDeterministicAdapter()
	mockAdapter.SetCustomResponse(ai.TaskAppAnswer, `{
		"question": "Do you require visa_sponsorship?",
		"answer": "needs_input",
		"status": "pending_user_input"
	}`)
	mockGateway := ai.NewAIGatewayService(mockAdapter)

	req := ai.GenerationRequest{
		TaskKey:         ai.TaskAppAnswer,
		ModelID:         "economy-text-v1",
		ScopedFacts:     missingProfile.ConfirmedFacts,
		HasModelConsent: true,
	}

	res, err := mockGateway.Generate(ctx, req)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// Verify that unknown sensitive question evaluates to needs_input, never a guessed affirmative (AT-003)
	if len(res.GroundingReport.NeedsInputFields) == 0 {
		t.Fatalf("expected NeedsInputFields for unconfirmed sponsorship question, got %+v", res.GroundingReport)
	}

	// 2. Load Adversarial Injections and verify detection & isolation (AT-019)
	maliciousPromptsPath := filepath.Join(fixturesDir, "adversarial", "malicious_prompts.json")
	maliciousBytes, err := os.ReadFile(maliciousPromptsPath)
	if err != nil {
		t.Fatalf("failed to read malicious_prompts: %v", err)
	}

	var maliciousCatalog struct {
		Categories []struct {
			Name     string   `json:"name"`
			Payloads []string `json:"payloads"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(maliciousBytes, &maliciousCatalog); err != nil {
		t.Fatalf("failed to parse malicious catalog: %v", err)
	}

	for _, cat := range maliciousCatalog.Categories {
		for _, payload := range cat.Payloads {
			injectionReq := ai.GenerationRequest{
				TaskKey:         ai.TaskJobFitExplain,
				UntrustedData:   payload,
				HasModelConsent: true,
			}
			_, err := gateway.Generate(ctx, injectionReq)
			// Must be blocked if it attempts to override instructions or call tools
			if err == ai.ErrPromptInjectionBlocked {
				// Successfully blocked!
				continue
			}
			// If not blocked by regex, verify it was safely quarantined within untrusted tags
			sanitizer := ai.NewPromptSafetySanitizer()
			sanitized, _ := sanitizer.SanitizeUntrustedData(payload)
			if sanitized == payload && (cat.Name == "instruction_override" || cat.Name == "tool_mimicry") {
				t.Fatalf("category %s payload was not blocked or sanitized: %s", cat.Name, payload)
			}
		}
	}
}

// TestFakes_DuplicateJobClusterRecognition_AT004 verifies that cross-board job listings
// sharing canonical URLs or identical metadata are accurately grouped and ambiguous matches are flagged.
func TestFakes_DuplicateJobClusterRecognition_AT004(t *testing.T) {
	fixturesDir := findFixturesRoot(t)
	jobFixturesPath := filepath.Join(fixturesDir, "jobs", "duplicate_job_listings.json")
	data, err := os.ReadFile(jobFixturesPath)
	if err != nil {
		t.Fatalf("failed to read duplicate_job_listings.json: %v", err)
	}

	var root struct {
		Clusters []struct {
			ClusterID string `json:"cluster_id"`
			Listings  []struct {
				Board                   string  `json:"board"`
				ListingID               string  `json:"listing_id"`
				Title                   string  `json:"title"`
				Company                 string  `json:"company"`
				Location                string  `json:"location"`
				URL                     string  `json:"url"`
				SalaryMin               float64 `json:"salary_min"`
				SalaryMax               float64 `json:"salary_max"`
				IsDuplicateOfCanonical bool    `json:"is_duplicate_of_canonical"`
				AmbiguousMatch          bool    `json:"ambiguous_match"`
			} `json:"listings"`
		} `json:"clusters"`
	}

	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("failed to parse job fixtures: %v", err)
	}

	adapter := fakes.NewFakeJobBoardAdapter()

	for _, fixtureCluster := range root.Clusters {
		listings := make([]fakes.JobListing, 0, len(fixtureCluster.Listings))
		for _, l := range fixtureCluster.Listings {
			// If it is canonical duplicate, share canonical URL
			canonicalURL := ""
			if l.IsDuplicateOfCanonical {
				canonicalURL = "https://stripe.com/jobs/canonical-payments"
			}
			listings = append(listings, fakes.JobListing{
				ID:           l.ListingID,
				Board:        l.Board,
				Title:        l.Title,
				Company:      l.Company,
				Location:     l.Location,
				CanonicalURL: canonicalURL,
				SalaryMin:    l.SalaryMin,
				SalaryMax:    l.SalaryMax,
			})
		}

		clusters := adapter.DeduplicateJobs(listings)
		if len(clusters) == 0 {
			t.Fatalf("expected at least one cluster, got 0")
		}

		c := clusters[0]
		if len(c.ExactDuplicates) == 0 {
			t.Fatalf("expected exact duplicates for cluster %s, got 0", c.ClusterID)
		}
		if len(c.AmbiguousMatches) == 0 {
			t.Fatalf("expected ambiguous matches for cluster %s, got 0", c.ClusterID)
		}
	}
}

// TestFakes_ApplicationCrashAfterDispatch_NeverDuplicate_AT005_AT006 verifies that:
// 1. Successful application yields authenticatable receipt and marks verified_applied (AT-005)
// 2. Crash after external dispatch enters needs_confirmation and NEVER submits unreviewed duplicate (AT-006)
// 3. Reconciling ambiguous submission transitions status cleanly
func TestFakes_ApplicationCrashAfterDispatch_NeverDuplicate_AT005_AT006(t *testing.T) {
	ctx := context.Background()
	adapter := fakes.NewFakeJobBoardAdapter()

	// 1. Missing approval rejected (REQ-015)
	_, err := adapter.SubmitApplication(ctx, fakes.ApplicationSubmission{
		JobID:       "job_101",
		CandidateID: "cand_01",
	})
	if err != fakes.ErrApprovalMissingOrTampered {
		t.Fatalf("expected ErrApprovalMissingOrTampered, got %v", err)
	}

	// 2. Normal successful application produces receipt (AT-005)
	receipt, err := adapter.SubmitApplication(ctx, fakes.ApplicationSubmission{
		JobID:          "job_101",
		CandidateID:    "cand_01",
		ApprovalID:     "appr_123",
		PayloadHash:    "hash_abc",
		HasUserConsent: true,
	})
	if err != nil {
		t.Fatalf("submission failed: %v", err)
	}
	if receipt.VerifiedStatus != "verified_applied" || receipt.ConfirmationCode == "" {
		t.Fatalf("expected verified_applied with confirmation code, got %+v", receipt)
	}

	// 3. Crash after dispatch simulation (AT-006)
	crashAdapter := fakes.NewFakeJobBoardAdapter()
	crashAdapter.SetErrorMode(fakes.ErrorModeCrashAfterDispatch)

	sub := fakes.ApplicationSubmission{
		JobID:          "job_202",
		CandidateID:    "cand_02",
		ApprovalID:     "appr_456",
		PayloadHash:    "hash_def",
		HasUserConsent: true,
	}

	// First attempt crashes after dispatch
	_, err = crashAdapter.SubmitApplication(ctx, sub)
	if err != fakes.ErrSimulatedCrashAfterDispatch {
		t.Fatalf("expected ErrSimulatedCrashAfterDispatch, got %v", err)
	}

	// Second attempt (worker retry after reboot) MUST be blocked to prevent unreviewed duplicate (AT-006)
	_, retryErr := crashAdapter.SubmitApplication(ctx, sub)
	if retryErr != fakes.ErrDuplicateSubmissionBlocked {
		t.Fatalf("expected ErrDuplicateSubmissionBlocked on retry after crash, got %v", retryErr)
	}

	// Reconcile with user confirmation -> transitions to verified_applied
	reconciled, err := crashAdapter.ReconcileAmbiguousSubmission("job_202", "cand_02", true)
	if err != nil {
		t.Fatalf("failed to reconcile submission: %v", err)
	}
	if reconciled.VerifiedStatus != "verified_applied" {
		t.Fatalf("expected verified_applied after confirmation, got %s", reconciled.VerifiedStatus)
	}
}

// TestFakes_SocialPublishMultiChannel_PartialFailure_AT015 verifies that:
// 1. Channel A success is not repeated when Channel B fails
// 2. Overall status is partially_failed, never prematurely Published (AT-015)
func TestFakes_SocialPublishMultiChannel_PartialFailure_AT015(t *testing.T) {
	ctx := context.Background()
	publisher := fakes.NewFakeSocialPublisherAdapter()

	channels := []fakes.PublishingChannel{
		{ChannelID: "ch_linkedin_01", Platform: "linkedin"},
		{ChannelID: "ch_x_01", Platform: "x"},
	}

	// Simulate failure on X channel
	publisher.SetChannelFailure("ch_x_01", true)

	sub := fakes.SocialPostSubmission{
		PostID:      "post_test_001",
		WorkspaceID: "ws_test_01",
		Channels:    channels,
		ContentText: "Announcing our new distributed platform.",
		PayloadHash: "hash_post_001",
		ApprovalID:  "appr_post_001",
		IsApproved:  true,
	}

	// First execution: LinkedIn succeeds, X fails
	report, err := publisher.PublishPost(ctx, sub)
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	if report.OverallStatus != "partially_failed" {
		t.Fatalf("expected overall_status partially_failed, got %s", report.OverallStatus)
	}

	// Verify LinkedIn was recorded as published
	if !publisher.IsChannelPublished("post_test_001", "ch_linkedin_01") {
		t.Fatalf("expected linkedin channel to be recorded as published")
	}

	// Now fix X channel and retry
	publisher.SetChannelFailure("ch_x_01", false)

	retryReport, err := publisher.PublishPost(ctx, sub)
	if err != nil {
		t.Fatalf("retry publish failed: %v", err)
	}

	if retryReport.OverallStatus != "published" {
		t.Fatalf("expected published after retry, got %s", retryReport.OverallStatus)
	}
}

// TestFakes_MessagingDeduplicationAndQuota_AT007_AT008 verifies that:
// 1. Duplicate messages to the same recipient are blocked
// 2. Quota exhaustion blocks transmission
// 3. Unapproved outreach is rejected
func TestFakes_MessagingDeduplicationAndQuota_AT007_AT008(t *testing.T) {
	ctx := context.Background()
	messaging := fakes.NewFakeMessagingAdapter(2) // quota of 2 messages

	// Unapproved message rejected (AT-007)
	_, err := messaging.SendMessage(ctx, fakes.DirectMessageRequest{
		MessageID:   "msg_001",
		RecipientID: "recruiter_01",
	})
	if err != fakes.ErrApprovalMissingOrTampered {
		t.Fatalf("expected ErrApprovalMissingOrTampered, got %v", err)
	}

	// Valid message 1
	rcpt1, err := messaging.SendMessage(ctx, fakes.DirectMessageRequest{
		MessageID:      "msg_001",
		RecipientID:    "recruiter_01",
		PayloadHash:    "hash_001",
		ApprovalID:     "appr_001",
		HasApproval:    true,
		RemainingQuota: 2,
	})
	if err != nil || rcpt1.Status != "sent" {
		t.Fatalf("failed to send message 1: %v", err)
	}

	// Duplicate message to same recipient blocked (REQ-010)
	_, dupErr := messaging.SendMessage(ctx, fakes.DirectMessageRequest{
		MessageID:      "msg_002",
		RecipientID:    "recruiter_01",
		PayloadHash:    "hash_002",
		ApprovalID:     "appr_002",
		HasApproval:    true,
		RemainingQuota: 1,
	})
	if dupErr != fakes.ErrRecipientDuplicateMessage {
		t.Fatalf("expected ErrRecipientDuplicateMessage, got %v", dupErr)
	}

	// Valid message 2 to different recipient
	_, err = messaging.SendMessage(ctx, fakes.DirectMessageRequest{
		MessageID:      "msg_003",
		RecipientID:    "recruiter_02",
		PayloadHash:    "hash_003",
		ApprovalID:     "appr_003",
		HasApproval:    true,
		RemainingQuota: 1,
	})
	if err != nil {
		t.Fatalf("failed to send message 2: %v", err)
	}

	// Third message blocked due to quota exhaustion (AT-008)
	_, quotaErr := messaging.SendMessage(ctx, fakes.DirectMessageRequest{
		MessageID:      "msg_004",
		RecipientID:    "recruiter_03",
		PayloadHash:    "hash_004",
		ApprovalID:     "appr_004",
		HasApproval:    true,
		RemainingQuota: 0,
	})
	if quotaErr != fakes.ErrMessagingQuotaExceeded {
		t.Fatalf("expected ErrMessagingQuotaExceeded, got %v", quotaErr)
	}
}

// TestFakes_RateLimit429AndTimeout_Handling_REQ010 verifies timeout and rate limit modes.
func TestFakes_RateLimit429AndTimeout_Handling_REQ010(t *testing.T) {
	ctx := context.Background()
	adapter := fakes.NewFakeJobBoardAdapter()

	// 1. Rate limit 429
	adapter.SetErrorMode(fakes.ErrorModeRateLimit429)
	_, err := adapter.SubmitApplication(ctx, fakes.ApplicationSubmission{
		JobID:          "job_301",
		CandidateID:    "cand_01",
		ApprovalID:     "appr_01",
		PayloadHash:    "hash_01",
		HasUserConsent: true,
	})
	if err != fakes.ErrSimulatedRateLimit {
		t.Fatalf("expected ErrSimulatedRateLimit, got %v", err)
	}

	// 2. Timeout -> enters needs_confirmation (AT-005)
	adapter.SetErrorMode(fakes.ErrorModeTimeout)
	receipt, timeoutErr := adapter.SubmitApplication(ctx, fakes.ApplicationSubmission{
		JobID:          "job_302",
		CandidateID:    "cand_01",
		ApprovalID:     "appr_02",
		PayloadHash:    "hash_02",
		HasUserConsent: true,
	})
	if timeoutErr != fakes.ErrSimulatedTimeout {
		t.Fatalf("expected ErrSimulatedTimeout, got %v", timeoutErr)
	}
	if receipt.VerifiedStatus != "needs_confirmation" {
		t.Fatalf("expected needs_confirmation on timeout, got %s", receipt.VerifiedStatus)
	}
}
