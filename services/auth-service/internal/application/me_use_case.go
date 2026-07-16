package application

import (
	"context"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
)

type MeUseCase struct {
	tokenService domain.TokenService
}

func NewMeUseCase(tokenService domain.TokenService) *MeUseCase {
	return &MeUseCase{
		tokenService: tokenService,
	}
}

func (uc *MeUseCase) Execute(ctx context.Context, token string) (*domain.User, error) {
	return uc.tokenService.ValidateAccessToken(ctx, token)
}
