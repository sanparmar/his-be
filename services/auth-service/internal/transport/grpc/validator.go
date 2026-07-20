package grpc

import (
	"context"
	"time"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/jwt"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/postgres"
	"github.com/google/uuid"
)

// AuthValidator defines the interface for validating authentication tokens.
type AuthValidator interface {
	ValidateToken(ctx context.Context, token string) (userID string, tenantID uuid.UUID, err error)
}

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

// ValidateToken implements AuthValidator
func (v *AuthValidatorImpl) ValidateToken(ctx context.Context, token string) (string, uuid.UUID, error) {
	user, err := v.jwtService.ValidateAccessToken(ctx, token)
	if err != nil {
		return "", uuid.Nil, err
	}

	// Check session in database
	session, err := v.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return "", uuid.Nil, err
	}
	if session == nil || session.ExpiresAt.Before(time.Now()) {
		return "", uuid.Nil, jwt.ErrInvalidToken
	}

	return user.ID.String(), user.TenantID, nil
}
