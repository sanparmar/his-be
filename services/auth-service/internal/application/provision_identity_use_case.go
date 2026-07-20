package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
)

type ProvisionIdentityUseCase struct {
	userRepo     domain.UserRepository
	sessionRepo  domain.SessionRepository
	tokenService domain.TokenService
}

func NewProvisionIdentityUseCase(userRepo domain.UserRepository, sessionRepo domain.SessionRepository, tokenService domain.TokenService) *ProvisionIdentityUseCase {
	return &ProvisionIdentityUseCase{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		tokenService: tokenService,
	}
}

func (uc *ProvisionIdentityUseCase) Execute(ctx context.Context, username, email, password string, tenantID, orgID, hospitalID uuid.UUID) (*domain.TokenPair, *domain.User, error) {
	// Check if username already exists
	existingUser, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, nil, err
	}
	if existingUser != nil {
		return nil, nil, domain.ErrUserAlreadyExists
	}

	// Check if email already exists
	existingUser, err = uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if existingUser != nil {
		return nil, nil, domain.ErrEmailAlreadyExists
	}

	// Hash password
	passwordHash, err := domain.HashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	user := &domain.User{
		ID:             uuid.New(),
		Username:       username,
		Email:          email,
		PasswordHash:   passwordHash,
		TenantID:       tenantID,
		OrganizationID: &orgID,
		HospitalID:     &hospitalID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// Generate initial token pair — a freshly provisioned user has no roles/
	// permissions assigned yet (that's a separate AssignRoles call).
	tokenPair, err := uc.tokenService.GenerateTokenPair(ctx, user, []string{}, []string{}, time.Now().Unix())
	if err != nil {
		return nil, nil, err
	}

	// Create session
	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, nil, err
	}

	return tokenPair, user, nil
}