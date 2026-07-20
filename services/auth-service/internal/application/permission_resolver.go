package application

import (
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
)

func NewPermissionResolver(
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	permissionRepo domain.PermissionRepository,
	permCache domain.PermCache,
) domain.PermissionResolver {
	return domain.NewPermissionResolver(userRoleRepo, roleRepo, permissionRepo, permCache)
}