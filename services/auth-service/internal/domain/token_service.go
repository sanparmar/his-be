package domain

import (
	"context"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type TokenService interface {
	GenerateTokenPair(ctx context.Context, user *User) (*TokenPair, error)
	ValidateAccessToken(ctx context.Context, token string) (*User, error)
	ValidateRefreshToken(ctx context.Context, token string) (*User, error)
}
