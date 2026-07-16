package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	authv1 "github.com/deloitte-us-consulting/his-be/api/auth/v1"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/application"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AuthServer implements the gRPC AuthService.
type AuthServer struct {
	authv1.UnimplementedAuthServiceServer

	loginUseCase    *application.LoginUseCase
	refreshUseCase  *application.RefreshUseCase
	logoutUseCase   *application.LogoutUseCase
	meUseCase       *application.MeUseCase
	userRepo        domain.UserRepository
	sessionRepo     domain.SessionRepository
	tokenService    domain.TokenService
}

// NewAuthServer creates a new gRPC auth server.
func NewAuthServer(
	loginUseCase *application.LoginUseCase,
	refreshUseCase *application.RefreshUseCase,
	logoutUseCase *application.LogoutUseCase,
	meUseCase *application.MeUseCase,
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	tokenService domain.TokenService,
) *AuthServer {
	return &AuthServer{
		loginUseCase:   loginUseCase,
		refreshUseCase: refreshUseCase,
		logoutUseCase:  logoutUseCase,
		meUseCase:      meUseCase,
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		tokenService:   tokenService,
	}
}

// 1. AuthenticateCredentials - Exchange credentials for JWT/Refresh token
func (s *AuthServer) AuthenticateCredentials(ctx context.Context, req *authv1.AuthenticateCredentialsRequest) (*authv1.AuthenticateCredentialsResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	tokenPair, err := s.loginUseCase.Execute(ctx, req.Username, req.Password)
	if err != nil {
		return nil, mapError(err)
	}

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &authv1.AuthenticateCredentialsResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int32(tokenPair.ExpiresIn),
		User:         domainUserToProto(user),
	}, nil
}

// 2. ValidateSession - Used by Gateway/Router to verify token state
func (s *AuthServer) ValidateSession(ctx context.Context, req *authv1.ValidateSessionRequest) (*authv1.ValidateSessionResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	user, err := s.tokenService.ValidateAccessToken(ctx, req.AccessToken)
	if err != nil {
		return &authv1.ValidateSessionResponse{
			Valid: false,
		}, nil
	}

	// Check session in database
	session, err := s.sessionRepo.GetByToken(ctx, req.AccessToken)
	if err != nil {
		return &authv1.ValidateSessionResponse{
			Valid: false,
		}, nil
	}

	if session == nil || session.ExpiresAt.Before(time.Now()) {
		return &authv1.ValidateSessionResponse{
			Valid: false,
		}, nil
	}

	// Get permissions from user roles
	permissions := getPermissionsFromRoles(user.Roles)

	return &authv1.ValidateSessionResponse{
		Valid:       true,
		User:        domainUserToProto(user),
		Permissions: permissions,
		ExpiresAt:   session.ExpiresAt.Unix(),
	}, nil
}

// 3. RefreshSession - Silent token rotation
func (s *AuthServer) RefreshSession(ctx context.Context, req *authv1.RefreshSessionRequest) (*authv1.RefreshSessionResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	tokenPair, err := s.refreshUseCase.Execute(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.RefreshSessionResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int32(tokenPair.ExpiresIn),
	}, nil
}

// 4. RevokeSession - Forced logout, immediate invalidation
func (s *AuthServer) RevokeSession(ctx context.Context, req *authv1.RevokeSessionRequest) (*authv1.RevokeSessionResponse, error) {
	if req.RefreshToken == "" && req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token or access_token is required")
	}

	// Use refresh token if available, otherwise access token
	token := req.RefreshToken
	if token == "" {
		token = req.AccessToken
	}

	err := s.logoutUseCase.Execute(ctx, token)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.RevokeSessionResponse{
		Success: true,
	}, nil
}

// 5. ProvisionIdentity - Creates the base UUID and credential record
func (s *AuthServer) ProvisionIdentity(ctx context.Context, req *authv1.ProvisionIdentityRequest) (*authv1.ProvisionIdentityResponse, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" || req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "username, email, password, and tenant_id are required")
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id format")
	}

	orgID := uuid.Nil
	if req.OrganizationId != "" {
		orgID, err = uuid.Parse(req.OrganizationId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
		}
	}

	hospitalID := uuid.Nil
	if req.HospitalId != "" {
		hospitalID, err = uuid.Parse(req.HospitalId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid hospital_id format")
		}
	}

	// Hash password
	hashedPassword, err := domain.HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	user := &domain.User{
		Username:       req.Username,
		Email:          req.Email,
		PasswordHash:   hashedPassword,
		TenantID:       tenantID,
		OrganizationID: orgID,
		HospitalID:     hospitalID,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	// Generate initial token pair
	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	// Create session
	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, status.Error(codes.Internal, "failed to create session")
	}

	return &authv1.ProvisionIdentityResponse{
		UserId:        user.ID.String(),
		AccessToken:   tokenPair.AccessToken,
		RefreshToken:  tokenPair.RefreshToken,
		ExpiresIn:     int32(tokenPair.ExpiresIn),
	}, nil
}

// 6. UpdateCredentials - Password changes, MFA enrolment
func (s *AuthServer) UpdateCredentials(ctx context.Context, req *authv1.UpdateCredentialsRequest) (*authv1.UpdateCredentialsResponse, error) {
	if req.UserId == "" || req.CurrentPassword == "" || req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, current_password, and new_password are required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Verify current password
	if err := domain.VerifyPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return nil, status.Error(codes.Unauthenticated, "current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := domain.HashPassword(req.NewPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash new password")
	}

	user.PasswordHash = hashedPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	// TODO: Handle MFA enrolment if req.EnableMfa is true

	return &authv1.UpdateCredentialsResponse{
		Success: true,
	}, nil
}

// 7. AssignRoles - RBAC mapping to UUIDs
func (s *AuthServer) AssignRoles(ctx context.Context, req *authv1.AssignRolesRequest) (*authv1.AssignRolesResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// TODO: Implement role assignment logic
	// This would typically involve a UserRole repository

	return &authv1.AssignRolesResponse{
		Success: true,
		UserRoles: []*authv1.UserRole{},
	}, nil
}

// 8. VerifyMFAChallenge - TOTP/WebAuthn verification
func (s *AuthServer) VerifyMFAChallenge(ctx context.Context, req *authv1.VerifyMFAChallengeRequest) (*authv1.VerifyMFAChallengeResponse, error) {
	if req.UserId == "" || req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and code are required")
	}

	// TODO: Implement MFA verification logic
	// This would check TOTP code or WebAuthn credential

	return &authv1.VerifyMFAChallengeResponse{
		Verified: false,
	}, nil
}

// 9. GetEffectivePermissions - Returns the flattened permission tree for a user
func (s *AuthServer) GetEffectivePermissions(ctx context.Context, req *authv1.GetEffectivePermissionsRequest) (*authv1.GetEffectivePermissionsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	permissions := getPermissionsFromRoles(user.Roles)
	roles := make([]*authv1.UserRole, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = &authv1.UserRole{
			RoleId:   r.ID.String(),
			Name:     r.Name,
			TenantId: r.Category, // Using category as tenant_id for now
		}
	}

	return &authv1.GetEffectivePermissionsResponse{
		Permissions: permissions,
		Roles:       roles,
	}, nil
}

// HealthCheck - Health check endpoint
func (s *AuthServer) HealthCheck(ctx context.Context, _ *emptypb.Empty) (*authv1.HealthCheckResponse, error) {
	return &authv1.HealthCheckResponse{
		Status:  "ok",
		Version: "1.0.0",
	}, nil
}

// Helper functions

func domainUserToProto(user *domain.User) *authv1.UserInfo {
	if user == nil {
		return nil
	}

	roles := make([]*authv1.UserRole, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = &authv1.UserRole{
			RoleId:   r.ID.String(),
			Name:     r.Name,
			TenantId: r.Category,
		}
	}

	return &authv1.UserInfo{
		Id:             user.ID.String(),
		Username:       user.Username,
		Email:          user.Email,
		TenantId:       user.TenantID.String(),
		OrganizationId: user.OrganizationID.String(),
		HospitalId:     user.HospitalID.String(),
		Roles:          roles,
	}
}

func getPermissionsFromRoles(roles []domain.Role) []string {
	permissions := make([]string, 0)
	for _, r := range roles {
		// In a real implementation, you would fetch permissions for each role
		// For now, we'll use role name as permission
		permissions = append(permissions, r.Name)
	}
	return permissions
}

func mapError(err error) error {
	switch err {
	case domain.ErrUserNotFound:
		return status.Error(codes.NotFound, "user not found")
	case domain.ErrInvalidCredentials:
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case domain.ErrInvalidToken:
		return status.Error(codes.Unauthenticated, "invalid token")
	case domain.ErrSessionExpired:
		return status.Error(codes.Unauthenticated, "session expired")
	case domain.ErrInvalidSession:
		return status.Error(codes.Unauthenticated, "invalid session")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
