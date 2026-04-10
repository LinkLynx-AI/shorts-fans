package creatoravatar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const uploadKeyPrefix = "creator_avatar_upload:"

type redisClient interface {
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
	Get(ctx context.Context, key string) *goredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
}

type storedUpload struct {
	AvatarAssetID string     `json:"avatarAssetId,omitempty"`
	AvatarURL     string     `json:"avatarUrl,omitempty"`
	ConsumedAt    *time.Time `json:"consumedAt,omitempty"`
	DeliveryKey   string     `json:"deliveryKey,omitempty"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	FileName      string     `json:"fileName"`
	FileSizeBytes int64      `json:"fileSizeBytes"`
	MimeType      string     `json:"mimeType"`
	State         string     `json:"state"`
	UploadKey     string     `json:"uploadKey"`
	ViewerUserID  string     `json:"viewerUserId"`
}

// RedisUploadStore は avatar upload state を Redis に保存します。
type RedisUploadStore struct {
	client redisClient
}

// NewRedisUploadStore は Redis-backed upload store を構築します。
func NewRedisUploadStore(client *goredis.Client) *RedisUploadStore {
	return &RedisUploadStore{client: client}
}

// SaveUpload は avatar upload token をキーに upload state を保存します。
func (s *RedisUploadStore) SaveUpload(ctx context.Context, avatarUploadToken string, upload storedUpload, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("creator avatar upload store is not initialized")
	}
	if strings.TrimSpace(avatarUploadToken) == "" {
		return fmt.Errorf("avatar upload token is required")
	}
	if ttl <= 0 {
		return fmt.Errorf("ttl must be greater than zero")
	}

	payload, err := json.Marshal(upload)
	if err != nil {
		return fmt.Errorf("marshal avatar upload token=%s: %w", avatarUploadToken, err)
	}

	if err := s.client.Set(ctx, redisUploadKey(avatarUploadToken), payload, ttl).Err(); err != nil {
		return fmt.Errorf("set avatar upload token=%s: %w", avatarUploadToken, err)
	}

	return nil
}

// GetUpload は avatar upload token に対応する upload state を返します。
func (s *RedisUploadStore) GetUpload(ctx context.Context, avatarUploadToken string) (storedUpload, error) {
	if s == nil || s.client == nil {
		return storedUpload{}, fmt.Errorf("creator avatar upload store is not initialized")
	}
	if strings.TrimSpace(avatarUploadToken) == "" {
		return storedUpload{}, fmt.Errorf("avatar upload token is required")
	}

	payload, err := s.client.Get(ctx, redisUploadKey(avatarUploadToken)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return storedUpload{}, ErrUploadNotFound
		}
		return storedUpload{}, fmt.Errorf("get avatar upload token=%s: %w", avatarUploadToken, err)
	}

	var upload storedUpload
	if err := json.Unmarshal(payload, &upload); err != nil {
		return storedUpload{}, fmt.Errorf("unmarshal avatar upload token=%s: %w", avatarUploadToken, err)
	}

	return upload, nil
}

// DeleteUpload は avatar upload token に対応する upload state を削除します。
func (s *RedisUploadStore) DeleteUpload(ctx context.Context, avatarUploadToken string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("creator avatar upload store is not initialized")
	}
	if strings.TrimSpace(avatarUploadToken) == "" {
		return fmt.Errorf("avatar upload token is required")
	}

	if err := s.client.Del(ctx, redisUploadKey(avatarUploadToken)).Err(); err != nil {
		return fmt.Errorf("delete avatar upload token=%s: %w", avatarUploadToken, err)
	}

	return nil
}

func redisUploadKey(avatarUploadToken string) string {
	return uploadKeyPrefix + strings.TrimSpace(avatarUploadToken)
}
