package grpc

import (
	"context"
	"time"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/jwt"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/postgres"
)

// AuthValidatorImpl implements the AuthValidator interface using JWTService and SessionRepository
type AuthValidatorImpl struct {
	jwtService  *jwt.JWTService
	sessionRepo *postgres.SessionRepository
}

func NewAuthValidatorImpl(jwtService *jwt.JWTService, sessionRepo *postgres.SessionRepository) *AuthValidatorImpl {
	return &AuthValidatorImpl{
		jwtService:  jwtService,
		sessionRepo: sessionRepo,
	}
}

// ValidateToken implements interceptors.AuthValidator
func (v *AuthValidatorImpl) ValidateToken(ctx context.Context, token string) (string, error) {
	user, err := v.jwtService.ValidateAccessToken(ctx, token)
	if err != nil {
		return "", err
	}
	
	// Check session in database
	session, err := v.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return "", err
	}
	if session == nil || session.ExpiresAt.Before(time.Now()) {
		return "", jwt.ErrInvalidToken
	}
	
	return user.ID.String(), nil
}
