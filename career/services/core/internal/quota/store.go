package quota

import (
	"context"
	"time"
)

// QuotaStore abstracts the persistence and atomic concurrency mechanics for quotas.
type QuotaStore interface {
	// Reserve atomically checks limit & parent limit, and records reservation if available.
	Reserve(ctx context.Context, res *Reservation, rule QuotaDefinition, parentRule *QuotaDefinition, now time.Time) error

	// Settle commits a reservation into settled usage upon confirmed successful execution.
	Settle(ctx context.Context, reservationID string, now time.Time) error

	// Release voids a reservation for a proven non-executed failure.
	Release(ctx context.Context, reservationID string, now time.Time) error

	// Retain marks a reservation as retained when outcome is ambiguous (e.g. timeout),
	// preventing blind duplicate executions and continuing to count against quota.
	Retain(ctx context.Context, reservationID string, reason string, now time.Time) error

	// GetUsage returns the consumed and active (pending + retained) reservations for a metric within the rolling window.
	GetUsage(ctx context.Context, accountID string, metric ActionMetric, window time.Duration, now time.Time) (consumed int, reserved int, err error)

	// GetReservation retrieves a reservation by its ID.
	GetReservation(ctx context.Context, reservationID string) (*Reservation, error)

	// GetAccountState retrieves the circuit-breaker / safety status of an external account.
	GetAccountState(ctx context.Context, accountID string) (*AccountState, error)

	// SetAccountState updates the circuit-breaker / safety status of an external account.
	SetAccountState(ctx context.Context, state *AccountState) error
}
