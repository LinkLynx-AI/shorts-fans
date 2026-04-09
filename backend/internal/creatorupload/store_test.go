package creatorupload

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type redisClientStub struct {
	del func(context.Context, ...string) *goredis.IntCmd
	get func(context.Context, string) *goredis.StringCmd
	set func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd
}

func (s redisClientStub) Del(ctx context.Context, keys ...string) *goredis.IntCmd {
	return s.del(ctx, keys...)
}

func (s redisClientStub) Get(ctx context.Context, key string) *goredis.StringCmd {
	return s.get(ctx, key)
}

func (s redisClientStub) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	return s.set(ctx, key, value, expiration)
}

func TestNewRedisPackageStoreAndRedisPackageKey(t *testing.T) {
	t.Parallel()

	store := NewRedisPackageStore(nil)
	if store == nil {
		t.Fatal("NewRedisPackageStore() = nil, want non-nil")
	}
	if got := redisPackageKey("  pkg-token  "); got != packageKeyPrefix+"pkg-token" {
		t.Fatalf("redisPackageKey() got %q want %q", got, packageKeyPrefix+"pkg-token")
	}
}

func TestRedisPackageStoreSaveGetDeleteSuccess(t *testing.T) {
	t.Parallel()

	var savedPayload []byte
	var savedTTL time.Duration
	store := &RedisPackageStore{
		client: redisClientStub{
			set: func(_ context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
				if key != packageKeyPrefix+"pkg-token" {
					t.Fatalf("Set() key got %q want %q", key, packageKeyPrefix+"pkg-token")
				}
				payload, ok := value.([]byte)
				if !ok {
					t.Fatalf("Set() payload type got %T want []byte", value)
				}
				savedPayload = payload
				savedTTL = expiration
				return goredis.NewStatusResult("OK", nil)
			},
			get: func(_ context.Context, key string) *goredis.StringCmd {
				if key != packageKeyPrefix+"pkg-token" {
					t.Fatalf("Get() key got %q want %q", key, packageKeyPrefix+"pkg-token")
				}
				return goredis.NewStringResult(string(savedPayload), nil)
			},
			del: func(_ context.Context, keys ...string) *goredis.IntCmd {
				if len(keys) != 1 || keys[0] != packageKeyPrefix+"pkg-token" {
					t.Fatalf("Del() keys got %#v want [%q]", keys, packageKeyPrefix+"pkg-token")
				}
				return goredis.NewIntResult(1, nil)
			},
		},
	}

	expected := storedPackage{
		CreatorUserID: "creator-id",
		ExpiresAt:     time.Unix(1710000000, 0).UTC(),
		Main: storedEntry{
			UploadEntryID: "main-entry",
			Role:          roleMain,
		},
		Shorts: []storedEntry{{
			UploadEntryID: "short-entry",
			Role:          roleShort,
		}},
	}

	if err := store.SavePackage(context.Background(), "pkg-token", expected, 15*time.Minute); err != nil {
		t.Fatalf("SavePackage() error = %v, want nil", err)
	}
	if savedTTL != 15*time.Minute {
		t.Fatalf("SavePackage() ttl got %s want %s", savedTTL, 15*time.Minute)
	}

	var marshaled storedPackage
	if err := json.Unmarshal(savedPayload, &marshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if marshaled.CreatorUserID != expected.CreatorUserID {
		t.Fatalf("SavePackage() payload got %#v want creator id %q", marshaled, expected.CreatorUserID)
	}

	got, err := store.GetPackage(context.Background(), "pkg-token")
	if err != nil {
		t.Fatalf("GetPackage() error = %v, want nil", err)
	}
	if got.Main.UploadEntryID != "main-entry" {
		t.Fatalf("GetPackage() main got %#v want uploadEntryId", got.Main)
	}

	if err := store.DeletePackage(context.Background(), "pkg-token"); err != nil {
		t.Fatalf("DeletePackage() error = %v, want nil", err)
	}
}

func TestRedisPackageStoreErrors(t *testing.T) {
	t.Parallel()

	if err := (&RedisPackageStore{}).SavePackage(context.Background(), "pkg-token", storedPackage{}, time.Minute); err == nil {
		t.Fatal("SavePackage() error = nil, want error for uninitialized store")
	}
	if _, err := (&RedisPackageStore{}).GetPackage(context.Background(), "pkg-token"); err == nil {
		t.Fatal("GetPackage() error = nil, want error for uninitialized store")
	}
	if err := (&RedisPackageStore{}).DeletePackage(context.Background(), "pkg-token"); err == nil {
		t.Fatal("DeletePackage() error = nil, want error for uninitialized store")
	}
	if err := (&RedisPackageStore{client: redisClientStub{}}).SavePackage(context.Background(), "   ", storedPackage{}, time.Minute); err == nil {
		t.Fatal("SavePackage() error = nil, want error for blank package token")
	}
	if err := (&RedisPackageStore{client: redisClientStub{}}).SavePackage(context.Background(), "pkg-token", storedPackage{}, 0); err == nil {
		t.Fatal("SavePackage() error = nil, want error for non-positive ttl")
	}
	if _, err := (&RedisPackageStore{client: redisClientStub{}}).GetPackage(context.Background(), "   "); err == nil {
		t.Fatal("GetPackage() error = nil, want error for blank package token")
	}
	if err := (&RedisPackageStore{client: redisClientStub{}}).DeletePackage(context.Background(), "   "); err == nil {
		t.Fatal("DeletePackage() error = nil, want error for blank package token")
	}

	setErr := errors.New("set failed")
	store := &RedisPackageStore{
		client: redisClientStub{
			set: func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd {
				return goredis.NewStatusResult("", setErr)
			},
			get: func(context.Context, string) *goredis.StringCmd {
				return goredis.NewStringResult("", goredis.Nil)
			},
			del: func(context.Context, ...string) *goredis.IntCmd {
				return goredis.NewIntResult(0, errors.New("delete failed"))
			},
		},
	}

	if err := store.SavePackage(context.Background(), "pkg-token", storedPackage{}, time.Minute); !errors.Is(err, setErr) {
		t.Fatalf("SavePackage() error got %v want wrapped %v", err, setErr)
	}

	if _, err := store.GetPackage(context.Background(), "pkg-token"); !errors.Is(err, ErrPackageNotFound) {
		t.Fatalf("GetPackage() error got %v want %v", err, ErrPackageNotFound)
	}

	store.client = redisClientStub{
		set: func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd {
			return goredis.NewStatusResult("OK", nil)
		},
		get: func(context.Context, string) *goredis.StringCmd {
			return goredis.NewStringResult("{invalid", nil)
		},
		del: func(context.Context, ...string) *goredis.IntCmd {
			return goredis.NewIntResult(0, errors.New("delete failed"))
		},
	}
	if _, err := store.GetPackage(context.Background(), "pkg-token"); err == nil {
		t.Fatal("GetPackage() error = nil, want unmarshal error")
	}

	deleteErr := errors.New("delete failed")
	store.client = redisClientStub{
		set: func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd {
			return goredis.NewStatusResult("OK", nil)
		},
		get: func(context.Context, string) *goredis.StringCmd {
			return goredis.NewStringResult("", nil)
		},
		del: func(context.Context, ...string) *goredis.IntCmd {
			return goredis.NewIntResult(0, deleteErr)
		},
	}
	if err := store.DeletePackage(context.Background(), "pkg-token"); !errors.Is(err, deleteErr) {
		t.Fatalf("DeletePackage() error got %v want wrapped %v", err, deleteErr)
	}
}
