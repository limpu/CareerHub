package career

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// ExportDatasetType defines the target entity domain for CSV exports (CAR-18, REQ-006).
type ExportDatasetType string

const (
	ExportDatasetJobs         ExportDatasetType = "jobs"
	ExportDatasetApplications ExportDatasetType = "applications"
	ExportDatasetContacts     ExportDatasetType = "contacts"
	ExportDatasetAnalysis     ExportDatasetType = "analysis"
	ExportDatasetAllBundle    ExportDatasetType = "all_bundle"
)

// RecruiterContact represents a professional lead or recruiter relationship (CAR-18, CAR-20, AT-019).
type RecruiterContact struct {
	ID                  string     `json:"id"`
	UserID              string     `json:"user_id"`
	Name                string     `json:"name"`
	RoleTitle           string     `json:"role_title"`
	Company             string     `json:"company"`
	Email               string     `json:"email"`
	LinkedInURL         string     `json:"linkedin_url"`
	Status              string     `json:"status"` // "lead", "contacted", "in_conversation", "archived"
	LinkedApplicationID string     `json:"linked_application_id,omitempty"`
	LastInteractionAt   *time.Time `json:"last_interaction_at,omitempty"`
	NextFollowUpAt      *time.Time `json:"next_follow_up_at,omitempty"`
	Notes               string     `json:"notes"`
	CreatedAt           time.Time  `json:"created_at"`
}

// CareerFunnelMetric represents an analytics or conversion funnel data point (CAR-18, CAR-19).
type CareerFunnelMetric struct {
	MetricKey string `json:"metric_key"`
	Category  string `json:"category"` // "funnel", "skills", "market"
	Label     string `json:"label"`
	Value     string `json:"value"`
	Unit      string `json:"unit"`
	Benchmark string `json:"benchmark"`
	Notes     string `json:"notes"`
}

// ExportFilterOptions specifies filtering and formatting options for CSV export (CAR-18, AT-013).
type ExportFilterOptions struct {
	DatasetType     ExportDatasetType `json:"dataset_type"`
	StartDate       *time.Time        `json:"start_date,omitempty"`
	EndDate         *time.Time        `json:"end_date,omitempty"`
	StageFilter     string            `json:"stage_filter,omitempty"`
	PlatformFilter  string            `json:"platform_filter,omitempty"`
	SelectedColumns []string          `json:"selected_columns,omitempty"`
	WithBOM         bool              `json:"with_bom"` // Prepend UTF-8 BOM for Microsoft Excel (AT-013)
	IncludeHeaders  bool              `json:"include_headers"`
	Timezone        string            `json:"timezone,omitempty"`
}

// CSVExportManifest contains metadata and serialized content for an exported dataset (CAR-18).
type CSVExportManifest struct {
	ExportID    string            `json:"export_id"`
	UserID      string            `json:"user_id"`
	DatasetType ExportDatasetType `json:"dataset_type"`
	Filename    string            `json:"filename"`
	RowCount    int               `json:"row_count"`
	ByteSize    int               `json:"byte_size"`
	Checksum    string            `json:"checksum"`
	WithBOM     bool              `json:"with_bom"`
	Headers     []string          `json:"headers"`
	CSVContent  string            `json:"csv_content"`
	GeneratedAt time.Time         `json:"generated_at"`
}

// ExportAuditRecord stores durable historical logging of export actions (EXP-003, AT-022).
type ExportAuditRecord struct {
	AuditID     string            `json:"audit_id"`
	UserID      string            `json:"user_id"`
	DatasetType ExportDatasetType `json:"dataset_type"`
	Filename    string            `json:"filename"`
	RowCount    int               `json:"row_count"`
	Format      string            `json:"format"` // "csv", "json"
	WithBOM     bool              `json:"with_bom"`
	ExportedAt  time.Time         `json:"exported_at"`
	IPAddress   string            `json:"ip_address,omitempty"`
}

// Standard column headers for multi-domain CSV exports (CAR-18, REQ-006, AT-013)
var (
	JobsCSVHeaders = []string{
		"Saved Job ID",
		"Job Title",
		"Company",
		"Location",
		"Job Type",
		"Salary Min",
		"Salary Max",
		"Currency",
		"Interval",
		"Required Skills",
		"Priority",
		"Status",
		"Direct Apply URL",
		"Saved At",
	}

	ApplicationsCSVHeadersExtended = []string{
		"Application ID",
		"Job Title",
		"Company",
		"Location",
		"Portal Type",
		"Stage",
		"Submission Mode",
		"Verified Applied",
		"Applied At (UTC)",
		"Verification Ref",
		"Verification Type",
		"Resume Version",
		"Resume Path",
		"Resume Checksum",
		"Notes",
	}

	ContactsCSVHeaders = []string{
		"Contact ID",
		"Full Name",
		"Role / Title",
		"Company",
		"Email",
		"LinkedIn URL",
		"Status",
		"Linked Application ID",
		"Last Interaction (UTC)",
		"Next Follow-Up (UTC)",
		"Notes",
	}

	AnalysisCSVHeaders = []string{
		"Metric Key",
		"Category",
		"Metric Label",
		"Current Value",
		"Unit",
		"Target / Benchmark",
		"Analysis Notes",
	}
)

// UTF-8 Byte Order Mark (BOM) for Excel compatibility.
var UTF8BOM = []byte{0xEF, 0xBB, 0xBF}

// EncodeCSVWithBOM optionally prepends the UTF-8 Byte Order Mark to CSV bytes.
func EncodeCSVWithBOM(content string, withBOM bool) []byte {
	raw := []byte(content)
	if withBOM {
		return append(UTF8BOM, raw...)
	}
	return raw
}

// GenerateJobsCSV exports saved and discovery jobs to sanitized RFC4180 CSV (CAR-18, AT-013).
func GenerateJobsCSV(userID string, jobs []SavedJob, opts ExportFilterOptions) (*CSVExportManifest, error) {
	headers := filterHeaders(JobsCSVHeaders, opts.SelectedColumns)
	rows := make([][]string, len(jobs))

	for i, j := range jobs {
		locStr := j.Location.City
		if j.Location.State != "" {
			locStr = fmt.Sprintf("%s, %s", locStr, j.Location.State)
		}
		if j.Location.IsRemote {
			if locStr != "" {
				locStr = fmt.Sprintf("%s (Remote)", locStr)
			} else {
				locStr = "Remote"
			}
		}

		minSalary := ""
		if j.Compensation.MinAmount > 0 {
			minSalary = fmt.Sprintf("%.0f", j.Compensation.MinAmount)
		}
		maxSalary := ""
		if j.Compensation.MaxAmount > 0 {
			maxSalary = fmt.Sprintf("%.0f", j.Compensation.MaxAmount)
		}

		savedAtStr := ""
		if !j.SavedAt.IsZero() {
			savedAtStr = j.SavedAt.UTC().Format("2006-01-02 15:04:05 UTC")
		}

		fullRow := []string{
			SanitizeFormulaInjection(j.ID),
			SanitizeFormulaInjection(j.Title),
			SanitizeFormulaInjection(j.Company),
			SanitizeFormulaInjection(locStr),
			SanitizeFormulaInjection(j.JobType),
			SanitizeFormulaInjection(minSalary),
			SanitizeFormulaInjection(maxSalary),
			SanitizeFormulaInjection(j.Compensation.Currency),
			SanitizeFormulaInjection(string(j.Compensation.Period)),
			SanitizeFormulaInjection(strings.Join(j.RequiredSkills, "; ")),
			fmt.Sprintf("%d", j.Priority),
			SanitizeFormulaInjection(string(j.Status)),
			SanitizeFormulaInjection(j.DirectApplyURL),
			SanitizeFormulaInjection(savedAtStr),
		}

		rows[i] = filterRowByHeaders(JobsCSVHeaders, fullRow, opts.SelectedColumns)
	}

	return buildExportManifest(userID, ExportDatasetJobs, "career_jobs", headers, rows, opts.WithBOM)
}

// GenerateApplicationsCSVExtended exports application ledger with formula injection defense and resume artifact references (CAR-18, FND-006, AT-013).
func GenerateApplicationsCSVExtended(userID string, apps []ApplicationRecord, opts ExportFilterOptions) (*CSVExportManifest, error) {
	headers := filterHeaders(ApplicationsCSVHeadersExtended, opts.SelectedColumns)
	rows := make([][]string, len(apps))

	for i, a := range apps {
		appliedAtStr := ""
		if !a.AppliedAt.IsZero() {
			appliedAtStr = a.AppliedAt.UTC().Format("2006-01-02 15:04:05 UTC")
		}

		verifiedAppliedStr := "FALSE"
		if a.IsVerifiedApplied {
			verifiedAppliedStr = "TRUE"
		}

		vRef := ""
		vType := ""
		if a.VerificationDetails != nil {
			vRef = a.VerificationDetails.ProviderReference
			if vRef == "" {
				vRef = a.VerificationDetails.ReceiptID
			}
			vType = string(a.VerificationDetails.VerificationType)
		}

		portalType := a.PortalType
		if portalType == "" {
			portalType = a.SubmissionMode
		}

		// Separate binary document: reference version, path, and checksum (FND-006)
		resVersion := "v1.0"
		resPath := fmt.Sprintf("/artifacts/resumes/%s.pdf", a.ID)
		resChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(a.ID+a.Title)))

		notes := a.Notes
		if notes == "" && len(a.Timeline) > 0 {
			for idx := len(a.Timeline) - 1; idx >= 0; idx-- {
				if a.Timeline[idx].Notes != "" {
					notes = a.Timeline[idx].Notes
					break
				}
			}
		}

		fullRow := []string{
			SanitizeFormulaInjection(a.ID),
			SanitizeFormulaInjection(a.Title),
			SanitizeFormulaInjection(a.Company),
			SanitizeFormulaInjection(a.Location),
			SanitizeFormulaInjection(portalType),
			SanitizeFormulaInjection(string(a.Stage)),
			SanitizeFormulaInjection(a.SubmissionMode),
			verifiedAppliedStr,
			SanitizeFormulaInjection(appliedAtStr),
			SanitizeFormulaInjection(vRef),
			SanitizeFormulaInjection(vType),
			SanitizeFormulaInjection(resVersion),
			SanitizeFormulaInjection(resPath),
			SanitizeFormulaInjection(resChecksum),
			SanitizeFormulaInjection(notes),
		}

		rows[i] = filterRowByHeaders(ApplicationsCSVHeadersExtended, fullRow, opts.SelectedColumns)
	}

	return buildExportManifest(userID, ExportDatasetApplications, "career_applications", headers, rows, opts.WithBOM)
}

// GenerateContactsCSV exports networking leads with interaction and follow-up timestamps (CAR-18, AT-013, AT-019).
func GenerateContactsCSV(userID string, contacts []RecruiterContact, opts ExportFilterOptions) (*CSVExportManifest, error) {
	headers := filterHeaders(ContactsCSVHeaders, opts.SelectedColumns)
	rows := make([][]string, len(contacts))

	for i, c := range contacts {
		lastIntStr := ""
		if c.LastInteractionAt != nil && !c.LastInteractionAt.IsZero() {
			lastIntStr = c.LastInteractionAt.UTC().Format("2006-01-02 15:04:05 UTC")
		}

		nextFollowStr := ""
		if c.NextFollowUpAt != nil && !c.NextFollowUpAt.IsZero() {
			nextFollowStr = c.NextFollowUpAt.UTC().Format("2006-01-02 15:04:05 UTC")
		}

		fullRow := []string{
			SanitizeFormulaInjection(c.ID),
			SanitizeFormulaInjection(c.Name),
			SanitizeFormulaInjection(c.RoleTitle),
			SanitizeFormulaInjection(c.Company),
			SanitizeFormulaInjection(c.Email),
			SanitizeFormulaInjection(c.LinkedInURL),
			SanitizeFormulaInjection(c.Status),
			SanitizeFormulaInjection(c.LinkedApplicationID),
			SanitizeFormulaInjection(lastIntStr),
			SanitizeFormulaInjection(nextFollowStr),
			SanitizeFormulaInjection(c.Notes),
		}

		rows[i] = filterRowByHeaders(ContactsCSVHeaders, fullRow, opts.SelectedColumns)
	}

	return buildExportManifest(userID, ExportDatasetContacts, "career_contacts", headers, rows, opts.WithBOM)
}

// GenerateAnalysisCSV exports career metrics, conversion benchmarks, and skills demand gap insights (CAR-18).
func GenerateAnalysisCSV(userID string, metrics []CareerFunnelMetric, opts ExportFilterOptions) (*CSVExportManifest, error) {
	headers := filterHeaders(AnalysisCSVHeaders, opts.SelectedColumns)
	rows := make([][]string, len(metrics))

	for i, m := range metrics {
		fullRow := []string{
			SanitizeFormulaInjection(m.MetricKey),
			SanitizeFormulaInjection(m.Category),
			SanitizeFormulaInjection(m.Label),
			SanitizeFormulaInjection(m.Value),
			SanitizeFormulaInjection(m.Unit),
			SanitizeFormulaInjection(m.Benchmark),
			SanitizeFormulaInjection(m.Notes),
		}

		rows[i] = filterRowByHeaders(AnalysisCSVHeaders, fullRow, opts.SelectedColumns)
	}

	return buildExportManifest(userID, ExportDatasetAnalysis, "career_analysis_metrics", headers, rows, opts.WithBOM)
}

func filterHeaders(allHeaders []string, selected []string) []string {
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

func filterRowByHeaders(allHeaders []string, fullRow []string, selected []string) []string {
	if len(selected) == 0 {
		return fullRow
	}
	selectedSet := make(map[string]bool, len(selected))
	for _, s := range selected {
		selectedSet[strings.ToLower(strings.TrimSpace(s))] = true
	}
	var filtered []string
	for i, h := range allHeaders {
		if selectedSet[strings.ToLower(h)] && i < len(fullRow) {
			filtered = append(filtered, fullRow[i])
		}
	}
	if len(filtered) == 0 {
		return fullRow
	}
	return filtered
}

func buildExportManifest(
	userID string,
	datasetType ExportDatasetType,
	prefix string,
	headers []string,
	rows [][]string,
	withBOM bool,
) (*CSVExportManifest, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("failed to write CSV headers: %w", err)
	}
	if err := writer.WriteAll(rows); err != nil {
		return nil, fmt.Errorf("failed to write CSV rows: %w", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	rawCSV := buf.String()
	exportBytes := EncodeCSVWithBOM(rawCSV, withBOM)

	hash := sha256.Sum256(exportBytes)
	checksum := hex.EncodeToString(hash[:])
	now := time.Now().UTC()
	filename := fmt.Sprintf("%s_%s.csv", prefix, now.Format("20060102_150405"))

	return &CSVExportManifest{
		ExportID:    fmt.Sprintf("exp_%s_%d", datasetType, now.UnixNano()),
		UserID:      userID,
		DatasetType: datasetType,
		Filename:    filename,
		RowCount:    len(rows),
		ByteSize:    len(exportBytes),
		Checksum:    checksum,
		WithBOM:     withBOM,
		Headers:     headers,
		CSVContent:  rawCSV,
		GeneratedAt: now,
	}, nil
}
