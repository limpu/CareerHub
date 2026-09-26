package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// ModelInfo describes a supported model in the platform.
type ModelInfo struct {
	ID           string            `json:"id"`
	DisplayName  string            `json:"display_name"`
	Provider     string            `json:"provider"`
	Tier         string            `json:"tier"` // economy, balanced, advanced
	Capabilities ModelCapabilities `json:"capabilities"`
}

// AIGatewayService is the central controller for AI generation, safety, fact grounding,
// and token/cost metering across all logical agents (AI-AGENT-ARCHITECTURE Section 2).
type AIGatewayService struct {
	mu         sync.RWMutex
	adapter    ProviderAdapter
	grounding  *FactGroundingValidator
	safety     *PromptSafetySanitizer
	models     map[string]ModelInfo
}

func NewAIGatewayService(adapter ProviderAdapter) *AIGatewayService {
	s := &AIGatewayService{
		adapter:   adapter,
		grounding: NewFactGroundingValidator(),
		safety:    NewPromptSafetySanitizer(),
		models:    make(map[string]ModelInfo),
	}

	// Register standard model catalog presets
	s.registerDefaultModels()
	return s
}

func (s *AIGatewayService) registerDefaultModels() {
	s.models["economy-text-v1"] = ModelInfo{
		ID:          "economy-text-v1",
		DisplayName: "Economy Text (Default)",
		Provider:    "deterministic-mock",
		Tier:        "economy",
		Capabilities: ModelCapabilities{
			ContextWindow:      128000,
			MaxOutputTokens:    2000,
			SupportsJSONSchema: true,
			SupportsStreaming:  true,
			SupportsTools:      false,
			ReportsUsage:       true,
		},
	}
	s.models["quality-text-v1"] = ModelInfo{
		ID:          "quality-text-v1",
		DisplayName: "Quality Text (Balanced)",
		Provider:    "deterministic-mock",
		Tier:        "balanced",
		Capabilities: ModelCapabilities{
			ContextWindow:      200000,
			MaxOutputTokens:    4000,
			SupportsJSONSchema: true,
			SupportsStreaming:  true,
			SupportsTools:      false,
			ReportsUsage:       true,
		},
	}
}

// ListModels returns all registered AI models and their verified capabilities.
func (s *AIGatewayService) ListModels(ctx context.Context) ([]ModelInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]ModelInfo, 0, len(s.models))
	for _, m := range s.models {
		result = append(result, m)
	}
	return result, nil
}

// Estimate provides preflight token and monetary estimates without executing LLM generation.
func (s *AIGatewayService) Estimate(ctx context.Context, req GenerationRequest) (UsageEstimate, error) {
	return s.adapter.Estimate(ctx, req)
}

// Generate orchestrates the end-to-end safe, grounded AI generation workflow:
// 1. Gating & Model Consent Check (REQ-016)
// 2. Untrusted Prompt Injection Sanitization (AT-019)
// 3. Pre-Dispatch Token & Budget Reservation (REQ-021)
// 4. Provider Execution
// 5. Hard Fact-Grounding & Unknown Answer Verification (REQ-016, AT-003)
func (s *AIGatewayService) Generate(ctx context.Context, req GenerationRequest) (*GenerationResult, error) {
	// 1. Check Model Consent for private career facts (REQ-016)
	if len(req.ScopedFacts) > 0 && !req.HasModelConsent {
		return nil, ErrModelConsentRequired
	}

	// 2. Prompt Safety & Injection Defense (AT-019)
	if req.UntrustedData != "" {
		sanitized, err := s.safety.SanitizeUntrustedData(req.UntrustedData)
		if err != nil {
			return nil, err // Rejects injection attempts with ErrPromptInjectionBlocked
		}
		req.UntrustedData = sanitized
	}

	// Set default model if empty
	if req.ModelID == "" {
		req.ModelID = "economy-text-v1"
	}

	// 3. Estimate usage and verify within budget caps (REQ-021)
	est, err := s.adapter.Estimate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to compute usage estimate: %w", err)
	}
	if est.EstimatedInputTokens > 100000 {
		return nil, ErrBudgetExceeded
	}

	// 4. Dispatch Generation to Provider Adapter
	result, err := s.adapter.Generate(ctx, req)
	if err != nil {
		return nil, err
	}

	// 5. Fact-Grounding Verification (REQ-016, AT-003)
	// Parse output to extract claims or answers to check against confirmed facts
	var claims []string
	formAnswers := make(map[string]string)

	var parsedMap map[string]interface{}
	if err := json.Unmarshal(result.OutputContent, &parsedMap); err == nil {
		// Check for extracted skills
		if skills, ok := parsedMap["skills"].([]interface{}); ok {
			for _, sk := range skills {
				if sStr, ok := sk.(string); ok {
					claims = append(claims, sStr)
				}
			}
		}
		if matchingSkills, ok := parsedMap["matching_skills"].([]interface{}); ok {
			for _, sk := range matchingSkills {
				if sStr, ok := sk.(string); ok {
					claims = append(claims, sStr)
				}
			}
		}

		// Check for form answers
		if ans, ok := parsedMap["answer"].(string); ok {
			if ans != "needs_input" && ans != "unresolved" {
				claims = append(claims, ans)
			}
			if q, ok := parsedMap["question"].(string); ok {
				formAnswers[q] = ans
			}
		}

		// Direct question answers map if present
		if answersMap, ok := parsedMap["answers"].(map[string]interface{}); ok {
			for k, v := range answersMap {
				formAnswers[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// Run grounding validation when task has confirmed user facts or form questions (REQ-016, AT-003)
	if len(req.ScopedFacts) > 0 || len(formAnswers) > 0 {
		report, err := s.grounding.ValidateGrounding(req.ScopedFacts, claims, formAnswers)
		result.GroundingReport = report
		if err != nil {
			return result, err // Fails closed on fabricated facts
		}
	} else {
		result.GroundingReport = GroundingReport{
			IsValid: true,
		}
	}

	return result, nil
}
