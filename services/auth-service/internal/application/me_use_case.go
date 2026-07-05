package application

import (
	"context"

	"github.com/his-platform/auth-service/internal/domain"
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
