package interceptors

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
)

// PermissionResolver is the interface for permission resolution
type PermissionResolver interface {
	HasPermission(ctx context.Context, userID uuid.UUID, userTenantID uuid.UUID, permission string, resourceCtx *domain.ResourceContext) (bool, error)
}

// rpcToPermission maps gRPC method names to required permissions
var rpcToPermission = map[string]string{
	"/auth.v1.AuthService/AssignRoles":             "role:assign",
	"/auth.v1.AuthService/GetEffectivePermissions": "user:read",
}

// isProtectedMethod checks if a method requires permission check
func isProtectedMethod(method string) bool {
	_, protected := rpcToPermission[method]
	return protected
}

// getRequiredPermission returns the required permission for a method
func getRequiredPermission(method string) string {
	return rpcToPermission[method]
}

// PermissionInterceptor returns a unary server interceptor that validates authorization
func PermissionInterceptor(logger *zap.Logger, permResolver PermissionResolver) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// Skip auth for public methods
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Check if method requires permission
		if !isProtectedMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Get required permission
		requiredPermission := getRequiredPermission(info.FullMethod)

		// Get user ID from context (set by auth interceptor)
		userID, ok := GetUserID(ctx)
		if !ok {
			logger.Warn("user ID not found in context", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "user ID not found in context")
		}

		// Get tenant ID from context
		tenantID, ok := GetTenantID(ctx)
		if !ok {
			logger.Warn("tenant ID not found in context", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "tenant ID not found in context")
		}

		// Create minimal resource context for scope evaluation
		resourceCtx := &domain.ResourceContext{
			TenantID: tenantID,
		}

		// Check permission
		hasPerm, err := permResolver.HasPermission(ctx, userID, tenantID, requiredPermission, resourceCtx)
		if err != nil {
			logger.Error("permission check failed",
				zap.String("method", info.FullMethod),
				zap.String("user_id", userID.String()),
				zap.String("tenant_id", tenantID.String()),
				zap.String("required_permission", requiredPermission),
				zap.Error(err))
			return nil, status.Error(codes.Internal, "permission check failed")
		}

		// Log authorization decision
		logger.Info("authorization check",
			zap.String("method", info.FullMethod),
			zap.String("user_id", userID.String()),
			zap.String("tenant_id", tenantID.String()),
			zap.String("required_permission", requiredPermission),
			zap.Bool("allowed", hasPerm))

		if !hasPerm {
			return nil, status.Error(codes.PermissionDenied, "permission denied: "+requiredPermission)
		}

		return handler(ctx, req)
	}
}