package mfa

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"image/png"
	"time"

	"github.com/pquerna/otp/totp"
)

type TOTPService struct{}

func NewTOTPService() *TOTPService {
	return &TOTPService{}
}

func (s *TOTPService) GenerateSecret(accountName string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      Issuer,
		AccountName: accountName,
		Digits:      Digits,
		Period:      Period,
	})
	if err != nil {
		return "", err
	}
	return key.Secret(), nil
}

func (s *TOTPService) GenerateCurrentCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

func (s *TOTPService) GenerateQRCode(secret, accountName string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      Issuer,
		AccountName: accountName,
		Secret:      []byte(secret),
		Digits:      Digits,
		Period:      Period,
	})
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	img, err := key.Image(256, 256)
	if err != nil {
		return "", err
	}
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (s *TOTPService) VerifyCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

func (s *TOTPService) GenerateBackupCodes() ([]string, error) {
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