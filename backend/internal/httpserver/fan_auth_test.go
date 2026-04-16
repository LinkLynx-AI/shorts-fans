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
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type fanAuthServiceStub struct {
	confirmPasswordReset func(context.Context, string, string, string) error
	confirmSignUp        func(context.Context, string, string) (auth.AuthenticatedSession, error)
	logout               func(context.Context, string) error
	reAuthenticate       func(context.Context, string, string) (auth.AuthenticatedSession, error)
	signIn               func(context.Context, string, string) (auth.AuthenticatedSession, error)
	startPasswordReset   func(context.Context, string) (auth.FanAuthAcceptedStep, error)
	startSignUp          func(context.Context, string, string, string, string) (auth.FanAuthAcceptedStep, error)
}

func (s fanAuthServiceStub) ConfirmPasswordReset(ctx context.Context, email string, confirmationCode string, newPassword string) error {
	return s.confirmPasswordReset(ctx, email, confirmationCode, newPassword)
}

func (s fanAuthServiceStub) ConfirmSignUp(ctx context.Context, email string, confirmationCode string) (auth.AuthenticatedSession, error) {
	return s.confirmSignUp(ctx, email, confirmationCode)
}

func (s fanAuthServiceStub) Logout(ctx context.Context, rawSessionToken string) error {
	return s.logout(ctx, rawSessionToken)
}

func (s fanAuthServiceStub) ReAuthenticate(ctx context.Context, rawSessionToken string, password string) (auth.AuthenticatedSession, error) {
	return s.reAuthenticate(ctx, rawSessionToken, password)
}

func (s fanAuthServiceStub) SignIn(ctx context.Context, email string, password string) (auth.AuthenticatedSession, error) {
	return s.signIn(ctx, email, password)
}

func (s fanAuthServiceStub) StartPasswordReset(ctx context.Context, email string) (auth.FanAuthAcceptedStep, error) {
	return s.startPasswordReset(ctx, email)
}

func (s fanAuthServiceStub) StartSignUp(
	ctx context.Context,
	email string,
	displayName string,
	handle string,
	password string,
) (auth.FanAuthAcceptedStep, error) {
	return s.startSignUp(ctx, email, displayName, handle, password)
}

func TestSignInSetsSessionCookie(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().UTC().Add(time.Hour)
	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			signIn: func(_ context.Context, email string, password string) (auth.AuthenticatedSession, error) {
				if email != "fan@example.com" {
					t.Fatalf("SignIn() email got %q want %q", email, "fan@example.com")
				}
				if password != "VeryStrongPass123!" {
					t.Fatalf("SignIn() password got %q want %q", password, "VeryStrongPass123!")
				}

				return auth.AuthenticatedSession{
					Token:     "raw-session-token",
					ExpiresAt: expiresAt,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/sign-in", bytes.NewBufferString(`{"email":"fan@example.com","password":"VeryStrongPass123!"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /api/fan/auth/sign-in status got %d want %d", rec.Code, http.StatusNoContent)
	}
	setCookie := rec.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "shorts_fans_session=raw-session-token") {
		t.Fatalf("POST /api/fan/auth/sign-in set-cookie got %q want session token", setCookie)
	}
}

func TestSignInMapsInvalidCredentials(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			signIn: func(context.Context, string, string) (auth.AuthenticatedSession, error) {
				return auth.AuthenticatedSession{}, auth.ErrInvalidCredentials
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/sign-in", bytes.NewBufferString(`{"email":"fan@example.com","password":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /api/fan/auth/sign-in status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_credentials"`) {
		t.Fatalf("POST /api/fan/auth/sign-in body got %q want invalid_credentials", rec.Body.String())
	}
}

func TestSignInMapsConfirmationRequired(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			signIn: func(context.Context, string, string) (auth.AuthenticatedSession, error) {
				return auth.AuthenticatedSession{}, auth.ErrConfirmationRequired
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/sign-in", bytes.NewBufferString(`{"email":"fan@example.com","password":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST /api/fan/auth/sign-in status got %d want %d", rec.Code, http.StatusForbidden)
	}
	if !strings.Contains(rec.Body.String(), `"code":"confirmation_required"`) {
		t.Fatalf("POST /api/fan/auth/sign-in body got %q want confirmation_required", rec.Body.String())
	}
}

func TestSignInRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			signIn: func(context.Context, string, string) (auth.AuthenticatedSession, error) {
				t.Fatal("SignIn() should not be called for malformed JSON")
				return auth.AuthenticatedSession{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/sign-in", bytes.NewBufferString(`{"email":"fan@example.com"}{"extra":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/fan/auth/sign-in status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_email"`) {
		t.Fatalf("POST /api/fan/auth/sign-in body got %q want invalid_email", rec.Body.String())
	}
}

func TestSignUpReturnsAcceptedResponse(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			startSignUp: func(_ context.Context, email string, displayName string, handle string, password string) (auth.FanAuthAcceptedStep, error) {
				if email != "fan@example.com" {
					t.Fatalf("StartSignUp() email got %q want %q", email, "fan@example.com")
				}
				if displayName != "Mina" {
					t.Fatalf("StartSignUp() displayName got %q want %q", displayName, "Mina")
				}
				if handle != "@mina" {
					t.Fatalf("StartSignUp() handle got %q want %q", handle, "@mina")
				}
				if password != "VeryStrongPass123!" {
					t.Fatalf("StartSignUp() password got %q want %q", password, "VeryStrongPass123!")
				}

				return auth.FanAuthAcceptedStep{
					DeliveryDestinationHint: stringPointer("f***@example.com"),
					NextStep:                auth.FanAuthNextStepConfirmSignUp,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/auth/sign-up",
		bytes.NewBufferString(`{"email":"fan@example.com","displayName":"Mina","handle":"@mina","password":"VeryStrongPass123!"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/fan/auth/sign-up status got %d want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Data struct {
			DeliveryDestinationHint *string `json:"deliveryDestinationHint"`
			NextStep                string  `json:"nextStep"`
		} `json:"data"`
		Error any `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if body.Data.NextStep != auth.FanAuthNextStepConfirmSignUp {
		t.Fatalf("POST /api/fan/auth/sign-up next step got %q want %q", body.Data.NextStep, auth.FanAuthNextStepConfirmSignUp)
	}
	if body.Data.DeliveryDestinationHint == nil || *body.Data.DeliveryDestinationHint != "f***@example.com" {
		t.Fatalf("POST /api/fan/auth/sign-up delivery hint got %v want %q", body.Data.DeliveryDestinationHint, "f***@example.com")
	}
}

func TestSignUpMapsHandleAlreadyTaken(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			startSignUp: func(context.Context, string, string, string, string) (auth.FanAuthAcceptedStep, error) {
				return auth.FanAuthAcceptedStep{}, auth.ErrHandleAlreadyTaken
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/auth/sign-up",
		bytes.NewBufferString(`{"email":"fan@example.com","displayName":"Mina","handle":"@mina","password":"VeryStrongPass123!"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("POST /api/fan/auth/sign-up status got %d want %d", rec.Code, http.StatusConflict)
	}
	if !strings.Contains(rec.Body.String(), `"code":"handle_already_taken"`) {
		t.Fatalf("POST /api/fan/auth/sign-up body got %q want handle_already_taken", rec.Body.String())
	}
}

func TestSignUpConfirmSetsSessionCookie(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().UTC().Add(time.Hour)
	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			confirmSignUp: func(_ context.Context, email string, confirmationCode string) (auth.AuthenticatedSession, error) {
				if email != "fan@example.com" {
					t.Fatalf("ConfirmSignUp() email got %q want %q", email, "fan@example.com")
				}
				if confirmationCode != "123456" {
					t.Fatalf("ConfirmSignUp() confirmationCode got %q want %q", confirmationCode, "123456")
				}

				return auth.AuthenticatedSession{
					Token:     "raw-session-token",
					ExpiresAt: expiresAt,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/sign-up/confirm", bytes.NewBufferString(`{"email":"fan@example.com","confirmationCode":"123456"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /api/fan/auth/sign-up/confirm status got %d want %d", rec.Code, http.StatusNoContent)
	}
	if !strings.Contains(rec.Header().Get("Set-Cookie"), "shorts_fans_session=raw-session-token") {
		t.Fatalf("POST /api/fan/auth/sign-up/confirm set-cookie got %q want session token", rec.Header().Get("Set-Cookie"))
	}
}

func TestSignUpConfirmMapsExpiredCode(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			confirmSignUp: func(context.Context, string, string) (auth.AuthenticatedSession, error) {
				return auth.AuthenticatedSession{}, auth.ErrConfirmationCodeExpired
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/sign-up/confirm", bytes.NewBufferString(`{"email":"fan@example.com","confirmationCode":"123456"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/fan/auth/sign-up/confirm status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"confirmation_code_expired"`) {
		t.Fatalf("POST /api/fan/auth/sign-up/confirm body got %q want confirmation_code_expired", rec.Body.String())
	}
}

func TestPasswordResetReturnsAcceptedResponse(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			startPasswordReset: func(_ context.Context, email string) (auth.FanAuthAcceptedStep, error) {
				if email != "fan@example.com" {
					t.Fatalf("StartPasswordReset() email got %q want %q", email, "fan@example.com")
				}

				return auth.FanAuthAcceptedStep{
					DeliveryDestinationHint: stringPointer("f***@example.com"),
					NextStep:                auth.FanAuthNextStepConfirmPasswordReset,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/password-reset", bytes.NewBufferString(`{"email":"fan@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/fan/auth/password-reset status got %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"nextStep":"confirm_password_reset"`) {
		t.Fatalf("POST /api/fan/auth/password-reset body got %q want confirm_password_reset", rec.Body.String())
	}
}

func TestPasswordResetMapsRateLimited(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			startPasswordReset: func(context.Context, string) (auth.FanAuthAcceptedStep, error) {
				return auth.FanAuthAcceptedStep{}, auth.ErrRateLimited
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/password-reset", bytes.NewBufferString(`{"email":"fan@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("POST /api/fan/auth/password-reset status got %d want %d", rec.Code, http.StatusTooManyRequests)
	}
	if !strings.Contains(rec.Body.String(), `"code":"rate_limited"`) {
		t.Fatalf("POST /api/fan/auth/password-reset body got %q want rate_limited", rec.Body.String())
	}
}

func TestPasswordResetRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			startPasswordReset: func(context.Context, string) (auth.FanAuthAcceptedStep, error) {
				t.Fatal("StartPasswordReset() should not be called for malformed JSON")
				return auth.FanAuthAcceptedStep{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/password-reset", bytes.NewBufferString(`{"email":"fan@example.com"}{"extra":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/fan/auth/password-reset status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_email"`) {
		t.Fatalf("POST /api/fan/auth/password-reset body got %q want invalid_email", rec.Body.String())
	}
}

func TestPasswordResetConfirmReturnsNoContent(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			confirmPasswordReset: func(_ context.Context, email string, confirmationCode string, newPassword string) error {
				if email != "fan@example.com" {
					t.Fatalf("ConfirmPasswordReset() email got %q want %q", email, "fan@example.com")
				}
				if confirmationCode != "123456" {
					t.Fatalf("ConfirmPasswordReset() confirmationCode got %q want %q", confirmationCode, "123456")
				}
				if newPassword != "AnotherStrongPass456!" {
					t.Fatalf("ConfirmPasswordReset() newPassword got %q want %q", newPassword, "AnotherStrongPass456!")
				}

				return nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/auth/password-reset/confirm",
		bytes.NewBufferString(`{"email":"fan@example.com","confirmationCode":"123456","newPassword":"AnotherStrongPass456!"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /api/fan/auth/password-reset/confirm status got %d want %d", rec.Code, http.StatusNoContent)
	}
}

func TestPasswordResetConfirmMapsPasswordPolicyViolation(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			confirmPasswordReset: func(context.Context, string, string, string) error {
				return auth.ErrPasswordPolicyViolation
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/auth/password-reset/confirm",
		bytes.NewBufferString(`{"email":"fan@example.com","confirmationCode":"123456","newPassword":"AnotherStrongPass456!"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/fan/auth/password-reset/confirm status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"password_policy_violation"`) {
		t.Fatalf("POST /api/fan/auth/password-reset/confirm body got %q want password_policy_violation", rec.Body.String())
	}
}

func TestPasswordResetConfirmMapsInvalidConfirmationCode(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			confirmPasswordReset: func(context.Context, string, string, string) error {
				return auth.ErrInvalidConfirmationCode
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/auth/password-reset/confirm",
		bytes.NewBufferString(`{"email":"fan@example.com","confirmationCode":"123456","newPassword":"AnotherStrongPass456!"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/fan/auth/password-reset/confirm status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_confirmation_code"`) {
		t.Fatalf("POST /api/fan/auth/password-reset/confirm body got %q want invalid_confirmation_code", rec.Body.String())
	}
}

func TestWriteMappedAuthError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err        error
		wantCode   string
		wantStatus int
	}{
		{err: auth.ErrInvalidDisplayName, wantCode: "invalid_display_name", wantStatus: http.StatusBadRequest},
		{err: auth.ErrInvalidHandle, wantCode: "invalid_handle", wantStatus: http.StatusBadRequest},
		{err: auth.ErrInvalidPassword, wantCode: "invalid_password", wantStatus: http.StatusBadRequest},
		{err: auth.ErrInvalidConfirmationCode, wantCode: "invalid_confirmation_code", wantStatus: http.StatusBadRequest},
		{err: auth.ErrRateLimited, wantCode: "rate_limited", wantStatus: http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.wantCode, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)

			if !writeMappedAuthError(c, tt.err, "fan_auth_test") {
				t.Fatal("writeMappedAuthError() returned false for mapped error")
			}
			if rec.Code != tt.wantStatus {
				t.Fatalf("writeMappedAuthError() status got %d want %d", rec.Code, tt.wantStatus)
			}
			if !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Fatalf("writeMappedAuthError() body got %q want code %q", rec.Body.String(), tt.wantCode)
			}
		})
	}
}

func TestReAuthSetsSessionCookie(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().UTC().Add(time.Hour)
	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			reAuthenticate: func(_ context.Context, rawSessionToken string, password string) (auth.AuthenticatedSession, error) {
				if rawSessionToken != "raw-session-token" {
					t.Fatalf("ReAuthenticate() raw session token got %q want %q", rawSessionToken, "raw-session-token")
				}
				if password != "VeryStrongPass123!" {
					t.Fatalf("ReAuthenticate() password got %q want %q", password, "VeryStrongPass123!")
				}

				return auth.AuthenticatedSession{
					Token:     "rotated-session-token",
					ExpiresAt: expiresAt,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/re-auth", bytes.NewBufferString(`{"password":"VeryStrongPass123!"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /api/fan/auth/re-auth status got %d want %d", rec.Code, http.StatusNoContent)
	}
	if !strings.Contains(rec.Header().Get("Set-Cookie"), "shorts_fans_session=rotated-session-token") {
		t.Fatalf("POST /api/fan/auth/re-auth set-cookie got %q want rotated session token", rec.Header().Get("Set-Cookie"))
	}
}

func TestReAuthMapsAuthRequired(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			reAuthenticate: func(context.Context, string, string) (auth.AuthenticatedSession, error) {
				return auth.AuthenticatedSession{}, auth.ErrAuthenticationRequired
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/re-auth", bytes.NewBufferString(`{"password":"VeryStrongPass123!"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /api/fan/auth/re-auth status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(rec.Body.String(), `"code":"auth_required"`) {
		t.Fatalf("POST /api/fan/auth/re-auth body got %q want auth_required", rec.Body.String())
	}
}

func TestLogoutClearsSessionCookie(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			logout: func(_ context.Context, rawSessionToken string) error {
				if rawSessionToken != "raw-session-token" {
					t.Fatalf("Logout() raw session token got %q want %q", rawSessionToken, "raw-session-token")
				}
				return nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/fan/auth/session", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE /api/fan/auth/session status got %d want %d", rec.Code, http.StatusNoContent)
	}
	setCookie := rec.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "shorts_fans_session=") || !strings.Contains(setCookie, "Max-Age=0") {
		t.Fatalf("DELETE /api/fan/auth/session set-cookie got %q want cleared session cookie", setCookie)
	}
}

func TestLogoutReturnsInternalErrorForUnexpectedFailure(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			logout: func(context.Context, string) error {
				return errors.New("boom")
			},
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/fan/auth/session", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("DELETE /api/fan/auth/session status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func stringPointer(value string) *string {
	return &value
}
