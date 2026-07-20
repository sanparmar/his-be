package grpc

import (
	"context"

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

// ValidateToken implements AuthValidator. Access tokens are validated by JWT
// signature/expiry alone (stateless) — the sessions table stores refresh
// tokens (see login_use_case.go/refresh_use_case.go/
// provision_identity_use_case.go, all of which write tokenPair.RefreshToken
// with a 7-day expiry), so looking sessions up by access token here always
// returned nil and rejected every authenticated request.
func (v *AuthValidatorImpl) ValidateToken(ctx context.Context, token string) (string, uuid.UUID, error) {
	user, err := v.jwtService.ValidateAccessToken(ctx, token)
	if err != nil {
		return "", uuid.Nil, err
	}

	return user.ID.String(), user.TenantID, nil
}
