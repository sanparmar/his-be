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

type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByNameAndTenant(ctx context.Context, name string, tenantID uuid.UUID) (*Role, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Role, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*Role, error)
}

type UserRoleRepository interface {
	Add(ctx context.Context, userID, roleID uuid.UUID) error
	Remove(ctx context.Context, userID, roleID uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*UserRole, error)
	GetByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]*UserRole, error)
}

type MFARepository interface {
	Create(ctx context.Context, mfa *MFACredential) error
	GetByUserID(ctx context.Context, userID uuid.UUID, mfaType MFAType) (*MFACredential, error)
	Update(ctx context.Context, mfa *MFACredential) error
	Delete(ctx context.Context, userID uuid.UUID, mfaType MFAType) error
}
