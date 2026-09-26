package workspace

import (
	"errors"
	"time"

	"github.com/social-platform/services/core/internal/identity"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrNotWorkspaceMember = errors.New("user is not a member of this workspace")
	ErrAccessDenied      = errors.New("access denied: resource is owner-private and requires an explicit active grant")
	ErrGrantNotFound     = errors.New("resource grant not found")
	ErrGrantExpired      = errors.New("resource grant has expired")
	ErrUnauthorizedGrant = errors.New("only the resource owner can grant access to another member")
)

type ResourceClass string

const (
	ResourceClassOwnerPrivate    ResourceClass = "owner_private"
	ResourceClassWorkspaceShared ResourceClass = "workspace_shared"
	ResourceClassDelegated       ResourceClass = "delegated"
)

type Permission string

const (
	PermissionRead    Permission = "read"
	PermissionReview  Permission = "review"
	PermissionComment Permission = "comment"
	PermissionAdmin   Permission = "admin"
)

type ResourceGrant struct {
	ID           string     `json:"id"`
	WorkspaceID  string     `json:"workspace_id"`
	GranterID    string     `json:"granter_id"`
	GranteeID    string     `json:"grantee_id"`
	ResourceType string     `json:"resource_type"` // e.g., 'resume', 'profile', 'application'
	ResourceID   string     `json:"resource_id"`
	Permission   Permission `json:"permission"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type WorkspaceInfo struct {
	Workspace  identity.Workspace   `json:"workspace"`
	UserRole   identity.WorkspaceRole `json:"user_role"`
}