package career

import (
	"context"
	"testing"
	"time"
)

func TestSavedJobs_CRUD_And_Lifecycle(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-alice"

	// 1. Save a new job
	saveInput := SaveJobInput{
		Job: DiscoveredJob{
			ID:           "job-101",
			Title:        "Staff Go Engineer",
			Company:      "CloudScale Labs",
			CanonicalURL: "https://cloudscale.io/careers/staff-go",
			Location: JobLocation{
				City:        "Austin",
				State:       "TX",
				Country:     "US",
				IsRemote:    true,
				RawLocation: "Austin, TX (Remote)",
			},
			JobType:        "Full-time",
			RequiredSkills: []string{"Go", "Distributed Systems"},
			DiscoveredAt:   time.Now().UTC(),
		},
		Status:   SavedJobStatusSaved,
		Priority: 4,
		Notes:    "Looks like a great match for distributed systems experience",
		Tags:     []string{"golang", "distributed", "remote"},
	}

	saved, err := service.SaveJob(ctx, userID, saveInput)
	if err != nil {
		t.Fatalf("SaveJob failed: %v", err)
	}
	if saved.ID == "" {
		t.Errorf("expected generated ID, got empty")
	}
	if saved.Fingerprint == "" {
		t.Errorf("expected computed fingerprint, got empty")
	}
	if saved.Status != SavedJobStatusSaved {
		t.Errorf("expected status 'saved', got %s", saved.Status)
	}
	if saved.Priority != 4 {
		t.Errorf("expected priority 4, got %d", saved.Priority)
	}

	// 2. Get saved job by ID
	fetched, err := service.GetSavedJob(ctx, userID, saved.ID)
	if err != nil {
		t.Fatalf("GetSavedJob failed: %v", err)
	}
	if fetched.Title != "Staff Go Engineer" {
		t.Errorf("expected title 'Staff Go Engineer', got %s", fetched.Title)
	}

	// 3. Update status to shortlisted, priority and notes
	newStatus := SavedJobStatusShortlisted
	newNotes := "Screening scheduled next week"
	newPriority := 5
	updated, err := service.UpdateSavedJob(ctx, userID, saved.ID, UpdateSavedJobInput{
		Status:   &newStatus,
		Notes:    &newNotes,
		Priority: &newPriority,
		Tags:     []string{"golang", "interviewing"},
	})
	if err != nil {
		t.Fatalf("UpdateSavedJob failed: %v", err)
	}
	if updated.Status != SavedJobStatusShortlisted {
		t.Errorf("expected status 'shortlisted', got %s", updated.Status)
	}
	if updated.Priority != 5 {
		t.Errorf("expected priority 5, got %d", updated.Priority)
	}
	if len(updated.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(updated.Tags))
	}

	// 4. List saved jobs
	list, err := service.ListSavedJobs(ctx, userID, "", "")
	if err != nil {
		t.Fatalf("ListSavedJobs failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 saved job, got %d", len(list))
	}

	// Filter by status
	filtered, err := service.ListSavedJobs(ctx, userID, SavedJobStatusShortlisted, "")
	if err != nil {
		t.Fatalf("ListSavedJobs filtered failed: %v", err)
	}
	if len(filtered) != 1 {
		t.Errorf("expected 1 shortlisted job, got %d", len(filtered))
	}

	// Filter by tag
	tagFiltered, err := service.ListSavedJobs(ctx, userID, "", "golang")
	if err != nil {
		t.Fatalf("ListSavedJobs tag filtered failed: %v", err)
	}
	if len(tagFiltered) != 1 {
		t.Errorf("expected 1 tag-matched job, got %d", len(tagFiltered))
	}

	emptyFiltered, err := service.ListSavedJobs(ctx, userID, SavedJobStatusArchived, "")
	if err != nil {
		t.Fatalf("ListSavedJobs empty filter failed: %v", err)
	}
	if len(emptyFiltered) != 0 {
		t.Errorf("expected 0 archived jobs, got %d", len(emptyFiltered))
	}

	// 5. Delete saved job
	err = service.DeleteSavedJob(ctx, userID, saved.ID)
	if err != nil {
		t.Fatalf("DeleteSavedJob failed: %v", err)
	}

	remaining, err := service.ListSavedJobs(ctx, userID, "", "")
	if err != nil {
		t.Fatalf("ListSavedJobs after delete failed: %v", err)
	}
	if len(remaining) != 0 {
		t.Errorf("expected 0 saved jobs after delete, got %d", len(remaining))
	}
}

func TestDedupe_ExactCanonicalURL_And_Fingerprint(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-bob"

	// Save an initial job
	saveInput := SaveJobInput{
		Job: DiscoveredJob{
			ID:           "job-201",
			Title:        "Senior Backend Developer",
			Company:      "FinTech Global",
			CanonicalURL: "https://fintech.com/jobs/sr-backend?ref=linkedin",
			Location: JobLocation{
				City:        "New York",
				State:       "NY",
				Country:     "US",
				IsHybrid:    true,
				RawLocation: "New York, NY",
			},
		},
		Status: SavedJobStatusSaved,
	}
	_, err := service.SaveJob(ctx, userID, saveInput)
	if err != nil {
		t.Fatalf("SaveJob failed: %v", err)
	}

	// Case A: Exact canonical URL match
	resList1, err := service.CheckJobDeduplication(ctx, userID, []DiscoveredJob{
		{
			Title:        "Senior Backend Developer",
			Company:      "FinTech Global",
			CanonicalURL: "https://fintech.com/jobs/sr-backend?ref=linkedin",
			Location: JobLocation{
				City:        "New York",
				State:       "NY",
				Country:     "US",
				IsHybrid:    true,
				RawLocation: "New York, NY",
			},
		},
	})
	if err != nil {
		t.Fatalf("CheckJobDeduplication failed: %v", err)
	}
	if len(resList1) == 0 {
		t.Fatalf("expected at least 1 result")
	}
	res1 := resList1[0]
	if !res1.IsDuplicate {
		t.Errorf("expected job to be flagged as duplicate")
	}
	if len(res1.Matches) == 0 || res1.Matches[0].MatchLevel != DedupeMatchExact {
		t.Errorf("expected match level exact")
	}
	if res1.ActionTaken != "linked" {
		t.Errorf("expected action_taken 'linked', got %s", res1.ActionTaken)
	}

	// Case B: Different URL, but exact same Title + Company + RawLocation -> exact fingerprint match
	resList2, err := service.CheckJobDeduplication(ctx, userID, []DiscoveredJob{
		{
			Title:        "Senior Backend Developer",
			Company:      "FinTech Global",
			CanonicalURL: "https://fintech.com/jobs/sr-backend-alternate-board",
			Location: JobLocation{
				City:        "New York",
				State:       "NY",
				Country:     "US",
				IsHybrid:    true,
				RawLocation: "New York, NY",
			},
		},
	})
	if err != nil {
		t.Fatalf("CheckJobDeduplication failed: %v", err)
	}
	if len(resList2) == 0 {
		t.Fatalf("expected at least 1 result")
	}
	res2 := resList2[0]
	if !res2.IsDuplicate {
		t.Errorf("expected fingerprint match to be flagged as duplicate")
	}
	if len(res2.Matches) == 0 || res2.Matches[0].MatchLevel != DedupeMatchExact {
		t.Errorf("expected match level exact for same title/company fingerprint")
	}
}

func TestDedupe_AmbiguousMatch_TextSimilarity_AT004(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-carol"

	// Save "Lead Distributed Systems Engineer" at "Acme Cloud"
	saveInput := SaveJobInput{
		Job: DiscoveredJob{
			ID:           "job-301",
			Title:        "Lead Distributed Systems Engineer",
			Company:      "Acme Cloud",
			CanonicalURL: "https://acme.com/careers/lead-distributed-systems",
			Location: JobLocation{
				City:        "San Francisco",
				State:       "CA",
				Country:     "US",
				IsRemote:    true,
				RawLocation: "San Francisco, CA (Remote)",
			},
		},
		Status: SavedJobStatusSaved,
	}
	_, err := service.SaveJob(ctx, userID, saveInput)
	if err != nil {
		t.Fatalf("SaveJob failed: %v", err)
	}

	// New candidate job: "Lead Distributed Systems Software Engineer" at "Acme Cloud"
	// Words in A: {lead, distributed, systems, engineer} (size 4)
	// Words in B: {lead, distributed, systems, software, engineer} (size 5)
	// Intersection: 4, Union: 5 -> Similarity: 4/5 = 0.80 >= 0.75 threshold
	resList, err := service.CheckJobDeduplication(ctx, userID, []DiscoveredJob{
		{
			Title:        "Lead Distributed Systems Software Engineer",
			Company:      "Acme Cloud",
			CanonicalURL: "https://otherjobboard.com/acme/lead-systems",
			Location: JobLocation{
				City:        "San Francisco",
				State:       "CA",
				Country:     "US",
				IsRemote:    true,
				RawLocation: "San Francisco, CA (Remote)",
			},
		},
	})
	if err != nil {
		t.Fatalf("CheckJobDeduplication failed: %v", err)
	}
	if len(resList) == 0 {
		t.Fatalf("expected results")
	}
	res := resList[0]

	if !res.IsAmbiguous {
		t.Errorf("expected ambiguous match to be flagged with IsAmbiguous=true")
	}
	if len(res.Matches) == 0 || res.Matches[0].MatchLevel != DedupeMatchAmbiguous {
		t.Errorf("expected match level ambiguous")
	}
	if !res.Matches[0].IsReviewable {
		t.Errorf("expected AT-004 reviewable flag to be true for ambiguous match")
	}
	if res.Matches[0].SimilarityScore < 0.75 {
		t.Errorf("expected similarity score >= 0.75, got %f", res.Matches[0].SimilarityScore)
	}
	if res.ActionTaken != "flagged_for_review" {
		t.Errorf("expected action_taken 'flagged_for_review', got %s", res.ActionTaken)
	}

	// Sub-threshold title: "Junior Python Developer" at "Acme Cloud"
	// Similarity is 0.0 -> Should NOT be flagged as duplicate
	resList2, err := service.CheckJobDeduplication(ctx, userID, []DiscoveredJob{
		{
			Title:        "Junior Python Developer",
			Company:      "Acme Cloud",
			CanonicalURL: "https://otherjobboard.com/acme/jr-python",
		},
	})
	if err != nil {
		t.Fatalf("CheckJobDeduplication failed: %v", err)
	}
	if len(resList2) == 0 {
		t.Fatalf("expected results")
	}
	res2 := resList2[0]
	if res2.IsDuplicate {
		t.Errorf("expected non-matching title to not be duplicate")
	}
	if res2.ActionTaken != "retained" {
		t.Errorf("expected action_taken 'retained', got %s", res2.ActionTaken)
	}
}

func TestDedupe_SeparateDiscoveryVsApplication_CAR10(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-david"

	targetJob := DiscoveredJob{
		ID:           "job-401",
		Title:        "Staff Infrastructure Engineer",
		Company:      "ScaleInfra Corp",
		CanonicalURL: "https://scaleinfra.com/jobs/staff-infra",
		Location: JobLocation{
			City:        "Seattle",
			State:       "WA",
			Country:     "US",
			IsRemote:    true,
			RawLocation: "Seattle, WA",
		},
	}

	// 1. Candidate saves job initially
	saved, err := service.SaveJob(ctx, userID, SaveJobInput{
		Job:    targetJob,
		Status: SavedJobStatusSaved,
	})
	if err != nil {
		t.Fatalf("SaveJob failed: %v", err)
	}

	// 2. Candidate records application submission via ApplicationRecord (AT-004 / AT-006 / CAR-10)
	appRecord, err := service.RecordApplication(ctx, userID, RecordApplicationInput{
		Job:            targetJob,
		Status:         "submitted",
		SubmissionMode: "manual_browser",
		Notes:          "Received confirmation email ref #A9281",
	})
	if err != nil {
		t.Fatalf("RecordApplication failed: %v", err)
	}
	if appRecord.ID == "" {
		t.Errorf("expected application record ID, got empty")
	}

	// 3. Verify SavedJob status automatically transitioned to 'applied'
	savedUpdated, err := service.GetSavedJob(ctx, userID, saved.ID)
	if err != nil {
		t.Fatalf("GetSavedJob failed: %v", err)
	}
	if savedUpdated.Status != SavedJobStatusApplied {
		t.Errorf("expected saved job status to be 'applied', got %s", savedUpdated.Status)
	}

	// 4. Verify separate deduplication: Application dedupe prevents applying twice!
	resList, err := service.CheckJobDeduplication(ctx, userID, []DiscoveredJob{targetJob})
	if err != nil {
		t.Fatalf("CheckJobDeduplication failed: %v", err)
	}
	if len(resList) == 0 {
		t.Fatalf("expected results")
	}
	res := resList[0]
	if !res.IsDuplicate {
		t.Errorf("expected job to be flagged as duplicate due to prior application")
	}
	foundAppMatch := false
	for _, m := range res.Matches {
		if m.TargetLedger == "applications" && m.MatchLevel == DedupeMatchExact {
			foundAppMatch = true
			break
		}
	}
	if !foundAppMatch {
		t.Errorf("expected application exact match in dedupe result")
	}

	// 5. List applications
	apps, err := service.ListApplications(ctx, userID)
	if err != nil {
		t.Fatalf("ListApplications failed: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("expected 1 application record, got %d", len(apps))
	}
	if apps[0].Company != "ScaleInfra Corp" {
		t.Errorf("expected company ScaleInfra Corp, got %s", apps[0].Company)
	}
}

func TestExclusions_PersistentBlocklist_NeverSeeAgain(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()
	userID := "user-emma"

	// 1. Add company exclusion ("Never see jobs from BadCorp again" - CAR-10 / C8)
	compExcl, err := service.AddExclusion(ctx, userID, AddExclusionInput{
		Type:        ExclusionTypeCompanyName,
		Value:       "BadCorp Industries",
		Reason:      "Poor work-life balance feedback",
		CompanyName: "BadCorp Industries",
	})
	if err != nil {
		t.Fatalf("AddExclusion company failed: %v", err)
	}
	if compExcl.ID == "" {
		t.Errorf("expected exclusion ID, got empty")
	}

	// 2. Add specific canonical URL exclusion
	urlExcl, err := service.AddExclusion(ctx, userID, AddExclusionInput{
		Type:   ExclusionTypeJobCanonicalURL,
		Value:  "https://spammyrecruit.com/job/12345",
		Reason: "Third-party agency spam",
	})
	if err != nil {
		t.Fatalf("AddExclusion url failed: %v", err)
	}
	if urlExcl.ID == "" {
		t.Errorf("expected url exclusion ID, got empty")
	}

	// 3. Test company exclusion filters out jobs from BadCorp
	resListComp, err := service.CheckJobDeduplication(ctx, userID, []DiscoveredJob{
		{
			Title:        "Senior SRE",
			Company:      "BadCorp Industries",
			CanonicalURL: "https://badcorp.com/careers/sre",
		},
	})
	if err != nil {
		t.Fatalf("CheckJobDeduplication failed: %v", err)
	}
	if len(resListComp) == 0 {
		t.Fatalf("expected result")
	}
	resComp := resListComp[0]
	if !resComp.IsExcluded {
		t.Errorf("expected excluded company job to have IsExcluded=true")
	}
	if resComp.ActionTaken != "filtered_out" {
		t.Errorf("expected action_taken 'filtered_out', got %s", resComp.ActionTaken)
	}

	// 4. Test canonical URL exclusion filters out spam URL
	resListURL, err := service.CheckJobDeduplication(ctx, userID, []DiscoveredJob{
		{
			Title:        "Frontend Lead",
			Company:      "GreatCorp",
			CanonicalURL: "https://spammyrecruit.com/job/12345",
		},
	})
	if err != nil {
		t.Fatalf("CheckJobDeduplication failed: %v", err)
	}
	if len(resListURL) == 0 {
		t.Fatalf("expected result")
	}
	resURL := resListURL[0]
	if !resURL.IsExcluded {
		t.Errorf("expected excluded URL to have IsExcluded=true")
	}
	if resURL.ActionTaken != "filtered_out" {
		t.Errorf("expected action_taken 'filtered_out', got %s", resURL.ActionTaken)
	}

	// 5. Test legitimate job is eligible (retained)
	resListGood, err := service.CheckJobDeduplication(ctx, userID, []DiscoveredJob{
		{
			Title:        "Backend Architect",
			Company:      "AwesomeCorp",
			CanonicalURL: "https://awesomecorp.io/careers/architect",
		},
	})
	if err != nil {
		t.Fatalf("CheckJobDeduplication failed: %v", err)
	}
	if len(resListGood) == 0 {
		t.Fatalf("expected result")
	}
	resGood := resListGood[0]
	if resGood.IsExcluded || resGood.IsDuplicate {
		t.Errorf("expected eligible job to not be excluded or duplicate")
	}
	if resGood.ActionTaken != "retained" {
		t.Errorf("expected action_taken 'retained', got %s", resGood.ActionTaken)
	}

	// 6. List and delete exclusions
	exclList, err := service.ListExclusions(ctx, userID)
	if err != nil {
		t.Fatalf("ListExclusions failed: %v", err)
	}
	if len(exclList) != 2 {
		t.Errorf("expected 2 exclusions, got %d", len(exclList))
	}

	err = service.DeleteExclusion(ctx, userID, compExcl.ID)
	if err != nil {
		t.Fatalf("DeleteExclusion failed: %v", err)
	}
	remainingExcl, err := service.ListExclusions(ctx, userID)
	if err != nil {
		t.Fatalf("ListExclusions after delete failed: %v", err)
	}
	if len(remainingExcl) != 1 {
		t.Errorf("expected 1 exclusion remaining, got %d", len(remainingExcl))
	}
}

func TestDedupe_OwnerIsolation_REQ018(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	commonJob := DiscoveredJob{
		ID:           "job-501",
		Title:        "VP of Engineering",
		Company:      "Metaverse Inc",
		CanonicalURL: "https://metaverse.com/careers/vpe",
	}

	// User 1 saves a job
	_, err := service.SaveJob(ctx, "user-1", SaveJobInput{
		Job:    commonJob,
		Status: SavedJobStatusSaved,
	})
	if err != nil {
		t.Fatalf("SaveJob user-1 failed: %v", err)
	}

	// User 2 saves same job - should succeed independently (owner isolated per REQ-018)
	u2Saved, err := service.SaveJob(ctx, "user-2", SaveJobInput{
		Job:    commonJob,
		Status: SavedJobStatusShortlisted,
	})
	if err != nil {
		t.Fatalf("SaveJob user-2 failed: %v", err)
	}
	if u2Saved.UserID != "user-2" {
		t.Errorf("expected UserID user-2, got %s", u2Saved.UserID)
	}

	// User 1's list should have 1 item
	u1List, err := service.ListSavedJobs(ctx, "user-1", "", "")
	if err != nil {
		t.Fatalf("ListSavedJobs user-1 failed: %v", err)
	}
	if len(u1List) != 1 || u1List[0].UserID != "user-1" {
		t.Errorf("user-1 isolation violated")
	}

	// User 2's dedupe check should not see user-1's applications or exclusions
	resListU2, err := service.CheckJobDeduplication(ctx, "user-2", []DiscoveredJob{
		{
			Title:        "Different Role",
			Company:      "Different Co",
			CanonicalURL: "https://different.com",
		},
	})
	if err != nil {
		t.Fatalf("CheckJobDeduplication user-2 failed: %v", err)
	}
	if len(resListU2) == 0 {
		t.Fatalf("expected result")
	}
	resU2 := resListU2[0]
	if resU2.IsDuplicate {
		t.Errorf("user-2 should have no false dedupe triggers from user-1")
	}
}
