package linkedin_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/social-platform/services/core/internal/linkedin"
)

type CommentsSweepFixtureData struct {
	Scenarios struct {
		TechnicalPostCommentDrafting struct {
			WorkspaceID string                   `json:"workspace_id"`
			TenantID    string                   `json:"tenant_id"`
			TargetPost  linkedin.SweptTargetPost `json:"target_post"`
			CandidateFacts []string              `json:"candidate_facts"`
			ExpectedAngles []string              `json:"expected_angles"`
			ForbiddenGenericPhrases []string     `json:"forbidden_generic_phrases"`
			RequiresApproval bool                `json:"requires_approval"`
			RequiresOneClickCopy bool            `json:"requires_one_click_copy"`
		} `json:"technical_post_comment_drafting"`
		HiringAnnouncementThreadReply struct {
			WorkspaceID string                   `json:"workspace_id"`
			TenantID    string                   `json:"tenant_id"`
			TargetPost  linkedin.SweptTargetPost `json:"target_post"`
			CandidateFacts []string              `json:"candidate_facts"`
			ExpectedAngle string                 `json:"expected_angle"`
			ExpectedContentSnippet string        `json:"expected_content_snippet"`
			RequiresApproval bool                `json:"requires_approval"`
		} `json:"hiring_announcement_thread_reply"`
		BulkAutoEngagementFailClosed struct {
			WorkspaceID string `json:"workspace_id"`
			TenantID    string `json:"tenant_id"`
			AttemptedMode string `json:"attempted_mode"`
			ExpectedError string `json:"expected_error"`
		} `json:"bulk_auto_engagement_fail_closed"`
		CrossTenantSweepIsolation struct {
			AlphaSweep struct {
				SweepID     string `json:"sweep_id"`
				WorkspaceID string `json:"workspace_id"`
				TenantID    string `json:"tenant_id"`
				PostID      string `json:"post_id"`
			} `json:"alpha_sweep"`
			BetaQuery struct {
				WorkspaceID string `json:"workspace_id"`
				TenantID    string `json:"tenant_id"`
			} `json:"beta_query"`
			ExpectedIsolationError string `json:"expected_isolation_error"`
		} `json:"cross_tenant_sweep_isolation"`
	} `json:"scenarios"`
}

func loadCommentsSweepFixture(t *testing.T) *CommentsSweepFixtureData {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "forms", "linkedin_comments_sweep.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read comments sweep fixture: %v", err)
	}
	var data CommentsSweepFixtureData
	if err := json.Unmarshal(bytes, &data); err != nil {
		t.Fatalf("failed to unmarshal comments sweep fixture: %v", err)
	}
	return &data
}

func setupCommentsSweepTest() (*linkedin.MemoryRepository, *linkedin.Service) {
	repo := linkedin.NewMemoryRepository()
	svc := linkedin.NewService(repo)
	return repo, svc
}

func TestCommentsSweep_FixtureGroundTruth(t *testing.T) {
	fixture := loadCommentsSweepFixture(t)
	_, svc := setupCommentsSweepTest()
	ctx := context.Background()

	// Scenario 1: Technical Post Comment Drafting
	t.Run("technical_post_comment_drafting", func(t *testing.T) {
		sc := fixture.Scenarios.TechnicalPostCommentDrafting
		post := sc.TargetPost
		post.WorkspaceID = sc.WorkspaceID
		post.TenantID = sc.TenantID

		err := svc.SaveSweptTargetPost(ctx, &post)
		if err != nil {
			t.Fatalf("SaveSweptTargetPost failed: %v", err)
		}

		fetched, err := svc.GetSweptTargetPost(ctx, post.PostID, post.TenantID)
		if err != nil {
			t.Fatalf("GetSweptTargetPost failed: %v", err)
		}
		if fetched.Category != linkedin.CategoryTechnicalDiscussion {
			t.Errorf("expected category %s, got %s", linkedin.CategoryTechnicalDiscussion, fetched.Category)
		}

		drafts, err := svc.GenerateCommentDrafts(ctx, post.PostID, post.TenantID, "manual", sc.CandidateFacts)
		if err != nil {
			t.Fatalf("GenerateCommentDrafts failed: %v", err)
		}
		if len(drafts) != len(sc.ExpectedAngles) {
			t.Fatalf("expected %d drafts, got %d", len(sc.ExpectedAngles), len(drafts))
		}

		for i, expAngle := range sc.ExpectedAngles {
			if string(drafts[i].Angle) != expAngle {
				t.Errorf("draft[%d] angle expected %s, got %s", i, expAngle, drafts[i].Angle)
			}
			if err := linkedin.ValidateCommentText(drafts[i].CommentText); err != nil {
				t.Errorf("draft[%d] failed validation: %v", i, err)
			}
		}

		for _, forbidden := range sc.ForbiddenGenericPhrases {
			if err := linkedin.ValidateCommentText(forbidden); err != linkedin.ErrGenericHollowCommentBlocked {
				t.Errorf("expected ErrGenericHollowCommentBlocked for %q, got %v", forbidden, err)
			}
		}

		// Approval test
		if sc.RequiresApproval {
			appr, err := svc.ApproveCommentDraft(ctx, drafts[0].DraftID, post.TenantID, "reviewer")
			if err != nil {
				t.Fatalf("ApproveCommentDraft failed: %v", err)
			}
			if appr.Status != linkedin.CommentDraftApproved || appr.ApprovalToken == "" {
				t.Errorf("expected approved draft with token")
			}
			if sc.RequiresOneClickCopy {
				copied, err := svc.MarkCommentDraftCopied(ctx, drafts[0].DraftID, post.TenantID)
				if err != nil {
					t.Fatalf("MarkCommentDraftCopied failed: %v", err)
				}
				if copied.Status != linkedin.CommentDraftCopiedToClipboard {
					t.Errorf("expected copied status, got %s", copied.Status)
				}
			}
		}
	})

	// Scenario 2: Hiring Announcement Thread Reply
	t.Run("hiring_announcement_thread_reply", func(t *testing.T) {
		sc := fixture.Scenarios.HiringAnnouncementThreadReply
		post := sc.TargetPost
		post.WorkspaceID = sc.WorkspaceID
		post.TenantID = sc.TenantID

		err := svc.SaveSweptTargetPost(ctx, &post)
		if err != nil {
			t.Fatalf("SaveSweptTargetPost failed: %v", err)
		}

		drafts, err := svc.GenerateCommentDrafts(ctx, post.PostID, post.TenantID, "manual", sc.CandidateFacts)
		if err != nil {
			t.Fatalf("GenerateCommentDrafts failed: %v", err)
		}
		if len(drafts) == 0 {
			t.Fatalf("expected drafts generated")
		}

		foundSnippet := false
		for _, d := range drafts {
			if strings.Contains(d.CommentText, sc.ExpectedContentSnippet) {
				foundSnippet = true
				break
			}
		}
		if !foundSnippet {
			t.Errorf("expected draft comment to contain snippet %q", sc.ExpectedContentSnippet)
		}
	})

	// Scenario 3: Bulk Auto Engagement Fail-Closed
	t.Run("bulk_auto_engagement_fail_closed", func(t *testing.T) {
		sc := fixture.Scenarios.BulkAutoEngagementFailClosed
		err := linkedin.RejectUnsupportedBulkEngagement(sc.AttemptedMode)
		if err == nil || err.Error() != sc.ExpectedError {
			t.Errorf("expected error %q, got %v", sc.ExpectedError, err)
		}

		// Service call also rejects
		_, err = svc.GenerateCommentDrafts(ctx, "any-post", sc.TenantID, sc.AttemptedMode, nil)
		if err == nil || err.Error() != sc.ExpectedError {
			t.Errorf("expected service error %q, got %v", sc.ExpectedError, err)
		}
	})

	// Scenario 4: Cross Tenant Sweep Isolation
	t.Run("cross_tenant_sweep_isolation", func(t *testing.T) {
		sc := fixture.Scenarios.CrossTenantSweepIsolation
		post := linkedin.SweptTargetPost{
			PostID:      sc.AlphaSweep.PostID,
			WorkspaceID: sc.AlphaSweep.WorkspaceID,
			TenantID:    sc.AlphaSweep.TenantID,
			AuthorName:  "Alpha Leader",
			PostURL:     "https://www.linkedin.com/feed/update/urn:li:activity:9999",
			Content:     "Alpha technical discussion",
			Category:    linkedin.CategoryTechnicalDiscussion,
		}
		_ = svc.SaveSweptTargetPost(ctx, &post)

		_, err := svc.GetSweptTargetPost(ctx, post.PostID, sc.BetaQuery.TenantID)
		if err == nil || err.Error() != sc.ExpectedIsolationError {
			t.Errorf("expected error %q, got %v", sc.ExpectedIsolationError, err)
		}
	})
}

func TestCommentsSweep_Categorization(t *testing.T) {
	cases := []struct {
		content  string
		expected linkedin.PostSweepCategory
	}{
		{"We are expanding our distributed database team and hiring senior engineers!", linkedin.CategoryHiringAnnouncement},
		{"Deep dive into Raft consensus algorithms and quorum leases in distributed clusters.", linkedin.CategoryTechnicalDiscussion},
		{"Reflections on engineering leadership, high trust culture, and team mentorship.", linkedin.CategoryThoughtLeadership},
		{"Today we announced the official release of our platform 2.0!", linkedin.CategoryIndustryNews},
		{"Sharing some random thoughts from the team's coffee meetup today.", linkedin.CategoryGeneralDiscussion},
	}

	for _, c := range cases {
		cat := linkedin.CategorizePostContent(c.content)
		if cat != c.expected {
			t.Errorf("for content %q, expected %s, got %s", c.content, c.expected, cat)
		}
	}
}

func TestCommentsSweep_ValidateCommentText(t *testing.T) {
	// Hollow phrases blocked
	badPhrases := []string{
		"Great post!",
		"Thanks for sharing!!",
		"Totally agree 100%",
		"Insightful read, thanks!",
		"nice post",
	}
	for _, p := range badPhrases {
		if err := linkedin.ValidateCommentText(p); err != linkedin.ErrGenericHollowCommentBlocked {
			t.Errorf("expected hollow comment error for %q, got %v", p, err)
		}
	}

	// Empty text blocked
	if err := linkedin.ValidateCommentText("   "); err != linkedin.ErrEmptyCommentText {
		t.Errorf("expected ErrEmptyCommentText, got %v", err)
	}

	// Substantive text passes
	goodText := "In our multi-region Kubernetes deployments, asymmetric partitions often cause lease drift before leader heartbeats expire. Tuning the timeout constants reduced failover latency by 40%."
	if err := linkedin.ValidateCommentText(goodText); err != nil {
		t.Errorf("expected good comment to pass, got %v", err)
	}
}

func TestCommentsSweep_ApprovalAndTamperInvalidation(t *testing.T) {
	_, svc := setupCommentsSweepTest()
	ctx := context.Background()

	post := linkedin.SweptTargetPost{
		PostID:      "post-tamper-01",
		WorkspaceID: "ws-01",
		TenantID:    "tenant-01",
		AuthorName:  "Jane Doe",
		PostURL:     "https://www.linkedin.com/feed/update/urn:li:activity:999999",
		Content:     "Discussing multi-region consensus latency.",
		Category:    linkedin.CategoryTechnicalDiscussion,
	}
	if err := svc.SaveSweptTargetPost(ctx, &post); err != nil {
		t.Fatalf("SaveSweptTargetPost failed: %v", err)
	}

	drafts, err := svc.GenerateCommentDrafts(ctx, post.PostID, post.TenantID, "manual_review", []string{"10 yrs distributed systems experience"})
	if err != nil {
		t.Fatalf("GenerateCommentDrafts failed: %v", err)
	}
	target := drafts[0]

	// Clipboard copy before approval should fail
	_, err = svc.MarkCommentDraftCopied(ctx, target.DraftID, post.TenantID)
	if err == nil {
		t.Fatalf("expected error copying unapproved draft, got nil")
	}

	// Approve draft
	approved, err := svc.ApproveCommentDraft(ctx, target.DraftID, post.TenantID, "user-alice")
	if err != nil {
		t.Fatalf("ApproveCommentDraft failed: %v", err)
	}
	if approved.Status != linkedin.CommentDraftApproved {
		t.Errorf("expected status %s, got %s", linkedin.CommentDraftApproved, approved.Status)
	}
	if approved.ApprovalToken == "" || !strings.HasPrefix(approved.ApprovalToken, "hmac-sha256:") {
		t.Errorf("expected valid hmac token, got %s", approved.ApprovalToken)
	}

	// Now copy to clipboard should succeed
	copied, err := svc.MarkCommentDraftCopied(ctx, target.DraftID, post.TenantID)
	if err != nil {
		t.Fatalf("MarkCommentDraftCopied failed: %v", err)
	}
	if copied.Status != linkedin.CommentDraftCopiedToClipboard {
		t.Errorf("expected status %s, got %s", linkedin.CommentDraftCopiedToClipboard, copied.Status)
	}

	// Tamper: Update text after approval
	updatedText := "In our multi-region Kubernetes deployments, asymmetric partitions often cause lease drift before leader heartbeats expire. We also added Raft pre-vote checks."
	edited, err := svc.UpdateCommentDraftText(ctx, target.DraftID, post.TenantID, updatedText)
	if err != nil {
		t.Fatalf("UpdateCommentDraftText failed: %v", err)
	}
	// Verify tamper invalidation: status resets to draft, approval token cleared
	if edited.Status != linkedin.CommentDraftPendingApproval {
		t.Errorf("expected status reset to %s after edit, got %s", linkedin.CommentDraftPendingApproval, edited.Status)
	}
	if edited.ApprovalToken != "" {
		t.Errorf("expected approval token to be wiped after edit, got %s", edited.ApprovalToken)
	}
	if edited.ApprovedBy != "" || edited.ApprovedAt != nil {
		t.Errorf("expected approved metadata to be cleared")
	}

	// Copy again without re-approving should fail
	_, err = svc.MarkCommentDraftCopied(ctx, target.DraftID, post.TenantID)
	if err == nil {
		t.Fatalf("expected copy after tamper to fail, got nil")
	}
}

func TestCommentsSweep_RejectCommentDraft(t *testing.T) {
	_, svc := setupCommentsSweepTest()
	ctx := context.Background()

	post := linkedin.SweptTargetPost{
		PostID:      "post-reject-01",
		WorkspaceID: "ws-01",
		TenantID:    "tenant-01",
		AuthorName:  "Jane Doe",
		PostURL:     "https://www.linkedin.com/feed/update/urn:li:activity:888888",
		Content:     "Technical discussions on throughput.",
		Category:    linkedin.CategoryTechnicalDiscussion,
	}
	_ = svc.SaveSweptTargetPost(ctx, &post)

	drafts, _ := svc.GenerateCommentDrafts(ctx, post.PostID, post.TenantID, "manual", nil)
	target := drafts[0]

	rejected, err := svc.RejectCommentDraft(ctx, target.DraftID, post.TenantID, "Not relevant to our focus")
	if err != nil {
		t.Fatalf("RejectCommentDraft failed: %v", err)
	}
	if rejected.Status != linkedin.CommentDraftRejected {
		t.Errorf("expected status %s, got %s", linkedin.CommentDraftRejected, rejected.Status)
	}
	if !strings.Contains(rejected.Rationale, "Not relevant to our focus") {
		t.Errorf("expected rationale to contain rejection reason, got %s", rejected.Rationale)
	}
}

func TestCommentsSweep_CrossTenantSecurity(t *testing.T) {
	_, svc := setupCommentsSweepTest()
	ctx := context.Background()

	post := linkedin.SweptTargetPost{
		PostID:      "post-cross-01",
		WorkspaceID: "ws-alpha",
		TenantID:    "tenant-alpha",
		AuthorName:  "Alpha Author",
		PostURL:     "https://www.linkedin.com/feed/update/urn:li:activity:777777",
		Content:     "Alpha organization confidential engineering discuss.",
		Category:    linkedin.CategoryTechnicalDiscussion,
	}
	_ = svc.SaveSweptTargetPost(ctx, &post)

	drafts, _ := svc.GenerateCommentDrafts(ctx, post.PostID, "tenant-alpha", "manual", nil)
	draftID := drafts[0].DraftID

	// Tenant Beta tries to access Alpha's post
	_, err := svc.GetSweptTargetPost(ctx, post.PostID, "tenant-beta")
	if err != linkedin.ErrCrossTenantAccessDenied {
		t.Errorf("expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// Tenant Beta tries to delete Alpha's post
	err = svc.DeleteSweptTargetPost(ctx, post.PostID, "tenant-beta")
	if err != linkedin.ErrCrossTenantAccessDenied {
		t.Errorf("expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// Tenant Beta tries to view Alpha's draft
	_, err = svc.GetCommentDraft(ctx, draftID, "tenant-beta")
	if err != linkedin.ErrCrossTenantAccessDenied {
		t.Errorf("expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// Tenant Beta tries to approve Alpha's draft
	_, err = svc.ApproveCommentDraft(ctx, draftID, "tenant-beta", "intruder")
	if err != linkedin.ErrCrossTenantAccessDenied {
		t.Errorf("expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// Tenant Beta tries to update Alpha's draft
	_, err = svc.UpdateCommentDraftText(ctx, draftID, "tenant-beta", "Malicious comment text")
	if err != linkedin.ErrCrossTenantAccessDenied {
		t.Errorf("expected ErrCrossTenantAccessDenied, got %v", err)
	}

	// Tenant Beta tries to delete Alpha's draft
	_, err = svc.RejectCommentDraft(ctx, draftID, "tenant-beta", "hack")
	if err == nil {
		t.Errorf("expected error rejecting foreign draft, got nil")
	}
}

func TestCommentsSweep_ListFilters(t *testing.T) {
	_, svc := setupCommentsSweepTest()
	ctx := context.Background()

	p1 := linkedin.SweptTargetPost{
		PostID:        "post-filter-1",
		WorkspaceID:   "ws-main",
		TenantID:      "tenant-main",
		AuthorName:    "Sarah Tech",
		AuthorCompany: "CloudCorp",
		Content:       "Kubernetes architecture optimization.",
		Category:      linkedin.CategoryTechnicalDiscussion,
	}
	p2 := linkedin.SweptTargetPost{
		PostID:        "post-filter-2",
		WorkspaceID:   "ws-main",
		TenantID:      "tenant-main",
		AuthorName:    "Recruiter Bob",
		AuthorCompany: "StaffingInc",
		Content:       "We are hiring senior engineers!",
		Category:      linkedin.CategoryHiringAnnouncement,
	}
	_ = svc.SaveSweptTargetPost(ctx, &p1)
	_ = svc.SaveSweptTargetPost(ctx, &p2)

	// Filter by category
	techPosts, err := svc.ListSweptTargetPosts(ctx, linkedin.SweptPostFilter{
		WorkspaceID: "ws-main",
		TenantID:    "tenant-main",
		Category:    linkedin.CategoryTechnicalDiscussion,
	})
	if err != nil {
		t.Fatalf("ListSweptTargetPosts failed: %v", err)
	}
	if len(techPosts) != 1 || techPosts[0].PostID != "post-filter-1" {
		t.Errorf("expected 1 tech post, got %d", len(techPosts))
	}

	// Filter by query (company search)
	staffingPosts, err := svc.ListSweptTargetPosts(ctx, linkedin.SweptPostFilter{
		WorkspaceID: "ws-main",
		TenantID:    "tenant-main",
		Query:       "StaffingInc",
	})
	if err != nil {
		t.Fatalf("ListSweptTargetPosts with query failed: %v", err)
	}
	if len(staffingPosts) != 1 || staffingPosts[0].PostID != "post-filter-2" {
		t.Errorf("expected 1 staffing post, got %d", len(staffingPosts))
	}
}
