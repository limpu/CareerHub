package career

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProfileNotFound            = errors.New("master career profile not found")
	ErrPrivateProfileAccessDenied = errors.New("private career profile: workspace admin access not granted by owner (REQ-018)")
	ErrInvalidConsent             = errors.New("data processing consent is required to process career facts")
	ErrSectionNotFound            = errors.New("career profile section not found")
)

// ContactInfo holds verified personal contact and biographical data.
// Invariant AT-003: Unknown or missing fields remain blank; never guessed.
type ContactInfo struct {
	FullName          string `json:"full_name"`
	Email             string `json:"email"`
	Phone             string `json:"phone,omitempty"`
	Location          string `json:"location,omitempty"`
	Headline          string `json:"headline,omitempty"`
	Summary           string `json:"summary,omitempty"`
	Citizenship       string `json:"citizenship,omitempty"`
	WorkAuthorization string `json:"work_authorization,omitempty"`
}

// ProfileLink represents social and portfolio URLs.
type ProfileLink struct {
	ID        string `json:"id"`
	Label     string `json:"label"` // e.g. "GitHub", "Portfolio", "LinkedIn"
	URL       string `json:"url"`
	LinkType  string `json:"link_type"` // "github", "linkedin", "portfolio", "blog", "other"
	Confirmed bool   `json:"confirmed"`
}

// ExperienceItem represents a professional role or position.
type ExperienceItem struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Company     string     `json:"company"`
	Location    string     `json:"location,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	IsCurrent   bool       `json:"is_current"`
	Description string     `json:"description,omitempty"`
	Highlights  []string   `json:"highlights"`
	SkillsUsed  []string   `json:"skills_used"`
	Confirmed   bool       `json:"confirmed"`
}

// EducationItem represents an academic degree or diploma.
type EducationItem struct {
	ID           string     `json:"id"`
	Institution  string     `json:"institution"`
	Degree       string     `json:"degree"`
	FieldOfStudy string     `json:"field_of_study,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Grade        string     `json:"grade,omitempty"`
	Highlights   []string   `json:"highlights"`
	Confirmed    bool       `json:"confirmed"`
}

// SkillProficiency defines validated skill tiers.
type SkillProficiency string

const (
	ProficiencyUnspecified  SkillProficiency = "unspecified"
	ProficiencyBeginner     SkillProficiency = "beginner"
	ProficiencyIntermediate SkillProficiency = "intermediate"
	ProficiencyAdvanced     SkillProficiency = "advanced"
	ProficiencyExpert       SkillProficiency = "expert"
)

// SkillItem represents a distinct technical, domain, or soft skill.
type SkillItem struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	Category          string           `json:"category"` // "technical", "soft", "tool", "domain"
	Proficiency       SkillProficiency `json:"proficiency"`
	YearsOfExperience int              `json:"years_of_experience,omitempty"`
	Confirmed         bool             `json:"confirmed"`
}

// ProjectItem represents an engineering or professional project.
type ProjectItem struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Role         string     `json:"role,omitempty"`
	URL          string     `json:"url,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Description  string     `json:"description"`
	Highlights   []string   `json:"highlights"`
	Technologies []string   `json:"technologies"`
	Confirmed    bool       `json:"confirmed"`
}

// CertificateItem represents an industry license or certification.
type CertificateItem struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Issuer        string     `json:"issuer"`
	IssueDate     *time.Time `json:"issue_date,omitempty"`
	ExpiryDate    *time.Time `json:"expiry_date,omitempty"`
	CredentialID  string     `json:"credential_id,omitempty"`
	CredentialURL string     `json:"credential_url,omitempty"`
	Confirmed     bool       `json:"confirmed"`
}

// LanguageProficiency defines standard language fluency levels.
type LanguageProficiency string

const (
	LangBasic          LanguageProficiency = "basic"
	LangConversational LanguageProficiency = "conversational"
	LangProfessional   LanguageProficiency = "professional"
	LangFluent         LanguageProficiency = "fluent"
	LangNative         LanguageProficiency = "native"
)

// LanguageItem represents a spoken or written language.
type LanguageItem struct {
	ID          string              `json:"id"`
	Language    string              `json:"language"`
	Proficiency LanguageProficiency `json:"proficiency"`
	Confirmed   bool                `json:"confirmed"`
}

// UserCareerConsent tracks explicit user consent for career data processing (REQ-023).
type UserCareerConsent struct {
	DataProcessingConsent    bool      `json:"data_processing_consent"`
	JobMatchingConsent       bool      `json:"job_matching_consent"`
	ThirdPartySharingConsent bool      `json:"third_party_sharing_consent"`
	ConsentTimestamp         time.Time `json:"consent_timestamp"`
	ConsentVersion           string    `json:"consent_version"` // e.g. "v1.0"
}

// CareerPrivacySettings defines access controls and workspace admin boundaries (REQ-018).
// Invariant REQ-018: ShareWithWorkspaceAdmin is FALSE by default. Admin access to workspace does not imply access to personal profile.
type CareerPrivacySettings struct {
	ProfileVisibility        PrivacyTier `json:"profile_visibility"`
	AllowRecruiterView       bool        `json:"allow_recruiter_view"`
	ShareWithWorkspaceAdmin  bool        `json:"share_with_workspace_admin"` // default false
	AnonymizeBeforeJobSearch bool        `json:"anonymize_before_job_search"`
}

// ProfileFieldAudit records granular field-level modification history.
type ProfileFieldAudit struct {
	ID        string    `json:"id"`
	Field     string    `json:"field"`
	OldValue  string    `json:"old_value,omitempty"`
	NewValue  string    `json:"new_value,omitempty"`
	ChangedBy string    `json:"changed_by"`
	ChangedAt time.Time `json:"changed_at"`
	Reason    string    `json:"reason,omitempty"`
}

// CompletenessReport measures profile quality without fabricating missing facts.
type CompletenessReport struct {
	Score              float64  `json:"score"` // 0.0 to 100.0
	MissingSections    []string `json:"missing_sections"`
	PopulatedCount     int      `json:"populated_count"`
	TotalSectionsCount int      `json:"total_sections_count"`
}

// MasterCareerProfile is the canonical user-confirmed career record (REQ-002, REQ-003).
// Governs master resume generation, job matching, and privacy isolation.
type MasterCareerProfile struct {
	ID           string                `json:"id"`
	UserID       string                `json:"user_id"`
	Contact      ContactInfo           `json:"contact"`
	Links        []ProfileLink         `json:"links"`
	Experiences  []ExperienceItem      `json:"experiences"`
	Education    []EducationItem       `json:"education"`
	Skills       []SkillItem           `json:"skills"`
	Projects     []ProjectItem         `json:"projects"`
	Certificates []CertificateItem     `json:"certificates"`
	Languages    []LanguageItem        `json:"languages"`
	Consent      UserCareerConsent     `json:"consent"`
	Privacy      CareerPrivacySettings `json:"privacy"`
	AuditTrail   []ProfileFieldAudit   `json:"audit_trail"`
	Completeness CompletenessReport    `json:"completeness"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

// NewMasterCareerProfile creates an initial blank profile with default strict privacy (REQ-018).
func NewMasterCareerProfile(userID string) *MasterCareerProfile {
	now := time.Now()
	p := &MasterCareerProfile{
		ID:     uuid.New().String(),
		UserID: userID,
		Contact: ContactInfo{
			FullName: "",
			Email:    "",
		},
		Links:        make([]ProfileLink, 0),
		Experiences:  make([]ExperienceItem, 0),
		Education:    make([]EducationItem, 0),
		Skills:       make([]SkillItem, 0),
		Projects:     make([]ProjectItem, 0),
		Certificates: make([]CertificateItem, 0),
		Languages:    make([]LanguageItem, 0),
		Consent: UserCareerConsent{
			DataProcessingConsent:    true,
			JobMatchingConsent:       true,
			ThirdPartySharingConsent: false,
			ConsentTimestamp:         now,
			ConsentVersion:           "v1.0",
		},
		Privacy: CareerPrivacySettings{
			ProfileVisibility:        PrivacyOwnerPrivate,
			AllowRecruiterView:       false,
			ShareWithWorkspaceAdmin:  false, // REQ-018: Strictly isolated from workspace admin by default
			AnonymizeBeforeJobSearch: false,
		},
		AuditTrail: make([]ProfileFieldAudit, 0),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	p.Completeness = p.CalculateCompleteness()
	return p
}

// CalculateCompleteness evaluates completeness without fabricating facts (AT-003).
func (p *MasterCareerProfile) CalculateCompleteness() CompletenessReport {
	var missing []string
	populated := 0
	total := 8

	if p.Contact.FullName != "" && p.Contact.Email != "" {
		populated++
	} else {
		missing = append(missing, "contact")
	}

	if len(p.Links) > 0 {
		populated++
	} else {
		missing = append(missing, "links")
	}

	if len(p.Experiences) > 0 {
		populated++
	} else {
		missing = append(missing, "experience")
	}

	if len(p.Education) > 0 {
		populated++
	} else {
		missing = append(missing, "education")
	}

	if len(p.Skills) > 0 {
		populated++
	} else {
		missing = append(missing, "skills")
	}

	if len(p.Projects) > 0 {
		populated++
	} else {
		missing = append(missing, "projects")
	}

	if len(p.Certificates) > 0 {
		populated++
	} else {
		missing = append(missing, "certificates")
	}

	if len(p.Languages) > 0 {
		populated++
	} else {
		missing = append(missing, "languages")
	}

	score := (float64(populated) / float64(total)) * 100.0

	return CompletenessReport{
		Score:              score,
		MissingSections:    missing,
		PopulatedCount:     populated,
		TotalSectionsCount: total,
	}
}

// RecordAudit logs a field modification.
func (p *MasterCareerProfile) RecordAudit(field, oldVal, newVal, changedBy, reason string) {
	p.AuditTrail = append(p.AuditTrail, ProfileFieldAudit{
		ID:        uuid.New().String(),
		Field:     field,
		OldValue:  oldVal,
		NewValue:  newVal,
		ChangedBy: changedBy,
		ChangedAt: time.Now(),
		Reason:    reason,
	})
	p.UpdatedAt = time.Now()
}
