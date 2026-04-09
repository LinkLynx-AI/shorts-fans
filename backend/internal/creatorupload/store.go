package creatorupload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const packageKeyPrefix = "creator_upload_package:"

// ErrPackageNotFound は temporary upload package が見つからないことを表します。
var ErrPackageNotFound = errors.New("creator upload package not found")

type redisClient interface {
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
	Get(ctx context.Context, key string) *goredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
}

// RedisPackageStore は upload package を Redis に保存します。
type RedisPackageStore struct {
	client redisClient
}

// NewRedisPackageStore は Redis-backed package store を構築します。
func NewRedisPackageStore(client *goredis.Client) *RedisPackageStore {
	return &RedisPackageStore{client: client}
}

// SavePackage は package token をキーに upload package を保存します。
func (s *RedisPackageStore) SavePackage(ctx context.Context, packageToken string, pkg storedPackage, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("creator upload package store is not initialized")
	}
	if strings.TrimSpace(packageToken) == "" {
		return fmt.Errorf("package token is required")
	}
	if ttl <= 0 {
		return fmt.Errorf("ttl must be greater than zero")
	}

	payload, err := json.Marshal(pkg)
	if err != nil {
		return fmt.Errorf("marshal upload package token=%s: %w", packageToken, err)
	}

	if err := s.client.Set(ctx, redisPackageKey(packageToken), payload, ttl).Err(); err != nil {
		return fmt.Errorf("set upload package token=%s: %w", packageToken, err)
	}

	return nil
}

// GetPackage は token に対応する upload package を返します。
func (s *RedisPackageStore) GetPackage(ctx context.Context, packageToken string) (storedPackage, error) {
	if s == nil || s.client == nil {
		return storedPackage{}, fmt.Errorf("creator upload package store is not initialized")
	}
	if strings.TrimSpace(packageToken) == "" {
		return storedPackage{}, fmt.Errorf("package token is required")
	}

	payload, err := s.client.Get(ctx, redisPackageKey(packageToken)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return storedPackage{}, ErrPackageNotFound
		}
		return storedPackage{}, fmt.Errorf("get upload package token=%s: %w", packageToken, err)
	}

	var pkg storedPackage
	if err := json.Unmarshal(payload, &pkg); err != nil {
		return storedPackage{}, fmt.Errorf("unmarshal upload package token=%s: %w", packageToken, err)
	}

	return pkg, nil
}

// DeletePackage は token に対応する upload package を削除します。
func (s *RedisPackageStore) DeletePackage(ctx context.Context, packageToken string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("creator upload package store is not initialized")
	}
	if strings.TrimSpace(packageToken) == "" {
		return fmt.Errorf("package token is required")
	}

	if err := s.client.Del(ctx, redisPackageKey(packageToken)).Err(); err != nil {
		return fmt.Errorf("delete upload package token=%s: %w", packageToken, err)
	}

	return nil
}

func redisPackageKey(packageToken string) string {
	return packageKeyPrefix + strings.TrimSpace(packageToken)
}
