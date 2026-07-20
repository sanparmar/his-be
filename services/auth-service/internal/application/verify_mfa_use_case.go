package application

import (
	"context"
	"fmt"
	"time"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/mfa"
	"github.com/google/uuid"
)

type VerifyMFAChallengeUseCase struct {
	userRepo      domain.UserRepository
	mfaRepo       domain.MFARepository
	mfaService    *mfa.MFAService
	sessionRepo   domain.SessionRepository
	userRoleRepo  domain.UserRoleRepository
	roleRepo      domain.RoleRepository
	tokenSvc      domain.TokenService
	permResolver  domain.PermissionResolver
}

func NewVerifyMFAChallengeUseCase(
	userRepo domain.UserRepository,
	mfaRepo domain.MFARepository,
	mfaService *mfa.MFAService,
	sessionRepo domain.SessionRepository,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	tokenSvc domain.TokenService,
	permResolver domain.PermissionResolver,
) *VerifyMFAChallengeUseCase {
	return &VerifyMFAChallengeUseCase{
		userRepo:      userRepo,
		mfaRepo:       mfaRepo,
		mfaService:    mfaService,
		sessionRepo:   sessionRepo,
		userRoleRepo:  userRoleRepo,
		roleRepo:      roleRepo,
		tokenSvc:      tokenSvc,
		permResolver:  permResolver,
	}
}

type VerifyMFAResult struct {
	Verified    bool
	TokenPair   *domain.TokenPair
	MFAEnabled  bool
}

func (uc *VerifyMFAChallengeUseCase) Execute(ctx context.Context, userID uuid.UUID, challengeID, code string) (*VerifyMFAResult, error) {
	if uc.mfaService == nil {
		return nil, domain.ErrMFANotEnabled
	}

	valid, err := uc.mfaService.VerifyChallenge(ctx, userID, challengeID, code)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, domain.ErrMFAInvalidCode
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	mfaCred, err := uc.mfaRepo.GetByUserID(ctx, userID, domain.MFATypeTOTP)
	if err != nil {
		return nil, err
	}

	firstTimeEnable := false
	if mfaCred != nil && !mfaCred.Enabled {
		mfaCred.Enabled = true
		mfaCred.UpdatedAt = time.Now()
		if err := uc.mfaRepo.Update(ctx, mfaCred); err != nil {
			return nil, fmt.Errorf("failed to enable MFA: %w", err)
		}
		firstTimeEnable = true
	}

	if err := uc.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to revoke sessions: %w", err)
	}

	perms, err := uc.permResolver.ResolvePermissions(ctx, userID, user.TenantID)
	if err != nil {
		return nil, err
	}

	roles := make([]string, 0)
	userRoles, err := uc.userRoleRepo.GetActiveByUserAndTenant(ctx, userID, user.TenantID)
	if err == nil {
		for _, ur := range userRoles {
			role, _ := uc.roleRepo.GetByID(ctx, ur.RoleID)
			if role != nil {
				roles = append(roles, role.Name)
			}
		}
	}

	permStrings := make([]string, len(perms))
	for i, p := range perms {
		permStrings[i] = p.Name
	}

	permVersion := time.Now().Unix()

	tokenPair, err := uc.tokenSvc.GenerateTokenPair(ctx, user, roles, permStrings, permVersion)
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &VerifyMFAResult{
		Verified:   true,
		TokenPair:  tokenPair,
		MFAEnabled: firstTimeEnable,
	}, nil
}