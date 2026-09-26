package fakes

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FakeSocialPublisherAdapter provides a deterministic mock social publisher
// satisfying REQ-015, AT-007, and AT-015.
type FakeSocialPublisherAdapter struct {
	mu               sync.RWMutex
	failedChannels   map[string]bool // channelID -> simulate failure
	publishedRecords map[string]ChannelPublishResult // key: "postID:channelID"
}

func NewFakeSocialPublisherAdapter() *FakeSocialPublisherAdapter {
	return &FakeSocialPublisherAdapter{
		failedChannels:   make(map[string]bool),
		publishedRecords: make(map[string]ChannelPublishResult),
	}
}

// SetChannelFailure configures an intentional failure for a targeted channel.
func (s *FakeSocialPublisherAdapter) SetChannelFailure(channelID string, fail bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fail {
		s.failedChannels[channelID] = true
	} else {
		delete(s.failedChannels, channelID)
	}
}

// PublishPost executes multi-channel social post publication (REQ-015, AT-015).
func (s *FakeSocialPublisherAdapter) PublishPost(ctx context.Context, sub SocialPostSubmission) (*SocialPublishReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Approval gate check (REQ-015, AT-007)
	if !sub.IsApproved || sub.ApprovalID == "" || sub.PayloadHash == "" {
		return nil, ErrApprovalMissingOrTampered
	}

	report := &SocialPublishReport{
		PostID:  sub.PostID,
		Results: make([]ChannelPublishResult, 0, len(sub.Channels)),
	}

	allSuccess := true
	anySuccess := false

	for _, ch := range sub.Channels {
		recordKey := fmt.Sprintf("%s:%s", sub.PostID, ch.ChannelID)

		// Check if this channel was already successfully published (AT-015: no repeated dispatch)
		if existing, exists := s.publishedRecords[recordKey]; exists && existing.Success {
			report.Results = append(report.Results, existing)
			anySuccess = true
			continue
		}

		// Check simulated failure
		if s.failedChannels[ch.ChannelID] {
			res := ChannelPublishResult{
				ChannelID:    ch.ChannelID,
				Platform:     ch.Platform,
				Success:      false,
				ErrorMessage: fmt.Sprintf("provider API rejected publication on %s", ch.Platform),
				DispatchedAt: time.Now().UTC(),
			}
			report.Results = append(report.Results, res)
			allSuccess = false
			continue
		}

		// Successful dispatch on this channel
		res := ChannelPublishResult{
			ChannelID:    ch.ChannelID,
			Platform:     ch.Platform,
			Success:      true,
			ExternalID:   fmt.Sprintf("ext_%s_%d", ch.Platform, time.Now().UnixNano()),
			DispatchedAt: time.Now().UTC(),
		}
		s.publishedRecords[recordKey] = res
		report.Results = append(report.Results, res)
		anySuccess = true
	}

	// Determine overall status (AT-015: processing is not prematurely Published)
	if allSuccess && len(sub.Channels) > 0 {
		report.OverallStatus = "published"
	} else if anySuccess {
		report.OverallStatus = "partially_failed"
	} else {
		report.OverallStatus = "failed"
	}

	return report, nil
}

// IsChannelPublished returns whether a specific channel has already completed publication.
func (s *FakeSocialPublisherAdapter) IsChannelPublished(postID, channelID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.publishedRecords[fmt.Sprintf("%s:%s", postID, channelID)]
	return ok && rec.Success
}
