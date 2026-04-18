package recommendation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type unlockConversionRetryRedisClientStub struct {
	del    func(context.Context, ...string) *goredis.IntCmd
	exists func(context.Context, ...string) *goredis.IntCmd
	set    func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd
}

func (s unlockConversionRetryRedisClientStub) Del(ctx context.Context, keys ...string) *goredis.IntCmd {
	return s.del(ctx, keys...)
}

func (s unlockConversionRetryRedisClientStub) Exists(ctx context.Context, keys ...string) *goredis.IntCmd {
	return s.exists(ctx, keys...)
}

func (s unlockConversionRetryRedisClientStub) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	return s.set(ctx, key, value, expiration)
}

func TestNewRedisUnlockConversionRetryStore(t *testing.T) {
	t.Parallel()

	store := NewRedisUnlockConversionRetryStore(nil)
	if store == nil {
		t.Fatal("NewRedisUnlockConversionRetryStore() = nil, want non-nil store")
	}
	if store.ttl != defaultUnlockConversionRetryTTL {
		t.Fatalf("NewRedisUnlockConversionRetryStore() ttl got %s want %s", store.ttl, defaultUnlockConversionRetryTTL)
	}
}

func TestRedisUnlockConversionRetryStoreRoundTrip(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	shortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	key := unlockConversionRetryKey(viewerID, mainID, shortID)
	var savedTTL time.Duration

	store := &RedisUnlockConversionRetryStore{
		client: unlockConversionRetryRedisClientStub{
			del: func(_ context.Context, keys ...string) *goredis.IntCmd {
				if len(keys) != 1 || keys[0] != key {
					t.Fatalf("Del() keys got %v want [%s]", keys, key)
				}

				return goredis.NewIntResult(1, nil)
			},
			exists: func(_ context.Context, keys ...string) *goredis.IntCmd {
				if len(keys) != 1 || keys[0] != key {
					t.Fatalf("Exists() keys got %v want [%s]", keys, key)
				}

				return goredis.NewIntResult(1, nil)
			},
			set: func(_ context.Context, gotKey string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
				if gotKey != key {
					t.Fatalf("Set() key got %q want %q", gotKey, key)
				}
				if value != "1" {
					t.Fatalf("Set() value got %#v want %q", value, "1")
				}

				savedTTL = expiration
				return goredis.NewStatusResult("OK", nil)
			},
		},
		ttl: defaultUnlockConversionRetryTTL,
	}

	if err := store.MarkPendingUnlockConversion(context.Background(), viewerID, mainID, shortID); err != nil {
		t.Fatalf("MarkPendingUnlockConversion() error = %v, want nil", err)
	}
	if savedTTL != defaultUnlockConversionRetryTTL {
		t.Fatalf("saved ttl got %s want %s", savedTTL, defaultUnlockConversionRetryTTL)
	}

	hasPending, err := store.HasPendingUnlockConversion(context.Background(), viewerID, mainID, shortID)
	if err != nil {
		t.Fatalf("HasPendingUnlockConversion() error = %v, want nil", err)
	}
	if !hasPending {
		t.Fatal("HasPendingUnlockConversion() got false want true")
	}

	if err := store.ClearPendingUnlockConversion(context.Background(), viewerID, mainID, shortID); err != nil {
		t.Fatalf("ClearPendingUnlockConversion() error = %v, want nil", err)
	}
}

func TestRedisUnlockConversionRetryStoreErrors(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	shortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	store := &RedisUnlockConversionRetryStore{}
	if err := store.MarkPendingUnlockConversion(context.Background(), viewerID, mainID, shortID); err == nil {
		t.Fatal("MarkPendingUnlockConversion() error = nil, want initialization error")
	}
	if _, err := store.HasPendingUnlockConversion(context.Background(), viewerID, mainID, shortID); err == nil {
		t.Fatal("HasPendingUnlockConversion() error = nil, want initialization error")
	}
	if err := store.ClearPendingUnlockConversion(context.Background(), viewerID, mainID, shortID); err == nil {
		t.Fatal("ClearPendingUnlockConversion() error = nil, want initialization error")
	}

	setErr := errors.New("redis set failed")
	existsErr := errors.New("redis exists failed")
	delErr := errors.New("redis del failed")

	store = &RedisUnlockConversionRetryStore{
		client: unlockConversionRetryRedisClientStub{
			del: func(context.Context, ...string) *goredis.IntCmd {
				return goredis.NewIntResult(0, delErr)
			},
			exists: func(context.Context, ...string) *goredis.IntCmd {
				return goredis.NewIntResult(0, existsErr)
			},
			set: func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd {
				return goredis.NewStatusResult("", setErr)
			},
		},
		ttl: defaultUnlockConversionRetryTTL,
	}

	if err := store.MarkPendingUnlockConversion(context.Background(), viewerID, mainID, shortID); !errors.Is(err, setErr) {
		t.Fatalf("MarkPendingUnlockConversion() error got %v want %v", err, setErr)
	}
	if _, err := store.HasPendingUnlockConversion(context.Background(), viewerID, mainID, shortID); !errors.Is(err, existsErr) {
		t.Fatalf("HasPendingUnlockConversion() error got %v want %v", err, existsErr)
	}
	if err := store.ClearPendingUnlockConversion(context.Background(), viewerID, mainID, shortID); !errors.Is(err, delErr) {
		t.Fatalf("ClearPendingUnlockConversion() error got %v want %v", err, delErr)
	}

	store.ttl = 0
	if err := store.MarkPendingUnlockConversion(context.Background(), viewerID, mainID, shortID); err == nil {
		t.Fatal("MarkPendingUnlockConversion() error = nil, want ttl validation error")
	}
}
