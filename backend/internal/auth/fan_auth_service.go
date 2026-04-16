package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/smithy-go"
	"github.com/google/uuid"
)

const (
	fanAuthStartCooldownTTL = time.Minute
	signUpDraftTTL          = 24 * time.Hour
)

var (
	// ErrAuthenticationRequired は current authenticated session が必要なことを表します。
	ErrAuthenticationRequired = errors.New("authentication is required")
	// ErrConfirmationCodeExpired は confirmation code が期限切れなことを表します。
	ErrConfirmationCodeExpired = errors.New("confirmation code expired")
	// ErrInvalidConfirmationCode は confirmation code が不正なことを表します。
	ErrInvalidConfirmationCode = errors.New("confirmation code is invalid")
	// ErrInvalidCredentials は email/password が不正なことを表します。
	ErrInvalidCredentials = errors.New("credentials are invalid")
	// ErrInvalidPassword は password 入力が空、または request boundary を満たさないことを表します。
	ErrInvalidPassword = errors.New("password is invalid")
	// ErrPasswordPolicyViolation は provider password policy を満たさないことを表します。
	ErrPasswordPolicyViolation = errors.New("password policy was violated")
	// ErrConfirmationRequired は sign in 前に email confirmation が必要なことを表します。
	ErrConfirmationRequired = errors.New("confirmation is required")
	// ErrRateLimited は retry guardrail または provider throttle に到達したことを表します。
	ErrRateLimited = errors.New("rate limited")
)

// Next steps for accepted auth flows.
const (
	FanAuthNextStepConfirmPasswordReset = "confirm_password_reset"
	FanAuthNextStepConfirmSignUp        = "confirm_sign_up"
)

// FanAuthAcceptedStep は accepted response の data payload を表します。
type FanAuthAcceptedStep struct {
	DeliveryDestinationHint *string
	NextStep                string
}

type fanAuthProvider interface {
	ConfirmPasswordReset(ctx context.Context, email string, confirmationCode string, newPassword string) error
	ConfirmSignUp(ctx context.Context, email string, confirmationCode string) error
	ResendSignUpCode(ctx context.Context, email string) error
	SignIn(ctx context.Context, email string, password string) (CognitoSessionInput, error)
	SignUp(ctx context.Context, email string, password string) error
	StartPasswordReset(ctx context.Context, email string) error
}

type fanAuthSessionManager interface {
	StartSession(ctx context.Context, input CognitoSessionInput) (AuthenticatedSession, error)
	StartSignUpSession(ctx context.Context, input CognitoSessionInput, displayName string, handle string) (AuthenticatedSession, error)
}

type fanAuthRepository interface {
	GetIdentityByProviderAndSubject(ctx context.Context, provider string, providerSubject string) (Identity, error)
	GetIdentityByNormalizedEmail(ctx context.Context, emailNormalized string) (Identity, error)
	GetPreferredEmailByUserID(ctx context.Context, userID uuid.UUID) (string, error)
	HandleExists(ctx context.Context, handle string) (bool, error)
	RevokeActiveSessionByTokenHash(ctx context.Context, sessionTokenHash string, revokedAt time.Time) (SessionRecord, error)
	UpdateActiveModeByTokenHash(ctx context.Context, sessionTokenHash string, activeMode ActiveMode) (SessionRecord, error)
}

type signUpDraftStore interface {
	DeleteDraft(ctx context.Context, email string) error
	GetDraft(ctx context.Context, email string) (SignUpDraft, error)
	SaveDraft(ctx context.Context, email string, draft SignUpDraft, ttl time.Duration) error
}

type authCooldownStore interface {
	Release(ctx context.Context, key string) error
	TryActivate(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

type fanAuthViewerReader interface {
	ReadCurrentViewer(ctx context.Context, rawSessionToken string) (Bootstrap, error)
}

// FanAuthService は Cognito-backed fan auth flow をまとめます。
type FanAuthService struct {
	cooldownStore  authCooldownStore
	draftStore     signUpDraftStore
	now            func() time.Time
	provider       fanAuthProvider
	repository     fanAuthRepository
	sessionManager fanAuthSessionManager
	viewerReader   fanAuthViewerReader
}

// NewFanAuthService は fan auth service を構築します。
func NewFanAuthService(
	provider fanAuthProvider,
	sessionManager fanAuthSessionManager,
	repository fanAuthRepository,
	viewerReader fanAuthViewerReader,
	draftStore signUpDraftStore,
	cooldownStore authCooldownStore,
) *FanAuthService {
	return &FanAuthService{
		cooldownStore:  cooldownStore,
		draftStore:     draftStore,
		now:            time.Now,
		provider:       provider,
		repository:     repository,
		sessionManager: sessionManager,
		viewerReader:   viewerReader,
	}
}

// SignIn は email/password で sign in し app session を返します。
func (s *FanAuthService) SignIn(ctx context.Context, email string, password string) (AuthenticatedSession, error) {
	if err := s.validateRuntime(); err != nil {
		return AuthenticatedSession{}, err
	}

	normalizedEmail, resolvedPassword, err := normalizeCredentials(email, password)
	if err != nil {
		return AuthenticatedSession{}, err
	}

	principal, err := s.provider.SignIn(ctx, normalizedEmail, resolvedPassword)
	if err != nil {
		return AuthenticatedSession{}, mapSignInError(err)
	}

	draft, draftErr := s.readDraft(ctx, normalizedEmail)
	switch {
	case draftErr == nil:
		session, err := s.sessionManager.StartSignUpSession(ctx, principal, draft.DisplayName, draft.Handle)
		if err != nil {
			return AuthenticatedSession{}, fmt.Errorf("start sign in session with sign up draft: %w", err)
		}
		_ = s.draftStore.DeleteDraft(ctx, normalizedEmail)
		return session, nil
	case !errors.Is(draftErr, ErrSignUpDraftNotFound):
		if _, err := s.repository.GetIdentityByProviderAndSubject(ctx, identityProviderCognito, principal.Subject); err == nil {
			break
		} else if !errors.Is(err, ErrIdentityNotFound) {
			return AuthenticatedSession{}, fmt.Errorf("resolve sign in fallback identity subject=%s: %w", principal.Subject, err)
		}

		if _, err := s.repository.GetIdentityByNormalizedEmail(ctx, normalizedEmail); err != nil {
			if errors.Is(err, ErrIdentityNotFound) {
				return AuthenticatedSession{}, draftErr
			}
			return AuthenticatedSession{}, fmt.Errorf("resolve sign in fallback identity email=%s: %w", normalizedEmail, err)
		}
	}

	session, err := s.sessionManager.StartSession(ctx, principal)
	if err != nil {
		return AuthenticatedSession{}, err
	}

	return session, nil
}

// StartSignUp は sign up を accepted state まで進めます。
func (s *FanAuthService) StartSignUp(ctx context.Context, email string, displayName string, handle string, password string) (FanAuthAcceptedStep, error) {
	if err := s.validateRuntime(); err != nil {
		return FanAuthAcceptedStep{}, err
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return FanAuthAcceptedStep{}, err
	}
	normalizedDisplayName, normalizedHandle, err := normalizeSignUpProfileInput(displayName, handle)
	if err != nil {
		return FanAuthAcceptedStep{}, err
	}
	resolvedPassword, err := normalizePassword(password)
	if err != nil {
		return FanAuthAcceptedStep{}, err
	}

	draft, err := s.readDraft(ctx, normalizedEmail)
	hasExistingDraft := err == nil
	switch {
	case hasExistingDraft:
		normalizedDisplayName = draft.DisplayName
		normalizedHandle = draft.Handle
		resolvedPassword = draft.Password
	case !errors.Is(err, ErrSignUpDraftNotFound):
		return FanAuthAcceptedStep{}, err
	}

	if !hasExistingDraft {
		handleExists, err := s.repository.HandleExists(ctx, normalizedHandle)
		if err != nil {
			return FanAuthAcceptedStep{}, fmt.Errorf("check sign up handle email=%s handle=%s: %w", normalizedEmail, normalizedHandle, err)
		}
		if handleExists {
			return FanAuthAcceptedStep{}, ErrHandleAlreadyTaken
		}

		draft = SignUpDraft{
			DisplayName: normalizedDisplayName,
			Email:       normalizedEmail,
			Handle:      normalizedHandle,
			Password:    resolvedPassword,
		}
	}
	draft.ExpiresAt = s.now().UTC().Add(signUpDraftTTL)

	cooldownKey := signUpCooldownKey(normalizedEmail)
	cooldownActivated, err := s.cooldownStore.TryActivate(ctx, cooldownKey, fanAuthStartCooldownTTL)
	if err != nil {
		return FanAuthAcceptedStep{}, err
	}
	if !cooldownActivated {
		if err := s.draftStore.SaveDraft(ctx, normalizedEmail, draft, signUpDraftTTL); err != nil {
			return FanAuthAcceptedStep{}, err
		}
		return acceptedSignUpStep(normalizedEmail), nil
	}
	if err := s.draftStore.SaveDraft(ctx, normalizedEmail, draft, signUpDraftTTL); err != nil {
		_ = s.cooldownStore.Release(ctx, cooldownKey)
		return FanAuthAcceptedStep{}, err
	}

	_, err = s.startRemoteSignUp(ctx, normalizedEmail, resolvedPassword)
	if err != nil {
		if !hasExistingDraft {
			_ = s.draftStore.DeleteDraft(ctx, normalizedEmail)
		}
		_ = s.cooldownStore.Release(ctx, cooldownKey)
		return FanAuthAcceptedStep{}, err
	}

	return acceptedSignUpStep(normalizedEmail), nil
}

// ConfirmSignUp は confirmation code を消費し session を開始します。
func (s *FanAuthService) ConfirmSignUp(ctx context.Context, email string, confirmationCode string) (AuthenticatedSession, error) {
	if err := s.validateRuntime(); err != nil {
		return AuthenticatedSession{}, err
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return AuthenticatedSession{}, err
	}
	normalizedCode, err := normalizeConfirmationCode(confirmationCode)
	if err != nil {
		return AuthenticatedSession{}, err
	}

	draft, err := s.readDraft(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, ErrSignUpDraftNotFound) {
			return AuthenticatedSession{}, ErrConfirmationCodeExpired
		}
		return AuthenticatedSession{}, err
	}

	if err := s.provider.ConfirmSignUp(ctx, normalizedEmail, normalizedCode); err != nil {
		return AuthenticatedSession{}, mapConfirmSignUpError(err)
	}

	principal, err := s.provider.SignIn(ctx, normalizedEmail, draft.Password)
	if err != nil {
		if isRateLimitError(err) {
			return AuthenticatedSession{}, ErrRateLimited
		}
		return AuthenticatedSession{}, fmt.Errorf("sign in after sign up confirmation email=%s: %w", normalizedEmail, err)
	}

	session, err := s.sessionManager.StartSignUpSession(ctx, principal, draft.DisplayName, draft.Handle)
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("start sign up session after confirmation email=%s: %w", normalizedEmail, err)
	}
	_ = s.draftStore.DeleteDraft(ctx, normalizedEmail)

	return session, nil
}

// StartPasswordReset は password reset code delivery を開始します。
func (s *FanAuthService) StartPasswordReset(ctx context.Context, email string) (FanAuthAcceptedStep, error) {
	if err := s.validateRuntime(); err != nil {
		return FanAuthAcceptedStep{}, err
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return FanAuthAcceptedStep{}, err
	}

	cooldownKey := passwordResetCooldownKey(normalizedEmail)
	activated, err := s.cooldownStore.TryActivate(ctx, cooldownKey, fanAuthStartCooldownTTL)
	if err != nil {
		return FanAuthAcceptedStep{}, err
	}
	if !activated {
		return acceptedPasswordResetStep(normalizedEmail), nil
	}

	if err := s.provider.StartPasswordReset(ctx, normalizedEmail); err != nil {
		if accepted, mappedErr := mapStartPasswordResetError(err); mappedErr != nil {
			_ = s.cooldownStore.Release(ctx, cooldownKey)
			return FanAuthAcceptedStep{}, mappedErr
		} else if !accepted {
			_ = s.cooldownStore.Release(ctx, cooldownKey)
			return FanAuthAcceptedStep{}, fmt.Errorf("start password reset email=%s: %w", normalizedEmail, err)
		}
	}

	return acceptedPasswordResetStep(normalizedEmail), nil
}

// ConfirmPasswordReset は new password を確定します。
func (s *FanAuthService) ConfirmPasswordReset(ctx context.Context, email string, confirmationCode string, newPassword string) error {
	if err := s.validateRuntime(); err != nil {
		return err
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return err
	}
	normalizedCode, err := normalizeConfirmationCode(confirmationCode)
	if err != nil {
		return err
	}
	resolvedPassword, err := normalizePassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.provider.ConfirmPasswordReset(ctx, normalizedEmail, normalizedCode, resolvedPassword); err != nil {
		return mapConfirmPasswordResetError(err)
	}

	return nil
}

// ReAuthenticate は current session が存在する viewer に対して fresh auth を更新します。
func (s *FanAuthService) ReAuthenticate(ctx context.Context, rawSessionToken string, password string) (AuthenticatedSession, error) {
	if err := s.validateRuntime(); err != nil {
		return AuthenticatedSession{}, err
	}

	resolvedPassword, err := normalizePassword(password)
	if err != nil {
		return AuthenticatedSession{}, err
	}

	trimmedToken := strings.TrimSpace(rawSessionToken)
	if trimmedToken == "" {
		return AuthenticatedSession{}, ErrAuthenticationRequired
	}

	bootstrap, err := s.viewerReader.ReadCurrentViewer(ctx, trimmedToken)
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("read current viewer for re-auth: %w", err)
	}
	if bootstrap.CurrentViewer == nil {
		return AuthenticatedSession{}, ErrAuthenticationRequired
	}

	email, err := s.repository.GetPreferredEmailByUserID(ctx, bootstrap.CurrentViewer.ID)
	if err != nil {
		if errors.Is(err, ErrIdentityNotFound) {
			return AuthenticatedSession{}, ErrAuthenticationRequired
		}
		return AuthenticatedSession{}, fmt.Errorf("resolve re-auth email user=%s: %w", bootstrap.CurrentViewer.ID, err)
	}

	principal, err := s.provider.SignIn(ctx, email, resolvedPassword)
	if err != nil {
		return AuthenticatedSession{}, mapSignInError(err)
	}

	session, err := s.sessionManager.StartSession(ctx, principal)
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("start re-auth session user=%s: %w", bootstrap.CurrentViewer.ID, err)
	}
	if bootstrap.CurrentViewer.ActiveMode != ActiveModeFan {
		if _, err := s.repository.UpdateActiveModeByTokenHash(
			ctx,
			HashSessionToken(session.Token),
			bootstrap.CurrentViewer.ActiveMode,
		); err != nil {
			return AuthenticatedSession{}, fmt.Errorf(
				"preserve re-auth active mode user=%s mode=%s: %w",
				bootstrap.CurrentViewer.ID,
				bootstrap.CurrentViewer.ActiveMode,
				err,
			)
		}
	}

	return session, nil
}

// Logout は current session を revoke します。cookie が空でも成功扱いです。
func (s *FanAuthService) Logout(ctx context.Context, rawSessionToken string) error {
	if s == nil || s.repository == nil {
		return fmt.Errorf("fan auth service is not initialized")
	}

	trimmedToken := strings.TrimSpace(rawSessionToken)
	if trimmedToken == "" {
		return nil
	}

	if _, err := s.repository.RevokeActiveSessionByTokenHash(ctx, HashSessionToken(trimmedToken), s.now().UTC()); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil
		}
		return fmt.Errorf("logout session revoke: %w", err)
	}

	return nil
}

func (s *FanAuthService) validateRuntime() error {
	switch {
	case s == nil:
		return fmt.Errorf("fan auth service is nil")
	case s.provider == nil:
		return fmt.Errorf("fan auth provider is not initialized")
	case s.sessionManager == nil:
		return fmt.Errorf("fan auth session manager is not initialized")
	case s.repository == nil:
		return fmt.Errorf("fan auth repository is not initialized")
	case s.viewerReader == nil:
		return fmt.Errorf("fan auth viewer reader is not initialized")
	case s.draftStore == nil:
		return fmt.Errorf("fan auth draft store is not initialized")
	case s.cooldownStore == nil:
		return fmt.Errorf("fan auth cooldown store is not initialized")
	default:
		return nil
	}
}

func (s *FanAuthService) startRemoteSignUp(ctx context.Context, email string, password string) (bool, error) {
	if err := s.provider.SignUp(ctx, email, password); err != nil {
		switch apiErrorCode(err) {
		case "UsernameExistsException":
			if err := s.provider.ResendSignUpCode(ctx, email); err != nil {
				switch apiErrorCode(err) {
				case "InvalidParameterException", "NotAuthorizedException":
					return true, nil
				default:
					if isRateLimitError(err) {
						return false, ErrRateLimited
					}
					return false, fmt.Errorf("resend sign up code email=%s: %w", email, err)
				}
			}
			return true, nil
		default:
			return false, mapStartSignUpError(err)
		}
	}

	return true, nil
}

func (s *FanAuthService) readDraft(ctx context.Context, email string) (SignUpDraft, error) {
	if s.draftStore == nil {
		return SignUpDraft{}, fmt.Errorf("fan auth draft store is not initialized")
	}

	draft, err := s.draftStore.GetDraft(ctx, email)
	if err != nil {
		return SignUpDraft{}, err
	}

	return draft, nil
}

func acceptedSignUpStep(email string) FanAuthAcceptedStep {
	return FanAuthAcceptedStep{
		DeliveryDestinationHint: maskedEmailPointer(email),
		NextStep:                FanAuthNextStepConfirmSignUp,
	}
}

func acceptedPasswordResetStep(email string) FanAuthAcceptedStep {
	return FanAuthAcceptedStep{
		DeliveryDestinationHint: maskedEmailPointer(email),
		NextStep:                FanAuthNextStepConfirmPasswordReset,
	}
}

func normalizeCredentials(email string, password string) (string, string, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return "", "", err
	}
	normalizedPassword, err := normalizePassword(password)
	if err != nil {
		return "", "", err
	}

	return normalizedEmail, normalizedPassword, nil
}

func normalizePassword(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", ErrInvalidPassword
	}

	return value, nil
}

func normalizeConfirmationCode(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", ErrInvalidConfirmationCode
	}

	return normalized, nil
}

func maskedEmailPointer(email string) *string {
	trimmed := strings.TrimSpace(email)
	parts := strings.Split(trimmed, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil
	}

	masked := parts[0][:1] + "***@" + parts[1]
	return &masked
}

func signUpCooldownKey(email string) string {
	return "sign_up:" + strings.TrimSpace(email)
}

func passwordResetCooldownKey(email string) string {
	return "password_reset:" + strings.TrimSpace(email)
}

func mapSignInError(err error) error {
	switch apiErrorCode(err) {
	case "NotAuthorizedException", "PasswordResetRequiredException":
		return ErrInvalidCredentials
	case "UserNotConfirmedException":
		return ErrConfirmationRequired
	default:
		if isRateLimitError(err) {
			return ErrRateLimited
		}
		return fmt.Errorf("sign in with cognito: %w", err)
	}
}

func mapStartSignUpError(err error) error {
	switch apiErrorCode(err) {
	case "InvalidPasswordException":
		return ErrPasswordPolicyViolation
	default:
		if isRateLimitError(err) {
			return ErrRateLimited
		}
		return fmt.Errorf("sign up with cognito: %w", err)
	}
}

func mapConfirmSignUpError(err error) error {
	switch apiErrorCode(err) {
	case "CodeMismatchException", "NotAuthorizedException", "AliasExistsException", "InvalidParameterException":
		return ErrInvalidConfirmationCode
	case "ExpiredCodeException":
		return ErrConfirmationCodeExpired
	default:
		if isRateLimitError(err) {
			return ErrRateLimited
		}
		return fmt.Errorf("confirm sign up with cognito: %w", err)
	}
}

func mapStartPasswordResetError(err error) (bool, error) {
	switch apiErrorCode(err) {
	case "InvalidParameterException", "UserNotFoundException", "NotAuthorizedException":
		return true, nil
	default:
		if isRateLimitError(err) {
			return false, ErrRateLimited
		}
		return false, fmt.Errorf("start password reset with cognito: %w", err)
	}
}

func mapConfirmPasswordResetError(err error) error {
	switch apiErrorCode(err) {
	case "CodeMismatchException", "NotAuthorizedException", "InvalidParameterException":
		return ErrInvalidConfirmationCode
	case "ExpiredCodeException":
		return ErrConfirmationCodeExpired
	case "InvalidPasswordException":
		return ErrPasswordPolicyViolation
	default:
		if isRateLimitError(err) {
			return ErrRateLimited
		}
		return fmt.Errorf("confirm password reset with cognito: %w", err)
	}
}

func apiErrorCode(err error) string {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode()
	}

	return ""
}

func isRateLimitError(err error) bool {
	switch apiErrorCode(err) {
	case "LimitExceededException", "TooManyFailedAttemptsException", "TooManyRequestsException":
		return true
	default:
		return false
	}
}
