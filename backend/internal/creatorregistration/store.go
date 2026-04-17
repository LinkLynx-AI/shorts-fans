package creatorregistration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const evidenceUploadKeyPrefix = "creator_registration_evidence_upload:"

type redisClient interface {
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
	Get(ctx context.Context, key string) *goredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
}

type storedEvidenceUpload struct {
	CompletedEvidence *Evidence              `json:"completedEvidence,omitempty"`
	ExpiresAt         time.Time              `json:"expiresAt"`
	FileName          string                 `json:"fileName"`
	FileSizeBytes     int64                  `json:"fileSizeBytes"`
	Kind              string                 `json:"kind"`
	MimeType          string                 `json:"mimeType"`
	PendingDelete     *EvidenceStorageObject `json:"pendingDelete,omitempty"`
	State             string                 `json:"state"`
	UploadKey         string                 `json:"uploadKey"`
	ViewerUserID      string                 `json:"viewerUserId"`
}

// RedisEvidenceUploadStore は evidence upload state を Redis に保存します。
type RedisEvidenceUploadStore struct {
	client redisClient
}

// NewRedisEvidenceUploadStore は Redis-backed evidence upload store を構築します。
func NewRedisEvidenceUploadStore(client *goredis.Client) *RedisEvidenceUploadStore {
	return &RedisEvidenceUploadStore{client: client}
}

func (s *RedisEvidenceUploadStore) DeleteUpload(ctx context.Context, evidenceUploadToken string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("creator registration evidence upload store is not initialized")
	}
	if strings.TrimSpace(evidenceUploadToken) == "" {
		return fmt.Errorf("evidence upload token is required")
	}

	if err := s.client.Del(ctx, redisEvidenceUploadKey(evidenceUploadToken)).Err(); err != nil {
		return fmt.Errorf("delete evidence upload token=%s: %w", evidenceUploadToken, err)
	}

	return nil
}

func (s *RedisEvidenceUploadStore) GetUpload(ctx context.Context, evidenceUploadToken string) (storedEvidenceUpload, error) {
	if s == nil || s.client == nil {
		return storedEvidenceUpload{}, fmt.Errorf("creator registration evidence upload store is not initialized")
	}
	if strings.TrimSpace(evidenceUploadToken) == "" {
		return storedEvidenceUpload{}, fmt.Errorf("evidence upload token is required")
	}

	payload, err := s.client.Get(ctx, redisEvidenceUploadKey(evidenceUploadToken)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return storedEvidenceUpload{}, ErrEvidenceUploadNotFound
		}
		return storedEvidenceUpload{}, fmt.Errorf("get evidence upload token=%s: %w", evidenceUploadToken, err)
	}

	var upload storedEvidenceUpload
	if err := json.Unmarshal(payload, &upload); err != nil {
		return storedEvidenceUpload{}, fmt.Errorf("unmarshal evidence upload token=%s: %w", evidenceUploadToken, err)
	}

	return upload, nil
}

func (s *RedisEvidenceUploadStore) SaveUpload(ctx context.Context, evidenceUploadToken string, upload storedEvidenceUpload, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("creator registration evidence upload store is not initialized")
	}
	if strings.TrimSpace(evidenceUploadToken) == "" {
		return fmt.Errorf("evidence upload token is required")
	}
	if ttl <= 0 {
		return fmt.Errorf("ttl must be greater than zero")
	}

	payload, err := json.Marshal(upload)
	if err != nil {
		return fmt.Errorf("marshal evidence upload token=%s: %w", evidenceUploadToken, err)
	}

	if err := s.client.Set(ctx, redisEvidenceUploadKey(evidenceUploadToken), payload, ttl).Err(); err != nil {
		return fmt.Errorf("set evidence upload token=%s: %w", evidenceUploadToken, err)
	}

	return nil
}

func redisEvidenceUploadKey(evidenceUploadToken string) string {
	return evidenceUploadKeyPrefix + strings.TrimSpace(evidenceUploadToken)
}
