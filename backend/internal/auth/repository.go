package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	identityProviderEmail        = "email"
	identityProviderCognito      = "cognito"
	loginChallengePurpose        = "login"
	identityUniqueConstraint     = "auth_identities_provider_provider_subject_key"
	emailUniqueConstraint        = "idx_auth_identities_email_normalized"
	cognitoEmailUniqueConstraint = "idx_auth_identities_cognito_email_normalized"
	userProfileHandleUnique      = "user_profiles_handle_unique_idx"
	emailClaimLockRetryInterval  = 10 * time.Millisecond
)

var (
	// ErrCurrentViewerNotFound は有効な session に紐づく viewer が見つからないことを表します。
	ErrCurrentViewerNotFound = errors.New("current viewer が見つかりません")
	// ErrIdentityNotFound は auth identity が見つからないことを表します。
	ErrIdentityNotFound = errors.New("auth identity が見つかりません")
	// ErrIdentityAlreadyExists は auth identity が既に存在することを表します。
	ErrIdentityAlreadyExists = errors.New("auth identity は既に存在します")
	// ErrLoginChallengeNotFound は有効な login challenge が見つからないことを表します。
	ErrLoginChallengeNotFound = errors.New("login challenge が見つかりません")
	// ErrSessionNotFound は有効な session が見つからないことを表します。
	ErrSessionNotFound = errors.New("auth session が見つかりません")
	// ErrInvalidProviderSubject は provider subject が不正なことを表します。
	ErrInvalidProviderSubject = errors.New("provider subject が不正です")
	// ErrEmailVerificationRequired は verified email が必要なことを表します。
	ErrEmailVerificationRequired = errors.New("email verification が必要です")
)

type queries interface {
	GetCurrentViewerBySessionTokenHash(ctx context.Context, sessionTokenHash string) (sqlc.GetCurrentViewerBySessionTokenHashRow, error)
	TouchAuthSessionLastSeenByTokenHash(ctx context.Context, arg sqlc.TouchAuthSessionLastSeenByTokenHashParams) (sqlc.AppAuthSession, error)
	UpdateActiveAuthSessionModeByTokenHash(ctx context.Context, arg sqlc.UpdateActiveAuthSessionModeByTokenHashParams) (sqlc.AppAuthSession, error)
}

// Repository は auth 関連の永続化操作を包みます。
type Repository struct {
	txBeginner postgres.TxBeginner
	db         sqlc.DBTX
	queries    queries
}

// Identity は domain 向けの auth identity レコードです。
type Identity struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	Provider            string
	ProviderSubject     string
	EmailNormalized     *string
	VerifiedAt          *time.Time
	LastAuthenticatedAt *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Challenge は domain 向けの login challenge レコードです。
type Challenge struct {
	ID                 uuid.UUID
	Provider           string
	ProviderSubject    string
	EmailNormalized    *string
	ChallengeTokenHash string
	Purpose            string
	ExpiresAt          time.Time
	ConsumedAt         *time.Time
	AttemptCount       int32
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// SessionRecord は domain 向けの auth session レコードです。
type SessionRecord struct {
	ID                    uuid.UUID
	UserID                uuid.UUID
	ActiveMode            ActiveMode
	SessionTokenHash      string
	ExpiresAt             time.Time
	RecentAuthenticatedAt time.Time
	LastSeenAt            time.Time
	RevokedAt             *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// CreateIdentityInput は CreateIdentity の入力です。
type CreateIdentityInput struct {
	UserID              uuid.UUID
	Provider            string
	ProviderSubject     string
	EmailNormalized     *string
	VerifiedAt          *time.Time
	LastAuthenticatedAt *time.Time
}

// CreateLoginChallengeInput は CreateLoginChallenge の入力です。
type CreateLoginChallengeInput struct {
	EmailNormalized    string
	ChallengeTokenHash string
	ExpiresAt          time.Time
}

// RecordIdentityAuthenticationInput は RecordIdentityAuthentication の入力です。
type RecordIdentityAuthenticationInput struct {
	ID                  uuid.UUID
	EmailNormalized     string
	VerifiedAt          *time.Time
	LastAuthenticatedAt time.Time
}

// CreateSessionInput は CreateSession の入力です。
type CreateSessionInput struct {
	UserID                uuid.UUID
	ActiveMode            ActiveMode
	SessionTokenHash      string
	ExpiresAt             time.Time
	RecentAuthenticatedAt time.Time
}

// CreateUserWithIdentityAndSessionInput は identity と session の一括作成入力です。
type CreateUserWithIdentityAndSessionInput struct {
	Provider              string
	ProviderSubject       string
	EmailNormalized       *string
	SessionTokenHash      string
	VerifiedAt            *time.Time
	LastAuthenticatedAt   *time.Time
	ExpiresAt             time.Time
	RecentAuthenticatedAt time.Time
}

// CreateUserWithEmailIdentityAndSessionInput は sign-up 完了時の一括作成入力です。
type CreateUserWithEmailIdentityAndSessionInput struct {
	DisplayName           string
	EmailNormalized       string
	Handle                string
	SessionTokenHash      string
	VerifiedAt            time.Time
	LastAuthenticatedAt   time.Time
	ExpiresAt             time.Time
	RecentAuthenticatedAt time.Time
}

// NewRepository は pgxpool ベースの auth repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	repository := &Repository{}
	if pool == nil {
		return repository
	}

	repository.txBeginner = pool
	repository.db = pool
	repository.queries = sqlc.New(pool)

	return repository
}

func newRepository(q queries) *Repository {
	return &Repository{queries: q}
}

// GetIdentityByProviderAndSubject は provider / subject で auth identity を取得します。
func (r *Repository) GetIdentityByProviderAndSubject(ctx context.Context, provider string, providerSubject string) (Identity, error) {
	q, err := r.dbQueries()
	if err != nil {
		return Identity{}, err
	}

	row, err := q.GetAuthIdentityByProviderAndSubject(ctx, sqlc.GetAuthIdentityByProviderAndSubjectParams{
		Provider:        provider,
		ProviderSubject: providerSubject,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Identity{}, fmt.Errorf(
				"auth identity 取得 provider=%s provider_subject=%s: %w",
				provider,
				providerSubject,
				ErrIdentityNotFound,
			)
		}

		return Identity{}, fmt.Errorf(
			"auth identity 取得 provider=%s provider_subject=%s: %w",
			provider,
			providerSubject,
			err,
		)
	}

	identity, err := mapIdentity(row)
	if err != nil {
		return Identity{}, fmt.Errorf(
			"auth identity 取得結果の変換 provider=%s provider_subject=%s: %w",
			provider,
			providerSubject,
			err,
		)
	}

	return identity, nil
}

// GetIdentityByEmail は email provider の auth identity を取得します。
func (r *Repository) GetIdentityByEmail(ctx context.Context, emailNormalized string) (Identity, error) {
	return r.GetIdentityByProviderAndSubject(ctx, identityProviderEmail, emailNormalized)
}

// GetIdentityByNormalizedEmail は provider を問わず email 正規化値で auth identity を取得します。
func (r *Repository) GetIdentityByNormalizedEmail(ctx context.Context, emailNormalized string) (Identity, error) {
	q, err := r.dbQueries()
	if err != nil {
		return Identity{}, err
	}

	row, err := q.GetAuthIdentityByEmailNormalized(ctx, postgres.TextToPG(&emailNormalized))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Identity{}, fmt.Errorf("auth identity 取得 email=%s: %w", emailNormalized, ErrIdentityNotFound)
		}

		return Identity{}, fmt.Errorf("auth identity 取得 email=%s: %w", emailNormalized, err)
	}

	identity, err := mapIdentity(row)
	if err != nil {
		return Identity{}, fmt.Errorf("auth identity 取得結果の変換 email=%s: %w", emailNormalized, err)
	}

	return identity, nil
}

// CreateIdentity は auth identity を作成します。
func (r *Repository) CreateIdentity(ctx context.Context, input CreateIdentityInput) (Identity, error) {
	q, err := r.dbQueries()
	if err != nil {
		return Identity{}, err
	}

	row, err := q.CreateAuthIdentity(ctx, sqlc.CreateAuthIdentityParams{
		UserID:              postgres.UUIDToPG(input.UserID),
		Provider:            input.Provider,
		ProviderSubject:     input.ProviderSubject,
		EmailNormalized:     postgres.TextToPG(input.EmailNormalized),
		VerifiedAt:          postgres.TimeToPG(input.VerifiedAt),
		LastAuthenticatedAt: postgres.TimeToPG(input.LastAuthenticatedAt),
	})
	if err != nil {
		return Identity{}, fmt.Errorf(
			"auth identity 作成 provider=%s provider_subject=%s: %w",
			input.Provider,
			input.ProviderSubject,
			mapIdentityWriteError(err),
		)
	}

	identity, err := mapIdentity(row)
	if err != nil {
		return Identity{}, fmt.Errorf(
			"auth identity 作成結果の変換 provider=%s provider_subject=%s: %w",
			input.Provider,
			input.ProviderSubject,
			err,
		)
	}

	return identity, nil
}

// CreateLoginChallenge は login challenge を作成します。
func (r *Repository) CreateLoginChallenge(ctx context.Context, input CreateLoginChallengeInput) (Challenge, error) {
	q, err := r.dbQueries()
	if err != nil {
		return Challenge{}, err
	}

	row, err := q.CreateAuthLoginChallenge(ctx, sqlc.CreateAuthLoginChallengeParams{
		Provider:           identityProviderEmail,
		ProviderSubject:    input.EmailNormalized,
		EmailNormalized:    postgres.TextToPG(&input.EmailNormalized),
		ChallengeTokenHash: input.ChallengeTokenHash,
		Purpose:            loginChallengePurpose,
		ExpiresAt:          postgres.TimeToPG(&input.ExpiresAt),
		AttemptCount:       0,
	})
	if err != nil {
		return Challenge{}, fmt.Errorf("login challenge 作成 email=%s: %w", input.EmailNormalized, err)
	}

	challenge, err := mapChallenge(row)
	if err != nil {
		return Challenge{}, fmt.Errorf("login challenge 作成結果の変換 email=%s: %w", input.EmailNormalized, err)
	}

	return challenge, nil
}

// GetLatestPendingLoginChallengeByEmail は最新の有効 challenge を取得します。
func (r *Repository) GetLatestPendingLoginChallengeByEmail(ctx context.Context, emailNormalized string) (Challenge, error) {
	q, err := r.dbQueries()
	if err != nil {
		return Challenge{}, err
	}

	row, err := q.GetLatestPendingAuthLoginChallengeByProviderAndSubject(ctx, sqlc.GetLatestPendingAuthLoginChallengeByProviderAndSubjectParams{
		Provider:        identityProviderEmail,
		ProviderSubject: emailNormalized,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Challenge{}, fmt.Errorf("login challenge 取得 email=%s: %w", emailNormalized, ErrLoginChallengeNotFound)
		}

		return Challenge{}, fmt.Errorf("login challenge 取得 email=%s: %w", emailNormalized, err)
	}

	challenge, err := mapChallenge(row)
	if err != nil {
		return Challenge{}, fmt.Errorf("login challenge 取得結果の変換 email=%s: %w", emailNormalized, err)
	}

	return challenge, nil
}

// IncrementLoginChallengeAttemptCount は challenge の attempt count を増やします。
func (r *Repository) IncrementLoginChallengeAttemptCount(ctx context.Context, id uuid.UUID) (Challenge, error) {
	q, err := r.dbQueries()
	if err != nil {
		return Challenge{}, err
	}

	row, err := q.IncrementAuthLoginChallengeAttemptCount(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Challenge{}, fmt.Errorf("login challenge attempt 更新 id=%s: %w", id, ErrLoginChallengeNotFound)
		}

		return Challenge{}, fmt.Errorf("login challenge attempt 更新 id=%s: %w", id, err)
	}

	challenge, err := mapChallenge(row)
	if err != nil {
		return Challenge{}, fmt.Errorf("login challenge attempt 更新結果の変換 id=%s: %w", id, err)
	}

	return challenge, nil
}

// ConsumeLoginChallenge は challenge を消費済みにします。
func (r *Repository) ConsumeLoginChallenge(ctx context.Context, id uuid.UUID, consumedAt time.Time) (Challenge, error) {
	q, err := r.dbQueries()
	if err != nil {
		return Challenge{}, err
	}

	row, err := q.ConsumeAuthLoginChallenge(ctx, sqlc.ConsumeAuthLoginChallengeParams{
		ID:         postgres.UUIDToPG(id),
		ConsumedAt: postgres.TimeToPG(&consumedAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Challenge{}, fmt.Errorf("login challenge consume id=%s: %w", id, ErrLoginChallengeNotFound)
		}

		return Challenge{}, fmt.Errorf("login challenge consume id=%s: %w", id, err)
	}

	challenge, err := mapChallenge(row)
	if err != nil {
		return Challenge{}, fmt.Errorf("login challenge consume 結果の変換 id=%s: %w", id, err)
	}

	return challenge, nil
}

// RecordIdentityAuthentication は認証成功時刻を記録します。
func (r *Repository) RecordIdentityAuthentication(ctx context.Context, input RecordIdentityAuthenticationInput) (Identity, error) {
	q, err := r.dbQueries()
	if err != nil {
		return Identity{}, err
	}

	row, err := q.RecordAuthIdentityAuthentication(ctx, sqlc.RecordAuthIdentityAuthenticationParams{
		ID:                  postgres.UUIDToPG(input.ID),
		EmailNormalized:     postgres.TextToPG(&input.EmailNormalized),
		VerifiedAt:          postgres.TimeToPG(input.VerifiedAt),
		LastAuthenticatedAt: postgres.TimeToPG(&input.LastAuthenticatedAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Identity{}, fmt.Errorf("auth identity 認証記録 id=%s: %w", input.ID, ErrIdentityNotFound)
		}

		return Identity{}, fmt.Errorf("auth identity 認証記録 id=%s: %w", input.ID, err)
	}

	identity, err := mapIdentity(row)
	if err != nil {
		return Identity{}, fmt.Errorf("auth identity 認証記録結果の変換 id=%s: %w", input.ID, err)
	}

	return identity, nil
}

// CreateSession は auth session を作成します。
func (r *Repository) CreateSession(ctx context.Context, input CreateSessionInput) (SessionRecord, error) {
	q, err := r.dbQueries()
	if err != nil {
		return SessionRecord{}, err
	}

	row, err := q.CreateAuthSession(ctx, sqlc.CreateAuthSessionParams{
		UserID:                postgres.UUIDToPG(input.UserID),
		ActiveMode:            string(input.ActiveMode),
		SessionTokenHash:      input.SessionTokenHash,
		ExpiresAt:             postgres.TimeToPG(&input.ExpiresAt),
		RecentAuthenticatedAt: postgres.TimeToPG(&input.RecentAuthenticatedAt),
	})
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session 作成 user=%s: %w", input.UserID, err)
	}

	session, err := mapSession(row)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session 作成結果の変換 user=%s: %w", input.UserID, err)
	}

	return session, nil
}

// CreateUserWithIdentityAndSession は identity と session の一括作成を transaction で行います。
func (r *Repository) CreateUserWithIdentityAndSession(
	ctx context.Context,
	input CreateUserWithIdentityAndSessionInput,
) (SessionRecord, error) {
	if r.txBeginner == nil {
		return SessionRecord{}, fmt.Errorf("auth repository pool が初期化されていません")
	}

	var sessionRow sqlc.AppAuthSession
	err := postgres.RunInTx(ctx, r.txBeginner, func(tx pgx.Tx) error {
		q := sqlc.New(tx)

		if err := lockEmailClaimIfNeeded(ctx, tx, input.EmailNormalized); err != nil {
			return err
		}
		if err := ensureNormalizedEmailAvailable(ctx, q, input.EmailNormalized); err != nil {
			return fmt.Errorf("auth identity 競合確認: %w", err)
		}

		user, err := q.CreateUser(ctx)
		if err != nil {
			return fmt.Errorf("user 作成: %w", err)
		}

		if _, err := q.CreateAuthIdentity(ctx, sqlc.CreateAuthIdentityParams{
			UserID:              user.ID,
			Provider:            input.Provider,
			ProviderSubject:     input.ProviderSubject,
			EmailNormalized:     postgres.TextToPG(input.EmailNormalized),
			VerifiedAt:          postgres.TimeToPG(input.VerifiedAt),
			LastAuthenticatedAt: postgres.TimeToPG(input.LastAuthenticatedAt),
		}); err != nil {
			return fmt.Errorf("auth identity 作成: %w", mapIdentityWriteError(err))
		}

		sessionRow, err = q.CreateAuthSession(ctx, sqlc.CreateAuthSessionParams{
			UserID:                user.ID,
			ActiveMode:            string(ActiveModeFan),
			SessionTokenHash:      input.SessionTokenHash,
			ExpiresAt:             postgres.TimeToPG(&input.ExpiresAt),
			RecentAuthenticatedAt: postgres.TimeToPG(&input.RecentAuthenticatedAt),
		})
		if err != nil {
			return fmt.Errorf("auth session 作成: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrIdentityAlreadyExists) {
			return SessionRecord{}, err
		}

		return SessionRecord{}, fmt.Errorf(
			"identity/session 一括作成 provider=%s provider_subject=%s: %w",
			input.Provider,
			input.ProviderSubject,
			err,
		)
	}

	session, err := mapSession(sessionRow)
	if err != nil {
		return SessionRecord{}, fmt.Errorf(
			"identity/session 一括作成結果の変換 provider=%s provider_subject=%s: %w",
			input.Provider,
			input.ProviderSubject,
			err,
		)
	}

	return session, nil
}

// CreateUserWithEmailIdentityAndSession は sign-up 完了時の一括作成を transaction で行います。
func (r *Repository) CreateUserWithEmailIdentityAndSession(ctx context.Context, input CreateUserWithEmailIdentityAndSessionInput) (SessionRecord, error) {
	if r.txBeginner == nil {
		return SessionRecord{}, fmt.Errorf("auth repository pool が初期化されていません")
	}

	var sessionRow sqlc.AppAuthSession
	err := postgres.RunInTx(ctx, r.txBeginner, func(tx pgx.Tx) error {
		q := sqlc.New(tx)

		user, err := q.CreateUser(ctx)
		if err != nil {
			return fmt.Errorf("user 作成: %w", err)
		}

		if _, err := q.CreateAuthIdentity(ctx, sqlc.CreateAuthIdentityParams{
			UserID:              user.ID,
			Provider:            identityProviderEmail,
			ProviderSubject:     input.EmailNormalized,
			EmailNormalized:     postgres.TextToPG(&input.EmailNormalized),
			VerifiedAt:          postgres.TimeToPG(&input.VerifiedAt),
			LastAuthenticatedAt: postgres.TimeToPG(&input.LastAuthenticatedAt),
		}); err != nil {
			return fmt.Errorf("auth identity 作成: %w", mapIdentityWriteError(err))
		}

		sessionRow, err = q.CreateAuthSession(ctx, sqlc.CreateAuthSessionParams{
			UserID:                user.ID,
			ActiveMode:            string(ActiveModeFan),
			SessionTokenHash:      input.SessionTokenHash,
			ExpiresAt:             postgres.TimeToPG(&input.ExpiresAt),
			RecentAuthenticatedAt: postgres.TimeToPG(&input.RecentAuthenticatedAt),
		})
		if err != nil {
			return fmt.Errorf("auth session 作成: %w", err)
		}

		if _, err := q.CreateUserProfile(ctx, sqlc.CreateUserProfileParams{
			UserID:      user.ID,
			DisplayName: input.DisplayName,
			Handle:      input.Handle,
			AvatarUrl:   postgres.TextToPG(nil),
		}); err != nil {
			return fmt.Errorf("user profile 作成: %w", mapUserProfileWriteError(err))
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrIdentityAlreadyExists) {
			return SessionRecord{}, err
		}
		if errors.Is(err, ErrHandleAlreadyTaken) {
			return SessionRecord{}, err
		}

		return SessionRecord{}, fmt.Errorf("sign up 一括作成 email=%s: %w", input.EmailNormalized, err)
	}

	session, err := mapSession(sessionRow)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("sign up session 結果の変換 email=%s: %w", input.EmailNormalized, err)
	}

	return session, nil
}

// RefreshSessionRecentAuthenticatedAtByTokenHash は recent auth proof を更新します。
func (r *Repository) RefreshSessionRecentAuthenticatedAtByTokenHash(
	ctx context.Context,
	sessionTokenHash string,
	recentAuthenticatedAt time.Time,
) (SessionRecord, error) {
	q, err := r.dbQueries()
	if err != nil {
		return SessionRecord{}, err
	}

	row, err := q.RefreshAuthSessionRecentAuthenticatedAtByTokenHash(
		ctx,
		sqlc.RefreshAuthSessionRecentAuthenticatedAtByTokenHashParams{
			SessionTokenHash:      sessionTokenHash,
			RecentAuthenticatedAt: postgres.TimeToPG(&recentAuthenticatedAt),
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionRecord{}, fmt.Errorf("auth session recent auth 更新 token=%s: %w", sessionTokenHash, ErrSessionNotFound)
		}

		return SessionRecord{}, fmt.Errorf("auth session recent auth 更新 token=%s: %w", sessionTokenHash, err)
	}

	session, err := mapSession(row)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session recent auth 更新結果の変換 token=%s: %w", sessionTokenHash, err)
	}

	return session, nil
}

// TouchSessionLastSeenByTokenHash は session の last seen を更新します。
func (r *Repository) TouchSessionLastSeenByTokenHash(ctx context.Context, sessionTokenHash string, lastSeenAt time.Time) (SessionRecord, error) {
	q, err := r.bootstrapQueries()
	if err != nil {
		return SessionRecord{}, err
	}

	row, err := q.TouchAuthSessionLastSeenByTokenHash(ctx, sqlc.TouchAuthSessionLastSeenByTokenHashParams{
		SessionTokenHash: sessionTokenHash,
		LastSeenAt:       postgres.TimeToPG(&lastSeenAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionRecord{}, fmt.Errorf("auth session touch token=%s: %w", sessionTokenHash, ErrSessionNotFound)
		}

		return SessionRecord{}, fmt.Errorf("auth session touch token=%s: %w", sessionTokenHash, err)
	}

	session, err := mapSession(row)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session touch 結果の変換 token=%s: %w", sessionTokenHash, err)
	}

	return session, nil
}

// RevokeActiveSessionByTokenHash は有効な session を revoke します。
func (r *Repository) RevokeActiveSessionByTokenHash(ctx context.Context, sessionTokenHash string, revokedAt time.Time) (SessionRecord, error) {
	q, err := r.dbQueries()
	if err != nil {
		return SessionRecord{}, err
	}

	row, err := q.RevokeActiveAuthSessionByTokenHash(ctx, sqlc.RevokeActiveAuthSessionByTokenHashParams{
		SessionTokenHash: sessionTokenHash,
		RevokedAt:        postgres.TimeToPG(&revokedAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionRecord{}, fmt.Errorf("auth session revoke token=%s: %w", sessionTokenHash, ErrSessionNotFound)
		}

		return SessionRecord{}, fmt.Errorf("auth session revoke token=%s: %w", sessionTokenHash, err)
	}

	session, err := mapSession(row)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session revoke 結果の変換 token=%s: %w", sessionTokenHash, err)
	}

	return session, nil
}

// UpdateActiveModeByTokenHash は session token hash で active mode を更新します。
func (r *Repository) UpdateActiveModeByTokenHash(ctx context.Context, sessionTokenHash string, activeMode ActiveMode) (SessionRecord, error) {
	q, err := r.dbQueries()
	if err != nil {
		return SessionRecord{}, err
	}

	row, err := q.UpdateActiveAuthSessionModeByTokenHash(ctx, sqlc.UpdateActiveAuthSessionModeByTokenHashParams{
		ActiveMode:       string(activeMode),
		SessionTokenHash: sessionTokenHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionRecord{}, fmt.Errorf("auth session active mode 更新 token=%s: %w", sessionTokenHash, ErrSessionNotFound)
		}

		return SessionRecord{}, fmt.Errorf("auth session active mode 更新 token=%s: %w", sessionTokenHash, err)
	}

	session, err := mapSession(row)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session active mode 更新結果の変換 token=%s: %w", sessionTokenHash, err)
	}

	return session, nil
}

// GetCurrentViewerBySessionTokenHash は session token hash から current viewer を取得します。
func (r *Repository) GetCurrentViewerBySessionTokenHash(ctx context.Context, sessionTokenHash string) (CurrentViewer, error) {
	q, err := r.bootstrapQueries()
	if err != nil {
		return CurrentViewer{}, err
	}

	row, err := q.GetCurrentViewerBySessionTokenHash(ctx, sessionTokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CurrentViewer{}, fmt.Errorf("current viewer 取得 session=%s: %w", sessionTokenHash, ErrCurrentViewerNotFound)
		}

		return CurrentViewer{}, fmt.Errorf("current viewer 取得 session=%s: %w", sessionTokenHash, err)
	}

	viewer, err := mapCurrentViewer(row)
	if err != nil {
		return CurrentViewer{}, fmt.Errorf("current viewer 取得結果の変換 session=%s: %w", sessionTokenHash, err)
	}

	return viewer, nil
}

func (r *Repository) dbQueries() (*sqlc.Queries, error) {
	if r.db == nil {
		return nil, fmt.Errorf("auth repository pool が初期化されていません")
	}

	return sqlc.New(r.db), nil
}

func (r *Repository) bootstrapQueries() (queries, error) {
	if r.queries != nil {
		return r.queries, nil
	}
	if r.db == nil {
		return nil, fmt.Errorf("auth repository queries が初期化されていません")
	}

	return sqlc.New(r.db), nil
}

func mapCurrentViewer(row sqlc.GetCurrentViewerBySessionTokenHashRow) (CurrentViewer, error) {
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return CurrentViewer{}, fmt.Errorf("current viewer の user id 変換: %w", err)
	}

	return CurrentViewer{
		ID:                   userID,
		ActiveMode:           ActiveMode(row.ActiveMode),
		CanAccessCreatorMode: row.CanAccessCreatorMode,
	}, nil
}

func mapIdentity(row sqlc.AppAuthIdentity) (Identity, error) {
	id, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return Identity{}, fmt.Errorf("auth identity id 変換: %w", err)
	}

	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return Identity{}, fmt.Errorf("auth identity user id 変換: %w", err)
	}

	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Identity{}, fmt.Errorf("auth identity created_at 変換: %w", err)
	}

	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Identity{}, fmt.Errorf("auth identity updated_at 変換: %w", err)
	}

	return Identity{
		ID:                  id,
		UserID:              userID,
		Provider:            row.Provider,
		ProviderSubject:     row.ProviderSubject,
		EmailNormalized:     postgres.OptionalTextFromPG(row.EmailNormalized),
		VerifiedAt:          postgres.OptionalTimeFromPG(row.VerifiedAt),
		LastAuthenticatedAt: postgres.OptionalTimeFromPG(row.LastAuthenticatedAt),
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}, nil
}

func mapChallenge(row sqlc.AppAuthLoginChallenge) (Challenge, error) {
	id, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return Challenge{}, fmt.Errorf("login challenge id 変換: %w", err)
	}

	expiresAt, err := postgres.RequiredTimeFromPG(row.ExpiresAt)
	if err != nil {
		return Challenge{}, fmt.Errorf("login challenge expires_at 変換: %w", err)
	}

	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Challenge{}, fmt.Errorf("login challenge created_at 変換: %w", err)
	}

	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Challenge{}, fmt.Errorf("login challenge updated_at 変換: %w", err)
	}

	return Challenge{
		ID:                 id,
		Provider:           row.Provider,
		ProviderSubject:    row.ProviderSubject,
		EmailNormalized:    postgres.OptionalTextFromPG(row.EmailNormalized),
		ChallengeTokenHash: row.ChallengeTokenHash,
		Purpose:            row.Purpose,
		ExpiresAt:          expiresAt,
		ConsumedAt:         postgres.OptionalTimeFromPG(row.ConsumedAt),
		AttemptCount:       row.AttemptCount,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}, nil
}

func mapSession(row sqlc.AppAuthSession) (SessionRecord, error) {
	id, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session id 変換: %w", err)
	}

	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session user id 変換: %w", err)
	}

	expiresAt, err := postgres.RequiredTimeFromPG(row.ExpiresAt)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session expires_at 変換: %w", err)
	}

	recentAuthenticatedAt, err := postgres.RequiredTimeFromPG(row.RecentAuthenticatedAt)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session recent_authenticated_at 変換: %w", err)
	}

	lastSeenAt, err := postgres.RequiredTimeFromPG(row.LastSeenAt)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session last_seen_at 変換: %w", err)
	}

	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session created_at 変換: %w", err)
	}

	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("auth session updated_at 変換: %w", err)
	}

	return SessionRecord{
		ID:                    id,
		UserID:                userID,
		ActiveMode:            ActiveMode(row.ActiveMode),
		SessionTokenHash:      row.SessionTokenHash,
		ExpiresAt:             expiresAt,
		RecentAuthenticatedAt: recentAuthenticatedAt,
		LastSeenAt:            lastSeenAt,
		RevokedAt:             postgres.OptionalTimeFromPG(row.RevokedAt),
		CreatedAt:             createdAt,
		UpdatedAt:             updatedAt,
	}, nil
}

func mapIdentityWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case identityUniqueConstraint, emailUniqueConstraint, cognitoEmailUniqueConstraint:
			return ErrIdentityAlreadyExists
		}
	}

	return err
}

func lockEmailClaimIfNeeded(ctx context.Context, tx pgx.Tx, emailNormalized *string) error {
	if emailNormalized == nil {
		return nil
	}

	for {
		var acquired bool
		err := tx.QueryRow(
			ctx,
			`SELECT pg_try_advisory_xact_lock(hashtextextended($1, 0))`,
			*emailNormalized,
		).Scan(&acquired)
		if err != nil {
			return fmt.Errorf("auth identity email claim lock 取得 email=%s: %w", *emailNormalized, err)
		}
		if acquired {
			return nil
		}

		timer := time.NewTimer(emailClaimLockRetryInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}

			return fmt.Errorf("auth identity email claim lock 待機 email=%s: %w", *emailNormalized, ctx.Err())
		case <-timer.C:
		}
	}
}

func ensureNormalizedEmailAvailable(ctx context.Context, q *sqlc.Queries, emailNormalized *string) error {
	if emailNormalized == nil {
		return nil
	}

	_, err := q.GetAuthIdentityByEmailNormalized(ctx, postgres.TextToPG(emailNormalized))
	if err == nil {
		return ErrIdentityAlreadyExists
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	return fmt.Errorf("auth identity 取得 email=%s: %w", *emailNormalized, err)
}

func mapUserProfileWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case userProfileHandleUnique:
			return ErrHandleAlreadyTaken
		}
	}

	return err
}
