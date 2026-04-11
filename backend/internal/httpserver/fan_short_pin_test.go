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

type stubFanShortPinWriter struct {
	pinPublicShort   func(context.Context, uuid.UUID, uuid.UUID) (shorts.PinMutationResult, error)
	unpinPublicShort func(context.Context, uuid.UUID, uuid.UUID) (shorts.PinMutationResult, error)
}

func (s stubFanShortPinWriter) PinPublicShort(ctx context.Context, viewerUserID uuid.UUID, shortID uuid.UUID) (shorts.PinMutationResult, error) {
	return s.pinPublicShort(ctx, viewerUserID, shortID)
}

func (s stubFanShortPinWriter) UnpinPublicShort(ctx context.Context, viewerUserID uuid.UUID, shortID uuid.UUID) (shorts.PinMutationResult, error) {
	return s.unpinPublicShort(ctx, viewerUserID, shortID)
}

func TestFanShortPinPutRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		FanShortPin: stubFanShortPinWriter{
			pinPublicShort: func(_ context.Context, gotViewerID uuid.UUID, gotShortID uuid.UUID) (shorts.PinMutationResult, error) {
				if gotViewerID != viewerID {
					t.Fatalf("PinPublicShort() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotShortID != shortID {
					t.Fatalf("PinPublicShort() shortID got %s want %s", gotShortID, shortID)
				}

				return shorts.PinMutationResult{
					IsPinned: true,
				}, nil
			},
			unpinPublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.PinMutationResult, error) {
				t.Fatal("UnpinPublicShort() was called on PUT route")
				return shorts.PinMutationResult{}, nil
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

	req := httptest.NewRequest(http.MethodPut, "/api/fan/shorts/"+shorts.FormatPublicShortID(shortID)+"/pin", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /api/fan/shorts/{shortId}/pin status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanShortPinResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || !response.Data.Viewer.IsPinned {
		t.Fatalf("response.Data got %#v want pin success", response.Data)
	}
	if response.Meta.Page != nil {
		t.Fatalf("response.Meta.Page got %#v want nil", response.Meta.Page)
	}
}

func TestFanShortPinDeleteRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		FanShortPin: stubFanShortPinWriter{
			pinPublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.PinMutationResult, error) {
				t.Fatal("PinPublicShort() was called on DELETE route")
				return shorts.PinMutationResult{}, nil
			},
			unpinPublicShort: func(_ context.Context, gotViewerID uuid.UUID, gotShortID uuid.UUID) (shorts.PinMutationResult, error) {
				if gotViewerID != viewerID {
					t.Fatalf("UnpinPublicShort() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotShortID != shortID {
					t.Fatalf("UnpinPublicShort() shortID got %s want %s", gotShortID, shortID)
				}

				return shorts.PinMutationResult{
					IsPinned: false,
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

	req := httptest.NewRequest(http.MethodDelete, "/api/fan/shorts/"+shorts.FormatPublicShortID(shortID)+"/pin", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("DELETE /api/fan/shorts/{shortId}/pin status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanShortPinResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Viewer.IsPinned {
		t.Fatalf("response.Data got %#v want unpin success", response.Data)
	}
}

func TestFanShortPinRouteRejectsUnauthenticatedRequest(t *testing.T) {
	t.Parallel()

	writerCalled := false
	router := NewHandler(HandlerConfig{
		FanShortPin: stubFanShortPinWriter{
			pinPublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.PinMutationResult, error) {
				writerCalled = true
				return shorts.PinMutationResult{}, nil
			},
			unpinPublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.PinMutationResult, error) {
				writerCalled = true
				return shorts.PinMutationResult{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/fan/shorts/short_missing/pin", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("PUT /api/fan/shorts/{shortId}/pin status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if writerCalled {
		t.Fatal("PUT /api/fan/shorts/{shortId}/pin writerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"auth_required"`) {
		t.Fatalf("PUT /api/fan/shorts/{shortId}/pin body got %q want auth_required", rec.Body.String())
	}
}

func TestFanShortPinRouteReturnsNotFound(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	router := NewHandler(HandlerConfig{
		FanShortPin: stubFanShortPinWriter{
			pinPublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.PinMutationResult, error) {
				return shorts.PinMutationResult{}, shorts.ErrShortNotFound
			},
			unpinPublicShort: func(context.Context, uuid.UUID, uuid.UUID) (shorts.PinMutationResult, error) {
				return shorts.PinMutationResult{}, nil
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

	req := httptest.NewRequest(http.MethodPut, "/api/fan/shorts/short_missing/pin", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("PUT /api/fan/shorts/{shortId}/pin status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("PUT /api/fan/shorts/{shortId}/pin body got %q want not_found", rec.Body.String())
	}
}
