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
)

type fanAuthServiceStub struct {
	issueSignInChallenge func(context.Context, string) (auth.IssuedChallenge, error)
	issueSignUpChallenge func(context.Context, string) (auth.IssuedChallenge, error)
	startSignInSession   func(context.Context, string, string) (auth.AuthenticatedSession, error)
	startSignUpSession   func(context.Context, string, string) (auth.AuthenticatedSession, error)
	logout               func(context.Context, string) error
}

func (s fanAuthServiceStub) IssueSignInChallenge(ctx context.Context, email string) (auth.IssuedChallenge, error) {
	return s.issueSignInChallenge(ctx, email)
}

func (s fanAuthServiceStub) IssueSignUpChallenge(ctx context.Context, email string) (auth.IssuedChallenge, error) {
	return s.issueSignUpChallenge(ctx, email)
}

func (s fanAuthServiceStub) StartSignInSession(ctx context.Context, email string, challengeToken string) (auth.AuthenticatedSession, error) {
	return s.startSignInSession(ctx, email, challengeToken)
}

func (s fanAuthServiceStub) StartSignUpSession(ctx context.Context, email string, challengeToken string) (auth.AuthenticatedSession, error) {
	return s.startSignUpSession(ctx, email, challengeToken)
}

func (s fanAuthServiceStub) Logout(ctx context.Context, rawSessionToken string) error {
	return s.logout(ctx, rawSessionToken)
}

func TestSignInChallengeReturnsChallengeToken(t *testing.T) {
	t.Parallel()

	expiresAt := time.Unix(1710000000, 0).UTC()
	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			issueSignInChallenge: func(_ context.Context, email string) (auth.IssuedChallenge, error) {
				if email != "fan@example.com" {
					t.Fatalf("IssueSignInChallenge() email got %q want %q", email, "fan@example.com")
				}

				return auth.IssuedChallenge{
					Token:     "challenge-token",
					ExpiresAt: expiresAt,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/sign-in/challenges", bytes.NewBufferString(`{"email":"fan@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/fan/auth/sign-in/challenges status got %d want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Data struct {
			ChallengeToken string `json:"challengeToken"`
			ExpiresAt      string `json:"expiresAt"`
		} `json:"data"`
		Error any `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if body.Data.ChallengeToken != "challenge-token" {
		t.Fatalf("POST /api/fan/auth/sign-in/challenges token got %q want %q", body.Data.ChallengeToken, "challenge-token")
	}
	if body.Data.ExpiresAt != expiresAt.Format(time.RFC3339) {
		t.Fatalf("POST /api/fan/auth/sign-in/challenges expiresAt got %q want %q", body.Data.ExpiresAt, expiresAt.Format(time.RFC3339))
	}
}

func TestSignUpChallengeMapsConflict(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			issueSignInChallenge: func(context.Context, string) (auth.IssuedChallenge, error) {
				t.Fatal("IssueSignInChallenge() should not be called")
				return auth.IssuedChallenge{}, nil
			},
			issueSignUpChallenge: func(context.Context, string) (auth.IssuedChallenge, error) {
				return auth.IssuedChallenge{}, auth.ErrEmailAlreadyRegistered
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/sign-up/challenges", bytes.NewBufferString(`{"email":"fan@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("POST /api/fan/auth/sign-up/challenges status got %d want %d", rec.Code, http.StatusConflict)
	}
	if !strings.Contains(rec.Body.String(), `"code":"email_already_registered"`) {
		t.Fatalf("POST /api/fan/auth/sign-up/challenges body got %q want conflict code", rec.Body.String())
	}
}

func TestSignInSessionSetsSessionCookie(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().UTC().Add(time.Hour)
	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			startSignInSession: func(_ context.Context, email string, challengeToken string) (auth.AuthenticatedSession, error) {
				if email != "fan@example.com" {
					t.Fatalf("StartSignInSession() email got %q want %q", email, "fan@example.com")
				}
				if challengeToken != "challenge-token" {
					t.Fatalf("StartSignInSession() challengeToken got %q want %q", challengeToken, "challenge-token")
				}

				return auth.AuthenticatedSession{
					Token:     "raw-session-token",
					ExpiresAt: expiresAt,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/sign-in/session", bytes.NewBufferString(`{"email":"fan@example.com","challengeToken":"challenge-token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /api/fan/auth/sign-in/session status got %d want %d", rec.Code, http.StatusNoContent)
	}

	setCookie := rec.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "shorts_fans_session=raw-session-token") {
		t.Fatalf("POST /api/fan/auth/sign-in/session set-cookie got %q want session token", setCookie)
	}
	if !strings.Contains(setCookie, "HttpOnly") {
		t.Fatalf("POST /api/fan/auth/sign-in/session set-cookie got %q want HttpOnly", setCookie)
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
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
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

func TestSignInSessionMapsInvalidChallenge(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanAuth: fanAuthServiceStub{
			startSignInSession: func(context.Context, string, string) (auth.AuthenticatedSession, error) {
				return auth.AuthenticatedSession{}, auth.ErrInvalidChallenge
			},
			logout: func(context.Context, string) error { return nil },
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/auth/sign-in/session", bytes.NewBufferString(`{"email":"fan@example.com","challengeToken":"bad-token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/fan/auth/sign-in/session status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_challenge"`) {
		t.Fatalf("POST /api/fan/auth/sign-in/session body got %q want invalid_challenge", rec.Body.String())
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
