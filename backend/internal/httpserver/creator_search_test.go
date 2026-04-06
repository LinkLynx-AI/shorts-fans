package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/google/uuid"
)

type stubCreatorSearchReader struct {
	listRecent func(context.Context, *creator.PublicProfileCursor, int) ([]creator.Profile, *creator.PublicProfileCursor, error)
	search     func(context.Context, string, *creator.PublicProfileCursor, int) ([]creator.Profile, *creator.PublicProfileCursor, error)
}

func (s stubCreatorSearchReader) ListRecentPublicProfiles(ctx context.Context, cursor *creator.PublicProfileCursor, limit int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
	return s.listRecent(ctx, cursor, limit)
}

func (s stubCreatorSearchReader) SearchPublicProfiles(ctx context.Context, query string, cursor *creator.PublicProfileCursor, limit int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
	return s.search(ctx, query, cursor, limit)
}

func TestCreatorSearchRecentRoute(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	router := NewHandler(HandlerConfig{
		CreatorSearch: stubCreatorSearchReader{
			listRecent: func(context.Context, *creator.PublicProfileCursor, int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
				return []creator.Profile{
					{
						UserID:      creatorID,
						DisplayName: stringPtr("Aoi N"),
						Handle:      stringPtr("aoina"),
						AvatarURL:   stringPtr("https://cdn.example.com/creator/aoi/avatar.jpg"),
						Bio:         "soft light と close framing の short を中心に更新中。",
						PublishedAt: timePtr(now),
					},
				}, nil, nil
			},
			search: func(context.Context, string, *creator.PublicProfileCursor, int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
				t.Fatal("SearchPublicProfiles() was called for empty query")
				return nil, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/search", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/search status got %d want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"query":""`) {
		t.Fatalf("GET /api/fan/creators/search body got %q want empty query", body)
	}
	if !strings.Contains(body, `"handle":"@aoina"`) {
		t.Fatalf("GET /api/fan/creators/search body got %q want @aoina", body)
	}
	if !strings.Contains(body, `"id":"creator_11111111111111111111111111111111"`) {
		t.Fatalf("GET /api/fan/creators/search body got %q want stable creator id", body)
	}
	if !strings.Contains(body, `"id":"asset_creator_11111111111111111111111111111111_avatar"`) {
		t.Fatalf("GET /api/fan/creators/search body got %q want stable avatar id", body)
	}
}

func TestCreatorSearchFilteredRoute(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	cursor := &creator.PublicProfileCursor{
		PublishedAt: now.Add(-time.Hour),
		Handle:      "minarei",
	}
	router := NewHandler(HandlerConfig{
		CreatorSearch: stubCreatorSearchReader{
			listRecent: func(context.Context, *creator.PublicProfileCursor, int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
				t.Fatal("ListRecentPublicProfiles() was called for non-empty query")
				return nil, nil, nil
			},
			search: func(_ context.Context, query string, gotCursor *creator.PublicProfileCursor, limit int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
				if query != "@mina" {
					t.Fatalf("SearchPublicProfiles() query got %q want %q", query, "@mina")
				}
				if limit != creatorSearchPageSize {
					t.Fatalf("SearchPublicProfiles() limit got %d want %d", limit, creatorSearchPageSize)
				}
				if gotCursor == nil || gotCursor.Handle != "aoina" {
					t.Fatalf("SearchPublicProfiles() cursor got %#v want handle aoina", gotCursor)
				}
				return []creator.Profile{
					{
						UserID:      creatorID,
						DisplayName: stringPtr("Mina Rei"),
						Handle:      stringPtr("minarei"),
						AvatarURL:   stringPtr("https://cdn.example.com/creator/mina/avatar.jpg"),
						Bio:         "quiet rooftop と hotel light の preview を軸に投稿。",
						PublishedAt: timePtr(now),
					},
				}, cursor, nil
			},
		},
	})

	requestCursor := encodeCreatorSearchCursor(&creator.PublicProfileCursor{
		PublishedAt: now,
		Handle:      "aoina",
	})
	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/search?q=%40mina&cursor="+*requestCursor, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/search?q=@mina status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorSearchResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Query != "@mina" {
		t.Fatalf("response.Data.Query got %#v want %q", response.Data, "@mina")
	}
	if len(response.Data.Items) != 1 {
		t.Fatalf("response.Data.Items len got %d want %d", len(response.Data.Items), 1)
	}
	if response.Data.Items[0].Creator.ID != "creator_22222222222222222222222222222222" {
		t.Fatalf("response.Data.Items[0].Creator.ID got %q want %q", response.Data.Items[0].Creator.ID, "creator_22222222222222222222222222222222")
	}
	if response.Data.Items[0].Creator.Avatar == nil {
		t.Fatal("response.Data.Items[0].Creator.Avatar = nil, want non-nil")
	}
	if response.Data.Items[0].Creator.Avatar.ID != "asset_creator_22222222222222222222222222222222_avatar" {
		t.Fatalf("response.Data.Items[0].Creator.Avatar.ID got %q want %q", response.Data.Items[0].Creator.Avatar.ID, "asset_creator_22222222222222222222222222222222_avatar")
	}
	if response.Meta.Page == nil || !response.Meta.Page.HasNext || response.Meta.Page.NextCursor == nil {
		t.Fatalf("response.Meta.Page got %#v want next cursor", response.Meta.Page)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(*response.Meta.Page.NextCursor)
	if err != nil {
		t.Fatalf("DecodeString(next cursor) error = %v, want nil", err)
	}
	if !strings.Contains(string(decoded), `"handle":"minarei"`) {
		t.Fatalf("decoded next cursor got %q want minarei", string(decoded))
	}
}

func TestCreatorSearchEmptyFilteredRoute(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorSearch: stubCreatorSearchReader{
			listRecent: func(context.Context, *creator.PublicProfileCursor, int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
				t.Fatal("ListRecentPublicProfiles() was called for filtered empty result")
				return nil, nil, nil
			},
			search: func(_ context.Context, query string, gotCursor *creator.PublicProfileCursor, limit int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
				if query != "missing" {
					t.Fatalf("SearchPublicProfiles() query got %q want %q", query, "missing")
				}
				if gotCursor != nil {
					t.Fatalf("SearchPublicProfiles() cursor got %#v want nil", gotCursor)
				}
				if limit != creatorSearchPageSize {
					t.Fatalf("SearchPublicProfiles() limit got %d want %d", limit, creatorSearchPageSize)
				}

				return []creator.Profile{}, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/search?q=missing", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/search?q=missing status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorSearchResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Query != "missing" {
		t.Fatalf("response.Data.Query got %#v want %q", response.Data, "missing")
	}
	if len(response.Data.Items) != 0 {
		t.Fatalf("response.Data.Items len got %d want %d", len(response.Data.Items), 0)
	}
	if response.Meta.Page == nil || response.Meta.Page.HasNext || response.Meta.Page.NextCursor != nil {
		t.Fatalf("response.Meta.Page got %#v want empty page info", response.Meta.Page)
	}
}

func TestCreatorSearchMalformedCursorFallsBackToFirstPage(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorSearch: stubCreatorSearchReader{
			listRecent: func(_ context.Context, gotCursor *creator.PublicProfileCursor, limit int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
				if gotCursor != nil {
					t.Fatalf("ListRecentPublicProfiles() cursor got %#v want nil", gotCursor)
				}
				if limit != creatorSearchPageSize {
					t.Fatalf("ListRecentPublicProfiles() limit got %d want %d", limit, creatorSearchPageSize)
				}

				return []creator.Profile{}, nil, nil
			},
			search: func(context.Context, string, *creator.PublicProfileCursor, int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
				t.Fatal("SearchPublicProfiles() was called for malformed recent cursor")
				return nil, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/search?cursor=***", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/search?cursor=*** status got %d want %d", rec.Code, http.StatusOK)
	}
}

func TestCreatorSearchAllowsMissingAvatar(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorSearch: stubCreatorSearchReader{
			listRecent: func(context.Context, *creator.PublicProfileCursor, int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
				return []creator.Profile{
					{
						UserID:      uuid.MustParse("33333333-3333-3333-3333-333333333333"),
						DisplayName: stringPtr("Aoi N"),
						Handle:      stringPtr("aoina"),
						Bio:         "missing avatar should return null avatar",
					},
				}, nil, nil
			},
			search: func(context.Context, string, *creator.PublicProfileCursor, int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
				t.Fatal("SearchPublicProfiles() was called for missing-avatar recent route")
				return nil, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/search", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/search status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorSearchResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error != nil {
		t.Fatalf("response.Error got %#v want nil", response.Error)
	}
	if response.Data == nil || len(response.Data.Items) != 1 {
		t.Fatalf("response.Data got %#v want one item", response.Data)
	}
	if response.Data.Items[0].Creator.Avatar != nil {
		t.Fatalf("response.Data.Items[0].Creator.Avatar got %#v want nil", response.Data.Items[0].Creator.Avatar)
	}
}

func stringPtr(value string) *string {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
