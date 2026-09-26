package linkedin_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/social-platform/services/core/internal/linkedin"
)

type CompanyFollowFixture struct {
	Scenarios struct {
		Standard struct {
			WorkspaceID   string `json:"workspace_id"`
			TenantID      string `json:"tenant_id"`
			WatchlistItems []struct {
				ItemID          string   `json:"item_id"`
				CompanyRecordID string   `json:"company_record_id"`
				CompanyName     string   `json:"company_name"`
				UniversalName   string   `json:"universal_name"`
				Domain          string   `json:"domain"`
				Priority        string   `json:"priority"`
				TargetReason    string   `json:"target_reason"`
				Tags            []string `json:"tags"`
				Status          string   `json:"status"`
			} `json:"watchlist_items"`
			BatchPlan struct {
				PlanName          string   `json:"plan_name"`
				PacingIntervalSec int      `json:"pacing_interval_sec"`
				SelectedItemIDs   []string `json:"selected_item_ids"`
			} `json:"batch_plan"`
			ExpectedGeneratedItems int      `json:"expected_generated_items"`
			ExpectedCompanyURLs    []string `json:"expected_company_page_urls"`
		} `json:"standard_company_watchlist_and_batch_follow"`
		Deduplication struct {
			WorkspaceID  string `json:"workspace_id"`
			TenantID     string `json:"tenant_id"`
			ExistingItem struct {
				ItemID        string `json:"item_id"`
				CompanyName   string `json:"company_name"`
				UniversalName string `json:"universal_name"`
			} `json:"existing_item"`
			DuplicateAttempt struct {
				ItemID        string `json:"item_id"`
				CompanyName   string `json:"company_name"`
				UniversalName string `json:"universal_name"`
			} `json:"duplicate_attempt"`
			ExpectedError string `json:"expected_error"`
		} `json:"company_deduplication_guardrail"`
		BudgetExhaustion struct {
			WorkspaceID   string `json:"workspace_id"`
			TenantID      string `json:"tenant_id"`
			InitialBudget struct {
				DailyLimit  int `json:"daily_limit"`
				WeeklyLimit int `json:"weekly_limit"`
				DailyUsed   int `json:"daily_used"`
				WeeklyUsed  int `json:"weekly_used"`
			} `json:"initial_budget"`
			ExpectedLocked bool   `json:"expected_locked"`
			ExpectedError  string `json:"expected_error"`
		} `json:"budget_exhaustion_and_throttling"`
		UnsupportedAutomation struct {
			WorkspaceID   string `json:"workspace_id"`
			TenantID      string `json:"tenant_id"`
			AttemptedMode string `json:"attempted_mode"`
			ExpectedError string `json:"expected_error"`
		} `json:"unsupported_automation_fail_closed"`
	} `json:"scenarios"`
}

func loadCompanyFollowFixture(t *testing.T) *CompanyFollowFixture {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_company_follow.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture file: %v", err)
	}

	var f CompanyFollowFixture
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("failed to unmarshal fixture json: %v", err)
	}
	return &f
}

func TestLinkedIn_CompanyFollow_FixtureGroundtruth(t *testing.T) {
	ctx := context.Background()
	fixture := loadCompanyFollowFixture(t)
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)

	// Scenario 1: standard_company_watchlist_and_batch_follow
	sc1 := fixture.Scenarios.Standard
	for _, item := range sc1.WatchlistItems {
		wlItem := &linkedin.CompanyWatchlistItem{
			ItemID:          item.ItemID,
			WorkspaceID:     sc1.WorkspaceID,
			TenantID:        sc1.TenantID,
			CompanyRecordID: item.CompanyRecordID,
			CompanyName:     item.CompanyName,
			UniversalName:   item.UniversalName,
			Domain:          item.Domain,
			Priority:        linkedin.CompanyFollowPriority(item.Priority),
			TargetReason:    item.TargetReason,
			Tags:            item.Tags,
			Status:          linkedin.WatchlistStatus(item.Status),
		}
		saved, err := svc.AddCompanyToWatchlist(ctx, wlItem)
		if err != nil {
			t.Fatalf("unexpected error adding company to watchlist: %v", err)
		}
		if saved.CompanyPageURL == "" {
			t.Errorf("expected generated CompanyPageURL, got empty")
		}
	}

	// Create Batch Follow Plan
	plan, err := svc.CreateBatchFollowPlan(ctx, sc1.WorkspaceID, sc1.TenantID, sc1.BatchPlan.PlanName, sc1.BatchPlan.SelectedItemIDs, sc1.BatchPlan.PacingIntervalSec)
	if err != nil {
		t.Fatalf("unexpected error creating batch follow plan: %v", err)
	}
	if plan.TotalCount != sc1.ExpectedGeneratedItems {
		t.Errorf("expected %d items, got %d", sc1.ExpectedGeneratedItems, plan.TotalCount)
	}
	for i, expectedURL := range sc1.ExpectedCompanyURLs {
		if plan.Items[i].CompanyPageURL != expectedURL {
			t.Errorf("expected URL %s, got %s", expectedURL, plan.Items[i].CompanyPageURL)
		}
	}

	// Confirm first item manual follow
	updatedPlan, err := svc.ConfirmCompanyFollowed(ctx, plan.PlanID, plan.Items[0].ItemID, sc1.TenantID)
	if err != nil {
		t.Fatalf("unexpected error confirming manual follow: %v", err)
	}
	if updatedPlan.FollowedCount != 1 {
		t.Errorf("expected FollowedCount 1, got %d", updatedPlan.FollowedCount)
	}
	if updatedPlan.Items[0].Status != linkedin.PlanItemFollowed {
		t.Errorf("expected item status %s, got %s", linkedin.PlanItemFollowed, updatedPlan.Items[0].Status)
	}

	// Scenario 2: company_deduplication_guardrail (AT-004, AT-007)
	sc2 := fixture.Scenarios.Deduplication
	dupItem := &linkedin.CompanyWatchlistItem{
		ItemID:        sc2.DuplicateAttempt.ItemID,
		WorkspaceID:   sc2.WorkspaceID,
		TenantID:      sc2.TenantID,
		CompanyName:   sc2.DuplicateAttempt.CompanyName,
		UniversalName: sc2.DuplicateAttempt.UniversalName,
		Priority:      linkedin.PriorityHigh,
	}
	_, dupErr := svc.AddCompanyToWatchlist(ctx, dupItem)
	if dupErr == nil {
		t.Fatalf("expected duplicate error, got nil")
	}
	if dupErr != linkedin.ErrWatchlistCompanyDuplicate {
		t.Errorf("expected ErrWatchlistCompanyDuplicate, got %v", dupErr)
	}

	// Scenario 3: budget_exhaustion_and_throttling (FND-011, REQ-010)
	sc3 := fixture.Scenarios.BudgetExhaustion
	budget := &linkedin.CompanyFollowBudget{
		WorkspaceID:     sc3.WorkspaceID,
		TenantID:        sc3.TenantID,
		DailyLimit:      sc3.InitialBudget.DailyLimit,
		WeeklyLimit:     sc3.InitialBudget.WeeklyLimit,
		DailyUsed:       sc3.InitialBudget.DailyUsed,
		WeeklyUsed:      sc3.InitialBudget.WeeklyUsed,
		LastResetDate:   time.Now().UTC().Format("2006-01-02"),
		WeeklyResetDate: time.Now().UTC().Format("2006-01-02"),
		IsLocked:        false,
	}
	if err := svc.UpdateCompanyFollowBudget(ctx, budget); err != nil {
		t.Fatalf("failed to update budget: %v", err)
	}
	// Attempt follow consumption on exhausted budget
	consumeErr := linkedin.ConsumeFollowBudget(budget, time.Now().UTC())
	if consumeErr == nil {
		t.Fatalf("expected budget exhaustion error, got nil")
	}
	if consumeErr != linkedin.ErrCompanyFollowBudgetExhausted {
		t.Errorf("expected ErrCompanyFollowBudgetExhausted, got %v", consumeErr)
	}
	if !budget.IsLocked {
		t.Errorf("expected budget to be locked upon exhaustion")
	}

	// Scenario 4: unsupported_automation_fail_closed (AT-010, REQ-009)
	sc4 := fixture.Scenarios.UnsupportedAutomation
	autoErr := linkedin.RejectUnsupportedLiveFollow(sc4.AttemptedMode)
	if autoErr == nil {
		t.Fatalf("expected unsupported automation error, got nil")
	}
	if autoErr != linkedin.ErrUnsupportedLiveFollow {
		t.Errorf("expected ErrUnsupportedLiveFollow, got %v", autoErr)
	}
}

func TestLinkedIn_CompanyFollow_TenantIsolation(t *testing.T) {
	ctx := context.Background()
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)

	// Workspace Alpha
	_, err := svc.AddCompanyToWatchlist(ctx, &linkedin.CompanyWatchlistItem{
		ItemID:        "cwl-alpha-1",
		WorkspaceID:   "ws-alpha",
		TenantID:      "tenant-alpha",
		CompanyName:   "Alpha Corp",
		UniversalName: "alphacorp",
		Priority:      linkedin.PriorityHigh,
	})
	if err != nil {
		t.Fatalf("failed to add alpha item: %v", err)
	}

	// Workspace Beta
	_, err = svc.AddCompanyToWatchlist(ctx, &linkedin.CompanyWatchlistItem{
		ItemID:        "cwl-beta-1",
		WorkspaceID:   "ws-beta",
		TenantID:      "tenant-beta",
		CompanyName:   "Beta Corp",
		UniversalName: "betacorp",
		Priority:      linkedin.PriorityLow,
	})
	if err != nil {
		t.Fatalf("failed to add beta item: %v", err)
	}

	// Listing with Alpha credentials
	alphaList, err := svc.GetCompanyWatchlist(ctx, "ws-alpha", "tenant-alpha", "")
	if err != nil {
		t.Fatalf("failed to list alpha watchlist: %v", err)
	}
	if len(alphaList) != 1 || alphaList[0].ItemID != "cwl-alpha-1" {
		t.Errorf("cross-tenant leakage or missing alpha item: %v", alphaList)
	}

	// Cross-tenant plan creation denial
	_, crossErr := svc.CreateBatchFollowPlan(ctx, "ws-alpha", "tenant-alpha", "Hacked Plan", []string{"cwl-beta-1"}, 120)
	if crossErr == nil {
		t.Fatalf("expected cross-tenant access denial, got nil")
	}
}

func TestLinkedIn_CompanyFollow_PacingAndSkip(t *testing.T) {
	ctx := context.Background()
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)

	item, err := svc.AddCompanyToWatchlist(ctx, &linkedin.CompanyWatchlistItem{
		ItemID:        "cwl-test-1",
		WorkspaceID:   "ws-test",
		TenantID:      "tenant-test",
		CompanyName:   "Test Labs",
		UniversalName: "test-labs",
		Priority:      linkedin.PriorityMedium,
	})
	if err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	plan, err := svc.CreateBatchFollowPlan(ctx, "ws-test", "tenant-test", "Test Batch", []string{item.ItemID}, 150)
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	if plan.Items[0].PacingIntervalSec != 150 {
		t.Errorf("expected pacing 150s, got %d", plan.Items[0].PacingIntervalSec)
	}

	// Skip item
	updatedPlan, err := svc.SkipCompanyFollow(ctx, plan.PlanID, plan.Items[0].ItemID, "tenant-test", "Already followed on personal account")
	if err != nil {
		t.Fatalf("failed to skip item: %v", err)
	}
	if updatedPlan.Items[0].Status != linkedin.PlanItemSkipped {
		t.Errorf("expected status %s, got %s", linkedin.PlanItemSkipped, updatedPlan.Items[0].Status)
	}
	if updatedPlan.Status != linkedin.BatchPlanCompleted {
		t.Errorf("expected batch status completed after skipping all items, got %s", updatedPlan.Status)
	}
}
