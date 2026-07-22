package redis

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestPermCache_GetSetInvalidate(t *testing.T) {
	client := NewMockRedisClient()
	cache := NewPermCache(client, 15*time.Minute)

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	// Test Get on empty cache
	perms, err := cache.Get(ctx, userID, tenantID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if perms != nil {
		t.Error("expected nil for empty cache")
	}

	// Test Set
	permsToSet := []string{"patient:read:own", "patient:write:own"}
	err = cache.Set(ctx, userID, tenantID, permsToSet)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test Get after Set
	perms, err = cache.Get(ctx, userID, tenantID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(perms) != len(permsToSet) {
		t.Errorf("expected %d permissions, got %d", len(permsToSet), len(perms))
	}
	for i, p := range perms {
		if p != permsToSet[i] {
			t.Errorf("permission %d: expected %s, got %s", i, permsToSet[i], p)
		}
	}

	// Test Invalidate
	err = cache.Invalidate(ctx, userID, tenantID)
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	// Test Get after Invalidate
	perms, err = cache.Get(ctx, userID, tenantID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if perms != nil {
		t.Error("expected nil after invalidation")
	}
}

func TestPermCache_KeyFormat(t *testing.T) {
	client := NewMockRedisClient()
	cache := NewPermCache(client, 15*time.Minute)

	userID := uuid.New()
	tenantID := uuid.New()

	key := cache.key(userID, tenantID)
	expected := "perms:" + userID.String() + ":" + tenantID.String()

	if key != expected {
		t.Errorf("key = %s, want %s", key, expected)
	}
}

func TestPermCache_DifferentUsers(t *testing.T) {
	client := NewMockRedisClient()
	cache := NewPermCache(client, 15*time.Minute)

	ctx := context.Background()
	userID1 := uuid.New()
	userID2 := uuid.New()
	tenantID := uuid.New()

	// Set perms for user 1
	err := cache.Set(ctx, userID1, tenantID, []string{"patient:read:own"})
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Set perms for user 2
	err = cache.Set(ctx, userID2, tenantID, []string{"order:read:own"})
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get perms for user 1
	perms1, err := cache.Get(ctx, userID1, tenantID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(perms1) != 1 || perms1[0] != "patient:read:own" {
		t.Errorf("unexpected perms for user1: %v", perms1)
	}

	// Get perms for user 2
	perms2, err := cache.Get(ctx, userID2, tenantID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(perms2) != 1 || perms2[0] != "order:read:own" {
		t.Errorf("unexpected perms for user2: %v", perms2)
	}

	// Invalidate user 1
	err = cache.Invalidate(ctx, userID1, tenantID)
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	// User 1 should be invalidated
	perms1, err = cache.Get(ctx, userID1, tenantID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if perms1 != nil {
		t.Error("expected nil for invalidated user")
	}

	// User 2 should still have perms
	perms2, err = cache.Get(ctx, userID2, tenantID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(perms2) != 1 || perms2[0] != "order:read:own" {
		t.Errorf("unexpected perms for user2 after invalidating user1: %v", perms2)
	}
}

type mockRedisClient struct {
	data map[string][]byte
}

func NewMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		data: make(map[string][]byte),
	}
}

func (m *mockRedisClient) Get(ctx context.Context, key string) *MockStringCmd {
	if val, ok := m.data[key]; ok {
		return &MockStringCmd{val: val, err: nil}
	}
	return &MockStringCmd{val: nil, err: redis.Nil}
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *MockStatusCmd {
	data, _ := value.([]byte)
	m.data[key] = data
	return &MockStatusCmd{err: nil}
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) *MockIntCmd {
	for _, key := range keys {
		delete(m.data, key)
	}
	return &MockIntCmd{val: int64(len(keys)), err: nil}
}

type MockStringCmd struct {
	val []byte
	err error
}

func (m *MockStringCmd) Bytes() ([]byte, error) {
	return m.val, m.err
}

type MockStatusCmd struct {
	err error
}

func (m *MockStatusCmd) Err() error {
	return m.err
}

type MockIntCmd struct {
	val int64
	err error
}

func (m *MockIntCmd) Err() error {
	return m.err
}