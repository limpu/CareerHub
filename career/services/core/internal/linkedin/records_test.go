package linkedin

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type recordsFixtureData struct {
	Workspaces map[string]struct {
		Persons   []PersonRecord  `json:"persons"`
		Companies []CompanyRecord `json:"companies"`
		Jobs      []JobRecord     `json:"jobs"`
		Posts     []PostRecord    `json:"posts"`
	} `json:"workspaces"`
}

func loadRecordsFixture(t *testing.T) *recordsFixtureData {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_records.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read linkedin_records.json fixture: %v", err)
	}

	var fix recordsFixtureData
	if err := json.Unmarshal(bytes, &fix); err != nil {
		t.Fatalf("failed to parse linkedin_records.json fixture: %v", err)
	}
	return &fix
}

func TestLinkedInRecords_FixtureLoading(t *testing.T) {
	fix := loadRecordsFixture(t)
	if _, ok := fix.Workspaces["ws-alpha"]; !ok {
		t.Fatalf("expected ws-alpha in fixture workspaces")
	}
	if _, ok := fix.Workspaces["ws-beta"]; !ok {
		t.Fatalf("expected ws-beta in fixture workspaces")
	}

	alpha := fix.Workspaces["ws-alpha"]
	if len(alpha.Persons) == 0 {
		t.Errorf("expected persons in ws-alpha")
	}
	if len(alpha.Companies) == 0 {
		t.Errorf("expected companies in ws-alpha")
	}
	if len(alpha.Jobs) == 0 {
		t.Errorf("expected jobs in ws-alpha")
	}
	if len(alpha.Posts) == 0 {
		t.Errorf("expected posts in ws-alpha")
	}

	if alpha.Persons[0].FullName != "Sarah Chen" {
		t.Errorf("expected Sarah Chen, got %s", alpha.Persons[0].FullName)
	}
	if alpha.Companies[0].CompanyName != "CloudScale Systems" {
		t.Errorf("expected CloudScale Systems, got %s", alpha.Companies[0].CompanyName)
	}
	if alpha.Jobs[0].Title != "Principal Systems Engineer" {
		t.Errorf("expected Principal Systems Engineer, got %s", alpha.Jobs[0].Title)
	}
	if alpha.Posts[0].URN != "urn:li:activity:7283940182" {
		t.Errorf("expected urn:li:activity:7283940182, got %s", alpha.Posts[0].URN)
	}
}

func TestLinkedInRecords_CRUD_And_Versioning(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	// 1. Person Record Upsert & Version Increment
	person := &PersonRecord{
		RecordID:         "rec-p-01",
		WorkspaceID:      "ws-alpha",
		TenantID:         "tenant-alpha",
		VanitySlug:       "sarah-chen-arch",
		FullName:         "Sarah Chen",
		Headline:         "Lead Systems Architect @ CloudScale",
		CurrentCompany:   "CloudScale Systems",
		CurrentRole:      "Lead Systems Architect",
		Location:         "San Francisco Bay Area",
		ConnectionTier:   "1st",
		PublicProfileURL: "https://linkedin.com/in/sarah-chen-arch",
		Skills:           []string{"Go", "Distributed Systems"},
		ObservedAt:       time.Now().UTC(),
	}

	if err := svc.UpsertPersonRecord(ctx, person); err != nil {
		t.Fatalf("failed to upsert person record: %v", err)
	}

	fetchedP, err := svc.GetPersonRecord(ctx, "rec-p-01", "tenant-alpha")
	if err != nil {
		t.Fatalf("failed to get person record: %v", err)
	}
	if fetchedP.Version != 1 {
		t.Errorf("expected version 1, got %d", fetchedP.Version)
	}
	if fetchedP.FullName != "Sarah Chen" {
		t.Errorf("expected Sarah Chen, got %s", fetchedP.FullName)
	}

	// Update headline and upsert again - version should bump to 2
	person.Headline = "VP of Architecture @ CloudScale"
	if err := svc.UpsertPersonRecord(ctx, person); err != nil {
		t.Fatalf("failed to update person record: %v", err)
	}

	fetchedP2, err := svc.GetPersonRecord(ctx, "rec-p-01", "tenant-alpha")
	if err != nil {
		t.Fatalf("failed to get updated person record: %v", err)
	}
	if fetchedP2.Version != 2 {
		t.Errorf("expected version 2 after second upsert, got %d", fetchedP2.Version)
	}
	if fetchedP2.Headline != "VP of Architecture @ CloudScale" {
		t.Errorf("expected headline updated, got %s", fetchedP2.Headline)
	}

	// 2. Company Record
	company := &CompanyRecord{
		RecordID:      "rec-c-01",
		WorkspaceID:   "ws-alpha",
		TenantID:      "tenant-alpha",
		UniversalName: "cloudscale-systems",
		CompanyName:   "CloudScale Systems",
		Domain:        "cloudscale.tech",
		Industry:      "Software Development",
		SizeTier:      "201-500 employees",
		Specialties:   []string{"Cloud Native", "Observability"},
		Verified:      true,
		ObservedAt:    time.Now().UTC(),
	}
	if err := svc.UpsertCompanyRecord(ctx, company); err != nil {
		t.Fatalf("failed to upsert company record: %v", err)
	}

	fetchedC, err := svc.GetCompanyRecord(ctx, "rec-c-01", "tenant-alpha")
	if err != nil {
		t.Fatalf("failed to get company record: %v", err)
	}
	if fetchedC.Version != 1 {
		t.Errorf("expected company version 1, got %d", fetchedC.Version)
	}
	if fetchedC.CompanyName != "CloudScale Systems" {
		t.Errorf("expected CloudScale Systems, got %s", fetchedC.CompanyName)
	}

	// 3. Job Record
	job := &JobRecord{
		RecordID:           "rec-j-01",
		WorkspaceID:        "ws-alpha",
		TenantID:           "tenant-alpha",
		JobID:              "job-10101",
		Title:              "Principal Systems Engineer",
		CompanyName:        "CloudScale Systems",
		WorkplaceType:      "Remote",
		ApplicantCount:     34,
		PostedAt:           time.Now().UTC(),
		DescriptionSnippet: "Leading distributed systems team",
		ObservedAt:         time.Now().UTC(),
	}
	if err := svc.UpsertJobRecord(ctx, job); err != nil {
		t.Fatalf("failed to upsert job record: %v", err)
	}

	fetchedJ, err := svc.GetJobRecord(ctx, "rec-j-01", "tenant-alpha")
	if err != nil {
		t.Fatalf("failed to get job record: %v", err)
	}
	if fetchedJ.Version != 1 {
		t.Errorf("expected job version 1, got %d", fetchedJ.Version)
	}
	if fetchedJ.Title != "Principal Systems Engineer" {
		t.Errorf("expected Principal Systems Engineer, got %s", fetchedJ.Title)
	}

	// 4. Post Record
	post := &PostRecord{
		RecordID:       "rec-post-01",
		WorkspaceID:    "ws-alpha",
		TenantID:       "tenant-alpha",
		URN:            "urn:li:activity:7283940182",
		AuthorName:     "Sarah Chen",
		Commentary:     "Sub-millisecond p99 latency achieved!",
		MediaType:      "article",
		ReactionsCount: 340,
		PublishedAt:    time.Now().UTC(),
		ObservedAt:     time.Now().UTC(),
	}
	if err := svc.UpsertPostRecord(ctx, post); err != nil {
		t.Fatalf("failed to upsert post record: %v", err)
	}

	fetchedPost, err := svc.GetPostRecord(ctx, "rec-post-01", "tenant-alpha")
	if err != nil {
		t.Fatalf("failed to get post record: %v", err)
	}
	if fetchedPost.Version != 1 {
		t.Errorf("expected post version 1, got %d", fetchedPost.Version)
	}
	if fetchedPost.ReactionsCount != 340 {
		t.Errorf("expected 340 reactions, got %d", fetchedPost.ReactionsCount)
	}
}

func TestLinkedInRecords_CrossTenantIsolation_AT012(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	fix := loadRecordsFixture(t)

	// Populate ws-alpha records
	for _, p := range fix.Workspaces["ws-alpha"].Persons {
		pCopy := p
		if err := svc.UpsertPersonRecord(ctx, &pCopy); err != nil {
			t.Fatalf("failed to upsert ws-alpha person: %v", err)
		}
	}
	for _, c := range fix.Workspaces["ws-alpha"].Companies {
		cCopy := c
		if err := svc.UpsertCompanyRecord(ctx, &cCopy); err != nil {
			t.Fatalf("failed to upsert ws-alpha company: %v", err)
		}
	}
	for _, j := range fix.Workspaces["ws-alpha"].Jobs {
		jCopy := j
		if err := svc.UpsertJobRecord(ctx, &jCopy); err != nil {
			t.Fatalf("failed to upsert ws-alpha job: %v", err)
		}
	}
	for _, post := range fix.Workspaces["ws-alpha"].Posts {
		postCopy := post
		if err := svc.UpsertPostRecord(ctx, &postCopy); err != nil {
			t.Fatalf("failed to upsert ws-alpha post: %v", err)
		}
	}

	// Populate ws-beta records
	for _, p := range fix.Workspaces["ws-beta"].Persons {
		pCopy := p
		if err := svc.UpsertPersonRecord(ctx, &pCopy); err != nil {
			t.Fatalf("failed to upsert ws-beta person: %v", err)
		}
	}

	// 1. Cross-tenant individual GET access test
	// tenant-beta tries to get tenant-alpha's person record
	_, err := svc.GetPersonRecord(ctx, "rec-p-01", "tenant-beta")
	if !errors.Is(err, ErrCrossTenantAccessDenied) {
		t.Fatalf("expected ErrCrossTenantAccessDenied when tenant-beta accesses tenant-alpha record, got: %v", err)
	}

	// 2. Listing scoped to tenant-alpha
	alphaPersons, err := svc.ListPersonRecords(ctx, RecordFilter{
		WorkspaceID: "ws-alpha",
		TenantID:    "tenant-alpha",
	})
	if err != nil {
		t.Fatalf("failed to list tenant-alpha persons: %v", err)
	}
	if len(alphaPersons) != 2 {
		t.Errorf("expected 2 persons for tenant-alpha, got %d", len(alphaPersons))
	}
	for _, p := range alphaPersons {
		if p.TenantID != "tenant-alpha" {
			t.Errorf("leakage detected: person %s has tenant %s in tenant-alpha query", p.RecordID, p.TenantID)
		}
	}

	// 3. Listing scoped to tenant-beta
	betaPersons, err := svc.ListPersonRecords(ctx, RecordFilter{
		WorkspaceID: "ws-beta",
		TenantID:    "tenant-beta",
	})
	if err != nil {
		t.Fatalf("failed to list tenant-beta persons: %v", err)
	}
	if len(betaPersons) != 1 {
		t.Fatalf("expected 1 person for tenant-beta, got %d", len(betaPersons))
	}
	if betaPersons[0].RecordID != "rec-p-99" || betaPersons[0].FullName != "Elena Rostova" {
		t.Errorf("unexpected record for tenant-beta: %+v", betaPersons[0])
	}
}

func TestLinkedInRecords_ForbiddenBiometric_Validation_FND009(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	// Prohibited biometric field in person
	badPerson := &PersonRecord{
		RecordID:    "rec-bad-01",
		WorkspaceID: "ws-alpha",
		TenantID:    "tenant-alpha",
		FullName:    "Hacker X",
		Headline:    "Contains biometric_iris_scan data",
	}

	err := svc.UpsertPersonRecord(ctx, badPerson)
	if !errors.Is(err, ErrForbiddenBiometricField) {
		t.Fatalf("expected ErrForbiddenBiometricField for iris scan, got %v", err)
	}

	// Prohibited sensitive SSN field
	badPerson2 := &PersonRecord{
		RecordID:    "rec-bad-02",
		WorkspaceID: "ws-alpha",
		TenantID:    "tenant-alpha",
		FullName:    "Hacker Y",
		Headline:    "Developer",
		Location:    "social_security number 123-45-6789",
	}
	err = svc.UpsertPersonRecord(ctx, badPerson2)
	if !errors.Is(err, ErrForbiddenBiometricField) {
		t.Fatalf("expected ErrForbiddenBiometricField for SSN, got %v", err)
	}
}
