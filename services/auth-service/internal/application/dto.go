package application

import "github.com/google/uuid"

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type UserResponse struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	TenantID       uuid.UUID `json:"tenant_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	HospitalID     uuid.UUID `json:"hospital_id"`
	Roles          []string  `json:"roles"`
	Permissions    []string  `json:"permissions,omitempty"`
}

type AssignRoleRequest struct {
	UserID         uuid.UUID  `json:"user_id" validate:"required"`
	RoleID         uuid.UUID  `json:"role_id" validate:"required"`
	TenantID       uuid.UUID  `json:"tenant_id" validate:"required"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	HospitalID     *uuid.UUID `json:"hospital_id,omitempty"`
	DepartmentID   *uuid.UUID `json:"department_id,omitempty"`
	AssignedBy     uuid.UUID  `json:"assigned_by" validate:"required"`
	ExpiresAt      *string    `json:"expires_at,omitempty"`
}

type RevokeRoleRequest struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	RoleID   uuid.UUID `json:"role_id" validate:"required"`
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
}

type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Category    string    `json:"category"`
	IsSystem    bool      `json:"is_system"`
	Description string    `json:"description"`
}

type PermissionResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Scope       string `json:"scope"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type UserPermissionsResponse struct {
	UserID       uuid.UUID            `json:"user_id"`
	TenantID     uuid.UUID            `json:"tenant_id"`
	Roles        []RoleResponse       `json:"roles"`
	Permissions  []PermissionResponse `json:"permissions"`
	EffectivePerms []string           `json:"effective_permissions"`
}

type ListRolesRequest struct {
	Category string `json:"category,omitempty"`
}

type ListPermissionsRequest struct {
	Category string `json:"category,omitempty"`
	Resource string `json:"resource,omitempty"`
}