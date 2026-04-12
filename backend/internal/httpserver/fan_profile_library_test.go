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

type stubFanProfileLibraryReader struct {
	listLibrary func(context.Context, uuid.UUID, *fanprofile.LibraryCursor, int) ([]fanprofile.LibraryItem, *fanprofile.LibraryCursor, error)
}

func (s stubFanProfileLibraryReader) ListLibrary(
	ctx context.Context,
	viewerID uuid.UUID,
	cursor *fanprofile.LibraryCursor,
	limit int,
) ([]fanprofile.LibraryItem, *fanprofile.LibraryCursor, error) {
	return s.listLibrary(ctx, viewerID, cursor, limit)
}

func TestFanProfileLibraryRoute(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	shortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	creatorID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	mediaAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	requestCursor := &fanprofile.LibraryCursor{
		MainID:          uuid.MustParse("66666666-6666-6666-6666-666666666666"),
		PurchasedAt:     now.Add(-time.Hour),
		UnlockCreatedAt: now.Add(-2 * time.Hour),
	}
	nextCursor := &fanprofile.LibraryCursor{
		MainID:          mainID,
		PurchasedAt:     now.Add(-3 * time.Hour),
		UnlockCreatedAt: now.Add(-3 * time.Hour),
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
		FanProfileLibrary: stubFanProfileLibraryReader{
			listLibrary: func(_ context.Context, gotViewerID uuid.UUID, gotCursor *fanprofile.LibraryCursor, limit int) ([]fanprofile.LibraryItem, *fanprofile.LibraryCursor, error) {
				if gotViewerID != viewerID {
					t.Fatalf("ListLibrary() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCursor == nil || gotCursor.MainID != requestCursor.MainID {
					t.Fatalf("ListLibrary() cursor got %#v want %#v", gotCursor, requestCursor)
				}
				if limit != fanprofile.DefaultLibraryPageSize {
					t.Fatalf("ListLibrary() limit got %d want %d", limit, fanprofile.DefaultLibraryPageSize)
				}

				return []fanprofile.LibraryItem{
					{
						CreatorAvatarURL:                 stringPtr("https://cdn.example.com/creator/mina/avatar.jpg"),
						CreatorBio:                       "quiet rooftop と hotel light の preview を軸に投稿。",
						CreatorDisplayName:               "Mina Rei",
						CreatorHandle:                    "minarei",
						CreatorUserID:                    creatorID,
						EntryShortCaption:                "quiet rooftop preview",
						EntryShortCanonicalMainID:        mainID,
						EntryShortID:                     shortID,
						EntryShortMediaAssetID:           mediaAssetID,
						EntryShortPreviewDurationSeconds: 16,
						MainDurationSeconds:              480,
						MainID:                           mainID,
						PurchasedAt:                      now,
						UnlockCreatedAt:                  now.Add(-time.Minute),
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
					DurationSeconds: 16,
					ID:              mediaAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/shorts/poster.jpg",
					URL:             "https://cdn.example.com/shorts/playback.mp4",
				}, nil
			},
		},
	})

	encodedRequestCursor := encodeFanProfileLibraryCursor(requestCursor)
	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/library?cursor="+*encodedRequestCursor, nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/profile/library status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanProfileLibraryResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || len(response.Data.Items) != 1 {
		t.Fatalf("response.Data.Items got %#v want len 1", response.Data)
	}
	if response.Data.Items[0].Main.ID != "main_22222222222222222222222222222222" {
		t.Fatalf("response.Data.Items[0].Main.ID got %q want %q", response.Data.Items[0].Main.ID, "main_22222222222222222222222222222222")
	}
	if response.Data.Items[0].Access.Status != "unlocked" || response.Data.Items[0].Access.Reason != "session_unlocked" {
		t.Fatalf("response.Data.Items[0].Access got %#v want unlocked/session_unlocked", response.Data.Items[0].Access)
	}
	if response.Data.Items[0].EntryShort.ID != "short_33333333333333333333333333333333" {
		t.Fatalf("response.Data.Items[0].EntryShort.ID got %q want %q", response.Data.Items[0].EntryShort.ID, "short_33333333333333333333333333333333")
	}
	if response.Meta.Page == nil || !response.Meta.Page.HasNext || response.Meta.Page.NextCursor == nil {
		t.Fatalf("response.Meta.Page got %#v want next cursor", response.Meta.Page)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(*response.Meta.Page.NextCursor)
	if err != nil {
		t.Fatalf("DecodeString(next cursor) error = %v, want nil", err)
	}
	if !strings.Contains(string(decoded), mainID.String()) {
		t.Fatalf("decoded next cursor got %q want main id %q", string(decoded), mainID.String())
	}
}

func TestFanProfileLibraryEmptyRoute(t *testing.T) {
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
		FanProfileLibrary: stubFanProfileLibraryReader{
			listLibrary: func(_ context.Context, gotViewerID uuid.UUID, gotCursor *fanprofile.LibraryCursor, limit int) ([]fanprofile.LibraryItem, *fanprofile.LibraryCursor, error) {
				if gotViewerID != viewerID {
					t.Fatalf("ListLibrary() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCursor != nil {
					t.Fatalf("ListLibrary() cursor got %#v want nil", gotCursor)
				}
				if limit != fanprofile.DefaultLibraryPageSize {
					t.Fatalf("ListLibrary() limit got %d want %d", limit, fanprofile.DefaultLibraryPageSize)
				}

				return []fanprofile.LibraryItem{}, nil, nil
			},
		},
		ShortDisplayAssets: stubShortDisplayAssetResolver{
			resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
				t.Fatal("ResolveShortDisplayAsset() was called unexpectedly")
				return media.VideoDisplayAsset{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/library", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/profile/library status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanProfileLibraryResponseData]
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

func TestFanProfileLibraryMalformedCursorFallsBackToFirstPage(t *testing.T) {
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
		FanProfileLibrary: stubFanProfileLibraryReader{
			listLibrary: func(_ context.Context, gotViewerID uuid.UUID, gotCursor *fanprofile.LibraryCursor, limit int) ([]fanprofile.LibraryItem, *fanprofile.LibraryCursor, error) {
				if gotViewerID != viewerID {
					t.Fatalf("ListLibrary() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCursor != nil {
					t.Fatalf("ListLibrary() cursor got %#v want nil", gotCursor)
				}
				if limit != fanprofile.DefaultLibraryPageSize {
					t.Fatalf("ListLibrary() limit got %d want %d", limit, fanprofile.DefaultLibraryPageSize)
				}

				return []fanprofile.LibraryItem{}, nil, nil
			},
		},
		ShortDisplayAssets: stubShortDisplayAssetResolver{
			resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
				t.Fatal("ResolveShortDisplayAsset() was called unexpectedly")
				return media.VideoDisplayAsset{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/library?cursor=***", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/profile/library?cursor=*** status got %d want %d", rec.Code, http.StatusOK)
	}
}

func TestFanProfileLibraryRejectsUnauthenticatedRequest(t *testing.T) {
	t.Parallel()

	readerCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
		FanProfileLibrary: stubFanProfileLibraryReader{
			listLibrary: func(context.Context, uuid.UUID, *fanprofile.LibraryCursor, int) ([]fanprofile.LibraryItem, *fanprofile.LibraryCursor, error) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/library", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/fan/profile/library status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if readerCalled {
		t.Fatal("GET /api/fan/profile/library readerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"auth_required"`) {
		t.Fatalf("GET /api/fan/profile/library body got %q want auth_required", rec.Body.String())
	}
}

func TestFanProfileLibraryReturnsInternalErrorWhenReaderFails(t *testing.T) {
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
		FanProfileLibrary: stubFanProfileLibraryReader{
			listLibrary: func(context.Context, uuid.UUID, *fanprofile.LibraryCursor, int) ([]fanprofile.LibraryItem, *fanprofile.LibraryCursor, error) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/library", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/fan/profile/library status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("GET /api/fan/profile/library body got %q want internal_error", rec.Body.String())
	}
}

func TestFanProfileLibraryReturnsInternalErrorWithoutAuthenticatedViewer(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{})
	router.GET("/api/fan/profile/library", func(c *gin.Context) {
		handleFanProfileLibrary(c, stubFanProfileLibraryReader{
			listLibrary: func(context.Context, uuid.UUID, *fanprofile.LibraryCursor, int) ([]fanprofile.LibraryItem, *fanprofile.LibraryCursor, error) {
				t.Fatal("ListLibrary() was called without authenticated viewer")
				return nil, nil, nil
			},
		}, stubShortDisplayAssetResolver{
			resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
				t.Fatal("ResolveShortDisplayAsset() was called unexpectedly")
				return media.VideoDisplayAsset{}, nil
			},
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/library", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/fan/profile/library status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("GET /api/fan/profile/library body got %q want internal_error", rec.Body.String())
	}
}
