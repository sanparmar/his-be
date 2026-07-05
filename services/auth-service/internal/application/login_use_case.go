package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/his-platform/auth-service/internal/domain"
)

type LoginUseCase struct {
	userRepo     domain.UserRepository
	sessionRepo  domain.SessionRepository
	tokenService domain.TokenService
}

func NewLoginUseCase(userRepo domain.UserRepository, sessionRepo domain.SessionRepository, tokenService domain.TokenService) *LoginUseCase {
	return &LoginUseCase{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		tokenService: tokenService,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, username, password string) (*domain.TokenPair, error) {
	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	_ = password

	// Password verification should be done here with bcrypt
	// For now, we'll skip it for the purpose of structure.

	tokenPair, err := uc.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return tokenPair, nil
}
