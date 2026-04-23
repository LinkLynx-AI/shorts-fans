package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/google/uuid"
)

type stubFanShortLikeWriter struct {
	likePublicShort   func(context.Context, uuid.UUID, uuid.UUID) (shorts.LikeMutationResult, error)
	unlikePublicShort func(context.Context, uuid.UUID, uuid.UUID) (shorts.LikeMutationResult, error)
}

func (s stubFanShortLikeWriter) LikePublicShort(ctx context.Context, viewerUserID uuid.UUID, shortID uuid.UUID) (shorts.LikeMutationResult, error) {
	return s.likePublicShort(ctx, viewerUserID, shortID)
}

func (s stubFanShortLikeWriter) UnlikePublicShort(ctx context.Context, viewerUserID uuid.UUID, shortID uuid.UUID) (shorts.LikeMutationResult, error) {
	return s.unlikePublicShort(ctx, viewerUserID, shortID)
}

func TestFanShortLikePutRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		FanShortLike: stubFanShortLikeWriter{
			likePublicShort: func(_ context.Context, gotViewerID uuid.UUID, gotShortID uuid.UUID) (shorts.LikeMutationResult, error) {
				if gotViewerID != viewerID {
					t.Fatalf("LikePublicShort() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotShortID != shortID {
					t.Fatalf("LikePublicShort() shortID got %s want %s", gotShortID, shortID)
				}

				return shorts.LikeMutationResult{
					HasLiked:  true,
					LikeCount: 42,
				}, nil
			},
			unlikePublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.LikeMutationResult, error) {
				t.Fatal("UnlikePublicShort() was called on PUT route")
				return shorts.LikeMutationResult{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/fan/shorts/"+shorts.FormatPublicShortID(shortID)+"/like", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /api/fan/shorts/{shortId}/like status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanShortLikeResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || !response.Data.Viewer.HasLiked || response.Data.Engagement.LikeCount != 42 {
		t.Fatalf("response.Data got %#v want like success", response.Data)
	}
	if response.Meta.Page != nil {
		t.Fatalf("response.Meta.Page got %#v want nil", response.Meta.Page)
	}
}

func TestFanShortLikeDeleteRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		FanShortLike: stubFanShortLikeWriter{
			likePublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.LikeMutationResult, error) {
				t.Fatal("LikePublicShort() was called on DELETE route")
				return shorts.LikeMutationResult{}, nil
			},
			unlikePublicShort: func(_ context.Context, gotViewerID uuid.UUID, gotShortID uuid.UUID) (shorts.LikeMutationResult, error) {
				if gotViewerID != viewerID {
					t.Fatalf("UnlikePublicShort() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotShortID != shortID {
					t.Fatalf("UnlikePublicShort() shortID got %s want %s", gotShortID, shortID)
				}

				return shorts.LikeMutationResult{
					HasLiked:  false,
					LikeCount: 41,
				}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/fan/shorts/"+shorts.FormatPublicShortID(shortID)+"/like", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("DELETE /api/fan/shorts/{shortId}/like status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanShortLikeResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Viewer.HasLiked || response.Data.Engagement.LikeCount != 41 {
		t.Fatalf("response.Data got %#v want unlike success", response.Data)
	}
}

func TestFanShortLikeRouteRejectsUnauthenticatedRequest(t *testing.T) {
	t.Parallel()

	writerCalled := false
	router := NewHandler(HandlerConfig{
		FanShortLike: stubFanShortLikeWriter{
			likePublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.LikeMutationResult, error) {
				writerCalled = true
				return shorts.LikeMutationResult{}, nil
			},
			unlikePublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.LikeMutationResult, error) {
				writerCalled = true
				return shorts.LikeMutationResult{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/fan/shorts/short_missing/like", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("PUT /api/fan/shorts/{shortId}/like status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if writerCalled {
		t.Fatal("PUT /api/fan/shorts/{shortId}/like writerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"auth_required"`) {
		t.Fatalf("PUT /api/fan/shorts/{shortId}/like body got %q want auth_required", rec.Body.String())
	}
}

func TestFanShortLikeRouteReturnsNotFound(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	router := NewHandler(HandlerConfig{
		FanShortLike: stubFanShortLikeWriter{
			likePublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.LikeMutationResult, error) {
				return shorts.LikeMutationResult{}, shorts.ErrShortNotFound
			},
			unlikePublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.LikeMutationResult, error) {
				return shorts.LikeMutationResult{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/fan/shorts/short_missing/like", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("PUT /api/fan/shorts/{shortId}/like status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("PUT /api/fan/shorts/{shortId}/like body got %q want not_found", rec.Body.String())
	}
}
