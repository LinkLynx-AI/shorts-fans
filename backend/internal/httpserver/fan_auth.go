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
	ConfirmPasswordReset(ctx context.Context, email string, confirmationCode string, newPassword string) error
	ConfirmSignUp(ctx context.Context, email string, confirmationCode string) (auth.AuthenticatedSession, error)
	Logout(ctx context.Context, rawSessionToken string) error
	ReAuthenticate(ctx context.Context, rawSessionToken string, password string) (auth.AuthenticatedSession, error)
	SignIn(ctx context.Context, email string, password string) (auth.AuthenticatedSession, error)
	StartPasswordReset(ctx context.Context, email string) (auth.FanAuthAcceptedStep, error)
	StartSignUp(ctx context.Context, email string, displayName string, handle string, password string) (auth.FanAuthAcceptedStep, error)
}

// AuthCookieConfig は auth session cookie の transport 設定です。
type AuthCookieConfig struct {
	Secure bool
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signUpRequest struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Handle      string `json:"handle"`
	Password    string `json:"password"`
}

type confirmationRequest struct {
	ConfirmationCode string `json:"confirmationCode"`
	Email            string `json:"email"`
}

type passwordResetConfirmRequest struct {
	ConfirmationCode string `json:"confirmationCode"`
	Email            string `json:"email"`
	NewPassword      string `json:"newPassword"`
}

type reAuthRequest struct {
	Password string `json:"password"`
}

type acceptedStepResponseData struct {
	DeliveryDestinationHint *string `json:"deliveryDestinationHint"`
	NextStep                string  `json:"nextStep"`
}

// registerFanAuthRoutes は fan auth transport を router に登録します。
func registerFanAuthRoutes(router gin.IRouter, service FanAuthService, cookieConfig AuthCookieConfig) {
	if service == nil {
		return
	}

	authGroup := router.Group("/api/fan/auth")
	authGroup.POST("/sign-in", func(c *gin.Context) {
		handleSignIn(c, service, cookieConfig)
	})
	authGroup.POST("/sign-up", func(c *gin.Context) {
		handleSignUp(c, service)
	})
	authGroup.POST("/sign-up/confirm", func(c *gin.Context) {
		handleSignUpConfirm(c, service, cookieConfig)
	})
	authGroup.POST("/password-reset", func(c *gin.Context) {
		handlePasswordReset(c, service)
	})
	authGroup.POST("/password-reset/confirm", func(c *gin.Context) {
		handlePasswordResetConfirm(c, service)
	})
	authGroup.POST("/re-auth", func(c *gin.Context) {
		handleReAuth(c, service, cookieConfig)
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

func handleSignIn(c *gin.Context, service FanAuthService, cookieConfig AuthCookieConfig) {
	var request signInRequest
	if !decodeAuthJSON(c, &request, "invalid_email", "email is invalid", "fan_auth_sign_in") {
		return
	}

	session, err := service.SignIn(c.Request.Context(), request.Email, request.Password)
	if err != nil {
		if writeMappedAuthError(c, err, "fan_auth_sign_in") {
			return
		}
		writeAuthError(c, http.StatusInternalServerError, "internal_error", "auth request could not be completed", "fan_auth_sign_in")
		return
	}

	setSessionCookie(c, session.Token, session.ExpiresAt, cookieConfig)
	c.Status(http.StatusNoContent)
}

func handleSignUp(c *gin.Context, service FanAuthService) {
	var request signUpRequest
	if !decodeAuthJSON(c, &request, "invalid_email", "email is invalid", "fan_auth_sign_up") {
		return
	}

	step, err := service.StartSignUp(c.Request.Context(), request.Email, request.DisplayName, request.Handle, request.Password)
	if err != nil {
		if writeMappedAuthError(c, err, "fan_auth_sign_up") {
			return
		}
		writeAuthError(c, http.StatusInternalServerError, "internal_error", "auth request could not be completed", "fan_auth_sign_up")
		return
	}

	writeAcceptedAuthStep(c, "fan_auth_sign_up_accepted", step)
}

func handleSignUpConfirm(c *gin.Context, service FanAuthService, cookieConfig AuthCookieConfig) {
	var request confirmationRequest
	if !decodeAuthJSON(c, &request, "invalid_confirmation_code", "confirmation code is invalid", "fan_auth_sign_up_confirm") {
		return
	}

	session, err := service.ConfirmSignUp(c.Request.Context(), request.Email, request.ConfirmationCode)
	if err != nil {
		if writeMappedAuthError(c, err, "fan_auth_sign_up_confirm") {
			return
		}
		writeAuthError(c, http.StatusInternalServerError, "internal_error", "auth request could not be completed", "fan_auth_sign_up_confirm")
		return
	}

	setSessionCookie(c, session.Token, session.ExpiresAt, cookieConfig)
	c.Status(http.StatusNoContent)
}

func handlePasswordReset(c *gin.Context, service FanAuthService) {
	var request struct {
		Email string `json:"email"`
	}
	if !decodeAuthJSON(c, &request, "invalid_email", "email is invalid", "fan_auth_password_reset") {
		return
	}

	step, err := service.StartPasswordReset(c.Request.Context(), request.Email)
	if err != nil {
		if writeMappedAuthError(c, err, "fan_auth_password_reset") {
			return
		}
		writeAuthError(c, http.StatusInternalServerError, "internal_error", "auth request could not be completed", "fan_auth_password_reset")
		return
	}

	writeAcceptedAuthStep(c, "fan_auth_password_reset_accepted", step)
}

func handlePasswordResetConfirm(c *gin.Context, service FanAuthService) {
	var request passwordResetConfirmRequest
	if !decodeAuthJSON(c, &request, "invalid_confirmation_code", "confirmation code is invalid", "fan_auth_password_reset_confirm") {
		return
	}

	if err := service.ConfirmPasswordReset(c.Request.Context(), request.Email, request.ConfirmationCode, request.NewPassword); err != nil {
		if writeMappedAuthError(c, err, "fan_auth_password_reset_confirm") {
			return
		}
		writeAuthError(c, http.StatusInternalServerError, "internal_error", "auth request could not be completed", "fan_auth_password_reset_confirm")
		return
	}

	c.Status(http.StatusNoContent)
}

func handleReAuth(c *gin.Context, service FanAuthService, cookieConfig AuthCookieConfig) {
	var request reAuthRequest
	if !decodeAuthJSON(c, &request, "invalid_password", "password is invalid", "fan_auth_re_auth") {
		return
	}

	rawSessionToken, _ := c.Cookie(auth.SessionCookieName)
	session, err := service.ReAuthenticate(c.Request.Context(), rawSessionToken, request.Password)
	if err != nil {
		if writeMappedAuthError(c, err, "fan_auth_re_auth") {
			return
		}
		writeAuthError(c, http.StatusInternalServerError, "internal_error", "auth request could not be completed", "fan_auth_re_auth")
		return
	}

	setSessionCookie(c, session.Token, session.ExpiresAt, cookieConfig)
	c.Status(http.StatusNoContent)
}

func writeAcceptedAuthStep(c *gin.Context, requestScope string, step auth.FanAuthAcceptedStep) {
	c.JSON(http.StatusOK, responseEnvelope[acceptedStepResponseData]{
		Data: &acceptedStepResponseData{
			DeliveryDestinationHint: step.DeliveryDestinationHint,
			NextStep:                step.NextStep,
		},
		Meta: responseMeta{
			Page:      nil,
			RequestID: newRequestID(requestScope),
		},
		Error: nil,
	})
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
	case errors.Is(err, auth.ErrInvalidHandle):
		writeAuthError(c, http.StatusBadRequest, "invalid_handle", "handle is invalid", requestScope)
	case errors.Is(err, auth.ErrInvalidPassword):
		writeAuthError(c, http.StatusBadRequest, "invalid_password", "password is invalid", requestScope)
	case errors.Is(err, auth.ErrInvalidConfirmationCode):
		writeAuthError(c, http.StatusBadRequest, "invalid_confirmation_code", "confirmation code is invalid", requestScope)
	case errors.Is(err, auth.ErrConfirmationCodeExpired):
		writeAuthError(c, http.StatusBadRequest, "confirmation_code_expired", "confirmation code has expired", requestScope)
	case errors.Is(err, auth.ErrPasswordPolicyViolation):
		writeAuthError(c, http.StatusBadRequest, "password_policy_violation", "password does not satisfy the policy", requestScope)
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeAuthError(c, http.StatusUnauthorized, "invalid_credentials", "email or password is invalid", requestScope)
	case errors.Is(err, auth.ErrAuthenticationRequired), errors.Is(err, auth.ErrSessionNotFound):
		writeAuthError(c, http.StatusUnauthorized, "auth_required", "authentication is required", requestScope)
	case errors.Is(err, auth.ErrConfirmationRequired), errors.Is(err, auth.ErrEmailVerificationRequired):
		writeAuthError(c, http.StatusForbidden, "confirmation_required", "email confirmation is required", requestScope)
	case errors.Is(err, auth.ErrHandleAlreadyTaken):
		writeAuthError(c, http.StatusConflict, "handle_already_taken", "handle is already taken", requestScope)
	case errors.Is(err, auth.ErrRateLimited):
		writeAuthError(c, http.StatusTooManyRequests, "rate_limited", "auth request was rate limited", requestScope)
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
