package workspace

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/social-platform/services/core/internal/identity"
)

type Store interface {
	CreateWorkspace(ctx context.Context, w *identity.Workspace) error
	GetWorkspace(ctx context.Context, id string) (*identity.Workspace, error)
	AddMembership(ctx context.Context, m *identity.Membership) error
	GetMembership(ctx context.Context, workspaceID, userID string) (*identity.Membership, error)
	ListUserMemberships(ctx context.Context, userID string) ([]identity.Membership, error)

	SaveGrant(ctx context.Context, grant *ResourceGrant) error
	GetGrant(ctx context.Context, id string) (*ResourceGrant, error)
	DeleteGrant(ctx context.Context, id string) error
	FindActiveGrant(ctx context.Context, workspaceID, granteeID, resourceID string, perm Permission) (*ResourceGrant, error)
}

type MemoryStore struct {
	mu           sync.RWMutex
	workspaces   map[string]*identity.Workspace
	memberships  map[string]*identity.Membership // key: workspaceID + ":" + userID
	grants       map[string]*ResourceGrant       // id -> grant
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		workspaces:  make(map[string]*identity.Workspace),
		memberships: make(map[string]*identity.Membership),
		grants:      make(map[string]*ResourceGrant),
	}
}

func (m *MemoryStore) CreateWorkspace(ctx context.Context, w *identity.Workspace) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workspaces[w.ID] = w
	return nil
}

func (m *MemoryStore) GetWorkspace(ctx context.Context, id string) (*identity.Workspace, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, exists := m.workspaces[id]
	if !exists {
		return nil, ErrWorkspaceNotFound
	}
	return w, nil
}

func (m *MemoryStore) AddMembership(ctx context.Context, mem *identity.Membership) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := mem.WorkspaceID + ":" + mem.UserID
	m.memberships[key] = mem
	return nil
}

func (m *MemoryStore) GetMembership(ctx context.Context, workspaceID, userID string) (*identity.Membership, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := workspaceID + ":" + userID
	mem, exists := m.memberships[key]
	if !exists {
		return nil, ErrNotWorkspaceMember
	}
	return mem, nil
}

func (m *MemoryStore) ListUserMemberships(ctx context.Context, userID string) ([]identity.Membership, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []identity.Membership
	for _, mem := range m.memberships {
		if mem.UserID == userID {
			list = append(list, *mem)
		}
	}
	return list, nil
}

func (m *MemoryStore) SaveGrant(ctx context.Context, grant *ResourceGrant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.grants[grant.ID] = grant
	return nil
}

func (m *MemoryStore) GetGrant(ctx context.Context, id string) (*ResourceGrant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	g, exists := m.grants[id]
	if !exists {
		return nil, ErrGrantNotFound
	}
	return g, nil
}

func (m *MemoryStore) DeleteGrant(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.grants, id)
	return nil
}

func (m *MemoryStore) FindActiveGrant(ctx context.Context, workspaceID, granteeID, resourceID string, perm Permission) (*ResourceGrant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now().UTC()
	for _, g := range m.grants {
		if g.WorkspaceID == workspaceID && g.GranteeID == granteeID && g.ResourceID == resourceID {
			if g.ExpiresAt != nil && now.After(*g.ExpiresAt) {
				continue // expired
			}
			return g, nil
		}
	}
	return nil, ErrGrantNotFound
}

type WorkspaceService struct {
	store Store
}

func NewWorkspaceService(store Store) *WorkspaceService {
	return &WorkspaceService{store: store}
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, ownerID, name, slug string) (*identity.Workspace, error) {
	now := time.Now().UTC()
	wsID := newUUID()
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(name, " ", "-")) + "-" + wsID[:6]
	}

	ws := &identity.Workspace{
		ID:        wsID,
		Name:      name,
		Slug:      slug,
		OwnerID:   ownerID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.store.CreateWorkspace(ctx, ws); err != nil {
		return nil, err
	}

	// Owner is automatically workspace Admin
	_ = s.store.AddMembership(ctx, &identity.Membership{
		WorkspaceID: ws.ID,
		UserID:      ownerID,
		Role:        identity.WorkspaceRoleAdmin,
		CreatedAt:   now,
	})

	return ws, nil
}

func (s *WorkspaceService) AddMember(ctx context.Context, actorID, workspaceID, targetUserID string, role identity.WorkspaceRole) error {
	// Actor must be Admin in the workspace
	mem, err := s.store.GetMembership(ctx, workspaceID, actorID)
	if err != nil || mem.Role != identity.WorkspaceRoleAdmin {
		return errors.New("only workspace admins can add members")
	}

	now := time.Now().UTC()
	return s.store.AddMembership(ctx, &identity.Membership{
		WorkspaceID: workspaceID,
		UserID:      targetUserID,
		Role:        role,
		CreatedAt:   now,
	})
}

// IsAdmin checks if a user is an administrator of a workspace.
func (s *WorkspaceService) IsAdmin(ctx context.Context, workspaceID, userID string) bool {
	mem, err := s.store.GetMembership(ctx, workspaceID, userID)
	return err == nil && mem != nil && mem.Role == identity.WorkspaceRoleAdmin
}

// CreateGrant delegates access from the resource owner to another member.
func (s *WorkspaceService) CreateGrant(ctx context.Context, granterID, workspaceID, granteeID, resourceType, resourceID string, perm Permission, duration time.Duration) (*ResourceGrant, error) {
	// Verify granter & grantee are members of the workspace
	if _, err := s.store.GetMembership(ctx, workspaceID, granterID); err != nil {
		return nil, ErrNotWorkspaceMember
	}
	if _, err := s.store.GetMembership(ctx, workspaceID, granteeID); err != nil {
		return nil, ErrNotWorkspaceMember
	}

	now := time.Now().UTC()
	var expiresAt *time.Time
	if duration > 0 {
		exp := now.Add(duration)
		expiresAt = &exp
	}

	grant := &ResourceGrant{
		ID:           newUUID(),
		WorkspaceID:  workspaceID,
		GranterID:    granterID,
		GranteeID:    granteeID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Permission:   perm,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
	}

	if err := s.store.SaveGrant(ctx, grant); err != nil {
		return nil, err
	}
	return grant, nil
}

func (s *WorkspaceService) RevokeGrant(ctx context.Context, actorID, grantID string) error {
	grant, err := s.store.GetGrant(ctx, grantID)
	if err != nil {
		return err
	}

	// Only the granter (owner) or the grantee can revoke a grant
	if actorID != grant.GranterID && actorID != grant.GranteeID {
		return errors.New("unauthorized to revoke this grant")
	}

	return s.store.DeleteGrant(ctx, grantID)
}

// AuthorizeResourceAccess enforces AT-011:
// An admin CANNOT read another member's private resume without an explicit grant.
func (s *WorkspaceService) AuthorizeResourceAccess(
	ctx context.Context,
	actorID string,
	workspaceID string,
	resourceOwnerID string,
	resourceType string,
	resourceID string,
	requiredPerm Permission,
) error {
	// 1. Owner always has full access to their own resource
	if actorID == resourceOwnerID {
		return nil
	}

	// 2. Actor must be a member of the workspace
	actorMem, err := s.store.GetMembership(ctx, workspaceID, actorID)
	if err != nil {
		return ErrNotWorkspaceMember
	}

	// 3. For owner-private resources (like 'resume', 'profile', 'application'),
	// workspace ADMIN role DOES NOT imply access!
	// Must have an explicit active grant.
	_, err = s.store.FindActiveGrant(ctx, workspaceID, actorID, resourceID, requiredPerm)
	if err != nil {
		// Explicit failure under AT-011 / REQ-018:
		return fmt.Errorf("%w: user %s has role %s but lacks explicit delegation on %s:%s",
			ErrAccessDenied, actorID, actorMem.Role, resourceType, resourceID)
	}

	return nil
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}