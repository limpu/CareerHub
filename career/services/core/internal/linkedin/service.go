package linkedin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Service provides high-level business operations for LinkedIn integrations (IMP-LI-01).
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ConnectAccount establishes or updates a user's LinkedIn connection and audits granted capabilities.
func (s *Service) ConnectAccount(ctx context.Context, userID, memberID, displayName, email string, grantedScopes []string, rawToken string, expiresInHours int) (*LinkedInConnection, error) {
	if userID == "" {
		return nil, errorsNew("user_id cannot be empty")
	}
	if len(grantedScopes) == 0 {
		return nil, ErrInvalidScopeRequest
	}

	now := time.Now().UTC()
	var expiresAt *time.Time
	if expiresInHours > 0 {
		t := now.Add(time.Duration(expiresInHours) * time.Hour)
		expiresAt = &t
	} else if expiresInHours < 0 {
		// Past date for expired token testing
		t := now.Add(time.Duration(expiresInHours) * time.Hour)
		expiresAt = &t
	}

	tokenStatus := "active"
	if expiresAt != nil && now.After(*expiresAt) {
		tokenStatus = "expired"
	}

	// Mask secret for user-facing and audit storage (AT-016)
	masked := "tok_li_****_unknown"
	if len(rawToken) >= 8 {
		masked = fmt.Sprintf("tok_li_****_%s", rawToken[len(rawToken)-4:])
	}

	// Calculate denied scopes from standard list
	allKnownScopes := []string{"openid", "profile", "email", "r_basicprofile", "w_member_social", "r_messages", "w_messages"}
	scopeSet := make(map[string]bool)
	for _, sc := range grantedScopes {
		scopeSet[strings.ToLower(strings.TrimSpace(sc))] = true
	}

	var deniedScopes []string
	for _, ks := range allKnownScopes {
		if !scopeSet[ks] {
			deniedScopes = append(deniedScopes, ks)
		}
	}

	connStatus := ConnConnected
	if tokenStatus == "expired" {
		connStatus = ConnReauthRequired
	} else if len(deniedScopes) > 0 {
		connStatus = ConnPartiallyAuthorized
	}

	conn := &LinkedInConnection{
		ConnectionID:    fmt.Sprintf("conn-li-%d", time.Now().UnixNano()),
		UserID:          userID,
		MemberID:        memberID,
		DisplayName:     displayName,
		Email:           email,
		GrantedScopes:   grantedScopes,
		DeniedScopes:    deniedScopes,
		MaskedToken:     masked,
		TokenStatus:     tokenStatus,
		ConnectedAt:     now,
		ExpiresAt:       expiresAt,
		Status:          connStatus,
		LastInspectedAt: now,
	}

	// Compute capability inspection
	conn.Capabilities = InspectCapabilities(conn)

	if err := s.repo.SaveConnection(ctx, conn); err != nil {
		return nil, err
	}

	return conn, nil
}

// GetConnection retrieves the stored connection record for a user.
func (s *Service) GetConnection(ctx context.Context, userID string) (*LinkedInConnection, error) {
	return s.repo.GetConnection(ctx, userID)
}

// InspectCapabilities runs fresh audit on the user's connection.
func (s *Service) InspectCapabilities(ctx context.Context, userID string) (*LinkedInConnection, error) {
	conn, err := s.repo.GetConnection(ctx, userID)
	if err != nil {
		return nil, err
	}

	conn.Capabilities = InspectCapabilities(conn)
	conn.LastInspectedAt = time.Now().UTC()

	if conn.ExpiresAt != nil && time.Now().UTC().After(*conn.ExpiresAt) {
		conn.TokenStatus = "expired"
		conn.Status = ConnReauthRequired
	}

	_ = s.repo.SaveConnection(ctx, conn)
	return conn, nil
}

// DisconnectAccount revokes and removes connection records for a user.
func (s *Service) DisconnectAccount(ctx context.Context, userID string) error {
	return s.repo.DeleteConnection(ctx, userID)
}

// CheckActionPermission evaluates if a live action can be dispatched or fails closed (AT-010).
func (s *Service) CheckActionPermission(ctx context.Context, userID, action string) (bool, *InspectedCapability, error) {
	conn, err := s.repo.GetConnection(ctx, userID)
	if err != nil {
		return false, nil, err
	}

	return CanExecuteAction(conn, action)
}

// ImportProfileArchive processes an uploaded Basic_LinkedInData.zip archive.
func (s *Service) ImportProfileArchive(ctx context.Context, userID string, zipBytes []byte) (*LinkedInImportedProfile, error) {
	p, err := ParseLinkedInZipArchive(zipBytes)
	if err != nil {
		return nil, err
	}
	p.UserID = userID
	if err := s.repo.SaveProfileImport(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ImportProfileArchiveFiles processes already-extracted CSV files map.
func (s *Service) ImportProfileArchiveFiles(ctx context.Context, userID string, files map[string]string) (*LinkedInImportedProfile, error) {
	p, err := ParseLinkedInArchiveFiles(files)
	if err != nil {
		return nil, err
	}
	p.UserID = userID
	if err := s.repo.SaveProfileImport(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ImportPastedProfile parses unstructured text pasted from a public profile page.
func (s *Service) ImportPastedProfile(ctx context.Context, userID, rawText string) (*LinkedInImportedProfile, error) {
	p, err := ParseLinkedInPastedText(rawText)
	if err != nil {
		return nil, err
	}
	p.UserID = userID
	if err := s.repo.SaveProfileImport(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ImportEnterpriseProfile parses structured profile response from LinkedIn partner APIs.
func (s *Service) ImportEnterpriseProfile(ctx context.Context, userID string, payload []byte) (*LinkedInImportedProfile, error) {
	p, err := ParseLinkedInEnterprisePayload(payload)
	if err != nil {
		return nil, err
	}
	p.UserID = userID
	if err := s.repo.SaveProfileImport(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ImportOIDCProfileFallback provides truth-in-advertising fallback when full scanning is requested on standard consumer OAuth (AT-010).
func (s *Service) ImportOIDCProfileFallback(ctx context.Context, userID string, claims map[string]interface{}) (*LinkedInImportedProfile, error) {
	p := ResolveOIDCBasicFallback(claims)
	p.UserID = userID
	if err := s.repo.SaveProfileImport(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetLatestProfileImport retrieves the most recent profile import for the user.
func (s *Service) GetLatestProfileImport(ctx context.Context, userID string) (*LinkedInImportedProfile, error) {
	return s.repo.GetLatestProfileImport(ctx, userID)
}

// GetProfileImport retrieves a specific profile import by ID.
func (s *Service) GetProfileImport(ctx context.Context, importID string) (*LinkedInImportedProfile, error) {
	return s.repo.GetProfileImport(ctx, importID)
}

// MarkProfileImportMerged updates the import status to merged (AT-007).
func (s *Service) MarkProfileImportMerged(ctx context.Context, importID string) (*LinkedInImportedProfile, error) {
	p, err := s.repo.GetProfileImport(ctx, importID)
	if err != nil {
		return nil, err
	}
	p.Status = "merged"
	if err := s.repo.SaveProfileImport(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// GenerateOptimizationReport produces fact-grounded profile enhancement recommendations (LI-03, AT-003).
func (s *Service) GenerateOptimizationReport(ctx context.Context, in CandidateOptimizationInput) (*ProfileOptimizationReport, error) {
	if in.UserID == "" {
		return nil, errorsNew("user_id is required")
	}

	report := GenerateProfileOptimization(in)
	if err := s.repo.SaveOptimizationReport(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

// GetLatestOptimizationReport retrieves the most recent optimization recommendations for a user.
func (s *Service) GetLatestOptimizationReport(ctx context.Context, userID string) (*ProfileOptimizationReport, error) {
	return s.repo.GetLatestOptimizationReport(ctx, userID)
}

// ApproveOptimizationSuggestion registers candidate approval for an AI recommendation (AT-007).
func (s *Service) ApproveOptimizationSuggestion(ctx context.Context, reportID, suggestionID string) (*SectionSuggestion, error) {
	rep, err := s.repo.GetOptimizationReport(ctx, reportID)
	if err != nil {
		return nil, err
	}

	sug, err := ApproveSuggestion(rep, suggestionID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveOptimizationReport(ctx, rep); err != nil {
		return nil, err
	}
	return sug, nil
}

// RejectOptimizationSuggestion registers candidate rejection of an AI recommendation.
func (s *Service) RejectOptimizationSuggestion(ctx context.Context, reportID, suggestionID string) (*SectionSuggestion, error) {
	rep, err := s.repo.GetOptimizationReport(ctx, reportID)
	if err != nil {
		return nil, err
	}

	sug, err := RejectSuggestion(rep, suggestionID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveOptimizationReport(ctx, rep); err != nil {
		return nil, err
	}
	return sug, nil
}

// CustomEditOptimizationSuggestion applies human edits and invalidates previous approvals (AT-007).
func (s *Service) CustomEditOptimizationSuggestion(ctx context.Context, reportID, suggestionID, customText string) (*SectionSuggestion, error) {
	rep, err := s.repo.GetOptimizationReport(ctx, reportID)
	if err != nil {
		return nil, err
	}

	sug, err := CustomEditSuggestion(rep, suggestionID, customText)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveOptimizationReport(ctx, rep); err != nil {
		return nil, err
	}
	return sug, nil
}

// -------------------------------------------------------------------------
// Records Engine Service Methods (IMP-LI-04, AT-011, AT-012, FND-009, FND-012)
// -------------------------------------------------------------------------

// UpsertPersonRecord creates or updates a person record with tenant scoping & validation.
func (s *Service) UpsertPersonRecord(ctx context.Context, p *PersonRecord) error {
	if p == nil {
		return ErrInvalidRecordData
	}
	if err := ValidateTenantScope(p.WorkspaceID, p.TenantID); err != nil {
		return err
	}
	// Permitted fields validation (FND-009, FND-012)
	dataMap := map[string]interface{}{
		"headline": p.Headline,
		"skills":   strings.Join(p.Skills, " "),
		"location": p.Location,
	}
	if err := ValidatePermittedFields(dataMap); err != nil {
		return err
	}
	return s.repo.SavePersonRecord(ctx, p)
}

func (s *Service) GetPersonRecord(ctx context.Context, recordID, requestingTenantID string) (*PersonRecord, error) {
	return s.repo.GetPersonRecord(ctx, recordID, requestingTenantID)
}

func (s *Service) ListPersonRecords(ctx context.Context, filter RecordFilter) ([]PersonRecord, error) {
	if err := ValidateTenantScope(filter.WorkspaceID, filter.TenantID); err != nil {
		return nil, err
	}
	return s.repo.ListPersonRecords(ctx, filter)
}

// UpsertCompanyRecord creates or updates a company record with tenant scoping & validation.
func (s *Service) UpsertCompanyRecord(ctx context.Context, c *CompanyRecord) error {
	if c == nil {
		return ErrInvalidRecordData
	}
	if err := ValidateTenantScope(c.WorkspaceID, c.TenantID); err != nil {
		return err
	}
	dataMap := map[string]interface{}{
		"company_name": c.CompanyName,
		"specialties":  strings.Join(c.Specialties, " "),
		"domain":       c.Domain,
	}
	if err := ValidatePermittedFields(dataMap); err != nil {
		return err
	}
	return s.repo.SaveCompanyRecord(ctx, c)
}

func (s *Service) GetCompanyRecord(ctx context.Context, recordID, requestingTenantID string) (*CompanyRecord, error) {
	return s.repo.GetCompanyRecord(ctx, recordID, requestingTenantID)
}

func (s *Service) ListCompanyRecords(ctx context.Context, filter RecordFilter) ([]CompanyRecord, error) {
	if err := ValidateTenantScope(filter.WorkspaceID, filter.TenantID); err != nil {
		return nil, err
	}
	return s.repo.ListCompanyRecords(ctx, filter)
}

// UpsertJobRecord creates or updates a job record with tenant scoping & validation.
func (s *Service) UpsertJobRecord(ctx context.Context, j *JobRecord) error {
	if j == nil {
		return ErrInvalidRecordData
	}
	if err := ValidateTenantScope(j.WorkspaceID, j.TenantID); err != nil {
		return err
	}
	dataMap := map[string]interface{}{
		"title":               j.Title,
		"description_snippet": j.DescriptionSnippet,
	}
	if err := ValidatePermittedFields(dataMap); err != nil {
		return err
	}
	return s.repo.SaveJobRecord(ctx, j)
}

func (s *Service) GetJobRecord(ctx context.Context, recordID, requestingTenantID string) (*JobRecord, error) {
	return s.repo.GetJobRecord(ctx, recordID, requestingTenantID)
}

func (s *Service) ListJobRecords(ctx context.Context, filter RecordFilter) ([]JobRecord, error) {
	if err := ValidateTenantScope(filter.WorkspaceID, filter.TenantID); err != nil {
		return nil, err
	}
	return s.repo.ListJobRecords(ctx, filter)
}

// UpsertPostRecord creates or updates a post record with tenant scoping & validation.
func (s *Service) UpsertPostRecord(ctx context.Context, post *PostRecord) error {
	if post == nil {
		return ErrInvalidRecordData
	}
	if err := ValidateTenantScope(post.WorkspaceID, post.TenantID); err != nil {
		return err
	}
	dataMap := map[string]interface{}{
		"commentary": post.Commentary,
	}
	if err := ValidatePermittedFields(dataMap); err != nil {
		return err
	}
	return s.repo.SavePostRecord(ctx, post)
}

func (s *Service) GetPostRecord(ctx context.Context, recordID, requestingTenantID string) (*PostRecord, error) {
	return s.repo.GetPostRecord(ctx, recordID, requestingTenantID)
}

func (s *Service) ListPostRecords(ctx context.Context, filter RecordFilter) ([]PostRecord, error) {
	if err := ValidateTenantScope(filter.WorkspaceID, filter.TenantID); err != nil {
		return nil, err
	}
	return s.repo.ListPostRecords(ctx, filter)
}

// -------------------------------------------------------------------------
// Discovery Engine (IMP-LI-05, LI-05, AT-010, FND-011, SRC-L1, SRC-L3)
// -------------------------------------------------------------------------

// DiscoverEntities executes professional candidate & company discovery within tenant scope (AT-010, AT-011, AT-012).
func (s *Service) DiscoverEntities(ctx context.Context, criteria DiscoveryCriteria, tenantID, workspaceID string) (*DiscoveryResult, error) {
	if err := ValidateTenantScope(workspaceID, tenantID); err != nil {
		return nil, err
	}
	if err := ValidateDiscoveryCriteria(criteria); err != nil {
		return nil, err
	}

	boolQuery, err := BuildBooleanQuery(criteria)
	if err != nil {
		return nil, err
	}

	// 1. Match Vault Persons
	persons, err := s.repo.ListPersonRecords(ctx, RecordFilter{
		WorkspaceID: workspaceID,
		TenantID:    tenantID,
	})
	if err != nil {
		return nil, err
	}

	var matchedPersons []PersonRecord
	for _, p := range persons {
		matched := false
		lowerRole := strings.ToLower(p.CurrentRole + " " + p.Headline)
		if criteria.TargetRole != "" && strings.Contains(lowerRole, strings.ToLower(criteria.TargetRole)) {
			matched = true
		}
		for _, alt := range criteria.RoleAlternatives {
			if strings.Contains(lowerRole, strings.ToLower(alt)) {
				matched = true
			}
		}
		for _, comp := range criteria.CurrentCompanies {
			if strings.Contains(strings.ToLower(p.CurrentCompany), strings.ToLower(comp)) {
				matched = true
			}
		}
		// Check exclusions
		if matched {
			for _, excl := range criteria.Exclusions {
				if strings.Contains(lowerRole, strings.ToLower(excl)) {
					matched = false
					break
				}
			}
		}
		if matched {
			matchedPersons = append(matchedPersons, p)
		}
	}

	// 2. Match Vault Companies
	companies, err := s.repo.ListCompanyRecords(ctx, RecordFilter{
		WorkspaceID: workspaceID,
		TenantID:    tenantID,
	})
	if err != nil {
		return nil, err
	}

	var matchedCompanies []CompanyRecord
	for _, c := range companies {
		matched := false
		for _, comp := range criteria.CurrentCompanies {
			if strings.Contains(strings.ToLower(c.CompanyName), strings.ToLower(comp)) {
				matched = true
				break
			}
		}
		for _, ind := range criteria.Industries {
			if strings.Contains(strings.ToLower(c.Industry), strings.ToLower(ind)) {
				matched = true
				break
			}
		}
		if matched {
			matchedCompanies = append(matchedCompanies, c)
		}
	}

	platformNotice := "Live unrestricted LinkedIn search requires enterprise LinkedIn Talent/Recruiter licenses. Matches above reflect verified Vault records. Use the generated Boolean Query and Deep Link to view native external results without account safety risks (AT-010)."

	return &DiscoveryResult{
		Criteria:          criteria,
		BooleanQuery:      boolQuery,
		MatchedPersons:    matchedPersons,
		MatchedCompanies:  matchedCompanies,
		TotalVaultMatches: len(matchedPersons) + len(matchedCompanies),
		PlatformNotice:    platformNotice,
		GeneratedAt:       time.Now().UTC(),
	}, nil
}

// --- Recruiter & Hiring Lead Workspace (IMP-LI-06, LI-06, AT-011) ---

func (s *Service) CreateRecruiterLead(ctx context.Context, lead *RecruiterLead) (*RecruiterLead, error) {
	if lead == nil {
		return nil, errorsNew("recruiter lead cannot be nil")
	}
	if lead.ID == "" {
		lead.ID = fmt.Sprintf("lead-%d", time.Now().UnixNano())
	}
	if err := s.repo.SaveRecruiterLead(ctx, lead); err != nil {
		return nil, err
	}
	return s.repo.GetRecruiterLead(ctx, lead.ID, lead.TenantID)
}

func (s *Service) CreateLeadFromPersonRecord(ctx context.Context, recordID, workspaceID, tenantID string) (*RecruiterLead, error) {
	p, err := s.repo.GetPersonRecord(ctx, recordID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve source person record: %w", err)
	}

	profileURL := p.PublicProfileURL
	if profileURL == "" {
		profileURL = fmt.Sprintf("https://www.linkedin.com/in/%s", p.VanitySlug)
	}

	companyName := p.CurrentCompany
	if companyName == "" {
		companyName = "Unknown Company"
	}

	lead := &RecruiterLead{
		ID:               fmt.Sprintf("lead-person-%s", recordID),
		WorkspaceID:      workspaceID,
		TenantID:         tenantID,
		RecruiterName:    p.FullName,
		RecruiterTitle:   p.Headline,
		Company:          companyName,
		LinkedInURL:      profileURL,
		SourceEntityType: "person_record",
		SourceEntityID:   recordID,
		Status:           LeadStatusNew,
		OutreachStage:    OutreachDraft,
		Notes:            []LeadNote{},
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	lead.AddNote("system", fmt.Sprintf("Imported lead from Person Record (%s at %s)", p.FullName, companyName))

	if err := s.repo.SaveRecruiterLead(ctx, lead); err != nil {
		return nil, err
	}
	return s.repo.GetRecruiterLead(ctx, lead.ID, tenantID)
}

func (s *Service) CreateLeadFromPostRecord(ctx context.Context, recordID, workspaceID, tenantID string) (*RecruiterLead, error) {
	post, err := s.repo.GetPostRecord(ctx, recordID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve source post record: %w", err)
	}

	authorURL := fmt.Sprintf("https://www.linkedin.com/in/%s", post.AuthorVanity)
	if post.AuthorVanity == "" {
		authorURL = "https://www.linkedin.com/feed/"
	}
	postURL := fmt.Sprintf("https://www.linkedin.com/feed/update/%s", post.URN)

	lead := &RecruiterLead{
		ID:               fmt.Sprintf("lead-post-%s", recordID),
		WorkspaceID:      workspaceID,
		TenantID:         tenantID,
		RecruiterName:    post.AuthorName,
		RecruiterTitle:   "Post Author / Hiring Lead",
		Company:          "LinkedIn Network",
		LinkedInURL:      authorURL,
		SourceEntityType: "post_record",
		SourceEntityID:   recordID,
		Status:           LeadStatusNew,
		OutreachStage:    OutreachDraft,
		Notes:            []LeadNote{},
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	lead.AddNote("system", fmt.Sprintf("Discovered lead from Post Record: %s", postURL))

	if err := s.repo.SaveRecruiterLead(ctx, lead); err != nil {
		return nil, err
	}
	return s.repo.GetRecruiterLead(ctx, lead.ID, tenantID)
}

func (s *Service) GetRecruiterLead(ctx context.Context, leadID, requestingTenantID string) (*RecruiterLead, error) {
	return s.repo.GetRecruiterLead(ctx, leadID, requestingTenantID)
}

func (s *Service) ListRecruiterLeads(ctx context.Context, filter RecruiterLeadFilter) ([]RecruiterLead, error) {
	return s.repo.ListRecruiterLeads(ctx, filter)
}

func (s *Service) AddLeadNote(ctx context.Context, leadID, requestingTenantID, author, content string) (*LeadNote, error) {
	lead, err := s.repo.GetRecruiterLead(ctx, leadID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	note, err := lead.AddNote(author, content)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveRecruiterLead(ctx, lead); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *Service) SetLeadReminder(ctx context.Context, leadID, requestingTenantID string, dueDate time.Time, message string) (*FollowUpReminder, error) {
	lead, err := s.repo.GetRecruiterLead(ctx, leadID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	rem, err := lead.SetReminder(dueDate, message)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveRecruiterLead(ctx, lead); err != nil {
		return nil, err
	}
	return rem, nil
}

func (s *Service) CompleteLeadReminder(ctx context.Context, leadID, requestingTenantID string) error {
	lead, err := s.repo.GetRecruiterLead(ctx, leadID, requestingTenantID)
	if err != nil {
		return err
	}
	if err := lead.CompleteReminder(); err != nil {
		return err
	}
	return s.repo.SaveRecruiterLead(ctx, lead)
}

func (s *Service) UpdateLeadStatus(ctx context.Context, leadID, requestingTenantID string, status LeadStatus, outreachStage LeadOutreachStage) (*RecruiterLead, error) {
	lead, err := s.repo.GetRecruiterLead(ctx, leadID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	if status != "" {
		lead.Status = status
	}
	if outreachStage != "" {
		lead.OutreachStage = outreachStage
	}
	lead.UpdatedAt = time.Now().UTC()
	if err := s.repo.SaveRecruiterLead(ctx, lead); err != nil {
		return nil, err
	}
	return s.repo.GetRecruiterLead(ctx, leadID, requestingTenantID)
}

func (s *Service) LinkLeadApplication(ctx context.Context, leadID, requestingTenantID, jobID, applicationID string) (*RecruiterLead, error) {
	lead, err := s.repo.GetRecruiterLead(ctx, leadID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	lead.LinkApplication(jobID, applicationID)
	if err := s.repo.SaveRecruiterLead(ctx, lead); err != nil {
		return nil, err
	}
	return s.repo.GetRecruiterLead(ctx, leadID, requestingTenantID)
}

// Connection Note Drafts & Queue Service (IMP-LI-07, LI-07, AT-007, AT-010, FND-010, FND-011, SRC-L1)

func (s *Service) DraftConnectionNote(candidateName, targetRole, recipientName, recipientCompany, recipientTitle, keyOverlap string) (string, []string) {
	return GenerateConnectionNoteDraft(candidateName, targetRole, recipientName, recipientCompany, recipientTitle, keyOverlap)
}

func (s *Service) EnqueueConnectionNote(ctx context.Context, item *ConnectionQueueItem) (*ConnectionQueueItem, error) {
	if item == nil {
		return nil, errors.New("item cannot be nil")
	}
	if item.ID == "" {
		item.ID = fmt.Sprintf("cq-%d", time.Now().UnixNano())
	}
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now
	item.Status = QueueItemPendingApproval

	if err := ValidateConnectionQueueItem(item); err != nil {
		return nil, err
	}

	// Recipient Deduplication (AT-007, AT-004)
	existing, err := s.repo.GetConnectionQueueItemByRecipientURL(ctx, item.WorkspaceID, item.RecipientLinkedInURL)
	if err == nil && existing != nil {
		if existing.Status == QueueItemPendingApproval ||
			existing.Status == QueueItemApproved ||
			existing.Status == QueueItemCopied ||
			existing.Status == QueueItemCompleted {
			return nil, ErrDuplicateRecipient
		}
	}

	if err := s.repo.SaveConnectionQueueItem(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetConnectionQueueItem(ctx, item.ID, item.TenantID)
}

func (s *Service) ApproveConnectionQueueItem(ctx context.Context, itemID, requestingTenantID, approver string) (*ConnectionQueueItem, error) {
	item, err := s.repo.GetConnectionQueueItem(ctx, itemID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	// Check safe daily/weekly invite limits before approving (FND-011, AT-008)
	budget, err := s.repo.GetConnectionBudget(ctx, item.WorkspaceID, item.TenantID)
	if err != nil {
		return nil, err
	}
	if err := budget.CanConsume(); err != nil {
		return nil, err
	}

	// Generate cryptographic approval token (FND-010, AT-007)
	token := ComputeApprovalToken("", item.ID, item.NoteText)
	now := time.Now().UTC()
	item.Status = QueueItemApproved
	item.ApprovalToken = token
	item.ApprovedBy = approver
	item.ApprovedAt = &now
	item.UpdatedAt = now

	if err := s.repo.SaveConnectionQueueItem(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetConnectionQueueItem(ctx, itemID, requestingTenantID)
}

func (s *Service) UpdateConnectionNoteText(ctx context.Context, itemID, requestingTenantID, newText string) (*ConnectionQueueItem, error) {
	item, err := s.repo.GetConnectionQueueItem(ctx, itemID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	trimmedText := strings.TrimSpace(newText)
	if trimmedText == "" {
		return nil, ErrEmptyNoteText
	}

	// Tamper Invalidation (FND-010, AT-007): Invalidate prior approval if payload changed
	if trimmedText != strings.TrimSpace(item.NoteText) && item.Status == QueueItemApproved {
		item.Status = QueueItemPendingApproval
		item.ApprovalToken = ""
		item.ApprovedBy = ""
		item.ApprovedAt = nil
	}

	item.NoteText = trimmedText
	item.CharacterCount = len([]rune(trimmedText))
	item.WithinLimit = item.CharacterCount <= LinkedInMaxNoteCharacters
	if !item.WithinLimit {
		return nil, ErrNoteExceedsLimit
	}
	item.UpdatedAt = time.Now().UTC()

	if err := s.repo.SaveConnectionQueueItem(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetConnectionQueueItem(ctx, itemID, requestingTenantID)
}

func (s *Service) MarkConnectionNoteCopied(ctx context.Context, itemID, requestingTenantID string) (*ConnectionQueueItem, error) {
	item, err := s.repo.GetConnectionQueueItem(ctx, itemID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	if item.Status != QueueItemApproved && item.Status != QueueItemCopied {
		return nil, ErrNotApproved
	}

	// Verify cryptographic token integrity
	if !VerifyApprovalToken("", item.ID, item.NoteText, item.ApprovalToken) {
		item.Status = QueueItemPendingApproval
		item.ApprovalToken = ""
		_ = s.repo.SaveConnectionQueueItem(ctx, item)
		return nil, ErrInvalidApprovalToken
	}

	item.Status = QueueItemCopied
	item.UpdatedAt = time.Now().UTC()

	if err := s.repo.SaveConnectionQueueItem(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetConnectionQueueItem(ctx, itemID, requestingTenantID)
}

func (s *Service) ConfirmConnectionSent(ctx context.Context, itemID, requestingTenantID string) (*ConnectionQueueItem, error) {
	item, err := s.repo.GetConnectionQueueItem(ctx, itemID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	if item.Status != QueueItemApproved && item.Status != QueueItemCopied {
		return nil, ErrNotApproved
	}

	if !VerifyApprovalToken("", item.ID, item.NoteText, item.ApprovalToken) {
		return nil, ErrInvalidApprovalToken
	}

	// Consume 1 unit from rolling budget upon confirmed send
	budget, err := s.repo.GetConnectionBudget(ctx, item.WorkspaceID, item.TenantID)
	if err != nil {
		return nil, err
	}
	if err := budget.Consume(); err != nil {
		return nil, err
	}
	if err := s.repo.SaveConnectionBudget(ctx, budget); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	item.Status = QueueItemCompleted
	item.CompletedAt = &now
	item.UpdatedAt = now

	if err := s.repo.SaveConnectionQueueItem(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetConnectionQueueItem(ctx, itemID, requestingTenantID)
}

func (s *Service) RejectConnectionQueueItem(ctx context.Context, itemID, requestingTenantID, reason string) (*ConnectionQueueItem, error) {
	item, err := s.repo.GetConnectionQueueItem(ctx, itemID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	item.Status = QueueItemRejected
	item.UpdatedAt = time.Now().UTC()

	if err := s.repo.SaveConnectionQueueItem(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetConnectionQueueItem(ctx, itemID, requestingTenantID)
}

func (s *Service) GetConnectionQueueItem(ctx context.Context, itemID, requestingTenantID string) (*ConnectionQueueItem, error) {
	return s.repo.GetConnectionQueueItem(ctx, itemID, requestingTenantID)
}

func (s *Service) ListConnectionQueueItems(ctx context.Context, filter ConnectionQueueFilter) ([]ConnectionQueueItem, error) {
	return s.repo.ListConnectionQueueItems(ctx, filter)
}

func (s *Service) GetConnectionBudget(ctx context.Context, workspaceID, tenantID string) (*ConnectionBudget, error) {
	return s.repo.GetConnectionBudget(ctx, workspaceID, tenantID)
}

func (s *Service) SaveConnectionBudget(ctx context.Context, budget *ConnectionBudget) error {
	return s.repo.SaveConnectionBudget(ctx, budget)
}

// Company-Follow Planning Service Methods (IMP-LI-08, AT-010, FND-011, SRC-L1)

func (s *Service) AddCompanyToWatchlist(ctx context.Context, item *CompanyWatchlistItem) (*CompanyWatchlistItem, error) {
	if item == nil {
		return nil, errors.New("company_follow: watchlist item is nil")
	}
	if strings.TrimSpace(item.WorkspaceID) == "" || strings.TrimSpace(item.TenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}

	// Deduplication Guardrail (AT-004, AT-007)
	cleanUniversal := strings.TrimSpace(item.UniversalName)
	if cleanUniversal != "" {
		existing, err := s.repo.GetWatchlistItemByUniversalName(ctx, item.WorkspaceID, item.TenantID, cleanUniversal)
		if err == nil && existing != nil && existing.Status != WatchlistArchived {
			return nil, ErrWatchlistCompanyDuplicate
		}
	}

	if item.ItemID == "" {
		item.ItemID = fmt.Sprintf("cwl-%d", time.Now().UTC().UnixNano())
	}
	if item.CompanyPageURL == "" && cleanUniversal != "" {
		item.CompanyPageURL = BuildCompanyPageURL(cleanUniversal)
	}

	if err := s.repo.SaveWatchlistItem(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetWatchlistItem(ctx, item.ItemID, item.TenantID)
}

func (s *Service) GetCompanyWatchlist(ctx context.Context, workspaceID, requestingTenantID string, status WatchlistStatus) ([]CompanyWatchlistItem, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListWatchlistItems(ctx, workspaceID, requestingTenantID, status)
}

func (s *Service) DeleteWatchlistItem(ctx context.Context, itemID, requestingTenantID string) error {
	return s.repo.DeleteWatchlistItem(ctx, itemID, requestingTenantID)
}

func (s *Service) CreateBatchFollowPlan(ctx context.Context, workspaceID, tenantID, planName string, itemIDs []string, pacingIntervalSec int) (*BatchCompanyFollowPlan, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(tenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	if len(itemIDs) == 0 {
		return nil, ErrInvalidBatchPlan
	}

	watchlist := make([]*CompanyWatchlistItem, 0, len(itemIDs))
	for _, id := range itemIDs {
		item, err := s.repo.GetWatchlistItem(ctx, id, tenantID)
		if err != nil {
			return nil, fmt.Errorf("company_follow: item %s not found: %w", id, err)
		}
		if item.WorkspaceID != workspaceID {
			return nil, ErrCrossTenantAccessDenied
		}
		watchlist = append(watchlist, item)
	}

	plan, err := GenerateBatchFollowPlan(workspaceID, tenantID, planName, watchlist, pacingIntervalSec, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveBatchFollowPlan(ctx, plan); err != nil {
		return nil, err
	}
	return s.repo.GetBatchFollowPlan(ctx, plan.PlanID, tenantID)
}

func (s *Service) GetBatchFollowPlan(ctx context.Context, planID, requestingTenantID string) (*BatchCompanyFollowPlan, error) {
	return s.repo.GetBatchFollowPlan(ctx, planID, requestingTenantID)
}

func (s *Service) ListBatchFollowPlans(ctx context.Context, workspaceID, requestingTenantID string) ([]BatchCompanyFollowPlan, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListBatchFollowPlans(ctx, workspaceID, requestingTenantID)
}

func (s *Service) ConfirmCompanyFollowed(ctx context.Context, planID, itemID, requestingTenantID string) (*BatchCompanyFollowPlan, error) {
	plan, err := s.repo.GetBatchFollowPlan(ctx, planID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	// Conservative safety budget consumption (FND-011, REQ-010)
	budget, err := s.repo.GetCompanyFollowBudget(ctx, plan.WorkspaceID, plan.TenantID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := ConsumeFollowBudget(budget, now); err != nil {
		// Save locked status if modified
		_ = s.repo.SaveCompanyFollowBudget(ctx, budget)
		return nil, err
	}
	if err := s.repo.SaveCompanyFollowBudget(ctx, budget); err != nil {
		return nil, err
	}

	found := false
	allProcessed := true
	for i := range plan.Items {
		if plan.Items[i].ItemID == itemID {
			found = true
			plan.Items[i].Status = PlanItemFollowed
			plan.Items[i].FollowedAt = &now
			plan.FollowedCount++
		}
		if plan.Items[i].Status == PlanItemReady || plan.Items[i].Status == PlanItemPlanned {
			allProcessed = false
		}
	}

	if !found {
		return nil, errors.New("company_follow: plan item not found")
	}

	if allProcessed {
		plan.Status = BatchPlanCompleted
	} else {
		plan.Status = BatchPlanInProgress
	}

	if err := s.repo.SaveBatchFollowPlan(ctx, plan); err != nil {
		return nil, err
	}
	return s.repo.GetBatchFollowPlan(ctx, planID, requestingTenantID)
}

func (s *Service) SkipCompanyFollow(ctx context.Context, planID, itemID, requestingTenantID, reason string) (*BatchCompanyFollowPlan, error) {
	plan, err := s.repo.GetBatchFollowPlan(ctx, planID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	found := false
	allProcessed := true
	for i := range plan.Items {
		if plan.Items[i].ItemID == itemID {
			found = true
			plan.Items[i].Status = PlanItemSkipped
			if reason != "" {
				plan.Items[i].Notes = reason
			}
		}
		if plan.Items[i].Status == PlanItemReady || plan.Items[i].Status == PlanItemPlanned {
			allProcessed = false
		}
	}

	if !found {
		return nil, errors.New("company_follow: plan item not found")
	}

	if allProcessed {
		plan.Status = BatchPlanCompleted
	} else {
		plan.Status = BatchPlanInProgress
	}

	if err := s.repo.SaveBatchFollowPlan(ctx, plan); err != nil {
		return nil, err
	}
	return s.repo.GetBatchFollowPlan(ctx, planID, requestingTenantID)
}

func (s *Service) GetCompanyFollowBudget(ctx context.Context, workspaceID, tenantID string) (*CompanyFollowBudget, error) {
	return s.repo.GetCompanyFollowBudget(ctx, workspaceID, tenantID)
}

func (s *Service) UpdateCompanyFollowBudget(ctx context.Context, budget *CompanyFollowBudget) error {
	return s.repo.SaveCompanyFollowBudget(ctx, budget)
}

// Inbox & Conversation Triage Service Methods (IMP-LI-09, LI-09, AT-007, AT-010, AT-011, FND-010, SRC-C3)

func (s *Service) SaveConversationThread(ctx context.Context, thread *ConversationThread) (*ConversationThread, error) {
	if thread == nil {
		return nil, ErrThreadNotFound
	}
	if thread.ThreadID == "" {
		thread.ThreadID = fmt.Sprintf("th-%d", time.Now().UnixNano())
	}
	if err := s.repo.SaveConversationThread(ctx, thread); err != nil {
		return nil, err
	}
	return s.repo.GetConversationThread(ctx, thread.ThreadID, thread.TenantID)
}

func (s *Service) GetConversationThread(ctx context.Context, threadID, requestingTenantID string) (*ConversationThread, error) {
	return s.repo.GetConversationThread(ctx, threadID, requestingTenantID)
}

func (s *Service) ListConversationThreads(ctx context.Context, filter ConversationThreadFilter) ([]ConversationThread, error) {
	if strings.TrimSpace(filter.WorkspaceID) == "" || strings.TrimSpace(filter.TenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListConversationThreads(ctx, filter)
}

func (s *Service) AddMessageToThread(ctx context.Context, threadID, requestingTenantID string, msg InboxMessage) (*ConversationThread, bool, error) {
	thread, err := s.repo.GetConversationThread(ctx, threadID, requestingTenantID)
	if err != nil {
		return nil, false, err
	}
	if strings.TrimSpace(msg.Content) == "" {
		return nil, false, ErrInvalidMessagePayload
	}
	if msg.MessageID == "" {
		msg.MessageID = fmt.Sprintf("msg-%d", time.Now().UnixNano())
	}
	if msg.ThreadID == "" {
		msg.ThreadID = threadID
	}
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now().UTC()
	}

	halted := ProcessInboundMessage(thread, msg)

	if err := s.repo.SaveConversationThread(ctx, thread); err != nil {
		return nil, false, err
	}
	updated, err := s.repo.GetConversationThread(ctx, threadID, requestingTenantID)
	return updated, halted, err
}

func (s *Service) GenerateReplyDraft(ctx context.Context, threadID, requestingTenantID string, tone ReplyDraftTone, candidateAvailability string, verifiedFacts []string) (*ReplyDraft, error) {
	thread, err := s.repo.GetConversationThread(ctx, threadID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	draft, err := GenerateContextualReplyDraft(thread, tone, candidateAvailability, verifiedFacts)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveReplyDraft(ctx, draft); err != nil {
		return nil, err
	}
	return s.repo.GetReplyDraft(ctx, draft.DraftID, requestingTenantID)
}

func (s *Service) ApproveReplyDraft(ctx context.Context, draftID, requestingTenantID, approver string) (*ReplyDraft, error) {
	draft, err := s.repo.GetReplyDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(draft.SuggestedText) == "" {
		return nil, ErrEmptyReplyText
	}

	token := ComputeReplyApprovalToken("", draft.DraftID, draft.SuggestedText)
	now := time.Now().UTC()
	draft.Status = ReplyDraftApproved
	draft.ApprovalToken = token
	draft.ApprovedBy = approver
	draft.ApprovedAt = &now
	draft.UpdatedAt = now

	if err := s.repo.SaveReplyDraft(ctx, draft); err != nil {
		return nil, err
	}
	return s.repo.GetReplyDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) UpdateReplyDraftText(ctx context.Context, draftID, requestingTenantID, newText string) (*ReplyDraft, error) {
	draft, err := s.repo.GetReplyDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	cleanText := strings.TrimSpace(newText)
	if cleanText == "" {
		return nil, ErrEmptyReplyText
	}

	draft.SuggestedText = cleanText
	draft.CharacterCount = len([]rune(cleanText))
	draft.UpdatedAt = time.Now().UTC()

	// Tamper invalidation: If previously approved, any text change clears approval token and resets to draft (FND-010, AT-007)
	if draft.Status == ReplyDraftApproved || draft.ApprovalToken != "" {
		draft.Status = ReplyDraftPendingApproval
		draft.ApprovalToken = ""
		draft.ApprovedBy = ""
		draft.ApprovedAt = nil
	}

	if err := s.repo.SaveReplyDraft(ctx, draft); err != nil {
		return nil, err
	}
	return s.repo.GetReplyDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) MarkReplyDraftCopied(ctx context.Context, draftID, requestingTenantID string) (*ReplyDraft, error) {
	draft, err := s.repo.GetReplyDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	if draft.Status != ReplyDraftApproved {
		return nil, errors.New("inbox: reply draft must be approved prior to copy")
	}
	if !VerifyReplyApprovalToken("", draft.DraftID, draft.SuggestedText, draft.ApprovalToken) {
		return nil, ErrInvalidReplyApprovalToken
	}

	draft.Status = ReplyDraftCopiedToClipboard
	draft.UpdatedAt = time.Now().UTC()

	if err := s.repo.SaveReplyDraft(ctx, draft); err != nil {
		return nil, err
	}
	return s.repo.GetReplyDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) RejectReplyDraft(ctx context.Context, draftID, requestingTenantID, reason string) (*ReplyDraft, error) {
	draft, err := s.repo.GetReplyDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	draft.Status = ReplyDraftRejected
	if reason != "" {
		draft.Rationale = fmt.Sprintf("Rejected: %s (Previous: %s)", reason, draft.Rationale)
	}
	draft.UpdatedAt = time.Now().UTC()

	if err := s.repo.SaveReplyDraft(ctx, draft); err != nil {
		return nil, err
	}
	return s.repo.GetReplyDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) GetReplyDraft(ctx context.Context, draftID, requestingTenantID string) (*ReplyDraft, error) {
	return s.repo.GetReplyDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) ListReplyDraftsForThread(ctx context.Context, threadID, requestingTenantID string) ([]ReplyDraft, error) {
	if strings.TrimSpace(threadID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListReplyDraftsForThread(ctx, threadID, requestingTenantID)
}

// SweptTargetPost and CommentDraft Service methods (IMP-LI-10, LI-10, AT-007, AT-010, AT-011, SRC-L1, SRC-L2)

func (s *Service) SaveSweptTargetPost(ctx context.Context, post *SweptTargetPost) error {
	if post == nil {
		return errors.New("post cannot be nil")
	}
	if strings.TrimSpace(post.PostID) == "" || strings.TrimSpace(post.TenantID) == "" || strings.TrimSpace(post.WorkspaceID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	if post.Category == "" {
		post.Category = CategorizePostContent(post.Content)
	}
	if post.SweptAt.IsZero() {
		post.SweptAt = time.Now().UTC()
	}
	return s.repo.SaveSweptTargetPost(ctx, post)
}

func (s *Service) GetSweptTargetPost(ctx context.Context, postID, requestingTenantID string) (*SweptTargetPost, error) {
	if strings.TrimSpace(postID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.GetSweptTargetPost(ctx, postID, requestingTenantID)
}

func (s *Service) ListSweptTargetPosts(ctx context.Context, filter SweptPostFilter) ([]SweptTargetPost, error) {
	if strings.TrimSpace(filter.WorkspaceID) == "" || strings.TrimSpace(filter.TenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListSweptTargetPosts(ctx, filter)
}

func (s *Service) DeleteSweptTargetPost(ctx context.Context, postID, requestingTenantID string) error {
	if strings.TrimSpace(postID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return s.repo.DeleteSweptTargetPost(ctx, postID, requestingTenantID)
}

func (s *Service) GenerateCommentDrafts(ctx context.Context, postID, requestingTenantID, mode string, candidateFacts []string) ([]CommentDraft, error) {
	if strings.TrimSpace(postID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	// Fail-closed against automated bulk spamming / engagement pods (AT-010, REQ-015, LI-10)
	if err := RejectUnsupportedBulkEngagement(mode); err != nil {
		return nil, err
	}

	post, err := s.repo.GetSweptTargetPost(ctx, postID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	drafts, err := GenerateCommentDrafts(post, candidateFacts)
	if err != nil {
		return nil, err
	}

	for i := range drafts {
		if err := s.repo.SaveCommentDraft(ctx, &drafts[i]); err != nil {
			return nil, err
		}
	}
	return drafts, nil
}

func (s *Service) ApproveCommentDraft(ctx context.Context, draftID, requestingTenantID, approvedBy string) (*CommentDraft, error) {
	if strings.TrimSpace(draftID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	draft, err := s.repo.GetCommentDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	if err := ValidateCommentText(draft.CommentText); err != nil {
		return nil, err
	}

	token := ComputeCommentApprovalToken("", draft.DraftID, draft.CommentText)
	now := time.Now().UTC()
	draft.Status = CommentDraftApproved
	draft.ApprovalToken = token
	draft.ApprovedBy = approvedBy
	draft.ApprovedAt = &now
	draft.UpdatedAt = now

	if err := s.repo.SaveCommentDraft(ctx, draft); err != nil {
		return nil, err
	}
	return s.repo.GetCommentDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) UpdateCommentDraftText(ctx context.Context, draftID, requestingTenantID, newText string) (*CommentDraft, error) {
	if strings.TrimSpace(draftID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	draft, err := s.repo.GetCommentDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	cleanText := strings.TrimSpace(newText)
	if err := ValidateCommentText(cleanText); err != nil {
		return nil, err
	}

	draft.CommentText = cleanText
	draft.CharacterCount = len([]rune(cleanText))
	// Tamper invalidation: any text modification invalidates approval (FND-010, AT-007)
	draft.Status = CommentDraftPendingApproval
	draft.ApprovalToken = ""
	draft.ApprovedBy = ""
	draft.ApprovedAt = nil
	draft.UpdatedAt = time.Now().UTC()

	if err := s.repo.SaveCommentDraft(ctx, draft); err != nil {
		return nil, err
	}
	return s.repo.GetCommentDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) MarkCommentDraftCopied(ctx context.Context, draftID, requestingTenantID string) (*CommentDraft, error) {
	if strings.TrimSpace(draftID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	draft, err := s.repo.GetCommentDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	if draft.Status != CommentDraftApproved {
		return nil, errors.New("comments: comment draft must be approved prior to clipboard copy")
	}
	if !VerifyCommentApprovalToken("", draft.DraftID, draft.CommentText, draft.ApprovalToken) {
		return nil, ErrInvalidCommentApprovalToken
	}

	draft.Status = CommentDraftCopiedToClipboard
	draft.UpdatedAt = time.Now().UTC()

	if err := s.repo.SaveCommentDraft(ctx, draft); err != nil {
		return nil, err
	}
	return s.repo.GetCommentDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) RejectCommentDraft(ctx context.Context, draftID, requestingTenantID, reason string) (*CommentDraft, error) {
	if strings.TrimSpace(draftID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	draft, err := s.repo.GetCommentDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	draft.Status = CommentDraftRejected
	if reason != "" {
		draft.Rationale = fmt.Sprintf("Rejected: %s (Previous: %s)", reason, draft.Rationale)
	}
	draft.UpdatedAt = time.Now().UTC()

	if err := s.repo.SaveCommentDraft(ctx, draft); err != nil {
		return nil, err
	}
	return s.repo.GetCommentDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) GetCommentDraft(ctx context.Context, draftID, requestingTenantID string) (*CommentDraft, error) {
	if strings.TrimSpace(draftID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.GetCommentDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) ListCommentDraftsForPost(ctx context.Context, postID, requestingTenantID string) ([]CommentDraft, error) {
	if strings.TrimSpace(postID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListCommentDraftsForPost(ctx, postID, requestingTenantID)
}

func errorsNew(msg string) error {
	return errors.New(msg)
}

// Post Writing, Hooks and Audits (IMP-LI-11, LI-11, AT-003, AT-019, FND-015, SRC-L2)

func (s *Service) GeneratePostDraft(ctx context.Context, req PostWritingRequest) (*PostDraft, error) {
	if strings.TrimSpace(req.WorkspaceID) == "" || strings.TrimSpace(req.TenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	draft, err := GeneratePostDraft(req)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SavePostDraft(ctx, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

func (s *Service) SavePostDraft(ctx context.Context, draft *PostDraft) error {
	if draft == nil || strings.TrimSpace(draft.WorkspaceID) == "" || strings.TrimSpace(draft.TenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return s.repo.SavePostDraft(ctx, draft)
}

func (s *Service) GetPostDraft(ctx context.Context, draftID, requestingTenantID string) (*PostDraft, error) {
	if strings.TrimSpace(draftID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.GetPostDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) ListPostDrafts(ctx context.Context, filter PostDraftFilter) ([]PostDraft, error) {
	if strings.TrimSpace(filter.WorkspaceID) == "" || strings.TrimSpace(filter.TenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListPostDrafts(ctx, filter)
}

func (s *Service) DeletePostDraft(ctx context.Context, draftID, requestingTenantID string) error {
	if strings.TrimSpace(draftID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return s.repo.DeletePostDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) GenerateHookVariantsForDraft(ctx context.Context, draftID, requestingTenantID string) ([]HookVariant, error) {
	draft, err := s.GetPostDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	hooks := GenerateHookVariants(draft.Topic, draft.Angle, draft.VerifiedFactsUsed)
	draft.AvailableHooks = hooks
	draft.UpdatedAt = time.Now().UTC()
	if err := s.repo.SavePostDraft(ctx, draft); err != nil {
		return nil, err
	}
	return hooks, nil
}

func (s *Service) AuditPostDraftContent(ctx context.Context, draftID, requestingTenantID string) (*EditorialAuditReport, error) {
	draft, err := s.GetPostDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	audit := AuditPostDraft(*draft, draft.VerifiedFactsUsed)
	draft.LatestAudit = &audit
	draft.UpdatedAt = time.Now().UTC()
	if err := s.repo.SavePostDraft(ctx, draft); err != nil {
		return nil, err
	}
	return &audit, nil
}

func (s *Service) ApprovePostDraft(ctx context.Context, draftID, requestingTenantID, approvedBy string) (*PostDraft, error) {
	draft, err := s.GetPostDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	if err := ValidatePostText(draft.FullPostText); err != nil {
		return nil, err
	}
	audit := AuditPostDraft(*draft, draft.VerifiedFactsUsed)
	draft.LatestAudit = &audit
	if !audit.Passed {
		return nil, fmt.Errorf("posts: cannot approve draft failing editorial audit (blocking issues: %d)", len(audit.Issues))
	}

	token := ComputePostApprovalToken(*draft, "")
	now := time.Now().UTC()
	draft.Status = PostDraftApproved
	draft.ApprovalToken = token
	draft.ApprovedBy = approvedBy
	draft.ApprovedAt = &now
	draft.UpdatedAt = now

	if err := s.repo.SavePostDraft(ctx, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

func (s *Service) UpdatePostDraftText(ctx context.Context, draftID, requestingTenantID, selectedHookType, selectedHookText, fullPostText string) (*PostDraft, error) {
	draft, err := s.GetPostDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	cleanText := strings.TrimSpace(fullPostText)
	if err := ValidatePostText(cleanText); err != nil {
		return nil, err
	}

	draft.FullPostText = cleanText
	if selectedHookText != "" {
		draft.SelectedHookText = selectedHookText
	}
	if selectedHookType != "" {
		draft.SelectedHookType = HookType(selectedHookType)
	}

	// Re-audit with updated text
	audit := AuditPostDraft(*draft, draft.VerifiedFactsUsed)
	draft.LatestAudit = &audit

	// Tamper invalidation: any text edit invalidates approval (FND-010, AT-007)
	draft.Status = PostDraftPendingApproval
	draft.ApprovalToken = ""
	draft.ApprovedBy = ""
	draft.ApprovedAt = nil
	draft.UpdatedAt = time.Now().UTC()

	if err := s.repo.SavePostDraft(ctx, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

func (s *Service) MarkPostDraftCopied(ctx context.Context, draftID, requestingTenantID string) (*PostDraft, error) {
	draft, err := s.GetPostDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	if draft.Status != PostDraftApproved {
		return nil, errors.New("posts: post draft must be approved prior to clipboard copy")
	}
	if !VerifyPostApprovalToken(*draft, draft.ApprovalToken, "") {
		return nil, ErrInvalidPostApprovalToken
	}

	draft.Status = PostDraftCopiedToClipboard
	draft.UpdatedAt = time.Now().UTC()

	if err := s.repo.SavePostDraft(ctx, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

func (s *Service) RejectPostDraft(ctx context.Context, draftID, requestingTenantID, reason string) (*PostDraft, error) {
	draft, err := s.GetPostDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	draft.Status = PostDraftRejected
	draft.UpdatedAt = time.Now().UTC()

	if err := s.repo.SavePostDraft(ctx, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

// -------------------------------------------------------------------------
// Humanizer and Reusable Voice (IMP-LI-12, LI-12, AT-003, FND-015, SRC-L2, SRC-S4)
// -------------------------------------------------------------------------

func (s *Service) CreateVoiceProfile(ctx context.Context, profile *VoiceProfile) (*VoiceProfile, error) {
	if profile == nil {
		return nil, errors.New("profile cannot be nil")
	}
	if profile.ProfileID == "" {
		profile.ProfileID = fmt.Sprintf("vp-%d", time.Now().UnixNano())
	}
	if profile.Status == "" {
		profile.Status = VoiceProfileDraft
	}
	if profile.CadenceStyle == "" {
		profile.CadenceStyle = CadenceBalancedRhythm
	}
	if profile.Perspective == "" {
		profile.Perspective = PerspectiveFirstPersonSingular
	}
	if err := s.repo.SaveVoiceProfile(ctx, profile); err != nil {
		return nil, err
	}
	return s.repo.GetVoiceProfile(ctx, profile.ProfileID, profile.TenantID)
}

func (s *Service) UpdateVoiceProfile(ctx context.Context, profile *VoiceProfile, requestingTenantID string) (*VoiceProfile, error) {
	if profile == nil {
		return nil, errors.New("profile cannot be nil")
	}
	existing, err := s.repo.GetVoiceProfile(ctx, profile.ProfileID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	// If modified after approval, invalidate approval token (AT-007)
	if existing.Status == VoiceProfileApproved {
		profile.Status = VoiceProfileDraft
		profile.ApprovalToken = ""
		profile.ApprovedBy = ""
		profile.ApprovedAt = ""
	}
	if err := s.repo.SaveVoiceProfile(ctx, profile); err != nil {
		return nil, err
	}
	return s.repo.GetVoiceProfile(ctx, profile.ProfileID, requestingTenantID)
}

func (s *Service) GetVoiceProfile(ctx context.Context, profileID, requestingTenantID string) (*VoiceProfile, error) {
	return s.repo.GetVoiceProfile(ctx, profileID, requestingTenantID)
}

func (s *Service) ListVoiceProfiles(ctx context.Context, workspaceID, requestingTenantID string) ([]VoiceProfile, error) {
	return s.repo.ListVoiceProfiles(ctx, workspaceID, requestingTenantID)
}

func (s *Service) DeleteVoiceProfile(ctx context.Context, profileID, requestingTenantID string) error {
	return s.repo.DeleteVoiceProfile(ctx, profileID, requestingTenantID)
}

func (s *Service) ApproveVoiceProfile(ctx context.Context, profileID, approvedBy, requestingTenantID string) (*VoiceProfile, error) {
	profile, err := s.repo.GetVoiceProfile(ctx, profileID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	token, err := ComputeVoiceApprovalToken(profile, "")
	if err != nil {
		return nil, err
	}
	profile.Status = VoiceProfileApproved
	profile.ApprovalToken = token
	profile.ApprovedBy = approvedBy
	profile.ApprovedAt = time.Now().UTC().Format(time.RFC3339)

	if err := s.repo.SaveVoiceProfile(ctx, profile); err != nil {
		return nil, err
	}
	return s.repo.GetVoiceProfile(ctx, profileID, requestingTenantID)
}

func (s *Service) HumanizeDraftContent(ctx context.Context, text, voiceProfileID, workspaceID, tenantID string, candidateFacts []string) (*HumanizeResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("text cannot be empty")
	}

	var profile *VoiceProfile
	if voiceProfileID != "" {
		p, err := s.repo.GetVoiceProfile(ctx, voiceProfileID, tenantID)
		if err == nil {
			profile = p
		}
	}

	result := HumanizeText(text, profile, candidateFacts)
	result.WorkspaceID = workspaceID
	result.TenantID = tenantID
	result.VoiceProfileID = voiceProfileID

	if err := s.repo.SaveHumanizeResult(ctx, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) ApproveHumanizedDraft(ctx context.Context, resultID, approvedBy, requestingTenantID string) (*HumanizeResult, error) {
	result, err := s.repo.GetHumanizeResult(ctx, resultID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	if !result.Passed {
		return nil, errors.New("humanizer: cannot approve draft with blocking audit issues (AT-003)")
	}

	token, err := ComputeHumanizeApprovalToken(result, "")
	if err != nil {
		return nil, err
	}
	result.Status = "approved"
	result.ApprovalToken = token
	result.ApprovedBy = approvedBy
	result.ApprovedAt = time.Now().UTC().Format(time.RFC3339)

	if err := s.repo.SaveHumanizeResult(ctx, result); err != nil {
		return nil, err
	}
	return s.repo.GetHumanizeResult(ctx, resultID, requestingTenantID)
}

func (s *Service) UpdateHumanizedText(ctx context.Context, resultID, newText, requestingTenantID string, candidateFacts []string) (*HumanizeResult, error) {
	result, err := s.repo.GetHumanizeResult(ctx, resultID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	var profile *VoiceProfile
	if result.VoiceProfileID != "" {
		p, err := s.repo.GetVoiceProfile(ctx, result.VoiceProfileID, requestingTenantID)
		if err == nil {
			profile = p
		}
	}

	updated := HumanizeText(newText, profile, candidateFacts)
	result.CleanedText = updated.CleanedText
	result.Tier1SlopReplacements = updated.Tier1SlopReplacements
	result.Tier2CadenceNotes = updated.Tier2CadenceNotes
	result.Tier3Issues = updated.Tier3Issues
	result.SlopScoreBefore = updated.SlopScoreBefore
	result.SlopScoreAfter = updated.SlopScoreAfter
	result.ReadabilityScoreBefore = updated.ReadabilityScoreBefore
	result.ReadabilityScoreAfter = updated.ReadabilityScoreAfter
	result.Passed = updated.Passed

	// Tamper Invalidation (AT-007)
	if result.Status == "approved" {
		result.Status = "draft"
		result.ApprovalToken = ""
		result.ApprovedBy = ""
		result.ApprovedAt = ""
	}

	if err := s.repo.SaveHumanizeResult(ctx, result); err != nil {
		return nil, err
	}
	return s.repo.GetHumanizeResult(ctx, resultID, requestingTenantID)
}

func (s *Service) MarkHumanizedDraftCopied(ctx context.Context, resultID, requestingTenantID string) (*HumanizeResult, error) {
	result, err := s.repo.GetHumanizeResult(ctx, resultID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	if result.Status != "approved" {
		return nil, errors.New("humanizer: draft must be approved prior to clipboard copy")
	}
	if !VerifyHumanizeApprovalToken(result, result.ApprovalToken, "") {
		return nil, errors.New("humanizer: invalid cryptographic approval token (tamper detected)")
	}

	result.Status = "copied_to_clipboard"
	if err := s.repo.SaveHumanizeResult(ctx, result); err != nil {
		return nil, err
	}
	return s.repo.GetHumanizeResult(ctx, resultID, requestingTenantID)
}

// -------------------------------------------------------------------------
// Story Bank and Guided Interviewer (IMP-LI-13, LI-13, AT-003, SRC-L2)
// -------------------------------------------------------------------------

func (s *Service) GetCuratedInterviewPrompts(ctx context.Context) ([]InterviewPrompt, error) {
	return GetCuratedInterviewPrompts(), nil
}

func (s *Service) StartInterviewSession(ctx context.Context, workspaceID, requestingTenantID string, category StoryCategory) (*InterviewSession, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}

	session := &InterviewSession{
		ID:                  fmt.Sprintf("sess-interview-%d", time.Now().UnixNano()),
		WorkspaceID:         workspaceID,
		TenantID:            requestingTenantID,
		Category:            category,
		Status:              "in_progress",
		QuestionsAndAnswers: make([]InterviewQA, 0),
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	if err := s.repo.SaveInterviewSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Service) RecordInterviewAnswer(ctx context.Context, sessionID, requestingTenantID, promptID, answer string) (*InterviewSession, error) {
	session, err := s.repo.GetInterviewSession(ctx, sessionID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	curated := GetCuratedPromptByID(promptID)
	questionText := "Interview Question"
	var followUpProbe string
	if curated != nil {
		questionText = curated.Question
		if len(curated.FollowUpProbes) > 0 {
			followUpProbe = curated.FollowUpProbes[0]
		}
	}

	found := false
	for i, qa := range session.QuestionsAndAnswers {
		if qa.PromptID == promptID {
			session.QuestionsAndAnswers[i].CandidateAnswer = answer
			if followUpProbe != "" {
				session.QuestionsAndAnswers[i].FollowUpProbe = followUpProbe
			}
			found = true
			break
		}
	}

	if !found {
		session.QuestionsAndAnswers = append(session.QuestionsAndAnswers, InterviewQA{
			PromptID:        promptID,
			Question:        questionText,
			CandidateAnswer: answer,
			FollowUpProbe:   followUpProbe,
		})
	}

	session.UpdatedAt = time.Now().UTC()
	if err := s.repo.SaveInterviewSession(ctx, session); err != nil {
		return nil, err
	}
	return s.repo.GetInterviewSession(ctx, sessionID, requestingTenantID)
}

func (s *Service) RecordInterviewFollowUpAnswer(ctx context.Context, sessionID, requestingTenantID, promptID, followUpAnswer string) (*InterviewSession, error) {
	session, err := s.repo.GetInterviewSession(ctx, sessionID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	for i, qa := range session.QuestionsAndAnswers {
		if qa.PromptID == promptID {
			session.QuestionsAndAnswers[i].CandidateFollowUpAnswer = followUpAnswer
			break
		}
	}

	session.UpdatedAt = time.Now().UTC()
	if err := s.repo.SaveInterviewSession(ctx, session); err != nil {
		return nil, err
	}
	return s.repo.GetInterviewSession(ctx, sessionID, requestingTenantID)
}

func (s *Service) SynthesizeInterviewStory(ctx context.Context, sessionID, requestingTenantID, authorName string, verifiedFacts []string) (*StoryEntry, error) {
	session, err := s.repo.GetInterviewSession(ctx, sessionID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	entry, err := SynthesizeStoryFromInterview(session, authorName, nil)
	if err != nil {
		return nil, err
	}

	audit := AuditStoryProvenanceAndMetrics(entry.Narrative, verifiedFacts)
	entry.AuditReport = &audit

	if err := s.repo.SaveStoryEntry(ctx, entry); err != nil {
		return nil, err
	}
	if err := s.repo.SaveInterviewSession(ctx, session); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *Service) CreateStoryEntry(ctx context.Context, entry *StoryEntry, verifiedFacts []string) (*StoryEntry, error) {
	if entry == nil {
		return nil, ErrInvalidStoryData
	}
	if strings.TrimSpace(entry.WorkspaceID) == "" || strings.TrimSpace(entry.TenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("story-%d", time.Now().UnixNano())
	}
	if entry.ApprovalStatus == "" {
		entry.ApprovalStatus = StoryStatusDraft
	}

	audit := AuditStoryProvenanceAndMetrics(entry.Narrative, verifiedFacts)
	entry.AuditReport = &audit

	if err := s.repo.SaveStoryEntry(ctx, entry); err != nil {
		return nil, err
	}
	return s.repo.GetStoryEntry(ctx, entry.ID, entry.TenantID)
}

func (s *Service) GetStoryEntry(ctx context.Context, entryID, requestingTenantID string) (*StoryEntry, error) {
	if strings.TrimSpace(entryID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.GetStoryEntry(ctx, entryID, requestingTenantID)
}

func (s *Service) ListStoryEntries(ctx context.Context, filter StoryFilter) ([]StoryEntry, error) {
	if strings.TrimSpace(filter.WorkspaceID) == "" || strings.TrimSpace(filter.RequestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListStoryEntries(ctx, filter)
}

func (s *Service) DeleteStoryEntry(ctx context.Context, entryID, requestingTenantID string) error {
	if strings.TrimSpace(entryID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return s.repo.DeleteStoryEntry(ctx, entryID, requestingTenantID)
}

func (s *Service) UpdateStoryEntry(ctx context.Context, entryID, requestingTenantID string, narrative StoryNarrative, verifiedFacts []string) (*StoryEntry, error) {
	entry, err := s.repo.GetStoryEntry(ctx, entryID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	entry.Narrative = narrative
	audit := AuditStoryProvenanceAndMetrics(narrative, verifiedFacts)
	entry.AuditReport = &audit

	// Tamper Invalidation (AT-007)
	if entry.ApprovalStatus == StoryStatusApproved {
		entry.ApprovalStatus = StoryStatusDraft
		entry.ApprovalToken = ""
		entry.LastApprovedAt = nil
	}

	if err := s.repo.SaveStoryEntry(ctx, entry); err != nil {
		return nil, err
	}
	return s.repo.GetStoryEntry(ctx, entryID, requestingTenantID)
}

func (s *Service) ApproveStoryEntry(ctx context.Context, entryID, requestingTenantID, secretKey string, verifiedFacts []string) (*StoryEntry, error) {
	entry, err := s.repo.GetStoryEntry(ctx, entryID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	audit := AuditStoryProvenanceAndMetrics(entry.Narrative, verifiedFacts)
	entry.AuditReport = &audit

	if audit.HasBlockingIssues {
		return nil, fmt.Errorf("story approval blocked: narrative contains %d unverified metric issue(s) under AT-003", len(audit.Issues))
	}

	token := ComputeStoryApprovalToken(secretKey, entry)
	now := time.Now().UTC()
	entry.ApprovalStatus = StoryStatusApproved
	entry.ApprovalToken = token
	entry.LastApprovedAt = &now

	if err := s.repo.SaveStoryEntry(ctx, entry); err != nil {
		return nil, err
	}
	return s.repo.GetStoryEntry(ctx, entryID, requestingTenantID)
}

func (s *Service) ArchiveStoryEntry(ctx context.Context, entryID, requestingTenantID string) (*StoryEntry, error) {
	entry, err := s.repo.GetStoryEntry(ctx, entryID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	entry.ApprovalStatus = StoryStatusArchived
	if err := s.repo.SaveStoryEntry(ctx, entry); err != nil {
		return nil, err
	}
	return s.repo.GetStoryEntry(ctx, entryID, requestingTenantID)
}

func (s *Service) MarkStoryEntryCopied(ctx context.Context, entryID, requestingTenantID, secretKey string) (*StoryEntry, string, error) {
	entry, err := s.repo.GetStoryEntry(ctx, entryID, requestingTenantID)
	if err != nil {
		return nil, "", err
	}

	if entry.ApprovalStatus != StoryStatusApproved {
		return nil, "", fmt.Errorf("story must be approved prior to clipboard export (AT-007)")
	}

	if !VerifyStoryApprovalToken(secretKey, entry, entry.ApprovalToken) {
		return nil, "", fmt.Errorf("invalid cryptographic approval token (tamper detected)")
	}

	formattedMarkdown := FormatStoryMarkdown(entry)
	return entry, formattedMarkdown, nil
}

// -------------------------------------------------------------------------
// Content Planning and Repurposing (IMP-LI-14, AT-018, AT-010, AT-007, FND-008, SRC-L2)
// -------------------------------------------------------------------------

func (s *Service) RepurposeArtifact(ctx context.Context, source SourceArtifact, format RepurposedFormat) (*RepurposedDraft, error) {
	if err := ValidateTenantScope(source.WorkspaceID, source.TenantID); err != nil {
		return nil, err
	}
	draft, err := RepurposeContentArtifact(source, format)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveRepurposedDraft(ctx, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

func (s *Service) RepurposeArtifactAllFormats(ctx context.Context, source SourceArtifact) ([]*RepurposedDraft, error) {
	if err := ValidateTenantScope(source.WorkspaceID, source.TenantID); err != nil {
		return nil, err
	}
	drafts, err := RepurposeAllFormats(source)
	if err != nil {
		return nil, err
	}
	for _, d := range drafts {
		if err := s.repo.SaveRepurposedDraft(ctx, d); err != nil {
			return nil, err
		}
	}
	return drafts, nil
}

func (s *Service) GetRepurposedDraft(ctx context.Context, draftID, requestingTenantID string) (*RepurposedDraft, error) {
	if strings.TrimSpace(draftID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.GetRepurposedDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) ListRepurposedDrafts(ctx context.Context, workspaceID, requestingTenantID string) ([]RepurposedDraft, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListRepurposedDrafts(ctx, workspaceID, requestingTenantID)
}

func (s *Service) UpdateRepurposedDraft(ctx context.Context, draftID, requestingTenantID, title, contentBody string) (*RepurposedDraft, error) {
	draft, err := s.repo.GetRepurposedDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	mutated := false
	if title != "" && title != draft.Title {
		draft.Title = title
		mutated = true
	}
	if contentBody != "" && contentBody != draft.ContentBody {
		draft.ContentBody = contentBody
		draft.CharacterCount = len([]rune(contentBody))
		mutated = true
	}

	// AT-007 Tamper Invalidation: Any content mutation resets approval
	if mutated {
		draft.Status = PlanStatusDraft
		draft.ApprovalToken = ""
		draft.LastApprovedAt = nil
	}

	if err := s.repo.SaveRepurposedDraft(ctx, draft); err != nil {
		return nil, err
	}
	return s.repo.GetRepurposedDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) ApproveRepurposedDraft(ctx context.Context, draftID, requestingTenantID, secretKey string) (*RepurposedDraft, error) {
	draft, err := s.repo.GetRepurposedDraft(ctx, draftID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	token := ComputeDraftApprovalToken(secretKey, draft)
	now := time.Now().UTC()
	draft.Status = PlanStatusApproved
	draft.ApprovalToken = token
	draft.LastApprovedAt = &now

	if err := s.repo.SaveRepurposedDraft(ctx, draft); err != nil {
		return nil, err
	}
	return s.repo.GetRepurposedDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) DeleteRepurposedDraft(ctx context.Context, draftID, requestingTenantID string) error {
	if strings.TrimSpace(draftID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return s.repo.DeleteRepurposedDraft(ctx, draftID, requestingTenantID)
}

func (s *Service) ScheduleContentPlan(
	ctx context.Context,
	workspaceID, tenantID, draftID string,
	requestedLocalTime time.Time,
	userTZ string,
	maxDailyBudget int,
) (*ContentPlanItem, error) {
	if err := ValidateTenantScope(workspaceID, tenantID); err != nil {
		return nil, err
	}
	draft, err := s.repo.GetRepurposedDraft(ctx, draftID, tenantID)
	if err != nil {
		return nil, err
	}

	// Query existing plans for this workspace to check collisions and daily budget
	existingPlans, err := s.repo.ListContentPlanItems(ctx, ContentCalendarFilter{
		WorkspaceID:        workspaceID,
		RequestingTenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	existingSlots := make([]time.Time, 0, len(existingPlans))
	for _, p := range existingPlans {
		existingSlots = append(existingSlots, p.ScheduledSlotUTC)
	}

	utcSlot, localFormatted, err := CalculateNextSafeScheduleSlot(requestedLocalTime, userTZ, existingSlots, maxDailyBudget)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	planID := fmt.Sprintf("plan-%d", now.UnixNano())
	composeURL := GenerateLinkedInComposeURL(draft.ContentBody)

	var slideCount int
	if draft.Format == FormatCarouselOutline {
		slideCount = 5
	}

	assetMeta := ContentAssetMetadata{
		SlideCount:           slideCount,
		EstimatedReadTimeSec: (draft.CharacterCount / 20) + 10,
		Hashtags:             draft.Tags,
	}

	item := &ContentPlanItem{
		ID:                       planID,
		WorkspaceID:              workspaceID,
		TenantID:                 tenantID,
		DraftID:                  draft.ID,
		Title:                    draft.Title,
		Format:                   draft.Format,
		ScheduledSlotUTC:         utcSlot,
		UserTimezone:             userTZ,
		LocalSlotFormatted:       localFormatted,
		Status:                   PlanStatusScheduled,
		AssetMetadata:            assetMeta,
		DirectPublishBlocked:     true, // AT-010 Strict Compliance
		ClipboardExportAvailable: true,
		ComposeURL:               composeURL,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	if err := s.repo.SaveContentPlanItem(ctx, item); err != nil {
		return nil, err
	}

	// Update draft with scheduled slot
	draft.ScheduledSlotUTC = &utcSlot
	draft.Status = PlanStatusScheduled
	_ = s.repo.SaveRepurposedDraft(ctx, draft)

	return item, nil
}

func (s *Service) GetContentPlanItem(ctx context.Context, itemID, requestingTenantID string) (*ContentPlanItem, error) {
	if strings.TrimSpace(itemID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.GetContentPlanItem(ctx, itemID, requestingTenantID)
}

func (s *Service) ListContentPlanItems(ctx context.Context, filter ContentCalendarFilter) ([]ContentPlanItem, error) {
	if strings.TrimSpace(filter.WorkspaceID) == "" || strings.TrimSpace(filter.RequestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListContentPlanItems(ctx, filter)
}

func (s *Service) UpdateContentPlanItem(
	ctx context.Context,
	itemID, requestingTenantID string,
	newTitle string,
	newLocalTime *time.Time,
	userTZ string,
	maxDailyBudget int,
) (*ContentPlanItem, error) {
	item, err := s.repo.GetContentPlanItem(ctx, itemID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	mutated := false
	if newTitle != "" && newTitle != item.Title {
		item.Title = newTitle
		mutated = true
	}

	if newLocalTime != nil {
		// Recheck slot safety
		existingPlans, err := s.repo.ListContentPlanItems(ctx, ContentCalendarFilter{
			WorkspaceID:        item.WorkspaceID,
			RequestingTenantID: requestingTenantID,
		})
		if err != nil {
			return nil, err
		}

		existingSlots := make([]time.Time, 0)
		for _, p := range existingPlans {
			if p.ID != item.ID { // Exclude self from collision check
				existingSlots = append(existingSlots, p.ScheduledSlotUTC)
			}
		}

		utcSlot, localFormatted, err := CalculateNextSafeScheduleSlot(*newLocalTime, userTZ, existingSlots, maxDailyBudget)
		if err != nil {
			return nil, err
		}
		item.ScheduledSlotUTC = utcSlot
		item.LocalSlotFormatted = localFormatted
		item.UserTimezone = userTZ
		mutated = true
	}

	// AT-007 Tamper Invalidation
	if mutated {
		item.Status = PlanStatusDraft
		item.ApprovalToken = ""
		item.LastApprovedAt = nil
	}

	if err := s.repo.SaveContentPlanItem(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetContentPlanItem(ctx, itemID, requestingTenantID)
}

func (s *Service) ApproveContentPlanItem(ctx context.Context, itemID, requestingTenantID, secretKey string) (*ContentPlanItem, error) {
	item, err := s.repo.GetContentPlanItem(ctx, itemID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	token := ComputePlanApprovalToken(secretKey, item)
	now := time.Now().UTC()
	item.Status = PlanStatusApproved
	item.ApprovalToken = token
	item.LastApprovedAt = &now

	if err := s.repo.SaveContentPlanItem(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetContentPlanItem(ctx, itemID, requestingTenantID)
}

func (s *Service) DeleteContentPlanItem(ctx context.Context, itemID, requestingTenantID string) error {
	if strings.TrimSpace(itemID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return s.repo.DeleteContentPlanItem(ctx, itemID, requestingTenantID)
}

// AttemptDirectPublish explicitly blocks automated direct posting under AT-010 & REQ-015 governance.
func (s *Service) AttemptDirectPublish(ctx context.Context, itemID, requestingTenantID string) error {
	_, err := s.repo.GetContentPlanItem(ctx, itemID, requestingTenantID)
	if err != nil {
		return err
	}
	return ErrDirectAutopublishDisallowed
}

// ExportPlanForClipboard returns the formatted markdown draft and LinkedIn Compose URL for 1-click human publishing.
func (s *Service) ExportPlanForClipboard(ctx context.Context, itemID, requestingTenantID, secretKey string) (*ContentPlanItem, string, string, error) {
	item, err := s.repo.GetContentPlanItem(ctx, itemID, requestingTenantID)
	if err != nil {
		return nil, "", "", err
	}

	if item.ApprovalToken != "" {
		if !VerifyPlanApprovalToken(secretKey, item, item.ApprovalToken) {
			return nil, "", "", fmt.Errorf("cryptographic verification failed: plan item has been tampered (AT-007)")
		}
	}

	draft, err := s.repo.GetRepurposedDraft(ctx, item.DraftID, requestingTenantID)
	if err != nil {
		return nil, "", "", err
	}

	formattedMD := FormatRepurposedMarkdown(draft)
	composeURL := GenerateLinkedInComposeURL(draft.ContentBody)

	return item, formattedMD, composeURL, nil
}

// Engagement Monitoring & Analytics Operations (IMP-LI-15, LI-15, AT-028, AT-010, AT-012, SRC-L2)

func (s *Service) RecordPostAnalytics(ctx context.Context, snapshot *PostAnalyticsSnapshot) error {
	if snapshot == nil {
		return ErrInvalidAnalyticsData
	}
	if snapshot.SnapshotID == "" {
		snapshot.SnapshotID = fmt.Sprintf("snap-%d", time.Now().UnixNano())
	}
	now := time.Now().UTC()
	if snapshot.ObservedAt.IsZero() {
		snapshot.ObservedAt = now
	}
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = now
	}
	snapshot.UpdatedAt = now

	// Compute engager segments and ICP breakdown if engagers are provided
	for i := range snapshot.Engagers {
		if snapshot.Engagers[i].Segment == "" {
			snapshot.Engagers[i].Segment = CategorizeEngager(snapshot.Engagers[i].Headline, snapshot.Engagers[i].Company)
		}
		if snapshot.Engagers[i].ObservedAt.IsZero() {
			snapshot.Engagers[i].ObservedAt = now
		}
	}
	snapshot.ICPBreakdown = ComputeICPBreakdown(snapshot.Engagers)

	return s.repo.SavePostAnalytics(ctx, snapshot)
}

func (s *Service) GetPostAnalytics(ctx context.Context, snapshotID, requestingTenantID string) (*PostAnalyticsSnapshot, error) {
	if strings.TrimSpace(snapshotID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.GetPostAnalytics(ctx, snapshotID, requestingTenantID)
}

func (s *Service) GetPostAnalyticsByURN(ctx context.Context, postURN, requestingTenantID string) (*PostAnalyticsSnapshot, error) {
	if strings.TrimSpace(postURN) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.GetPostAnalyticsByURN(ctx, postURN, requestingTenantID)
}

func (s *Service) ListPostAnalytics(ctx context.Context, workspaceID, requestingTenantID string) ([]PostAnalyticsSnapshot, error) {
	if strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListPostAnalytics(ctx, workspaceID, requestingTenantID)
}

func (s *Service) DeletePostAnalytics(ctx context.Context, snapshotID, requestingTenantID string) error {
	if strings.TrimSpace(snapshotID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return s.repo.DeletePostAnalytics(ctx, snapshotID, requestingTenantID)
}

func (s *Service) SegmentEngagers(ctx context.Context, engagers []EngagerProfile) ([]EngagerProfile, ICPBreakdown) {
	results := make([]EngagerProfile, len(engagers))
	for i, e := range engagers {
		copied := e
		copied.Segment = CategorizeEngager(e.Headline, e.Company)
		if copied.ObservedAt.IsZero() {
			copied.ObservedAt = time.Now().UTC()
		}
		results[i] = copied
	}
	breakdown := ComputeICPBreakdown(results)
	return results, breakdown
}

// Employee Advocacy Service Methods (IMP-LI-16)

func (s *Service) CreateAdvocacyCampaign(ctx context.Context, c *AdvocacyCampaign) (*AdvocacyCampaign, error) {
	if c == nil || strings.TrimSpace(c.WorkspaceID) == "" || strings.TrimSpace(c.TenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	if err := ValidateAdvocacyCampaign(c); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if c.CampaignID == "" {
		c.CampaignID = fmt.Sprintf("camp-adv-%d", now.UnixNano())
	}
	if c.Status == "" {
		c.Status = AdvocacyStatusDraft
	}
	c.CreatedAt = now.Format(time.RFC3339)
	c.UpdatedAt = now.Format(time.RFC3339)

	// Check if any variant contains forbidden keywords
	for _, v := range c.Variants {
		violations := CheckBrandCompliance(v.SuggestedText, c.Governance)
		if len(violations) > 0 {
			return nil, fmt.Errorf("%w: variant %s contains: %s", ErrForbiddenKeywordViolation, v.VariantID, strings.Join(violations, ", "))
		}
	}

	if err := s.repo.SaveAdvocacyCampaign(ctx, c); err != nil {
		return nil, err
	}
	return s.repo.GetAdvocacyCampaign(ctx, c.CampaignID, c.TenantID)
}

func (s *Service) GetAdvocacyCampaign(ctx context.Context, campaignID, requestingTenantID string) (*AdvocacyCampaign, error) {
	if strings.TrimSpace(campaignID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.GetAdvocacyCampaign(ctx, campaignID, requestingTenantID)
}

func (s *Service) ListAdvocacyCampaigns(ctx context.Context, workspaceID, requestingTenantID string) ([]AdvocacyCampaign, error) {
	if strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListAdvocacyCampaigns(ctx, workspaceID, requestingTenantID)
}

func (s *Service) UpdateAdvocacyCampaign(ctx context.Context, updated *AdvocacyCampaign, requestingTenantID string) (*AdvocacyCampaign, error) {
	if updated == nil || strings.TrimSpace(updated.CampaignID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	existing, err := s.repo.GetAdvocacyCampaign(ctx, updated.CampaignID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	// Check brand compliance
	for _, v := range updated.Variants {
		violations := CheckBrandCompliance(v.SuggestedText, updated.Governance)
		if len(violations) > 0 {
			return nil, fmt.Errorf("%w: variant %s contains: %s", ErrForbiddenKeywordViolation, v.VariantID, strings.Join(violations, ", "))
		}
	}

	// Invalidate approval token if previously approved (AT-007)
	if existing.Status == AdvocacyStatusApproved {
		updated.Status = AdvocacyStatusDraft
		updated.ApprovalToken = ""
		updated.ApprovedBy = ""
		updated.ApprovedAt = ""
	}

	updated.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := s.repo.SaveAdvocacyCampaign(ctx, updated); err != nil {
		return nil, err
	}
	return s.repo.GetAdvocacyCampaign(ctx, updated.CampaignID, requestingTenantID)
}

func (s *Service) ApproveAdvocacyCampaign(ctx context.Context, campaignID, approverID, requestingTenantID, secretKey string) (*AdvocacyCampaign, error) {
	if strings.TrimSpace(campaignID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	c, err := s.repo.GetAdvocacyCampaign(ctx, campaignID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	// Verify brand compliance before approving
	for _, v := range c.Variants {
		violations := CheckBrandCompliance(v.SuggestedText, c.Governance)
		if len(violations) > 0 {
			return nil, fmt.Errorf("%w: variant %s contains: %s", ErrForbiddenKeywordViolation, v.VariantID, strings.Join(violations, ", "))
		}
	}

	if secretKey == "" {
		secretKey = "advocacy-signing-secret"
	}

	token := ComputeCampaignApprovalToken(*c, secretKey)
	now := time.Now().UTC().Format(time.RFC3339)
	c.Status = AdvocacyStatusApproved
	c.ApprovalToken = token
	c.ApprovedBy = approverID
	c.ApprovedAt = now
	c.UpdatedAt = now

	if err := s.repo.SaveAdvocacyCampaign(ctx, c); err != nil {
		return nil, err
	}
	return s.repo.GetAdvocacyCampaign(ctx, campaignID, requestingTenantID)
}

func (s *Service) DeleteAdvocacyCampaign(ctx context.Context, campaignID, requestingTenantID string) error {
	if strings.TrimSpace(campaignID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return s.repo.DeleteAdvocacyCampaign(ctx, campaignID, requestingTenantID)
}

func (s *Service) ShareAdvocacyContent(
	ctx context.Context,
	campaignID, variantID, employeeID, customizedText, workspaceID, requestingTenantID, secretKey string,
) (*ShareAdvocacyResult, error) {
	if strings.TrimSpace(campaignID) == "" || strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	c, err := s.repo.GetAdvocacyCampaign(ctx, campaignID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	if workspaceID != "" && c.WorkspaceID != workspaceID {
		return nil, ErrCrossTenantAnalyticsDeny
	}
	if c.Status != AdvocacyStatusApproved {
		return nil, ErrCampaignNotApproved
	}

	if secretKey == "" {
		secretKey = "advocacy-signing-secret"
	}
	if !VerifyCampaignApprovalToken(*c, secretKey) {
		return nil, ErrInvalidApprovalToken
	}

	var selectedVariant *AdvocacyCopyVariant
	for _, v := range c.Variants {
		if v.VariantID == variantID {
			selectedVariant = &v
			break
		}
	}
	if selectedVariant == nil && len(c.Variants) > 0 {
		selectedVariant = &c.Variants[0]
	}
	if selectedVariant == nil {
		return nil, errors.New("no copy variant found in campaign")
	}

	finalText := selectedVariant.SuggestedText
	if strings.TrimSpace(customizedText) != "" {
		violations := CheckBrandCompliance(customizedText, c.Governance)
		if len(violations) > 0 {
			return nil, fmt.Errorf("%w: customized text contains: %s", ErrForbiddenKeywordViolation, strings.Join(violations, ", "))
		}
		finalText = customizedText
	}

	now := time.Now().UTC()
	shareID := fmt.Sprintf("share-%d", now.UnixNano())
	event := &EmployeeShareEvent{
		ShareID:        shareID,
		TenantID:       requestingTenantID,
		WorkspaceID:    c.WorkspaceID,
		CampaignID:     c.CampaignID,
		VariantID:      selectedVariant.VariantID,
		EmployeeID:     employeeID,
		CustomizedText: finalText,
		SharedAt:       now.Format(time.RFC3339),
		Platform:       "linkedin",
		ShareMethod:    "clipboard_compose_assist",
	}

	if err := s.repo.RecordEmployeeShare(ctx, event); err != nil {
		return nil, err
	}

	deepLink := GenerateComposeDeepLink(finalText)

	return &ShareAdvocacyResult{
		Success:         true,
		ShareID:         shareID,
		ClipboardText:   finalText,
		ComposeDeepLink: deepLink,
		Message:         "Content prepared for voluntary sharing. 1-click clipboard copied and LinkedIn Compose deep-link generated.",
	}, nil
}

func (s *Service) ListEmployeeShares(ctx context.Context, campaignID, workspaceID, requestingTenantID string) ([]EmployeeShareEvent, error) {
	if strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListEmployeeShares(ctx, campaignID, workspaceID, requestingTenantID)
}

func (s *Service) TriggerCoordinatedPod(ctx context.Context, req CoordinatedEngagementRequest) error {
	return RejectCoordinatedEngagementPod(req)
}

// -------------------------------------------------------------------------
// Provider Fallback & Diagnostics Service Methods (IMP-LI-17, LI-17, AT-010, AT-016, SRC-L5)
// -------------------------------------------------------------------------

func (s *Service) RegisterProviderAdapter(ctx context.Context, adapter *ProviderAdapter) error {
	if adapter == nil {
		return ErrInvalidProviderAdapter
	}
	if strings.TrimSpace(adapter.TenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return s.repo.RegisterProviderAdapter(ctx, adapter)
}

func (s *Service) GetProviderAdapter(ctx context.Context, providerID, requestingTenantID string) (*ProviderAdapter, error) {
	if strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.GetProviderAdapter(ctx, providerID, requestingTenantID)
}

func (s *Service) ListProviderAdapters(ctx context.Context, workspaceID, requestingTenantID string) ([]ProviderAdapter, error) {
	if strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	return s.repo.ListProviderAdapters(ctx, workspaceID, requestingTenantID)
}

func (s *Service) UpdateProviderHealth(ctx context.Context, providerID, requestingTenantID string, status ProviderHealthStatus, latencyMs int64, errStr string) error {
	if strings.TrimSpace(requestingTenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return s.repo.UpdateProviderHealth(ctx, providerID, requestingTenantID, status, latencyMs, errStr)
}

func (s *Service) DeleteProviderAdapter(ctx context.Context, providerID, requestingTenantID string) error {
	if strings.TrimSpace(requestingTenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return s.repo.DeleteProviderAdapter(ctx, providerID, requestingTenantID)
}

func (s *Service) RunDiagnosticDoctor(ctx context.Context, workspaceID, requestingTenantID string) (*ProviderDiagnosticReport, error) {
	if strings.TrimSpace(requestingTenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	providers, err := s.repo.ListProviderAdapters(ctx, workspaceID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	report := &ProviderDiagnosticReport{
		WorkspaceID:    workspaceID,
		TenantID:       requestingTenantID,
		TotalProviders: len(providers),
		GeneratedAt:    time.Now().UTC(),
	}

	healthyCount := 0
	degradedCount := 0

	for i := range providers {
		probe := RunDiagnosticProbe(&providers[i])
		_ = s.repo.RecordDiagnosticProbe(ctx, &probe)
		report.Probes = append(report.Probes, probe)

		if probe.HealthStatus == HealthStatusHealthy {
			healthyCount++
		} else {
			degradedCount++
		}
	}

	report.HealthyProviders = healthyCount
	report.DegradedProviders = degradedCount

	if len(providers) == 0 {
		report.OverallHealth = "critical"
	} else if degradedCount == 0 {
		report.OverallHealth = "healthy"
	} else if healthyCount > 0 {
		report.OverallHealth = "degraded"
	} else {
		report.OverallHealth = "critical"
	}

	return report, nil
}

func (s *Service) DispatchWithFallback(ctx context.Context, req DispatchActionRequest) (*ProviderDispatchResult, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.WorkspaceID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	providers, err := s.repo.ListProviderAdapters(ctx, req.WorkspaceID, req.TenantID)
	if err != nil {
		return nil, err
	}
	return ExecuteProviderActionWithFallback(providers, req)
}

func (s *Service) ScrubDiagnosticText(raw string) string {
	return ScrubDiagnosticSecrets(raw)
}

// -------------------------------------------------------------------------
// Limits, CAPTCHA/reauth handling, activity (IMP-LI-18, LI-18, AT-008, AT-009, AT-010, REQ-009, REQ-010)
// -------------------------------------------------------------------------

func (s *Service) GetAccountSafetyState(ctx context.Context, accountID, requestingTenantID string) (*AccountSafetyState, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, errors.New("accountID cannot be empty")
	}
	return s.repo.GetAccountSafetyState(ctx, accountID, requestingTenantID)
}

func (s *Service) DetectAndRecordRestriction(
	ctx context.Context,
	accountID, tenantID, workspaceID, actionAttempted string,
	statusCode int,
	rawError string,
) (*PlatformRestrictionIncident, *AccountSafetyState, error) {
	if strings.TrimSpace(accountID) == "" || strings.TrimSpace(tenantID) == "" {
		return nil, nil, errors.New("accountID and tenantID cannot be empty")
	}

	now := time.Now().UTC()
	incident := InspectPlatformErrorForRestrictions(accountID, tenantID, workspaceID, actionAttempted, statusCode, rawError, now)
	if incident == nil {
		return nil, nil, errors.New("no platform restriction pattern identified in error")
	}

	if err := s.repo.SaveRestrictionIncident(ctx, incident); err != nil {
		return nil, nil, err
	}

	state, err := s.repo.GetAccountSafetyState(ctx, accountID, tenantID)
	if err != nil {
		return nil, nil, err
	}
	if state.WorkspaceID == "" {
		state.WorkspaceID = workspaceID
	}

	ApplyRestrictionToAccount(state, incident, now)
	if err := s.repo.SaveAccountSafetyState(ctx, state); err != nil {
		return nil, nil, err
	}

	// Also log into activity log
	activity := &ActivityLogEntry{
		EntryID:     fmt.Sprintf("act-%d", now.UnixNano()),
		AccountID:   accountID,
		TenantID:    tenantID,
		WorkspaceID: workspaceID,
		Module:      "linkedin",
		Action:      actionAttempted,
		Amount:      0,
		Status:      "challenge_triggered",
		Details:     fmt.Sprintf("Restriction applied: %s (%s)", incident.LimitType, incident.RawErrorMessageScrubbed),
		Timestamp:   now,
	}
	_ = s.repo.RecordActivityLog(ctx, activity)

	return incident, state, nil
}

func (s *Service) ResolveChallenge(
	ctx context.Context,
	accountID, requestingTenantID string,
	req ResolveChallengeRequest,
) (*AccountSafetyState, error) {
	state, err := s.repo.GetAccountSafetyState(ctx, accountID, requestingTenantID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := ResolveAccountChallenge(state, req, now); err != nil {
		return nil, err
	}

	if err := s.repo.SaveAccountSafetyState(ctx, state); err != nil {
		return nil, err
	}

	activity := &ActivityLogEntry{
		EntryID:     fmt.Sprintf("act-%d", now.UnixNano()),
		AccountID:   accountID,
		TenantID:    state.TenantID,
		WorkspaceID: state.WorkspaceID,
		Module:      "linkedin",
		Action:      "resolve_challenge",
		Amount:      0,
		Status:      "success",
		Details:     fmt.Sprintf("Challenge resolved via %s by user: %s", req.ResolutionMethod, req.Notes),
		Timestamp:   now,
	}
	_ = s.repo.RecordActivityLog(ctx, activity)

	return state, nil
}

func (s *Service) RecordAccountActivity(
	ctx context.Context,
	entry *ActivityLogEntry,
) error {
	if entry == nil || strings.TrimSpace(entry.AccountID) == "" {
		return errors.New("invalid activity log entry")
	}
	return s.repo.RecordActivityLog(ctx, entry)
}

func (s *Service) ListAccountActivity(
	ctx context.Context,
	accountID, requestingTenantID string,
	limit int,
) ([]ActivityLogEntry, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, errors.New("accountID cannot be empty")
	}
	return s.repo.ListActivityLog(ctx, accountID, requestingTenantID, limit)
}

func (s *Service) GetCombinedBudgetReport(
	ctx context.Context,
	accountID, requestingTenantID string,
) (*CombinedAccountBudgetReport, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, errors.New("accountID cannot be empty")
	}

	entries, err := s.repo.ListActivityLog(ctx, accountID, requestingTenantID, 1000)
	if err != nil {
		return nil, err
	}

	report := ComputeCombinedAccountBudget(accountID, requestingTenantID, entries, time.Now().UTC())
	return &report, nil
}

func (s *Service) CheckSafetyGate(ctx context.Context, accountID, requestingTenantID, action string) (*SafetyGateDecision, error) {
	state, err := s.repo.GetAccountSafetyState(ctx, accountID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	gateErr := CheckAccountSafetyGate(state, action, now)
	decision := &SafetyGateDecision{
		Allowed: gateErr == nil,
		Status:  state.Status,
	}
	if gateErr != nil {
		decision.Reason = gateErr.Error()
	}
	return decision, nil
}

// -------------------------------------------------------------------------
// Session Lifecycle & Optional Local Tools (IMP-LI-19, LI-19, AT-016, FND-009, SRC-L3)
// -------------------------------------------------------------------------

func (s *Service) CreateSession(ctx context.Context, req CreateSessionRequest) (*LinkedInSessionRecord, error) {
	now := time.Now().UTC()
	session, err := CreateIsolatedSession(req, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Service) GetSession(ctx context.Context, sessionID, requestingTenantID string) (*LinkedInSessionRecord, error) {
	return s.repo.GetSession(ctx, sessionID, requestingTenantID)
}

func (s *Service) ListSessions(ctx context.Context, ownerID, requestingTenantID string) ([]LinkedInSessionRecord, error) {
	return s.repo.ListSessions(ctx, ownerID, requestingTenantID)
}

func (s *Service) RevokeSession(ctx context.Context, sessionID, requestingTenantID, reason string) (*LinkedInSessionRecord, error) {
	session, err := s.repo.GetSession(ctx, sessionID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := RevokeSession(session, reason, now); err != nil {
		return nil, err
	}
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Service) AcquireViewerLease(ctx context.Context, sessionID, requestingTenantID string, req AcquireViewerLeaseRequest) (*ShortLivedViewerLease, error) {
	session, err := s.repo.GetSession(ctx, sessionID, requestingTenantID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	lease, err := AcquireShortLivedViewerLease(session, req, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return nil, err
	}
	if err := s.repo.SaveViewerLease(ctx, lease); err != nil {
		return nil, err
	}
	return lease, nil
}

func (s *Service) ValidateViewerLease(ctx context.Context, leaseID string) (*ShortLivedViewerLease, error) {
	lease, err := s.repo.GetViewerLease(ctx, leaseID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := ValidateViewerLease(lease, now); err != nil {
		_ = s.repo.SaveViewerLease(ctx, lease) // Persist expired status
		return lease, err
	}
	return lease, nil
}

func (s *Service) ReleaseViewerLease(ctx context.Context, leaseID, requestingTenantID string) error {
	lease, err := s.repo.GetViewerLease(ctx, leaseID)
	if err != nil {
		return err
	}
	session, err := s.repo.GetSession(ctx, lease.SessionID, requestingTenantID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := ReleaseViewerLease(lease, session, now); err != nil {
		return err
	}
	_ = s.repo.SaveViewerLease(ctx, lease)
	_ = s.repo.SaveSession(ctx, session)
	return nil
}

// --- Relationship CSV/Sheets Export Service Methods (IMP-LI-20, LI-20, REQ-006, REQ-019, AT-013, AT-014) ---

func (s *Service) ExportRelationshipsCSV(
	ctx context.Context,
	filter RelationshipExportFilter,
) (*RelationshipExportManifest, error) {
	if strings.TrimSpace(filter.WorkspaceID) == "" {
		return nil, ErrEmptyExportWorkspace
	}
	if strings.TrimSpace(filter.TenantID) == "" {
		return nil, ErrEmptyExportTenant
	}
	if filter.RequestingTenantID != "" && filter.RequestingTenantID != filter.TenantID {
		return nil, ErrCrossTenantExportDenied
	}

	// Fetch scoped leads
	leadFilter := RecruiterLeadFilter{
		WorkspaceID: filter.WorkspaceID,
		TenantID:    filter.TenantID,
		Status:      filter.StatusFilter,
	}
	leads, err := s.repo.ListRecruiterLeads(ctx, leadFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list recruiter leads for export: %w", err)
	}

	// Fetch person records map if requested
	var personMap map[string]*PersonRecord
	if filter.IncludePersonRecords {
		personMap = make(map[string]*PersonRecord)
		persons, err := s.repo.ListPersonRecords(ctx, RecordFilter{
			WorkspaceID: filter.WorkspaceID,
			TenantID:    filter.TenantID,
		})
		if err == nil {
			for i := range persons {
				personMap[persons[i].RecordID] = &persons[i]
			}
		}
	}

	manifest, err := GenerateRelationshipCSV(leads, personMap, filter)
	if err != nil {
		return nil, err
	}

	// Record durable export audit log (REQ-019, AT-022)
	auditRec := &RelationshipExportAuditRecord{
		AuditID:      fmt.Sprintf("audit_rel_csv_%d", time.Now().UnixNano()),
		WorkspaceID:  filter.WorkspaceID,
		TenantID:     filter.TenantID,
		OwnerID:      filter.RequestingOwnerID,
		Destination:  DestinationCSVDownload,
		Filename:     manifest.Filename,
		RowCount:     manifest.RowCount,
		IncludeNotes: filter.IncludeNotes,
		ExportedAt:   time.Now().UTC(),
	}
	_ = s.repo.SaveRelationshipExportAudit(ctx, auditRec)

	return manifest, nil
}

func (s *Service) SyncRelationshipsToSheets(
	ctx context.Context,
	existingRows [][]string,
	config RelationshipSheetsSyncConfig,
	requestingTenantID, requestingOwnerID string,
) ([][]string, RelationshipSheetsSyncResult, error) {
	if strings.TrimSpace(config.WorkspaceID) == "" {
		return nil, RelationshipSheetsSyncResult{}, ErrEmptyExportWorkspace
	}
	if strings.TrimSpace(config.TenantID) == "" {
		return nil, RelationshipSheetsSyncResult{}, ErrEmptyExportTenant
	}
	if requestingTenantID != "" && requestingTenantID != config.TenantID {
		return nil, RelationshipSheetsSyncResult{}, ErrCrossTenantExportDenied
	}

	leads, err := s.repo.ListRecruiterLeads(ctx, RecruiterLeadFilter{
		WorkspaceID: config.WorkspaceID,
		TenantID:    config.TenantID,
	})
	if err != nil {
		return nil, RelationshipSheetsSyncResult{}, fmt.Errorf("failed to list leads for sheets sync: %w", err)
	}

	personMap := make(map[string]*PersonRecord)
	persons, err := s.repo.ListPersonRecords(ctx, RecordFilter{
		WorkspaceID: config.WorkspaceID,
		TenantID:    config.TenantID,
	})
	if err == nil {
		for i := range persons {
			personMap[persons[i].RecordID] = &persons[i]
		}
	}

	outRows, result, err := ReconcileRelationshipSheetsProjection(existingRows, leads, personMap, config)
	if err != nil {
		return nil, result, err
	}

	// Record audit record
	auditRec := &RelationshipExportAuditRecord{
		AuditID:      fmt.Sprintf("audit_rel_sheets_%d", time.Now().UnixNano()),
		WorkspaceID:  config.WorkspaceID,
		TenantID:     config.TenantID,
		OwnerID:      requestingOwnerID,
		Destination:  DestinationSheetsSync,
		Filename:     fmt.Sprintf("%s/%s", config.SpreadsheetID, config.SheetName),
		RowCount:     result.TotalSynced,
		IncludeNotes: config.IncludeNotes,
		ExportedAt:   time.Now().UTC(),
	}
	_ = s.repo.SaveRelationshipExportAudit(ctx, auditRec)

	return outRows, result, nil
}

func (s *Service) ListRelationshipExportAudits(
	ctx context.Context,
	workspaceID, requestingTenantID string,
) ([]RelationshipExportAuditRecord, error) {
	return s.repo.ListRelationshipExportAudits(ctx, workspaceID, requestingTenantID)
}

