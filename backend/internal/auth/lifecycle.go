package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	challengeTokenByteLength = 32
	sessionTokenByteLength   = 32
	defaultChallengeTTL      = 10 * time.Minute
	defaultSessionTTL        = 30 * 24 * time.Hour
)

var (
	// ErrEmailNotFound は sign-in 対象の email identity が存在しないことを表します。
	ErrEmailNotFound = errors.New("email が見つかりません")
	// ErrEmailAlreadyRegistered は sign-up 対象の email が既に使われていることを表します。
	ErrEmailAlreadyRegistered = errors.New("email は既に登録されています")
	// ErrInvalidChallenge は challenge token が不正、期限切れ、または消費済みであることを表します。
	ErrInvalidChallenge = errors.New("challenge が不正です")
)

// IssuedChallenge は client へ返す challenge token を表します。
type IssuedChallenge struct {
	Token     string
	ExpiresAt time.Time
}

// AuthenticatedSession は発行済み session token を表します。
type AuthenticatedSession struct {
	Token     string
	ExpiresAt time.Time
}

type lifecycleRepository interface {
	GetIdentityByEmail(ctx context.Context, emailNormalized string) (Identity, error)
	CreateLoginChallenge(ctx context.Context, input CreateLoginChallengeInput) (Challenge, error)
	GetLatestPendingLoginChallengeByEmail(ctx context.Context, emailNormalized string) (Challenge, error)
	IncrementLoginChallengeAttemptCount(ctx context.Context, id uuid.UUID) (Challenge, error)
	ConsumeLoginChallenge(ctx context.Context, id uuid.UUID, consumedAt time.Time) (Challenge, error)
	RecordIdentityAuthentication(ctx context.Context, input RecordIdentityAuthenticationInput) (Identity, error)
	CreateSession(ctx context.Context, input CreateSessionInput) (SessionRecord, error)
	CreateUserWithEmailIdentityAndSession(ctx context.Context, input CreateUserWithEmailIdentityAndSessionInput) (SessionRecord, error)
	RevokeActiveSessionByTokenHash(ctx context.Context, sessionTokenHash string, revokedAt time.Time) (SessionRecord, error)
}

// Lifecycle は auth transport から使う challenge / session lifecycle を扱います。
type Lifecycle struct {
	repository        lifecycleRepository
	now               func() time.Time
	newChallengeToken func() (string, error)
	newSessionToken   func() (string, error)
}

// NewLifecycle は auth lifecycle service を構築します。
func NewLifecycle(repository lifecycleRepository) *Lifecycle {
	return &Lifecycle{
		repository: repository,
		now:        time.Now,
		newChallengeToken: func() (string, error) {
			return generateOpaqueToken(challengeTokenByteLength)
		},
		newSessionToken: func() (string, error) {
			return generateOpaqueToken(sessionTokenByteLength)
		},
	}
}

// IssueSignInChallenge は既存 identity 向けの sign-in challenge を発行します。
func (l *Lifecycle) IssueSignInChallenge(ctx context.Context, email string) (IssuedChallenge, error) {
	if l == nil || l.repository == nil {
		return IssuedChallenge{}, fmt.Errorf("auth lifecycle が初期化されていません")
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return IssuedChallenge{}, err
	}

	if _, err := l.repository.GetIdentityByEmail(ctx, normalizedEmail); err != nil {
		if errors.Is(err, ErrIdentityNotFound) {
			return IssuedChallenge{}, ErrEmailNotFound
		}

		return IssuedChallenge{}, fmt.Errorf("sign in identity 確認 email=%s: %w", normalizedEmail, err)
	}

	return l.issueChallenge(ctx, normalizedEmail)
}

// IssueSignUpChallenge は未登録 email 向けの sign-up challenge を発行します。
func (l *Lifecycle) IssueSignUpChallenge(ctx context.Context, email string) (IssuedChallenge, error) {
	if l == nil || l.repository == nil {
		return IssuedChallenge{}, fmt.Errorf("auth lifecycle が初期化されていません")
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return IssuedChallenge{}, err
	}

	if _, err := l.repository.GetIdentityByEmail(ctx, normalizedEmail); err == nil {
		return IssuedChallenge{}, ErrEmailAlreadyRegistered
	} else if !errors.Is(err, ErrIdentityNotFound) {
		return IssuedChallenge{}, fmt.Errorf("sign up identity 確認 email=%s: %w", normalizedEmail, err)
	}

	return l.issueChallenge(ctx, normalizedEmail)
}

// StartSignInSession は sign-in challenge を消費して session を開始します。
func (l *Lifecycle) StartSignInSession(ctx context.Context, email string, challengeToken string) (AuthenticatedSession, error) {
	if l == nil || l.repository == nil {
		return AuthenticatedSession{}, fmt.Errorf("auth lifecycle が初期化されていません")
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return AuthenticatedSession{}, err
	}

	identity, err := l.repository.GetIdentityByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, ErrIdentityNotFound) {
			return AuthenticatedSession{}, ErrEmailNotFound
		}

		return AuthenticatedSession{}, fmt.Errorf("sign in identity 取得 email=%s: %w", normalizedEmail, err)
	}

	now := l.now().UTC()
	if err := l.consumeChallenge(ctx, normalizedEmail, challengeToken, now); err != nil {
		return AuthenticatedSession{}, err
	}

	verifiedAt := identity.VerifiedAt
	if verifiedAt == nil {
		verifiedAt = &now
	}

	if _, err := l.repository.RecordIdentityAuthentication(ctx, RecordIdentityAuthenticationInput{
		ID:                  identity.ID,
		EmailNormalized:     normalizedEmail,
		VerifiedAt:          verifiedAt,
		LastAuthenticatedAt: now,
	}); err != nil {
		return AuthenticatedSession{}, fmt.Errorf("sign in 認証記録 email=%s: %w", normalizedEmail, err)
	}
	if l.newSessionToken == nil {
		return AuthenticatedSession{}, fmt.Errorf("session token generator が初期化されていません")
	}

	rawSessionToken, err := l.newSessionToken()
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("sign in session token 生成 email=%s: %w", normalizedEmail, err)
	}

	session, err := l.repository.CreateSession(ctx, CreateSessionInput{
		UserID:           identity.UserID,
		ActiveMode:       ActiveModeFan,
		SessionTokenHash: HashSessionToken(rawSessionToken),
		ExpiresAt:        now.Add(defaultSessionTTL),
	})
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("sign in session 作成 email=%s: %w", normalizedEmail, err)
	}

	return AuthenticatedSession{
		Token:     rawSessionToken,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// StartSignUpSession は sign-up challenge を消費して user / identity / session を開始します。
func (l *Lifecycle) StartSignUpSession(ctx context.Context, email string, challengeToken string) (AuthenticatedSession, error) {
	if l == nil || l.repository == nil {
		return AuthenticatedSession{}, fmt.Errorf("auth lifecycle が初期化されていません")
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return AuthenticatedSession{}, err
	}

	if _, err := l.repository.GetIdentityByEmail(ctx, normalizedEmail); err == nil {
		return AuthenticatedSession{}, ErrEmailAlreadyRegistered
	} else if !errors.Is(err, ErrIdentityNotFound) {
		return AuthenticatedSession{}, fmt.Errorf("sign up identity 確認 email=%s: %w", normalizedEmail, err)
	}

	now := l.now().UTC()
	if err := l.consumeChallenge(ctx, normalizedEmail, challengeToken, now); err != nil {
		return AuthenticatedSession{}, err
	}
	if l.newSessionToken == nil {
		return AuthenticatedSession{}, fmt.Errorf("session token generator が初期化されていません")
	}

	rawSessionToken, err := l.newSessionToken()
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("sign up session token 生成 email=%s: %w", normalizedEmail, err)
	}

	session, err := l.repository.CreateUserWithEmailIdentityAndSession(ctx, CreateUserWithEmailIdentityAndSessionInput{
		EmailNormalized:     normalizedEmail,
		SessionTokenHash:    HashSessionToken(rawSessionToken),
		VerifiedAt:          now,
		LastAuthenticatedAt: now,
		ExpiresAt:           now.Add(defaultSessionTTL),
	})
	if err != nil {
		if errors.Is(err, ErrIdentityAlreadyExists) {
			return AuthenticatedSession{}, ErrEmailAlreadyRegistered
		}

		return AuthenticatedSession{}, fmt.Errorf("sign up session 作成 email=%s: %w", normalizedEmail, err)
	}

	return AuthenticatedSession{
		Token:     rawSessionToken,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// Logout は raw session token に紐づく session を revoke します。
func (l *Lifecycle) Logout(ctx context.Context, rawSessionToken string) error {
	if l == nil || l.repository == nil {
		return fmt.Errorf("auth lifecycle が初期化されていません")
	}

	trimmedToken := strings.TrimSpace(rawSessionToken)
	if trimmedToken == "" {
		return nil
	}

	if _, err := l.repository.RevokeActiveSessionByTokenHash(ctx, HashSessionToken(trimmedToken), l.now().UTC()); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil
		}

		return fmt.Errorf("logout session revoke: %w", err)
	}

	return nil
}

func (l *Lifecycle) issueChallenge(ctx context.Context, emailNormalized string) (IssuedChallenge, error) {
	if l == nil || l.repository == nil || l.newChallengeToken == nil {
		return IssuedChallenge{}, fmt.Errorf("auth lifecycle が初期化されていません")
	}

	rawToken, err := l.newChallengeToken()
	if err != nil {
		return IssuedChallenge{}, fmt.Errorf("challenge token 生成 email=%s: %w", emailNormalized, err)
	}

	expiresAt := l.now().UTC().Add(defaultChallengeTTL)
	challenge, err := l.repository.CreateLoginChallenge(ctx, CreateLoginChallengeInput{
		EmailNormalized:    emailNormalized,
		ChallengeTokenHash: HashChallengeToken(rawToken),
		ExpiresAt:          expiresAt,
	})
	if err != nil {
		return IssuedChallenge{}, fmt.Errorf("challenge 作成 email=%s: %w", emailNormalized, err)
	}

	return IssuedChallenge{
		Token:     rawToken,
		ExpiresAt: challenge.ExpiresAt,
	}, nil
}

func (l *Lifecycle) consumeChallenge(ctx context.Context, emailNormalized string, rawChallengeToken string, consumedAt time.Time) error {
	if l == nil || l.repository == nil {
		return fmt.Errorf("auth lifecycle が初期化されていません")
	}

	trimmedToken := strings.TrimSpace(rawChallengeToken)
	if trimmedToken == "" {
		return ErrInvalidChallenge
	}

	challenge, err := l.repository.GetLatestPendingLoginChallengeByEmail(ctx, emailNormalized)
	if err != nil {
		if errors.Is(err, ErrLoginChallengeNotFound) {
			return ErrInvalidChallenge
		}

		return fmt.Errorf("challenge 取得 email=%s: %w", emailNormalized, err)
	}

	if challenge.ChallengeTokenHash != HashChallengeToken(trimmedToken) {
		if _, err := l.repository.IncrementLoginChallengeAttemptCount(ctx, challenge.ID); err != nil && !errors.Is(err, ErrLoginChallengeNotFound) {
			return fmt.Errorf("challenge attempt 更新 email=%s: %w", emailNormalized, err)
		}

		return ErrInvalidChallenge
	}

	if _, err := l.repository.ConsumeLoginChallenge(ctx, challenge.ID, consumedAt); err != nil {
		if errors.Is(err, ErrLoginChallengeNotFound) {
			return ErrInvalidChallenge
		}

		return fmt.Errorf("challenge consume email=%s: %w", emailNormalized, err)
	}

	return nil
}
