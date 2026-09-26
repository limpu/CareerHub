package career

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// PortalType identifies the recognized job board or ATS vendor
type PortalType string

const (
	PortalIndeedEasyApply        PortalType = "indeed_easy_apply"
	PortalLinkedInEasyApply      PortalType = "linkedin_easy_apply"
	PortalGreenhouse             PortalType = "greenhouse_direct"
	PortalLever                  PortalType = "lever_direct"
	PortalAshby                  PortalType = "ashby_direct"
	PortalWorkdayExternal        PortalType = "workday_external"
	PortalTaleoExternal          PortalType = "taleo_external"
	PortalSuccessFactorsExternal PortalType = "successfactors_external"
	PortalGenericExternal        PortalType = "generic_external"
)

// PortalSupportLevel defines whether the platform can run native assisted apply or must use assisted manual flow
type PortalSupportLevel string

const (
	SupportLevelNativeEasyApply PortalSupportLevel = "native_easy_apply"
	SupportLevelAssistedManual  PortalSupportLevel = "assisted_manual"
)

// PortalCapability specifies operational boundaries and disclosures (CAR-14, AT-010)
type PortalCapability struct {
	CanAutoFill              bool     `json:"can_auto_fill"`
	RequiresExternalRedirect bool     `json:"requires_external_redirect"`
	ManualFallbackMandatory  bool     `json:"manual_fallback_mandatory"`
	Description              string   `json:"description"`
	KnownLimitations         []string `json:"known_limitations"`
}

// ExternalPortalInfo holds classified portal metadata and compliance directives (CAR-14, AT-010)
type ExternalPortalInfo struct {
	OriginalURL         string             `json:"original_url"`
	CleanURL            string             `json:"clean_url"`
	Domain              string             `json:"domain"`
	PortalType          PortalType         `json:"portal_type"`
	SupportLevel        PortalSupportLevel `json:"support_level"`
	Capability          PortalCapability   `json:"capability"`
	IsEasyApply         bool               `json:"is_easy_apply"`
	ApplyButtonSelector string             `json:"apply_button_selector,omitempty"`
	ModalSelector       string             `json:"modal_selector,omitempty"`
	RecommendedAction   string             `json:"recommended_action"`
}

// ClipboardItem is a single field item formatted for one-click candidate copy-paste
type ClipboardItem struct {
	FieldID string `json:"field_id"`
	Label   string `json:"label"`
	Value   string `json:"value"`
	Section string `json:"section"`
}

// ExternalApplicationBundle represents an application packet prepared for an external portal
type ExternalApplicationBundle struct {
	SessionID              string                 `json:"session_id"`
	JobID                  string                 `json:"job_id"`
	JobTitle               string                 `json:"job_title"`
	Company                string                 `json:"company"`
	TargetURL              string                 `json:"target_url"`
	PortalInfo             ExternalPortalInfo     `json:"portal_info"`
	TailoredResumeID       string                 `json:"tailored_resume_id"`
	CoverLetterID          string                 `json:"cover_letter_id,omitempty"`
	ClipboardItems         []ClipboardItem        `json:"clipboard_items"`
	FormattedClipboardText string                 `json:"formatted_clipboard_text"`
	Status                 ApplicationWorkflowStatus `json:"status"`
	CreatedAt              time.Time              `json:"created_at"`
	DispatchedAt           *time.Time             `json:"dispatched_at,omitempty"`
	ConfirmedAt            *time.Time             `json:"confirmed_at,omitempty"`
	SubmissionReference    string                 `json:"submission_reference,omitempty"`
	ConfirmationNotes      string                 `json:"confirmation_notes,omitempty"`
}

var (
	ErrUnsupportedLiveApplyAction = errors.New("live automated submission is not supported on this external portal; open in browser with assisted manual flow (AT-010)")
	ErrExternalBundleNotDispatched = errors.New("cannot confirm application submission before external dispatch has occurred (AT-005)")
	ErrMissingSubmissionReceipt    = errors.New("explicit confirmation requires submission reference or candidate confirmation statement (REQ-005, AT-005)")
)

// ClassifyPortal inspects a job application URL and deterministically identifies portal vendor, support level, and limitations
func ClassifyPortal(rawURL string, redirectTarget ...string) ExternalPortalInfo {
	if len(redirectTarget) > 0 && strings.TrimSpace(redirectTarget[0]) != "" {
		targetInfo := ClassifyPortal(redirectTarget[0])
		targetInfo.OriginalURL = rawURL
		return targetInfo
	}

	cleanURL := NormalizeCanonicalURL(rawURL)
	parsed, err := url.Parse(cleanURL)
	domain := ""
	if err == nil {
		domain = strings.ToLower(parsed.Hostname())
	}

	info := ExternalPortalInfo{
		OriginalURL: rawURL,
		CleanURL:    cleanURL,
		Domain:      domain,
	}

	// 1. Indeed detection (SRC-C3, CAR-14)
	if strings.Contains(domain, "indeed.com") {
		// Differentiate Easy Apply vs External Click-Through
		if strings.Contains(cleanURL, "rc/clk") || strings.Contains(cleanURL, "fccid=") || strings.Contains(cleanURL, "vjs=3") && strings.Contains(cleanURL, "applyurl=") {
			// External redirect through Indeed
			info.PortalType = PortalGenericExternal
			info.SupportLevel = SupportLevelAssistedManual
			info.IsEasyApply = false
			info.ApplyButtonSelector = "a#viewJobButtonLink"
			info.RecommendedAction = "Open in browser with prepared clipboard answers (Assisted Manual Flow)"
			info.Capability = PortalCapability{
				CanAutoFill:              false,
				RequiresExternalRedirect: true,
				ManualFallbackMandatory:  true,
				Description:              "Employer requires direct application on company portal. Automated application is disabled to avoid silent failures.",
				KnownLimitations: []string{
					"External company ATS requires manual candidate navigation (AT-010)",
					"Candidate must confirm submission after completion (AT-005)",
				},
			}
			return info
		}

		// Native Indeed Easy Apply
		info.PortalType = PortalIndeedEasyApply
		info.SupportLevel = SupportLevelNativeEasyApply
		info.IsEasyApply = true
		info.ApplyButtonSelector = "button.ia-IndeedApplyButton"
		info.ModalSelector = "#indeed-apply-widget"
		info.RecommendedAction = "Proceed with Indeed Easy Apply In-Page Flow"
		info.Capability = PortalCapability{
			CanAutoFill:              true,
			RequiresExternalRedirect: false,
			ManualFallbackMandatory:  false,
			Description:              "Native Indeed Easy Apply modal with structured question fields.",
			KnownLimitations: []string{
				"Indeed phone screening questions require explicit candidate answers (AT-003)",
				"CAPTCHA challenges must be resolved interactively by candidate (AT-010)",
			},
		}
		return info
	}

	// 2. LinkedIn Easy Apply
	if strings.Contains(domain, "linkedin.com") {
		info.PortalType = PortalLinkedInEasyApply
		info.SupportLevel = SupportLevelNativeEasyApply
		info.IsEasyApply = true
		info.ApplyButtonSelector = "button.jobs-apply-button"
		info.ModalSelector = "div.jobs-easy-apply-modal"
		info.RecommendedAction = "Proceed with LinkedIn Easy Apply Flow"
		info.Capability = PortalCapability{
			CanAutoFill:              true,
			RequiresExternalRedirect: false,
			ManualFallbackMandatory:  false,
			Description:              "Standard LinkedIn Easy Apply modal with Shadow DOM outlet.",
			KnownLimitations: []string{
				"Requires Shadow DOM traversal for modern widgets (SRC-C4)",
				"Unconfirmed questions halt for review (AT-003)",
			},
		}
		return info
	}

	// 3. Greenhouse Direct Board
	if strings.Contains(domain, "greenhouse.io") {
		info.PortalType = PortalGreenhouse
		info.SupportLevel = SupportLevelNativeEasyApply
		info.IsEasyApply = true
		info.ApplyButtonSelector = "form#application_form input[type='submit']"
		info.RecommendedAction = "Proceed with Greenhouse Direct Form Submission"
		info.Capability = PortalCapability{
			CanAutoFill:              true,
			RequiresExternalRedirect: false,
			ManualFallbackMandatory:  false,
			Description:              "Greenhouse ATS direct application board.",
			KnownLimitations: []string{
				"Custom employer questions require candidate grounding (AT-003)",
			},
		}
		return info
	}

	// 4. Lever Direct Board
	if strings.Contains(domain, "lever.co") {
		info.PortalType = PortalLever
		info.SupportLevel = SupportLevelNativeEasyApply
		info.IsEasyApply = true
		info.ApplyButtonSelector = "button.template-btn-submit"
		info.RecommendedAction = "Proceed with Lever Direct Form Submission"
		info.Capability = PortalCapability{
			CanAutoFill:              true,
			RequiresExternalRedirect: false,
			ManualFallbackMandatory:  false,
			Description:              "Lever direct job board application form.",
			KnownLimitations: []string{
				"Resume upload required (PDF/DOCX)",
			},
		}
		return info
	}

	// 5. Ashby Direct Board
	if strings.Contains(domain, "ashbyhq.com") {
		info.PortalType = PortalAshby
		info.SupportLevel = SupportLevelNativeEasyApply
		info.IsEasyApply = true
		info.ApplyButtonSelector = "button[type='submit']"
		info.RecommendedAction = "Proceed with Ashby Direct Form Submission"
		info.Capability = PortalCapability{
			CanAutoFill:              true,
			RequiresExternalRedirect: false,
			ManualFallbackMandatory:  false,
			Description:              "Ashby single-page application board.",
			KnownLimitations: []string{
				"Multi-step questionnaires must not guess unspecified skills (AT-003)",
			},
		}
		return info
	}

	// 6. Workday External ATS (CAR-14, AT-010)
	if strings.Contains(domain, "myworkdayjobs.com") || strings.Contains(domain, "workday.com") {
		info.PortalType = PortalWorkdayExternal
		info.SupportLevel = SupportLevelAssistedManual
		info.IsEasyApply = false
		info.RecommendedAction = "Open Workday in browser with Assisted Manual Flow"
		info.Capability = PortalCapability{
			CanAutoFill:              false,
			RequiresExternalRedirect: true,
			ManualFallbackMandatory:  true,
			Description:              "Enterprise Workday ATS with isolated corporate authentication. Automated application is disabled by design (truth-in-advertising AT-010).",
			KnownLimitations: []string{
				"Workday multi-tenant portals require candidate login and proprietary file upload workflows",
				"Full automation claim rejected; system generates clipboard bundle and tracks manual confirmation (CAR-14, AT-005)",
			},
		}
		return info
	}

	// 7. Oracle Taleo External
	if strings.Contains(domain, "taleo.net") {
		info.PortalType = PortalTaleoExternal
		info.SupportLevel = SupportLevelAssistedManual
		info.IsEasyApply = false
		info.RecommendedAction = "Open Taleo in browser with Assisted Manual Flow"
		info.Capability = PortalCapability{
			CanAutoFill:              false,
			RequiresExternalRedirect: true,
			ManualFallbackMandatory:  true,
			Description:              "Legacy enterprise Taleo portal. Automated form filling unsupported; open in browser.",
			KnownLimitations: []string{
				"Legacy frames and session state require manual candidate execution",
			},
		}
		return info
	}

	// 8. SAP SuccessFactors External
	if strings.Contains(domain, "successfactors.com") || strings.Contains(domain, "successfactors.eu") {
		info.PortalType = PortalSuccessFactorsExternal
		info.SupportLevel = SupportLevelAssistedManual
		info.IsEasyApply = false
		info.RecommendedAction = "Open SAP SuccessFactors in browser with Assisted Manual Flow"
		info.Capability = PortalCapability{
			CanAutoFill:              false,
			RequiresExternalRedirect: true,
			ManualFallbackMandatory:  true,
			Description:              "Enterprise SAP portal with complex multi-stage registration.",
			KnownLimitations: []string{
				"Automated cross-domain submission unsupported; open in browser",
			},
		}
		return info
	}

	// 9. Generic External Website
	info.PortalType = PortalGenericExternal
	info.SupportLevel = SupportLevelAssistedManual
	info.IsEasyApply = false
	info.RecommendedAction = "Open company website in browser with Assisted Manual Flow"
	info.Capability = PortalCapability{
		CanAutoFill:              false,
		RequiresExternalRedirect: true,
		ManualFallbackMandatory:  true,
		Description:              "External employer portal. No universal automated apply supported (truth-in-advertising AT-010). Use clipboard assisted flow.",
		KnownLimitations: []string{
			"Third-party custom application flow requires candidate completion in browser",
			"Candidate must confirm submission upon receiving employer receipt (AT-005)",
		},
	}
	return info
}

// BuildClipboardBundle generates structured candidate facts and answers for quick copy-paste into external portals
func BuildClipboardBundle(profile MasterCareerProfile, resume TailoredResume, coverLetter CoverLetter, customAnswers map[string]string) ([]ClipboardItem, string) {
	var items []ClipboardItem

	// 1. Contact Info
	items = append(items, ClipboardItem{
		FieldID: "full_name",
		Label:   "Full Name",
		Value:   profile.Contact.FullName,
		Section: "Contact",
	})
	items = append(items, ClipboardItem{
		FieldID: "email",
		Label:   "Email Address",
		Value:   profile.Contact.Email,
		Section: "Contact",
	})
	if profile.Contact.Phone != "" {
		items = append(items, ClipboardItem{
			FieldID: "phone",
			Label:   "Phone Number",
			Value:   profile.Contact.Phone,
			Section: "Contact",
		})
	}
	if profile.Contact.Location != "" {
		items = append(items, ClipboardItem{
			FieldID: "location",
			Label:   "Location / City",
			Value:   profile.Contact.Location,
			Section: "Contact",
		})
	}
	if profile.Contact.Headline != "" {
		items = append(items, ClipboardItem{
			FieldID: "headline",
			Label:   "Professional Headline",
			Value:   profile.Contact.Headline,
			Section: "Summary",
		})
	}
	if profile.Contact.Summary != "" {
		items = append(items, ClipboardItem{
			FieldID: "summary",
			Label:   "Professional Summary",
			Value:   profile.Contact.Summary,
			Section: "Summary",
		})
	}

	// 2. Links
	for _, link := range profile.Links {
		items = append(items, ClipboardItem{
			FieldID: "link_" + strings.ToLower(link.Label),
			Label:   link.Label + " URL",
			Value:   link.URL,
			Section: "Links",
		})
	}

	// 3. Custom / Screening answers
	for k, v := range customAnswers {
		if strings.TrimSpace(v) != "" {
			items = append(items, ClipboardItem{
				FieldID: k,
				Label:   strings.ReplaceAll(k, "_", " "),
				Value:   v,
				Section: "Questions",
			})
		}
	}

	// 4. Cover Letter text if present
	if coverLetter.FullText != "" {
		items = append(items, ClipboardItem{
			FieldID: "cover_letter",
			Label:   "Cover Letter Text",
			Value:   coverLetter.FullText,
			Section: "Documents",
		})
	}

	// Build plain-text aggregated clipboard representation
	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== APPLICATION MATERIALS FOR %s ===\n\n", strings.ToUpper(profile.Contact.FullName)))
	b.WriteString(fmt.Sprintf("Name: %s\nEmail: %s\nPhone: %s\nLocation: %s\n\n", profile.Contact.FullName, profile.Contact.Email, profile.Contact.Phone, profile.Contact.Location))

	if len(profile.Links) > 0 {
		b.WriteString("--- LINKS ---\n")
		for _, l := range profile.Links {
			b.WriteString(fmt.Sprintf("%s: %s\n", l.Label, l.URL))
		}
		b.WriteString("\n")
	}

	if profile.Contact.Summary != "" {
		b.WriteString("--- PROFESSIONAL SUMMARY ---\n")
		b.WriteString(profile.Contact.Summary + "\n\n")
	}

	if len(customAnswers) > 0 {
		b.WriteString("--- SCREENING ANSWERS ---\n")
		for k, v := range customAnswers {
			b.WriteString(fmt.Sprintf("%s: %s\n", strings.ReplaceAll(k, "_", " "), v))
		}
		b.WriteString("\n")
	}

	if coverLetter.FullText != "" {
		b.WriteString("--- COVER LETTER ---\n")
		b.WriteString(coverLetter.FullText + "\n\n")
	}

	b.WriteString(fmt.Sprintf("Resume File: %s (ID: %s)\n", resume.FileName, resume.ID))
	b.WriteString("========================================")

	return items, b.String()
}

// PrepareExternalApplication creates an application bundle ready for external dispatch
func PrepareExternalApplication(
	sessionID string,
	job DiscoveredJob,
	profile MasterCareerProfile,
	resume TailoredResume,
	coverLetter CoverLetter,
	customAnswers map[string]string,
) (*ExternalApplicationBundle, error) {
	if sessionID == "" {
		sessionID = fmt.Sprintf("ext_sess_%d", time.Now().UnixNano())
	}

	targetURL := job.DirectApplyURL
	if targetURL == "" {
		targetURL = job.CanonicalURL
	}
	portalInfo := ClassifyPortal(targetURL)
	items, formattedText := BuildClipboardBundle(profile, resume, coverLetter, customAnswers)

	bundle := &ExternalApplicationBundle{
		SessionID:              sessionID,
		JobID:                  job.ID,
		JobTitle:               job.Title,
		Company:                job.Company,
		TargetURL:              portalInfo.CleanURL,
		PortalInfo:             portalInfo,
		TailoredResumeID:       resume.ID,
		CoverLetterID:          coverLetter.ID,
		ClipboardItems:         items,
		FormattedClipboardText: formattedText,
		Status:                 WorkflowStatusReadyForReview,
		CreatedAt:              time.Now().UTC(),
	}

	return bundle, nil
}

// DispatchExternalApplication marks the external session as dispatched (CAR-14, AT-005)
// Invariant: Clicking Apply or opening external link never marks the application as 'applied' directly (AT-005).
func DispatchExternalApplication(bundle *ExternalApplicationBundle) (*ExternalApplicationBundle, error) {
	now := time.Now().UTC()
	bundle.DispatchedAt = &now
	bundle.Status = WorkflowStatusDispatched
	return bundle, nil
}

// ConfirmExternalApplication records candidate's explicit confirmation or receipt (REQ-005, AT-005, CAR-10)
func ConfirmExternalApplication(
	bundle *ExternalApplicationBundle,
	userID string,
	submissionRef string,
	notes string,
) (*ExternalApplicationBundle, *ApplicationRecord, error) {
	if bundle.Status != WorkflowStatusDispatched && bundle.Status != WorkflowStatusNeedsConfirmation {
		return nil, nil, ErrExternalBundleNotDispatched
	}

	cleanRef := strings.TrimSpace(submissionRef)
	cleanNotes := strings.TrimSpace(notes)
	if cleanRef == "" && cleanNotes == "" {
		return nil, nil, ErrMissingSubmissionReceipt
	}

	now := time.Now().UTC()
	bundle.ConfirmedAt = &now
	bundle.SubmissionReference = cleanRef
	bundle.ConfirmationNotes = cleanNotes
	bundle.Status = WorkflowStatusApplied

	// Create durable ApplicationRecord for permanent ledger (CAR-10)
	notesCombined := cleanNotes
	if cleanRef != "" {
		notesCombined = fmt.Sprintf("Receipt: %s | %s", cleanRef, cleanNotes)
	}

	appRecord := &ApplicationRecord{
		ID:             fmt.Sprintf("app_rec_%d", now.UnixNano()),
		UserID:         userID,
		JobID:          bundle.JobID,
		Company:        bundle.Company,
		Title:          bundle.JobTitle,
		CanonicalURL:   bundle.TargetURL,
		Fingerprint:    NewDedupeEngine().ComputeFingerprint(bundle.Company, bundle.JobTitle, ""),
		AppliedAt:      now,
		SubmissionMode: string(bundle.PortalInfo.PortalType),
		Status:         "applied",
		Notes:          notesCombined,
	}

	return bundle, appRecord, nil
}

