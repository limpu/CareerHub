package workspace

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/social-platform/services/core/internal/identity"
)

func TestWorkspace_AT011_AdminCannotAccessPrivateResumeWithoutGrant(t *testing.T) {
	store := NewMemoryStore()
	svc := NewWorkspaceService(store)
	ctx := context.Background()

	adminID := "admin-user-001"
	memberID := "member-user-002"
	outsiderID := "outsider-user-999"

	// 1. Admin creates a team workspace
	ws, err := svc.CreateWorkspace(ctx, adminID, "Engineering Workspace", "eng-ws")
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// 2. Admin adds regular member to workspace
	err = svc.AddMember(ctx, adminID, ws.ID, memberID, identity.WorkspaceRoleUser)
	if err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	memberResumeID := "resume-uuid-xyz-456"

	// 3. Member accessing their own resume -> MUST BE ALLOWED
	err = svc.AuthorizeResourceAccess(ctx, memberID, ws.ID, memberID, "resume", memberResumeID, PermissionRead)
	if err != nil {
		t.Fatalf("expected owner to access their own resume, got: %v", err)
	}

	// 4. AT-011 CRITICAL TEST:
	// Admin attempts to access member's private resume WITHOUT an explicit grant -> MUST BE DENIED!
	err = svc.AuthorizeResourceAccess(ctx, adminID, ws.ID, memberID, "resume", memberResumeID, PermissionRead)
	if err == nil || !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("AT-011 VIOLATION: expected ErrAccessDenied for admin without grant, got: %v", err)
	}

	// 5. Member explicitly grants 'review' permission to the Admin for 1 hour
	grant, err := svc.CreateGrant(ctx, memberID, ws.ID, adminID, "resume", memberResumeID, PermissionReview, time.Hour)
	if err != nil {
		t.Fatalf("failed to create grant: %v", err)
	}
	if grant.GranteeID != adminID || grant.ResourceID != memberResumeID {
		t.Fatal("grant details mismatch")
	}

	// 6. Admin attempts to access member's resume WITH active grant -> MUST BE ALLOWED
	err = svc.AuthorizeResourceAccess(ctx, adminID, ws.ID, memberID, "resume", memberResumeID, PermissionReview)
	if err != nil {
		t.Fatalf("expected admin with active grant to have access, got: %v", err)
	}

	// 7. Member explicitly revokes the grant
	err = svc.RevokeGrant(ctx, memberID, grant.ID)
	if err != nil {
		t.Fatalf("failed to revoke grant: %v", err)
	}

	// 8. Admin attempts to access resume after revocation -> MUST BE DENIED AGAIN
	err = svc.AuthorizeResourceAccess(ctx, adminID, ws.ID, memberID, "resume", memberResumeID, PermissionReview)
	if err == nil || !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("expected ErrAccessDenied after grant revocation, got: %v", err)
	}

	// 9. Outsider (non-workspace member) attempts access -> MUST BE REJECTED
	err = svc.AuthorizeResourceAccess(ctx, outsiderID, ws.ID, memberID, "resume", memberResumeID, PermissionRead)
	if err == nil || !errors.Is(err, ErrNotWorkspaceMember) {
		t.Fatalf("expected ErrNotWorkspaceMember for non-member, got: %v", err)
	}
}

func TestWorkspace_GrantExpiration(t *testing.T) {
	store := NewMemoryStore()
	svc := NewWorkspaceService(store)
	ctx := context.Background()

	adminID := "admin-user-001"
	memberID := "member-user-002"

	ws, _ := svc.CreateWorkspace(ctx, adminID, "Dev Workspace", "dev-ws")
	_ = svc.AddMember(ctx, adminID, ws.ID, memberID, identity.WorkspaceRoleUser)

	resumeID := "resume-expiring-999"

	// Create grant that expires in 10 milliseconds
	_, err := svc.CreateGrant(ctx, memberID, ws.ID, adminID, "resume", resumeID, PermissionRead, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create grant: %v", err)
	}

	// Immediately should be valid
	err = svc.AuthorizeResourceAccess(ctx, adminID, ws.ID, memberID, "resume", resumeID, PermissionRead)
	if err != nil {
		t.Fatalf("expected grant to be active immediately, got: %v", err)
	}

	// Wait for expiration
	time.Sleep(25 * time.Millisecond)

	// Now should be expired and denied
	err = svc.AuthorizeResourceAccess(ctx, adminID, ws.ID, memberID, "resume", resumeID, PermissionRead)
	if err == nil || !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("expected ErrAccessDenied after grant expiration, got: %v", err)
	}
}