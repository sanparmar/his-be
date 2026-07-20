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
	mockUserRepo := &mfaMockUserRepository{
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

	mockSessionRepo := &mfaMockSessionRepository{}
	mfaMockTokenService := &mfaMockTokenService{
		tokenPair: &domain.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    900,
		},
	}

	mockPermResolver := &mfaMockPermissionResolver{
		perms: []domain.ResolvedPermission{
			{Permission: domain.Permission{Name: "patient:read:own"}},
		},
	}

	mockMFAService := &mockMFAService{
		verifyFunc: func(ctx context.Context, userID uuid.UUID, challengeID, code string) (bool, error) {
			return code == "123456", nil
		},
	}

	mockUserRoleRepo := &mfaMockUserRoleRepository{}
	mockRoleRepo := &mfaMockRoleRepository{}

	useCase := NewVerifyMFAChallengeUseCase(
		mockUserRepo,
		mockMFARepo,
		mockMFAService,
		mockSessionRepo,
		mockUserRoleRepo,
		mockRoleRepo,
		mfaMockTokenService,
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

	mockUserRepo := &mfaMockUserRepository{user: nil}
	mockMFARepo := &mockMFARepository{}
	mockSessionRepo := &mfaMockSessionRepository{}
	mfaMockTokenService := &mfaMockTokenService{}
	mockPermResolver := &mfaMockPermissionResolver{}
	mockMFAService := &mockMFAService{}
	mockUserRoleRepo := &mfaMockUserRoleRepository{}
	mockRoleRepo := &mfaMockRoleRepository{}

	useCase := NewVerifyMFAChallengeUseCase(
		mockUserRepo,
		mockMFARepo,
		mockMFAService,
		mockSessionRepo,
		mockUserRoleRepo,
		mockRoleRepo,
		mfaMockTokenService,
		mockPermResolver,
	)

	_, err := useCase.Execute(ctx, uuid.New(), "challenge-123", "123456")
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

type mfaMockUserRepository struct {
	user *domain.User
}

func (m *mfaMockUserRepository) Create(ctx context.Context, user *domain.User) error { return nil }
func (m *mfaMockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) { return m.user, nil }
func (m *mfaMockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) { return m.user, nil }
func (m *mfaMockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) { return m.user, nil }
func (m *mfaMockUserRepository) Update(ctx context.Context, user *domain.User) error { return nil }

type mockMFARepository struct {
	mfaCred *domain.MFACredential
}

func (m *mockMFARepository) Create(ctx context.Context, mfa *domain.MFACredential) error { return nil }
func (m *mockMFARepository) GetByUserID(ctx context.Context, userID uuid.UUID, mfaType domain.MFAType) (*domain.MFACredential, error) { return m.mfaCred, nil }
func (m *mockMFARepository) Update(ctx context.Context, mfa *domain.MFACredential) error { return nil }
func (m *mockMFARepository) Delete(ctx context.Context, userID uuid.UUID, mfaType domain.MFAType) error { return nil }

type mfaMockSessionRepository struct{}

func (m *mfaMockSessionRepository) Create(ctx context.Context, session *domain.Session) error { return nil }
func (m *mfaMockSessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) { return nil, nil }
func (m *mfaMockSessionRepository) Delete(ctx context.Context, token string) error { return nil }
func (m *mfaMockSessionRepository) UpdateLastActivity(ctx context.Context, token string) error { return nil }
func (m *mfaMockSessionRepository) UpdateToken(ctx context.Context, oldToken, newToken string, expiresAt time.Time) error { return nil }
func (m *mfaMockSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error { return nil }

type mfaMockTokenService struct {
	tokenPair *domain.TokenPair
}

func (m *mfaMockTokenService) GenerateTokenPair(ctx context.Context, user *domain.User, roles, perms []string, permVersion int64) (*domain.TokenPair, error) {
	return m.tokenPair, nil
}
func (m *mfaMockTokenService) ValidateAccessToken(ctx context.Context, token string) (*domain.User, error) { return nil, nil }
func (m *mfaMockTokenService) ValidateRefreshToken(ctx context.Context, token string) (*domain.User, error) { return nil, nil }

type mfaMockPermissionResolver struct {
	perms []domain.ResolvedPermission
}

func (m *mfaMockPermissionResolver) ResolvePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.ResolvedPermission, error) { return m.perms, nil }
func (m *mfaMockPermissionResolver) HasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string, resourceCtx *domain.ResourceContext) (bool, error) { return true, nil }
func (m *mfaMockPermissionResolver) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.Role, error) { return nil, nil }
func (m *mfaMockPermissionResolver) GetEffectivePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) { return nil, nil }
func (m *mfaMockPermissionResolver) InvalidateCache(ctx context.Context, userID, tenantID uuid.UUID) error { return nil }

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

type mfaMockUserRoleRepository struct{}

func (m *mfaMockUserRoleRepository) Assign(ctx context.Context, ur *domain.UserRole) error { return nil }
func (m *mfaMockUserRoleRepository) Revoke(ctx context.Context, userID, roleID, tenantID uuid.UUID) error { return nil }
func (m *mfaMockUserRoleRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mfaMockUserRoleRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mfaMockUserRoleRepository) GetByRole(ctx context.Context, roleID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mfaMockUserRoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]domain.UserRole, error) { return nil, nil }
func (m *mfaMockUserRoleRepository) IsAssigned(ctx context.Context, userID, roleID, tenantID uuid.UUID) (bool, error) { return false, nil }
func (m *mfaMockUserRoleRepository) GetActiveByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) { return []domain.UserRole{}, nil }
func (m *mfaMockUserRoleRepository) CountByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (int, error) { return 0, nil }

type mfaMockRoleRepository struct{}

func (m *mfaMockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) { return nil, nil }
func (m *mfaMockRoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) { return nil, nil }
func (m *mfaMockRoleRepository) GetWithPermissions(ctx context.Context, id uuid.UUID) (*domain.Role, error) { return nil, nil }
func (m *mfaMockRoleRepository) GetInheritedRoles(ctx context.Context, roleID uuid.UUID) ([]domain.Role, error) { return nil, nil }
func (m *mfaMockRoleRepository) List(ctx context.Context) ([]domain.Role, error) { return nil, nil }
func (m *mfaMockRoleRepository) ListByCategory(ctx context.Context, category string) ([]domain.Role, error) { return nil, nil }
func (m *mfaMockRoleRepository) Create(ctx context.Context, role *domain.Role) error { return nil }
func (m *mfaMockRoleRepository) Update(ctx context.Context, role *domain.Role) error { return nil }
func (m *mfaMockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mfaMockRoleRepository) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error { return nil }
func (m *mfaMockRoleRepository) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error { return nil }
func (m *mfaMockRoleRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) { return nil, nil }