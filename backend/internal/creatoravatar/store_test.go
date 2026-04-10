package creatoravatar

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

func TestNewRedisUploadStoreAndRedisUploadKey(t *testing.T) {
	t.Parallel()

	store := NewRedisUploadStore(nil)
	if store == nil {
		t.Fatal("NewRedisUploadStore() = nil, want non-nil")
	}
	if got := redisUploadKey("  token  "); got != uploadKeyPrefix+"token" {
		t.Fatalf("redisUploadKey() got %q want %q", got, uploadKeyPrefix+"token")
	}
}

func TestRedisUploadStoreSaveGetDeleteSuccess(t *testing.T) {
	t.Parallel()

	var savedPayload []byte
	store := &RedisUploadStore{
		client: redisClientStub{
			set: func(_ context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
				if key != uploadKeyPrefix+"token" {
					t.Fatalf("Set() key got %q want %q", key, uploadKeyPrefix+"token")
				}
				if expiration != 10*time.Minute {
					t.Fatalf("Set() expiration got %s want %s", expiration, 10*time.Minute)
				}
				payload, ok := value.([]byte)
				if !ok {
					t.Fatalf("Set() payload type got %T want []byte", value)
				}
				savedPayload = payload
				return goredis.NewStatusResult("OK", nil)
			},
			get: func(_ context.Context, key string) *goredis.StringCmd {
				if key != uploadKeyPrefix+"token" {
					t.Fatalf("Get() key got %q want %q", key, uploadKeyPrefix+"token")
				}
				return goredis.NewStringResult(string(savedPayload), nil)
			},
			del: func(_ context.Context, keys ...string) *goredis.IntCmd {
				if len(keys) != 1 || keys[0] != uploadKeyPrefix+"token" {
					t.Fatalf("Del() keys got %#v want [%q]", keys, uploadKeyPrefix+"token")
				}
				return goredis.NewIntResult(1, nil)
			},
		},
	}

	expected := storedUpload{
		ExpiresAt:     time.Unix(1710000000, 0).UTC(),
		FileName:      "mina.png",
		FileSizeBytes: 42,
		MimeType:      "image/png",
		State:         uploadStateCreated,
		UploadKey:     "creator-avatar-upload/viewer/token/mina.png",
		ViewerUserID:  "viewer-id",
	}

	if err := store.SaveUpload(context.Background(), "token", expected, 10*time.Minute); err != nil {
		t.Fatalf("SaveUpload() error = %v, want nil", err)
	}

	var marshaled storedUpload
	if err := json.Unmarshal(savedPayload, &marshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if marshaled.FileName != expected.FileName {
		t.Fatalf("SaveUpload() payload got %#v want file name %q", marshaled, expected.FileName)
	}

	got, err := store.GetUpload(context.Background(), "token")
	if err != nil {
		t.Fatalf("GetUpload() error = %v, want nil", err)
	}
	if got.UploadKey != expected.UploadKey {
		t.Fatalf("GetUpload() upload key got %q want %q", got.UploadKey, expected.UploadKey)
	}

	if err := store.DeleteUpload(context.Background(), "token"); err != nil {
		t.Fatalf("DeleteUpload() error = %v, want nil", err)
	}
}

func TestRedisUploadStoreErrors(t *testing.T) {
	t.Parallel()

	if err := (&RedisUploadStore{}).SaveUpload(context.Background(), "token", storedUpload{}, time.Minute); err == nil {
		t.Fatal("SaveUpload() error = nil, want error for uninitialized store")
	}
	if _, err := (&RedisUploadStore{}).GetUpload(context.Background(), "token"); err == nil {
		t.Fatal("GetUpload() error = nil, want error for uninitialized store")
	}
	if err := (&RedisUploadStore{}).DeleteUpload(context.Background(), "token"); err == nil {
		t.Fatal("DeleteUpload() error = nil, want error for uninitialized store")
	}

	setErr := errors.New("set failed")
	store := &RedisUploadStore{
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

	if err := store.SaveUpload(context.Background(), "token", storedUpload{}, time.Minute); !errors.Is(err, setErr) {
		t.Fatalf("SaveUpload() error got %v want wrapped %v", err, setErr)
	}

	if _, err := store.GetUpload(context.Background(), "token"); !errors.Is(err, ErrUploadNotFound) {
		t.Fatalf("GetUpload() error got %v want %v", err, ErrUploadNotFound)
	}

	store.client = redisClientStub{
		set: func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd {
			return goredis.NewStatusResult("OK", nil)
		},
		get: func(context.Context, string) *goredis.StringCmd {
			return goredis.NewStringResult("{invalid", nil)
		},
		del: func(context.Context, ...string) *goredis.IntCmd {
			return goredis.NewIntResult(0, nil)
		},
	}
	if _, err := store.GetUpload(context.Background(), "token"); err == nil {
		t.Fatal("GetUpload() error = nil, want unmarshal error")
	}
}
