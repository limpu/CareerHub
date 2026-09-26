package linkedin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type recruiterWorkspaceFixture struct {
	ScenarioTitle string `json:"scenario_title"`
	Workspaces    []struct {
		WorkspaceID string          `json:"workspace_id"`
		TenantID    string          `json:"tenant_id"`
		Leads       []RecruiterLead `json:"leads"`
	} `json:"workspaces"`
}

func loadRecruiterFixture(t *testing.T) *recruiterWorkspaceFixture {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_recruiter_workspace.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read recruiter workspace fixture at %s: %v", path, err)
	}
	var fix recruiterWorkspaceFixture
	if err := json.Unmarshal(data, &fix); err != nil {
		t.Fatalf("failed to unmarshal recruiter workspace fixture: %v", err)
	}
	return &fix
}

func TestRecruiterLead_FixtureLoading(t *testing.T) {
	fix := loadRecruiterFixture(t)
	if len(fix.Workspaces) < 2 {
		t.Fatalf("expected at least 2 workspaces in fixture, got %d", len(fix.Workspaces))
	}

	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	// Seed leads from fixture
	for _, ws := range fix.Workspaces {
		for _, lead := range ws.Leads {
			created, err := svc.CreateRecruiterLead(ctx, &lead)
			if err != nil {
				t.Fatalf("failed to create lead %s: %v", lead.ID, err)
			}
			if created.RecruiterName != lead.RecruiterName {
				t.Errorf("expected recruiter name %q, got %q", lead.RecruiterName, created.RecruiterName)
			}
		}
	}

	// Verify ws-alpha leads count
	alphaLeads, err := svc.ListRecruiterLeads(ctx, RecruiterLeadFilter{
		WorkspaceID: "ws-alpha",
		TenantID:    "tenant-alpha",
	})
	if err != nil {
		t.Fatalf("failed to list alpha leads: %v", err)
	}
	if len(alphaLeads) != 2 {
		t.Errorf("expected 2 leads for ws-alpha, got %d", len(alphaLeads))
	}

	// Verify ws-beta leads count
	betaLeads, err := svc.ListRecruiterLeads(ctx, RecruiterLeadFilter{
		WorkspaceID: "ws-beta",
		TenantID:    "tenant-beta",
	})
	if err != nil {
		t.Fatalf("failed to list beta leads: %v", err)
	}
	if len(betaLeads) != 1 {
		t.Errorf("expected 1 lead for ws-beta, got %d", len(betaLeads))
	}
}

func TestRecruiterLead_Lifecycle(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	lead := &RecruiterLead{
		WorkspaceID:    "ws-alpha",
		TenantID:       "tenant-alpha",
		RecruiterName:  "Rachel Zane",
		RecruiterTitle: "Senior Tech Recruiter",
		Company:        "Pearson Hardman",
		LinkedInURL:    "https://www.linkedin.com/in/rachel-zane",
		Status:         LeadStatusNew,
	}

	created, err := svc.CreateRecruiterLead(ctx, lead)
	if err != nil {
		t.Fatalf("failed to create recruiter lead: %v", err)
	}

	// 1. Add Note
	note, err := svc.AddLeadNote(ctx, created.ID, "tenant-alpha", "candidate-1", "Met at conference; interested in Go backend platform roles.")
	if err != nil {
		t.Fatalf("failed to add note: %v", err)
	}
	if note.Content == "" {
		t.Errorf("expected non-empty note content")
	}

	// 2. Schedule Reminder
	remDate := time.Now().Add(48 * time.Hour)
	rem, err := svc.SetLeadReminder(ctx, created.ID, "tenant-alpha", remDate, "Send technical architecture deck")
	if err != nil {
		t.Fatalf("failed to set reminder: %v", err)
	}
	if rem.Status != ReminderPending {
		t.Errorf("expected reminder status pending, got %s", rem.Status)
	}

	// 3. Complete Reminder
	if err := svc.CompleteLeadReminder(ctx, created.ID, "tenant-alpha"); err != nil {
		t.Fatalf("failed to complete reminder: %v", err)
	}

	// 4. Update Status & Link Application
	updated, err := svc.UpdateLeadStatus(ctx, created.ID, "tenant-alpha", LeadStatusInDialogue, OutreachContacted)
	if err != nil {
		t.Fatalf("failed to update lead status: %v", err)
	}
	if updated.Status != LeadStatusInDialogue || updated.OutreachStage != OutreachContacted {
		t.Errorf("expected status %s and outreach %s, got %s / %s",
			LeadStatusInDialogue, OutreachContacted, updated.Status, updated.OutreachStage)
	}

	linked, err := svc.LinkLeadApplication(ctx, created.ID, "tenant-alpha", "job-101", "app-999")
	if err != nil {
		t.Fatalf("failed to link application: %v", err)
	}
	if linked.RelatedJobID != "job-101" || linked.RelatedApplicationID != "app-999" {
		t.Errorf("expected linked job-101 and app-999, got %s / %s", linked.RelatedJobID, linked.RelatedApplicationID)
	}

	// 5. Generate Outreach Draft (AT-010: no fake live send)
	draft := linked.GenerateOutreachDraft("Alex Chen", "Principal Infrastructure Architect")
	if !draft.WithinLimit {
		t.Errorf("expected outreach draft to fit within LinkedIn 300-char connection limit")
	}
	if draft.DirectChatURL != "https://www.linkedin.com/in/rachel-zane/" {
		t.Errorf("expected correct direct chat URL, got %s", draft.DirectChatURL)
	}
}

func TestRecruiterLead_TenantIsolation(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	alphaLead := &RecruiterLead{
		ID:            "lead-alpha-secure",
		WorkspaceID:   "ws-alpha",
		TenantID:      "tenant-alpha",
		RecruiterName: "Harvey Specter",
		Company:       "Specter Litt",
		LinkedInURL:   "https://www.linkedin.com/in/harvey-specter",
	}
	if _, err := svc.CreateRecruiterLead(ctx, alphaLead); err != nil {
		t.Fatalf("failed to create alpha lead: %v", err)
	}

	// Attempt access from tenant-beta must fail closed (AT-011)
	_, err := svc.GetRecruiterLead(ctx, "lead-alpha-secure", "tenant-beta")
	if err == nil {
		t.Fatalf("expected cross-tenant access violation error, got nil")
	}

	// Attempt note addition from tenant-beta must fail
	_, err = svc.AddLeadNote(ctx, "lead-alpha-secure", "tenant-beta", "attacker", "hacked")
	if err == nil {
		t.Fatalf("expected cross-tenant note error, got nil")
	}

	// Attempt reminder update from tenant-beta must fail
	_, err = svc.SetLeadReminder(ctx, "lead-alpha-secure", "tenant-beta", time.Now().Add(time.Hour), "reminder")
	if err == nil {
		t.Fatalf("expected cross-tenant reminder error, got nil")
	}

	// List from tenant-beta should not return alpha lead
	betaList, err := svc.ListRecruiterLeads(ctx, RecruiterLeadFilter{
		WorkspaceID: "ws-beta",
		TenantID:    "tenant-beta",
	})
	if err != nil {
		t.Fatalf("failed to list beta leads: %v", err)
	}
	for _, l := range betaList {
		if l.ID == "lead-alpha-secure" {
			t.Errorf("cross-tenant data leakage detected: alpha lead found in beta list")
		}
	}
}

func TestRecruiterLead_SourceLinkage(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	// 1. Create a person record
	person := &PersonRecord{
		RecordID:         "person-source-001",
		WorkspaceID:      "ws-alpha",
		TenantID:         "tenant-alpha",
		FullName:         "Donna Paulsen",
		Headline:         "Executive Talent Partner",
		CurrentCompany:   "Pearson Specter",
		PublicProfileURL: "https://www.linkedin.com/in/donna-paulsen",
	}
	if err := repo.SavePersonRecord(ctx, person); err != nil {
		t.Fatalf("failed to save person record: %v", err)
	}

	// Create lead from person record
	lead, err := svc.CreateLeadFromPersonRecord(ctx, "person-source-001", "ws-alpha", "tenant-alpha")
	if err != nil {
		t.Fatalf("failed to create lead from person record: %v", err)
	}
	if lead.RecruiterName != "Donna Paulsen" || lead.SourceEntityType != "person_record" {
		t.Errorf("expected lead Donna Paulsen with person_record source, got %s / %s",
			lead.RecruiterName, lead.SourceEntityType)
	}
	if len(lead.Notes) == 0 {
		t.Errorf("expected initial system audit note on lead creation")
	}

	// 2. Create a post record
	post := &PostRecord{
		RecordID:     "post-source-001",
		WorkspaceID:  "ws-alpha",
		TenantID:     "tenant-alpha",
		AuthorName:   "Mike Ross",
		AuthorVanity: "mike-ross-legal",
		URN:          "urn:li:activity:998877",
		Commentary:   "We are hiring Senior Associates and Systems Architects in New York!",
	}
	if err := repo.SavePostRecord(ctx, post); err != nil {
		t.Fatalf("failed to save post record: %v", err)
	}

	// Create lead from post record
	postLead, err := svc.CreateLeadFromPostRecord(ctx, "post-source-001", "ws-alpha", "tenant-alpha")
	if err != nil {
		t.Fatalf("failed to create lead from post record: %v", err)
	}
	if postLead.RecruiterName != "Mike Ross" || postLead.SourceEntityType != "post_record" {
		t.Errorf("expected lead Mike Ross with post_record source, got %s / %s",
			postLead.RecruiterName, postLead.SourceEntityType)
	}
}
