package career

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// DiscoveryEngine coordinates multi-board queries, deduplication, and capability reporting (IMP-CAR-07, CAR-07, AT-004, AT-010).
type DiscoveryEngine struct {
	mu       sync.RWMutex
	adapters map[BoardSource]JobBoardAdapter
}

func NewDiscoveryEngine() *DiscoveryEngine {
	engine := &DiscoveryEngine{
		adapters: make(map[BoardSource]JobBoardAdapter),
	}

	// Register default standard adapters
	engine.RegisterAdapter(NewGreenhouseAdapter())
	engine.RegisterAdapter(NewLeverAdapter())
	engine.RegisterAdapter(NewAshbyAdapter())
	engine.RegisterAdapter(NewLinkedInAdapter())
	engine.RegisterAdapter(NewIndeedAdapter())

	return engine
}

// RegisterAdapter registers a job board adapter.
func (e *DiscoveryEngine) RegisterAdapter(adapter JobBoardAdapter) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.adapters[adapter.Source()] = adapter
}

// GetCapabilitiesMatrix returns the capability matrix for all supported boards (AT-010).
func (e *DiscoveryEngine) GetCapabilitiesMatrix() []BoardCapabilities {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var capabilities []BoardCapabilities
	// Deterministic order
	sources := []BoardSource{
		BoardSourceGreenhouse,
		BoardSourceLever,
		BoardSourceAshby,
		BoardSourceLinkedIn,
		BoardSourceIndeed,
	}

	for _, src := range sources {
		if adapter, exists := e.adapters[src]; exists {
			capabilities = append(capabilities, adapter.Capabilities())
		}
	}
	return capabilities
}

// SearchMultiBoard queries all requested job boards, aggregates results, and performs deduplication (AT-004).
func (e *DiscoveryEngine) SearchMultiBoard(ctx context.Context, query JobSearchQuery) ([]DiscoveredJob, error) {
	e.mu.RLock()
	var targetAdapters []JobBoardAdapter

	if len(query.Sources) == 0 {
		// Default to all registered adapters
		for _, a := range e.adapters {
			targetAdapters = append(targetAdapters, a)
		}
	} else {
		for _, src := range query.Sources {
			if a, exists := e.adapters[src]; exists {
				targetAdapters = append(targetAdapters, a)
			}
		}
	}
	e.mu.RUnlock()

	if len(targetAdapters) == 0 {
		return nil, fmt.Errorf("no matching job board adapters available for requested sources")
	}

	type searchResult struct {
		jobs []DiscoveredJob
		err  error
	}

	resultsChan := make(chan searchResult, len(targetAdapters))
	var wg sync.WaitGroup

	for _, adapter := range targetAdapters {
		wg.Add(1)
		go func(ad JobBoardAdapter) {
			defer wg.Done()
			jobs, err := ad.DiscoverJobs(ctx, query)
			resultsChan <- searchResult{jobs: jobs, err: err}
		}(adapter)
	}

	wg.Wait()
	close(resultsChan)

	var allJobs []DiscoveredJob
	for res := range resultsChan {
		if res.err == nil && len(res.jobs) > 0 {
			allJobs = append(allJobs, res.jobs...)
		}
	}

	// Canonical URL & Title+Company Deduplication (AT-004)
	deduped := DeduplicateDiscoveredJobs(allJobs)
	return deduped, nil
}

// DeduplicateDiscoveredJobs eliminates duplicates based on normalized canonical URL or (Company + Title).
// Invariant AT-004: Direct employer postings (e.g. Greenhouse/Lever) take precedence over aggregators.
func DeduplicateDiscoveredJobs(jobs []DiscoveredJob) []DiscoveredJob {
	seenURL := make(map[string]int)     // canonical URL -> index in result
	seenFingerprint := make(map[string]int) // company:title -> index in result
	var result []DiscoveredJob

	for _, job := range jobs {
		cleanURL := NormalizeCanonicalURL(job.CanonicalURL)
		if cleanURL != "" {
			job.CanonicalURL = cleanURL
		}

		fingerprint := makeFingerprint(job.Company, job.Title)

		// Check if already seen by URL
		if idx, found := seenURL[cleanURL]; found && cleanURL != "" {
			// If the newly seen job is direct employer and the previous one was not, supersede it
			if job.IsDirectEmployer && !result[idx].IsDirectEmployer {
				result[idx] = job
			}
			continue
		}

		// Check if already seen by Company + Title fingerprint
		if idx, found := seenFingerprint[fingerprint]; found && fingerprint != "" {
			if job.IsDirectEmployer && !result[idx].IsDirectEmployer {
				result[idx] = job
			}
			continue
		}

		// New unique job
		idx := len(result)
		result = append(result, job)
		if cleanURL != "" {
			seenURL[cleanURL] = idx
		}
		if fingerprint != "" {
			seenFingerprint[fingerprint] = idx
		}
	}

	return result
}

func makeFingerprint(company, title string) string {
	c := strings.ToLower(strings.TrimSpace(company))
	t := strings.ToLower(strings.TrimSpace(title))
	// Clean common prefixes/suffixes
	c = strings.TrimSuffix(c, " inc.")
	c = strings.TrimSuffix(c, " inc")
	c = strings.TrimSuffix(c, " llc")
	c = strings.TrimSuffix(c, " ltd")
	c = strings.TrimSpace(c)
	t = strings.TrimSpace(t)

	if c == "" && t == "" {
		return fmt.Sprintf("random-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%s::%s", c, t)
}
