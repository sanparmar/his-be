package interceptors

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ErrInvalidToken is returned when token validation fails
var ErrInvalidToken = errors.New("invalid or expired token")

// AuthValidator defines the interface for validating authentication tokens.
type AuthValidator interface {
	ValidateToken(ctx context.Context, token string) (userID string, tenantID uuid.UUID, err error)
}

// AuthInterceptor returns a unary server interceptor that validates authentication.
func AuthInterceptor(logger *zap.Logger, validator AuthValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// Skip auth for public methods
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("missing metadata", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "missing authentication metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			logger.Warn("missing authorization header", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := extractBearerToken(authHeaders[0])
		if token == "" {
			logger.Warn("invalid authorization header format", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		// Validate token
		userID, tenantID, err := validator.ValidateToken(ctx, token)
		if err != nil {
			logger.Warn("token validation failed", zap.String("method", info.FullMethod), zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// Add user ID and tenant ID to context
		ctx = context.WithValue(ctx, userIDKey{}, userID)
		ctx = context.WithValue(ctx, tenantIDKey{}, tenantID)

		return handler(ctx, req)
	}
}

type userIDKey struct{}
type tenantIDKey struct{}

// GetUserID extracts the user ID from context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userIDStr, ok := ctx.Value(userIDKey{}).(string)
	if !ok {
		return uuid.Nil, false
	}
	userID, err := uuid.Parse(userIDStr)
	return userID, err == nil
}

// GetTenantID extracts the tenant ID from context.
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(tenantIDKey{}).(uuid.UUID)
	return tenantID, ok
}

func isPublicMethod(method string) bool {
	publicMethods := map[string]bool{
		"/auth.v1.AuthService/AuthenticateCredentials": true,
		"/auth.v1.AuthService/RefreshSession":          true,
		"/auth.v1.AuthService/ProvisionIdentity":       true,
		"/auth.v1.AuthService/InitiateMFAChallenge":    true,
		"/auth.v1.AuthService/VerifyMFAChallenge":      true,
		"/auth.v1.AuthService/HealthCheck":             true,
		"/grpc.health.v1.Health/Check":                 true,
		"/grpc.health.v1.Health/Watch":                 true,
	}
	return publicMethods[method]
}

func extractBearerToken(authHeader string) string {
	const prefix = "Bearer "
	if strings.HasPrefix(authHeader, prefix) {
		return strings.TrimPrefix(authHeader, prefix)
	}
	return ""
}
