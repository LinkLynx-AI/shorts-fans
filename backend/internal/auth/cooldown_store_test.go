package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type cooldownRedisClientStub struct {
	del   func(context.Context, ...string) *goredis.IntCmd
	setNX func(context.Context, string, interface{}, time.Duration) *goredis.BoolCmd
}

func (s cooldownRedisClientStub) Del(ctx context.Context, keys ...string) *goredis.IntCmd {
	return s.del(ctx, keys...)
}

func (s cooldownRedisClientStub) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.BoolCmd {
	return s.setNX(ctx, key, value, expiration)
}

func TestNewRedisCooldownStore(t *testing.T) {
	t.Parallel()

	if got := NewRedisCooldownStore(nil); got == nil {
		t.Fatal("NewRedisCooldownStore() = nil, want non-nil store")
	}
}

func TestRedisCooldownStoreTryActivateAndRelease(t *testing.T) {
	t.Parallel()

	var activatedKey string
	var activatedTTL time.Duration
	var releasedKey string

	store := &RedisCooldownStore{
		client: cooldownRedisClientStub{
			setNX: func(_ context.Context, key string, _ interface{}, expiration time.Duration) *goredis.BoolCmd {
				activatedKey = key
				activatedTTL = expiration
				return goredis.NewBoolResult(true, nil)
			},
			del: func(_ context.Context, keys ...string) *goredis.IntCmd {
				if len(keys) != 1 {
					t.Fatalf("Del() key count got %d want %d", len(keys), 1)
				}
				releasedKey = keys[0]
				return goredis.NewIntResult(1, nil)
			},
		},
	}

	activated, err := store.TryActivate(context.Background(), " sign_up:fan@example.com ", time.Minute)
	if err != nil {
		t.Fatalf("TryActivate() error = %v, want nil", err)
	}
	if !activated {
		t.Fatal("TryActivate() activated got false want true")
	}
	if activatedKey != "fan_auth_cooldown:sign_up:fan@example.com" {
		t.Fatalf("TryActivate() key got %q want %q", activatedKey, "fan_auth_cooldown:sign_up:fan@example.com")
	}
	if activatedTTL != time.Minute {
		t.Fatalf("TryActivate() ttl got %s want %s", activatedTTL, time.Minute)
	}

	if err := store.Release(context.Background(), " sign_up:fan@example.com "); err != nil {
		t.Fatalf("Release() error = %v, want nil", err)
	}
	if releasedKey != "fan_auth_cooldown:sign_up:fan@example.com" {
		t.Fatalf("Release() key got %q want %q", releasedKey, "fan_auth_cooldown:sign_up:fan@example.com")
	}
}

func TestRedisCooldownStoreValidatesInput(t *testing.T) {
	t.Parallel()

	store := &RedisCooldownStore{}
	if _, err := store.TryActivate(context.Background(), "sign-up", time.Minute); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("TryActivate() error got %v want initialization error", err)
	}

	store = &RedisCooldownStore{
		client: cooldownRedisClientStub{
			setNX: func(context.Context, string, interface{}, time.Duration) *goredis.BoolCmd {
				return goredis.NewBoolResult(true, nil)
			},
			del: func(context.Context, ...string) *goredis.IntCmd {
				return goredis.NewIntResult(1, nil)
			},
		},
	}
	if _, err := store.TryActivate(context.Background(), " ", time.Minute); err == nil || !strings.Contains(err.Error(), "key is required") {
		t.Fatalf("TryActivate() error got %v want key validation error", err)
	}
	if _, err := store.TryActivate(context.Background(), "sign-up", 0); err == nil || !strings.Contains(err.Error(), "greater than zero") {
		t.Fatalf("TryActivate() error got %v want ttl validation error", err)
	}
	if err := store.Release(context.Background(), " "); err == nil || !strings.Contains(err.Error(), "key is required") {
		t.Fatalf("Release() error got %v want key validation error", err)
	}
}
