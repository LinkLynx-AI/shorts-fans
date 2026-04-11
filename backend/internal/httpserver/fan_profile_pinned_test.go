package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanprofile"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubFanProfilePinnedShortsReader struct {
	listPinnedShorts func(context.Context, uuid.UUID, *fanprofile.PinnedShortCursor, int) ([]fanprofile.PinnedShortItem, *fanprofile.PinnedShortCursor, error)
}

func (s stubFanProfilePinnedShortsReader) ListPinnedShorts(
	ctx context.Context,
	viewerID uuid.UUID,
	cursor *fanprofile.PinnedShortCursor,
	limit int,
) ([]fanprofile.PinnedShortItem, *fanprofile.PinnedShortCursor, error) {
	return s.listPinnedShorts(ctx, viewerID, cursor, limit)
}

func TestFanProfilePinnedShortsRoute(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	creatorID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mainID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	mediaAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	requestCursor := &fanprofile.PinnedShortCursor{
		PinnedAt: now.Add(-time.Hour),
		ShortID:  uuid.MustParse("66666666-6666-6666-6666-666666666666"),
	}
	nextCursor := &fanprofile.PinnedShortCursor{
		PinnedAt: now.Add(-2 * time.Hour),
		ShortID:  shortID,
	}

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         viewerID,
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		FanProfilePinnedShorts: stubFanProfilePinnedShortsReader{
			listPinnedShorts: func(_ context.Context, gotViewerID uuid.UUID, gotCursor *fanprofile.PinnedShortCursor, limit int) ([]fanprofile.PinnedShortItem, *fanprofile.PinnedShortCursor, error) {
				if gotViewerID != viewerID {
					t.Fatalf("ListPinnedShorts() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCursor == nil || gotCursor.ShortID != requestCursor.ShortID {
					t.Fatalf("ListPinnedShorts() cursor got %#v want %#v", gotCursor, requestCursor)
				}
				if limit != fanprofile.DefaultPinnedShortsPageSize {
					t.Fatalf("ListPinnedShorts() limit got %d want %d", limit, fanprofile.DefaultPinnedShortsPageSize)
				}

				return []fanprofile.PinnedShortItem{
					{
						CreatorAvatarURL:            stringPtr("https://cdn.example.com/creator/sora/avatar.jpg"),
						CreatorBio:                  "after rain と balcony mood の short をまとめています。",
						CreatorDisplayName:          "Sora Vale",
						CreatorHandle:               "soravale",
						CreatorUserID:               creatorID,
						PinnedAt:                    now,
						ShortCaption:                "after rain preview",
						ShortCanonicalMainID:        mainID,
						ShortID:                     shortID,
						ShortMediaAssetID:           mediaAssetID,
						ShortPreviewDurationSeconds: 17,
					},
				}, nextCursor, nil
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
					DurationSeconds: 17,
					ID:              mediaAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/shorts/poster.jpg",
					URL:             "https://cdn.example.com/shorts/playback.mp4",
				}, nil
			},
		},
	})

	encodedRequestCursor := encodeFanProfilePinnedShortsCursor(requestCursor)
	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/pinned-shorts?cursor="+*encodedRequestCursor, nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/profile/pinned-shorts status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanProfilePinnedShortsResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || len(response.Data.Items) != 1 {
		t.Fatalf("response.Data.Items got %#v want len 1", response.Data)
	}
	if response.Data.Items[0].Short.ID != "short_22222222222222222222222222222222" {
		t.Fatalf("response.Data.Items[0].Short.ID got %q want %q", response.Data.Items[0].Short.ID, "short_22222222222222222222222222222222")
	}
	if response.Data.Items[0].Short.Caption != "after rain preview" {
		t.Fatalf("response.Data.Items[0].Short.Caption got %q want %q", response.Data.Items[0].Short.Caption, "after rain preview")
	}
	if response.Meta.Page == nil || !response.Meta.Page.HasNext || response.Meta.Page.NextCursor == nil {
		t.Fatalf("response.Meta.Page got %#v want next cursor", response.Meta.Page)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(*response.Meta.Page.NextCursor)
	if err != nil {
		t.Fatalf("DecodeString(next cursor) error = %v, want nil", err)
	}
	if !strings.Contains(string(decoded), shortID.String()) {
		t.Fatalf("decoded next cursor got %q want short id %q", string(decoded), shortID.String())
	}
}

func TestFanProfilePinnedShortsEmptyRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         viewerID,
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		FanProfilePinnedShorts: stubFanProfilePinnedShortsReader{
			listPinnedShorts: func(_ context.Context, gotViewerID uuid.UUID, gotCursor *fanprofile.PinnedShortCursor, limit int) ([]fanprofile.PinnedShortItem, *fanprofile.PinnedShortCursor, error) {
				if gotViewerID != viewerID {
					t.Fatalf("ListPinnedShorts() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCursor != nil {
					t.Fatalf("ListPinnedShorts() cursor got %#v want nil", gotCursor)
				}
				if limit != fanprofile.DefaultPinnedShortsPageSize {
					t.Fatalf("ListPinnedShorts() limit got %d want %d", limit, fanprofile.DefaultPinnedShortsPageSize)
				}

				return []fanprofile.PinnedShortItem{}, nil, nil
			},
		},
		ShortDisplayAssets: stubShortDisplayAssetResolver{
			resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
				t.Fatal("ResolveShortDisplayAsset() was called unexpectedly")
				return media.VideoDisplayAsset{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/pinned-shorts", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/profile/pinned-shorts status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanProfilePinnedShortsResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || len(response.Data.Items) != 0 {
		t.Fatalf("response.Data.Items got %#v want empty", response.Data)
	}
	if response.Meta.Page == nil || response.Meta.Page.HasNext || response.Meta.Page.NextCursor != nil {
		t.Fatalf("response.Meta.Page got %#v want empty page info", response.Meta.Page)
	}
}

func TestFanProfilePinnedShortsMalformedCursorFallsBackToFirstPage(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         viewerID,
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		FanProfilePinnedShorts: stubFanProfilePinnedShortsReader{
			listPinnedShorts: func(_ context.Context, gotViewerID uuid.UUID, gotCursor *fanprofile.PinnedShortCursor, limit int) ([]fanprofile.PinnedShortItem, *fanprofile.PinnedShortCursor, error) {
				if gotViewerID != viewerID {
					t.Fatalf("ListPinnedShorts() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCursor != nil {
					t.Fatalf("ListPinnedShorts() cursor got %#v want nil", gotCursor)
				}
				if limit != fanprofile.DefaultPinnedShortsPageSize {
					t.Fatalf("ListPinnedShorts() limit got %d want %d", limit, fanprofile.DefaultPinnedShortsPageSize)
				}

				return []fanprofile.PinnedShortItem{}, nil, nil
			},
		},
		ShortDisplayAssets: stubShortDisplayAssetResolver{
			resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
				t.Fatal("ResolveShortDisplayAsset() was called unexpectedly")
				return media.VideoDisplayAsset{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/pinned-shorts?cursor=***", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/profile/pinned-shorts?cursor=*** status got %d want %d", rec.Code, http.StatusOK)
	}
}

func TestFanProfilePinnedShortsRejectsUnauthenticatedRequest(t *testing.T) {
	t.Parallel()

	readerCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
		FanProfilePinnedShorts: stubFanProfilePinnedShortsReader{
			listPinnedShorts: func(context.Context, uuid.UUID, *fanprofile.PinnedShortCursor, int) ([]fanprofile.PinnedShortItem, *fanprofile.PinnedShortCursor, error) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/pinned-shorts", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/fan/profile/pinned-shorts status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if readerCalled {
		t.Fatal("GET /api/fan/profile/pinned-shorts readerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"auth_required"`) {
		t.Fatalf("GET /api/fan/profile/pinned-shorts body got %q want auth_required", rec.Body.String())
	}
}

func TestFanProfilePinnedShortsReturnsInternalErrorWhenReaderFails(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         viewerID,
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		FanProfilePinnedShorts: stubFanProfilePinnedShortsReader{
			listPinnedShorts: func(context.Context, uuid.UUID, *fanprofile.PinnedShortCursor, int) ([]fanprofile.PinnedShortItem, *fanprofile.PinnedShortCursor, error) {
				return nil, nil, errors.New("boom")
			},
		},
		ShortDisplayAssets: stubShortDisplayAssetResolver{
			resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
				t.Fatal("ResolveShortDisplayAsset() was called unexpectedly")
				return media.VideoDisplayAsset{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/pinned-shorts", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/fan/profile/pinned-shorts status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("GET /api/fan/profile/pinned-shorts body got %q want internal_error", rec.Body.String())
	}
}

func TestFanProfilePinnedShortsReturnsInternalErrorWithoutAuthenticatedViewer(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{})
	router.GET("/api/fan/profile/pinned-shorts", func(c *gin.Context) {
		handleFanProfilePinnedShorts(c, stubFanProfilePinnedShortsReader{
			listPinnedShorts: func(context.Context, uuid.UUID, *fanprofile.PinnedShortCursor, int) ([]fanprofile.PinnedShortItem, *fanprofile.PinnedShortCursor, error) {
				t.Fatal("ListPinnedShorts() was called without authenticated viewer")
				return nil, nil, nil
			},
		}, stubShortDisplayAssetResolver{
			resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
				t.Fatal("ResolveShortDisplayAsset() was called unexpectedly")
				return media.VideoDisplayAsset{}, nil
			},
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/pinned-shorts", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/fan/profile/pinned-shorts status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("GET /api/fan/profile/pinned-shorts body got %q want internal_error", rec.Body.String())
	}
}
