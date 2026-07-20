package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
)

type RefreshUseCase struct {
	sessionRepo  domain.SessionRepository
	tokenService domain.TokenService
	permResolver domain.PermissionResolver
	userRoleRepo domain.UserRoleRepository
	roleRepo     domain.RoleRepository
}

func NewRefreshUseCase(
	sessionRepo domain.SessionRepository,
	tokenService domain.TokenService,
	permResolver domain.PermissionResolver,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
) *RefreshUseCase {
	return &RefreshUseCase{
		sessionRepo:  sessionRepo,
		tokenService: tokenService,
		permResolver: permResolver,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
	}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	user, err := uc.tokenService.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	_ = uc.sessionRepo.Delete(ctx, refreshToken)

	perms, err := uc.permResolver.ResolvePermissions(ctx, user.ID, user.TenantID)
	if err != nil {
		return nil, err
	}

	permStrings := make([]string, len(perms))
	for i, p := range perms {
		permStrings[i] = p.Name
	}

	userRoles, _ := uc.userRoleRepo.GetActiveByUserAndTenant(ctx, user.ID, user.TenantID)
	roles := make([]string, 0)
	for _, ur := range userRoles {
		role, _ := uc.roleRepo.GetByID(ctx, ur.RoleID)
		if role != nil {
			roles = append(roles, role.Name)
		}
	}

	permVersion := time.Now().Unix()

	tokenPair, err := uc.tokenService.GenerateTokenPair(ctx, user, roles, permStrings, permVersion)
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return tokenPair, nil
}