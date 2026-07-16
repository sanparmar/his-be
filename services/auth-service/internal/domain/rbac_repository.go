package domain

import (
	"context"
	"github.com/google/uuid"
)

type PermissionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Permission, error)
	GetByName(ctx context.Context, name string) (*Permission, error)
	GetByResourceAction(ctx context.Context, resource, action string) ([]Permission, error)
	ListByCategory(ctx context.Context, category string) ([]Permission, error)
	ListAll(ctx context.Context) ([]Permission, error)
	Create(ctx context.Context, perm *Permission) error
	Update(ctx context.Context, perm *Permission) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	GetWithPermissions(ctx context.Context, id uuid.UUID) (*Role, error)
	GetInheritedRoles(ctx context.Context, id uuid.UUID) ([]Role, error)
	List(ctx context.Context) ([]Role, error)
	ListByCategory(ctx context.Context, category string) ([]Role, error)
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddPermission(ctx context.Context, roleID, permID uuid.UUID) error
	RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error
	GetPermissions(ctx context.Context, roleID uuid.UUID) ([]Permission, error)
}

type UserRoleRepository interface {
	Assign(ctx context.Context, ur *UserRole) error
	Revoke(ctx context.Context, userID, roleID, tenantID uuid.UUID) error
	GetByUser(ctx context.Context, userID uuid.UUID) ([]UserRole, error)
	GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]UserRole, error)
	GetByRole(ctx context.Context, roleID uuid.UUID) ([]UserRole, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]UserRole, error)
	IsAssigned(ctx context.Context, userID, roleID, tenantID uuid.UUID) (bool, error)
}