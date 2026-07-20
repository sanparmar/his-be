package domain

import (
	"context"

	"github.com/google/uuid"
)

type ResolvedPermission struct {
	Permission
	GrantedByRole string `json:"granted_by_role"`
}

type PermissionResolver interface {
	ResolvePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]ResolvedPermission, error)
	HasPermission(ctx context.Context, userID uuid.UUID, userTenantID uuid.UUID, permission string, resourceCtx *ResourceContext) (bool, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]Role, error)
	GetEffectivePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]string, error)
	InvalidateCache(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) error
}

// AdminRolesPermission is the existing permission ("Manage roles and
// permissions" in migrations/000006_seed_demo_rbac.up.sql) required to view
// another user's effective permissions — see CanAccessEffectivePermissions.
// Also what interceptors/permission.go requires for AssignRoles.
const AdminRolesPermission = "admin:roles"

// CanAccessEffectivePermissions decides whether callerID may read
// targetID's effective permissions: always true for a user reading their
// own (no extra permission needed — this is self-service, like reading
// your own profile), otherwise only with AdminRolesPermission. Without
// this, GetEffectivePermissions let any authenticated caller pass any
// user_id and read that user's roles/permissions.
func CanAccessEffectivePermissions(ctx context.Context, resolver PermissionResolver, callerID, targetID, tenantID uuid.UUID) (bool, error) {
	if callerID == targetID {
		return true, nil
	}
	return resolver.HasPermission(ctx, callerID, tenantID, AdminRolesPermission, &ResourceContext{TenantID: tenantID})
}

type PermCache interface {
	Get(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error)
	Set(ctx context.Context, userID, tenantID uuid.UUID, perms []string) error
	Invalidate(ctx context.Context, userID, tenantID uuid.UUID) error
}

type permissionResolver struct {
	userRoleRepo   UserRoleRepository
	roleRepo       RoleRepository
	permissionRepo PermissionRepository
	permCache      PermCache
}

func NewPermissionResolver(
	userRoleRepo UserRoleRepository,
	roleRepo RoleRepository,
	permissionRepo PermissionRepository,
	permCache PermCache,
) PermissionResolver {
	return &permissionResolver{
		userRoleRepo:   userRoleRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		permCache:      permCache,
	}
}

// GetUserRoles resolves a user's roles via user_roles -> roles, scoped to
// tenant and excluding expired assignments — the real, deployed join path
// (no Role.Permissions field, no UserRole tenant column: see entity.go).
func (r *permissionResolver) GetUserRoles(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]Role, error) {
	userRoles, err := r.userRoleRepo.GetActiveByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]uuid.UUID, 0, len(userRoles))
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	roles, err := r.roleRepo.GetByIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	result := make([]Role, 0, len(roles))
	for _, role := range roles {
		result = append(result, *role)
	}
	return result, nil
}

func (r *permissionResolver) ResolvePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]ResolvedPermission, error) {
	if r.permCache != nil {
		cached, err := r.permCache.Get(ctx, userID, tenantID)
		if err == nil && cached != nil {
			perms := make([]ResolvedPermission, len(cached))
			for i, name := range cached {
				perms[i] = ResolvedPermission{Permission: Permission{Name: name}}
			}
			return perms, nil
		}
	}

	roles, err := r.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]uuid.UUID, len(roles))
	roleByID := make(map[uuid.UUID]Role, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.ID
		roleByID[role.ID] = role
	}

	perms, err := r.permissionRepo.GetByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	// Dedupe by permission name (a permission can be granted by more than one role).
	permMap := make(map[string]ResolvedPermission, len(perms))
	for _, perm := range perms {
		if _, exists := permMap[perm.Name]; exists {
			continue
		}
		permMap[perm.Name] = ResolvedPermission{Permission: perm}
	}

	result := make([]ResolvedPermission, 0, len(permMap))
	for _, p := range permMap {
		result = append(result, p)
	}

	if r.permCache != nil {
		permStrings := make([]string, len(result))
		for i, p := range result {
			permStrings[i] = p.Name
		}
		_ = r.permCache.Set(ctx, userID, tenantID, permStrings)
	}

	return result, nil
}

// HasPermission does an exact resource:action match — the deployed schema
// carries no scope column to evaluate hierarchically (resourceCtx is
// accepted for interface compatibility with the gRPC interceptor but unused
// beyond the permission-string match itself).
func (r *permissionResolver) HasPermission(ctx context.Context, userID uuid.UUID, userTenantID uuid.UUID, permission string, _ *ResourceContext) (bool, error) {
	perms, err := r.ResolvePermissions(ctx, userID, userTenantID)
	if err != nil {
		return false, err
	}

	for _, p := range perms {
		if p.Name == permission {
			return true, nil
		}
	}
	return false, nil
}

func (r *permissionResolver) GetEffectivePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]string, error) {
	if r.permCache != nil {
		cached, err := r.permCache.Get(ctx, userID, tenantID)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	perms, err := r.ResolvePermissions(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = p.Name
	}
	return result, nil
}

func (r *permissionResolver) InvalidateCache(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) error {
	if r.permCache != nil {
		return r.permCache.Invalidate(ctx, userID, tenantID)
	}
	return nil
}
