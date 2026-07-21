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

// ValidateToken implements AuthValidator
func (v *AuthValidatorImpl) ValidateToken(ctx context.Context, token string) (string, uuid.UUID, error) {
	// Access tokens are short-lived, stateless JWTs: validity rests on
	// signature + expiry here. Sessions are keyed by refresh token (see
	// login_use_case.go), not access token, so there is no session row to
	// look up for an access token — explicit access-token revocation is
	// handled separately via the Redis revocation list (see
	// ValidateSession/RevokeSession in transport/grpc/handlers.go).
	user, err := v.jwtService.ValidateAccessToken(ctx, token)
	if err != nil {
		return "", uuid.Nil, err
	}

	return user.ID.String(), user.TenantID, nil
}
