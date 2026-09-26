package linkedin

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrProtectedAttributeProhibited = errors.New("discovery: filtering or scoring by protected attributes (age, gender, race, religion, sexual orientation, disability) is strictly prohibited (LI-05)")
	ErrInvalidDiscoveryCriteria     = errors.New("discovery: discovery criteria must specify a target role or company, and a valid user goal")
	ErrInvalidLinkedInURL           = errors.New("discovery: input is not a recognized valid LinkedIn profile or company URL")
)

// UserGoal represents the candidate/user's stated objective for discovery.
type UserGoal string

const (
	GoalNetworking   UserGoal = "networking"
	GoalRecruiting   UserGoal = "recruiting"
	GoalPartnerships UserGoal = "partnerships"
	GoalBenchmarking UserGoal = "benchmarking"
)

// DiscoveryCriteria encapsulates professional query filters (LI-05).
type DiscoveryCriteria struct {
	TargetRole        string            `json:"target_role"`
	RoleAlternatives  []string          `json:"role_alternatives,omitempty"`
	Industries        []string          `json:"industries,omitempty"`
	Locations         []string          `json:"locations,omitempty"`
	CurrentCompanies  []string          `json:"current_companies,omitempty"`
	ConnectionTiers   []string          `json:"connection_tiers,omitempty"` // "1st", "2nd", "3rd+"
	CompanySizeTiers  []string          `json:"company_size_tiers,omitempty"`
	Exclusions        []string          `json:"exclusions,omitempty"`
	UserGoal          UserGoal          `json:"user_goal"`
	CustomFilters     map[string]string `json:"custom_filters,omitempty"`
}

// BooleanQuery stores structured Boolean expressions tailored for LinkedIn native search (SRC-L1).
type BooleanQuery struct {
	RawQuery          string `json:"raw_query"`
	RoleClause        string `json:"role_clause"`
	CompanyClause     string `json:"company_clause"`
	ExclusionClause   string `json:"exclusion_clause"`
	LinkedInSearchURL string `json:"linkedin_search_url"`
}

// CanonicalURL holds normalized, tracker-free LinkedIn entity permalinks (SRC-L2).
type CanonicalURL struct {
	RawURL        string `json:"raw_url"`
	NormalizedURL string `json:"normalized_url"`
	EntityType    string `json:"entity_type"` // "person" or "company"
	Slug          string `json:"slug"`
}

// DiscoveryResult delivers direct Vault matches plus native dispatch assistance (AT-010).
type DiscoveryResult struct {
	Criteria          DiscoveryCriteria `json:"criteria"`
	BooleanQuery      BooleanQuery      `json:"boolean_query"`
	MatchedPersons    []PersonRecord    `json:"matched_persons"`
	MatchedCompanies  []CompanyRecord   `json:"matched_companies"`
	TotalVaultMatches int               `json:"total_vault_matches"`
	PlatformNotice    string            `json:"platform_notice"`
	GeneratedAt       time.Time         `json:"generated_at"`
}

// ValidateDiscoveryCriteria enforces protection against prohibited protected-attribute discrimination (LI-05).
func ValidateDiscoveryCriteria(criteria DiscoveryCriteria) error {
	if strings.TrimSpace(criteria.TargetRole) == "" && len(criteria.CurrentCompanies) == 0 {
		return fmt.Errorf("%w: target_role or current_companies must be provided", ErrInvalidDiscoveryCriteria)
	}

	// Valid user goals
	switch criteria.UserGoal {
	case GoalNetworking, GoalRecruiting, GoalPartnerships, GoalBenchmarking:
	case "":
		// Default to networking if empty
	default:
		return fmt.Errorf("%w: unrecognized user_goal '%s'", ErrInvalidDiscoveryCriteria, criteria.UserGoal)
	}

	// Prohibited protected attributes list (Title VII, ADEA, ADA compliance)
	prohibitedKeywords := []string{
		"age", "under_30", "over_40", "under_40", "older", "young", "gender", "female", "male", "woman", "man",
		"race", "ethnicity", "black", "white", "asian", "hispanic", "latino", "religion", "christian", "jewish",
		"muslim", "hindu", "sexual_orientation", "marital_status", "pregnant", "disability", "handicap",
	}

	checkString := func(s string, context string) error {
		lower := strings.ToLower(s)
		for _, kw := range prohibitedKeywords {
			// Precise word boundary or snake_case token boundary
			pattern := fmt.Sprintf(`(?i)(?:^|[\s_\W])%s(?:$|[\s_\W])`, regexp.QuoteMeta(kw))
			matched, _ := regexp.MatchString(pattern, lower)
			if matched {
				return fmt.Errorf("%w: detected prohibited protected attribute '%s' in %s", ErrProtectedAttributeProhibited, kw, context)
			}
		}
		return nil
	}

	if err := checkString(criteria.TargetRole, "target_role"); err != nil {
		return err
	}
	for _, role := range criteria.RoleAlternatives {
		if err := checkString(role, "role_alternatives"); err != nil {
			return err
		}
	}
	for _, excl := range criteria.Exclusions {
		if err := checkString(excl, "exclusions"); err != nil {
			return err
		}
	}
	for k, v := range criteria.CustomFilters {
		if err := checkString(k, "custom_filters key"); err != nil {
			return err
		}
		if err := checkString(v, "custom_filters value"); err != nil {
			return err
		}
	}

	return nil
}

// BuildBooleanQuery constructs an optimized, grouped Boolean search string for LinkedIn native search (SRC-L1).
func BuildBooleanQuery(criteria DiscoveryCriteria) (BooleanQuery, error) {
	if err := ValidateDiscoveryCriteria(criteria); err != nil {
		return BooleanQuery{}, err
	}

	var clauses []string
	var roleClause, companyClause, exclClause string

	// 1. Role / Title Clause
	var roles []string
	if strings.TrimSpace(criteria.TargetRole) != "" {
		roles = append(roles, fmt.Sprintf("\"%s\"", strings.TrimSpace(criteria.TargetRole)))
	}
	for _, alt := range criteria.RoleAlternatives {
		clean := strings.TrimSpace(alt)
		if clean != "" {
			roles = append(roles, fmt.Sprintf("\"%s\"", clean))
		}
	}
	if len(roles) > 0 {
		roleClause = fmt.Sprintf("(%s)", strings.Join(roles, " OR "))
		clauses = append(clauses, roleClause)
	}

	// 2. Company Clause
	var companies []string
	for _, c := range criteria.CurrentCompanies {
		clean := strings.TrimSpace(c)
		if clean != "" {
			companies = append(companies, fmt.Sprintf("\"%s\"", clean))
		}
	}
	if len(companies) > 0 {
		companyClause = fmt.Sprintf("(%s)", strings.Join(companies, " OR "))
		clauses = append(clauses, companyClause)
	}

	// 3. Exclusions Clause
	var excls []string
	for _, e := range criteria.Exclusions {
		clean := strings.TrimSpace(e)
		if clean != "" {
			excls = append(excls, fmt.Sprintf("\"%s\"", clean))
		}
	}
	if len(excls) > 0 {
		exclClause = fmt.Sprintf("NOT (%s)", strings.Join(excls, " OR "))
		clauses = append(clauses, exclClause)
	}

	rawQuery := strings.Join(clauses, " AND ")

	// Generate deep-link to LinkedIn Native People Search with prefilled keywords
	encodedQuery := url.QueryEscape(rawQuery)
	linkedInSearchURL := fmt.Sprintf("https://www.linkedin.com/search/results/people/?keywords=%s&origin=GLOBAL_SEARCH_HEADER", encodedQuery)

	return BooleanQuery{
		RawQuery:          rawQuery,
		RoleClause:        roleClause,
		CompanyClause:     companyClause,
		ExclusionClause:   exclClause,
		LinkedInSearchURL: linkedInSearchURL,
	}, nil
}

// NormalizeLinkedInURL cleans and normalizes tracking-polluted URLs into canonical permalinks (SRC-L2).
func NormalizeLinkedInURL(rawURL string) (CanonicalURL, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return CanonicalURL{}, ErrInvalidLinkedInURL
	}

	// Ensure scheme prefix for proper parsing
	parsedInput := trimmed
	if !strings.HasPrefix(parsedInput, "http://") && !strings.HasPrefix(parsedInput, "https://") {
		parsedInput = "https://" + parsedInput
	}

	u, err := url.Parse(parsedInput)
	if err != nil {
		return CanonicalURL{}, fmt.Errorf("%w: %v", ErrInvalidLinkedInURL, err)
	}

	host := strings.ToLower(u.Host)
	if !strings.Contains(host, "linkedin.com") {
		return CanonicalURL{}, fmt.Errorf("%w: host must be linkedin.com", ErrInvalidLinkedInURL)
	}

	// Clean path segments
	cleanPath := strings.Trim(u.Path, "/")
	parts := strings.Split(cleanPath, "/")

	if len(parts) >= 2 && (parts[0] == "in" || parts[0] == "company") {
		entityType := "person"
		if parts[0] == "company" {
			entityType = "company"
		}
		slug := parts[1]

		normalized := fmt.Sprintf("https://www.linkedin.com/%s/%s", parts[0], slug)
		return CanonicalURL{
			RawURL:        trimmed,
			NormalizedURL: normalized,
			EntityType:    entityType,
			Slug:          slug,
		}, nil
	}

	return CanonicalURL{}, fmt.Errorf("%w: unrecognized LinkedIn entity path in '%s'", ErrInvalidLinkedInURL, rawURL)
}
