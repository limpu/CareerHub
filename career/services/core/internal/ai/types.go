package ai

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrModelConsentRequired        = errors.New("explicit model consent is required before dispatching private data to AI provider (REQ-016)")
	ErrFabricatedFactsDetected     = errors.New("output contains fabricated qualifications or ungrounded claims (REQ-016, AT-003)")
	ErrNeedsUserInput              = errors.New("required factual field cannot be inferred without explicit user input (AT-003)")
	ErrPromptInjectionBlocked      = errors.New("untrusted external text attempted prompt injection or unauthorized tool authority (AT-019)")
	ErrBudgetExceeded              = errors.New("ai token or cost budget exceeded (REQ-021)")
	ErrUnsupportedModelCapability  = errors.New("selected model does not support required task capabilities")
	ErrInvalidStructuredOutput     = errors.New("ai provider output did not conform to required JSON schema")
)

// TaskKey identifies the logical agent task being executed (AI-AGENT-ARCHITECTURE Section 6).
type TaskKey string

const (
	TaskProfileExtract    TaskKey = "profile_extract"
	TaskJobFitExplain     TaskKey = "job_fit_explain"
	TaskResumeTailor      TaskKey = "resume_tailor"
	TaskCoverLetter       TaskKey = "cover_letter"
	TaskAppAnswer         TaskKey = "app_answer"
	TaskLinkedInOptimize  TaskKey = "linkedin_optimize"
	TaskConnectionDraft   TaskKey = "connection_draft"
	TaskSocialPost        TaskKey = "social_post"
	TaskContentIdeas      TaskKey = "content_ideas"
	TaskResearchBrief     TaskKey = "research_brief"
)

// FactCategory categorizes confirmed user career facts.
type FactCategory string

const (
	CategoryExperience    FactCategory = "experience"
	CategoryEducation     FactCategory = "education"
	CategorySkill         FactCategory = "skill"
	CategoryCertification FactCategory = "certification"
	CategoryMetric        FactCategory = "metric"
	CategoryLegalStatus   FactCategory = "legal_status"
	CategorySalary        FactCategory = "salary"
	CategorySponsorship   FactCategory = "sponsorship"
)

// ScopedFact represents a verified, confirmed user fact used for fact-grounding (REQ-016).
type ScopedFact struct {
	ID         string       `json:"id"`
	Category   FactCategory `json:"category"`
	Statement  string       `json:"statement"`
	VerifiedAt time.Time    `json:"verified_at"`
}

// ModelCapabilities captures supported features of a model endpoint.
type ModelCapabilities struct {
	ContextWindow       int  `json:"context_window"`
	MaxOutputTokens     int  `json:"max_output_tokens"`
	SupportsJSONSchema  bool `json:"supports_json_schema"`
	SupportsStreaming   bool `json:"supports_streaming"`
	SupportsTools       bool `json:"supports_tools"`
	ReportsUsage        bool `json:"reports_usage"`
}

// TokenUsage details the metered token consumption of a generation.
type TokenUsage struct {
	InputTokens     int `json:"input_tokens"`
	OutputTokens    int `json:"output_tokens"`
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
	TotalTokens     int `json:"total_tokens"`
}

// UsageEstimate represents preflight token and monetary estimates.
type UsageEstimate struct {
	EstimatedInputTokens  int   `json:"estimated_input_tokens"`
	EstimatedOutputTokens int   `json:"estimated_output_tokens"`
	EstimatedCostMicroUSD int64 `json:"estimated_cost_micro_usd"`
}

// GroundingReport provides proof that generated output strictly adheres to verified facts (REQ-016, AT-003).
type GroundingReport struct {
	IsValid          bool     `json:"is_valid"`
	GroundedClaims   []string `json:"grounded_claims"`
	UnverifiedClaims []string `json:"unverified_claims,omitempty"`
	NeedsInputFields []string `json:"needs_input_fields,omitempty"`
	Violations       []string `json:"violations,omitempty"`
}

// GenerationRequest contains typed inputs for an AI generation.
type GenerationRequest struct {
	TaskKey        TaskKey           `json:"task_key"`
	WorkspaceID    string            `json:"workspace_id"`
	ActorID        string            `json:"actor_id"`
	ModelID        string            `json:"model_id"`
	ScopedFacts    []ScopedFact      `json:"scoped_facts"`
	SystemPrompt   string            `json:"system_prompt"`
	UntrustedData  string            `json:"untrusted_data"` // Bounded external text (resumes, job text, etc.)
	MaxTokens      int               `json:"max_tokens"`
	Temperature    float64           `json:"temperature"`
	HasModelConsent bool             `json:"has_model_consent"`
	IdempotencyKey string            `json:"idempotency_key"`
}

// GenerationResult represents the verified, structured output of an AI generation.
type GenerationResult struct {
	OutputContent   json.RawMessage  `json:"output_content"`
	RawText         string           `json:"raw_text"`
	Usage           TokenUsage       `json:"usage"`
	CostMicroUSD    int64            `json:"cost_micro_usd"`
	GroundingReport GroundingReport  `json:"grounding_report"`
	ModelUsed       string           `json:"model_used"`
	GeneratedAt     time.Time        `json:"generated_at"`
}
