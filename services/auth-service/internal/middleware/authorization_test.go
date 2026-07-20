package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/his-platform/auth-service/internal/domain"
	"github.com/his-platform/auth-service/internal/infrastructure/jwt"
)

func TestAuthorizationMiddleware_RequirePermission(t *testing.T) {
	mockJWTService := &mockJWTService{}
	mockPermResolver := &mockPermissionResolver{
		hasPermFunc: func(ctx context.Context, userID, tenantID uuid.UUID, perm string, rc *domain.ResourceContext) (bool, error) {
			return perm == "patient:read:own", nil
		},
	}

	middleware := NewAuthorizationMiddleware(mockJWTService, mockPermResolver)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		TenantID: uuid.New(),
		Roles:    []string{"doctor"},
		Permissions: []string{"patient:read:own"},
	}

	validToken := "valid-token"
	mockJWTService.validateFunc = func(token string) (*domain.User, error) {
		if token == validToken {
			return user, nil
		}
		return nil, jwt.ErrInvalidToken
	}

	handler := middleware.RequirePermission("patient:read:own")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test with permission
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Test without permission
	handlerNoPerm := middleware.RequirePermission("patient:write:own")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rr = httptest.NewRecorder()
	handlerNoPerm.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestAuthorizationMiddleware_RequireModuleAccess(t *testing.T) {
	mockJWTService := &mockJWTService{}
	mockPermResolver := &mockPermissionResolver{
		getEffectivePermsFunc: func(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
			return []string{"patient:read:own", "patient:write:own", "order:read:own"}, nil
		},
	}

	middleware := NewAuthorizationMiddleware(mockJWTService, mockPermResolver)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		TenantID: uuid.New(),
		Roles:    []string{"doctor"},
	}

	mockJWTService.validateFunc = func(token string) (*domain.User, error) {
		return user, nil
	}

	handler := middleware.RequireModuleAccess("patient")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Test without module access
	handlerNoAccess := middleware.RequireModuleAccess("billing")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr = httptest.NewRecorder()
	handlerNoAccess.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestAuthorizationMiddleware_RequireScope(t *testing.T) {
	mockJWTService := &mockJWTService{}
	mockPermResolver := &mockPermissionResolver{}

	middleware := NewAuthorizationMiddleware(mockJWTService, mockPermResolver)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		TenantID: uuid.New(),
		Roles:    []string{"doctor"},
		Permissions: []string{
			"patient:read:department",
			"order:read:hospital",
		},
	}

	mockJWTService.validateFunc = func(token string) (*domain.User, error) {
		return user, nil
	}

	// Test with sufficient scope (department)
	handler := middleware.RequireScope(domain.ScopeDepartment)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Test with insufficient scope (hospital)
	handler = middleware.RequireScope(domain.ScopeHospital)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestAuthorizationMiddleware_RequireAnyPermission(t *testing.T) {
	mockJWTService := &mockJWTService{}
	mockPermResolver := &mockPermissionResolver{
		hasPermFunc: func(ctx context.Context, userID, tenantID uuid.UUID, perm string, rc *domain.ResourceContext) (bool, error) {
			return perm == "patient:read:own", nil
		},
	}

	middleware := NewAuthorizationMiddleware(mockJWTService, mockPermResolver)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		TenantID: uuid.New(),
		Roles:    []string{"doctor"},
	}

	mockJWTService.validateFunc = func(token string) (*domain.User, error) {
		return user, nil
	}

	handler := middleware.RequireAnyPermission("patient:write:own", "patient:read:own", "order:read:own")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Test with none matching
	handler = middleware.RequireAnyPermission("billing:read:own", "claim:read:own")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestAuthorizationMiddleware_RequireAllPermissions(t *testing.T) {
	mockJWTService := &mockJWTService{}
	mockPermResolver := &mockPermissionResolver{
		hasPermFunc: func(ctx context.Context, userID, tenantID uuid.UUID, perm string, rc *domain.ResourceContext) (bool, error) {
			return perm == "patient:read:own" || perm == "order:read:own", nil
		},
	}

	middleware := NewAuthorizationMiddleware(mockJWTService, mockPermResolver)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		TenantID: uuid.New(),
		Roles:    []string{"doctor"},
	}

	mockJWTService.validateFunc = func(token string) (*domain.User, error) {
		return user, nil
	}

	handler := middleware.RequireAllPermissions("patient:read:own", "order:read:own")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Test with one missing
	handler = middleware.RequireAllPermissions("patient:read:own", "billing:read:own")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestGetUserFromContext(t *testing.T) {
	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		TenantID: uuid.New(),
	}

	ctx := context.WithValue(context.Background(), UserContextKey, user)
	result := GetUserFromContext(ctx)

	if result == nil || result.ID != user.ID {
		t.Errorf("GetUserFromContext() = %v, want %v", result, user)
	}

	// Test with empty context
	ctx = context.Background()
	result = GetUserFromContext(ctx)
	if result != nil {
		t.Errorf("GetUserFromContext() should return nil for empty context, got %v", result)
	}
}

type mockJWTService struct {
	validateFunc func(token string) (*domain.User, error)
}

func (m *mockJWTService) GenerateTokenPair(ctx context.Context, user *domain.User, roles, perms []string, permVersion int64) (*domain.TokenPair, error) {
	return nil, nil
}

func (m *mockJWTService) ValidateAccessToken(ctx context.Context, token string) (*domain.User, error) {
	if m.validateFunc != nil {
		return m.validateFunc(token)
	}
	return nil, jwt.ErrInvalidToken
}

func (m *mockJWTService) ValidateRefreshToken(ctx context.Context, token string) (*domain.User, error) {
	return nil, nil
}

func (m *mockJWTService) ParseClaims(tokenStr string) (*jwt.CustomClaims, error) {
	return nil, nil
}

type mockPermissionResolver struct {
	hasPermFunc       func(ctx context.Context, userID, tenantID uuid.UUID, perm string, rc *domain.ResourceContext) (bool, error)
	getEffectivePermsFunc func(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error)
}

func (m *mockPermissionResolver) ResolvePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]domain.ResolvedPermission, error) {
	return nil, nil
}

func (m *mockPermissionResolver) HasPermission(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, permission string, resourceCtx *domain.ResourceContext) (bool, error) {
	if m.hasPermFunc != nil {
		return m.hasPermFunc(ctx, userID, tenantID, permission, resourceCtx)
	}
	return true, nil
}

func (m *mockPermissionResolver) GetUserRoles(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]domain.Role, error) {
	return nil, nil
}

func (m *mockPermissionResolver) GetEffectivePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]string, error) {
	if m.getEffectivePermsFunc != nil {
		return m.getEffectivePermsFunc(ctx, userID, tenantID)
	}
	return nil, nil
}

func (m *mockPermissionResolver) InvalidateCache(userID, tenantID uuid.UUID) {}
func (m *mockPermissionResolver) GetCachedPermissions(userID, tenantID uuid.UUID) ([]domain.ResolvedPermission, bool) {
	return nil, false
}