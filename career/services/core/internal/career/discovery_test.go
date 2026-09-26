package career

import (
	"context"
	"testing"
)

func TestNormalizeCanonicalURL_AT004(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Strip LinkedIn tracking params",
			input:    "https://www.linkedin.com/jobs/view/382910401/?trackingId=e9xK2%3D%3D&refId=feed_1&trk=flagship_job_search",
			expected: "https://www.linkedin.com/jobs/view/382910401",
		},
		{
			name:     "Strip Greenhouse utm params and anchors",
			input:    "https://boards.greenhouse.io/stripe/jobs/5492011?utm_source=linkedin&refId=992#app",
			expected: "https://boards.greenhouse.io/stripe/jobs/5492011",
		},
		{
			name:     "Strip Lever marketing origin query",
			input:    "https://jobs.lever.co/datadog/dd-89102-lever/?lever-origin=applied&utm_medium=jobboard",
			expected: "https://jobs.lever.co/datadog/dd-89102-lever",
		},
		{
			name:     "Handle clean URL with trailing slash removed",
			input:    "https://jobs.ashbyhq.com/linear/linear-eng-441/",
			expected: "https://jobs.ashbyhq.com/linear/linear-eng-441",
		},
		{
			name:     "Empty input",
			input:    "   ",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeCanonicalURL(tc.input)
			if got != tc.expected {
				t.Errorf("NormalizeCanonicalURL(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestDiscoveryEngine_PerBoardCapabilities_AT010(t *testing.T) {
	engine := NewDiscoveryEngine()
	caps := engine.GetCapabilitiesMatrix()

	if len(caps) != 5 {
		t.Fatalf("expected 5 board capabilities, got %d", len(caps))
	}

	capsMap := make(map[BoardSource]BoardCapabilities)
	for _, c := range caps {
		capsMap[c.BoardName] = c
	}

	// 1. Greenhouse verification
	gh, ok := capsMap[BoardSourceGreenhouse]
	if !ok {
		t.Fatal("missing Greenhouse capabilities")
	}
	if !gh.IsDirectATS {
		t.Error("Greenhouse must have IsDirectATS=true")
	}
	if gh.SupportsSalaryFilter {
		t.Error("Greenhouse public board does not support server-side salary filter (AT-010)")
	}

	// 2. Lever verification
	lever, ok := capsMap[BoardSourceLever]
	if !ok {
		t.Fatal("missing Lever capabilities")
	}
	if !lever.IsDirectATS {
		t.Error("Lever must have IsDirectATS=true")
	}

	// 3. Ashby verification
	ashby, ok := capsMap[BoardSourceAshby]
	if !ok {
		t.Fatal("missing Ashby capabilities")
	}
	if !ashby.IsDirectATS {
		t.Error("Ashby must have IsDirectATS=true")
	}
	if !ashby.SupportsSalaryFilter {
		t.Error("Ashby supports structured salary disclosure")
	}

	// 4. LinkedIn verification
	li, ok := capsMap[BoardSourceLinkedIn]
	if !ok {
		t.Fatal("missing LinkedIn capabilities")
	}
	if li.IsDirectATS {
		t.Error("LinkedIn is an aggregator, IsDirectATS must be false")
	}

	// 5. Indeed verification
	ind, ok := capsMap[BoardSourceIndeed]
	if !ok {
		t.Fatal("missing Indeed capabilities")
	}
	if ind.IsDirectATS {
		t.Error("Indeed is an aggregator, IsDirectATS must be false")
	}
}

func TestDiscoveryEngine_Deduplication_AT004(t *testing.T) {
	jobs := []DiscoveredJob{
		{
			ID:               "li-shopify-dup",
			Source:           BoardSourceLinkedIn,
			CanonicalURL:     "https://www.linkedin.com/jobs/view/99912/?refId=feed",
			Title:            "Backend Systems Engineer",
			Company:          "Shopify",
			IsDirectEmployer: false,
		},
		{
			ID:               "gh-shopify-direct",
			Source:           BoardSourceGreenhouse,
			CanonicalURL:     "https://boards.greenhouse.io/shopify/jobs/99912?utm_source=gh",
			Title:            "Backend Systems Engineer",
			Company:          "Shopify",
			IsDirectEmployer: true, // Direct employer ATS
		},
		{
			ID:               "ind-datadog",
			Source:           BoardSourceIndeed,
			CanonicalURL:     "https://www.indeed.com/viewjob?jk=dd123&from=vj",
			Title:            "Lead Site Reliability Engineer",
			Company:          "Datadog",
			IsDirectEmployer: false,
		},
	}

	deduped := DeduplicateDiscoveredJobs(jobs)

	// Invariant AT-004: 2 unique positions (Shopify and Datadog)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 unique jobs after deduplication, got %d", len(deduped))
	}

	// Verify that the direct employer (Greenhouse) superseded the aggregator (LinkedIn)
	var shopifyJob *DiscoveredJob
	for _, j := range deduped {
		if j.Company == "Shopify" {
			shopifyJob = &j
			break
		}
	}
	if shopifyJob == nil {
		t.Fatal("Shopify job missing from deduplicated results")
	}
	if !shopifyJob.IsDirectEmployer || shopifyJob.Source != BoardSourceGreenhouse {
		t.Errorf("expected direct Greenhouse job to supersede aggregator, got %s (direct=%v)",
			shopifyJob.Source, shopifyJob.IsDirectEmployer)
	}
}

func TestDiscoveryEngine_UndisclosedSalary_AT003_AT028(t *testing.T) {
	engine := NewDiscoveryEngine()
	ctx := context.Background()

	jobs, err := engine.SearchMultiBoard(ctx, JobSearchQuery{
		Keywords: "Datadog",
	})
	if err != nil {
		t.Fatalf("SearchMultiBoard failed: %v", err)
	}
	if len(jobs) == 0 {
		t.Fatal("expected at least one job for Datadog")
	}

	datadogJob := jobs[0]
	// Invariant AT-003: If employer didn't disclose salary, IsDisclosed must be false and never defaulted to 0
	if datadogJob.Compensation.IsDisclosed {
		t.Error("expected Datadog job compensation IsDisclosed=false")
	}
	if datadogJob.Compensation.MinAmount != 0 || datadogJob.Compensation.MaxAmount != 0 {
		t.Errorf("undisclosed salary should not have amounts invented: %f - %f",
			datadogJob.Compensation.MinAmount, datadogJob.Compensation.MaxAmount)
	}
}

func TestCareerService_MultiBoardSearch(t *testing.T) {
	repo := NewMemoryCareerRepository()
	service := NewCareerService(repo)
	ctx := context.Background()

	// 1. Search with remote filter
	jobs, err := service.DiscoverJobs(ctx, JobSearchQuery{
		RemoteOnly: true,
	})
	if err != nil {
		t.Fatalf("DiscoverJobs failed: %v", err)
	}
	if len(jobs) == 0 {
		t.Fatal("expected remote jobs to be discovered")
	}

	for _, j := range jobs {
		if !j.Location.IsRemote {
			t.Errorf("job %s should be remote", j.Title)
		}
	}

	// 2. Test Capabilities Matrix
	caps := service.GetJobBoardCapabilities(ctx)
	if len(caps) != 5 {
		t.Errorf("expected 5 board capabilities, got %d", len(caps))
	}

	// 3. Test URL normalization helper
	clean := service.NormalizeJobURL("https://jobs.lever.co/company/job-123/?utm_source=twitter&refId=12")
	if clean != "https://jobs.lever.co/company/job-123" {
		t.Errorf("unexpected normalized URL: %s", clean)
	}
}
