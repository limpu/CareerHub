package fakes

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrRecipientDuplicateMessage = errors.New("fakes: duplicate message to the same recipient blocked")
	ErrMessagingQuotaExceeded    = errors.New("fakes: messaging quota exceeded for recipient tier")
)

// DirectMessageRequest represents a direct outreach payload to a recruiter or connection.
type DirectMessageRequest struct {
	MessageID      string `json:"message_id"`
	RecipientID    string `json:"recipient_id"`
	RecipientType  string `json:"recipient_type"` // "recruiter", "connection", "creator"
	ContentText    string `json:"content_text"`
	PayloadHash    string `json:"payload_hash"`
	ApprovalID     string `json:"approval_id"`
	HasApproval    bool   `json:"has_approval"`
	RemainingQuota int    `json:"remaining_quota"`
}

// DirectMessageReceipt records message transmission.
type DirectMessageReceipt struct {
	ReceiptID    string    `json:"receipt_id"`
	MessageID    string    `json:"message_id"`
	RecipientID  string    `json:"recipient_id"`
	DispatchedAt time.Time `json:"dispatched_at"`
	Status       string    `json:"status"` // "sent", "rate_limited", "rejected"
}

// FakeMessagingAdapter provides deterministic messaging testing satisfying AT-007 and AT-008.
type FakeMessagingAdapter struct {
	mu            sync.RWMutex
	sentTo        map[string]bool // key: recipientID
	sentMessages  map[string]*DirectMessageReceipt
	simulateQuota int
}

func NewFakeMessagingAdapter(quota int) *FakeMessagingAdapter {
	return &FakeMessagingAdapter{
		sentTo:        make(map[string]bool),
		sentMessages:  make(map[string]*DirectMessageReceipt),
		simulateQuota: quota,
	}
}

// SendMessage dispatches a direct message ensuring approval, deduplication, and quota safety.
func (m *FakeMessagingAdapter) SendMessage(ctx context.Context, req DirectMessageRequest) (*DirectMessageReceipt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Approval Gate verification (AT-007)
	if !req.HasApproval || req.ApprovalID == "" || req.PayloadHash == "" {
		return nil, ErrApprovalMissingOrTampered
	}

	// 2. Recipient deduplication check (REQ-010)
	if m.sentTo[req.RecipientID] {
		return nil, ErrRecipientDuplicateMessage
	}

	// 3. Quota check (AT-008, AT-010)
	if req.RemainingQuota <= 0 || m.simulateQuota <= 0 {
		return nil, ErrMessagingQuotaExceeded
	}

	// Record transmission
	receipt := &DirectMessageReceipt{
		ReceiptID:    fmt.Sprintf("rcpt_msg_%d", time.Now().UnixNano()),
		MessageID:    req.MessageID,
		RecipientID:  req.RecipientID,
		DispatchedAt: time.Now().UTC(),
		Status:       "sent",
	}

	m.sentTo[req.RecipientID] = true
	m.sentMessages[req.MessageID] = receipt
	m.simulateQuota--

	return receipt, nil
}

// HasMessagedRecipient checks if an outreach has already been sent to a recipient.
func (m *FakeMessagingAdapter) HasMessagedRecipient(recipientID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sentTo[recipientID]
}
