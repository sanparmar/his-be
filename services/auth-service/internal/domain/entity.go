package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role matches the deployed schema (roles table: id, name, description,
// tenant_id, is_system_role, created_at, updated_at — no category column).
type Role struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	TenantID     uuid.UUID `json:"tenant_id" db:"tenant_id"`
	IsSystemRole bool      `json:"is_system_role" db:"is_system_role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type UserRole struct {
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	RoleID     uuid.UUID  `json:"role_id" db:"role_id"`
	AssignedAt time.Time  `json:"assigned_at" db:"assigned_at"`
	AssignedBy *uuid.UUID `json:"assigned_by,omitempty" db:"assigned_by"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

func (ur *UserRole) IsExpired() bool {
	if ur.ExpiresAt == nil {
		return false
	}
	return ur.ExpiresAt.Before(time.Now())
}

// Permission matches the deployed schema exactly (see
// migrations/000004_add_rbac_tables.up.sql's `permissions` table: id,
// resource, action, description, created_at — no name/scope/category
// columns). Name is derived in-memory (resource:action), not persisted.
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Name        string    `json:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Username       string     `json:"username" db:"username"`
	Email          string     `json:"email" db:"email"`
	PasswordHash   string     `json:"-" db:"password_hash"`
	TenantID       uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" db:"organization_id"`
	HospitalID     *uuid.UUID `json:"hospital_id,omitempty" db:"hospital_id"`
	DepartmentID   *uuid.UUID `json:"department_id,omitempty" db:"department_id"`
	Roles          []Role     `json:"roles,omitempty"`
	Permissions    []string   `json:"permissions,omitempty"`
	PermVersion    int64      `json:"perm_version,omitempty"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type Session struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
}

type MFAType string

const (
	MFATypeTOTP     MFAType = "totp"
	MFATypeWebAuthn MFAType = "webauthn"
)

type MFACredential struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Type        MFAType   `json:"type" db:"type"`
	Secret      string    `json:"-" db:"secret"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	BackupCodes []string  `json:"-" db:"backup_codes"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}