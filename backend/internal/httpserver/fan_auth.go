package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
)

// FanAuthService は fan auth transport が依存する auth lifecycle 境界です。
type FanAuthService interface {
	IssueSignInChallenge(ctx context.Context, email string) (auth.IssuedChallenge, error)
	IssueSignUpChallenge(ctx context.Context, email string) (auth.IssuedChallenge, error)
	StartSignInSession(ctx context.Context, email string, challengeToken string) (auth.AuthenticatedSession, error)
	StartSignUpSession(ctx context.Context, email string, challengeToken string, displayName string, handle string) (auth.AuthenticatedSession, error)
	Logout(ctx context.Context, rawSessionToken string) error
}

// AuthCookieConfig は auth session cookie の transport 設定です。
type AuthCookieConfig struct {
	Secure bool
}

type authChallengeRequest struct {
	Email string `json:"email"`
}

type authSessionRequest struct {
	ChallengeToken string `json:"challengeToken"`
	Email          string `json:"email"`
}

type authSignUpSessionRequest struct {
	ChallengeToken string `json:"challengeToken"`
	DisplayName    string `json:"displayName"`
	Email          string `json:"email"`
	Handle         string `json:"handle"`
}

type authChallengeResponseData struct {
	ChallengeToken string `json:"challengeToken"`
	ExpiresAt      string `json:"expiresAt"`
}

// registerFanAuthRoutes は fan auth transport を router に登録します。
func registerFanAuthRoutes(router gin.IRouter, service FanAuthService, cookieConfig AuthCookieConfig) {
	if service == nil {
		return
	}

	authGroup := router.Group("/api/fan/auth")
	authGroup.POST("/sign-in/challenges", func(c *gin.Context) {
		handleAuthChallengeIssue(c, "fan_auth_sign_in_challenge", "invalid_email", "email is invalid", service.IssueSignInChallenge)
	})
	authGroup.POST("/sign-up/challenges", func(c *gin.Context) {
		handleAuthChallengeIssue(c, "fan_auth_sign_up_challenge", "invalid_email", "email is invalid", service.IssueSignUpChallenge)
	})
	authGroup.POST("/sign-in/session", func(c *gin.Context) {
		handleAuthSessionStart(c, "fan_auth_sign_in_session", service.StartSignInSession, cookieConfig)
	})
	authGroup.POST("/sign-up/session", func(c *gin.Context) {
		handleAuthSignUpSessionStart(c, "fan_auth_sign_up_session", service.StartSignUpSession, cookieConfig)
	})
	authGroup.DELETE("/session", func(c *gin.Context) {
		rawSessionToken, _ := c.Cookie(auth.SessionCookieName)
		if err := service.Logout(c.Request.Context(), rawSessionToken); err != nil {
			writeAuthError(c, http.StatusInternalServerError, "internal_error", "auth request could not be completed", "fan_auth_logout")
			return
		}

		clearSessionCookie(c, cookieConfig)
		c.Status(http.StatusNoContent)
	})
}

func handleAuthChallengeIssue(
	c *gin.Context,
	requestScope string,
	invalidCode string,
	invalidMessage string,
	issue func(context.Context, string) (auth.IssuedChallenge, error),
) {
	var request authChallengeRequest
	if !decodeAuthJSON(c, &request, invalidCode, invalidMessage, requestScope) {
		return
	}

	challenge, err := issue(c.Request.Context(), request.Email)
	if err != nil {
		if writeMappedAuthError(c, err, requestScope) {
			return
		}

		writeAuthError(c, http.StatusInternalServerError, "internal_error", "auth request could not be completed", requestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[authChallengeResponseData]{
		Data: &authChallengeResponseData{
			ChallengeToken: challenge.Token,
			ExpiresAt:      challenge.ExpiresAt.Format(time.RFC3339),
		},
		Meta: responseMeta{
			Page:      nil,
			RequestID: newRequestID(requestScope),
		},
		Error: nil,
	})
}

func handleAuthSessionStart(
	c *gin.Context,
	requestScope string,
	start func(context.Context, string, string) (auth.AuthenticatedSession, error),
	cookieConfig AuthCookieConfig,
) {
	var request authSessionRequest
	if !decodeAuthJSON(c, &request, "invalid_challenge", "challenge is invalid", requestScope) {
		return
	}

	session, err := start(c.Request.Context(), request.Email, request.ChallengeToken)
	if err != nil {
		if writeMappedAuthError(c, err, requestScope) {
			return
		}

		writeAuthError(c, http.StatusInternalServerError, "internal_error", "auth request could not be completed", requestScope)
		return
	}

	setSessionCookie(c, session.Token, session.ExpiresAt, cookieConfig)
	c.Status(http.StatusNoContent)
}

func handleAuthSignUpSessionStart(
	c *gin.Context,
	requestScope string,
	start func(context.Context, string, string, string, string) (auth.AuthenticatedSession, error),
	cookieConfig AuthCookieConfig,
) {
	var request authSignUpSessionRequest
	if !decodeAuthJSON(c, &request, "invalid_challenge", "challenge is invalid", requestScope) {
		return
	}

	session, err := start(
		c.Request.Context(),
		request.Email,
		request.ChallengeToken,
		request.DisplayName,
		request.Handle,
	)
	if err != nil {
		if writeMappedAuthError(c, err, requestScope) {
			return
		}

		writeAuthError(c, http.StatusInternalServerError, "internal_error", "auth request could not be completed", requestScope)
		return
	}

	setSessionCookie(c, session.Token, session.ExpiresAt, cookieConfig)
	c.Status(http.StatusNoContent)
}

func decodeAuthJSON[T any](c *gin.Context, target *T, invalidCode string, invalidMessage string, requestScope string) bool {
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(target); err != nil {
		writeAuthError(c, http.StatusBadRequest, invalidCode, invalidMessage, requestScope)
		return false
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != nil && !errors.Is(err, io.EOF) {
		writeAuthError(c, http.StatusBadRequest, invalidCode, invalidMessage, requestScope)
		return false
	}
	if len(extra) > 0 {
		writeAuthError(c, http.StatusBadRequest, invalidCode, invalidMessage, requestScope)
		return false
	}

	return true
}

func writeMappedAuthError(c *gin.Context, err error, requestScope string) bool {
	switch {
	case errors.Is(err, auth.ErrInvalidEmail):
		writeAuthError(c, http.StatusBadRequest, "invalid_email", "email is invalid", requestScope)
	case errors.Is(err, auth.ErrInvalidDisplayName):
		writeAuthError(c, http.StatusBadRequest, "invalid_display_name", "display name is invalid", requestScope)
	case errors.Is(err, auth.ErrEmailNotFound):
		writeAuthError(c, http.StatusNotFound, "email_not_found", "email was not found", requestScope)
	case errors.Is(err, auth.ErrEmailAlreadyRegistered):
		writeAuthError(c, http.StatusConflict, "email_already_registered", "email is already registered", requestScope)
	case errors.Is(err, auth.ErrInvalidChallenge):
		writeAuthError(c, http.StatusBadRequest, "invalid_challenge", "challenge is invalid", requestScope)
	case errors.Is(err, auth.ErrInvalidHandle):
		writeAuthError(c, http.StatusBadRequest, "invalid_handle", "handle is invalid", requestScope)
	case errors.Is(err, auth.ErrHandleAlreadyTaken):
		writeAuthError(c, http.StatusConflict, "handle_already_taken", "handle is already taken", requestScope)
	default:
		return false
	}

	return true
}

func writeAuthError(c *gin.Context, status int, code string, message string, requestScope string) {
	c.JSON(status, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			Page:      nil,
			RequestID: newRequestID(requestScope),
		},
		Error: &responseError{
			Code:    code,
			Message: message,
		},
	})
}

func setSessionCookie(c *gin.Context, rawSessionToken string, expiresAt time.Time, config AuthCookieConfig) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 1 {
		maxAge = 1
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.SessionCookieName, rawSessionToken, maxAge, "/", "", config.Secure, true)
}

func clearSessionCookie(c *gin.Context, config AuthCookieConfig) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.SessionCookieName, "", -1, "/", "", config.Secure, true)
}
