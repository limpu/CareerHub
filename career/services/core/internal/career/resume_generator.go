package career

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrUnsupportedResumeFormat = errors.New("unsupported resume format")
	ErrEmptyMasterProfile      = errors.New("cannot generate resume from empty or unconfirmed master profile")
	ErrParseBackVerification   = errors.New("parse-back verification failed: generated resume is unreadable or dropped critical facts")
)

type MasterResumeFormat string

const (
	ResumeFormatPDF  MasterResumeFormat = "pdf"
	ResumeFormatDOCX MasterResumeFormat = "docx"
	ResumeFormatTXT  MasterResumeFormat = "txt"
)

type ResumeTemplateType string

const (
	TemplateSingleColumnModern  ResumeTemplateType = "single_column_modern"
	TemplateSingleColumnClassic ResumeTemplateType = "single_column_classic"
	TemplateSingleColumnMinimal ResumeTemplateType = "single_column_minimal"
)

// Standard deterministic section order for ATS-friendly resumes (CAR-04, AT-002)
var DeterministicSectionOrder = []string{
	"CONTACT",
	"SUMMARY",
	"EXPERIENCE",
	"EDUCATION",
	"SKILLS",
	"PROJECTS",
	"CERTIFICATES",
	"LANGUAGES",
	"LINKS",
}

type GeneratedResume struct {
	ID                        string             `json:"id"`
	UserID                    string             `json:"user_id"`
	Format                    MasterResumeFormat `json:"format"`
	Template                  ResumeTemplateType `json:"template"`
	FileName                  string             `json:"file_name"`
	MimeType                  string             `json:"mime_type"`
	ByteContent               []byte             `json:"-"` // Not serialized in direct JSON summary
	ContentLength             int                `json:"content_length"`
	ChecksumSHA256            string             `json:"checksum_sha256"`
	DeterministicSectionOrder []string           `json:"deterministic_section_order"`
	ParseBackVerified         bool               `json:"parse_back_verified"`
	CreatedAt                 time.Time          `json:"created_at"`
}

type ParseBackVerificationResult struct {
	Success           bool      `json:"success"`
	Format            string    `json:"format"`
	ExtractedWords    int       `json:"extracted_words"`
	MatchedFactsCount int       `json:"matched_facts_count"`
	MissingFacts      []string  `json:"missing_facts,omitempty"`
	Warnings          []string  `json:"warnings,omitempty"`
	VerifiedAt        time.Time `json:"verified_at"`
}

// ResumeGenerator handles deterministic resume generation across PDF, DOCX, and TXT.
type ResumeGenerator struct {
	extractor DocumentExtractor
}

func NewResumeGenerator(extractor DocumentExtractor) *ResumeGenerator {
	if extractor == nil {
		extractor = NewNativeDocumentExtractor()
	}
	return &ResumeGenerator{
		extractor: extractor,
	}
}

// GenerateMasterResume produces an ATS-friendly single-column resume and validates it via parse-back.
func (rg *ResumeGenerator) GenerateMasterResume(profile *MasterCareerProfile, format MasterResumeFormat, template ResumeTemplateType) (*GeneratedResume, *ParseBackVerificationResult, error) {
	if profile == nil || strings.TrimSpace(profile.Contact.FullName) == "" {
		return nil, nil, ErrEmptyMasterProfile
	}

	if template == "" {
		template = TemplateSingleColumnModern
	}

	var (
		data     []byte
		mimeType string
		ext      string
		err      error
	)

	switch format {
	case ResumeFormatPDF:
		mimeType = "application/pdf"
		ext = "pdf"
		data, err = rg.generatePDF(profile, template)
	case ResumeFormatDOCX:
		mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		ext = "docx"
		data, err = rg.generateDOCX(profile, template)
	case ResumeFormatTXT:
		mimeType = "text/plain"
		ext = "txt"
		data, err = rg.generateTXT(profile, template)
	default:
		return nil, nil, fmt.Errorf("%w: %s", ErrUnsupportedResumeFormat, format)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate resume in %s format: %w", format, err)
	}

	// Calculate SHA256 checksum
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	// Sanitize name for file name
	safeName := strings.ToLower(strings.ReplaceAll(profile.Contact.FullName, " ", "_"))
	fileName := fmt.Sprintf("%s_resume_%s.%s", safeName, template, ext)

	resumeID := fmt.Sprintf("res_%d", time.Now().UnixNano())

	res := &GeneratedResume{
		ID:                        resumeID,
		UserID:                    profile.UserID,
		Format:                    format,
		Template:                  template,
		FileName:                  fileName,
		MimeType:                  mimeType,
		ByteContent:               data,
		ContentLength:             len(data),
		ChecksumSHA256:            checksum,
		DeterministicSectionOrder: DeterministicSectionOrder,
		ParseBackVerified:         false,
		CreatedAt:                 time.Now().UTC(),
	}

	// AT-002: Independent parse-back verification check
	vResult, err := rg.VerifyParseBack(res, profile)
	if err != nil {
		return res, vResult, fmt.Errorf("%w: %v", ErrParseBackVerification, err)
	}

	res.ParseBackVerified = vResult.Success
	return res, vResult, nil
}

// generateTXT produces a clean, single-column plain text representation.
func (rg *ResumeGenerator) generateTXT(p *MasterCareerProfile, template ResumeTemplateType) ([]byte, error) {
	var b strings.Builder

	// 1. CONTACT
	b.WriteString(strings.ToUpper(p.Contact.FullName) + "\n")
	if p.Contact.Headline != "" {
		b.WriteString(p.Contact.Headline + "\n")
	}
	var contactParts []string
	if p.Contact.Email != "" {
		contactParts = append(contactParts, p.Contact.Email)
	}
	if p.Contact.Phone != "" {
		contactParts = append(contactParts, p.Contact.Phone)
	}
	if p.Contact.Location != "" {
		contactParts = append(contactParts, p.Contact.Location)
	}
	if len(contactParts) > 0 {
		b.WriteString(strings.Join(contactParts, " | ") + "\n")
	}
	b.WriteString("\n" + strings.Repeat("-", 72) + "\n\n")

	// 2. SUMMARY
	if strings.TrimSpace(p.Contact.Summary) != "" {
		b.WriteString("PROFESSIONAL SUMMARY\n")
		b.WriteString(strings.Repeat("-", 20) + "\n")
		b.WriteString(p.Contact.Summary + "\n\n")
	}

	// 3. EXPERIENCE (Confirmed only - AT-003)
	var confirmedExp []ExperienceItem
	for _, exp := range p.Experiences {
		if exp.Confirmed {
			confirmedExp = append(confirmedExp, exp)
		}
	}
	if len(confirmedExp) > 0 {
		b.WriteString("WORK EXPERIENCE\n")
		b.WriteString(strings.Repeat("-", 15) + "\n")
		for _, exp := range confirmedExp {
			dateStr := formatRange(exp.StartDate, exp.EndDate, exp.IsCurrent)
			b.WriteString(fmt.Sprintf("%s at %s (%s)\n", exp.Title, exp.Company, dateStr))
			if exp.Location != "" {
				b.WriteString(fmt.Sprintf("Location: %s\n", exp.Location))
			}
			if exp.Description != "" {
				b.WriteString(exp.Description + "\n")
			}
			for _, h := range exp.Highlights {
				b.WriteString(fmt.Sprintf("  * %s\n", h))
			}
			if len(exp.SkillsUsed) > 0 {
				b.WriteString(fmt.Sprintf("  Skills: %s\n", strings.Join(exp.SkillsUsed, ", ")))
			}
			b.WriteString("\n")
		}
	}

	// 4. EDUCATION (Confirmed only)
	var confirmedEdu []EducationItem
	for _, edu := range p.Education {
		if edu.Confirmed {
			confirmedEdu = append(confirmedEdu, edu)
		}
	}
	if len(confirmedEdu) > 0 {
		b.WriteString("EDUCATION\n")
		b.WriteString(strings.Repeat("-", 9) + "\n")
		for _, edu := range confirmedEdu {
			dateStr := formatRange(edu.StartDate, edu.EndDate, false)
			b.WriteString(fmt.Sprintf("%s, %s - %s (%s)\n", edu.Degree, edu.FieldOfStudy, edu.Institution, dateStr))
			if edu.Grade != "" {
				b.WriteString(fmt.Sprintf("  Grade/GPA: %s\n", edu.Grade))
			}
			for _, h := range edu.Highlights {
				b.WriteString(fmt.Sprintf("  * %s\n", h))
			}
			b.WriteString("\n")
		}
	}

	// 5. SKILLS (Confirmed only)
	var confirmedSkills []SkillItem
	for _, s := range p.Skills {
		if s.Confirmed {
			confirmedSkills = append(confirmedSkills, s)
		}
	}
	if len(confirmedSkills) > 0 {
		b.WriteString("SKILLS\n")
		b.WriteString(strings.Repeat("-", 6) + "\n")
		byCat := make(map[string][]string)
		for _, s := range confirmedSkills {
			cat := strings.ToUpper(s.Category)
			if cat == "" {
				cat = "TECHNICAL"
			}
			byCat[cat] = append(byCat[cat], s.Name)
		}
		for cat, names := range byCat {
			b.WriteString(fmt.Sprintf("%s: %s\n", cat, strings.Join(names, ", ")))
		}
		b.WriteString("\n")
	}

	// 6. PROJECTS (Confirmed only)
	var confirmedProj []ProjectItem
	for _, prj := range p.Projects {
		if prj.Confirmed {
			confirmedProj = append(confirmedProj, prj)
		}
	}
	if len(confirmedProj) > 0 {
		b.WriteString("KEY PROJECTS\n")
		b.WriteString(strings.Repeat("-", 12) + "\n")
		for _, prj := range confirmedProj {
			b.WriteString(fmt.Sprintf("%s (%s)\n", prj.Title, prj.Role))
			if prj.URL != "" {
				b.WriteString(fmt.Sprintf("URL: %s\n", prj.URL))
			}
			if prj.Description != "" {
				b.WriteString(prj.Description + "\n")
			}
			for _, h := range prj.Highlights {
				b.WriteString(fmt.Sprintf("  * %s\n", h))
			}
			if len(prj.Technologies) > 0 {
				b.WriteString(fmt.Sprintf("  Technologies: %s\n", strings.Join(prj.Technologies, ", ")))
			}
			b.WriteString("\n")
		}
	}

	// 7. CERTIFICATES (Confirmed only)
	var confirmedCerts []CertificateItem
	for _, c := range p.Certificates {
		if c.Confirmed {
			confirmedCerts = append(confirmedCerts, c)
		}
	}
	if len(confirmedCerts) > 0 {
		b.WriteString("CERTIFICATIONS\n")
		b.WriteString(strings.Repeat("-", 14) + "\n")
		for _, c := range confirmedCerts {
			b.WriteString(fmt.Sprintf("%s - %s (%s)\n", c.Name, c.Issuer, formatDate(c.IssueDate)))
		}
		b.WriteString("\n")
	}

	// 8. LANGUAGES (Confirmed only)
	var confirmedLangs []LanguageItem
	for _, l := range p.Languages {
		if l.Confirmed {
			confirmedLangs = append(confirmedLangs, l)
		}
	}
	if len(confirmedLangs) > 0 {
		b.WriteString("LANGUAGES\n")
		b.WriteString(strings.Repeat("-", 9) + "\n")
		var langParts []string
		for _, l := range confirmedLangs {
			langParts = append(langParts, fmt.Sprintf("%s (%s)", l.Language, l.Proficiency))
		}
		b.WriteString(strings.Join(langParts, ", ") + "\n\n")
	}

	// 9. LINKS
	if len(p.Links) > 0 {
		b.WriteString("LINKS & PORTFOLIO\n")
		b.WriteString(strings.Repeat("-", 17) + "\n")
		for _, lnk := range p.Links {
			b.WriteString(fmt.Sprintf("%s: %s\n", lnk.Label, lnk.URL))
		}
		b.WriteString("\n")
	}

	return []byte(b.String()), nil
}

// generatePDF constructs a pure-Go, ATS-compatible single-column PDF with selectable text.
func (rg *ResumeGenerator) generatePDF(p *MasterCareerProfile, template ResumeTemplateType) ([]byte, error) {
	var lines []string

	// Title / Contact
	lines = append(lines, strings.ToUpper(p.Contact.FullName))
	if p.Contact.Headline != "" {
		lines = append(lines, p.Contact.Headline)
	}
	var contactParts []string
	if p.Contact.Email != "" {
		contactParts = append(contactParts, p.Contact.Email)
	}
	if p.Contact.Phone != "" {
		contactParts = append(contactParts, p.Contact.Phone)
	}
	if p.Contact.Location != "" {
		contactParts = append(contactParts, p.Contact.Location)
	}
	if len(contactParts) > 0 {
		lines = append(lines, strings.Join(contactParts, " | "))
	}
	lines = append(lines, "")

	// Summary
	if strings.TrimSpace(p.Contact.Summary) != "" {
		lines = append(lines, "PROFESSIONAL SUMMARY")
		lines = append(lines, p.Contact.Summary)
		lines = append(lines, "")
	}

	// Experience
	var confirmedExp []ExperienceItem
	for _, exp := range p.Experiences {
		if exp.Confirmed {
			confirmedExp = append(confirmedExp, exp)
		}
	}
	if len(confirmedExp) > 0 {
		lines = append(lines, "WORK EXPERIENCE")
		for _, exp := range confirmedExp {
			dateStr := formatRange(exp.StartDate, exp.EndDate, exp.IsCurrent)
			lines = append(lines, fmt.Sprintf("%s at %s (%s)", exp.Title, exp.Company, dateStr))
			if exp.Description != "" {
				lines = append(lines, exp.Description)
			}
			for _, h := range exp.Highlights {
				lines = append(lines, "- "+h)
			}
			if len(exp.SkillsUsed) > 0 {
				lines = append(lines, "Skills: "+strings.Join(exp.SkillsUsed, ", "))
			}
			lines = append(lines, "")
		}
	}

	// Education
	var confirmedEdu []EducationItem
	for _, edu := range p.Education {
		if edu.Confirmed {
			confirmedEdu = append(confirmedEdu, edu)
		}
	}
	if len(confirmedEdu) > 0 {
		lines = append(lines, "EDUCATION")
		for _, edu := range confirmedEdu {
			dateStr := formatRange(edu.StartDate, edu.EndDate, false)
			lines = append(lines, fmt.Sprintf("%s, %s - %s (%s)", edu.Degree, edu.FieldOfStudy, edu.Institution, dateStr))
			for _, h := range edu.Highlights {
				lines = append(lines, "- "+h)
			}
			lines = append(lines, "")
		}
	}

	// Skills
	var confirmedSkills []SkillItem
	for _, s := range p.Skills {
		if s.Confirmed {
			confirmedSkills = append(confirmedSkills, s)
		}
	}
	if len(confirmedSkills) > 0 {
		lines = append(lines, "SKILLS")
		var sNames []string
		for _, s := range confirmedSkills {
			sNames = append(sNames, s.Name)
		}
		lines = append(lines, strings.Join(sNames, ", "))
		lines = append(lines, "")
	}

	// Projects
	var confirmedProj []ProjectItem
	for _, prj := range p.Projects {
		if prj.Confirmed {
			confirmedProj = append(confirmedProj, prj)
		}
	}
	if len(confirmedProj) > 0 {
		lines = append(lines, "KEY PROJECTS")
		for _, prj := range confirmedProj {
			lines = append(lines, fmt.Sprintf("%s (%s)", prj.Title, prj.Role))
			if prj.URL != "" {
				lines = append(lines, "URL: "+prj.URL)
			}
			if prj.Description != "" {
				lines = append(lines, prj.Description)
			}
			for _, h := range prj.Highlights {
				lines = append(lines, "- "+h)
			}
			lines = append(lines, "")
		}
	}

	// Certificates
	var confirmedCerts []CertificateItem
	for _, c := range p.Certificates {
		if c.Confirmed {
			confirmedCerts = append(confirmedCerts, c)
		}
	}
	if len(confirmedCerts) > 0 {
		lines = append(lines, "CERTIFICATIONS")
		for _, c := range confirmedCerts {
			lines = append(lines, fmt.Sprintf("%s - %s (%s)", c.Name, c.Issuer, formatDate(c.IssueDate)))
		}
		lines = append(lines, "")
	}

	// Languages
	var confirmedLangs []LanguageItem
	for _, l := range p.Languages {
		if l.Confirmed {
			confirmedLangs = append(confirmedLangs, l)
		}
	}
	if len(confirmedLangs) > 0 {
		lines = append(lines, "LANGUAGES")
		var langParts []string
		for _, l := range confirmedLangs {
			langParts = append(langParts, fmt.Sprintf("%s (%s)", l.Language, l.Proficiency))
		}
		lines = append(lines, strings.Join(langParts, ", "))
		lines = append(lines, "")
	}

	// Links
	if len(p.Links) > 0 {
		lines = append(lines, "LINKS")
		for _, lnk := range p.Links {
			lines = append(lines, fmt.Sprintf("%s: %s", lnk.Label, lnk.URL))
		}
		lines = append(lines, "")
	}

	// Build valid PDF 1.4 binary structure
	return buildSimplePDF(lines)
}

// buildSimplePDF outputs a standard single-page or multi-section PDF 1.4 file
// with font dictionary and BT ... ET selectable text operators.
func buildSimplePDF(lines []string) ([]byte, error) {
	var streamBuf strings.Builder

	// Text matrix setup
	streamBuf.WriteString("BT\n")
	streamBuf.WriteString("/F1 10 Tf\n")
	streamBuf.WriteString("14 TL\n") // 14pt line leading
	streamBuf.WriteString("50 750 Td\n")

	for _, line := range lines {
		escaped := escapePDFString(line)
		if escaped == "" {
			streamBuf.WriteString("T*\n") // New line
		} else {
			// Write selectable text chunk with Tj and line jump
			streamBuf.WriteString(fmt.Sprintf("(%s) Tj T*\n", escaped))
		}
	}
	streamBuf.WriteString("ET\n")

	contentStream := streamBuf.String()
	contentLen := len(contentStream)

	var doc bytes.Buffer
	doc.WriteString("%PDF-1.4\n")
	doc.WriteString("%\xE2\xE3\xCF\xD3\n") // Binary marker

	var offsets []int

	// Obj 1: Catalog
	offsets = append(offsets, doc.Len())
	doc.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Obj 2: Pages
	offsets = append(offsets, doc.Len())
	doc.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	// Obj 3: Page (Letter: 612 x 792 pt)
	offsets = append(offsets, doc.Len())
	doc.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n")

	// Obj 4: Content Stream
	offsets = append(offsets, doc.Len())
	doc.WriteString(fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", contentLen, contentStream))

	// Obj 5: Font (Helvetica standard type 1)
	offsets = append(offsets, doc.Len())
	doc.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	// Xref table
	xrefOffset := doc.Len()
	doc.WriteString("xref\n0 6\n0000000000 65535 f \n")
	for _, o := range offsets {
		doc.WriteString(fmt.Sprintf("%010d 00000 n \n", o))
	}

	// Trailer
	doc.WriteString(fmt.Sprintf("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefOffset))

	return doc.Bytes(), nil
}

func escapePDFString(raw string) string {
	raw = strings.ReplaceAll(raw, `\`, `\\`)
	raw = strings.ReplaceAll(raw, `(`, `\(`)
	raw = strings.ReplaceAll(raw, `)`, `\)`)
	// replace non-ascii or tabs with space
	raw = strings.ReplaceAll(raw, "\t", "    ")
	return raw
}

// generateDOCX constructs a pure-Go, ATS-compatible single-column Word document (.docx).
func (rg *ResumeGenerator) generateDOCX(p *MasterCareerProfile, template ResumeTemplateType) ([]byte, error) {
	txtBytes, err := rg.generateTXT(p, template)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(txtBytes), "\n")

	var docXML strings.Builder
	docXML.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	docXML.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` + "\n")
	docXML.WriteString(`  <w:body>` + "\n")

	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		clean := escapeXML(trimmed)
		if strings.HasPrefix(trimmed, "---") {
			// Separator
			docXML.WriteString(`    <w:p><w:pPr><w:pBdr><w:bottom w:val="single" w:sz="6" w:space="1" w:color="CCCCCC"/></w:pBdr></w:pPr></w:p>` + "\n")
		} else if strings.ToUpper(trimmed) == trimmed && len(trimmed) > 3 && !strings.Contains(trimmed, "@") && !strings.Contains(trimmed, "|") {
			// Header
			docXML.WriteString(fmt.Sprintf(`    <w:p><w:pPr><w:rPr><w:b/><w:sz w:val="24"/></w:rPr></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="24"/></w:rPr><w:t>%s</w:t></w:r></w:p>`+"\n", clean))
		} else {
			// Regular text
			docXML.WriteString(fmt.Sprintf(`    <w:p><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`+"\n", clean))
		}
	}

	// Standard margins
	docXML.WriteString(`    <w:sectPr><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440"/></w:sectPr>` + "\n")
	docXML.WriteString(`  </w:body>` + "\n")
	docXML.WriteString(`</w:document>`)

	// Package into zip
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)

	// 1. [Content_Types].xml
	ctW, err := zw.Create("[Content_Types].xml")
	if err != nil {
		return nil, err
	}
	ctW.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`))

	// 2. _rels/.rels
	relW, err := zw.Create("_rels/.rels")
	if err != nil {
		return nil, err
	}
	relW.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`))

	// 3. word/document.xml
	docW, err := zw.Create("word/document.xml")
	if err != nil {
		return nil, err
	}
	docW.Write([]byte(docXML.String()))

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return zipBuf.Bytes(), nil
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, `'`, "&apos;")
	return s
}

func formatDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("Jan 2006")
}

func formatRange(start, end *time.Time, isCurrent bool) string {
	s := formatDate(start)
	if isCurrent {
		if s != "" {
			return s + " - Present"
		}
		return "Present"
	}
	e := formatDate(end)
	if s != "" && e != "" {
		return s + " - " + e
	}
	if s != "" {
		return s
	}
	return e
}

// VerifyParseBack executes an independent parse-back check (AT-002) using our DocumentExtractor
// to ensure the generated document contains confirmed contact facts and experience history.
func (rg *ResumeGenerator) VerifyParseBack(resume *GeneratedResume, originalProfile *MasterCareerProfile) (*ParseBackVerificationResult, error) {
	if resume == nil || len(resume.ByteContent) == 0 {
		return nil, errors.New("cannot verify empty resume byte content")
	}

	extractedText, err := rg.extractor.ExtractText(resume.ByteContent, resume.MimeType)
	if err != nil {
		return &ParseBackVerificationResult{
			Success:  false,
			Format:   string(resume.Format),
			Warnings: []string{fmt.Sprintf("Extraction failed: %v", err)},
		}, fmt.Errorf("extractor failed on generated resume: %w", err)
	}

	words := strings.Fields(extractedText)
	wordCount := len(words)
	lowerText := strings.ToLower(extractedText)

	var (
		matchedCount int
		missingFacts []string
		warnings     []string
	)

	// Check 1: Candidate Full Name must be present
	if strings.TrimSpace(originalProfile.Contact.FullName) != "" {
		nameTokens := strings.Fields(strings.ToLower(originalProfile.Contact.FullName))
		allNameTokensFound := true
		for _, tok := range nameTokens {
			if !strings.Contains(lowerText, tok) {
				allNameTokensFound = false
				break
			}
		}
		if allNameTokensFound {
			matchedCount++
		} else {
			missingFacts = append(missingFacts, fmt.Sprintf("Candidate Name: %s", originalProfile.Contact.FullName))
		}
	}

	// Check 2: Candidate Email must be present if configured
	if strings.TrimSpace(originalProfile.Contact.Email) != "" {
		emailLower := strings.ToLower(strings.TrimSpace(originalProfile.Contact.Email))
		if strings.Contains(lowerText, emailLower) {
			matchedCount++
		} else {
			missingFacts = append(missingFacts, fmt.Sprintf("Email: %s", originalProfile.Contact.Email))
		}
	}

	// Check 3: Confirmed Experience Companies and Titles
	for _, exp := range originalProfile.Experiences {
		if !exp.Confirmed {
			continue // Skip unconfirmed facts
		}
		companyLower := strings.ToLower(strings.TrimSpace(exp.Company))
		if companyLower != "" {
			if strings.Contains(lowerText, companyLower) {
				matchedCount++
			} else {
				missingFacts = append(missingFacts, fmt.Sprintf("Experience Company: %s", exp.Company))
			}
		}

		titleLower := strings.ToLower(strings.TrimSpace(exp.Title))
		if titleLower != "" {
			if strings.Contains(lowerText, titleLower) {
				matchedCount++
			} else {
				missingFacts = append(missingFacts, fmt.Sprintf("Experience Title: %s", exp.Title))
			}
		}
	}

	// Check 4: Confirmed Skills
	for _, skill := range originalProfile.Skills {
		if !skill.Confirmed {
			continue
		}
		skillLower := strings.ToLower(strings.TrimSpace(skill.Name))
		if skillLower != "" {
			if strings.Contains(lowerText, skillLower) {
				matchedCount++
			} else {
				warnings = append(warnings, fmt.Sprintf("Skill not parsed verbatim: %s", skill.Name))
			}
		}
	}

	success := len(missingFacts) == 0 && wordCount > 20

	return &ParseBackVerificationResult{
		Success:           success,
		Format:            string(resume.Format),
		ExtractedWords:    wordCount,
		MatchedFactsCount: matchedCount,
		MissingFacts:      missingFacts,
		Warnings:          warnings,
		VerifiedAt:        time.Now().UTC(),
	}, nil
}
