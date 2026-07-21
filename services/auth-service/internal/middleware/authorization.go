package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/jwt"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserContextKey    contextKey = "user"
	ClaimsContextKey  contextKey = "claims"
)

type AuthorizationMiddleware struct {
	jwtService   *jwt.JWTService
	permResolver domain.PermissionResolver
}

func NewAuthorizationMiddleware(jwtService *jwt.JWTService, permResolver domain.PermissionResolver) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		jwtService:   jwtService,
		permResolver: permResolver,
	}
}

func (m *AuthorizationMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		user, err := m.jwtService.ValidateAccessToken(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, ClaimsContextKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthorizationMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(*domain.User)
			if !ok || user == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			hasPerm, err := m.permResolver.HasPermission(r.Context(), user.ID, user.TenantID, permission, nil)
			if err != nil || !hasPerm {
				http.Error(w, "permission denied: "+permission, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthorizationMiddleware) RequireModuleAccess(module string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(*domain.User)
			if !ok || user == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			perms, err := m.permResolver.GetEffectivePermissions(r.Context(), user.ID, user.TenantID)
			if err != nil {
				http.Error(w, "permission check failed", http.StatusInternalServerError)
				return
			}

			hasAccess := false
			for _, p := range perms {
				if strings.HasPrefix(p, module+":") {
					hasAccess = true
					break
				}
			}

			if !hasAccess {
				http.Error(w, "access denied to module: "+module, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthorizationMiddleware) RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(*domain.User)
			if !ok || user == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			requiredPerm := "resource:action:" + scope
			hasPerm, err := m.permResolver.HasPermission(r.Context(), user.ID, user.TenantID, requiredPerm, nil)
			if err != nil || !hasPerm {
				http.Error(w, "insufficient scope: "+scope, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserFromContext(ctx context.Context) *domain.User {
	user, _ := ctx.Value(UserContextKey).(*domain.User)
	return user
}

func GetTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(ClaimsContextKey).(string)
	return token
}

func GetUserID(ctx context.Context) uuid.UUID {
	user := GetUserFromContext(ctx)
	if user == nil {
		return uuid.Nil
	}
	return user.ID
}

func GetTenantID(ctx context.Context) uuid.UUID {
	user := GetUserFromContext(ctx)
	if user == nil {
		return uuid.Nil
	}
	return user.TenantID
}

func GetUserRoles(ctx context.Context) []string {
	user := GetUserFromContext(ctx)
	if user == nil {
		return nil
	}
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Name
	}
	return roles
}

func HasPermission(ctx context.Context, perm string) bool {
	user := GetUserFromContext(ctx)
	if user == nil {
		return false
	}
	for _, p := range user.Permissions {
		if p == perm || strings.HasPrefix(p, perm+":") {
			return true
		}
	}
	return false
}

func HasRole(ctx context.Context, role string) bool {
	roles := GetUserRoles(ctx)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}