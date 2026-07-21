package application

import (
	"context"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
)

type AssignRolesUseCase struct {
	userRepo     domain.UserRepository
	roleRepo     domain.RoleRepository
	userRoleRepo domain.UserRoleRepository
	permResolver domain.PermissionResolver
}

func NewAssignRolesUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	userRoleRepo domain.UserRoleRepository,
	permResolver domain.PermissionResolver,
) *AssignRolesUseCase {
	return &AssignRolesUseCase{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
		permResolver: permResolver,
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
		roles, err := uc.roleRepo.List(ctx)
		if err != nil {
			return nil, err
		}

		roleMap := make(map[uuid.UUID]domain.Role)
		for _, r := range roles {
			roleMap[r.ID] = r
		}

		for _, roleID := range addRoleIDs {
			if _, ok := roleMap[roleID]; !ok {
				return nil, domain.ErrRoleNotFound
			}
		}
	}

	// Add roles
	for _, roleID := range addRoleIDs {
		ur := &domain.UserRole{
			UserID:     userID,
			RoleID:     roleID,
			TenantID:   user.TenantID,
			AssignedBy: userID, // TODO: get actual assigner from context
		}

		if err := uc.userRoleRepo.Assign(ctx, ur); err != nil {
			return nil, err
		}
	}

	// Remove roles
	for _, roleID := range removeRoleIDs {
		if err := uc.userRoleRepo.Revoke(ctx, userID, roleID, user.TenantID); err != nil {
			return nil, err
		}
	}

	// Invalidate permission cache
	if uc.permResolver != nil {
		_ = uc.permResolver.InvalidateCache(ctx, userID, user.TenantID)
	}

	// Get updated user roles
	userRoles, err := uc.userRoleRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert to []*domain.UserRole
	userRolesPtr := make([]*domain.UserRole, len(userRoles))
	for i := range userRoles {
		userRolesPtr[i] = &userRoles[i]
	}

	return &AssignRolesResult{
		Success:   true,
		UserRoles: userRolesPtr,
	}, nil
}