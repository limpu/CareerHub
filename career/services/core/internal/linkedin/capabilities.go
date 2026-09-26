package linkedin

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrCapabilityNotGranted             = errors.New("capability not granted by credential scopes")
	ErrTokenExpiredOrRevoked            = errors.New("token is expired or revoked; reauthentication required (AT-016)")
	ErrDirectMessagingPartnerRequired   = errors.New("direct messaging requires LinkedIn enterprise partner approval; automated live background execution is prohibited (AT-010)")
	ErrEasyApplyLiveUnsupported         = errors.New("automated Easy Apply background write is unsupported on live LinkedIn (AT-010)")
	ErrConnectionNotFound               = errors.New("linkedin connection not found")
	ErrInvalidScopeRequest              = errors.New("invalid or empty scope request")
)

// ScopeDomain categorizes capabilities into distinct permission boundaries (LI-01, REQ-004).
type ScopeDomain string

const (
	DomainIdentity       ScopeDomain = "identity"
	DomainProfileRead    ScopeDomain = "profile_read"
	DomainContentPublish ScopeDomain = "content_publish"
	DomainMessaging      ScopeDomain = "messaging"
	DomainEasyApply      ScopeDomain = "easy_apply"
)

// CapabilityStatus describes the availability and authorization state of a capability (AT-010).
type CapabilityStatus string

const (
	StatusGranted                CapabilityStatus = "granted"
	StatusDenied                 CapabilityStatus = "denied"
	StatusUnavailablePartnerOnly CapabilityStatus = "unavailable_partner_only"
	StatusUnsupportedPlatform   CapabilityStatus = "unsupported_platform"
)

// ConnectionStatus tracks token freshness and re-authentication requirements (AT-016).
type ConnectionStatus string

const (
	ConnConnected           ConnectionStatus = "connected"
	ConnPartiallyAuthorized ConnectionStatus = "partially_authorized"
	ConnReauthRequired      ConnectionStatus = "reauth_required"
	ConnDisconnected        ConnectionStatus = "disconnected"
)

// FallbackMode defines the safe alternative path when direct live automation is disallowed (AT-010).
type FallbackMode string

const (
	FallbackNone                 FallbackMode = "none"
	FallbackUserExportUpload     FallbackMode = "user_export_upload"
	FallbackManualClipboardAssist FallbackMode = "manual_clipboard_assist"
)

// InspectedCapability is an individual audited capability with truth-in-advertising disclosures.
type InspectedCapability struct {
	CapabilityID            string           `json:"capability_id"`
	Domain                  ScopeDomain      `json:"domain"`
	Name                    string           `json:"name"`
	Description             string           `json:"description"`
	RequiredScope           string           `json:"required_scope"`
	Status                  CapabilityStatus `json:"status"`
	IsGranted               bool             `json:"is_granted"`
	TruthInAdvertisingNote  string           `json:"truth_in_advertising_note"`
	FallbackMode            FallbackMode     `json:"fallback_mode"`
}

// LinkedInConnection holds user-authorized LinkedIn connection state and credentials (AT-016).
type LinkedInConnection struct {
	ConnectionID     string                `json:"connection_id"`
	UserID           string                `json:"user_id"`
	MemberID         string                `json:"member_id"`
	DisplayName      string                `json:"display_name"`
	Email            string                `json:"email"`
	ProfilePicture   string                `json:"profile_picture,omitempty"`
	GrantedScopes    []string              `json:"granted_scopes"`
	DeniedScopes     []string              `json:"denied_scopes"`
	MaskedToken      string                `json:"masked_token"` // e.g. "tok_li_****_88fa"
	TokenStatus      string                `json:"token_status"` // "active", "expired", "revoked"
	ConnectedAt      time.Time             `json:"connected_at"`
	ExpiresAt        *time.Time            `json:"expires_at,omitempty"`
	Status           ConnectionStatus      `json:"status"`
	Capabilities     []InspectedCapability `json:"capabilities"`
	LastInspectedAt  time.Time             `json:"last_inspected_at"`
}

// CapabilityTemplate defines canonical capability specifications.
type CapabilityTemplate struct {
	ID                     string
	Domain                 ScopeDomain
	Name                   string
	Description            string
	RequiredScope          string
	IsPartnerOnly          bool
	IsPlatformUnsupported  bool
	TruthInAdvertisingNote string
	Fallback               FallbackMode
}

// Catalog of canonical LinkedIn capabilities (IMP-LI-01, LI-01).
var standardCapabilityTemplates = []CapabilityTemplate{
	{
		ID:                     "identity_signin",
		Domain:                 DomainIdentity,
		Name:                   "Sign In with LinkedIn (OpenID Connect)",
		Description:            "Authenticate user identity, verified email address, and core OIDC claims.",
		RequiredScope:          "openid",
		IsPartnerOnly:          false,
		TruthInAdvertisingNote: "Standard OpenID Connect provides identity authentication only; it conveys zero access to full career history, messaging, or job applications.",
		Fallback:               FallbackNone,
	},
	{
		ID:                     "basic_profile_read",
		Domain:                 DomainProfileRead,
		Name:                   "Basic Profile Read",
		Description:            "Read public name, vanity slug, profile photo, and localized headline.",
		RequiredScope:          "profile",
		IsPartnerOnly:          false,
		TruthInAdvertisingNote: "Grants basic headline and portrait photo only. Complete work experience and resume history are not returned by consumer profile scopes.",
		Fallback:               FallbackNone,
	},
	{
		ID:                     "career_history_import",
		Domain:                 DomainProfileRead,
		Name:                   "Career History & Experience Ingestion",
		Description:            "Direct programmatic import of structured employment history, education, and licenses.",
		RequiredScope:          "r_fullprofile",
		IsPartnerOnly:          true,
		TruthInAdvertisingNote: "Full profile scraping via API is gated to vetted Talent Solutions Enterprise partners. Supported platform fallback: upload your official LinkedIn Profile Export ZIP (Basic_LinkedInData.zip).",
		Fallback:               FallbackUserExportUpload,
	},
	{
		ID:                     "content_publish",
		Domain:                 DomainContentPublish,
		Name:                   "Member Post & Article Publishing",
		Description:            "Publish user-approved professional updates and post drafts to the candidate's personal feed.",
		RequiredScope:          "w_member_social",
		IsPartnerOnly:          false,
		TruthInAdvertisingNote: "Permits publishing user-generated content strictly with explicit human confirmation (AT-007). Coordinated engagement rings or automated bulk reposts are strictly blocked.",
		Fallback:               FallbackManualClipboardAssist,
	},
	{
		ID:                     "direct_messaging",
		Domain:                 DomainMessaging,
		Name:                   "Direct Messaging & InMail",
		Description:            "Programmatic message read and write across 1st-degree connections and recruiter threads.",
		RequiredScope:          "r_messages",
		IsPartnerOnly:          true,
		TruthInAdvertisingNote: "Standard consumer applications are not granted messaging automation. The platform provides assisted-manual copy-paste drafts to prevent account bans (AT-010).",
		Fallback:               FallbackManualClipboardAssist,
	},
	{
		ID:                     "automated_easy_apply",
		Domain:                 DomainEasyApply,
		Name:                   "Automated Easy Apply Submissions",
		Description:            "Background execution of multi-step job application submissions on LinkedIn job posts.",
		RequiredScope:          "w_job_applications",
		IsPartnerOnly:          false,
		IsPlatformUnsupported:  true,
		TruthInAdvertisingNote: "Direct automated background submission without human-in-the-loop review violates LinkedIn terms of service and is permanently blocked (REQ-015, AT-010). Use Assisted-Manual flow with pre-filled clipboard answers.",
		Fallback:               FallbackManualClipboardAssist,
	},
}

// InspectCapabilities evaluates connection scopes against the standard capability catalog (LI-01).
func InspectCapabilities(conn *LinkedInConnection) []InspectedCapability {
	if conn == nil {
		return nil
	}

	scopeSet := make(map[string]bool)
	for _, s := range conn.GrantedScopes {
		scopeSet[strings.ToLower(strings.TrimSpace(s))] = true
	}

	isExpired := conn.TokenStatus == "expired" || conn.TokenStatus == "revoked"
	if conn.ExpiresAt != nil && time.Now().UTC().After(*conn.ExpiresAt) {
		isExpired = true
	}

	capabilities := make([]InspectedCapability, 0, len(standardCapabilityTemplates))

	for _, tmpl := range standardCapabilityTemplates {
		cap := InspectedCapability{
			CapabilityID:           tmpl.ID,
			Domain:                 tmpl.Domain,
			Name:                   tmpl.Name,
			Description:            tmpl.Description,
			RequiredScope:          tmpl.RequiredScope,
			TruthInAdvertisingNote: tmpl.TruthInAdvertisingNote,
			FallbackMode:           tmpl.Fallback,
		}

		if isExpired {
			cap.IsGranted = false
			cap.Status = StatusDenied
			capabilities = append(capabilities, cap)
			continue
		}

		if tmpl.IsPlatformUnsupported {
			cap.IsGranted = false
			cap.Status = StatusUnsupportedPlatform
		} else if tmpl.IsPartnerOnly {
			// Check if partner scope is granted
			if scopeSet[tmpl.RequiredScope] {
				cap.IsGranted = true
				cap.Status = StatusGranted
			} else {
				cap.IsGranted = false
				cap.Status = StatusUnavailablePartnerOnly
			}
		} else {
			// Standard consumer scope
			if scopeSet[tmpl.RequiredScope] {
				cap.IsGranted = true
				cap.Status = StatusGranted
			} else {
				cap.IsGranted = false
				cap.Status = StatusDenied
			}
		}

		capabilities = append(capabilities, cap)
	}

	return capabilities
}

// CanExecuteAction performs fail-closed capability authorization before executing any action (AT-010, AT-016).
func CanExecuteAction(conn *LinkedInConnection, action string) (bool, *InspectedCapability, error) {
	if conn == nil {
		return false, nil, ErrConnectionNotFound
	}

	// Check token expiration
	if conn.TokenStatus == "expired" || conn.TokenStatus == "revoked" {
		return false, nil, ErrTokenExpiredOrRevoked
	}
	if conn.ExpiresAt != nil && time.Now().UTC().After(*conn.ExpiresAt) {
		return false, nil, ErrTokenExpiredOrRevoked
	}

	var targetCapID string
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "sign_in", "auth_identity", "get_me":
		targetCapID = "identity_signin"
	case "read_profile", "read_vanity":
		targetCapID = "basic_profile_read"
	case "import_experience", "sync_work_history":
		targetCapID = "career_history_import"
	case "publish_post", "create_post", "share_article":
		targetCapID = "content_publish"
	case "send_direct_message", "send_inmail", "reply_message":
		targetCapID = "direct_messaging"
	case "automated_easy_apply", "submit_easy_apply":
		targetCapID = "automated_easy_apply"
	default:
		return false, nil, ErrCapabilityNotGranted
	}

	capabilities := InspectCapabilities(conn)
	for i := range capabilities {
		cap := &capabilities[i]
		if cap.CapabilityID == targetCapID {
			if cap.IsGranted {
				return true, cap, nil
			}

			// Specific descriptive errors (AT-010)
			switch cap.Status {
			case StatusUnavailablePartnerOnly:
				if cap.CapabilityID == "direct_messaging" {
					return false, cap, ErrDirectMessagingPartnerRequired
				}
				return false, cap, ErrCapabilityNotGranted
			case StatusUnsupportedPlatform:
				if cap.CapabilityID == "automated_easy_apply" {
					return false, cap, ErrEasyApplyLiveUnsupported
				}
				return false, cap, ErrCapabilityNotGranted
			default:
				return false, cap, ErrCapabilityNotGranted
			}
		}
	}

	return false, nil, ErrCapabilityNotGranted
}
