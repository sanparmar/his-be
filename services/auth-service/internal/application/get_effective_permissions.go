package application

import (
	"context"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
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