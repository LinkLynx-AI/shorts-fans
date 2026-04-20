package recommendation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type signalExposureRedisClientStub struct {
	exists    func(context.Context, ...string) *goredis.IntCmd
	pipelined func(context.Context, func(signalExposurePipeline) error) error
	set       func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd
}

func (s signalExposureRedisClientStub) Exists(ctx context.Context, keys ...string) *goredis.IntCmd {
	return s.exists(ctx, keys...)
}

func (s signalExposureRedisClientStub) Pipelined(ctx context.Context, fn func(signalExposurePipeline) error) error {
	return s.pipelined(ctx, fn)
}

func (s signalExposureRedisClientStub) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	return s.set(ctx, key, value, expiration)
}

type signalExposurePipelineStub struct {
	set func(context.Context, string, interface{}, time.Duration)
}

func (s signalExposurePipelineStub) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) {
	s.set(ctx, key, value, expiration)
}

func TestNewRedisSignalExposureStore(t *testing.T) {
	t.Parallel()

	store := NewRedisSignalExposureStore(nil)
	if store == nil {
		t.Fatal("NewRedisSignalExposureStore() = nil, want non-nil store")
	}
	if store.ttl != defaultSignalExposureTTL {
		t.Fatalf("NewRedisSignalExposureStore() ttl got %s want %s", store.ttl, defaultSignalExposureTTL)
	}
}

func TestRedisSignalExposureStoreRememberAndHasExposure(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	creatorUserID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	var savedKeys []string
	var savedTTL time.Duration

	store := &RedisSignalExposureStore{
		client: signalExposureRedisClientStub{
			exists: func(_ context.Context, keys ...string) *goredis.IntCmd {
				if len(keys) != 1 {
					t.Fatalf("Exists() keys len got %d want %d", len(keys), 1)
				}

				return goredis.NewIntResult(1, nil)
			},
			pipelined: func(context.Context, func(signalExposurePipeline) error) error {
				t.Fatal("Pipelined() should not be called for single-key exposure writes")
				return nil
			},
			set: func(_ context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
				if value != "1" {
					t.Fatalf("Set() value got %#v want %q", value, "1")
				}

				savedKeys = append(savedKeys, key)
				savedTTL = expiration

				return goredis.NewStatusResult("OK", nil)
			},
		},
		ttl: defaultSignalExposureTTL,
	}

	if err := store.RememberShortExposure(context.Background(), viewerID, shortID); err != nil {
		t.Fatalf("RememberShortExposure() error = %v, want nil", err)
	}
	if err := store.RememberCreatorExposure(context.Background(), viewerID, creatorUserID); err != nil {
		t.Fatalf("RememberCreatorExposure() error = %v, want nil", err)
	}

	hasShort, err := store.HasShortExposure(context.Background(), viewerID, shortID)
	if err != nil {
		t.Fatalf("HasShortExposure() error = %v, want nil", err)
	}
	if !hasShort {
		t.Fatal("HasShortExposure() got false want true")
	}

	hasCreator, err := store.HasCreatorExposure(context.Background(), viewerID, creatorUserID)
	if err != nil {
		t.Fatalf("HasCreatorExposure() error = %v, want nil", err)
	}
	if !hasCreator {
		t.Fatal("HasCreatorExposure() got false want true")
	}

	if len(savedKeys) != 2 {
		t.Fatalf("saved keys len got %d want %d", len(savedKeys), 2)
	}
	if savedKeys[0] != shortExposureKey(viewerID, shortID) {
		t.Fatalf("saved short key got %q want %q", savedKeys[0], shortExposureKey(viewerID, shortID))
	}
	if savedKeys[1] != creatorExposureKey(viewerID, creatorUserID) {
		t.Fatalf("saved creator key got %q want %q", savedKeys[1], creatorExposureKey(viewerID, creatorUserID))
	}
	if savedTTL != defaultSignalExposureTTL {
		t.Fatalf("saved ttl got %s want %s", savedTTL, defaultSignalExposureTTL)
	}
}

func TestRedisSignalExposureStoreRememberExposureBatchUsesPipeline(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortIDs := []uuid.UUID{
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		uuid.MustParse("33333333-3333-3333-3333-333333333333"),
	}
	creatorUserIDs := []uuid.UUID{
		uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		uuid.MustParse("55555555-5555-5555-5555-555555555555"),
	}
	var pipelinedKeys []string
	var pipelinedTTL time.Duration

	store := &RedisSignalExposureStore{
		client: signalExposureRedisClientStub{
			exists: func(context.Context, ...string) *goredis.IntCmd {
				t.Fatal("Exists() should not be called during batch remember")
				return goredis.NewIntResult(0, nil)
			},
			pipelined: func(ctx context.Context, fn func(signalExposurePipeline) error) error {
				return fn(signalExposurePipelineStub{
					set: func(_ context.Context, key string, value interface{}, expiration time.Duration) {
						if value != "1" {
							t.Fatalf("pipeline Set() value got %#v want %q", value, "1")
						}

						pipelinedKeys = append(pipelinedKeys, key)
						pipelinedTTL = expiration
					},
				})
			},
			set: func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd {
				t.Fatal("Set() should not be called for batch exposure writes")
				return goredis.NewStatusResult("", nil)
			},
		},
		ttl: defaultSignalExposureTTL,
	}

	if err := store.RememberShortExposures(context.Background(), viewerID, shortIDs); err != nil {
		t.Fatalf("RememberShortExposures() error = %v, want nil", err)
	}
	if err := store.RememberCreatorExposures(context.Background(), viewerID, creatorUserIDs); err != nil {
		t.Fatalf("RememberCreatorExposures() error = %v, want nil", err)
	}

	expectedKeys := []string{
		shortExposureKey(viewerID, shortIDs[0]),
		shortExposureKey(viewerID, shortIDs[1]),
		creatorExposureKey(viewerID, creatorUserIDs[0]),
		creatorExposureKey(viewerID, creatorUserIDs[1]),
	}
	if len(pipelinedKeys) != len(expectedKeys) {
		t.Fatalf("pipelined keys len got %d want %d", len(pipelinedKeys), len(expectedKeys))
	}
	for index, key := range expectedKeys {
		if pipelinedKeys[index] != key {
			t.Fatalf("pipelined key[%d] got %q want %q", index, pipelinedKeys[index], key)
		}
	}
	if pipelinedTTL != defaultSignalExposureTTL {
		t.Fatalf("pipelined ttl got %s want %s", pipelinedTTL, defaultSignalExposureTTL)
	}
}

func TestRedisSignalExposureStoreErrors(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	creatorUserID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	store := &RedisSignalExposureStore{}
	if err := store.RememberShortExposure(context.Background(), viewerID, shortID); err == nil {
		t.Fatal("RememberShortExposure() error = nil, want initialization error")
	}
	if _, err := store.HasCreatorExposure(context.Background(), viewerID, creatorUserID); err == nil {
		t.Fatal("HasCreatorExposure() error = nil, want initialization error")
	}

	setErr := errors.New("redis set failed")
	existsErr := errors.New("redis exists failed")
	pipelinedErr := errors.New("redis pipeline failed")
	store = &RedisSignalExposureStore{
		client: signalExposureRedisClientStub{
			exists: func(context.Context, ...string) *goredis.IntCmd {
				return goredis.NewIntResult(0, existsErr)
			},
			pipelined: func(context.Context, func(signalExposurePipeline) error) error {
				return pipelinedErr
			},
			set: func(context.Context, string, interface{}, time.Duration) *goredis.StatusCmd {
				return goredis.NewStatusResult("", setErr)
			},
		},
		ttl: defaultSignalExposureTTL,
	}

	if err := store.RememberCreatorExposure(context.Background(), viewerID, creatorUserID); !errors.Is(err, setErr) {
		t.Fatalf("RememberCreatorExposure() error got %v want %v", err, setErr)
	}
	if _, err := store.HasShortExposure(context.Background(), viewerID, shortID); !errors.Is(err, existsErr) {
		t.Fatalf("HasShortExposure() error got %v want %v", err, existsErr)
	}
	if err := store.RememberShortExposures(context.Background(), viewerID, []uuid.UUID{shortID, creatorUserID}); !errors.Is(err, pipelinedErr) {
		t.Fatalf("RememberShortExposures() error got %v want %v", err, pipelinedErr)
	}

	store.ttl = 0
	if err := store.RememberShortExposure(context.Background(), viewerID, shortID); err == nil {
		t.Fatal("RememberShortExposure() error = nil, want ttl validation error")
	}
}
