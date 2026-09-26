package provider

import (
	"fmt"
	"sync"
)

// CapabilityRegistry manages verified provider manifests and action authorization (REQ-021, REQ-024).
type CapabilityRegistry struct {
	mu        sync.RWMutex
	manifests map[string]*ProviderManifest
}

func NewCapabilityRegistry() *CapabilityRegistry {
	reg := &CapabilityRegistry{
		manifests: make(map[string]*ProviderManifest),
	}
	reg.registerDefaultManifests()
	return reg
}

// registerDefaultManifests populates vetted provider contracts with review dates (REQ-024).
func (r *CapabilityRegistry) registerDefaultManifests() {
	// 1. LinkedIn Connector Manifest
	r.RegisterManifest(&ProviderManifest{
		ID:              "linkedin",
		Name:            "LinkedIn Official API Connector",
		Channel:         "linkedin",
		SupportedScopes: []string{"openid", "profile", "email", "w_member_social"},
		Capabilities: []ActionCapability{
			{
				Action:        "linkedin.profile.read",
				RequiredScope: "profile",
				Description:   "Read authentic LinkedIn profile details",
			},
			{
				Action:        "linkedin.post.create",
				RequiredScope: "w_member_social",
				Description:   "Publish organic posts to member profile feed",
			},
			{
				Action:        "linkedin.dm.send",
				RequiredScope: "w_member_social",
				Description:   "Send direct message to verified connection",
			},
			{
				Action:        "linkedin.connection.request",
				RequiredScope: "w_member_social",
				Description:   "Send invitation or connection request",
			},
		},
		ReviewDate:       "2026-09-17",
		Enabled:          true,
		DocumentationURL: "https://learn.microsoft.com/en-us/linkedin/consumer/integrations/self-serve/share-on-linkedin",
	})

	// 2. Google Sheets Connector Manifest
	r.RegisterManifest(&ProviderManifest{
		ID:              "google_sheets",
		Name:            "Google Sheets Export Connector",
		Channel:         "google_sheets",
		SupportedScopes: []string{"https://www.googleapis.com/auth/spreadsheets"},
		Capabilities: []ActionCapability{
			{
				Action:        "sheets.export",
				RequiredScope: "https://www.googleapis.com/auth/spreadsheets",
				Description:   "Idempotent export of job tracking and application records",
			},
		},
		ReviewDate:       "2026-09-17",
		Enabled:          true,
		DocumentationURL: "https://developers.google.com/sheets/api/guides/concepts",
	})

	// 3. Job Boards Connector Manifest
	r.RegisterManifest(&ProviderManifest{
		ID:              "job_boards",
		Name:            "Job Boards API & Portal Connector",
		Channel:         "job_board",
		SupportedScopes: []string{"job_search", "job_apply"},
		Capabilities: []ActionCapability{
			{
				Action:        "job.search",
				RequiredScope: "job_search",
				Description:   "Query job postings from verified provider search feeds",
			},
			{
				Action:        "job.apply",
				RequiredScope: "job_apply",
				Description:   "Direct partner API job application submission",
			},
		},
		ReviewDate:       "2026-09-17",
		Enabled:          true,
		DocumentationURL: "https://docs.social-platform.local/providers/job-boards",
	})
}

// RegisterManifest registers or updates a provider manifest.
func (r *CapabilityRegistry) RegisterManifest(manifest *ProviderManifest) {
	r.mu.Lock()
	defer r.mu.Unlock()

	clone := *manifest
	r.manifests[manifest.ID] = &clone
}

// GetManifest retrieves a provider manifest by its ID.
func (r *CapabilityRegistry) GetManifest(providerID string) (*ProviderManifest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	manifest, ok := r.manifests[providerID]
	if !ok {
		return nil, ErrProviderNotFound
	}
	clone := *manifest
	return &clone, nil
}

// ListManifests returns all registered provider manifests.
func (r *CapabilityRegistry) ListManifests() []*ProviderManifest {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*ProviderManifest
	for _, m := range r.manifests {
		clone := *m
		list = append(list, &clone)
	}
	return list
}

// AssertCapability verifies that an action is supported by the provider AND granted by scopes (REQ-021, AT-010).
func (r *CapabilityRegistry) AssertCapability(providerID, action string, grantedScopes []string) error {
	r.mu.RLock()
	manifest, ok := r.manifests[providerID]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, providerID)
	}

	if !manifest.Enabled {
		return fmt.Errorf("%w: provider %s is disabled", ErrActionUnsupported, providerID)
	}

	// 1. Check if the action is explicitly supported in the manifest
	var targetCap *ActionCapability
	for _, cap := range manifest.Capabilities {
		if cap.Action == action {
			targetCap = &cap
			break
		}
	}

	if targetCap == nil {
		// AT-010 Invariant: Unsupported actions must fail closed and NEVER be simulated
		return fmt.Errorf("%w: action '%s' is not supported by provider '%s'", ErrActionUnsupported, action, providerID)
	}

	// 2. Check if granted scopes satisfy the required scope for this action
	if targetCap.RequiredScope == "" {
		return nil // No scope required
	}

	scopeSet := make(map[string]bool)
	for _, s := range grantedScopes {
		scopeSet[s] = true
	}

	if !scopeSet[targetCap.RequiredScope] {
		return fmt.Errorf("%w: action '%s' requires scope '%s'", ErrCapabilityNotGranted, action, targetCap.RequiredScope)
	}

	return nil
}

// IsActionSupported checks if an action is registered and enabled on a provider.
func (r *CapabilityRegistry) IsActionSupported(providerID, action string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	manifest, ok := r.manifests[providerID]
	if !ok || !manifest.Enabled {
		return false
	}
	for _, cap := range manifest.Capabilities {
		if cap.Action == action {
			return true
		}
	}
	return false
}
