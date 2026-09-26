package social

import "time"

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPending   PostStatus = "pending_approval"
	PostStatusScheduled PostStatus = "scheduled"
	PostStatusPublished PostStatus = "published"
	PostStatusFailed    PostStatus = "failed"
)

type Post struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	AuthorID    string     `json:"author_id"`
	Content     string     `json:"content"`
	Platforms   []string   `json:"platforms"`
	Status      PostStatus `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
