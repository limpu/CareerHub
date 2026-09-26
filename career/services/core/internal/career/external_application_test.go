package career

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPortalClassification_Indeed_NativeVsRedirect(t *testing.T) {
	// 1. Native Indeed Easy Apply URL
	nativeURL := "https://www.indeed.com/viewjob?jk=9876543210abcdef&tk=1h2j3k4l5m"
	info := ClassifyPortal(nativeURL)

	if info.PortalType != PortalIndeedEasyApply {
		t.Fatalf("expected portal_type %s, got %s", PortalIndeedEasyApply, info.PortalType)
	}
	if info.SupportLevel != SupportLevelNativeEasyApply {
		t.Fatalf("expected support_level %s, got %s", SupportLevelNativeEasyApply, info.SupportLevel)
	}
	if !info.Capability.CanAutoFill {
		t.Fatalf("expected can_auto_fill to be true for native indeed easy apply")
	}
	if info.Capability.RequiresExternalRedirect {
		t.Fatalf("expected requires_external_redirect to be false for native indeed easy apply")
	}
	if info.ApplyButtonSelector != "button.ia-IndeedApplyButton" {
		t.Fatalf("expected button selector button.ia-IndeedApplyButton, got %s", info.ApplyButtonSelector)
	}

	// 2. Indeed External Click-Through Redirect URL (CAR-14)
	redirectURL := "https://www.indeed.com/rc/clk?jk=abcdef0123456789&from=vj&pos=top"
	redirectInfo := ClassifyPortal(redirectURL)

	if redirectInfo.PortalType != PortalGenericExternal {
		t.Fatalf("expected portal_type %s, got %s", PortalGenericExternal, redirectInfo.PortalType)
	}
	if redirectInfo.SupportLevel != SupportLevelAssistedManual {
		t.Fatalf("expected support_level %s, got %s", SupportLevelAssistedManual, redirectInfo.SupportLevel)
	}
	if redirectInfo.Capability.CanAutoFill {
		t.Fatalf("expected can_auto_fill to be false for indeed redirect (AT-010)")
	}
	if !redirectInfo.Capability.RequiresExternalRedirect {
		t.Fatalf("expected requires_external_redirect to be true for indeed redirect")
	}
	if !redirectInfo.Capability.ManualFallbackMandatory {
		t.Fatalf("expected manual_fallback_mandatory to be true (CAR-14)")
	}
}

func TestPortalClassification_UnsupportedEnterprisePortals_AT010(t *testing.T) {
	testCases := []struct {
		url          string
		expectedType PortalType
	}{
		{"https://acme.myworkdayjobs.com/en-US/careers/job/123", PortalWorkdayExternal},
		{"https://oracle.taleo.net/careersection/2/jobdetail.ftl?job=99", PortalTaleoExternal},
		{"https://jobs.successfactors.eu/career?company=SapCorp", PortalSuccessFactorsExternal},
		{"https://www.examplecorp.com/careers/apply/engineer", PortalGenericExternal},
	}

	for _, tc := range testCases {
		info := ClassifyPortal(tc.url)
		if info.PortalType != tc.expectedType {
			t.Errorf("url %s: expected portal type %s, got %s", tc.url, tc.expectedType, info.PortalType)
		}
		if info.SupportLevel != SupportLevelAssistedManual {
			t.Errorf("url %s: expected support level assisted_manual, got %s", tc.url, info.SupportLevel)
		}
		// Truth-in-advertising: never claim full auto-apply for enterprise/unsupported external portals (AT-010)
		if info.Capability.CanAutoFill {
			t.Errorf("url %s: can_auto_fill must be false for unsupported domain (AT-010)", tc.url)
		}
		if !info.Capability.RequiresExternalRedirect {
			t.Errorf("url %s: requires_external_redirect must be true (AT-010)", tc.url)
		}
		if !info.Capability.ManualFallbackMandatory {
			t.Errorf("url %s: manual_fallback_mandatory must be true", tc.url)
		}
	}
}

func TestPortalClassification_SupportedATSBoards(t *testing.T) {
	testCases := []struct {
		url          string
		expectedType PortalType
	}{
		{"https://boards.greenhouse.io/stripe/jobs/123", PortalGreenhouse},
		{"https://jobs.lever.co/netflix/456", PortalLever},
		{"https://jobs.ashbyhq.com/openai/789", PortalAshby},
		{"https://www.linkedin.com/jobs/view/1000", PortalLinkedInEasyApply},
	}

	for _, tc := range testCases {
		info := ClassifyPortal(tc.url)
		if info.PortalType != tc.expectedType {
			t.Errorf("url %s: expected %s, got %s", tc.url, tc.expectedType, info.PortalType)
		}
		if info.SupportLevel != SupportLevelNativeEasyApply {
			t.Errorf("url %s: expected native_easy_apply, got %s", tc.url, info.SupportLevel)
		}
		if !info.Capability.CanAutoFill {
			t.Errorf("url %s: expected can_auto_fill to be true", tc.url)
		}
	}
}

func TestExternalBundle_ClipboardGeneration_AT003(t *testing.T) {
	profile := MasterCareerProfile{
		Contact: ContactInfo{
			FullName: "Alice Developer",
			Email:    "alice@example.com",
			Phone:    "+1 (555) 123-4567",
			Location: "Austin, TX",
			Summary:  "Staff Backend Engineer specializing in resilient Go distributed systems.",
		},
		Links: []ProfileLink{
			{Label: "GitHub", URL: "https://github.com/alice"},
			{Label: "LinkedIn", URL: "https://linkedin.com/in/alice"},
		},
	}

	resume := TailoredResume{
		ID:       "res_tailored_101",
		FileName: "Alice_Developer_Go_Resume.pdf",
	}

	coverLetter := CoverLetter{
		ID:       "cov_letter_202",
		FullText: "Dear Hiring Team, I am excited to apply for this backend role.",
	}

	answers := map[string]string{
		"sponsorship_required": "No",
		"years_of_go":          "5",
	}

	job := DiscoveredJob{
		ID:             "job_workday_999",
		Title:          "Lead Backend Engineer",
		Company:        "Acme Global",
		DirectApplyURL: "https://acme.myworkdayjobs.com/careers/job/999",
	}

	bundle, err := PrepareExternalApplication("test_session_1", job, profile, resume, coverLetter, answers)
	if err != nil {
		t.Fatalf("failed to prepare external bundle: %v", err)
	}

	if bundle.Status != WorkflowStatusReadyForReview {
		t.Fatalf("expected initial status ready_for_review, got %s", bundle.Status)
	}
	if bundle.PortalInfo.PortalType != PortalWorkdayExternal {
		t.Fatalf("expected portal type workday_external, got %s", bundle.PortalInfo.PortalType)
	}
	if len(bundle.ClipboardItems) == 0 {
		t.Fatalf("expected clipboard items, got 0")
	}

	// Verify clipboard formatted text includes all confirmed facts and no fabricated extras
	txt := bundle.FormattedClipboardText
	if !containsStr(txt, "ALICE DEVELOPER") || !containsStr(txt, "alice@example.com") {
		t.Fatalf("formatted clipboard missing candidate contact info")
	}
	if !containsStr(txt, "https://github.com/alice") {
		t.Fatalf("formatted clipboard missing links")
	}
	if !containsStr(txt, "years of go: 5") {
		t.Fatalf("formatted clipboard missing custom answers")
	}
	if !containsStr(txt, "Alice_Developer_Go_Resume.pdf") {
		t.Fatalf("formatted clipboard missing resume file reference")
	}
}

func TestExternalApplication_DispatchDoesNotMarkApplied_AT005(t *testing.T) {
	bundle := &ExternalApplicationBundle{
		SessionID: "sess_ext_dispatch_test",
		JobID:     "job_123",
		JobTitle:  "Staff Engineer",
		Company:   "TechCorp",
		TargetURL: "https://techcorp.myworkdayjobs.com/job/123",
		Status:    WorkflowStatusReadyForReview,
	}

	dispatched, err := DispatchExternalApplication(bundle)
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	// Invariant: Clicking Apply or opening external link never marks Applied directly (AT-005)
	if dispatched.Status != WorkflowStatusDispatched {
		t.Fatalf("expected status dispatched, got %s (violates AT-005)", dispatched.Status)
	}
	if dispatched.DispatchedAt == nil {
		t.Fatalf("expected DispatchedAt timestamp to be recorded")
	}
	if dispatched.ConfirmedAt != nil {
		t.Fatalf("ConfirmedAt must be nil on dispatch (AT-005)")
	}
}

func TestExternalApplication_ExplicitConfirmationMarksApplied_REQ005_AT005(t *testing.T) {
	bundle := &ExternalApplicationBundle{
		SessionID: "sess_confirm_test",
		JobID:     "job_456",
		JobTitle:  "Principal Architect",
		Company:   "Enterprise Inc",
		TargetURL: "https://oracle.taleo.net/career/456",
		Status:    WorkflowStatusDispatched,
	}

	// Confirming without receipt or notes must fail closed
	_, _, err := ConfirmExternalApplication(bundle, "user_1", "", "")
	if err != ErrMissingSubmissionReceipt {
		t.Fatalf("expected ErrMissingSubmissionReceipt, got %v", err)
	}

	// Confirming with reference and notes succeeds
	confirmed, record, err := ConfirmExternalApplication(bundle, "user_1", "TALEO-CONF-98765", "Received confirmation email at 10:15am")
	if err != nil {
		t.Fatalf("unexpected confirmation error: %v", err)
	}

	if confirmed.Status != WorkflowStatusApplied {
		t.Fatalf("expected status applied, got %s", confirmed.Status)
	}
	if confirmed.ConfirmedAt == nil {
		t.Fatalf("expected ConfirmedAt timestamp to be set")
	}
	if confirmed.SubmissionReference != "TALEO-CONF-98765" {
		t.Fatalf("expected submission reference TALEO-CONF-98765, got %s", confirmed.SubmissionReference)
	}

	// Verify ApplicationRecord for permanent ledger (CAR-10)
	if record == nil {
		t.Fatalf("expected ApplicationRecord to be generated for ledger")
	}
	if record.JobID != "job_456" || record.Status != "applied" {
		t.Fatalf("invalid record: %+v", record)
	}
	if !containsStr(record.Notes, "TALEO-CONF-98765") {
		t.Fatalf("expected record notes to contain TALEO-CONF-98765, got %s", record.Notes)
	}
}

func TestExternalApplication_FixtureEvaluation(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "indeed_and_external_portals.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture file %s: %v", fixturePath, err)
	}

	var root struct {
		Fixtures []struct {
			FixtureID          string `json:"fixture_id"`
			URL                string `json:"url"`
			RedirectTarget     string `json:"redirect_target,omitempty"`
			PortalType         string `json:"portal_type"`
			SupportLevel       string `json:"support_level"`
			ExpectedCapability struct {
				CanAutoFill              bool `json:"can_auto_fill"`
				RequiresExternalRedirect bool `json:"requires_external_redirect"`
				ManualFallbackMandatory  bool `json:"manual_fallback_mandatory"`
			} `json:"expected_capability"`
		} `json:"fixtures"`
	}

	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	for _, fix := range root.Fixtures {
		info := ClassifyPortal(fix.URL, fix.RedirectTarget)

		if string(info.PortalType) != fix.PortalType {
			t.Errorf("[%s] expected portal type %s, got %s", fix.FixtureID, fix.PortalType, info.PortalType)
		}
		if string(info.SupportLevel) != fix.SupportLevel {
			t.Errorf("[%s] expected support level %s, got %s", fix.FixtureID, fix.SupportLevel, info.SupportLevel)
		}
		if info.Capability.CanAutoFill != fix.ExpectedCapability.CanAutoFill {
			t.Errorf("[%s] expected can_auto_fill %v, got %v", fix.FixtureID, fix.ExpectedCapability.CanAutoFill, info.Capability.CanAutoFill)
		}
		if info.Capability.RequiresExternalRedirect != fix.ExpectedCapability.RequiresExternalRedirect {
			t.Errorf("[%s] expected requires_external_redirect %v, got %v", fix.FixtureID, fix.ExpectedCapability.RequiresExternalRedirect, info.Capability.RequiresExternalRedirect)
		}
	}
}

func containsStr(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
