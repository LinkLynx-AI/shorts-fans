package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type viewerCreatorRegistrationWriterStub struct {
	registerApprovedCreator func(context.Context, creator.SelfServeRegistrationInput) (creator.SelfServeRegistrationResult, error)
}

func (s viewerCreatorRegistrationWriterStub) RegisterApprovedCreator(
	ctx context.Context,
	input creator.SelfServeRegistrationInput,
) (creator.SelfServeRegistrationResult, error) {
	return s.registerApprovedCreator(ctx, input)
}

type viewerActiveModeSwitcherStub struct {
	switchActiveMode func(context.Context, string, auth.ActiveMode) error
}

func (s viewerActiveModeSwitcherStub) SwitchActiveMode(
	ctx context.Context,
	rawSessionToken string,
	activeMode auth.ActiveMode,
) error {
	return s.switchActiveMode(ctx, rawSessionToken, activeMode)
}

func TestViewerCreatorRegistrationRequiresAuth(t *testing.T) {
	t.Parallel()

	writerCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationWriterStub{
			registerApprovedCreator: func(context.Context, creator.SelfServeRegistrationInput) (creator.SelfServeRegistrationResult, error) {
				writerCalled = true
				return creator.SelfServeRegistrationResult{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/viewer/creator-registration", bytes.NewBufferString(`{"displayName":"Mina","handle":"mina","bio":"bio"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if writerCalled {
		t.Fatal("POST /api/viewer/creator-registration writerCalled = true, want false")
	}
}

func TestViewerCreatorRegistrationSuccess(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var gotInput creator.SelfServeRegistrationInput

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
		CreatorRegistration: viewerCreatorRegistrationWriterStub{
			registerApprovedCreator: func(_ context.Context, input creator.SelfServeRegistrationInput) (creator.SelfServeRegistrationResult, error) {
				gotInput = input
				return creator.SelfServeRegistrationResult{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/viewer/creator-registration", bytes.NewBufferString(`{"displayName":"Mina","handle":"mina","bio":"bio"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusNoContent)
	}
	if gotInput.UserID != viewerID {
		t.Fatalf("POST /api/viewer/creator-registration user id got %s want %s", gotInput.UserID, viewerID)
	}
	if gotInput.DisplayName != "Mina" {
		t.Fatalf("POST /api/viewer/creator-registration display name got %q want %q", gotInput.DisplayName, "Mina")
	}
	if gotInput.Handle != "mina" {
		t.Fatalf("POST /api/viewer/creator-registration handle got %q want %q", gotInput.Handle, "mina")
	}
	if gotInput.Bio != "bio" {
		t.Fatalf("POST /api/viewer/creator-registration bio got %q want %q", gotInput.Bio, "bio")
	}
}

func TestViewerCreatorRegistrationInvalidDisplayName(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.New(),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationWriterStub{
			registerApprovedCreator: func(context.Context, creator.SelfServeRegistrationInput) (creator.SelfServeRegistrationResult, error) {
				return creator.SelfServeRegistrationResult{}, creator.ErrInvalidDisplayName
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/viewer/creator-registration", bytes.NewBufferString(`{"displayName":" ","handle":"mina","bio":"bio"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_display_name"`) {
		t.Fatalf("POST /api/viewer/creator-registration body got %q want invalid_display_name", rec.Body.String())
	}
}

func TestViewerCreatorRegistrationInvalidHandle(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.New(),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationWriterStub{
			registerApprovedCreator: func(context.Context, creator.SelfServeRegistrationInput) (creator.SelfServeRegistrationResult, error) {
				return creator.SelfServeRegistrationResult{}, creator.ErrInvalidHandle
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/viewer/creator-registration", bytes.NewBufferString(`{"displayName":"Mina","handle":"@","bio":"bio"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_handle"`) {
		t.Fatalf("POST /api/viewer/creator-registration body got %q want invalid_handle", rec.Body.String())
	}
}

func TestViewerCreatorRegistrationHandleAlreadyTaken(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.New(),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationWriterStub{
			registerApprovedCreator: func(context.Context, creator.SelfServeRegistrationInput) (creator.SelfServeRegistrationResult, error) {
				return creator.SelfServeRegistrationResult{}, creator.ErrHandleAlreadyTaken
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/viewer/creator-registration", bytes.NewBufferString(`{"displayName":"Mina","handle":"mina","bio":"bio"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusConflict)
	}
	if !strings.Contains(rec.Body.String(), `"code":"handle_already_taken"`) {
		t.Fatalf("POST /api/viewer/creator-registration body got %q want handle_already_taken", rec.Body.String())
	}
}

func TestViewerActiveModeSwitchSuccess(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var gotRawToken string
	var gotActiveMode auth.ActiveMode

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(_ context.Context, rawSessionToken string, activeMode auth.ActiveMode) error {
				gotRawToken = rawSessionToken
				gotActiveMode = activeMode
				return nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/viewer/active-mode", bytes.NewBufferString(`{"activeMode":"creator"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusNoContent)
	}
	if gotRawToken != "raw-session-token" {
		t.Fatalf("PUT /api/viewer/active-mode raw token got %q want %q", gotRawToken, "raw-session-token")
	}
	if gotActiveMode != auth.ActiveModeCreator {
		t.Fatalf("PUT /api/viewer/active-mode active mode got %q want %q", gotActiveMode, auth.ActiveModeCreator)
	}
}

func TestViewerActiveModeSwitchRejectsUnavailableCreatorMode(t *testing.T) {
	t.Parallel()

	switcherCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
				switcherCalled = true
				return nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/viewer/active-mode", bytes.NewBufferString(`{"activeMode":"creator"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusForbidden)
	}
	if switcherCalled {
		t.Fatal("PUT /api/viewer/active-mode switcherCalled = true, want false")
	}

	var response responseEnvelope[json.RawMessage]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "creator_mode_unavailable" {
		t.Fatalf("PUT /api/viewer/active-mode error got %#v want creator_mode_unavailable", response.Error)
	}
}

func TestViewerCreatorRegistrationRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	writerCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.New(),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationWriterStub{
			registerApprovedCreator: func(context.Context, creator.SelfServeRegistrationInput) (creator.SelfServeRegistrationResult, error) {
				writerCalled = true
				return creator.SelfServeRegistrationResult{}, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/creator-registration",
		bytes.NewBufferString(`{"displayName":"Mina","handle":"mina"}{"bio":"bio"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if writerCalled {
		t.Fatal("POST /api/viewer/creator-registration writerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("POST /api/viewer/creator-registration body got %q want invalid_request", rec.Body.String())
	}
}

func TestViewerCreatorRegistrationRejectsUnknownField(t *testing.T) {
	t.Parallel()

	writerCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.New(),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationWriterStub{
			registerApprovedCreator: func(context.Context, creator.SelfServeRegistrationInput) (creator.SelfServeRegistrationResult, error) {
				writerCalled = true
				return creator.SelfServeRegistrationResult{}, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/creator-registration",
		bytes.NewBufferString(`{"displayName":"Mina","handle":"mina","bio":"bio","unexpected":true}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if writerCalled {
		t.Fatal("POST /api/viewer/creator-registration writerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("POST /api/viewer/creator-registration body got %q want invalid_request", rec.Body.String())
	}
}

func TestViewerCreatorRegistrationReturnsInternalErrorWhenWriterFails(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.New(),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationWriterStub{
			registerApprovedCreator: func(context.Context, creator.SelfServeRegistrationInput) (creator.SelfServeRegistrationResult, error) {
				return creator.SelfServeRegistrationResult{}, errors.New("boom")
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/viewer/creator-registration", bytes.NewBufferString(`{"displayName":"Mina","handle":"mina","bio":"bio"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestViewerActiveModeSwitchRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	switcherCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
				switcherCalled = true
				return nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/active-mode",
		bytes.NewBufferString(`{"activeMode":"creator"}{"extra":true}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if switcherCalled {
		t.Fatal("PUT /api/viewer/active-mode switcherCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("PUT /api/viewer/active-mode body got %q want invalid_request", rec.Body.String())
	}
}

func TestViewerActiveModeSwitchRejectsUnknownField(t *testing.T) {
	t.Parallel()

	switcherCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
				switcherCalled = true
				return nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/active-mode",
		bytes.NewBufferString(`{"activeMode":"fan","unexpected":true}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if switcherCalled {
		t.Fatal("PUT /api/viewer/active-mode switcherCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("PUT /api/viewer/active-mode body got %q want invalid_request", rec.Body.String())
	}
}

func TestViewerActiveModeSwitchRejectsInvalidMode(t *testing.T) {
	t.Parallel()

	switcherCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
				switcherCalled = true
				return nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/viewer/active-mode", bytes.NewBufferString(`{"activeMode":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if switcherCalled {
		t.Fatal("PUT /api/viewer/active-mode switcherCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_active_mode"`) {
		t.Fatalf("PUT /api/viewer/active-mode body got %q want invalid_active_mode", rec.Body.String())
	}
}

func TestViewerActiveModeSwitchMapsInvalidModeFromSwitcher(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						ActiveMode:           auth.ActiveModeCreator,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
				return auth.ErrInvalidActiveMode
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/viewer/active-mode", bytes.NewBufferString(`{"activeMode":"fan"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_active_mode"`) {
		t.Fatalf("PUT /api/viewer/active-mode body got %q want invalid_active_mode", rec.Body.String())
	}
}

func TestViewerActiveModeSwitchReturnsInternalErrorWhenCookieMissing(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/viewer/active-mode", bytes.NewBufferString(`{"activeMode":"fan"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set(authenticatedViewerContextKey, auth.CurrentViewer{
		ID:                   uuid.New(),
		ActiveMode:           auth.ActiveModeCreator,
		CanAccessCreatorMode: true,
	})

	handleViewerActiveModeSwitch(ctx, viewerActiveModeSwitcherStub{
		switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
			t.Fatal("SwitchActiveMode() should not be called")
			return nil
		},
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("handleViewerActiveModeSwitch() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}
