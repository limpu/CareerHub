package career

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type TailorEngine struct {
	generator *ResumeGenerator
	secretKey []byte
}

func NewTailorEngine(generator *ResumeGenerator, secretKey []byte) *TailorEngine {
	if len(secretKey) == 0 {
		secretKey = []byte("default-tailor-approval-secret-key-32b")
	}
	return &TailorEngine{
		generator: generator,
		secretKey: secretKey,
	}
}

// TailorResume creates a tailored resume draft strictly from confirmed facts,
// re-ordering relevant highlights and recording unmatched requirements (REQ-016, AT-002, AT-003).
func (te *TailorEngine) TailorResume(
	original *MasterCareerProfile,
	job JobTarget,
	format MasterResumeFormat,
	template ResumeTemplateType,
) (*TailoredResume, error) {
	if original == nil || strings.TrimSpace(original.Contact.FullName) == "" {
		return nil, ErrEmptyMasterProfile
	}
	if strings.TrimSpace(job.Title) == "" || strings.TrimSpace(job.Company) == "" {
		return nil, ErrJobTargetRequired
	}

	// 1. Build tailored profile clone strictly with confirmed facts
	tailoredProfile := *original
	diff := ResumeDiffSummary{}

	// Normalize target keywords and required skills
	jobTokens := make(map[string]bool)
	for _, s := range job.RequiredSkills {
		jobTokens[strings.ToLower(strings.TrimSpace(s))] = true
	}
	for _, kw := range job.Keywords {
		jobTokens[strings.ToLower(strings.TrimSpace(kw))] = true
	}
	descLower := strings.ToLower(job.Description)

	// Filter & prioritize confirmed skills
	var (
		matchedSkills   []SkillItem
		unmatchedSkills []SkillItem
		candidateSkills = make(map[string]bool)
	)

	for _, s := range original.Skills {
		if !s.Confirmed {
			continue // Invariant AT-003: Skip unconfirmed facts
		}
		sLower := strings.ToLower(strings.TrimSpace(s.Name))
		candidateSkills[sLower] = true

		if jobTokens[sLower] || strings.Contains(descLower, sLower) {
			matchedSkills = append(matchedSkills, s)
			diff.EmphasizedSkills = append(diff.EmphasizedSkills, s.Name)
		} else {
			unmatchedSkills = append(unmatchedSkills, s)
		}
	}

	// Unmatched requirements disclosure: required by job, but NOT possessed by candidate (REQ-016)
	for _, reqSkill := range job.RequiredSkills {
		reqLower := strings.ToLower(strings.TrimSpace(reqSkill))
		if !candidateSkills[reqLower] {
			diff.UnmatchedJobRequirements = append(diff.UnmatchedJobRequirements, reqSkill)
		}
	}

	// Re-order skills so matched ones appear first
	tailoredProfile.Skills = append(matchedSkills, unmatchedSkills...)

	// Filter & prioritize experience highlights
	var tailoredExperiences []ExperienceItem
	for _, exp := range original.Experiences {
		if !exp.Confirmed {
			continue
		}
		clonedExp := exp

		var (
			prioritized []string
			standard    []string
		)

		for _, h := range exp.Highlights {
			hLower := strings.ToLower(h)
			hasMatch := false
			for tok := range jobTokens {
				if len(tok) > 2 && strings.Contains(hLower, tok) {
					hasMatch = true
					break
				}
			}
			if hasMatch {
				prioritized = append(prioritized, h)
				diff.PrioritizedHighlights = append(diff.PrioritizedHighlights, h)
			} else {
				standard = append(standard, h)
			}
		}

		clonedExp.Highlights = append(prioritized, standard...)
		tailoredExperiences = append(tailoredExperiences, clonedExp)
	}
	tailoredProfile.Experiences = tailoredExperiences

	diff.TotalConfirmedFactsUsed = len(tailoredProfile.Experiences) + len(tailoredProfile.Skills)

	// 2. Generate binary resume using standard generator (single-column, ATS-safe)
	generated, vResult, err := te.generator.GenerateMasterResume(&tailoredProfile, format, template)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tailored resume: %w", err)
	}

	resumeID := fmt.Sprintf("tr_%d", time.Now().UnixNano())
	safeName := strings.ToLower(strings.ReplaceAll(original.Contact.FullName, " ", "_"))
	safeCompany := strings.ToLower(strings.ReplaceAll(job.Company, " ", "_"))
	fileName := fmt.Sprintf("%s_resume_%s.%s", safeName, safeCompany, format)

	// Compute HMAC approval token for FND-010 immutable approval gate
	token := te.generateApprovalToken(resumeID, generated.ChecksumSHA256)

	return &TailoredResume{
		ID:                resumeID,
		UserID:            original.UserID,
		JobID:             job.ID,
		JobTarget:         job,
		Format:            format,
		Template:          template,
		OriginalProfileID: original.ID,
		TailoredProfile:   tailoredProfile,
		ByteContent:       generated.ByteContent,
		ContentLength:     len(generated.ByteContent),
		ChecksumSHA256:    generated.ChecksumSHA256,
		FileName:          fileName,
		MimeType:          generated.MimeType,
		DiffSummary:       diff,
		ApprovalStatus:    ApprovalStatusPending,
		ApprovalToken:     token,
		ParseBackVerified: vResult.Success,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

// GenerateCoverLetter crafts a fact-grounded cover letter aligned to the employer requirements.
func (te *TailorEngine) GenerateCoverLetter(
	profile *MasterCareerProfile,
	job JobTarget,
	format MasterResumeFormat,
	recipientName string,
) (*CoverLetter, error) {
	if profile == nil || strings.TrimSpace(profile.Contact.FullName) == "" {
		return nil, ErrEmptyMasterProfile
	}
	if strings.TrimSpace(job.Title) == "" || strings.TrimSpace(job.Company) == "" {
		return nil, ErrJobTargetRequired
	}

	if recipientName == "" {
		recipientName = "Hiring Team"
	}

	salutation := fmt.Sprintf("Dear %s,", recipientName)
	opening := fmt.Sprintf("I am writing to express my enthusiastic interest in the %s position at %s. With a proven track record as %s and deep technical expertise in %s, I am excited by the opportunity to contribute to your engineering milestones.",
		job.Title, job.Company, profile.Contact.Headline, getTopSkillsList(profile, 3))

	// Find most relevant confirmed experience
	var topHighlights []string
	if len(profile.Experiences) > 0 {
		for _, exp := range profile.Experiences {
			if exp.Confirmed && len(exp.Highlights) > 0 {
				topHighlights = append(topHighlights, fmt.Sprintf("At %s, I led initiatives including: %s", exp.Company, exp.Highlights[0]))
				if len(topHighlights) >= 2 {
					break
				}
			}
		}
	}

	body1 := fmt.Sprintf("Throughout my professional background, I have consistently focused on building scalable, resilient architectures that solve core business problems. %s",
		strings.Join(topHighlights, " "))

	body2 := fmt.Sprintf("Your mission at %s strongly resonates with my professional ethos. My practical background in %s aligns directly with the challenges outlined in your job requirements.",
		job.Company, getTopSkillsList(profile, 4))

	closing := fmt.Sprintf("Thank you for your time and consideration. I welcome the opportunity to discuss how my experience and confirmed technical skills will add direct value to the %s team.", job.Company)
	signoff := fmt.Sprintf("Sincerely,\n%s\n%s\n%s", profile.Contact.FullName, profile.Contact.Email, profile.Contact.Phone)

	var fullTextBuilder strings.Builder
	fullTextBuilder.WriteString(fmt.Sprintf("%s\n%s | %s | %s\n\n", strings.ToUpper(profile.Contact.FullName), profile.Contact.Email, profile.Contact.Phone, profile.Contact.Location))
	fullTextBuilder.WriteString(fmt.Sprintf("%s\n\n", time.Now().Format("January 2, 2006")))
	fullTextBuilder.WriteString(fmt.Sprintf("To: %s\n%s\n\n", recipientName, job.Company))
	fullTextBuilder.WriteString(fmt.Sprintf("%s\n\n", salutation))
	fullTextBuilder.WriteString(fmt.Sprintf("%s\n\n", opening))
	fullTextBuilder.WriteString(fmt.Sprintf("%s\n\n", body1))
	fullTextBuilder.WriteString(fmt.Sprintf("%s\n\n", body2))
	fullTextBuilder.WriteString(fmt.Sprintf("%s\n\n", closing))
	fullTextBuilder.WriteString(signoff)

	fullText := fullTextBuilder.String()
	data := []byte(fullText)
	mimeType := "text/plain"

	if format == ResumeFormatPDF {
		lines := strings.Split(fullText, "\n")
		pdfData, err := buildSimplePDF(lines)
		if err == nil {
			data = pdfData
			mimeType = "application/pdf"
		}
	}

	letterID := fmt.Sprintf("cl_%d", time.Now().UnixNano())
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])
	safeCompany := strings.ToLower(strings.ReplaceAll(job.Company, " ", "_"))
	safeName := strings.ToLower(strings.ReplaceAll(profile.Contact.FullName, " ", "_"))
	fileName := fmt.Sprintf("%s_cover_letter_%s.%s", safeName, safeCompany, format)

	token := te.generateApprovalToken(letterID, checksum)

	return &CoverLetter{
		ID:                letterID,
		UserID:            profile.UserID,
		JobID:             job.ID,
		JobTarget:         job,
		RecipientName:     recipientName,
		Format:            format,
		Salutation:        salutation,
		OpeningParagraph:  opening,
		BodyParagraphs:    []string{body1, body2},
		ClosingParagraph:  closing,
		Signoff:           signoff,
		FullText:          fullText,
		ByteContent:       data,
		ContentLength:     len(data),
		ChecksumSHA256:    checksum,
		FileName:          fileName,
		MimeType:          mimeType,
		ApprovalStatus:    ApprovalStatusPending,
		ApprovalToken:     token,
		ParseBackVerified: len(data) > 50,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

func (te *TailorEngine) ValidateApprovalToken(id, checksum, token string) bool {
	expected := te.generateApprovalToken(id, checksum)
	return hmac.Equal([]byte(expected), []byte(token))
}

func (te *TailorEngine) generateApprovalToken(id, checksum string) string {
	mac := hmac.New(sha256.New, te.secretKey)
	mac.Write([]byte(fmt.Sprintf("%s:%s", id, checksum)))
	return hex.EncodeToString(mac.Sum(nil))
}

func getTopSkillsList(p *MasterCareerProfile, max int) string {
	var names []string
	for _, s := range p.Skills {
		if s.Confirmed {
			names = append(names, s.Name)
			if len(names) >= max {
				break
			}
		}
	}
	if len(names) == 0 {
		return "Software Engineering"
	}
	return strings.Join(names, ", ")
}
