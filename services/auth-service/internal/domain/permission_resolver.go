package domain

import (
	"context"
	"github.com/google/uuid"
)

type ResolvedPermission struct {
	Permission
	GrantedByRole string `json:"granted_by_role"`
	GrantedByScope string `json:"granted_by_scope"`
}

type PermissionResolver interface {
	ResolvePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]ResolvedPermission, error)
	HasPermission(ctx context.Context, userID uuid.UUID, permission string, resourceCtx *ResourceContext) (bool, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]Role, error)
	GetEffectivePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]string, error)
}

type permissionResolver struct {
	userRoleRepo   UserRoleRepository
	roleRepo       RoleRepository
	permissionRepo PermissionRepository
	scopeEval      *ScopeEvaluator
	permHierarchy  *PermissionHierarchy
}

func NewPermissionResolver(
	userRoleRepo UserRoleRepository,
	roleRepo RoleRepository,
	permissionRepo PermissionRepository,
) PermissionResolver {
	return &permissionResolver{
		userRoleRepo:   userRoleRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		scopeEval:      NewScopeEvaluator(),
		permHierarchy:  NewPermissionHierarchy(),
	}
}

func (r *permissionResolver) ResolvePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]ResolvedPermission, error) {
	roles, err := r.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	permMap := make(map[string]ResolvedPermission)

	for _, role := range roles {
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

	return result, nil
}

func (r *permissionResolver) HasPermission(ctx context.Context, userID uuid.UUID, permission string, resourceCtx *ResourceContext) (bool, error) {
	perms, err := r.ResolvePermissions(ctx, userID, resourceCtx.TenantID)
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
				TenantID:       resourceCtx.TenantID,
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