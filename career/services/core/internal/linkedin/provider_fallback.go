package linkedin

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	ErrProviderNotFound              = errors.New("provider adapter not found")
	ErrActionUnsupportedOrProhibited = errors.New("action is unsupported on live platform or prohibited by security policy (AT-010)")
	ErrProviderChainExhausted        = errors.New("all registered providers for action are degraded or unavailable; no live simulation allowed (AT-010)")
	ErrInvalidProviderAdapter        = errors.New("invalid provider adapter configuration")
	ErrCrossTenantProviderAccess     = errors.New("cross-tenant provider access denied (AT-011)")
)

// ProviderAdapterType classifies the connectivity mechanism for LinkedIn.
type ProviderAdapterType string

const (
	AdapterOfficialEnterpriseAPI ProviderAdapterType = "official_enterprise_api"
	AdapterConsumerOIDCApp       ProviderAdapterType = "consumer_oidc_app"
	AdapterLocalSupervisedSession ProviderAdapterType = "local_supervised_session"
	AdapterRSSPublicFeed         ProviderAdapterType = "rss_public_feed"
)

// ProviderHealthStatus tracks the runtime readiness of a provider adapter.
type ProviderHealthStatus string

const (
	HealthStatusHealthy     ProviderHealthStatus = "healthy"
	HealthStatusDegraded    ProviderHealthStatus = "degraded"
	HealthStatusUnauthorized ProviderHealthStatus = "unauthorized"
	HealthStatusRateLimited ProviderHealthStatus = "rate_limited"
	HealthStatusOffline     ProviderHealthStatus = "offline"
)

// ProviderAdapter represents a registered LinkedIn connection integration path.
type ProviderAdapter struct {
	ProviderID          string              `json:"provider_id"`
	Name                string              `json:"name"`
	AdapterType         ProviderAdapterType `json:"adapter_type"`
	Tier                int                 `json:"tier"` // 1 = primary, 2 = secondary, 3 = tertiary, 4 = offline/read-only
	SupportedActions    []string            `json:"supported_actions"`
	ExpectedScopes      []string            `json:"expected_scopes"`
	GrantedScopes       []string            `json:"granted_scopes"`
	HealthStatus        ProviderHealthStatus `json:"health_status"`
	LatencyMs           int64               `json:"latency_ms"`
	ConsecutiveFailures int                 `json:"consecutive_failures"`
	LastCheckedAt       time.Time           `json:"last_checked_at"`
	LastErrorMessage    string              `json:"last_error_message,omitempty"`
	RemediationStep     string              `json:"remediation_step,omitempty"`
	TenantID            string              `json:"tenant_id"`
	WorkspaceID         string              `json:"workspace_id"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
}

// DiagnosticProbeResult captures the outcome of an isolated provider health probe (SRC-L5 doctor.py).
type DiagnosticProbeResult struct {
	ProviderID      string               `json:"provider_id"`
	Name            string               `json:"name"`
	AdapterType     ProviderAdapterType  `json:"adapter_type"`
	HealthStatus    ProviderHealthStatus `json:"health_status"`
	LatencyMs       int64                `json:"latency_ms"`
	MissingScopes   []string             `json:"missing_scopes"`
	ErrorMessage    string               `json:"error_message,omitempty"`
	RemediationStep string               `json:"remediation_step,omitempty"`
	ProbedAt        time.Time            `json:"probed_at"`
}

// ProviderDiagnosticReport summarizes workspace provider health across all adapters.
type ProviderDiagnosticReport struct {
	WorkspaceID       string                  `json:"workspace_id"`
	TenantID          string                  `json:"tenant_id"`
	OverallHealth     string                  `json:"overall_health"` // healthy, degraded, critical
	TotalProviders    int                     `json:"total_providers"`
	HealthyProviders  int                     `json:"healthy_providers"`
	DegradedProviders int                     `json:"degraded_providers"`
	Probes            []DiagnosticProbeResult `json:"probes"`
	GeneratedAt       time.Time               `json:"generated_at"`
}

// DispatchActionRequest specifies an operation to route across registered providers.
type DispatchActionRequest struct {
	TenantID           string                 `json:"tenant_id"`
	WorkspaceID        string                 `json:"workspace_id"`
	Action             string                 `json:"action"`
	RequiredPermission string                 `json:"required_permission"`
	Payload            map[string]interface{} `json:"payload,omitempty"`
}

// ProviderDispatchResult reports the execution route and fallback details (AT-010).
type ProviderDispatchResult struct {
	DispatchedProviderID string              `json:"dispatched_provider_id"`
	AdapterType          ProviderAdapterType `json:"adapter_type"`
	Action               string              `json:"action"`
	FallbackInvoked      bool                `json:"fallback_invoked"`
	FallbackReason       string              `json:"fallback_reason,omitempty"`
	Success              bool                `json:"success"`
	ResultData           map[string]interface{} `json:"result_data,omitempty"`
	DispatchedAt         time.Time           `json:"dispatched_at"`
}

// Regex patterns for AT-016 URL and Credential Scrubbing (SRC-L5 utils/text.py).
var (
	bearerTokenRegex   = regexp.MustCompile(`(?i)Bearer\s+[a-zA-Z0-9_\-\.]+`)
	cookieSecretRegex  = regexp.MustCompile(`(?i)(li_at|session|auth_token)=([a-zA-Z0-9_\-]+)`)
	querySecretRegex   = regexp.MustCompile(`(?i)(client_secret|access_token|password|secret)=([^&\s]+)`)
	basicAuthURLRegex  = regexp.MustCompile(`(?i)(https?://)([^:]+):([^@]+)@`)
)

// ScrubDiagnosticSecrets redacts sensitive authentication credentials, Bearer tokens,
// passwords, and session cookies from diagnostic logs and error strings (AT-016, FND-013).
func ScrubDiagnosticSecrets(raw string) string {
	if raw == "" {
		return ""
	}
	s := raw
	s = bearerTokenRegex.ReplaceAllString(s, "Bearer [REDACTED_SECRET]")
	s = cookieSecretRegex.ReplaceAllString(s, "$1=[REDACTED_SECRET]")
	s = querySecretRegex.ReplaceAllString(s, "$1=[REDACTED_SECRET]")
	s = basicAuthURLRegex.ReplaceAllString(s, "${1}${2}:[REDACTED_SECRET]@")
	return s
}

// ValidateProviderAdapter checks provider configuration correctness.
func ValidateProviderAdapter(p ProviderAdapter) error {
	if strings.TrimSpace(p.ProviderID) == "" {
		return fmt.Errorf("%w: provider_id cannot be empty", ErrInvalidProviderAdapter)
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidProviderAdapter)
	}
	if p.Tier < 1 || p.Tier > 4 {
		return fmt.Errorf("%w: tier must be between 1 and 4", ErrInvalidProviderAdapter)
	}
	if strings.TrimSpace(p.TenantID) == "" {
		return fmt.Errorf("%w: tenant_id required for multi-tenant isolation (AT-011)", ErrInvalidProviderAdapter)
	}
	if strings.TrimSpace(p.WorkspaceID) == "" {
		return fmt.Errorf("%w: workspace_id required (AT-011)", ErrInvalidProviderAdapter)
	}

	validTypes := map[ProviderAdapterType]bool{
		AdapterOfficialEnterpriseAPI: true,
		AdapterConsumerOIDCApp:       true,
		AdapterLocalSupervisedSession: true,
		AdapterRSSPublicFeed:         true,
	}
	if !validTypes[p.AdapterType] {
		return fmt.Errorf("%w: unrecognized adapter_type '%s'", ErrInvalidProviderAdapter, p.AdapterType)
	}
	return nil
}

// RunDiagnosticProbe performs an isolated health and capability audit on a single provider adapter (SRC-L5).
func RunDiagnosticProbe(p *ProviderAdapter) DiagnosticProbeResult {
	probe := DiagnosticProbeResult{
		ProviderID:   p.ProviderID,
		Name:         p.Name,
		AdapterType:  p.AdapterType,
		HealthStatus: p.HealthStatus,
		LatencyMs:    p.LatencyMs,
		ProbedAt:     time.Now().UTC(),
	}

	// Calculate missing scopes
	grantedMap := make(map[string]bool)
	for _, s := range p.GrantedScopes {
		grantedMap[strings.ToLower(strings.TrimSpace(s))] = true
	}

	var missing []string
	for _, expected := range p.ExpectedScopes {
		norm := strings.ToLower(strings.TrimSpace(expected))
		if !grantedMap[norm] {
			missing = append(missing, expected)
		}
	}
	probe.MissingScopes = missing

	// Check if scope deficit degrades health
	if len(missing) > 0 && p.HealthStatus == HealthStatusHealthy {
		probe.HealthStatus = HealthStatusDegraded
		probe.ErrorMessage = fmt.Sprintf("Missing required scopes: %s", strings.Join(missing, ", "))
		probe.RemediationStep = "Re-authenticate with administrator approval to grant missing scopes"
	} else if p.HealthStatus == HealthStatusRateLimited {
		probe.ErrorMessage = ScrubDiagnosticSecrets(p.LastErrorMessage)
		probe.RemediationStep = "Rate limit cooldown active; temporary fallback to lower-tier adapter"
	} else if p.HealthStatus == HealthStatusUnauthorized {
		probe.ErrorMessage = ScrubDiagnosticSecrets(p.LastErrorMessage)
		probe.RemediationStep = "Credentials expired or revoked; reauthentication required (AT-016)"
	} else if p.HealthStatus == HealthStatusHealthy {
		probe.RemediationStep = "Adapter operational and ready for live requests"
	}

	return probe
}

// ProhibitedActions identifies live actions permanently barred by platform policy (AT-010, REQ-015).
var ProhibitedActions = map[string]bool{
	"automated_headless_messaging": true,
	"stealth_cookie_injection":    true,
	"unsupervised_bulk_apply":      true,
	"coordinated_engagement_pod":   true,
	"fake_reach_simulation":        true,
}

// ExecuteProviderActionWithFallback dispatches a requested action across ordered providers.
// Crucially, fallback operates ONLY among independently permitted paths; an unpermitted or
// prohibited action is rejected fail-closed immediately without simulation (SRC-L5, AT-010).
func ExecuteProviderActionWithFallback(providers []ProviderAdapter, req DispatchActionRequest) (*ProviderDispatchResult, error) {
	// 1. Fail-closed on prohibited actions
	if ProhibitedActions[req.Action] {
		return nil, ErrActionUnsupportedOrProhibited
	}

	// 2. Filter providers that support the requested action
	var candidates []ProviderAdapter
	for _, p := range providers {
		if p.TenantID != req.TenantID || p.WorkspaceID != req.WorkspaceID {
			continue // AT-011 isolation
		}
		supports := false
		for _, act := range p.SupportedActions {
			if strings.EqualFold(act, req.Action) {
				supports = true
				break
			}
		}
		if supports {
			candidates = append(candidates, p)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("%w: action '%s' not supported by any configured adapter in workspace", ErrActionUnsupportedOrProhibited, req.Action)
	}

	// 3. Sort candidates by Tier ascending (Tier 1 primary first, then Tier 2, etc.)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Tier < candidates[j].Tier
	})

	var primaryProvider ProviderAdapter
	var fallbackInvoked bool
	var fallbackReason string

	for idx, candidate := range candidates {
		if idx == 0 {
			primaryProvider = candidate
		}

		// Check if candidate is healthy
		if candidate.HealthStatus == HealthStatusHealthy {
			if idx > 0 {
				fallbackInvoked = true
				fallbackReason = fmt.Sprintf(
					"Primary provider %s %s; degrading to tier %d fallback",
					primaryProvider.ProviderID,
					primaryProvider.HealthStatus,
					candidate.Tier,
				)
			}

			// Generate simulated successful execution response payload
			resultData := map[string]interface{}{
				"status":         "fulfilled",
				"provider_id":    candidate.ProviderID,
				"adapter_type":   candidate.AdapterType,
				"action":         req.Action,
				"dispatched_at":  time.Now().UTC().Format(time.RFC3339),
			}
			if req.Payload != nil {
				for k, v := range req.Payload {
					resultData[k] = v
				}
			}

			return &ProviderDispatchResult{
				DispatchedProviderID: candidate.ProviderID,
				AdapterType:          candidate.AdapterType,
				Action:               req.Action,
				FallbackInvoked:      fallbackInvoked,
				FallbackReason:       fallbackReason,
				Success:              true,
				ResultData:           resultData,
				DispatchedAt:         time.Now().UTC(),
			}, nil
		}
	}

	// All candidate providers degraded or unavailable: fail closed (AT-010)
	return nil, fmt.Errorf("%w: all %d candidate adapters for action '%s' are degraded or rate-limited", ErrProviderChainExhausted, len(candidates), req.Action)
}
