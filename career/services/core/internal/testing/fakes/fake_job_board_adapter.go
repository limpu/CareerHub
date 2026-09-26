package fakes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// JobListing represents an aggregated job listing for testing deduplication.
type JobListing struct {
	ID           string  `json:"id"`
	Board        string  `json:"board"`
	Title        string  `json:"title"`
	Company      string  `json:"company"`
	Location     string  `json:"location"`
	CanonicalURL string  `json:"canonical_url"`
	SalaryMin    float64 `json:"salary_min"`
	SalaryMax    float64 `json:"salary_max"`
}

// DeduplicationCluster groups duplicate and ambiguous job matches (AT-004).
type DeduplicationCluster struct {
	ClusterID        string       `json:"cluster_id"`
	CanonicalListing JobListing   `json:"canonical_listing"`
	ExactDuplicates  []JobListing `json:"exact_duplicates"`
	AmbiguousMatches []JobListing `json:"ambiguous_matches"`
}

// FakeJobBoardAdapter provides a deterministic mock job board provider
// satisfying AT-003, AT-004, AT-005, and AT-006.
type FakeJobBoardAdapter struct {
	mu                   sync.RWMutex
	errorMode            ErrorMode
	dispatchedRuns       map[string]bool // key: "jobID:candidateID"
	receipts             map[string]*SubmissionReceipt
	ambiguousSubmissions map[string]*SubmissionReceipt
}

func NewFakeJobBoardAdapter() *FakeJobBoardAdapter {
	return &FakeJobBoardAdapter{
		errorMode:            ErrorModeNone,
		dispatchedRuns:       make(map[string]bool),
		receipts:             make(map[string]*SubmissionReceipt),
		ambiguousSubmissions: make(map[string]*SubmissionReceipt),
	}
}

// SetErrorMode configures the simulation mode for testing failure edge cases.
func (a *FakeJobBoardAdapter) SetErrorMode(mode ErrorMode) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errorMode = mode
}

// DeduplicateJobs groups job listings into exact duplicates and ambiguous matches (AT-004).
func (a *FakeJobBoardAdapter) DeduplicateJobs(listings []JobListing) []DeduplicationCluster {
	clusters := make([]DeduplicationCluster, 0)
	visited := make(map[string]bool)

	for i, l := range listings {
		if visited[l.ID] {
			continue
		}
		cluster := DeduplicationCluster{
			ClusterID:        fmt.Sprintf("cluster_%s", l.ID),
			CanonicalListing: l,
			ExactDuplicates:  make([]JobListing, 0),
			AmbiguousMatches: make([]JobListing, 0),
		}
		visited[l.ID] = true

		for j := i + 1; j < len(listings); j++ {
			other := listings[j]
			if visited[other.ID] {
				continue
			}

			// 1. Exact duplicate condition: matching canonical URL or identical title+company+location
			isExactURL := l.CanonicalURL != "" && other.CanonicalURL != "" && l.CanonicalURL == other.CanonicalURL
			isExactMeta := strings.EqualFold(l.Title, other.Title) &&
				strings.EqualFold(l.Company, other.Company) &&
				strings.EqualFold(l.Location, other.Location)

			if isExactURL || isExactMeta {
				cluster.ExactDuplicates = append(cluster.ExactDuplicates, other)
				visited[other.ID] = true
				continue
			}

			// 2. Ambiguous match condition: same company and substantial title overlap (AT-004)
			if strings.EqualFold(l.Company, other.Company) {
				lTitleLower := strings.ToLower(l.Title)
				oTitleLower := strings.ToLower(other.Title)
				if strings.Contains(lTitleLower, "engineer") && strings.Contains(oTitleLower, "engineer") {
					cluster.AmbiguousMatches = append(cluster.AmbiguousMatches, other)
					visited[other.ID] = true
				}
			}
		}
		clusters = append(clusters, cluster)
	}

	return clusters
}

// SubmitApplication submits a candidate job application with strict safety invariants (REQ-015, AT-005, AT-006).
func (a *FakeJobBoardAdapter) SubmitApplication(ctx context.Context, sub ApplicationSubmission) (*SubmissionReceipt, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 1. Enforce Approval Gate invariant (REQ-015)
	if sub.ApprovalID == "" || sub.PayloadHash == "" || !sub.HasUserConsent {
		return nil, ErrApprovalMissingOrTampered
	}

	dispatchKey := fmt.Sprintf("%s:%s", sub.JobID, sub.CandidateID)

	// 2. Enforce AT-006: Crash after external dispatch does not trigger an unreviewed duplicate submission
	if a.dispatchedRuns[dispatchKey] {
		// If it was already dispatched during a crash or ambiguous drop, refuse to blindly re-submit
		return nil, ErrDuplicateSubmissionBlocked
	}

	// 3. Handle simulated error scenarios
	switch a.errorMode {
	case ErrorModeTimeout:
		// AT-005: Clicking Apply does not mark Applied; timeout enters needs_confirmation
		receipt := &SubmissionReceipt{
			ReceiptID:        fmt.Sprintf("rcpt_timeout_%d", time.Now().UnixNano()),
			Provider:         "fake-job-board",
			ExternalTargetID: sub.JobID,
			DispatchedAt:     time.Now().UTC(),
			PayloadSHA256:    sub.PayloadHash,
			VerifiedStatus:   "needs_confirmation",
		}
		a.ambiguousSubmissions[dispatchKey] = receipt
		return receipt, ErrSimulatedTimeout

	case ErrorModeRateLimit429:
		return nil, ErrSimulatedRateLimit

	case ErrorModeServer500:
		return nil, ErrSimulatedServer500

	case ErrorModeServer503:
		return nil, ErrSimulatedServer503

	case ErrorModeCrashAfterDispatch:
		// Remote dispatch reached external provider, but client/worker crashed before receiving ack
		a.dispatchedRuns[dispatchKey] = true
		receipt := &SubmissionReceipt{
			ReceiptID:        fmt.Sprintf("rcpt_ambiguous_%d", time.Now().UnixNano()),
			Provider:         "fake-job-board",
			ExternalTargetID: sub.JobID,
			DispatchedAt:     time.Now().UTC(),
			PayloadSHA256:    sub.PayloadHash,
			VerifiedStatus:   "needs_confirmation",
		}
		a.ambiguousSubmissions[dispatchKey] = receipt
		return nil, ErrSimulatedCrashAfterDispatch

	case ErrorModeAmbiguousDrop:
		a.dispatchedRuns[dispatchKey] = true
		receipt := &SubmissionReceipt{
			ReceiptID:        fmt.Sprintf("rcpt_drop_%d", time.Now().UnixNano()),
			Provider:         "fake-job-board",
			ExternalTargetID: sub.JobID,
			DispatchedAt:     time.Now().UTC(),
			PayloadSHA256:    sub.PayloadHash,
			VerifiedStatus:   "needs_confirmation",
		}
		a.ambiguousSubmissions[dispatchKey] = receipt
		return receipt, ErrAmbiguousSubmission

	case ErrorModeNone:
		// Normal successful submission: produces verified receipt (AT-005)
		a.dispatchedRuns[dispatchKey] = true
		confirmationCode := generateConfirmationCode(sub.JobID, sub.CandidateID)
		receipt := &SubmissionReceipt{
			ReceiptID:        fmt.Sprintf("rcpt_success_%d", time.Now().UnixNano()),
			Provider:         "fake-job-board",
			ExternalTargetID: sub.JobID,
			DispatchedAt:     time.Now().UTC(),
			PayloadSHA256:    sub.PayloadHash,
			ConfirmationCode: confirmationCode,
			VerifiedStatus:   "verified_applied",
		}
		a.receipts[dispatchKey] = receipt
		return receipt, nil

	default:
		return nil, fmt.Errorf("unknown error mode: %s", a.errorMode)
	}
}

// ReconcileAmbiguousSubmission allows human or verified receipt reconciliation of an ambiguous dispatch (AT-005, AT-006).
func (a *FakeJobBoardAdapter) ReconcileAmbiguousSubmission(jobID, candidateID string, userConfirmed bool) (*SubmissionReceipt, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	dispatchKey := fmt.Sprintf("%s:%s", jobID, candidateID)
	receipt, exists := a.ambiguousSubmissions[dispatchKey]
	if !exists {
		return nil, fmt.Errorf("no ambiguous submission found for %s", dispatchKey)
	}

	if !userConfirmed {
		receipt.VerifiedStatus = "rejected"
		delete(a.dispatchedRuns, dispatchKey)
		delete(a.ambiguousSubmissions, dispatchKey)
		return receipt, nil
	}

	// Confirmed by user or provider receipt
	receipt.VerifiedStatus = "verified_applied"
	receipt.ConfirmationCode = generateConfirmationCode(jobID, candidateID)
	a.receipts[dispatchKey] = receipt
	delete(a.ambiguousSubmissions, dispatchKey)
	return receipt, nil
}

// GetReceipt retrieves a verified receipt by job and candidate ID.
func (a *FakeJobBoardAdapter) GetReceipt(jobID, candidateID string) (*SubmissionReceipt, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	rcpt, ok := a.receipts[fmt.Sprintf("%s:%s", jobID, candidateID)]
	return rcpt, ok
}

func generateConfirmationCode(jobID, candidateID string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", jobID, candidateID, time.Now().UnixNano())))
	return "CONF-" + strings.ToUpper(hex.EncodeToString(sum[:4]))
}
