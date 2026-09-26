package ai

import (
	"fmt"
	"strings"
)

// FactGroundingValidator enforces the strict invariant that AI models cannot invent
// qualifications, experience years, salary expectations, sponsorship needs, or legal answers (REQ-016, AT-003).
type FactGroundingValidator struct{}

func NewFactGroundingValidator() *FactGroundingValidator {
	return &FactGroundingValidator{}
}

// ValidateGrounding verifies that claims and extracted answers are grounded in user's confirmed facts.
// Missing required answers for sensitive fields must be set to "needs_input", not guessed defaults (AT-003).
func (v *FactGroundingValidator) ValidateGrounding(
	facts []ScopedFact,
	claims []string,
	formAnswers map[string]string,
) (GroundingReport, error) {
	report := GroundingReport{
		IsValid:          true,
		GroundedClaims:   make([]string, 0),
		UnverifiedClaims: make([]string, 0),
		NeedsInputFields: make([]string, 0),
		Violations:       make([]string, 0),
	}

	// 1. Build knowledge map of confirmed user facts
	confirmedFacts := make(map[string]bool)
	confirmedCategories := make(map[FactCategory][]string)
	for _, f := range facts {
		norm := strings.ToLower(strings.TrimSpace(f.Statement))
		confirmedFacts[norm] = true
		confirmedCategories[f.Category] = append(confirmedCategories[f.Category], norm)
	}

	// 2. Validate individual claims against confirmed facts (REQ-016)
	if len(facts) > 0 {
		for _, claim := range claims {
			normClaim := strings.ToLower(strings.TrimSpace(claim))
			if normClaim == "" {
				continue
			}

			matched := false
			for fact := range confirmedFacts {
				if strings.Contains(normClaim, fact) || strings.Contains(fact, normClaim) {
					matched = true
					break
				}
			}

			if matched {
				report.GroundedClaims = append(report.GroundedClaims, claim)
			} else {
				report.UnverifiedClaims = append(report.UnverifiedClaims, claim)
				report.Violations = append(report.Violations, fmt.Sprintf("unverified claim not found in user confirmed facts: '%s'", claim))
				report.IsValid = false
			}
		}
	}

	// 3. Inspect Form Question Answers (AT-003)
	// Unknown skill years, salary, sponsorship or legal answers MUST NOT become fabricated defaults
	sensitiveFields := []string{"salary", "sponsorship", "visa", "legal_eligibility", "years_experience"}
	for field, val := range formAnswers {
		normField := strings.ToLower(strings.TrimSpace(field))
		normVal := strings.ToLower(strings.TrimSpace(val))

		for _, sensitive := range sensitiveFields {
			if strings.Contains(normField, sensitive) {
				// Check if user has a confirmed fact for this sensitive field
				hasConfirmedFact := false
				for _, cf := range confirmedCategories[FactCategory(sensitive)] {
					if strings.Contains(normVal, cf) || strings.Contains(cf, normVal) {
						hasConfirmedFact = true
						break
					}
				}

				// If there is no confirmed fact, and the AI guessed an affirmative or fabricated default:
				if !hasConfirmedFact {
					if normVal != "needs_input" && normVal != "unresolved" && normVal != "" {
						report.Violations = append(report.Violations,
							fmt.Sprintf("field '%s' has answer '%s' but no confirmed fact exists. Guessing unknown answers is prohibited (AT-003)", field, val))
						report.IsValid = false
					} else {
						report.NeedsInputFields = append(report.NeedsInputFields, field)
					}
				}
				break
			}
		}
	}

	if !report.IsValid {
		return report, ErrFabricatedFactsDetected
	}

	return report, nil
}
