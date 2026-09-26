package identity

import "time"

type PlatformRole string

const (
	PlatformRoleSuperAdmin PlatformRole = "super_admin"
	PlatformRoleOperator   PlatformRole = "operator"
	PlatformRoleStandard   PlatformRole = "standard"
)

type WorkspaceRole string

const (
	WorkspaceRoleAdmin WorkspaceRole = "admin"
	WorkspaceRoleUser  WorkspaceRole = "user"
)

type User struct {
	ID           string       `json:"id"`
	Email        string       `json:"email"`
	PasswordHash string       `json:"-"`
	FullName     string       `json:"full_name"`
	PlatformRole PlatformRole `json:"platform_role"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Membership struct {
	WorkspaceID string        `json:"workspace_id"`
	UserID      string        `json:"user_id"`
	Role        WorkspaceRole `json:"role"`
	CreatedAt   time.Time     `json:"created_at"`
}

type Session struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	IPAddress string     `json:"ip_address,omitempty"`
	UserAgent string     `json:"user_agent,omitempty"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}