package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	authv1 "github.com/deloitte-us-consulting/his-be/api/auth/v1"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/application"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/jwt"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/postgres"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/redis"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/validation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthService struct {
	authv1.UnimplementedAuthServiceServer

	loginUseCase             *application.LoginUseCase
	refreshUseCase           *application.RefreshUseCase
	logoutUseCase            *application.LogoutUseCase
	meUseCase                *application.MeUseCase
	provisionIdentityUseCase *application.ProvisionIdentityUseCase
	updateCredentialsUseCase *application.UpdateCredentialsUseCase
	assignRolesUseCase       *application.AssignRolesUseCase
	verifyMFAUseCase         *application.VerifyMFAChallengeUseCase
	userRepo                 *postgres.UserRepository
	sessionRepo              *postgres.SessionRepository
	roleRepo                 *postgres.RoleRepository
	userRoleRepo             *postgres.UserRoleRepository
	jwtService               *jwt.JWTService
	revocationStore          *redis.TokenRevocationStore
	permResolver             domain.PermissionResolver
}

func NewAuthService(
	loginUseCase *application.LoginUseCase,
	refreshUseCase *application.RefreshUseCase,
	logoutUseCase *application.LogoutUseCase,
	meUseCase *application.MeUseCase,
	provisionIdentityUseCase *application.ProvisionIdentityUseCase,
	updateCredentialsUseCase *application.UpdateCredentialsUseCase,
	assignRolesUseCase *application.AssignRolesUseCase,
	verifyMFAUseCase *application.VerifyMFAChallengeUseCase,
	userRepo *postgres.UserRepository,
	sessionRepo *postgres.SessionRepository,
	roleRepo *postgres.RoleRepository,
	userRoleRepo *postgres.UserRoleRepository,
	jwtService *jwt.JWTService,
	revocationStore *redis.TokenRevocationStore,
	permResolver domain.PermissionResolver,
) *AuthService {
	return &AuthService{
		loginUseCase:             loginUseCase,
		refreshUseCase:           refreshUseCase,
		logoutUseCase:            logoutUseCase,
		meUseCase:                meUseCase,
		provisionIdentityUseCase: provisionIdentityUseCase,
		updateCredentialsUseCase: updateCredentialsUseCase,
		assignRolesUseCase:       assignRolesUseCase,
		verifyMFAUseCase:         verifyMFAUseCase,
		userRepo:                 userRepo,
		sessionRepo:              sessionRepo,
		roleRepo:                 roleRepo,
		userRoleRepo:             userRoleRepo,
		jwtService:               jwtService,
		revocationStore:          revocationStore,
		permResolver:             permResolver,
	}
}

// loadUserRoles resolves a user's active roles via user_roles/roles directly
// — domain.User.Roles is never populated by UserRepository (it only reads
// the users table), so every caller that needs roles alongside user info
// must resolve them this way instead (same pattern login_use_case.go and
// refresh_use_case.go already use for JWT claims).
func (s *AuthService) loadUserRoles(ctx context.Context, userID, tenantID uuid.UUID) []domain.Role {
	if s.userRoleRepo == nil {
		return nil
	}
	userRoles, err := s.userRoleRepo.GetActiveByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil
	}
	roles := make([]domain.Role, 0, len(userRoles))
	for _, ur := range userRoles {
		role, err := s.roleRepo.GetByID(ctx, ur.RoleID)
		if err == nil && role != nil {
			roles = append(roles, *role)
		}
	}
	return roles
}

// 1. AuthenticateCredentials - Exchange credentials for JWT/Refresh token
func (s *AuthService) AuthenticateCredentials(ctx context.Context, req *authv1.AuthenticateCredentialsRequest) (*authv1.AuthenticateCredentialsResponse, error) {
	if err := validation.ValidateUsername(req.Username); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validation.ValidateOptionalUUID(req.TenantId, "tenant_id"); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tokenPair, err := s.loginUseCase.Execute(ctx, req.Username, req.Password)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound, domain.ErrInvalidCredentials:
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		default:
			return nil, status.Error(codes.Internal, "authentication failed")
		}
	}

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user info")
	}
	user.Roles = s.loadUserRoles(ctx, user.ID, user.TenantID)

	return &authv1.AuthenticateCredentialsResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int32(tokenPair.ExpiresIn),
		User:         s.domainUserToProto(user),
	}, nil
}

// 2. ValidateSession - Used by Gateway/Router to verify token state
func (s *AuthService) ValidateSession(ctx context.Context, req *authv1.ValidateSessionRequest) (*authv1.ValidateSessionResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access token is required")
	}

	// Check if token is revoked
	if s.revocationStore != nil {
		revoked, err := s.revocationStore.IsRevoked(ctx, req.AccessToken)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to check token revocation")
		}
		if revoked {
			return &authv1.ValidateSessionResponse{Valid: false}, nil
		}
	}

	tokenUser, err := s.jwtService.ValidateAccessToken(ctx, req.AccessToken)
	if err != nil {
		if err == jwt.ErrExpiredToken {
			return &authv1.ValidateSessionResponse{Valid: false}, nil
		}
		return &authv1.ValidateSessionResponse{Valid: false}, nil
	}

	// Check session in Postgres/Redis
	session, err := s.sessionRepo.GetByToken(ctx, req.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check session")
	}
	if session == nil {
		return &authv1.ValidateSessionResponse{Valid: false}, nil
	}

	// Get full user from database
	user, err := s.userRepo.GetByID(ctx, tokenUser.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}
	if user == nil {
		return &authv1.ValidateSessionResponse{Valid: false}, nil
	}
	user.Roles = s.loadUserRoles(ctx, user.ID, user.TenantID)

	permissions, err := s.permResolver.GetEffectivePermissions(ctx, user.ID, user.TenantID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get permissions")
	}

	return &authv1.ValidateSessionResponse{
		Valid:       true,
		User:        s.domainUserToProto(user),
		Permissions: permissions,
		ExpiresAt:   session.ExpiresAt.Unix(),
	}, nil
}

// 3. RefreshSession - Silent token rotation
func (s *AuthService) RefreshSession(ctx context.Context, req *authv1.RefreshSessionRequest) (*authv1.RefreshSessionResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	tokenPair, err := s.refreshUseCase.Execute(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	return &authv1.RefreshSessionResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int32(tokenPair.ExpiresIn),
	}, nil
}

// 4. RevokeSession - Forced logout, immediate invalidation
func (s *AuthService) RevokeSession(ctx context.Context, req *authv1.RevokeSessionRequest) (*authv1.RevokeSessionResponse, error) {
	if req.RefreshToken == "" && req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token or access token is required")
	}

	tokenToRevoke := req.RefreshToken
	if tokenToRevoke == "" {
		tokenToRevoke = req.AccessToken
	}

	err := s.logoutUseCase.Execute(ctx, tokenToRevoke)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke session")
	}

	// Add to revocation list in Redis
	if s.revocationStore != nil {
		// Add refresh token to revocation list with 7 day TTL (same as refresh token expiry)
		if req.RefreshToken != "" {
			expiresAt := time.Now().Add(7 * 24 * time.Hour)
			_ = s.revocationStore.AddRefreshTokenToRevocationList(ctx, req.RefreshToken, expiresAt)
		}
		// Add access token to revocation list with 15 min TTL (same as access token expiry)
		if req.AccessToken != "" {
			expiresAt := time.Now().Add(15 * time.Minute)
			_ = s.revocationStore.AddAccessTokenToRevocationList(ctx, req.AccessToken, expiresAt)
		}
	}

	return &authv1.RevokeSessionResponse{Success: true}, nil
}

// 5. ProvisionIdentity - Creates the base UUID and credential record
func (s *AuthService) ProvisionIdentity(ctx context.Context, req *authv1.ProvisionIdentityRequest) (*authv1.ProvisionIdentityResponse, error) {
	if err := validation.ValidateUsername(req.Username); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validation.ValidateEmail(req.Email); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validation.ValidatePassword(req.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validation.ValidateUUID(req.TenantId, "tenant_id"); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validation.ValidateOptionalUUID(req.OrganizationId, "organization_id"); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validation.ValidateOptionalUUID(req.HospitalId, "hospital_id"); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}

	orgID := uuid.Nil
	if req.OrganizationId != "" {
		orgID, err = uuid.Parse(req.OrganizationId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid organization_id")
		}
	}

	hospitalID := uuid.Nil
	if req.HospitalId != "" {
		hospitalID, err = uuid.Parse(req.HospitalId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid hospital_id")
		}
	}

	tokenPair, user, err := s.provisionIdentityUseCase.Execute(ctx, req.Username, req.Email, req.Password, tenantID, orgID, hospitalID)
	if err != nil {
		switch err {
		case domain.ErrUserAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, "username already exists")
		case domain.ErrEmailAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		default:
			return nil, status.Error(codes.Internal, "failed to provision identity")
		}
	}

	return &authv1.ProvisionIdentityResponse{
		UserId:        user.ID.String(),
		AccessToken:   tokenPair.AccessToken,
		RefreshToken:  tokenPair.RefreshToken,
		ExpiresIn:     int32(tokenPair.ExpiresIn),
	}, nil
}

// 6. UpdateCredentials - Password changes, MFA enrolment
func (s *AuthService) UpdateCredentials(ctx context.Context, req *authv1.UpdateCredentialsRequest) (*authv1.UpdateCredentialsResponse, error) {
	if err := validation.ValidateUUID(req.UserId, "user_id"); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if req.CurrentPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "current_password is required")
	}
	if err := validation.ValidatePassword(req.NewPassword); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validation.ValidateMFAType(req.MfaType); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	result, err := s.updateCredentialsUseCase.Execute(ctx, userID, req.CurrentPassword, req.NewPassword, req.EnableMfa, req.MfaType)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, "user not found")
		case domain.ErrInvalidCredentials:
			return nil, status.Error(codes.Unauthenticated, "current password is incorrect")
		case domain.ErrMFAAlreadyEnabled:
			return nil, status.Error(codes.AlreadyExists, "MFA already enabled")
		default:
			return nil, status.Error(codes.Internal, "failed to update credentials")
		}
	}

	return &authv1.UpdateCredentialsResponse{
		Success:      result.Success,
		MfaSecret:    result.MFASecret,
		MfaQrCode:    result.MFAQRCode,
		BackupCodes:  result.BackupCodes,
	}, nil
}

// 7. AssignRoles - RBAC mapping to UUIDs
func (s *AuthService) AssignRoles(ctx context.Context, req *authv1.AssignRolesRequest) (*authv1.AssignRolesResponse, error) {
	if err := validation.ValidateUUID(req.UserId, "user_id"); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validation.ValidateRoleIDs(req.RoleIds); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validation.ValidateRoleIDs(req.RemoveRoleIds); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	var addRoleIDs, removeRoleIDs []uuid.UUID
	for _, id := range req.RoleIds {
		if parsed, err := uuid.Parse(id); err == nil {
			addRoleIDs = append(addRoleIDs, parsed)
		}
	}
	for _, id := range req.RemoveRoleIds {
		if parsed, err := uuid.Parse(id); err == nil {
			removeRoleIDs = append(removeRoleIDs, parsed)
		}
	}

	result, err := s.assignRolesUseCase.Execute(ctx, userID, addRoleIDs, removeRoleIDs)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, "user not found")
		case domain.ErrRoleNotFound:
			return nil, status.Error(codes.NotFound, "one or more roles not found")
		case domain.ErrRoleTenantMismatch:
			return nil, status.Error(codes.InvalidArgument, "role tenant does not match user tenant")
		default:
			return nil, status.Error(codes.Internal, "failed to assign roles")
		}
	}

	var userRoles []*authv1.UserRole
	for _, ur := range result.UserRoles {
		role, _ := s.roleRepo.GetByID(ctx, ur.RoleID)
		if role != nil {
			userRoles = append(userRoles, &authv1.UserRole{
				RoleId:   role.ID.String(),
				Name:     role.Name,
				TenantId: ur.TenantID.String(),
			})
		}
	}

	return &authv1.AssignRolesResponse{
		Success:   result.Success,
		UserRoles: userRoles,
	}, nil
}

// 8. VerifyMFAChallenge - TOTP/WebAuthn verification
func (s *AuthService) VerifyMFAChallenge(ctx context.Context, req *authv1.VerifyMFAChallengeRequest) (*authv1.VerifyMFAChallengeResponse, error) {
	if req.UserId == "" || req.Code == "" || req.ChallengeId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, code, and challenge_id are required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	result, err := s.verifyMFAUseCase.Execute(ctx, userID, req.ChallengeId, req.Code)
	if err != nil {
		switch err {
		case domain.ErrMFAInvalidCode:
			return &authv1.VerifyMFAChallengeResponse{Verified: false}, nil
		default:
			return nil, status.Error(codes.Internal, "MFA verification failed")
		}
	}

	return &authv1.VerifyMFAChallengeResponse{
		Verified:     result.Verified,
		AccessToken:  result.TokenPair.AccessToken,
		RefreshToken: result.TokenPair.RefreshToken,
	}, nil
}

// 9. GetEffectivePermissions - Returns the flattened permission tree for a user
func (s *AuthService) GetEffectivePermissions(ctx context.Context, req *authv1.GetEffectivePermissionsRequest) (*authv1.GetEffectivePermissionsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	permissions, err := s.permResolver.GetEffectivePermissions(ctx, userID, user.TenantID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get effective permissions")
	}
	user.Roles = s.loadUserRoles(ctx, user.ID, user.TenantID)

	return &authv1.GetEffectivePermissionsResponse{
		Permissions: permissions,
		Roles:       s.domainRolesToProto(user.Roles, user.TenantID),
	}, nil
}

// Health check
func (s *AuthService) HealthCheck(ctx context.Context, _ *emptypb.Empty) (*authv1.HealthCheckResponse, error) {
	return &authv1.HealthCheckResponse{
		Status:  "ok",
		Version: "1.0.0",
	}, nil
}

func (s *AuthService) domainUserToProto(user *domain.User) *authv1.UserInfo {
	if user == nil {
		return nil
	}
	var orgID, hospID string
	if user.OrganizationID != nil {
		orgID = user.OrganizationID.String()
	}
	if user.HospitalID != nil {
		hospID = user.HospitalID.String()
	}
	return &authv1.UserInfo{
		Id:             user.ID.String(),
		Username:       user.Username,
		Email:          user.Email,
		TenantId:       user.TenantID.String(),
		OrganizationId: orgID,
		HospitalId:     hospID,
		Roles:          s.domainRolesToProto(user.Roles, user.TenantID),
	}
}

func (s *AuthService) domainRolesToProto(roles []domain.Role, tenantID uuid.UUID) []*authv1.UserRole {
	protoRoles := make([]*authv1.UserRole, len(roles))
	for i, r := range roles {
		protoRoles[i] = &authv1.UserRole{
			RoleId:   r.ID.String(),
			Name:     r.Name,
			TenantId: tenantID.String(),
		}
	}
	return protoRoles
}

func (s *AuthService) getUserPermissions(user *domain.User) []string {
	// TODO: Implement actual permission resolution based on roles
	// For now, return a placeholder based on roles
	var permissions []string
	for _, role := range user.Roles {
		permissions = append(permissions, fmt.Sprintf("%s:*", role.Name))
	}
	if len(permissions) == 0 {
		permissions = []string{"user:read"}
	}
	return permissions
}
