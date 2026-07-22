package application

import (
	"context"
	"time"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/mfa"
	"github.com/google/uuid"
)

type UpdateCredentialsUseCase struct {
	userRepo     domain.UserRepository
	sessionRepo  domain.SessionRepository
	mfaRepo      domain.MFARepository
	totpService  *mfa.TOTPService
	tokenService domain.TokenService
}

func NewUpdateCredentialsUseCase(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	mfaRepo domain.MFARepository,
	tokenService domain.TokenService,
) *UpdateCredentialsUseCase {
	return &UpdateCredentialsUseCase{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		mfaRepo:      mfaRepo,
		totpService:  mfa.NewTOTPService(),
		tokenService: tokenService,
	}
}

type UpdateCredentialsResult struct {
	Success     bool
	MFASecret   string
	MFAQRCode   string
	BackupCodes []string
}

func (uc *UpdateCredentialsUseCase) Execute(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string, enableMFA bool, mfaType string) (*UpdateCredentialsResult, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Verify current password
	if err := domain.VerifyPassword(user.PasswordHash, currentPassword); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	result := &UpdateCredentialsResult{Success: false}

	// Hash new password
	newHash, err := domain.HashPassword(newPassword)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = newHash
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Revoke all existing sessions for this user (force re-login)
	if err := uc.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return nil, err
	}

	// Handle MFA enrollment
	if enableMFA && mfaType == "totp" {
		// Generate TOTP secret
		secret, err := uc.totpService.GenerateSecret(user.Username)
		if err != nil {
			return nil, err
		}

		// Generate QR code
		qrCode, err := uc.totpService.GenerateQRCode(secret, user.Username)
		if err != nil {
			return nil, err
		}

		// Generate backup codes
		backupCodes, err := uc.totpService.GenerateBackupCodes()
		if err != nil {
			return nil, err
		}

		// Store MFA credential (disabled until verified)
		mfaCred := &domain.MFACredential{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        domain.MFATypeTOTP,
			Secret:      secret,
			Enabled:     false,
			BackupCodes: backupCodes,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := uc.mfaRepo.Create(ctx, mfaCred); err != nil {
			return nil, err
		}

		result.MFASecret = secret
		result.MFAQRCode = qrCode
		result.BackupCodes = backupCodes
	}

	result.Success = true
	return result, nil
}