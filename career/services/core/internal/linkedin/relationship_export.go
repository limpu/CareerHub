package linkedin

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

// RelationshipExportDestination defines where relationship records are directed (REQ-006, EXP-001).
type RelationshipExportDestination string

const (
	DestinationCSVDownload   RelationshipExportDestination = "csv_download"
	DestinationSheetsSync    RelationshipExportDestination = "google_sheets_sync"
)

var (
	ErrEmptyExportWorkspace          = errors.New("relationship_export: workspace_id is required")
	ErrEmptyExportTenant             = errors.New("relationship_export: tenant_id is required")
	ErrEmptyExportDestination        = errors.New("relationship_export: valid export destination is required")
	ErrInvalidSheetsExportConfig     = errors.New("relationship_export: invalid sheets sync config, spreadsheet_id and sheet_name required")
	ErrCrossTenantExportDenied       = errors.New("relationship_export: cross-tenant relationship export denied (AT-011, AT-012)")
	ErrUnauthorizedPrivateNotes      = errors.New("relationship_export: private notes require explicit user consent (include_notes: true)")
)

// Standard headers for Relationship CSV & Sheets exports (LI-20, AT-013).
var DefaultRelationshipHeaders = []string{
	"Lead ID",
	"Recruiter Name",
	"Recruiter Title",
	"Company",
	"LinkedIn URL",
	"Status",
	"Outreach Stage",
	"Source Entity Type",
	"Source Entity ID",
	"Related Job ID",
	"Related Application ID",
	"Person Location",
	"Person Headline",
	"Person Connection Tier",
	"Follow-Up Due Date",
	"Follow-Up Message",
	"Follow-Up Status",
	"Created At (UTC)",
	"Updated At (UTC)",
}

// NotesHeader defines the optional authorized notes column.
const NotesHeader = "Interaction Notes"

// UTF8BOM represents the UTF-8 Byte Order Mark for Microsoft Excel compatibility (AT-013).
var UTF8BOM = []byte{0xEF, 0xBB, 0xBF}

// RelationshipExportFilter configures export scoping, destinations, and format options.
type RelationshipExportFilter struct {
	WorkspaceID          string                        `json:"workspace_id"`
	TenantID             string                        `json:"tenant_id"`
	RequestingTenantID   string                        `json:"requesting_tenant_id"`
	RequestingOwnerID    string                        `json:"requesting_owner_id"`
	Destination          RelationshipExportDestination `json:"destination"`
	StatusFilter         LeadStatus                    `json:"status_filter,omitempty"`
	CompanyFilter        string                        `json:"company_filter,omitempty"`
	IncludeNotes         bool                          `json:"include_notes"`
	IncludePersonRecords bool                          `json:"include_person_records"`
	SelectedColumns      []string                      `json:"selected_columns,omitempty"`
	WithBOM              bool                          `json:"with_bom"`
	Timezone             string                        `json:"timezone,omitempty"`
}

// RelationshipExportManifest captures metadata and serialized RFC4180 CSV content.
type RelationshipExportManifest struct {
	ExportID    string                        `json:"export_id"`
	WorkspaceID string                        `json:"workspace_id"`
	TenantID    string                        `json:"tenant_id"`
	Destination RelationshipExportDestination `json:"destination"`
	Filename    string                        `json:"filename"`
	RowCount    int                           `json:"row_count"`
	ByteSize    int                           `json:"byte_size"`
	Checksum    string                        `json:"checksum"`
	WithBOM     bool                          `json:"with_bom"`
	Headers     []string                      `json:"headers"`
	CSVContent  string                        `json:"csv_content"`
	GeneratedAt time.Time                     `json:"generated_at"`
}

// RelationshipSheetsSyncConfig defines the Google Sheets one-way projection target.
type RelationshipSheetsSyncConfig struct {
	ConfigID      string    `json:"config_id"`
	WorkspaceID   string    `json:"workspace_id"`
	TenantID      string    `json:"tenant_id"`
	SpreadsheetID string    `json:"spreadsheet_id"`
	SheetName     string    `json:"sheet_name"`
	Timezone      string    `json:"timezone"`
	IncludeNotes  bool      `json:"include_notes"`
	SyncMode      string    `json:"sync_mode"` // "one_way_upsert"
	CreatedAt     time.Time `json:"created_at"`
	LastSyncedAt  *time.Time `json:"last_synced_at,omitempty"`
}

// RelationshipSheetsSyncResult summarizes the reconciliation run.
type RelationshipSheetsSyncResult struct {
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

// RelationshipExportAuditRecord tracks audit logging for exports (REQ-019, AT-022).
type RelationshipExportAuditRecord struct {
	AuditID      string                        `json:"audit_id"`
	WorkspaceID  string                        `json:"workspace_id"`
	TenantID     string                        `json:"tenant_id"`
	OwnerID      string                        `json:"owner_id"`
	Destination  RelationshipExportDestination `json:"destination"`
	Filename     string                        `json:"filename"`
	RowCount     int                           `json:"row_count"`
	IncludeNotes bool                          `json:"include_notes"`
	ExportedAt   time.Time                     `json:"exported_at"`
}

// SanitizeFormulaInjection neutralizes CSV / Spreadsheet formula injection vulnerabilities (AT-013).
// If a cell begins with '=', '+', '-', '@', '\t', or '\r', it prefixes a single quote (')
// so spreadsheet engines interpret it strictly as literal text, never as executable code.
func SanitizeFormulaInjection(value string) string {
	if value == "" {
		return ""
	}

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

// BuildRelationshipCSVHeaders returns standard headers optionally augmented with notes.
func BuildRelationshipCSVHeaders(includeNotes bool, selected []string) []string {
	allHeaders := make([]string, len(DefaultRelationshipHeaders))
	copy(allHeaders, DefaultRelationshipHeaders)

	if includeNotes {
		allHeaders = append(allHeaders, NotesHeader)
	}

	if len(selected) == 0 {
		return allHeaders
	}

	selectedSet := make(map[string]bool, len(selected))
	for _, s := range selected {
		selectedSet[strings.ToLower(strings.TrimSpace(s))] = true
	}

	var filtered []string
	for _, h := range allHeaders {
		if selectedSet[strings.ToLower(h)] {
			filtered = append(filtered, h)
		}
	}

	if len(filtered) == 0 {
		return allHeaders
	}
	return filtered
}

// BuildRelationshipRow converts a RecruiterLead and optional PersonRecord into a sanitized row.
func BuildRelationshipRow(
	lead RecruiterLead,
	person *PersonRecord,
	loc *time.Location,
	includeNotes bool,
	allHeaders []string,
	selectedHeaders []string,
) []string {
	if loc == nil {
		loc = time.UTC
	}

	tzLabel := loc.String()
	if tzLabel == "" {
		tzLabel = "UTC"
	}

	createdAtStr := ""
	if !lead.CreatedAt.IsZero() {
		createdAtStr = lead.CreatedAt.In(loc).Format(fmt.Sprintf("2006-01-02 15:04:05 (%s)", tzLabel))
	}
	updatedAtStr := ""
	if !lead.UpdatedAt.IsZero() {
		updatedAtStr = lead.UpdatedAt.In(loc).Format(fmt.Sprintf("2006-01-02 15:04:05 (%s)", tzLabel))
	}

	personLoc := ""
	personHeadline := ""
	personConnTier := ""
	if person != nil {
		personLoc = person.Location
		personHeadline = person.Headline
		personConnTier = person.ConnectionTier
	}

	reminderDueStr := ""
	reminderMsg := ""
	reminderStatus := ""
	if lead.Reminder != nil {
		if !lead.Reminder.DueDate.IsZero() {
			reminderDueStr = lead.Reminder.DueDate.In(loc).Format(fmt.Sprintf("2006-01-02 15:04:05 (%s)", tzLabel))
		}
		reminderMsg = lead.Reminder.Message
		reminderStatus = string(lead.Reminder.Status)
	}

	rowValues := []string{
		SanitizeFormulaInjection(lead.ID),
		SanitizeFormulaInjection(lead.RecruiterName),
		SanitizeFormulaInjection(lead.RecruiterTitle),
		SanitizeFormulaInjection(lead.Company),
		SanitizeFormulaInjection(lead.LinkedInURL),
		SanitizeFormulaInjection(string(lead.Status)),
		SanitizeFormulaInjection(string(lead.OutreachStage)),
		SanitizeFormulaInjection(lead.SourceEntityType),
		SanitizeFormulaInjection(lead.SourceEntityID),
		SanitizeFormulaInjection(lead.RelatedJobID),
		SanitizeFormulaInjection(lead.RelatedApplicationID),
		SanitizeFormulaInjection(personLoc),
		SanitizeFormulaInjection(personHeadline),
		SanitizeFormulaInjection(personConnTier),
		SanitizeFormulaInjection(reminderDueStr),
		SanitizeFormulaInjection(reminderMsg),
		SanitizeFormulaInjection(reminderStatus),
		SanitizeFormulaInjection(createdAtStr),
		SanitizeFormulaInjection(updatedAtStr),
	}

	if includeNotes {
		notesStr := ""
		if len(lead.Notes) > 0 {
			var noteSnippets []string
			for _, n := range lead.Notes {
				tStr := ""
				if !n.CreatedAt.IsZero() {
					tStr = n.CreatedAt.In(loc).Format("2006-01-02 15:04")
				}
				noteSnippets = append(noteSnippets, fmt.Sprintf("[%s by %s]: %s", tStr, n.Author, n.Content))
			}
			notesStr = strings.Join(noteSnippets, " | ")
		}
		rowValues = append(rowValues, SanitizeFormulaInjection(notesStr))
	}

	// Filter columns if subset requested
	if len(selectedHeaders) == 0 || len(selectedHeaders) == len(allHeaders) {
		return rowValues
	}

	selectedSet := make(map[string]bool, len(selectedHeaders))
	for _, s := range selectedHeaders {
		selectedSet[strings.ToLower(strings.TrimSpace(s))] = true
	}

	var filteredRow []string
	for i, h := range allHeaders {
		if selectedSet[strings.ToLower(h)] && i < len(rowValues) {
			filteredRow = append(filteredRow, rowValues[i])
		}
	}

	if len(filteredRow) == 0 {
		return rowValues
	}
	return filteredRow
}

// GenerateRelationshipCSV generates RFC4180 compliant CSV content for scoped recruiter leads.
func GenerateRelationshipCSV(
	leads []RecruiterLead,
	personMap map[string]*PersonRecord,
	filter RelationshipExportFilter,
) (*RelationshipExportManifest, error) {
	if strings.TrimSpace(filter.WorkspaceID) == "" {
		return nil, ErrEmptyExportWorkspace
	}
	if strings.TrimSpace(filter.TenantID) == "" {
		return nil, ErrEmptyExportTenant
	}

	// Enforce strict tenant isolation (AT-011, AT-012)
	if filter.RequestingTenantID != "" && filter.RequestingTenantID != filter.TenantID {
		return nil, ErrCrossTenantExportDenied
	}

	loc := time.UTC
	if filter.Timezone != "" {
		if l, err := time.LoadLocation(filter.Timezone); err == nil {
			loc = l
		}
	}

	allHeaders := BuildRelationshipCSVHeaders(filter.IncludeNotes, nil)
	activeHeaders := BuildRelationshipCSVHeaders(filter.IncludeNotes, filter.SelectedColumns)

	// Filter leads by criteria if provided
	var filteredLeads []RecruiterLead
	for _, lead := range leads {
		if lead.TenantID != filter.TenantID || lead.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.StatusFilter != "" && lead.Status != filter.StatusFilter {
			continue
		}
		if filter.CompanyFilter != "" && !strings.EqualFold(lead.Company, filter.CompanyFilter) {
			continue
		}
		filteredLeads = append(filteredLeads, lead)
	}

	rows := make([][]string, len(filteredLeads))
	for i, lead := range filteredLeads {
		var person *PersonRecord
		if filter.IncludePersonRecords && lead.SourceEntityID != "" && personMap != nil {
			person = personMap[lead.SourceEntityID]
		}
		rows[i] = BuildRelationshipRow(lead, person, loc, filter.IncludeNotes, allHeaders, activeHeaders)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err := writer.Write(activeHeaders); err != nil {
		return nil, fmt.Errorf("failed to write relationship CSV headers: %w", err)
	}
	if err := writer.WriteAll(rows); err != nil {
		return nil, fmt.Errorf("failed to write relationship CSV rows: %w", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush relationship CSV writer: %w", err)
	}

	rawCSV := buf.String()
	exportBytes := []byte(rawCSV)
	if filter.WithBOM {
		exportBytes = append(UTF8BOM, exportBytes...)
	}

	hash := sha256.Sum256(exportBytes)
	checksum := hex.EncodeToString(hash[:])
	now := time.Now().UTC()
	filename := fmt.Sprintf("linkedin_relationships_%s_%s.csv", filter.WorkspaceID, now.Format("20060102_150405"))

	return &RelationshipExportManifest{
		ExportID:    fmt.Sprintf("rel_exp_%d", now.UnixNano()),
		WorkspaceID: filter.WorkspaceID,
		TenantID:    filter.TenantID,
		Destination: DestinationCSVDownload,
		Filename:    filename,
		RowCount:    len(rows),
		ByteSize:    len(exportBytes),
		Checksum:    checksum,
		WithBOM:     filter.WithBOM,
		Headers:     activeHeaders,
		CSVContent:  rawCSV,
		GeneratedAt: now,
	}, nil
}

// ReconcileRelationshipSheetsProjection performs an idempotent one-way sync into spreadsheet rows (AT-014).
// Rule 1: Lead ID (column 0) is the immutable primary key.
// Rule 2: Re-syncing after candidate sorting or filtering updates rows in-place without duplicating.
// Rule 3: User custom columns (columns > standard column count) are strictly preserved, not overwritten.
func ReconcileRelationshipSheetsProjection(
	existingRows [][]string,
	leads []RecruiterLead,
	personMap map[string]*PersonRecord,
	config RelationshipSheetsSyncConfig,
) ([][]string, RelationshipSheetsSyncResult, error) {
	if strings.TrimSpace(config.SpreadsheetID) == "" || strings.TrimSpace(config.SheetName) == "" {
		return nil, RelationshipSheetsSyncResult{}, ErrInvalidSheetsExportConfig
	}

	loc := time.UTC
	if config.Timezone != "" {
		if l, err := time.LoadLocation(config.Timezone); err == nil {
			loc = l
		}
	}

	standardHeaders := BuildRelationshipCSVHeaders(config.IncludeNotes, nil)

	// Filter leads belonging to config's workspace & tenant
	var targetLeads []RecruiterLead
	for _, l := range leads {
		if l.WorkspaceID == config.WorkspaceID && l.TenantID == config.TenantID {
			targetLeads = append(targetLeads, l)
		}
	}

	appended := 0
	updated := 0
	unchanged := 0

	// Case 1: Empty or freshly initialized sheet
	if len(existingRows) == 0 {
		outRows := make([][]string, 0, len(targetLeads)+1)
		outRows = append(outRows, standardHeaders)

		for _, lead := range targetLeads {
			var person *PersonRecord
			if lead.SourceEntityID != "" && personMap != nil {
				person = personMap[lead.SourceEntityID]
			}
			row := BuildRelationshipRow(lead, person, loc, config.IncludeNotes, standardHeaders, standardHeaders)
			outRows = append(outRows, row)
			appended++
		}

		result := RelationshipSheetsSyncResult{
			SpreadsheetID:  config.SpreadsheetID,
			SheetName:      config.SheetName,
			TotalSynced:    len(targetLeads),
			AppendedCount:  appended,
			UpdatedCount:   0,
			UnchangedCount: 0,
			SyncedAt:       time.Now().UTC(),
			Checksum:       computeRowsChecksum(outRows),
			Message:        fmt.Sprintf("Initialized relationship sheet with %d records.", appended),
		}
		return outRows, result, nil
	}

	// Case 2: Sheet with existing data - Map rows by Lead ID (Column 0)
	idColIdx := 0
	existingIDToRowIdx := make(map[string]int)
	for rIdx := 1; rIdx < len(existingRows); rIdx++ {
		row := existingRows[rIdx]
		if len(row) > idColIdx {
			leadID := strings.TrimPrefix(row[idColIdx], "'") // Unquote if escaped
			if leadID != "" {
				existingIDToRowIdx[leadID] = rIdx
			}
		}
	}

	// Deep copy existing rows to safely mutate
	outRows := make([][]string, len(existingRows))
	for i, r := range existingRows {
		cp := make([]string, len(r))
		copy(cp, r)
		outRows[i] = cp
	}

	// Reconcile each target lead
	for _, lead := range targetLeads {
		var person *PersonRecord
		if lead.SourceEntityID != "" && personMap != nil {
			person = personMap[lead.SourceEntityID]
		}
		newCoreCells := BuildRelationshipRow(lead, person, loc, config.IncludeNotes, standardHeaders, standardHeaders)

		if rIdx, found := existingIDToRowIdx[lead.ID]; found {
			existingRow := outRows[rIdx]

			// Check if any core cell changed
			hasChanged := false
			for cIdx := 0; cIdx < len(newCoreCells); cIdx++ {
				if cIdx >= len(existingRow) || existingRow[cIdx] != newCoreCells[cIdx] {
					hasChanged = true
					break
				}
			}

			if hasChanged {
				// Reconstruct row: core platform cells + preserved user custom columns
				reconciledRow := make([]string, 0, len(existingRow))
				reconciledRow = append(reconciledRow, newCoreCells...)

				// Preserve any user-created custom columns past standard header length
				if len(existingRow) > len(standardHeaders) {
					customSuffix := existingRow[len(standardHeaders):]
					reconciledRow = append(reconciledRow, customSuffix...)
				}
				outRows[rIdx] = reconciledRow
				updated++
			} else {
				unchanged++
			}
		} else {
			// New lead not previously present on sheet: append
			outRows = append(outRows, newCoreCells)
			appended++
		}
	}

	result := RelationshipSheetsSyncResult{
		SpreadsheetID:  config.SpreadsheetID,
		SheetName:      config.SheetName,
		TotalSynced:    len(targetLeads),
		AppendedCount:  appended,
		UpdatedCount:   updated,
		UnchangedCount: unchanged,
		SyncedAt:       time.Now().UTC(),
		Checksum:       computeRowsChecksum(outRows),
		Message:        fmt.Sprintf("Idempotent sync reconciled: %d appended, %d updated, %d unchanged.", appended, updated, unchanged),
	}

	return outRows, result, nil
}

func computeRowsChecksum(rows [][]string) string {
	var buf bytes.Buffer
	for _, r := range rows {
		buf.WriteString(strings.Join(r, ","))
		buf.WriteString("\n")
	}
	hash := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(hash[:])
}
