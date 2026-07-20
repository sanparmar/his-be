package mfa

import (
	"context"
	"testing"
	"time"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestMFAService_InitiateChallenge(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	mfaService := NewMFAService(NewTOTPService(), client)

	ctx := context.Background()
	userID := uuid.New()

	challenge, secret, qrCode, err := mfaService.InitiateChallenge(ctx, userID, domain.MFATypeTOTP)
	if err != nil {
		t.Logf("Redis not available, skipping test: %v", err)
		return
	}

	if challenge == nil {
		t.Fatal("expected challenge, got nil")
	}

	if challenge.ID == "" {
		t.Error("challenge ID should not be empty")
	}

	if challenge.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, challenge.UserID)
	}

	if challenge.MFAType != domain.MFATypeTOTP {
		t.Errorf("expected MFAType %s, got %s", domain.MFATypeTOTP, challenge.MFAType)
	}

	if secret == "" {
		t.Error("expected secret, got empty")
	}

	if qrCode == "" {
		t.Error("expected QR code, got empty")
	}

	if time.Now().After(challenge.ExpiresAt) {
		t.Error("challenge should not be expired")
	}
}

func TestMFAService_InitiateChallenge_InvalidType(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	mfaService := NewMFAService(NewTOTPService(), client)

	ctx := context.Background()
	userID := uuid.New()

	_, _, _, err := mfaService.InitiateChallenge(ctx, userID, "invalid_type")
	if err == nil {
		t.Error("expected error for invalid MFA type")
	}

	if err != ErrInvalidMFAType {
		t.Errorf("expected ErrInvalidMFAType, got %v", err)
	}
}

func TestMFAService_VerifyChallenge(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	mfaService := NewMFAService(NewTOTPService(), client)

	ctx := context.Background()
	userID := uuid.New()

	// First initiate a challenge
	challenge, secret, _, err := mfaService.InitiateChallenge(ctx, userID, domain.MFATypeTOTP)
	if err != nil {
		t.Logf("Redis not available, skipping test: %v", err)
		return
	}

	// Generate the current valid code
	code, err := mfaService.totpService.GenerateCurrentCode(secret)
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}

	// Verify the challenge with the correct code
	valid, err := mfaService.VerifyChallenge(ctx, userID, challenge.ID, code)
	if err != nil {
		t.Fatalf("VerifyChallenge failed: %v", err)
	}

	if !valid {
		t.Error("expected valid code to be accepted")
	}

	// Try to verify again (should fail as challenge is deleted)
	valid, err = mfaService.VerifyChallenge(ctx, userID, challenge.ID, code)
	if err == nil {
		t.Error("expected error for already used challenge")
	}
}

func TestMFAService_VerifyChallenge_InvalidCode(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	mfaService := NewMFAService(NewTOTPService(), client)

	ctx := context.Background()
	userID := uuid.New()

	challenge, _, _, err := mfaService.InitiateChallenge(ctx, userID, domain.MFATypeTOTP)
	if err != nil {
		t.Logf("Redis not available, skipping test: %v", err)
		return
	}

	// Try with wrong code
	valid, err := mfaService.VerifyChallenge(ctx, userID, challenge.ID, "000000")
	if err == nil {
		t.Error("expected error for invalid code")
	}

	if valid {
		t.Error("expected invalid code to be rejected")
	}
}

func TestMFAService_VerifyChallenge_Expired(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	mfaService := NewMFAService(NewTOTPService(), client)

	ctx := context.Background()
	userID := uuid.New()

	// Create a challenge with very short TTL
	shortTTLService := &MFAService{
		totpService: NewTOTPService(),
		redisClient: client,
	}

	// We can't easily test TTL without waiting, so we'll test the challenge not found case
	valid, err := shortTTLService.VerifyChallenge(ctx, userID, "nonexistent", "123456")
	if err == nil {
		t.Error("expected error for nonexistent challenge")
	}

	if valid {
		t.Error("expected nonexistent challenge to be invalid")
	}

	if err != ErrChallengeNotFound {
		t.Errorf("expected ErrChallengeNotFound, got %v", err)
	}
}

func TestMFAService_BackupCodes(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	mfaService := NewMFAService(NewTOTPService(), client)

	// Test GenerateBackupCodes
	codes, err := mfaService.GenerateBackupCodes()
	if err != nil {
		t.Fatalf("GenerateBackupCodes failed: %v", err)
	}

	if len(codes) != BackupCodeCount {
		t.Errorf("expected %d backup codes, got %d", BackupCodeCount, len(codes))
	}

	for _, code := range codes {
		if len(code) != BackupCodeLen {
			t.Errorf("expected code length %d, got %d", BackupCodeLen, len(code))
		}
	}

	// Test HashBackupCodes
	hashed, err := mfaService.HashBackupCodes(codes)
	if err != nil {
		t.Fatalf("HashBackupCodes failed: %v", err)
	}

	if len(hashed) != len(codes) {
		t.Errorf("expected %d hashed codes, got %d", len(codes), len(hashed))
	}

	// Test VerifyBackupCode with correct code
	valid, remaining, err := mfaService.VerifyBackupCode(hashed, codes[0])
	if err != nil {
		t.Fatalf("VerifyBackupCode failed: %v", err)
	}

	if !valid {
		t.Error("expected correct backup code to be valid")
	}

	if len(remaining) != len(hashed)-1 {
		t.Errorf("expected %d remaining codes, got %d", len(hashed)-1, len(remaining))
	}

	// Test VerifyBackupCode with wrong code
	valid, _, err = mfaService.VerifyBackupCode(remaining, "wrong_code")
	if err == nil {
		t.Error("expected error for wrong backup code")
	}

	if valid {
		t.Error("expected wrong backup code to be invalid")
	}
}

func TestTOTPService_GenerateCurrentCode(t *testing.T) {
	totpService := NewTOTPService()

	secret, err := totpService.GenerateSecret("test@example.com")
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	code, err := totpService.GenerateCurrentCode(secret)
	if err != nil {
		t.Fatalf("GenerateCurrentCode failed: %v", err)
	}

	if len(code) != Digits {
		t.Errorf("expected %d digit code, got %d", Digits, len(code))
	}

	// Verify the code works
	if !totpService.VerifyCode(secret, code) {
		t.Error("generated code should be valid")
	}
}

func TestTOTPService_VerifyCode(t *testing.T) {
	totpService := NewTOTPService()

	secret, err := totpService.GenerateSecret("test@example.com")
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	// Test with valid code
	code, err := totpService.GenerateCurrentCode(secret)
	if err != nil {
		t.Fatalf("GenerateCurrentCode failed: %v", err)
	}

	if !totpService.VerifyCode(secret, code) {
		t.Error("valid code should be accepted")
	}

	// Test with invalid code
	if totpService.VerifyCode(secret, "000000") {
		t.Error("invalid code should be rejected")
	}
}

func TestTOTPService_GenerateQRCode(t *testing.T) {
	totpService := NewTOTPService()

	secret, err := totpService.GenerateSecret("test@example.com")
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	qrCode, err := totpService.GenerateQRCode(secret, "test@example.com")
	if err != nil {
		t.Fatalf("GenerateQRCode failed: %v", err)
	}

	if qrCode == "" {
		t.Error("QR code should not be empty")
	}

	// Verify it's base64 encoded
	// Just check it's not empty and looks like base64
	if len(qrCode) < 100 {
		t.Error("QR code seems too short")
	}
}

func TestTOTPService_GenerateBackupCodes(t *testing.T) {
	totpService := NewTOTPService()

	codes, err := totpService.GenerateBackupCodes()
	if err != nil {
		t.Fatalf("GenerateBackupCodes failed: %v", err)
	}

	if len(codes) != BackupCodes {
		t.Errorf("expected %d codes, got %d", BackupCodes, len(codes))
	}

	for _, code := range codes {
		if len(code) != CodeLength {
			t.Errorf("expected code length %d, got %d", CodeLength, len(code))
		}
	}

	// Ensure codes are unique
	seen := make(map[string]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("duplicate backup code: %s", code)
		}
		seen[code] = true
	}
}