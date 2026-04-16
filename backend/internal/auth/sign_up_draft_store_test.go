package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type signUpDraftRedisClientStub struct {
	values  map[string]string
	delErr  error
	getErr  error
	setErr  error
	setNXErr error
}

func (s *signUpDraftRedisClientStub) Del(_ context.Context, keys ...string) *goredis.IntCmd {
	if s.delErr != nil {
		return goredis.NewIntResult(0, s.delErr)
	}

	var count int64
	for _, key := range keys {
		if _, ok := s.values[key]; ok {
			delete(s.values, key)
			count++
		}
	}

	return goredis.NewIntResult(count, nil)
}

func (s *signUpDraftRedisClientStub) Get(_ context.Context, key string) *goredis.StringCmd {
	if s.getErr != nil {
		return goredis.NewStringResult("", s.getErr)
	}
	value, ok := s.values[key]
	if !ok {
		return goredis.NewStringResult("", goredis.Nil)
	}

	return goredis.NewStringResult(value, nil)
}

func (s *signUpDraftRedisClientStub) Set(_ context.Context, key string, value interface{}, _ time.Duration) *goredis.StatusCmd {
	if s.setErr != nil {
		return goredis.NewStatusResult("", s.setErr)
	}
	s.values[key] = stringifyRedisValue(value)
	return goredis.NewStatusResult("OK", nil)
}

func (s *signUpDraftRedisClientStub) SetNX(_ context.Context, key string, value interface{}, _ time.Duration) *goredis.BoolCmd {
	if s.setNXErr != nil {
		return goredis.NewBoolResult(false, s.setNXErr)
	}
	if _, ok := s.values[key]; ok {
		return goredis.NewBoolResult(false, nil)
	}

	s.values[key] = stringifyRedisValue(value)
	return goredis.NewBoolResult(true, nil)
}

func TestNewRedisSignUpDraftStore(t *testing.T) {
	t.Parallel()

	if got := NewRedisSignUpDraftStore(nil); got == nil {
		t.Fatal("NewRedisSignUpDraftStore() = nil, want non-nil store")
	}
}

func TestRedisSignUpDraftStoreSaveGetAndDeleteDraft(t *testing.T) {
	t.Parallel()

	client := &signUpDraftRedisClientStub{values: map[string]string{}}
	store := &RedisSignUpDraftStore{client: client}
	draft := SignUpDraft{
		DisplayName: "Mina",
		Email:       "fan@example.com",
		ExpiresAt:   time.Unix(1712000500, 0).UTC(),
		Handle:      "mina",
		Password:    "VeryStrongPass123!",
	}

	if err := store.SaveDraft(context.Background(), " fan@example.com ", draft, time.Hour); err != nil {
		t.Fatalf("SaveDraft() error = %v, want nil", err)
	}

	got, err := store.GetDraft(context.Background(), "fan@example.com")
	if err != nil {
		t.Fatalf("GetDraft() error = %v, want nil", err)
	}
	if got.DisplayName != "Mina" {
		t.Fatalf("GetDraft() display name got %q want %q", got.DisplayName, "Mina")
	}
	if got.Handle != "mina" {
		t.Fatalf("GetDraft() handle got %q want %q", got.Handle, "mina")
	}
	if client.values["fan_auth_sign_up_handle:mina"] != "fan@example.com" {
		t.Fatalf("SaveDraft() reserved handle got %q want %q", client.values["fan_auth_sign_up_handle:mina"], "fan@example.com")
	}

	if err := store.DeleteDraft(context.Background(), "fan@example.com"); err != nil {
		t.Fatalf("DeleteDraft() error = %v, want nil", err)
	}
	if _, ok := client.values["fan_auth_sign_up_handle:mina"]; ok {
		t.Fatal("DeleteDraft() did not release handle reservation")
	}
	if _, err := store.GetDraft(context.Background(), "fan@example.com"); err == nil || !strings.Contains(err.Error(), ErrSignUpDraftNotFound.Error()) {
		t.Fatalf("GetDraft() after delete error got %v want %v", err, ErrSignUpDraftNotFound)
	}
}

func TestRedisSignUpDraftStoreIsHandleReserved(t *testing.T) {
	t.Parallel()

	store := &RedisSignUpDraftStore{client: &signUpDraftRedisClientStub{
		values: map[string]string{
			"fan_auth_sign_up_handle:mina": "fan@example.com",
		},
	}}

	reserved, err := store.IsHandleReserved(context.Background(), " mina ")
	if err != nil {
		t.Fatalf("IsHandleReserved() error = %v, want nil", err)
	}
	if !reserved {
		t.Fatal("IsHandleReserved() = false, want true")
	}

	reserved, err = store.IsHandleReserved(context.Background(), "other")
	if err != nil {
		t.Fatalf("IsHandleReserved() missing handle error = %v, want nil", err)
	}
	if reserved {
		t.Fatal("IsHandleReserved() = true, want false for missing handle")
	}
}

func TestRedisSignUpDraftStoreIsHandleReservedValidatesInput(t *testing.T) {
	t.Parallel()

	var store *RedisSignUpDraftStore
	if _, err := store.IsHandleReserved(context.Background(), "mina"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("IsHandleReserved() error got %v want initialization error", err)
	}

	store = &RedisSignUpDraftStore{client: &signUpDraftRedisClientStub{values: map[string]string{}}}
	if _, err := store.IsHandleReserved(context.Background(), "   "); err == nil || !strings.Contains(err.Error(), "handle is required") {
		t.Fatalf("IsHandleReserved() error got %v want handle validation error", err)
	}
}

func TestRedisSignUpDraftStoreSaveDraftDetectsTakenHandle(t *testing.T) {
	t.Parallel()

	client := &signUpDraftRedisClientStub{
		values: map[string]string{
			"fan_auth_sign_up_handle:mina": "other@example.com",
		},
	}
	store := &RedisSignUpDraftStore{client: client}

	err := store.SaveDraft(context.Background(), "fan@example.com", SignUpDraft{
		DisplayName: "Mina",
		Handle:      "mina",
		Password:    "VeryStrongPass123!",
	}, time.Hour)
	if err == nil || !strings.Contains(err.Error(), ErrHandleAlreadyTaken.Error()) {
		t.Fatalf("SaveDraft() error got %v want %v", err, ErrHandleAlreadyTaken)
	}
}

func TestRedisSignUpDraftStoreSaveDraftReleasesOldHandleOnHandleChange(t *testing.T) {
	t.Parallel()

	client := &signUpDraftRedisClientStub{values: map[string]string{}}
	store := &RedisSignUpDraftStore{client: client}

	if err := store.SaveDraft(context.Background(), "fan@example.com", SignUpDraft{
		DisplayName: "Mina",
		Handle:      "mina",
		Password:    "VeryStrongPass123!",
	}, time.Hour); err != nil {
		t.Fatalf("first SaveDraft() error = %v, want nil", err)
	}

	if err := store.SaveDraft(context.Background(), "fan@example.com", SignUpDraft{
		DisplayName: "Mina",
		Handle:      "mina-new",
		Password:    "VeryStrongPass123!",
	}, time.Hour); err != nil {
		t.Fatalf("second SaveDraft() error = %v, want nil", err)
	}

	if _, ok := client.values["fan_auth_sign_up_handle:mina"]; ok {
		t.Fatal("SaveDraft() kept old handle reservation")
	}
	if got := client.values["fan_auth_sign_up_handle:mina-new"]; got != "fan@example.com" {
		t.Fatalf("SaveDraft() new handle reservation got %q want %q", got, "fan@example.com")
	}
}

func TestRedisSignUpDraftStoreValidatesInput(t *testing.T) {
	t.Parallel()

	store := &RedisSignUpDraftStore{}
	if _, err := store.GetDraft(context.Background(), "fan@example.com"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("GetDraft() error got %v want initialization error", err)
	}

	store = &RedisSignUpDraftStore{client: &signUpDraftRedisClientStub{values: map[string]string{}}}
	if err := store.SaveDraft(context.Background(), " ", SignUpDraft{Handle: "mina"}, time.Hour); err == nil || !strings.Contains(err.Error(), "email is required") {
		t.Fatalf("SaveDraft() error got %v want email validation error", err)
	}
	if err := store.SaveDraft(context.Background(), "fan@example.com", SignUpDraft{}, time.Hour); err == nil || !strings.Contains(err.Error(), "handle is required") {
		t.Fatalf("SaveDraft() error got %v want handle validation error", err)
	}
	if err := store.SaveDraft(context.Background(), "fan@example.com", SignUpDraft{Handle: "mina"}, 0); err == nil || !strings.Contains(err.Error(), "greater than zero") {
		t.Fatalf("SaveDraft() error got %v want ttl validation error", err)
	}
}

func stringifyRedisValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}
