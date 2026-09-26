package linkedin

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrWatchlistCompanyDuplicate    = errors.New("company_follow: company already present in target watchlist (AT-004, AT-007)")
	ErrCompanyFollowBudgetExhausted = errors.New("company_follow: daily company follow limit exhausted, fail-closed throttling active (FND-011, REQ-010)")
	ErrWeeklyFollowBudgetExhausted  = errors.New("company_follow: weekly company follow limit exhausted, fail-closed throttling active (FND-011, REQ-010)")
	ErrUnsupportedLiveFollow        = errors.New("company_follow: background headless follow mutation permanently rejected under platform truth-in-advertising policy (AT-010, REQ-009)")
	ErrInvalidBatchPlan             = errors.New("company_follow: invalid batch plan payload or empty target list")
	ErrWatchlistItemNotFound        = errors.New("company_follow: target company watchlist item not found")
	ErrBatchPlanNotFound            = errors.New("company_follow: batch follow plan not found")
	ErrInvalidCompanyURL            = errors.New("company_follow: missing or invalid company universal name / URL")
)

type CompanyFollowPriority string

const (
	PriorityHigh   CompanyFollowPriority = "high"
	PriorityMedium CompanyFollowPriority = "medium"
	PriorityLow    CompanyFollowPriority = "low"
)

type WatchlistStatus string

const (
	WatchlistActive   WatchlistStatus = "active"
	WatchlistPaused   WatchlistStatus = "paused"
	WatchlistArchived WatchlistStatus = "archived"
)

type PlanItemStatus string

const (
	PlanItemPlanned   PlanItemStatus = "planned"
	PlanItemReady     PlanItemStatus = "ready_for_manual_follow"
	PlanItemFollowed  PlanItemStatus = "manual_followed"
	PlanItemSkipped   PlanItemStatus = "skipped"
)

type BatchPlanStatus string

const (
	BatchPlanDraft      BatchPlanStatus = "draft"
	BatchPlanInProgress BatchPlanStatus = "in_progress"
	BatchPlanCompleted  BatchPlanStatus = "completed"
)

// CompanyWatchlistItem represents an audited, target organization tracked by a candidate.
type CompanyWatchlistItem struct {
	ItemID          string                `json:"item_id"`
	WorkspaceID     string                `json:"workspace_id"`
	TenantID        string                `json:"tenant_id"`
	CompanyRecordID string                `json:"company_record_id,omitempty"`
	CompanyName     string                `json:"company_name"`
	UniversalName   string                `json:"universal_name"`
	Domain          string                `json:"domain,omitempty"`
	CompanyPageURL  string                `json:"company_page_url"`
	Priority        CompanyFollowPriority `json:"priority"`
	TargetReason    string                `json:"target_reason"`
	Tags            []string              `json:"tags"`
	Status          WatchlistStatus       `json:"status"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// CompanyFollowBudget enforces strict conservative account-level follow caps (FND-011, REQ-010).
type CompanyFollowBudget struct {
	WorkspaceID     string    `json:"workspace_id"`
	TenantID        string    `json:"tenant_id"`
	DailyLimit      int       `json:"daily_limit"`  // Default 15
	WeeklyLimit     int       `json:"weekly_limit"` // Default 50
	DailyUsed       int       `json:"daily_used"`
	WeeklyUsed      int       `json:"weekly_used"`
	LastResetDate   string    `json:"last_reset_date"`   // YYYY-MM-DD
	WeeklyResetDate string    `json:"weekly_reset_date"` // YYYY-MM-DD
	IsLocked        bool      `json:"is_locked"`
	LockReason      string    `json:"lock_reason,omitempty"`
}

// CompanyFollowPlanItem represents a single planned follow step within a batch.
type CompanyFollowPlanItem struct {
	ItemID            string                `json:"item_id"`
	CompanyRecordID   string                `json:"company_record_id,omitempty"`
	CompanyName       string                `json:"company_name"`
	UniversalName     string                `json:"universal_name"`
	CompanyPageURL    string                `json:"company_page_url"`
	Priority          CompanyFollowPriority `json:"priority"`
	TargetReason      string                `json:"target_reason"`
	PacingIntervalSec int                   `json:"pacing_interval_sec"`
	Status            PlanItemStatus        `json:"status"`
	ScheduledFor      time.Time             `json:"scheduled_for"`
	FollowedAt        *time.Time            `json:"followed_at,omitempty"`
	Notes             string                `json:"notes,omitempty"`
}

// BatchCompanyFollowPlan represents a structured sequence of paced follow actions (SRC-L1).
type BatchCompanyFollowPlan struct {
	PlanID        string                  `json:"plan_id"`
	WorkspaceID   string                  `json:"workspace_id"`
	TenantID      string                  `json:"tenant_id"`
	PlanName      string                  `json:"plan_name"`
	Items         []CompanyFollowPlanItem `json:"items"`
	TotalCount    int                     `json:"total_count"`
	FollowedCount int                     `json:"followed_count"`
	Status        BatchPlanStatus         `json:"status"`
	CreatedAt     time.Time               `json:"created_at"`
}

// BuildCompanyPageURL constructs a canonical LinkedIn company profile URL from a universal name.
func BuildCompanyPageURL(universalName string) string {
	cleanSlug := strings.TrimSpace(strings.ToLower(universalName))
	cleanSlug = strings.TrimPrefix(cleanSlug, "https://www.linkedin.com/company/")
	cleanSlug = strings.TrimPrefix(cleanSlug, "http://www.linkedin.com/company/")
	cleanSlug = strings.Trim(cleanSlug, "/")
	if cleanSlug == "" {
		return ""
	}
	return fmt.Sprintf("https://www.linkedin.com/company/%s", cleanSlug)
}

// ValidateWatchlistItem validates data completeness and tenant boundaries.
func ValidateWatchlistItem(item *CompanyWatchlistItem) error {
	if item == nil {
		return errors.New("company_follow: watchlist item cannot be nil")
	}
	if strings.TrimSpace(item.WorkspaceID) == "" || strings.TrimSpace(item.TenantID) == "" {
		return ErrInvalidTenantOrWorkspace
	}
	if strings.TrimSpace(item.CompanyName) == "" {
		return errors.New("company_follow: company name is required")
	}
	cleanUniversal := strings.TrimSpace(item.UniversalName)
	if cleanUniversal == "" {
		return ErrInvalidCompanyURL
	}
	if item.CompanyPageURL == "" {
		item.CompanyPageURL = BuildCompanyPageURL(cleanUniversal)
	}
	if item.Priority == "" {
		item.Priority = PriorityMedium
	}
	if item.Status == "" {
		item.Status = WatchlistActive
	}
	return nil
}

// DefaultCompanyFollowBudget provides conservative defaults satisfying REQ-010 and FND-011.
func DefaultCompanyFollowBudget(workspaceID, tenantID string, now time.Time) *CompanyFollowBudget {
	return &CompanyFollowBudget{
		WorkspaceID:     workspaceID,
		TenantID:        tenantID,
		DailyLimit:      15,
		WeeklyLimit:     50,
		DailyUsed:       0,
		WeeklyUsed:      0,
		LastResetDate:   now.Format("2006-01-02"),
		WeeklyResetDate: now.Format("2006-01-02"),
		IsLocked:        false,
	}
}

// ResetExpiredBudgets updates daily/weekly reset dates if rolled over.
func (b *CompanyFollowBudget) ResetExpiredBudgets(now time.Time) {
	currentDay := now.Format("2006-01-02")
	if b.LastResetDate != currentDay {
		b.DailyUsed = 0
		b.LastResetDate = currentDay
		if b.IsLocked && strings.Contains(b.LockReason, "daily") {
			b.IsLocked = false
			b.LockReason = ""
		}
	}

	lastReset, err := time.Parse("2006-01-02", b.WeeklyResetDate)
	if err == nil && now.Sub(lastReset) >= 7*24*time.Hour {
		b.WeeklyUsed = 0
		b.WeeklyResetDate = currentDay
		if b.IsLocked && strings.Contains(b.LockReason, "weekly") {
			b.IsLocked = false
			b.LockReason = ""
		}
	}
}

// CanConsumeFollowBudget checks if the budget allows 1 follow action.
func CanConsumeFollowBudget(b *CompanyFollowBudget, now time.Time) (bool, error) {
	if b == nil {
		return false, errors.New("company_follow: budget cannot be nil")
	}
	b.ResetExpiredBudgets(now)
	if b.IsLocked {
		return false, fmt.Errorf("company_follow: budget locked: %s", b.LockReason)
	}
	if b.DailyUsed >= b.DailyLimit {
		b.IsLocked = true
		b.LockReason = "daily company follow limit exhausted, fail-closed throttling active (FND-011, REQ-010)"
		return false, ErrCompanyFollowBudgetExhausted
	}
	if b.WeeklyUsed >= b.WeeklyLimit {
		b.IsLocked = true
		b.LockReason = "weekly company follow limit exhausted, fail-closed throttling active (FND-011, REQ-010)"
		return false, ErrWeeklyFollowBudgetExhausted
	}
	return true, nil
}

// ConsumeFollowBudget increments follow consumption and locks if thresholds reached.
func ConsumeFollowBudget(b *CompanyFollowBudget, now time.Time) error {
	can, err := CanConsumeFollowBudget(b, now)
	if !can {
		return err
	}
	b.DailyUsed++
	b.WeeklyUsed++
	if b.DailyUsed >= b.DailyLimit {
		b.IsLocked = true
		b.LockReason = "daily company follow limit exhausted, fail-closed throttling active (FND-011, REQ-010)"
	} else if b.WeeklyUsed >= b.WeeklyLimit {
		b.IsLocked = true
		b.LockReason = "weekly company follow limit exhausted, fail-closed throttling active (FND-011, REQ-010)"
	}
	return nil
}

// RejectUnsupportedLiveFollow rejects headless auto-actions under AT-010 policy.
func RejectUnsupportedLiveFollow(mode string) error {
	cleanMode := strings.ToLower(strings.TrimSpace(mode))
	if cleanMode == "headless_auto_follow" || cleanMode == "background_auto" || cleanMode == "unassisted_bot" {
		return ErrUnsupportedLiveFollow
	}
	return nil
}

// GenerateBatchFollowPlan builds a pacing-aware batch sequence for targeted watchlist items.
func GenerateBatchFollowPlan(
	workspaceID, tenantID, planName string,
	watchlist []*CompanyWatchlistItem,
	pacingSec int,
	startTime time.Time,
) (*BatchCompanyFollowPlan, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(tenantID) == "" {
		return nil, ErrInvalidTenantOrWorkspace
	}
	if len(watchlist) == 0 {
		return nil, ErrInvalidBatchPlan
	}
	if pacingSec <= 0 {
		pacingSec = 120 // Default safe pacing 2 minutes
	}
	cleanPlanName := strings.TrimSpace(planName)
	if cleanPlanName == "" {
		cleanPlanName = fmt.Sprintf("Company Follow Batch - %s", startTime.Format("2006-01-02 15:04"))
	}

	planID := fmt.Sprintf("plan-%d", startTime.UnixNano())
	planItems := make([]CompanyFollowPlanItem, 0, len(watchlist))

	for i, item := range watchlist {
		scheduled := startTime.Add(time.Duration(i*pacingSec) * time.Second)
		pageURL := item.CompanyPageURL
		if pageURL == "" {
			pageURL = BuildCompanyPageURL(item.UniversalName)
		}

		planItems = append(planItems, CompanyFollowPlanItem{
			ItemID:            fmt.Sprintf("item-%s-%d", planID, i+1),
			CompanyRecordID:   item.CompanyRecordID,
			CompanyName:       item.CompanyName,
			UniversalName:     item.UniversalName,
			CompanyPageURL:    pageURL,
			Priority:          item.Priority,
			TargetReason:      item.TargetReason,
			PacingIntervalSec: pacingSec,
			Status:            PlanItemReady,
			ScheduledFor:      scheduled,
		})
	}

	return &BatchCompanyFollowPlan{
		PlanID:        planID,
		WorkspaceID:   workspaceID,
		TenantID:      tenantID,
		PlanName:      cleanPlanName,
		Items:         planItems,
		TotalCount:    len(planItems),
		FollowedCount: 0,
		Status:        BatchPlanDraft,
		CreatedAt:     startTime,
	}, nil
}
