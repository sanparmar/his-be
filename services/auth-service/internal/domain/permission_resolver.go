package domain

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
)

type ResolvedPermission struct {
	Permission
	GrantedByRole  string `json:"granted_by_role"`
	GrantedByScope string `json:"granted_by_scope"`
}

type PermissionResolver interface {
	ResolvePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]ResolvedPermission, error)
	HasPermission(ctx context.Context, userID uuid.UUID, userTenantID uuid.UUID, permission string, resourceCtx *ResourceContext) (bool, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]Role, error)
	GetEffectivePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]string, error)
	InvalidateCache(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) error
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
	scopeEval      *ScopeEvaluator
	permHierarchy  *PermissionHierarchy
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
		scopeEval:      NewScopeEvaluator(),
		permHierarchy:  NewPermissionHierarchy(),
		permCache:      permCache,
	}
}

func (r *permissionResolver) ResolvePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]ResolvedPermission, error) {
	if r.permCache != nil {
		cached, err := r.permCache.Get(ctx, userID, tenantID)
		if err == nil && cached != nil {
			perms := make([]ResolvedPermission, len(cached))
			for i, p := range cached {
				perm, _ := ParsePermission(p)
				perms[i] = ResolvedPermission{Permission: *perm}
			}
			return perms, nil
		}
	}

	roles, err := r.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	permMap := make(map[string]ResolvedPermission)

	for _, role := range roles {
		fullRole, err := r.roleRepo.GetWithPermissions(ctx, role.ID)
		if err != nil {
			continue
		}
		if fullRole != nil {
			role = *fullRole
		}

		for _, perm := range role.Permissions {
			key := perm.Name
			if existing, ok := permMap[key]; !ok || perm.ScopeLevel() > existing.Permission.ScopeLevel() {
				permMap[key] = ResolvedPermission{
					Permission:    perm,
					GrantedByRole: role.Name,
					GrantedByScope: perm.Scope,
				}
			}
		}
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

func (r *permissionResolver) HasPermission(ctx context.Context, userID uuid.UUID, userTenantID uuid.UUID, permission string, resourceCtx *ResourceContext) (bool, error) {
	perms, err := r.ResolvePermissions(ctx, userID, userTenantID)
	if err != nil {
		return false, err
	}

	requiredPerm, err := ParsePermission(permission)
	if err != nil {
		return false, err
	}

	for _, p := range perms {
		if r.permHierarchy.Implies(p.Name, requiredPerm.Name) {
			userCtx := &UserContext{
				UserID:         userID,
				TenantID:       userTenantID,
				OrganizationID: resourceCtx.OrganizationID,
				HospitalID:     resourceCtx.HospitalID,
				DepartmentID:   resourceCtx.DepartmentID,
			}
			if r.scopeEval.CanAccessResource(userCtx, resourceCtx, &p.Permission) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (r *permissionResolver) GetUserRoles(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]Role, error) {
	userRoles, err := r.userRoleRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	var roles []Role
	for _, ur := range userRoles {
		if ur.IsExpired() {
			continue
		}
		role, err := r.roleRepo.GetByID(ctx, ur.RoleID)
		if err != nil {
			continue
		}
		if role != nil {
			roles = append(roles, *role)
		}
	}

	return roles, nil
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

	var result []string
	for _, p := range perms {
		result = append(result, p.Name)
	}
	return result, nil
}

func (r *permissionResolver) InvalidateCache(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) error {
	if r.permCache != nil {
		return r.permCache.Invalidate(ctx, userID, tenantID)
	}
	return nil
}