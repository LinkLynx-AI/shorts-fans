package creatorregistration

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type creatorRegistrationRedisClientStub struct {
	del func(context.Context, ...string) *goredis.IntCmd
	get func(context.Context, string) *goredis.StringCmd
	set func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd
}

func (s creatorRegistrationRedisClientStub) Del(ctx context.Context, keys ...string) *goredis.IntCmd {
	return s.del(ctx, keys...)
}

func (s creatorRegistrationRedisClientStub) Get(ctx context.Context, key string) *goredis.StringCmd {
	return s.get(ctx, key)
}

func (s creatorRegistrationRedisClientStub) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	return s.set(ctx, key, value, expiration)
}

func TestNewRedisEvidenceUploadStoreAndKey(t *testing.T) {
	t.Parallel()

	store := NewRedisEvidenceUploadStore(nil)
	if store == nil {
		t.Fatal("NewRedisEvidenceUploadStore() = nil, want non-nil")
	}
	if got := redisEvidenceUploadKey("  token  "); got != evidenceUploadKeyPrefix+"token" {
		t.Fatalf("redisEvidenceUploadKey() got %q want %q", got, evidenceUploadKeyPrefix+"token")
	}
}

func TestRedisEvidenceUploadStoreSaveGetDeleteSuccess(t *testing.T) {
	t.Parallel()

	var savedPayload []byte
	store := &RedisEvidenceUploadStore{
		client: creatorRegistrationRedisClientStub{
			set: func(_ context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
				if key != evidenceUploadKeyPrefix+"token" {
					t.Fatalf("Set() key got %q want %q", key, evidenceUploadKeyPrefix+"token")
				}
				if expiration != 5*time.Minute {
					t.Fatalf("Set() expiration got %s want %s", expiration, 5*time.Minute)
				}
				payload, ok := value.([]byte)
				if !ok {
					t.Fatalf("Set() payload type got %T want []byte", value)
				}
				savedPayload = payload
				return goredis.NewStatusResult("OK", nil)
			},
			get: func(_ context.Context, key string) *goredis.StringCmd {
				if key != evidenceUploadKeyPrefix+"token" {
					t.Fatalf("Get() key got %q want %q", key, evidenceUploadKeyPrefix+"token")
				}
				return goredis.NewStringResult(string(savedPayload), nil)
			},
			del: func(_ context.Context, keys ...string) *goredis.IntCmd {
				if len(keys) != 1 || keys[0] != evidenceUploadKeyPrefix+"token" {
					t.Fatalf("Del() keys got %#v want [%q]", keys, evidenceUploadKeyPrefix+"token")
				}
				return goredis.NewIntResult(1, nil)
			},
		},
	}

	expected := storedEvidenceUpload{
		ExpiresAt:     time.Unix(1710000000, 0).UTC(),
		FileName:      "government-id.png",
		FileSizeBytes: 123,
		Kind:          EvidenceKindGovernmentID,
		MimeType:      "image/png",
		State:         evidenceUploadStateCreated,
		UploadKey:     "creator-registration/evidence/upload.png",
		ViewerUserID:  "viewer-id",
	}

	if err := store.SaveUpload(context.Background(), "token", expected, 5*time.Minute); err != nil {
		t.Fatalf("SaveUpload() error = %v, want nil", err)
	}

	var marshaled storedEvidenceUpload
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
		t.Fatalf("GetUpload() UploadKey got %q want %q", got.UploadKey, expected.UploadKey)
	}

	if err := store.DeleteUpload(context.Background(), "token"); err != nil {
		t.Fatalf("DeleteUpload() error = %v, want nil", err)
	}
}

func TestRedisEvidenceUploadStoreErrors(t *testing.T) {
	t.Parallel()

	if err := (&RedisEvidenceUploadStore{}).SaveUpload(context.Background(), "token", storedEvidenceUpload{}, time.Minute); err == nil {
		t.Fatal("SaveUpload() error = nil, want error for uninitialized store")
	}
	if _, err := (&RedisEvidenceUploadStore{}).GetUpload(context.Background(), "token"); err == nil {
		t.Fatal("GetUpload() error = nil, want error for uninitialized store")
	}
	if err := (&RedisEvidenceUploadStore{}).DeleteUpload(context.Background(), "token"); err == nil {
		t.Fatal("DeleteUpload() error = nil, want error for uninitialized store")
	}

	setErr := errors.New("set failed")
	deleteErr := errors.New("delete failed")
	store := &RedisEvidenceUploadStore{
		client: creatorRegistrationRedisClientStub{
			set: func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd {
				return goredis.NewStatusResult("", setErr)
			},
			get: func(context.Context, string) *goredis.StringCmd {
				return goredis.NewStringResult("", goredis.Nil)
			},
			del: func(context.Context, ...string) *goredis.IntCmd {
				return goredis.NewIntResult(0, deleteErr)
			},
		},
	}

	if err := store.SaveUpload(context.Background(), "token", storedEvidenceUpload{}, time.Minute); !errors.Is(err, setErr) {
		t.Fatalf("SaveUpload() error got %v want wrapped %v", err, setErr)
	}
	if _, err := store.GetUpload(context.Background(), "token"); !errors.Is(err, ErrEvidenceUploadNotFound) {
		t.Fatalf("GetUpload() error got %v want %v", err, ErrEvidenceUploadNotFound)
	}
	if err := store.DeleteUpload(context.Background(), "token"); !errors.Is(err, deleteErr) {
		t.Fatalf("DeleteUpload() error got %v want wrapped %v", err, deleteErr)
	}

	store.client = creatorRegistrationRedisClientStub{
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

func TestRedisEvidenceUploadStoreInputValidation(t *testing.T) {
	t.Parallel()

	store := &RedisEvidenceUploadStore{
		client: creatorRegistrationRedisClientStub{
			set: func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd {
				return goredis.NewStatusResult("OK", nil)
			},
			get: func(context.Context, string) *goredis.StringCmd {
				return goredis.NewStringResult("", nil)
			},
			del: func(context.Context, ...string) *goredis.IntCmd {
				return goredis.NewIntResult(0, nil)
			},
		},
	}

	if err := store.DeleteUpload(context.Background(), "   "); err == nil {
		t.Fatal("DeleteUpload() error = nil, want token validation error")
	}
	if _, err := store.GetUpload(context.Background(), "   "); err == nil {
		t.Fatal("GetUpload() error = nil, want token validation error")
	}
	if err := store.SaveUpload(context.Background(), "   ", storedEvidenceUpload{}, time.Minute); err == nil {
		t.Fatal("SaveUpload() error = nil, want token validation error")
	}
	if err := store.SaveUpload(context.Background(), "token", storedEvidenceUpload{}, 0); err == nil {
		t.Fatal("SaveUpload() error = nil, want ttl validation error")
	}
}
