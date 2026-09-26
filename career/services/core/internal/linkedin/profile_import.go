package linkedin

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

var (
	ErrEmptyArchive        = errors.New("empty linkedin archive or files")
	ErrEmptyText           = errors.New("empty linkedin profile text")
	ErrInvalidJSONPayload  = errors.New("invalid enterprise linkedin json payload")
	ErrProfileDataNotFound = errors.New("no parseable profile data found")
)

// IngestionSourceMode identifies how LinkedIn profile data was collected (LI-02, REQ-004, AT-010).
type IngestionSourceMode string

const (
	SourceArchiveZip    IngestionSourceMode = "archive_zip"
	SourceManualPaste   IngestionSourceMode = "manual_paste"
	SourceEnterpriseAPI IngestionSourceMode = "enterprise_api"
	SourceOIDCBasicOnly IngestionSourceMode = "oidc_basic_only"
)

// LinkedInImportedExperience represents a parsed employment or role entry.
type LinkedInImportedExperience struct {
	CompanyName string `json:"company_name"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
	IsCurrent   bool   `json:"is_current"`
}

// LinkedInImportedEducation represents a parsed academic qualification.
type LinkedInImportedEducation struct {
	SchoolName   string `json:"school_name"`
	DegreeName   string `json:"degree_name,omitempty"`
	FieldOfStudy string `json:"field_of_study,omitempty"`
	StartDate    string `json:"start_date,omitempty"`
	EndDate      string `json:"end_date,omitempty"`
	Notes        string `json:"notes,omitempty"`
}

// LinkedInImportedCertification represents a parsed license or certificate.
type LinkedInImportedCertification struct {
	Name          string `json:"name"`
	Authority     string `json:"authority,omitempty"`
	Url           string `json:"url,omitempty"`
	LicenseNumber string `json:"license_number,omitempty"`
	StartDate     string `json:"start_date,omitempty"`
	EndDate       string `json:"end_date,omitempty"`
}

// LinkedInImportedProfile captures unified LinkedIn profile information across all ingestion methods (AT-001, AT-010).
type LinkedInImportedProfile struct {
	ImportID                string                          `json:"import_id"`
	UserID                  string                          `json:"user_id"`
	SourceMode              IngestionSourceMode             `json:"source_mode"`
	DisplayName             string                          `json:"display_name"`
	Email                   string                          `json:"email,omitempty"`
	Headline                string                          `json:"headline,omitempty"`
	Summary                 string                          `json:"summary,omitempty"`
	Location                string                          `json:"location,omitempty"`
	Industry                string                          `json:"industry,omitempty"`
	Websites                []string                        `json:"websites,omitempty"`
	Experiences             []LinkedInImportedExperience    `json:"experiences"`
	Education               []LinkedInImportedEducation     `json:"education"`
	Skills                  []string                        `json:"skills"`
	Certifications          []LinkedInImportedCertification `json:"certifications,omitempty"`
	RequiresFallback        bool                            `json:"requires_fallback"`
	RecommendedFallback     string                          `json:"recommended_fallback,omitempty"`
	TruthInAdvertisingNote  string                          `json:"truth_in_advertising_note,omitempty"`
	ImportedAt              time.Time                       `json:"imported_at"`
	Status                  string                          `json:"status"` // "imported", "merged", "pending_review"
}

// ParseLinkedInArchiveFiles parses the map of extracted CSV files (Profile.csv, Positions.csv, etc.)
func ParseLinkedInArchiveFiles(files map[string]string) (*LinkedInImportedProfile, error) {
	if len(files) == 0 {
		return nil, ErrEmptyArchive
	}

	profile := &LinkedInImportedProfile{
		ImportID:       fmt.Sprintf("lip-imp-%d", time.Now().UnixNano()),
		SourceMode:     SourceArchiveZip,
		Experiences:    make([]LinkedInImportedExperience, 0),
		Education:      make([]LinkedInImportedEducation, 0),
		Skills:         make([]string, 0),
		Certifications: make([]LinkedInImportedCertification, 0),
		ImportedAt:     time.Now().UTC(),
		Status:         "imported",
	}

	for fileName, content := range files {
		cleanName := strings.TrimSpace(fileName)
		// Match case-insensitively for various export configurations
		lower := strings.ToLower(cleanName)
		switch {
		case strings.Contains(lower, "profile.csv"):
			parseProfileCSV(content, profile)
		case strings.Contains(lower, "positions.csv"):
			parsePositionsCSV(content, profile)
		case strings.Contains(lower, "education.csv"):
			parseEducationCSV(content, profile)
		case strings.Contains(lower, "skills.csv"):
			parseSkillsCSV(content, profile)
		case strings.Contains(lower, "certifications.csv"):
			parseCertificationsCSV(content, profile)
		}
	}

	return profile, nil
}

// ParseLinkedInZipArchive extracts and parses files directly from a zip byte slice.
func ParseLinkedInZipArchive(zipBytes []byte) (*LinkedInImportedProfile, error) {
	if len(zipBytes) == 0 {
		return nil, ErrEmptyArchive
	}

	r, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip archive: %w", err)
	}

	files := make(map[string]string)
	for _, f := range r.File {
		// Limit reading to reasonable size
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rc)
		_ = rc.Close()
		files[f.Name] = buf.String()
	}

	return ParseLinkedInArchiveFiles(files)
}

func parseProfileCSV(content string, p *LinkedInImportedProfile) {
	r := csv.NewReader(strings.NewReader(content))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	rows, err := r.ReadAll()
	if err != nil || len(rows) < 2 {
		return
	}

	header := rows[0]
	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[strings.ToLower(strings.TrimSpace(col))] = i
	}

	data := rows[1]
	getVal := func(name string) string {
		idx, ok := colIdx[strings.ToLower(name)]
		if ok && idx < len(data) {
			return strings.TrimSpace(data[idx])
		}
		return ""
	}

	firstName := getVal("first name")
	lastName := getVal("last name")
	if firstName != "" || lastName != "" {
		p.DisplayName = strings.TrimSpace(firstName + " " + lastName)
	}
	if p.Headline == "" {
		p.Headline = getVal("headline")
	}
	if p.Summary == "" {
		p.Summary = getVal("summary")
	}
	if p.Industry == "" {
		p.Industry = getVal("industry")
	}
	geo := getVal("geo location")
	if geo == "" {
		geo = getVal("address")
	}
	if p.Location == "" {
		p.Location = geo
	}
	websites := getVal("websites")
	if websites != "" {
		parts := strings.Split(websites, ",")
		for _, w := range parts {
			w = strings.TrimSpace(w)
			if w != "" {
				p.Websites = append(p.Websites, w)
			}
		}
	}
}

func parsePositionsCSV(content string, p *LinkedInImportedProfile) {
	r := csv.NewReader(strings.NewReader(content))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	rows, err := r.ReadAll()
	if err != nil || len(rows) < 2 {
		return
	}

	header := rows[0]
	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[strings.ToLower(strings.TrimSpace(col))] = i
	}

	for _, data := range rows[1:] {
		getVal := func(name string) string {
			idx, ok := colIdx[strings.ToLower(name)]
			if ok && idx < len(data) {
				return strings.TrimSpace(data[idx])
			}
			return ""
		}

		comp := getVal("company name")
		title := getVal("title")
		if comp == "" && title == "" {
			continue
		}

		startedOn := getVal("started on")
		finishedOn := getVal("finished on")
		isCurrent := finishedOn == ""

		p.Experiences = append(p.Experiences, LinkedInImportedExperience{
			CompanyName: comp,
			Title:       title,
			Description: getVal("description"),
			Location:    getVal("location"),
			StartDate:   startedOn,
			EndDate:     finishedOn,
			IsCurrent:   isCurrent,
		})
	}
}

func parseEducationCSV(content string, p *LinkedInImportedProfile) {
	r := csv.NewReader(strings.NewReader(content))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	rows, err := r.ReadAll()
	if err != nil || len(rows) < 2 {
		return
	}

	header := rows[0]
	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[strings.ToLower(strings.TrimSpace(col))] = i
	}

	for _, data := range rows[1:] {
		getVal := func(name string) string {
			idx, ok := colIdx[strings.ToLower(name)]
			if ok && idx < len(data) {
				return strings.TrimSpace(data[idx])
			}
			return ""
		}

		school := getVal("school name")
		if school == "" {
			continue
		}

		p.Education = append(p.Education, LinkedInImportedEducation{
			SchoolName: school,
			DegreeName: getVal("degree name"),
			StartDate:  getVal("start date"),
			EndDate:    getVal("end date"),
			Notes:      getVal("notes"),
		})
	}
}

func parseSkillsCSV(content string, p *LinkedInImportedProfile) {
	r := csv.NewReader(strings.NewReader(content))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	rows, err := r.ReadAll()
	if err != nil || len(rows) < 2 {
		return
	}

	header := rows[0]
	colIdx := -1
	for i, col := range header {
		if strings.ToLower(strings.TrimSpace(col)) == "name" {
			colIdx = i
			break
		}
	}
	if colIdx == -1 {
		colIdx = 0
	}

	for _, data := range rows[1:] {
		if colIdx < len(data) {
			s := strings.TrimSpace(data[colIdx])
			if s != "" {
				p.Skills = append(p.Skills, s)
			}
		}
	}
}

func parseCertificationsCSV(content string, p *LinkedInImportedProfile) {
	r := csv.NewReader(strings.NewReader(content))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	rows, err := r.ReadAll()
	if err != nil || len(rows) < 2 {
		return
	}

	header := rows[0]
	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[strings.ToLower(strings.TrimSpace(col))] = i
	}

	for _, data := range rows[1:] {
		getVal := func(name string) string {
			idx, ok := colIdx[strings.ToLower(name)]
			if ok && idx < len(data) {
				return strings.TrimSpace(data[idx])
			}
			return ""
		}

		name := getVal("name")
		if name == "" {
			continue
		}

		p.Certifications = append(p.Certifications, LinkedInImportedCertification{
			Name:          name,
			Authority:     getVal("authority"),
			Url:           getVal("url"),
			LicenseNumber: getVal("license number"),
			StartDate:     getVal("started on"),
			EndDate:       getVal("finished on"),
		})
	}
}

// ParseLinkedInPastedText extracts profile components from unformatted copied text.
func ParseLinkedInPastedText(rawText string) (*LinkedInImportedProfile, error) {
	trimmed := strings.TrimSpace(rawText)
	if trimmed == "" {
		return nil, ErrEmptyText
	}

	p := &LinkedInImportedProfile{
		ImportID:       fmt.Sprintf("lip-imp-%d", time.Now().UnixNano()),
		SourceMode:     SourceManualPaste,
		Experiences:    make([]LinkedInImportedExperience, 0),
		Education:      make([]LinkedInImportedEducation, 0),
		Skills:         make([]string, 0),
		Certifications: make([]LinkedInImportedCertification, 0),
		ImportedAt:     time.Now().UTC(),
		Status:         "imported",
	}

	lines := strings.Split(trimmed, "\n")
	var nonBlank []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t != "" {
			nonBlank = append(nonBlank, t)
		}
	}

	if len(nonBlank) == 0 {
		return nil, ErrEmptyText
	}

	// Identify primary sections
	type sectionBlock struct {
		name  string
		lines []string
	}

	var sections []sectionBlock
	currentSec := "header"
	secLines := []string{}

	isSectionHeader := func(l string) (string, bool) {
		low := strings.ToLower(strings.TrimSpace(l))
		switch low {
		case "about":
			return "about", true
		case "experience", "experiences":
			return "experience", true
		case "education":
			return "education", true
		case "skills":
			return "skills", true
		case "certifications", "licenses & certifications":
			return "certifications", true
		}
		return "", false
	}

	for _, line := range nonBlank {
		if secName, match := isSectionHeader(line); match {
			if len(secLines) > 0 {
				sections = append(sections, sectionBlock{name: currentSec, lines: secLines})
			}
			currentSec = secName
			secLines = []string{}
			continue
		}
		secLines = append(secLines, line)
	}
	if len(secLines) > 0 {
		sections = append(sections, sectionBlock{name: currentSec, lines: secLines})
	}

	for _, sec := range sections {
		switch sec.name {
		case "header":
			if len(sec.lines) > 0 {
				p.DisplayName = sec.lines[0]
			}
			if len(sec.lines) > 1 {
				p.Headline = sec.lines[1]
			}
			if len(sec.lines) > 2 {
				// Often contains location & connections: e.g. "Greater New York City Area • 500+ connections"
				locParts := strings.Split(sec.lines[2], "•")
				p.Location = strings.TrimSpace(locParts[0])
			}
		case "about":
			p.Summary = strings.Join(sec.lines, "\n")
		case "experience":
			parsePastedExperiences(sec.lines, p)
		case "education":
			parsePastedEducation(sec.lines, p)
		case "skills":
			parsePastedSkills(sec.lines, p)
		}
	}

	return p, nil
}

func parsePastedExperiences(lines []string, p *LinkedInImportedProfile) {
	// A typical paste pattern per job:
	// Line 0: Company
	// Line 1: Title
	// Line 2: Jan 2023 - Present • 3 yrs
	// Line 3: Location (optional)
	// Line 4+: Description
	var currentExp *LinkedInImportedExperience

	isDateLine := func(l string) bool {
		low := strings.ToLower(l)
		return strings.Contains(low, "present") || strings.Contains(low, " - ") || strings.Contains(low, " • ")
	}

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Check if next or next-next line is a date range, indicating a new company/title block
		if i+2 < len(lines) && isDateLine(lines[i+2]) {
			if currentExp != nil {
				p.Experiences = append(p.Experiences, *currentExp)
			}
			company := line
			title := lines[i+1]
			dateLine := lines[i+2]
			dates := strings.Split(dateLine, "•")[0]
			start, end, current := parseDateRange(dates)

			exp := LinkedInImportedExperience{
				CompanyName: company,
				Title:       title,
				StartDate:   start,
				EndDate:     end,
				IsCurrent:   current,
			}
			i += 3
			// Check for location
			if i < len(lines) && !isDateLine(lines[i]) && (strings.Contains(lines[i], ",") || strings.Contains(lines[i], "United States")) {
				exp.Location = lines[i]
				i++
			}
			// Collect description lines until next experience
			var descLines []string
			for i < len(lines) {
				if i+2 < len(lines) && isDateLine(lines[i+2]) {
					break
				}
				descLines = append(descLines, lines[i])
				i++
			}
			exp.Description = strings.Join(descLines, "\n")
			currentExp = &exp
			continue
		}

		i++
	}

	if currentExp != nil {
		p.Experiences = append(p.Experiences, *currentExp)
	}
}

func parsePastedEducation(lines []string, p *LinkedInImportedProfile) {
	// Pattern:
	// School
	// Degree / Field
	// Years (2017 - 2019)
	i := 0
	for i < len(lines) {
		school := lines[i]
		edu := LinkedInImportedEducation{
			SchoolName: school,
		}
		i++
		if i < len(lines) {
			edu.DegreeName = lines[i]
			i++
		}
		if i < len(lines) && (strings.Contains(lines[i], "-") || strings.Contains(lines[i], "20")) {
			start, end, _ := parseDateRange(lines[i])
			edu.StartDate = start
			edu.EndDate = end
			i++
		}
		p.Education = append(p.Education, edu)
	}
}

func parsePastedSkills(lines []string, p *LinkedInImportedProfile) {
	for _, l := range lines {
		// Split by bullet •, comma ,, or pipe |
		delims := []string{"•", ",", "|"}
		foundDelim := false
		for _, d := range delims {
			if strings.Contains(l, d) {
				parts := strings.Split(l, d)
				for _, part := range parts {
					s := strings.TrimSpace(part)
					if s != "" {
						p.Skills = append(p.Skills, s)
					}
				}
				foundDelim = true
				break
			}
		}
		if !foundDelim {
			s := strings.TrimSpace(l)
			if s != "" {
				p.Skills = append(p.Skills, s)
			}
		}
	}
}

func parseDateRange(s string) (start string, end string, isCurrent bool) {
	clean := strings.TrimSpace(s)
	parts := strings.Split(clean, "-")
	if len(parts) == 1 {
		parts = strings.Split(clean, "–")
	}
	if len(parts) >= 2 {
		start = strings.TrimSpace(parts[0])
		endStr := strings.TrimSpace(parts[1])
		if strings.EqualFold(endStr, "present") {
			isCurrent = true
			end = ""
		} else {
			end = endStr
		}
		return
	}
	start = clean
	return
}

// Enterprise API representation
type enterpriseProfilePayload struct {
	ID        string `json:"id"`
	FirstName struct {
		Localized map[string]string `json:"localized"`
	} `json:"firstName"`
	LastName struct {
		Localized map[string]string `json:"localized"`
	} `json:"lastName"`
	Headline struct {
		Localized map[string]string `json:"localized"`
	} `json:"headline"`
	Summary   string `json:"summary"`
	Positions []struct {
		CompanyName string `json:"companyName"`
		Title       string `json:"title"`
		Location    string `json:"location"`
		StartDate   struct {
			Year  int `json:"year"`
			Month int `json:"month"`
		} `json:"startDate"`
		EndDate struct {
			Year  int `json:"year"`
			Month int `json:"month"`
		} `json:"endDate"`
		IsCurrent   bool   `json:"isCurrent"`
		Description string `json:"description"`
	} `json:"positions"`
	Educations []struct {
		SchoolName   string `json:"schoolName"`
		DegreeName   string `json:"degreeName"`
		FieldOfStudy string `json:"fieldOfStudy"`
		StartDate    struct {
			Year int `json:"year"`
		} `json:"startDate"`
		EndDate struct {
			Year int `json:"year"`
		} `json:"endDate"`
	} `json:"educations"`
	Skills []string `json:"skills"`
}

// ParseLinkedInEnterprisePayload parses JSON output from approved enterprise partner APIs.
func ParseLinkedInEnterprisePayload(payload []byte) (*LinkedInImportedProfile, error) {
	if len(payload) == 0 {
		return nil, ErrInvalidJSONPayload
	}

	var data enterpriseProfilePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSONPayload, err)
	}

	p := &LinkedInImportedProfile{
		ImportID:       fmt.Sprintf("lip-imp-%d", time.Now().UnixNano()),
		SourceMode:     SourceEnterpriseAPI,
		Experiences:    make([]LinkedInImportedExperience, 0),
		Education:      make([]LinkedInImportedEducation, 0),
		Skills:         make([]string, 0),
		Certifications: make([]LinkedInImportedCertification, 0),
		ImportedAt:     time.Now().UTC(),
		Status:         "imported",
		Summary:        data.Summary,
	}

	// Localized name extraction
	fn := ""
	for _, v := range data.FirstName.Localized {
		fn = v
		break
	}
	ln := ""
	for _, v := range data.LastName.Localized {
		ln = v
		break
	}
	if fn != "" || ln != "" {
		p.DisplayName = strings.TrimSpace(fn + " " + ln)
	}

	for _, v := range data.Headline.Localized {
		p.Headline = v
		break
	}

	for _, pos := range data.Positions {
		start := ""
		if pos.StartDate.Year > 0 {
			if pos.StartDate.Month > 0 {
				start = fmt.Sprintf("%04d-%02d", pos.StartDate.Year, pos.StartDate.Month)
			} else {
				start = fmt.Sprintf("%04d", pos.StartDate.Year)
			}
		}
		end := ""
		if pos.EndDate.Year > 0 {
			if pos.EndDate.Month > 0 {
				end = fmt.Sprintf("%04d-%02d", pos.EndDate.Year, pos.EndDate.Month)
			} else {
				end = fmt.Sprintf("%04d", pos.EndDate.Year)
			}
		}

		p.Experiences = append(p.Experiences, LinkedInImportedExperience{
			CompanyName: pos.CompanyName,
			Title:       pos.Title,
			Location:    pos.Location,
			StartDate:   start,
			EndDate:     end,
			IsCurrent:   pos.IsCurrent,
			Description: pos.Description,
		})
	}

	for _, edu := range data.Educations {
		start := ""
		if edu.StartDate.Year > 0 {
			start = fmt.Sprintf("%04d", edu.StartDate.Year)
		}
		end := ""
		if edu.EndDate.Year > 0 {
			end = fmt.Sprintf("%04d", edu.EndDate.Year)
		}

		p.Education = append(p.Education, LinkedInImportedEducation{
			SchoolName:   edu.SchoolName,
			DegreeName:   edu.DegreeName,
			FieldOfStudy: edu.FieldOfStudy,
			StartDate:    start,
			EndDate:      end,
		})
	}

	p.Skills = append(p.Skills, data.Skills...)

	return p, nil
}

// ResolveOIDCBasicFallback produces explicit truth-in-advertising guidance when full scanning fails due to consumer token scope (AT-010).
func ResolveOIDCBasicFallback(claims map[string]interface{}) *LinkedInImportedProfile {
	p := &LinkedInImportedProfile{
		ImportID:                fmt.Sprintf("lip-imp-%d", time.Now().UnixNano()),
		SourceMode:              SourceOIDCBasicOnly,
		Experiences:             make([]LinkedInImportedExperience, 0),
		Education:               make([]LinkedInImportedEducation, 0),
		Skills:                  make([]string, 0),
		Certifications:          make([]LinkedInImportedCertification, 0),
		ImportedAt:              time.Now().UTC(),
		Status:                  "pending_review",
		RequiresFallback:        true,
		RecommendedFallback:     "user_export_upload",
		TruthInAdvertisingNote: "Standard LinkedIn Sign-in (OpenID Connect) provides identity claims only. Full career history API access is restricted to LinkedIn Talent Partners. Please upload your Basic_LinkedInData.zip export.",
	}

	if name, ok := claims["name"].(string); ok {
		p.DisplayName = name
	}
	if email, ok := claims["email"].(string); ok {
		p.Email = email
	}

	return p
}
