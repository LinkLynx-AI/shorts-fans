package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestProtectedFanAuthGuardRejectsMissingSession(t *testing.T) {
	t.Parallel()

	readerCalled := false
	handlerCalled := false

	router := NewHandler(HandlerConfig{})
	router.GET(
		"/api/fan/profile",
		buildProtectedFanAuthGuard(viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				readerCalled = true
				return auth.Bootstrap{}, nil
			},
		}, "fan_profile", "fan profile requires authentication"),
		func(c *gin.Context) {
			handlerCalled = true
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/fan/profile status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if readerCalled {
		t.Fatal("GET /api/fan/profile readerCalled = true, want false")
	}
	if handlerCalled {
		t.Fatal("GET /api/fan/profile handlerCalled = true, want false")
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data != nil {
		t.Fatalf("response.Data got %#v want nil", response.Data)
	}
	if response.Meta.Page != nil {
		t.Fatalf("response.Meta.Page got %#v want nil", response.Meta.Page)
	}
	if response.Error == nil || response.Error.Code != "auth_required" {
		t.Fatalf("response.Error got %#v want auth_required", response.Error)
	}
	if response.Error.Message != "fan profile requires authentication" {
		t.Fatalf("response.Error.Message got %q want %q", response.Error.Message, "fan profile requires authentication")
	}
}

func TestProtectedFanAuthGuardRejectsUnresolvedViewer(t *testing.T) {
	t.Parallel()

	handlerCalled := false
	var gotRawSessionToken string

	router := NewHandler(HandlerConfig{})
	router.GET(
		"/api/fan/profile",
		buildProtectedFanAuthGuard(viewerBootstrapReaderStub{
			readCurrentViewer: func(_ context.Context, rawSessionToken string) (auth.Bootstrap, error) {
				gotRawSessionToken = rawSessionToken
				return auth.Bootstrap{}, nil
			},
		}, "fan_profile", "fan profile requires authentication"),
		func(c *gin.Context) {
			handlerCalled = true
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/fan/profile status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if gotRawSessionToken != "raw-session-token" {
		t.Fatalf("GET /api/fan/profile raw session token got %q want %q", gotRawSessionToken, "raw-session-token")
	}
	if handlerCalled {
		t.Fatal("GET /api/fan/profile handlerCalled = true, want false")
	}
}

func TestProtectedFanAuthGuardPassesAuthenticatedViewerToContext(t *testing.T) {
	t.Parallel()

	expectedViewer := auth.CurrentViewer{
		ID:                   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ActiveMode:           auth.ActiveModeFan,
		CanAccessCreatorMode: false,
	}
	handlerCalled := false

	router := NewHandler(HandlerConfig{})
	router.GET(
		"/api/fan/profile",
		buildProtectedFanAuthGuard(viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &expectedViewer,
				}, nil
			},
		}, "fan_profile", "fan profile requires authentication"),
		func(c *gin.Context) {
			handlerCalled = true

			viewer, ok := authenticatedViewerFromContext(c)
			if !ok {
				c.Status(http.StatusInternalServerError)
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"activeMode":           string(viewer.ActiveMode),
				"canAccessCreatorMode": viewer.CanAccessCreatorMode,
				"id":                   viewer.ID.String(),
			})
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/profile status got %d want %d", rec.Code, http.StatusOK)
	}
	if !handlerCalled {
		t.Fatal("GET /api/fan/profile handlerCalled = false, want true")
	}

	var body struct {
		ActiveMode           string `json:"activeMode"`
		CanAccessCreatorMode bool   `json:"canAccessCreatorMode"`
		ID                   string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if body.ID != expectedViewer.ID.String() {
		t.Fatalf("response id got %q want %q", body.ID, expectedViewer.ID.String())
	}
	if body.ActiveMode != string(expectedViewer.ActiveMode) {
		t.Fatalf("response activeMode got %q want %q", body.ActiveMode, expectedViewer.ActiveMode)
	}
	if body.CanAccessCreatorMode != expectedViewer.CanAccessCreatorMode {
		t.Fatalf("response canAccessCreatorMode got %v want %v", body.CanAccessCreatorMode, expectedViewer.CanAccessCreatorMode)
	}
}

func TestProtectedFanAuthGuardReturnsInternalErrorWhenReaderFails(t *testing.T) {
	t.Parallel()

	handlerCalled := false

	router := NewHandler(HandlerConfig{})
	router.GET(
		"/api/fan/profile",
		buildProtectedFanAuthGuard(viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, errors.New("boom")
			},
		}, "fan_profile", "fan profile requires authentication"),
		func(c *gin.Context) {
			handlerCalled = true
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/fan/profile status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
	if handlerCalled {
		t.Fatal("GET /api/fan/profile handlerCalled = true, want false")
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "internal_error" {
		t.Fatalf("response.Error got %#v want internal_error", response.Error)
	}
}
