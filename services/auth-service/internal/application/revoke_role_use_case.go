package application

import (
	"context"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
)

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

	if uc.permResolver != nil {
		_ = uc.permResolver.InvalidateCache(ctx, req.UserID, req.TenantID)
	}

	return nil
}