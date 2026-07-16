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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthService struct {
	authv1.UnimplementedAuthServiceServer

	loginUseCase          *application.LoginUseCase
	refreshUseCase        *application.RefreshUseCase
	logoutUseCase         *application.LogoutUseCase
	meUseCase             *application.MeUseCase
	provisionIdentityUseCase *application.ProvisionIdentityUseCase
	userRepo        *postgres.UserRepository
	sessionRepo     *postgres.SessionRepository
	jwtService      *jwt.JWTService
	revocationStore *redis.TokenRevocationStore
}

func NewAuthService(
	loginUseCase *application.LoginUseCase,
	refreshUseCase *application.RefreshUseCase,
	logoutUseCase *application.LogoutUseCase,
	meUseCase *application.MeUseCase,
	provisionIdentityUseCase *application.ProvisionIdentityUseCase,
	userRepo *postgres.UserRepository,
	sessionRepo *postgres.SessionRepository,
	jwtService *jwt.JWTService,
	revocationStore *redis.TokenRevocationStore,
) *AuthService {
	return &AuthService{
		loginUseCase:          loginUseCase,
		refreshUseCase:        refreshUseCase,
		logoutUseCase:         logoutUseCase,
		meUseCase:             meUseCase,
		provisionIdentityUseCase: provisionIdentityUseCase,
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		jwtService:      jwtService,
		revocationStore: revocationStore,
	}
}

// 1. AuthenticateCredentials - Exchange credentials for JWT/Refresh token
func (s *AuthService) AuthenticateCredentials(ctx context.Context, req *authv1.AuthenticateCredentialsRequest) (*authv1.AuthenticateCredentialsResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
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

	permissions := s.getUserPermissions(user)

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
	if req.Username == "" || req.Email == "" || req.Password == "" || req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "username, email, password, and tenant_id are required")
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
	if req.UserId == "" || req.CurrentPassword == "" || req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, current_password, and new_password are required")
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

	// Verify current password
	if err := domain.VerifyPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return nil, status.Error(codes.Unauthenticated, "current password is incorrect")
	}

	// Hash new password
	newHash, err := domain.HashPassword(req.NewPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash new password")
	}

	user.PasswordHash = newHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, status.Error(codes.Internal, "failed to update password")
	}

	// Revoke all existing sessions for this user (force re-login)
	// TODO: Implement session revocation by user ID

	response := &authv1.UpdateCredentialsResponse{
		Success: true,
	}

	// TODO: Handle MFA enrolment if req.EnableMfa is true

	return response, nil
}

// 7. AssignRoles - RBAC mapping to UUIDs
func (s *AuthService) AssignRoles(ctx context.Context, req *authv1.AssignRolesRequest) (*authv1.AssignRolesResponse, error) {
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

	// TODO: Implement role assignment using role repository
	// For now, return success with current roles
	return &authv1.AssignRolesResponse{
		Success: true,
		UserRoles: []*authv1.UserRole{
			// Populate from user.Roles
		},
	}, nil
}

// 8. VerifyMFAChallenge - TOTP/WebAuthn verification
func (s *AuthService) VerifyMFAChallenge(ctx context.Context, req *authv1.VerifyMFAChallengeRequest) (*authv1.VerifyMFAChallengeResponse, error) {
	if req.UserId == "" || req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and code are required")
	}

	// TODO: Implement MFA verification
	return &authv1.VerifyMFAChallengeResponse{
		Verified: false,
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

	permissions := s.getUserPermissions(user)

	return &authv1.GetEffectivePermissionsResponse{
		Permissions: permissions,
		Roles:       s.domainRolesToProto(user.Roles),
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
	return &authv1.UserInfo{
		Id:             user.ID.String(),
		Username:       user.Username,
		Email:          user.Email,
		TenantId:       user.TenantID.String(),
		OrganizationId: user.OrganizationID.String(),
		HospitalId:     user.HospitalID.String(),
		Roles:          s.domainRolesToProto(user.Roles),
	}
}

func (s *AuthService) domainRolesToProto(roles []domain.Role) []*authv1.UserRole {
	protoRoles := make([]*authv1.UserRole, len(roles))
	for i, r := range roles {
		protoRoles[i] = &authv1.UserRole{
			RoleId:   r.ID.String(),
			Name:     r.Name,
			TenantId: r.Category, // Using category as tenant_id for now
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
