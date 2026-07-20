package domain

import (
	"github.com/google/uuid"
)

// UserContext and ResourceContext are consumed by
// transport/grpc/interceptors.PermissionInterceptor. The scope-hierarchy
// evaluation that used to live in this file (own/department/hospital/...)
// depended on a Permission.Scope column that does not exist in the deployed
// schema (see migrations/000004_add_rbac_tables.up.sql's `permissions`
// table: id, resource, action, description, created_at — no scope column),
// so it's been removed rather than adapted to fake data.
type UserContext struct {
	UserID         uuid.UUID
	TenantID       uuid.UUID
	OrganizationID *uuid.UUID
	HospitalID     *uuid.UUID
	DepartmentID   *uuid.UUID
	Roles          []Role
}

type ResourceContext struct {
	ResourceID     uuid.UUID
	ResourceType   string
	TenantID       uuid.UUID
	OrganizationID *uuid.UUID
	HospitalID     *uuid.UUID
	DepartmentID   *uuid.UUID
}
