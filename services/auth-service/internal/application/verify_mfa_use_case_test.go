package application

import (
	"context"
	"testing"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/mfa"
	"github.com/google/uuid"
)

func TestVerifyMFAChallengeUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	// Mock dependencies
	mockUserRepo := &mockUserRepository{
		user: &domain.User{
			ID:       uuid.New(),
			Username: "testuser",
			Email:    "test@example.com",
			TenantID: uuid.New(),
			Roles:    []domain.Role{{Name: "doctor"}},
		},
	}

	mockMFARepo := &mockMFARepository{
		mfaCred: &domain.MFACredential{
			ID:        uuid.New(),
			UserID:    mockUserRepo.user.ID,
			Type:      domain.MFATypeTOTP,
			Secret:    "JBSWY3DPEHPK3PXP",
			Enabled:   false,
			BackupCodes: []string{"backup1", "backup2"},
		},
	}

	mockSessionRepo := &mockSessionRepository{}
	mockTokenService := &mockTokenService{
		tokenPair: &domain.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    900,
		},
	}

	mockPermResolver := &mockPermissionResolver{
		perms: []domain.ResolvedPermission{
			{Permission: domain.Permission{Name: "patient:read:own"}},
		},
	}

	mockMFAService := &mockMFAService{
		verifyFunc: func(ctx context.Context, userID uuid.UUID, challengeID, code string) (bool, error) {
			return code == "123456", nil
		},
	}

	mockUserRoleRepo := &mockUserRoleRepository{}
	mockRoleRepo := &mockRoleRepository{}

	useCase := NewVerifyMFAChallengeUseCase(
		mockUserRepo,
		mockMFARepo,
		mockMFAService,
		mockSessionRepo,
		mockUserRoleRepo,
		mockRoleRepo,
		mockTokenService,
		mockPermResolver,
	)

	// Test successful verification
	result, err := useCase.Execute(ctx, mockUserRepo.user.ID, "challenge-123", "123456")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Verified {
		t.Error("expected verification to succeed")
	}
	if result.TokenPair == nil {
		t.Error("expected token pair")
	}
	if !result.MFAEnabled {
		t.Error("expected MFA to be enabled on first verification")
	}

	// Test invalid code
	_, err = useCase.Execute(ctx, mockUserRepo.user.ID, "challenge-123", "wrong")
	if err == nil {
		t.Error("expected error for invalid code")
	}
}

func TestVerifyMFAChallengeUseCase_UserNotFound(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := &mockUserRepository{user: nil}
	mockMFARepo := &mockMFARepository{}
	mockSessionRepo := &mockSessionRepository{}
	mockTokenService := &mockTokenService{}
	mockPermResolver := &mockPermissionResolver{}
	mockMFAService := &mockMFAService{}
	mockUserRoleRepo := &mockUserRoleRepository{}
	mockRoleRepo := &mockRoleRepository{}

	useCase := NewVerifyMFAChallengeUseCase(
		mockUserRepo,
		mockMFARepo,
		mockMFAService,
		mockSessionRepo,
		mockUserRoleRepo,
		mockRoleRepo,
		mockTokenService,
		mockPermResolver,
	)

	_, err := useCase.Execute(ctx, uuid.New(), "challenge-123", "123456")
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

type mockUserRepository struct {
	user *domain.User
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) { return m.user, nil }
func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) { return m.user, nil }
func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) { return m.user, nil }
func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error { return nil }

type mockMFARepository struct {
	mfaCred *domain.MFACredential
}

func (m *mockMFARepository) Create(ctx context.Context, mfa *domain.MFACredential) error { return nil }
func (m *mockMFARepository) GetByUserID(ctx context.Context, userID uuid.UUID, mfaType domain.MFAType) (*domain.MFACredential, error) { return m.mfaCred, nil }
func (m *mockMFARepository) Update(ctx context.Context, mfa *domain.MFACredential) error { return nil }
func (m *mockMFARepository) Delete(ctx context.Context, userID uuid.UUID, mfaType domain.MFAType) error { return nil }

type mockSessionRepository struct{}

func (m *mockSessionRepository) Create(ctx context.Context, session *domain.Session) error { return nil }
func (m *mockSessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) { return nil, nil }
func (m *mockSessionRepository) Delete(ctx context.Context, token string) error { return nil }
func (m *mockSessionRepository) UpdateLastActivity(ctx context.Context, token string) error { return nil }
func (m *mockSessionRepository) UpdateToken(ctx context.Context, oldToken, newToken string, expiresAt time.Time) error { return nil }
func (m *mockSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error { return nil }

type mockTokenService struct {
	tokenPair *domain.TokenPair
}

func (m *mockTokenService) GenerateTokenPair(ctx context.Context, user *domain.User, roles, perms []string, permVersion int64) (*domain.TokenPair, error) {
	return m.tokenPair, nil
}
func (m *mockTokenService) ValidateAccessToken(ctx context.Context, token string) (*domain.User, error) { return nil, nil }
func (m *mockTokenService) ValidateRefreshToken(ctx context.Context, token string) (*domain.User, error) { return nil, nil }

type mockPermissionResolver struct {
	perms []domain.ResolvedPermission
}

func (m *mockPermissionResolver) ResolvePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.ResolvedPermission, error) { return m.perms, nil }
func (m *mockPermissionResolver) HasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string, resourceCtx *domain.ResourceContext) (bool, error) { return true, nil }
func (m *mockPermissionResolver) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.Role, error) { return nil, nil }
func (m *mockPermissionResolver) GetEffectivePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) { return nil, nil }
func (m *mockPermissionResolver) InvalidateCache(ctx context.Context, userID, tenantID uuid.UUID) error { return nil }

type mockMFAService struct {
	verifyFunc func(ctx context.Context, userID uuid.UUID, challengeID, code string) (bool, error)
}

func (m *mockMFAService) InitiateChallenge(ctx context.Context, userID uuid.UUID, mfaType domain.MFAType) (*mfa.MFAChallenge, string, string, error) {
	return nil, "", "", nil
}
func (m *mockMFAService) VerifyChallenge(ctx context.Context, userID uuid.UUID, challengeID, code string) (bool, error) {
	if m.verifyFunc != nil {
		return m.verifyFunc(ctx, userID, challengeID, code)
	}
	return false, nil
}
func (m *mockMFAService) HashBackupCodes(codes []string) ([]string, error) { return nil, nil }
func (m *mockMFAService) VerifyBackupCode(hashedCodes []string, providedCode string) (bool, []string, error) { return false, nil, nil }
func (m *mockMFAService) GenerateBackupCodes() ([]string, error) { return nil, nil }

type mockUserRoleRepository struct{}

func (m *mockUserRoleRepository) Assign(ctx context.Context, ur *domain.UserRole) error { return nil }
func (m *mockUserRoleRepository) Revoke(ctx context.Context, userID, roleID, tenantID uuid.UUID) error { return nil }
func (m *mockUserRoleRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) GetByRole(ctx context.Context, roleID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mockUserRoleRepository) IsAssigned(ctx context.Context, userID, roleID, tenantID uuid.UUID) (bool, error) { return false, nil }
func (m *mockUserRoleRepository) GetActiveByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) { return []domain.UserRole{}, nil }
func (m *mockUserRoleRepository) CountByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (int, error) { return 0, nil }

type mockRoleRepository struct{}

func (m *mockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) GetWithPermissions(ctx context.Context, id uuid.UUID) (*domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) GetInheritedRoles(ctx context.Context, roleID uuid.UUID) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) List(ctx context.Context) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) ListByCategory(ctx context.Context, category string) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepository) Create(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepository) Update(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockRoleRepository) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error { return nil }
func (m *mockRoleRepository) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error { return nil }
func (m *mockRoleRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) { return nil, nil }