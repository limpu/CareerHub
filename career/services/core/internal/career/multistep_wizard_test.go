package career

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestStateMachine_ForwardBackwardTransitions(t *testing.T) {
	steps := []FormStep{
		{
			StepIndex: 1,
			StepName:  "Contact",
			Fields: []FormField{
				{FieldID: "name", Label: "Full Name", Required: true},
			},
		},
		{
			StepIndex: 2,
			StepName:  "Experience",
			Fields: []FormField{
				{FieldID: "years", Label: "Years of Experience", Required: true},
			},
		},
		{
			StepIndex: 3,
			StepName:  "Review",
			Fields: []FormField{
				{FieldID: "consent", Label: "Agree to terms", Required: true},
			},
		},
	}

	sm := NewFormStateMachine("test_wiz_1", "Test Wizard", "linkedin", steps)

	// Step 1: Advance with required name
	answers1 := map[string]interface{}{"name": "Alice Developer"}
	nextStep, err := sm.AdvanceStep(answers1)
	if err != nil {
		t.Fatalf("failed advancing step 1: %v", err)
	}
	if nextStep.StepIndex != 2 {
		t.Errorf("expected to reach step 2, got %d", nextStep.StepIndex)
	}
	if sm.CurrentStepIndex != 2 {
		t.Errorf("expected state machine step index 2, got %d", sm.CurrentStepIndex)
	}

	// Step 2: Advance with required years
	answers2 := map[string]interface{}{"years": 5}
	nextStep, err = sm.AdvanceStep(answers2)
	if err != nil {
		t.Fatalf("failed advancing step 2: %v", err)
	}
	if nextStep.StepIndex != 3 {
		t.Errorf("expected to reach step 3, got %d", nextStep.StepIndex)
	}

	// Navigate backward to Step 2
	prevStep, err := sm.PreviousStep()
	if err != nil {
		t.Fatalf("failed navigating previous step: %v", err)
	}
	if prevStep.StepIndex != 2 {
		t.Errorf("expected previous step to be 2, got %d", prevStep.StepIndex)
	}
	if sm.CurrentStepIndex != 2 {
		t.Errorf("expected current step index 2, got %d", sm.CurrentStepIndex)
	}

	// Move forward to Step 3 again
	nextStep, err = sm.AdvanceStep(answers2)
	if err != nil {
		t.Fatalf("failed re-advancing to step 3: %v", err)
	}
	if nextStep.StepIndex != 3 {
		t.Errorf("expected step 3, got %d", nextStep.StepIndex)
	}

	// Complete final step
	answers3 := map[string]interface{}{"consent": true}
	finalStep, err := sm.AdvanceStep(answers3)
	if err != nil {
		t.Fatalf("failed completing step 3: %v", err)
	}
	if sm.Status != StepStatusCompleted {
		t.Errorf("expected status completed, got %s", sm.Status)
	}
	if finalStep.StepIndex != 3 {
		t.Errorf("expected final step index 3, got %d", finalStep.StepIndex)
	}
}

func TestStateMachine_LoopDetection_HaltsGracefully(t *testing.T) {
	steps := []FormStep{
		{
			StepIndex: 1,
			StepName:  "Work Authorization Modal",
			Fields: []FormField{
				{FieldID: "q_auth", Label: "Authorized to work", Required: true},
			},
		},
		{
			StepIndex: 2,
			StepName:  "Review",
			Fields: []FormField{
				{FieldID: "q_review", Label: "Review", Required: false},
			},
		},
	}

	sm := NewFormStateMachine("test_wiz_loop", "Loop Wizard", "greenhouse", steps)
	sm.MaxConsecutiveLoopThreshold = 2 // Detect after 2 consecutive identical failures

	answers := map[string]interface{}{"q_auth": "Yes"}

	// Simulate broken external portal where clicking Next does not progress DOM and repeats state
	_, err := sm.ForceLoopSimulation(answers)
	if err != nil {
		t.Fatalf("first simulation click should succeed: %v", err)
	}

	_, err = sm.ForceLoopSimulation(answers)
	if err != nil {
		t.Fatalf("second simulation click should succeed: %v", err)
	}

	// Third attempt with identical state MUST trigger ErrLoopDetected (CAR-13, AT-006)
	_, err = sm.ForceLoopSimulation(answers)
	if err == nil {
		t.Fatalf("expected ErrLoopDetected on repeated identical step loop, got nil")
	}
	if !errors.Is(err, ErrLoopDetected) {
		t.Errorf("expected error to be ErrLoopDetected, got %v", err)
	}
	if sm.Status != StepStatusLoopDetected {
		t.Errorf("expected state machine status to be %s, got %s", StepStatusLoopDetected, sm.Status)
	}
	if sm.HaltReason == "" {
		t.Errorf("expected descriptive halt reason for loop detection")
	}
}

func TestShadowDOM_PiercingAndExtraction(t *testing.T) {
	steps := []FormStep{
		{
			StepIndex:          1,
			StepName:           "Modern LinkedIn Shadow Container",
			IsShadowDOM:        true,
			ShadowRootSelector: "#interop-outlet",
			Fields: []FormField{
				{FieldID: "shadow_radio_1", Label: "Visa Requirement", Type: FieldTypeRadio, Required: true},
				{FieldID: "shadow_input_2", Label: "Current Visa Code", Type: FieldTypeText, Required: false},
			},
		},
	}

	sm := NewFormStateMachine("shadow_wiz", "LinkedIn Modern Modal", "linkedin", steps)

	mockRawHTML := `<html><body><div id="interop-outlet" class="shadow-host"></div></body></html>`

	traversedInfo, fields, err := sm.TraverseShadowDOM(mockRawHTML, "#interop-outlet")
	if err != nil {
		t.Fatalf("failed traversing shadow DOM: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields pierced from shadow DOM, got %d", len(fields))
	}
	if fields[0].FieldID != "shadow_radio_1" {
		t.Errorf("expected first field shadow_radio_1, got %s", fields[0].FieldID)
	}
	if !strings.Contains(traversedInfo, "pierced_shadow_root:#interop-outlet") {
		t.Errorf("expected traversedInfo to contain pierced selector info, got %s", traversedInfo)
	}

	// Verify piercing selector generator
	piercingSel := PiercingShadowDOMSelector("#interop-outlet", "input[type='radio']")
	expectedSel := "#interop-outlet >> input[type='radio']"
	if piercingSel != expectedSel {
		t.Errorf("expected piercing selector %s, got %s", expectedSel, piercingSel)
	}
}

func TestStateMachine_ZeroFabrication_BlocksTransition(t *testing.T) {
	steps := []FormStep{
		{
			StepIndex: 1,
			StepName:  "Qualifications",
			Fields: []FormField{
				{FieldID: "req_go", Label: "Years with Go", Required: true},
				{FieldID: "opt_comment", Label: "Additional notes", Required: false},
			},
		},
		{
			StepIndex: 2,
			StepName:  "Submission",
		},
	}

	sm := NewFormStateMachine("zero_fab_sm", "Zero Fab Guard", "lever", steps)

	// Missing required field "req_go"
	emptyAnswers := map[string]interface{}{
		"opt_comment": "Excited for this role",
	}

	_, err := sm.AdvanceStep(emptyAnswers)
	if err == nil {
		t.Fatalf("AT-003 VIOLATION: Expected advance step to be blocked when required field is unanswered")
	}
	if !errors.Is(err, ErrStepBlockedNeedsInput) {
		t.Errorf("expected ErrStepBlockedNeedsInput, got %v", err)
	}
	if sm.Status != StepStatusBlockedNeedsInput {
		t.Errorf("expected status %s, got %s", StepStatusBlockedNeedsInput, sm.Status)
	}
	if len(sm.UnresolvedFields) != 1 || sm.UnresolvedFields[0] != "req_go" {
		t.Errorf("expected unresolved fields ['req_go'], got %v", sm.UnresolvedFields)
	}
}

func TestEvidence_PIIScrubbing(t *testing.T) {
	rawEvidence := `User candidate entered email: john.doe@securemail.com, phone: +1-555-432-8765, ssn: 123-45-6789, credit_card: 4111 2222 3333 4444, bearer: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.payload`

	scrubbed := ScrubStepEvidence(rawEvidence)

	if scrubbed == rawEvidence {
		t.Errorf("expected evidence to be scrubbed, but got raw text unchanged")
	}

	sensitiveSubstrings := []string{
		"john.doe@securemail.com",
		"+1-555-432-8765",
		"123-45-6789",
		"4111 2222 3333 4444",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
	}

	for _, s := range sensitiveSubstrings {
		if strings.Contains(scrubbed, s) {
			t.Errorf("CAR-13 / AT-011 VIOLATION: Sensitive PII '%s' was not redacted in scrubbed evidence:\n%s", s, scrubbed)
		}
	}
}

func TestStateMachine_MultistepFixture(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "multistep_shadow_dom_form.json")
	sm, err := LoadWizardFromJSON(fixturePath)
	if err != nil {
		t.Fatalf("failed to load multistep shadow DOM fixture: %v", err)
	}

	if sm.TotalSteps != 3 {
		t.Fatalf("expected 3 steps in fixture, got %d", sm.TotalSteps)
	}

	// Step 1: Contact (Light DOM)
	s1, err := sm.GetCurrentStep()
	if err != nil {
		t.Fatalf("failed getting step 1: %v", err)
	}
	if s1.IsShadowDOM {
		t.Errorf("step 1 should not be shadow DOM")
	}

	ans1 := map[string]interface{}{
		"step1_name":  "Alice Candidate",
		"step1_email": "alice@candidate.com",
		"step1_phone": "+1 555 0192",
	}
	s2, err := sm.AdvanceStep(ans1)
	if err != nil {
		t.Fatalf("failed advancing to step 2: %v", err)
	}
	if s2.StepIndex != 2 {
		t.Errorf("expected step index 2, got %d", s2.StepIndex)
	}

	// Step 2: Shadow DOM Work Auth (#interop-outlet)
	if !s2.IsShadowDOM || s2.ShadowRootSelector != "#interop-outlet" {
		t.Errorf("expected step 2 to be Shadow DOM with #interop-outlet, got %t / %s", s2.IsShadowDOM, s2.ShadowRootSelector)
	}

	ans2 := map[string]interface{}{
		"step2_work_auth":   "Yes",
		"step2_sponsorship": "No",
		"step2_go_years":    4,
	}
	s3, err := sm.AdvanceStep(ans2)
	if err != nil {
		t.Fatalf("failed advancing to step 3: %v", err)
	}
	if s3.StepIndex != 3 {
		t.Errorf("expected step index 3, got %d", s3.StepIndex)
	}

	// Step 3: Complete Review
	ans3 := map[string]interface{}{
		"step3_confirm_terms": true,
	}
	finalStep, err := sm.AdvanceStep(ans3)
	if err != nil {
		t.Fatalf("failed completing step 3: %v", err)
	}
	if sm.Status != StepStatusCompleted {
		t.Errorf("expected wizard status completed, got %s", sm.Status)
	}
	if finalStep.StepIndex != 3 {
		t.Errorf("expected final step index 3, got %d", finalStep.StepIndex)
	}
}
