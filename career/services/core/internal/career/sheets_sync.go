package career

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Standard error definitions for Google Sheets Sync (CAR-17, REQ-006, AT-013, AT-014)
var (
	ErrInvalidSheetsConfig    = errors.New("invalid sheets sync config: missing spreadsheet_id or sheet_name")
	ErrUnsupportedSyncMode    = errors.New("unsupported sync mode: must be one_way_upsert, append_only, or snapshot_overwrite")
	ErrEmptySyncApplicationSet = errors.New("cannot sync empty application set")
)

// SheetsSyncMode defines how records project into Google Sheets.
type SheetsSyncMode string

const (
	SyncModeOneWayUpsert       SheetsSyncMode = "one_way_upsert"
	SyncModeAppendOnly         SheetsSyncMode = "append_only"
	SyncModeSnapshotOverwrite  SheetsSyncMode = "snapshot_overwrite"
)

// Standard Header for Google Sheets Projection (CAR-17, REQ-006)
var DefaultSheetHeaders = []string{
	"Application ID",
	"Job Title",
	"Company",
	"Location",
	"Portal Type",
	"Stage",
	"Verified Applied",
	"Applied At (UTC)",
	"Provider Reference",
	"Verification Type",
	"Platform Notes",
}

// SheetsSyncConfig represents the user's configured Google Sheet export target (CAR-17, EXP-001).
type SheetsSyncConfig struct {
	ConfigID      string         `json:"config_id"`
	UserID        string         `json:"user_id"`
	SpreadsheetID string         `json:"spreadsheet_id"`
	SheetName     string         `json:"sheet_name"`
	Timezone      string         `json:"timezone"`
	SyncMode      SheetsSyncMode `json:"sync_mode"`
	IncludeNotes  bool           `json:"include_notes"`
	CreatedAt     time.Time      `json:"created_at"`
	LastSyncedAt  *time.Time     `json:"last_synced_at,omitempty"`
}

// SheetsSyncResult details the outcome of an idempotent sync reconciliation run (AT-014).
type SheetsSyncResult struct {
	SpreadsheetID  string    `json:"spreadsheet_id"`
	SheetName      string    `json:"sheet_name"`
	TotalSynced    int       `json:"total_synced"`
	AppendedCount  int       `json:"appended_count"`
	UpdatedCount   int       `json:"updated_count"`
	UnchangedCount int       `json:"unchanged_count"`
	Checksum       string    `json:"checksum"`
	SyncedAt       time.Time `json:"synced_at"`
	Message        string    `json:"message"`
}

// SheetsExportPayload contains raw rows and ready-to-paste/import CSV text (AT-013).
type SheetsExportPayload struct {
	SpreadsheetID string     `json:"spreadsheet_id,omitempty"`
	SheetName     string     `json:"sheet_name,omitempty"`
	Headers       []string   `json:"headers"`
	Rows          [][]string `json:"rows"`
	CSVContent    string     `json:"csv_content"`
	GeneratedAt   time.Time  `json:"generated_at"`
	RowCount      int        `json:"row_count"`
}

// SanitizeFormulaInjection neutralizes CSV / Spreadsheet formula injection vulnerabilities (AT-013).
// If a cell begins with '=', '+', '-', '@', '\t', or '\r', it prefixes a single quote (')
// so spreadsheet engines interpret it strictly as literal text, never as executable code.
func SanitizeFormulaInjection(value string) string {
	if value == "" {
		return ""
	}

	// Examine trimmed leading character
	trimmed := strings.TrimLeft(value, " ")
	if len(trimmed) == 0 {
		return value
	}

	firstChar := trimmed[0]
	switch firstChar {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + value
	default:
		return value
	}
}

// BuildApplicationSheetRow converts an ApplicationRecord into a sanitized spreadsheet row (AT-013).
func BuildApplicationSheetRow(app ApplicationRecord, loc *time.Location, includeNotes bool) []string {
	if loc == nil {
		loc = time.UTC
	}

	appliedAtStr := ""
	if !app.AppliedAt.IsZero() {
		appliedAtStr = app.AppliedAt.In(loc).Format("2006-01-02 15:04:05 MST")
	} else if !app.UpdatedAt.IsZero() {
		appliedAtStr = app.UpdatedAt.In(loc).Format("2006-01-02 15:04:05 MST")
	}

	verifiedAppliedStr := "FALSE"
	if app.IsVerifiedApplied {
		verifiedAppliedStr = "TRUE"
	}

	ref := ""
	vType := ""
	if app.VerificationDetails != nil {
		ref = app.VerificationDetails.ProviderReference
		if ref == "" {
			ref = app.VerificationDetails.ReceiptID
		}
		vType = string(app.VerificationDetails.VerificationType)
	}

	portalType := app.PortalType
	if portalType == "" {
		portalType = app.SubmissionMode
	}

	notes := ""
	if includeNotes {
		if app.Notes != "" {
			notes = app.Notes
		} else if len(app.Timeline) > 0 {
			// Capture the latest non-empty timeline note
			for i := len(app.Timeline) - 1; i >= 0; i-- {
				if app.Timeline[i].Notes != "" {
					notes = app.Timeline[i].Notes
					break
				}
			}
		}
	}

	row := []string{
		SanitizeFormulaInjection(app.ID),
		SanitizeFormulaInjection(app.Title),
		SanitizeFormulaInjection(app.Company),
		SanitizeFormulaInjection(app.Location),
		SanitizeFormulaInjection(portalType),
		SanitizeFormulaInjection(string(app.Stage)),
		verifiedAppliedStr,
		SanitizeFormulaInjection(appliedAtStr),
		SanitizeFormulaInjection(ref),
		SanitizeFormulaInjection(vType),
		SanitizeFormulaInjection(notes),
	}

	return row
}

// ReconcileSheetProjection performs an idempotent one-way reconciliation into spreadsheet rows (CAR-17, AT-014).
// Rule 1: Platform-owned Application ID (column 0) is the immutable primary key.
// Rule 2: Re-syncing after candidate sorting or filtering updates rows in-place without duplicating.
// Rule 3: User custom columns (columns > 10) are strictly preserved, not overwritten.
func ReconcileSheetProjection(
	existingRows [][]string,
	apps []ApplicationRecord,
	config SheetsSyncConfig,
) ([][]string, SheetsSyncResult, error) {
	if config.SpreadsheetID == "" || config.SheetName == "" {
		return nil, SheetsSyncResult{}, ErrInvalidSheetsConfig
	}

	loc := time.UTC
	if config.Timezone != "" {
		if l, err := time.LoadLocation(config.Timezone); err == nil {
			loc = l
		}
	}

	appended := 0
	updated := 0
	unchanged := 0

	// Case 1: Fresh or empty sheet - initialize standard headers and rows
	if len(existingRows) == 0 {
		outRows := make([][]string, 0, len(apps)+1)
		outRows = append(outRows, DefaultSheetHeaders)

		for _, app := range apps {
			row := BuildApplicationSheetRow(app, loc, config.IncludeNotes)
			outRows = append(outRows, row)
			appended++
		}

		result := SheetsSyncResult{
			SpreadsheetID:  config.SpreadsheetID,
			SheetName:      config.SheetName,
			TotalSynced:    len(apps),
			AppendedCount:  appended,
			UpdatedCount:   0,
			UnchangedCount: 0,
			SyncedAt:       time.Now().UTC(),
			Checksum:       computeRowsChecksum(outRows),
			Message:        fmt.Sprintf("Initialized sheet with %d new application records.", appended),
		}

		return outRows, result, nil
	}

	// Case 2: Existing sheet with header - map existing rows by Application ID
	// Verify header row presence
	header := existingRows[0]
	idColIdx := 0 // Column 0 is Application ID

	// Map existing row index by Application ID
	// Skip header (row 0)
	existingIDToRowIdx := make(map[string]int)
	for rIdx := 1; rIdx < len(existingRows); rIdx++ {
		row := existingRows[rIdx]
		if len(row) > idColIdx {
			appID := strings.TrimPrefix(row[idColIdx], "'") // Unquote if formula-escaped
			if appID != "" {
				existingIDToRowIdx[appID] = rIdx
			}
		}
	}

	// Deep copy existing rows so we can mutate safely
	outRows := make([][]string, len(existingRows))
	for i, r := range existingRows {
		copiedRow := make([]string, len(r))
		copy(copiedRow, r)
		outRows[i] = copiedRow
	}

	// Reconcile each platform application
	for _, app := range apps {
		newCoreCells := BuildApplicationSheetRow(app, loc, config.IncludeNotes)

		if rIdx, found := existingIDToRowIdx[app.ID]; found {
			// Existing row found: update core platform columns, preserve user custom columns (AT-014)
			existingRow := outRows[rIdx]

			isModified := false
			// Check if any core cell changed
			for c := 0; c < len(DefaultSheetHeaders); c++ {
				existingVal := ""
				if c < len(existingRow) {
					existingVal = existingRow[c]
				}
				newVal := newCoreCells[c]
				if existingVal != newVal {
					isModified = true
					break
				}
			}

			if isModified {
				// Construct merged row: new core cells + existing custom user cells beyond DefaultSheetHeaders
				mergedRow := make([]string, len(newCoreCells))
				copy(mergedRow, newCoreCells)

				if len(existingRow) > len(DefaultSheetHeaders) {
					// Append user's preserved custom columns
					mergedRow = append(mergedRow, existingRow[len(DefaultSheetHeaders):]...)
				}

				outRows[rIdx] = mergedRow
				updated++
			} else {
				unchanged++
			}
		} else {
			// New record not yet in sheet: append new row
			// If existing table has wider custom headers, pad with empty strings
			if len(header) > len(newCoreCells) {
				paddedRow := make([]string, len(header))
				copy(paddedRow, newCoreCells)
				outRows = append(outRows, paddedRow)
			} else {
				outRows = append(outRows, newCoreCells)
			}
			appended++
		}
	}

	result := SheetsSyncResult{
		SpreadsheetID:  config.SpreadsheetID,
		SheetName:      config.SheetName,
		TotalSynced:    len(apps),
		AppendedCount:  appended,
		UpdatedCount:   updated,
		UnchangedCount: unchanged,
		SyncedAt:       time.Now().UTC(),
		Checksum:       computeRowsChecksum(outRows),
		Message:        fmt.Sprintf("Idempotent sync complete: %d appended, %d updated, %d unchanged.", appended, updated, unchanged),
	}

	return outRows, result, nil
}

// GenerateSheetsCSV serializes rows into a valid RFC-4180 CSV string (AT-013).
func GenerateSheetsCSV(headers []string, rows [][]string) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if len(headers) > 0 {
		if err := w.Write(headers); err != nil {
			return "", fmt.Errorf("failed to write CSV headers: %w", err)
		}
	}

	for _, r := range rows {
		if err := w.Write(r); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return buf.String(), nil
}

// computeRowsChecksum generates a SHA-256 hash across all cells for integrity tracking.
func computeRowsChecksum(rows [][]string) string {
	h := sha256.New()
	for _, r := range rows {
		h.Write([]byte(strings.Join(r, "|") + "\n"))
	}
	return hex.EncodeToString(h.Sum(nil))
}
