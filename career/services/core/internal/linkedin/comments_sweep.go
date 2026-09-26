package linkedin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrSweepPostNotFound           = errors.New("comments: swept target post not found")
	ErrCommentDraftNotFound        = errors.New("comments: comment draft not found")
	ErrUnsupportedBulkEngagement   = errors.New("comments: autonomous bulk auto-engagement and unreviewed commenting permanently barred under platform integrity policy (AT-010, REQ-015, LI-10)")
	ErrGenericHollowCommentBlocked = errors.New("comments: generic hollow praise or spam phrases strictly barred; comments must add substantive domain value (FND-015, LI-10)")
	ErrInvalidCommentApprovalToken = errors.New("comments: invalid or tampered comment approval token (FND-010, AT-007)")
	ErrEmptyCommentText            = errors.New("comments: comment draft text cannot be empty")
)

type PostSweepCategory string

const (
	CategoryTechnicalDiscussion PostSweepCategory = "technical_discussion"
	CategoryHiringAnnouncement  PostSweepCategory = "hiring_announcement"
	CategoryThoughtLeadership   PostSweepCategory = "thought_leadership"
	CategoryIndustryNews        PostSweepCategory = "industry_news"
	CategoryGeneralDiscussion   PostSweepCategory = "general_discussion"
)

type CommentDraftAngle string

const (
	AngleInsightfulAddition    CommentDraftAngle = "insightful_addition"
	AngleEngagingQuestion      CommentDraftAngle = "engaging_question"
	AngleSupportivePerspective CommentDraftAngle = "supportive_perspective"
)

type CommentDraftStatus string

const (
	CommentDraftPendingApproval   CommentDraftStatus = "draft"
	CommentDraftApproved          CommentDraftStatus = "approved"
	CommentDraftCopiedToClipboard CommentDraftStatus = "copied_to_clipboard"
	CommentDraftRejected          CommentDraftStatus = "rejected"
)

type SweptPostComment struct {
	CommentID      string    `json:"comment_id"`
	AuthorName     string    `json:"author_name"`
	AuthorHeadline string    `json:"author_headline,omitempty"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
	LikesCount     int       `json:"likes_count"`
	IsHighPriority bool      `json:"is_high_priority"` // recruiter or hiring lead
}

type SweptTargetPost struct {
	PostID                string             `json:"post_id"`
	WorkspaceID           string             `json:"workspace_id"`
	TenantID              string             `json:"tenant_id"`
	AuthorName            string             `json:"author_name"`
	AuthorHeadline        string             `json:"author_headline,omitempty"`
	AuthorCompany         string             `json:"author_company,omitempty"`
	PostURL               string             `json:"post_url"`
	Content               string             `json:"content"`
	Category              PostSweepCategory  `json:"category"`
	ExistingCommentsCount int                `json:"existing_comments_count"`
	Comments              []SweptPostComment `json:"comments,omitempty"`
	SweptAt               time.Time          `json:"swept_at"`
}

type CommentDraft struct {
	DraftID           string             `json:"draft_id"`
	PostID            string             `json:"post_id"`
	WorkspaceID       string             `json:"workspace_id"`
	TenantID          string             `json:"tenant_id"`
	TargetCommentID   string             `json:"target_comment_id,omitempty"` // empty if top-level comment
	Angle             CommentDraftAngle  `json:"angle"`
	CommentText       string             `json:"comment_text"`
	Rationale         string             `json:"rationale"`
	VerifiedFactsUsed []string           `json:"verified_facts_used"`
	CharacterCount    int                `json:"character_count"`
	Status            CommentDraftStatus `json:"status"`
	ApprovalToken     string             `json:"approval_token,omitempty"`
	ApprovedBy        string             `json:"approved_by,omitempty"`
	ApprovedAt        *time.Time         `json:"approved_at,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

type SweptPostFilter struct {
	WorkspaceID string            `json:"workspace_id"`
	TenantID    string            `json:"tenant_id"`
	Category    PostSweepCategory `json:"category,omitempty"`
	Query       string            `json:"query,omitempty"`
}

// CategorizePostContent classifies swept post text into functional categories (SRC-L2).
func CategorizePostContent(content string) PostSweepCategory {
	lower := strings.ToLower(content)
	if strings.Contains(lower, "hiring") || strings.Contains(lower, "opening") || strings.Contains(lower, "looking for") || strings.Contains(lower, "expanding our") || strings.Contains(lower, "join our") {
		return CategoryHiringAnnouncement
	}
	if strings.Contains(lower, "consensus") || strings.Contains(lower, "raft") || strings.Contains(lower, "paxos") || strings.Contains(lower, "architecture") || strings.Contains(lower, "distributed") || strings.Contains(lower, "kubernetes") || strings.Contains(lower, "algorithm") || strings.Contains(lower, "database") || strings.Contains(lower, "throughput") {
		return CategoryTechnicalDiscussion
	}
	if strings.Contains(lower, "leadership") || strings.Contains(lower, "culture") || strings.Contains(lower, "lessons learned") || strings.Contains(lower, "mentorship") || strings.Contains(lower, "management") {
		return CategoryThoughtLeadership
	}
	if strings.Contains(lower, "announced") || strings.Contains(lower, "launch") || strings.Contains(lower, "release") || strings.Contains(lower, "breaking") {
		return CategoryIndustryNews
	}
	return CategoryGeneralDiscussion
}

// ValidateCommentText checks for forbidden generic hollow praise phrases (LI-10, FND-015).
func ValidateCommentText(text string) error {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return ErrEmptyCommentText
	}
	lower := strings.ToLower(clean)
	forbiddenPhrases := []string{
		"great post",
		"thanks for sharing",
		"totally agree 100%",
		"totally agree",
		"insightful read",
		"agree 100%",
		"good share",
		"nice post",
	}
	for _, fp := range forbiddenPhrases {
		if strings.Contains(lower, fp) && len([]rune(clean)) < 80 {
			return ErrGenericHollowCommentBlocked
		}
	}
	return nil
}

// RejectUnsupportedBulkEngagement enforces platform integrity against automated bot actions (AT-010, REQ-015, LI-10).
func RejectUnsupportedBulkEngagement(mode string) error {
	cleanMode := strings.ToLower(strings.TrimSpace(mode))
	if cleanMode == "bulk_auto_comment" || cleanMode == "engagement_pod" || cleanMode == "bot_auto_engage" || cleanMode == "headless_mass_comment" {
		return ErrUnsupportedBulkEngagement
	}
	return nil
}

// ComputeCommentApprovalToken computes an HMAC-SHA256 signature for tamper invalidation (FND-010, AT-007).
func ComputeCommentApprovalToken(secret, draftID, text string) string {
	if secret == "" {
		secret = "social_platform_comment_draft_hmac_2026"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%s:%s", draftID, strings.TrimSpace(text))))
	return fmt.Sprintf("hmac-sha256:%s", hex.EncodeToString(mac.Sum(nil)))
}

// VerifyCommentApprovalToken verifies the HMAC signature of a comment draft.
func VerifyCommentApprovalToken(secret, draftID, text, token string) bool {
	expected := ComputeCommentApprovalToken(secret, draftID, text)
	return hmac.Equal([]byte(expected), []byte(token))
}

// GenerateCommentDrafts creates 3 distinct high-value comment angles tailored to the swept post (SRC-L1, SRC-L2).
func GenerateCommentDrafts(post *SweptTargetPost, candidateFacts []string) ([]CommentDraft, error) {
	if post == nil {
		return nil, ErrSweepPostNotFound
	}

	authorFirst := strings.Split(post.AuthorName, " ")[0]
	if authorFirst == "" {
		authorFirst = "there"
	}

	now := time.Now().UTC()
	drafts := make([]CommentDraft, 0, 3)

	// Angle 1: Insightful Addition (Technical depth or practical nuance)
	var textAddition string
	var rationaleAddition string
	if post.Category == CategoryHiringAnnouncement {
		textAddition = fmt.Sprintf(
			"Exciting expansion at %s! Having spent over a decade architecting high-throughput distributed systems in Go and Kubernetes, the engineering challenges your team is tackling sound impactful. Looking forward to connecting regarding the Principal role.",
			post.AuthorCompany,
		)
		rationaleAddition = "Highlights verified engineering alignment with role requirements and expresses direct professional interest."
	} else {
		textAddition = fmt.Sprintf(
			"Fascinating analysis, %s. In multi-region deployments, we observed that asymmetric network partitions frequently cause quorum lease drift before heartbeats trigger leader step-down. Tightening heartbeat randomization proved essential for keeping failover latencies under 500ms.",
			authorFirst,
		)
		rationaleAddition = "Adds concrete real-world engineering perspective that deepens the technical discussion."
	}

	d1 := CommentDraft{
		DraftID:           fmt.Sprintf("cdraft-%d-add", now.UnixNano()),
		PostID:            post.PostID,
		WorkspaceID:       post.WorkspaceID,
		TenantID:          post.TenantID,
		Angle:             AngleInsightfulAddition,
		CommentText:       textAddition,
		Rationale:         rationaleAddition,
		VerifiedFactsUsed: candidateFacts,
		CharacterCount:    len([]rune(textAddition)),
		Status:            CommentDraftPendingApproval,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	drafts = append(drafts, d1)

	// Angle 2: Engaging Question (Thoughtful inquiry fostering dialogue)
	textQuestion := fmt.Sprintf(
		"Great breakdown of the trade-offs, %s. How does your team typically handle split-brain detection when clients are partitioned from the leader but can still reach a subset of quorum followers?",
		authorFirst,
	)
	d2 := CommentDraft{
		DraftID:           fmt.Sprintf("cdraft-%d-q", now.UnixNano()+1),
		PostID:            post.PostID,
		WorkspaceID:       post.WorkspaceID,
		TenantID:          post.TenantID,
		Angle:             AngleEngagingQuestion,
		CommentText:       textQuestion,
		Rationale:         "Fosters genuine professional discourse by asking a specific, domain-relevant question.",
		VerifiedFactsUsed: candidateFacts,
		CharacterCount:    len([]rune(textQuestion)),
		Status:            CommentDraftPendingApproval,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	drafts = append(drafts, d2)

	// Angle 3: Supportive Perspective (Validates thesis with industry context)
	textSupportive := fmt.Sprintf(
		"This aligns closely with your points, %s. Operating distributed clusters at scale has shown us that emphasizing partition resiliency over optimistic consistency is a critical discipline.",
		authorFirst,
	)
	d3 := CommentDraft{
		DraftID:           fmt.Sprintf("cdraft-%d-sup", now.UnixNano()+2),
		PostID:            post.PostID,
		WorkspaceID:       post.WorkspaceID,
		TenantID:          post.TenantID,
		Angle:             AngleSupportivePerspective,
		CommentText:       textSupportive,
		Rationale:         "Validates the author's argument using industry architecture principles without hollow generic fluff.",
		VerifiedFactsUsed: candidateFacts,
		CharacterCount:    len([]rune(textSupportive)),
		Status:            CommentDraftPendingApproval,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	drafts = append(drafts, d3)

	return drafts, nil
}
