package career

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	ErrInvalidStepTransition    = errors.New("invalid wizard step transition")
	ErrStepBlockedNeedsInput    = errors.New("cannot advance step: required questions need candidate input (AT-003)")
	ErrLoopDetected             = errors.New("form navigation loop detected: repeated identical form state without progression (CAR-13, AT-006)")
	ErrShadowDOMHostNotFound    = errors.New("shadow DOM host element not located")
	ErrWizardAlreadyCompleted   = errors.New("wizard is already in completed state")
	ErrStepIndexOutOfBounds     = errors.New("step index out of bounds")
)

// FormStepStatus defines the lifecycle state of a form step or state machine
type FormStepStatus string

const (
	StepStatusPending          FormStepStatus = "pending"
	StepStatusInProgress       FormStepStatus = "in_progress"
	StepStatusCompleted        FormStepStatus = "completed"
	StepStatusFailed           FormStepStatus = "failed"
	StepStatusBlockedNeedsInput FormStepStatus = "blocked_needs_input"
	StepStatusLoopDetected     FormStepStatus = "loop_detected"
)

// FormStep represents an individual step in a multi-step modal (CAR-13)
type FormStep struct {
	StepIndex          int       `json:"step_index"`
	StepName           string    `json:"step_name"`
	IsShadowDOM        bool      `json:"is_shadow_dom"`
	ShadowRootSelector string    `json:"shadow_root_selector,omitempty"` // e.g. "#interop-outlet"
	Fields             []FormField `json:"fields"`
	StepFingerprint    string    `json:"step_fingerprint,omitempty"`
	ScrubbedEvidence   string    `json:"scrubbed_evidence,omitempty"`
}

// FormStateMachine tracks progress, historical fingerprints, and loop detection (CAR-13, AT-006)
type FormStateMachine struct {
	WizardID                    string         `json:"wizard_id"`
	Title                       string         `json:"title"`
	Provider                    string         `json:"provider"`
	TotalSteps                  int            `json:"total_steps"`
	CurrentStepIndex            int            `json:"current_step_index"`
	Steps                       []FormStep     `json:"steps"`
	HistoryFingerprints         []string       `json:"history_fingerprints"`
	MaxConsecutiveLoopThreshold int            `json:"max_consecutive_loop_threshold"` // default 2
	Status                      FormStepStatus `json:"status"`
	HaltReason                  string         `json:"halt_reason,omitempty"`
	UnresolvedFields            []string       `json:"unresolved_fields,omitempty"`
	CreatedAt                   time.Time      `json:"created_at"`
	UpdatedAt                   time.Time      `json:"updated_at"`
}

// MultiStepWizardFixture represents serialized wizard schema from test fixtures
type MultiStepWizardFixture struct {
	WizardID   string     `json:"wizard_id"`
	Title      string     `json:"title"`
	Provider   string     `json:"provider"`
	TotalSteps int        `json:"total_steps"`
	Steps      []FormStep `json:"steps"`
}

// PII scrubbing regex patterns for form evidence (CAR-13, AT-011)
var (
	emailScrubRegex    = regexp.MustCompile(`(?i)[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	phoneScrubRegex    = regexp.MustCompile(`\+?[0-9]{1,3}?[-.\s]?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}`)
	ssnScrubRegex      = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	creditCardRegex    = regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`)
	tokenKeyScrubRegex = regexp.MustCompile(`(?i)(bearer\s+[a-zA-Z0-9_\-\.]{20,}|(?:api[_-]?key|secret|password|auth_token)\s*[:=]\s*["']?[a-zA-Z0-9_\-\.]{8,}["']?)`)
)

// ScrubStepEvidence neutralizes personal identifiable information and secrets from evidence snapshots
func ScrubStepEvidence(rawEvidence string) string {
	scrubbed := rawEvidence
	scrubbed = emailScrubRegex.ReplaceAllString(scrubbed, "[REDACTED_EMAIL]")
	scrubbed = phoneScrubRegex.ReplaceAllString(scrubbed, "[REDACTED_PHONE]")
	scrubbed = ssnScrubRegex.ReplaceAllString(scrubbed, "[REDACTED_SSN]")
	scrubbed = creditCardRegex.ReplaceAllString(scrubbed, "[REDACTED_CC]")
	scrubbed = tokenKeyScrubRegex.ReplaceAllString(scrubbed, "[REDACTED_SECRET]")
	return scrubbed
}

// PiercingShadowDOMSelector returns a Playwright-compatible shadow DOM piercing locator (SRC-C4)
func PiercingShadowDOMSelector(hostSelector, targetSelector string) string {
	host := strings.TrimSpace(hostSelector)
	target := strings.TrimSpace(targetSelector)
	if host == "" {
		return target
	}
	// Playwright piercing syntax uses >> or css deep piercing
	return fmt.Sprintf("%s >> %s", host, target)
}

// ComputeStepFingerprint creates a deterministic SHA-256 fingerprint for a step based on its fields and answers
func ComputeStepFingerprint(step *FormStep, answers map[string]interface{}) string {
	var fieldKeys []string
	for _, f := range step.Fields {
		val := ""
		if v, ok := answers[f.FieldID]; ok && v != nil {
			val = fmt.Sprintf("%v", v)
		}
		fieldKeys = append(fieldKeys, fmt.Sprintf("%s:%s:%s", f.FieldID, f.Type, val))
	}
	sort.Strings(fieldKeys)

	raw := fmt.Sprintf("step:%d:%s:shadow:%t:%s", step.StepIndex, step.StepName, step.IsShadowDOM, strings.Join(fieldKeys, "|"))
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:16]) // 32-character hex
}

// NewFormStateMachine initializes a multi-step form state machine
func NewFormStateMachine(wizardID, title, provider string, steps []FormStep) *FormStateMachine {
	if wizardID == "" {
		wizardID = fmt.Sprintf("wiz_%d", time.Now().UnixNano())
	}
	total := len(steps)
	return &FormStateMachine{
		WizardID:                    wizardID,
		Title:                       title,
		Provider:                    provider,
		TotalSteps:                  total,
		CurrentStepIndex:            1,
		Steps:                       steps,
		HistoryFingerprints:         make([]string, 0),
		MaxConsecutiveLoopThreshold: 2,
		Status:                      StepStatusInProgress,
		CreatedAt:                   time.Now().UTC(),
		UpdatedAt:                   time.Now().UTC(),
	}
}

// LoadWizardFromJSON loads a multi-step form fixture from JSON file (FND-016)
func LoadWizardFromJSON(filePath string) (*FormStateMachine, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wizard fixture file %s: %w", filePath, err)
	}

	var fixture MultiStepWizardFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		return nil, fmt.Errorf("failed to parse wizard fixture JSON: %w", err)
	}

	sm := NewFormStateMachine(fixture.WizardID, fixture.Title, fixture.Provider, fixture.Steps)
	return sm, nil
}

// GetCurrentStep returns the pointer to the active step
func (sm *FormStateMachine) GetCurrentStep() (*FormStep, error) {
	if sm.CurrentStepIndex < 1 || sm.CurrentStepIndex > len(sm.Steps) {
		return nil, ErrStepIndexOutOfBounds
	}
	return &sm.Steps[sm.CurrentStepIndex-1], nil
}

// CanAdvanceStep checks whether all required fields on the step are answered (AT-003: Zero fabrication)
func (sm *FormStateMachine) CanAdvanceStep(answers map[string]interface{}) (bool, []string) {
	currentStep, err := sm.GetCurrentStep()
	if err != nil {
		return false, []string{err.Error()}
	}

	var missingRequired []string
	for _, f := range currentStep.Fields {
		if !f.Required {
			continue
		}
		val, exists := answers[f.FieldID]
		if !exists || val == nil || strings.TrimSpace(fmt.Sprintf("%v", val)) == "" {
			missingRequired = append(missingRequired, f.FieldID)
		}
	}

	return len(missingRequired) == 0, missingRequired
}

// DetectLoop verifies if the last consecutive fingerprints match the new fingerprint (CAR-13, AT-006)
func (sm *FormStateMachine) DetectLoop(newFingerprint string) bool {
	if len(sm.HistoryFingerprints) < sm.MaxConsecutiveLoopThreshold {
		return false
	}

	consecutiveIdentical := 0
	for i := len(sm.HistoryFingerprints) - 1; i >= 0; i-- {
		if sm.HistoryFingerprints[i] == newFingerprint {
			consecutiveIdentical++
		} else {
			break
		}
	}

	return consecutiveIdentical >= sm.MaxConsecutiveLoopThreshold
}

// AdvanceStep transitions the form to the next step, evaluating loop detection and zero-fabrication gating
func (sm *FormStateMachine) AdvanceStep(answers map[string]interface{}) (*FormStep, error) {
	if sm.Status == StepStatusCompleted {
		return nil, ErrWizardAlreadyCompleted
	}
	if sm.Status == StepStatusLoopDetected {
		return nil, fmt.Errorf("%w: %s", ErrLoopDetected, sm.HaltReason)
	}

	currentStep, err := sm.GetCurrentStep()
	if err != nil {
		return nil, err
	}

	// 1. Invariant AT-003: Verify required questions are not missing/unconfirmed
	canAdvance, missing := sm.CanAdvanceStep(answers)
	if !canAdvance {
		sm.Status = StepStatusBlockedNeedsInput
		sm.UnresolvedFields = missing
		sm.HaltReason = fmt.Sprintf("Step %d requires candidate confirmation for fields: %s", sm.CurrentStepIndex, strings.Join(missing, ", "))
		sm.UpdatedAt = time.Now().UTC()
		return nil, fmt.Errorf("%w: %s", ErrStepBlockedNeedsInput, sm.HaltReason)
	}

	// 2. Compute current step fingerprint
	currentFP := ComputeStepFingerprint(currentStep, answers)
	currentStep.StepFingerprint = currentFP

	// 3. Invariant CAR-13, AT-006: Loop Detection
	if sm.DetectLoop(currentFP) {
		sm.Status = StepStatusLoopDetected
		sm.HaltReason = fmt.Sprintf("Repeated step fingerprint '%s' observed %d times without progression. Halting click loop.", currentFP, sm.MaxConsecutiveLoopThreshold)
		sm.UpdatedAt = time.Now().UTC()
		return nil, fmt.Errorf("%w: %s", ErrLoopDetected, sm.HaltReason)
	}

	// Record to history
	sm.HistoryFingerprints = append(sm.HistoryFingerprints, currentFP)

	// 4. Advance step index
	if sm.CurrentStepIndex >= sm.TotalSteps {
		// Last step completed
		sm.Status = StepStatusCompleted
		sm.HaltReason = ""
		sm.UnresolvedFields = nil
		sm.UpdatedAt = time.Now().UTC()
		return currentStep, nil
	}

	sm.CurrentStepIndex++
	sm.Status = StepStatusInProgress
	sm.HaltReason = ""
	sm.UnresolvedFields = nil
	sm.UpdatedAt = time.Now().UTC()

	nextStep, err := sm.GetCurrentStep()
	if err != nil {
		return nil, err
	}
	return nextStep, nil
}

// PreviousStep navigates backward in the wizard without clearing answered facts
func (sm *FormStateMachine) PreviousStep() (*FormStep, error) {
	if sm.CurrentStepIndex <= 1 {
		return nil, fmt.Errorf("%w: already at step 1", ErrInvalidStepTransition)
	}

	sm.CurrentStepIndex--
	sm.Status = StepStatusInProgress
	sm.HaltReason = ""
	sm.UpdatedAt = time.Now().UTC()

	return sm.GetCurrentStep()
}

// ForceLoopSimulation simulates an external portal bug where next clicks fail to advance DOM
func (sm *FormStateMachine) ForceLoopSimulation(answers map[string]interface{}) (*FormStep, error) {
	currentStep, err := sm.GetCurrentStep()
	if err != nil {
		return nil, err
	}

	currentFP := ComputeStepFingerprint(currentStep, answers)
	if sm.DetectLoop(currentFP) {
		sm.Status = StepStatusLoopDetected
		sm.HaltReason = fmt.Sprintf("Form navigation loop detected: identical fingerprint '%s' repeated %d times without advancing DOM (AT-006).", currentFP, sm.MaxConsecutiveLoopThreshold)
		sm.UpdatedAt = time.Now().UTC()
		return nil, fmt.Errorf("%w: %s", ErrLoopDetected, sm.HaltReason)
	}

	sm.HistoryFingerprints = append(sm.HistoryFingerprints, currentFP)
	sm.UpdatedAt = time.Now().UTC()
	return currentStep, nil
}

// TraverseShadowDOM simulates extracting shadow DOM content given host selector (SRC-C4)
func (sm *FormStateMachine) TraverseShadowDOM(rawHTML, shadowHostSelector string) (string, []FormField, error) {
	if strings.TrimSpace(shadowHostSelector) == "" {
		return "", nil, ErrShadowDOMHostNotFound
	}

	// Verify shadow host element presence
	if !strings.Contains(rawHTML, shadowHostSelector) && !strings.Contains(rawHTML, strings.TrimPrefix(shadowHostSelector, "#")) {
		return "", nil, fmt.Errorf("%w: selector %s", ErrShadowDOMHostNotFound, shadowHostSelector)
	}

	// Find active step matching shadow selector
	var matchedFields []FormField
	for _, step := range sm.Steps {
		if step.IsShadowDOM && step.ShadowRootSelector == shadowHostSelector {
			matchedFields = append(matchedFields, step.Fields...)
		}
	}

	traversedInfo := fmt.Sprintf("pierced_shadow_root:%s:found_%d_fields", shadowHostSelector, len(matchedFields))
	return traversedInfo, matchedFields, nil
}
