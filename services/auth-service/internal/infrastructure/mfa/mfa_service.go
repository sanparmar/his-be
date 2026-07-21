package mfa

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	ChallengeTTL    = 5 * time.Minute
	ChallengePrefix = "mfa_challenge:"
	BackupCodeCount = 8
	BackupCodeLen   = 10
	Issuer          = "HIS"
	Digits          = 6
	Period          = 30
)

var (
	ErrChallengeNotFound    = errors.New("MFA challenge not found or expired")
	ErrChallengeExpired     = errors.New("MFA challenge expired")
	ErrInvalidMFAType       = errors.New("invalid MFA type")
	ErrMFAAlreadyEnabled    = errors.New("MFA already enabled for this user")
	ErrInvalidCode          = errors.New("invalid MFA code")
	ErrNoMFACredential      = errors.New("no MFA credential found for user")
)

type MFAService struct {
	totpService *TOTPService
	redisClient *redis.Client
}

func NewMFAService(totp *TOTPService, redis *redis.Client) *MFAService {
	return &MFAService{
		totpService: totp,
		redisClient: redis,
	}
}

type MFAChallenge struct {
	ID         string
	UserID     uuid.UUID
	MFAType    domain.MFAType
	TOTPSecret string
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

func (s *MFAService) InitiateChallenge(ctx context.Context, userID uuid.UUID, mfaType domain.MFAType) (*MFAChallenge, string, string, error) {
	if mfaType != domain.MFATypeTOTP {
		return nil, "", "", ErrInvalidMFAType
	}

	challengeID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(ChallengeTTL)

	secret, err := s.totpService.GenerateSecret("user@" + Issuer)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	challenge := &MFAChallenge{
		ID:         challengeID,
		UserID:     userID,
		MFAType:    mfaType,
		TOTPSecret: secret,
		CreatedAt:  now,
		ExpiresAt:  expiresAt,
	}

	key := ChallengePrefix + challengeID
	data := fmt.Sprintf("%s|%s|%s|%d|%d",
		challenge.UserID.String(),
		challenge.MFAType,
		challenge.TOTPSecret,
		challenge.CreatedAt.Unix(),
		challenge.ExpiresAt.Unix(),
	)

	if err := s.redisClient.Set(ctx, key, data, ChallengeTTL).Err(); err != nil {
		return nil, "", "", fmt.Errorf("failed to store MFA challenge: %w", err)
	}

	qrCode, err := s.totpService.GenerateQRCode(secret, "user@"+Issuer)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	return challenge, secret, qrCode, nil
}

func (s *MFAService) VerifyChallenge(ctx context.Context, userID uuid.UUID, challengeID, code string) (bool, error) {
	key := ChallengePrefix + challengeID
	data, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, ErrChallengeNotFound
	}
	if err != nil {
		return false, fmt.Errorf("failed to get MFA challenge: %w", err)
	}

	parts := strings.Split(data, "|")
	if len(parts) < 5 {
		return false, ErrChallengeNotFound
	}

	storedUserID := parts[0]
	if storedUserID != userID.String() {
		return false, ErrChallengeNotFound
	}

	secret := parts[2]
	expiresAtUnix, _ := strconv.ParseInt(parts[4], 10, 64)
	expiresAt := time.Unix(expiresAtUnix, 0)
	if time.Now().After(expiresAt) {
		s.redisClient.Del(ctx, key)
		return false, ErrChallengeExpired
	}

	valid := s.totpService.VerifyCode(secret, code)
	if !valid {
		return false, ErrInvalidCode
	}

	s.redisClient.Del(ctx, key)
	return true, nil
}

func (s *MFAService) HashBackupCodes(codes []string) ([]string, error) {
	hashed := make([]string, len(codes))
	for i, code := range codes {
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hashed[i] = string(hash)
	}
	return hashed, nil
}

func (s *MFAService) VerifyBackupCode(hashedCodes []string, providedCode string) (bool, []string, error) {
	for _, hash := range hashedCodes {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(providedCode))
		if err == nil {
			return true, hashedCodes, nil
		}
	}
	return false, hashedCodes, ErrInvalidCode
}

func (s *MFAService) GenerateBackupCodes() ([]string, error) {
	codes := make([]string, BackupCodeCount)
	for i := 0; i < BackupCodeCount; i++ {
		b := make([]byte, BackupCodeLen/2)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		codes[i] = base64.StdEncoding.EncodeToString(b)[:BackupCodeLen]
	}
	return codes, nil
}