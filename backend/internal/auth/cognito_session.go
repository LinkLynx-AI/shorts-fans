package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CognitoSessionInput は verified Cognito principal から app session を開始する入力です。
type CognitoSessionInput struct {
	Subject         string
	Email           string
	EmailVerified   bool
	AuthenticatedAt time.Time
}

type cognitoSessionRepository interface {
	GetIdentityByProviderAndSubject(ctx context.Context, provider string, providerSubject string) (Identity, error)
	GetIdentityByEmail(ctx context.Context, emailNormalized string) (Identity, error)
	CreateIdentity(ctx context.Context, input CreateIdentityInput) (Identity, error)
	CreateSession(ctx context.Context, input CreateSessionInput) (SessionRecord, error)
	CreateUserWithIdentityAndSession(ctx context.Context, input CreateUserWithIdentityAndSessionInput) (SessionRecord, error)
	RecordIdentityAuthentication(ctx context.Context, input RecordIdentityAuthenticationInput) (Identity, error)
	RefreshSessionRecentAuthenticatedAtByTokenHash(ctx context.Context, sessionTokenHash string, recentAuthenticatedAt time.Time) (SessionRecord, error)
}

// CognitoSessionManager は Cognito identity と internal user / session を接続します。
type CognitoSessionManager struct {
	repository      cognitoSessionRepository
	now             func() time.Time
	newSessionToken func() (string, error)
}

// NewCognitoSessionManager は Cognito principal 向け session manager を構築します。
func NewCognitoSessionManager(repository cognitoSessionRepository) *CognitoSessionManager {
	return &CognitoSessionManager{
		repository: repository,
		now:        time.Now,
		newSessionToken: func() (string, error) {
			return generateOpaqueToken(sessionTokenByteLength)
		},
	}
}

// StartSession は Cognito principal を internal user へ解決し、app session を発行します。
func (m *CognitoSessionManager) StartSession(ctx context.Context, input CognitoSessionInput) (AuthenticatedSession, error) {
	if m == nil || m.repository == nil {
		return AuthenticatedSession{}, fmt.Errorf("cognito session manager が初期化されていません")
	}

	normalized, err := m.normalizeInput(input)
	if err != nil {
		return AuthenticatedSession{}, err
	}

	identity, err := m.repository.GetIdentityByProviderAndSubject(ctx, identityProviderCognito, normalized.subject)
	switch {
	case err == nil:
		if err := m.recordAuthentication(ctx, identity.ID, normalized.email, normalized.authenticatedAt); err != nil {
			return AuthenticatedSession{}, err
		}

		return m.issueSession(ctx, identity.UserID, normalized.authenticatedAt)
	case !errors.Is(err, ErrIdentityNotFound):
		return AuthenticatedSession{}, fmt.Errorf(
			"cognito identity 取得 subject=%s: %w",
			normalized.subject,
			err,
		)
	}

	legacyIdentity, err := m.repository.GetIdentityByEmail(ctx, normalized.email)
	switch {
	case err == nil:
		identity, err = m.ensureCognitoIdentity(ctx, legacyIdentity.UserID, normalized)
		if err != nil {
			return AuthenticatedSession{}, err
		}

		return m.issueSession(ctx, identity.UserID, normalized.authenticatedAt)
	case !errors.Is(err, ErrIdentityNotFound):
		return AuthenticatedSession{}, fmt.Errorf(
			"legacy email identity 取得 email=%s: %w",
			normalized.email,
			err,
		)
	}

	return m.createUserAndSession(ctx, normalized)
}

// RefreshRecentAuthentication は既存 session の recent auth proof を更新します。
func (m *CognitoSessionManager) RefreshRecentAuthentication(
	ctx context.Context,
	rawSessionToken string,
	authenticatedAt time.Time,
) error {
	if m == nil || m.repository == nil {
		return fmt.Errorf("cognito session manager が初期化されていません")
	}

	trimmedToken := strings.TrimSpace(rawSessionToken)
	if trimmedToken == "" {
		return ErrSessionNotFound
	}

	refreshedAt := authenticatedAt
	if refreshedAt.IsZero() {
		refreshedAt = m.now()
	}
	refreshedAt = refreshedAt.UTC()

	if _, err := m.repository.RefreshSessionRecentAuthenticatedAtByTokenHash(
		ctx,
		HashSessionToken(trimmedToken),
		refreshedAt,
	); err != nil {
		return fmt.Errorf("session recent auth 更新: %w", err)
	}

	return nil
}

type normalizedCognitoSessionInput struct {
	subject         string
	email           string
	authenticatedAt time.Time
}

func (m *CognitoSessionManager) normalizeInput(input CognitoSessionInput) (normalizedCognitoSessionInput, error) {
	subject := strings.TrimSpace(input.Subject)
	if subject == "" {
		return normalizedCognitoSessionInput{}, ErrInvalidProviderSubject
	}

	email, err := normalizeEmail(input.Email)
	if err != nil {
		return normalizedCognitoSessionInput{}, err
	}
	if !input.EmailVerified {
		return normalizedCognitoSessionInput{}, ErrEmailVerificationRequired
	}

	authenticatedAt := input.AuthenticatedAt
	if authenticatedAt.IsZero() {
		authenticatedAt = m.now()
	}

	return normalizedCognitoSessionInput{
		subject:         subject,
		email:           email,
		authenticatedAt: authenticatedAt.UTC(),
	}, nil
}

func (m *CognitoSessionManager) ensureCognitoIdentity(
	ctx context.Context,
	userID uuid.UUID,
	input normalizedCognitoSessionInput,
) (Identity, error) {
	email := input.email
	identity, err := m.repository.CreateIdentity(ctx, CreateIdentityInput{
		UserID:              userID,
		Provider:            identityProviderCognito,
		ProviderSubject:     input.subject,
		EmailNormalized:     &email,
		VerifiedAt:          &input.authenticatedAt,
		LastAuthenticatedAt: &input.authenticatedAt,
	})
	if err != nil {
		if !errors.Is(err, ErrIdentityAlreadyExists) {
			return Identity{}, fmt.Errorf(
				"cognito identity 作成 user=%s subject=%s: %w",
				userID,
				input.subject,
				err,
			)
		}

		identity, err = m.repository.GetIdentityByProviderAndSubject(ctx, identityProviderCognito, input.subject)
		if err != nil {
			return Identity{}, fmt.Errorf(
				"cognito identity 再取得 subject=%s: %w",
				input.subject,
				err,
			)
		}
	}

	if err := m.recordAuthentication(ctx, identity.ID, input.email, input.authenticatedAt); err != nil {
		return Identity{}, err
	}

	return identity, nil
}

func (m *CognitoSessionManager) recordAuthentication(
	ctx context.Context,
	identityID uuid.UUID,
	email string,
	authenticatedAt time.Time,
) error {
	if _, err := m.repository.RecordIdentityAuthentication(ctx, RecordIdentityAuthenticationInput{
		ID:                  identityID,
		EmailNormalized:     email,
		VerifiedAt:          &authenticatedAt,
		LastAuthenticatedAt: authenticatedAt,
	}); err != nil {
		return fmt.Errorf("identity 認証記録 id=%s: %w", identityID, err)
	}

	return nil
}

func (m *CognitoSessionManager) createUserAndSession(
	ctx context.Context,
	input normalizedCognitoSessionInput,
) (AuthenticatedSession, error) {
	if m.newSessionToken == nil {
		return AuthenticatedSession{}, fmt.Errorf("session token generator が初期化されていません")
	}

	rawSessionToken, err := m.newSessionToken()
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf(
			"cognito session token 生成 subject=%s: %w",
			input.subject,
			err,
		)
	}

	issuedAt := m.now().UTC()
	email := input.email
	session, err := m.repository.CreateUserWithIdentityAndSession(ctx, CreateUserWithIdentityAndSessionInput{
		Provider:              identityProviderCognito,
		ProviderSubject:       input.subject,
		EmailNormalized:       &email,
		SessionTokenHash:      HashSessionToken(rawSessionToken),
		VerifiedAt:            &input.authenticatedAt,
		LastAuthenticatedAt:   &input.authenticatedAt,
		ExpiresAt:             issuedAt.Add(defaultSessionTTL),
		RecentAuthenticatedAt: input.authenticatedAt,
	})
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf(
			"cognito user/session 作成 subject=%s: %w",
			input.subject,
			err,
		)
	}

	return AuthenticatedSession{
		Token:     rawSessionToken,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func (m *CognitoSessionManager) issueSession(
	ctx context.Context,
	userID uuid.UUID,
	authenticatedAt time.Time,
) (AuthenticatedSession, error) {
	if m.newSessionToken == nil {
		return AuthenticatedSession{}, fmt.Errorf("session token generator が初期化されていません")
	}

	rawSessionToken, err := m.newSessionToken()
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("session token 生成 user=%s: %w", userID, err)
	}

	issuedAt := m.now().UTC()
	session, err := m.repository.CreateSession(ctx, CreateSessionInput{
		UserID:                userID,
		ActiveMode:            ActiveModeFan,
		SessionTokenHash:      HashSessionToken(rawSessionToken),
		ExpiresAt:             issuedAt.Add(defaultSessionTTL),
		RecentAuthenticatedAt: authenticatedAt,
	})
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("session 作成 user=%s: %w", userID, err)
	}

	return AuthenticatedSession{
		Token:     rawSessionToken,
		ExpiresAt: session.ExpiresAt,
	}, nil
}
