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
	router := NewHandler(HandlerConfig{
		CreatorSearch: stubCreatorSearchReader{
			listRecent: func(context.Context, *creator.PublicProfileCursor, int) ([]creator.Profile, *creator.PublicProfileCursor, error) {
				return []creator.Profile{
					{
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
	if !strings.Contains(body, `"id":"aoina"`) {
		t.Fatalf("GET /api/fan/creators/search body got %q want id aoina", body)
	}
}

func TestCreatorSearchFilteredRoute(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
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

func stringPtr(value string) *string {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
