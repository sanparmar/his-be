package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
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

	// Token rotation: delete old session and create new one with new refresh token
	oldSession, err := uc.sessionRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	if oldSession != nil {
		if err := uc.sessionRepo.Delete(ctx, refreshToken); err != nil {
			return nil, err
		}
	}

	newSession := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	if err := uc.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, err
	}

	return tokenPair, nil
}
