package career

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrDraftNotFound  = errors.New("resume extraction draft not found")
	ErrFactNotFound   = errors.New("profile fact not found")
	ErrUnauthorized   = errors.New("unauthorized access to career profile record")
	ErrResumeNotFound = errors.New("generated master resume not found")
)

// CareerRepository manages persistence for drafts and confirmed facts.
type CareerRepository interface {
	SaveDraft(ctx context.Context, draft *ResumeExtractionDraft) error
	GetDraft(ctx context.Context, userID, draftID string) (*ResumeExtractionDraft, error)
	UpdateDraftStatus(ctx context.Context, userID, draftID string, status ExtractionStatus, confirmedAt *time.Time) error

	SaveFact(ctx context.Context, fact *ProfileFactItem) error
	SaveFactsBatch(ctx context.Context, facts []ProfileFactItem) error
	GetFactsByUser(ctx context.Context, userID string, category string) ([]ProfileFactItem, error)
	GetFactByID(ctx context.Context, userID, factID string) (*ProfileFactItem, error)
	UpdateFact(ctx context.Context, fact *ProfileFactItem) error
	DeleteFact(ctx context.Context, userID, factID string) error
	DeleteAllUserFacts(ctx context.Context, userID string) error

	SaveProfile(ctx context.Context, profile *MasterCareerProfile) error
	GetProfile(ctx context.Context, userID string) (*MasterCareerProfile, error)
	DeleteProfile(ctx context.Context, userID string) error

	SavePreferences(ctx context.Context, pref *CareerPreferences) error
	GetPreferences(ctx context.Context, userID string) (*CareerPreferences, error)
	DeletePreferences(ctx context.Context, userID string) error

	SaveResume(ctx context.Context, resume *GeneratedResume) error
	GetResume(ctx context.Context, userID, resumeID string) (*GeneratedResume, error)
	ListResumesByUser(ctx context.Context, userID string) ([]GeneratedResume, error)
	DeleteResume(ctx context.Context, userID, resumeID string) error

	SaveTailoredResume(ctx context.Context, resume *TailoredResume) error
	GetTailoredResume(ctx context.Context, userID, resumeID string) (*TailoredResume, error)
	ListTailoredResumes(ctx context.Context, userID string) ([]TailoredResume, error)
	UpdateTailoredResumeStatus(ctx context.Context, userID, resumeID string, status ApprovalStatus) error

	SaveCoverLetter(ctx context.Context, letter *CoverLetter) error
	GetCoverLetter(ctx context.Context, userID, letterID string) (*CoverLetter, error)
	ListCoverLetters(ctx context.Context, userID string) ([]CoverLetter, error)
	UpdateCoverLetterStatus(ctx context.Context, userID, letterID string, status ApprovalStatus) error

	SaveFeedbackReport(ctx context.Context, report *ResumeFeedbackReport) error
	GetLatestFeedbackReport(ctx context.Context, userID string) (*ResumeFeedbackReport, error)

	// Saved Jobs & Exclusion History (IMP-CAR-10, CAR-10, AT-004)
	SaveSavedJob(ctx context.Context, job *SavedJob) error
	GetSavedJob(ctx context.Context, userID, savedJobID string) (*SavedJob, error)
	GetSavedJobByCanonicalURL(ctx context.Context, userID, canonicalURL string) (*SavedJob, error)
	ListSavedJobs(ctx context.Context, userID string, status SavedJobStatus) ([]SavedJob, error)
	DeleteSavedJob(ctx context.Context, userID, savedJobID string) error

	SaveExclusion(ctx context.Context, exclusion *JobExclusion) error
	GetExclusion(ctx context.Context, userID, exclusionID string) (*JobExclusion, error)
	ListExclusions(ctx context.Context, userID string) ([]JobExclusion, error)
	DeleteExclusion(ctx context.Context, userID, exclusionID string) error

	SaveApplicationRecord(ctx context.Context, record *ApplicationRecord) error
	GetApplicationRecord(ctx context.Context, userID, recordID string) (*ApplicationRecord, error)
	GetApplicationByCanonicalURL(ctx context.Context, userID, canonicalURL string) (*ApplicationRecord, error)
	ListApplications(ctx context.Context, userID string) ([]ApplicationRecord, error)
	UpdateApplicationRecord(ctx context.Context, record *ApplicationRecord) error

	// Application Workflow & Human Review (IMP-CAR-11, CAR-11, REQ-005, REQ-015, AT-005)
	SaveReviewSession(ctx context.Context, session *ApplicationReviewSession) error
	GetReviewSession(ctx context.Context, userID, sessionID string) (*ApplicationReviewSession, error)
	ListReviewSessions(ctx context.Context, userID string, status ApplicationWorkflowStatus) ([]ApplicationReviewSession, error)
	ListDispatchedReviewSessions(ctx context.Context) ([]ApplicationReviewSession, error)
	UpdateReviewSession(ctx context.Context, session *ApplicationReviewSession) error
	DeleteReviewSession(ctx context.Context, userID, sessionID string) error

	// Application Evidence Bundle (IMP-CAR-16, CAR-16, AT-005, AT-011)
	SaveEvidenceBundle(ctx context.Context, bundle *ApplicationEvidenceBundle) error
	GetEvidenceBundle(ctx context.Context, bundleID string) (*ApplicationEvidenceBundle, error)
	GetEvidenceBundleByApplicationID(ctx context.Context, appID string) (*ApplicationEvidenceBundle, error)
	ListEvidenceBundlesByUser(ctx context.Context, userID string) ([]ApplicationEvidenceBundle, error)

	// Google Sheets Export & Sync (IMP-CAR-17, CAR-17, REQ-006, AT-013, AT-014)
	SaveSheetsSyncConfig(ctx context.Context, config *SheetsSyncConfig) error
	GetSheetsSyncConfig(ctx context.Context, userID string) (*SheetsSyncConfig, error)
	SaveSheetsSyncResult(ctx context.Context, userID string, result *SheetsSyncResult) error
	GetLatestSheetsSyncResult(ctx context.Context, userID string) (*SheetsSyncResult, error)

	// Multi-domain CSV Export & Audit Logging (IMP-CAR-18, CAR-18, REQ-006, AT-013, AT-022)
	SaveExportAuditRecord(ctx context.Context, record *ExportAuditRecord) error
	ListExportAuditRecords(ctx context.Context, userID string) ([]ExportAuditRecord, error)
	SaveRecruiterContact(ctx context.Context, contact *RecruiterContact) error
	ListRecruiterContacts(ctx context.Context, userID string) ([]RecruiterContact, error)

	// Daily Reports & Reminders (IMP-CAR-19, CAR-19, AT-018)
	SaveDailyReport(ctx context.Context, report *DailyReport) error
	GetDailyReport(ctx context.Context, userID, date string) (*DailyReport, error)
	ListDailyReports(ctx context.Context, userID string) ([]DailyReport, error)
	SaveDailyReportConfig(ctx context.Context, config *DailyReportConfig) error
	GetDailyReportConfig(ctx context.Context, userID string) (*DailyReportConfig, error)
	SaveReminder(ctx context.Context, reminder *CareerReminder) error
	GetReminder(ctx context.Context, userID, reminderID string) (*CareerReminder, error)
	ListReminders(ctx context.Context, userID string, status ReminderStatus) ([]CareerReminder, error)
	UpdateReminderStatus(ctx context.Context, userID, reminderID string, status ReminderStatus, snoozedUntil *time.Time) error

	// Hiring Posts & Recruiter Leads (IMP-CAR-20, CAR-20, AT-019)
	SaveHiringPost(ctx context.Context, post *HiringPost) error
	GetHiringPost(ctx context.Context, id string) (*HiringPost, error)
	ListHiringPosts(ctx context.Context, userID string) ([]HiringPost, error)
	SaveRecruiterLead(ctx context.Context, lead *RecruiterLead) error
	GetRecruiterLead(ctx context.Context, id string) (*RecruiterLead, error)
	ListRecruiterLeads(ctx context.Context, userID string) ([]RecruiterLead, error)

	// Run Controls & Crash Recovery (IMP-CAR-21, CAR-21, AT-006, AT-021, REQ-017, FND-011)
	SaveCareerRun(ctx context.Context, run *CareerRun) error
	GetCareerRun(ctx context.Context, id string) (*CareerRun, error)
	ListCareerRuns(ctx context.Context, userID string) ([]CareerRun, error)

	// PII Controls & Provider Choices (IMP-CAR-22, CAR-22, REQ-023)
	SaveProviderConsent(ctx context.Context, consent *UserProviderConsent) error
	GetProviderConsent(ctx context.Context, userID, providerID string) (*UserProviderConsent, error)
	ListProviderConsents(ctx context.Context, userID string) ([]UserProviderConsent, error)
	RevokeProviderConsent(ctx context.Context, userID, providerID string) error

	// Interview Preparation & Outcomes (IMP-CAR-23, CAR-23, AT-003)
	SaveInterviewPrepPack(ctx context.Context, pack *InterviewPreparationPack) error
	GetInterviewPrepPack(ctx context.Context, packID string) (*InterviewPreparationPack, error)
	GetInterviewPrepPackByAppID(ctx context.Context, appID string) (*InterviewPreparationPack, error)
	SaveInterviewSchedule(ctx context.Context, sched *InterviewSchedule) error
	GetInterviewSchedule(ctx context.Context, scheduleID string) (*InterviewSchedule, error)
	ListInterviewSchedules(ctx context.Context, appID string) ([]InterviewSchedule, error)
	UpdateInterviewScheduleOutcome(ctx context.Context, scheduleID string, outcome RoundOutcome) error
	SaveInterviewSessionNote(ctx context.Context, note *InterviewSessionNote) error
	GetInterviewSessionNote(ctx context.Context, roundID string) (*InterviewSessionNote, error)
	ListInterviewSessionNotes(ctx context.Context, appID string) ([]InterviewSessionNote, error)

	// Resume/Profile/LinkedIn Consistency Check (IMP-CAR-24, CAR-24, REQ-004, AT-007)
	SaveConsistencyReport(ctx context.Context, report *ConsistencyAuditReport) error
	GetConsistencyReport(ctx context.Context, reportID string) (*ConsistencyAuditReport, error)
	GetLatestConsistencyReport(ctx context.Context, userID string) (*ConsistencyAuditReport, error)
	SaveLinkedInSnapshot(ctx context.Context, snapshot *LinkedInProfileSnapshot) error
	GetLinkedInSnapshot(ctx context.Context, userID string) (*LinkedInProfileSnapshot, error)
}

var (
	ErrFeedbackReportNotFound = errors.New("resume feedback report not found")
	ErrSheetsConfigNotFound   = errors.New("google sheets sync config not found")
	ErrSheetsResultNotFound   = errors.New("no sheets sync results recorded")
	ErrContactNotFound        = errors.New("recruiter contact not found")
	ErrDailyReportNotFound    = errors.New("daily report not found")
	ErrReminderNotFound       = errors.New("career reminder not found")
	ErrConsentNotFound        = errors.New("user provider consent not found")
)

// MemoryCareerRepository provides an in-memory, thread-safe implementation.
type MemoryCareerRepository struct {
	mu              sync.RWMutex
	drafts          map[string]*ResumeExtractionDraft    // draftID -> draft
	facts           map[string]*ProfileFactItem          // factID -> fact
	profiles        map[string]*MasterCareerProfile      // userID -> profile
	preferences     map[string]*CareerPreferences        // userID -> preferences
	resumes         map[string]*GeneratedResume          // resumeID -> resume
	tailoredResumes map[string]*TailoredResume           // resumeID -> tailored resume
	coverLetters    map[string]*CoverLetter              // letterID -> cover letter
	reports         map[string]*ResumeFeedbackReport     // reportID -> report
	latestReport    map[string]string                    // userID -> latest reportID
	savedJobs       map[string]*SavedJob                 // savedJobID -> SavedJob
	exclusions      map[string]*JobExclusion             // exclusionID -> JobExclusion
	applications    map[string]*ApplicationRecord        // applicationID -> ApplicationRecord
	reviewSessions  map[string]*ApplicationReviewSession // sessionID -> ApplicationReviewSession
	evidenceBundles map[string]*ApplicationEvidenceBundle // bundleID -> ApplicationEvidenceBundle
	sheetsConfigs   map[string]*SheetsSyncConfig         // userID -> SheetsSyncConfig
	sheetsResults   map[string]*SheetsSyncResult         // userID -> latest SheetsSyncResult
	exportAudits    []ExportAuditRecord                  // historical audit records (EXP-003, AT-022)
	contacts        map[string]*RecruiterContact         // contactID -> RecruiterContact
	dailyReports    map[string]*DailyReport              // reportID -> DailyReport
	dailyConfigs    map[string]*DailyReportConfig        // userID -> DailyReportConfig
	reminders       map[string]*CareerReminder           // reminderID -> CareerReminder
	hiringPosts     map[string]*HiringPost               // postID -> HiringPost
	recruiterLeads  map[string]*RecruiterLead            // leadID -> RecruiterLead
	careerRuns        map[string]*CareerRun                // runID -> CareerRun
	providerConsents   map[string]*UserProviderConsent      // userID:providerID -> UserProviderConsent
	prepPacks          map[string]*InterviewPreparationPack // packID -> InterviewPreparationPack
	schedules          map[string]*InterviewSchedule        // scheduleID -> InterviewSchedule
	sessionNotes       map[string]*InterviewSessionNote     // roundID -> InterviewSessionNote
	consistencyReports map[string]*ConsistencyAuditReport   // reportID -> ConsistencyAuditReport
	latestConsistency  map[string]string                    // userID -> reportID
	linkedInSnapshots  map[string]*LinkedInProfileSnapshot  // userID -> LinkedInProfileSnapshot
}

func NewMemoryCareerRepository() *MemoryCareerRepository {
	return &MemoryCareerRepository{
		drafts:             make(map[string]*ResumeExtractionDraft),
		facts:              make(map[string]*ProfileFactItem),
		profiles:           make(map[string]*MasterCareerProfile),
		preferences:        make(map[string]*CareerPreferences),
		resumes:            make(map[string]*GeneratedResume),
		tailoredResumes:    make(map[string]*TailoredResume),
		coverLetters:       make(map[string]*CoverLetter),
		reports:            make(map[string]*ResumeFeedbackReport),
		latestReport:       make(map[string]string),
		savedJobs:          make(map[string]*SavedJob),
		exclusions:         make(map[string]*JobExclusion),
		applications:       make(map[string]*ApplicationRecord),
		reviewSessions:     make(map[string]*ApplicationReviewSession),
		evidenceBundles:    make(map[string]*ApplicationEvidenceBundle),
		sheetsConfigs:      make(map[string]*SheetsSyncConfig),
		sheetsResults:      make(map[string]*SheetsSyncResult),
		exportAudits:       make([]ExportAuditRecord, 0),
		contacts:           make(map[string]*RecruiterContact),
		dailyReports:       make(map[string]*DailyReport),
		dailyConfigs:       make(map[string]*DailyReportConfig),
		reminders:          make(map[string]*CareerReminder),
		hiringPosts:        make(map[string]*HiringPost),
		recruiterLeads:     make(map[string]*RecruiterLead),
		careerRuns:         make(map[string]*CareerRun),
		providerConsents:   make(map[string]*UserProviderConsent),
		prepPacks:          make(map[string]*InterviewPreparationPack),
		schedules:          make(map[string]*InterviewSchedule),
		sessionNotes:       make(map[string]*InterviewSessionNote),
		consistencyReports: make(map[string]*ConsistencyAuditReport),
		latestConsistency:  make(map[string]string),
		linkedInSnapshots:  make(map[string]*LinkedInProfileSnapshot),
	}
}

func (r *MemoryCareerRepository) SaveDraft(ctx context.Context, draft *ResumeExtractionDraft) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.drafts[draft.ID] = draft
	return nil
}

func (r *MemoryCareerRepository) GetDraft(ctx context.Context, userID, draftID string) (*ResumeExtractionDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	draft, exists := r.drafts[draftID]
	if !exists {
		return nil, ErrDraftNotFound
	}
	if draft.UserID != userID {
		return nil, ErrUnauthorized
	}
	return draft, nil
}

func (r *MemoryCareerRepository) UpdateDraftStatus(ctx context.Context, userID, draftID string, status ExtractionStatus, confirmedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	draft, exists := r.drafts[draftID]
	if !exists {
		return ErrDraftNotFound
	}
	if draft.UserID != userID {
		return ErrUnauthorized
	}

	draft.Status = status
	if confirmedAt != nil {
		draft.ConfirmedAt = confirmedAt
	}
	return nil
}

func (r *MemoryCareerRepository) SaveFact(ctx context.Context, fact *ProfileFactItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.facts[fact.ID] = fact
	return nil
}

func (r *MemoryCareerRepository) SaveFactsBatch(ctx context.Context, facts []ProfileFactItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range facts {
		r.facts[facts[i].ID] = &facts[i]
	}
	return nil
}

func (r *MemoryCareerRepository) GetFactsByUser(ctx context.Context, userID string, category string) ([]ProfileFactItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []ProfileFactItem
	for _, f := range r.facts {
		if f.UserID == userID {
			if category == "" || f.Category == category {
				result = append(result, *f)
			}
		}
	}
	return result, nil
}

func (r *MemoryCareerRepository) GetFactByID(ctx context.Context, userID, factID string) (*ProfileFactItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fact, exists := r.facts[factID]
	if !exists {
		return nil, ErrFactNotFound
	}
	if fact.UserID != userID {
		return nil, ErrUnauthorized
	}
	return fact, nil
}

func (r *MemoryCareerRepository) UpdateFact(ctx context.Context, fact *ProfileFactItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.facts[fact.ID]
	if !exists {
		return ErrFactNotFound
	}
	if existing.UserID != fact.UserID {
		return ErrUnauthorized
	}

	fact.UpdatedAt = time.Now()
	r.facts[fact.ID] = fact
	return nil
}

func (r *MemoryCareerRepository) DeleteFact(ctx context.Context, userID, factID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.facts[factID]
	if !exists {
		return ErrFactNotFound
	}
	if existing.UserID != userID {
		return ErrUnauthorized
	}

	delete(r.facts, factID)
	return nil
}

func (r *MemoryCareerRepository) DeleteAllUserFacts(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, f := range r.facts {
		if f.UserID == userID {
			delete(r.facts, id)
		}
	}
	for id, d := range r.drafts {
		if d.UserID == userID {
			delete(r.drafts, id)
		}
	}
	delete(r.profiles, userID)
	delete(r.preferences, userID)
	return nil
}

func (r *MemoryCareerRepository) SaveProfile(ctx context.Context, profile *MasterCareerProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.profiles[profile.UserID] = profile
	return nil
}

func (r *MemoryCareerRepository) GetProfile(ctx context.Context, userID string) (*MasterCareerProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.profiles[userID]
	if !exists {
		return nil, ErrProfileNotFound
	}
	return p, nil
}

func (r *MemoryCareerRepository) DeleteProfile(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.profiles, userID)
	return nil
}

func (r *MemoryCareerRepository) SavePreferences(ctx context.Context, pref *CareerPreferences) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.preferences[pref.UserID] = pref
	return nil
}

func (r *MemoryCareerRepository) GetPreferences(ctx context.Context, userID string) (*CareerPreferences, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.preferences[userID]
	if !exists {
		return nil, ErrPreferencesNotFound
	}
	return p, nil
}

func (r *MemoryCareerRepository) DeletePreferences(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.preferences, userID)
	return nil
}

func (r *MemoryCareerRepository) SaveResume(ctx context.Context, resume *GeneratedResume) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.resumes[resume.ID] = resume
	return nil
}

func (r *MemoryCareerRepository) GetResume(ctx context.Context, userID, resumeID string) (*GeneratedResume, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res, exists := r.resumes[resumeID]
	if !exists {
		return nil, ErrResumeNotFound
	}
	if res.UserID != userID {
		return nil, ErrUnauthorized
	}
	return res, nil
}

func (r *MemoryCareerRepository) ListResumesByUser(ctx context.Context, userID string) ([]GeneratedResume, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []GeneratedResume
	for _, res := range r.resumes {
		if res.UserID == userID {
			result = append(result, *res)
		}
	}
	return result, nil
}

func (r *MemoryCareerRepository) DeleteResume(ctx context.Context, userID, resumeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, exists := r.resumes[resumeID]
	if !exists {
		return ErrResumeNotFound
	}
	if res.UserID != userID {
		return ErrUnauthorized
	}
	delete(r.resumes, resumeID)
	return nil
}

func (r *MemoryCareerRepository) SaveTailoredResume(ctx context.Context, resume *TailoredResume) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tailoredResumes[resume.ID] = resume
	return nil
}

func (r *MemoryCareerRepository) GetTailoredResume(ctx context.Context, userID, resumeID string) (*TailoredResume, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res, exists := r.tailoredResumes[resumeID]
	if !exists {
		return nil, ErrTailoredResumeNotFound
	}
	if res.UserID != userID {
		return nil, ErrUnauthorized
	}
	return res, nil
}

func (r *MemoryCareerRepository) ListTailoredResumes(ctx context.Context, userID string) ([]TailoredResume, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []TailoredResume
	for _, res := range r.tailoredResumes {
		if res.UserID == userID {
			result = append(result, *res)
		}
	}
	return result, nil
}

func (r *MemoryCareerRepository) UpdateTailoredResumeStatus(ctx context.Context, userID, resumeID string, status ApprovalStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, exists := r.tailoredResumes[resumeID]
	if !exists {
		return ErrTailoredResumeNotFound
	}
	if res.UserID != userID {
		return ErrUnauthorized
	}

	res.ApprovalStatus = status
	if status == ApprovalStatusApproved {
		now := time.Now().UTC()
		res.ApprovedAt = &now
	}
	return nil
}

func (r *MemoryCareerRepository) SaveCoverLetter(ctx context.Context, letter *CoverLetter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.coverLetters[letter.ID] = letter
	return nil
}

func (r *MemoryCareerRepository) GetCoverLetter(ctx context.Context, userID, letterID string) (*CoverLetter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	let, exists := r.coverLetters[letterID]
	if !exists {
		return nil, ErrCoverLetterNotFound
	}
	if let.UserID != userID {
		return nil, ErrUnauthorized
	}
	return let, nil
}

func (r *MemoryCareerRepository) ListCoverLetters(ctx context.Context, userID string) ([]CoverLetter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []CoverLetter
	for _, let := range r.coverLetters {
		if let.UserID == userID {
			result = append(result, *let)
		}
	}
	return result, nil
}

func (r *MemoryCareerRepository) UpdateCoverLetterStatus(ctx context.Context, userID, letterID string, status ApprovalStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	let, exists := r.coverLetters[letterID]
	if !exists {
		return ErrCoverLetterNotFound
	}
	if let.UserID != userID {
		return ErrUnauthorized
	}

	let.ApprovalStatus = status
	if status == ApprovalStatusApproved {
		now := time.Now().UTC()
		let.ApprovedAt = &now
	}
	return nil
}

func (r *MemoryCareerRepository) SaveFeedbackReport(ctx context.Context, report *ResumeFeedbackReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reports[report.ID] = report
	r.latestReport[report.UserID] = report.ID
	return nil
}

func (r *MemoryCareerRepository) GetLatestFeedbackReport(ctx context.Context, userID string) (*ResumeFeedbackReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reportID, exists := r.latestReport[userID]
	if !exists {
		return nil, ErrFeedbackReportNotFound
	}

	report, exists := r.reports[reportID]
	if !exists {
		return nil, ErrFeedbackReportNotFound
	}
	return report, nil
}

// --- Saved Jobs Persistence (IMP-CAR-10, CAR-10) ---

func (r *MemoryCareerRepository) SaveSavedJob(ctx context.Context, job *SavedJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.savedJobs[job.ID] = job
	return nil
}

func (r *MemoryCareerRepository) GetSavedJob(ctx context.Context, userID, savedJobID string) (*SavedJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, exists := r.savedJobs[savedJobID]
	if !exists {
		return nil, ErrSavedJobNotFound
	}
	if job.UserID != userID {
		return nil, ErrUnauthorized
	}
	return job, nil
}

func (r *MemoryCareerRepository) GetSavedJobByCanonicalURL(ctx context.Context, userID, canonicalURL string) (*SavedJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normURL := NormalizeCanonicalURL(canonicalURL)
	for _, job := range r.savedJobs {
		if job.UserID == userID && strings.EqualFold(NormalizeCanonicalURL(job.CanonicalURL), normURL) {
			return job, nil
		}
	}
	return nil, ErrSavedJobNotFound
}

func (r *MemoryCareerRepository) ListSavedJobs(ctx context.Context, userID string, status SavedJobStatus) ([]SavedJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []SavedJob
	for _, job := range r.savedJobs {
		if job.UserID == userID {
			if status != "" && job.Status != status {
				continue
			}
			list = append(list, *job)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) DeleteSavedJob(ctx context.Context, userID, savedJobID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, exists := r.savedJobs[savedJobID]
	if !exists {
		return ErrSavedJobNotFound
	}
	if job.UserID != userID {
		return ErrUnauthorized
	}
	delete(r.savedJobs, savedJobID)
	return nil
}

// --- Persistent Exclusions Ledger (IMP-CAR-10, CAR-10, C8) ---

func (r *MemoryCareerRepository) SaveExclusion(ctx context.Context, exclusion *JobExclusion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.exclusions[exclusion.ID] = exclusion
	return nil
}

func (r *MemoryCareerRepository) GetExclusion(ctx context.Context, userID, exclusionID string) (*JobExclusion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ex, exists := r.exclusions[exclusionID]
	if !exists {
		return nil, ErrExclusionNotFound
	}
	if ex.UserID != userID {
		return nil, ErrUnauthorized
	}
	return ex, nil
}

func (r *MemoryCareerRepository) ListExclusions(ctx context.Context, userID string) ([]JobExclusion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []JobExclusion
	for _, ex := range r.exclusions {
		if ex.UserID == userID {
			list = append(list, *ex)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) DeleteExclusion(ctx context.Context, userID, exclusionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ex, exists := r.exclusions[exclusionID]
	if !exists {
		return ErrExclusionNotFound
	}
	if ex.UserID != userID {
		return ErrUnauthorized
	}
	delete(r.exclusions, exclusionID)
	return nil
}

// --- Separate Application Ledger (IMP-CAR-10, CAR-10, AT-004, AT-006) ---

func (r *MemoryCareerRepository) SaveApplicationRecord(ctx context.Context, record *ApplicationRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.applications[record.ID] = record
	return nil
}

func (r *MemoryCareerRepository) GetApplicationByCanonicalURL(ctx context.Context, userID, canonicalURL string) (*ApplicationRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normURL := NormalizeCanonicalURL(canonicalURL)
	for _, app := range r.applications {
		if app.UserID == userID && strings.EqualFold(NormalizeCanonicalURL(app.CanonicalURL), normURL) {
			return app, nil
		}
	}
	return nil, errors.New("application record not found")
}

func (r *MemoryCareerRepository) ListApplications(ctx context.Context, userID string) ([]ApplicationRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []ApplicationRecord
	for _, app := range r.applications {
		if app.UserID == userID {
			list = append(list, *app)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) GetApplicationRecord(ctx context.Context, userID, recordID string) (*ApplicationRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	app, exists := r.applications[recordID]
	if !exists {
		return nil, ErrApplicationRecordNotFound
	}
	if app.UserID != userID {
		return nil, ErrUnauthorized
	}
	return app, nil
}

func (r *MemoryCareerRepository) UpdateApplicationRecord(ctx context.Context, record *ApplicationRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.applications[record.ID]
	if !exists {
		return ErrApplicationRecordNotFound
	}
	if existing.UserID != record.UserID {
		return ErrUnauthorized
	}

	record.UpdatedAt = time.Now().UTC()
	r.applications[record.ID] = record
	return nil
}

// --- Application Workflow & Human Review (IMP-CAR-11, CAR-11, REQ-005, REQ-015, AT-005) ---

func (r *MemoryCareerRepository) SaveReviewSession(ctx context.Context, session *ApplicationReviewSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reviewSessions[session.ID] = session
	return nil
}

func (r *MemoryCareerRepository) GetReviewSession(ctx context.Context, userID, sessionID string) (*ApplicationReviewSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.reviewSessions[sessionID]
	if !exists {
		return nil, ErrWorkflowNotFound
	}
	if session.UserID != userID {
		return nil, ErrUnauthorized
	}
	return session, nil
}

func (r *MemoryCareerRepository) ListReviewSessions(ctx context.Context, userID string, status ApplicationWorkflowStatus) ([]ApplicationReviewSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []ApplicationReviewSession
	for _, session := range r.reviewSessions {
		if session.UserID == userID {
			if status != "" && session.Status != status {
				continue
			}
			list = append(list, *session)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) ListDispatchedReviewSessions(ctx context.Context) ([]ApplicationReviewSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []ApplicationReviewSession
	for _, session := range r.reviewSessions {
		if session.Status == WorkflowStatusDispatched || session.Status == WorkflowStatusNeedsConfirmation {
			list = append(list, *session)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) UpdateReviewSession(ctx context.Context, session *ApplicationReviewSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.reviewSessions[session.ID]
	if !exists {
		return ErrWorkflowNotFound
	}
	if existing.UserID != session.UserID {
		return ErrUnauthorized
	}

	session.UpdatedAt = time.Now().UTC()
	r.reviewSessions[session.ID] = session
	return nil
}

func (r *MemoryCareerRepository) DeleteReviewSession(ctx context.Context, userID, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.reviewSessions[sessionID]
	if !exists {
		return ErrWorkflowNotFound
	}
	if session.UserID != userID {
		return ErrUnauthorized
	}

	delete(r.reviewSessions, sessionID)
	return nil
}

// Application Evidence Bundle Methods (IMP-CAR-16, CAR-16, AT-005, AT-011)

func (r *MemoryCareerRepository) SaveEvidenceBundle(ctx context.Context, bundle *ApplicationEvidenceBundle) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if bundle == nil || bundle.BundleID == "" {
		return ErrInvalidEvidenceBundle
	}

	r.evidenceBundles[bundle.BundleID] = bundle
	return nil
}

func (r *MemoryCareerRepository) GetEvidenceBundle(ctx context.Context, bundleID string) (*ApplicationEvidenceBundle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bundle, exists := r.evidenceBundles[bundleID]
	if !exists {
		return nil, ErrEvidenceBundleNotFound
	}

	copied := *bundle
	return &copied, nil
}

func (r *MemoryCareerRepository) GetEvidenceBundleByApplicationID(ctx context.Context, appID string) (*ApplicationEvidenceBundle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, bundle := range r.evidenceBundles {
		if bundle.ApplicationRecordID == appID {
			copied := *bundle
			return &copied, nil
		}
	}
	return nil, ErrEvidenceBundleNotFound
}

func (r *MemoryCareerRepository) ListEvidenceBundlesByUser(ctx context.Context, userID string) ([]ApplicationEvidenceBundle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []ApplicationEvidenceBundle
	for _, bundle := range r.evidenceBundles {
		if bundle.UserID == userID {
			list = append(list, *bundle)
		}
	}
	return list, nil
}

// --- Google Sheets Sync Methods (IMP-CAR-17, CAR-17, REQ-006, AT-013, AT-014) ---

func (r *MemoryCareerRepository) SaveSheetsSyncConfig(ctx context.Context, config *SheetsSyncConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *config
	r.sheetsConfigs[config.UserID] = &copied
	return nil
}

func (r *MemoryCareerRepository) GetSheetsSyncConfig(ctx context.Context, userID string) (*SheetsSyncConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, exists := r.sheetsConfigs[userID]
	if !exists {
		return nil, ErrSheetsConfigNotFound
	}

	copied := *config
	return &copied, nil
}

func (r *MemoryCareerRepository) SaveSheetsSyncResult(ctx context.Context, userID string, result *SheetsSyncResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *result
	r.sheetsResults[userID] = &copied
	return nil
}

func (r *MemoryCareerRepository) GetLatestSheetsSyncResult(ctx context.Context, userID string) (*SheetsSyncResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result, exists := r.sheetsResults[userID]
	if !exists {
		return nil, ErrSheetsResultNotFound
	}

	copied := *result
	return &copied, nil
}

// --- Multi-domain CSV Export & Audit Logging Methods (IMP-CAR-18, CAR-18, AT-013, AT-022) ---

func (r *MemoryCareerRepository) SaveExportAuditRecord(ctx context.Context, record *ExportAuditRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *record
	r.exportAudits = append(r.exportAudits, copied)
	return nil
}

func (r *MemoryCareerRepository) ListExportAuditRecords(ctx context.Context, userID string) ([]ExportAuditRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userAudits []ExportAuditRecord
	for _, rec := range r.exportAudits {
		if rec.UserID == userID {
			userAudits = append(userAudits, rec)
		}
	}
	return userAudits, nil
}

func (r *MemoryCareerRepository) SaveRecruiterContact(ctx context.Context, contact *RecruiterContact) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *contact
	r.contacts[contact.ID] = &copied
	return nil
}

func (r *MemoryCareerRepository) ListRecruiterContacts(ctx context.Context, userID string) ([]RecruiterContact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []RecruiterContact
	for _, c := range r.contacts {
		if c.UserID == userID {
			list = append(list, *c)
		}
	}
	return list, nil
}

// --- Daily Reports & Reminders Store Methods (IMP-CAR-19, CAR-19, AT-018) ---

func (r *MemoryCareerRepository) SaveDailyReport(ctx context.Context, report *DailyReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *report
	r.dailyReports[report.ID] = &copied
	return nil
}

func (r *MemoryCareerRepository) GetDailyReport(ctx context.Context, userID, date string) (*DailyReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	targetID := fmt.Sprintf("rpt_%s_%s", userID, date)
	if report, exists := r.dailyReports[targetID]; exists {
		copied := *report
		return &copied, nil
	}
	// Fallback to searching user & date
	for _, report := range r.dailyReports {
		if report.UserID == userID && report.ReportDate == date {
			copied := *report
			return &copied, nil
		}
	}
	return nil, ErrDailyReportNotFound
}

func (r *MemoryCareerRepository) ListDailyReports(ctx context.Context, userID string) ([]DailyReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []DailyReport
	for _, report := range r.dailyReports {
		if report.UserID == userID {
			list = append(list, *report)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) SaveDailyReportConfig(ctx context.Context, config *DailyReportConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *config
	r.dailyConfigs[config.UserID] = &copied
	return nil
}

func (r *MemoryCareerRepository) GetDailyReportConfig(ctx context.Context, userID string) (*DailyReportConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if cfg, exists := r.dailyConfigs[userID]; exists {
		copied := *cfg
		return &copied, nil
	}
	return nil, errors.New("daily report config not found")
}

func (r *MemoryCareerRepository) SaveReminder(ctx context.Context, reminder *CareerReminder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *reminder
	r.reminders[reminder.ID] = &copied
	return nil
}

func (r *MemoryCareerRepository) GetReminder(ctx context.Context, userID, reminderID string) (*CareerReminder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if rem, exists := r.reminders[reminderID]; exists {
		if rem.UserID == userID {
			copied := *rem
			return &copied, nil
		}
	}
	return nil, ErrReminderNotFound
}

func (r *MemoryCareerRepository) ListReminders(ctx context.Context, userID string, status ReminderStatus) ([]CareerReminder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []CareerReminder
	for _, rem := range r.reminders {
		if rem.UserID == userID {
			if status != "" && rem.Status != status {
				continue
			}
			list = append(list, *rem)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) UpdateReminderStatus(ctx context.Context, userID, reminderID string, status ReminderStatus, snoozedUntil *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rem, exists := r.reminders[reminderID]
	if !exists || rem.UserID != userID {
		return ErrReminderNotFound
	}

	rem.Status = status
	rem.SnoozedUntil = snoozedUntil
	rem.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *MemoryCareerRepository) SaveHiringPost(ctx context.Context, post *HiringPost) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *post
	r.hiringPosts[post.ID] = &copied
	return nil
}

func (r *MemoryCareerRepository) GetHiringPost(ctx context.Context, id string) (*HiringPost, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, exists := r.hiringPosts[id]
	if !exists {
		return nil, ErrHiringPostNotFound
	}
	copied := *post
	return &copied, nil
}

func (r *MemoryCareerRepository) ListHiringPosts(ctx context.Context, userID string) ([]HiringPost, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []HiringPost
	for _, post := range r.hiringPosts {
		if post.UserID == userID {
			list = append(list, *post)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) SaveRecruiterLead(ctx context.Context, lead *RecruiterLead) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *lead
	r.recruiterLeads[lead.ID] = &copied
	return nil
}

func (r *MemoryCareerRepository) GetRecruiterLead(ctx context.Context, id string) (*RecruiterLead, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lead, exists := r.recruiterLeads[id]
	if !exists {
		return nil, ErrLeadNotFound
	}
	copied := *lead
	return &copied, nil
}

func (r *MemoryCareerRepository) ListRecruiterLeads(ctx context.Context, userID string) ([]RecruiterLead, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []RecruiterLead
	for _, lead := range r.recruiterLeads {
		if lead.UserID == userID {
			list = append(list, *lead)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) SaveCareerRun(ctx context.Context, run *CareerRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *run
	r.careerRuns[run.ID] = &copied
	return nil
}

func (r *MemoryCareerRepository) GetCareerRun(ctx context.Context, id string) (*CareerRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	run, exists := r.careerRuns[id]
	if !exists {
		return nil, ErrRunNotFound
	}
	copied := *run
	return &copied, nil
}

func (r *MemoryCareerRepository) ListCareerRuns(ctx context.Context, userID string) ([]CareerRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []CareerRun
	for _, run := range r.careerRuns {
		if run.UserID == userID {
			list = append(list, *run)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) SaveProviderConsent(ctx context.Context, consent *UserProviderConsent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *consent
	key := fmt.Sprintf("%s:%s", consent.UserID, consent.ProviderID)
	r.providerConsents[key] = &copied
	return nil
}

func (r *MemoryCareerRepository) GetProviderConsent(ctx context.Context, userID, providerID string) (*UserProviderConsent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", userID, providerID)
	consent, exists := r.providerConsents[key]
	if !exists {
		return nil, ErrConsentNotFound
	}
	copied := *consent
	return &copied, nil
}

func (r *MemoryCareerRepository) ListProviderConsents(ctx context.Context, userID string) ([]UserProviderConsent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []UserProviderConsent
	for _, consent := range r.providerConsents {
		if consent.UserID == userID {
			list = append(list, *consent)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) RevokeProviderConsent(ctx context.Context, userID, providerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, providerID)
	consent, exists := r.providerConsents[key]
	if !exists {
		return ErrConsentNotFound
	}
	now := time.Now().UTC()
	consent.Status = "revoked"
	consent.RevokedAt = &now
	return nil
}

func (r *MemoryCareerRepository) SaveInterviewPrepPack(ctx context.Context, pack *InterviewPreparationPack) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.prepPacks[pack.PackID] = pack
	return nil
}

func (r *MemoryCareerRepository) GetInterviewPrepPack(ctx context.Context, packID string) (*InterviewPreparationPack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pack, exists := r.prepPacks[packID]
	if !exists {
		return nil, ErrPrepPackNotFound
	}
	return pack, nil
}

func (r *MemoryCareerRepository) GetInterviewPrepPackByAppID(ctx context.Context, appID string) (*InterviewPreparationPack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, pack := range r.prepPacks {
		if pack.ApplicationRecordID == appID {
			return pack, nil
		}
	}
	return nil, ErrPrepPackNotFound
}

func (r *MemoryCareerRepository) SaveInterviewSchedule(ctx context.Context, sched *InterviewSchedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.schedules[sched.ScheduleID] = sched
	return nil
}

func (r *MemoryCareerRepository) GetInterviewSchedule(ctx context.Context, scheduleID string) (*InterviewSchedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sched, exists := r.schedules[scheduleID]
	if !exists {
		return nil, ErrScheduleNotFound
	}
	return sched, nil
}

func (r *MemoryCareerRepository) ListInterviewSchedules(ctx context.Context, appID string) ([]InterviewSchedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []InterviewSchedule
	for _, sched := range r.schedules {
		if appID == "" || sched.ApplicationRecordID == appID {
			list = append(list, *sched)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) UpdateInterviewScheduleOutcome(ctx context.Context, scheduleID string, outcome RoundOutcome) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sched, exists := r.schedules[scheduleID]
	if !exists {
		return ErrScheduleNotFound
	}
	sched.Outcome = outcome
	return nil
}

func (r *MemoryCareerRepository) SaveInterviewSessionNote(ctx context.Context, note *InterviewSessionNote) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessionNotes[note.RoundID] = note
	return nil
}

func (r *MemoryCareerRepository) GetInterviewSessionNote(ctx context.Context, roundID string) (*InterviewSessionNote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	note, exists := r.sessionNotes[roundID]
	if !exists {
		return nil, ErrSessionNoteNotFound
	}
	return note, nil
}

func (r *MemoryCareerRepository) ListInterviewSessionNotes(ctx context.Context, appID string) ([]InterviewSessionNote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []InterviewSessionNote
	for _, note := range r.sessionNotes {
		if appID == "" || note.ApplicationRecordID == appID {
			list = append(list, *note)
		}
	}
	return list, nil
}

func (r *MemoryCareerRepository) SaveConsistencyReport(ctx context.Context, report *ConsistencyAuditReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.consistencyReports[report.ReportID] = report
	r.latestConsistency[report.UserID] = report.ReportID
	return nil
}

func (r *MemoryCareerRepository) GetConsistencyReport(ctx context.Context, reportID string) (*ConsistencyAuditReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	report, exists := r.consistencyReports[reportID]
	if !exists {
		return nil, ErrReportNotFound
	}
	return report, nil
}

func (r *MemoryCareerRepository) GetLatestConsistencyReport(ctx context.Context, userID string) (*ConsistencyAuditReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reportID, exists := r.latestConsistency[userID]
	if !exists {
		return nil, ErrReportNotFound
	}
	report, exists := r.consistencyReports[reportID]
	if !exists {
		return nil, ErrReportNotFound
	}
	return report, nil
}

func (r *MemoryCareerRepository) SaveLinkedInSnapshot(ctx context.Context, snapshot *LinkedInProfileSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.linkedInSnapshots[snapshot.UserID] = snapshot
	return nil
}

func (r *MemoryCareerRepository) GetLinkedInSnapshot(ctx context.Context, userID string) (*LinkedInProfileSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot, exists := r.linkedInSnapshots[userID]
	if !exists {
		return nil, ErrLinkedInSnapshotNotFound
	}
	return snapshot, nil
}






