package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Category  string    `json:"category" db:"category"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UserRole struct {
	UserID uuid.UUID `json:"user_id" db:"user_id"`
	RoleID uuid.UUID `json:"role_id" db:"role_id"`
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