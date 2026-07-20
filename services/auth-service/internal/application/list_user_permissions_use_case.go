package application

import (
	"context"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
)

type ListUserPermissionsUseCase struct {
	userRoleRepo domain.UserRoleRepository
	roleRepo     domain.RoleRepository
	permResolver domain.PermissionResolver
}

func NewListUserPermissionsUseCase(
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	permResolver domain.PermissionResolver,
) *ListUserPermissionsUseCase {
	return &ListUserPermissionsUseCase{
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		permResolver: permResolver,
	}
}

func (uc *ListUserPermissionsUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID) (*UserPermissionsResponse, error) {
	userRoles, err := uc.userRoleRepo.GetActiveByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	var roles []RoleResponse
	permMap := make(map[string]PermissionResponse)

	for _, ur := range userRoles {
		role, err := uc.roleRepo.GetWithPermissions(ctx, ur.RoleID)
		if err != nil || role == nil {
			continue
		}

		roles = append(roles, RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Category:    role.Category,
			IsSystem:    role.IsSystem,
		})

		for _, p := range role.Permissions {
			if _, exists := permMap[p.Name]; !exists {
				permMap[p.Name] = PermissionResponse{
					ID:          p.ID,
					Name:        p.Name,
					Resource:    p.Resource,
					Action:      p.Action,
					Scope:       p.Scope,
					Category:    p.Category,
					Description: p.Description,
				}
			}
		}
	}

	perms := make([]PermissionResponse, 0, len(permMap))
	for _, p := range permMap {
		perms = append(perms, p)
	}

	effectivePerms, _ := uc.permResolver.GetEffectivePermissions(ctx, userID, tenantID)

	return &UserPermissionsResponse{
		UserID:         userID,
		Roles:          roles,
		Permissions:    perms,
		EffectivePerms: effectivePerms,
	}, nil
}