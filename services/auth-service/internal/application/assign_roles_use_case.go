package application

import (
	"context"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
)

type AssignRolesUseCase struct {
	userRepo      domain.UserRepository
	roleRepo      domain.RoleRepository
	userRoleRepo  domain.UserRoleRepository
	permResolver  domain.PermissionResolver
}

func NewAssignRolesUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	userRoleRepo domain.UserRoleRepository,
	permResolver domain.PermissionResolver,
) *AssignRolesUseCase {
	return &AssignRolesUseCase{
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		userRoleRepo:  userRoleRepo,
		permResolver:  permResolver,
	}
}

type AssignRolesResult struct {
	Success   bool
	UserRoles []*domain.UserRole
}

func (uc *AssignRolesUseCase) Execute(
	ctx context.Context,
	userID uuid.UUID,
	addRoleIDs []uuid.UUID,
	removeRoleIDs []uuid.UUID,
) (*AssignRolesResult, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Validate add role IDs exist
	if len(addRoleIDs) > 0 {
		roles, err := uc.roleRepo.GetByIDs(ctx, addRoleIDs)
		if err != nil {
			return nil, err
		}
		if len(roles) != len(addRoleIDs) {
			return nil, domain.ErrRoleNotFound
		}
		// Check tenant matches
		for _, role := range roles {
			if role.TenantID != user.TenantID {
				return nil, domain.ErrRoleTenantMismatch
			}
		}
	}

	// Add roles
	for _, roleID := range addRoleIDs {
		if err := uc.userRoleRepo.Add(ctx, userID, roleID); err != nil {
			return nil, err
		}
	}

	// Remove roles
	for _, roleID := range removeRoleIDs {
		if err := uc.userRoleRepo.Remove(ctx, userID, roleID); err != nil {
			return nil, err
		}
	}

	// Invalidate permission cache
	if uc.permResolver != nil {
		_ = uc.permResolver.InvalidateCache(ctx, userID, user.TenantID)
	}

	// Get updated user roles
	userRoles, err := uc.userRoleRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &AssignRolesResult{
		Success:   true,
		UserRoles: userRoles,
	}, nil
}