package application

import (
	"context"
	"testing"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
)

func TestAssignRoleUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	mockUserRoleRepo := &mockUserRoleRepository{}
	mockRoleRepo := &mockRoleRepository{
		role: &domain.Role{
			ID:       uuid.MustParse("88888888-8888-8888-8888-888888888888"),
			Name:     "doctor",
			IsSystem: true,
		},
	}
	mockPermResolver := &mockPermissionResolver{}

	assignUC := NewAssignRoleUseCase(mockUserRoleRepo, mockRoleRepo, mockPermResolver)

	req := AssignRoleRequest{
		UserID:         uuid.New(),
		RoleID:         uuid.MustParse("88888888-8888-8888-8888-888888888888"),
		TenantID:       uuid.New(),
		OrganizationID: uuidPtr(uuid.New()),
		AssignedBy:     uuid.New(),
	}

	err := assignUC.Execute(ctx, req)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func TestAssignRoleUseCase_RoleNotFound(t *testing.T) {
	ctx := context.Background()

	mockUserRoleRepo := &mockUserRoleRepository{}
	mockRoleRepo := &mockRoleRepository{role: nil}
	mockPermResolver := &mockPermissionResolver{}

	assignUC := NewAssignRoleUseCase(mockUserRoleRepo, mockRoleRepo, mockPermResolver)

	req := AssignRoleRequest{
		UserID:     uuid.New(),
		RoleID:     uuid.New(),
		TenantID:   uuid.New(),
		AssignedBy: uuid.New(),
	}

	err := assignUC.Execute(ctx, req)
	if err == nil {
		t.Error("expected error for role not found, got nil")
	}
}

func TestRevokeRoleUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	mockUserRoleRepo := &mockUserRoleRepository{}
	mockPermResolver := &mockPermissionResolver{}

	revokeUC := NewRevokeRoleUseCase(mockUserRoleRepo, mockPermResolver)

	req := RevokeRoleRequest{
		UserID:   uuid.New(),
		RoleID:   uuid.New(),
		TenantID: uuid.New(),
	}

	err := revokeUC.Execute(ctx, req)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

type mockUserRoleRepository struct {
	assignErr error
	revokeErr error
}

func (m *mockUserRoleRepository) Assign(ctx context.Context, ur *domain.UserRole) error          { return m.assignErr }
func (m *mockUserRoleRepository) Revoke(ctx context.Context, userID, roleID, tenantID uuid.UUID) error { return m.revokeErr }
func (m *mockUserRoleRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) GetByRole(ctx context.Context, roleID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) IsAssigned(ctx context.Context, userID, roleID, tenantID uuid.UUID) (bool, error) { return false, nil }
func (m *mockUserRoleRepository) GetActiveByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) CountByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (int, error) { return 0, nil }

type mockRoleRepository struct {
	role *domain.Role
}

func (m *mockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) { return m.role, nil }
func (m *mockRoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) GetWithPermissions(ctx context.Context, id uuid.UUID) (*domain.Role, error) { return m.role, nil }
func (m *mockRoleRepository) GetInheritedRoles(ctx context.Context, roleID uuid.UUID) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) List(ctx context.Context) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) ListByCategory(ctx context.Context, category string) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) Create(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepository) Update(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockRoleRepository) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error { return nil }
func (m *mockRoleRepository) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error { return nil }
func (m *mockRoleRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) { return nil, nil }

type mockPermissionResolver struct{}

func (m *mockPermissionResolver) ResolvePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]ResolvedPermission, error) { return nil, nil }
func (m *mockPermissionResolver) HasPermission(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, permission string, resourceCtx *domain.ResourceContext) (bool, error) { return true, nil }
func (m *mockPermissionResolver) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.Role, error) { return nil, nil }
func (m *mockPermissionResolver) GetEffectivePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) { return nil, nil }
func (m *mockPermissionResolver) InvalidateCache(userID, tenantID uuid.UUID) {}
func (m *mockPermissionResolver) GetCachedPermissions(userID, tenantID uuid.UUID) ([]ResolvedPermission, bool) { return nil, false }