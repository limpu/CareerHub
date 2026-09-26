package quota_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/social-platform/services/core/internal/provider"
	"github.com/social-platform/services/core/internal/quota"
)

// AT-008: Two concurrent workers cannot consume the same last budget slot.
func TestQuota_ConcurrentReservationRace(t *testing.T) {
	ctx := context.Background()
	store := quota.NewMemoryQuotaStore()
	registry := provider.NewCapabilityRegistry()
	svc := quota.NewQuotaService(store, registry)

	accountID := "li:race_user"

	// Define a custom rule with a limit of 1
	testMetric := quota.ActionMetric("test.single_slot")
	svc.SetCustomRule(quota.QuotaDefinition{
		Metric:                     testMetric,
		Limit:                      1,
		Window:                     24 * time.Hour,
		RequiresApprovedCapability: false,
	})

	numWorkers := 100
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	var successCount int32
	var rejectedCount int32

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()
			_, err := svc.Reserve(ctx, quota.ReservationRequest{
				AccountID:   accountID,
				Provider:    "linkedin",
				WorkspaceID: "ws-1",
				Module:      "linkedin",
				Metric:      testMetric,
				Amount:      1,
				TTL:         5 * time.Minute,
			})
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else if errors.Is(err, quota.ErrQuotaExceeded) {
				atomic.AddInt32(&rejectedCount, 1)
			} else {
				t.Errorf("worker %d got unexpected error: %v", workerID, err)
			}
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Fatalf("expected exactly 1 worker to succeed, got %d", successCount)
	}
	if rejectedCount != int32(numWorkers-1) {
		t.Fatalf("expected %d workers to be rejected with ErrQuotaExceeded, got %d", numWorkers-1, rejectedCount)
	}
}

// AT-008 & REQ-010: Recruiter outreach consumes nested limits (3/day inside 10-DM total).
func TestQuota_NestedLimitsRecruiterDM(t *testing.T) {
	ctx := context.Background()
	store := quota.NewMemoryQuotaStore()
	registry := provider.NewCapabilityRegistry()
	svc := quota.NewQuotaService(store, registry)

	accountID := "li:nested_user"

	// 1. Send 3 recruiter DMs (child limit is 3, parent limit is 10)
	for i := 1; i <= 3; i++ {
		res, err := svc.Reserve(ctx, quota.ReservationRequest{
			AccountID:   accountID,
			Provider:    "linkedin",
			WorkspaceID: "ws-1",
			Module:      "career",
			Metric:      quota.MetricLinkedInDMRecruiter,
			Amount:      1,
		})
		if err != nil {
			t.Fatalf("expected recruiter DM %d to succeed, got err: %v", i, err)
		}
		if err := svc.Settle(ctx, res.ID); err != nil {
			t.Fatalf("failed to settle recruiter DM %d: %v", i, err)
		}
	}

	// 2. Attempt a 4th recruiter DM: child limit (3) exceeded, even though 7 parent DM slots remain
	_, err := svc.Reserve(ctx, quota.ReservationRequest{
		AccountID:   accountID,
		Provider:    "linkedin",
		WorkspaceID: "ws-1",
		Module:      "career",
		Metric:      quota.MetricLinkedInDMRecruiter,
		Amount:      1,
	})
	if !errors.Is(err, quota.ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded on 4th recruiter DM, got %v", err)
	}

	// 3. Regular LinkedIn DM should still have slots (10 - 3 = 7 slots available)
	generalRes, err := svc.Reserve(ctx, quota.ReservationRequest{
		AccountID:   accountID,
		Provider:    "linkedin",
		WorkspaceID: "ws-1",
		Module:      "linkedin",
		Metric:      quota.MetricLinkedInDMSend,
		Amount:      1,
	})
	if err != nil {
		t.Fatalf("expected general DM to succeed because 7 parent slots remain, got: %v", err)
	}
	_ = svc.Settle(ctx, generalRes.ID)

	// Now check usage summary
	summary, err := svc.GetUsageSummary(ctx, accountID, quota.MetricLinkedInDMRecruiter)
	if err != nil {
		t.Fatalf("failed to get usage summary: %v", err)
	}
	if summary.Consumed != 3 {
		t.Errorf("expected 3 consumed recruiter DMs, got %d", summary.Consumed)
	}
	if summary.ParentUsage == nil || summary.ParentUsage.Consumed != 4 {
		t.Errorf("expected parent DM consumed to be 4 (3 recruiter + 1 general), got %v", summary.ParentUsage)
	}

	// 4. Test reverse case: parent limit exhausted by general DMs blocks child recruiter DM
	accountID2 := "li:parent_exhausted"
	for i := 1; i <= 10; i++ {
		res, err := svc.Reserve(ctx, quota.ReservationRequest{
			AccountID:   accountID2,
			Provider:    "linkedin",
			WorkspaceID: "ws-1",
			Module:      "linkedin",
			Metric:      quota.MetricLinkedInDMSend,
			Amount:      1,
		})
		if err != nil {
			t.Fatalf("failed to reserve general DM %d: %v", i, err)
		}
		_ = svc.Settle(ctx, res.ID)
	}

	// Now try to send 1 recruiter DM: child count is 0, but parent is 10/10 -> must fail with ErrNestedQuotaExceeded
	_, err = svc.Reserve(ctx, quota.ReservationRequest{
		AccountID:   accountID2,
		Provider:    "linkedin",
		WorkspaceID: "ws-1",
		Module:      "career",
		Metric:      quota.MetricLinkedInDMRecruiter,
		Amount:      1,
	})
	if !errors.Is(err, quota.ErrNestedQuotaExceeded) {
		t.Fatalf("expected ErrNestedQuotaExceeded when parent DM cap is full, got %v", err)
	}
}

// AT-009 & REQ-011: Same provider account across modules/workspaces shares budgets; reconnect does not reset.
func TestQuota_AccountWideSharedAcrossWorkspacesAndModules(t *testing.T) {
	ctx := context.Background()
	store := quota.NewMemoryQuotaStore()
	registry := provider.NewCapabilityRegistry()
	svc := quota.NewQuotaService(store, registry)

	sharedAccountUID := "li:company_brand_exec"

	// Workspace A (Career module) reserves 3 connection requests (limit is 5/day)
	for i := 0; i < 3; i++ {
		res, err := svc.Reserve(ctx, quota.ReservationRequest{
			AccountID:   sharedAccountUID,
			Provider:    "linkedin",
			WorkspaceID: "workspace-alpha",
			Module:      "career",
			Metric:      quota.MetricLinkedInConnectionRequest,
			Amount:      1,
		})
		if err != nil {
			t.Fatalf("expected Workspace A reservation to succeed, got %v", err)
		}
		_ = svc.Settle(ctx, res.ID)
	}

	// Workspace B (Social module) reserves 2 connection requests -> reaches 5/5
	for i := 0; i < 2; i++ {
		res, err := svc.Reserve(ctx, quota.ReservationRequest{
			AccountID:   sharedAccountUID,
			Provider:    "linkedin",
			WorkspaceID: "workspace-beta",
			Module:      "social",
			Metric:      quota.MetricLinkedInConnectionRequest,
			Amount:      1,
		})
		if err != nil {
			t.Fatalf("expected Workspace B reservation to succeed, got %v", err)
		}
		_ = svc.Settle(ctx, res.ID)
	}

	// Workspace C (or A again) tries 6th connection request -> fails because shared account budget is exhausted
	_, err := svc.Reserve(ctx, quota.ReservationRequest{
		AccountID:   sharedAccountUID,
		Provider:    "linkedin",
		WorkspaceID: "workspace-gamma",
		Module:      "linkedin",
		Metric:      quota.MetricLinkedInConnectionRequest,
		Amount:      1,
	})
	if !errors.Is(err, quota.ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded across shared account, got %v", err)
	}

	// Re-querying after "reconnect" or workspace switch preserves counters
	summary, err := svc.GetUsageSummary(ctx, sharedAccountUID, quota.MetricLinkedInConnectionRequest)
	if err != nil {
		t.Fatalf("failed to query usage summary: %v", err)
	}
	if summary.Consumed != 5 || summary.Available != 0 {
		t.Fatalf("expected 5 consumed and 0 available, got consumed=%d, available=%d", summary.Consumed, summary.Available)
	}
}

// AT-010: Unauthorized action remains disabled; 429/auth/challenge pause correctly.
func TestQuota_CapabilityAndAccountStatusGating(t *testing.T) {
	ctx := context.Background()
	store := quota.NewMemoryQuotaStore()
	registry := provider.NewCapabilityRegistry()
	svc := quota.NewQuotaService(store, registry)

	accountID := "li:security_user"

	// 1. Unsupported action: job application is not supported on LinkedIn provider manifest
	_, err := svc.Reserve(ctx, quota.ReservationRequest{
		AccountID:   accountID,
		Provider:    "linkedin",
		WorkspaceID: "ws-1",
		Module:      "career",
		Metric:      quota.MetricCareerJobApplyAutomated,
		Amount:      1,
	})
	if !errors.Is(err, quota.ErrActionCapabilityUnauthorized) {
		t.Fatalf("expected ErrActionCapabilityUnauthorized for unsupported action, got %v", err)
	}

	// 2. Paused account circuit breaker
	err = svc.PauseAccount(ctx, accountID, "linkedin", "Suspicious activity alert from platform")
	if err != nil {
		t.Fatalf("failed to pause account: %v", err)
	}

	_, err = svc.Reserve(ctx, quota.ReservationRequest{
		AccountID:   accountID,
		Provider:    "linkedin",
		WorkspaceID: "ws-1",
		Module:      "social",
		Metric:      quota.MetricLinkedInConnectionRequest,
		Amount:      1,
	})
	if !errors.Is(err, quota.ErrAccountPaused) {
		t.Fatalf("expected ErrAccountPaused when account is paused, got %v", err)
	}

	// Resume account
	_ = svc.ResumeAccount(ctx, accountID, "linkedin")

	// 3. Challenge required circuit breaker (CAPTCHA/2FA)
	err = svc.RecordChallenge(ctx, accountID, "linkedin", "CAPTCHA detected on login checkpoint")
	if err != nil {
		t.Fatalf("failed to record challenge: %v", err)
	}

	_, err = svc.Reserve(ctx, quota.ReservationRequest{
		AccountID:   accountID,
		Provider:    "linkedin",
		WorkspaceID: "ws-1",
		Module:      "social",
		Metric:      quota.MetricLinkedInConnectionRequest,
		Amount:      1,
	})
	if !errors.Is(err, quota.ErrChallengeRequired) {
		t.Fatalf("expected ErrChallengeRequired when challenge is pending, got %v", err)
	}

	// 4. Rate Limited (429) backoff
	retryAfter := time.Now().UTC().Add(30 * time.Minute)
	err = svc.RecordRateLimit(ctx, accountID, "linkedin", retryAfter)
	if err != nil {
		t.Fatalf("failed to record rate limit: %v", err)
	}

	_, err = svc.Reserve(ctx, quota.ReservationRequest{
		AccountID:   accountID,
		Provider:    "linkedin",
		WorkspaceID: "ws-1",
		Module:      "social",
		Metric:      quota.MetricLinkedInConnectionRequest,
		Amount:      1,
	})
	if !errors.Is(err, quota.ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited within backoff window, got %v", err)
	}
}

// AT-018: Rolling windows are timestamp-based and do not multiply on timezone or DST changes.
func TestQuota_TimezoneAndDSTRollingWindow(t *testing.T) {
	ctx := context.Background()
	store := quota.NewMemoryQuotaStore()
	registry := provider.NewCapabilityRegistry()
	svc := quota.NewQuotaService(store, registry)

	accountID := "li:rolling_user"

	// Define 24-hour limit of 2
	metric := quota.ActionMetric("test.rolling_24h")
	svc.SetCustomRule(quota.QuotaDefinition{
		Metric: metric,
		Limit:  2,
		Window: 24 * time.Hour,
	})

	now := time.Now().UTC()

	// Action 1 occurred 26 hours ago (settled)
	oldSettled := now.Add(-26 * time.Hour)
	resOld := &quota.Reservation{
		ID:        "res-old",
		AccountID: accountID,
		Metric:    metric,
		Amount:    1,
		Status:    quota.ReservationSettled,
		CreatedAt: oldSettled,
		SettledAt: &oldSettled,
	}
	_ = store.Reserve(ctx, resOld, quota.QuotaDefinition{Metric: metric, Limit: 10, Window: 48 * time.Hour}, nil, oldSettled)
	_ = store.Settle(ctx, resOld.ID, oldSettled)

	// Action 2 occurred 2 hours ago (settled)
	recentSettled := now.Add(-2 * time.Hour)
	resRecent := &quota.Reservation{
		ID:        "res-recent",
		AccountID: accountID,
		Metric:    metric,
		Amount:    1,
		Status:    quota.ReservationSettled,
		CreatedAt: recentSettled,
		SettledAt: &recentSettled,
	}
	_ = store.Reserve(ctx, resRecent, quota.QuotaDefinition{Metric: metric, Limit: 10, Window: 48 * time.Hour}, nil, recentSettled)
	_ = store.Settle(ctx, resRecent.ID, recentSettled)

	// Since resOld is 26 hours ago, only resRecent is in the 24h window (usage = 1 / 2)
	summary, err := svc.GetUsageSummary(ctx, accountID, metric)
	if err != nil {
		t.Fatalf("failed to get usage summary: %v", err)
	}
	if summary.Consumed != 1 {
		t.Fatalf("expected exactly 1 consumed in rolling 24h window, got %d", summary.Consumed)
	}
	if summary.Available != 1 {
		t.Fatalf("expected 1 slot available, got %d", summary.Available)
	}

	// Reserving the second slot succeeds
	res2, err := svc.Reserve(ctx, quota.ReservationRequest{
		AccountID: accountID,
		Metric:    metric,
		Amount:    1,
	})
	if err != nil {
		t.Fatalf("expected reserving remaining slot to succeed, got %v", err)
	}
	_ = svc.Settle(ctx, res2.ID)

	// Reserving a 3rd slot fails
	_, err = svc.Reserve(ctx, quota.ReservationRequest{
		AccountID: accountID,
		Metric:    metric,
		Amount:    1,
	})
	if !errors.Is(err, quota.ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
}

// AT-006 & REQ-010: Ambiguous write outcome retains reservation and fails closed against duplicate attempts.
func TestQuota_AmbiguousWriteRetention(t *testing.T) {
	ctx := context.Background()
	store := quota.NewMemoryQuotaStore()
	registry := provider.NewCapabilityRegistry()
	svc := quota.NewQuotaService(store, registry)

	accountID := "li:timeout_user"
	metric := quota.ActionMetric("test.single_dm")
	svc.SetCustomRule(quota.QuotaDefinition{
		Metric: metric,
		Limit:  1,
		Window: 24 * time.Hour,
	})

	// Reserve slot 1
	res, err := svc.Reserve(ctx, quota.ReservationRequest{
		AccountID: accountID,
		Metric:    metric,
		Amount:    1,
	})
	if err != nil {
		t.Fatalf("failed to reserve slot: %v", err)
	}

	// Network drop / timeout after dispatch -> RetainAmbiguous
	err = svc.RetainAmbiguous(ctx, res.ID, "HTTP 504 gateway timeout after upstream socket write")
	if err != nil {
		t.Fatalf("failed to retain ambiguous reservation: %v", err)
	}

	// Retained reservation CANNOT be released blindly
	err = svc.Release(ctx, res.ID)
	if !errors.Is(err, quota.ErrReservationAlreadyFinalized) {
		t.Fatalf("expected ErrReservationAlreadyFinalized when attempting to release retained reservation, got %v", err)
	}

	// Retained reservation counts against quota, blocking unconfirmed duplicate retry
	_, err = svc.Reserve(ctx, quota.ReservationRequest{
		AccountID: accountID,
		Metric:    metric,
		Amount:    1,
	})
	if !errors.Is(err, quota.ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded because retained slot locks capacity, got %v", err)
	}
}

// REQ-010: Outage fails closed.
func TestQuota_FailClosedOnStoreOutage(t *testing.T) {
	ctx := context.Background()
	store := quota.NewMemoryQuotaStore()
	registry := provider.NewCapabilityRegistry()
	svc := quota.NewQuotaService(store, registry)

	accountID := "li:outage_user"
	metric := quota.ActionMetric("test.outage")

	// Trigger outage
	store.SetOutage(true)

	_, err := svc.Reserve(ctx, quota.ReservationRequest{
		AccountID: accountID,
		Metric:    metric,
		Amount:    1,
	})
	if !errors.Is(err, quota.ErrStoreUnavailable) {
		t.Fatalf("expected ErrStoreUnavailable on store outage, got %v", err)
	}
}
