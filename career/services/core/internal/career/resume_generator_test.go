package career

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
	"time"
)

func parseTime(s string) *time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return &t
}

func createTestMasterProfile() *MasterCareerProfile {
	return &MasterCareerProfile{
		ID:     "prof-test-001",
		UserID: "user-test-42",
		Contact: ContactInfo{
			FullName: "Jordan Lee",
			Email:    "jordan.lee@example.com",
			Phone:    "+1 (555) 019-2834",
			Location: "Seattle, WA",
			Headline: "Senior Staff Distributed Systems Engineer",
			Summary:  "10+ years engineering high-scale distributed consensus engines and resilient microservices in Go.",
		},
		Experiences: []ExperienceItem{
			{
				ID:          "exp-1",
				Title:       "Principal Backend Architect",
				Company:     "CloudScale Solutions",
				Location:    "Remote",
				StartDate:   parseTime("2021-01-01"),
				IsCurrent:   true,
				Description: "Lead architectural evolution for multi-tenant data pipelines.",
				Highlights: []string{
					"Scaled Kafka & Redis stream ingest to 500k events/sec with zero packet drop",
					"Championed deterministic idempotency and transactional outbox across 14 microservices",
				},
				SkillsUsed: []string{"Go", "Kafka", "PostgreSQL", "Redis"},
				Confirmed:  true, // Confirmed fact (AT-003)
			},
			{
				ID:          "exp-unconfirmed",
				Title:       "Hallucinated CTO",
				Company:     "Phantom Corp",
				Location:    "Nowhere",
				StartDate:   parseTime("2019-01-01"),
				EndDate:     parseTime("2020-01-01"),
				Description: "Unconfirmed fact that must never appear in generated resume",
				Confirmed:  false, // UNCONFIRMED - MUST BE OMITTED (AT-003)
			},
		},
		Education: []EducationItem{
			{
				ID:           "edu-1",
				Degree:       "Bachelor of Science",
				FieldOfStudy: "Computer Science & Engineering",
				Institution:  "Bangladesh University of Engineering and Technology",
				StartDate:    parseTime("2012-01-01"),
				EndDate:      parseTime("2016-01-01"),
				Grade:        "3.92 / 4.00",
				Confirmed:    true,
			},
		},
		Skills: []SkillItem{
			{ID: "sk-1", Name: "Go", Category: "Languages", Proficiency: "expert", Confirmed: true},
			{ID: "sk-2", Name: "Distributed Systems", Category: "Architecture", Proficiency: "expert", Confirmed: true},
			{ID: "sk-3", Name: "PostgreSQL", Category: "Databases", Proficiency: "advanced", Confirmed: true},
			{ID: "sk-unconfirmed", Name: "Quantum Teleportation", Category: "Fake", Confirmed: false}, // OMITTED
		},
		Projects: []ProjectItem{
			{
				ID:           "prj-1",
				Title:        "Raft Consensus KV-Store",
				Role:         "Creator & Maintainer",
				URL:          "https://github.com/farhan/raft-kv",
				Description:  "High-availability replicated key-value storage engine.",
				Highlights:   []string{"Implemented log compaction and fast election benchmarks"},
				Technologies: []string{"Go", "gRPC"},
				Confirmed:    true,
			},
		},
		Certificates: []CertificateItem{
			{
				ID:        "crt-1",
				Name:      "AWS Certified Solutions Architect",
				Issuer:    "Amazon Web Services",
				IssueDate: parseTime("2023-05-15"),
				Confirmed: true,
			},
		},
		Languages: []LanguageItem{
			{ID: "lng-1", Language: "English", Proficiency: "fluent", Confirmed: true},
			{ID: "lng-2", Language: "Bengali", Proficiency: "native", Confirmed: true},
		},
		Links: []ProfileLink{
			{Label: "GitHub", URL: "https://github.com/farhan"},
			{Label: "LinkedIn", URL: "https://linkedin.com/in/farhan"},
		},
	}
}

func TestResumeGenerator_TXT_FormatAndDeterministicOrder(t *testing.T) {
	p := createTestMasterProfile()
	gen := NewResumeGenerator(NewNativeDocumentExtractor())

	res, vResult, err := gen.GenerateMasterResume(p, ResumeFormatTXT, TemplateSingleColumnModern)
	if err != nil {
		t.Fatalf("GenerateMasterResume(TXT) failed: %v", err)
	}

	if res.Format != ResumeFormatTXT {
		t.Errorf("expected format %s, got %s", ResumeFormatTXT, res.Format)
	}
	if !strings.HasSuffix(res.FileName, ".txt") {
		t.Errorf("expected .txt file extension, got %s", res.FileName)
	}

	content := string(res.ByteContent)

	// Check deterministic headers
	expectedSections := []string{
		"FARHAN AHMED\n",
		"PROFESSIONAL SUMMARY\n",
		"WORK EXPERIENCE\n",
		"EDUCATION\n",
		"SKILLS\n",
		"KEY PROJECTS\n",
		"CERTIFICATIONS\n",
		"LANGUAGES\n---------",
		"LINKS & PORTFOLIO\n",
	}

	lastIdx := -1
	for _, sec := range expectedSections {
		idx := strings.Index(content, sec)
		if idx == -1 {
			t.Errorf("missing expected section in TXT resume: %s", sec)
		} else if idx < lastIdx {
			t.Errorf("section %s appeared out of deterministic order (idx: %d, lastIdx: %d)", sec, idx, lastIdx)
		}
		lastIdx = idx
	}

	// AT-003: Unconfirmed facts must NOT be present
	if strings.Contains(content, "Phantom Corp") || strings.Contains(content, "Hallucinated CTO") {
		t.Errorf("AT-003 violation: unconfirmed experience appeared in generated resume")
	}
	if strings.Contains(content, "Quantum Teleportation") {
		t.Errorf("AT-003 violation: unconfirmed skill appeared in generated resume")
	}

	// Parse-back verification check
	if !vResult.Success {
		t.Errorf("expected parse-back verification to succeed on TXT, missing facts: %v", vResult.MissingFacts)
	}
	if !res.ParseBackVerified {
		t.Errorf("expected res.ParseBackVerified to be true")
	}
}

func TestResumeGenerator_PDF_SelectableText(t *testing.T) {
	p := createTestMasterProfile()
	gen := NewResumeGenerator(NewNativeDocumentExtractor())

	res, vResult, err := gen.GenerateMasterResume(p, ResumeFormatPDF, TemplateSingleColumnClassic)
	if err != nil {
		t.Fatalf("GenerateMasterResume(PDF) failed: %v", err)
	}

	if res.Format != ResumeFormatPDF {
		t.Errorf("expected format %s, got %s", ResumeFormatPDF, res.Format)
	}
	if !bytes.HasPrefix(res.ByteContent, []byte("%PDF-1.4")) {
		t.Errorf("generated PDF does not have standard %%PDF-1.4 header")
	}
	if !strings.HasSuffix(res.FileName, ".pdf") {
		t.Errorf("expected .pdf file extension, got %s", res.FileName)
	}

	// Verify selectable text BT ... ET operators exist
	if !bytes.Contains(res.ByteContent, []byte("BT")) || !bytes.Contains(res.ByteContent, []byte("ET")) {
		t.Errorf("PDF does not contain BT ... ET selectable text operators")
	}

	// AT-002 Parse-Back Verification
	if !vResult.Success {
		t.Fatalf("AT-002 failed: PDF parse-back verification was not successful: %v", vResult.MissingFacts)
	}
	if vResult.MatchedFactsCount < 3 {
		t.Errorf("expected at least 3 matched facts from PDF parse-back, got %d", vResult.MatchedFactsCount)
	}
}

func TestResumeGenerator_DOCX_OpenXML(t *testing.T) {
	p := createTestMasterProfile()
	gen := NewResumeGenerator(NewNativeDocumentExtractor())

	res, vResult, err := gen.GenerateMasterResume(p, ResumeFormatDOCX, TemplateSingleColumnMinimal)
	if err != nil {
		t.Fatalf("GenerateMasterResume(DOCX) failed: %v", err)
	}

	if res.Format != ResumeFormatDOCX {
		t.Errorf("expected format %s, got %s", ResumeFormatDOCX, res.Format)
	}
	if !strings.HasSuffix(res.FileName, ".docx") {
		t.Errorf("expected .docx file extension, got %s", res.FileName)
	}

	// Verify OpenXML zip archive structure
	r, err := zip.NewReader(bytes.NewReader(res.ByteContent), int64(len(res.ByteContent)))
	if err != nil {
		t.Fatalf("failed to open generated DOCX as zip archive: %v", err)
	}

	hasDocXML := false
	hasContentTypes := false
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			hasDocXML = true
		}
		if f.Name == "[Content_Types].xml" {
			hasContentTypes = true
		}
	}

	if !hasDocXML || !hasContentTypes {
		t.Errorf("generated DOCX missing essential OpenXML files: docXML=%v, contentTypes=%v", hasDocXML, hasContentTypes)
	}

	// AT-002 Parse-Back Verification
	if !vResult.Success {
		t.Fatalf("AT-002 failed: DOCX parse-back verification was not successful: %v", vResult.MissingFacts)
	}
	if vResult.ExtractedWords < 30 {
		t.Errorf("expected at least 30 extracted words from DOCX, got %d", vResult.ExtractedWords)
	}
}

func TestAT002_ParseBackVerification_AllFormats(t *testing.T) {
	p := createTestMasterProfile()
	gen := NewResumeGenerator(NewNativeDocumentExtractor())

	formats := []MasterResumeFormat{ResumeFormatTXT, ResumeFormatPDF, ResumeFormatDOCX}

	for _, fmt := range formats {
		t.Run(string(fmt), func(t *testing.T) {
			res, vResult, err := gen.GenerateMasterResume(p, fmt, TemplateSingleColumnModern)
			if err != nil {
				t.Fatalf("generation failed for format %s: %v", fmt, err)
			}
			if !vResult.Success {
				t.Fatalf("AT-002 assertion failed for %s: missing facts=%v, warnings=%v", fmt, vResult.MissingFacts, vResult.Warnings)
			}
			if len(vResult.MissingFacts) > 0 {
				t.Errorf("format %s dropped confirmed facts: %v", fmt, vResult.MissingFacts)
			}
			if !res.ParseBackVerified {
				t.Errorf("res.ParseBackVerified is false for format %s", fmt)
			}
		})
	}
}

func TestResumeGenerator_ZeroFabrication_ConfirmedFactsOnly(t *testing.T) {
	p := createTestMasterProfile()
	// Add another unconfirmed education and project
	p.Education = append(p.Education, EducationItem{
		ID:          "fake-edu",
		Institution: "Fabricated Ivy League",
		Degree:      "PhD",
		Confirmed:   false,
	})
	p.Projects = append(p.Projects, ProjectItem{
		ID:        "fake-prj",
		Title:     "Moon Landing Rover",
		Confirmed: false,
	})

	gen := NewResumeGenerator(NewNativeDocumentExtractor())
	res, _, err := gen.GenerateMasterResume(p, ResumeFormatTXT, TemplateSingleColumnModern)
	if err != nil {
		t.Fatalf("GenerateMasterResume failed: %v", err)
	}

	txt := string(res.ByteContent)
	if strings.Contains(txt, "Fabricated Ivy League") {
		t.Errorf("AT-003 violation: unconfirmed education found in resume")
	}
	if strings.Contains(txt, "Moon Landing Rover") {
		t.Errorf("AT-003 violation: unconfirmed project found in resume")
	}
}

func TestResumeGenerator_EmptyProfileRejection(t *testing.T) {
	gen := NewResumeGenerator(NewNativeDocumentExtractor())

	_, _, err := gen.GenerateMasterResume(nil, ResumeFormatPDF, TemplateSingleColumnModern)
	if err != ErrEmptyMasterProfile {
		t.Errorf("expected ErrEmptyMasterProfile for nil profile, got: %v", err)
	}

	emptyProfile := &MasterCareerProfile{}
	_, _, err = gen.GenerateMasterResume(emptyProfile, ResumeFormatPDF, TemplateSingleColumnModern)
	if err != ErrEmptyMasterProfile {
		t.Errorf("expected ErrEmptyMasterProfile for empty profile, got: %v", err)
	}
}
