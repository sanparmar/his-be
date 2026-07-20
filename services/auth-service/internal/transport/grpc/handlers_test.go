package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/application"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
	authv1 "github.com/deloitte-us-consulting/his-be/services/auth-service/proto/auth/v1"
)

func TestAuthHandlers_Login(t *testing.T) {
	ctx := context.Background()

	mockLoginUC := &mockLoginUseCase{
		tokenPair: &domain.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    900,
		},
	}
	mockMeUC := &mockMeUseCase{
		user: &domain.User{
			ID:       uuid.New(),
			Username: "doctor1",
			Email:    "doctor1@hospital.com",
			TenantID: uuid.New(),
		},
	}

	handlers := NewAuthHandlers(
		mockLoginUC,
		nil, nil, mockMeUC,
		nil, nil, nil, nil,
		&mockPermissionResolver{},
	)

	req := &authv1.LoginRequest{
		Username: "doctor1",
		Password: "password123",
	}

	resp, err := handlers.Login(ctx, req)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if resp.AccessToken != "access-token" {
		t.Errorf("expected access token 'access-token', got '%s'", resp.AccessToken)
	}

	if resp.RefreshToken != "refresh-token" {
		t.Errorf("expected refresh token 'refresh-token', got '%s'", resp.RefreshToken)
	}

	if resp.User == nil {
		t.Error("expected user in response, got nil")
	}
}

func TestAuthHandlers_Login_InvalidCredentials(t *testing.T) {
	ctx := context.Background()

	mockLoginUC := &mockLoginUseCase{
		err: domain.ErrInvalidCredentials,
	}

	handlers := NewAuthHandlers(
		mockLoginUC,
		nil, nil, nil,
		nil, nil, nil, nil,
		&mockPermissionResolver{},
	)

	req := &authv1.LoginRequest{
		Username: "doctor1",
		Password: "wrongpassword",
	}

	_, err := handlers.Login(ctx, req)
	if err == nil {
		t.Error("expected error for invalid credentials, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected code %v, got %v", codes.Unauthenticated, st.Code())
	}
}

func TestAuthHandlers_Refresh(t *testing.T) {
	ctx := context.Background()

	mockRefreshUC := &mockRefreshUseCase{
		tokenPair: &domain.TokenPair{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    900,
		},
	}

	handlers := NewAuthHandlers(
		nil,
		mockRefreshUC, nil, nil,
		nil, nil, nil, nil,
		&mockPermissionResolver{},
	)

	req := &authv1.RefreshRequest{
		RefreshToken: "refresh-token",
	}

	resp, err := handlers.Refresh(ctx, req)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if resp.AccessToken != "new-access-token" {
		t.Errorf("expected access token 'new-access-token', got '%s'", resp.AccessToken)
	}
}

func TestAuthHandlers_Logout(t *testing.T) {
	ctx := context.Background()

	mockLogoutUC := &mockLogoutUseCase{}

	handlers := NewAuthHandlers(
		nil, nil,
		mockLogoutUC, nil,
		nil, nil, nil, nil,
		&mockPermissionResolver{},
	)

	req := &authv1.LogoutRequest{
		AccessToken: "Bearer access-token",
	}

	resp, err := handlers.Logout(ctx, req)
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	if !resp.Success {
		t.Error("expected success = true")
	}
}

func TestAuthHandlers_Me(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	tenantID := uuid.New()

	mockMeUC := &mockMeUseCase{
		user: &domain.User{
			ID:       userID,
			Username: "doctor1",
			Email:    "doctor1@hospital.com",
			TenantID: tenantID,
			Roles:    []string{"doctor"},
		},
	}

	mockPermResolver := &mockPermissionResolver{
		getEffectiveFunc: func(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
			return []string{"patient:read:own"}, nil
		},
	}

	handlers := NewAuthHandlers(
		nil, nil, nil, mockMeUC,
		nil, nil, nil, nil,
		mockPermResolver,
	)

	req := &authv1.MeRequest{
		AccessToken: "Bearer access-token",
	}

	resp, err := handlers.Me(ctx, req)
	if err != nil {
		t.Fatalf("Me() error = %v", err)
	}

	if resp.User == nil {
		t.Error("expected user in response, got nil")
	}

	if resp.User.Id != userID.String() {
		t.Errorf("expected user ID %s, got %s", userID.String(), resp.User.Id)
	}
}

func TestAuthHandlers_ValidateToken(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	tenantID := uuid.New()

	mockMeUC := &mockMeUseCase{
		user: &domain.User{
			ID:       userID,
			Username: "doctor1",
			TenantID: tenantID,
			Roles:    []string{"doctor"},
		},
	}

	mockPermResolver := &mockPermissionResolver{
		hasPermFunc: func(ctx context.Context, uID, tID uuid.UUID, perm string, rc *domain.ResourceContext) (bool, error) {
			return true, nil
		},
		getEffectiveFunc: func(ctx context.Context, uID, tID uuid.UUID) ([]string, error) {
			return []string{"patient:read:own"}, nil
		},
	}

	handlers := NewAuthHandlers(
		nil, nil, nil, mockMeUC,
		nil, nil, nil, nil,
		mockPermResolver,
	)

	req := &authv1.ValidateTokenRequest{
		AccessToken: "Bearer access-token",
	}

	resp, err := handlers.ValidateToken(ctx, req)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if !resp.Valid {
		t.Error("expected valid = true")
	}

	if resp.User == nil {
		t.Error("expected user in response")
	}
}

func TestAuthHandlers_AssignRole(t *testing.T) {
	ctx := context.Background()

	mockAssignUC := &mockAssignRoleUseCase{}

	handlers := NewAuthHandlers(
		nil, nil, nil, nil,
		mockAssignUC, nil, nil, nil,
		&mockPermissionResolver{},
	)

	req := &authv1.AssignRoleRequest{
		UserId:         uuid.New().String(),
		RoleId:         uuid.New().String(),
		TenantId:       uuid.New().String(),
		OrganizationId: uuid.New().String(),
		AssignedBy:     uuid.New().String(),
	}

	resp, err := handlers.AssignRole(ctx, req)
	if err != nil {
		t.Fatalf("AssignRole() error = %v", err)
	}

	if !resp.Success {
		t.Error("expected success = true")
	}
}

func TestAuthHandlers_RevokeRole(t *testing.T) {
	ctx := context.Background()

	mockRevokeUC := &mockRevokeRoleUseCase{}

	handlers := NewAuthHandlers(
		nil, nil, nil, nil,
		nil, mockRevokeUC, nil, nil,
		&mockPermissionResolver{},
	)

	req := &authv1.RevokeRoleRequest{
		UserId:   uuid.New().String(),
		RoleId:   uuid.New().String(),
		TenantId: uuid.New().String(),
	}

	resp, err := handlers.RevokeRole(ctx, req)
	if err != nil {
		t.Fatalf("RevokeRole() error = %v", err)
	}

	if !resp.Success {
		t.Error("expected success = true")
	}
}

type mockLoginUseCase struct {
	tokenPair *domain.TokenPair
	err       error
}

func (m *mockLoginUseCase) Execute(ctx context.Context, username, password string) (*domain.TokenPair, error) {
	return m.tokenPair, m.err
}

type mockRefreshUseCase struct {
	tokenPair *domain.TokenPair
	err       error
}

func (m *mockRefreshUseCase) Execute(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	return m.tokenPair, m.err
}

type mockLogoutUseCase struct {
	err error
}

func (m *mockLogoutUseCase) Execute(ctx context.Context, token string) error {
	return m.err
}

type mockMeUseCase struct {
	user *domain.User
	err  error
}

func (m *mockMeUseCase) Execute(ctx context.Context, token string) (*domain.User, error) {
	return m.user, m.err
}

type mockAssignRoleUseCase struct {
	err error
}

func (m *mockAssignRoleUseCase) Execute(ctx context.Context, req application.AssignRoleRequest) error {
	return m.err
}

type mockRevokeRoleUseCase struct {
	err error
}

func (m *mockRevokeRoleUseCase) Execute(ctx context.Context, req application.RevokeRoleRequest) error {
	return m.err
}

type mockPermissionResolver struct {
	hasPermFunc    func(ctx context.Context, userID, tenantID uuid.UUID, perm string, rc *domain.ResourceContext) (bool, error)
	getEffectiveFunc func(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error)
}

func (m *mockPermissionResolver) ResolvePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.ResolvedPermission, error) {
	return nil, nil
}
func (m *mockPermissionResolver) HasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string, resourceCtx *domain.ResourceContext) (bool, error) {
	if m.hasPermFunc != nil {
		return m.hasPermFunc(ctx, userID, tenantID, permission, resourceCtx)
	}
	return true, nil
}
func (m *mockPermissionResolver) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.Role, error) {
	return nil, nil
}
func (m *mockPermissionResolver) GetEffectivePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	if m.getEffectiveFunc != nil {
		return m.getEffectiveFunc(ctx, userID, tenantID)
	}
	return nil, nil
}
func (m *mockPermissionResolver) InvalidateCache(userID, tenantID uuid.UUID) {}
func (m *mockPermissionResolver) GetCachedPermissions(userID, tenantID uuid.UUID) ([]domain.ResolvedPermission, bool) {
	return nil, false
}