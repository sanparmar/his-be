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
	UserID         uuid.UUID `json:"uid"`
	TenantID       uuid.UUID `json:"tid"`
	OrganizationID uuid.UUID `json:"oid"`
	HospitalID     uuid.UUID `json:"hid"`
	Username       string    `json:"sub_name"`
	Roles          []string  `json:"roles,omitempty"`
	Permissions    []string  `json:"perms,omitempty"`
	PermVersion    int64     `json:"pv,omitempty"`
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

	var orgID, hospID uuid.UUID
	if user.OrganizationID != nil {
		orgID = *user.OrganizationID
	}
	if user.HospitalID != nil {
		hospID = *user.HospitalID
	}

	accessClaims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:         user.ID,
		TenantID:       user.TenantID,
		OrganizationID: orgID,
		HospitalID:     hospID,
		Username:       user.Username,
		Roles:          roles,
		Permissions:    permissions,
		PermVersion:    permVersion,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshClaims := jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Subject:   user.ID.String(),
		ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
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

	return &domain.User{
		ID:             claims.UserID,
		Username:       claims.Username,
		TenantID:       claims.TenantID,
		OrganizationID: &claims.OrganizationID,
		HospitalID:     &claims.HospitalID,
		Roles:          rolesFromStrings(claims.Roles),
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

	// Refresh tokens are signed with only jwt.RegisteredClaims (see
	// GenerateTokenPair) — claims.UserID/TenantID/etc. are never set on
	// them and would silently decode as zero values. The one real piece
	// of identity on a refresh token is its Subject (user ID). Callers
	// that need the full user record (tenant, org, hospital, ...) must
	// look it up separately by this ID.
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return &domain.User{
		ID: userID,
	}, nil
}

func rolesFromStrings(names []string) []domain.Role {
	roles := make([]domain.Role, len(names))
	for i, n := range names {
		roles[i] = domain.Role{Name: n}
	}
	return roles
}