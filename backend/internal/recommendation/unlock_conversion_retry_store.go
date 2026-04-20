package recommendation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

const (
	defaultUnlockConversionRetryTTL = 24 * time.Hour
	unlockConversionRetryKeyPrefix  = "recommendation_unlock_conversion_pending:"
)

type unlockConversionRetryRedisClient interface {
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
	Exists(ctx context.Context, keys ...string) *goredis.IntCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
}

// RedisUnlockConversionRetryStore は retry が必要な unlock_conversion を Redis に保持します。
type RedisUnlockConversionRetryStore struct {
	client unlockConversionRetryRedisClient
	ttl    time.Duration
}

// NewRedisUnlockConversionRetryStore は Redis-backed unlock conversion retry store を構築します。
func NewRedisUnlockConversionRetryStore(client *goredis.Client) *RedisUnlockConversionRetryStore {
	return &RedisUnlockConversionRetryStore{
		client: client,
		ttl:    defaultUnlockConversionRetryTTL,
	}
}

// MarkPendingUnlockConversion は pending unlock conversion retry state を記録します。
func (s *RedisUnlockConversionRetryStore) MarkPendingUnlockConversion(ctx context.Context, viewerID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("unlock conversion retry store is not initialized")
	}
	if s.ttl <= 0 {
		return fmt.Errorf("unlock conversion retry ttl must be greater than zero")
	}

	key := unlockConversionRetryKey(viewerID, mainID, shortID)
	if err := s.client.Set(ctx, key, "1", s.ttl).Err(); err != nil {
		return fmt.Errorf("mark pending unlock conversion key=%s: %w", key, err)
	}

	return nil
}

// HasPendingUnlockConversion は pending unlock conversion retry state の存在を返します。
func (s *RedisUnlockConversionRetryStore) HasPendingUnlockConversion(ctx context.Context, viewerID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID) (bool, error) {
	if s == nil || s.client == nil {
		return false, fmt.Errorf("unlock conversion retry store is not initialized")
	}

	key := unlockConversionRetryKey(viewerID, mainID, shortID)
	count, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check pending unlock conversion key=%s: %w", key, err)
	}

	return count > 0, nil
}

// ClearPendingUnlockConversion は pending unlock conversion retry state を削除します。
func (s *RedisUnlockConversionRetryStore) ClearPendingUnlockConversion(ctx context.Context, viewerID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("unlock conversion retry store is not initialized")
	}

	key := unlockConversionRetryKey(viewerID, mainID, shortID)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("clear pending unlock conversion key=%s: %w", key, err)
	}

	return nil
}

func unlockConversionRetryKey(viewerID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID) string {
	return unlockConversionRetryKeyPrefix +
		strings.TrimSpace(viewerID.String()) + ":" +
		strings.TrimSpace(mainID.String()) + ":" +
		strings.TrimSpace(shortID.String())
}
