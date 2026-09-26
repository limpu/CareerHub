package linkedin_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/social-platform/services/core/internal/linkedin"
)

type RelationshipExportFixtureSuite struct {
	ScenarioTitle string                      `json:"scenario_title"`
	Version       string                      `json:"version"`
	Scenarios     []RelationshipExportScenario `json:"scenarios"`
}

type RelationshipExportScenario struct {
	ScenarioID            string `json:"scenario_id"`
	Description           string `json:"description"`
	WorkspaceID           string `json:"workspace_id"`
	TenantID              string `json:"tenant_id"`
	RequestingOwnerID     string `json:"requesting_owner_id"`
	RequestingWorkspaceID string `json:"requesting_workspace_id"`
	RequestingTenantID    string `json:"requesting_tenant_id"`
	TargetWorkspaceID     string `json:"target_workspace_id"`
	TargetTenantID        string `json:"target_tenant_id"`
	ExpectedAllowed       bool   `json:"expected_allowed"`
	ExpectedError         string `json:"expected_error"`

	TestInputs []struct {
		Field                  string `json:"field"`
		RawValue               string `json:"raw_value"`
		ExpectedSanitizedValue string `json:"expected_sanitized_value"`
	} `json:"test_inputs"`

	SyncConfig struct {
		SpreadsheetID string `json:"spreadsheet_id"`
		SheetName     string `json:"sheet_name"`
		Timezone      string `json:"timezone"`
		IncludeNotes  bool   `json:"include_notes"`
		SyncMode      string `json:"sync_mode"`
	} `json:"sync_config"`

	SimulatedExistingSheet struct {
		Headers []string   `json:"headers"`
		Rows    [][]string `json:"rows"`
	} `json:"simulated_existing_sheet"`
}

func loadRelationshipExportFixture(t *testing.T) *RelationshipExportFixtureSuite {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_relationship_export.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read relationship export fixture: %v", err)
	}

	var suite RelationshipExportFixtureSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("failed to parse relationship export fixture: %v", err)
	}
	return &suite
}

func TestRelationshipExport_FixtureGroundTruth(t *testing.T) {
	suite := loadRelationshipExportFixture(t)
	if len(suite.Scenarios) < 4 {
		t.Fatalf("expected at least 4 scenarios in relationship export fixture, got %d", len(suite.Scenarios))
	}
}

func TestRelationshipExport_FormulaInjectionDefense_AT013(t *testing.T) {
	suite := loadRelationshipExportFixture(t)
	var formulaScenario *RelationshipExportScenario
	for i := range suite.Scenarios {
		if suite.Scenarios[i].ScenarioID == "formula_injection_defense" {
			formulaScenario = &suite.Scenarios[i]
			break
		}
	}
	if formulaScenario == nil {
		t.Fatal("formula_injection_defense scenario not found in fixture")
	}

	for _, input := range formulaScenario.TestInputs {
		sanitized := linkedin.SanitizeFormulaInjection(input.RawValue)
		if sanitized != input.ExpectedSanitizedValue {
			t.Errorf("field %s: expected sanitized %q, got %q", input.Field, input.ExpectedSanitizedValue, sanitized)
		}
		// Invariant: MUST start with single quote to prevent spreadsheet code execution
		if !strings.HasPrefix(sanitized, "'") {
			t.Errorf("field %s: sanitized formula %q must start with single quote", input.Field, sanitized)
		}
	}
}

func TestRelationshipExport_ScopedCSVExport_AT013_REQ006(t *testing.T) {
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	// Seed recruiter leads in ws-alpha and tenant-alpha
	now := time.Now().UTC()
	lead1 := &linkedin.RecruiterLead{
		ID:             "lead-alpha-001",
		WorkspaceID:    "ws-alpha",
		TenantID:       "tenant-alpha",
		RecruiterName:  "Marcus Vance",
		RecruiterTitle: "Principal Talent Acquisition Partner",
		Company:        "Stripe",
		LinkedInURL:    "https://www.linkedin.com/in/marcus-vance-talent",
		Status:         linkedin.LeadStatusInDialogue,
		OutreachStage:  linkedin.OutreachReplied,
		Notes: []linkedin.LeadNote{
			{
				ID:        "note-001",
				Author:    "user-alpha",
				Content:   "=HYPERLINK(\"http://malicious.com\", \"Chat\") Initial discussion on Staff Infra",
				CreatedAt: now,
			},
		},
		Reminder: &linkedin.FollowUpReminder{
			DueDate: now.Add(24 * time.Hour),
			Message: "Follow up with Marcus",
			Status:  linkedin.ReminderPending,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	lead2 := &linkedin.RecruiterLead{
		ID:             "lead-alpha-002",
		WorkspaceID:    "ws-alpha",
		TenantID:       "tenant-alpha",
		RecruiterName:  "@Elena Rostova",
		RecruiterTitle: "+Director of Engineering",
		Company:        "-Cloudflare",
		LinkedInURL:    "https://www.linkedin.com/in/elena-rostova-tech",
		Status:         linkedin.LeadStatusNew,
		OutreachStage:  linkedin.OutreachDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := repo.SaveRecruiterLead(ctx, lead1); err != nil {
		t.Fatalf("failed to save lead1: %v", err)
	}
	if err := repo.SaveRecruiterLead(ctx, lead2); err != nil {
		t.Fatalf("failed to save lead2: %v", err)
	}

	filter := linkedin.RelationshipExportFilter{
		WorkspaceID:        "ws-alpha",
		TenantID:           "tenant-alpha",
		RequestingTenantID: "tenant-alpha",
		RequestingOwnerID:  "usr-alex-001",
		Destination:        linkedin.DestinationCSVDownload,
		IncludeNotes:       true,
		WithBOM:            true,
		Timezone:           "UTC",
	}

	manifest, err := svc.ExportRelationshipsCSV(ctx, filter)
	if err != nil {
		t.Fatalf("ExportRelationshipsCSV failed: %v", err)
	}

	if manifest.RowCount != 2 {
		t.Errorf("expected 2 exported rows, got %d", manifest.RowCount)
	}
	if !manifest.WithBOM {
		t.Error("expected with_bom to be true")
	}

	// Verify formula sanitization in generated CSV content
	if !strings.Contains(manifest.CSVContent, "'+Director of Engineering") {
		t.Errorf("expected formula sanitized title '+Director of Engineering in CSV, got content:\n%s", manifest.CSVContent)
	}
	if !strings.Contains(manifest.CSVContent, "'-Cloudflare") {
		t.Errorf("expected formula sanitized company '-Cloudflare in CSV, got content:\n%s", manifest.CSVContent)
	}
	if !strings.Contains(manifest.CSVContent, "'@Elena Rostova") {
		t.Errorf("expected formula sanitized recruiter name '@Elena Rostova in CSV, got content:\n%s", manifest.CSVContent)
	}

	// Verify audit logging
	audits, err := svc.ListRelationshipExportAudits(ctx, "ws-alpha", "tenant-alpha")
	if err != nil {
		t.Fatalf("failed to list audits: %v", err)
	}
	if len(audits) == 0 {
		t.Fatal("expected export audit to be saved")
	}
	if audits[0].RowCount != 2 || audits[0].Destination != linkedin.DestinationCSVDownload {
		t.Errorf("unexpected audit entry: %+v", audits[0])
	}
}

func TestRelationshipExport_CrossTenantExportDenial_AT011_AT012(t *testing.T) {
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	filter := linkedin.RelationshipExportFilter{
		WorkspaceID:        "ws-alpha",
		TenantID:           "tenant-alpha",
		RequestingTenantID: "tenant-beta", // Malicious / unauthorized tenant
		Destination:        linkedin.DestinationCSVDownload,
	}

	_, err := svc.ExportRelationshipsCSV(ctx, filter)
	if err == nil {
		t.Fatal("expected cross-tenant export to be blocked, but succeeded")
	}
	if err != linkedin.ErrCrossTenantExportDenied {
		t.Errorf("expected ErrCrossTenantExportDenied, got %v", err)
	}

	// Sheets sync cross tenant check
	sheetsConfig := linkedin.RelationshipSheetsSyncConfig{
		WorkspaceID:   "ws-alpha",
		TenantID:      "tenant-alpha",
		SpreadsheetID: "sheet-001",
		SheetName:     "Leads",
	}
	_, _, err = svc.SyncRelationshipsToSheets(ctx, nil, sheetsConfig, "tenant-beta", "usr-sarah-002")
	if err == nil {
		t.Fatal("expected cross-tenant sheets sync to be blocked, but succeeded")
	}
	if err != linkedin.ErrCrossTenantExportDenied {
		t.Errorf("expected ErrCrossTenantExportDenied, got %v", err)
	}
}

func TestRelationshipExport_IdempotentSheetsReconciliation_AT014(t *testing.T) {
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	ctx := context.Background()

	now := time.Now().UTC()
	lead1 := &linkedin.RecruiterLead{
		ID:             "lead-alpha-001",
		WorkspaceID:    "ws-alpha",
		TenantID:       "tenant-alpha",
		RecruiterName:  "Marcus Vance",
		RecruiterTitle: "Principal Talent Acquisition Partner",
		Company:        "Stripe",
		LinkedInURL:    "https://www.linkedin.com/in/marcus-vance-talent",
		Status:         linkedin.LeadStatusInDialogue,
		OutreachStage:  linkedin.OutreachReplied,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	lead2 := &linkedin.RecruiterLead{
		ID:             "lead-alpha-002",
		WorkspaceID:    "ws-alpha",
		TenantID:       "tenant-alpha",
		RecruiterName:  "Elena Rostova",
		RecruiterTitle: "Director of Engineering",
		Company:        "Cloudflare",
		LinkedInURL:    "https://www.linkedin.com/in/elena-rostova-tech",
		Status:         linkedin.LeadStatusNew,
		OutreachStage:  linkedin.OutreachDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_ = repo.SaveRecruiterLead(ctx, lead1)
	_ = repo.SaveRecruiterLead(ctx, lead2)

	// Step 1: Initial Sync (Fresh sheet)
	config := linkedin.RelationshipSheetsSyncConfig{
		WorkspaceID:   "ws-alpha",
		TenantID:      "tenant-alpha",
		SpreadsheetID: "sheet-crm-001",
		SheetName:     "CRM",
		IncludeNotes:  false,
		Timezone:      "UTC",
	}

	initialRows, result1, err := svc.SyncRelationshipsToSheets(ctx, nil, config, "tenant-alpha", "usr-alex-001")
	if err != nil {
		t.Fatalf("initial sync failed: %v", err)
	}
	if result1.AppendedCount != 2 {
		t.Errorf("expected 2 appended rows, got %d", result1.AppendedCount)
	}

	// Step 2: User sorts rows and adds a custom notes column at the end
	// Headers: standard headers + "User Custom Evaluation"
	// Row 1: lead-alpha-002 (sorted first) + "Top team"
	// Row 2: lead-alpha-001 (sorted second) + "Great compensation"
	standardHeaderCount := len(linkedin.DefaultRelationshipHeaders)
	simulatedUserSheet := make([][]string, len(initialRows))
	for i, r := range initialRows {
		cp := make([]string, len(r))
		copy(cp, r)
		simulatedUserSheet[i] = cp
	}

	// Add custom column to header
	simulatedUserSheet[0] = append(simulatedUserSheet[0], "Candidate Private Notes (Custom)")

	// Swap data rows (simulate user sorting) and append custom notes
	dataRow1 := append(initialRows[2], "Candidate Private Note: Top choice Cloudflare")
	dataRow2 := append(initialRows[1], "Candidate Private Note: Stripe interview")
	simulatedUserSheet[1] = dataRow1 // lead-alpha-002
	simulatedUserSheet[2] = dataRow2 // lead-alpha-001

	// Update lead-alpha-001 in database (e.g. status changed from in_dialogue to interview_scheduled)
	lead1.Status = linkedin.LeadStatusInterviewScheduled
	lead1.UpdatedAt = time.Now().UTC()
	_ = repo.SaveRecruiterLead(ctx, lead1)

	// Step 3: Re-syncing sheet with user custom columns and sorted rows
	reconciledRows, result2, err := svc.SyncRelationshipsToSheets(ctx, simulatedUserSheet, config, "tenant-alpha", "usr-alex-001")
	if err != nil {
		t.Fatalf("reconciled sync failed: %v", err)
	}

	// Assertions for AT-014:
	// 1. Zero duplicate rows created (total rows remains 1 header + 2 leads = 3)
	if len(reconciledRows) != 3 {
		t.Errorf("expected exactly 3 rows, got %d (duplicate rows created!)", len(reconciledRows))
	}
	if result2.AppendedCount != 0 {
		t.Errorf("expected 0 appended, got %d", result2.AppendedCount)
	}
	if result2.UpdatedCount != 1 {
		t.Errorf("expected 1 updated row, got %d", result2.UpdatedCount)
	}
	if result2.UnchangedCount != 1 {
		t.Errorf("expected 1 unchanged row, got %d", result2.UnchangedCount)
	}

	// 2. Custom column preserved in both rows
	for rIdx := 1; rIdx < len(reconciledRows); rIdx++ {
		row := reconciledRows[rIdx]
		if len(row) <= standardHeaderCount {
			t.Errorf("row %d lost user custom column! row: %+v", rIdx, row)
		}
	}

	// Verify the updated row reflects interview_scheduled
	// Row 2 is lead-alpha-001 in the sorted sheet
	if reconciledRows[2][5] != string(linkedin.LeadStatusInterviewScheduled) {
		t.Errorf("expected row 2 status to be updated to interview_scheduled, got %s", reconciledRows[2][5])
	}
	if reconciledRows[2][standardHeaderCount] != "Candidate Private Note: Stripe interview" {
		t.Errorf("expected preserved custom note on row 2, got %s", reconciledRows[2][standardHeaderCount])
	}
}
