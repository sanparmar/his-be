package domain

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// mockResolver lets tests control HasPermission's answer and observe
// whether/how it was called, without a real DB or Redis.
type mockResolver struct {
	hasPermission     bool
	hasPermissionErr  error
	calledWith        string
	hasPermissionCall int
}

func (m *mockResolver) ResolvePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]ResolvedPermission, error) {
	return nil, nil
}

func (m *mockResolver) HasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string, resourceCtx *ResourceContext) (bool, error) {
	m.hasPermissionCall++
	m.calledWith = permission
	return m.hasPermission, m.hasPermissionErr
}

func (m *mockResolver) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]Role, error) {
	return nil, nil
}

func (m *mockResolver) GetEffectivePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	return nil, nil
}

func (m *mockResolver) InvalidateCache(ctx context.Context, userID, tenantID uuid.UUID) error {
	return nil
}

func TestCanAccessEffectivePermissions_Self(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	// A caller reading their own effective permissions never needs
	// admin:roles — HasPermission must not even be consulted.
	resolver := &mockResolver{hasPermission: false}

	allowed, err := CanAccessEffectivePermissions(ctx, resolver, userID, userID, tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected a caller to always be allowed to read their own effective permissions")
	}
	if resolver.hasPermissionCall != 0 {
		t.Errorf("expected HasPermission not to be called for self-access, called %d time(s)", resolver.hasPermissionCall)
	}
}

func TestCanAccessEffectivePermissions_OtherUser_Denied(t *testing.T) {
	ctx := context.Background()
	callerID := uuid.New()
	targetID := uuid.New()
	tenantID := uuid.New()

	resolver := &mockResolver{hasPermission: false}

	allowed, err := CanAccessEffectivePermissions(ctx, resolver, callerID, targetID, tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected a caller without admin:roles to be denied reading another user's effective permissions")
	}
	if resolver.hasPermissionCall != 1 {
		t.Errorf("expected HasPermission to be checked exactly once, got %d", resolver.hasPermissionCall)
	}
	if resolver.calledWith != AdminRolesPermission {
		t.Errorf("expected the check to be for %q, got %q", AdminRolesPermission, resolver.calledWith)
	}
}

func TestCanAccessEffectivePermissions_OtherUser_AllowedWithAdminRoles(t *testing.T) {
	ctx := context.Background()
	callerID := uuid.New()
	targetID := uuid.New()
	tenantID := uuid.New()

	resolver := &mockResolver{hasPermission: true}

	allowed, err := CanAccessEffectivePermissions(ctx, resolver, callerID, targetID, tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected a caller with admin:roles to be allowed to read another user's effective permissions")
	}
	if resolver.calledWith != AdminRolesPermission {
		t.Errorf("expected the check to be for %q, got %q", AdminRolesPermission, resolver.calledWith)
	}
}
