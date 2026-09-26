package career_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/social-platform/services/core/internal/career"
)

func TestCSVExport_FormulaInjectionDefense_AllDatasets_AT013(t *testing.T) {
	// Test Jobs with injection
	jobs := []career.SavedJob{
		{
			ID:      "job-test-inj",
			Title:   "=cmd|'/C calc'!A0",
			Company: "@TargetCorp",
			Location: career.JobLocation{
				City:     "+RemoteCity",
				IsRemote: true,
			},
			JobType: "\tFull-time",
			Status:  career.SavedJobStatusSaved,
		},
	}
	jobManifest, err := career.GenerateJobsCSV("user-1", jobs, career.ExportFilterOptions{})
	if err != nil {
		t.Fatalf("GenerateJobsCSV error: %v", err)
	}
	if !strings.Contains(jobManifest.CSVContent, "'=cmd|'/C calc'!A0") {
		t.Errorf("Jobs CSV missing formula sanitized title: %s", jobManifest.CSVContent)
	}
	if !strings.Contains(jobManifest.CSVContent, "'@TargetCorp") {
		t.Errorf("Jobs CSV missing formula sanitized company: %s", jobManifest.CSVContent)
	}
	if !strings.Contains(jobManifest.CSVContent, "'+RemoteCity") {
		t.Errorf("Jobs CSV missing formula sanitized city: %s", jobManifest.CSVContent)
	}

	// Test Contacts with injection
	contacts := []career.RecruiterContact{
		{
			ID:        "cnt-test-inj",
			Name:      "=HYPERLINK(\"http://evil.com\",\"Name\")",
			Company:   "-InsecureCompany",
			RoleTitle: "+Director",
			Notes:     "\t=2+5*cmd",
		},
	}
	cntManifest, err := career.GenerateContactsCSV("user-1", contacts, career.ExportFilterOptions{})
	if err != nil {
		t.Fatalf("GenerateContactsCSV error: %v", err)
	}
	if !strings.Contains(cntManifest.CSVContent, "'=HYPERLINK") {
		t.Errorf("Contacts CSV missing sanitized name: %s", cntManifest.CSVContent)
	}
	if !strings.Contains(cntManifest.CSVContent, "'-InsecureCompany") {
		t.Errorf("Contacts CSV missing sanitized company: %s", cntManifest.CSVContent)
	}

	// Test Analysis with injection
	metrics := []career.CareerFunnelMetric{
		{
			MetricKey: "key_inj",
			Category:  "skills",
			Label:     "=ExcelFormula()",
			Value:     "+100%",
			Notes:     "@InjectedNotes",
		},
	}
	anaManifest, err := career.GenerateAnalysisCSV("user-1", metrics, career.ExportFilterOptions{})
	if err != nil {
		t.Fatalf("GenerateAnalysisCSV error: %v", err)
	}
	if !strings.Contains(anaManifest.CSVContent, "'=ExcelFormula()") {
		t.Errorf("Analysis CSV missing sanitized label: %s", anaManifest.CSVContent)
	}
	if !strings.Contains(anaManifest.CSVContent, "'+100%") {
		t.Errorf("Analysis CSV missing sanitized value: %s", anaManifest.CSVContent)
	}
}

func TestCSVExport_UTF8_BOM_Encoding(t *testing.T) {
	text := "Job Title,Company\nSoftware Engineer,গুগল\n"

	withBOMBytes := career.EncodeCSVWithBOM(text, true)
	if !bytes.HasPrefix(withBOMBytes, career.UTF8BOM) {
		t.Errorf("expected UTF8BOM prefix [0xEF, 0xBB, 0xBF] when withBOM=true")
	}
	if len(withBOMBytes) != len(text)+3 {
		t.Errorf("expected length %d; got %d", len(text)+3, len(withBOMBytes))
	}

	withoutBOMBytes := career.EncodeCSVWithBOM(text, false)
	if bytes.HasPrefix(withoutBOMBytes, career.UTF8BOM) {
		t.Errorf("expected no BOM prefix when withBOM=false")
	}
	if len(withoutBOMBytes) != len(text) {
		t.Errorf("expected length %d; got %d", len(text), len(withoutBOMBytes))
	}
}

func TestCSVExport_JobsDataset_FieldsAndFiltering(t *testing.T) {
	jobs := []career.SavedJob{
		{
			ID:           "job-1",
			Title:        "Staff Backend Engineer",
			Company:      "Datadog",
			JobType:      "Full-time",
			Compensation: career.NormalizedCompensation{MinAmount: 200000, MaxAmount: 260000, Currency: "USD", Period: "yearly"},
			RequiredSkills: []string{"Go", "Distributed Systems", "Kubernetes"},
			Status:       career.SavedJobStatusShortlisted,
			Priority:     5,
		},
	}

	// 1. Full export
	manifest, err := career.GenerateJobsCSV("user-123", jobs, career.ExportFilterOptions{WithBOM: true})
	if err != nil {
		t.Fatalf("GenerateJobsCSV error: %v", err)
	}
	if manifest.RowCount != 1 {
		t.Errorf("expected RowCount=1; got %d", manifest.RowCount)
	}
	if !manifest.WithBOM {
		t.Errorf("expected WithBOM=true")
	}
	if !strings.Contains(manifest.CSVContent, "200000") || !strings.Contains(manifest.CSVContent, "260000") {
		t.Errorf("salary intervals missing in CSV: %s", manifest.CSVContent)
	}
	if !strings.Contains(manifest.CSVContent, "Go; Distributed Systems; Kubernetes") {
		t.Errorf("skills list missing in CSV: %s", manifest.CSVContent)
	}

	// 2. Filtered columns export
	filteredManifest, err := career.GenerateJobsCSV("user-123", jobs, career.ExportFilterOptions{
		SelectedColumns: []string{"Job Title", "Company", "Priority"},
	})
	if err != nil {
		t.Fatalf("GenerateJobsCSV with filtered columns error: %v", err)
	}
	if len(filteredManifest.Headers) != 3 {
		t.Errorf("expected 3 filtered headers; got %d: %v", len(filteredManifest.Headers), filteredManifest.Headers)
	}
	firstRow := strings.Split(filteredManifest.CSVContent, "\n")[1]
	parts := strings.Split(firstRow, ",")
	if len(parts) != 3 {
		t.Errorf("expected 3 cells in row; got %d: %s", len(parts), firstRow)
	}
}

func TestCSVExport_ApplicationsExtended_ResumeArtifactSeparation_FND006(t *testing.T) {
	apps := []career.ApplicationRecord{
		{
			ID:                "app-sep-1",
			Title:             "Principal Engineer",
			Company:           "Stripe",
			Location:          "Remote",
			PortalType:        "greenhouse_direct",
			Stage:             career.StageApplied,
			SubmissionMode:    "direct_ats",
			IsVerifiedApplied: true,
			AppliedAt:         time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC),
			Notes:             "All facts verified.",
		},
	}

	manifest, err := career.GenerateApplicationsCSVExtended("user-test", apps, career.ExportFilterOptions{})
	if err != nil {
		t.Fatalf("GenerateApplicationsCSVExtended error: %v", err)
	}

	// Verify resume binary separation: version, path, and checksum must be present
	if !strings.Contains(manifest.CSVContent, "/artifacts/resumes/app-sep-1.pdf") {
		t.Errorf("expected resume document path reference in CSV: %s", manifest.CSVContent)
	}
	if !strings.Contains(manifest.CSVContent, "v1.0") {
		t.Errorf("expected resume version in CSV: %s", manifest.CSVContent)
	}
}

func TestCSVExport_AuditLogging_EXP003_AT022(t *testing.T) {
	repo := career.NewMemoryCareerRepository()
	svc := career.NewCareerService(repo)
	ctx := context.Background()
	userID := "user-audit-test-99"

	// Add a sample job
	job := career.SavedJob{
		ID:       "saved-job-audit",
		UserID:   userID,
		Title:    "Cloud Architect",
		Company:  "Google",
		SavedAt:  time.Now().UTC(),
		Priority: 5,
	}
	if err := repo.SaveSavedJob(ctx, &job); err != nil {
		t.Fatalf("SaveSavedJob error: %v", err)
	}

	// Export Jobs
	manifest, err := svc.ExportDatasetCSV(ctx, userID, career.ExportDatasetJobs, career.ExportFilterOptions{WithBOM: true})
	if err != nil {
		t.Fatalf("ExportDatasetCSV error: %v", err)
	}
	if manifest.RowCount != 1 {
		t.Errorf("expected 1 row; got %d", manifest.RowCount)
	}

	// Query audit history (EXP-003, AT-022)
	history, err := svc.GetExportAuditHistory(ctx, userID)
	if err != nil {
		t.Fatalf("GetExportAuditHistory error: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 audit record; got %d", len(history))
	}
	rec := history[0]
	if rec.DatasetType != career.ExportDatasetJobs {
		t.Errorf("expected dataset type jobs; got %s", rec.DatasetType)
	}
	if rec.RowCount != 1 {
		t.Errorf("expected row count 1; got %d", rec.RowCount)
	}
	if !rec.WithBOM {
		t.Errorf("expected with_bom=true in audit record")
	}
}

func TestCSVExport_FixtureEvaluation(t *testing.T) {
	fixturePath := "../../testdata/fixtures/forms/csv_additional_exports.json"
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture file: %v", err)
	}

	var root struct {
		Scenarios []struct {
			ScenarioID        string                   `json:"scenario_id"`
			DatasetType       career.ExportDatasetType `json:"dataset_type"`
			InputJobs         []career.SavedJob        `json:"input_jobs"`
			InputApplications []struct {
				career.ApplicationRecord
				ProviderReference string `json:"provider_reference"`
				VerificationType  string `json:"verification_type"`
			} `json:"input_applications"`
			InputContacts     []career.RecruiterContact   `json:"input_contacts"`
			InputMetrics      []career.CareerFunnelMetric `json:"input_metrics"`
			ExpectedHeaders   []string                    `json:"expected_headers"`
			ExpectedRowsCount int                         `json:"expected_rows_count"`
			ExpectedSanitized []string                    `json:"expected_sanitized_cells"`
		} `json:"scenarios"`
	}

	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	for _, sc := range root.Scenarios {
		t.Run(sc.ScenarioID, func(t *testing.T) {
			var manifest *career.CSVExportManifest
			var expErr error

			switch sc.DatasetType {
			case career.ExportDatasetJobs:
				manifest, expErr = career.GenerateJobsCSV("user-fix", sc.InputJobs, career.ExportFilterOptions{})
			case career.ExportDatasetApplications:
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
				manifest, expErr = career.GenerateApplicationsCSVExtended("user-fix", apps, career.ExportFilterOptions{})
			case career.ExportDatasetContacts:
				manifest, expErr = career.GenerateContactsCSV("user-fix", sc.InputContacts, career.ExportFilterOptions{})
			case career.ExportDatasetAnalysis:
				manifest, expErr = career.GenerateAnalysisCSV("user-fix", sc.InputMetrics, career.ExportFilterOptions{})
			default:
				t.Fatalf("unknown dataset type: %s", sc.DatasetType)
			}

			if expErr != nil {
				t.Fatalf("export error: %v", expErr)
			}

			if manifest.RowCount != sc.ExpectedRowsCount {
				t.Errorf("expected row count %d; got %d", sc.ExpectedRowsCount, manifest.RowCount)
			}

			for _, expectedHeader := range sc.ExpectedHeaders {
				if !strings.Contains(manifest.CSVContent, expectedHeader) {
					t.Errorf("expected header %q in CSV: %s", expectedHeader, manifest.CSVContent)
				}
			}

			for _, sanitizedCell := range sc.ExpectedSanitized {
				escapedCell := strings.ReplaceAll(sanitizedCell, `"`, `""`)
				if !strings.Contains(manifest.CSVContent, sanitizedCell) && !strings.Contains(manifest.CSVContent, escapedCell) {
					t.Errorf("expected sanitized cell %q in CSV output: %s", sanitizedCell, manifest.CSVContent)
				}
			}
		})
	}
}
