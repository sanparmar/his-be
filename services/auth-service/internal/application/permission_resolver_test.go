package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/his-platform/auth-service/internal/domain"
)

func TestPermissionResolver_ResolvePermissions(t *testing.T) {
	ctx := context.Background()

	mockUserRoleRepo := &mockUserRoleRepository{}
	mockRoleRepo := &mockRoleRepository{}
	mockPermRepo := &mockPermissionRepository{}

	resolver := NewPermissionResolver(mockUserRoleRepo, mockRoleRepo, mockPermRepo)

	userID := uuid.New()
	tenantID := uuid.New()

	perms, err := resolver.ResolvePermissions(ctx, userID, tenantID)
	if err != nil {
		t.Fatalf("ResolvePermissions() error = %v", err)
	}

	if len(perms) == 0 {
		t.Error("expected permissions, got empty slice")
	}
}

func TestPermissionResolver_HasPermission(t *testing.T) {
	ctx := context.Background()

	mockUserRoleRepo := &mockUserRoleRepository{}
	mockRoleRepo := &mockRoleRepository{}
	mockPermRepo := &mockPermissionRepository{}

	resolver := NewPermissionResolver(mockUserRoleRepo, mockRoleRepo, mockPermRepo)

	userID := uuid.New()
	tenantID := uuid.New()

	hasPerm, err := resolver.HasPermission(ctx, userID, tenantID, "patient:read:own", nil)
	if err != nil {
		t.Fatalf("HasPermission() error = %v", err)
	}

	_ = hasPerm
}

func TestPermissionResolver_GetEffectivePermissions(t *testing.T) {
	ctx := context.Background()

	mockUserRoleRepo := &mockUserRoleRepository{}
	mockRoleRepo := &mockRoleRepository{}
	mockPermRepo := &mockPermissionRepository{}

	resolver := NewPermissionResolver(mockUserRoleRepo, mockRoleRepo, mockPermRepo)

	userID := uuid.New()
	tenantID := uuid.New()

	perms, err := resolver.GetEffectivePermissions(ctx, userID, tenantID)
	if err != nil {
		t.Fatalf("GetEffectivePermissions() error = %v", err)
	}

	if perms == nil {
		t.Error("expected permissions slice, got nil")
	}
}

type mockUserRoleRepository struct{}

func (m *mockUserRoleRepository) Assign(ctx context.Context, ur *domain.UserRole) error          { return nil }
func (m *mockUserRoleRepository) Revoke(ctx context.Context, userID, roleID, tenantID uuid.UUID) error { return nil }
func (m *mockUserRoleRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserRole, error) {
	return []domain.UserRole{}, nil
}
func (m *mockUserRoleRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) {
	return []domain.UserRole{}, nil
}
func (m *mockUserRoleRepository) GetByRole(ctx context.Context, roleID uuid.UUID) ([]domain.UserRole, error) {
	return []domain.UserRole{}, nil
}
func (m *mockUserRoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]domain.UserRole, error) {
	return []domain.UserRole{}, nil
}
func (m *mockUserRoleRepository) IsAssigned(ctx context.Context, userID, roleID, tenantID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockUserRoleRepository) GetActiveByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) {
	return []domain.UserRole{}, nil
}
func (m *mockUserRoleRepository) CountByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (int, error) {
	return 0, nil
}

type mockRoleRepository struct{}

func (m *mockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return &domain.Role{
		ID:   id,
		Name: "doctor",
		Permissions: []domain.Permission{
			{Name: "patient:read:own", Resource: "patient", Action: "read", Scope: "own"},
		},
	}, nil
}
func (m *mockRoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) GetWithPermissions(ctx context.Context, id uuid.UUID) (*domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) GetInheritedRoles(ctx context.Context, roleID uuid.UUID) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) List(ctx context.Context) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) ListByCategory(ctx context.Context, category string) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) Create(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepository) Update(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockRoleRepository) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error { return nil }
func (m *mockRoleRepository) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error { return nil }
func (m *mockRoleRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) { return nil, nil }

type mockPermissionRepository struct{}

func (m *mockPermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) { return nil, nil }
func (m *mockPermissionRepository) GetByName(ctx context.Context, name string) (*domain.Permission, error) { return nil, nil }
func (m *mockPermissionRepository) GetByResourceAction(ctx context.Context, resource, action string) ([]domain.Permission, error) { return nil, nil }
func (m *mockPermissionRepository) ListByCategory(ctx context.Context, category string) ([]domain.Permission, error) { return nil, nil }
func (m *mockPermissionRepository) ListAll(ctx context.Context) ([]domain.Permission, error) { return nil, nil }
func (m *mockPermissionRepository) Create(ctx context.Context, perm *domain.Permission) error { return nil }
func (m *mockPermissionRepository) Update(ctx context.Context, perm *domain.Permission) error { return nil }
func (m *mockPermissionRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }