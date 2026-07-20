package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	PermCachePrefix = "perms:"
	PermCacheTTL    = 15 * time.Minute
)

type PermCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewPermCache(client *redis.Client, ttl time.Duration) *PermCache {
	if ttl == 0 {
		ttl = PermCacheTTL
	}
	return &PermCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *PermCache) key(userID, tenantID uuid.UUID) string {
	return fmt.Sprintf("%s%s:%s", PermCachePrefix, userID, tenantID)
}

func (c *PermCache) Get(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	data, err := c.client.Get(ctx, c.key(userID, tenantID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var perms []string
	if err := json.Unmarshal(data, &perms); err != nil {
		return nil, err
	}
	return perms, nil
}

func (c *PermCache) Set(ctx context.Context, userID, tenantID uuid.UUID, perms []string) error {
	data, err := json.Marshal(perms)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(userID, tenantID), data, c.ttl).Err()
}

func (c *PermCache) Invalidate(ctx context.Context, userID, tenantID uuid.UUID) error {
	return c.client.Del(ctx, c.key(userID, tenantID)).Err()
}