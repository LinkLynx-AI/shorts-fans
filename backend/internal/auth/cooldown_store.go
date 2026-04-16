package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const authCooldownKeyPrefix = "fan_auth_cooldown:"

type cooldownRedisClient interface {
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.BoolCmd
}

// RedisCooldownStore は short-lived auth cooldown state を Redis に保存します。
type RedisCooldownStore struct {
	client cooldownRedisClient
}

// NewRedisCooldownStore は Redis-backed cooldown store を構築します。
func NewRedisCooldownStore(client *goredis.Client) *RedisCooldownStore {
	return &RedisCooldownStore{client: client}
}

// TryActivate は key に対応する cooldown を開始し、既に active なら false を返します。
func (s *RedisCooldownStore) TryActivate(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if s == nil || s.client == nil {
		return false, fmt.Errorf("auth cooldown store is not initialized")
	}
	if strings.TrimSpace(key) == "" {
		return false, fmt.Errorf("cooldown key is required")
	}
	if ttl <= 0 {
		return false, fmt.Errorf("ttl must be greater than zero")
	}

	activated, err := s.client.SetNX(ctx, redisCooldownKey(key), "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("activate cooldown key=%s: %w", key, err)
	}

	return activated, nil
}

// Release は key に対応する cooldown を解除します。
func (s *RedisCooldownStore) Release(ctx context.Context, key string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("auth cooldown store is not initialized")
	}
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("cooldown key is required")
	}

	if err := s.client.Del(ctx, redisCooldownKey(key)).Err(); err != nil {
		return fmt.Errorf("release cooldown key=%s: %w", key, err)
	}

	return nil
}

func redisCooldownKey(key string) string {
	return authCooldownKeyPrefix + strings.TrimSpace(key)
}
