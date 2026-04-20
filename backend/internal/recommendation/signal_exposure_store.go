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
	defaultSignalExposureTTL = 30 * time.Minute
	signalExposureKeyPrefix  = "recommendation_signal_exposure:"
)

type signalExposureRedisClient interface {
	Exists(ctx context.Context, keys ...string) *goredis.IntCmd
	Pipelined(ctx context.Context, fn func(signalExposurePipeline) error) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
}

type signalExposurePipeline interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration)
}

type redisSignalExposureClient struct {
	client *goredis.Client
}

type redisSignalExposurePipeline struct {
	pipe goredis.Pipeliner
}

// RedisSignalExposureStore は recent surfaced short / creator を Redis に保存します。
type RedisSignalExposureStore struct {
	client signalExposureRedisClient
	ttl    time.Duration
}

// NewRedisSignalExposureStore は Redis-backed recommendation signal exposure store を構築します。
func NewRedisSignalExposureStore(client *goredis.Client) *RedisSignalExposureStore {
	var redisClient signalExposureRedisClient
	if client != nil {
		redisClient = redisSignalExposureClient{client: client}
	}

	return &RedisSignalExposureStore{
		client: redisClient,
		ttl:    defaultSignalExposureTTL,
	}
}

// RememberShortExposure は viewer に最近配信した short を記録します。
func (s *RedisSignalExposureStore) RememberShortExposure(ctx context.Context, viewerID uuid.UUID, shortID uuid.UUID) error {
	return s.rememberExposure(ctx, shortExposureKey(viewerID, shortID))
}

// RememberShortExposures は viewer に最近配信した short をまとめて記録します。
func (s *RedisSignalExposureStore) RememberShortExposures(ctx context.Context, viewerID uuid.UUID, shortIDs []uuid.UUID) error {
	return s.rememberExposures(ctx, exposureKeys(viewerID, shortIDs, shortExposureKey))
}

// RememberCreatorExposure は viewer に最近配信した creator を記録します。
func (s *RedisSignalExposureStore) RememberCreatorExposure(ctx context.Context, viewerID uuid.UUID, creatorUserID uuid.UUID) error {
	return s.rememberExposure(ctx, creatorExposureKey(viewerID, creatorUserID))
}

// RememberCreatorExposures は viewer に最近配信した creator をまとめて記録します。
func (s *RedisSignalExposureStore) RememberCreatorExposures(ctx context.Context, viewerID uuid.UUID, creatorUserIDs []uuid.UUID) error {
	return s.rememberExposures(ctx, exposureKeys(viewerID, creatorUserIDs, creatorExposureKey))
}

// HasShortExposure は viewer に最近配信した short が存在するかを返します。
func (s *RedisSignalExposureStore) HasShortExposure(ctx context.Context, viewerID uuid.UUID, shortID uuid.UUID) (bool, error) {
	return s.hasExposure(ctx, shortExposureKey(viewerID, shortID))
}

// HasCreatorExposure は viewer に最近配信した creator が存在するかを返します。
func (s *RedisSignalExposureStore) HasCreatorExposure(ctx context.Context, viewerID uuid.UUID, creatorUserID uuid.UUID) (bool, error) {
	return s.hasExposure(ctx, creatorExposureKey(viewerID, creatorUserID))
}

func (s *RedisSignalExposureStore) rememberExposure(ctx context.Context, key string) error {
	return s.rememberExposures(ctx, []string{key})
}

func (s *RedisSignalExposureStore) rememberExposures(ctx context.Context, keys []string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("recommendation signal exposure store is not initialized")
	}
	if s.ttl <= 0 {
		return fmt.Errorf("recommendation signal exposure ttl must be greater than zero")
	}
	if len(keys) == 0 {
		return nil
	}

	if len(keys) == 1 {
		if err := s.client.Set(ctx, keys[0], "1", s.ttl).Err(); err != nil {
			return fmt.Errorf("remember recommendation signal exposure key=%s: %w", keys[0], err)
		}

		return nil
	}

	if err := s.client.Pipelined(ctx, func(pipe signalExposurePipeline) error {
		for _, key := range keys {
			pipe.Set(ctx, key, "1", s.ttl)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("remember recommendation signal exposure keys=%v: %w", keys, err)
	}

	return nil
}

func (s *RedisSignalExposureStore) hasExposure(ctx context.Context, key string) (bool, error) {
	if s == nil || s.client == nil {
		return false, fmt.Errorf("recommendation signal exposure store is not initialized")
	}

	count, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check recommendation signal exposure key=%s: %w", key, err)
	}

	return count > 0, nil
}

func shortExposureKey(viewerID uuid.UUID, shortID uuid.UUID) string {
	return signalExposureKeyPrefix + "short:" + strings.TrimSpace(viewerID.String()) + ":" + strings.TrimSpace(shortID.String())
}

func creatorExposureKey(viewerID uuid.UUID, creatorUserID uuid.UUID) string {
	return signalExposureKeyPrefix + "creator:" + strings.TrimSpace(viewerID.String()) + ":" + strings.TrimSpace(creatorUserID.String())
}

func exposureKeys(viewerID uuid.UUID, ids []uuid.UUID, buildKey func(uuid.UUID, uuid.UUID) string) []string {
	if len(ids) == 0 {
		return nil
	}

	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, buildKey(viewerID, id))
	}

	return keys
}

func (c redisSignalExposureClient) Exists(ctx context.Context, keys ...string) *goredis.IntCmd {
	return c.client.Exists(ctx, keys...)
}

func (c redisSignalExposureClient) Pipelined(ctx context.Context, fn func(signalExposurePipeline) error) error {
	_, err := c.client.Pipelined(ctx, func(pipe goredis.Pipeliner) error {
		return fn(redisSignalExposurePipeline{pipe: pipe})
	})

	return err
}

func (c redisSignalExposureClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	return c.client.Set(ctx, key, value, expiration)
}

func (p redisSignalExposurePipeline) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) {
	p.pipe.Set(ctx, key, value, expiration)
}
