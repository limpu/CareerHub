package linkedin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// Invariants & Domain Errors
var (
	ErrCoordinatedEngagementProhibited = errors.New("coordinated_engagement_prohibited: coordinated engagement pods and automated reciprocal likes are strictly prohibited by platform integrity policies")
	ErrCampaignTitleRequired           = errors.New("campaign title is required")
	ErrCampaignVariantsRequired        = errors.New("at least one suggested copy variant is required")
	ErrCampaignNotFound                = errors.New("advocacy campaign not found")
	ErrCampaignNotApproved             = errors.New("advocacy campaign must be approved before employee sharing")
	ErrForbiddenKeywordViolation       = errors.New("content contains forbidden brand keywords")
)

// AdvocacyCampaignStatus represents the lifecycle state of an advocacy campaign.
type AdvocacyCampaignStatus string

const (
	AdvocacyStatusDraft    AdvocacyCampaignStatus = "draft"
	AdvocacyStatusApproved AdvocacyCampaignStatus = "approved"
	AdvocacyStatusArchived AdvocacyCampaignStatus = "archived"
)

// EmployeePersona represents target professional personas for copy variants.
type EmployeePersona string

const (
	PersonaEngineering   EmployeePersona = "engineering"
	PersonaProduct       EmployeePersona = "product"
	PersonaTalentCulture EmployeePersona = "talent_culture"
	PersonaGeneral       EmployeePersona = "general"
)

// AdvocacyCopyVariant represents a suggested post variant for employees.
type AdvocacyCopyVariant struct {
	VariantID     string          `json:"variant_id"`
	Persona       EmployeePersona `json:"persona"`
	Headline      string          `json:"headline"`
	SuggestedText string          `json:"suggested_text"`
	TargetTags    []string        `json:"target_tags,omitempty"`
}

// BrandGovernance encapsulates brand rules, allowed tags, and blacklisted keywords.
type BrandGovernance struct {
	Guidelines        string   `json:"guidelines"`
	AllowedHashtags   []string `json:"allowed_hashtags,omitempty"`
	ForbiddenKeywords []string `json:"forbidden_keywords,omitempty"`
}

// AdvocacyCampaign represents an employer advocacy campaign for team members.
type AdvocacyCampaign struct {
	CampaignID    string                 `json:"campaign_id"`
	TenantID      string                 `json:"tenant_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Governance    BrandGovernance        `json:"governance"`
	Variants      []AdvocacyCopyVariant  `json:"variants"`
	Status        AdvocacyCampaignStatus `json:"status"`
	ApprovalToken string                 `json:"approval_token,omitempty"`
	ApprovedBy    string                 `json:"approved_by,omitempty"`
	ApprovedAt    string                 `json:"approved_at,omitempty"`
	ShareCount    int                    `json:"share_count"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// EmployeeShareEvent records voluntary employee sharing for internal analytics.
type EmployeeShareEvent struct {
	ShareID        string `json:"share_id"`
	TenantID       string `json:"tenant_id"`
	WorkspaceID    string `json:"workspace_id"`
	CampaignID     string `json:"campaign_id"`
	VariantID      string `json:"variant_id"`
	EmployeeID     string `json:"employee_id"`
	CustomizedText string `json:"customized_text"`
	SharedAt       string `json:"shared_at"`
	Platform       string `json:"platform"`
	ShareMethod    string `json:"share_method"`
}

// CoordinatedEngagementRequest models an adversarial or prohibited pod engagement request.
type CoordinatedEngagementRequest struct {
	CampaignID                string   `json:"campaign_id"`
	TargetPostURN             string   `json:"target_post_urn"`
	Action                    string   `json:"action"`
	ParticipatingEmployeeIDs []string `json:"participating_employee_ids"`
	SyntheticComments         []string `json:"synthetic_comments,omitempty"`
}

// ShareAdvocacyResult returns clipboard formatted text and deep-link for assisted manual sharing.
type ShareAdvocacyResult struct {
	Success         bool   `json:"success"`
	ShareID         string `json:"share_id"`
	ClipboardText   string `json:"clipboard_text"`
	ComposeDeepLink string `json:"compose_deep_link"`
	Message         string `json:"message"`
}

// ValidateAdvocacyCampaign validates structural integrity of a campaign.
func ValidateAdvocacyCampaign(campaign *AdvocacyCampaign) error {
	if strings.TrimSpace(campaign.Title) == "" {
		return ErrCampaignTitleRequired
	}
	if len(campaign.Variants) == 0 {
		return ErrCampaignVariantsRequired
	}
	for _, v := range campaign.Variants {
		if strings.TrimSpace(v.SuggestedText) == "" {
			return fmt.Errorf("variant %s must have non-empty suggested_text", v.VariantID)
		}
	}
	return nil
}

// CheckBrandCompliance checks a given text against forbidden keywords in brand governance.
func CheckBrandCompliance(text string, gov BrandGovernance) []string {
	normalized := strings.ToLower(text)
	var violations []string
	for _, kw := range gov.ForbiddenKeywords {
		cleanKW := strings.TrimSpace(strings.ToLower(kw))
		if cleanKW != "" && strings.Contains(normalized, cleanKW) {
			violations = append(violations, kw)
		}
	}
	return violations
}

// RejectCoordinatedEngagementPod enforces the strict anti-pod / zero fake engagement policy.
func RejectCoordinatedEngagementPod(req CoordinatedEngagementRequest) error {
	// Any coordinated mutual auto-liking, automated employee syndication, or synthetic commenting is permanently barred.
	return ErrCoordinatedEngagementProhibited
}

// ComputeCampaignApprovalToken computes a deterministic HMAC-SHA256 signature for campaign state.
func ComputeCampaignApprovalToken(campaign AdvocacyCampaign, secret string) string {
	var canonical strings.Builder
	canonical.WriteString("tenant:")
	canonical.WriteString(campaign.TenantID)
	canonical.WriteString("|ws:")
	canonical.WriteString(campaign.WorkspaceID)
	canonical.WriteString("|campaign:")
	canonical.WriteString(campaign.CampaignID)
	canonical.WriteString("|title:")
	canonical.WriteString(campaign.Title)
	canonical.WriteString("|guidelines:")
	canonical.WriteString(campaign.Governance.Guidelines)

	// Sort and append forbidden keywords
	var keywords []string
	for _, kw := range campaign.Governance.ForbiddenKeywords {
		keywords = append(keywords, strings.TrimSpace(strings.ToLower(kw)))
	}
	sort.Strings(keywords)
	canonical.WriteString("|forbidden:")
	canonical.WriteString(strings.Join(keywords, ","))

	// Sort and append variant texts
	var variantTexts []string
	for _, v := range campaign.Variants {
		variantTexts = append(variantTexts, fmt.Sprintf("%s:%s", v.VariantID, strings.TrimSpace(v.SuggestedText)))
	}
	sort.Strings(variantTexts)
	canonical.WriteString("|variants:")
	canonical.WriteString(strings.Join(variantTexts, ";"))

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyCampaignApprovalToken checks if the campaign approval token matches the current campaign state.
func VerifyCampaignApprovalToken(campaign AdvocacyCampaign, secret string) bool {
	if campaign.ApprovalToken == "" {
		return false
	}
	expected := ComputeCampaignApprovalToken(campaign, secret)
	return hmac.Equal([]byte(campaign.ApprovalToken), []byte(expected))
}

// GenerateComposeDeepLink generates a native LinkedIn web compose share URL with prefilled text.
func GenerateComposeDeepLink(text string) string {
	baseURL := "https://www.linkedin.com/feed/?shareActive=true&text="
	return baseURL + url.QueryEscape(text)
}
