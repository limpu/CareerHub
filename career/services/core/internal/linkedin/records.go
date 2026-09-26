package linkedin

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidTenantOrWorkspace = errors.New("records: tenant_id and workspace_id are mandatory for tenant isolation (AT-011, AT-012)")
	ErrForbiddenBiometricField  = errors.New("records: forbidden biometric or sensitive unpermitted personal data detected (FND-009, FND-012)")
	ErrRecordNotFound           = errors.New("records: record not found in repository")
	ErrCrossTenantAccessDenied  = errors.New("records: cross-tenant access strictly denied (AT-012)")
	ErrInvalidRecordData        = errors.New("records: record payload contains invalid or missing required identifiers")
)

// PersonRecord represents a normalized, privacy-compliant snapshot of a professional profile (SRC-L3).
type PersonRecord struct {
	RecordID         string    `json:"record_id"`
	WorkspaceID      string    `json:"workspace_id"`
	TenantID         string    `json:"tenant_id"`
	EntityType       string    `json:"entity_type"` // always "person"
	VanitySlug       string    `json:"vanity_slug"`
	FullName         string    `json:"full_name"`
	Headline         string    `json:"headline"`
	CurrentCompany   string    `json:"current_company"`
	CurrentRole      string    `json:"current_role"`
	Location         string    `json:"location"`
	ConnectionTier   string    `json:"connection_tier"` // e.g. "1st", "2nd", "3rd", "out_of_network"
	PublicProfileURL string    `json:"public_profile_url"`
	Skills           []string  `json:"skills"`
	Version          int       `json:"version"`
	ObservedAt       time.Time `json:"observed_at"`
}

// CompanyRecord represents an audited organizational profile snapshot (SRC-L4).
type CompanyRecord struct {
	RecordID      string    `json:"record_id"`
	WorkspaceID   string    `json:"workspace_id"`
	TenantID      string    `json:"tenant_id"`
	EntityType    string    `json:"entity_type"` // always "company"
	UniversalName string    `json:"universal_name"`
	CompanyName   string    `json:"company_name"`
	Domain        string    `json:"domain"`
	Industry      string    `json:"industry"`
	SizeTier      string    `json:"size_tier"`
	Headquarters  string    `json:"headquarters"`
	Specialties   []string  `json:"specialties"`
	FollowerCount int       `json:"follower_count"`
	Verified      bool      `json:"verified"`
	Version       int       `json:"version"`
	ObservedAt    time.Time `json:"observed_at"`
}

// JobRecord represents an audited job opening record (LI-04, SRC-L3).
type JobRecord struct {
	RecordID           string    `json:"record_id"`
	WorkspaceID        string    `json:"workspace_id"`
	TenantID           string    `json:"tenant_id"`
	EntityType         string    `json:"entity_type"` // always "job"
	JobID              string    `json:"job_id"`
	Title              string    `json:"title"`
	CompanyName        string    `json:"company_name"`
	WorkplaceType      string    `json:"workplace_type"` // "Remote", "Hybrid", "On-site"
	Location           string    `json:"location"`
	EmploymentType     string    `json:"employment_type"` // "Full-time", "Contract", etc.
	DescriptionSnippet string    `json:"description_snippet"`
	ApplicantCount     int       `json:"applicant_count"`
	PostedAt           time.Time `json:"posted_at"`
	Version            int       `json:"version"`
	ObservedAt         time.Time `json:"observed_at"`
}

// PostRecord represents a social/UGC post record (LI-04, SRC-L4).
type PostRecord struct {
	RecordID       string    `json:"record_id"`
	WorkspaceID    string    `json:"workspace_id"`
	TenantID       string    `json:"tenant_id"`
	EntityType     string    `json:"entity_type"` // always "post"
	URN            string    `json:"urn"`
	AuthorName     string    `json:"author_name"`
	AuthorVanity   string    `json:"author_vanity"`
	Commentary     string    `json:"commentary"`
	MediaType      string    `json:"media_type"` // "none", "image", "article", "video"
	ReactionsCount int       `json:"reactions_count"`
	CommentsCount  int       `json:"comments_count"`
	SharesCount    int       `json:"shares_count"`
	PublishedAt    time.Time `json:"published_at"`
	Version        int       `json:"version"`
	ObservedAt     time.Time `json:"observed_at"`
}

// RecordFilter allows querying records with tenant scoping (AT-011, AT-012).
type RecordFilter struct {
	WorkspaceID string `json:"workspace_id"`
	TenantID    string `json:"tenant_id"`
	Query       string `json:"query,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

// ValidatePermittedFields guards against storing sensitive or prohibited biometric data (FND-009, FND-012).
func ValidatePermittedFields(data map[string]interface{}) error {
	forbiddenSubstrings := []string{
		"face_geometry", "biometric", "iris_scan", "voiceprint", "social_security",
		"national_id", "medical_record", "credit_card", "fingerprint",
	}

	for key, val := range data {
		lowerKey := strings.ToLower(key)
		for _, forbidden := range forbiddenSubstrings {
			if strings.Contains(lowerKey, forbidden) {
				return fmt.Errorf("%w: prohibited field key '%s'", ErrForbiddenBiometricField, key)
			}
		}

		if strVal, ok := val.(string); ok {
			lowerVal := strings.ToLower(strVal)
			for _, forbidden := range forbiddenSubstrings {
				if strings.Contains(lowerVal, forbidden) {
					return fmt.Errorf("%w: prohibited sensitive content detected in '%s'", ErrForbiddenBiometricField, key)
				}
			}
		}
	}
	return nil
}

// ValidateTenantScope checks that tenant and workspace identities are present (AT-011, AT-012).
func ValidateTenantScope(workspaceID, tenantID string) error {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(tenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	return nil
}

// CheckTenantAccess ensures the requesting tenant matches the resource tenant (AT-012).
func CheckTenantAccess(resourceTenantID, requestingTenantID string) error {
	if resourceTenantID == "" || requestingTenantID == "" || resourceTenantID != requestingTenantID {
		return ErrCrossTenantAccessDenied
	}
	return nil
}
