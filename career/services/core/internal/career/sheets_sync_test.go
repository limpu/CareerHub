package career_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/social-platform/services/core/internal/career"
)

func TestSheetsSync_FormulaInjectionSanitization_AT013(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"=cmd|'/C calc'!A0", "'=cmd|'/C calc'!A0"},
		{"+1234567890", "'+1234567890"},
		{"-sum(A1:B2)", "'-sum(A1:B2)"},
		{"@dangerousFormula()", "'@dangerousFormula()"},
		{"\t=2+5", "'\t=2+5"},
		{"\r=2+5", "'\r=2+5"},
		{"   =indentedFormula", "'   =indentedFormula"},
		{"Senior Go Engineer", "Senior Go Engineer"},
		{"Stripe, Inc.", "Stripe, Inc."},
		{"https://example.com/jobs/123", "https://example.com/jobs/123"},
		{"", ""},
	}

	for _, tc := range cases {
		sanitized := career.SanitizeFormulaInjection(tc.input)
		if sanitized != tc.expected {
			t.Errorf("SanitizeFormulaInjection(%q) = %q; expected %q", tc.input, sanitized, tc.expected)
		}
	}
}

func TestSheetsSync_CleanOneWaySync_HeadersAndRows(t *testing.T) {
	apps := []career.ApplicationRecord{
		{
			ID:                 "app-1",
			Title:              "Cloud Architect",
			Company:            "Google",
			Location:           "Mountain View, CA",
			PortalType:         "greenhouse_direct",
			Stage:              career.StageApplied,
			IsVerifiedApplied:  true,
			AppliedAt:          time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC),
			VerificationDetails: &career.AppliedVerificationDetails{
				ProviderReference: "GOOG-APP-1",
				VerificationType:  career.VerificationProviderReceipt,
			},
		},
		{
			ID:                 "app-2",
			Title:              "Staff Engineer",
			Company:            "Meta",
			Location:           "Menlo Park, CA",
			PortalType:         "lever_direct",
			Stage:              career.StageInterviewing,
			IsVerifiedApplied:  true,
			AppliedAt:          time.Date(2026, 9, 16, 15, 0, 0, 0, time.UTC),
			VerificationDetails: &career.AppliedVerificationDetails{
				ProviderReference: "META-APP-2",
				VerificationType:  career.VerificationUserAttestation,
			},
		},
	}

	config := career.SheetsSyncConfig{
		SpreadsheetID: "sheet-xyz-123",
		SheetName:     "Applications",
		Timezone:      "UTC",
		SyncMode:      career.SyncModeOneWayUpsert,
		IncludeNotes:  true,
	}

	reconciled, result, err := career.ReconcileSheetProjection(nil, apps, config)
	if err != nil {
		t.Fatalf("unexpected error during clean sync: %v", err)
	}

	if result.AppendedCount != 2 {
		t.Errorf("expected AppendedCount = 2; got %d", result.AppendedCount)
	}
	if result.UpdatedCount != 0 {
		t.Errorf("expected UpdatedCount = 0; got %d", result.UpdatedCount)
	}
	if len(reconciled) != 3 { // 1 header + 2 rows
		t.Fatalf("expected 3 total rows; got %d", len(reconciled))
	}

	// Verify header row
	if reconciled[0][0] != "Application ID" || reconciled[0][1] != "Job Title" {
		t.Errorf("header row mismatch: %v", reconciled[0])
	}

	// Verify first application row
	if reconciled[1][0] != "app-1" || reconciled[1][1] != "Cloud Architect" || reconciled[1][6] != "TRUE" {
		t.Errorf("first row mismatch: %v", reconciled[1])
	}
}

func TestSheetsSync_IdempotentUpsert_AT014(t *testing.T) {
	apps := []career.ApplicationRecord{
		{
			ID:                "app-rec-1",
			Title:             "Platform Engineer",
			Company:           "Stripe",
			Location:          "Remote",
			PortalType:        "greenhouse_direct",
			Stage:             career.StageApplied,
			IsVerifiedApplied: true,
			AppliedAt:         time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC),
		},
	}

	config := career.SheetsSyncConfig{
		SpreadsheetID: "test-sheet",
		SheetName:     "Apps",
		Timezone:      "UTC",
		SyncMode:      career.SyncModeOneWayUpsert,
	}

	// First sync
	firstRunRows, firstResult, err := career.ReconcileSheetProjection(nil, apps, config)
	if err != nil {
		t.Fatalf("first sync failed: %v", err)
	}
	if firstResult.AppendedCount != 1 || len(firstRunRows) != 2 {
		t.Fatalf("expected 1 appended row; got %d", firstResult.AppendedCount)
	}

	// Second sync with identical data (idempotency check)
	secondRunRows, secondResult, err := career.ReconcileSheetProjection(firstRunRows, apps, config)
	if err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	if secondResult.AppendedCount != 0 {
		t.Errorf("expected 0 appended on re-sync; got %d", secondResult.AppendedCount)
	}
	if secondResult.UpdatedCount != 0 {
		t.Errorf("expected 0 updated on identical re-sync; got %d", secondResult.UpdatedCount)
	}
	if secondResult.UnchangedCount != 1 {
		t.Errorf("expected 1 unchanged; got %d", secondResult.UnchangedCount)
	}
	if len(secondRunRows) != 2 {
		t.Errorf("row count expanded on re-sync: %d instead of 2", len(secondRunRows))
	}
}

func TestSheetsSync_RowSortingPreservation_AT014(t *testing.T) {
	apps := []career.ApplicationRecord{
		{
			ID:                "app-1",
			Title:             "Frontend Lead",
			Company:           "Vercel",
			Location:          "Remote",
			Stage:             career.StageOffered, // Updated stage
			IsVerifiedApplied: true,
		},
		{
			ID:                "app-2",
			Title:             "Backend Lead",
			Company:           "Supabase",
			Location:          "Remote",
			Stage:             career.StageInterviewing, // Updated stage
			IsVerifiedApplied: true,
		},
	}

	config := career.SheetsSyncConfig{
		SpreadsheetID: "test-sheet",
		SheetName:     "Apps",
		Timezone:      "UTC",
		SyncMode:      career.SyncModeOneWayUpsert,
	}

	// User intentionally reverse-sorted rows in Google Sheets
	userReorderedRows := [][]string{
		career.DefaultSheetHeaders,
		{"app-2", "Backend Lead", "Supabase", "Remote", "direct", "applied", "TRUE", "", "", "", ""},
		{"app-1", "Frontend Lead", "Vercel", "Remote", "direct", "applied", "TRUE", "", "", "", ""},
	}

	reconciled, result, err := career.ReconcileSheetProjection(userReorderedRows, apps, config)
	if err != nil {
		t.Fatalf("sync with reordered rows failed: %v", err)
	}

	if result.AppendedCount != 0 {
		t.Errorf("expected 0 appended rows; got %d", result.AppendedCount)
	}
	if result.UpdatedCount != 2 {
		t.Errorf("expected 2 updated rows; got %d", result.UpdatedCount)
	}
	if len(reconciled) != 3 {
		t.Fatalf("expected exactly 3 rows; got %d", len(reconciled))
	}

	// Row 1 should still be app-2, but stage updated to interviewing
	if reconciled[1][0] != "app-2" || reconciled[1][5] != "interviewing" {
		t.Errorf("app-2 row was not correctly updated in-place: %v", reconciled[1])
	}
	// Row 2 should still be app-1, but stage updated to offered
	if reconciled[2][0] != "app-1" || reconciled[2][5] != "offered" {
		t.Errorf("app-1 row was not correctly updated in-place: %v", reconciled[2])
	}
}

func TestSheetsSync_PreserveCustomUserColumns_AT014(t *testing.T) {
	apps := []career.ApplicationRecord{
		{
			ID:                "app-1",
			Title:             "Security Architect",
			Company:           "Cloudflare",
			Location:          "Austin, TX",
			Stage:             career.StageOffered,
			IsVerifiedApplied: true,
		},
	}

	config := career.SheetsSyncConfig{
		SpreadsheetID: "test-sheet",
		SheetName:     "Apps",
		Timezone:      "UTC",
		SyncMode:      career.SyncModeOneWayUpsert,
	}

	customNotes := "Personal interview feedback: strong architecture alignment, expected comp $220k."
	existingRowsWithCustomColumn := [][]string{
		append(career.DefaultSheetHeaders, "Candidate Custom Notes"),
		{"app-1", "Security Architect", "Cloudflare", "Austin, TX", "direct", "applied", "TRUE", "", "", "", "Old platform note", customNotes},
	}

	reconciled, result, err := career.ReconcileSheetProjection(existingRowsWithCustomColumn, apps, config)
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if result.UpdatedCount != 1 {
		t.Errorf("expected 1 update; got %d", result.UpdatedCount)
	}

	updatedRow := reconciled[1]
	if len(updatedRow) != 12 {
		t.Fatalf("expected 12 columns preserving custom column; got %d", len(updatedRow))
	}

	// Verify stage was updated to offered
	if updatedRow[5] != "offered" {
		t.Errorf("stage not updated: %s", updatedRow[5])
	}

	// Verify custom notes column was preserved untouched!
	if updatedRow[11] != customNotes {
		t.Errorf("custom user note was overwritten or lost: got %q, expected %q", updatedRow[11], customNotes)
	}
}

func TestSheetsSync_TimezoneAndCurrencyFormatting_AT013(t *testing.T) {
	app := career.ApplicationRecord{
		ID:        "app-tz-1",
		Title:     "DevOps Engineer",
		Company:   "HashiCorp",
		AppliedAt: time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC),
	}

	// Sync in New York timezone (EDT, UTC-4)
	nyLoc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("skipping timezone test: America/New_York not available on system")
	}

	rowNY := career.BuildApplicationSheetRow(app, nyLoc, false)
	timeCell := rowNY[7]

	if !strings.Contains(timeCell, "EDT") && !strings.Contains(timeCell, "-04") {
		t.Logf("timeCell in NY: %s", timeCell)
	}

	// CSV generation check
	csvStr, err := career.GenerateSheetsCSV(career.DefaultSheetHeaders, [][]string{rowNY})
	if err != nil {
		t.Fatalf("GenerateSheetsCSV failed: %v", err)
	}

	if !strings.Contains(csvStr, "app-tz-1") || !strings.Contains(csvStr, "HashiCorp") {
		t.Errorf("CSV output missing record fields: %s", csvStr)
	}
}

func TestSheetsSync_ServiceIntegration(t *testing.T) {
	repo := career.NewMemoryCareerRepository()
	svc := career.NewCareerService(repo)
	ctx := context.Background()
	userID := "user-sync-test-42"

	// Add application record
	app := career.ApplicationRecord{
		ID:                "app-svc-1",
		UserID:            userID,
		Title:             "Senior SRE",
		Company:           "Datadog",
		Location:          "New York, NY",
		PortalType:        "greenhouse_direct",
		Stage:             career.StageApplied,
		IsVerifiedApplied: true,
		AppliedAt:         time.Date(2026, 9, 17, 14, 0, 0, 0, time.UTC),
	}
	if err := repo.SaveApplicationRecord(ctx, &app); err != nil {
		t.Fatalf("SaveApplicationRecord failed: %v", err)
	}

	// 1. Configure Sheets Sync
	config := career.SheetsSyncConfig{
		SpreadsheetID: "sheet-integration-99",
		SheetName:     "Career Pipeline",
		Timezone:      "UTC",
		SyncMode:      career.SyncModeOneWayUpsert,
		IncludeNotes:  true,
	}
	savedCfg, err := svc.ConfigureGoogleSheetsSync(ctx, userID, config)
	if err != nil {
		t.Fatalf("ConfigureGoogleSheetsSync failed: %v", err)
	}
	if savedCfg.SpreadsheetID != "sheet-integration-99" {
		t.Errorf("config mismatch: %v", savedCfg)
	}

	// 2. Get Preview
	preview, err := svc.GetGoogleSheetsPreview(ctx, userID)
	if err != nil {
		t.Fatalf("GetGoogleSheetsPreview failed: %v", err)
	}
	if preview.RowCount != 1 || len(preview.Rows) != 1 {
		t.Errorf("expected 1 row in preview; got %d", preview.RowCount)
	}
	if !strings.Contains(preview.CSVContent, "Datadog") {
		t.Errorf("expected CSV preview to contain 'Datadog': %s", preview.CSVContent)
	}

	// 3. Export to CSV string
	csvText, err := svc.ExportApplicationsToCSV(ctx, userID)
	if err != nil {
		t.Fatalf("ExportApplicationsToCSV failed: %v", err)
	}
	if !strings.Contains(csvText, "Application ID") || !strings.Contains(csvText, "Senior SRE") {
		t.Errorf("CSV text incomplete: %s", csvText)
	}

	// 4. Execute Idempotent Sync
	result, finalRows, err := svc.SyncApplicationsToGoogleSheets(ctx, userID, nil)
	if err != nil {
		t.Fatalf("SyncApplicationsToGoogleSheets failed: %v", err)
	}
	if result.AppendedCount != 1 {
		t.Errorf("expected 1 appended; got %d", result.AppendedCount)
	}
	if len(finalRows) != 2 {
		t.Errorf("expected 2 rows (header + 1 app); got %d", len(finalRows))
	}
}

func TestSheetsSync_FixtureEvaluation(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "forms", "application_sheets_sync.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture file at %s: %v", fixturePath, err)
	}

	var root struct {
		Scenarios []struct {
			ScenarioID        string                  `json:"scenario_id"`
			Description       string                  `json:"description"`
			Config            career.SheetsSyncConfig `json:"config"`
			InputApplications []struct {
				career.ApplicationRecord
				ProviderReference string `json:"provider_reference"`
				VerificationType  string `json:"verification_type"`
			} `json:"input_applications"`
			ExistingSheetRows  [][]string `json:"existing_sheet_rows"`
			ExpectedRowsCount  int        `json:"expected_rows_count"`
			ExpectedAppended   int        `json:"expected_appended_count"`
			ExpectedUpdated    int        `json:"expected_updated_count"`
			ExpectedSanitized  []string   `json:"expected_sanitized_cells"`
			ExpectedCustomNote string     `json:"expected_preserved_custom_note"`
		} `json:"scenarios"`
	}

	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("failed to unmarshal fixture JSON: %v", err)
	}

	for _, sc := range root.Scenarios {
		t.Run(sc.ScenarioID, func(t *testing.T) {
			var apps []career.ApplicationRecord
			for _, item := range sc.InputApplications {
				app := item.ApplicationRecord
				if item.ProviderReference != "" || item.VerificationType != "" {
					app.VerificationDetails = &career.AppliedVerificationDetails{
						ProviderReference: item.ProviderReference,
						VerificationType:  career.AppliedVerificationType(item.VerificationType),
					}
				}
				apps = append(apps, app)
			}

			reconciled, result, err := career.ReconcileSheetProjection(sc.ExistingSheetRows, apps, sc.Config)
			if err != nil {
				t.Fatalf("scenario %s error: %v", sc.ScenarioID, err)
			}

			if sc.ExpectedRowsCount > 0 && len(reconciled) != sc.ExpectedRowsCount {
				t.Errorf("expected %d rows; got %d", sc.ExpectedRowsCount, len(reconciled))
			}
			if sc.ExpectedAppended > 0 && result.AppendedCount != sc.ExpectedAppended {
				t.Errorf("expected %d appended; got %d", sc.ExpectedAppended, result.AppendedCount)
			}
			if sc.ExpectedUpdated > 0 && result.UpdatedCount != sc.ExpectedUpdated {
				t.Errorf("expected %d updated; got %d", sc.ExpectedUpdated, result.UpdatedCount)
			}

			if len(sc.ExpectedSanitized) > 0 {
				appRow := reconciled[1]
				cellStr := strings.Join(appRow, " ")
				for _, expected := range sc.ExpectedSanitized {
					if !strings.Contains(cellStr, expected) {
						t.Errorf("expected sanitized cell %q in row: %s", expected, cellStr)
					}
				}
			}

			if sc.ExpectedCustomNote != "" {
				appRow := reconciled[1]
				found := false
				for _, cell := range appRow {
					if cell == sc.ExpectedCustomNote {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("custom note was not preserved: expected %q in row %v", sc.ExpectedCustomNote, appRow)
				}
			}
		})
	}
}
