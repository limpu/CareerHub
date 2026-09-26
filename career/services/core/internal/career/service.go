package career

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/social-platform/services/core/internal/storage"
)

var (
	ErrEmptyFileContent  = errors.New("uploaded file has 0 bytes")
	ErrQuarantinedFile   = errors.New("file failed quarantine security scan")
	ErrAlreadyConfirmed  = errors.New("extraction draft has already been confirmed")
	ErrInvalidFactUpdate = errors.New("invalid career fact update")
)

type CreateManualFactInput struct {
	Category    string      `json:"category"`
	Title       string      `json:"title"`
	Organization string     `json:"organization,omitempty"`
	StartDate   *time.Time  `json:"start_date,omitempty"`
	EndDate     *time.Time  `json:"end_date,omitempty"`
	IsCurrent   bool        `json:"is_current"`
	Details     []string    `json:"details"`
	PrivacyTier PrivacyTier `json:"privacy_tier"`
}

type ConfirmDraftInput struct {
	FactOverrides []ProfileFactItem `json:"fact_overrides,omitempty"`
}

type CareerService struct {
	repo             CareerRepository
	extractor        DocumentExtractor
	parser           *HeuristicFactParser
	generator        *ResumeGenerator
	tailorEngine     *TailorEngine
	feedbackAnalyzer *FeedbackAnalyzer
	discoveryEngine  *DiscoveryEngine
	matchEngine      *MatchEngine
	dedupeEngine     *DedupeEngine
	workflowEngine   *WorkflowEngine
	fieldEngine      *FieldRecognitionEngine
}

func NewCareerService(repo CareerRepository) *CareerService {
	extractor := NewNativeDocumentExtractor()
	generator := NewResumeGenerator(extractor)
	return &CareerService{
		repo:             repo,
		extractor:        extractor,
		parser:           NewHeuristicFactParser(),
		generator:        generator,
		tailorEngine:     NewTailorEngine(generator, []byte("career-service-tailor-approval-hmac-key")),
		feedbackAnalyzer: NewFeedbackAnalyzer(),
		discoveryEngine:  NewDiscoveryEngine(),
		matchEngine:      NewMatchEngine(),
		dedupeEngine:     NewDedupeEngine(),
		workflowEngine:   NewWorkflowEngine([]byte("career-service-workflow-approval-hmac-key")),
		fieldEngine:      NewFieldRecognitionEngine(),
	}
}

// ProcessUploadedDocument processes a resume file upload, validates quarantine, extracts text,
// generates reviewable facts, and stages a ResumeExtractionDraft (REQ-002, AT-001, AT-020).
func (s *CareerService) ProcessUploadedDocument(
	ctx context.Context,
	userID string,
	uploadID string,
	fileName string,
	mimeType string,
	data []byte,
) (*ResumeExtractionDraft, error) {
	if len(data) == 0 {
		return nil, ErrEmptyFileContent
	}

	// 1. Enforce quarantine security check (AT-020, FND-006)
	qStatus, qReason := storage.ScanForProhibitedContent(data)
	if qStatus == storage.QuarantineStatusRejected {
		return nil, fmt.Errorf("%w: %s", ErrQuarantinedFile, qReason)
	}

	// 2. Extract text natively (Pure-Go PDF / DOCX parser)
	extractedText, err := s.extractor.ExtractText(data, mimeType)
	if err != nil {
		return nil, fmt.Errorf("text extraction failed: %w", err)
	}

	// 3. Determine source provenance
	provenance := ProvenanceUploadPDF
	if strings.Contains(mimeType, "wordprocessingml") || strings.HasSuffix(strings.ToLower(fileName), ".docx") {
		provenance = ProvenanceUploadDOCX
	}

	// 4. Parse facts heuristically without guessing unstated facts (AT-003)
	facts := s.parser.ParseFacts(userID, extractedText, provenance)

	// Calculate draft confidence score
	var totalConfidence float64
	for _, f := range facts {
		totalConfidence += f.Confidence
	}
	confidenceScore := 0.0
	if len(facts) > 0 {
		confidenceScore = totalConfidence / float64(len(facts))
	}

	words := strings.Fields(extractedText)

	// 5. Stage draft for user review and confirmation
	draft := &ResumeExtractionDraft{
		ID:              uuid.New().String(),
		UserID:          userID,
		UploadID:        uploadID,
		FileName:        fileName,
		MimeType:        mimeType,
		Status:          ExtractionPending,
		RawTextPreview:  truncatePreview(extractedText, 1000),
		WordCount:       len(words),
		ConfidenceScore: confidenceScore,
		ExtractedFacts:  facts,
		CreatedAt:       time.Now(),
	}

	if err := s.repo.SaveDraft(ctx, draft); err != nil {
		return nil, fmt.Errorf("failed to save extraction draft: %w", err)
	}

	return draft, nil
}

// GetExtractionDraft retrieves a staged draft for review.
func (s *CareerService) GetExtractionDraft(ctx context.Context, userID, draftID string) (*ResumeExtractionDraft, error) {
	return s.repo.GetDraft(ctx, userID, draftID)
}

// ConfirmDraft approves staged facts, merges user overrides, and promotes them to canonical profile facts (REQ-002, AT-001).
func (s *CareerService) ConfirmDraft(ctx context.Context, userID, draftID string, in ConfirmDraftInput) ([]ProfileFactItem, error) {
	draft, err := s.repo.GetDraft(ctx, userID, draftID)
	if err != nil {
		return nil, err
	}
	if draft.Status == ExtractionConfirmed {
		return nil, ErrAlreadyConfirmed
	}

	now := time.Now()
	var finalFacts []ProfileFactItem

	// If overrides provided, use reviewed/corrected facts; otherwise promote extracted facts
	if len(in.FactOverrides) > 0 {
		for _, f := range in.FactOverrides {
			f.UserID = userID
			f.Confirmed = true
			if f.ID == "" {
				f.ID = uuid.New().String()
			}
			if f.PrivacyTier == "" {
				f.PrivacyTier = PrivacyPublic
			}
			f.UpdatedAt = now
			finalFacts = append(finalFacts, f)
		}
	} else {
		for _, f := range draft.ExtractedFacts {
			f.Confirmed = true
			f.UpdatedAt = now
			finalFacts = append(finalFacts, f)
		}
	}

	if err := s.repo.SaveFactsBatch(ctx, finalFacts); err != nil {
		return nil, fmt.Errorf("failed to save confirmed facts: %w", err)
	}

	if err := s.repo.UpdateDraftStatus(ctx, userID, draftID, ExtractionConfirmed, &now); err != nil {
		return nil, fmt.Errorf("failed to update draft status: %w", err)
	}

	return finalFacts, nil
}

// CreateManualFact directly saves a user-entered confirmed fact (AT-001: identical schema).
func (s *CareerService) CreateManualFact(ctx context.Context, userID string, in CreateManualFactInput) (*ProfileFactItem, error) {
	if in.Title == "" || in.Category == "" {
		return nil, errors.New("title and category are required")
	}

	now := time.Now()
	privacyTier := in.PrivacyTier
	if privacyTier == "" {
		privacyTier = PrivacyPublic
	}

	fact := &ProfileFactItem{
		ID:               uuid.New().String(),
		UserID:           userID,
		Category:         in.Category,
		Title:            in.Title,
		Organization:     in.Organization,
		StartDate:        in.StartDate,
		EndDate:          in.EndDate,
		IsCurrent:        in.IsCurrent,
		Details:          in.Details,
		Confidence:       1.0, // User entered facts have 100% confidence
		Confirmed:        true,
		SourceProvenance: ProvenanceManualEntry,
		SourceLocation:   "manual_form",
		PrivacyTier:      privacyTier,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.SaveFact(ctx, fact); err != nil {
		return nil, fmt.Errorf("failed to save manual profile fact: %w", err)
	}

	return fact, nil
}

// GetUserFacts returns canonical confirmed facts for a user.
func (s *CareerService) GetUserFacts(ctx context.Context, userID, category string) ([]ProfileFactItem, error) {
	return s.repo.GetFactsByUser(ctx, userID, category)
}

// UpdateFact modifies an existing profile fact.
func (s *CareerService) UpdateFact(ctx context.Context, userID string, fact *ProfileFactItem) error {
	if fact.ID == "" || fact.UserID != userID {
		return ErrInvalidFactUpdate
	}
	return s.repo.UpdateFact(ctx, fact)
}

// DeleteFact removes an existing profile fact.
func (s *CareerService) DeleteFact(ctx context.Context, userID, factID string) error {
	return s.repo.DeleteFact(ctx, userID, factID)
}

// GetMasterProfile retrieves the canonical career profile or initializes a fresh one (REQ-002, REQ-003).
func (s *CareerService) GetMasterProfile(ctx context.Context, userID string) (*MasterCareerProfile, error) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			profile = NewMasterCareerProfile(userID)
			// Populate from confirmed facts if present
			if _, syncErr := s.syncProfileFromFacts(ctx, profile); syncErr == nil {
				_ = s.repo.SaveProfile(ctx, profile)
			}
			return profile, nil
		}
		return nil, err
	}
	profile.Completeness = profile.CalculateCompleteness()
	return profile, nil
}

// GetMasterProfileWithPrivacyCheck enforces strict workspace admin isolation (REQ-018, AT-011).
// Admin access to a workspace does NOT grant access to personal career profiles unless explicitly shared by owner.
func (s *CareerService) GetMasterProfileWithPrivacyCheck(ctx context.Context, callerUserID, targetUserID string, isWorkspaceAdmin bool) (*MasterCareerProfile, error) {
	if callerUserID == targetUserID {
		return s.GetMasterProfile(ctx, targetUserID)
	}

	profile, err := s.GetMasterProfile(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	// Invariant REQ-018: personal career records are private unless explicitly shared
	if !profile.Privacy.ShareWithWorkspaceAdmin || !isWorkspaceAdmin {
		return nil, ErrPrivateProfileAccessDenied
	}

	return profile, nil
}

// SaveMasterProfile updates the entire master career profile and records an audit trail.
func (s *CareerService) SaveMasterProfile(ctx context.Context, userID string, profile *MasterCareerProfile, changedBy, reason string) (*MasterCareerProfile, error) {
	if profile.UserID != userID {
		return nil, errors.New("cannot update profile for another user")
	}

	now := time.Now()
	profile.UpdatedAt = now
	profile.RecordAudit("all", "", "updated master profile", changedBy, reason)
	profile.Completeness = profile.CalculateCompleteness()

	if err := s.repo.SaveProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return profile, nil
}

// UpdateProfileSection updates a specific section with field-level audit tracking.
func (s *CareerService) UpdateProfileSection(ctx context.Context, userID, section string, payload []byte, changedBy string) (*MasterCareerProfile, error) {
	profile, err := s.GetMasterProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(section) {
	case "contact":
		var c ContactInfo
		if err := json.Unmarshal(payload, &c); err != nil {
			return nil, fmt.Errorf("invalid contact payload: %w", err)
		}
		profile.RecordAudit("contact", profile.Contact.FullName, c.FullName, changedBy, "section update")
		profile.Contact = c

	case "links":
		var links []ProfileLink
		if err := json.Unmarshal(payload, &links); err != nil {
			return nil, fmt.Errorf("invalid links payload: %w", err)
		}
		profile.RecordAudit("links", fmt.Sprintf("%d links", len(profile.Links)), fmt.Sprintf("%d links", len(links)), changedBy, "section update")
		profile.Links = links

	case "experience", "experiences":
		var exp []ExperienceItem
		if err := json.Unmarshal(payload, &exp); err != nil {
			return nil, fmt.Errorf("invalid experience payload: %w", err)
		}
		profile.RecordAudit("experiences", fmt.Sprintf("%d items", len(profile.Experiences)), fmt.Sprintf("%d items", len(exp)), changedBy, "section update")
		profile.Experiences = exp

	case "education":
		var edu []EducationItem
		if err := json.Unmarshal(payload, &edu); err != nil {
			return nil, fmt.Errorf("invalid education payload: %w", err)
		}
		profile.RecordAudit("education", fmt.Sprintf("%d items", len(profile.Education)), fmt.Sprintf("%d items", len(edu)), changedBy, "section update")
		profile.Education = edu

	case "skills":
		var skills []SkillItem
		if err := json.Unmarshal(payload, &skills); err != nil {
			return nil, fmt.Errorf("invalid skills payload: %w", err)
		}
		profile.RecordAudit("skills", fmt.Sprintf("%d skills", len(profile.Skills)), fmt.Sprintf("%d skills", len(skills)), changedBy, "section update")
		profile.Skills = skills

	case "projects":
		var proj []ProjectItem
		if err := json.Unmarshal(payload, &proj); err != nil {
			return nil, fmt.Errorf("invalid projects payload: %w", err)
		}
		profile.RecordAudit("projects", fmt.Sprintf("%d projects", len(profile.Projects)), fmt.Sprintf("%d projects", len(proj)), changedBy, "section update")
		profile.Projects = proj

	case "certificates":
		var certs []CertificateItem
		if err := json.Unmarshal(payload, &certs); err != nil {
			return nil, fmt.Errorf("invalid certificates payload: %w", err)
		}
		profile.RecordAudit("certificates", fmt.Sprintf("%d certs", len(profile.Certificates)), fmt.Sprintf("%d certs", len(certs)), changedBy, "section update")
		profile.Certificates = certs

	case "languages":
		var langs []LanguageItem
		if err := json.Unmarshal(payload, &langs); err != nil {
			return nil, fmt.Errorf("invalid languages payload: %w", err)
		}
		profile.RecordAudit("languages", fmt.Sprintf("%d langs", len(profile.Languages)), fmt.Sprintf("%d langs", len(langs)), changedBy, "section update")
		profile.Languages = langs

	case "consent":
		var consent UserCareerConsent
		if err := json.Unmarshal(payload, &consent); err != nil {
			return nil, fmt.Errorf("invalid consent payload: %w", err)
		}
		consent.ConsentTimestamp = time.Now()
		profile.RecordAudit("consent", profile.Consent.ConsentVersion, consent.ConsentVersion, changedBy, "consent update")
		profile.Consent = consent

	case "privacy":
		var privacy CareerPrivacySettings
		if err := json.Unmarshal(payload, &privacy); err != nil {
			return nil, fmt.Errorf("invalid privacy payload: %w", err)
		}
		profile.RecordAudit("privacy", string(profile.Privacy.ProfileVisibility), string(privacy.ProfileVisibility), changedBy, "privacy update")
		profile.Privacy = privacy

	default:
		return nil, fmt.Errorf("%w: %s", ErrSectionNotFound, section)
	}

	profile.UpdatedAt = time.Now()
	profile.Completeness = profile.CalculateCompleteness()

	if err := s.repo.SaveProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save updated section: %w", err)
	}

	return profile, nil
}

// UpdateCareerConsent updates user consents for career services (REQ-023).
func (s *CareerService) UpdateCareerConsent(ctx context.Context, userID string, consent UserCareerConsent) (*MasterCareerProfile, error) {
	profile, err := s.GetMasterProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	consent.ConsentTimestamp = time.Now()
	if consent.ConsentVersion == "" {
		consent.ConsentVersion = "v1.0"
	}
	profile.RecordAudit("consent", profile.Consent.ConsentVersion, consent.ConsentVersion, userID, "user consent update")
	profile.Consent = consent
	profile.UpdatedAt = time.Now()

	if err := s.repo.SaveProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// UpdatePrivacySettings updates privacy visibility and workspace admin sharing flags (REQ-018).
func (s *CareerService) UpdatePrivacySettings(ctx context.Context, userID string, privacy CareerPrivacySettings) (*MasterCareerProfile, error) {
	profile, err := s.GetMasterProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile.RecordAudit("privacy", string(profile.Privacy.ProfileVisibility), string(privacy.ProfileVisibility), userID, "privacy settings update")
	profile.Privacy = privacy
	profile.UpdatedAt = time.Now()

	if err := s.repo.SaveProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// syncProfileFromFacts aggregates confirmed ProfileFactItems into MasterCareerProfile (AT-001, AT-003).
func (s *CareerService) syncProfileFromFacts(ctx context.Context, profile *MasterCareerProfile) (*MasterCareerProfile, error) {
	facts, err := s.repo.GetFactsByUser(ctx, profile.UserID, "")
	if err != nil {
		return profile, err
	}

	for _, f := range facts {
		if !f.Confirmed {
			continue
		}
		switch f.Category {
		case "contact":
			if profile.Contact.FullName == "" && f.Title != "" && !strings.Contains(f.Title, "@") {
				profile.Contact.FullName = f.Title
			}
			for _, detail := range f.Details {
				if strings.Contains(detail, "@") && profile.Contact.Email == "" {
					profile.Contact.Email = detail
				} else if strings.HasPrefix(detail, "Phone:") && profile.Contact.Phone == "" {
					profile.Contact.Phone = strings.TrimSpace(strings.TrimPrefix(detail, "Phone:"))
				}
			}
		case "experience":
			exists := false
			for _, existing := range profile.Experiences {
				if existing.Title == f.Title && existing.Company == f.Organization {
					exists = true
					break
				}
			}
			if !exists {
				profile.Experiences = append(profile.Experiences, ExperienceItem{
					ID:         f.ID,
					Title:      f.Title,
					Company:    f.Organization,
					StartDate:  f.StartDate,
					EndDate:    f.EndDate,
					IsCurrent:  f.IsCurrent,
					Highlights: f.Details,
					Confirmed:  true,
				})
			}
		case "education":
			exists := false
			for _, existing := range profile.Education {
				if existing.Degree == f.Title && existing.Institution == f.Organization {
					exists = true
					break
				}
			}
			if !exists {
				profile.Education = append(profile.Education, EducationItem{
					ID:          f.ID,
					Institution: f.Organization,
					Degree:      f.Title,
					StartDate:   f.StartDate,
					EndDate:     f.EndDate,
					Highlights:  f.Details,
					Confirmed:   true,
				})
			}
		case "skill":
			exists := false
			for _, existing := range profile.Skills {
				if strings.EqualFold(existing.Name, f.Title) {
					exists = true
					break
				}
			}
			if !exists {
				profile.Skills = append(profile.Skills, SkillItem{
					ID:          f.ID,
					Name:        f.Title,
					Category:    "technical",
					Proficiency: ProficiencyUnspecified,
					Confirmed:   true,
				})
			}
		}
	}

	profile.Completeness = profile.CalculateCompleteness()
	return profile, nil
}

func truncatePreview(s string, maxChars int) string {
	if len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "..."
}

// GetCareerPreferences retrieves explicit job matching criteria and compensation preferences (REQ-003, AT-003).
func (s *CareerService) GetCareerPreferences(ctx context.Context, userID string) (*CareerPreferences, error) {
	pref, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrPreferencesNotFound) {
			pref = NewCareerPreferences(userID)
			_ = s.repo.SavePreferences(ctx, pref)
			return pref, nil
		}
		return nil, err
	}
	return pref, nil
}

// SaveCareerPreferences stores user-confirmed search parameters and exclusions (REQ-003, AT-003).
func (s *CareerService) SaveCareerPreferences(ctx context.Context, userID string, pref *CareerPreferences) (*CareerPreferences, error) {
	if pref.UserID != userID {
		return nil, errors.New("cannot update preferences for another user")
	}

	if err := pref.Validate(); err != nil {
		return nil, err
	}

	pref.UpdatedAt = time.Now()
	if err := s.repo.SavePreferences(ctx, pref); err != nil {
		return nil, fmt.Errorf("failed to save preferences: %w", err)
	}

	return pref, nil
}

// CheckJobMatchExclusion verifies whether an opportunity matches the candidate's exclusions.
func (s *CareerService) CheckJobMatchExclusion(ctx context.Context, userID, companyName, jobTitle, description string) (bool, string, error) {
	pref, err := s.GetCareerPreferences(ctx, userID)
	if err != nil {
		return false, "", err
	}
	excluded, reason := pref.MatchesExclusion(companyName, jobTitle, description)
	return excluded, reason, nil
}

// GenerateMasterResume produces an ATS-friendly single-column resume (PDF, DOCX, TXT)
// with deterministic section ordering and runs an independent parse-back verification check (CAR-04, REQ-003, AT-002).
func (s *CareerService) GenerateMasterResume(
	ctx context.Context,
	userID string,
	format MasterResumeFormat,
	template ResumeTemplateType,
) (*GeneratedResume, *ParseBackVerificationResult, error) {
	profile, err := s.GetMasterProfile(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve master profile for resume generation: %w", err)
	}

	res, vResult, err := s.generator.GenerateMasterResume(profile, format, template)
	if err != nil {
		return nil, vResult, err
	}

	if err := s.repo.SaveResume(ctx, res); err != nil {
		return nil, vResult, fmt.Errorf("failed to persist generated resume: %w", err)
	}

	return res, vResult, nil
}

// GetGeneratedResume retrieves a generated resume by ID with owner isolation.
func (s *CareerService) GetGeneratedResume(ctx context.Context, userID, resumeID string) (*GeneratedResume, error) {
	return s.repo.GetResume(ctx, userID, resumeID)
}

// ListGeneratedResumes lists all generated resumes for a user.
func (s *CareerService) ListGeneratedResumes(ctx context.Context, userID string) ([]GeneratedResume, error) {
	return s.repo.ListResumesByUser(ctx, userID)
}

// VerifyResumeParseBack executes an independent parse-back check on an existing generated resume.
func (s *CareerService) VerifyResumeParseBack(ctx context.Context, userID, resumeID string) (*ParseBackVerificationResult, error) {
	resume, err := s.repo.GetResume(ctx, userID, resumeID)
	if err != nil {
		return nil, err
	}

	profile, err := s.GetMasterProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.generator.VerifyParseBack(resume, profile)
}

// CreateTailoredResume produces a job-specific tailored resume draft with diff tracking and AT-003 zero-fabrication.
func (s *CareerService) CreateTailoredResume(
	ctx context.Context,
	userID string,
	job JobTarget,
	format MasterResumeFormat,
	template ResumeTemplateType,
) (*TailoredResume, error) {
	profile, err := s.GetMasterProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve master profile: %w", err)
	}

	tailored, err := s.tailorEngine.TailorResume(profile, job, format, template)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveTailoredResume(ctx, tailored); err != nil {
		return nil, fmt.Errorf("failed to persist tailored resume draft: %w", err)
	}

	return tailored, nil
}

// ApproveTailoredResume marks a tailored resume as approved and immutable using FND-010 cryptographic token.
func (s *CareerService) ApproveTailoredResume(ctx context.Context, userID, resumeID, token string) (*TailoredResume, error) {
	tailored, err := s.repo.GetTailoredResume(ctx, userID, resumeID)
	if err != nil {
		return nil, err
	}

	if tailored.ApprovalStatus == ApprovalStatusApproved {
		return tailored, nil // Idempotent
	}

	if !s.tailorEngine.ValidateApprovalToken(tailored.ID, tailored.ChecksumSHA256, token) {
		return nil, ErrInvalidApprovalToken
	}

	if err := s.repo.UpdateTailoredResumeStatus(ctx, userID, resumeID, ApprovalStatusApproved); err != nil {
		return nil, err
	}

	return s.repo.GetTailoredResume(ctx, userID, resumeID)
}

// GetTailoredResume retrieves a tailored resume by ID.
func (s *CareerService) GetTailoredResume(ctx context.Context, userID, resumeID string) (*TailoredResume, error) {
	return s.repo.GetTailoredResume(ctx, userID, resumeID)
}

// ListTailoredResumes lists all tailored resumes for a user.
func (s *CareerService) ListTailoredResumes(ctx context.Context, userID string) ([]TailoredResume, error) {
	return s.repo.ListTailoredResumes(ctx, userID)
}

// CreateCoverLetter crafts a targeted, fact-grounded cover letter aligned to the employer requirements.
func (s *CareerService) CreateCoverLetter(
	ctx context.Context,
	userID string,
	job JobTarget,
	format MasterResumeFormat,
	recipientName string,
) (*CoverLetter, error) {
	profile, err := s.GetMasterProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve master profile: %w", err)
	}

	letter, err := s.tailorEngine.GenerateCoverLetter(profile, job, format, recipientName)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveCoverLetter(ctx, letter); err != nil {
		return nil, fmt.Errorf("failed to persist cover letter draft: %w", err)
	}

	return letter, nil
}

// ApproveCoverLetter marks a cover letter as approved and immutable using FND-010 cryptographic token.
func (s *CareerService) ApproveCoverLetter(ctx context.Context, userID, letterID, token string) (*CoverLetter, error) {
	letter, err := s.repo.GetCoverLetter(ctx, userID, letterID)
	if err != nil {
		return nil, err
	}

	if letter.ApprovalStatus == ApprovalStatusApproved {
		return letter, nil // Idempotent
	}

	if !s.tailorEngine.ValidateApprovalToken(letter.ID, letter.ChecksumSHA256, token) {
		return nil, ErrInvalidApprovalToken
	}

	if err := s.repo.UpdateCoverLetterStatus(ctx, userID, letterID, ApprovalStatusApproved); err != nil {
		return nil, err
	}

	return s.repo.GetCoverLetter(ctx, userID, letterID)
}

// GetCoverLetter retrieves a cover letter by ID.
func (s *CareerService) GetCoverLetter(ctx context.Context, userID, letterID string) (*CoverLetter, error) {
	return s.repo.GetCoverLetter(ctx, userID, letterID)
}

// ListCoverLetters lists all cover letters for a user.
func (s *CareerService) ListCoverLetters(ctx context.Context, userID string) ([]CoverLetter, error) {
	return s.repo.ListCoverLetters(ctx, userID)
}

// GenerateResumeFeedback analyzes the user's master career profile for ATS readability,
// quantification rate, and action verb density (IMP-CAR-06).
func (s *CareerService) GenerateResumeFeedback(ctx context.Context, userID string) (*ResumeFeedbackReport, error) {
	profile, err := s.GetMasterProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve master profile: %w", err)
	}

	report, err := s.feedbackAnalyzer.AnalyzeProfile(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze profile: %w", err)
	}

	if err := s.repo.SaveFeedbackReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to persist feedback report: %w", err)
	}

	return report, nil
}

// GetLatestResumeFeedback returns the most recently generated feedback report for the user.
func (s *CareerService) GetLatestResumeFeedback(ctx context.Context, userID string) (*ResumeFeedbackReport, error) {
	report, err := s.repo.GetLatestFeedbackReport(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrFeedbackReportNotFound) {
			// Auto-generate if not yet present
			return s.GenerateResumeFeedback(ctx, userID)
		}
		return nil, err
	}
	return report, nil
}

// GetSkillsDemandAnalysis provides market intelligence for a target role while strictly
// separating market suggestions from actual confirmed candidate facts (CAR-06, AT-003, AT-028).
func (s *CareerService) GetSkillsDemandAnalysis(ctx context.Context, userID, targetRole string) (*SkillsDemandAnalysis, error) {
	profile, err := s.GetMasterProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve master profile: %w", err)
	}

	return s.feedbackAnalyzer.AnalyzeSkillsDemand(profile, targetRole)
}

// DiscoverJobs coordinates multi-board discovery queries and deduplication (IMP-CAR-07, REQ-005, AT-004).
func (s *CareerService) DiscoverJobs(ctx context.Context, query JobSearchQuery) ([]DiscoveredJob, error) {
	return s.discoveryEngine.SearchMultiBoard(ctx, query)
}

// GetJobBoardCapabilities returns programmatic capability contracts for all supported job boards (AT-010).
func (s *CareerService) GetJobBoardCapabilities(ctx context.Context) []BoardCapabilities {
	return s.discoveryEngine.GetCapabilitiesMatrix()
}

// NormalizeJobURL strips marketing parameters and normalizes canonical job posting URLs (AT-004).
func (s *CareerService) NormalizeJobURL(rawURL string) string {
	return NormalizeCanonicalURL(rawURL)
}

// FilterDiscoveredJobs applies fine-grained post-filters (age, work mode, job type, company exclusions, salary)
// and returns filtered results accompanied by a transparent audit trail (IMP-CAR-08, CAR-08).
func (s *CareerService) FilterDiscoveredJobs(ctx context.Context, jobs []DiscoveredJob, filter AdvancedJobFilter) ([]DiscoveredJob, FilterAuditReport) {
	evaluator := NewFilterEvaluator()
	return evaluator.Evaluate(jobs, filter, time.Now().UTC())
}

// EvaluateJobMatch runs explainable job matching for a candidate against a target opportunity,
// enforcing hard gates, transparent scoring, and unknown-vs-mismatch distinctions (IMP-CAR-09, CAR-09, AT-003, AT-028).
func (s *CareerService) EvaluateJobMatch(ctx context.Context, userID string, job DiscoveredJob, weights *ScoringWeights) (*JobMatchResult, error) {
	profile, err := s.GetMasterProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve candidate master profile: %w", err)
	}

	pref, err := s.GetCareerPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve candidate preferences: %w", err)
	}

	return s.matchEngine.EvaluateJobMatch(profile, pref, &job, weights)
}

// EvaluateBatchJobMatches runs explainable job matching for a candidate against multiple jobs in batch,
// sorting candidates by eligibility and overall score descending.
func (s *CareerService) EvaluateBatchJobMatches(ctx context.Context, userID string, jobs []DiscoveredJob, weights *ScoringWeights) ([]JobMatchResult, error) {
	profile, err := s.GetMasterProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve candidate master profile: %w", err)
	}

	pref, err := s.GetCareerPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve candidate preferences: %w", err)
	}

	results := make([]JobMatchResult, 0, len(jobs))
	for _, job := range jobs {
		res, err := s.matchEngine.EvaluateJobMatch(profile, pref, &job, weights)
		if err != nil {
			return nil, err
		}
		results = append(results, *res)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].IsEligible != results[j].IsEligible {
			return results[i].IsEligible
		}
		return results[i].OverallScore > results[j].OverallScore
	})

	return results, nil
}

// GetDefaultScoringWeights returns the authoritative default weights that sum to 100 (CAR-09).
func (s *CareerService) GetDefaultScoringWeights() ScoringWeights {
	return DefaultScoringWeights()
}

// --- Saved Jobs, Deduplication & Exclusion History (IMP-CAR-10, CAR-10, AT-004) ---

// SaveJob saves or shortlists an opportunity to the candidate's personal ledger.
func (s *CareerService) SaveJob(ctx context.Context, userID string, input SaveJobInput) (*SavedJob, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}

	canonicalURL := NormalizeCanonicalURL(input.Job.CanonicalURL)
	// Check if already saved
	if existing, err := s.repo.GetSavedJobByCanonicalURL(ctx, userID, canonicalURL); err == nil && existing != nil {
		return nil, ErrDuplicateSavedJob
	}

	now := time.Now().UTC()
	status := input.Status
	if status == "" {
		status = SavedJobStatusSaved
	}
	priority := input.Priority
	if priority <= 0 {
		priority = 3
	}

	fingerprint := s.dedupeEngine.ComputeFingerprint(input.Job.Company, input.Job.Title, input.Job.Location.RawLocation)
	saved := &SavedJob{
		ID:             uuid.New().String(),
		UserID:         userID,
		JobID:          input.Job.ID,
		CanonicalURL:   canonicalURL,
		Fingerprint:    fingerprint,
		Title:          input.Job.Title,
		Company:        input.Job.Company,
		Location:       input.Job.Location,
		JobType:        input.Job.JobType,
		Compensation:   input.Job.Compensation,
		RequiredSkills: input.Job.RequiredSkills,
		DirectApplyURL: input.Job.DirectApplyURL,
		Status:         status,
		Notes:          input.Notes,
		Priority:       priority,
		Tags:           input.Tags,
		SavedAt:        now,
		UpdatedAt:      now,
	}

	if err := s.repo.SaveSavedJob(ctx, saved); err != nil {
		return nil, fmt.Errorf("failed to persist saved job: %w", err)
	}

	return saved, nil
}

// GetSavedJob retrieves a saved job by ID with owner isolation.
func (s *CareerService) GetSavedJob(ctx context.Context, userID, savedJobID string) (*SavedJob, error) {
	return s.repo.GetSavedJob(ctx, userID, savedJobID)
}

// UpdateSavedJob updates status, notes, priority, or tags on a saved job.
func (s *CareerService) UpdateSavedJob(ctx context.Context, userID, savedJobID string, input UpdateSavedJobInput) (*SavedJob, error) {
	saved, err := s.repo.GetSavedJob(ctx, userID, savedJobID)
	if err != nil {
		return nil, err
	}

	if input.Status != nil {
		saved.Status = *input.Status
	}
	if input.Notes != nil {
		saved.Notes = *input.Notes
	}
	if input.Priority != nil {
		saved.Priority = *input.Priority
	}
	if input.Tags != nil {
		saved.Tags = input.Tags
	}
	if input.AppliedAt != nil {
		saved.AppliedAt = input.AppliedAt
	}
	saved.UpdatedAt = time.Now().UTC()

	if err := s.repo.SaveSavedJob(ctx, saved); err != nil {
		return nil, err
	}
	return saved, nil
}

// ListSavedJobs returns candidate's saved jobs, optionally filtered by status and tag.
func (s *CareerService) ListSavedJobs(ctx context.Context, userID string, status SavedJobStatus, tag string) ([]SavedJob, error) {
	list, err := s.repo.ListSavedJobs(ctx, userID, status)
	if err != nil {
		return nil, err
	}

	if tag == "" {
		return list, nil
	}

	filtered := make([]SavedJob, 0)
	for _, item := range list {
		for _, t := range item.Tags {
			if strings.EqualFold(t, tag) {
				filtered = append(filtered, item)
				break
			}
		}
	}
	return filtered, nil
}

// DeleteSavedJob removes an opportunity from the saved list.
func (s *CareerService) DeleteSavedJob(ctx context.Context, userID, savedJobID string) error {
	return s.repo.DeleteSavedJob(ctx, userID, savedJobID)
}

// AddExclusion permanently excludes a company, canonical URL, or job fingerprint (CAR-10, C8).
func (s *CareerService) AddExclusion(ctx context.Context, userID string, input AddExclusionInput) (*JobExclusion, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}
	if input.Value == "" {
		return nil, ErrEmptyExclusionValue
	}
	if input.Type != ExclusionTypeJobCanonicalURL && input.Type != ExclusionTypeJobFingerprint && input.Type != ExclusionTypeCompanyName {
		return nil, ErrInvalidExclusionType
	}

	exclusion := &JobExclusion{
		ID:          uuid.New().String(),
		UserID:      userID,
		Type:        input.Type,
		Value:       input.Value,
		Reason:      input.Reason,
		JobTitle:    input.JobTitle,
		CompanyName: input.CompanyName,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.SaveExclusion(ctx, exclusion); err != nil {
		return nil, err
	}
	return exclusion, nil
}

// ListExclusions returns all active exclusions for the candidate.
func (s *CareerService) ListExclusions(ctx context.Context, userID string) ([]JobExclusion, error) {
	return s.repo.ListExclusions(ctx, userID)
}

// DeleteExclusion unblocks a previously excluded entity.
func (s *CareerService) DeleteExclusion(ctx context.Context, userID, exclusionID string) error {
	return s.repo.DeleteExclusion(ctx, userID, exclusionID)
}

// RecordApplication records an application submission in the separate application ledger (CAR-10, AT-004, AT-006).
func (s *CareerService) RecordApplication(ctx context.Context, userID string, input RecordApplicationInput) (*ApplicationRecord, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}

	canonicalURL := NormalizeCanonicalURL(input.Job.CanonicalURL)
	// Check if application already recorded
	if existing, err := s.repo.GetApplicationByCanonicalURL(ctx, userID, canonicalURL); err == nil && existing != nil {
		return nil, ErrDuplicateApplication
	}

	appliedAt := time.Now().UTC()
	if input.AppliedAt != nil {
		appliedAt = *input.AppliedAt
	}
	status := input.Status
	if status == "" {
		status = "submitted"
	}
	submissionMode := input.SubmissionMode
	if submissionMode == "" {
		submissionMode = "manual_browser"
	}

	fingerprint := s.dedupeEngine.ComputeFingerprint(input.Job.Company, input.Job.Title, input.Job.Location.RawLocation)
	record := &ApplicationRecord{
		ID:             uuid.New().String(),
		UserID:         userID,
		JobID:          input.Job.ID,
		CanonicalURL:   canonicalURL,
		Fingerprint:    fingerprint,
		Title:          input.Job.Title,
		Company:        input.Job.Company,
		Status:         status,
		SubmissionMode: submissionMode,
		AppliedAt:      appliedAt,
		Notes:          input.Notes,
	}

	if err := s.repo.SaveApplicationRecord(ctx, record); err != nil {
		return nil, err
	}

	// Also update any matching saved job to "applied" status
	if saved, err := s.repo.GetSavedJobByCanonicalURL(ctx, userID, canonicalURL); err == nil && saved != nil {
		saved.Status = SavedJobStatusApplied
		saved.AppliedAt = &appliedAt
		saved.UpdatedAt = time.Now().UTC()
		_ = s.repo.SaveSavedJob(ctx, saved)
	}

	return record, nil
}

// ListApplications retrieves candidate's application history.
func (s *CareerService) ListApplications(ctx context.Context, userID string) ([]ApplicationRecord, error) {
	return s.repo.ListApplications(ctx, userID)
}

// CheckJobDeduplication evaluates a batch of jobs against candidate's saved jobs, applications, and exclusions (CAR-10, AT-004).
func (s *CareerService) CheckJobDeduplication(ctx context.Context, userID string, jobs []DiscoveredJob) ([]JobDedupeResult, error) {
	savedJobs, err := s.repo.ListSavedJobs(ctx, userID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve saved jobs: %w", err)
	}

	applications, err := s.repo.ListApplications(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve applications: %w", err)
	}

	exclusions, err := s.repo.ListExclusions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve exclusions: %w", err)
	}

	results := make([]JobDedupeResult, 0, len(jobs))
	for _, job := range jobs {
		res := s.dedupeEngine.EvaluateJobDeduplication(job, savedJobs, applications, exclusions)
		results = append(results, res)
	}

	return results, nil
}



// --- Application Workflow & Human Review (IMP-CAR-11, CAR-11, REQ-005, REQ-015, AT-003, AT-005, AT-007) ---

func (s *CareerService) PrepareApplicationWorkflow(
	ctx context.Context,
	userID, workspaceID string,
	job DiscoveredJob,
	resumeID, coverLetterID string,
	mode ApplicationExecutionMode,
) (*ApplicationReviewSession, error) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve candidate profile: %w", err)
	}

	var resume *TailoredResume
	if resumeID != "" {
		res, err := s.repo.GetTailoredResume(ctx, userID, resumeID)
		if err == nil {
			resume = res
		}
	}

	var coverLetter *CoverLetter
	if coverLetterID != "" {
		cl, err := s.repo.GetCoverLetter(ctx, userID, coverLetterID)
		if err == nil {
			coverLetter = cl
		}
	}

	session, err := s.workflowEngine.PrepareApplicationSession(userID, workspaceID, job, *profile, resume, coverLetter, mode)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveReviewSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save review session: %w", err)
	}

	return session, nil
}

func (s *CareerService) GetApplicationWorkflowSession(ctx context.Context, userID, sessionID string) (*ApplicationReviewSession, error) {
	return s.repo.GetReviewSession(ctx, userID, sessionID)
}

func (s *CareerService) ListApplicationWorkflowSessions(ctx context.Context, userID string, status ApplicationWorkflowStatus) ([]ApplicationReviewSession, error) {
	return s.repo.ListReviewSessions(ctx, userID, status)
}

func (s *CareerService) UpdateApplicationWorkflowAnswers(
	ctx context.Context,
	userID, sessionID string,
	answers []ApplicationQuestionAnswer,
) (*ApplicationReviewSession, bool, error) {
	session, err := s.repo.GetReviewSession(ctx, userID, sessionID)
	if err != nil {
		return nil, false, err
	}

	approvalInvalidated, err := s.workflowEngine.UpdateAnswers(session, answers)
	if err != nil {
		return nil, false, err
	}

	if err := s.repo.UpdateReviewSession(ctx, session); err != nil {
		return nil, false, fmt.Errorf("failed to update review session: %w", err)
	}

	return session, approvalInvalidated, nil
}

func (s *CareerService) ApproveApplicationWorkflow(ctx context.Context, userID, sessionID, approverID string) (*ApplicationReviewSession, error) {
	session, err := s.repo.GetReviewSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if err := s.workflowEngine.ApproveSession(session, approverID); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateReviewSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update approved session: %w", err)
	}

	return session, nil
}

func (s *CareerService) DispatchApplicationWorkflow(ctx context.Context, userID, sessionID string, mode ApplicationExecutionMode) (*ApplicationReviewSession, error) {
	session, err := s.repo.GetReviewSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if err := s.workflowEngine.DispatchSession(session, mode); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateReviewSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update dispatched session: %w", err)
	}

	return session, nil
}

func (s *CareerService) CheckApplicationWorkflowTimeout(ctx context.Context, userID, sessionID string) (*ApplicationReviewSession, bool, error) {
	session, err := s.repo.GetReviewSession(ctx, userID, sessionID)
	if err != nil {
		return nil, false, err
	}

	timedOut := s.workflowEngine.CheckConfirmationTimeout(session)
	if timedOut {
		if err := s.repo.UpdateReviewSession(ctx, session); err != nil {
			return nil, false, fmt.Errorf("failed to update timed out session: %w", err)
		}
	}

	return session, timedOut, nil
}

func (s *CareerService) ConfirmApplicationWorkflowSubmission(
	ctx context.Context,
	userID, sessionID string,
	receipt SubmissionReceipt,
) (*ApplicationReviewSession, *ApplicationRecord, error) {
	session, err := s.repo.GetReviewSession(ctx, userID, sessionID)
	if err != nil {
		return nil, nil, err
	}

	appRecord, err := s.workflowEngine.ConfirmSubmission(session, receipt)
	if err != nil {
		return nil, nil, err
	}

	// 1. Update review session state to applied
	if err := s.repo.UpdateReviewSession(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("failed to update confirmed session: %w", err)
	}

	// 2. Persist application in permanent application ledger (CAR-10, REQ-005)
	if err := s.repo.SaveApplicationRecord(ctx, appRecord); err != nil {
		return nil, nil, fmt.Errorf("failed to save application ledger record: %w", err)
	}

	// 3. Update saved job status to applied if job was shortlisted/saved
	savedJob, err := s.repo.GetSavedJobByCanonicalURL(ctx, userID, session.ApplyURL)
	if err == nil && savedJob != nil {
		savedJob.Status = SavedJobStatusApplied
		_ = s.repo.SaveSavedJob(ctx, savedJob)
	}

	return session, appRecord, nil
}

func (s *CareerService) CancelApplicationWorkflow(ctx context.Context, userID, sessionID, reason string) (*ApplicationReviewSession, error) {
	session, err := s.repo.GetReviewSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if err := s.workflowEngine.CancelSession(session, reason); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateReviewSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update cancelled session: %w", err)
	}

	return session, nil
}

// RecognizeFieldsRequest carries form input fields or raw form JSON to recognize
type RecognizeFieldsRequest struct {
	FormID   string        `json:"form_id"`
	Title    string        `json:"title,omitempty"`
	Provider string        `json:"provider,omitempty"`
	Fields   []FormField   `json:"fields,omitempty"`
	Sections []FormSection `json:"sections,omitempty"`
}

// RecognizeFieldsResponse returns parsed fields enriched with types, categories, and security sanitization
type RecognizeFieldsResponse struct {
	FormID          string        `json:"form_id"`
	TotalFields     int           `json:"total_fields"`
	RecognizedForm  RecognizedForm `json:"recognized_form"`
	SecurityAlerts  []string      `json:"security_alerts,omitempty"`
}

// ResolveAnswersRequest carries form fields and session answers to ground against candidate profile (AT-003)
type ResolveAnswersRequest struct {
	Form           RecognizedForm         `json:"form"`
	CurrentAnswers map[string]interface{} `json:"current_answers,omitempty"`
}

// ResolveAnswersResponse returns field resolutions with confidence, provenance, and needs_input flags
type ResolveAnswersResponse struct {
	FormID          string            `json:"form_id"`
	Resolutions     []FieldResolution `json:"resolutions"`
	NeedsInputCount int               `json:"needs_input_count"`
	ResolvedCount   int               `json:"resolved_count"`
	ZeroFabrication bool              `json:"zero_fabrication"` // Always true per AT-003
}

// EvaluateConditionsRequest carries form fields and answers to evaluate dynamic dependencies (CAR-12)
type EvaluateConditionsRequest struct {
	Fields  []FormField            `json:"fields"`
	Answers map[string]interface{} `json:"answers"`
}

// EvaluateConditionsResponse returns computed condition states (active, hidden, disabled) for each field
type EvaluateConditionsResponse struct {
	States map[string]ConditionState `json:"states"`
}

// RecognizeFormFields classifies questions, inputs, and sanitizes adversarial prompts (CAR-12, AT-019)
func (s *CareerService) RecognizeFormFields(ctx context.Context, req RecognizeFieldsRequest) (*RecognizeFieldsResponse, error) {
	form := &RecognizedForm{
		FormID:   req.FormID,
		Title:    req.Title,
		Provider: req.Provider,
		Fields:   req.Fields,
		Sections: req.Sections,
	}

	recognized := s.fieldEngine.RecognizeForm(form)

	var securityAlerts []string
	totalFields := len(recognized.Fields)
	for _, f := range recognized.Fields {
		if len(f.SecurityFlags) > 0 {
			securityAlerts = append(securityAlerts, f.SecurityFlags...)
		}
	}
	for _, sec := range recognized.Sections {
		totalFields += len(sec.Fields)
		for _, f := range sec.Fields {
			if len(f.SecurityFlags) > 0 {
				securityAlerts = append(securityAlerts, f.SecurityFlags...)
			}
		}
	}

	return &RecognizeFieldsResponse{
		FormID:         recognized.FormID,
		TotalFields:    totalFields,
		RecognizedForm: *recognized,
		SecurityAlerts: securityAlerts,
	}, nil
}

// ResolveFormAnswers resolves form questions strictly against candidate's confirmed profile and preferences (AT-003)
func (s *CareerService) ResolveFormAnswers(ctx context.Context, userID string, req ResolveAnswersRequest) (*ResolveAnswersResponse, error) {
	profile, _ := s.repo.GetProfile(ctx, userID)
	preferences, _ := s.repo.GetPreferences(ctx, userID)

	resolutions := s.fieldEngine.ResolveFieldAnswers(&req.Form, profile, preferences, req.CurrentAnswers)

	needsInputCount := 0
	resolvedCount := 0
	for _, r := range resolutions {
		if r.NeedsInput {
			needsInputCount++
		} else if r.ConditionState == StateActive {
			resolvedCount++
		}
	}

	return &ResolveAnswersResponse{
		FormID:          req.Form.FormID,
		Resolutions:     resolutions,
		NeedsInputCount: needsInputCount,
		ResolvedCount:   resolvedCount,
		ZeroFabrication: true, // Guarantees zero-fabrication contract AT-003
	}, nil
}

// EvaluateConditionalFields calculates dynamic field visibility states given answer changes (CAR-12)
func (s *CareerService) EvaluateConditionalFields(ctx context.Context, req EvaluateConditionsRequest) (*EvaluateConditionsResponse, error) {
	states := s.fieldEngine.EvaluateFormConditions(req.Fields, req.Answers)
	return &EvaluateConditionsResponse{
		States: states,
	}, nil
}

// InitWizardRequest initializes a multi-step form wizard
type InitWizardRequest struct {
	WizardID string     `json:"wizard_id"`
	Title    string     `json:"title"`
	Provider string     `json:"provider"`
	Steps    []FormStep `json:"steps"`
}

// InitWizardResponse returns initialized state machine
type InitWizardResponse struct {
	StateMachine FormStateMachine `json:"state_machine"`
	CurrentStep  FormStep         `json:"current_step"`
}

// AdvanceWizardStepRequest advances the wizard with provided step answers
type AdvanceWizardStepRequest struct {
	StateMachine FormStateMachine       `json:"state_machine"`
	Answers      map[string]interface{} `json:"answers"`
	RawEvidence  string                 `json:"raw_evidence,omitempty"`
}

// AdvanceWizardStepResponse returns updated state machine and new active step
type AdvanceWizardStepResponse struct {
	StateMachine FormStateMachine `json:"state_machine"`
	ActiveStep   FormStep         `json:"active_step"`
	IsCompleted  bool             `json:"is_completed"`
	LoopDetected bool             `json:"loop_detected"`
	HaltReason   string           `json:"halt_reason,omitempty"`
}

// PreviousWizardStepRequest navigates to previous step
type PreviousWizardStepRequest struct {
	StateMachine FormStateMachine `json:"state_machine"`
}

// PreviousWizardStepResponse returns updated state machine
type PreviousWizardStepResponse struct {
	StateMachine FormStateMachine `json:"state_machine"`
	ActiveStep   FormStep         `json:"active_step"`
}

// TraverseShadowDOMRequest pierces shadow root (SRC-C4)
type TraverseShadowDOMRequest struct {
	RawHTML            string `json:"raw_html"`
	ShadowHostSelector string `json:"shadow_host_selector"`
}

// TraverseShadowDOMResponse returns pierced fields
type TraverseShadowDOMResponse struct {
	PiercedSelector string      `json:"pierced_selector"`
	Fields          []FormField `json:"fields"`
}

// InitFormStateMachine creates and validates a new form state machine (CAR-13)
func (s *CareerService) InitFormStateMachine(ctx context.Context, req InitWizardRequest) (*InitWizardResponse, error) {
	sm := NewFormStateMachine(req.WizardID, req.Title, req.Provider, req.Steps)
	currentStep, err := sm.GetCurrentStep()
	if err != nil {
		return nil, err
	}
	return &InitWizardResponse{
		StateMachine: *sm,
		CurrentStep:  *currentStep,
	}, nil
}

// AdvanceWizardStep validates zero-fabrication, evaluates loop detection, and advances step (CAR-13, AT-003, AT-006)
func (s *CareerService) AdvanceWizardStep(ctx context.Context, req AdvanceWizardStepRequest) (*AdvanceWizardStepResponse, error) {
	sm := req.StateMachine

	// Scrub raw evidence if provided (CAR-13, AT-011)
	if req.RawEvidence != "" {
		scrubbed := ScrubStepEvidence(req.RawEvidence)
		if sm.CurrentStepIndex <= len(sm.Steps) {
			sm.Steps[sm.CurrentStepIndex-1].ScrubbedEvidence = scrubbed
		}
	}

	activeStep, err := sm.AdvanceStep(req.Answers)
	if err != nil {
		if errors.Is(err, ErrLoopDetected) {
			return &AdvanceWizardStepResponse{
				StateMachine: sm,
				LoopDetected: true,
				HaltReason:   sm.HaltReason,
			}, nil
		}
		if errors.Is(err, ErrStepBlockedNeedsInput) {
			return &AdvanceWizardStepResponse{
				StateMachine: sm,
				HaltReason:   sm.HaltReason,
			}, nil
		}
		return nil, err
	}

	return &AdvanceWizardStepResponse{
		StateMachine: sm,
		ActiveStep:   *activeStep,
		IsCompleted:  sm.Status == StepStatusCompleted,
		LoopDetected: false,
	}, nil
}

// PreviousWizardStep navigates backwards in the wizard
func (s *CareerService) PreviousWizardStep(ctx context.Context, req PreviousWizardStepRequest) (*PreviousWizardStepResponse, error) {
	sm := req.StateMachine
	prevStep, err := sm.PreviousStep()
	if err != nil {
		return nil, err
	}
	return &PreviousWizardStepResponse{
		StateMachine: sm,
		ActiveStep:   *prevStep,
	}, nil
}

// TraverseShadowDOM pierces shadow root and returns selector (SRC-C4)
func (s *CareerService) TraverseShadowDOM(ctx context.Context, req TraverseShadowDOMRequest) (*TraverseShadowDOMResponse, error) {
	sm := NewFormStateMachine("temp_shadow", "Shadow Inspection", "web", nil)
	info, fields, err := sm.TraverseShadowDOM(req.RawHTML, req.ShadowHostSelector)
	if err != nil {
		return nil, err
	}
	return &TraverseShadowDOMResponse{
		PiercedSelector: info,
		Fields:          fields,
	}, nil
}

// ==========================================
// External & Indeed Applications (CAR-14, AT-005, AT-006, AT-010, SRC-C3)
// ==========================================

// ClassifyPortalRequest requests classification of a portal URL
type ClassifyPortalRequest struct {
	URL            string `json:"url"`
	RedirectTarget string `json:"redirect_target,omitempty"`
}

// ClassifyPortalResponse returns portal classification and capabilities
type ClassifyPortalResponse struct {
	PortalInfo ExternalPortalInfo `json:"portal_info"`
}

// PrepareExternalBundleRequest requests creation of a formatted external bundle
type PrepareExternalBundleRequest struct {
	JobID         string            `json:"job_id"`
	CustomAnswers map[string]string `json:"custom_answers,omitempty"`
}

// DispatchExternalApplicationRequest marks an external bundle as dispatched
type DispatchExternalApplicationRequest struct {
	Bundle ExternalApplicationBundle `json:"bundle"`
}

// ConfirmExternalApplicationRequest records explicit external submission
type ConfirmExternalApplicationRequest struct {
	Bundle              ExternalApplicationBundle `json:"bundle"`
	SubmissionReference string                    `json:"submission_reference,omitempty"`
	ConfirmationNotes   string                    `json:"confirmation_notes,omitempty"`
}

// ConfirmExternalApplicationResponse returns confirmed bundle and durable ledger record
type ConfirmExternalApplicationResponse struct {
	Bundle            ExternalApplicationBundle `json:"bundle"`
	ApplicationRecord ApplicationRecord         `json:"application_record"`
}

// ClassifyPortal classifies a target job portal URL (CAR-14, AT-010)
func (s *CareerService) ClassifyPortal(ctx context.Context, req ClassifyPortalRequest) (*ClassifyPortalResponse, error) {
	info := ClassifyPortal(req.URL, req.RedirectTarget)
	return &ClassifyPortalResponse{
		PortalInfo: info,
	}, nil
}

// PrepareExternalBundle prepares structured clipboard materials for assisted manual apply (CAR-14, CAR-11)
func (s *CareerService) PrepareExternalBundle(ctx context.Context, userID string, req PrepareExternalBundleRequest) (*ExternalApplicationBundle, error) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	savedJob, err := s.repo.GetSavedJob(ctx, userID, req.JobID)
	if err != nil {
		return nil, err
	}

	discJob := DiscoveredJob{
		ID:             savedJob.JobID,
		Title:          savedJob.Title,
		Company:        savedJob.Company,
		DirectApplyURL: savedJob.DirectApplyURL,
		CanonicalURL:   savedJob.CanonicalURL,
	}

	// Fetch tailored resume or fallback to master profile resume
	var tailoredResume TailoredResume
	tailoredList, _ := s.repo.ListTailoredResumes(ctx, userID)
	for _, tr := range tailoredList {
		if tr.JobID == savedJob.JobID {
			tailoredResume = tr
			break
		}
	}
	if tailoredResume.ID == "" {
		tailoredResume = TailoredResume{
			ID:       "master_profile_resume",
			FileName: fmt.Sprintf("%s_Resume.pdf", strings.ReplaceAll(profile.Contact.FullName, " ", "_")),
		}
	}

	// Fetch cover letter if generated
	var coverLetter CoverLetter
	letters, _ := s.repo.ListCoverLetters(ctx, userID)
	for _, cl := range letters {
		if cl.JobID == savedJob.JobID {
			coverLetter = cl
			break
		}
	}

	sessionID := fmt.Sprintf("ext_sess_%d", time.Now().UnixNano())
	return PrepareExternalApplication(sessionID, discJob, *profile, tailoredResume, coverLetter, req.CustomAnswers)
}

// DispatchExternalApplication transitions bundle to dispatched status (AT-005)
func (s *CareerService) DispatchExternalApplication(ctx context.Context, req DispatchExternalApplicationRequest) (*ExternalApplicationBundle, error) {
	b := req.Bundle
	return DispatchExternalApplication(&b)
}

// ConfirmExternalApplication verifies and saves the application record in the permanent ledger (REQ-005, AT-005, CAR-10)
func (s *CareerService) ConfirmExternalApplication(ctx context.Context, userID string, req ConfirmExternalApplicationRequest) (*ConfirmExternalApplicationResponse, error) {
	b := req.Bundle
	confirmedBundle, appRecord, err := ConfirmExternalApplication(&b, userID, req.SubmissionReference, req.ConfirmationNotes)
	if err != nil {
		return nil, err
	}

	// Persist application in permanent ledger (CAR-10)
	if err := s.repo.SaveApplicationRecord(ctx, appRecord); err != nil {
		return nil, err
	}

	// Update saved job status to applied if present
	if savedJob, err := s.repo.GetSavedJob(ctx, userID, b.JobID); err == nil && savedJob != nil {
		savedJob.Status = SavedJobStatusApplied
		now := time.Now().UTC()
		savedJob.AppliedAt = &now
		_ = s.repo.SaveSavedJob(ctx, savedJob)
	}

	return &ConfirmExternalApplicationResponse{
		Bundle:            *confirmedBundle,
		ApplicationRecord: *appRecord,
	}, nil
}

// ==========================================
// Application Status & Verified Applied Mark (IMP-CAR-15, CAR-15, REQ-005, REQ-016, AT-005, AT-006)
// ==========================================

type VerifyAppliedMarkRequest struct {
	ApplicationRecordID string                  `json:"application_record_id"`
	VerificationType    AppliedVerificationType `json:"verification_type"`
	ReceiptID           string                  `json:"receipt_id"`
	ProviderReference   string                  `json:"provider_reference,omitempty"`
	AttestationNotes    string                  `json:"attestation_notes,omitempty"`
}

type VerifyAppliedMarkResponse struct {
	Record        ApplicationRecord        `json:"record"`
	TimelineEvent ApplicationTimelineEvent `json:"timeline_event"`
}

type ReconcileApplicationRequest struct {
	SessionID string               `json:"session_id"`
	Action    ReconciliationAction `json:"action"` // confirm_applied, mark_abandoned, retry_dispatch
	Receipt   *SubmissionReceipt   `json:"receipt,omitempty"`
	Notes     string               `json:"notes,omitempty"`
}

type ReconcileApplicationResponse struct {
	Session       ApplicationReviewSession `json:"session"`
	Record        *ApplicationRecord       `json:"record,omitempty"`
	TimelineEvent ApplicationTimelineEvent `json:"timeline_event"`
}

type UpdateApplicationStageRequest struct {
	ApplicationRecordID string           `json:"application_record_id"`
	NewStage            ApplicationStage `json:"new_stage"`
	Notes               string           `json:"notes,omitempty"`
}

type UpdateApplicationStageResponse struct {
	Record        ApplicationRecord        `json:"record"`
	TimelineEvent ApplicationTimelineEvent `json:"timeline_event"`
}

type GetApplicationAuditLedgerResponse struct {
	Record   ApplicationRecord          `json:"record"`
	Timeline []ApplicationTimelineEvent `json:"timeline"`
}

// VerifyAppliedMark explicitly stamps an application record as verified applied (REQ-005, REQ-016, AT-005).
func (s *CareerService) VerifyAppliedMark(ctx context.Context, userID string, req VerifyAppliedMarkRequest) (*VerifyAppliedMarkResponse, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}

	record, err := s.repo.GetApplicationRecord(ctx, userID, req.ApplicationRecordID)
	if err != nil {
		return nil, err
	}

	if record.IsVerifiedApplied {
		return nil, ErrApplicationAlreadyApplied
	}

	if req.ReceiptID == "" && strings.TrimSpace(req.AttestationNotes) == "" {
		return nil, ErrMissingVerification
	}

	now := time.Now().UTC()
	vType := req.VerificationType
	if vType == "" {
		vType = VerificationUserAttestation
	}

	vDetails := &AppliedVerificationDetails{
		VerifiedAt:        now,
		VerificationType:  vType,
		ReceiptID:         req.ReceiptID,
		ProviderReference: req.ProviderReference,
		AttestationNotes:  req.AttestationNotes,
		VerifierID:        userID,
	}

	fromStage := record.Stage
	if fromStage == "" {
		fromStage = StageDispatched
	}

	record.IsVerifiedApplied = true
	record.Stage = StageApplied
	record.Status = "applied"
	record.AppliedAt = now
	record.UpdatedAt = now
	record.VerificationDetails = vDetails

	event := ApplicationTimelineEvent{
		EventID:             fmt.Sprintf("evt_ver_%d", now.UnixNano()),
		ApplicationRecordID: record.ID,
		Timestamp:           now,
		FromStage:           fromStage,
		ToStage:             StageApplied,
		Trigger:             "verify_applied_mark",
		ActorID:             userID,
		Notes:               req.AttestationNotes,
		Verification:        vDetails,
	}
	record.Timeline = append(record.Timeline, event)

	if err := s.repo.UpdateApplicationRecord(ctx, record); err != nil {
		return nil, err
	}

	return &VerifyAppliedMarkResponse{
		Record:        *record,
		TimelineEvent: event,
	}, nil
}

// ReconcileApplication resolves a session in dispatched or needs_confirmation status (AT-005, AT-006).
func (s *CareerService) ReconcileApplication(ctx context.Context, userID string, req ReconcileApplicationRequest) (*ReconcileApplicationResponse, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}

	session, err := s.repo.GetReviewSession(ctx, userID, req.SessionID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	event, err := ReconcileDispatchedSession(session, req.Action, req.Receipt, req.Notes, userID, now)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateReviewSession(ctx, session); err != nil {
		return nil, err
	}

	var appRecord *ApplicationRecord
	if req.Action == ReconcileActionConfirmApplied {
		// Look up or create permanent ApplicationRecord in ledger (CAR-10, REQ-005)
		if session.ApplicationRecordID != "" {
			if existing, err := s.repo.GetApplicationRecord(ctx, userID, session.ApplicationRecordID); err == nil && existing != nil {
				existing.IsVerifiedApplied = true
				existing.Stage = StageApplied
				existing.Status = "applied"
				existing.AppliedAt = now
				existing.UpdatedAt = now
				existing.VerificationDetails = event.Verification
				existing.Timeline = append(existing.Timeline, *event)
				_ = s.repo.UpdateApplicationRecord(ctx, existing)
				appRecord = existing
			}
		}

		if appRecord == nil {
			recID := fmt.Sprintf("rec_app_%d", now.UnixNano())
			appRecord = &ApplicationRecord{
				ID:                recID,
				UserID:            userID,
				JobID:             session.JobID,
				CanonicalURL:      session.ApplyURL,
				Title:             session.JobTitle,
				Company:           session.Company,
				Status:            "applied",
				Stage:             StageApplied,
				SubmissionMode:    string(session.ExecutionMode),
				IsVerifiedApplied: true,
				AppliedAt:         now,
				UpdatedAt:         now,
				Notes:             req.Notes,
				VerificationDetails: event.Verification,
				Timeline:          []ApplicationTimelineEvent{*event},
			}
			_ = s.repo.SaveApplicationRecord(ctx, appRecord)
			session.ApplicationRecordID = recID
			_ = s.repo.UpdateReviewSession(ctx, session)
		}
	}

	return &ReconcileApplicationResponse{
		Session:       *session,
		Record:        appRecord,
		TimelineEvent: *event,
	}, nil
}

// UpdateApplicationStage transitions an application through downstream lifecycle stages (CAR-15, SRC-C8).
func (s *CareerService) UpdateApplicationStage(ctx context.Context, userID string, req UpdateApplicationStageRequest) (*UpdateApplicationStageResponse, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}

	record, err := s.repo.GetApplicationRecord(ctx, userID, req.ApplicationRecordID)
	if err != nil {
		return nil, err
	}

	currentStage := record.Stage
	if currentStage == "" {
		currentStage = StageApplied
	}

	if !IsValidStageTransition(currentStage, req.NewStage) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStageTransition, currentStage, req.NewStage)
	}

	now := time.Now().UTC()
	record.Stage = req.NewStage
	record.UpdatedAt = now

	// Synchronize legacy status string
	switch req.NewStage {
	case StageApplied:
		record.Status = "applied"
	case StageInterviewing:
		record.Status = "interviewing"
	case StageOffered:
		record.Status = "offered"
	case StageRejected:
		record.Status = "rejected"
	case StageWithdrawn:
		record.Status = "withdrawn"
	case StageArchived:
		record.Status = "archived"
	}

	event := ApplicationTimelineEvent{
		EventID:             fmt.Sprintf("evt_stage_%d", now.UnixNano()),
		ApplicationRecordID: record.ID,
		Timestamp:           now,
		FromStage:           currentStage,
		ToStage:             req.NewStage,
		Trigger:             "manual_stage_update",
		ActorID:             userID,
		Notes:               req.Notes,
	}
	record.Timeline = append(record.Timeline, event)

	if err := s.repo.UpdateApplicationRecord(ctx, record); err != nil {
		return nil, err
	}

	return &UpdateApplicationStageResponse{
		Record:        *record,
		TimelineEvent: event,
	}, nil
}

// GetApplicationAuditLedger returns the immutable timeline history and verification details (CAR-15).
func (s *CareerService) GetApplicationAuditLedger(ctx context.Context, userID, recordID string) (*GetApplicationAuditLedgerResponse, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}

	record, err := s.repo.GetApplicationRecord(ctx, userID, recordID)
	if err != nil {
		return nil, err
	}

	return &GetApplicationAuditLedgerResponse{
		Record:   *record,
		Timeline: record.Timeline,
	}, nil
}

// CheckAndExpireDispatchedSessions scans all dispatched sessions and transitions timed-out ones to needs_confirmation (AT-005).
func (s *CareerService) CheckAndExpireDispatchedSessions(ctx context.Context) ([]ApplicationReviewSession, error) {
	sessions, err := s.repo.ListDispatchedReviewSessions(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var transitioned []ApplicationReviewSession

	for _, session := range sessions {
		if session.Status == WorkflowStatusDispatched {
			expired, newStatus := EvaluateSessionTimeout(&session, now)
			if expired && newStatus == WorkflowStatusNeedsConfirmation {
				session.Status = WorkflowStatusNeedsConfirmation
				session.UpdatedAt = now
				if err := s.repo.UpdateReviewSession(ctx, &session); err == nil {
					transitioned = append(transitioned, session)
				}
			}
		}
	}

	return transitioned, nil
}

// Application Evidence Bundle Methods (IMP-CAR-16, CAR-16, AT-005, AT-011)

type CreateEvidenceBundleRequest struct {
	ApplicationRecordID string                  `json:"application_record_id"`
	JobSnapshot         JobSnapshot             `json:"job_snapshot"`
	ResumeArtifact      ResumeArtifactSnapshot  `json:"resume_artifact"`
	CoverLetterArtifact *CoverLetterSnapshot    `json:"cover_letter_artifact,omitempty"`
	QuestionAnswers     []SubmittedAnswer       `json:"question_answers"`
	SubmissionTimestamp time.Time               `json:"submission_timestamp"`
	ConfirmationType    AppliedVerificationType `json:"confirmation_type"`
	ProviderReference   string                  `json:"provider_reference"`
	ScrubbedArtifacts   []ScrubbedArtifactItem  `json:"scrubbed_artifacts,omitempty"`
}

type CreateEvidenceBundleResponse struct {
	Bundle ApplicationEvidenceBundle `json:"bundle"`
}

type GetEvidenceBundleRequest struct {
	BundleID            string `json:"bundle_id"`
	ApplicationRecordID string `json:"application_record_id,omitempty"`
	RequestingUserID    string `json:"requesting_user_id"`
	IsWorkspaceAdmin    bool   `json:"is_workspace_admin"`
	HasExplicitGrant    bool   `json:"has_explicit_grant"`
}

type GetEvidenceBundleResponse struct {
	Bundle            ApplicationEvidenceBundle `json:"bundle"`
	IsIntegrityValid  bool                      `json:"is_integrity_valid"`
	IntegrityChecksum string                    `json:"integrity_checksum"`
}

type VerifyBundleIntegrityRequest struct {
	BundleID string `json:"bundle_id"`
}

type VerifyBundleIntegrityResponse struct {
	BundleID         string `json:"bundle_id"`
	IsValid          bool   `json:"is_valid"`
	ExpectedChecksum string `json:"expected_checksum"`
	ComputedChecksum string `json:"computed_checksum"`
}

// CreateEvidenceBundle seals and stores an immutable application evidence bundle (CAR-16, AT-005, AT-011).
func (s *CareerService) CreateEvidenceBundle(ctx context.Context, userID string, req CreateEvidenceBundleRequest) (*CreateEvidenceBundleResponse, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}

	bundleID := fmt.Sprintf("bundle_%s_%d", req.ApplicationRecordID, time.Now().UnixNano())
	subTime := req.SubmissionTimestamp
	if subTime.IsZero() {
		subTime = time.Now().UTC()
	}

	bundle, err := AssembleEvidenceBundle(
		bundleID,
		req.ApplicationRecordID,
		userID,
		req.JobSnapshot,
		req.ResumeArtifact,
		req.CoverLetterArtifact,
		req.QuestionAnswers,
		subTime,
		req.ConfirmationType,
		req.ProviderReference,
		req.ScrubbedArtifacts,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveEvidenceBundle(ctx, bundle); err != nil {
		return nil, err
	}

	return &CreateEvidenceBundleResponse{
		Bundle: *bundle,
	}, nil
}

// GetEvidenceBundle retrieves an evidence bundle subject to AT-011 owner-privacy checks.
func (s *CareerService) GetEvidenceBundle(ctx context.Context, req GetEvidenceBundleRequest) (*GetEvidenceBundleResponse, error) {
	var bundle *ApplicationEvidenceBundle
	var err error

	if req.BundleID != "" {
		bundle, err = s.repo.GetEvidenceBundle(ctx, req.BundleID)
	} else if req.ApplicationRecordID != "" {
		bundle, err = s.repo.GetEvidenceBundleByApplicationID(ctx, req.ApplicationRecordID)
	} else {
		return nil, ErrInvalidEvidenceBundle
	}

	if err != nil {
		return nil, err
	}

	// AT-011 / REQ-018: Enforce owner privacy authorization
	if err := AuthorizeEvidenceAccess(bundle, req.RequestingUserID, req.IsWorkspaceAdmin, req.HasExplicitGrant); err != nil {
		return nil, err
	}

	valid := VerifyEvidenceIntegrity(bundle)

	return &GetEvidenceBundleResponse{
		Bundle:            *bundle,
		IsIntegrityValid:  valid,
		IntegrityChecksum: bundle.IntegrityChecksum,
	}, nil
}

// VerifyEvidenceBundleIntegrity checks if an evidence bundle has been tampered with.
func (s *CareerService) VerifyEvidenceBundleIntegrity(ctx context.Context, bundleID string) (*VerifyBundleIntegrityResponse, error) {
	if bundleID == "" {
		return nil, ErrInvalidEvidenceBundle
	}

	bundle, err := s.repo.GetEvidenceBundle(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	computed, err := ComputeEvidenceChecksum(bundle)
	if err != nil {
		return nil, err
	}

	valid := strings.EqualFold(computed, bundle.IntegrityChecksum)

	return &VerifyBundleIntegrityResponse{
		BundleID:         bundle.BundleID,
		IsValid:          valid,
		ExpectedChecksum: bundle.IntegrityChecksum,
		ComputedChecksum: computed,
	}, nil
}

// GetEvidenceBundleByApplicationID retrieves evidence for an application with AT-011 authorization.
func (s *CareerService) GetEvidenceBundleByApplicationID(ctx context.Context, appID, requestingUserID string, isWorkspaceAdmin, hasExplicitGrant bool) (*GetEvidenceBundleResponse, error) {
	return s.GetEvidenceBundle(ctx, GetEvidenceBundleRequest{
		ApplicationRecordID: appID,
		RequestingUserID:    requestingUserID,
		IsWorkspaceAdmin:    isWorkspaceAdmin,
		HasExplicitGrant:    hasExplicitGrant,
	})
}

// --- Google Sheets Export & Sync (IMP-CAR-17, CAR-17, REQ-006, AT-013, AT-014) ---

func (s *CareerService) ConfigureGoogleSheetsSync(ctx context.Context, userID string, config SheetsSyncConfig) (*SheetsSyncConfig, error) {
	if config.SpreadsheetID == "" || config.SheetName == "" {
		return nil, ErrInvalidSheetsConfig
	}
	if config.SyncMode == "" {
		config.SyncMode = SyncModeOneWayUpsert
	}
	config.UserID = userID
	config.CreatedAt = time.Now().UTC()

	if err := s.repo.SaveSheetsSyncConfig(ctx, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *CareerService) GetGoogleSheetsSyncConfig(ctx context.Context, userID string) (*SheetsSyncConfig, error) {
	return s.repo.GetSheetsSyncConfig(ctx, userID)
}

func (s *CareerService) GetGoogleSheetsPreview(ctx context.Context, userID string) (*SheetsExportPayload, error) {
	apps, err := s.repo.ListApplications(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve applications: %w", err)
	}

	config, err := s.repo.GetSheetsSyncConfig(ctx, userID)
	if err != nil {
		// Default config if not yet customized
		config = &SheetsSyncConfig{
			SpreadsheetID: "preview-target-sheet",
			SheetName:     "Job Applications",
			Timezone:      "UTC",
			SyncMode:      SyncModeOneWayUpsert,
			IncludeNotes:  true,
		}
	}

	loc := time.UTC
	if config.Timezone != "" {
		if l, err := time.LoadLocation(config.Timezone); err == nil {
			loc = l
		}
	}

	rows := make([][]string, len(apps))
	for i, app := range apps {
		rows[i] = BuildApplicationSheetRow(app, loc, config.IncludeNotes)
	}

	csvStr, err := GenerateSheetsCSV(DefaultSheetHeaders, rows)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSV preview: %w", err)
	}

	return &SheetsExportPayload{
		SpreadsheetID: config.SpreadsheetID,
		SheetName:     config.SheetName,
		Headers:       DefaultSheetHeaders,
		Rows:          rows,
		CSVContent:    csvStr,
		GeneratedAt:   time.Now().UTC(),
		RowCount:      len(rows),
	}, nil
}

func (s *CareerService) ExportApplicationsToCSV(ctx context.Context, userID string) (string, error) {
	preview, err := s.GetGoogleSheetsPreview(ctx, userID)
	if err != nil {
		return "", err
	}
	return preview.CSVContent, nil
}

func (s *CareerService) SyncApplicationsToGoogleSheets(
	ctx context.Context,
	userID string,
	existingRows [][]string,
) (*SheetsSyncResult, [][]string, error) {
	config, err := s.repo.GetSheetsSyncConfig(ctx, userID)
	if err != nil {
		// Fallback to default one_way_upsert configuration
		config = &SheetsSyncConfig{
			UserID:        userID,
			SpreadsheetID: "default-applications-sheet",
			SheetName:     "Job Applications",
			Timezone:      "UTC",
			SyncMode:      SyncModeOneWayUpsert,
			IncludeNotes:  true,
		}
	}

	apps, err := s.repo.ListApplications(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list applications: %w", err)
	}

	reconciledRows, result, err := ReconcileSheetProjection(existingRows, apps, *config)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now().UTC()
	config.LastSyncedAt = &now
	_ = s.repo.SaveSheetsSyncConfig(ctx, config)
	_ = s.repo.SaveSheetsSyncResult(ctx, userID, &result)

	return &result, reconciledRows, nil
}

// --- Multi-Domain CSV Export & Audit Service Methods (IMP-CAR-18, CAR-18, REQ-006, AT-013, AT-022) ---

func (s *CareerService) ExportDatasetCSV(
	ctx context.Context,
	userID string,
	datasetType ExportDatasetType,
	opts ExportFilterOptions,
) (*CSVExportManifest, error) {
	var manifest *CSVExportManifest
	var err error

	switch datasetType {
	case ExportDatasetJobs:
		jobs, listErr := s.repo.ListSavedJobs(ctx, userID, "")
		if listErr != nil {
			return nil, fmt.Errorf("failed to list saved jobs: %w", listErr)
		}
		manifest, err = GenerateJobsCSV(userID, jobs, opts)

	case ExportDatasetApplications:
		apps, listErr := s.repo.ListApplications(ctx, userID)
		if listErr != nil {
			return nil, fmt.Errorf("failed to list applications: %w", listErr)
		}
		manifest, err = GenerateApplicationsCSVExtended(userID, apps, opts)

	case ExportDatasetContacts:
		contacts, listErr := s.repo.ListRecruiterContacts(ctx, userID)
		if listErr != nil {
			return nil, fmt.Errorf("failed to list contacts: %w", listErr)
		}
		manifest, err = GenerateContactsCSV(userID, contacts, opts)

	case ExportDatasetAnalysis:
		metrics, metricErr := s.GetCareerAnalyticsMetrics(ctx, userID)
		if metricErr != nil {
			return nil, fmt.Errorf("failed to gather analytics metrics: %w", metricErr)
		}
		manifest, err = GenerateAnalysisCSV(userID, metrics, opts)

	default:
		return nil, fmt.Errorf("unsupported dataset type %q for CSV export", datasetType)
	}

	if err != nil {
		return nil, err
	}

	// Persist audit record (EXP-003, AT-022)
	auditRec := ExportAuditRecord{
		AuditID:     fmt.Sprintf("audit_%s_%d", datasetType, time.Now().UnixNano()),
		UserID:      userID,
		DatasetType: datasetType,
		Filename:    manifest.Filename,
		RowCount:    manifest.RowCount,
		Format:      "csv",
		WithBOM:     manifest.WithBOM,
		ExportedAt:  time.Now().UTC(),
	}
	_ = s.repo.SaveExportAuditRecord(ctx, &auditRec)

	return manifest, nil
}

func (s *CareerService) GetExportAuditHistory(ctx context.Context, userID string) ([]ExportAuditRecord, error) {
	return s.repo.ListExportAuditRecords(ctx, userID)
}

func (s *CareerService) SaveRecruiterContact(ctx context.Context, contact *RecruiterContact) error {
	if contact.ID == "" {
		contact.ID = fmt.Sprintf("contact_%d", time.Now().UnixNano())
	}
	if contact.CreatedAt.IsZero() {
		contact.CreatedAt = time.Now().UTC()
	}
	return s.repo.SaveRecruiterContact(ctx, contact)
}

func (s *CareerService) ListRecruiterContacts(ctx context.Context, userID string) ([]RecruiterContact, error) {
	return s.repo.ListRecruiterContacts(ctx, userID)
}

func (s *CareerService) GetCareerAnalyticsMetrics(ctx context.Context, userID string) ([]CareerFunnelMetric, error) {
	apps, _ := s.repo.ListApplications(ctx, userID)
	jobs, _ := s.repo.ListSavedJobs(ctx, userID, "")

	totalApps := len(apps)
	interviewCount := 0
	verifiedCount := 0
	for _, a := range apps {
		if a.Stage == StageInterviewing || a.Stage == StageOffered {
			interviewCount++
		}
		if a.IsVerifiedApplied {
			verifiedCount++
		}
	}

	conversionRate := "0.0%"
	if totalApps > 0 {
		conversionRate = fmt.Sprintf("%.1f%%", float64(interviewCount)/float64(totalApps)*100.0)
	}

	return []CareerFunnelMetric{
		{
			MetricKey: "total_saved_jobs",
			Category:  "sourcing",
			Label:     "Total Opportunities Saved",
			Value:     fmt.Sprintf("%d", len(jobs)),
			Unit:      "count",
			Benchmark: "50",
			Notes:     "Active tracked target opportunities.",
		},
		{
			MetricKey: "total_applications",
			Category:  "funnel",
			Label:     "Total Applications Logged",
			Value:     fmt.Sprintf("%d", totalApps),
			Unit:      "count",
			Benchmark: "30",
			Notes:     "Formal applications recorded in platform ledger.",
		},
		{
			MetricKey: "verified_applied_count",
			Category:  "funnel",
			Label:     "Verified Applications (REQ-005)",
			Value:     fmt.Sprintf("%d", verifiedCount),
			Unit:      "count",
			Benchmark: "30",
			Notes:     "Applications backed by direct receipt or explicit attestation.",
		},
		{
			MetricKey: "interview_conversion_rate",
			Category:  "funnel",
			Label:     "Interview Funnel Conversion Rate",
			Value:     conversionRate,
			Unit:      "percentage",
			Benchmark: "15.0%",
			Notes:     "Percentage of applications advancing to interview rounds.",
		},
	}, nil
}

// --- Daily Report and Reminders Service Methods (IMP-CAR-19, CAR-19, AT-018) ---

// GetDailyReport generates or retrieves the daily digest for a specific date (default today).
func (s *CareerService) GetDailyReport(ctx context.Context, userID, workspaceID, dateStr string) (*DailyReport, error) {
	now := time.Now().UTC()
	if dateStr == "" {
		dateStr = now.Format("2006-01-02")
	}

	// First try to get cached/saved report for date
	existing, err := s.repo.GetDailyReport(ctx, userID, dateStr)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Fetch dependencies to generate on demand
	profile, _ := s.GetMasterProfile(ctx, userID)
	apps, _ := s.repo.ListApplications(ctx, userID)
	savedJobs, _ := s.repo.ListSavedJobs(ctx, userID, "")
	reminders, _ := s.repo.ListReminders(ctx, userID, "")
	config, err := s.repo.GetDailyReportConfig(ctx, userID)
	if err != nil || config == nil {
		config = DefaultDailyReportConfig(userID, workspaceID)
		_ = s.repo.SaveDailyReportConfig(ctx, config)
	}

	report, err := GenerateDailyReport(userID, workspaceID, dateStr, profile, apps, savedJobs, reminders, config, now)
	if err != nil {
		return nil, fmt.Errorf("failed to generate daily report: %w", err)
	}

	_ = s.repo.SaveDailyReport(ctx, report)
	return report, nil
}

// GetDailyReportConfig retrieves the user's scheduled delivery preferences.
func (s *CareerService) GetDailyReportConfig(ctx context.Context, userID, workspaceID string) (*DailyReportConfig, error) {
	config, err := s.repo.GetDailyReportConfig(ctx, userID)
	if err != nil || config == nil {
		config = DefaultDailyReportConfig(userID, workspaceID)
		_ = s.repo.SaveDailyReportConfig(ctx, config)
	}
	return config, nil
}

// UpdateDailyReportConfig updates report schedule and delivery channels, recalculating next delivery (AT-018).
func (s *CareerService) UpdateDailyReportConfig(ctx context.Context, userID, workspaceID string, update *DailyReportConfig) (*DailyReportConfig, error) {
	if update == nil {
		return nil, errors.New("daily report config cannot be nil")
	}

	update.UserID = userID
	update.WorkspaceID = workspaceID

	// Validate and calculate next delivery
	nextDelivery, err := CalculateNextReportDelivery(time.Now().UTC(), update.Timezone, update.ScheduledHour, update.ScheduledMinute, update.DeliveryDays)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next delivery: %w", err)
	}
	update.NextDeliveryUTC = nextDelivery
	update.UpdatedAt = time.Now().UTC()

	// Truth-in-advertising (AT-010): require valid URL if webhook is enabled
	if update.Channels.WebhookEnabled && strings.TrimSpace(update.Channels.WebhookURL) == "" {
		return nil, errors.New("webhook URL is required when webhook notifications are enabled")
	}

	if err := s.repo.SaveDailyReportConfig(ctx, update); err != nil {
		return nil, fmt.Errorf("failed to save daily report config: %w", err)
	}
	return update, nil
}

// ListReminders retrieves all active/pending/snoozed reminders for user.
func (s *CareerService) ListReminders(ctx context.Context, userID string, status ReminderStatus) ([]CareerReminder, error) {
	return s.repo.ListReminders(ctx, userID, status)
}

// UpdateReminderStatus transitions reminder (snooze, dismiss, complete).
func (s *CareerService) UpdateReminderStatus(ctx context.Context, userID, reminderID string, status ReminderStatus, snoozeHours int) error {
	var snoozedUntil *time.Time
	if status == ReminderStatusSnoozed {
		if snoozeHours <= 0 {
			snoozeHours = 24
		}
		t := time.Now().UTC().Add(time.Duration(snoozeHours) * time.Hour)
		snoozedUntil = &t
	}

	return s.repo.UpdateReminderStatus(ctx, userID, reminderID, status, snoozedUntil)
}

// CreateCustomReminder creates a user-defined career reminder.
func (s *CareerService) CreateCustomReminder(ctx context.Context, userID, workspaceID string, rem *CareerReminder) (*CareerReminder, error) {
	if rem == nil {
		return nil, errors.New("reminder cannot be nil")
	}
	if strings.TrimSpace(rem.Title) == "" {
		return nil, errors.New("reminder title cannot be empty")
	}

	now := time.Now().UTC()
	if rem.ID == "" {
		rem.ID = fmt.Sprintf("rem_%s_%d", userID, now.UnixNano())
	}
	rem.UserID = userID
	rem.WorkspaceID = workspaceID
	if rem.Type == "" {
		rem.Type = ReminderTypeCustom
	}
	if rem.Priority == "" {
		rem.Priority = ReminderPriorityMedium
	}
	if rem.Status == "" {
		rem.Status = ReminderStatusPending
	}
	if rem.DueAt.IsZero() {
		rem.DueAt = now.Add(24 * time.Hour)
	}
	rem.CreatedAt = now
	rem.UpdatedAt = now

	if err := s.repo.SaveReminder(ctx, rem); err != nil {
		return nil, fmt.Errorf("failed to save reminder: %w", err)
	}
	return rem, nil
}

// --- Hiring Posts & Recruiter Leads Service Methods (IMP-CAR-20, CAR-20, AT-019) ---

// IngestHiringPost extracts hiring signals, defends against prompt injection (AT-019), and generates a recruiter lead.
func (s *CareerService) IngestHiringPost(ctx context.Context, userID string, req IngestHiringPostRequest) (*HiringPost, *RecruiterLead, error) {
	post, err := ExtractHiringPost(userID, req)
	if err != nil {
		return nil, nil, err
	}

	if err := s.repo.SaveHiringPost(ctx, post); err != nil {
		return nil, nil, fmt.Errorf("failed to save hiring post: %w", err)
	}

	// Create actionable recruiter lead
	lead := CreateLeadFromPost(post, 0)
	if err := s.repo.SaveRecruiterLead(ctx, lead); err != nil {
		return nil, nil, fmt.Errorf("failed to save recruiter lead: %w", err)
	}

	// Also link or create a unified RecruiterContact record
	contact := &RecruiterContact{
		ID:          lead.ContactID,
		UserID:      userID,
		Name:        lead.RecruiterName,
		RoleTitle:   lead.RecruiterTitle,
		Company:     lead.Company,
		Email:       lead.RecruiterEmail,
		LinkedInURL: lead.LinkedInURL,
		Status:      "lead",
		Notes:       fmt.Sprintf("Discovered from hiring post: %s", post.PostURL),
		CreatedAt:   time.Now().UTC(),
	}
	_ = s.repo.SaveRecruiterContact(ctx, contact)

	return post, lead, nil
}

// ListHiringPosts returns all ingested hiring posts for a user.
func (s *CareerService) ListHiringPosts(ctx context.Context, userID string) ([]HiringPost, error) {
	return s.repo.ListHiringPosts(ctx, userID)
}

// ListRecruiterLeads lists discovered recruiter leads with optional review status filter.
func (s *CareerService) ListRecruiterLeads(ctx context.Context, userID string, filter LeadReviewStatus) ([]RecruiterLead, error) {
	leads, err := s.repo.ListRecruiterLeads(ctx, userID)
	if err != nil {
		return nil, err
	}

	if filter == "" {
		return leads, nil
	}

	var filtered []RecruiterLead
	for _, l := range leads {
		if l.ReviewStatus == filter {
			filtered = append(filtered, l)
		}
	}
	return filtered, nil
}

// GetRecruiterLead fetches a specific recruiter lead.
func (s *CareerService) GetRecruiterLead(ctx context.Context, userID, leadID string) (*RecruiterLead, error) {
	lead, err := s.repo.GetRecruiterLead(ctx, leadID)
	if err != nil {
		return nil, err
	}
	if lead.UserID != userID {
		return nil, ErrLeadNotFound
	}
	return lead, nil
}

// ReviewRecruiterLead executes the mandatory human review gate (CAR-20, REQ-015).
func (s *CareerService) ReviewRecruiterLead(ctx context.Context, userID, leadID string, action LeadReviewStatus, reviewerID, notes string) (*RecruiterLead, error) {
	lead, err := s.GetRecruiterLead(ctx, userID, leadID)
	if err != nil {
		return nil, err
	}

	if err := ReviewRecruiterLead(lead, action, reviewerID, notes); err != nil {
		return nil, err
	}

	if err := s.repo.SaveRecruiterLead(ctx, lead); err != nil {
		return nil, fmt.Errorf("failed to update recruiter lead: %w", err)
	}

	return lead, nil
}

// GenerateLeadOutreach constructs a fact-grounded outreach draft requiring prior human review approval.
func (s *CareerService) GenerateLeadOutreach(ctx context.Context, userID, leadID string, templateType OutreachTemplateType) (*OutreachDraft, error) {
	lead, err := s.GetRecruiterLead(ctx, userID, leadID)
	if err != nil {
		return nil, err
	}

	// Retrieve candidate profile facts
	candidateProfile, _ := s.GetMasterProfile(ctx, userID)
	var facts CandidateOutreachProfile
	if candidateProfile != nil {
		facts.CandidateName = candidateProfile.Contact.FullName
		facts.CurrentTitle = candidateProfile.Contact.Headline
		facts.YearsOfExp = len(candidateProfile.Experiences)
		for _, sk := range candidateProfile.Skills {
			facts.KeySkills = append(facts.KeySkills, sk.Name)
		}
		for _, l := range candidateProfile.Links {
			if l.LinkType == "portfolio" || l.LinkType == "github" {
				facts.PortfolioURL = l.URL
				break
			}
		}
	}
	if facts.CandidateName == "" {
		facts.CandidateName = "Candidate"
	}
	if facts.CurrentTitle == "" {
		facts.CurrentTitle = "Senior Software Engineer"
	}
	if len(facts.KeySkills) == 0 {
		facts.KeySkills = []string{"Go", "Distributed Systems", "Cloud"}
	}
	if facts.YearsOfExp == 0 {
		facts.YearsOfExp = 5
	}

	draft, err := GenerateOutreachDraft(lead, facts, templateType)
	if err != nil {
		return nil, err
	}

	// Persist generated draft
	_ = s.repo.SaveRecruiterLead(ctx, lead)

	return draft, nil
}

// RecordLeadOutreachAction records user action on a lead (e.g. copy draft, mark contacted).
func (s *CareerService) RecordLeadOutreachAction(ctx context.Context, userID, leadID string, action LeadOutreachStatus) (*RecruiterLead, error) {
	lead, err := s.GetRecruiterLead(ctx, userID, leadID)
	if err != nil {
		return nil, err
	}

	if err := RecordLeadOutreachAction(lead, action); err != nil {
		return nil, err
	}

	if err := s.repo.SaveRecruiterLead(ctx, lead); err != nil {
		return nil, fmt.Errorf("failed to save lead status: %w", err)
	}

	return lead, nil
}

// LinkLeadApplication links a recruiter lead to an active platform ApplicationRecord.
func (s *CareerService) LinkLeadApplication(ctx context.Context, userID, leadID, applicationID string) (*RecruiterLead, error) {
	lead, err := s.GetRecruiterLead(ctx, userID, leadID)
	if err != nil {
		return nil, err
	}

	if err := LinkLeadToApplication(lead, applicationID); err != nil {
		return nil, err
	}

	if err := s.repo.SaveRecruiterLead(ctx, lead); err != nil {
		return nil, fmt.Errorf("failed to save linked application: %w", err)
	}

	return lead, nil
}

// ---------------------------------------------------------------------------
// Run Controls & Crash Recovery (IMP-CAR-21, CAR-21, AT-006, AT-021, REQ-017, FND-011)
// ---------------------------------------------------------------------------

func (s *CareerService) CreateCareerRun(
	ctx context.Context,
	userID, workspaceID string,
	runType CareerRunType,
	totalItems, hourlyLimit int,
) (*CareerRun, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("userID cannot be empty")
	}
	if strings.TrimSpace(string(runType)) == "" {
		runType = RunTypeDiscoveryRun
	}
	run := NewCareerRun(userID, workspaceID, runType, totalItems, hourlyLimit)

	if err := s.repo.SaveCareerRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create career run: %w", err)
	}

	return run, nil
}

func (s *CareerService) GetCareerRun(ctx context.Context, userID, runID string) (*CareerRun, error) {
	run, err := s.repo.GetCareerRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run.UserID != userID {
		return nil, ErrUnauthorized
	}
	return run, nil
}

func (s *CareerService) ListCareerRuns(ctx context.Context, userID string) ([]CareerRun, error) {
	return s.repo.ListCareerRuns(ctx, userID)
}

func (s *CareerService) ControlCareerRun(ctx context.Context, userID, runID string, action RunControlAction, reason string) (*CareerRun, error) {
	run, err := s.GetCareerRun(ctx, userID, runID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	switch action {
	case ActionPause:
		if err := PauseCareerRun(run, reason); err != nil {
			return nil, err
		}
	case ActionResume:
		if err := ResumeCareerRun(run, now); err != nil {
			return nil, err
		}
	case ActionCancel:
		if err := CancelCareerRun(run, reason); err != nil {
			return nil, err
		}
	case ActionDrain:
		if err := DrainCareerRun(run); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidRunAction, action)
	}

	if err := s.repo.SaveCareerRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to save controlled run: %w", err)
	}

	return run, nil
}

func (s *CareerService) CommitRunAttempt(
	ctx context.Context,
	userID, runID string,
	token int64,
	stage string,
	safety StageSafetyLevel,
	succeeded bool,
	diagnostics string,
) (*CareerRun, error) {
	run, err := s.GetCareerRun(ctx, userID, runID)
	if err != nil {
		return nil, err
	}

	if err := CommitAttempt(run, token, stage, safety, succeeded, diagnostics); err != nil {
		return nil, err
	}

	if err := s.repo.SaveCareerRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to save run commit attempt: %w", err)
	}

	return run, nil
}

func (s *CareerService) RecoverCareerRun(ctx context.Context, userID, runID string) (*CareerRun, bool, error) {
	run, err := s.GetCareerRun(ctx, userID, runID)
	if err != nil {
		return nil, false, err
	}

	now := time.Now().UTC()
	recovered, err := RecoverCrashedRun(run, now)
	if err != nil {
		return nil, false, err
	}

	if recovered {
		if err := s.repo.SaveCareerRun(ctx, run); err != nil {
			return nil, false, fmt.Errorf("failed to save recovered run: %w", err)
		}
	}

	return run, recovered, nil
}

func (s *CareerService) ReconcileCareerRun(ctx context.Context, userID, runID string, resolution string, markSucceeded bool, notes string) (*CareerRun, error) {
	run, err := s.GetCareerRun(ctx, userID, runID)
	if err != nil {
		return nil, err
	}

	if err := ReconcileRun(run, resolution, markSucceeded, notes); err != nil {
		return nil, err
	}

	if err := s.repo.SaveCareerRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to save reconciled run: %w", err)
	}

	return run, nil
}

// ---------------------------------------------------------------------------
// PII Controls & LLM Provider Choices (IMP-CAR-22, CAR-22, REQ-023, AT-019)
// ---------------------------------------------------------------------------

func (s *CareerService) RedactCandidatePII(
	ctx context.Context,
	userID, text string,
	policy PIIRedactionPolicy,
) (*PIIRedactionResult, error) {
	var identities CandidateIdentities

	// Extract confirmed candidate identities from profile if available
	profile, err := s.repo.GetProfile(ctx, userID)
	if err == nil && profile != nil {
		identities.FullName = profile.Contact.FullName
		if profile.Contact.Email != "" {
			identities.Emails = append(identities.Emails, profile.Contact.Email)
		}
		if profile.Contact.Phone != "" {
			identities.Phones = append(identities.Phones, profile.Contact.Phone)
		}
		if profile.Contact.Location != "" {
			identities.Locations = append(identities.Locations, profile.Contact.Location)
		}
		for _, lk := range profile.Links {
			if lk.URL != "" {
				identities.Links = append(identities.Links, lk.URL)
			}
		}
	}

	// Also extract compensation from preferences if present
	pref, err := s.repo.GetPreferences(ctx, userID)
	if err == nil && pref != nil && pref.Salary.TargetAmount > 0 {
		identities.Compensation = append(identities.Compensation, fmt.Sprintf("$%.0f", pref.Salary.TargetAmount))
	}

	return RedactPII(text, identities, policy)
}

func (s *CareerService) RehydrateCandidateText(anonymizedText string, tokenMap map[string]string) string {
	return RehydrateText(anonymizedText, tokenMap)
}

func (s *CareerService) GrantProviderConsent(
	ctx context.Context,
	userID, providerID, providerName string,
	allowedTasks []string,
	zeroTrainingAffirmed bool,
	retentionDays int,
) (*UserProviderConsent, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("userID cannot be empty")
	}
	if strings.TrimSpace(providerID) == "" {
		return nil, errors.New("providerID cannot be empty")
	}
	if strings.TrimSpace(providerName) == "" {
		providerName = providerID
	}
	if retentionDays <= 0 {
		retentionDays = 30
	}

	consent := &UserProviderConsent{
		ID:                   fmt.Sprintf("consent_%s_%s", userID, providerID),
		UserID:               userID,
		ProviderID:           providerID,
		ProviderName:         providerName,
		Status:               "active",
		AllowedTasks:         allowedTasks,
		ZeroTrainingAffirmed: zeroTrainingAffirmed,
		ConsentedAt:          time.Now().UTC(),
		RetentionDays:        retentionDays,
	}

	if err := s.repo.SaveProviderConsent(ctx, consent); err != nil {
		return nil, fmt.Errorf("failed to save provider consent: %w", err)
	}

	return consent, nil
}

func (s *CareerService) RevokeProviderConsent(ctx context.Context, userID, providerID string) (*UserProviderConsent, error) {
	consent, err := s.repo.GetProviderConsent(ctx, userID, providerID)
	if err != nil {
		return nil, err
	}
	if consent.Status == "revoked" {
		return nil, ErrConsentAlreadyRevoked
	}

	if err := s.repo.RevokeProviderConsent(ctx, userID, providerID); err != nil {
		return nil, err
	}

	return s.repo.GetProviderConsent(ctx, userID, providerID)
}

func (s *CareerService) GetProviderConsent(ctx context.Context, userID, providerID string) (*UserProviderConsent, error) {
	return s.repo.GetProviderConsent(ctx, userID, providerID)
}

func (s *CareerService) ListProviderConsents(ctx context.Context, userID string) ([]UserProviderConsent, error) {
	return s.repo.ListProviderConsents(ctx, userID)
}

func (s *CareerService) VerifyProviderConsentForTask(ctx context.Context, userID, providerID, taskKey string) error {
	consent, err := s.repo.GetProviderConsent(ctx, userID, providerID)
	if err != nil {
		if errors.Is(err, ErrConsentNotFound) {
			return ErrProviderConsentRequired
		}
		return err
	}

	if consent.Status != "active" {
		return ErrProviderConsentRequired
	}

	if len(consent.AllowedTasks) > 0 {
		allowed := false
		for _, t := range consent.AllowedTasks {
			if t == taskKey || t == "*" {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("%w: task '%s' is not authorized for provider '%s'", ErrProviderConsentRequired, taskKey, providerID)
		}
	}

	return nil
}

// GenerateInterviewPrep synthesizes company brief and tailored interview questions grounded in verified candidate facts (CAR-23, AT-003).
func (s *CareerService) GenerateInterviewPrep(ctx context.Context, appID, candidateID string, company CompanyBrief, jobTitle string, requiredSkills []string) (*InterviewPreparationPack, error) {
	var profile *MasterCareerProfile
	if candidateID != "" {
		p, err := s.repo.GetProfile(ctx, candidateID)
		if err == nil {
			profile = p
		}
	}

	pack := CreateInterviewPreparationPack(appID, candidateID, company, jobTitle, requiredSkills, profile)
	if err := s.repo.SaveInterviewPrepPack(ctx, pack); err != nil {
		return nil, err
	}
	return pack, nil
}

func (s *CareerService) GetInterviewPrep(ctx context.Context, packID string) (*InterviewPreparationPack, error) {
	return s.repo.GetInterviewPrepPack(ctx, packID)
}

func (s *CareerService) GetInterviewPrepByApp(ctx context.Context, appID string) (*InterviewPreparationPack, error) {
	return s.repo.GetInterviewPrepPackByAppID(ctx, appID)
}

func (s *CareerService) ScheduleInterviewRound(ctx context.Context, appID, roundID, stageName string, scheduledAt time.Time, tz, format, link, interviewer string) (*InterviewSchedule, error) {
	sched, err := CreateInterviewSchedule(appID, roundID, stageName, scheduledAt, tz, format, link, interviewer)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveInterviewSchedule(ctx, sched); err != nil {
		return nil, err
	}
	return sched, nil
}

func (s *CareerService) GetInterviewSchedule(ctx context.Context, scheduleID string) (*InterviewSchedule, error) {
	return s.repo.GetInterviewSchedule(ctx, scheduleID)
}

func (s *CareerService) ListInterviewSchedules(ctx context.Context, appID string) ([]InterviewSchedule, error) {
	return s.repo.ListInterviewSchedules(ctx, appID)
}

func (s *CareerService) UpdateRoundOutcome(ctx context.Context, scheduleID string, outcome RoundOutcome) error {
	return s.repo.UpdateInterviewScheduleOutcome(ctx, scheduleID, outcome)
}

func (s *CareerService) RecordInterviewSessionNote(ctx context.Context, appID, roundID, candidateID, preNotes string, questionsAsked []string, postReflections, interviewerName, interviewerTitle string, selfRating int, followUp string) (*InterviewSessionNote, error) {
	note, err := RecordSessionNote(appID, roundID, candidateID, preNotes, questionsAsked, postReflections, interviewerName, interviewerTitle, selfRating, followUp)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveInterviewSessionNote(ctx, note); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *CareerService) GetInterviewSessionNote(ctx context.Context, roundID string) (*InterviewSessionNote, error) {
	return s.repo.GetInterviewSessionNote(ctx, roundID)
}

func (s *CareerService) ListInterviewSessionNotes(ctx context.Context, appID string) ([]InterviewSessionNote, error) {
	return s.repo.ListInterviewSessionNotes(ctx, appID)
}

func (s *CareerService) GetOutcomeFunnel(ctx context.Context, userID string) (OutcomeFunnel, error) {
	apps, err := s.repo.ListApplications(ctx, userID)
	if err != nil {
		return OutcomeFunnel{}, err
	}
	return CalculateOutcomeFunnel(apps), nil
}

// RunConsistencyAudit cross-references facts across profile, resume, and LinkedIn snapshot (CAR-24, REQ-004).
func (s *CareerService) RunConsistencyAudit(ctx context.Context, userID string) (*ConsistencyAuditReport, error) {
	profile, _ := s.repo.GetProfile(ctx, userID)
	resumes, _ := s.repo.ListResumesByUser(ctx, userID)
	var latestResume *GeneratedResume
	if len(resumes) > 0 {
		latestResume = &resumes[len(resumes)-1]
	}
	linkedIn, _ := s.repo.GetLinkedInSnapshot(ctx, userID)

	report := CompareCareerRecords(userID, profile, latestResume, linkedIn)
	if err := s.repo.SaveConsistencyReport(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (s *CareerService) GetLatestConsistencyReport(ctx context.Context, userID string) (*ConsistencyAuditReport, error) {
	return s.repo.GetLatestConsistencyReport(ctx, userID)
}

func (s *CareerService) ResolveConsistencyMismatch(ctx context.Context, userID, reportID, mismatchID string, decision ResolutionDecision) (*ConsistencyAuditReport, error) {
	report, err := s.repo.GetConsistencyReport(ctx, reportID)
	if err != nil {
		return nil, err
	}
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	updatedReport, err := ApplyConsistencyResolution(report, mismatchID, decision, profile)
	if err != nil {
		return nil, err
	}

	// Persist updated profile and report
	if err := s.repo.SaveProfile(ctx, profile); err != nil {
		return nil, err
	}
	if err := s.repo.SaveConsistencyReport(ctx, updatedReport); err != nil {
		return nil, err
	}

	return updatedReport, nil
}

func (s *CareerService) SaveLinkedInSnapshot(ctx context.Context, snapshot *LinkedInProfileSnapshot) error {
	return s.repo.SaveLinkedInSnapshot(ctx, snapshot)
}

func (s *CareerService) GetLinkedInSnapshot(ctx context.Context, userID string) (*LinkedInProfileSnapshot, error) {
	return s.repo.GetLinkedInSnapshot(ctx, userID)
}






