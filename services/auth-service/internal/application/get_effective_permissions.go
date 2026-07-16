package application

import (
	"context"
	"github.com/google/uuid"
	"github.com/his-platform/auth-service/internal/domain"
)

type GetEffectivePermissionsUseCase struct {
	permResolver domain.PermissionResolver
}

func NewGetEffectivePermissionsUseCase(permResolver domain.PermissionResolver) *GetEffectivePermissionsUseCase {
	return &GetEffectivePermissionsUseCase{
		permResolver: permResolver,
	}
}

func (uc *GetEffectivePermissionsUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	return uc.permResolver.GetEffectivePermissions(ctx, userID, tenantID)
}