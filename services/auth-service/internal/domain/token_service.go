package domain

import (
	"context"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenService interface {
	GenerateTokenPair(ctx context.Context, user *User) (*TokenPair, error)
	ValidateAccessToken(ctx context.Context, token string) (*User, error)
	ValidateRefreshToken(ctx context.Context, token string) (*User, error)
}
