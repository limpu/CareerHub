package identity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	// Too short password
	_, err := HashPassword("short")
	if err != ErrPasswordTooShort {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}

	// Valid password
	password := "SecurePassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !CheckPasswordHash(password, hash) {
		t.Fatal("expected password check to succeed")
	}

	if CheckPasswordHash("WrongPassword", hash) {
		t.Fatal("expected password check to fail for incorrect password")
	}
}

func TestAuthService_SignUpAndLogin(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAuthService(store, "test-secret-key-32-bytes-long!!")
	ctx := context.Background()

	// 1. Sign up
	res, err := svc.SignUp(ctx, "user@example.com", "SecretPass123!", "Test User")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}
	if res.User.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %s", res.User.Email)
	}
	if res.User.PlatformRole != PlatformRoleStandard {
		t.Fatalf("expected standard role, got %s", res.User.PlatformRole)
	}
	if res.Token == "" || res.SessionToken == "" {
		t.Fatal("expected non-empty tokens")
	}

	// 2. Duplicate signup must fail
	_, err = svc.SignUp(ctx, "user@example.com", "OtherPass123!", "Another User")
	if err != ErrUserAlreadyExists {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}

	// 3. Login with wrong password
	_, err = svc.Login(ctx, "user@example.com", "WrongPassword!", "127.0.0.1", "test-agent")
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	// 4. Login with correct password
	loginRes, err := svc.Login(ctx, "user@example.com", "SecretPass123!", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// 5. Validate token
	u, err := svc.ValidateToken(ctx, loginRes.Token)
	if err != nil {
		t.Fatalf("validate token failed: %v", err)
	}
	if u.ID != res.User.ID {
		t.Fatalf("expected user ID %s, got %s", res.User.ID, u.ID)
	}

	// 6. Logout / Revoke session
	err = svc.Logout(ctx, loginRes.SessionToken)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// Token must now be rejected because session is revoked
	_, err = svc.ValidateToken(ctx, loginRes.Token)
	if err != ErrSessionRevoked {
		t.Fatalf("expected ErrSessionRevoked, got %v", err)
	}
}

func TestAuthService_OneTimeBootstrapSuperAdmin(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAuthService(store, "test-secret-key-32-bytes-long!!")
	ctx := context.Background()

	// First bootstrap call must succeed
	admin, err := svc.BootstrapSuperAdmin(ctx, "admin@platform.local", "SuperAdminPass123!", "Root Admin")
	if err != nil {
		t.Fatalf("first bootstrap failed: %v", err)
	}
	if admin.PlatformRole != PlatformRoleSuperAdmin {
		t.Fatalf("expected role super_admin, got %s", admin.PlatformRole)
	}

	// Second bootstrap call must be permanently rejected
	_, err = svc.BootstrapSuperAdmin(ctx, "attacker@platform.local", "AttackerPass123!", "Attacker")
	if err != ErrAlreadyBootstrapped {
		t.Fatalf("expected ErrAlreadyBootstrapped, got %v", err)
	}
}

func TestAuthMiddleware(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAuthService(store, "test-secret-key-32-bytes-long!!")
	ctx := context.Background()

	userRes, _ := svc.SignUp(ctx, "member@platform.local", "Password123!", "Regular Member")
	adminUser, _ := svc.BootstrapSuperAdmin(ctx, "root@platform.local", "AdminPass123!", "Root Admin")
	adminRes, _ := svc.Login(ctx, adminUser.Email, "AdminPass123!", "127.0.0.1", "agent")

	protectedHandler := RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok-auth"))
	})

	superAdminHandler := RequireAuth(RequireSuperAdmin(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok-super-admin"))
	}))

	mw := AuthMiddleware(svc)

	// 1. Unauthenticated request -> 401
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()
	mw(protectedHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	// 2. Regular user accessing protected -> 200
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+userRes.Token)
	rr = httptest.NewRecorder()
	mw(protectedHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// 3. Regular user accessing super admin route -> 403 Forbidden
	req = httptest.NewRequest(http.MethodGet, "/super-admin", nil)
	req.Header.Set("Authorization", "Bearer "+userRes.Token)
	rr = httptest.NewRecorder()
	mw(superAdminHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}

	// 4. Super admin accessing super admin route -> 200 OK
	req = httptest.NewRequest(http.MethodGet, "/super-admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminRes.Token)
	rr = httptest.NewRecorder()
	mw(superAdminHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}