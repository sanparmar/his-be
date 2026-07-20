package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type CustomClaims struct {
	jwt.RegisteredClaims
	UserID         uuid.UUID  `json:"uid"`
	TenantID       uuid.UUID  `json:"tid"`
	OrganizationID *uuid.UUID `json:"oid,omitempty"`
	HospitalID     *uuid.UUID `json:"hid,omitempty"`
	Username       string     `json:"sub_name"`
	Roles          []string   `json:"roles,omitempty"`
	Permissions    []string   `json:"perms,omitempty"`
	PermVersion    int64      `json:"pv,omitempty"`
}

type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewJWTService(accessSecret, refreshSecret string) *JWTService {
	return &JWTService{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
	}
}

func (s *JWTService) GenerateTokenPair(ctx context.Context, user *domain.User, roles []string, permissions []string, permVersion int64) (*domain.TokenPair, error) {
	now := time.Now()

	accessClaims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:         user.ID,
		TenantID:       user.TenantID,
		OrganizationID: user.OrganizationID,
		HospitalID:     user.HospitalID,
		Username:       user.Username,
		Roles:          roles,
		Permissions:    permissions,
		PermVersion:    permVersion,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// ValidateRefreshToken parses into *CustomClaims (to reuse the same
	// parsing path as access tokens), so the refresh token must carry the
	// uid/tid/oid/hid fields too — signing it with bare RegisteredClaims
	// left those zero-valued on every refresh, which then failed at
	// sessionRepo.Create's user_id foreign key (uuid.Nil isn't a real user).
	// Deliberately omitting Roles/Permissions/PermVersion: refresh should
	// re-resolve current permissions from the DB, not replay a stale
	// snapshot from when the refresh token was minted.
	refreshClaims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:         user.ID,
		TenantID:       user.TenantID,
		OrganizationID: user.OrganizationID,
		HospitalID:     user.HospitalID,
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(15 * time.Minute.Seconds()),
	}, nil
}

func (s *JWTService) ValidateAccessToken(ctx context.Context, tokenStr string) (*domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.accessSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// domain.User.Roles is []Role (full objects), but the token only carries
	// role names ([]string, for JWT payload size) — left unset here, same as
	// the DB-load path (UserRepository never hydrates it either). Callers
	// needing roles/permissions resolve them via PermissionResolver against
	// the DB, not from token claims.
	return &domain.User{
		ID:             claims.UserID,
		Username:       claims.Username,
		TenantID:       claims.TenantID,
		OrganizationID: claims.OrganizationID,
		HospitalID:     claims.HospitalID,
		Permissions:    claims.Permissions,
		PermVersion:    claims.PermVersion,
	}, nil
}

func (s *JWTService) ValidateRefreshToken(ctx context.Context, tokenStr string) (*domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.refreshSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &domain.User{
		ID:             claims.UserID,
		TenantID:       claims.TenantID,
		OrganizationID: claims.OrganizationID,
		HospitalID:     claims.HospitalID,
	}, nil
}