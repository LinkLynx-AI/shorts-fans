package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	signUpDraftKeyPrefix        = "fan_auth_sign_up_draft:"
	signUpHandleReservationKey  = "fan_auth_sign_up_handle:"
)

// ErrSignUpDraftNotFound は sign-up draft が見つからないことを表します。
var ErrSignUpDraftNotFound = errors.New("sign up draft was not found")

// SignUpDraft は sign-up confirm まで保持する一時 state を表します。
type SignUpDraft struct {
	DisplayName string    `json:"displayName"`
	Email       string    `json:"email"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Handle      string    `json:"handle"`
	Password    string    `json:"password"`
}

type signUpDraftRedisClient interface {
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
	Get(ctx context.Context, key string) *goredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.BoolCmd
}

// RedisSignUpDraftStore は sign-up draft と handle reservation を Redis に保存します。
type RedisSignUpDraftStore struct {
	client signUpDraftRedisClient
}

// NewRedisSignUpDraftStore は Redis-backed sign-up draft store を構築します。
func NewRedisSignUpDraftStore(client *goredis.Client) *RedisSignUpDraftStore {
	return &RedisSignUpDraftStore{client: client}
}

// SaveDraft は email に紐づく sign-up draft を保存し、handle reservation を維持します。
func (s *RedisSignUpDraftStore) SaveDraft(ctx context.Context, email string, draft SignUpDraft, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("sign up draft store is not initialized")
	}
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email is required")
	}
	if strings.TrimSpace(draft.Handle) == "" {
		return fmt.Errorf("handle is required")
	}
	if ttl <= 0 {
		return fmt.Errorf("ttl must be greater than zero")
	}

	existingDraft, err := s.GetDraft(ctx, email)
	oldHandle := ""
	switch {
	case err == nil:
		oldHandle = existingDraft.Handle
	case errors.Is(err, ErrSignUpDraftNotFound):
	default:
		return err
	}

	if err := s.claimHandle(ctx, draft.Handle, email, ttl); err != nil {
		return err
	}

	payload, err := json.Marshal(draft)
	if err != nil {
		if oldHandle != draft.Handle {
			_ = s.releaseHandle(ctx, draft.Handle)
		}
		return fmt.Errorf("marshal sign up draft email=%s: %w", email, err)
	}

	if err := s.client.Set(ctx, redisSignUpDraftKey(email), payload, ttl).Err(); err != nil {
		if oldHandle != draft.Handle {
			_ = s.releaseHandle(ctx, draft.Handle)
		}
		return fmt.Errorf("set sign up draft email=%s: %w", email, err)
	}

	if oldHandle != "" && oldHandle != draft.Handle {
		if err := s.releaseHandle(ctx, oldHandle); err != nil {
			return err
		}
	}

	return nil
}

// GetDraft は email に紐づく sign-up draft を返します。
func (s *RedisSignUpDraftStore) GetDraft(ctx context.Context, email string) (SignUpDraft, error) {
	if s == nil || s.client == nil {
		return SignUpDraft{}, fmt.Errorf("sign up draft store is not initialized")
	}
	if strings.TrimSpace(email) == "" {
		return SignUpDraft{}, fmt.Errorf("email is required")
	}

	payload, err := s.client.Get(ctx, redisSignUpDraftKey(email)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return SignUpDraft{}, ErrSignUpDraftNotFound
		}
		return SignUpDraft{}, fmt.Errorf("get sign up draft email=%s: %w", email, err)
	}

	var draft SignUpDraft
	if err := json.Unmarshal(payload, &draft); err != nil {
		return SignUpDraft{}, fmt.Errorf("unmarshal sign up draft email=%s: %w", email, err)
	}

	return draft, nil
}

// IsHandleReserved は sign-up 中の handle reservation が存在するか返します。
func (s *RedisSignUpDraftStore) IsHandleReserved(ctx context.Context, handle string) (bool, error) {
	if s == nil || s.client == nil {
		return false, fmt.Errorf("sign up draft store is not initialized")
	}

	normalizedHandle := strings.TrimSpace(handle)
	if normalizedHandle == "" {
		return false, fmt.Errorf("handle is required")
	}

	_, err := s.client.Get(ctx, redisSignUpHandleKey(normalizedHandle)).Result()
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, goredis.Nil):
		return false, nil
	default:
		return false, fmt.Errorf("get sign up handle reservation handle=%s: %w", normalizedHandle, err)
	}
}

// DeleteDraft は sign-up draft と関連 handle reservation を削除します。
func (s *RedisSignUpDraftStore) DeleteDraft(ctx context.Context, email string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("sign up draft store is not initialized")
	}
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email is required")
	}

	draft, err := s.GetDraft(ctx, email)
	switch {
	case err == nil:
	case errors.Is(err, ErrSignUpDraftNotFound):
		return nil
	default:
		return err
	}

	if err := s.client.Del(ctx, redisSignUpDraftKey(email)).Err(); err != nil {
		return fmt.Errorf("delete sign up draft email=%s: %w", email, err)
	}

	if err := s.releaseHandle(ctx, draft.Handle); err != nil {
		return err
	}

	return nil
}

func (s *RedisSignUpDraftStore) claimHandle(ctx context.Context, handle string, email string, ttl time.Duration) error {
	normalizedHandle := strings.TrimSpace(handle)
	normalizedEmail := strings.TrimSpace(email)
	key := redisSignUpHandleKey(normalizedHandle)

	claimed, err := s.client.SetNX(ctx, key, normalizedEmail, ttl).Result()
	if err != nil {
		return fmt.Errorf("claim sign up handle handle=%s: %w", normalizedHandle, err)
	}
	if claimed {
		return nil
	}

	reservedBy, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return ErrHandleAlreadyTaken
		}
		return fmt.Errorf("get sign up handle reservation handle=%s: %w", normalizedHandle, err)
	}
	if strings.TrimSpace(reservedBy) != normalizedEmail {
		return ErrHandleAlreadyTaken
	}

	if err := s.client.Set(ctx, key, normalizedEmail, ttl).Err(); err != nil {
		return fmt.Errorf("refresh sign up handle reservation handle=%s: %w", normalizedHandle, err)
	}

	return nil
}

func (s *RedisSignUpDraftStore) releaseHandle(ctx context.Context, handle string) error {
	if strings.TrimSpace(handle) == "" {
		return nil
	}

	if err := s.client.Del(ctx, redisSignUpHandleKey(handle)).Err(); err != nil {
		return fmt.Errorf("release sign up handle handle=%s: %w", handle, err)
	}

	return nil
}

func redisSignUpDraftKey(email string) string {
	return signUpDraftKeyPrefix + strings.TrimSpace(email)
}

func redisSignUpHandleKey(handle string) string {
	return signUpHandleReservationKey + strings.TrimSpace(handle)
}
