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

type stubCreatorProfileReader struct {
	getHeader func(context.Context, string) (creator.PublicProfileHeader, error)
}

func (s stubCreatorProfileReader) GetPublicProfileHeader(ctx context.Context, creatorID string) (creator.PublicProfileHeader, error) {
	return s.getHeader(ctx, creatorID)
}

type stubCreatorProfileShortsReader struct {
	listShorts func(context.Context, string, *creator.PublicProfileShortCursor, int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error)
}

func (s stubCreatorProfileShortsReader) ListPublicProfileShorts(ctx context.Context, creatorID string, cursor *creator.PublicProfileShortCursor, limit int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error) {
	return s.listShorts(ctx, creatorID, cursor, limit)
}

func TestCreatorProfileRoute(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		CreatorProfile: stubCreatorProfileReader{
			getHeader: func(_ context.Context, gotCreatorID string) (creator.PublicProfileHeader, error) {
				if gotCreatorID != creator.FormatPublicID(creatorID) {
					t.Fatalf("GetPublicProfileHeader() creatorID got %q want %q", gotCreatorID, creator.FormatPublicID(creatorID))
				}

				return creator.PublicProfileHeader{
					Profile: creator.Profile{
						UserID:      creatorID,
						DisplayName: stringPtr("Mina Rei"),
						Handle:      stringPtr("minarei"),
						AvatarURL:   stringPtr("https://cdn.example.com/mina.jpg"),
						Bio:         "quiet rooftop と hotel light の preview を軸に投稿。",
						PublishedAt: timePtr(now),
					},
					ShortCount:  2,
					FanCount:    24,
					IsFollowing: false,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/"+creator.FormatPublicID(creatorID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/{creatorId} status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorProfileResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Profile.Creator.ID != creator.FormatPublicID(creatorID) {
		t.Fatalf("response.Data.Profile.Creator.ID got %#v want %q", response.Data, creator.FormatPublicID(creatorID))
	}
	if response.Data.Profile.Stats.ShortCount != 2 {
		t.Fatalf("response.Data.Profile.Stats.ShortCount got %d want %d", response.Data.Profile.Stats.ShortCount, 2)
	}
	if response.Meta.Page != nil {
		t.Fatalf("response.Meta.Page got %#v want nil", response.Meta.Page)
	}
	body := rec.Body.String()
	if strings.Contains(body, `"items"`) || strings.Contains(body, `"viewCount"`) {
		t.Fatalf("GET /api/fan/creators/{creatorId} body got %q want no items/viewCount", body)
	}
}

func TestCreatorProfileNotFoundRoute(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorProfile: stubCreatorProfileReader{
			getHeader: func(context.Context, string) (creator.PublicProfileHeader, error) {
				return creator.PublicProfileHeader{}, creator.ErrProfileNotFound
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/creator_missing", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/fan/creators/{creatorId} status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("GET /api/fan/creators/{creatorId} body got %q want not_found", rec.Body.String())
	}
}

func TestCreatorProfileShortsRoute(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	nextShortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mainID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	mediaAssetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	expectedCursor := &creator.PublicProfileShortCursor{
		PublishedAt: now.Add(-time.Hour),
		ShortID:     nextShortID,
	}

	router := NewHandler(HandlerConfig{
		CreatorProfileShorts: stubCreatorProfileShortsReader{
			listShorts: func(_ context.Context, gotCreatorID string, gotCursor *creator.PublicProfileShortCursor, limit int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error) {
				if gotCreatorID != creator.FormatPublicID(creatorID) {
					t.Fatalf("ListPublicProfileShorts() creatorID got %q want %q", gotCreatorID, creator.FormatPublicID(creatorID))
				}
				if gotCursor != nil {
					t.Fatalf("ListPublicProfileShorts() cursor got %#v want nil", gotCursor)
				}
				if limit != creator.DefaultPublicProfileShortGridPageSize {
					t.Fatalf("ListPublicProfileShorts() limit got %d want %d", limit, creator.DefaultPublicProfileShortGridPageSize)
				}

				return []creator.PublicProfileShort{
					{
						ID:                     shortID,
						CreatorUserID:          creatorID,
						CanonicalMainID:        mainID,
						MediaAssetID:           mediaAssetID,
						MediaURL:               "https://cdn.example.com/short-a.mp4",
						PreviewDurationSeconds: 16,
						PublishedAt:            now,
					},
				}, expectedCursor, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/"+creator.FormatPublicID(creatorID)+"/shorts", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorProfileShortGridResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || len(response.Data.Items) != 1 {
		t.Fatalf("response.Data.Items got %#v want len 1", response.Data)
	}
	if response.Data.Items[0].ID != shortPublicID(shortID) {
		t.Fatalf("response.Data.Items[0].ID got %q want %q", response.Data.Items[0].ID, shortPublicID(shortID))
	}
	if response.Meta.Page == nil || !response.Meta.Page.HasNext || response.Meta.Page.NextCursor == nil {
		t.Fatalf("response.Meta.Page got %#v want next cursor", response.Meta.Page)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(*response.Meta.Page.NextCursor)
	if err != nil {
		t.Fatalf("DecodeString(next cursor) error = %v, want nil", err)
	}
	if !strings.Contains(string(decoded), nextShortID.String()) {
		t.Fatalf("decoded next cursor got %q want %q", string(decoded), nextShortID.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, `"profile"`) || strings.Contains(body, `"title"`) || strings.Contains(body, `"caption"`) {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts body got %q want no profile/title/caption", body)
	}
}

func TestCreatorProfileShortsEmptyRoute(t *testing.T) {
	t.Parallel()

	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		CreatorProfileShorts: stubCreatorProfileShortsReader{
			listShorts: func(_ context.Context, gotCreatorID string, gotCursor *creator.PublicProfileShortCursor, limit int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error) {
				if gotCreatorID != creator.FormatPublicID(creatorID) {
					t.Fatalf("ListPublicProfileShorts() creatorID got %q want %q", gotCreatorID, creator.FormatPublicID(creatorID))
				}
				if gotCursor != nil {
					t.Fatalf("ListPublicProfileShorts() cursor got %#v want nil", gotCursor)
				}
				if limit != creator.DefaultPublicProfileShortGridPageSize {
					t.Fatalf("ListPublicProfileShorts() limit got %d want %d", limit, creator.DefaultPublicProfileShortGridPageSize)
				}
				return []creator.PublicProfileShort{}, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/"+creator.FormatPublicID(creatorID)+"/shorts", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorProfileShortGridResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || len(response.Data.Items) != 0 {
		t.Fatalf("response.Data.Items got %#v want empty", response.Data)
	}
	if response.Meta.Page == nil || response.Meta.Page.HasNext || response.Meta.Page.NextCursor != nil {
		t.Fatalf("response.Meta.Page got %#v want empty page info", response.Meta.Page)
	}
	body := rec.Body.String()
	if strings.Contains(body, `"profile"`) || strings.Contains(body, `"title"`) || strings.Contains(body, `"caption"`) {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts body got %q want no profile/title/caption", body)
	}
}

func TestCreatorProfileShortsMalformedCursorFallsBackToFirstPage(t *testing.T) {
	t.Parallel()

	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		CreatorProfileShorts: stubCreatorProfileShortsReader{
			listShorts: func(_ context.Context, gotCreatorID string, gotCursor *creator.PublicProfileShortCursor, limit int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error) {
				if gotCreatorID != creator.FormatPublicID(creatorID) {
					t.Fatalf("ListPublicProfileShorts() creatorID got %q want %q", gotCreatorID, creator.FormatPublicID(creatorID))
				}
				if gotCursor != nil {
					t.Fatalf("ListPublicProfileShorts() cursor got %#v want nil", gotCursor)
				}
				if limit != creator.DefaultPublicProfileShortGridPageSize {
					t.Fatalf("ListPublicProfileShorts() limit got %d want %d", limit, creator.DefaultPublicProfileShortGridPageSize)
				}
				return []creator.PublicProfileShort{}, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/"+creator.FormatPublicID(creatorID)+"/shorts?cursor=***", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts?cursor=*** status got %d want %d", rec.Code, http.StatusOK)
	}
}

func TestCreatorProfileShortsNotFoundRoute(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorProfileShorts: stubCreatorProfileShortsReader{
			listShorts: func(context.Context, string, *creator.PublicProfileShortCursor, int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error) {
				return nil, nil, creator.ErrProfileNotFound
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/creator_missing/shorts", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts body got %q want not_found", rec.Body.String())
	}
}
