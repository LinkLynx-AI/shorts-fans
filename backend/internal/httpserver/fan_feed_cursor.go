package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

const (
	defaultFanFeedCursorTTL = 24 * time.Hour
	fanFeedCursorKeyPrefix  = "fan_feed_cursor:"
	fanFeedCursorVersion    = 1
)

var errFanFeedCursorInvalid = errors.New("fan feed cursor is invalid")
var errFanFeedCursorNotFound = errors.New("fan feed cursor was not found")

type fanFeedCursorState struct {
	FollowingRemainingShortIDs []uuid.UUID `json:"followingRemainingShortIds,omitempty"`
	OwnerBinding               string      `json:"ownerBinding"`
	PublishedAt                *time.Time  `json:"publishedAt,omitempty"`
	ShortID                    *uuid.UUID  `json:"shortId,omitempty"`
	Tab                        string      `json:"tab"`
	Version                    int         `json:"version"`
}

type fanFeedCursorCodec struct {
	store fanFeedCursorStore
	ttl   time.Duration
}

type fanFeedCursorStore interface {
	Load(ctx context.Context, token string) ([]byte, error)
	Save(ctx context.Context, token string, payload []byte, ttl time.Duration) error
}

type fanFeedCursorRedisClient interface {
	Get(ctx context.Context, key string) *goredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
}

type redisFanFeedCursorStore struct {
	client fanFeedCursorRedisClient
}

type memoryFanFeedCursorStore struct {
	mu     sync.Mutex
	tokens map[string]memoryFanFeedCursorEntry
}

type memoryFanFeedCursorEntry struct {
	expiresAt time.Time
	payload   []byte
}

// NewRedisFanFeedCursorCodec は Redis-backed な fan feed cursor codec を構築します。
func NewRedisFanFeedCursorCodec(client *goredis.Client) FanFeedCursorCodec {
	if client == nil {
		return newMemoryFanFeedCursorCodec()
	}

	return &fanFeedCursorCodec{
		store: redisFanFeedCursorStore{client: client},
		ttl:   defaultFanFeedCursorTTL,
	}
}

func newMemoryFanFeedCursorCodec() FanFeedCursorCodec {
	return &fanFeedCursorCodec{
		store: &memoryFanFeedCursorStore{
			tokens: make(map[string]memoryFanFeedCursorEntry),
		},
		ttl: defaultFanFeedCursorTTL,
	}
}

func (c *fanFeedCursorCodec) Decode(ctx context.Context, tab string, ownerBinding string, encoded string) (*feed.Cursor, error) {
	if strings.TrimSpace(encoded) == "" {
		return nil, nil
	}
	if c == nil || c.store == nil {
		return nil, errFanFeedCursorInvalid
	}

	payload, err := c.store.Load(ctx, strings.TrimSpace(encoded))
	if err != nil {
		if errors.Is(err, errFanFeedCursorNotFound) {
			return nil, errFanFeedCursorInvalid
		}

		return nil, err
	}

	var state fanFeedCursorState
	if err := json.Unmarshal(payload, &state); err != nil {
		return nil, errFanFeedCursorInvalid
	}
	if err := validateFanFeedCursorState(tab, ownerBinding, state); err != nil {
		return nil, err
	}

	cursor := &feed.Cursor{}
	switch tab {
	case "recommended":
		cursor.PublishedAt = state.PublishedAt.UTC()
		cursor.ShortID = *state.ShortID
	case "following":
		cursor.FollowingRemainingShortIDs = append([]uuid.UUID(nil), state.FollowingRemainingShortIDs...)
	default:
		return nil, errFanFeedCursorInvalid
	}

	return cursor, nil
}

func (c *fanFeedCursorCodec) Encode(ctx context.Context, tab string, ownerBinding string, cursor *feed.Cursor) (*string, error) {
	if cursor == nil {
		return nil, nil
	}
	if c == nil || c.store == nil {
		return nil, errFanFeedCursorInvalid
	}

	state, err := buildFanFeedCursorState(tab, ownerBinding, cursor)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("marshal fan feed cursor: %w", err)
	}

	token := uuid.NewString()
	if err := c.store.Save(ctx, token, payload, c.ttl); err != nil {
		return nil, fmt.Errorf("save fan feed cursor token=%s: %w", token, err)
	}

	return &token, nil
}

func buildFanFeedCursorState(tab string, ownerBinding string, cursor *feed.Cursor) (fanFeedCursorState, error) {
	state := fanFeedCursorState{
		OwnerBinding: ownerBinding,
		Tab:          tab,
		Version:      fanFeedCursorVersion,
	}

	switch tab {
	case "recommended":
		if len(cursor.FollowingRemainingShortIDs) > 0 || cursor.PublishedAt.IsZero() || cursor.ShortID == uuid.Nil {
			return fanFeedCursorState{}, errFanFeedCursorInvalid
		}
		publishedAt := cursor.PublishedAt.UTC()
		shortID := cursor.ShortID
		state.PublishedAt = &publishedAt
		state.ShortID = &shortID
	case "following":
		if err := validateFollowingFeedCursor(cursor); err != nil {
			return fanFeedCursorState{}, errFanFeedCursorInvalid
		}
		state.FollowingRemainingShortIDs = append([]uuid.UUID(nil), cursor.FollowingRemainingShortIDs...)
	default:
		return fanFeedCursorState{}, errFanFeedCursorInvalid
	}

	if err := validateFanFeedCursorState(tab, ownerBinding, state); err != nil {
		return fanFeedCursorState{}, err
	}

	return state, nil
}

func validateFanFeedCursorState(tab string, ownerBinding string, state fanFeedCursorState) error {
	if strings.TrimSpace(ownerBinding) == "" {
		return errFanFeedCursorInvalid
	}
	if state.Version != fanFeedCursorVersion {
		return errFanFeedCursorInvalid
	}
	if strings.TrimSpace(state.OwnerBinding) != ownerBinding {
		return errFanFeedCursorInvalid
	}
	if strings.TrimSpace(state.Tab) != tab {
		return errFanFeedCursorInvalid
	}

	switch tab {
	case "recommended":
		if state.PublishedAt == nil || state.PublishedAt.IsZero() || state.ShortID == nil || *state.ShortID == uuid.Nil {
			return errFanFeedCursorInvalid
		}
		if len(state.FollowingRemainingShortIDs) > 0 {
			return errFanFeedCursorInvalid
		}
	case "following":
		if len(state.FollowingRemainingShortIDs) == 0 {
			return errFanFeedCursorInvalid
		}
		if state.PublishedAt != nil || state.ShortID != nil {
			return errFanFeedCursorInvalid
		}
		for _, shortID := range state.FollowingRemainingShortIDs {
			if shortID == uuid.Nil {
				return errFanFeedCursorInvalid
			}
		}
	default:
		return errFanFeedCursorInvalid
	}

	return nil
}

func validateFollowingFeedCursor(cursor *feed.Cursor) error {
	if cursor == nil {
		return errFanFeedCursorInvalid
	}
	if len(cursor.FollowingRemainingShortIDs) == 0 {
		return errFanFeedCursorInvalid
	}
	if !cursor.PublishedAt.IsZero() || cursor.ShortID != uuid.Nil {
		return errFanFeedCursorInvalid
	}
	for _, shortID := range cursor.FollowingRemainingShortIDs {
		if shortID == uuid.Nil {
			return errFanFeedCursorInvalid
		}
	}

	return nil
}

func (s redisFanFeedCursorStore) Load(ctx context.Context, token string) ([]byte, error) {
	if s.client == nil {
		return nil, errFanFeedCursorNotFound
	}

	payload, err := s.client.Get(ctx, fanFeedCursorRedisKey(token)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, errFanFeedCursorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get fan feed cursor token=%s: %w", token, err)
	}

	return payload, nil
}

func (s redisFanFeedCursorStore) Save(ctx context.Context, token string, payload []byte, ttl time.Duration) error {
	if s.client == nil {
		return errFanFeedCursorNotFound
	}

	if err := s.client.Set(ctx, fanFeedCursorRedisKey(token), payload, ttl).Err(); err != nil {
		return fmt.Errorf("set fan feed cursor token=%s: %w", token, err)
	}

	return nil
}

func (s *memoryFanFeedCursorStore) Load(_ context.Context, token string) ([]byte, error) {
	if s == nil {
		return nil, errFanFeedCursorNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.tokens[token]
	if !ok {
		return nil, errFanFeedCursorNotFound
	}
	if time.Now().After(entry.expiresAt) {
		delete(s.tokens, token)
		return nil, errFanFeedCursorNotFound
	}

	return append([]byte(nil), entry.payload...), nil
}

func (s *memoryFanFeedCursorStore) Save(_ context.Context, token string, payload []byte, ttl time.Duration) error {
	if s == nil {
		return errFanFeedCursorNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[token] = memoryFanFeedCursorEntry{
		expiresAt: time.Now().Add(ttl),
		payload:   append([]byte(nil), payload...),
	}

	return nil
}

func fanFeedCursorRedisKey(token string) string {
	return fanFeedCursorKeyPrefix + strings.TrimSpace(token)
}
