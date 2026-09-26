package career

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestMasterCareerProfile_CRUD(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-car-001"

	// 1. Get initial profile (should initialize default blank profile with REQ-018 isolation)
	profile, err := service.GetMasterProfile(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error getting master profile: %v", err)
	}

	if profile.UserID != userID {
		t.Errorf("expected UserID %q, got %q", userID, profile.UserID)
	}
	if profile.Privacy.ShareWithWorkspaceAdmin {
		t.Errorf("REQ-018 violation: ShareWithWorkspaceAdmin must be false by default")
	}
	if profile.Completeness.Score != 0 {
		t.Errorf("expected initial completeness score 0, got %v", profile.Completeness.Score)
	}

	// 2. Update Contact Section
	contactPayload, _ := json.Marshal(ContactInfo{
		FullName:    "Alice Rahman",
		Email:       "alice@example.com",
		Phone:       "+1234567890",
		Location:    "Dhaka, Bangladesh",
		Headline:    "Senior Systems Engineer",
		Summary:     "Specializing in distributed Go architectures",
		Citizenship: "Bangladeshi",
	})
	profile, err = service.UpdateProfileSection(ctx, userID, "contact", contactPayload, userID)
	if err != nil {
		t.Fatalf("failed to update contact section: %v", err)
	}
	if profile.Contact.FullName != "Alice Rahman" {
		t.Errorf("expected full name 'Alice Rahman', got %q", profile.Contact.FullName)
	}

	// 3. Update Skills Section
	skillsPayload, _ := json.Marshal([]SkillItem{
		{ID: "sk-1", Name: "Go", Category: "technical", Proficiency: ProficiencyExpert, YearsOfExperience: 6, Confirmed: true},
		{ID: "sk-2", Name: "PostgreSQL", Category: "technical", Proficiency: ProficiencyAdvanced, YearsOfExperience: 5, Confirmed: true},
	})
	profile, err = service.UpdateProfileSection(ctx, userID, "skills", skillsPayload, userID)
	if err != nil {
		t.Fatalf("failed to update skills section: %v", err)
	}
	if len(profile.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(profile.Skills))
	}

	// 4. Update Experiences Section
	now := time.Now()
	expPayload, _ := json.Marshal([]ExperienceItem{
		{
			ID:          "exp-1",
			Title:       "Staff Engineer",
			Company:     "Acme Corp",
			Location:    "Remote",
			StartDate:   &now,
			IsCurrent:   true,
			Description: "Leading core platform team",
			Highlights:  []string{"Reduced p99 latency by 45%"},
			SkillsUsed:  []string{"Go", "Redis"},
			Confirmed:   true,
		},
	})
	profile, err = service.UpdateProfileSection(ctx, userID, "experiences", expPayload, userID)
	if err != nil {
		t.Fatalf("failed to update experiences section: %v", err)
	}
	if len(profile.Experiences) != 1 {
		t.Errorf("expected 1 experience, got %d", len(profile.Experiences))
	}

	// 5. Update Education Section
	eduPayload, _ := json.Marshal([]EducationItem{
		{
			ID:           "edu-1",
			Institution:  "Engineering University",
			Degree:       "B.Sc. in Computer Science",
			FieldOfStudy: "Software Engineering",
			Confirmed:    true,
		},
	})
	profile, err = service.UpdateProfileSection(ctx, userID, "education", eduPayload, userID)
	if err != nil {
		t.Fatalf("failed to update education section: %v", err)
	}
	if len(profile.Education) != 1 {
		t.Errorf("expected 1 education item, got %d", len(profile.Education))
	}

	// 6. Update Links Section
	linksPayload, _ := json.Marshal([]ProfileLink{
		{ID: "lnk-1", Label: "GitHub", URL: "https://github.com/alicerahman", LinkType: "github", Confirmed: true},
	})
	profile, err = service.UpdateProfileSection(ctx, userID, "links", linksPayload, userID)
	if err != nil {
		t.Fatalf("failed to update links: %v", err)
	}

	// 7. Update Projects Section
	projPayload, _ := json.Marshal([]ProjectItem{
		{ID: "proj-1", Title: "Distributed Cache", Description: "Raft-based key-value store", Technologies: []string{"Go", "gRPC"}, Confirmed: true},
	})
	profile, err = service.UpdateProfileSection(ctx, userID, "projects", projPayload, userID)
	if err != nil {
		t.Fatalf("failed to update projects: %v", err)
	}

	// 8. Update Certificates & Languages
	certPayload, _ := json.Marshal([]CertificateItem{
		{ID: "cert-1", Name: "AWS Certified Solutions Architect", Issuer: "Amazon", Confirmed: true},
	})
	profile, err = service.UpdateProfileSection(ctx, userID, "certificates", certPayload, userID)
	if err != nil {
		t.Fatalf("failed to update certificates: %v", err)
	}

	langPayload, _ := json.Marshal([]LanguageItem{
		{ID: "lang-1", Language: "English", Proficiency: LangFluent, Confirmed: true},
		{ID: "lang-2", Language: "Bengali", Proficiency: LangNative, Confirmed: true},
	})
	profile, err = service.UpdateProfileSection(ctx, userID, "languages", langPayload, userID)
	if err != nil {
		t.Fatalf("failed to update languages: %v", err)
	}

	// Check Completeness: All 8 sections populated => Score should be 100%
	if profile.Completeness.Score != 100.0 {
		t.Errorf("expected completeness score 100.0, got %v (missing: %v)", profile.Completeness.Score, profile.Completeness.MissingSections)
	}
}

func TestMasterCareerProfile_FieldLevelAudit(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-audit-001"

	// Update contact twice
	c1, _ := json.Marshal(ContactInfo{FullName: "Initial Name", Email: "initial@test.com"})
	_, _ = service.UpdateProfileSection(ctx, userID, "contact", c1, "admin-agent")

	c2, _ := json.Marshal(ContactInfo{FullName: "Updated Name", Email: "initial@test.com"})
	profile, err := service.UpdateProfileSection(ctx, userID, "contact", c2, "user-self")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if len(profile.AuditTrail) < 2 {
		t.Fatalf("expected at least 2 audit entries, got %d", len(profile.AuditTrail))
	}

	lastAudit := profile.AuditTrail[len(profile.AuditTrail)-1]
	if lastAudit.Field != "contact" {
		t.Errorf("expected audit field 'contact', got %q", lastAudit.Field)
	}
	if lastAudit.OldValue != "Initial Name" || lastAudit.NewValue != "Updated Name" {
		t.Errorf("unexpected audit values: old=%q, new=%q", lastAudit.OldValue, lastAudit.NewValue)
	}
	if lastAudit.ChangedBy != "user-self" {
		t.Errorf("expected changed by 'user-self', got %q", lastAudit.ChangedBy)
	}
}

func TestMasterCareerProfile_ConsentManagement(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-consent-001"

	consent := UserCareerConsent{
		DataProcessingConsent:    true,
		JobMatchingConsent:       false,
		ThirdPartySharingConsent: false,
		ConsentVersion:           "v1.2",
	}

	profile, err := service.UpdateCareerConsent(ctx, userID, consent)
	if err != nil {
		t.Fatalf("failed to update consent: %v", err)
	}

	if !profile.Consent.DataProcessingConsent {
		t.Errorf("expected DataProcessingConsent true")
	}
	if profile.Consent.JobMatchingConsent {
		t.Errorf("expected JobMatchingConsent false")
	}
	if profile.Consent.ConsentVersion != "v1.2" {
		t.Errorf("expected consent version 'v1.2', got %q", profile.Consent.ConsentVersion)
	}
	if profile.Consent.ConsentTimestamp.IsZero() {
		t.Errorf("expected non-zero consent timestamp")
	}
}

func TestREQ018_WorkspaceAdminIsolation(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	ownerUserID := "candidate-owner-001"
	workspaceAdminID := "workspace-admin-002"
	unrelatedUserID := "regular-user-003"

	// 1. Owner populates profile with default privacy (ShareWithWorkspaceAdmin = false)
	profile, err := service.GetMasterProfile(ctx, ownerUserID)
	if err != nil {
		t.Fatalf("failed to get profile: %v", err)
	}
	if profile.Privacy.ShareWithWorkspaceAdmin {
		t.Fatal("REQ-018 failure: default privacy must NOT share with workspace admin")
	}

	// 2. Owner can read their own profile
	ownerProfile, err := service.GetMasterProfileWithPrivacyCheck(ctx, ownerUserID, ownerUserID, false)
	if err != nil {
		t.Fatalf("owner should be able to access own profile: %v", err)
	}
	if ownerProfile == nil {
		t.Fatal("owner profile is nil")
	}

	// 3. Workspace Admin attempts to access private personal profile -> MUST FAIL CLOSED (REQ-018)
	_, err = service.GetMasterProfileWithPrivacyCheck(ctx, workspaceAdminID, ownerUserID, true)
	if err == nil {
		t.Fatalf("REQ-018 violation: workspace admin accessed personal career profile without explicit grant")
	}
	if !errors.Is(err, ErrPrivateProfileAccessDenied) {
		t.Errorf("expected ErrPrivateProfileAccessDenied, got: %v", err)
	}

	// 4. Regular unrelated user attempts access -> MUST FAIL CLOSED
	_, err = service.GetMasterProfileWithPrivacyCheck(ctx, unrelatedUserID, ownerUserID, false)
	if err == nil {
		t.Fatalf("unrelated user accessed private profile")
	}
	if !errors.Is(err, ErrPrivateProfileAccessDenied) {
		t.Errorf("expected ErrPrivateProfileAccessDenied, got: %v", err)
	}

	// 5. Owner explicitly grants permission to Workspace Admin
	_, err = service.UpdatePrivacySettings(ctx, ownerUserID, CareerPrivacySettings{
		ProfileVisibility:       PrivacyPublic,
		ShareWithWorkspaceAdmin: true, // Explicitly granted!
		AllowRecruiterView:      true,
	})
	if err != nil {
		t.Fatalf("failed to update privacy settings: %v", err)
	}

	// 6. Now Workspace Admin CAN access
	adminView, err := service.GetMasterProfileWithPrivacyCheck(ctx, workspaceAdminID, ownerUserID, true)
	if err != nil {
		t.Fatalf("workspace admin should have access once explicitly granted: %v", err)
	}
	if adminView.UserID != ownerUserID {
		t.Errorf("expected profile user %q, got %q", ownerUserID, adminView.UserID)
	}
}

func TestAT003_ZeroFabrication_Completeness(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-zero-fab-001"

	// Profile with only Contact Info filled
	contactPayload, _ := json.Marshal(ContactInfo{
		FullName: "Only Contact Person",
		Email:    "only@contact.com",
	})
	profile, err := service.UpdateProfileSection(ctx, userID, "contact", contactPayload, userID)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// AT-003: Unspecified fields must remain empty; no fake company, no fake degree
	if len(profile.Experiences) != 0 {
		t.Errorf("expected 0 experiences, got %d", len(profile.Experiences))
	}
	if len(profile.Education) != 0 {
		t.Errorf("expected 0 education items, got %d", len(profile.Education))
	}
	if len(profile.Skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(profile.Skills))
	}

	// Completeness: 1/8 populated = 12.5%
	if profile.Completeness.PopulatedCount != 1 {
		t.Errorf("expected 1 populated section, got %d", profile.Completeness.PopulatedCount)
	}
	if profile.Completeness.Score != 12.5 {
		t.Errorf("expected score 12.5, got %v", profile.Completeness.Score)
	}

	// Missing sections must list remaining 7 sections
	expectedMissing := []string{"links", "experience", "education", "skills", "projects", "certificates", "languages"}
	if len(profile.Completeness.MissingSections) != len(expectedMissing) {
		t.Errorf("expected %d missing sections, got %d (%v)", len(expectedMissing), len(profile.Completeness.MissingSections), profile.Completeness.MissingSections)
	}
}

func TestProfile_AggregateFromFacts(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-agg-001"

	// Simulate confirmed facts from previous upload or manual entry (AT-001)
	facts := []ProfileFactItem{
		{
			ID:               "fact-1",
			UserID:           userID,
			Category:         "contact",
			Title:            "John Doe",
			Details:          []string{"john@doe.com", "Phone: +1-555-0199"},
			Confidence:       0.95,
			Confirmed:        true,
			SourceProvenance: ProvenanceUploadPDF,
		},
		{
			ID:               "fact-2",
			UserID:           userID,
			Category:         "experience",
			Title:            "Backend Architect",
			Organization:     "FinTech Corp",
			Details:          []string{"Designed high-throughput ledger"},
			Confidence:       0.90,
			Confirmed:        true,
			SourceProvenance: ProvenanceUploadPDF,
		},
		{
			ID:               "fact-3",
			UserID:           userID,
			Category:         "skill",
			Title:            "Kubernetes",
			Confidence:       0.85,
			Confirmed:        true,
			SourceProvenance: ProvenanceUploadPDF,
		},
	}
	if err := repo.SaveFactsBatch(ctx, facts); err != nil {
		t.Fatalf("failed to save facts: %v", err)
	}

	// GetMasterProfile should auto-sync confirmed facts
	profile, err := service.GetMasterProfile(ctx, userID)
	if err != nil {
		t.Fatalf("failed to get profile: %v", err)
	}

	if profile.Contact.FullName != "John Doe" {
		t.Errorf("expected full name 'John Doe', got %q", profile.Contact.FullName)
	}
	if profile.Contact.Email != "john@doe.com" {
		t.Errorf("expected email 'john@doe.com', got %q", profile.Contact.Email)
	}
	if profile.Contact.Phone != "+1-555-0199" {
		t.Errorf("expected phone '+1-555-0199', got %q", profile.Contact.Phone)
	}
	if len(profile.Experiences) != 1 || profile.Experiences[0].Title != "Backend Architect" {
		t.Errorf("expected 1 experience 'Backend Architect', got: %v", profile.Experiences)
	}
	if len(profile.Skills) != 1 || profile.Skills[0].Name != "Kubernetes" {
		t.Errorf("expected 1 skill 'Kubernetes', got: %v", profile.Skills)
	}
}
