package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanmain"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubFanUnlockMainService struct {
	getPlaybackSurface func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, string) (fanmain.PlaybackSurface, error)
	getUnlockSurface   func(context.Context, uuid.UUID, string, uuid.UUID) (fanmain.UnlockSurface, error)
	issueAccessEntry   func(context.Context, string, fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error)
}

func (s stubFanUnlockMainService) GetPlaybackSurface(ctx context.Context, viewerID uuid.UUID, sessionBinding string, mainID uuid.UUID, fromShortID uuid.UUID, grantToken string) (fanmain.PlaybackSurface, error) {
	return s.getPlaybackSurface(ctx, viewerID, sessionBinding, mainID, fromShortID, grantToken)
}

func (s stubFanUnlockMainService) GetUnlockSurface(ctx context.Context, viewerID uuid.UUID, sessionBinding string, shortID uuid.UUID) (fanmain.UnlockSurface, error) {
	return s.getUnlockSurface(ctx, viewerID, sessionBinding, shortID)
}

func (s stubFanUnlockMainService) IssueAccessEntry(ctx context.Context, sessionBinding string, input fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error) {
	return s.issueAccessEntry(ctx, sessionBinding, input)
}

type stubMainDisplayAssetResolver struct {
	resolve func(context.Context, media.MainDisplaySource, media.AccessBoundary, time.Duration) (media.VideoDisplayAsset, error)
}

func (s stubMainDisplayAssetResolver) ResolveMainDisplayAsset(ctx context.Context, source media.MainDisplaySource, boundary media.AccessBoundary, ttl time.Duration) (media.VideoDisplayAsset, error) {
	return s.resolve(ctx, source, boundary, ttl)
}

func TestFanShortUnlockRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	router := NewHandler(HandlerConfig{
		FanUnlockMain: stubFanUnlockMainService{
			getUnlockSurface: func(_ context.Context, gotViewerID uuid.UUID, sessionBinding string, gotShortID uuid.UUID) (fanmain.UnlockSurface, error) {
				if gotViewerID != viewerID {
					t.Fatalf("GetUnlockSurface() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if sessionBinding == "" {
					t.Fatal("GetUnlockSurface() sessionBinding = empty, want value")
				}
				if gotShortID != shortID {
					t.Fatalf("GetUnlockSurface() shortID got %s want %s", gotShortID, shortID)
				}

				mainDurationSeconds := int64(480)
				priceJPY := int64(1800)

				return fanmain.UnlockSurface{
					Access: fanmain.MainAccessState{
						MainID: mainID,
						Reason: "unlock_required",
						Status: "locked",
					},
					Creator: fanmain.CreatorSummary{
						Bio:         "quiet rooftop specialist",
						DisplayName: "Mina Rei",
						Handle:      "minarei",
						ID:          viewerID,
					},
					Main: fanmain.MainSummary{
						DurationSeconds: 480,
						ID:              mainID,
						MediaAssetID:    uuid.MustParse("55555555-5555-5555-5555-555555555555"),
						PriceJPY:        1800,
					},
					MainAccessToken: "signed-entry-token",
					Setup:           fanmain.UnlockSetupState{},
					Short: fanmain.ShortSummary{
						Caption:                "quiet rooftop preview",
						CanonicalMainID:        mainID,
						CreatorUserID:          viewerID,
						ID:                     shortID,
						MediaAssetID:           shortAssetID,
						PreviewDurationSeconds: 16,
					},
					UnlockCta: fanmain.UnlockCtaState{
						MainDurationSeconds: &mainDurationSeconds,
						PriceJPY:            &priceJPY,
						State:               "unlock_available",
					},
				}, nil
			},
		},
		MainDisplayAssets: stubMainDisplayAssetResolver{
			resolve: func(context.Context, media.MainDisplaySource, media.AccessBoundary, time.Duration) (media.VideoDisplayAsset, error) {
				t.Fatal("ResolveMainDisplayAsset() should not be called")
				return media.VideoDisplayAsset{}, nil
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
					ID:              shortAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/shorts/poster.jpg",
					URL:             "https://cdn.example.com/shorts/playback.mp4",
				}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID: viewerID,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/shorts/short_22222222222222222222222222222222/unlock", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/shorts/:shortId/unlock status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[unlockSurfaceResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.MainAccessEntry.Token != "signed-entry-token" {
		t.Fatalf("response.Data.MainAccessEntry.Token got %#v want %q", response.Data, "signed-entry-token")
	}
	if response.Data.Short.ID != "short_22222222222222222222222222222222" {
		t.Fatalf("response.Data.Short.ID got %q want %q", response.Data.Short.ID, "short_22222222222222222222222222222222")
	}
}

func TestFanMainAccessEntryRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	router := NewHandler(HandlerConfig{
		FanUnlockMain: stubFanUnlockMainService{
			issueAccessEntry: func(_ context.Context, sessionBinding string, input fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error) {
				if sessionBinding == "" {
					t.Fatal("IssueAccessEntry() sessionBinding = empty, want value")
				}
				if input.ViewerID != viewerID {
					t.Fatalf("IssueAccessEntry() viewerID got %s want %s", input.ViewerID, viewerID)
				}
				if input.MainID != mainID {
					t.Fatalf("IssueAccessEntry() mainID got %s want %s", input.MainID, mainID)
				}
				if input.FromShortID != shortID {
					t.Fatalf("IssueAccessEntry() fromShortID got %s want %s", input.FromShortID, shortID)
				}
				if input.EntryToken != "signed-entry-token" {
					t.Fatalf("IssueAccessEntry() entryToken got %q want %q", input.EntryToken, "signed-entry-token")
				}

				return fanmain.AccessEntryResult{
					GrantKind:  fanmain.MainPlaybackGrantKindUnlocked,
					GrantToken: "signed-grant-token",
				}, nil
			},
		},
		MainDisplayAssets: stubMainDisplayAssetResolver{
			resolve: func(context.Context, media.MainDisplaySource, media.AccessBoundary, time.Duration) (media.VideoDisplayAsset, error) {
				t.Fatal("ResolveMainDisplayAsset() should not be called")
				return media.VideoDisplayAsset{}, nil
			},
		},
		ShortDisplayAssets: stubShortDisplayAssetResolver{
			resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
				t.Fatal("ResolveShortDisplayAsset() should not be called")
				return media.VideoDisplayAsset{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID: viewerID,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/mains/main_33333333333333333333333333333333/access-entry",
		strings.NewReader(`{"acceptedAge":true,"acceptedTerms":true,"entryToken":"signed-entry-token","fromShortId":"short_22222222222222222222222222222222"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/fan/mains/:mainId/access-entry status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[mainAccessEntryResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || !strings.Contains(response.Data.Href, "signed-grant-token") {
		t.Fatalf("response.Data.Href got %#v want grant token", response.Data)
	}
}

func TestFanMainPlaybackRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	mainAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")

	router := NewHandler(HandlerConfig{
		FanUnlockMain: stubFanUnlockMainService{
			getPlaybackSurface: func(_ context.Context, gotViewerID uuid.UUID, sessionBinding string, gotMainID uuid.UUID, gotFromShortID uuid.UUID, gotGrant string) (fanmain.PlaybackSurface, error) {
				if gotViewerID != viewerID {
					t.Fatalf("GetPlaybackSurface() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if sessionBinding == "" {
					t.Fatal("GetPlaybackSurface() sessionBinding = empty, want value")
				}
				if gotMainID != mainID {
					t.Fatalf("GetPlaybackSurface() mainID got %s want %s", gotMainID, mainID)
				}
				if gotFromShortID != shortID {
					t.Fatalf("GetPlaybackSurface() fromShortID got %s want %s", gotFromShortID, shortID)
				}
				if gotGrant != "signed-grant-token" {
					t.Fatalf("GetPlaybackSurface() grant got %q want %q", gotGrant, "signed-grant-token")
				}

				return fanmain.PlaybackSurface{
					Access: fanmain.MainAccessState{
						MainID: mainID,
						Reason: "session_unlocked",
						Status: "unlocked",
					},
					Creator: fanmain.CreatorSummary{
						Bio:         "quiet rooftop specialist",
						DisplayName: "Mina Rei",
						Handle:      "minarei",
						ID:          viewerID,
					},
					EntryShort: fanmain.ShortSummary{
						Caption:                "quiet rooftop preview",
						CanonicalMainID:        mainID,
						CreatorUserID:          viewerID,
						ID:                     shortID,
						MediaAssetID:           shortAssetID,
						PreviewDurationSeconds: 16,
					},
					Main: fanmain.MainSummary{
						DurationSeconds: 480,
						ID:              mainID,
						MediaAssetID:    mainAssetID,
						PriceJPY:        1800,
					},
				}, nil
			},
		},
		MainDisplayAssets: stubMainDisplayAssetResolver{
			resolve: func(_ context.Context, source media.MainDisplaySource, boundary media.AccessBoundary, ttl time.Duration) (media.VideoDisplayAsset, error) {
				if source.MainID != mainID {
					t.Fatalf("ResolveMainDisplayAsset() mainID got %s want %s", source.MainID, mainID)
				}
				if boundary != media.AccessBoundaryPrivate {
					t.Fatalf("ResolveMainDisplayAsset() boundary got %s want %s", boundary, media.AccessBoundaryPrivate)
				}
				if ttl != 0 {
					t.Fatalf("ResolveMainDisplayAsset() ttl got %s want 0", ttl)
				}

				return media.VideoDisplayAsset{
					DurationSeconds: 480,
					ID:              mainAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/mains/poster.jpg",
					URL:             "https://cdn.example.com/mains/playback.mp4",
				}, nil
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
					ID:              shortAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/shorts/poster.jpg",
					URL:             "https://cdn.example.com/shorts/playback.mp4",
				}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID: viewerID,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/fan/mains/main_33333333333333333333333333333333/playback?fromShortId=short_22222222222222222222222222222222&grant=signed-grant-token",
		nil,
	)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/mains/:mainId/playback status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[mainPlaybackResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Main.Media.URL != "https://cdn.example.com/mains/playback.mp4" {
		t.Fatalf("response.Data.Main.Media.URL got %#v want playback URL", response.Data)
	}
}

func TestFanMainPlaybackRouteRemembersCreatorExposureForUnlockedPlayback(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	shortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mainID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	shortAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	mainAssetID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	rememberedCreator := uuid.Nil

	router := NewHandler(HandlerConfig{
		FanUnlockMain: stubFanUnlockMainService{
			getPlaybackSurface: func(_ context.Context, gotViewerID uuid.UUID, _ string, gotMainID uuid.UUID, gotFromShortID uuid.UUID, gotGrant string) (fanmain.PlaybackSurface, error) {
				if gotViewerID != viewerID {
					t.Fatalf("GetPlaybackSurface() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotMainID != mainID {
					t.Fatalf("GetPlaybackSurface() mainID got %s want %s", gotMainID, mainID)
				}
				if gotFromShortID != shortID {
					t.Fatalf("GetPlaybackSurface() fromShortID got %s want %s", gotFromShortID, shortID)
				}
				if gotGrant != "signed-grant-token" {
					t.Fatalf("GetPlaybackSurface() grant got %q want %q", gotGrant, "signed-grant-token")
				}

				return fanmain.PlaybackSurface{
					Access: fanmain.MainAccessState{
						MainID: mainID,
						Reason: "session_unlocked",
						Status: "unlocked",
					},
					Creator: fanmain.CreatorSummary{
						DisplayName: "Mina Rei",
						Handle:      "minarei",
						ID:          creatorID,
					},
					EntryShort: fanmain.ShortSummary{
						CanonicalMainID:        mainID,
						CreatorUserID:          creatorID,
						ID:                     shortID,
						MediaAssetID:           shortAssetID,
						PreviewDurationSeconds: 16,
					},
					Main: fanmain.MainSummary{
						DurationSeconds: 480,
						ID:              mainID,
						MediaAssetID:    mainAssetID,
					},
				}, nil
			},
			getUnlockSurface: func(context.Context, uuid.UUID, string, uuid.UUID) (fanmain.UnlockSurface, error) {
				t.Fatal("GetUnlockSurface() should not be called")
				return fanmain.UnlockSurface{}, nil
			},
			issueAccessEntry: func(context.Context, string, fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error) {
				t.Fatal("IssueAccessEntry() should not be called")
				return fanmain.AccessEntryResult{}, nil
			},
		},
		MainDisplayAssets: stubMainDisplayAssetResolver{
			resolve: func(_ context.Context, source media.MainDisplaySource, boundary media.AccessBoundary, _ time.Duration) (media.VideoDisplayAsset, error) {
				if source.MainID != mainID {
					t.Fatalf("ResolveMainDisplayAsset() mainID got %s want %s", source.MainID, mainID)
				}
				if boundary != media.AccessBoundaryPrivate {
					t.Fatalf("ResolveMainDisplayAsset() boundary got %s want %s", boundary, media.AccessBoundaryPrivate)
				}

				return media.VideoDisplayAsset{
					DurationSeconds: 480,
					ID:              mainAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/mains/poster.jpg",
					URL:             "https://cdn.example.com/mains/playback.mp4",
				}, nil
			},
		},
		RecommendationSignalExposure: stubRecommendationSignalExposureStore{
			hasCreatorExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
				return false, nil
			},
			hasShortExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
				return false, nil
			},
			rememberCreatorExposure: func(_ context.Context, gotViewerID uuid.UUID, gotCreatorID uuid.UUID) error {
				if gotViewerID != viewerID {
					t.Fatalf("RememberCreatorExposure() viewerID got %s want %s", gotViewerID, viewerID)
				}
				rememberedCreator = gotCreatorID
				return nil
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
					ID:              shortAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/shorts/poster.jpg",
					URL:             "https://cdn.example.com/shorts/playback.mp4",
				}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{CurrentViewer: &auth.CurrentViewer{ID: viewerID}}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/mains/main_44444444444444444444444444444444/playback?fromShortId=short_33333333333333333333333333333333&grant=signed-grant-token", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/mains/:mainId/playback status got %d want %d", rec.Code, http.StatusOK)
	}
	if rememberedCreator != creatorID {
		t.Fatalf("RememberCreatorExposure() creatorID got %s want %s", rememberedCreator, creatorID)
	}
}

func TestFanUnlockMainRouteErrors(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	buildRouter := func(service stubFanUnlockMainService) http.Handler {
		return NewHandler(HandlerConfig{
			FanUnlockMain: service,
			MainDisplayAssets: stubMainDisplayAssetResolver{
				resolve: func(context.Context, media.MainDisplaySource, media.AccessBoundary, time.Duration) (media.VideoDisplayAsset, error) {
					return media.VideoDisplayAsset{}, errors.New("boom")
				},
			},
			ShortDisplayAssets: stubShortDisplayAssetResolver{
				resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
					return media.VideoDisplayAsset{}, errors.New("boom")
				},
			},
			ViewerBootstrap: viewerBootstrapReaderStub{
				readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
					return auth.Bootstrap{
						CurrentViewer: &auth.CurrentViewer{
							ID: viewerID,
						},
					}, nil
				},
			},
		})
	}

	t.Run("unlock not found", func(t *testing.T) {
		t.Parallel()

		router := buildRouter(stubFanUnlockMainService{
			getUnlockSurface: func(context.Context, uuid.UUID, string, uuid.UUID) (fanmain.UnlockSurface, error) {
				return fanmain.UnlockSurface{}, fanmain.ErrShortUnlockNotFound
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/api/fan/shorts/short_22222222222222222222222222222222/unlock", nil)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("unlock not found status got %d want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("access entry invalid body", func(t *testing.T) {
		t.Parallel()

		router := buildRouter(stubFanUnlockMainService{
			issueAccessEntry: func(context.Context, string, fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error) {
				t.Fatal("IssueAccessEntry() should not be called")
				return fanmain.AccessEntryResult{}, nil
			},
		})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/fan/mains/main_33333333333333333333333333333333/access-entry",
			strings.NewReader("{"),
		)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("access entry invalid body status got %d want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("access entry locked", func(t *testing.T) {
		t.Parallel()

		router := buildRouter(stubFanUnlockMainService{
			issueAccessEntry: func(context.Context, string, fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error) {
				return fanmain.AccessEntryResult{}, fanmain.ErrMainLocked
			},
		})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/fan/mains/main_33333333333333333333333333333333/access-entry",
			strings.NewReader(`{"acceptedAge":true,"acceptedTerms":true,"entryToken":"signed-entry-token","fromShortId":"short_22222222222222222222222222222222"}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("access entry locked status got %d want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("playback missing grant", func(t *testing.T) {
		t.Parallel()

		router := buildRouter(stubFanUnlockMainService{
			getPlaybackSurface: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, string) (fanmain.PlaybackSurface, error) {
				t.Fatal("GetPlaybackSurface() should not be called")
				return fanmain.PlaybackSurface{}, nil
			},
		})

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/fan/mains/main_33333333333333333333333333333333/playback?fromShortId=short_22222222222222222222222222222222",
			nil,
		)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("playback missing grant status got %d want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("playback locked", func(t *testing.T) {
		t.Parallel()

		router := buildRouter(stubFanUnlockMainService{
			getPlaybackSurface: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, string) (fanmain.PlaybackSurface, error) {
				return fanmain.PlaybackSurface{}, fanmain.ErrMainLocked
			},
		})

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/fan/mains/main_33333333333333333333333333333333/playback?fromShortId=short_22222222222222222222222222222222&grant=signed-grant-token",
			nil,
		)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("playback locked status got %d want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("playback not found", func(t *testing.T) {
		t.Parallel()

		router := buildRouter(stubFanUnlockMainService{
			getPlaybackSurface: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, string) (fanmain.PlaybackSurface, error) {
				return fanmain.PlaybackSurface{}, fanmain.ErrPlaybackNotFound
			},
		})

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/fan/mains/main_33333333333333333333333333333333/playback?fromShortId=short_22222222222222222222222222222222&grant=signed-grant-token",
			nil,
		)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("playback not found status got %d want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("playback asset resolution failure does not remember exposure", func(t *testing.T) {
		t.Parallel()

		mainAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
		shortAssetID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
		creatorID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
		rememberedCreator := uuid.Nil

		router := NewHandler(HandlerConfig{
			FanUnlockMain: stubFanUnlockMainService{
				getPlaybackSurface: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, string) (fanmain.PlaybackSurface, error) {
					return fanmain.PlaybackSurface{
						Access: fanmain.MainAccessState{MainID: mainID, Reason: "session_unlocked", Status: "unlocked"},
						Creator: fanmain.CreatorSummary{
							DisplayName: "Mina Rei",
							Handle:      "minarei",
							ID:          creatorID,
						},
						EntryShort: fanmain.ShortSummary{
							CanonicalMainID:        mainID,
							CreatorUserID:          creatorID,
							ID:                     shortID,
							MediaAssetID:           shortAssetID,
							PreviewDurationSeconds: 16,
						},
						Main: fanmain.MainSummary{
							DurationSeconds: 480,
							ID:              mainID,
							MediaAssetID:    mainAssetID,
						},
					}, nil
				},
			},
			MainDisplayAssets: stubMainDisplayAssetResolver{
				resolve: func(context.Context, media.MainDisplaySource, media.AccessBoundary, time.Duration) (media.VideoDisplayAsset, error) {
					return media.VideoDisplayAsset{}, errors.New("boom")
				},
			},
			RecommendationSignalExposure: stubRecommendationSignalExposureStore{
				hasCreatorExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
					return false, nil
				},
				hasShortExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
					return false, nil
				},
				rememberCreatorExposure: func(context.Context, uuid.UUID, uuid.UUID) error {
					rememberedCreator = creatorID
					return nil
				},
			},
			ShortDisplayAssets: stubShortDisplayAssetResolver{
				resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
					return media.VideoDisplayAsset{
						DurationSeconds: 16,
						ID:              shortAssetID,
						Kind:            "video",
						URL:             "https://cdn.example.com/shorts/playback.mp4",
					}, nil
				},
			},
			ViewerBootstrap: viewerBootstrapReaderStub{
				readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
					return auth.Bootstrap{CurrentViewer: &auth.CurrentViewer{ID: viewerID}}, nil
				},
			},
		})

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/fan/mains/main_33333333333333333333333333333333/playback?fromShortId=short_22222222222222222222222222222222&grant=signed-grant-token",
			nil,
		)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("playback asset resolution failure status got %d want %d", rec.Code, http.StatusInternalServerError)
		}
		if rememberedCreator != uuid.Nil {
			t.Fatalf("RememberCreatorExposure() creatorID got %s want nil UUID", rememberedCreator)
		}
	})

	t.Run("access entry invalid short", func(t *testing.T) {
		t.Parallel()

		router := buildRouter(stubFanUnlockMainService{
			issueAccessEntry: func(context.Context, string, fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error) {
				t.Fatal("IssueAccessEntry() should not be called")
				return fanmain.AccessEntryResult{}, nil
			},
		})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/fan/mains/main_33333333333333333333333333333333/access-entry",
			strings.NewReader(`{"acceptedAge":true,"acceptedTerms":true,"entryToken":"signed-entry-token","fromShortId":"invalid-short-id"}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("access entry invalid short status got %d want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("playback success owner boundary", func(t *testing.T) {
		t.Parallel()

		mainAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
		shortAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
		var gotBoundary media.AccessBoundary

		router := NewHandler(HandlerConfig{
			FanUnlockMain: stubFanUnlockMainService{
				getPlaybackSurface: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, string) (fanmain.PlaybackSurface, error) {
					return fanmain.PlaybackSurface{
						Access: fanmain.MainAccessState{MainID: mainID, Reason: "owner_preview", Status: "owner"},
						Creator: fanmain.CreatorSummary{
							DisplayName: "Mina Rei",
							Handle:      "minarei",
							ID:          viewerID,
						},
						EntryShort: fanmain.ShortSummary{
							Caption:                "quiet rooftop preview",
							CanonicalMainID:        mainID,
							CreatorUserID:          viewerID,
							ID:                     shortID,
							MediaAssetID:           shortAssetID,
							PreviewDurationSeconds: 16,
						},
						Main: fanmain.MainSummary{
							DurationSeconds: 480,
							ID:              mainID,
							MediaAssetID:    mainAssetID,
						},
					}, nil
				},
			},
			MainDisplayAssets: stubMainDisplayAssetResolver{
				resolve: func(_ context.Context, source media.MainDisplaySource, boundary media.AccessBoundary, ttl time.Duration) (media.VideoDisplayAsset, error) {
					gotBoundary = boundary
					return media.VideoDisplayAsset{
						DurationSeconds: 480,
						ID:              mainAssetID,
						Kind:            "video",
						URL:             "https://cdn.example.com/mains/playback.mp4",
					}, nil
				},
			},
			ShortDisplayAssets: stubShortDisplayAssetResolver{
				resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
					return media.VideoDisplayAsset{
						DurationSeconds: 16,
						ID:              shortAssetID,
						Kind:            "video",
						URL:             "https://cdn.example.com/shorts/playback.mp4",
					}, nil
				},
			},
			ViewerBootstrap: viewerBootstrapReaderStub{
				readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
					return auth.Bootstrap{
						CurrentViewer: &auth.CurrentViewer{ID: viewerID},
					}, nil
				},
			},
		})

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/fan/mains/main_33333333333333333333333333333333/playback?fromShortId=short_22222222222222222222222222222222&grant=signed-grant-token",
			nil,
		)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("playback success owner status got %d want %d", rec.Code, http.StatusOK)
		}
		if gotBoundary != media.AccessBoundaryOwner {
			t.Fatalf("ResolveMainDisplayAsset() boundary got %s want %s", gotBoundary, media.AccessBoundaryOwner)
		}
	})
}

func TestFanUnlockMainHelpers(t *testing.T) {
	t.Parallel()

	t.Run("resolve authenticated viewer request", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		if _, _, ok := resolveAuthenticatedViewerRequest(c); ok {
			t.Fatal("resolveAuthenticatedViewerRequest() ok = true, want false without viewer")
		}

		c.Set(authenticatedViewerContextKey, auth.CurrentViewer{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111")})
		if _, _, ok := resolveAuthenticatedViewerRequest(c); ok {
			t.Fatal("resolveAuthenticatedViewerRequest() ok = true, want false without cookie")
		}

		c.Request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		if _, sessionBinding, ok := resolveAuthenticatedViewerRequest(c); !ok || sessionBinding == "" {
			t.Fatalf("resolveAuthenticatedViewerRequest() got ok=%v sessionBinding=%q want true/non-empty", ok, sessionBinding)
		}
	})

	t.Run("playback boundary", func(t *testing.T) {
		t.Parallel()

		if got := resolveMainPlaybackBoundary("owner"); got != media.AccessBoundaryOwner {
			t.Fatalf("resolveMainPlaybackBoundary(owner) got %s want %s", got, media.AccessBoundaryOwner)
		}
		if got := resolveMainPlaybackBoundary("unlocked"); got != media.AccessBoundaryPrivate {
			t.Fatalf("resolveMainPlaybackBoundary(unlocked) got %s want %s", got, media.AccessBoundaryPrivate)
		}
	})
}
