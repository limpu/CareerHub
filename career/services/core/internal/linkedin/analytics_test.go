package linkedin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type AnalyticsFixtureScenario struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	WorkspaceID  string `json:"workspace_id"`
	TenantID     string `json:"tenant_id"`
	PostURN      string `json:"post_urn"`
	PostTitle    string `json:"post_title"`
	PublishedAt  string `json:"published_at"`
	Metrics      map[string]struct {
		Value             float64 `json:"value"`
		Status            string  `json:"status"`
		UnavailableReason string  `json:"unavailable_reason,omitempty"`
		IsModeled         bool    `json:"is_modeled"`
		ModelBasis        string  `json:"model_basis,omitempty"`
	} `json:"metrics"`
	Engagers []struct {
		AuthorURN       string `json:"author_urn"`
		Name            string `json:"name"`
		Headline        string `json:"headline"`
		Company         string `json:"company"`
		InteractionType string `json:"interaction_type"`
		CommentText     string `json:"comment_text,omitempty"`
		ExpectedSegment string `json:"expected_segment"`
	} `json:"engagers,omitempty"`
	ExpectedICPBreakdown struct {
		DecisionMakers         int `json:"decision_makers"`
		PeerPractitioners      int `json:"peer_practitioners"`
		TalentPartners         int `json:"talent_partners"`
		OtherNetwork           int `json:"other_network"`
		HighValueEngagersCount int `json:"high_value_engagers_count"`
	} `json:"expected_icp_breakdown,omitempty"`
	ForbiddenAccessTenant string `json:"forbidden_access_tenant,omitempty"`
	ExpectedRejection     string `json:"expected_rejection,omitempty"`
}

type AnalyticsFixtureSuite struct {
	Version           string                     `json:"version"`
	Task              string                     `json:"task"`
	Scenarios         []AnalyticsFixtureScenario `json:"scenarios"`
	CreatorBenchmarks []PerformanceBenchmark     `json:"creator_benchmarks"`
}

func loadAnalyticsFixture(t *testing.T) AnalyticsFixtureSuite {
	t.Helper()
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_analytics_monitoring.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture file %s: %v", fixturePath, err)
	}
	var suite AnalyticsFixtureSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("Failed to unmarshal fixture JSON: %v", err)
	}
	return suite
}

func TestAnalytics_FixtureGroundTruth(t *testing.T) {
	suite := loadAnalyticsFixture(t)
	if suite.Task != "IMP-LI-15" {
		t.Fatalf("Expected task IMP-LI-15, got %s", suite.Task)
	}
	if len(suite.Scenarios) != 4 {
		t.Fatalf("Expected 4 scenarios, got %d", len(suite.Scenarios))
	}
	if len(suite.CreatorBenchmarks) != 2 {
		t.Fatalf("Expected 2 creator benchmarks, got %d", len(suite.CreatorBenchmarks))
	}
}

func TestAnalytics_VerifiedMetricsAndICPSegmentation(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)
	suite := loadAnalyticsFixture(t)

	sc := suite.Scenarios[0]
	if sc.ID != "verified_post_metrics_with_icp_segmentation" {
		t.Fatalf("Expected scenario 1, got %s", sc.ID)
	}

	metricsMap := make(map[string]MetricItem)
	for k, v := range sc.Metrics {
		metricsMap[k] = MetricItem{
			Value:             v.Value,
			Status:            MetricAvailabilityStatus(v.Status),
			UnavailableReason: v.UnavailableReason,
			IsModeled:         v.IsModeled,
			ModelBasis:        v.ModelBasis,
		}
	}

	engagers := make([]EngagerProfile, len(sc.Engagers))
	for i, e := range sc.Engagers {
		engagers[i] = EngagerProfile{
			AuthorURN:       e.AuthorURN,
			Name:            e.Name,
			Headline:        e.Headline,
			Company:         e.Company,
			InteractionType: InteractionType(e.InteractionType),
			CommentText:     e.CommentText,
		}
	}

	publishedAt, _ := time.Parse(time.RFC3339, sc.PublishedAt)
	snapshot := &PostAnalyticsSnapshot{
		SnapshotID:  "snap-verified-001",
		WorkspaceID: sc.WorkspaceID,
		TenantID:    sc.TenantID,
		PostURN:     sc.PostURN,
		PostTitle:   sc.PostTitle,
		PublishedAt: publishedAt,
		Metrics:     metricsMap,
		Engagers:    engagers,
	}

	err := svc.RecordPostAnalytics(ctx, snapshot)
	if err != nil {
		t.Fatalf("RecordPostAnalytics failed: %v", err)
	}

	retrieved, err := svc.GetPostAnalytics(ctx, "snap-verified-001", sc.TenantID)
	if err != nil {
		t.Fatalf("GetPostAnalytics failed: %v", err)
	}

	if retrieved.PostTitle != sc.PostTitle {
		t.Errorf("Expected title %s, got %s", sc.PostTitle, retrieved.PostTitle)
	}

	// Verify ICP Breakdown matches fixture expectations (SRC-L2)
	bd := retrieved.ICPBreakdown
	if bd.DecisionMakers != sc.ExpectedICPBreakdown.DecisionMakers {
		t.Errorf("Expected %d decision makers, got %d", sc.ExpectedICPBreakdown.DecisionMakers, bd.DecisionMakers)
	}
	if bd.PeerPractitioners != sc.ExpectedICPBreakdown.PeerPractitioners {
		t.Errorf("Expected %d peer practitioners, got %d", sc.ExpectedICPBreakdown.PeerPractitioners, bd.PeerPractitioners)
	}
	if bd.TalentPartners != sc.ExpectedICPBreakdown.TalentPartners {
		t.Errorf("Expected %d talent partners, got %d", sc.ExpectedICPBreakdown.TalentPartners, bd.TalentPartners)
	}
	if bd.OtherNetwork != sc.ExpectedICPBreakdown.OtherNetwork {
		t.Errorf("Expected %d other network, got %d", sc.ExpectedICPBreakdown.OtherNetwork, bd.OtherNetwork)
	}
	if bd.HighValueEngagersCount != sc.ExpectedICPBreakdown.HighValueEngagersCount {
		t.Errorf("Expected %d high value engagers, got %d", sc.ExpectedICPBreakdown.HighValueEngagersCount, bd.HighValueEngagersCount)
	}
}

func TestAnalytics_UnavailableMetricsScopeBoundary_AT010_AT028(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)
	suite := loadAnalyticsFixture(t)

	sc := suite.Scenarios[1]
	metricsMap := make(map[string]MetricItem)
	for k, v := range sc.Metrics {
		metricsMap[k] = MetricItem{
			Value:             v.Value,
			Status:            MetricAvailabilityStatus(v.Status),
			UnavailableReason: v.UnavailableReason,
			IsModeled:         v.IsModeled,
		}
	}

	publishedAt, _ := time.Parse(time.RFC3339, sc.PublishedAt)
	snapshot := &PostAnalyticsSnapshot{
		SnapshotID:       "snap-unavailable-001",
		WorkspaceID:      sc.WorkspaceID,
		TenantID:         sc.TenantID,
		PostURN:          sc.PostURN,
		PostTitle:        sc.PostTitle,
		PublishedAt:      publishedAt,
		Metrics:          metricsMap,
		DisclosureNotice: "Impression and view analytics require official LinkedIn Marketing Developer Partner permissions.",
	}

	err := svc.RecordPostAnalytics(ctx, snapshot)
	if err != nil {
		t.Fatalf("RecordPostAnalytics failed on legitimate unavailable payload: %v", err)
	}

	// Invariant Test: An unavailable metric CANNOT have a fake non-zero value (AT-028)
	tamperedSnapshot := &PostAnalyticsSnapshot{
		SnapshotID:  "snap-faked-001",
		WorkspaceID: sc.WorkspaceID,
		TenantID:    sc.TenantID,
		PostURN:     "urn:li:activity:99999",
		Metrics: map[string]MetricItem{
			"impressions": {
				Value:             5000, // Non-zero value while marked unavailable!
				Status:            MetricStatusUnavailable,
				UnavailableReason: "missing scope",
			},
		},
	}
	err = svc.RecordPostAnalytics(ctx, tamperedSnapshot)
	if err == nil {
		t.Fatalf("Expected error when unavailable metric has non-zero value, got nil")
	}

	// Invariant Test: An unavailable metric MUST have an explanatory reason (AT-028)
	missingReasonSnapshot := &PostAnalyticsSnapshot{
		SnapshotID:  "snap-missing-reason-001",
		WorkspaceID: sc.WorkspaceID,
		TenantID:    sc.TenantID,
		PostURN:     "urn:li:activity:88888",
		Metrics: map[string]MetricItem{
			"impressions": {
				Value:  0,
				Status: MetricStatusUnavailable,
				// Empty reason
			},
		},
	}
	err = svc.RecordPostAnalytics(ctx, missingReasonSnapshot)
	if err == nil {
		t.Fatalf("Expected error when unavailable metric lacks reason, got nil")
	}
}

func TestAnalytics_ModeledEstimateDistinction_AT028(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)
	suite := loadAnalyticsFixture(t)

	sc := suite.Scenarios[2]
	metricsMap := make(map[string]MetricItem)
	for k, v := range sc.Metrics {
		metricsMap[k] = MetricItem{
			Value:      v.Value,
			Status:     MetricAvailabilityStatus(v.Status),
			IsModeled:  v.IsModeled,
			ModelBasis: v.ModelBasis,
		}
	}

	publishedAt, _ := time.Parse(time.RFC3339, sc.PublishedAt)
	snapshot := &PostAnalyticsSnapshot{
		SnapshotID:  "snap-modeled-001",
		WorkspaceID: sc.WorkspaceID,
		TenantID:    sc.TenantID,
		PostURN:     sc.PostURN,
		PostTitle:   sc.PostTitle,
		PublishedAt: publishedAt,
		Metrics:     metricsMap,
	}

	err := svc.RecordPostAnalytics(ctx, snapshot)
	if err != nil {
		t.Fatalf("RecordPostAnalytics failed for modeled estimate: %v", err)
	}

	// Invariant Test: Modeled estimate MUST provide transparent model_basis explanation (AT-028)
	tamperedModeled := &PostAnalyticsSnapshot{
		SnapshotID:  "snap-modeled-bad",
		WorkspaceID: sc.WorkspaceID,
		TenantID:    sc.TenantID,
		PostURN:     "urn:li:activity:77777",
		Metrics: map[string]MetricItem{
			"reach": {
				Value:      3000,
				Status:     MetricStatusModeledEstimate,
				IsModeled:  true,
				ModelBasis: "", // Missing explanation
			},
		},
	}
	err = svc.RecordPostAnalytics(ctx, tamperedModeled)
	if err == nil {
		t.Fatalf("Expected error when modeled estimate lacks transparent basis explanation, got nil")
	}
}

func TestAnalytics_CrossTenantIsolation_AT012(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	svc := NewService(repo)
	suite := loadAnalyticsFixture(t)

	sc := suite.Scenarios[3]
	snapshot := &PostAnalyticsSnapshot{
		SnapshotID:  "snap-isolated-001",
		WorkspaceID: sc.WorkspaceID,
		TenantID:    sc.TenantID,
		PostURN:     sc.PostURN,
		PostTitle:   "Alpha Post",
		Metrics: map[string]MetricItem{
			"reactions": {Value: 10, Status: MetricStatusAvailable},
		},
	}

	if err := svc.RecordPostAnalytics(ctx, snapshot); err != nil {
		t.Fatalf("RecordPostAnalytics failed: %v", err)
	}

	// Attempt access from forbidden tenant
	_, err := svc.GetPostAnalytics(ctx, "snap-isolated-001", sc.ForbiddenAccessTenant)
	if err == nil {
		t.Fatalf("Expected cross-tenant access denial, but succeeded")
	}
	if err != ErrCrossTenantAnalyticsDeny {
		t.Errorf("Expected ErrCrossTenantAnalyticsDeny, got %v", err)
	}

	// By URN
	_, err = svc.GetPostAnalyticsByURN(ctx, sc.PostURN, sc.ForbiddenAccessTenant)
	if err == nil {
		t.Fatalf("Expected cross-tenant access denial by URN, but succeeded")
	}
	if err != ErrCrossTenantAnalyticsDeny {
		t.Errorf("Expected ErrCrossTenantAnalyticsDeny, got %v", err)
	}

	// List scoping
	list, err := svc.ListPostAnalytics(ctx, "", sc.ForbiddenAccessTenant)
	if err != nil {
		t.Fatalf("ListPostAnalytics failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("Expected 0 results for forbidden tenant list, got %d", len(list))
	}
}
