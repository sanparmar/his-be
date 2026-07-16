package application

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/his-platform/auth-service/internal/domain"
)

func TestLoginUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()

	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:             uuid.New(),
		Username:       "doctor1",
		Email:          "doctor1@hospital.com",
		PasswordHash:   string(hash),
		TenantID:       uuid.New(),
		OrganizationID: uuidPtr(uuid.New()),
		HospitalID:     uuidPtr(uuid.New()),
		DepartmentID:   uuidPtr(uuid.New()),
		Roles:          []domain.Role{{Name: "doctor"}},
	}

	mockUserRepo := &mockUserRepository{user: user}
	mockSessionRepo := &mockSessionRepository{}
	mockTokenService := &mockTokenService{
		generateFunc: func(ctx context.Context, u *domain.User, roles, perms []string, pv int64) (*domain.TokenPair, error) {
			return &domain.TokenPair{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				ExpiresIn:    900,
			}, nil
		},
	}
	mockPermResolver := &mockPermissionResolver{
		resolveFunc: func(ctx context.Context, userID, tenantID uuid.UUID) ([]ResolvedPermission, error) {
			return []ResolvedPermission{
				{Permission: domain.Permission{Name: "patient:read:own"}},
				{Permission: domain.Permission{Name: "encounter:write:own"}},
			}, nil
		},
	}
	mockUserRoleRepo := &mockUserRoleRepository{}
	mockRoleRepo := &mockRoleRepository{}

	loginUC := NewLoginUseCase(mockUserRepo, mockSessionRepo, mockTokenService, mockPermResolver, mockUserRoleRepo, mockRoleRepo)

	tokenPair, err := loginUC.Execute(ctx, "doctor1", password)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if tokenPair == nil {
		t.Error("expected token pair, got nil")
	}

	if tokenPair.AccessToken != "access-token" {
		t.Errorf("expected access token 'access-token', got '%s'", tokenPair.AccessToken)
	}

	if tokenPair.RefreshToken != "refresh-token" {
		t.Errorf("expected refresh token 'refresh-token', got '%s'", tokenPair.RefreshToken)
	}
}

func TestLoginUseCase_Execute_InvalidPassword(t *testing.T) {
	ctx := context.Background()

	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:             uuid.New(),
		Username:       "doctor1",
		Email:          "doctor1@hospital.com",
		PasswordHash:   string(hash),
		TenantID:       uuid.New(),
		OrganizationID: uuidPtr(uuid.New()),
		HospitalID:     uuidPtr(uuid.New()),
	}

	mockUserRepo := &mockUserRepository{user: user}
	mockSessionRepo := &mockSessionRepository{}
	mockTokenService := &mockTokenService{}
	mockPermResolver := &mockPermissionResolver{}
	mockUserRoleRepo := &mockUserRoleRepository{}
	mockRoleRepo := &mockRoleRepository{}

	loginUC := NewLoginUseCase(mockUserRepo, mockSessionRepo, mockTokenService, mockPermResolver, mockUserRoleRepo, mockRoleRepo)

	_, err := loginUC.Execute(ctx, "doctor1", "wrongpassword")
	if err == nil {
		t.Error("expected error for wrong password, got nil")
	}
}

func TestLoginUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := &mockUserRepository{user: nil}
	mockSessionRepo := &mockSessionRepository{}
	mockTokenService := &mockTokenService{}
	mockPermResolver := &mockPermissionResolver{}
	mockUserRoleRepo := &mockUserRoleRepository{}
	mockRoleRepo := &mockRoleRepository{}

	loginUC := NewLoginUseCase(mockUserRepo, mockSessionRepo, mockTokenService, mockPermResolver, mockUserRoleRepo, mockRoleRepo)

	_, err := loginUC.Execute(ctx, "nonexistent", "password")
	if err == nil {
		t.Error("expected error for nonexistent user, got nil")
	}
}

func TestLoginUseCase_Execute_ResolvesPermissions(t *testing.T) {
	ctx := context.Background()

	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:             uuid.New(),
		Username:       "doctor1",
		Email:          "doctor1@hospital.com",
		PasswordHash:   string(hash),
		TenantID:       uuid.New(),
		OrganizationID: uuidPtr(uuid.New()),
		HospitalID:     uuidPtr(uuid.New()),
		DepartmentID:   uuidPtr(uuid.New()),
		Roles:          []domain.Role{{Name: "doctor"}},
	}

	mockUserRepo := &mockUserRepository{user: user}
	mockSessionRepo := &mockSessionRepository{}
	mockTokenService := &mockTokenService{
		generateFunc: func(ctx context.Context, u *domain.User, roles, perms []string, pv int64) (*domain.TokenPair, error) {
			if len(perms) == 0 {
				t.Error("expected permissions to be passed to token service")
			}
			if len(roles) == 0 {
				t.Error("expected roles to be passed to token service")
			}
			return &domain.TokenPair{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				ExpiresIn:    900,
			}, nil
		},
	}
	mockPermResolver := &mockPermissionResolver{
		resolveFunc: func(ctx context.Context, userID, tenantID uuid.UUID) ([]ResolvedPermission, error) {
			if userID != user.ID || tenantID != user.TenantID {
				t.Errorf("ResolvePermissions called with wrong user/tenant: userID=%v, tenantID=%v", userID, tenantID)
			}
			return []ResolvedPermission{
				{Permission: domain.Permission{Name: "patient:read:own"}},
			}, nil
		},
	}
	mockUserRoleRepo := &mockUserRoleRepository{
		getActiveFunc: func(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) {
			return []domain.UserRole{
				{UserID: user.ID, RoleID: uuid.MustParse("88888888-8888-8888-8888-888888888888"), TenantID: user.TenantID},
			}, nil
		},
	}
	mockRoleRepo := &mockRoleRepository{
		role: &domain.Role{
			ID:   uuid.MustParse("88888888-8888-8888-8888-888888888888"),
			Name: "doctor",
		},
	}

	loginUC := NewLoginUseCase(mockUserRepo, mockSessionRepo, mockTokenService, mockPermResolver, mockUserRoleRepo, mockRoleRepo)

	_, err := loginUC.Execute(ctx, "doctor1", password)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

type mockUserRepository struct {
	user *domain.User
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) { return m.user, nil }
func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) { return m.user, nil }
func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error { return nil }

type mockSessionRepository struct{}

func (m *mockSessionRepository) Create(ctx context.Context, session *domain.Session) error { return nil }
func (m *mockSessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) { return nil, nil }
func (m *mockSessionRepository) Delete(ctx context.Context, token string) error { return nil }
func (m *mockSessionRepository) UpdateLastActivity(ctx context.Context, token string) error { return nil }

type mockTokenService struct {
	generateFunc func(ctx context.Context, user *domain.User, roles, perms []string, permVersion int64) (*domain.TokenPair, error)
}

func (m *mockTokenService) GenerateTokenPair(ctx context.Context, user *domain.User, roles, perms []string, permVersion int64) (*domain.TokenPair, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, user, roles, perms, permVersion)
	}
	return &domain.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    900,
	}, nil
}
func (m *mockTokenService) ValidateAccessToken(ctx context.Context, token string) (*domain.User, error) { return nil, nil }
func (m *mockTokenService) ValidateRefreshToken(ctx context.Context, token string) (*domain.User, error) { return nil, nil }

type mockPermissionResolver struct {
	resolveFunc func(ctx context.Context, userID, tenantID uuid.UUID) ([]ResolvedPermission, error)
}

func (m *mockPermissionResolver) ResolvePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]ResolvedPermission, error) {
	if m.resolveFunc != nil {
		return m.resolveFunc(ctx, userID, tenantID)
	}
	return []ResolvedPermission{}, nil
}
func (m *mockPermissionResolver) HasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string, resourceCtx *domain.ResourceContext) (bool, error) { return true, nil }
func (m *mockPermissionResolver) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.Role, error) { return nil, nil }
func (m *mockPermissionResolver) GetEffectivePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) { return nil, nil }
func (m *mockPermissionResolver) InvalidateCache(userID, tenantID uuid.UUID) {}
func (m *mockPermissionResolver) GetCachedPermissions(userID, tenantID uuid.UUID) ([]ResolvedPermission, bool) { return nil, false }

type mockUserRoleRepository struct {
	getActiveFunc func(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error)
}

func (m *mockUserRoleRepository) Assign(ctx context.Context, ur *domain.UserRole) error { return nil }
func (m *mockUserRoleRepository) Revoke(ctx context.Context, userID, roleID, tenantID uuid.UUID) error { return nil }
func (m *mockUserRoleRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) GetByRole(ctx context.Context, roleID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) IsAssigned(ctx context.Context, userID, roleID, tenantID uuid.UUID) (bool, error) { return false, nil }
func (m *mockUserRoleRepository) GetActiveByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) {
	if m.getActiveFunc != nil {
		return m.getActiveFunc(ctx, userID, tenantID)
	}
	return []domain.UserRole{}, nil
}
func (m *mockUserRoleRepository) CountByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (int, error) { return 0, nil }

type mockRoleRepository struct {
	role *domain.Role
}

func (m *mockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) { return m.role, nil }
func (m *mockRoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) GetWithPermissions(ctx context.Context, id uuid.UUID) (*domain.Role, error) { return m.role, nil }
func (m *mockRoleRepository) GetInheritedRoles(ctx context.Context, roleID uuid.UUID) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) List(ctx context.Context) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) ListByCategory(ctx context.Context, category string) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) Create(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepository) Update(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockRoleRepository) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error { return nil }
func (m *mockRoleRepository) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error { return nil }
func (m *mockRoleRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) { return nil, nil }