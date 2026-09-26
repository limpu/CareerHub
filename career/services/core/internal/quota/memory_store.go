package quota

import (
	"context"
	"sync"
	"time"
)

// MemoryQuotaStore is an in-memory, thread-safe implementation of QuotaStore.
type MemoryQuotaStore struct {
	mu           sync.RWMutex
	reservations map[string]*Reservation
	accounts     map[string]*AccountState
	isOutage     bool // Used to simulate store failures and verify fail-closed invariants
}

// NewMemoryQuotaStore creates a new MemoryQuotaStore.
func NewMemoryQuotaStore() *MemoryQuotaStore {
	return &MemoryQuotaStore{
		reservations: make(map[string]*Reservation),
		accounts:     make(map[string]*AccountState),
	}
}

// SetOutage allows tests to simulate database/Redis outages to verify fail-closed behavior.
func (s *MemoryQuotaStore) SetOutage(outage bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isOutage = outage
}

// Reserve atomically checks both metric limit and parent metric limit (if nested),
// and stores the reservation if within quota.
func (s *MemoryQuotaStore) Reserve(ctx context.Context, res *Reservation, rule QuotaDefinition, parentRule *QuotaDefinition, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isOutage {
		return ErrStoreUnavailable
	}

	// 1. Check metric limit
	consumed, reserved := s.calculateUsageLocked(res.AccountID, rule.Metric, rule.Window, now)
	if consumed+reserved+res.Amount > rule.Limit {
		return ErrQuotaExceeded
	}

	// 2. If nested, check parent metric limit
	if parentRule != nil {
		parentConsumed, parentReserved := s.calculateUsageLocked(res.AccountID, parentRule.Metric, parentRule.Window, now)
		if parentConsumed+parentReserved+res.Amount > parentRule.Limit {
			return ErrNestedQuotaExceeded
		}
	}

	// 3. Atomically record reservation
	s.reservations[res.ID] = res
	return nil
}

// Settle marks a pending reservation as settled upon verified execution.
func (s *MemoryQuotaStore) Settle(ctx context.Context, reservationID string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isOutage {
		return ErrStoreUnavailable
	}

	res, exists := s.reservations[reservationID]
	if !exists {
		return ErrReservationNotFound
	}
	if res.Status != ReservationPending && res.Status != ReservationRetained {
		return ErrReservationAlreadyFinalized
	}

	res.Status = ReservationSettled
	res.SettledAt = &now
	return nil
}

// Release voids a reservation for proven non-executed failures.
func (s *MemoryQuotaStore) Release(ctx context.Context, reservationID string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isOutage {
		return ErrStoreUnavailable
	}

	res, exists := s.reservations[reservationID]
	if !exists {
		return ErrReservationNotFound
	}
	if res.Status != ReservationPending {
		return ErrReservationAlreadyFinalized
	}

	res.Status = ReservationReleased
	res.ReleasedAt = &now
	return nil
}

// Retain locks a reservation when the external outcome is ambiguous (e.g. timeout),
// ensuring it continues to count against the quota and cannot be blindly retried.
func (s *MemoryQuotaStore) Retain(ctx context.Context, reservationID string, reason string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isOutage {
		return ErrStoreUnavailable
	}

	res, exists := s.reservations[reservationID]
	if !exists {
		return ErrReservationNotFound
	}
	if res.Status != ReservationPending {
		return ErrReservationAlreadyFinalized
	}

	res.Status = ReservationRetained
	res.RetainedAt = &now
	res.RetainReason = reason
	return nil
}

// GetUsage computes consumed and active reserved units within the rolling window.
func (s *MemoryQuotaStore) GetUsage(ctx context.Context, accountID string, metric ActionMetric, window time.Duration, now time.Time) (consumed int, reserved int, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.isOutage {
		return 0, 0, ErrStoreUnavailable
	}

	c, r := s.calculateUsageLocked(accountID, metric, window, now)
	return c, r, nil
}

// GetReservation retrieves a reservation by ID.
func (s *MemoryQuotaStore) GetReservation(ctx context.Context, reservationID string) (*Reservation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.isOutage {
		return nil, ErrStoreUnavailable
	}

	res, exists := s.reservations[reservationID]
	if !exists {
		return nil, ErrReservationNotFound
	}

	copyRes := *res
	return &copyRes, nil
}

// GetAccountState retrieves account circuit-breaker status.
func (s *MemoryQuotaStore) GetAccountState(ctx context.Context, accountID string) (*AccountState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.isOutage {
		return nil, ErrStoreUnavailable
	}

	state, exists := s.accounts[accountID]
	if !exists {
		return &AccountState{
			AccountID: accountID,
			Status:    AccountStatusActive,
			UpdatedAt: time.Now().UTC(),
		}, nil
	}

	copyState := *state
	return &copyState, nil
}

// SetAccountState records account circuit-breaker status.
func (s *MemoryQuotaStore) SetAccountState(ctx context.Context, state *AccountState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isOutage {
		return ErrStoreUnavailable
	}

	s.accounts[state.AccountID] = state
	return nil
}

// calculateUsageLocked computes consumed and active reserved items within rolling window.
// Note: If an action has a parent metric (e.g. recruiter DM is a child of DM),
// querying the parent metric MUST also include the child metric's consumption and reservations!
func (s *MemoryQuotaStore) calculateUsageLocked(accountID string, metric ActionMetric, window time.Duration, now time.Time) (int, int) {
	windowStart := now.Add(-window)
	consumed := 0
	reserved := 0

	for _, r := range s.reservations {
		if r.AccountID != accountID {
			continue
		}

		// Match direct metric OR if querying a parent metric, match reservations where this metric is the parent
		matches := (r.Metric == metric) || (r.ParentMetric == metric)
		if !matches {
			continue
		}

		// Check window boundary based on creation or settlement
		timestamp := r.CreatedAt
		if r.SettledAt != nil {
			timestamp = *r.SettledAt
		}
		if timestamp.Before(windowStart) {
			continue
		}

		switch r.Status {
		case ReservationSettled:
			consumed += r.Amount
		case ReservationPending:
			// If expired and not settled/retained, ignore as expired lease
			if now.After(r.ExpiresAt) {
				continue
			}
			reserved += r.Amount
		case ReservationRetained:
			// Retained reservations ALWAYS count against quota until investigated
			reserved += r.Amount
		case ReservationReleased:
			// Released reservations do not count
		}
	}

	return consumed, reserved
}
