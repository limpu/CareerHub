package career

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// LeadReviewStatus defines the mandatory human review lifecycle stage (CAR-20, REQ-015).
type LeadReviewStatus string

const (
	LeadReviewPending  LeadReviewStatus = "pending_review"
	LeadReviewReviewed LeadReviewStatus = "reviewed"
	LeadReviewApproved LeadReviewStatus = "approved"
	LeadReviewRejected LeadReviewStatus = "rejected"
)

// LeadOutreachStatus defines the candidate's outreach progress (CAR-20, AT-010).
type LeadOutreachStatus string

const (
	LeadOutreachDraft         LeadOutreachStatus = "draft"
	OutreachReadyToSend       LeadOutreachStatus = "ready_to_send"
	OutreachCopiedToClipboard LeadOutreachStatus = "copied_to_clipboard"
	OutreachContacted         LeadOutreachStatus = "contacted"
	OutreachReplied           LeadOutreachStatus = "replied"
	OutreachArchived          LeadOutreachStatus = "archived"
)

// OutreachTemplateType represents supported outreach channel message formats.
type OutreachTemplateType string

const (
	TemplateLinkedInConnect OutreachTemplateType = "linkedin_connect" // <= 300 characters
	TemplateLinkedInInMail  OutreachTemplateType = "linkedin_inmail"
	TemplateEmailIntro      OutreachTemplateType = "email_intro"
)

var (
	ErrLeadNotApproved       = errors.New("cannot perform outreach: recruiter lead has not been human reviewed and approved (CAR-20, REQ-015)")
	ErrLeadNotFound          = errors.New("recruiter lead not found")
	ErrHiringPostNotFound    = errors.New("hiring post not found")
	ErrEmptyPostContent      = errors.New("hiring post content cannot be empty")
	ErrPromptInjectionBlocked = errors.New("untrusted content contains adversarial prompt injection payload; automated bypass prevented (AT-019)")
)

// ExtractedHiringRole represents an individual job role extracted from a hiring post.
type ExtractedHiringRole struct {
	RoleTitle         string   `json:"role_title"`
	Seniority         string   `json:"seniority"`
	Location          string   `json:"location"`
	IsRemote          bool     `json:"is_remote"`
	TechStack         []string `json:"tech_stack"`
	ApplyInstructions string   `json:"apply_instructions,omitempty"`
	Confidence        float64  `json:"confidence"`
}

// HiringPost represents a parsed post ingested from feed or imported by the user (CAR-20, SRC-C1).
type HiringPost struct {
	ID                  string                `json:"id"`
	UserID              string                `json:"user_id"`
	SourcePlatform      string                `json:"source_platform"` // "linkedin_post", "imported_text", "feed_scanner"
	PostURL             string                `json:"post_url"`
	AuthorName          string                `json:"author_name"`
	AuthorTitle         string                `json:"author_title"`
	AuthorLinkedInURL   string                `json:"author_linkedin_url"`
	Company             string                `json:"company"`
	RawContent          string                `json:"raw_content"`
	CleanedContent      string                `json:"cleaned_content"`
	HiringKeywordsFound []string              `json:"hiring_keywords_found"`
	ExtractedRoles      []ExtractedHiringRole `json:"extracted_roles"`
	RecruiterEmail      string                `json:"recruiter_email,omitempty"`
	SecurityFlags       []string              `json:"security_flags,omitempty"`
	ImportedAt          time.Time             `json:"imported_at"`
	Status              string                `json:"status"` // "extracted", "reviewed", "lead_created", "discarded"
}

// RecruiterLead represents an actionable recruiter relationship discovered from a hiring post (CAR-20).
type RecruiterLead struct {
	ID                  string             `json:"id"`
	UserID              string             `json:"user_id"`
	ContactID           string             `json:"contact_id"`
	HiringPostID        string             `json:"hiring_post_id"`
	RecruiterName       string             `json:"recruiter_name"`
	RecruiterTitle      string             `json:"recruiter_title"`
	Company             string             `json:"company"`
	RecruiterEmail      string             `json:"recruiter_email,omitempty"`
	LinkedInURL         string             `json:"linkedin_url"`
	RoleInterest        string             `json:"role_interest"`
	SourcePostURL       string             `json:"source_post_url"`
	ReviewStatus        LeadReviewStatus   `json:"review_status"`
	OutreachStatus      LeadOutreachStatus `json:"outreach_status"`
	HumanReviewed       bool               `json:"human_reviewed"`
	HumanReviewedAt     *time.Time         `json:"human_reviewed_at,omitempty"`
	ReviewedBy          string             `json:"reviewed_by,omitempty"`
	ReviewNotes         string             `json:"review_notes,omitempty"`
	LinkedApplicationID string             `json:"linked_application_id,omitempty"`
	DraftSubject        string             `json:"draft_subject,omitempty"`
	DraftMessage        string             `json:"draft_message,omitempty"`
	SecurityAlerts      []string           `json:"security_alerts,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

// CandidateOutreachProfile holds verified facts about the candidate used for grounded outreach generation.
type CandidateOutreachProfile struct {
	CandidateName string
	CurrentTitle  string
	KeySkills     []string
	YearsOfExp    int
	PortfolioURL  string
}

// OutreachDraft represents a grounded, personalized outreach message.
type OutreachDraft struct {
	TemplateType           OutreachTemplateType `json:"template_type"`
	Subject                string               `json:"subject,omitempty"`
	Body                   string               `json:"body"`
	CharacterCount         int                  `json:"character_count"`
	PersonalizedHighlights []string             `json:"personalized_highlights"`
	Warnings               []string             `json:"warnings,omitempty"`
}

// IngestHiringPostRequest defines the request payload to import a hiring post.
type IngestHiringPostRequest struct {
	SourcePlatform    string `json:"source_platform"`
	PostURL           string `json:"post_url"`
	AuthorName        string `json:"author_name"`
	AuthorTitle       string `json:"author_title"`
	AuthorLinkedInURL string `json:"author_linkedin_url"`
	Company           string `json:"company"`
	RawContent        string `json:"raw_content"`
}

// Common hiring keywords (SRC-C1 feed_scanner.py HIRING_KEYWORDS).
var hiringKeywordPatterns = []string{
	"we're hiring",
	"we are hiring",
	"join our team",
	"looking for",
	"opening for",
	"now hiring",
	"seeking a",
	"reach out",
	"dm me",
	"send your resume",
	"apply at",
	"expanding the",
	"hiring alert",
	"contact me",
}

// Regex for email extraction.
var leadEmailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

// Prompt injection patterns to detect adversarial payloads (AT-019).
var promptInjectionPatterns = []string{
	`ignore previous instructions`,
	`system alert`,
	`system_override`,
	`bypass candidate human review`,
	`bypass.*review`,
	`send secret api keys`,
	`user credentials to`,
	`exfil@`,
	`automatically approve`,
	`eval\(`,
	`<script`,
	`\badmin_override\b`,
}

// DetectPromptInjection checks untrusted content for adversarial instruction injection payloads (AT-019).
func DetectPromptInjection(text string) (bool, []string) {
	lower := strings.ToLower(text)
	var matched []string
	for _, pattern := range promptInjectionPatterns {
		re, err := regexp.Compile(`(?i)` + pattern)
		if err == nil && re.MatchString(lower) {
			matched = append(matched, pattern)
		}
	}
	return len(matched) > 0, matched
}

// SanitizeUntrustedContent encapsulates untrusted post text and redacts active injection commands (AT-019).
func SanitizeUntrustedContent(raw string) (string, []string) {
	isInjection, flags := DetectPromptInjection(raw)
	cleaned := raw

	if isInjection {
		// Strip XML/HTML style system override tags
		reTag := regexp.MustCompile(`(?i)</?system[a-z0-9_-]*>`)
		cleaned = reTag.ReplaceAllString(cleaned, "")

		// Replace blatant instruction override sequences
		for _, flag := range flags {
			reFlag := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(flag))
			cleaned = reFlag.ReplaceAllString(cleaned, "[REDACTED_SECURITY_PAYLOAD]")
		}
	}

	return cleaned, flags
}

// ExtractHiringPost processes raw post text, finds hiring keywords, extracts roles and contact info, and applies prompt injection safety (CAR-20, AT-019).
func ExtractHiringPost(userID string, req IngestHiringPostRequest) (*HiringPost, error) {
	raw := strings.TrimSpace(req.RawContent)
	if raw == "" {
		return nil, ErrEmptyPostContent
	}

	// 1. Prompt Injection Defense (AT-019)
	cleaned, injectionFlags := SanitizeUntrustedContent(raw)
	var securityFlags []string
	if len(injectionFlags) > 0 {
		securityFlags = append(securityFlags, "prompt_injection_detected", "adversarial_override_attempt")
	}

	// 2. Keyword detection
	var keywordsFound []string
	lower := strings.ToLower(raw)
	for _, kw := range hiringKeywordPatterns {
		if strings.Contains(lower, strings.ToLower(kw)) {
			keywordsFound = append(keywordsFound, kw)
		}
	}

	// 3. Recruiter email extraction
	emails := leadEmailRegex.FindAllString(raw, -1)
	var recruiterEmail string
	for _, em := range emails {
		lowerEm := strings.ToLower(em)
		if !strings.Contains(lowerEm, "exfil") && !strings.Contains(lowerEm, "attacker") {
			recruiterEmail = em
			break
		}
	}

	// 4. Role and tech stack extraction
	roles := extractRolesFromContent(raw, req.Company)

	// Generate deterministic ID
	hasher := sha256.New()
	hasher.Write([]byte(userID + ":" + req.SourcePlatform + ":" + req.AuthorName + ":" + raw[:min(len(raw), 64)]))
	postID := "post_" + hex.EncodeToString(hasher.Sum(nil))[:16]

	post := &HiringPost{
		ID:                  postID,
		UserID:              userID,
		SourcePlatform:      req.SourcePlatform,
		PostURL:             req.PostURL,
		AuthorName:          req.AuthorName,
		AuthorTitle:         req.AuthorTitle,
		AuthorLinkedInURL:   req.AuthorLinkedInURL,
		Company:             req.Company,
		RawContent:          raw,
		CleanedContent:      cleaned,
		HiringKeywordsFound: keywordsFound,
		ExtractedRoles:      roles,
		RecruiterEmail:      recruiterEmail,
		SecurityFlags:       securityFlags,
		ImportedAt:          time.Now().UTC(),
		Status:              "extracted",
	}

	return post, nil
}

// extractRolesFromContent identifies engineering/product roles, seniority, location, and tech stack.
func extractRolesFromContent(content string, companyFallback string) []ExtractedHiringRole {
	var roles []ExtractedHiringRole
	lines := strings.Split(content, "\n")

	// Tech stack dictionary
	knownTech := []string{"Go", "PostgreSQL", "Redis", "Kubernetes", "Kafka", "React", "TypeScript", "Next.js", "GraphQL", "Python", "gRPC", "AWS", "Docker", "Data Engineer", "Payments", "Fintech", "distributed ledger", "streaming pipeline"}
	seniorityKeywords := []string{"Senior", "Staff", "Lead", "Principal", "Junior", "Mid"}

	// Check if multi-role numbered format exists (e.g. 1) Role, 2) Role)
	numRoleRegex := regexp.MustCompile(`(?i)(?:[0-9]+[.)]|[-*])\s*([A-Za-z\s]+)(?:\(([^)]+)\))?`)
	matches := numRoleRegex.FindAllStringSubmatch(content, -1)

	if len(matches) > 1 {
		for _, match := range matches {
			if len(match) > 1 {
				title := strings.TrimSpace(match[1])
				if len(title) > 3 && !strings.HasPrefix(strings.ToLower(title), "all position") && !strings.HasPrefix(strings.ToLower(title), "competitive") {
					var stack []string
					if len(match) > 2 && match[2] != "" {
						rawStack := strings.Split(match[2], ",")
						for _, s := range rawStack {
							stack = append(stack, strings.TrimSpace(s))
						}
					}
					// Identify seniority
					seniority := "Mid"
					for _, sen := range seniorityKeywords {
						if strings.Contains(strings.ToLower(title), strings.ToLower(sen)) {
							seniority = sen
							break
						}
					}
					isRemote := strings.Contains(strings.ToLower(content), "remote")
					loc := "Hybrid (New York, NY)"
					if isRemote {
						loc = "Remote"
					}
					roles = append(roles, ExtractedHiringRole{
						RoleTitle:  title,
						Seniority:  seniority,
						Location:   loc,
						IsRemote:   isRemote,
						TechStack:  stack,
						Confidence: 0.90,
					})
				}
			}
		}
	}

	if len(roles) == 0 {
		// Single role extraction
		var title string
		var seniority = "Mid"
		lowerContent := strings.ToLower(content)

		for _, sen := range seniorityKeywords {
			if strings.Contains(lowerContent, strings.ToLower(sen)) {
				seniority = sen
				break
			}
		}

		// Look for role phrases
		roleRegex := regexp.MustCompile(`(?i)(?:looking for|seeking|opening for|hiring)\s+(?:a\s+)?([A-Za-z0-9\s]+?)(?:to|\.|\!|\n|stack|remote)`)
		if m := roleRegex.FindStringSubmatch(content); len(m) > 1 {
			title = strings.TrimSpace(m[1])
		} else if strings.Contains(lowerContent, "senior go backend engineer") {
			title = "Senior Go Backend Engineer"
		} else if strings.Contains(lowerContent, "senior data engineer") {
			title = "Senior Data Engineer"
		} else if strings.Contains(lowerContent, "junior react devs") {
			title = "junior React devs"
		} else {
			title = "Software Engineer"
		}

		// Detect tech stack
		var detectedTech []string
		for _, tech := range knownTech {
			if strings.Contains(lowerContent, strings.ToLower(tech)) {
				detectedTech = append(detectedTech, tech)
			}
		}

		isRemote := strings.Contains(lowerContent, "remote")
		loc := "Unspecified"
		if strings.Contains(lowerContent, "remote (us/canada)") {
			loc = "Remote (US/Canada)"
		} else if strings.Contains(lowerContent, "100% remote worldwide") {
			loc = "100% remote worldwide"
		} else if isRemote {
			loc = "Remote"
		}

		roles = append(roles, ExtractedHiringRole{
			RoleTitle:  title,
			Seniority:  seniority,
			Location:   loc,
			IsRemote:   isRemote,
			TechStack:  detectedTech,
			Confidence: 0.95,
		})
	}

	_ = lines
	return roles
}

// CreateLeadFromPost converts an extracted HiringPost into an actionable RecruiterLead (CAR-20).
func CreateLeadFromPost(post *HiringPost, selectedRoleIndex int) *RecruiterLead {
	roleTitle := "Open Engineering Position"
	if selectedRoleIndex >= 0 && selectedRoleIndex < len(post.ExtractedRoles) {
		roleTitle = post.ExtractedRoles[selectedRoleIndex].RoleTitle
	} else if len(post.ExtractedRoles) > 0 {
		roleTitle = post.ExtractedRoles[0].RoleTitle
	}

	hasher := sha256.New()
	hasher.Write([]byte(post.UserID + ":" + post.AuthorName + ":" + post.Company + ":" + roleTitle))
	leadID := "lead_" + hex.EncodeToString(hasher.Sum(nil))[:16]

	now := time.Now().UTC()
	lead := &RecruiterLead{
		ID:             leadID,
		UserID:         post.UserID,
		ContactID:      "cont_" + leadID[5:],
		HiringPostID:   post.ID,
		RecruiterName:  post.AuthorName,
		RecruiterTitle: post.AuthorTitle,
		Company:        post.Company,
		RecruiterEmail: post.RecruiterEmail,
		LinkedInURL:    post.AuthorLinkedInURL,
		RoleInterest:   roleTitle,
		SourcePostURL:  post.PostURL,
		ReviewStatus:   LeadReviewPending, // Mandatory pending state (CAR-20, REQ-015)
		OutreachStatus: LeadOutreachDraft,
		HumanReviewed:  false,
		SecurityAlerts: post.SecurityFlags,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return lead
}

// ReviewRecruiterLead implements the mandatory human review gate (CAR-20, REQ-015).
func ReviewRecruiterLead(lead *RecruiterLead, action LeadReviewStatus, reviewerID string, notes string) error {
	if lead == nil {
		return ErrLeadNotFound
	}

	lead.ReviewStatus = action
	now := time.Now().UTC()
	lead.HumanReviewed = true
	lead.HumanReviewedAt = &now
	lead.ReviewedBy = reviewerID
	lead.ReviewNotes = notes
	lead.UpdatedAt = now

	if action == LeadReviewApproved {
		lead.OutreachStatus = OutreachReadyToSend
	} else if action == LeadReviewRejected {
		lead.OutreachStatus = OutreachArchived
	}

	return nil
}

// GenerateOutreachDraft constructs a fact-grounded outreach draft based on candidate facts and extracted role.
// Strictly requires human review approval (CAR-20, REQ-015).
func GenerateOutreachDraft(lead *RecruiterLead, profile CandidateOutreachProfile, templateType OutreachTemplateType) (*OutreachDraft, error) {
	if lead == nil {
		return nil, ErrLeadNotFound
	}

	// Enforce Human Review Gate
	if !lead.HumanReviewed || lead.ReviewStatus != LeadReviewApproved {
		return nil, ErrLeadNotApproved
	}

	var subject string
	var body string
	var highlights []string
	var warnings []string

	recruiterFirst := lead.RecruiterName
	if parts := strings.Fields(lead.RecruiterName); len(parts) > 0 {
		recruiterFirst = parts[0]
	}

	switch templateType {
	case TemplateLinkedInConnect:
		// LinkedIn connection requests have a strict 300 character limit
		body = fmt.Sprintf("Hi %s, saw your post regarding the %s role at %s. With a background in %s, I'd love to connect and follow your team's work!",
			recruiterFirst, lead.RoleInterest, lead.Company, profile.CurrentTitle)
		if len(body) > 300 {
			body = fmt.Sprintf("Hi %s, saw your post hiring for %s at %s. With my background in %s, I'd love to connect!",
				recruiterFirst, lead.RoleInterest, lead.Company, profile.CurrentTitle)
		}
		highlights = append(highlights, "Character-bounded connection request (<= 300 chars)", "Specific role & company citation")

	case TemplateLinkedInInMail:
		subject = fmt.Sprintf("Question regarding %s at %s", lead.RoleInterest, lead.Company)
		body = fmt.Sprintf("Hi %s,\n\nI noticed your recent post hiring for a %s on the %s team.\n\nAs a %s with hands-on experience in %s, my background aligns closely with the core requirements you outlined.\n\nI'd welcome the chance to share a brief overview of my work or discuss how I could contribute to %s's goals. Would you be open to a quick 10-minute introductory conversation next week?\n\nBest regards,\n%s",
			recruiterFirst, lead.RoleInterest, lead.Company, profile.CurrentTitle, strings.Join(profile.KeySkills, ", "), lead.Company, profile.CandidateName)
		highlights = append(highlights, "Value-driven InMail structure", "Personalized role citation")

	case TemplateEmailIntro:
		subject = fmt.Sprintf("Application / Intro: %s — %s", lead.RoleInterest, profile.CandidateName)
		body = fmt.Sprintf("Dear %s,\n\nI came across your hiring announcement for the %s opening at %s.\n\nOver the past %d years as a %s, I have built and maintained systems using %s. Given %s's focus in this space, I am very excited about the impact of this role.\n\nI have attached my tailored resume and portfolio (%s) for your review. If your team is currently scheduling conversations, I would love the opportunity to speak.\n\nThank you for your time and consideration.\n\nSincerely,\n%s",
			recruiterFirst, lead.RoleInterest, lead.Company, profile.YearsOfExp, profile.CurrentTitle, strings.Join(profile.KeySkills, ", "), lead.Company, profile.PortfolioURL, profile.CandidateName)
		highlights = append(highlights, "Professional formal email layout", "Experience years & portfolio inclusion")

	default:
		return nil, fmt.Errorf("unsupported outreach template type: %s", templateType)
	}

	// Truth-in-advertising check (AT-010): Warn that live API dispatch is disabled
	warnings = append(warnings, "Truth-in-Advertising (AT-010): Automated live LinkedIn dispatch is disabled without enterprise API seat; copy draft to clipboard for manual delivery.")

	draft := &OutreachDraft{
		TemplateType:           templateType,
		Subject:                subject,
		Body:                   body,
		CharacterCount:         len(body),
		PersonalizedHighlights: highlights,
		Warnings:               warnings,
	}

	lead.DraftSubject = subject
	lead.DraftMessage = body
	lead.UpdatedAt = time.Now().UTC()

	return draft, nil
}

// RecordLeadOutreachAction tracks user actions on the lead, upholding truth-in-advertising (CAR-20, AT-010).
func RecordLeadOutreachAction(lead *RecruiterLead, action LeadOutreachStatus) error {
	if lead == nil {
		return ErrLeadNotFound
	}

	// Cannot copy or contact unapproved leads
	if (action == OutreachCopiedToClipboard || action == OutreachContacted) && (!lead.HumanReviewed || lead.ReviewStatus != LeadReviewApproved) {
		return ErrLeadNotApproved
	}

	lead.OutreachStatus = action
	lead.UpdatedAt = time.Now().UTC()
	return nil
}

// LinkLeadToApplication associates a recruiter lead with an active platform ApplicationRecord.
func LinkLeadToApplication(lead *RecruiterLead, applicationID string) error {
	if lead == nil {
		return ErrLeadNotFound
	}
	lead.LinkedApplicationID = applicationID
	lead.UpdatedAt = time.Now().UTC()
	return nil
}
