package application

import (
	"context"
	"github.com/google/uuid"
	"github.com/his-platform/auth-service/internal/domain"
)

type RevokeRoleRequest struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	RoleID   uuid.UUID `json:"role_id" validate:"required"`
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
}

type RevokeRoleUseCase struct {
	userRoleRepo domain.UserRoleRepository
	permResolver domain.PermissionResolver
}

func NewRevokeRoleUseCase(
	userRoleRepo domain.UserRoleRepository,
	permResolver domain.PermissionResolver,
) *RevokeRoleUseCase {
	return &RevokeRoleUseCase{
		userRoleRepo: userRoleRepo,
		permResolver: permResolver,
	}
}

func (uc *RevokeRoleUseCase) Execute(ctx context.Context, req RevokeRoleRequest) error {
	if err := uc.userRoleRepo.Revoke(ctx, req.UserID, req.RoleID, req.TenantID); err != nil {
		return err
	}

	uc.permResolver.InvalidateCache(req.UserID, req.TenantID)

	return nil
}