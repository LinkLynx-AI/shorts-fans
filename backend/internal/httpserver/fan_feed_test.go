package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/google/uuid"
)

type stubFanFeedReader struct {
	getDetail       func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error)
	listFollowing   func(context.Context, uuid.UUID, *feed.Cursor, int) ([]feed.Item, *feed.Cursor, error)
	listRecommended func(context.Context, *uuid.UUID, *feed.Cursor, int) ([]feed.Item, *feed.Cursor, error)
}

func (s stubFanFeedReader) GetDetail(ctx context.Context, shortID uuid.UUID, viewerUserID *uuid.UUID) (feed.Detail, error) {
	return s.getDetail(ctx, shortID, viewerUserID)
}

func (s stubFanFeedReader) ListFollowing(ctx context.Context, viewerUserID uuid.UUID, cursor *feed.Cursor, limit int) ([]feed.Item, *feed.Cursor, error) {
	return s.listFollowing(ctx, viewerUserID, cursor, limit)
}

func (s stubFanFeedReader) ListRecommended(ctx context.Context, viewerUserID *uuid.UUID, cursor *feed.Cursor, limit int) ([]feed.Item, *feed.Cursor, error) {
	return s.listRecommended(ctx, viewerUserID, cursor, limit)
}

type stubShortDisplayAssetResolver struct {
	resolve func(source media.ShortDisplaySource, boundary media.AccessBoundary) (media.VideoDisplayAsset, error)
}

func (s stubShortDisplayAssetResolver) ResolveShortDisplayAsset(source media.ShortDisplaySource, boundary media.AccessBoundary) (media.VideoDisplayAsset, error) {
	return s.resolve(source, boundary)
}

func TestFanFeedRecommendedRoute(t *testing.T) {
	t.Parallel()

	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mediaAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	publishedAt := time.Unix(1710000000, 0).UTC()

	router := NewHandler(HandlerConfig{
		FanFeed: stubFanFeedReader{
			listRecommended: func(_ context.Context, viewerUserID *uuid.UUID, cursor *feed.Cursor, limit int) ([]feed.Item, *feed.Cursor, error) {
				if viewerUserID != nil {
					t.Fatalf("ListRecommended() viewerUserID got %v want nil", viewerUserID)
				}
				if cursor != nil {
					t.Fatalf("ListRecommended() cursor got %#v want nil", cursor)
				}
				if limit != feed.DefaultPageSize {
					t.Fatalf("ListRecommended() limit got %d want %d", limit, feed.DefaultPageSize)
				}

				item := feed.Item{
					Creator: feed.CreatorSummary{
						Bio:         "quiet rooftop と hotel light の preview を軸に投稿。",
						DisplayName: "Mina Rei",
						Handle:      "minarei",
						ID:          creatorID,
					},
					Short: feed.ShortSummary{
						Caption:                "quiet rooftop preview",
						CanonicalMainID:        mainID,
						CreatorUserID:          creatorID,
						ID:                     shortID,
						MediaAssetID:           mediaAssetID,
						PreviewDurationSeconds: 16,
						PublishedAt:            publishedAt,
					},
					Unlock: feed.UnlockPreview{
						IsOwner:             false,
						IsUnlocked:          false,
						MainDurationSeconds: 480,
						PriceJPY:            1800,
					},
				}
				item.Viewer.IsPinned = true

				return []feed.Item{item}, nil, nil
			},
		},
		ShortDisplayAssets: stubShortDisplayAssetResolver{
			resolve: func(source media.ShortDisplaySource, boundary media.AccessBoundary) (media.VideoDisplayAsset, error) {
				if source.AssetID != mediaAssetID {
					t.Fatalf("ResolveShortDisplayAsset() assetID got %s want %s", source.AssetID, mediaAssetID)
				}
				if source.ShortID != shortID {
					t.Fatalf("ResolveShortDisplayAsset() shortID got %s want %s", source.ShortID, shortID)
				}
				if boundary != media.AccessBoundaryPublic {
					t.Fatalf("ResolveShortDisplayAsset() boundary got %s want %s", boundary, media.AccessBoundaryPublic)
				}

				return media.VideoDisplayAsset{
					DurationSeconds: 16,
					ID:              mediaAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/shorts/poster.jpg",
					URL:             "https://cdn.example.com/shorts/playback.mp4",
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/feed?tab=recommended", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/feed?tab=recommended status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanFeedResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Tab != "recommended" {
		t.Fatalf("response.Data got %#v want tab recommended", response.Data)
	}
	if len(response.Data.Items) != 1 {
		t.Fatalf("response.Data.Items len got %d want %d", len(response.Data.Items), 1)
	}
	if response.Data.Items[0].Short.ID != shorts.FormatPublicShortID(shortID) {
		t.Fatalf("response.Data.Items[0].Short.ID got %q want %q", response.Data.Items[0].Short.ID, shorts.FormatPublicShortID(shortID))
	}
	if response.Data.Items[0].UnlockCta.State != "unlock_available" {
		t.Fatalf("response.Data.Items[0].UnlockCta.State got %q want %q", response.Data.Items[0].UnlockCta.State, "unlock_available")
	}
	if !strings.Contains(rec.Body.String(), `"caption":"quiet rooftop preview"`) {
		t.Fatalf("response body got %q want caption", rec.Body.String())
	}
}

func TestFanFeedFollowingRouteRequiresAuth(t *testing.T) {
	t.Parallel()

	readerCalled := false
	router := NewHandler(HandlerConfig{
		FanFeed: stubFanFeedReader{
			listFollowing: func(context.Context, uuid.UUID, *feed.Cursor, int) ([]feed.Item, *feed.Cursor, error) {
				readerCalled = true
				return nil, nil, nil
			},
		},
		ShortDisplayAssets: stubShortDisplayAssetResolver{
			resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
				t.Fatal("ResolveShortDisplayAsset() was called unexpectedly")
				return media.VideoDisplayAsset{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/feed?tab=following", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/fan/feed?tab=following status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if readerCalled {
		t.Fatal("GET /api/fan/feed?tab=following readerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"auth_required"`) {
		t.Fatalf("GET /api/fan/feed?tab=following body got %q want auth_required", rec.Body.String())
	}
}

func TestFanFeedRouteIsNotRegisteredWithoutShortDisplayAssets(t *testing.T) {
	t.Parallel()

	readerCalled := false
	router := NewHandler(HandlerConfig{
		FanFeed: stubFanFeedReader{
			listRecommended: func(context.Context, *uuid.UUID, *feed.Cursor, int) ([]feed.Item, *feed.Cursor, error) {
				readerCalled = true
				return nil, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/feed?tab=recommended", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/fan/feed?tab=recommended status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if readerCalled {
		t.Fatal("GET /api/fan/feed?tab=recommended readerCalled = true, want false")
	}
}

func TestFanShortDetailRoute(t *testing.T) {
	t.Parallel()

	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mediaAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	publishedAt := time.Unix(1710000000, 0).UTC()

	router := NewHandler(HandlerConfig{
		FanFeed: stubFanFeedReader{
			getDetail: func(_ context.Context, gotShortID uuid.UUID, viewerUserID *uuid.UUID) (feed.Detail, error) {
				if gotShortID != shortID {
					t.Fatalf("GetDetail() shortID got %s want %s", gotShortID, shortID)
				}
				if viewerUserID != nil {
					t.Fatalf("GetDetail() viewerUserID got %v want nil", viewerUserID)
				}

				detail := feed.Detail{
					Item: feed.Item{
						Creator: feed.CreatorSummary{
							Bio:         "quiet rooftop と hotel light の preview を軸に投稿。",
							DisplayName: "Mina Rei",
							Handle:      "minarei",
							ID:          creatorID,
						},
						Short: feed.ShortSummary{
							Caption:                "quiet rooftop preview",
							CanonicalMainID:        mainID,
							CreatorUserID:          creatorID,
							ID:                     shortID,
							MediaAssetID:           mediaAssetID,
							PreviewDurationSeconds: 16,
							PublishedAt:            publishedAt,
						},
						Unlock: feed.UnlockPreview{
							IsOwner:             false,
							IsUnlocked:          true,
							MainDurationSeconds: 480,
							PriceJPY:            1800,
						},
					},
				}
				detail.Item.Viewer.IsPinned = true
				detail.Viewer.IsFollowingCreator = true

				return detail, nil
			},
		},
		ShortDisplayAssets: stubShortDisplayAssetResolver{
			resolve: func(source media.ShortDisplaySource, boundary media.AccessBoundary) (media.VideoDisplayAsset, error) {
				if source.ShortID != shortID {
					t.Fatalf("ResolveShortDisplayAsset() shortID got %s want %s", source.ShortID, shortID)
				}
				if boundary != media.AccessBoundaryPublic {
					t.Fatalf("ResolveShortDisplayAsset() boundary got %s want %s", boundary, media.AccessBoundaryPublic)
				}

				return media.VideoDisplayAsset{
					DurationSeconds: 16,
					ID:              mediaAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/shorts/poster.jpg",
					URL:             "https://cdn.example.com/shorts/playback.mp4",
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/shorts/"+shorts.FormatPublicShortID(shortID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/shorts/{shortId} status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[shortDetailResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Detail.Short.ID != shorts.FormatPublicShortID(shortID) {
		t.Fatalf("response.Data.Detail.Short.ID got %#v want %q", response.Data, shorts.FormatPublicShortID(shortID))
	}
	if !response.Data.Detail.Viewer.IsFollowingCreator {
		t.Fatalf("response.Data.Detail.Viewer.IsFollowingCreator got %#v want true", response.Data.Detail.Viewer)
	}
	if response.Data.Detail.UnlockCta.State != "continue_main" {
		t.Fatalf("response.Data.Detail.UnlockCta.State got %q want %q", response.Data.Detail.UnlockCta.State, "continue_main")
	}
}
