package identity

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrUserAlreadyExists   = errors.New("user with this email already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrSessionRevoked      = errors.New("session has been revoked")
	ErrSessionExpired      = errors.New("session has expired")
	ErrAlreadyBootstrapped = errors.New("super admin already exists; bootstrap is permanently disabled")
)

type Store interface {
	CreateUser(ctx context.Context, u *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	CountSuperAdmins(ctx context.Context) (int, error)

	CreateSession(ctx context.Context, s *Session) error
	GetSessionByHash(ctx context.Context, hash string) (*Session, error)
	RevokeSession(ctx context.Context, hash string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error

	CreateWorkspace(ctx context.Context, w *Workspace) error
	AddMembership(ctx context.Context, m *Membership) error
}

type MemoryStore struct {
	mu         sync.RWMutex
	users      map[string]*User       // id -> user
	usersByEmail map[string]*User     // email -> user
	sessions   map[string]*Session    // hash -> session
	workspaces map[string]*Workspace  // id -> workspace
	memberships []Membership
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:        make(map[string]*User),
		usersByEmail: make(map[string]*User),
		sessions:     make(map[string]*Session),
		workspaces:   make(map[string]*Workspace),
		memberships:  make([]Membership, 0),
	}
}

func (m *MemoryStore) CreateUser(ctx context.Context, u *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	normEmail := strings.ToLower(strings.TrimSpace(u.Email))
	if _, exists := m.usersByEmail[normEmail]; exists {
		return ErrUserAlreadyExists
	}
	m.users[u.ID] = u
	m.usersByEmail[normEmail] = u
	return nil
}

func (m *MemoryStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	normEmail := strings.ToLower(strings.TrimSpace(email))
	u, exists := m.usersByEmail[normEmail]
	if !exists {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *MemoryStore) GetUserByID(ctx context.Context, id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, exists := m.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *MemoryStore) CountSuperAdmins(ctx context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, u := range m.users {
		if u.PlatformRole == PlatformRoleSuperAdmin {
			count++
		}
	}
	return count, nil
}

func (m *MemoryStore) CreateSession(ctx context.Context, s *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[s.TokenHash] = s
	return nil
}

func (m *MemoryStore) GetSessionByHash(ctx context.Context, hash string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, exists := m.sessions[hash]
	if !exists {
		return nil, errors.New("session not found")
	}
	return s, nil
}

func (m *MemoryStore) RevokeSession(ctx context.Context, hash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, exists := m.sessions[hash]
	if !exists {
		return nil
	}
	now := time.Now().UTC()
	s.RevokedAt = &now
	return nil
}

func (m *MemoryStore) RevokeAllUserSessions(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	for _, s := range m.sessions {
		if s.UserID == userID && s.RevokedAt == nil {
			s.RevokedAt = &now
		}
	}
	return nil
}

func (m *MemoryStore) CreateWorkspace(ctx context.Context, w *Workspace) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.workspaces[w.ID] = w
	return nil
}

func (m *MemoryStore) AddMembership(ctx context.Context, mem *Membership) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.memberships = append(m.memberships, *mem)
	return nil
}

type AuthService struct {
	store     Store
	jwtSecret string
}

func NewAuthService(store Store, jwtSecret string) *AuthService {
	if jwtSecret == "" {
		jwtSecret = "default_jwt_secret_must_change_in_production"
	}
	return &AuthService{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

type AuthResult struct {
	User         *User    `json:"user"`
	Token        string   `json:"token"`
	SessionToken string   `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (s *AuthService) SignUp(ctx context.Context, email, password, fullName string) (*AuthResult, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	uid := newUUID()
	now := time.Now().UTC()

	user := &User{
		ID:           uid,
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: hash,
		FullName:     strings.TrimSpace(fullName),
		PlatformRole: PlatformRoleStandard,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Create personal default workspace
	wsID := newUUID()
	workspace := &Workspace{
		ID:        wsID,
		Name:      fmt.Sprintf("%s's Workspace", user.FullName),
		Slug:      fmt.Sprintf("ws-%s", uid[:8]),
		OwnerID:   user.ID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = s.store.CreateWorkspace(ctx, workspace)
	_ = s.store.AddMembership(ctx, &Membership{
		WorkspaceID: wsID,
		UserID:      user.ID,
		Role:        WorkspaceRoleAdmin,
		CreatedAt:   now,
	})

	return s.createSessionAndJWT(ctx, user, "", "")
}

func (s *AuthService) Login(ctx context.Context, email, password, ip, userAgent string) (*AuthResult, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	return s.createSessionAndJWT(ctx, user, ip, userAgent)
}

func (s *AuthService) Logout(ctx context.Context, sessionToken string) error {
	hash := HashSessionToken(sessionToken)
	return s.store.RevokeSession(ctx, hash)
}

func (s *AuthService) RevokeAllSessions(ctx context.Context, userID string) error {
	return s.store.RevokeAllUserSessions(ctx, userID)
}

func (s *AuthService) Me(ctx context.Context, userID string) (*User, error) {
	return s.store.GetUserByID(ctx, userID)
}

// BootstrapSuperAdmin creates the initial Super Admin. If one already exists, it is permanently disabled.
func (s *AuthService) BootstrapSuperAdmin(ctx context.Context, email, password, fullName string) (*User, error) {
	count, err := s.store.CountSuperAdmins(ctx)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrAlreadyBootstrapped
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := &User{
		ID:           newUUID(),
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: hash,
		FullName:     strings.TrimSpace(fullName),
		PlatformRole: PlatformRoleSuperAdmin,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*User, error) {
	claims, err := VerifyJWT(tokenString, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Verify session in store
	sess, err := s.store.GetSessionByHash(ctx, claims.SessionID)
	if err != nil || sess.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if time.Now().UTC().After(sess.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return s.store.GetUserByID(ctx, claims.UserID)
}

func (s *AuthService) createSessionAndJWT(ctx context.Context, user *User, ip, userAgent string) (*AuthResult, error) {
	sessionToken, err := GenerateSessionToken()
	if err != nil {
		return nil, err
	}
	tokenHash := HashSessionToken(sessionToken)

	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)

	session := &Session{
		ID:        newUUID(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		IPAddress: ip,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	claims := JWTClaims{
		UserID:       user.ID,
		Email:        user.Email,
		PlatformRole: user.PlatformRole,
		SessionID:    tokenHash,
		IssuedAt:     now.Unix(),
		ExpiresAt:    expiresAt.Unix(),
	}

	jwtStr, err := SignJWT(claims, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		Token:        jwtStr,
		SessionToken: sessionToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}