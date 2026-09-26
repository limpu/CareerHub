package linkedin

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Standard domain errors for platform limits, challenges, and safety gatekeeping
var (
	ErrAccountCircuitBreakerOpen = errors.New("account circuit breaker is open; execution halted for safety (AT-010)")
	ErrAccountChallengeRequired  = errors.New("legitimate user-facing re-authentication required; anti-bot bypass permanently prohibited (AT-010)")
	ErrAccountRateLimited        = errors.New("account is currently in a cooling-off rate limit period (REQ-009)")
	ErrStealthBypassProhibited   = errors.New("automated bypass of CAPTCHA or security checkpoint is strictly prohibited by platform policy (AT-010)")
	ErrAccountBudgetExceeded     = errors.New("account-wide combined budget exceeded across workspaces (AT-008, AT-009)")
	ErrInvalidResolution        = errors.New("invalid challenge resolution request: user verification required")
	ErrAccountSafetyNotFound     = errors.New("account safety state record not found")
)

// PlatformLimitType represents categorized external limits enforced by LinkedIn.
type PlatformLimitType string

const (
	LimitWeeklyInvitations  PlatformLimitType = "weekly_invitations"
	LimitDailyMessages      PlatformLimitType = "daily_messages"
	LimitDailyRecruiterDM   PlatformLimitType = "daily_recruiter_dm"
	LimitProfileSearches    PlatformLimitType = "profile_searches"
	LimitCompanyFollows     PlatformLimitType = "company_follows"
	LimitRateLimit429       PlatformLimitType = "rate_limit_429"
	LimitSecurityCheckpoint PlatformLimitType = "security_checkpoint"
	LimitSessionExpired     PlatformLimitType = "session_expired"
)

// ChallengeType represents specific security challenges presented by LinkedIn.
type ChallengeType string

const (
	ChallengeCaptcha               ChallengeType = "captcha"
	ChallengeEmailPIN              ChallengeType = "email_pin"
	ChallengeSMS2FA                ChallengeType = "sms_2fa"
	ChallengeCredentialReauth      ChallengeType = "credential_reauth"
	ChallengeCommercialUseCap      ChallengeType = "commercial_use_cap"
	ChallengeWeeklyInvitationLimit ChallengeType = "weekly_invitation_limit"
)

// AccountSafetyStatus defines the circuit-breaker status of a connected account.
type AccountSafetyStatus string

const (
	SafetyStatusActive            AccountSafetyStatus = "active"
	SafetyStatusPaused            AccountSafetyStatus = "paused"
	SafetyStatusChallengeRequired AccountSafetyStatus = "challenge_required"
	SafetyStatusRateLimited       AccountSafetyStatus = "rate_limited"
	SafetyStatusRevoked           AccountSafetyStatus = "revoked"
)

// AccountSafetyState holds the live circuit breaker and safety status of an external account.
type AccountSafetyState struct {
	AccountID                string              `json:"account_id"`
	TenantID                 string              `json:"tenant_id"`
	WorkspaceID              string              `json:"workspace_id"`
	Status                   AccountSafetyStatus `json:"status"`
	ActiveLimitType          *PlatformLimitType  `json:"active_limit_type,omitempty"`
	ActiveChallenge          *ChallengeType      `json:"active_challenge,omitempty"`
	ChallengeReason          string              `json:"challenge_reason,omitempty"`
	CooldownUntil            *time.Time          `json:"cooldown_until,omitempty"`
	ConsecutiveRestrictions  int                 `json:"consecutive_restrictions"`
	LastIncidentAt           *time.Time          `json:"last_incident_at,omitempty"`
	LastResolvedAt           *time.Time          `json:"last_resolved_at,omitempty"`
	UpdatedAt                time.Time           `json:"updated_at"`
}

// PlatformRestrictionIncident documents a detected upstream limitation or challenge.
type PlatformRestrictionIncident struct {
	IncidentID               string            `json:"incident_id"`
	AccountID                string            `json:"account_id"`
	TenantID                 string            `json:"tenant_id"`
	WorkspaceID              string            `json:"workspace_id"`
	StatusCode               int               `json:"status_code"`
	RawErrorMessageScrubbed  string            `json:"raw_error_message_scrubbed"`
	LimitType                PlatformLimitType `json:"limit_type"`
	ChallengeType            *ChallengeType    `json:"challenge_type,omitempty"`
	ActionAttempted          string            `json:"action_attempted"`
	CooldownHours            int               `json:"cooldown_hours"`
	OccurredAt               time.Time         `json:"occurred_at"`
}

// ActivityLogEntry tracks actions across modules and workspaces for the same account (AT-009).
type ActivityLogEntry struct {
	EntryID     string    `json:"entry_id"`
	AccountID   string    `json:"account_id"`
	TenantID    string    `json:"tenant_id"`
	WorkspaceID string    `json:"workspace_id"`
	Module      string    `json:"module"` // linkedin, career, social
	Action      string    `json:"action"` // connection_request, recruiter_dm, company_follow, profile_search
	Amount      int       `json:"amount"`
	Status      string    `json:"status"` // dispatched, blocked, throttled, challenge_triggered
	Details     string    `json:"details,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// CombinedAccountBudgetReport summarizes account-wide shared usage and remaining quotas.
type CombinedAccountBudgetReport struct {
	AccountID                    string    `json:"account_id"`
	TenantID                     string    `json:"tenant_id"`
	ComputedAt                   time.Time `json:"computed_at"`
	DailyInvitationsLimit        int       `json:"daily_invitations_limit"`
	DailyInvitationsUsed         int       `json:"daily_invitations_used"`
	DailyInvitationsRemaining    int       `json:"daily_invitations_remaining"`
	WeeklyInvitationsLimit       int       `json:"weekly_invitations_limit"`
	WeeklyInvitationsUsed        int       `json:"weekly_invitations_used"`
	WeeklyInvitationsRemaining   int       `json:"weekly_invitations_remaining"`
	DailyInMailLimit             int       `json:"daily_inmail_limit"`
	DailyInMailUsed              int       `json:"daily_inmail_used"`
	DailyInMailRemaining         int       `json:"daily_inmail_remaining"`
	DailyRecruiterDMLimit        int       `json:"daily_recruiter_dm_limit"`
	DailyRecruiterDMUsed         int       `json:"daily_recruiter_dm_used"`
	DailyRecruiterDMRemaining    int       `json:"daily_recruiter_dm_remaining"`
	DailyCompanyFollowsLimit     int       `json:"daily_company_follows_limit"`
	DailyCompanyFollowsUsed      int       `json:"daily_company_follows_used"`
	DailyCompanyFollowsRemaining int       `json:"daily_company_follows_remaining"`
	ContributingWorkspaces       []string  `json:"contributing_workspaces"`
}

// ResolveChallengeRequest defines the user-directed reauth clearance payload.
type ResolveChallengeRequest struct {
	ResolutionMethod string `json:"resolution_method"`
	VerifiedByUser   bool   `json:"verified_by_user"`
	Notes            string `json:"notes"`
}

// SafetyGateDecision represents the evaluation result of the circuit breaker gatekeeper.
type SafetyGateDecision struct {
	Allowed bool                `json:"allowed"`
	Status  AccountSafetyStatus `json:"status"`
	Reason  string              `json:"reason,omitempty"`
}

// Upstream regex matchers for restriction and challenge identification (SRC-S3, SRC-L1).
var (
	weeklyInviteRegex    = regexp.MustCompile(`(?i)(weekly invitation limit|invitation_limit_exceeded|invitations will be refreshed next week)`)
	checkpointRegex      = regexp.MustCompile(`(?i)(checkpoint|challenge/captcha|security checkpoint|user interaction required|automated access prohibited)`)
	commercialSearchRegex = regexp.MustCompile(`(?i)(commercial use limit|out of free commercial searches|commercial_use_cap)`)
	rateLimit429Regex    = regexp.MustCompile(`(?i)(too many requests|rate limit exceeded|rate_limit_exceeded)`)
	sessionExpiredRegex  = regexp.MustCompile(`(?i)(session expired|unauthorized|invalid_token|revoked_token)`)
)

// InspectPlatformErrorForRestrictions analyzes HTTP status codes and error payloads
// to accurately categorize LinkedIn limits without false simulations (SRC-S3, REQ-009).
func InspectPlatformErrorForRestrictions(
	accountID, tenantID, workspaceID, actionAttempted string,
	statusCode int,
	rawError string,
	now time.Time,
) *PlatformRestrictionIncident {
	scrubbed := ScrubDiagnosticSecrets(rawError)
	incident := &PlatformRestrictionIncident{
		IncidentID:              fmt.Sprintf("inc-%d", now.UnixNano()),
		AccountID:               accountID,
		TenantID:                tenantID,
		WorkspaceID:             workspaceID,
		StatusCode:              statusCode,
		RawErrorMessageScrubbed: scrubbed,
		ActionAttempted:         actionAttempted,
		OccurredAt:              now,
	}

	lower := strings.ToLower(rawError)

	// 1. Security Checkpoint & CAPTCHA (AT-010, SRC-S3)
	if statusCode == 403 && checkpointRegex.MatchString(lower) {
		incident.LimitType = LimitSecurityCheckpoint
		chal := ChallengeCaptcha
		incident.ChallengeType = &chal
		incident.CooldownHours = 0 // Stays blocked until legitimate reauth
		return incident
	}

	// 2. Weekly Invitation Limit (REQ-009, REQ-010, SRC-L1)
	if weeklyInviteRegex.MatchString(lower) || (statusCode == 429 && strings.Contains(lower, "invitation")) {
		incident.LimitType = LimitWeeklyInvitations
		chal := ChallengeWeeklyInvitationLimit
		incident.ChallengeType = &chal
		incident.CooldownHours = 72 // 3-day minimum safe cooldown until next calendar cycle
		return incident
	}

	// 3. Commercial Search / Profile Use Cap
	if commercialSearchRegex.MatchString(lower) {
		incident.LimitType = LimitProfileSearches
		chal := ChallengeCommercialUseCap
		incident.ChallengeType = &chal
		incident.CooldownHours = 24
		return incident
	}

	// 4. General HTTP 429 Too Many Requests
	if statusCode == 429 || rateLimit429Regex.MatchString(lower) {
		incident.LimitType = LimitRateLimit429
		incident.CooldownHours = 4 // 4-hour adaptive backoff
		return incident
	}

	// 5. Session Expiry / Reauth
	if statusCode == 401 || sessionExpiredRegex.MatchString(lower) {
		incident.LimitType = LimitSessionExpired
		chal := ChallengeCredentialReauth
		incident.ChallengeType = &chal
		incident.CooldownHours = 0
		return incident
	}

	// Default fallback restriction
	if statusCode >= 400 {
		incident.LimitType = LimitRateLimit429
		incident.CooldownHours = 1
		return incident
	}

	return nil
}

// ApplyRestrictionToAccount transitions the account circuit breaker to a protective state (fail-closed).
func ApplyRestrictionToAccount(state *AccountSafetyState, incident *PlatformRestrictionIncident, now time.Time) {
	if state == nil || incident == nil {
		return
	}

	state.LastIncidentAt = &now
	state.ConsecutiveRestrictions++
	state.ActiveLimitType = &incident.LimitType
	state.ChallengeReason = incident.RawErrorMessageScrubbed
	state.UpdatedAt = now

	if incident.ChallengeType != nil {
		state.ActiveChallenge = incident.ChallengeType
	}

	// State machine transition: Checkpoints require re-auth; Rate limits require cooldown
	if incident.LimitType == LimitSecurityCheckpoint || incident.LimitType == LimitSessionExpired {
		state.Status = SafetyStatusChallengeRequired
		state.CooldownUntil = nil
	} else if incident.CooldownHours > 0 {
		state.Status = SafetyStatusRateLimited
		cooldown := now.Add(time.Duration(incident.CooldownHours) * time.Hour)
		state.CooldownUntil = &cooldown
	} else {
		state.Status = SafetyStatusPaused
	}
}

// CheckAccountSafetyGate enforces fail-closed gatekeeping prior to any live action execution (AT-010).
func CheckAccountSafetyGate(state *AccountSafetyState, action string, now time.Time) error {
	if state == nil {
		return nil
	}

	// Check if cooldown has expired for rate-limited accounts
	if state.Status == SafetyStatusRateLimited && state.CooldownUntil != nil {
		if now.After(*state.CooldownUntil) {
			state.Status = SafetyStatusActive
			state.ActiveLimitType = nil
			state.ActiveChallenge = nil
			state.CooldownUntil = nil
			state.UpdatedAt = now
			return nil
		}
		return fmt.Errorf("%w: account cooldown active until %s for limit %v",
			ErrAccountRateLimited, state.CooldownUntil.Format(time.RFC3339), state.ActiveLimitType)
	}

	switch state.Status {
	case SafetyStatusChallengeRequired:
		return fmt.Errorf("%w: active challenge (%v) requires legitimate re-authentication. Reason: %s",
			ErrAccountChallengeRequired, state.ActiveChallenge, state.ChallengeReason)
	case SafetyStatusPaused:
		return fmt.Errorf("%w: account is paused for safety reviews", ErrAccountCircuitBreakerOpen)
	case SafetyStatusRevoked:
		return fmt.Errorf("%w: account access credentials have been revoked", ErrAccountCircuitBreakerOpen)
	case SafetyStatusActive:
		return nil
	default:
		return nil
	}
}

// ResolveAccountChallenge processes legitimate user-directed re-authentication (AT-010, AT-016).
// Evasive stealth bypasses (CDP injection, turnstile solvers) are strictly rejected.
func ResolveAccountChallenge(state *AccountSafetyState, req ResolveChallengeRequest, now time.Time) error {
	if state == nil {
		return ErrAccountSafetyNotFound
	}

	if !req.VerifiedByUser {
		return fmt.Errorf("%w: user verification must be explicitly confirmed. Automated bypasses are prohibited (AT-010)", ErrInvalidResolution)
	}

	// Disallow any bypass claiming automated challenge solve
	lowerMethod := strings.ToLower(req.ResolutionMethod)
	if strings.Contains(lowerMethod, "stealth") || strings.Contains(lowerMethod, "bypass") || strings.Contains(lowerMethod, "auto_solve") {
		return fmt.Errorf("%w: resolution method '%s' violates platform integrity policy (AT-010)", ErrStealthBypassProhibited, req.ResolutionMethod)
	}

	state.Status = SafetyStatusActive
	state.ActiveChallenge = nil
	state.ActiveLimitType = nil
	state.ChallengeReason = ""
	state.CooldownUntil = nil
	state.LastResolvedAt = &now
	state.UpdatedAt = now

	return nil
}

// ComputeCombinedAccountBudget aggregates cross-workspace usage for the same provider account (AT-008, AT-009, REQ-010).
func ComputeCombinedAccountBudget(
	accountID, tenantID string,
	entries []ActivityLogEntry,
	now time.Time,
) CombinedAccountBudgetReport {
	// Standard conservative defaults (AT-008, REQ-010, CONTEXT.md Section 8.2)
	report := CombinedAccountBudgetReport{
		AccountID:                    accountID,
		TenantID:                     tenantID,
		ComputedAt:                   now,
		DailyInvitationsLimit:        20,
		WeeklyInvitationsLimit:       80,
		DailyInMailLimit:             15,
		DailyRecruiterDMLimit:        5, // Nested cap inside DailyInMailLimit
		DailyCompanyFollowsLimit:     15,
		ContributingWorkspaces:       []string{},
	}

	oneDayAgo := now.Add(-24 * time.Hour)
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)

	workspaceSet := make(map[string]bool)

	for _, e := range entries {
		if e.AccountID != accountID {
			continue
		}
		if e.Status != "dispatched" && e.Status != "success" && e.Status != "" {
			continue
		}

		workspaceSet[e.WorkspaceID] = true

		switch e.Action {
		case "connection_request":
			if e.Timestamp.After(sevenDaysAgo) {
				report.WeeklyInvitationsUsed += e.Amount
			}
			if e.Timestamp.After(oneDayAgo) {
				report.DailyInvitationsUsed += e.Amount
			}
		case "recruiter_dm":
			if e.Timestamp.After(oneDayAgo) {
				report.DailyRecruiterDMUsed += e.Amount
				report.DailyInMailUsed += e.Amount // Nested consumption
			}
		case "inmail_message", "inmail", "direct_message":
			if e.Timestamp.After(oneDayAgo) {
				report.DailyInMailUsed += e.Amount
			}
		case "company_follow":
			if e.Timestamp.After(oneDayAgo) {
				report.DailyCompanyFollowsUsed += e.Amount
			}
		}
	}

	// Calculate remaining allowances
	report.DailyInvitationsRemaining = report.DailyInvitationsLimit - report.DailyInvitationsUsed
	if report.DailyInvitationsRemaining < 0 {
		report.DailyInvitationsRemaining = 0
	}

	report.WeeklyInvitationsRemaining = report.WeeklyInvitationsLimit - report.WeeklyInvitationsUsed
	if report.WeeklyInvitationsRemaining < 0 {
		report.WeeklyInvitationsRemaining = 0
	}

	report.DailyInMailRemaining = report.DailyInMailLimit - report.DailyInMailUsed
	if report.DailyInMailRemaining < 0 {
		report.DailyInMailRemaining = 0
	}

	report.DailyRecruiterDMRemaining = report.DailyRecruiterDMLimit - report.DailyRecruiterDMUsed
	if report.DailyRecruiterDMRemaining < 0 {
		report.DailyRecruiterDMRemaining = 0
	}

	report.DailyCompanyFollowsRemaining = report.DailyCompanyFollowsLimit - report.DailyCompanyFollowsUsed
	if report.DailyCompanyFollowsRemaining < 0 {
		report.DailyCompanyFollowsRemaining = 0
	}

	for ws := range workspaceSet {
		if ws != "" {
			report.ContributingWorkspaces = append(report.ContributingWorkspaces, ws)
		}
	}

	return report
}
