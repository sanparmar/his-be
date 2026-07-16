package domain

import "errors"

var (
	ErrPermissionDenied     = errors.New("permission denied")
	ErrRoleNotFound         = errors.New("role not found")
	ErrPermissionNotFound   = errors.New("permission not found")
	ErrUserRoleNotFound     = errors.New("user role assignment not found")
	ErrRoleAlreadyAssigned  = errors.New("role already assigned to user")
	ErrRoleAssignmentExpired = errors.New("role assignment has expired")
	ErrInvalidRoleHierarchy = errors.New("invalid role hierarchy")
	ErrSystemRoleModification = errors.New("cannot modify system role")
	ErrInvalidScope         = errors.New("invalid scope for resource")
	ErrTenantMismatch       = errors.New("tenant mismatch in role assignment")
	ErrPermissionNotGranted = errors.New("permission not granted to role")
	ErrCircularRoleHierarchy = errors.New("circular role hierarchy detected")
	ErrInvalidPermission    = errors.New("invalid permission format")
)