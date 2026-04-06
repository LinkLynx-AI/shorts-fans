package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/google/uuid"
)

type viewerBootstrapReaderStub struct {
	readCurrentViewer func(context.Context, string) (auth.Bootstrap, error)
}

func (s viewerBootstrapReaderStub) ReadCurrentViewer(ctx context.Context, rawSessionToken string) (auth.Bootstrap, error) {
	return s.readCurrentViewer(ctx, rawSessionToken)
}

func TestViewerBootstrapReturnsUnauthenticatedState(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/viewer/bootstrap", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/viewer/bootstrap status got %d want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Data struct {
			CurrentViewer *struct {
				ID string `json:"id"`
			} `json:"currentViewer"`
		} `json:"data"`
		Error any `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if body.Data.CurrentViewer != nil {
		t.Fatalf("GET /api/viewer/bootstrap current viewer got %#v want nil", body.Data.CurrentViewer)
	}
	if body.Error != nil {
		t.Fatalf("GET /api/viewer/bootstrap error got %#v want nil", body.Error)
	}
}

func TestViewerBootstrapReturnsCurrentViewer(t *testing.T) {
	t.Parallel()

	expectedID := uuid.New()
	var gotRawSessionToken string
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(_ context.Context, rawSessionToken string) (auth.Bootstrap, error) {
				gotRawSessionToken = rawSessionToken

				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   expectedID,
						ActiveMode:           auth.ActiveModeCreator,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/viewer/bootstrap", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/viewer/bootstrap status got %d want %d", rec.Code, http.StatusOK)
	}
	if gotRawSessionToken != "raw-session-token" {
		t.Fatalf("GET /api/viewer/bootstrap raw token got %q want %q", gotRawSessionToken, "raw-session-token")
	}

	var body struct {
		Data struct {
			CurrentViewer *struct {
				ID                   string `json:"id"`
				ActiveMode           string `json:"activeMode"`
				CanAccessCreatorMode bool   `json:"canAccessCreatorMode"`
			} `json:"currentViewer"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if body.Data.CurrentViewer == nil {
		t.Fatal("GET /api/viewer/bootstrap current viewer = nil, want viewer")
	}
	if body.Data.CurrentViewer.ID != expectedID.String() {
		t.Fatalf("GET /api/viewer/bootstrap id got %q want %q", body.Data.CurrentViewer.ID, expectedID.String())
	}
	if body.Data.CurrentViewer.ActiveMode != "creator" {
		t.Fatalf("GET /api/viewer/bootstrap active mode got %q want %q", body.Data.CurrentViewer.ActiveMode, "creator")
	}
	if !body.Data.CurrentViewer.CanAccessCreatorMode {
		t.Fatal("GET /api/viewer/bootstrap can access creator mode = false, want true")
	}
}

func TestViewerBootstrapReturnsInternalErrorEnvelope(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, errors.New("query failed")
			},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/viewer/bootstrap", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/viewer/bootstrap status got %d want %d", rec.Code, http.StatusInternalServerError)
	}

	var body struct {
		Data  any `json:"data"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if body.Data != nil {
		t.Fatalf("GET /api/viewer/bootstrap data got %#v want nil", body.Data)
	}
	if body.Error.Code != "internal_error" {
		t.Fatalf("GET /api/viewer/bootstrap error code got %q want %q", body.Error.Code, "internal_error")
	}
}
