package career

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidStageTransition       = errors.New("invalid application stage transition")
	ErrApplicationAlreadyApplied     = errors.New("application record is already marked as verified applied")
	ErrMissingVerification           = errors.New("cannot mark applied: verified receipt or candidate attestation required (REQ-005, AT-005)")
	ErrReconciliationFailed         = errors.New("application reconciliation action failed")
	ErrDuplicateSubmissionBlocked   = errors.New("duplicate submission blocked: application already dispatched or submitted (AT-006)")
	ErrApplicationRecordNotFound     = errors.New("application record not found in ledger")
)

// ApplicationStage represents the explicit multi-stage job application lifecycle (CAR-15, SRC-C8, REQ-002).
type ApplicationStage string

const (
	StageInterested        ApplicationStage = "interested"
	StagePreparing         ApplicationStage = "preparing"
	StageDispatched        ApplicationStage = "dispatched"         // Portal opened / assistant pre-filled; NOT applied (AT-005)
	StageNeedsConfirmation ApplicationStage = "needs_confirmation" // Timed out or ambiguous dispatch awaiting reconciliation (AT-005)
	StageApplied           ApplicationStage = "applied"            // Verified Applied Mark (REQ-005, REQ-016)
	StageInterviewing      ApplicationStage = "interviewing"       // Multi-round interview process (phone, tech, onsite)
	StageOffered           ApplicationStage = "offered"            // Official job offer received
	StageRejected          ApplicationStage = "rejected"           // Company rejection received
	StageWithdrawn         ApplicationStage = "withdrawn"          // Candidate proactively withdrew
	StageArchived          ApplicationStage = "archived"           // Closed/archived historical record
)

// AppliedVerificationType specifies the proof source for the verified Applied mark (REQ-005, AT-005).
type AppliedVerificationType string

const (
	VerificationProviderReceipt AppliedVerificationType = "provider_receipt"  // Direct ATS response (Greenhouse, Ashby, etc.)
	VerificationUserAttestation AppliedVerificationType = "user_attestation"  // Candidate explicit manual attestation
	VerificationEmailReceipt    AppliedVerificationType = "email_confirmation" // Parsed provider confirmation email
)

// AppliedVerificationDetails stores immutable verification proof (REQ-005, REQ-016).
type AppliedVerificationDetails struct {
	VerifiedAt        time.Time               `json:"verified_at"`
	VerificationType  AppliedVerificationType `json:"verification_type"`
	ReceiptID         string                  `json:"receipt_id"`
	ProviderReference string                  `json:"provider_reference,omitempty"`
	AttestationNotes  string                  `json:"attestation_notes,omitempty"`
	VerifierID        string                  `json:"verifier_id"`
}

// InterviewRound represents an active or completed interview stage (CAR-15).
type InterviewRound struct {
	RoundID     string    `json:"round_id"`
	StageName   string    `json:"stage_name"` // "Recruiter Screen", "Technical Assessment", "System Design", "Onsite/Behavioral"
	ScheduledAt time.Time `json:"scheduled_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Interviewer string    `json:"interviewer,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	Outcome     string    `json:"outcome,omitempty"` // "passed", "pending", "failed"
}

// ApplicationTimelineEvent records an immutable stage transition or audit note in the ledger (CAR-15, FND-007).
type ApplicationTimelineEvent struct {
	EventID             string                      `json:"event_id"`
	ApplicationRecordID string                      `json:"application_record_id"`
	Timestamp           time.Time                   `json:"timestamp"`
	FromStage           ApplicationStage            `json:"from_stage"`
	ToStage             ApplicationStage            `json:"to_stage"`
	Trigger             string                      `json:"trigger"` // e.g. "dispatch", "timeout_expiry", "user_confirm", "ats_webhook"
	ActorID             string                      `json:"actor_id"`
	Notes               string                      `json:"notes,omitempty"`
	Verification        *AppliedVerificationDetails `json:"verification,omitempty"`
}

// ReconciliationAction represents the resolution decision for a session in needs_confirmation (AT-005, AT-006).
type ReconciliationAction string

const (
	ReconcileActionConfirmApplied ReconciliationAction = "confirm_applied" // Candidate confirms external submission completed
	ReconcileActionMarkAbandoned  ReconciliationAction = "mark_abandoned"  // Candidate did not complete external submission
	ReconcileActionRetryDispatch  ReconciliationAction = "retry_dispatch"  // Restart dispatch timer for another attempt (AT-006 crash safe)
)

// AllowedTransitions maps valid source stages to valid destination stages.
var AllowedTransitions = map[ApplicationStage][]ApplicationStage{
	StageInterested:        {StagePreparing, StageDispatched, StageArchived},
	StagePreparing:         {StageDispatched, StageInterested, StageArchived},
	StageDispatched:        {StageNeedsConfirmation, StageApplied, StagePreparing, StageWithdrawn, StageArchived},
	StageNeedsConfirmation: {StageApplied, StagePreparing, StageDispatched, StageWithdrawn, StageArchived},
	StageApplied:           {StageInterviewing, StageOffered, StageRejected, StageWithdrawn, StageArchived},
	StageInterviewing:      {StageInterviewing, StageOffered, StageRejected, StageWithdrawn, StageArchived},
	StageOffered:           {StageWithdrawn, StageArchived},
	StageRejected:          {StageArchived},
	StageWithdrawn:         {StageArchived, StageInterested},
	StageArchived:          {StageInterested}, // Can un-archive if reconsidered
}

// IsValidStageTransition evaluates whether moving from current to next is permitted.
func IsValidStageTransition(current, next ApplicationStage) bool {
	if current == next {
		return true
	}
	allowed, exists := AllowedTransitions[current]
	if !exists {
		return false
	}
	for _, target := range allowed {
		if target == next {
			return true
		}
	}
	return false
}

// EvaluateSessionTimeout checks if a dispatched session has expired and must transition to needs_confirmation (AT-005).
// Invariant: An expired dispatched session MUST NOT mark applied automatically.
func EvaluateSessionTimeout(session *ApplicationReviewSession, now time.Time) (bool, ApplicationWorkflowStatus) {
	if session == nil {
		return false, ""
	}
	if session.Status != WorkflowStatusDispatched {
		return false, session.Status
	}

	timeoutSecs := session.TimeoutDurationSecs
	if timeoutSecs <= 0 {
		timeoutSecs = 600 // Default 10 minutes
	}

	var dispatchedTime time.Time
	if session.DispatchedAt != nil {
		dispatchedTime = *session.DispatchedAt
	} else {
		dispatchedTime = session.CreatedAt
	}

	elapsed := now.Sub(dispatchedTime)
	if elapsed >= time.Duration(timeoutSecs)*time.Second {
		return true, WorkflowStatusNeedsConfirmation
	}

	return false, WorkflowStatusDispatched
}

// ReconcileDispatchedSession resolves a session in needs_confirmation or dispatched state (AT-005, AT-006).
func ReconcileDispatchedSession(
	session *ApplicationReviewSession,
	action ReconciliationAction,
	receipt *SubmissionReceipt,
	note string,
	actorID string,
	now time.Time,
) (*ApplicationTimelineEvent, error) {
	if session == nil {
		return nil, ErrWorkflowNotFound
	}

	// Must be in dispatched or needs_confirmation to reconcile
	if session.Status != WorkflowStatusDispatched && session.Status != WorkflowStatusNeedsConfirmation {
		return nil, fmt.Errorf("%w: session %s is in status %s", ErrWorkflowInvalidTransition, session.ID, session.Status)
	}

	switch action {
	case ReconcileActionConfirmApplied:
		if receipt == nil && strings.TrimSpace(note) == "" {
			return nil, ErrMissingVerification
		}

		session.Status = WorkflowStatusApplied
		session.UpdatedAt = now

		recID := fmt.Sprintf("rec_%d", now.UnixNano())
		if receipt != nil && receipt.ReceiptID != "" {
			recID = receipt.ReceiptID
		}

		var vType AppliedVerificationType = VerificationUserAttestation
		if receipt != nil && receipt.ProviderReference != "" {
			vType = VerificationProviderReceipt
		}

		verification := &AppliedVerificationDetails{
			VerifiedAt:        now,
			VerificationType:  vType,
			ReceiptID:         recID,
			ProviderReference: "",
			AttestationNotes:  note,
			VerifierID:        actorID,
		}
		if receipt != nil {
			verification.ProviderReference = receipt.ProviderReference
		}

		event := &ApplicationTimelineEvent{
			EventID:             fmt.Sprintf("evt_%d", now.UnixNano()),
			ApplicationRecordID: session.ApplicationRecordID,
			Timestamp:           now,
			FromStage:           StageNeedsConfirmation,
			ToStage:             StageApplied,
			Trigger:             "reconcile_confirm_applied",
			ActorID:             actorID,
			Notes:               note,
			Verification:        verification,
		}

		session.Receipt = &SubmissionReceipt{
			ReceiptID:         recID,
			ProviderReference: verification.ProviderReference,
			SubmissionURL:     session.ApplyURL,
			ConfirmedAt:       now,
			ConfirmedByUser:   true,
			Notes:             note,
		}

		return event, nil

	case ReconcileActionMarkAbandoned:
		session.Status = WorkflowStatusCancelled
		session.UpdatedAt = now

		event := &ApplicationTimelineEvent{
			EventID:             fmt.Sprintf("evt_%d", now.UnixNano()),
			ApplicationRecordID: session.ApplicationRecordID,
			Timestamp:           now,
			FromStage:           StageNeedsConfirmation,
			ToStage:             StageArchived,
			Trigger:             "reconcile_mark_abandoned",
			ActorID:             actorID,
			Notes:               fmt.Sprintf("Candidate abandoned external dispatch: %s", note),
		}
		return event, nil

	case ReconcileActionRetryDispatch:
		// Reset dispatch timer without duplicating records (AT-006 crash recovery safe)
		session.Status = WorkflowStatusDispatched
		session.DispatchedAt = &now
		session.UpdatedAt = now

		event := &ApplicationTimelineEvent{
			EventID:             fmt.Sprintf("evt_%d", now.UnixNano()),
			ApplicationRecordID: session.ApplicationRecordID,
			Timestamp:           now,
			FromStage:           StageNeedsConfirmation,
			ToStage:             StageDispatched,
			Trigger:             "reconcile_retry_dispatch",
			ActorID:             actorID,
			Notes:               "Dispatch timer reset for renewed external application attempt.",
		}
		return event, nil

	default:
		return nil, fmt.Errorf("%w: unknown action %s", ErrReconciliationFailed, action)
	}
}
