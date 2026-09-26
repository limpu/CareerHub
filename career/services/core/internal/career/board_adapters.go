package career

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ----------------------------------------------------------------------
// 1. Greenhouse Adapter (Direct ATS)
// ----------------------------------------------------------------------

type GreenhouseAdapter struct{}

func NewGreenhouseAdapter() *GreenhouseAdapter {
	return &GreenhouseAdapter{}
}

func (a *GreenhouseAdapter) Source() BoardSource {
	return BoardSourceGreenhouse
}

func (a *GreenhouseAdapter) Capabilities() BoardCapabilities {
	return BoardCapabilities{
		BoardName:             BoardSourceGreenhouse,
		DisplayName:           "Greenhouse ATS",
		IsDirectATS:           true,
		SupportsKeywordSearch: true,
		SupportsLocationFilter: true,
		SupportsRemoteFilter:  true,
		SupportsSalaryFilter:  false, // Greenhouse public board lacks server-side salary filter (AT-010)
		RequiresAuthentication: false,
		RateLimitPerMinute:    60,
		Notes:                 "Direct employer applicant tracking system. Discloses exact application form fields.",
	}
}

func (a *GreenhouseAdapter) DiscoverJobs(ctx context.Context, query JobSearchQuery) ([]DiscoveredJob, error) {
	// Curated canonical sample dataset simulating public Greenhouse boards
	samplePostings := []DiscoveredJob{
		{
			ID:             "gh-stripe-backend-staff",
			Source:         BoardSourceGreenhouse,
			SourceJobID:    "5492011",
			CanonicalURL:   NormalizeCanonicalURL("https://boards.greenhouse.io/stripe/jobs/5492011?utm_source=linkedin&refId=992"),
			DirectApplyURL: "https://boards.greenhouse.io/stripe/jobs/5492011#app",
			Title:          "Staff Software Engineer - Payment Infrastructure",
			Company:        "Stripe",
			CompanyLogoURL: "https://assets.example.com/logos/stripe.png",
			Location: JobLocation{
				RawLocation: "San Francisco, CA (Remote Friendly)",
				City:        "San Francisco",
				State:       "CA",
				Country:     "US",
				IsRemote:    true,
			},
			JobType:     "full_time",
			Description: "Join Stripe Payment Processing team building high-throughput Go and distributed ledger engines with 99.999% availability.",
			RequiredSkills: []string{
				"Go", "Distributed Systems", "Kubernetes", "PostgreSQL", "Kafka",
			},
			Compensation: NormalizedCompensation{
				IsDisclosed: true,
				MinAmount:   215000,
				MaxAmount:   285000,
				Currency:    "USD",
				Period:      PeriodYearly,
				IsEstimated: false,
			},
			DatePosted:       time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC),
			DiscoveredAt:     time.Now().UTC(),
			IsDirectEmployer: true,
		},
		{
			ID:             "gh-figma-systems-lead",
			Source:         BoardSourceGreenhouse,
			SourceJobID:    "4810292",
			CanonicalURL:   NormalizeCanonicalURL("https://boards.greenhouse.io/figma/jobs/4810292?gh_src=careers"),
			DirectApplyURL: "https://boards.greenhouse.io/figma/jobs/4810292#app",
			Title:          "Senior Infrastructure Engineer - Cloud SRE",
			Company:        "Figma",
			CompanyLogoURL: "https://assets.example.com/logos/figma.png",
			Location: JobLocation{
				RawLocation: "New York, NY",
				City:        "New York",
				State:       "NY",
				Country:     "US",
				IsRemote:    false,
				IsHybrid:    true,
			},
			JobType:     "full_time",
			Description: "Scale collaborative graphics engine infrastructure and multi-region Kubernetes clusters.",
			RequiredSkills: []string{
				"Kubernetes", "AWS", "Terraform", "Go", "Docker",
			},
			Compensation: NormalizedCompensation{
				IsDisclosed: true,
				MinAmount:   190000,
				MaxAmount:   240000,
				Currency:    "USD",
				Period:      PeriodYearly,
				IsEstimated: false,
			},
			DatePosted:       time.Date(2026, 9, 12, 14, 30, 0, 0, time.UTC),
			DiscoveredAt:     time.Now().UTC(),
			IsDirectEmployer: true,
		},
	}

	return filterJobs(samplePostings, query), nil
}

// ----------------------------------------------------------------------
// 2. Lever Adapter (Direct ATS)
// ----------------------------------------------------------------------

type LeverAdapter struct{}

func NewLeverAdapter() *LeverAdapter {
	return &LeverAdapter{}
}

func (a *LeverAdapter) Source() BoardSource {
	return BoardSourceLever
}

func (a *LeverAdapter) Capabilities() BoardCapabilities {
	return BoardCapabilities{
		BoardName:             BoardSourceLever,
		DisplayName:           "Lever ATS",
		IsDirectATS:           true,
		SupportsKeywordSearch: true,
		SupportsLocationFilter: true,
		SupportsRemoteFilter:  true,
		SupportsSalaryFilter:  false, // Lever REST endpoints lack server-side salary filter
		RequiresAuthentication: false,
		RateLimitPerMinute:    60,
		Notes:                 "Structured REST API without DOM scraping. Direct employer application pipeline.",
	}
}

func (a *LeverAdapter) DiscoverJobs(ctx context.Context, query JobSearchQuery) ([]DiscoveredJob, error) {
	samplePostings := []DiscoveredJob{
		{
			ID:             "lever-datadog-sre-lead",
			Source:         BoardSourceLever,
			SourceJobID:    "dd-89102-lever",
			CanonicalURL:   NormalizeCanonicalURL("https://jobs.lever.co/datadog/dd-89102-lever?lever-origin=applied&lever-source%5B%5D=LinkedIn"),
			DirectApplyURL: "https://jobs.lever.co/datadog/dd-89102-lever/apply",
			Title:          "Lead Site Reliability Engineer",
			Company:        "Datadog",
			CompanyLogoURL: "https://assets.example.com/logos/datadog.png",
			Location: JobLocation{
				RawLocation: "Remote (Global)",
				IsRemote:    true,
			},
			JobType:     "full_time",
			Description: "Architect massive telemetry pipelines processing billions of events per second with high availability.",
			RequiredSkills: []string{
				"Go", "Python", "Kubernetes", "Kafka", "Linux",
			},
			Compensation: NormalizedCompensation{
				IsDisclosed: false, // Invariant AT-003: Salary undisclosed by employer; preserved as false, never defaulted to 0
			},
			DatePosted:       time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC),
			DiscoveredAt:     time.Now().UTC(),
			IsDirectEmployer: true,
		},
	}

	return filterJobs(samplePostings, query), nil
}

// ----------------------------------------------------------------------
// 3. Ashby Adapter (Direct ATS)
// ----------------------------------------------------------------------

type AshbyAdapter struct{}

func NewAshbyAdapter() *AshbyAdapter {
	return &AshbyAdapter{}
}

func (a *AshbyAdapter) Source() BoardSource {
	return BoardSourceAshby
}

func (a *AshbyAdapter) Capabilities() BoardCapabilities {
	return BoardCapabilities{
		BoardName:             BoardSourceAshby,
		DisplayName:           "Ashby HQ",
		IsDirectATS:           true,
		SupportsKeywordSearch: true,
		SupportsLocationFilter: true,
		SupportsRemoteFilter:  true,
		SupportsSalaryFilter:  true, // Ashby supports structured compensation disclosures
		RequiresAuthentication: false,
		RateLimitPerMinute:    80,
		Notes:                 "Modern ATS platform with structured compensation and remote tier fields.",
	}
}

func (a *AshbyAdapter) DiscoverJobs(ctx context.Context, query JobSearchQuery) ([]DiscoveredJob, error) {
	samplePostings := []DiscoveredJob{
		{
			ID:             "ashby-linear-fullstack",
			Source:         BoardSourceAshby,
			SourceJobID:    "linear-eng-441",
			CanonicalURL:   NormalizeCanonicalURL("https://jobs.ashbyhq.com/linear/linear-eng-441?source=referral&tracking_id=linear_home"),
			DirectApplyURL: "https://jobs.ashbyhq.com/linear/linear-eng-441/application",
			Title:          "Senior FullStack Product Engineer",
			Company:        "Linear",
			CompanyLogoURL: "https://assets.example.com/logos/linear.png",
			Location: JobLocation{
				RawLocation: "Remote (Europe or Americas)",
				IsRemote:    true,
			},
			JobType:     "full_time",
			Description: "Craft high-performance, real-time sync engine features using TypeScript, React, and Node.js microservices.",
			RequiredSkills: []string{
				"TypeScript", "React", "Node.js", "GraphQL", "PostgreSQL",
			},
			Compensation: NormalizedCompensation{
				IsDisclosed: true,
				MinAmount:   175000,
				MaxAmount:   220000,
				Currency:    "USD",
				Period:      PeriodYearly,
				IsEstimated: false,
			},
			DatePosted:       time.Date(2026, 9, 13, 11, 15, 0, 0, time.UTC),
			DiscoveredAt:     time.Now().UTC(),
			IsDirectEmployer: true,
		},
	}

	return filterJobs(samplePostings, query), nil
}

// ----------------------------------------------------------------------
// 4. LinkedIn Adapter (Board Aggregator)
// ----------------------------------------------------------------------

type LinkedInAdapter struct{}

func NewLinkedInAdapter() *LinkedInAdapter {
	return &LinkedInAdapter{}
}

func (a *LinkedInAdapter) Source() BoardSource {
	return BoardSourceLinkedIn
}

func (a *LinkedInAdapter) Capabilities() BoardCapabilities {
	return BoardCapabilities{
		BoardName:             BoardSourceLinkedIn,
		DisplayName:           "LinkedIn Jobs",
		IsDirectATS:           false,
		SupportsKeywordSearch: true,
		SupportsLocationFilter: true,
		SupportsRemoteFilter:  true,
		SupportsSalaryFilter:  true,
		RequiresAuthentication: false,
		RateLimitPerMinute:    30,
		Notes:                 "Aggregator job search. Discloses poster details and company size signals.",
	}
}

func (a *LinkedInAdapter) DiscoverJobs(ctx context.Context, query JobSearchQuery) ([]DiscoveredJob, error) {
	samplePostings := []DiscoveredJob{
		{
			ID:             "li-job-382910401",
			Source:         BoardSourceLinkedIn,
			SourceJobID:    "382910401",
			CanonicalURL:   NormalizeCanonicalURL("https://www.linkedin.com/jobs/view/382910401/?trackingId=e9xK2%3D%3D&refId=feed_1"),
			DirectApplyURL: "", // Aggregator posting without direct employer ATS URL
			Title:          "Backend Systems Engineer - Core Platform",
			Company:        "Shopify",
			CompanyLogoURL: "https://assets.example.com/logos/shopify.png",
			Location: JobLocation{
				RawLocation: "Toronto, ON, Canada (Remote)",
				City:        "Toronto",
				State:       "ON",
				Country:     "CA",
				IsRemote:    true,
			},
			JobType:     "full_time",
			Description: "Scale checkout backend services processing hundreds of thousands of requests per second during flash sales.",
			RequiredSkills: []string{
				"Go", "Ruby", "Kafka", "MySQL", "Docker",
			},
			Compensation: NormalizedCompensation{
				IsDisclosed: true,
				MinAmount:   160000,
				MaxAmount:   210000,
				Currency:    "CAD",
				Period:      PeriodYearly,
				IsEstimated: true, // Estimated by board
			},
			DatePosted:       time.Date(2026, 9, 14, 16, 45, 0, 0, time.UTC),
			DiscoveredAt:     time.Now().UTC(),
			IsDirectEmployer: false,
		},
	}

	return filterJobs(samplePostings, query), nil
}

// ----------------------------------------------------------------------
// 5. Indeed Adapter (Board Aggregator)
// ----------------------------------------------------------------------

type IndeedAdapter struct{}

func NewIndeedAdapter() *IndeedAdapter {
	return &IndeedAdapter{}
}

func (a *IndeedAdapter) Source() BoardSource {
	return BoardSourceIndeed
}

func (a *IndeedAdapter) Capabilities() BoardCapabilities {
	return BoardCapabilities{
		BoardName:             BoardSourceIndeed,
		DisplayName:           "Indeed Jobs",
		IsDirectATS:           false,
		SupportsKeywordSearch: true,
		SupportsLocationFilter: true,
		SupportsRemoteFilter:  true,
		SupportsSalaryFilter:  true,
		RequiresAuthentication: false,
		RateLimitPerMinute:    40,
		Notes:                 "Aggregator job indexing. Normalizes employer compensation intervals.",
	}
}

func (a *IndeedAdapter) DiscoverJobs(ctx context.Context, query JobSearchQuery) ([]DiscoveredJob, error) {
	samplePostings := []DiscoveredJob{
		{
			ID:             "ind-job-9018281",
			Source:         BoardSourceIndeed,
			SourceJobID:    "jk_9018281a",
			CanonicalURL:   NormalizeCanonicalURL("https://www.indeed.com/viewjob?jk=9018281a&utm_source=indeed_feed&from=vj"),
			DirectApplyURL: "",
			Title:          "DevOps & Cloud Automation Specialist",
			Company:        "HashiCorp",
			CompanyLogoURL: "https://assets.example.com/logos/hashicorp.png",
			Location: JobLocation{
				RawLocation: "Austin, TX (Remote)",
				City:        "Austin",
				State:       "TX",
				Country:     "US",
				IsRemote:    true,
			},
			JobType:     "full_time",
			Description: "Design declarative Terraform and Vault automation pipelines across multi-cloud environments.",
			RequiredSkills: []string{
				"Terraform", "Vault", "AWS", "Go", "Kubernetes",
			},
			Compensation: NormalizedCompensation{
				IsDisclosed: true,
				MinAmount:   155000,
				MaxAmount:   195000,
				Currency:    "USD",
				Period:      PeriodYearly,
				IsEstimated: false,
			},
			DatePosted:       time.Date(2026, 9, 11, 9, 20, 0, 0, time.UTC),
			DiscoveredAt:     time.Now().UTC(),
			IsDirectEmployer: false,
		},
	}

	return filterJobs(samplePostings, query), nil
}

// ----------------------------------------------------------------------
// Helper Filtering Logic
// ----------------------------------------------------------------------

func filterJobs(jobs []DiscoveredJob, query JobSearchQuery) []DiscoveredJob {
	if query.Keywords == "" && query.Location == "" && !query.RemoteOnly && len(query.JobTypes) == 0 {
		return jobs
	}

	var results []DiscoveredJob
	keywordTokens := tokenize(query.Keywords)
	locationToken := strings.ToLower(strings.TrimSpace(query.Location))

	for _, job := range jobs {
		// Remote filter
		if query.RemoteOnly && !job.Location.IsRemote {
			continue
		}

		// Location filter
		if locationToken != "" {
			locString := strings.ToLower(fmt.Sprintf("%s %s %s %s",
				job.Location.RawLocation, job.Location.City, job.Location.State, job.Location.Country))
			if !strings.Contains(locString, locationToken) {
				continue
			}
		}

		// Job type filter
		if len(query.JobTypes) > 0 {
			matchedType := false
			for _, jt := range query.JobTypes {
				if strings.EqualFold(job.JobType, jt) {
					matchedType = true
					break
				}
			}
			if !matchedType {
				continue
			}
		}

		// Keyword filter
		if len(keywordTokens) > 0 {
			fullText := strings.ToLower(fmt.Sprintf("%s %s %s %s",
				job.Title, job.Company, job.Description, strings.Join(job.RequiredSkills, " ")))

			matchedAll := true
			for _, token := range keywordTokens {
				if !strings.Contains(fullText, token) {
					matchedAll = false
					break
				}
			}
			if !matchedAll {
				continue
			}
		}

		results = append(results, job)
	}

	return results
}

func tokenize(s string) []string {
	fields := strings.Fields(strings.ToLower(s))
	var clean []string
	for _, f := range fields {
		f = strings.Trim(f, ",.-_!?;:")
		if len(f) > 1 {
			clean = append(clean, f)
		}
	}
	return clean
}
