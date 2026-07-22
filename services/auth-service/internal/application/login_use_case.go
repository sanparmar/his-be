package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
)

type LoginUseCase struct {
	userRepo     domain.UserRepository
	sessionRepo  domain.SessionRepository
	tokenService domain.TokenService
	permResolver domain.PermissionResolver
	userRoleRepo domain.UserRoleRepository
	roleRepo     domain.RoleRepository
}

func NewLoginUseCase(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	tokenService domain.TokenService,
	permResolver domain.PermissionResolver,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		tokenService: tokenService,
		permResolver: permResolver,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, username, password string) (*domain.TokenPair, error) {
	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

if err := domain.VerifyPassword(user.PasswordHash, password); err != nil {
	return nil, domain.ErrInvalidCredentials
}

userRoles, err := uc.userRoleRepo.GetActiveByUserAndTenant(ctx, user.ID, user.TenantID)
if err != nil {
	return nil, err
}
	perms, err := uc.permResolver.ResolvePermissions(ctx, user.ID, user.TenantID)
	if err != nil {
		return nil, err
	}

	roles := make([]string, 0, len(userRoles))
	permStrings := make([]string, 0, len(perms))
	for _, ur := range userRoles {
		role, _ := uc.roleRepo.GetByID(ctx, ur.RoleID)
		if role != nil {
			roles = append(roles, role.Name)
		}
	}
	for _, p := range perms {
		permStrings = append(permStrings, p.Name)
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
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return tokenPair, nil
}