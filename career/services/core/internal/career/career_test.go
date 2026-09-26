package career

import (
	"archive/zip"
	"bytes"
	"context"
	"strings"
	"testing"
)

// Helper to synthesize a valid, non-quarantined test PDF with BT...ET text blocks
func createMockPDF(name, email, skills string) []byte {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	buf.WriteString("1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n")
	buf.WriteString("2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj\n")
	buf.WriteString("3 0 obj << /Type /Page /Parent 2 0 R /Contents 4 0 R >> endobj\n")
	buf.WriteString("4 0 obj << /Length 200 >> stream\n")
	buf.WriteString("BT\n")
	buf.WriteString("/F1 12 Tf\n")
	buf.WriteString("(" + name + ") Tj\n")
	buf.WriteString("ET\n")
	buf.WriteString("BT\n")
	buf.WriteString("(" + email + ") Tj\n")
	buf.WriteString("ET\n")
	buf.WriteString("BT\n")
	buf.WriteString("(Skills: " + skills + ") Tj\n")
	buf.WriteString("ET\n")
	buf.WriteString("endstream\nendobj\n")
	buf.WriteString("xref\n0 5\n0000000000 65535 f \n")
	buf.WriteString("trailer << /Size 5 /Root 1 0 R >>\nstartxref\n500\n%%EOF\n")
	return buf.Bytes()
}

// Helper to synthesize a valid mock DOCX zip archive with word/document.xml
func createMockDOCX(name, email, skills string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	docXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>` + name + `</w:t></w:r></w:p>
    <w:p><w:r><w:t>` + email + `</w:t></w:r></w:p>
    <w:p><w:r><w:t>Professional Experience</w:t></w:r></w:p>
    <w:p><w:r><w:t>Staff Backend Engineer at Acme Corp</w:t></w:r></w:p>
    <w:p><w:r><w:t>Skills: ` + skills + `</w:t></w:r></w:p>
  </w:body>
</w:document>`

	w, _ := zw.Create("word/document.xml")
	_, _ = w.Write([]byte(docXML))
	_ = zw.Close()

	return buf.Bytes()
}

func TestCareerService_ProcessUploadedPDF(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryCareerRepository()
	svc := NewCareerService(repo)

	pdfData := createMockPDF("Jane Developer", "jane.dev@example.com", "Go, PostgreSQL, Docker, Kubernetes")

	draft, err := svc.ProcessUploadedDocument(ctx, "user_123", "up_001", "resume.pdf", "application/pdf", pdfData)
	if err != nil {
		t.Fatalf("ProcessUploadedDocument failed: %v", err)
	}

	if draft.UserID != "user_123" {
		t.Errorf("expected userID user_123, got %s", draft.UserID)
	}
	if draft.Status != ExtractionPending {
		t.Errorf("expected draft status pending, got %s", draft.Status)
	}
	if len(draft.ExtractedFacts) == 0 {
		t.Fatalf("expected extracted facts, got 0")
	}

	// Verify contact info extraction
	var foundName, foundEmail, foundSkills bool
	for _, f := range draft.ExtractedFacts {
		if f.Category == "contact" && f.Title == "Jane Developer" {
			foundName = true
			if f.SourceProvenance != ProvenanceUploadPDF {
				t.Errorf("expected provenance upload_pdf, got %s", f.SourceProvenance)
			}
		}
		if f.Category == "contact" && f.Title == "jane.dev@example.com" {
			foundEmail = true
		}
		if f.Category == "skill" && strings.Contains(f.Title, "Go") {
			foundSkills = true
		}
	}

	if !foundName {
		t.Errorf("expected candidate name extracted from PDF")
	}
	if !foundEmail {
		t.Errorf("expected candidate email extracted from PDF")
	}
	if !foundSkills {
		t.Errorf("expected recognized skills extracted from PDF")
	}
}

func TestCareerService_ProcessUploadedDOCX(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryCareerRepository()
	svc := NewCareerService(repo)

	docxData := createMockDOCX("Alex Smith", "alex.smith@example.com", "TypeScript, React, Next.js, Redis")

	draft, err := svc.ProcessUploadedDocument(
		ctx,
		"user_456",
		"up_002",
		"resume.docx",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		docxData,
	)
	if err != nil {
		t.Fatalf("ProcessUploadedDocument DOCX failed: %v", err)
	}

	if draft.UserID != "user_456" {
		t.Errorf("expected userID user_456, got %s", draft.UserID)
	}

	var foundName, foundEmail, foundSkills bool
	for _, f := range draft.ExtractedFacts {
		if f.Category == "contact" && f.Title == "Alex Smith" {
			foundName = true
			if f.SourceProvenance != ProvenanceUploadDOCX {
				t.Errorf("expected provenance upload_docx, got %s", f.SourceProvenance)
			}
		}
		if f.Category == "contact" && f.Title == "alex.smith@example.com" {
			foundEmail = true
		}
		if f.Category == "skill" && strings.Contains(f.Title, "TypeScript") {
			foundSkills = true
		}
	}

	if !foundName || !foundEmail || !foundSkills {
		t.Errorf("expected name, email, skills extracted from DOCX: name=%v email=%v skills=%v", foundName, foundEmail, foundSkills)
	}
}

// TestAT001_SchemaEquivalence verifies that PDF, DOCX, and manual entry map to the identical ProfileFactItem schema
func TestAT001_SchemaEquivalence(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryCareerRepository()
	svc := NewCareerService(repo)
	userID := "user_at001"

	// 1. Process PDF
	pdfBytes := createMockPDF("Charlie Doe", "charlie@example.com", "Go, Python")
	pdfDraft, err := svc.ProcessUploadedDocument(ctx, userID, "up_pdf", "charlie.pdf", "application/pdf", pdfBytes)
	if err != nil {
		t.Fatalf("PDF upload failed: %v", err)
	}
	confirmedPDFFacts, err := svc.ConfirmDraft(ctx, userID, pdfDraft.ID, ConfirmDraftInput{})
	if err != nil {
		t.Fatalf("Confirm PDF draft failed: %v", err)
	}

	// 2. Process DOCX
	docxBytes := createMockDOCX("Charlie Doe", "charlie@example.com", "Go, Python")
	docxDraft, err := svc.ProcessUploadedDocument(ctx, userID, "up_docx", "charlie.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
	if err != nil {
		t.Fatalf("DOCX upload failed: %v", err)
	}
	confirmedDOCXFacts, err := svc.ConfirmDraft(ctx, userID, docxDraft.ID, ConfirmDraftInput{})
	if err != nil {
		t.Fatalf("Confirm DOCX draft failed: %v", err)
	}

	// 3. Manual Entry
	manualFact, err := svc.CreateManualFact(ctx, userID, CreateManualFactInput{
		Category:    "contact",
		Title:       "Charlie Doe",
		Details:     []string{"Full Legal Name"},
		PrivacyTier: PrivacyPublic,
	})
	if err != nil {
		t.Fatalf("CreateManualFact failed: %v", err)
	}

	// Verify Schema Equivalence:
	// All confirmed facts share the same ProfileFactItem struct fields: Category, Title, Confirmed=true, PrivacyTier, CreatedAt
	if !manualFact.Confirmed {
		t.Errorf("manual fact must be confirmed=true")
	}
	for _, f := range confirmedPDFFacts {
		if !f.Confirmed {
			t.Errorf("confirmed PDF fact must have Confirmed=true")
		}
		if f.Category == "" || f.Title == "" {
			t.Errorf("confirmed PDF fact missing required category/title")
		}
	}
	for _, f := range confirmedDOCXFacts {
		if !f.Confirmed {
			t.Errorf("confirmed DOCX fact must have Confirmed=true")
		}
		if f.Category == "" || f.Title == "" {
			t.Errorf("confirmed DOCX fact missing required category/title")
		}
	}
}

// TestAT020_QuarantineRejection verifies that encrypted/script-injected uploads fail closed
func TestAT020_QuarantineRejection(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryCareerRepository()
	svc := NewCareerService(repo)

	// Encrypted PDF payload
	maliciousPDF := []byte("%PDF-1.4\n1 0 obj << /Encrypt 2 0 R >> endobj\n%%EOF")
	_, err := svc.ProcessUploadedDocument(ctx, "user_sec", "up_bad", "encrypted.pdf", "application/pdf", maliciousPDF)
	if err == nil {
		t.Fatalf("expected quarantine rejection for encrypted PDF, got nil")
	}
	if !strings.Contains(err.Error(), "quarantine") {
		t.Errorf("expected error message mentioning quarantine, got: %v", err)
	}

	// Script action PDF payload
	scriptPDF := []byte("%PDF-1.4\n1 0 obj << /JavaScript (app.alert('pwned')) >> endobj\n%%EOF")
	_, err = svc.ProcessUploadedDocument(ctx, "user_sec", "up_bad2", "script.pdf", "application/pdf", scriptPDF)
	if err == nil {
		t.Fatalf("expected quarantine rejection for JavaScript PDF, got nil")
	}
}

// TestAT003_ZeroFabrication verifies unstated facts (experience years, salary) are not invented
func TestAT003_ZeroFabrication(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryCareerRepository()
	svc := NewCareerService(repo)

	pdfBytes := createMockPDF("Zero Guess", "zero@example.com", "Go")
	draft, err := svc.ProcessUploadedDocument(ctx, "user_zero", "up_zero", "resume.pdf", "application/pdf", pdfBytes)
	if err != nil {
		t.Fatalf("failed to process document: %v", err)
	}

	// Ensure no salary or legal sponsorship facts are fabricated when absent from document
	for _, f := range draft.ExtractedFacts {
		if f.Category == "salary" {
			t.Errorf("salary fact must NOT be fabricated when unmentioned in source (AT-003)")
		}
		if f.Category == "legal_eligibility" || f.Category == "sponsorship" {
			t.Errorf("sponsorship fact must NOT be fabricated when unmentioned in source (AT-003)")
		}
	}
}

func TestCareerService_DraftConfirmationWithOverrides(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryCareerRepository()
	svc := NewCareerService(repo)
	userID := "user_override"

	pdfBytes := createMockPDF("Typo Naim", "typo@example.com", "Go")
	draft, err := svc.ProcessUploadedDocument(ctx, userID, "up_ovr", "resume.pdf", "application/pdf", pdfBytes)
	if err != nil {
		t.Fatalf("failed to process document: %v", err)
	}

	// Correct the candidate name during review
	overrides := []ProfileFactItem{
		{
			Category:         "contact",
			Title:            "Corrected Name",
			Details:          []string{"Full Legal Name"},
			Confidence:       1.0,
			SourceProvenance: ProvenanceUploadPDF,
			PrivacyTier:      PrivacyPublic,
		},
	}

	confirmedFacts, err := svc.ConfirmDraft(ctx, userID, draft.ID, ConfirmDraftInput{
		FactOverrides: overrides,
	})
	if err != nil {
		t.Fatalf("ConfirmDraft with overrides failed: %v", err)
	}

	if len(confirmedFacts) != 1 || confirmedFacts[0].Title != "Corrected Name" {
		t.Errorf("expected 1 confirmed fact with 'Corrected Name', got %+v", confirmedFacts)
	}

	// Confirming second time must fail closed (ErrAlreadyConfirmed)
	_, err = svc.ConfirmDraft(ctx, userID, draft.ID, ConfirmDraftInput{})
	if err != ErrAlreadyConfirmed {
		t.Errorf("expected ErrAlreadyConfirmed, got %v", err)
	}

	// Verify persistence in repository
	savedFacts, err := svc.GetUserFacts(ctx, userID, "")
	if err != nil {
		t.Fatalf("GetUserFacts failed: %v", err)
	}
	if len(savedFacts) != 1 || savedFacts[0].Title != "Corrected Name" {
		t.Errorf("expected saved fact 'Corrected Name', got %+v", savedFacts)
	}
}
