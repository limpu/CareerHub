package quota

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/social-platform/services/core/internal/provider"
)

// QuotaService orchestrates multi-gate authorization, atomic reservations,
// nested limits, account circuit breaking, and usage metering.
type QuotaService struct {
	store              QuotaStore
	capabilityRegistry *provider.CapabilityRegistry
	mu                 sync.RWMutex
	rules              map[ActionMetric]QuotaDefinition
}

// NewQuotaService creates a new QuotaService initialized with standard planning limits.
func NewQuotaService(store QuotaStore, registry *provider.CapabilityRegistry) *QuotaService {
	svc := &QuotaService{
		store:              store,
		capabilityRegistry: registry,
		rules:              make(map[ActionMetric]QuotaDefinition),
	}

	// Register default conservative planning quotas from CONTEXT.md Section 8.2
	svc.registerDefaultRules()
	return svc
}

// registerDefaultRules defines the initial planning budgets and nested hierarchies.
func (s *QuotaService) registerDefaultRules() {
	// LinkedIn Actions (REQ-010, AT-008)
	s.rules[MetricLinkedInConnectionRequest] = QuotaDefinition{
		Metric:                     MetricLinkedInConnectionRequest,
		Limit:                      5, // 5/day
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: true,
		AutoExecutionDefault:       0,
	}

	s.rules[MetricLinkedInFollow] = QuotaDefinition{
		Metric:                     MetricLinkedInFollow,
		Limit:                      5, // 5/day
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: true,
		AutoExecutionDefault:       0,
	}

	s.rules[MetricLinkedInDMSend] = QuotaDefinition{
		Metric:                     MetricLinkedInDMSend,
		Limit:                      10, // 10/day total
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: true,
		AutoExecutionDefault:       0,
	}

	// Recruiter outreach is nested inside MetricLinkedInDMSend: 3/day inside the 10-DM total
	s.rules[MetricLinkedInDMRecruiter] = QuotaDefinition{
		Metric:                     MetricLinkedInDMRecruiter,
		ParentMetric:               MetricLinkedInDMSend,
		Limit:                      3, // 3/day nested
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: true,
		AutoExecutionDefault:       0,
	}

	// Career Actions
	s.rules[MetricCareerJobApplyManual] = QuotaDefinition{
		Metric:                     MetricCareerJobApplyManual,
		Limit:                      10,
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: false,
		AutoExecutionDefault:       10,
	}

	s.rules[MetricCareerJobApplyAutomated] = QuotaDefinition{
		Metric:                     MetricCareerJobApplyAutomated,
		Limit:                      0, // Automated application disabled by default
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: true,
		AutoExecutionDefault:       0,
	}

	// Social Actions
	s.rules[MetricSocialPublishPost] = QuotaDefinition{
		Metric:                     MetricSocialPublishPost,
		Limit:                      5,
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: true,
		AutoExecutionDefault:       0,
	}

	s.rules[MetricSocialFollow] = QuotaDefinition{
		Metric:                     MetricSocialFollow,
		Limit:                      5,
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: true,
		AutoExecutionDefault:       0,
	}

	s.rules[MetricSocialCommentDraft] = QuotaDefinition{
		Metric:                     MetricSocialCommentDraft,
		Limit:                      10,
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: false,
		AutoExecutionDefault:       10,
	}

	s.rules[MetricSocialMessageOptIn] = QuotaDefinition{
		Metric:                     MetricSocialMessageOptIn,
		Limit:                      5,
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: true,
		AutoExecutionDefault:       0,
	}

	s.rules[MetricSocialResearchRead] = QuotaDefinition{
		Metric:                     MetricSocialResearchRead,
		Limit:                      100,
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: false,
		AutoExecutionDefault:       100,
	}
}

// SetCustomRule allows overriding or adding custom quota definitions.
func (s *QuotaService) SetCustomRule(rule QuotaDefinition) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[rule.Metric] = rule
}

// GetRule returns the quota definition for a metric.
func (s *QuotaService) GetRule(metric ActionMetric) (QuotaDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rule, exists := s.rules[metric]
	return rule, exists
}

// Reserve executes multi-gate validation and atomically reserves quota slots (REQ-021, AT-008, AT-010).
func (s *QuotaService) Reserve(ctx context.Context, req ReservationRequest) (*Reservation, error) {
	now := time.Now().UTC()

	// 1. Gate 1: Account Circuit-Breaker State (AT-010)
	accountState, err := s.store.GetAccountState(ctx, req.AccountID)
	if err != nil {
		return nil, ErrStoreUnavailable
	}
	if accountState.Status == AccountStatusPaused {
		return nil, ErrAccountPaused
	}
	if accountState.Status == AccountStatusChallengeRequired {
		return nil, ErrChallengeRequired
	}
	if accountState.Status == AccountStatusRateLimited {
		if accountState.RetryAfter != nil && now.Before(*accountState.RetryAfter) {
			return nil, ErrRateLimited
		}
	}

	// 2. Gate 2: Quota Rule Definition
	s.mu.RLock()
	rule, exists := s.rules[req.Metric]
	var parentRule *QuotaDefinition
	if exists && rule.ParentMetric != "" {
		if pr, pExists := s.rules[rule.ParentMetric]; pExists {
			parentRule = &pr
		}
	}
	s.mu.RUnlock()

	if !exists {
		// If no explicit quota rule exists, default to 1 unit with 24h window
		rule = QuotaDefinition{
			Metric: req.Metric,
			Limit:  1,
			Window: 24 * time.Hour,
		}
	}

	// 3. Gate 3: Provider Capability & Scope Validation (REQ-021, AT-010)
	if rule.RequiresApprovedCapability && s.capabilityRegistry != nil && req.Provider != "" {
		actionID := s.metricToActionID(req.Metric)
		if !s.capabilityRegistry.IsActionSupported(req.Provider, actionID) {
			return nil, ErrActionCapabilityUnauthorized
		}
	}

	// 4. Gate 4: Atomic Reservation via Store (AT-008, AT-009)
	ttl := req.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute // Default lease expiration for reservation
	}

	res := &Reservation{
		ID:           generateID(),
		AccountID:    req.AccountID,
		WorkspaceID:  req.WorkspaceID,
		Module:       req.Module,
		Metric:       req.Metric,
		ParentMetric: rule.ParentMetric,
		Amount:       req.Amount,
		Status:       ReservationPending,
		CreatedAt:    now,
		ExpiresAt:    now.Add(ttl),
	}

	if err := s.store.Reserve(ctx, res, rule, parentRule, now); err != nil {
		return nil, err
	}

	return res, nil
}

// Settle commits a reservation into settled usage upon successful external dispatch.
func (s *QuotaService) Settle(ctx context.Context, reservationID string) error {
	now := time.Now().UTC()
	return s.store.Settle(ctx, reservationID, now)
}

// Release voids a reservation for a proven non-executed failure.
func (s *QuotaService) Release(ctx context.Context, reservationID string) error {
	now := time.Now().UTC()
	return s.store.Release(ctx, reservationID, now)
}

// RetainAmbiguous retains a reservation when external execution outcome is uncertain (e.g. timeout)
// preventing blind duplicates and preserving quota accountability (REQ-010, AT-006).
func (s *QuotaService) RetainAmbiguous(ctx context.Context, reservationID string, reason string) error {
	now := time.Now().UTC()
	return s.store.Retain(ctx, reservationID, reason, now)
}

// RecordChallenge puts the account into a challenge-required state, halting automated actions (AT-010).
func (s *QuotaService) RecordChallenge(ctx context.Context, accountID string, providerName string, reason string) error {
	state := &AccountState{
		AccountID:       accountID,
		Provider:        providerName,
		Status:          AccountStatusChallengeRequired,
		ChallengeReason: reason,
		UpdatedAt:       time.Now().UTC(),
	}
	return s.store.SetAccountState(ctx, state)
}

// RecordRateLimit records a 429 rate limit backoff until retryAfter (AT-010).
func (s *QuotaService) RecordRateLimit(ctx context.Context, accountID string, providerName string, retryAfter time.Time) error {
	state := &AccountState{
		AccountID:  accountID,
		Provider:   providerName,
		Status:     AccountStatusRateLimited,
		RetryAfter: &retryAfter,
		UpdatedAt:  time.Now().UTC(),
	}
	return s.store.SetAccountState(ctx, state)
}

// PauseAccount halts all actions on an account (e.g., on warning or user request).
func (s *QuotaService) PauseAccount(ctx context.Context, accountID string, providerName string, reason string) error {
	state := &AccountState{
		AccountID:       accountID,
		Provider:        providerName,
		Status:          AccountStatusPaused,
		ChallengeReason: reason,
		UpdatedAt:       time.Now().UTC(),
	}
	return s.store.SetAccountState(ctx, state)
}

// ResumeAccount restores an account to active status after issue resolution.
func (s *QuotaService) ResumeAccount(ctx context.Context, accountID string, providerName string) error {
	state := &AccountState{
		AccountID: accountID,
		Provider:  providerName,
		Status:    AccountStatusActive,
		UpdatedAt: time.Now().UTC(),
	}
	return s.store.SetAccountState(ctx, state)
}

// GetUsageSummary returns the complete consumption and reservation metrics for an account and action.
func (s *QuotaService) GetUsageSummary(ctx context.Context, accountID string, metric ActionMetric) (*UsageSummary, error) {
	now := time.Now().UTC()
	s.mu.RLock()
	rule, exists := s.rules[metric]
	s.mu.RUnlock()

	if !exists {
		rule = QuotaDefinition{
			Metric: metric,
			Limit:  1,
			Window: 24 * time.Hour,
		}
	}

	consumed, reserved, err := s.store.GetUsage(ctx, accountID, metric, rule.Window, now)
	if err != nil {
		return nil, err
	}

	available := rule.Limit - (consumed + reserved)
	if available < 0 {
		available = 0
	}

	summary := &UsageSummary{
		AccountID:    accountID,
		Metric:       metric,
		Limit:        rule.Limit,
		Consumed:     consumed,
		Reserved:     reserved,
		Available:    available,
		Window:       rule.Window,
		ParentMetric: rule.ParentMetric,
	}

	// If nested, also fetch parent summary
	if rule.ParentMetric != "" {
		parentSummary, err := s.GetUsageSummary(ctx, accountID, rule.ParentMetric)
		if err == nil {
			summary.ParentUsage = parentSummary
		}
	}

	return summary, nil
}

// metricToActionID maps ActionMetric to provider capability action IDs in CapabilityRegistry.
func (s *QuotaService) metricToActionID(metric ActionMetric) string {
	switch metric {
	case MetricLinkedInConnectionRequest:
		return "linkedin.connection.request"
	case MetricLinkedInFollow:
		return "linkedin.follow"
	case MetricLinkedInDMSend, MetricLinkedInDMRecruiter:
		return "linkedin.dm.send"
	case MetricCareerJobApplyAutomated:
		return "job.apply"
	case MetricSocialPublishPost:
		return "linkedin.post.create"
	default:
		return string(metric)
	}
}

func generateID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
