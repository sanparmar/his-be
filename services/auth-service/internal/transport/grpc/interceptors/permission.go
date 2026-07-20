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

// rpcToPermission maps gRPC method names to a required permission that's
// unconditional for the whole method, regardless of request content.
//
// "role:assign" and "permission:read" (the original values here) don't
// exist anywhere in the deployed permission taxonomy (see
// migrations/000006_seed_demo_rbac.up.sql — resources are patients,
// appointments, billing, clinical, laboratory, radiology, admin, reports),
// so every caller, including admin, was unconditionally denied.
//
// AssignRoles is remapped to domain.AdminRolesPermission ("Manage roles and
// permissions" in the seed data), the existing permission that already
// matches its intent. GetEffectivePermissions is NOT listed here even
// though it also requires domain.AdminRolesPermission in one case: whether
// it's required depends on request content (a caller reading their own
// effective permissions needs no extra permission at all), which this
// static per-method map can't express — see handlers.go's
// GetEffectivePermissions, which calls domain.CanAccessEffectivePermissions
// directly instead.
var rpcToPermission = map[string]string{
	"/auth.v1.AuthService/AssignRoles": domain.AdminRolesPermission,
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