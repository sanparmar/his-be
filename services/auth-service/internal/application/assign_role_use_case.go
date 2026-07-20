package application

import (
	"context"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
)

type AssignRoleRequest struct {
	UserID         uuid.UUID  `json:"user_id" validate:"required"`
	RoleID         uuid.UUID  `json:"role_id" validate:"required"`
	TenantID       uuid.UUID  `json:"tenant_id" validate:"required"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	HospitalID     *uuid.UUID `json:"hospital_id,omitempty"`
	DepartmentID   *uuid.UUID `json:"department_id,omitempty"`
	AssignedBy     uuid.UUID  `json:"assigned_by" validate:"required"`
	ExpiresAt      *string    `json:"expires_at,omitempty"`
}

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