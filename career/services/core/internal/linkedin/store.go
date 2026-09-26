package linkedin

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Repository defines storage operations for LinkedIn connections and inspection logs (AT-016).
type Repository interface {
	SaveConnection(ctx context.Context, conn *LinkedInConnection) error
	GetConnection(ctx context.Context, userID string) (*LinkedInConnection, error)
	DeleteConnection(ctx context.Context, userID string) error
	ListConnections(ctx context.Context) ([]LinkedInConnection, error)

	SaveProfileImport(ctx context.Context, profile *LinkedInImportedProfile) error
	GetLatestProfileImport(ctx context.Context, userID string) (*LinkedInImportedProfile, error)
	GetProfileImport(ctx context.Context, importID string) (*LinkedInImportedProfile, error)

	SaveOptimizationReport(ctx context.Context, report *ProfileOptimizationReport) error
	GetLatestOptimizationReport(ctx context.Context, userID string) (*ProfileOptimizationReport, error)
	GetOptimizationReport(ctx context.Context, reportID string) (*ProfileOptimizationReport, error)

	// Records Repository methods (IMP-LI-04, AT-011, AT-012)
	SavePersonRecord(ctx context.Context, p *PersonRecord) error
	GetPersonRecord(ctx context.Context, recordID, requestingTenantID string) (*PersonRecord, error)
	ListPersonRecords(ctx context.Context, filter RecordFilter) ([]PersonRecord, error)

	SaveCompanyRecord(ctx context.Context, c *CompanyRecord) error
	GetCompanyRecord(ctx context.Context, recordID, requestingTenantID string) (*CompanyRecord, error)
	ListCompanyRecords(ctx context.Context, filter RecordFilter) ([]CompanyRecord, error)

	SaveJobRecord(ctx context.Context, j *JobRecord) error
	GetJobRecord(ctx context.Context, recordID, requestingTenantID string) (*JobRecord, error)
	ListJobRecords(ctx context.Context, filter RecordFilter) ([]JobRecord, error)

	SavePostRecord(ctx context.Context, post *PostRecord) error
	GetPostRecord(ctx context.Context, recordID, requestingTenantID string) (*PostRecord, error)
	ListPostRecords(ctx context.Context, filter RecordFilter) ([]PostRecord, error)

	// Recruiter / Hiring Lead Workspace methods (IMP-LI-06, LI-06, AT-011)
	SaveRecruiterLead(ctx context.Context, lead *RecruiterLead) error
	GetRecruiterLead(ctx context.Context, leadID, requestingTenantID string) (*RecruiterLead, error)
	ListRecruiterLeads(ctx context.Context, filter RecruiterLeadFilter) ([]RecruiterLead, error)
	DeleteRecruiterLead(ctx context.Context, leadID, requestingTenantID string) error

	// Connection Note Drafts and Queue (IMP-LI-07, LI-07, AT-007, AT-010, FND-010, FND-011)
	SaveConnectionQueueItem(ctx context.Context, item *ConnectionQueueItem) error
	GetConnectionQueueItem(ctx context.Context, itemID, requestingTenantID string) (*ConnectionQueueItem, error)
	ListConnectionQueueItems(ctx context.Context, filter ConnectionQueueFilter) ([]ConnectionQueueItem, error)
	DeleteConnectionQueueItem(ctx context.Context, itemID, requestingTenantID string) error
	GetConnectionQueueItemByRecipientURL(ctx context.Context, workspaceID, linkedinURL string) (*ConnectionQueueItem, error)

	SaveConnectionBudget(ctx context.Context, budget *ConnectionBudget) error
	GetConnectionBudget(ctx context.Context, workspaceID, tenantID string) (*ConnectionBudget, error)

	// Company-Follow Planning (IMP-LI-08, LI-08, AT-010, FND-011, SRC-L1)
	SaveWatchlistItem(ctx context.Context, item *CompanyWatchlistItem) error
	GetWatchlistItem(ctx context.Context, itemID, requestingTenantID string) (*CompanyWatchlistItem, error)
	ListWatchlistItems(ctx context.Context, workspaceID, requestingTenantID string, status WatchlistStatus) ([]CompanyWatchlistItem, error)
	DeleteWatchlistItem(ctx context.Context, itemID, requestingTenantID string) error
	GetWatchlistItemByUniversalName(ctx context.Context, workspaceID, requestingTenantID, universalName string) (*CompanyWatchlistItem, error)

	SaveBatchFollowPlan(ctx context.Context, plan *BatchCompanyFollowPlan) error
	GetBatchFollowPlan(ctx context.Context, planID, requestingTenantID string) (*BatchCompanyFollowPlan, error)
	ListBatchFollowPlans(ctx context.Context, workspaceID, requestingTenantID string) ([]BatchCompanyFollowPlan, error)

	SaveCompanyFollowBudget(ctx context.Context, budget *CompanyFollowBudget) error
	GetCompanyFollowBudget(ctx context.Context, workspaceID, tenantID string) (*CompanyFollowBudget, error)

	// Inbox / Conversation Threads & Reply Drafts (IMP-LI-09, LI-09, AT-007, AT-010, AT-011, SRC-C3)
	SaveConversationThread(ctx context.Context, thread *ConversationThread) error
	GetConversationThread(ctx context.Context, threadID, requestingTenantID string) (*ConversationThread, error)
	ListConversationThreads(ctx context.Context, filter ConversationThreadFilter) ([]ConversationThread, error)
	DeleteConversationThread(ctx context.Context, threadID, requestingTenantID string) error

	SaveReplyDraft(ctx context.Context, draft *ReplyDraft) error
	GetReplyDraft(ctx context.Context, draftID, requestingTenantID string) (*ReplyDraft, error)
	ListReplyDraftsForThread(ctx context.Context, threadID, requestingTenantID string) ([]ReplyDraft, error)
	DeleteReplyDraft(ctx context.Context, draftID, requestingTenantID string) error

	// Comments, Replies and Thread Sweep (IMP-LI-10, LI-10, AT-007, AT-010, AT-011, SRC-L1, SRC-L2)
	SaveSweptTargetPost(ctx context.Context, post *SweptTargetPost) error
	GetSweptTargetPost(ctx context.Context, postID, requestingTenantID string) (*SweptTargetPost, error)
	ListSweptTargetPosts(ctx context.Context, filter SweptPostFilter) ([]SweptTargetPost, error)
	DeleteSweptTargetPost(ctx context.Context, postID, requestingTenantID string) error

	SaveCommentDraft(ctx context.Context, draft *CommentDraft) error
	GetCommentDraft(ctx context.Context, draftID, requestingTenantID string) (*CommentDraft, error)
	ListCommentDraftsForPost(ctx context.Context, postID, requestingTenantID string) ([]CommentDraft, error)
	DeleteCommentDraft(ctx context.Context, draftID, requestingTenantID string) error

	// Post Writing, Hooks and Audits (IMP-LI-11, LI-11, AT-003, AT-019, FND-015, SRC-L2)
	SavePostDraft(ctx context.Context, draft *PostDraft) error
	GetPostDraft(ctx context.Context, draftID, requestingTenantID string) (*PostDraft, error)
	ListPostDrafts(ctx context.Context, filter PostDraftFilter) ([]PostDraft, error)
	DeletePostDraft(ctx context.Context, draftID, requestingTenantID string) error

	// Humanizer and Reusable Voice (IMP-LI-12, LI-12, AT-003, FND-015, SRC-L2, SRC-S4)
	SaveVoiceProfile(ctx context.Context, profile *VoiceProfile) error
	GetVoiceProfile(ctx context.Context, profileID, requestingTenantID string) (*VoiceProfile, error)
	ListVoiceProfiles(ctx context.Context, workspaceID, requestingTenantID string) ([]VoiceProfile, error)
	DeleteVoiceProfile(ctx context.Context, profileID, requestingTenantID string) error

	SaveHumanizeResult(ctx context.Context, result *HumanizeResult) error
	GetHumanizeResult(ctx context.Context, resultID, requestingTenantID string) (*HumanizeResult, error)
	ListHumanizeResults(ctx context.Context, workspaceID, requestingTenantID string) ([]HumanizeResult, error)
	DeleteHumanizeResult(ctx context.Context, resultID, requestingTenantID string) error

	// Story Bank and Interviewer (IMP-LI-13, LI-13, AT-003, SRC-L2)
	SaveStoryEntry(ctx context.Context, entry *StoryEntry) error
	GetStoryEntry(ctx context.Context, entryID, requestingTenantID string) (*StoryEntry, error)
	ListStoryEntries(ctx context.Context, filter StoryFilter) ([]StoryEntry, error)
	DeleteStoryEntry(ctx context.Context, entryID, requestingTenantID string) error

	SaveInterviewSession(ctx context.Context, session *InterviewSession) error
	GetInterviewSession(ctx context.Context, sessionID, requestingTenantID string) (*InterviewSession, error)
	ListInterviewSessions(ctx context.Context, workspaceID, requestingTenantID string) ([]InterviewSession, error)
	DeleteInterviewSession(ctx context.Context, sessionID, requestingTenantID string) error

	// Content Planning and Repurposing (IMP-LI-14, LI-14, AT-018, AT-010, AT-007, FND-008, SRC-L2)
	SaveRepurposedDraft(ctx context.Context, draft *RepurposedDraft) error
	GetRepurposedDraft(ctx context.Context, draftID, requestingTenantID string) (*RepurposedDraft, error)
	ListRepurposedDrafts(ctx context.Context, workspaceID, requestingTenantID string) ([]RepurposedDraft, error)
	DeleteRepurposedDraft(ctx context.Context, draftID, requestingTenantID string) error

	SaveContentPlanItem(ctx context.Context, item *ContentPlanItem) error
	GetContentPlanItem(ctx context.Context, itemID, requestingTenantID string) (*ContentPlanItem, error)
	ListContentPlanItems(ctx context.Context, filter ContentCalendarFilter) ([]ContentPlanItem, error)
	DeleteContentPlanItem(ctx context.Context, itemID, requestingTenantID string) error

	// Engagement Monitoring and Analytics (IMP-LI-15, LI-15, AT-028, AT-010, AT-012, SRC-L2)
	SavePostAnalytics(ctx context.Context, snapshot *PostAnalyticsSnapshot) error
	GetPostAnalytics(ctx context.Context, snapshotID, requestingTenantID string) (*PostAnalyticsSnapshot, error)
	GetPostAnalyticsByURN(ctx context.Context, postURN, requestingTenantID string) (*PostAnalyticsSnapshot, error)
	ListPostAnalytics(ctx context.Context, workspaceID, requestingTenantID string) ([]PostAnalyticsSnapshot, error)
	DeletePostAnalytics(ctx context.Context, snapshotID, requestingTenantID string) error

	// Employee Advocacy (IMP-LI-16, LI-16, AT-007, AT-011, SRC-L2)
	SaveAdvocacyCampaign(ctx context.Context, campaign *AdvocacyCampaign) error
	GetAdvocacyCampaign(ctx context.Context, campaignID, requestingTenantID string) (*AdvocacyCampaign, error)
	ListAdvocacyCampaigns(ctx context.Context, workspaceID, requestingTenantID string) ([]AdvocacyCampaign, error)
	DeleteAdvocacyCampaign(ctx context.Context, campaignID, requestingTenantID string) error
	RecordEmployeeShare(ctx context.Context, event *EmployeeShareEvent) error
	ListEmployeeShares(ctx context.Context, campaignID, workspaceID, requestingTenantID string) ([]EmployeeShareEvent, error)

	// Provider Fallback & Diagnostics (IMP-LI-17, LI-17, AT-010, AT-016, SRC-L5)
	RegisterProviderAdapter(ctx context.Context, adapter *ProviderAdapter) error
	GetProviderAdapter(ctx context.Context, providerID, requestingTenantID string) (*ProviderAdapter, error)
	ListProviderAdapters(ctx context.Context, workspaceID, requestingTenantID string) ([]ProviderAdapter, error)
	UpdateProviderHealth(ctx context.Context, providerID, requestingTenantID string, status ProviderHealthStatus, latencyMs int64, errStr string) error
	DeleteProviderAdapter(ctx context.Context, providerID, requestingTenantID string) error
	RecordDiagnosticProbe(ctx context.Context, probe *DiagnosticProbeResult) error
	ListDiagnosticProbes(ctx context.Context, providerID string) ([]DiagnosticProbeResult, error)

	// Limits, CAPTCHA/reauth handling, activity (IMP-LI-18, LI-18, AT-008, AT-009, AT-010, REQ-009, REQ-010)
	SaveAccountSafetyState(ctx context.Context, state *AccountSafetyState) error
	GetAccountSafetyState(ctx context.Context, accountID, requestingTenantID string) (*AccountSafetyState, error)
	SaveRestrictionIncident(ctx context.Context, incident *PlatformRestrictionIncident) error
	ListRestrictionIncidents(ctx context.Context, accountID, requestingTenantID string) ([]PlatformRestrictionIncident, error)
	RecordActivityLog(ctx context.Context, entry *ActivityLogEntry) error
	ListActivityLog(ctx context.Context, accountID, requestingTenantID string, limit int) ([]ActivityLogEntry, error)

	// Session lifecycle and optional local tools (IMP-LI-19, LI-19, AT-016, FND-009, SRC-L3)
	SaveSession(ctx context.Context, session *LinkedInSessionRecord) error
	GetSession(ctx context.Context, sessionID, requestingTenantID string) (*LinkedInSessionRecord, error)
	ListSessions(ctx context.Context, ownerID, requestingTenantID string) ([]LinkedInSessionRecord, error)
	SaveViewerLease(ctx context.Context, lease *ShortLivedViewerLease) error
	GetViewerLease(ctx context.Context, leaseID string) (*ShortLivedViewerLease, error)

	// Relationship CSV/Sheets Exports (IMP-LI-20, LI-20, REQ-006, REQ-019, AT-013, AT-014)
	SaveRelationshipExportAudit(ctx context.Context, audit *RelationshipExportAuditRecord) error
	ListRelationshipExportAudits(ctx context.Context, workspaceID, requestingTenantID string) ([]RelationshipExportAuditRecord, error)
}

// StoryFilter defines search/filtering criteria for Story Bank entries.
type StoryFilter struct {
	WorkspaceID        string
	RequestingTenantID string
	Category           StoryCategory
	Status             StoryStatus
}

// MemoryRepository provides a thread-safe in-memory store for LinkedIn data.
type MemoryRepository struct {
	mu                  sync.RWMutex
	connections         map[string]*LinkedInConnection         // userID -> connection
	profileImports      map[string]*LinkedInImportedProfile    // importID -> profile
	userImports         map[string]string                      // userID -> latest importID
	optimizationReports map[string]*ProfileOptimizationReport  // reportID -> report
	userOptimizations   map[string]string                      // userID -> latest reportID

	// Records maps (IMP-LI-04)
	persons   map[string]*PersonRecord
	companies map[string]*CompanyRecord
	jobs      map[string]*JobRecord
	posts     map[string]*PostRecord

	// Recruiter Leads map (IMP-LI-06)
	leads map[string]*RecruiterLead

	// Connection Queue & Budgets (IMP-LI-07)
	queueItems map[string]*ConnectionQueueItem
	budgets    map[string]*ConnectionBudget

	// Company Follow Planning (IMP-LI-08)
	watchlistItems map[string]*CompanyWatchlistItem
	followPlans    map[string]*BatchCompanyFollowPlan
	followBudgets  map[string]*CompanyFollowBudget

	// Inbox / Conversations & Reply Drafts (IMP-LI-09)
	conversationThreads map[string]*ConversationThread
	replyDrafts         map[string]*ReplyDraft

	// Comments, Replies and Thread Sweep (IMP-LI-10)
	sweptPosts   map[string]*SweptTargetPost
	commentDrafts map[string]*CommentDraft

	// Post Writing, Hooks and Audits (IMP-LI-11)
	postDrafts map[string]*PostDraft

	// Humanizer and Reusable Voice (IMP-LI-12)
	voiceProfiles   map[string]*VoiceProfile
	humanizeResults map[string]*HumanizeResult

	// Story Bank and Interviewer (IMP-LI-13)
	storyEntries      map[string]*StoryEntry
	interviewSessions map[string]*InterviewSession

	// Content Planning and Repurposing (IMP-LI-14)
	repurposedDrafts map[string]*RepurposedDraft
	contentPlans     map[string]*ContentPlanItem

	// Engagement Monitoring and Analytics (IMP-LI-15)
	postAnalytics map[string]*PostAnalyticsSnapshot

	// Employee Advocacy (IMP-LI-16)
	advocacyCampaigns map[string]*AdvocacyCampaign
	employeeShares    map[string]*EmployeeShareEvent

	// Provider Fallback & Diagnostics (IMP-LI-17)
	providerAdapters map[string]*ProviderAdapter
	diagnosticProbes map[string]*DiagnosticProbeResult

	// Limits, CAPTCHA/reauth handling, activity (IMP-LI-18)
	accountSafetyStates  map[string]*AccountSafetyState
	restrictionIncidents map[string][]*PlatformRestrictionIncident
	activityLogs         map[string][]*ActivityLogEntry

	// Session lifecycle & Optional Local Tools (IMP-LI-19)
	sessions     map[string]*LinkedInSessionRecord
	viewerLeases map[string]*ShortLivedViewerLease

	// Relationship CSV/Sheets Exports (IMP-LI-20)
	relationshipExportAudits map[string][]*RelationshipExportAuditRecord
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		connections:         make(map[string]*LinkedInConnection),
		profileImports:      make(map[string]*LinkedInImportedProfile),
		userImports:         make(map[string]string),
		optimizationReports: make(map[string]*ProfileOptimizationReport),
		userOptimizations:   make(map[string]string),
		persons:             make(map[string]*PersonRecord),
		companies:           make(map[string]*CompanyRecord),
		jobs:                make(map[string]*JobRecord),
		posts:               make(map[string]*PostRecord),
		leads:               make(map[string]*RecruiterLead),
		queueItems:          make(map[string]*ConnectionQueueItem),
		budgets:             make(map[string]*ConnectionBudget),
		watchlistItems:      make(map[string]*CompanyWatchlistItem),
		followPlans:         make(map[string]*BatchCompanyFollowPlan),
		followBudgets:       make(map[string]*CompanyFollowBudget),
		conversationThreads: make(map[string]*ConversationThread),
		replyDrafts:         make(map[string]*ReplyDraft),
		sweptPosts:          make(map[string]*SweptTargetPost),
		commentDrafts:       make(map[string]*CommentDraft),
		postDrafts:          make(map[string]*PostDraft),
		voiceProfiles:       make(map[string]*VoiceProfile),
		humanizeResults:     make(map[string]*HumanizeResult),
		storyEntries:        make(map[string]*StoryEntry),
		interviewSessions:   make(map[string]*InterviewSession),
		repurposedDrafts:    make(map[string]*RepurposedDraft),
		contentPlans:        make(map[string]*ContentPlanItem),
		postAnalytics:       make(map[string]*PostAnalyticsSnapshot),
		advocacyCampaigns:   make(map[string]*AdvocacyCampaign),
		employeeShares:      make(map[string]*EmployeeShareEvent),
		providerAdapters:    make(map[string]*ProviderAdapter),
		diagnosticProbes:    make(map[string]*DiagnosticProbeResult),
		accountSafetyStates:  make(map[string]*AccountSafetyState),
		restrictionIncidents: make(map[string][]*PlatformRestrictionIncident),
		activityLogs:         make(map[string][]*ActivityLogEntry),
		sessions:             make(map[string]*LinkedInSessionRecord),
		viewerLeases:         make(map[string]*ShortLivedViewerLease),
		relationshipExportAudits: make(map[string][]*RelationshipExportAuditRecord),
	}
}

func (r *MemoryRepository) SaveProfileImport(ctx context.Context, profile *LinkedInImportedProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *profile
	r.profileImports[profile.ImportID] = &copied
	if profile.UserID != "" {
		r.userImports[profile.UserID] = profile.ImportID
	}
	return nil
}

func (r *MemoryRepository) GetLatestProfileImport(ctx context.Context, userID string) (*LinkedInImportedProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	impID, ok := r.userImports[userID]
	if !ok {
		return nil, ErrProfileDataNotFound
	}
	p, ok := r.profileImports[impID]
	if !ok {
		return nil, ErrProfileDataNotFound
	}
	copied := *p
	return &copied, nil
}

func (r *MemoryRepository) GetProfileImport(ctx context.Context, importID string) (*LinkedInImportedProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.profileImports[importID]
	if !ok {
		return nil, ErrProfileDataNotFound
	}
	copied := *p
	return &copied, nil
}

func (r *MemoryRepository) SaveOptimizationReport(ctx context.Context, report *ProfileOptimizationReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *report
	copied.Suggestions = make([]SectionSuggestion, len(report.Suggestions))
	copy(copied.Suggestions, report.Suggestions)

	r.optimizationReports[report.ReportID] = &copied
	if report.UserID != "" {
		r.userOptimizations[report.UserID] = report.ReportID
	}
	return nil
}

func (r *MemoryRepository) GetLatestOptimizationReport(ctx context.Context, userID string) (*ProfileOptimizationReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repID, ok := r.userOptimizations[userID]
	if !ok {
		return nil, ErrOptimizationReportNotFound
	}
	p, ok := r.optimizationReports[repID]
	if !ok {
		return nil, ErrOptimizationReportNotFound
	}

	copied := *p
	copied.Suggestions = make([]SectionSuggestion, len(p.Suggestions))
	copy(copied.Suggestions, p.Suggestions)
	return &copied, nil
}

func (r *MemoryRepository) GetOptimizationReport(ctx context.Context, reportID string) (*ProfileOptimizationReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.optimizationReports[reportID]
	if !ok {
		return nil, ErrOptimizationReportNotFound
	}

	copied := *p
	copied.Suggestions = make([]SectionSuggestion, len(p.Suggestions))
	copy(copied.Suggestions, p.Suggestions)
	return &copied, nil
}

func (r *MemoryRepository) SaveConnection(ctx context.Context, conn *LinkedInConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clone to avoid data races
	copied := *conn
	copied.Capabilities = make([]InspectedCapability, len(conn.Capabilities))
	copy(copied.Capabilities, conn.Capabilities)
	copied.GrantedScopes = make([]string, len(conn.GrantedScopes))
	copy(copied.GrantedScopes, conn.GrantedScopes)
	copied.DeniedScopes = make([]string, len(conn.DeniedScopes))
	copy(copied.DeniedScopes, conn.DeniedScopes)

	r.connections[conn.UserID] = &copied
	return nil
}

func (r *MemoryRepository) GetConnection(ctx context.Context, userID string) (*LinkedInConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conn, exists := r.connections[userID]
	if !exists {
		return nil, ErrConnectionNotFound
	}

	copied := *conn
	copied.Capabilities = make([]InspectedCapability, len(conn.Capabilities))
	copy(copied.Capabilities, conn.Capabilities)
	copied.GrantedScopes = make([]string, len(conn.GrantedScopes))
	copy(copied.GrantedScopes, conn.GrantedScopes)
	copied.DeniedScopes = make([]string, len(conn.DeniedScopes))
	copy(copied.DeniedScopes, conn.DeniedScopes)

	return &copied, nil
}

func (r *MemoryRepository) DeleteConnection(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.connections, userID)
	return nil
}

func (r *MemoryRepository) ListConnections(ctx context.Context) ([]LinkedInConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []LinkedInConnection
	for _, c := range r.connections {
		result = append(result, *c)
	}
	return result, nil
}

// -------------------------------------------------------------------------
// Person Records (IMP-LI-04, AT-011, AT-012)
// -------------------------------------------------------------------------

func (r *MemoryRepository) SavePersonRecord(ctx context.Context, p *PersonRecord) error {
	if p == nil {
		return ErrInvalidRecordData
	}
	if err := ValidateTenantScope(p.WorkspaceID, p.TenantID); err != nil {
		return err
	}
	if p.RecordID == "" {
		return fmt.Errorf("%w: record_id is required", ErrInvalidRecordData)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.persons[p.RecordID]
	copied := *p
	if exists {
		copied.Version = existing.Version + 1
	} else if copied.Version == 0 {
		copied.Version = 1
	}
	if copied.ObservedAt.IsZero() {
		copied.ObservedAt = time.Now().UTC()
	}
	copied.EntityType = "person"
	copied.Skills = make([]string, len(p.Skills))
	copy(copied.Skills, p.Skills)

	r.persons[p.RecordID] = &copied
	return nil
}

func (r *MemoryRepository) GetPersonRecord(ctx context.Context, recordID, requestingTenantID string) (*PersonRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.persons[recordID]
	if !exists {
		return nil, ErrRecordNotFound
	}
	if err := CheckTenantAccess(p.TenantID, requestingTenantID); err != nil {
		return nil, err
	}

	copied := *p
	copied.Skills = make([]string, len(p.Skills))
	copy(copied.Skills, p.Skills)
	return &copied, nil
}

func (r *MemoryRepository) ListPersonRecords(ctx context.Context, filter RecordFilter) ([]PersonRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []PersonRecord
	for _, p := range r.persons {
		if filter.TenantID != "" && p.TenantID != filter.TenantID {
			continue
		}
		if filter.WorkspaceID != "" && p.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.Query != "" {
			q := strings.ToLower(filter.Query)
			if !strings.Contains(strings.ToLower(p.FullName), q) &&
				!strings.Contains(strings.ToLower(p.Headline), q) &&
				!strings.Contains(strings.ToLower(p.CurrentCompany), q) {
				continue
			}
		}
		copied := *p
		copied.Skills = make([]string, len(p.Skills))
		copy(copied.Skills, p.Skills)
		result = append(result, copied)

		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result, nil
}

// -------------------------------------------------------------------------
// Company Records (IMP-LI-04, AT-011, AT-012)
// -------------------------------------------------------------------------

func (r *MemoryRepository) SaveCompanyRecord(ctx context.Context, c *CompanyRecord) error {
	if c == nil {
		return ErrInvalidRecordData
	}
	if err := ValidateTenantScope(c.WorkspaceID, c.TenantID); err != nil {
		return err
	}
	if c.RecordID == "" {
		return fmt.Errorf("%w: record_id is required", ErrInvalidRecordData)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.companies[c.RecordID]
	copied := *c
	if exists {
		copied.Version = existing.Version + 1
	} else if copied.Version == 0 {
		copied.Version = 1
	}
	if copied.ObservedAt.IsZero() {
		copied.ObservedAt = time.Now().UTC()
	}
	copied.EntityType = "company"
	copied.Specialties = make([]string, len(c.Specialties))
	copy(copied.Specialties, c.Specialties)

	r.companies[c.RecordID] = &copied
	return nil
}

func (r *MemoryRepository) GetCompanyRecord(ctx context.Context, recordID, requestingTenantID string) (*CompanyRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, exists := r.companies[recordID]
	if !exists {
		return nil, ErrRecordNotFound
	}
	if err := CheckTenantAccess(c.TenantID, requestingTenantID); err != nil {
		return nil, err
	}

	copied := *c
	copied.Specialties = make([]string, len(c.Specialties))
	copy(copied.Specialties, c.Specialties)
	return &copied, nil
}

func (r *MemoryRepository) ListCompanyRecords(ctx context.Context, filter RecordFilter) ([]CompanyRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []CompanyRecord
	for _, c := range r.companies {
		if filter.TenantID != "" && c.TenantID != filter.TenantID {
			continue
		}
		if filter.WorkspaceID != "" && c.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.Query != "" {
			q := strings.ToLower(filter.Query)
			if !strings.Contains(strings.ToLower(c.CompanyName), q) &&
				!strings.Contains(strings.ToLower(c.Industry), q) &&
				!strings.Contains(strings.ToLower(c.Domain), q) {
				continue
			}
		}
		copied := *c
		copied.Specialties = make([]string, len(c.Specialties))
		copy(copied.Specialties, c.Specialties)
		result = append(result, copied)

		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result, nil
}

// -------------------------------------------------------------------------
// Job Records (IMP-LI-04, AT-011, AT-012)
// -------------------------------------------------------------------------

func (r *MemoryRepository) SaveJobRecord(ctx context.Context, j *JobRecord) error {
	if j == nil {
		return ErrInvalidRecordData
	}
	if err := ValidateTenantScope(j.WorkspaceID, j.TenantID); err != nil {
		return err
	}
	if j.RecordID == "" {
		return fmt.Errorf("%w: record_id is required", ErrInvalidRecordData)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.jobs[j.RecordID]
	copied := *j
	if exists {
		copied.Version = existing.Version + 1
	} else if copied.Version == 0 {
		copied.Version = 1
	}
	if copied.ObservedAt.IsZero() {
		copied.ObservedAt = time.Now().UTC()
	}
	copied.EntityType = "job"

	r.jobs[j.RecordID] = &copied
	return nil
}

func (r *MemoryRepository) GetJobRecord(ctx context.Context, recordID, requestingTenantID string) (*JobRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	j, exists := r.jobs[recordID]
	if !exists {
		return nil, ErrRecordNotFound
	}
	if err := CheckTenantAccess(j.TenantID, requestingTenantID); err != nil {
		return nil, err
	}

	copied := *j
	return &copied, nil
}

func (r *MemoryRepository) ListJobRecords(ctx context.Context, filter RecordFilter) ([]JobRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []JobRecord
	for _, j := range r.jobs {
		if filter.TenantID != "" && j.TenantID != filter.TenantID {
			continue
		}
		if filter.WorkspaceID != "" && j.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.Query != "" {
			q := strings.ToLower(filter.Query)
			if !strings.Contains(strings.ToLower(j.Title), q) &&
				!strings.Contains(strings.ToLower(j.CompanyName), q) &&
				!strings.Contains(strings.ToLower(j.Location), q) {
				continue
			}
		}
		copied := *j
		result = append(result, copied)

		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result, nil
}

// -------------------------------------------------------------------------
// Post Records (IMP-LI-04, AT-011, AT-012)
// -------------------------------------------------------------------------

func (r *MemoryRepository) SavePostRecord(ctx context.Context, post *PostRecord) error {
	if post == nil {
		return ErrInvalidRecordData
	}
	if err := ValidateTenantScope(post.WorkspaceID, post.TenantID); err != nil {
		return err
	}
	if post.RecordID == "" {
		return fmt.Errorf("%w: record_id is required", ErrInvalidRecordData)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.posts[post.RecordID]
	copied := *post
	if exists {
		copied.Version = existing.Version + 1
	} else if copied.Version == 0 {
		copied.Version = 1
	}
	if copied.ObservedAt.IsZero() {
		copied.ObservedAt = time.Now().UTC()
	}
	copied.EntityType = "post"

	r.posts[post.RecordID] = &copied
	return nil
}

func (r *MemoryRepository) GetPostRecord(ctx context.Context, recordID, requestingTenantID string) (*PostRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, exists := r.posts[recordID]
	if !exists {
		return nil, ErrRecordNotFound
	}
	if err := CheckTenantAccess(post.TenantID, requestingTenantID); err != nil {
		return nil, err
	}

	copied := *post
	return &copied, nil
}

func (r *MemoryRepository) ListPostRecords(ctx context.Context, filter RecordFilter) ([]PostRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []PostRecord
	for _, p := range r.posts {
		if filter.TenantID != "" && p.TenantID != filter.TenantID {
			continue
		}
		if filter.WorkspaceID != "" && p.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.Query != "" {
			q := strings.ToLower(filter.Query)
			if !strings.Contains(strings.ToLower(p.AuthorName), q) &&
				!strings.Contains(strings.ToLower(p.Commentary), q) {
				continue
			}
		}
		copied := *p
		result = append(result, copied)

		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result, nil
}

func (r *MemoryRepository) SaveRecruiterLead(ctx context.Context, lead *RecruiterLead) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ValidateRecruiterLead(lead); err != nil {
		return err
	}

	copied := *lead
	// Deep copy notes
	copied.Notes = make([]LeadNote, len(lead.Notes))
	copy(copied.Notes, lead.Notes)

	// Copy reminder if present
	if lead.Reminder != nil {
		rem := *lead.Reminder
		copied.Reminder = &rem
	}

	now := time.Now().UTC()
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = now
	}
	copied.UpdatedAt = now

	r.leads[lead.ID] = &copied
	return nil
}

func (r *MemoryRepository) GetRecruiterLead(ctx context.Context, leadID, requestingTenantID string) (*RecruiterLead, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lead, exists := r.leads[leadID]
	if !exists {
		return nil, ErrLeadNotFound
	}
	if err := CheckTenantAccess(lead.TenantID, requestingTenantID); err != nil {
		return nil, err
	}

	copied := *lead
	copied.Notes = make([]LeadNote, len(lead.Notes))
	copy(copied.Notes, lead.Notes)
	if lead.Reminder != nil {
		rem := *lead.Reminder
		copied.Reminder = &rem
	}
	return &copied, nil
}

func (r *MemoryRepository) ListRecruiterLeads(ctx context.Context, filter RecruiterLeadFilter) ([]RecruiterLead, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []RecruiterLead
	for _, l := range r.leads {
		if filter.TenantID != "" && l.TenantID != filter.TenantID {
			continue
		}
		if filter.WorkspaceID != "" && l.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.Company != "" && !strings.EqualFold(l.Company, filter.Company) {
			continue
		}
		if filter.Status != "" && l.Status != filter.Status {
			continue
		}
		if filter.SearchQuery != "" {
			q := strings.ToLower(filter.SearchQuery)
			matchesName := strings.Contains(strings.ToLower(l.RecruiterName), q)
			matchesTitle := strings.Contains(strings.ToLower(l.RecruiterTitle), q)
			matchesCompany := strings.Contains(strings.ToLower(l.Company), q)
			if !matchesName && !matchesTitle && !matchesCompany {
				continue
			}
		}

		copied := *l
		copied.Notes = make([]LeadNote, len(l.Notes))
		copy(copied.Notes, l.Notes)
		if l.Reminder != nil {
			rem := *l.Reminder
			copied.Reminder = &rem
		}
		result = append(result, copied)
	}
	return result, nil
}

func (r *MemoryRepository) DeleteRecruiterLead(ctx context.Context, leadID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	lead, exists := r.leads[leadID]
	if !exists {
		return ErrLeadNotFound
	}
	if err := CheckTenantAccess(lead.TenantID, requestingTenantID); err != nil {
		return err
	}

	delete(r.leads, leadID)
	return nil
}

// Connection Queue and Budget Storage Implementation (IMP-LI-07, LI-07, AT-007, AT-010, FND-010, FND-011)

func (r *MemoryRepository) SaveConnectionQueueItem(ctx context.Context, item *ConnectionQueueItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ValidateConnectionQueueItem(item); err != nil {
		return err
	}

	copied := *item
	if item.ApprovedAt != nil {
		t := *item.ApprovedAt
		copied.ApprovedAt = &t
	}
	if item.CompletedAt != nil {
		t := *item.CompletedAt
		copied.CompletedAt = &t
	}
	if item.ContextFactors != nil {
		copied.ContextFactors = make([]string, len(item.ContextFactors))
		copy(copied.ContextFactors, item.ContextFactors)
	}

	r.queueItems[item.ID] = &copied
	return nil
}

func (r *MemoryRepository) GetConnectionQueueItem(ctx context.Context, itemID, requestingTenantID string) (*ConnectionQueueItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.queueItems[itemID]
	if !exists {
		return nil, ErrQueueItemNotFound
	}
	if err := CheckTenantAccess(item.TenantID, requestingTenantID); err != nil {
		return nil, err
	}

	copied := *item
	if item.ApprovedAt != nil {
		t := *item.ApprovedAt
		copied.ApprovedAt = &t
	}
	if item.CompletedAt != nil {
		t := *item.CompletedAt
		copied.CompletedAt = &t
	}
	if item.ContextFactors != nil {
		copied.ContextFactors = make([]string, len(item.ContextFactors))
		copy(copied.ContextFactors, item.ContextFactors)
	}
	return &copied, nil
}

func (r *MemoryRepository) ListConnectionQueueItems(ctx context.Context, filter ConnectionQueueFilter) ([]ConnectionQueueItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []ConnectionQueueItem
	for _, item := range r.queueItems {
		if filter.TenantID != "" {
			if err := CheckTenantAccess(item.TenantID, filter.TenantID); err != nil {
				continue
			}
		}
		if filter.WorkspaceID != "" && item.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		if filter.SearchQuery != "" {
			q := strings.ToLower(filter.SearchQuery)
			matchesName := strings.Contains(strings.ToLower(item.RecipientName), q)
			matchesCompany := strings.Contains(strings.ToLower(item.RecipientCompany), q)
			matchesTitle := strings.Contains(strings.ToLower(item.RecipientTitle), q)
			matchesNote := strings.Contains(strings.ToLower(item.NoteText), q)
			if !matchesName && !matchesCompany && !matchesTitle && !matchesNote {
				continue
			}
		}

		copied := *item
		if item.ApprovedAt != nil {
			t := *item.ApprovedAt
			copied.ApprovedAt = &t
		}
		if item.CompletedAt != nil {
			t := *item.CompletedAt
			copied.CompletedAt = &t
		}
		if item.ContextFactors != nil {
			copied.ContextFactors = make([]string, len(item.ContextFactors))
			copy(copied.ContextFactors, item.ContextFactors)
		}
		result = append(result, copied)
	}
	return result, nil
}

func (r *MemoryRepository) DeleteConnectionQueueItem(ctx context.Context, itemID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.queueItems[itemID]
	if !exists {
		return ErrQueueItemNotFound
	}
	if err := CheckTenantAccess(item.TenantID, requestingTenantID); err != nil {
		return err
	}

	delete(r.queueItems, itemID)
	return nil
}

func (r *MemoryRepository) GetConnectionQueueItemByRecipientURL(ctx context.Context, workspaceID, linkedinURL string) (*ConnectionQueueItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cleanTarget := strings.TrimRight(strings.ToLower(strings.TrimSpace(linkedinURL)), "/")
	for _, item := range r.queueItems {
		if item.WorkspaceID == workspaceID {
			cleanItem := strings.TrimRight(strings.ToLower(strings.TrimSpace(item.RecipientLinkedInURL)), "/")
			if cleanItem == cleanTarget {
				copied := *item
				return &copied, nil
			}
		}
	}
	return nil, ErrQueueItemNotFound
}

func (r *MemoryRepository) SaveConnectionBudget(ctx context.Context, budget *ConnectionBudget) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if budget == nil {
		return errors.New("budget cannot be nil")
	}
	key := fmt.Sprintf("%s:%s", budget.WorkspaceID, budget.TenantID)
	copied := *budget
	r.budgets[key] = &copied
	return nil
}

func (r *MemoryRepository) GetConnectionBudget(ctx context.Context, workspaceID, tenantID string) (*ConnectionBudget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", workspaceID, tenantID)
	budget, exists := r.budgets[key]
	if !exists {
		// Initialize default budget if not set yet
		b := NewConnectionBudget(workspaceID, tenantID)
		return b, nil
	}
	copied := *budget
	return &copied, nil
}

// Company-Follow Planning Implementations (IMP-LI-08, AT-010, FND-011, SRC-L1)

func (r *MemoryRepository) SaveWatchlistItem(ctx context.Context, item *CompanyWatchlistItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ValidateWatchlistItem(item); err != nil {
		return err
	}

	copied := *item
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = time.Now().UTC()
	}
	copied.UpdatedAt = time.Now().UTC()
	r.watchlistItems[copied.ItemID] = &copied
	return nil
}

func (r *MemoryRepository) GetWatchlistItem(ctx context.Context, itemID, requestingTenantID string) (*CompanyWatchlistItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.watchlistItems[itemID]
	if !exists {
		return nil, ErrWatchlistItemNotFound
	}
	if item.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}
	copied := *item
	return &copied, nil
}

func (r *MemoryRepository) ListWatchlistItems(ctx context.Context, workspaceID, requestingTenantID string, status WatchlistStatus) ([]CompanyWatchlistItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]CompanyWatchlistItem, 0)
	for _, item := range r.watchlistItems {
		if item.WorkspaceID == workspaceID {
			if item.TenantID != requestingTenantID {
				continue
			}
			if status != "" && item.Status != status {
				continue
			}
			res = append(res, *item)
		}
	}
	return res, nil
}

func (r *MemoryRepository) DeleteWatchlistItem(ctx context.Context, itemID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.watchlistItems[itemID]
	if !exists {
		return ErrWatchlistItemNotFound
	}
	if item.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.watchlistItems, itemID)
	return nil
}

func (r *MemoryRepository) GetWatchlistItemByUniversalName(ctx context.Context, workspaceID, requestingTenantID, universalName string) (*CompanyWatchlistItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cleanTarget := strings.TrimSpace(strings.ToLower(universalName))
	cleanTarget = strings.TrimPrefix(cleanTarget, "https://www.linkedin.com/company/")
	cleanTarget = strings.TrimPrefix(cleanTarget, "http://www.linkedin.com/company/")
	cleanTarget = strings.Trim(cleanTarget, "/")

	for _, item := range r.watchlistItems {
		if item.WorkspaceID == workspaceID && item.TenantID == requestingTenantID {
			itemUniversal := strings.TrimSpace(strings.ToLower(item.UniversalName))
			itemUniversal = strings.TrimPrefix(itemUniversal, "https://www.linkedin.com/company/")
			itemUniversal = strings.TrimPrefix(itemUniversal, "http://www.linkedin.com/company/")
			itemUniversal = strings.Trim(itemUniversal, "/")
			if itemUniversal == cleanTarget {
				copied := *item
				return &copied, nil
			}
		}
	}
	return nil, ErrWatchlistItemNotFound
}

func (r *MemoryRepository) SaveBatchFollowPlan(ctx context.Context, plan *BatchCompanyFollowPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if plan == nil || plan.PlanID == "" {
		return ErrInvalidBatchPlan
	}
	copied := *plan
	r.followPlans[plan.PlanID] = &copied
	return nil
}

func (r *MemoryRepository) GetBatchFollowPlan(ctx context.Context, planID, requestingTenantID string) (*BatchCompanyFollowPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plan, exists := r.followPlans[planID]
	if !exists {
		return nil, ErrBatchPlanNotFound
	}
	if plan.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}
	copied := *plan
	return &copied, nil
}

func (r *MemoryRepository) ListBatchFollowPlans(ctx context.Context, workspaceID, requestingTenantID string) ([]BatchCompanyFollowPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]BatchCompanyFollowPlan, 0)
	for _, plan := range r.followPlans {
		if plan.WorkspaceID == workspaceID {
			if plan.TenantID != requestingTenantID {
				continue
			}
			res = append(res, *plan)
		}
	}
	return res, nil
}

func (r *MemoryRepository) SaveCompanyFollowBudget(ctx context.Context, budget *CompanyFollowBudget) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if budget == nil {
		return errors.New("company_follow: budget cannot be nil")
	}
	key := fmt.Sprintf("%s:%s", budget.WorkspaceID, budget.TenantID)
	copied := *budget
	r.followBudgets[key] = &copied
	return nil
}

func (r *MemoryRepository) GetCompanyFollowBudget(ctx context.Context, workspaceID, tenantID string) (*CompanyFollowBudget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", workspaceID, tenantID)
	budget, exists := r.followBudgets[key]
	if !exists {
		b := DefaultCompanyFollowBudget(workspaceID, tenantID, time.Now().UTC())
		return b, nil
	}
	copied := *budget
	return &copied, nil
}

// Conversation Thread & Reply Draft Methods (IMP-LI-09)

func (r *MemoryRepository) SaveConversationThread(ctx context.Context, thread *ConversationThread) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if thread == nil || thread.ThreadID == "" {
		return errors.New("inbox: thread cannot be nil")
	}
	copied := *thread
	copied.Messages = make([]InboxMessage, len(thread.Messages))
	copy(copied.Messages, thread.Messages)
	r.conversationThreads[thread.ThreadID] = &copied
	return nil
}

func (r *MemoryRepository) GetConversationThread(ctx context.Context, threadID, requestingTenantID string) (*ConversationThread, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	thread, exists := r.conversationThreads[threadID]
	if !exists {
		return nil, ErrThreadNotFound
	}
	if thread.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}
	copied := *thread
	copied.Messages = make([]InboxMessage, len(thread.Messages))
	copy(copied.Messages, thread.Messages)
	return &copied, nil
}

func (r *MemoryRepository) ListConversationThreads(ctx context.Context, filter ConversationThreadFilter) ([]ConversationThread, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]ConversationThread, 0)
	for _, thread := range r.conversationThreads {
		if filter.WorkspaceID != "" && thread.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.TenantID != "" && thread.TenantID != filter.TenantID {
			continue
		}
		if filter.ThreadType != "" && thread.ThreadType != filter.ThreadType {
			continue
		}
		if filter.Classification != "" && thread.Classification != filter.Classification {
			continue
		}
		if filter.OnlyUnread && thread.UnreadCount == 0 {
			continue
		}
		if filter.Query != "" {
			q := strings.ToLower(filter.Query)
			matches := strings.Contains(strings.ToLower(thread.ParticipantName), q) ||
				strings.Contains(strings.ToLower(thread.ParticipantCompany), q) ||
				strings.Contains(strings.ToLower(thread.Subject), q) ||
				strings.Contains(strings.ToLower(thread.LastMessageSnippet), q)
			if !matches {
				continue
			}
		}
		copied := *thread
		copied.Messages = make([]InboxMessage, len(thread.Messages))
		copy(copied.Messages, thread.Messages)
		res = append(res, copied)
	}
	return res, nil
}

func (r *MemoryRepository) DeleteConversationThread(ctx context.Context, threadID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	thread, exists := r.conversationThreads[threadID]
	if !exists {
		return ErrThreadNotFound
	}
	if thread.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.conversationThreads, threadID)
	return nil
}

func (r *MemoryRepository) SaveReplyDraft(ctx context.Context, draft *ReplyDraft) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if draft == nil || draft.DraftID == "" {
		return errors.New("inbox: invalid reply draft")
	}
	copied := *draft
	copied.VerifiedFactsUsed = make([]string, len(draft.VerifiedFactsUsed))
	copy(copied.VerifiedFactsUsed, draft.VerifiedFactsUsed)
	r.replyDrafts[draft.DraftID] = &copied
	return nil
}

func (r *MemoryRepository) GetReplyDraft(ctx context.Context, draftID, requestingTenantID string) (*ReplyDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	draft, exists := r.replyDrafts[draftID]
	if !exists {
		return nil, ErrReplyDraftNotFound
	}
	if draft.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}
	copied := *draft
	copied.VerifiedFactsUsed = make([]string, len(draft.VerifiedFactsUsed))
	copy(copied.VerifiedFactsUsed, draft.VerifiedFactsUsed)
	return &copied, nil
}

func (r *MemoryRepository) ListReplyDraftsForThread(ctx context.Context, threadID, requestingTenantID string) ([]ReplyDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]ReplyDraft, 0)
	for _, draft := range r.replyDrafts {
		if draft.ThreadID == threadID {
			if draft.TenantID != requestingTenantID {
				continue
			}
			copied := *draft
			copied.VerifiedFactsUsed = make([]string, len(draft.VerifiedFactsUsed))
			copy(copied.VerifiedFactsUsed, draft.VerifiedFactsUsed)
			res = append(res, copied)
		}
	}
	return res, nil
}

func (r *MemoryRepository) DeleteReplyDraft(ctx context.Context, draftID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	draft, exists := r.replyDrafts[draftID]
	if !exists {
		return ErrReplyDraftNotFound
	}
	if draft.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.replyDrafts, draftID)
	return nil
}

// SweptTargetPost and CommentDraft store methods (IMP-LI-10, LI-10, AT-007, AT-010, AT-011)

func (r *MemoryRepository) SaveSweptTargetPost(ctx context.Context, post *SweptTargetPost) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *post
	if len(post.Comments) > 0 {
		copied.Comments = make([]SweptPostComment, len(post.Comments))
		copy(copied.Comments, post.Comments)
	}
	r.sweptPosts[post.PostID] = &copied
	return nil
}

func (r *MemoryRepository) GetSweptTargetPost(ctx context.Context, postID, requestingTenantID string) (*SweptTargetPost, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, exists := r.sweptPosts[postID]
	if !exists {
		return nil, ErrSweepPostNotFound
	}
	if post.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}
	copied := *post
	if len(post.Comments) > 0 {
		copied.Comments = make([]SweptPostComment, len(post.Comments))
		copy(copied.Comments, post.Comments)
	}
	return &copied, nil
}

func (r *MemoryRepository) ListSweptTargetPosts(ctx context.Context, filter SweptPostFilter) ([]SweptTargetPost, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	queryLower := strings.ToLower(strings.TrimSpace(filter.Query))
	res := make([]SweptTargetPost, 0)
	for _, post := range r.sweptPosts {
		if filter.WorkspaceID != "" && post.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.TenantID != "" && post.TenantID != filter.TenantID {
			continue
		}
		if filter.Category != "" && post.Category != filter.Category {
			continue
		}
		if queryLower != "" {
			authorMatches := strings.Contains(strings.ToLower(post.AuthorName), queryLower)
			contentMatches := strings.Contains(strings.ToLower(post.Content), queryLower)
			companyMatches := strings.Contains(strings.ToLower(post.AuthorCompany), queryLower)
			if !authorMatches && !contentMatches && !companyMatches {
				continue
			}
		}

		copied := *post
		if len(post.Comments) > 0 {
			copied.Comments = make([]SweptPostComment, len(post.Comments))
			copy(copied.Comments, post.Comments)
		}
		res = append(res, copied)
	}
	return res, nil
}

func (r *MemoryRepository) DeleteSweptTargetPost(ctx context.Context, postID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, exists := r.sweptPosts[postID]
	if !exists {
		return ErrSweepPostNotFound
	}
	if post.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.sweptPosts, postID)
	return nil
}

func (r *MemoryRepository) SaveCommentDraft(ctx context.Context, draft *CommentDraft) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *draft
	copied.VerifiedFactsUsed = make([]string, len(draft.VerifiedFactsUsed))
	copy(copied.VerifiedFactsUsed, draft.VerifiedFactsUsed)
	r.commentDrafts[draft.DraftID] = &copied
	return nil
}

func (r *MemoryRepository) GetCommentDraft(ctx context.Context, draftID, requestingTenantID string) (*CommentDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	draft, exists := r.commentDrafts[draftID]
	if !exists {
		return nil, ErrCommentDraftNotFound
	}
	if draft.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}
	copied := *draft
	copied.VerifiedFactsUsed = make([]string, len(draft.VerifiedFactsUsed))
	copy(copied.VerifiedFactsUsed, draft.VerifiedFactsUsed)
	return &copied, nil
}

func (r *MemoryRepository) ListCommentDraftsForPost(ctx context.Context, postID, requestingTenantID string) ([]CommentDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]CommentDraft, 0)
	for _, draft := range r.commentDrafts {
		if draft.PostID == postID {
			if draft.TenantID != requestingTenantID {
				continue
			}
			copied := *draft
			copied.VerifiedFactsUsed = make([]string, len(draft.VerifiedFactsUsed))
			copy(copied.VerifiedFactsUsed, draft.VerifiedFactsUsed)
			res = append(res, copied)
		}
	}
	return res, nil
}

func (r *MemoryRepository) DeleteCommentDraft(ctx context.Context, draftID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	draft, exists := r.commentDrafts[draftID]
	if !exists {
		return ErrCommentDraftNotFound
	}
	if draft.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.commentDrafts, draftID)
	return nil
}

// Post Writing, Hooks and Audits (IMP-LI-11, LI-11, AT-003, AT-019, FND-015, SRC-L2)

func (r *MemoryRepository) SavePostDraft(ctx context.Context, draft *PostDraft) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *draft
	copied.VerifiedFactsUsed = make([]string, len(draft.VerifiedFactsUsed))
	copy(copied.VerifiedFactsUsed, draft.VerifiedFactsUsed)
	copied.AvailableHooks = make([]HookVariant, len(draft.AvailableHooks))
	copy(copied.AvailableHooks, draft.AvailableHooks)
	if draft.LatestAudit != nil {
		auditCopy := *draft.LatestAudit
		auditCopy.Issues = make([]AuditIssue, len(draft.LatestAudit.Issues))
		copy(auditCopy.Issues, draft.LatestAudit.Issues)
		copied.LatestAudit = &auditCopy
	}
	r.postDrafts[draft.DraftID] = &copied
	return nil
}

func (r *MemoryRepository) GetPostDraft(ctx context.Context, draftID, requestingTenantID string) (*PostDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	draft, exists := r.postDrafts[draftID]
	if !exists {
		return nil, ErrPostDraftNotFound
	}
	if draft.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}
	copied := *draft
	copied.VerifiedFactsUsed = make([]string, len(draft.VerifiedFactsUsed))
	copy(copied.VerifiedFactsUsed, draft.VerifiedFactsUsed)
	copied.AvailableHooks = make([]HookVariant, len(draft.AvailableHooks))
	copy(copied.AvailableHooks, draft.AvailableHooks)
	if draft.LatestAudit != nil {
		auditCopy := *draft.LatestAudit
		auditCopy.Issues = make([]AuditIssue, len(draft.LatestAudit.Issues))
		copy(auditCopy.Issues, draft.LatestAudit.Issues)
		copied.LatestAudit = &auditCopy
	}
	return &copied, nil
}

func (r *MemoryRepository) ListPostDrafts(ctx context.Context, filter PostDraftFilter) ([]PostDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]PostDraft, 0)
	q := strings.ToLower(strings.TrimSpace(filter.Query))
	for _, draft := range r.postDrafts {
		if draft.WorkspaceID == filter.WorkspaceID {
			if draft.TenantID != filter.TenantID {
				continue
			}
			if filter.Angle != "" && draft.Angle != filter.Angle {
				continue
			}
			if filter.Status != "" && draft.Status != filter.Status {
				continue
			}
			if q != "" {
				topicMatch := strings.Contains(strings.ToLower(draft.Topic), q)
				textMatch := strings.Contains(strings.ToLower(draft.FullPostText), q)
				if !topicMatch && !textMatch {
					continue
				}
			}
			copied := *draft
			copied.VerifiedFactsUsed = make([]string, len(draft.VerifiedFactsUsed))
			copy(copied.VerifiedFactsUsed, draft.VerifiedFactsUsed)
			copied.AvailableHooks = make([]HookVariant, len(draft.AvailableHooks))
			copy(copied.AvailableHooks, draft.AvailableHooks)
			if draft.LatestAudit != nil {
				auditCopy := *draft.LatestAudit
				auditCopy.Issues = make([]AuditIssue, len(draft.LatestAudit.Issues))
				copy(auditCopy.Issues, draft.LatestAudit.Issues)
				copied.LatestAudit = &auditCopy
			}
			res = append(res, copied)
		}
	}
	return res, nil
}

func (r *MemoryRepository) DeletePostDraft(ctx context.Context, draftID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	draft, exists := r.postDrafts[draftID]
	if !exists {
		return ErrPostDraftNotFound
	}
	if draft.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.postDrafts, draftID)
	return nil
}

// -------------------------------------------------------------------------
// Humanizer and Reusable Voice (IMP-LI-12, LI-12, AT-003, FND-015, SRC-L2, SRC-S4)
// -------------------------------------------------------------------------

func (r *MemoryRepository) SaveVoiceProfile(ctx context.Context, profile *VoiceProfile) error {
	if profile == nil {
		return ErrInvalidRecordData
	}
	if err := ValidateTenantScope(profile.WorkspaceID, profile.TenantID); err != nil {
		return err
	}
	if profile.ProfileID == "" {
		return fmt.Errorf("%w: profile_id is required", ErrInvalidRecordData)
	}
	if err := ValidateVoiceProfile(profile); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *profile
	copied.PreferredTerms = make([]string, len(profile.PreferredTerms))
	copy(copied.PreferredTerms, profile.PreferredTerms)
	copied.BlacklistedTerms = make([]string, len(profile.BlacklistedTerms))
	copy(copied.BlacklistedTerms, profile.BlacklistedTerms)
	copied.ApprovedWritingSamples = make([]string, len(profile.ApprovedWritingSamples))
	copy(copied.ApprovedWritingSamples, profile.ApprovedWritingSamples)

	nowStr := time.Now().UTC().Format(time.RFC3339)
	if copied.CreatedAt == "" {
		copied.CreatedAt = nowStr
	}
	copied.UpdatedAt = nowStr

	r.voiceProfiles[profile.ProfileID] = &copied
	return nil
}

func (r *MemoryRepository) GetVoiceProfile(ctx context.Context, profileID, requestingTenantID string) (*VoiceProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	profile, exists := r.voiceProfiles[profileID]
	if !exists {
		return nil, ErrVoiceProfileNotFound
	}
	if profile.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}

	copied := *profile
	copied.PreferredTerms = make([]string, len(profile.PreferredTerms))
	copy(copied.PreferredTerms, profile.PreferredTerms)
	copied.BlacklistedTerms = make([]string, len(profile.BlacklistedTerms))
	copy(copied.BlacklistedTerms, profile.BlacklistedTerms)
	copied.ApprovedWritingSamples = make([]string, len(profile.ApprovedWritingSamples))
	copy(copied.ApprovedWritingSamples, profile.ApprovedWritingSamples)
	return &copied, nil
}

func (r *MemoryRepository) ListVoiceProfiles(ctx context.Context, workspaceID, requestingTenantID string) ([]VoiceProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []VoiceProfile
	for _, p := range r.voiceProfiles {
		if p.WorkspaceID == workspaceID {
			if p.TenantID != requestingTenantID {
				continue
			}
			copied := *p
			copied.PreferredTerms = make([]string, len(p.PreferredTerms))
			copy(copied.PreferredTerms, p.PreferredTerms)
			copied.BlacklistedTerms = make([]string, len(p.BlacklistedTerms))
			copy(copied.BlacklistedTerms, p.BlacklistedTerms)
			copied.ApprovedWritingSamples = make([]string, len(p.ApprovedWritingSamples))
			copy(copied.ApprovedWritingSamples, p.ApprovedWritingSamples)
			result = append(result, copied)
		}
	}
	return result, nil
}

func (r *MemoryRepository) DeleteVoiceProfile(ctx context.Context, profileID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	profile, exists := r.voiceProfiles[profileID]
	if !exists {
		return ErrVoiceProfileNotFound
	}
	if profile.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.voiceProfiles, profileID)
	return nil
}

func (r *MemoryRepository) SaveHumanizeResult(ctx context.Context, result *HumanizeResult) error {
	if result == nil {
		return ErrInvalidRecordData
	}
	if err := ValidateTenantScope(result.WorkspaceID, result.TenantID); err != nil {
		return err
	}
	if result.ResultID == "" {
		return fmt.Errorf("%w: result_id is required", ErrInvalidRecordData)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *result
	copied.Tier1SlopReplacements = make([]SlopReplacement, len(result.Tier1SlopReplacements))
	copy(copied.Tier1SlopReplacements, result.Tier1SlopReplacements)
	copied.Tier2CadenceNotes = make([]string, len(result.Tier2CadenceNotes))
	copy(copied.Tier2CadenceNotes, result.Tier2CadenceNotes)
	copied.Tier3Issues = make([]HumanizerAuditIssue, len(result.Tier3Issues))
	copy(copied.Tier3Issues, result.Tier3Issues)

	r.humanizeResults[result.ResultID] = &copied
	return nil
}

func (r *MemoryRepository) GetHumanizeResult(ctx context.Context, resultID, requestingTenantID string) (*HumanizeResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res, exists := r.humanizeResults[resultID]
	if !exists {
		return nil, ErrHumanizeResultNotFound
	}
	if res.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}

	copied := *res
	copied.Tier1SlopReplacements = make([]SlopReplacement, len(res.Tier1SlopReplacements))
	copy(copied.Tier1SlopReplacements, res.Tier1SlopReplacements)
	copied.Tier2CadenceNotes = make([]string, len(res.Tier2CadenceNotes))
	copy(copied.Tier2CadenceNotes, res.Tier2CadenceNotes)
	copied.Tier3Issues = make([]HumanizerAuditIssue, len(res.Tier3Issues))
	copy(copied.Tier3Issues, res.Tier3Issues)
	return &copied, nil
}

func (r *MemoryRepository) ListHumanizeResults(ctx context.Context, workspaceID, requestingTenantID string) ([]HumanizeResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []HumanizeResult
	for _, res := range r.humanizeResults {
		if res.WorkspaceID == workspaceID {
			if res.TenantID != requestingTenantID {
				continue
			}
			copied := *res
			copied.Tier1SlopReplacements = make([]SlopReplacement, len(res.Tier1SlopReplacements))
			copy(copied.Tier1SlopReplacements, res.Tier1SlopReplacements)
			copied.Tier2CadenceNotes = make([]string, len(res.Tier2CadenceNotes))
			copy(copied.Tier2CadenceNotes, res.Tier2CadenceNotes)
			copied.Tier3Issues = make([]HumanizerAuditIssue, len(res.Tier3Issues))
			copy(copied.Tier3Issues, res.Tier3Issues)
			result = append(result, copied)
		}
	}
	return result, nil
}

func (r *MemoryRepository) DeleteHumanizeResult(ctx context.Context, resultID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, exists := r.humanizeResults[resultID]
	if !exists {
		return ErrHumanizeResultNotFound
	}
	if res.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.humanizeResults, resultID)
	return nil
}

// -------------------------------------------------------------------------
// Story Bank and Interviewer (IMP-LI-13, LI-13, AT-003, SRC-L2)
// -------------------------------------------------------------------------

func (r *MemoryRepository) SaveStoryEntry(ctx context.Context, entry *StoryEntry) error {
	if entry == nil {
		return ErrInvalidStoryData
	}
	if err := ValidateTenantScope(entry.WorkspaceID, entry.TenantID); err != nil {
		return err
	}
	if entry.ID == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidStoryData)
	}
	if err := ValidateStoryEntry(entry); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *entry
	copied.Provenance.GroundedFactIDs = make([]string, len(entry.Provenance.GroundedFactIDs))
	copy(copied.Provenance.GroundedFactIDs, entry.Provenance.GroundedFactIDs)
	copied.Tags = make([]string, len(entry.Tags))
	copy(copied.Tags, entry.Tags)

	if entry.AuditReport != nil {
		auditCopy := *entry.AuditReport
		auditCopy.Issues = make([]StoryAuditIssue, len(entry.AuditReport.Issues))
		copy(auditCopy.Issues, entry.AuditReport.Issues)
		copied.AuditReport = &auditCopy
	}

	now := time.Now().UTC()
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = now
	}
	copied.UpdatedAt = now

	r.storyEntries[entry.ID] = &copied
	return nil
}

func (r *MemoryRepository) GetStoryEntry(ctx context.Context, entryID, requestingTenantID string) (*StoryEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.storyEntries[entryID]
	if !exists {
		return nil, ErrStoryEntryNotFound
	}
	if entry.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}

	copied := *entry
	copied.Provenance.GroundedFactIDs = make([]string, len(entry.Provenance.GroundedFactIDs))
	copy(copied.Provenance.GroundedFactIDs, entry.Provenance.GroundedFactIDs)
	copied.Tags = make([]string, len(entry.Tags))
	copy(copied.Tags, entry.Tags)

	if entry.AuditReport != nil {
		auditCopy := *entry.AuditReport
		auditCopy.Issues = make([]StoryAuditIssue, len(entry.AuditReport.Issues))
		copy(auditCopy.Issues, entry.AuditReport.Issues)
		copied.AuditReport = &auditCopy
	}
	return &copied, nil
}

func (r *MemoryRepository) ListStoryEntries(ctx context.Context, filter StoryFilter) ([]StoryEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []StoryEntry
	for _, entry := range r.storyEntries {
		if entry.WorkspaceID == filter.WorkspaceID {
			if entry.TenantID != filter.RequestingTenantID {
				continue
			}
			if filter.Category != "" && entry.Category != filter.Category {
				continue
			}
			if filter.Status != "" && entry.ApprovalStatus != filter.Status {
				continue
			}

			copied := *entry
			copied.Provenance.GroundedFactIDs = make([]string, len(entry.Provenance.GroundedFactIDs))
			copy(copied.Provenance.GroundedFactIDs, entry.Provenance.GroundedFactIDs)
			copied.Tags = make([]string, len(entry.Tags))
			copy(copied.Tags, entry.Tags)

			if entry.AuditReport != nil {
				auditCopy := *entry.AuditReport
				auditCopy.Issues = make([]StoryAuditIssue, len(entry.AuditReport.Issues))
				copy(auditCopy.Issues, entry.AuditReport.Issues)
				copied.AuditReport = &auditCopy
			}
			result = append(result, copied)
		}
	}
	return result, nil
}

func (r *MemoryRepository) DeleteStoryEntry(ctx context.Context, entryID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.storyEntries[entryID]
	if !exists {
		return ErrStoryEntryNotFound
	}
	if entry.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.storyEntries, entryID)
	return nil
}

func (r *MemoryRepository) SaveInterviewSession(ctx context.Context, session *InterviewSession) error {
	if session == nil {
		return ErrInterviewSessionNotFound
	}
	if err := ValidateTenantScope(session.WorkspaceID, session.TenantID); err != nil {
		return err
	}
	if session.ID == "" {
		return fmt.Errorf("%w: session id is required", ErrInvalidStoryData)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *session
	copied.QuestionsAndAnswers = make([]InterviewQA, len(session.QuestionsAndAnswers))
	copy(copied.QuestionsAndAnswers, session.QuestionsAndAnswers)

	now := time.Now().UTC()
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = now
	}
	copied.UpdatedAt = now

	r.interviewSessions[session.ID] = &copied
	return nil
}

func (r *MemoryRepository) GetInterviewSession(ctx context.Context, sessionID, requestingTenantID string) (*InterviewSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.interviewSessions[sessionID]
	if !exists {
		return nil, ErrInterviewSessionNotFound
	}
	if session.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}

	copied := *session
	copied.QuestionsAndAnswers = make([]InterviewQA, len(session.QuestionsAndAnswers))
	copy(copied.QuestionsAndAnswers, session.QuestionsAndAnswers)
	return &copied, nil
}

func (r *MemoryRepository) ListInterviewSessions(ctx context.Context, workspaceID, requestingTenantID string) ([]InterviewSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []InterviewSession
	for _, session := range r.interviewSessions {
		if session.WorkspaceID == workspaceID {
			if session.TenantID != requestingTenantID {
				continue
			}
			copied := *session
			copied.QuestionsAndAnswers = make([]InterviewQA, len(session.QuestionsAndAnswers))
			copy(copied.QuestionsAndAnswers, session.QuestionsAndAnswers)
			result = append(result, copied)
		}
	}
	return result, nil
}

func (r *MemoryRepository) DeleteInterviewSession(ctx context.Context, sessionID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.interviewSessions[sessionID]
	if !exists {
		return ErrInterviewSessionNotFound
	}
	if session.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.interviewSessions, sessionID)
	return nil
}

// -------------------------------------------------------------------------
// Content Planning and Repurposing (IMP-LI-14, AT-018, AT-010, AT-007, FND-008)
// -------------------------------------------------------------------------

func (r *MemoryRepository) SaveRepurposedDraft(ctx context.Context, draft *RepurposedDraft) error {
	if draft == nil {
		return ErrRepurposedDraftNotFound
	}
	if err := ValidateTenantScope(draft.WorkspaceID, draft.TenantID); err != nil {
		return err
	}
	if draft.ID == "" {
		return fmt.Errorf("%w: draft id is required", ErrInvalidPlanData)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *draft
	if len(draft.Tags) > 0 {
		copied.Tags = make([]string, len(draft.Tags))
		copy(copied.Tags, draft.Tags)
	}
	if draft.LastApprovedAt != nil {
		t := *draft.LastApprovedAt
		copied.LastApprovedAt = &t
	}
	if draft.ScheduledSlotUTC != nil {
		t := *draft.ScheduledSlotUTC
		copied.ScheduledSlotUTC = &t
	}

	now := time.Now().UTC()
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = now
	}
	copied.UpdatedAt = now

	r.repurposedDrafts[draft.ID] = &copied
	return nil
}

func (r *MemoryRepository) GetRepurposedDraft(ctx context.Context, draftID, requestingTenantID string) (*RepurposedDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	draft, exists := r.repurposedDrafts[draftID]
	if !exists {
		return nil, ErrRepurposedDraftNotFound
	}
	if draft.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}

	copied := *draft
	if len(draft.Tags) > 0 {
		copied.Tags = make([]string, len(draft.Tags))
		copy(copied.Tags, draft.Tags)
	}
	if draft.LastApprovedAt != nil {
		t := *draft.LastApprovedAt
		copied.LastApprovedAt = &t
	}
	if draft.ScheduledSlotUTC != nil {
		t := *draft.ScheduledSlotUTC
		copied.ScheduledSlotUTC = &t
	}
	return &copied, nil
}

func (r *MemoryRepository) ListRepurposedDrafts(ctx context.Context, workspaceID, requestingTenantID string) ([]RepurposedDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []RepurposedDraft
	for _, draft := range r.repurposedDrafts {
		if draft.WorkspaceID == workspaceID {
			if draft.TenantID != requestingTenantID {
				continue
			}
			copied := *draft
			if len(draft.Tags) > 0 {
				copied.Tags = make([]string, len(draft.Tags))
				copy(copied.Tags, draft.Tags)
			}
			if draft.LastApprovedAt != nil {
				t := *draft.LastApprovedAt
				copied.LastApprovedAt = &t
			}
			if draft.ScheduledSlotUTC != nil {
				t := *draft.ScheduledSlotUTC
				copied.ScheduledSlotUTC = &t
			}
			result = append(result, copied)
		}
	}
	return result, nil
}

func (r *MemoryRepository) DeleteRepurposedDraft(ctx context.Context, draftID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	draft, exists := r.repurposedDrafts[draftID]
	if !exists {
		return ErrRepurposedDraftNotFound
	}
	if draft.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.repurposedDrafts, draftID)
	return nil
}

func (r *MemoryRepository) SaveContentPlanItem(ctx context.Context, item *ContentPlanItem) error {
	if item == nil {
		return ErrContentPlanNotFound
	}
	if err := ValidateTenantScope(item.WorkspaceID, item.TenantID); err != nil {
		return err
	}
	if item.ID == "" {
		return fmt.Errorf("%w: plan item id is required", ErrInvalidPlanData)
	}
	if err := ValidateContentPlan(item); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *item
	if len(item.AssetMetadata.Hashtags) > 0 {
		copied.AssetMetadata.Hashtags = make([]string, len(item.AssetMetadata.Hashtags))
		copy(copied.AssetMetadata.Hashtags, item.AssetMetadata.Hashtags)
	}
	if item.LastApprovedAt != nil {
		t := *item.LastApprovedAt
		copied.LastApprovedAt = &t
	}

	now := time.Now().UTC()
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = now
	}
	copied.UpdatedAt = now

	r.contentPlans[item.ID] = &copied
	return nil
}

func (r *MemoryRepository) GetContentPlanItem(ctx context.Context, itemID, requestingTenantID string) (*ContentPlanItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.contentPlans[itemID]
	if !exists {
		return nil, ErrContentPlanNotFound
	}
	if item.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAccessDenied
	}

	copied := *item
	if len(item.AssetMetadata.Hashtags) > 0 {
		copied.AssetMetadata.Hashtags = make([]string, len(item.AssetMetadata.Hashtags))
		copy(copied.AssetMetadata.Hashtags, item.AssetMetadata.Hashtags)
	}
	if item.LastApprovedAt != nil {
		t := *item.LastApprovedAt
		copied.LastApprovedAt = &t
	}
	return &copied, nil
}

func (r *MemoryRepository) ListContentPlanItems(ctx context.Context, filter ContentCalendarFilter) ([]ContentPlanItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []ContentPlanItem
	for _, item := range r.contentPlans {
		if filter.WorkspaceID != "" && item.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.RequestingTenantID != "" && item.TenantID != filter.RequestingTenantID {
			continue
		}
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		if filter.Format != "" && item.Format != filter.Format {
			continue
		}
		if filter.StartDateUTC != nil && item.ScheduledSlotUTC.Before(*filter.StartDateUTC) {
			continue
		}
		if filter.EndDateUTC != nil && item.ScheduledSlotUTC.After(*filter.EndDateUTC) {
			continue
		}

		copied := *item
		if len(item.AssetMetadata.Hashtags) > 0 {
			copied.AssetMetadata.Hashtags = make([]string, len(item.AssetMetadata.Hashtags))
			copy(copied.AssetMetadata.Hashtags, item.AssetMetadata.Hashtags)
		}
		if item.LastApprovedAt != nil {
			t := *item.LastApprovedAt
			copied.LastApprovedAt = &t
		}
		result = append(result, copied)
	}
	return result, nil
}

func (r *MemoryRepository) DeleteContentPlanItem(ctx context.Context, itemID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.contentPlans[itemID]
	if !exists {
		return ErrContentPlanNotFound
	}
	if item.TenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	delete(r.contentPlans, itemID)
	return nil
}

// Post Analytics Implementation (IMP-LI-15, AT-028, AT-010, AT-012)

func (r *MemoryRepository) SavePostAnalytics(ctx context.Context, snapshot *PostAnalyticsSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ValidateAnalyticsSnapshot(snapshot); err != nil {
		return err
	}

	copied := *snapshot
	copied.Metrics = make(map[string]MetricItem, len(snapshot.Metrics))
	for k, v := range snapshot.Metrics {
		copied.Metrics[k] = v
	}
	if len(snapshot.Engagers) > 0 {
		copied.Engagers = make([]EngagerProfile, len(snapshot.Engagers))
		copy(copied.Engagers, snapshot.Engagers)
	}

	r.postAnalytics[snapshot.SnapshotID] = &copied
	return nil
}

func (r *MemoryRepository) GetPostAnalytics(ctx context.Context, snapshotID, requestingTenantID string) (*PostAnalyticsSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snap, exists := r.postAnalytics[snapshotID]
	if !exists {
		return nil, ErrPostAnalyticsNotFound
	}
	if snap.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAnalyticsDeny
	}

	copied := *snap
	copied.Metrics = make(map[string]MetricItem, len(snap.Metrics))
	for k, v := range snap.Metrics {
		copied.Metrics[k] = v
	}
	if len(snap.Engagers) > 0 {
		copied.Engagers = make([]EngagerProfile, len(snap.Engagers))
		copy(copied.Engagers, snap.Engagers)
	}
	return &copied, nil
}

func (r *MemoryRepository) GetPostAnalyticsByURN(ctx context.Context, postURN, requestingTenantID string) (*PostAnalyticsSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, snap := range r.postAnalytics {
		if snap.PostURN == postURN {
			if snap.TenantID != requestingTenantID {
				return nil, ErrCrossTenantAnalyticsDeny
			}
			copied := *snap
			copied.Metrics = make(map[string]MetricItem, len(snap.Metrics))
			for k, v := range snap.Metrics {
				copied.Metrics[k] = v
			}
			if len(snap.Engagers) > 0 {
				copied.Engagers = make([]EngagerProfile, len(snap.Engagers))
				copy(copied.Engagers, snap.Engagers)
			}
			return &copied, nil
		}
	}
	return nil, ErrPostAnalyticsNotFound
}

func (r *MemoryRepository) ListPostAnalytics(ctx context.Context, workspaceID, requestingTenantID string) ([]PostAnalyticsSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []PostAnalyticsSnapshot
	for _, snap := range r.postAnalytics {
		if workspaceID != "" && snap.WorkspaceID != workspaceID {
			continue
		}
		if requestingTenantID != "" && snap.TenantID != requestingTenantID {
			continue
		}

		copied := *snap
		copied.Metrics = make(map[string]MetricItem, len(snap.Metrics))
		for k, v := range snap.Metrics {
			copied.Metrics[k] = v
		}
		if len(snap.Engagers) > 0 {
			copied.Engagers = make([]EngagerProfile, len(snap.Engagers))
			copy(copied.Engagers, snap.Engagers)
		}
		results = append(results, copied)
	}
	return results, nil
}

func (r *MemoryRepository) DeletePostAnalytics(ctx context.Context, snapshotID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	snap, exists := r.postAnalytics[snapshotID]
	if !exists {
		return ErrPostAnalyticsNotFound
	}
	if snap.TenantID != requestingTenantID {
		return ErrCrossTenantAnalyticsDeny
	}
	delete(r.postAnalytics, snapshotID)
	return nil
}

// Employee Advocacy Repository Implementation (IMP-LI-16)

func (r *MemoryRepository) SaveAdvocacyCampaign(ctx context.Context, campaign *AdvocacyCampaign) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *campaign
	if len(campaign.Governance.AllowedHashtags) > 0 {
		copied.Governance.AllowedHashtags = append([]string{}, campaign.Governance.AllowedHashtags...)
	}
	if len(campaign.Governance.ForbiddenKeywords) > 0 {
		copied.Governance.ForbiddenKeywords = append([]string{}, campaign.Governance.ForbiddenKeywords...)
	}
	if len(campaign.Variants) > 0 {
		copied.Variants = make([]AdvocacyCopyVariant, len(campaign.Variants))
		for i, v := range campaign.Variants {
			varCopy := v
			if len(v.TargetTags) > 0 {
				varCopy.TargetTags = append([]string{}, v.TargetTags...)
			}
			copied.Variants[i] = varCopy
		}
	}
	r.advocacyCampaigns[campaign.CampaignID] = &copied
	return nil
}

func (r *MemoryRepository) GetAdvocacyCampaign(ctx context.Context, campaignID, requestingTenantID string) (*AdvocacyCampaign, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, exists := r.advocacyCampaigns[campaignID]
	if !exists {
		return nil, ErrCampaignNotFound
	}
	if requestingTenantID != "" && c.TenantID != requestingTenantID {
		return nil, ErrCrossTenantAnalyticsDeny
	}
	copied := *c
	if len(c.Governance.AllowedHashtags) > 0 {
		copied.Governance.AllowedHashtags = append([]string{}, c.Governance.AllowedHashtags...)
	}
	if len(c.Governance.ForbiddenKeywords) > 0 {
		copied.Governance.ForbiddenKeywords = append([]string{}, c.Governance.ForbiddenKeywords...)
	}
	if len(c.Variants) > 0 {
		copied.Variants = make([]AdvocacyCopyVariant, len(c.Variants))
		for i, v := range c.Variants {
			varCopy := v
			if len(v.TargetTags) > 0 {
				varCopy.TargetTags = append([]string{}, v.TargetTags...)
			}
			copied.Variants[i] = varCopy
		}
	}
	return &copied, nil
}

func (r *MemoryRepository) ListAdvocacyCampaigns(ctx context.Context, workspaceID, requestingTenantID string) ([]AdvocacyCampaign, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []AdvocacyCampaign
	for _, c := range r.advocacyCampaigns {
		if workspaceID != "" && c.WorkspaceID != workspaceID {
			continue
		}
		if requestingTenantID != "" && c.TenantID != requestingTenantID {
			continue
		}
		copied := *c
		if len(c.Governance.AllowedHashtags) > 0 {
			copied.Governance.AllowedHashtags = append([]string{}, c.Governance.AllowedHashtags...)
		}
		if len(c.Governance.ForbiddenKeywords) > 0 {
			copied.Governance.ForbiddenKeywords = append([]string{}, c.Governance.ForbiddenKeywords...)
		}
		if len(c.Variants) > 0 {
			copied.Variants = make([]AdvocacyCopyVariant, len(c.Variants))
			for i, v := range c.Variants {
				varCopy := v
				if len(v.TargetTags) > 0 {
					varCopy.TargetTags = append([]string{}, v.TargetTags...)
				}
				copied.Variants[i] = varCopy
			}
		}
		results = append(results, copied)
	}
	return results, nil
}

func (r *MemoryRepository) DeleteAdvocacyCampaign(ctx context.Context, campaignID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, exists := r.advocacyCampaigns[campaignID]
	if !exists {
		return ErrCampaignNotFound
	}
	if requestingTenantID != "" && c.TenantID != requestingTenantID {
		return ErrCrossTenantAnalyticsDeny
	}
	delete(r.advocacyCampaigns, campaignID)
	return nil
}

func (r *MemoryRepository) RecordEmployeeShare(ctx context.Context, event *EmployeeShareEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *event
	r.employeeShares[event.ShareID] = &copied

	if c, ok := r.advocacyCampaigns[event.CampaignID]; ok {
		c.ShareCount++
	}
	return nil
}

func (r *MemoryRepository) ListEmployeeShares(ctx context.Context, campaignID, workspaceID, requestingTenantID string) ([]EmployeeShareEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []EmployeeShareEvent
	for _, s := range r.employeeShares {
		if campaignID != "" && s.CampaignID != campaignID {
			continue
		}
		if workspaceID != "" && s.WorkspaceID != workspaceID {
			continue
		}
		if requestingTenantID != "" && s.TenantID != requestingTenantID {
			continue
		}
		results = append(results, *s)
	}
	return results, nil
}

// -------------------------------------------------------------------------
// Provider Fallback & Diagnostics (IMP-LI-17, LI-17, AT-010, AT-016, SRC-L5)
// -------------------------------------------------------------------------

func (r *MemoryRepository) RegisterProviderAdapter(ctx context.Context, adapter *ProviderAdapter) error {
	if adapter == nil {
		return ErrInvalidProviderAdapter
	}
	if err := ValidateProviderAdapter(*adapter); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *adapter
	copied.SupportedActions = make([]string, len(adapter.SupportedActions))
	copy(copied.SupportedActions, adapter.SupportedActions)
	copied.ExpectedScopes = make([]string, len(adapter.ExpectedScopes))
	copy(copied.ExpectedScopes, adapter.ExpectedScopes)
	copied.GrantedScopes = make([]string, len(adapter.GrantedScopes))
	copy(copied.GrantedScopes, adapter.GrantedScopes)

	now := time.Now().UTC()
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = now
	}
	copied.UpdatedAt = now

	r.providerAdapters[adapter.ProviderID] = &copied
	return nil
}

func (r *MemoryRepository) GetProviderAdapter(ctx context.Context, providerID, requestingTenantID string) (*ProviderAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.providerAdapters[providerID]
	if !exists {
		return nil, ErrProviderNotFound
	}
	if requestingTenantID != "" && p.TenantID != requestingTenantID {
		return nil, ErrCrossTenantProviderAccess
	}

	copied := *p
	copied.SupportedActions = append([]string(nil), p.SupportedActions...)
	copied.ExpectedScopes = append([]string(nil), p.ExpectedScopes...)
	copied.GrantedScopes = append([]string(nil), p.GrantedScopes...)
	return &copied, nil
}

func (r *MemoryRepository) ListProviderAdapters(ctx context.Context, workspaceID, requestingTenantID string) ([]ProviderAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []ProviderAdapter
	for _, p := range r.providerAdapters {
		if workspaceID != "" && p.WorkspaceID != workspaceID {
			continue
		}
		if requestingTenantID != "" && p.TenantID != requestingTenantID {
			continue
		}
		copied := *p
		copied.SupportedActions = append([]string(nil), p.SupportedActions...)
		copied.ExpectedScopes = append([]string(nil), p.ExpectedScopes...)
		copied.GrantedScopes = append([]string(nil), p.GrantedScopes...)
		results = append(results, copied)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Tier < results[j].Tier
	})
	return results, nil
}

func (r *MemoryRepository) UpdateProviderHealth(ctx context.Context, providerID, requestingTenantID string, status ProviderHealthStatus, latencyMs int64, errStr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.providerAdapters[providerID]
	if !exists {
		return ErrProviderNotFound
	}
	if requestingTenantID != "" && p.TenantID != requestingTenantID {
		return ErrCrossTenantProviderAccess
	}

	p.HealthStatus = status
	p.LatencyMs = latencyMs
	p.LastCheckedAt = time.Now().UTC()
	p.LastErrorMessage = ScrubDiagnosticSecrets(errStr)
	if status == HealthStatusHealthy {
		p.ConsecutiveFailures = 0
	} else {
		p.ConsecutiveFailures++
	}
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *MemoryRepository) DeleteProviderAdapter(ctx context.Context, providerID, requestingTenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.providerAdapters[providerID]
	if !exists {
		return ErrProviderNotFound
	}
	if requestingTenantID != "" && p.TenantID != requestingTenantID {
		return ErrCrossTenantProviderAccess
	}

	delete(r.providerAdapters, providerID)
	return nil
}

func (r *MemoryRepository) RecordDiagnosticProbe(ctx context.Context, probe *DiagnosticProbeResult) error {
	if probe == nil {
		return errors.New("nil probe result")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *probe
	r.diagnosticProbes[probe.ProviderID] = &copied
	return nil
}

func (r *MemoryRepository) ListDiagnosticProbes(ctx context.Context, providerID string) ([]DiagnosticProbeResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []DiagnosticProbeResult
	for _, pr := range r.diagnosticProbes {
		if providerID != "" && pr.ProviderID != providerID {
			continue
		}
		results = append(results, *pr)
	}
	return results, nil
}

// ============================================================================
// Limits, CAPTCHA/reauth handling, activity (IMP-LI-18, LI-18, AT-008, AT-009, AT-010, REQ-009, REQ-010)
// ============================================================================

func (r *MemoryRepository) SaveAccountSafetyState(ctx context.Context, state *AccountSafetyState) error {
	if state == nil {
		return errors.New("nil account safety state")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *state
	if copied.UpdatedAt.IsZero() {
		copied.UpdatedAt = time.Now().UTC()
	}
	r.accountSafetyStates[state.AccountID] = &copied
	return nil
}

func (r *MemoryRepository) GetAccountSafetyState(ctx context.Context, accountID, requestingTenantID string) (*AccountSafetyState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, exists := r.accountSafetyStates[accountID]
	if !exists {
		// Return a default active state if not yet recorded
		return &AccountSafetyState{
			AccountID: accountID,
			TenantID:  requestingTenantID,
			Status:    SafetyStatusActive,
			UpdatedAt: time.Now().UTC(),
		}, nil
	}
	if requestingTenantID != "" && state.TenantID != "" && state.TenantID != requestingTenantID {
		return nil, ErrCrossTenantProviderAccess
	}

	copied := *state
	return &copied, nil
}

func (r *MemoryRepository) SaveRestrictionIncident(ctx context.Context, incident *PlatformRestrictionIncident) error {
	if incident == nil {
		return errors.New("nil restriction incident")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *incident
	if copied.OccurredAt.IsZero() {
		copied.OccurredAt = time.Now().UTC()
	}
	r.restrictionIncidents[incident.AccountID] = append(r.restrictionIncidents[incident.AccountID], &copied)
	return nil
}

func (r *MemoryRepository) ListRestrictionIncidents(ctx context.Context, accountID, requestingTenantID string) ([]PlatformRestrictionIncident, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []PlatformRestrictionIncident
	for accID, list := range r.restrictionIncidents {
		if accountID != "" && accID != accountID {
			continue
		}
		for _, inc := range list {
			if requestingTenantID != "" && inc.TenantID != "" && inc.TenantID != requestingTenantID {
				continue
			}
			results = append(results, *inc)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].OccurredAt.After(results[j].OccurredAt)
	})

	return results, nil
}

func (r *MemoryRepository) RecordActivityLog(ctx context.Context, entry *ActivityLogEntry) error {
	if entry == nil {
		return errors.New("nil activity log entry")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *entry
	if copied.Timestamp.IsZero() {
		copied.Timestamp = time.Now().UTC()
	}
	r.activityLogs[entry.AccountID] = append(r.activityLogs[entry.AccountID], &copied)
	return nil
}

func (r *MemoryRepository) ListActivityLog(ctx context.Context, accountID, requestingTenantID string, limit int) ([]ActivityLogEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []ActivityLogEntry
	for accID, list := range r.activityLogs {
		if accountID != "" && accID != accountID {
			continue
		}
		for _, e := range list {
			if requestingTenantID != "" && e.TenantID != "" && e.TenantID != requestingTenantID {
				continue
			}
			results = append(results, *e)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// -------------------------------------------------------------------------
// Session Lifecycle & Optional Local Tools (IMP-LI-19, LI-19, AT-016, FND-009, SRC-L3)
// -------------------------------------------------------------------------

func (r *MemoryRepository) SaveSession(ctx context.Context, session *LinkedInSessionRecord) error {
	if session == nil || session.SessionID == "" {
		return errors.New("invalid session record")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *session
	r.sessions[session.SessionID] = &copied
	return nil
}

func (r *MemoryRepository) GetSession(ctx context.Context, sessionID, requestingTenantID string) (*LinkedInSessionRecord, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, errors.New("sessionID cannot be empty")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	sess, ok := r.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	// Per-owner & multi-tenant isolation barrier (AT-011, AT-012)
	if requestingTenantID != "" && sess.TenantID != "" && sess.TenantID != requestingTenantID {
		return nil, ErrCrossTenantSessionAccess
	}
	copied := *sess
	return &copied, nil
}

func (r *MemoryRepository) ListSessions(ctx context.Context, ownerID, requestingTenantID string) ([]LinkedInSessionRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []LinkedInSessionRecord
	for _, sess := range r.sessions {
		if requestingTenantID != "" && sess.TenantID != "" && sess.TenantID != requestingTenantID {
			continue
		}
		if ownerID != "" && sess.OwnerID != ownerID {
			continue
		}
		results = append(results, *sess)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	return results, nil
}

func (r *MemoryRepository) SaveViewerLease(ctx context.Context, lease *ShortLivedViewerLease) error {
	if lease == nil || lease.LeaseID == "" {
		return errors.New("invalid viewer lease record")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *lease
	r.viewerLeases[lease.LeaseID] = &copied
	return nil
}

func (r *MemoryRepository) GetViewerLease(ctx context.Context, leaseID string) (*ShortLivedViewerLease, error) {
	if strings.TrimSpace(leaseID) == "" {
		return nil, errors.New("leaseID cannot be empty")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	lease, ok := r.viewerLeases[leaseID]
	if !ok {
		return nil, errors.New("viewer lease not found")
	}
	copied := *lease
	return &copied, nil
}

// Relationship CSV/Sheets Export Audit methods (IMP-LI-20, LI-20, REQ-019, AT-013)
func (r *MemoryRepository) SaveRelationshipExportAudit(ctx context.Context, audit *RelationshipExportAuditRecord) error {
	if audit == nil {
		return errors.New("audit record cannot be nil")
	}
	if strings.TrimSpace(audit.WorkspaceID) == "" {
		return ErrEmptyExportWorkspace
	}
	if strings.TrimSpace(audit.TenantID) == "" {
		return ErrEmptyExportTenant
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *audit
	key := audit.WorkspaceID
	r.relationshipExportAudits[key] = append(r.relationshipExportAudits[key], &copied)
	return nil
}

func (r *MemoryRepository) ListRelationshipExportAudits(ctx context.Context, workspaceID, requestingTenantID string) ([]RelationshipExportAuditRecord, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, ErrEmptyExportWorkspace
	}
	if strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrEmptyExportTenant
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	records := r.relationshipExportAudits[workspaceID]
	var out []RelationshipExportAuditRecord
	for _, rec := range records {
		if rec.TenantID != requestingTenantID {
			continue
		}
		out = append(out, *rec)
	}

	// Sort newest first
	sort.Slice(out, func(i, j int) bool {
		return out[i].ExportedAt.After(out[j].ExportedAt)
	})

	return out, nil
}


