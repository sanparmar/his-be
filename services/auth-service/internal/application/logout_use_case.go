package application

import (
	"context"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
)

type LogoutUseCase struct {
	sessionRepo domain.SessionRepository
}

func NewLogoutUseCase(sessionRepo domain.SessionRepository) *LogoutUseCase {
	return &LogoutUseCase{
		sessionRepo: sessionRepo,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, token string) error {
	return uc.sessionRepo.Delete(ctx, token)
}
