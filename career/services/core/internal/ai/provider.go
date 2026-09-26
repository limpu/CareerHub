package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ProviderAdapter defines the unified contract for external AI providers (AI-AGENT-ARCHITECTURE Section 19.1).
type ProviderAdapter interface {
	ValidateConfig(ctx context.Context, endpoint, key string) error
	Capabilities(ctx context.Context, modelID string) (ModelCapabilities, error)
	Estimate(ctx context.Context, req GenerationRequest) (UsageEstimate, error)
	Generate(ctx context.Context, req GenerationRequest) (*GenerationResult, error)
}

// MockDeterministicAdapter provides deterministic, verifiable outputs for all 10 agent tasks
// without incurring live cloud API costs during automated testing and CI.
type MockDeterministicAdapter struct {
	customResponses map[TaskKey]string
}

func NewMockDeterministicAdapter() *MockDeterministicAdapter {
	return &MockDeterministicAdapter{
		customResponses: make(map[TaskKey]string),
	}
}

func (m *MockDeterministicAdapter) SetCustomResponse(task TaskKey, responseJSON string) {
	m.customResponses[task] = responseJSON
}

func (m *MockDeterministicAdapter) ValidateConfig(ctx context.Context, endpoint, key string) error {
	if endpoint == "" {
		return fmt.Errorf("provider endpoint cannot be empty")
	}
	return nil
}

func (m *MockDeterministicAdapter) Capabilities(ctx context.Context, modelID string) (ModelCapabilities, error) {
	return ModelCapabilities{
		ContextWindow:      128000,
		MaxOutputTokens:    4096,
		SupportsJSONSchema: true,
		SupportsStreaming:  true,
		SupportsTools:      false, // In our architecture, AI does not directly call tools (AT-019)
		ReportsUsage:       true,
	}, nil
}

func (m *MockDeterministicAdapter) Estimate(ctx context.Context, req GenerationRequest) (UsageEstimate, error) {
	inputTokens := (len(req.SystemPrompt) + len(req.UntrustedData) + len(fmt.Sprintf("%v", req.ScopedFacts))) / 4
	if inputTokens < 100 {
		inputTokens = 100
	}

	outputTokens := req.MaxTokens
	if outputTokens <= 0 || outputTokens > 1000 {
		outputTokens = 500
	}

	// Example rates: $1.00 per 1M input tokens = 1 micro-USD per 1000 tokens
	// $3.00 per 1M output tokens = 3 micro-USD per 1000 tokens
	costMicroUSD := int64((inputTokens*1 + outputTokens*3) / 1000)
	if costMicroUSD == 0 {
		costMicroUSD = 10 // Minimum 10 micro-USD
	}

	return UsageEstimate{
		EstimatedInputTokens:  inputTokens,
		EstimatedOutputTokens: outputTokens,
		EstimatedCostMicroUSD: costMicroUSD,
	}, nil
}

func (m *MockDeterministicAdapter) Generate(ctx context.Context, req GenerationRequest) (*GenerationResult, error) {
	est, _ := m.Estimate(ctx, req)

	// Check if a custom mock response was registered
	var outputJSON string
	if custom, ok := m.customResponses[req.TaskKey]; ok {
		outputJSON = custom
	} else {
		outputJSON = m.generateDefaultTaskOutput(req)
	}

	// Validate that output is well-formed JSON
	var parsed json.RawMessage
	if err := json.Unmarshal([]byte(outputJSON), &parsed); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidStructuredOutput, err)
	}

	actualInputTokens := est.EstimatedInputTokens
	actualOutputTokens := len(outputJSON) / 4
	if actualOutputTokens < 50 {
		actualOutputTokens = 50
	}

	costMicroUSD := int64((actualInputTokens*1 + actualOutputTokens*3) / 1000)
	if costMicroUSD == 0 {
		costMicroUSD = 10
	}

	return &GenerationResult{
		OutputContent: parsed,
		RawText:       outputJSON,
		Usage: TokenUsage{
			InputTokens:  actualInputTokens,
			OutputTokens: actualOutputTokens,
			TotalTokens:  actualInputTokens + actualOutputTokens,
		},
		CostMicroUSD: costMicroUSD,
		ModelUsed:    req.ModelID,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

func (m *MockDeterministicAdapter) generateDefaultTaskOutput(req GenerationRequest) string {
	switch req.TaskKey {
	case TaskProfileExtract:
		return `{
			"candidate_name": "Alice Candidate",
			"headline": "Full-Stack Software Engineer",
			"skills": ["Go", "TypeScript", "PostgreSQL"],
			"extracted_experience_years": 5
		}`
	case TaskJobFitExplain:
		return `{
			"match_score": 92,
			"matching_skills": ["Go", "PostgreSQL"],
			"missing_skills": [],
			"summary": "Strong match based on verified backend experience."
		}`
	case TaskResumeTailor:
		return `{
			"target_role": "Senior Backend Engineer",
			"tailored_summary": "Experienced Go engineer specializing in high-throughput transactional workflows.",
			"selected_fact_ids": ["fact-1", "fact-2"]
		}`
	case TaskCoverLetter:
		return `{
			"salutation": "Dear Hiring Team,",
			"body": "I am writing to express my strong interest in the Backend Engineer role.",
			"referenced_facts": ["fact-1"]
		}`
	case TaskAppAnswer:
		return `{
			"question": "Do you have experience with Go?",
			"answer": "Yes, 5 years of production experience.",
			"supporting_fact_id": "fact-1",
			"status": "ready"
		}`
	case TaskLinkedInOptimize:
		return `{
			"suggested_headline": "Senior Backend Engineer | Go, Distributed Systems & Next.js",
			"suggested_about": "Building resilient multi-tenant architectures and high-velocity applications.",
			"diff_summary": "Added Go and Distributed Systems keywords from confirmed career facts."
		}`
	case TaskConnectionDraft:
		return `{
			"recipient_name": "Alex",
			"message": "Hi Alex, noticed your team is building with Go and distributed event streaming. Would love to connect!",
			"slot_cost": 1
		}`
	case TaskSocialPost:
		return `{
			"content": "Designing resilient event relays: why PostgreSQL transactional outbox paired with Redis Streams prevents duplicate side effects.",
			"hashtags": ["#SoftwareEngineering", "#SystemDesign", "#GoLang"],
			"character_count": 154
		}`
	case TaskContentIdeas:
		return `{
			"ideas": [
				{"title": "Zero-Hallucination AI Gateways", "angle": "Architectural contract testing"},
				{"title": "Multi-Tenant Isolation in Meilisearch", "angle": "Privacy bounds"}
			]
		}`
	case TaskResearchBrief:
		return `{
			"topic": "Current hiring trends in distributed computing",
			"key_findings": ["Growing demand for Go and Rust in event-driven architectures"],
			"source_citations": ["https://example.com/report"]
		}`
	default:
		return `{"status": "completed", "task": "` + string(req.TaskKey) + `"}`
	}
}
