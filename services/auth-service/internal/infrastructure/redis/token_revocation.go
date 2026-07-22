package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenRevocationStore struct {
	client *redis.Client
}

func NewTokenRevocationStore(client *redis.Client) *TokenRevocationStore {
	return &TokenRevocationStore{client: client}
}

func (s *TokenRevocationStore) AddToRevocationList(ctx context.Context, token string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil // Token already expired, no need to add
	}
	key := "revoked_token:" + token
	return s.client.Set(ctx, key, "1", ttl).Err()
}

func (s *TokenRevocationStore) IsRevoked(ctx context.Context, token string) (bool, error) {
	key := "revoked_token:" + token
	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (s *TokenRevocationStore) AddAccessTokenToRevocationList(ctx context.Context, token string, expiresAt time.Time) error {
	return s.AddToRevocationList(ctx, token, expiresAt)
}

func (s *TokenRevocationStore) AddRefreshTokenToRevocationList(ctx context.Context, token string, expiresAt time.Time) error {
	return s.AddToRevocationList(ctx, token, expiresAt)
}