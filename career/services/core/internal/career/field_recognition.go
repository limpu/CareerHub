package career

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// FieldType defines the HTML/UI input type of an application question
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeTextarea FieldType = "textarea"
	FieldTypeNumber   FieldType = "number"
	FieldTypeSelect   FieldType = "select"
	FieldTypeRadio    FieldType = "radio"
	FieldTypeCheckbox FieldType = "checkbox"
	FieldTypeFile     FieldType = "file"
)

// FieldCategory classifies the semantic meaning of the field for candidate profile mapping
type FieldCategory string

const (
	CategoryContact           FieldCategory = "contact"
	CategoryExperience        FieldCategory = "experience"
	CategoryEducation         FieldCategory = "education"
	CategorySkills            FieldCategory = "skills"
	CategoryWorkAuthorization FieldCategory = "work_authorization"
	CategorySponsorship       FieldCategory = "sponsorship"
	CategorySalary            FieldCategory = "salary"
	CategoryNoticePeriod      FieldCategory = "notice_period"
	CategoryPortfolio         FieldCategory = "portfolio"
	CategoryDemographic       FieldCategory = "demographic"
	CategoryCustom            FieldCategory = "custom"
)

// ConditionOperator defines how parent field answers are evaluated
type ConditionOperator string

const (
	OpEquals    ConditionOperator = "equals"
	OpNotEquals ConditionOperator = "not_equals"
	OpContains  ConditionOperator = "contains"
	OpIn        ConditionOperator = "in"
)

// ConditionAction defines how the child field behaves when the condition matches
type ConditionAction string

const (
	ActionShow    ConditionAction = "show"
	ActionHide    ConditionAction = "hide"
	ActionEnable  ConditionAction = "enable"
	ActionDisable ConditionAction = "disable"
	ActionRequire ConditionAction = "require"
)

// ConditionState represents the computed runtime state of a field in a multi-step form
type ConditionState string

const (
	StateActive   ConditionState = "active"
	StateHidden   ConditionState = "hidden"
	StateDisabled ConditionState = "disabled"
)

// FieldCondition represents conditional dependency between fields (CAR-12)
type FieldCondition struct {
	ParentFieldID string            `json:"parent_field_id"`
	Operator      ConditionOperator `json:"operator"`
	ExpectedValue interface{}       `json:"expected_value"`
	Action        ConditionAction   `json:"action"`
}

// FieldValidation defines validation constraints for an input
type FieldValidation struct {
	MinLength         int      `json:"min_length,omitempty"`
	MaxLength         int      `json:"max_length,omitempty"`
	MinValue          *float64 `json:"min_value,omitempty"`
	MaxValue          *float64 `json:"max_value,omitempty"`
	Pattern           string   `json:"pattern,omitempty"`
	AllowedExtensions []string `json:"allowed_extensions,omitempty"`
}

// FormField defines an extracted or recognized question field from an application modal
type FormField struct {
	FieldID         string           `json:"field_id"`
	Label           string           `json:"label"`
	SanitizedLabel  string           `json:"sanitized_label"`
	Type            FieldType        `json:"type"`
	Required        bool             `json:"required"`
	Placeholder     string           `json:"placeholder,omitempty"`
	HelpText        string           `json:"help_text,omitempty"`
	Options         []string         `json:"options,omitempty"`
	Category        FieldCategory    `json:"category"`
	TargetSubKey    string           `json:"target_sub_key,omitempty"` // e.g. "go", "python", "phone", "email"
	IsSensitive     bool             `json:"is_sensitive"`
	Validation      *FieldValidation `json:"validation,omitempty"`
	Condition       *FieldCondition  `json:"condition,omitempty"`
	SecurityFlags   []string         `json:"security_flags,omitempty"`
}

// FormSection groups fields in multi-step application wizards
type FormSection struct {
	SectionID   string      `json:"section_id"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Fields      []FormField `json:"fields"`
}

// RecognizedForm represents a fully parsed application form or modal snapshot
type RecognizedForm struct {
	FormID      string        `json:"form_id"`
	Title       string        `json:"title"`
	SourceURL   string        `json:"source_url,omitempty"`
	Provider    string        `json:"provider,omitempty"` // e.g. "linkedin_easy_apply", "greenhouse", "lever", "ashby"
	Sections    []FormSection `json:"sections"`
	Fields      []FormField   `json:"fields"`
}

// FieldResolution records the grounded value, confidence, and provenance for each recognized field
type FieldResolution struct {
	FieldID          string         `json:"field_id"`
	Label            string         `json:"label"`
	Type             FieldType      `json:"type"`
	Category         FieldCategory  `json:"category"`
	ResolvedValue    interface{}    `json:"resolved_value,omitempty"`
	NeedsInput       bool           `json:"needs_input"`
	Confidence       float64        `json:"confidence"` // 0.0 to 1.0
	Provenance       string         `json:"provenance"` // e.g. "profile.contact.email", "unconfirmed_needs_input"
	ResolutionReason string         `json:"resolution_reason"`
	ConditionState   ConditionState `json:"condition_state"`
	SecurityFlags    []string       `json:"security_flags,omitempty"`
}

// RecognitionRule matches field labels across languages and structures (SRC-C4, SRC-C6)
type RecognitionRule struct {
	Category     FieldCategory
	SubKey       string
	FieldType    FieldType
	IsSensitive  bool
	Keywords     []string
	RegexPattern *regexp.Regexp
}

// Multilingual keywords catalog for pure-logic matching (English, German, French, Spanish)
var defaultRecognitionRules = []RecognitionRule{
	// Contact: Full Name
	{
		Category:  CategoryContact,
		SubKey:    "name",
		FieldType: FieldTypeText,
		Keywords: []string{
			"full name", "legal name", "full legal name", "first and last name", "complete name", "candidate name", "your name",
			"vollständiger name", "vor- und nachname", "nom et prénom", "nom complet",
			"nombre completo", "nombre y apellidos",
		},
	},
	// Contact: Email
	{
		Category:  CategoryContact,
		SubKey:    "email",
		FieldType: FieldTypeText,
		Keywords: []string{
			"email", "e-mail", "email address", "e-mail-adresse", "adresse e-mail", "courriel", "correo electrónico", "email de contacto",
		},
	},
	// Contact: Phone
	{
		Category:  CategoryContact,
		SubKey:    "phone",
		FieldType: FieldTypeText,
		Keywords: []string{
			"phone", "mobile", "telephone", "phone number", "telefonnummer", "handynummer", "numéro de téléphone", "téléphone", "número de teléfono", "móvil",
		},
	},
	// Work Authorization (Sensitive)
	{
		Category:    CategoryWorkAuthorization,
		FieldType:   FieldTypeRadio,
		IsSensitive: true,
		Keywords: []string{
			"legally authorized to work", "legal authorization", "right to work", "work permit",
			"arbeitserlaubnis", "gesetzliche arbeitsberechtigung", "autorisation de travail", "légalement autorisé à travailler",
			"autorización legal para trabajar", "permiso de trabajo",
		},
	},
	// Visa Sponsorship (Sensitive)
	{
		Category:    CategorySponsorship,
		FieldType:   FieldTypeRadio,
		IsSensitive: true,
		Keywords: []string{
			"require sponsorship", "visa sponsorship", "need sponsorship", "require a visa",
			"visum-sponsoring", "visumsponsoring", "parrainage de visa", "besoin d'un parrainage",
			"patrocinio de visa", "necesita patrocinio",
		},
	},
	// Salary Expectation (Sensitive)
	{
		Category:    CategorySalary,
		FieldType:   FieldTypeNumber,
		IsSensitive: true,
		Keywords: []string{
			"salary expectation", "expected annual base", "compensation expectation", "desired salary", "target comp",
			"gehaltserwartung", "jahresgehalt", "wunschgehalt", "prétention salariale", "rémunération souhaitée",
			"expectativa salarial", "sueldo deseado", "salario esperado",
		},
	},
	// Notice Period / Availability
	{
		Category:  CategoryNoticePeriod,
		FieldType: FieldTypeText,
		Keywords: []string{
			"notice period", "earliest start date", "availability", "when can you start",
			"kündigungsfrist", "frühestmögliches startdatum", "verfügbarkeit", "préavis", "date de début",
			"disponibilité", "período de preaviso", "disponibilidad", "fecha de inicio",
		},
	},
	// Portfolio / Links
	{
		Category:  CategoryPortfolio,
		SubKey:    "github",
		FieldType: FieldTypeText,
		Keywords: []string{
			"github", "github profile", "github url", "github repo",
		},
	},
	{
		Category:  CategoryPortfolio,
		SubKey:    "linkedin",
		FieldType: FieldTypeText,
		Keywords: []string{
			"linkedin", "linkedin profile", "linkedin url",
		},
	},
	{
		Category:  CategoryPortfolio,
		SubKey:    "portfolio",
		FieldType: FieldTypeText,
		Keywords: []string{
			"portfolio", "portfolio url", "personal website", "persönliche website", "site personnel", "sitio web personal",
		},
	},
	// File Upload: Resume
	{
		Category:  CategoryContact,
		SubKey:    "resume",
		FieldType: FieldTypeFile,
		Keywords: []string{
			"resume", "cv", "curriculum vitae", "lebenslauf", "mon cv", "adjuntar cv",
		},
	},
	// File Upload: Cover Letter
	{
		Category:  CategoryContact,
		SubKey:    "cover_letter",
		FieldType: FieldTypeFile,
		Keywords: []string{
			"cover letter", "anschreiben", "lettre de motivation", "carta de presentación",
		},
	},
}

// Suspicious injection patterns for AT-019 defense
var adversarialPromptPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignore\s+(all\s+)?previous\s+instructions`),
	regexp.MustCompile(`(?i)system\s+update\s*:`),
	regexp.MustCompile(`(?i)unrestricted\s+developer`),
	regexp.MustCompile(`(?i)call\s+tool\s*:`),
	regexp.MustCompile(`(?i)execute\s+tool\s*:`),
	regexp.MustCompile(`(?i)function\s+call\s*:`),
	regexp.MustCompile(`(?i)bypass\s+approval`),
	regexp.MustCompile(`(?i)skip\s+human\s+review`),
	regexp.MustCompile(`(?i)reveal\s+api_key`),
	regexp.MustCompile(`(?i)</?system_instruction>`),
	regexp.MustCompile(`(?i)</?untrusted_context`),
	regexp.MustCompile(`(?i)\[SYSTEM\]`),
}

// FieldRecognitionEngine orchestrates question classification and zero-fabrication matching
type FieldRecognitionEngine struct {
	rules []RecognitionRule
}

// NewFieldRecognitionEngine initializes the engine with default rules
func NewFieldRecognitionEngine() *FieldRecognitionEngine {
	return &FieldRecognitionEngine{
		rules: defaultRecognitionRules,
	}
}

// SanitizePromptInjection detects and neutralizes adversarial prompt injection in form text (AT-019)
func (e *FieldRecognitionEngine) SanitizePromptInjection(input string) (string, []string) {
	var flags []string
	clean := input

	for _, pattern := range adversarialPromptPatterns {
		if pattern.MatchString(clean) {
			flags = append(flags, fmt.Sprintf("adversarial_pattern_neutralized:%s", pattern.String()))
			clean = pattern.ReplaceAllString(clean, "[REDACTED_INJECTION_PAYLOAD]")
		}
	}

	// Delimiter escaping neutralization
	clean = strings.ReplaceAll(clean, "<", "&lt;")
	clean = strings.ReplaceAll(clean, ">", "&gt;")

	return strings.TrimSpace(clean), flags
}

// NormalizeText normalizes whitespace, accents, and lowercase for robust matching
func NormalizeText(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "?", "")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	words := strings.Fields(s)
	return strings.Join(words, " ")
}

// RecognizeField analyzes a field's label and properties to identify its category and type
func (e *FieldRecognitionEngine) RecognizeField(field FormField) FormField {
	sanitizedLabel, flags := e.SanitizePromptInjection(field.Label)
	field.SanitizedLabel = sanitizedLabel
	if len(flags) > 0 {
		field.SecurityFlags = append(field.SecurityFlags, flags...)
	}

	norm := NormalizeText(sanitizedLabel)

	// Check for programming language / tech skill questions (e.g. "How many years of work experience do you have with Go?")
	skillRegex := regexp.MustCompile(`(?i)(?:years of (?:work )?experience|berufserfahrung|années d'expérience|años de experiencia).*(?:with|in|mit|avec|con)\s+([a-zA-Z0-9#+.]+)`)
	if matches := skillRegex.FindStringSubmatch(sanitizedLabel); len(matches) > 1 {
		field.Category = CategoryExperience
		field.TargetSubKey = strings.ToLower(strings.TrimSpace(matches[1]))
		if field.Type == "" {
			field.Type = FieldTypeNumber
		}
		return field
	}

	// Match against standard rule catalog
	for _, rule := range e.rules {
		for _, kw := range rule.Keywords {
			if strings.Contains(norm, NormalizeText(kw)) {
				if field.Category == "" || field.Category == CategoryCustom {
					field.Category = rule.Category
				}
				if field.TargetSubKey == "" {
					field.TargetSubKey = rule.SubKey
				}
				if field.Type == "" {
					field.Type = rule.FieldType
				}
				if rule.IsSensitive {
					field.IsSensitive = true
				}
				return field
			}
		}
	}

	// Default fallback for unrecognized fields
	if field.Category == "" {
		field.Category = CategoryCustom
	}
	if field.Type == "" {
		field.Type = FieldTypeText
	}

	return field
}

// RecognizeForm processes all fields in a form
func (e *FieldRecognitionEngine) RecognizeForm(form *RecognizedForm) *RecognizedForm {
	if form == nil {
		return nil
	}

	for i := range form.Fields {
		form.Fields[i] = e.RecognizeField(form.Fields[i])
	}

	for s := range form.Sections {
		for f := range form.Sections[s].Fields {
			form.Sections[s].Fields[f] = e.RecognizeField(form.Sections[s].Fields[f])
		}
	}

	return form
}

// EvaluateFormConditions evaluates runtime states (active, hidden, disabled) based on current answers (CAR-12)
func (e *FieldRecognitionEngine) EvaluateFormConditions(fields []FormField, currentAnswers map[string]interface{}) map[string]ConditionState {
	states := make(map[string]ConditionState)

	// Initial pass: default all fields to active
	for _, field := range fields {
		states[field.FieldID] = StateActive
	}

	// Evaluate conditions
	for _, field := range fields {
		if field.Condition == nil {
			continue
		}

		parentVal, hasAnswer := currentAnswers[field.Condition.ParentFieldID]
		if !hasAnswer || parentVal == nil {
			// Parent has not been answered yet -> child conditional field is hidden or disabled
			switch field.Condition.Action {
			case ActionShow:
				states[field.FieldID] = StateHidden
			case ActionEnable:
				states[field.FieldID] = StateDisabled
			case ActionRequire:
				// If parent not matching, child is not required / inactive
				states[field.FieldID] = StateHidden
			default:
				states[field.FieldID] = StateHidden
			}
			continue
		}

		// Evaluate operator
		matches := false
		switch field.Condition.Operator {
		case OpEquals:
			matches = strings.EqualFold(fmt.Sprintf("%v", parentVal), fmt.Sprintf("%v", field.Condition.ExpectedValue))
		case OpNotEquals:
			matches = !strings.EqualFold(fmt.Sprintf("%v", parentVal), fmt.Sprintf("%v", field.Condition.ExpectedValue))
		case OpContains:
			matches = strings.Contains(strings.ToLower(fmt.Sprintf("%v", parentVal)), strings.ToLower(fmt.Sprintf("%v", field.Condition.ExpectedValue)))
		case OpIn:
			if slice, ok := field.Condition.ExpectedValue.([]string); ok {
				for _, item := range slice {
					if strings.EqualFold(fmt.Sprintf("%v", parentVal), item) {
						matches = true
						break
					}
				}
			}
		}

		if matches {
			switch field.Condition.Action {
			case ActionShow, ActionRequire, ActionEnable:
				states[field.FieldID] = StateActive
			case ActionHide:
				states[field.FieldID] = StateHidden
			case ActionDisable:
				states[field.FieldID] = StateDisabled
			}
		} else {
			// Condition did not match
			switch field.Condition.Action {
			case ActionShow, ActionRequire:
				states[field.FieldID] = StateHidden
			case ActionEnable:
				states[field.FieldID] = StateDisabled
			case ActionHide, ActionDisable:
				states[field.FieldID] = StateActive
			}
		}
	}

	return states
}

// ResolveFieldAnswers resolves form questions strictly against confirmed user profile (AT-003: Zero fabrication)
func (e *FieldRecognitionEngine) ResolveFieldAnswers(
	form *RecognizedForm,
	profile *MasterCareerProfile,
	preferences *CareerPreferences,
	currentAnswers map[string]interface{},
) []FieldResolution {
	if form == nil {
		return nil
	}

	// Flatten fields from sections if needed
	var allFields []FormField
	if len(form.Fields) > 0 {
		allFields = form.Fields
	} else {
		for _, sec := range form.Sections {
			allFields = append(allFields, sec.Fields...)
		}
	}

	// 1. First ensure all fields are recognized
	for i := range allFields {
		allFields[i] = e.RecognizeField(allFields[i])
	}

	// 2. Evaluate conditional field states
	conditionStates := e.EvaluateFormConditions(allFields, currentAnswers)

	var resolutions []FieldResolution

	for _, field := range allFields {
		condState := conditionStates[field.FieldID]
		res := FieldResolution{
			FieldID:        field.FieldID,
			Label:          field.Label,
			Type:           field.Type,
			Category:       field.Category,
			ConditionState: condState,
			SecurityFlags:  field.SecurityFlags,
		}

		// If field is hidden or disabled, it should not demand input from user
		if condState == StateHidden || condState == StateDisabled {
			res.NeedsInput = false
			res.Confidence = 1.0
			res.Provenance = "conditional_inactive"
			res.ResolutionReason = fmt.Sprintf("Field is %s based on conditional dependency", condState)
			resolutions = append(resolutions, res)
			continue
		}

		// Check if user already provided an explicit answer in current session
		if userVal, exists := currentAnswers[field.FieldID]; exists && userVal != nil && fmt.Sprintf("%v", userVal) != "" {
			res.ResolvedValue = userVal
			res.NeedsInput = false
			res.Confidence = 1.0
			res.Provenance = "session.user_provided"
			res.ResolutionReason = "User explicitly confirmed answer in this session"
			resolutions = append(resolutions, res)
			continue
		}

		// Ground against candidate profile and preferences with ZERO FABRICATION (AT-003)
		switch field.Category {
		case CategoryContact:
			if profile != nil {
				switch field.TargetSubKey {
				case "name":
					if profile.Contact.FullName != "" {
						res.ResolvedValue = profile.Contact.FullName
						res.NeedsInput = false
						res.Confidence = 1.0
						res.Provenance = "profile.contact.full_name"
						res.ResolutionReason = "Matched candidate full name from confirmed profile"
					}
				case "email":
					if profile.Contact.Email != "" {
						res.ResolvedValue = profile.Contact.Email
						res.NeedsInput = false
						res.Confidence = 1.0
						res.Provenance = "profile.contact.email"
						res.ResolutionReason = "Matched candidate email from confirmed profile"
					}
				case "phone":
					if profile.Contact.Phone != "" {
						res.ResolvedValue = profile.Contact.Phone
						res.NeedsInput = false
						res.Confidence = 1.0
						res.Provenance = "profile.contact.phone"
						res.ResolutionReason = "Matched candidate phone from confirmed profile"
					}
				}
			}

		case CategoryExperience:
			// Match specific skill years
			if profile != nil && field.TargetSubKey != "" {
				skillTarget := strings.ToLower(field.TargetSubKey)
				found := false
				for _, sk := range profile.Skills {
					if strings.EqualFold(sk.Name, skillTarget) {
						res.ResolvedValue = sk.YearsOfExperience
						res.NeedsInput = false
						res.Confidence = 0.95
						res.Provenance = fmt.Sprintf("profile.skills.%s", sk.Name)
						res.ResolutionReason = fmt.Sprintf("Grounded skill experience from confirmed skill item (%d years)", sk.YearsOfExperience)
						found = true
						break
					}
				}
				// AT-003 INVARIANT: If skill not confirmed in candidate profile, NEVER GUESS '5' OR ARBITRARY DEFAULTS
				if !found {
					res.NeedsInput = true
					res.Confidence = 0.0
					res.Provenance = "unconfirmed_needs_input"
					res.ResolutionReason = fmt.Sprintf("Candidate has no confirmed experience record for '%s'. Under AT-003 zero-fabrication rules, this requires candidate input.", field.TargetSubKey)
				}
			} else {
				res.NeedsInput = true
				res.Confidence = 0.0
				res.Provenance = "unconfirmed_needs_input"
				res.ResolutionReason = "General experience question requires candidate confirmation."
			}

		case CategoryWorkAuthorization:
			// Invariant AT-003: Check confirmed preferences
			if preferences != nil && preferences.Sponsorship != SponsorshipUnspecified {
				if preferences.Sponsorship == SponsorshipAuthorizedNoSponsorship {
					res.ResolvedValue = "Yes"
					res.NeedsInput = false
					res.Confidence = 0.95
					res.Provenance = "preferences.work_authorization"
					res.ResolutionReason = "Candidate confirmed no sponsorship is required (legally authorized)"
				} else {
					res.ResolvedValue = "No"
					res.NeedsInput = false
					res.Confidence = 0.90
					res.Provenance = "preferences.work_authorization"
					res.ResolutionReason = "Candidate indicated sponsorship is required"
				}
			} else {
				// AT-003: Fail closed to needs_input
				res.NeedsInput = true
				res.Confidence = 0.0
				res.Provenance = "unconfirmed_needs_input"
				res.ResolutionReason = "Work authorization status is unspecified in preferences. Failing closed to needs_input (AT-003)."
			}

		case CategorySponsorship:
			if preferences != nil && preferences.Sponsorship != SponsorshipUnspecified {
				if preferences.Sponsorship == SponsorshipRequiresSponsorship {
					res.ResolvedValue = "Yes"
					res.NeedsInput = false
					res.Confidence = 0.95
					res.Provenance = "preferences.sponsorship"
					res.ResolutionReason = "Candidate confirmed sponsorship is required"
				} else {
					res.ResolvedValue = "No"
					res.NeedsInput = false
					res.Confidence = 0.95
					res.Provenance = "preferences.sponsorship"
					res.ResolutionReason = "Candidate confirmed sponsorship is not required"
				}
			} else {
				// AT-003: Fail closed to needs_input
				res.NeedsInput = true
				res.Confidence = 0.0
				res.Provenance = "unconfirmed_needs_input"
				res.ResolutionReason = "Visa sponsorship preference is unspecified. Failing closed to needs_input (AT-003)."
			}

		case CategorySalary:
			if preferences != nil && preferences.Salary.MinimumAmount > 0 {
				res.ResolvedValue = preferences.Salary.MinimumAmount
				res.NeedsInput = false
				res.Confidence = 0.90
				res.Provenance = "preferences.salary.minimum_amount"
				res.ResolutionReason = fmt.Sprintf("Grounded from candidate minimum target compensation (%.0f %s)", preferences.Salary.MinimumAmount, preferences.Salary.Currency)
			} else {
				res.NeedsInput = field.Required
				res.Confidence = 0.0
				res.Provenance = "unconfirmed_needs_input"
				res.ResolutionReason = "Salary expectation is unspecified in candidate preferences."
			}

		case CategoryNoticePeriod:
			if preferences != nil && preferences.NoticePeriodDays > 0 {
				res.ResolvedValue = fmt.Sprintf("%d days", preferences.NoticePeriodDays)
				res.NeedsInput = false
				res.Confidence = 0.90
				res.Provenance = "preferences.notice_period_days"
				res.ResolutionReason = fmt.Sprintf("Grounded notice period (%d days) from candidate preferences", preferences.NoticePeriodDays)
			} else {
				res.NeedsInput = field.Required
				res.Confidence = 0.0
				res.Provenance = "unconfirmed_needs_input"
				res.ResolutionReason = "Notice period is unspecified in candidate preferences."
			}

		case CategoryPortfolio:
			if profile != nil {
				for _, link := range profile.Links {
					if strings.Contains(strings.ToLower(link.LinkType), strings.ToLower(field.TargetSubKey)) ||
						strings.Contains(strings.ToLower(link.Label), strings.ToLower(field.TargetSubKey)) ||
						strings.Contains(strings.ToLower(link.URL), strings.ToLower(field.TargetSubKey)) {
						res.ResolvedValue = link.URL
						res.NeedsInput = false
						res.Confidence = 1.0
						res.Provenance = fmt.Sprintf("profile.links.%s", link.LinkType)
						res.ResolutionReason = fmt.Sprintf("Matched %s URL from candidate confirmed links", link.LinkType)
						break
					}
				}
			}
			if res.ResolvedValue == nil {
				res.NeedsInput = field.Required
				res.Confidence = 0.0
				res.Provenance = "unconfirmed_needs_input"
				res.ResolutionReason = fmt.Sprintf("Portfolio link for '%s' is not in candidate profile.", field.TargetSubKey)
			}

		default:
			// Custom/unknown question -> Fail closed to needs_input
			res.NeedsInput = true
			res.Confidence = 0.0
			res.Provenance = "unconfirmed_needs_input"
			res.ResolutionReason = "Unrecognized or custom question requires candidate confirmation (AT-003 zero-fabrication)."
		}

		resolutions = append(resolutions, res)
	}

	return resolutions
}

// LoadFormFromJSON loads a groundtruth form fixture from a JSON file (FND-016)
func LoadFormFromJSON(filePath string) (*RecognizedForm, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read form fixture file %s: %w", filePath, err)
	}

	var form RecognizedForm
	if err := json.Unmarshal(data, &form); err != nil {
		return nil, fmt.Errorf("failed to parse form fixture JSON: %w", err)
	}

	return &form, nil
}

// ParseOptionToNumeric attempts to parse user answer or option to numeric value
func ParseOptionToNumeric(val interface{}) (float64, error) {
	switch v := val.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		// Clean string of currency symbols or extra words
		cleaned := strings.TrimSpace(v)
		cleaned = strings.ReplaceAll(cleaned, "$", "")
		cleaned = strings.ReplaceAll(cleaned, "€", "")
		cleaned = strings.ReplaceAll(cleaned, "£", "")
		cleaned = strings.ReplaceAll(cleaned, ",", "")
		return strconv.ParseFloat(cleaned, 64)
	default:
		return 0, fmt.Errorf("cannot parse value of type %T to numeric", val)
	}
}
