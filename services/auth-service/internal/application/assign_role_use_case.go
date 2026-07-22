package application

import (
	"context"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
)

type AssignRoleUseCase struct {
	userRoleRepo domain.UserRoleRepository
	roleRepo     domain.RoleRepository
	permResolver domain.PermissionResolver
}

func NewAssignRoleUseCase(
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	permResolver domain.PermissionResolver,
) *AssignRoleUseCase {
	return &AssignRoleUseCase{
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		permResolver: permResolver,
	}
}

func (uc *AssignRoleUseCase) Execute(ctx context.Context, req AssignRoleRequest) error {
	role, err := uc.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		return err
	}
	if role == nil {
		return domain.ErrRoleNotFound
	}

	if role.IsSystem && req.AssignedBy != uuid.Nil {
		return domain.ErrSystemRoleModification
	}

	ur := &domain.UserRole{
		UserID:         req.UserID,
		RoleID:         req.RoleID,
		TenantID:       req.TenantID,
		OrganizationID: req.OrganizationID,
		HospitalID:     req.HospitalID,
		DepartmentID:   req.DepartmentID,
		AssignedBy:     req.AssignedBy,
	}

	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		// Parse expires_at if provided
		// ur.ExpiresAt = parsedTime
	}

	if err := uc.userRoleRepo.Assign(ctx, ur); err != nil {
		return err
	}

	if uc.permResolver != nil {
		_ = uc.permResolver.InvalidateCache(ctx, req.UserID, req.TenantID)
	}

	return nil
}