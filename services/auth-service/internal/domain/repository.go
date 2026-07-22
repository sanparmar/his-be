package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Update(ctx context.Context, user *User) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByToken(ctx context.Context, token string) (*Session, error)
	Delete(ctx context.Context, token string) error
	UpdateLastActivity(ctx context.Context, token string) error
	UpdateToken(ctx context.Context, oldToken, newToken string, expiresAt time.Time) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type MFARepository interface {
	Create(ctx context.Context, mfa *MFACredential) error
	GetByUserID(ctx context.Context, userID uuid.UUID, mfaType MFAType) (*MFACredential, error)
	Update(ctx context.Context, mfa *MFACredential) error
	Delete(ctx context.Context, userID uuid.UUID, mfaType MFAType) error
}

type PermissionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Permission, error)
	GetByName(ctx context.Context, name string) (*Permission, error)
	GetByResourceAction(ctx context.Context, resource, action string) ([]Permission, error)
	ListByCategory(ctx context.Context, category string) ([]Permission, error)
	ListAll(ctx context.Context) ([]Permission, error)
	Create(ctx context.Context, perm *Permission) error
}

type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	GetWithPermissions(ctx context.Context, id uuid.UUID) (*Role, error)
	GetInheritedRoles(ctx context.Context, roleID uuid.UUID) ([]Role, error)
	List(ctx context.Context) ([]Role, error)
	ListByCategory(ctx context.Context, category string) ([]Role, error)
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	AddPermission(ctx context.Context, roleID, permID uuid.UUID) error
	RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error
	GetPermissions(ctx context.Context, roleID uuid.UUID) ([]Permission, error)
}

type UserRoleRepository interface {
	Assign(ctx context.Context, ur *UserRole) error
	Revoke(ctx context.Context, userID, roleID, tenantID uuid.UUID) error
	GetByUser(ctx context.Context, userID uuid.UUID) ([]UserRole, error)
	GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]UserRole, error)
	GetActiveByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]UserRole, error)
	ListByRole(ctx context.Context, roleID uuid.UUID) ([]UserRole, error)
	CountByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (int, error)
}
