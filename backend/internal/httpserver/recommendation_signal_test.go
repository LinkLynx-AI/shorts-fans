package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/recommendation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubRecommendationSignalWriter struct {
	recordProfileClick func(context.Context, uuid.UUID, uuid.UUID, string) (recommendation.RecordEventResult, error)
	recordShortSignal  func(context.Context, uuid.UUID, uuid.UUID, recommendation.EventKind, string) (recommendation.RecordEventResult, error)
}

func (s stubRecommendationSignalWriter) RecordProfileClick(
	ctx context.Context,
	viewerID uuid.UUID,
	creatorUserID uuid.UUID,
	idempotencyKey string,
) (recommendation.RecordEventResult, error) {
	return s.recordProfileClick(ctx, viewerID, creatorUserID, idempotencyKey)
}

func (s stubRecommendationSignalWriter) RecordShortSignal(
	ctx context.Context,
	viewerID uuid.UUID,
	shortID uuid.UUID,
	eventKind recommendation.EventKind,
	idempotencyKey string,
) (recommendation.RecordEventResult, error) {
	return s.recordShortSignal(ctx, viewerID, shortID, eventKind, idempotencyKey)
}

type stubRecommendationSignalExposureStore struct {
	hasCreatorExposure       func(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	hasShortExposure         func(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	rememberCreatorExposure  func(context.Context, uuid.UUID, uuid.UUID) error
	rememberCreatorExposures func(context.Context, uuid.UUID, []uuid.UUID) error
	rememberShortExposure    func(context.Context, uuid.UUID, uuid.UUID) error
	rememberShortExposures   func(context.Context, uuid.UUID, []uuid.UUID) error
}

func (s stubRecommendationSignalExposureStore) HasCreatorExposure(ctx context.Context, viewerID uuid.UUID, creatorUserID uuid.UUID) (bool, error) {
	return s.hasCreatorExposure(ctx, viewerID, creatorUserID)
}

func (s stubRecommendationSignalExposureStore) HasShortExposure(ctx context.Context, viewerID uuid.UUID, shortID uuid.UUID) (bool, error) {
	return s.hasShortExposure(ctx, viewerID, shortID)
}

func (s stubRecommendationSignalExposureStore) RememberCreatorExposure(ctx context.Context, viewerID uuid.UUID, creatorUserID uuid.UUID) error {
	if s.rememberCreatorExposure == nil {
		return nil
	}

	return s.rememberCreatorExposure(ctx, viewerID, creatorUserID)
}

func (s stubRecommendationSignalExposureStore) RememberCreatorExposures(ctx context.Context, viewerID uuid.UUID, creatorUserIDs []uuid.UUID) error {
	if s.rememberCreatorExposures != nil {
		return s.rememberCreatorExposures(ctx, viewerID, creatorUserIDs)
	}

	for _, creatorUserID := range creatorUserIDs {
		if err := s.RememberCreatorExposure(ctx, viewerID, creatorUserID); err != nil {
			return err
		}
	}

	return nil
}

func (s stubRecommendationSignalExposureStore) RememberShortExposure(ctx context.Context, viewerID uuid.UUID, shortID uuid.UUID) error {
	if s.rememberShortExposure == nil {
		return nil
	}

	return s.rememberShortExposure(ctx, viewerID, shortID)
}

func (s stubRecommendationSignalExposureStore) RememberShortExposures(ctx context.Context, viewerID uuid.UUID, shortIDs []uuid.UUID) error {
	if s.rememberShortExposures != nil {
		return s.rememberShortExposures(ctx, viewerID, shortIDs)
	}

	for _, shortID := range shortIDs {
		if err := s.RememberShortExposure(ctx, viewerID, shortID); err != nil {
			return err
		}
	}

	return nil
}

func TestRecommendationSignalRouteRecordsShortSignal(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	router := NewHandler(HandlerConfig{
		RecommendationSignalExposure: stubRecommendationSignalExposureStore{
			hasCreatorExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
				t.Fatal("HasCreatorExposure() should not be called")
				return false, nil
			},
			hasShortExposure: func(_ context.Context, gotViewerID uuid.UUID, gotShortID uuid.UUID) (bool, error) {
				if gotViewerID != viewerID {
					t.Fatalf("HasShortExposure() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotShortID != shortID {
					t.Fatalf("HasShortExposure() shortID got %s want %s", gotShortID, shortID)
				}

				return true, nil
			},
		},
		RecommendationSignals: stubRecommendationSignalWriter{
			recordProfileClick: func(context.Context, uuid.UUID, uuid.UUID, string) (recommendation.RecordEventResult, error) {
				t.Fatal("RecordProfileClick() should not be called")
				return recommendation.RecordEventResult{}, nil
			},
			recordShortSignal: func(_ context.Context, gotViewerID uuid.UUID, gotShortID uuid.UUID, gotEventKind recommendation.EventKind, gotIdempotencyKey string) (recommendation.RecordEventResult, error) {
				if gotViewerID != viewerID {
					t.Fatalf("RecordShortSignal() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotShortID != shortID {
					t.Fatalf("RecordShortSignal() shortID got %s want %s", gotShortID, shortID)
				}
				if gotEventKind != recommendation.EventKindImpression {
					t.Fatalf("RecordShortSignal() eventKind got %q want %q", gotEventKind, recommendation.EventKindImpression)
				}
				if gotIdempotencyKey != "impression:session-1" {
					t.Fatalf("RecordShortSignal() idempotencyKey got %q want %q", gotIdempotencyKey, "impression:session-1")
				}

				return recommendation.RecordEventResult{Recorded: true}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{CurrentViewer: &auth.CurrentViewer{ID: viewerID}}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/recommendation/events", strings.NewReader(`{"eventKind":"impression","idempotencyKey":"impression:session-1","shortId":"short_22222222222222222222222222222222"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/fan/recommendation/events status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[recommendationSignalResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || !response.Data.Recorded {
		t.Fatalf("response.Data got %#v want recorded=true", response.Data)
	}
}

func TestRecommendationSignalRouteRecordsProfileClick(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorUserID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	router := NewHandler(HandlerConfig{
		RecommendationSignalExposure: stubRecommendationSignalExposureStore{
			hasCreatorExposure: func(_ context.Context, gotViewerID uuid.UUID, gotCreatorUserID uuid.UUID) (bool, error) {
				if gotViewerID != viewerID {
					t.Fatalf("HasCreatorExposure() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCreatorUserID != creatorUserID {
					t.Fatalf("HasCreatorExposure() creatorUserID got %s want %s", gotCreatorUserID, creatorUserID)
				}

				return true, nil
			},
			hasShortExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
				t.Fatal("HasShortExposure() should not be called")
				return false, nil
			},
		},
		RecommendationSignals: stubRecommendationSignalWriter{
			recordProfileClick: func(_ context.Context, gotViewerID uuid.UUID, gotCreatorUserID uuid.UUID, gotIdempotencyKey string) (recommendation.RecordEventResult, error) {
				if gotViewerID != viewerID {
					t.Fatalf("RecordProfileClick() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCreatorUserID != creatorUserID {
					t.Fatalf("RecordProfileClick() creatorUserID got %s want %s", gotCreatorUserID, creatorUserID)
				}
				if gotIdempotencyKey != "profile:session-1" {
					t.Fatalf("RecordProfileClick() idempotencyKey got %q want %q", gotIdempotencyKey, "profile:session-1")
				}

				return recommendation.RecordEventResult{Recorded: false}, nil
			},
			recordShortSignal: func(context.Context, uuid.UUID, uuid.UUID, recommendation.EventKind, string) (recommendation.RecordEventResult, error) {
				t.Fatal("RecordShortSignal() should not be called")
				return recommendation.RecordEventResult{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{CurrentViewer: &auth.CurrentViewer{ID: viewerID}}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/recommendation/events", strings.NewReader(`{"eventKind":"profile_click","idempotencyKey":"profile:session-1","creatorId":"creator_22222222222222222222222222222222"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/fan/recommendation/events status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[recommendationSignalResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Recorded {
		t.Fatalf("response.Data got %#v want recorded=false", response.Data)
	}
}

func TestRecommendationSignalRouteRejectsInvalidRequest(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		RecommendationSignalExposure: stubRecommendationSignalExposureStore{
			hasCreatorExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
				t.Fatal("HasCreatorExposure() should not be called")
				return false, nil
			},
			hasShortExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
				t.Fatal("HasShortExposure() should not be called")
				return false, nil
			},
		},
		RecommendationSignals: stubRecommendationSignalWriter{
			recordProfileClick: func(context.Context, uuid.UUID, uuid.UUID, string) (recommendation.RecordEventResult, error) {
				t.Fatal("RecordProfileClick() should not be called")
				return recommendation.RecordEventResult{}, nil
			},
			recordShortSignal: func(context.Context, uuid.UUID, uuid.UUID, recommendation.EventKind, string) (recommendation.RecordEventResult, error) {
				t.Fatal("RecordShortSignal() should not be called")
				return recommendation.RecordEventResult{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{CurrentViewer: &auth.CurrentViewer{ID: uuid.New()}}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/recommendation/events", strings.NewReader(`{"eventKind":"impression","idempotencyKey":"impression:session-1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/fan/recommendation/events status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("response body got %q want invalid_request", rec.Body.String())
	}
}

func TestRecommendationSignalRouteRejectsMalformedShortID(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		RecommendationSignalExposure: stubRecommendationSignalExposureStore{
			hasCreatorExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
				t.Fatal("HasCreatorExposure() should not be called")
				return false, nil
			},
			hasShortExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
				t.Fatal("HasShortExposure() should not be called")
				return false, nil
			},
		},
		RecommendationSignals: stubRecommendationSignalWriter{
			recordProfileClick: func(context.Context, uuid.UUID, uuid.UUID, string) (recommendation.RecordEventResult, error) {
				t.Fatal("RecordProfileClick() should not be called")
				return recommendation.RecordEventResult{}, nil
			},
			recordShortSignal: func(context.Context, uuid.UUID, uuid.UUID, recommendation.EventKind, string) (recommendation.RecordEventResult, error) {
				t.Fatal("RecordShortSignal() should not be called")
				return recommendation.RecordEventResult{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{CurrentViewer: &auth.CurrentViewer{ID: uuid.New()}}, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/recommendation/events",
		strings.NewReader(`{"eventKind":"impression","idempotencyKey":"impression:session-1","shortId":"not-a-short-id"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/fan/recommendation/events status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("response body got %q want invalid_request", rec.Body.String())
	}
}

func TestRecommendationSignalRouteMapsNotFoundAndConflict(t *testing.T) {
	t.Parallel()

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		router := NewHandler(HandlerConfig{
			RecommendationSignalExposure: stubRecommendationSignalExposureStore{
				hasCreatorExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
					t.Fatal("HasCreatorExposure() should not be called")
					return false, nil
				},
				hasShortExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
					return true, nil
				},
			},
			RecommendationSignals: stubRecommendationSignalWriter{
				recordProfileClick: func(context.Context, uuid.UUID, uuid.UUID, string) (recommendation.RecordEventResult, error) {
					t.Fatal("RecordProfileClick() should not be called")
					return recommendation.RecordEventResult{}, nil
				},
				recordShortSignal: func(context.Context, uuid.UUID, uuid.UUID, recommendation.EventKind, string) (recommendation.RecordEventResult, error) {
					return recommendation.RecordEventResult{}, recommendation.ErrSignalTargetNotFound
				},
			},
			ViewerBootstrap: viewerBootstrapReaderStub{
				readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
					return auth.Bootstrap{CurrentViewer: &auth.CurrentViewer{ID: uuid.New()}}, nil
				},
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/api/fan/recommendation/events", strings.NewReader(`{"eventKind":"impression","idempotencyKey":"impression:session-1","shortId":"short_22222222222222222222222222222222"}`))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("POST /api/fan/recommendation/events status got %d want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()

		router := NewHandler(HandlerConfig{
			RecommendationSignalExposure: stubRecommendationSignalExposureStore{
				hasCreatorExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
					t.Fatal("HasCreatorExposure() should not be called")
					return false, nil
				},
				hasShortExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
					return true, nil
				},
			},
			RecommendationSignals: stubRecommendationSignalWriter{
				recordProfileClick: func(context.Context, uuid.UUID, uuid.UUID, string) (recommendation.RecordEventResult, error) {
					t.Fatal("RecordProfileClick() should not be called")
					return recommendation.RecordEventResult{}, nil
				},
				recordShortSignal: func(context.Context, uuid.UUID, uuid.UUID, recommendation.EventKind, string) (recommendation.RecordEventResult, error) {
					return recommendation.RecordEventResult{}, recommendation.ErrIdempotencyConflict
				},
			},
			ViewerBootstrap: viewerBootstrapReaderStub{
				readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
					return auth.Bootstrap{CurrentViewer: &auth.CurrentViewer{ID: uuid.New()}}, nil
				},
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/api/fan/recommendation/events", strings.NewReader(`{"eventKind":"impression","idempotencyKey":"impression:session-1","shortId":"short_22222222222222222222222222222222"}`))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("POST /api/fan/recommendation/events status got %d want %d", rec.Code, http.StatusConflict)
		}
		if !strings.Contains(rec.Body.String(), `"code":"idempotency_conflict"`) {
			t.Fatalf("response body got %q want idempotency_conflict", rec.Body.String())
		}
	})
}

func TestHandleRecommendationSignalReturnsInternalServerErrorWithoutViewerContext(t *testing.T) {
	t.Parallel()

	writer := stubRecommendationSignalWriter{
		recordProfileClick: func(context.Context, uuid.UUID, uuid.UUID, string) (recommendation.RecordEventResult, error) {
			t.Fatal("RecordProfileClick() should not be called")
			return recommendation.RecordEventResult{}, nil
		},
		recordShortSignal: func(context.Context, uuid.UUID, uuid.UUID, recommendation.EventKind, string) (recommendation.RecordEventResult, error) {
			t.Fatal("RecordShortSignal() should not be called")
			return recommendation.RecordEventResult{}, nil
		},
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/api/fan/recommendation/events", strings.NewReader(`{}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	c.Request = req

	handleRecommendationSignal(c, writer, stubRecommendationSignalExposureStore{
		hasCreatorExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
			return false, nil
		},
		hasShortExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
			return false, nil
		},
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("handleRecommendationSignal() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestRecordRecommendationSignalRejectsUnknownEventKind(t *testing.T) {
	t.Parallel()

	_, err := recordRecommendationSignal(context.Background(), stubRecommendationSignalWriter{
		recordProfileClick: func(context.Context, uuid.UUID, uuid.UUID, string) (recommendation.RecordEventResult, error) {
			return recommendation.RecordEventResult{}, nil
		},
		recordShortSignal: func(context.Context, uuid.UUID, uuid.UUID, recommendation.EventKind, string) (recommendation.RecordEventResult, error) {
			return recommendation.RecordEventResult{}, nil
		},
	}, stubRecommendationSignalExposureStore{
		hasCreatorExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
			return true, nil
		},
		hasShortExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
			return true, nil
		},
	}, uuid.New(), recommendationSignalRequestPayload{EventKind: "unknown"})
	if !errors.Is(err, recommendation.ErrEventKindInvalid) {
		t.Fatalf("recordRecommendationSignal() error got %v want %v", err, recommendation.ErrEventKindInvalid)
	}
}

func TestRecommendationSignalRouteRejectsUnsurfacedTarget(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		RecommendationSignalExposure: stubRecommendationSignalExposureStore{
			hasCreatorExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
				return false, nil
			},
			hasShortExposure: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
				return false, nil
			},
		},
		RecommendationSignals: stubRecommendationSignalWriter{
			recordProfileClick: func(context.Context, uuid.UUID, uuid.UUID, string) (recommendation.RecordEventResult, error) {
				t.Fatal("RecordProfileClick() should not be called")
				return recommendation.RecordEventResult{}, nil
			},
			recordShortSignal: func(context.Context, uuid.UUID, uuid.UUID, recommendation.EventKind, string) (recommendation.RecordEventResult, error) {
				t.Fatal("RecordShortSignal() should not be called")
				return recommendation.RecordEventResult{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{CurrentViewer: &auth.CurrentViewer{ID: uuid.New()}}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/recommendation/events", strings.NewReader(`{"eventKind":"impression","idempotencyKey":"impression:session-1","shortId":"short_22222222222222222222222222222222"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/fan/recommendation/events status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("response body got %q want invalid_request", rec.Body.String())
	}
}
