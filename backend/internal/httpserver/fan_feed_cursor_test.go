package httpserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type fanFeedCursorRedisClientStub struct {
	values map[string]string
	getErr error
	setErr error
}

func (s *fanFeedCursorRedisClientStub) Get(_ context.Context, key string) *goredis.StringCmd {
	if s.getErr != nil {
		return goredis.NewStringResult("", s.getErr)
	}
	value, ok := s.values[key]
	if !ok {
		return goredis.NewStringResult("", goredis.Nil)
	}

	return goredis.NewStringResult(value, nil)
}

func (s *fanFeedCursorRedisClientStub) Set(_ context.Context, key string, value interface{}, _ time.Duration) *goredis.StatusCmd {
	if s.setErr != nil {
		return goredis.NewStatusResult("", s.setErr)
	}

	switch typed := value.(type) {
	case []byte:
		s.values[key] = string(typed)
	default:
		payload, _ := json.Marshal(typed)
		s.values[key] = string(payload)
	}

	return goredis.NewStatusResult("OK", nil)
}

func TestMemoryFanFeedCursorCodecFollowingRoundTrip(t *testing.T) {
	t.Parallel()

	codec := newMemoryFanFeedCursorCodec()
	cursor := &feed.Cursor{
		FollowingRemainingShortIDs: []uuid.UUID{
			uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		},
	}
	ownerBinding := "viewer:11111111-1111-1111-1111-111111111111"

	encoded, err := codec.Encode(context.Background(), "following", ownerBinding, cursor)
	if err != nil {
		t.Fatalf("Encode() error = %v, want nil", err)
	}
	if encoded == nil {
		t.Fatal("Encode() = nil, want token")
	}
	if strings.Contains(*encoded, cursor.FollowingRemainingShortIDs[0].String()) {
		t.Fatalf("Encode() token got %q want opaque token without raw cursor fields", *encoded)
	}

	decoded, err := codec.Decode(context.Background(), "following", ownerBinding, *encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v, want nil", err)
	}
	if decoded == nil {
		t.Fatal("Decode() = nil, want cursor")
	}
	if len(decoded.FollowingRemainingShortIDs) != len(cursor.FollowingRemainingShortIDs) {
		t.Fatalf(
			"Decode() remaining short ids len got %d want %d",
			len(decoded.FollowingRemainingShortIDs),
			len(cursor.FollowingRemainingShortIDs),
		)
	}
	for index, wantShortID := range cursor.FollowingRemainingShortIDs {
		if decoded.FollowingRemainingShortIDs[index] != wantShortID {
			t.Fatalf(
				"Decode() remaining short ids[%d] got %s want %s",
				index,
				decoded.FollowingRemainingShortIDs[index],
				wantShortID,
			)
		}
	}
	if !decoded.PublishedAt.IsZero() || decoded.ShortID != uuid.Nil {
		t.Fatalf("Decode() got %#v want %#v", decoded, cursor)
	}
}

func TestMemoryFanFeedCursorCodecRejectsUnknownOrForeignTabToken(t *testing.T) {
	t.Parallel()

	codec := newMemoryFanFeedCursorCodec()
	if _, err := codec.Decode(context.Background(), "following", "viewer:11111111-1111-1111-1111-111111111111", uuid.NewString()); err == nil {
		t.Fatal("Decode(unknown token) error = nil, want invalid cursor error")
	}

	recommendedCursor := &feed.Cursor{
		PublishedAt: time.Unix(1710000200, 0).UTC(),
		ShortID:     uuid.MustParse("33333333-3333-3333-3333-333333333333"),
	}
	encoded, err := codec.Encode(context.Background(), "recommended", "public", recommendedCursor)
	if err != nil {
		t.Fatalf("Encode(recommended) error = %v, want nil", err)
	}
	if _, err := codec.Decode(context.Background(), "following", "public", *encoded); err == nil {
		t.Fatal("Decode(foreign tab token) error = nil, want invalid cursor error")
	}
}

func TestRedisFanFeedCursorCodecRoundTrip(t *testing.T) {
	t.Parallel()

	client := &fanFeedCursorRedisClientStub{values: map[string]string{}}
	codec := &fanFeedCursorCodec{
		store: redisFanFeedCursorStore{client: client},
		ttl:   defaultFanFeedCursorTTL,
	}
	cursor := &feed.Cursor{
		PublishedAt: time.Unix(1710000200, 0).UTC(),
		ShortID:     uuid.MustParse("33333333-3333-3333-3333-333333333333"),
	}

	encoded, err := codec.Encode(context.Background(), "recommended", "public", cursor)
	if err != nil {
		t.Fatalf("Encode() error = %v, want nil", err)
	}
	if encoded == nil {
		t.Fatal("Encode() = nil, want token")
	}
	if _, ok := client.values[fanFeedCursorRedisKey(*encoded)]; !ok {
		t.Fatalf("redis cursor store missing key %q", fanFeedCursorRedisKey(*encoded))
	}

	decoded, err := codec.Decode(context.Background(), "recommended", "public", *encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v, want nil", err)
	}
	if decoded == nil || !decoded.PublishedAt.Equal(cursor.PublishedAt) || decoded.ShortID != cursor.ShortID {
		t.Fatalf("Decode() got %#v want %#v", decoded, cursor)
	}
}
