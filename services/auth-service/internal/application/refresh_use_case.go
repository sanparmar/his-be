package application

import (
	"context"

	"github.com/his-platform/auth-service/internal/domain"
)

type RefreshUseCase struct {
	sessionRepo  domain.SessionRepository
	tokenService domain.TokenService
}

func NewRefreshUseCase(sessionRepo domain.SessionRepository, tokenService domain.TokenService) *RefreshUseCase {
	return &RefreshUseCase{
		sessionRepo:  sessionRepo,
		tokenService: tokenService,
	}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	user, err := uc.tokenService.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	tokenPair, err := uc.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	// Update session or create new one
	// ...

	return tokenPair, nil
}
