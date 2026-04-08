package auth

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type stubQueries struct {
	touchAuthSessionLastSeenByTokenHash func(context.Context, sqlc.TouchAuthSessionLastSeenByTokenHashParams) (sqlc.AppAuthSession, error)
	updateActiveModeByTokenHash         func(context.Context, sqlc.UpdateActiveAuthSessionModeByTokenHashParams) (sqlc.AppAuthSession, error)
	getCurrentViewerBySessionTokenHash  func(context.Context, string) (sqlc.GetCurrentViewerBySessionTokenHashRow, error)
}

func (s stubQueries) TouchAuthSessionLastSeenByTokenHash(
	ctx context.Context,
	arg sqlc.TouchAuthSessionLastSeenByTokenHashParams,
) (sqlc.AppAuthSession, error) {
	if s.touchAuthSessionLastSeenByTokenHash == nil {
		return sqlc.AppAuthSession{}, nil
	}

	return s.touchAuthSessionLastSeenByTokenHash(ctx, arg)
}

func (s stubQueries) GetCurrentViewerBySessionTokenHash(
	ctx context.Context,
	sessionTokenHash string,
) (sqlc.GetCurrentViewerBySessionTokenHashRow, error) {
	return s.getCurrentViewerBySessionTokenHash(ctx, sessionTokenHash)
}

func (s stubQueries) UpdateActiveAuthSessionModeByTokenHash(
	ctx context.Context,
	arg sqlc.UpdateActiveAuthSessionModeByTokenHashParams,
) (sqlc.AppAuthSession, error) {
	if s.updateActiveModeByTokenHash == nil {
		return sqlc.AppAuthSession{}, nil
	}

	return s.updateActiveModeByTokenHash(ctx, arg)
}

type dbtxStub struct {
	queryRow func(context.Context, string, ...any) pgx.Row
}

func (s dbtxStub) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, fmt.Errorf("unexpected Exec call")
}

func (s dbtxStub) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, fmt.Errorf("unexpected Query call")
}

func (s dbtxStub) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if s.queryRow == nil {
		return rowErr(fmt.Errorf("unexpected QueryRow call: %s", sql))
	}

	return s.queryRow(ctx, sql, args...)
}

type rowStub struct {
	err    error
	scanFn func(dest ...any) error
}

func (s rowStub) Scan(dest ...any) error {
	if s.scanFn != nil {
		return s.scanFn(dest...)
	}

	return s.err
}

type txBeginnerStub struct {
	beginErr error
	tx       pgx.Tx
	began    bool
}

func (s *txBeginnerStub) Begin(context.Context) (pgx.Tx, error) {
	s.began = true
	if s.beginErr != nil {
		return nil, s.beginErr
	}

	return s.tx, nil
}

type txStub struct {
	queryRow    func(context.Context, string, ...any) pgx.Row
	commitErr   error
	rollbackErr error
	committed   bool
	rolledBack  bool
}

func (s *txStub) Begin(context.Context) (pgx.Tx, error) {
	return nil, fmt.Errorf("unexpected nested Begin call")
}

func (s *txStub) Commit(context.Context) error {
	s.committed = true
	return s.commitErr
}

func (s *txStub) Rollback(context.Context) error {
	s.rolledBack = true
	return s.rollbackErr
}

func (s *txStub) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, fmt.Errorf("unexpected CopyFrom call")
}

func (s *txStub) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	return nil
}

func (s *txStub) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (s *txStub) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, fmt.Errorf("unexpected Prepare call")
}

func (s *txStub) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, fmt.Errorf("unexpected Exec call")
}

func (s *txStub) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, fmt.Errorf("unexpected Query call")
}

func (s *txStub) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if s.queryRow == nil {
		return rowErr(fmt.Errorf("unexpected QueryRow call: %s", sql))
	}

	return s.queryRow(ctx, sql, args...)
}

func (s *txStub) Conn() *pgx.Conn {
	return nil
}

func rowWithValues(values ...any) pgx.Row {
	return rowStub{
		scanFn: func(dest ...any) error {
			if len(dest) != len(values) {
				return fmt.Errorf("scan dest count got %d want %d", len(dest), len(values))
			}

			for i, value := range values {
				destValue := reflect.ValueOf(dest[i])
				if destValue.Kind() != reflect.Pointer || destValue.IsNil() {
					return fmt.Errorf("scan dest[%d] must be a non-nil pointer", i)
				}

				sourceValue := reflect.ValueOf(value)
				if !sourceValue.Type().AssignableTo(destValue.Elem().Type()) {
					return fmt.Errorf("scan value[%d] type %s is not assignable to %s", i, sourceValue.Type(), destValue.Elem().Type())
				}

				destValue.Elem().Set(sourceValue)
			}

			return nil
		},
	}
}

func rowErr(err error) pgx.Row {
	return rowStub{err: err}
}

func TestNewRepositoryInitializesDependencies(t *testing.T) {
	t.Parallel()

	repository := NewRepository(&pgxpool.Pool{})
	if repository.txBeginner == nil {
		t.Fatal("NewRepository() txBeginner = nil, want initialized")
	}
	if repository.db == nil {
		t.Fatal("NewRepository() db = nil, want initialized")
	}
	if repository.queries == nil {
		t.Fatal("NewRepository() queries = nil, want initialized")
	}
}

func TestGetIdentityByEmail(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	identityID := uuid.New()
	userID := uuid.New()
	repository := &Repository{
		db: dbtxStub{
			queryRow: func(context.Context, string, ...any) pgx.Row {
				return rowWithValues(
					pgUUID(identityID),
					pgUUID(userID),
					identityProviderEmail,
					"fan@example.com",
					pgText("fan@example.com"),
					pgTime(now),
					pgTime(now.Add(time.Minute)),
					pgTime(now.Add(-time.Hour)),
					pgTime(now.Add(-30*time.Minute)),
				)
			},
		},
	}

	got, err := repository.GetIdentityByEmail(context.Background(), "fan@example.com")
	if err != nil {
		t.Fatalf("GetIdentityByEmail() error = %v, want nil", err)
	}
	if got.ID != identityID {
		t.Fatalf("GetIdentityByEmail() id got %s want %s", got.ID, identityID)
	}
	if got.UserID != userID {
		t.Fatalf("GetIdentityByEmail() user id got %s want %s", got.UserID, userID)
	}
	if got.EmailNormalized == nil || *got.EmailNormalized != "fan@example.com" {
		t.Fatalf("GetIdentityByEmail() email got %v want %q", got.EmailNormalized, "fan@example.com")
	}
}

func TestGetIdentityByEmailNotFound(t *testing.T) {
	t.Parallel()

	repository := &Repository{
		db: dbtxStub{
			queryRow: func(context.Context, string, ...any) pgx.Row {
				return rowErr(pgx.ErrNoRows)
			},
		},
	}

	if _, err := repository.GetIdentityByEmail(context.Background(), "fan@example.com"); !errors.Is(err, ErrIdentityNotFound) {
		t.Fatalf("GetIdentityByEmail() error got %v want %v", err, ErrIdentityNotFound)
	}
}

func TestCreateLoginChallenge(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000100, 0).UTC()
	challengeID := uuid.New()
	repository := &Repository{
		db: dbtxStub{
			queryRow: func(context.Context, string, ...any) pgx.Row {
				return rowWithValues(
					pgUUID(challengeID),
					identityProviderEmail,
					"fan@example.com",
					pgText("fan@example.com"),
					"challenge-hash",
					loginChallengePurpose,
					pgTime(now.Add(10*time.Minute)),
					pgtype.Timestamptz{},
					int32(0),
					pgTime(now),
					pgTime(now),
				)
			},
		},
	}

	got, err := repository.CreateLoginChallenge(context.Background(), CreateLoginChallengeInput{
		EmailNormalized:    "fan@example.com",
		ChallengeTokenHash: "challenge-hash",
		ExpiresAt:          now.Add(10 * time.Minute),
	})
	if err != nil {
		t.Fatalf("CreateLoginChallenge() error = %v, want nil", err)
	}
	if got.ID != challengeID {
		t.Fatalf("CreateLoginChallenge() id got %s want %s", got.ID, challengeID)
	}
	if got.ChallengeTokenHash != "challenge-hash" {
		t.Fatalf("CreateLoginChallenge() token hash got %q want %q", got.ChallengeTokenHash, "challenge-hash")
	}
}

func TestGetLatestPendingLoginChallengeByEmail(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000200, 0).UTC()
	challengeID := uuid.New()
	repository := &Repository{
		db: dbtxStub{
			queryRow: func(context.Context, string, ...any) pgx.Row {
				return rowWithValues(
					pgUUID(challengeID),
					identityProviderEmail,
					"fan@example.com",
					pgText("fan@example.com"),
					"challenge-hash",
					loginChallengePurpose,
					pgTime(now.Add(10*time.Minute)),
					pgtype.Timestamptz{},
					int32(2),
					pgTime(now),
					pgTime(now),
				)
			},
		},
	}

	got, err := repository.GetLatestPendingLoginChallengeByEmail(context.Background(), "fan@example.com")
	if err != nil {
		t.Fatalf("GetLatestPendingLoginChallengeByEmail() error = %v, want nil", err)
	}
	if got.ID != challengeID {
		t.Fatalf("GetLatestPendingLoginChallengeByEmail() id got %s want %s", got.ID, challengeID)
	}
	if got.AttemptCount != 2 {
		t.Fatalf("GetLatestPendingLoginChallengeByEmail() attempt count got %d want %d", got.AttemptCount, 2)
	}
}

func TestIncrementLoginChallengeAttemptCount(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000300, 0).UTC()
	challengeID := uuid.New()
	repository := &Repository{
		db: dbtxStub{
			queryRow: func(context.Context, string, ...any) pgx.Row {
				return rowWithValues(
					pgUUID(challengeID),
					identityProviderEmail,
					"fan@example.com",
					pgText("fan@example.com"),
					"challenge-hash",
					loginChallengePurpose,
					pgTime(now.Add(10*time.Minute)),
					pgtype.Timestamptz{},
					int32(3),
					pgTime(now),
					pgTime(now),
				)
			},
		},
	}

	got, err := repository.IncrementLoginChallengeAttemptCount(context.Background(), challengeID)
	if err != nil {
		t.Fatalf("IncrementLoginChallengeAttemptCount() error = %v, want nil", err)
	}
	if got.AttemptCount != 3 {
		t.Fatalf("IncrementLoginChallengeAttemptCount() attempt count got %d want %d", got.AttemptCount, 3)
	}
}

func TestConsumeLoginChallenge(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000400, 0).UTC()
	consumedAt := now.Add(time.Minute)
	challengeID := uuid.New()
	repository := &Repository{
		db: dbtxStub{
			queryRow: func(context.Context, string, ...any) pgx.Row {
				return rowWithValues(
					pgUUID(challengeID),
					identityProviderEmail,
					"fan@example.com",
					pgText("fan@example.com"),
					"challenge-hash",
					loginChallengePurpose,
					pgTime(now.Add(10*time.Minute)),
					pgTime(consumedAt),
					int32(1),
					pgTime(now),
					pgTime(now),
				)
			},
		},
	}

	got, err := repository.ConsumeLoginChallenge(context.Background(), challengeID, consumedAt)
	if err != nil {
		t.Fatalf("ConsumeLoginChallenge() error = %v, want nil", err)
	}
	if got.ConsumedAt == nil || !got.ConsumedAt.Equal(consumedAt) {
		t.Fatalf("ConsumeLoginChallenge() consumed at got %v want %s", got.ConsumedAt, consumedAt)
	}
}

func TestRecordIdentityAuthentication(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000500, 0).UTC()
	identityID := uuid.New()
	userID := uuid.New()
	repository := &Repository{
		db: dbtxStub{
			queryRow: func(context.Context, string, ...any) pgx.Row {
				return rowWithValues(
					pgUUID(identityID),
					pgUUID(userID),
					identityProviderEmail,
					"fan@example.com",
					pgText("fan@example.com"),
					pgTime(now),
					pgTime(now.Add(time.Minute)),
					pgTime(now.Add(-time.Hour)),
					pgTime(now.Add(-30*time.Minute)),
				)
			},
		},
	}

	got, err := repository.RecordIdentityAuthentication(context.Background(), RecordIdentityAuthenticationInput{
		ID:                  identityID,
		EmailNormalized:     "fan@example.com",
		VerifiedAt:          &now,
		LastAuthenticatedAt: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("RecordIdentityAuthentication() error = %v, want nil", err)
	}
	if got.ID != identityID {
		t.Fatalf("RecordIdentityAuthentication() id got %s want %s", got.ID, identityID)
	}
}

func TestCreateSession(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000600, 0).UTC()
	sessionID := uuid.New()
	userID := uuid.New()
	repository := &Repository{
		db: dbtxStub{
			queryRow: func(context.Context, string, ...any) pgx.Row {
				return rowWithValues(
					pgUUID(sessionID),
					pgUUID(userID),
					string(ActiveModeFan),
					"session-hash",
					pgTime(now.Add(24*time.Hour)),
					pgTime(now),
					pgtype.Timestamptz{},
					pgTime(now),
					pgTime(now),
				)
			},
		},
	}

	got, err := repository.CreateSession(context.Background(), CreateSessionInput{
		UserID:           userID,
		ActiveMode:       ActiveModeFan,
		SessionTokenHash: "session-hash",
		ExpiresAt:        now.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v, want nil", err)
	}
	if got.ID != sessionID {
		t.Fatalf("CreateSession() id got %s want %s", got.ID, sessionID)
	}
	if got.ActiveMode != ActiveModeFan {
		t.Fatalf("CreateSession() active mode got %q want %q", got.ActiveMode, ActiveModeFan)
	}
}

func TestCreateUserWithEmailIdentityAndSession(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000700, 0).UTC()
	userID := uuid.New()
	identityID := uuid.New()
	sessionID := uuid.New()
	tx := &txStub{}
	callCount := 0
	tx.queryRow = func(context.Context, string, ...any) pgx.Row {
		callCount++
		switch callCount {
		case 1:
			return rowWithValues(
				pgUUID(userID),
				pgTime(now),
				pgTime(now),
			)
		case 2:
			return rowWithValues(
				pgUUID(identityID),
				pgUUID(userID),
				identityProviderEmail,
				"fan@example.com",
				pgText("fan@example.com"),
				pgTime(now),
				pgTime(now),
				pgTime(now),
				pgTime(now),
			)
		case 3:
			return rowWithValues(
				pgUUID(sessionID),
				pgUUID(userID),
				string(ActiveModeFan),
				"session-hash",
				pgTime(now.Add(24*time.Hour)),
				pgTime(now),
				pgtype.Timestamptz{},
				pgTime(now),
				pgTime(now),
			)
		default:
			return rowErr(fmt.Errorf("unexpected QueryRow call count: %d", callCount))
		}
	}

	beginner := &txBeginnerStub{tx: tx}
	repository := &Repository{txBeginner: beginner}

	got, err := repository.CreateUserWithEmailIdentityAndSession(context.Background(), CreateUserWithEmailIdentityAndSessionInput{
		EmailNormalized:     "fan@example.com",
		SessionTokenHash:    "session-hash",
		VerifiedAt:          now,
		LastAuthenticatedAt: now,
		ExpiresAt:           now.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("CreateUserWithEmailIdentityAndSession() error = %v, want nil", err)
	}
	if got.ID != sessionID {
		t.Fatalf("CreateUserWithEmailIdentityAndSession() session id got %s want %s", got.ID, sessionID)
	}
	if !beginner.began {
		t.Fatal("CreateUserWithEmailIdentityAndSession() did not begin transaction")
	}
	if !tx.committed {
		t.Fatal("CreateUserWithEmailIdentityAndSession() did not commit transaction")
	}
	if tx.rolledBack {
		t.Fatal("CreateUserWithEmailIdentityAndSession() rolled back successful transaction")
	}
}

func TestCreateUserWithEmailIdentityAndSessionMapsDuplicateIdentity(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000800, 0).UTC()
	tx := &txStub{}
	callCount := 0
	tx.queryRow = func(context.Context, string, ...any) pgx.Row {
		callCount++
		switch callCount {
		case 1:
			return rowWithValues(
				pgUUID(uuid.New()),
				pgTime(now),
				pgTime(now),
			)
		case 2:
			return rowErr(&pgconn.PgError{
				Code:           "23505",
				ConstraintName: emailUniqueConstraint,
			})
		default:
			return rowErr(fmt.Errorf("unexpected QueryRow call count: %d", callCount))
		}
	}

	repository := &Repository{txBeginner: &txBeginnerStub{tx: tx}}

	if _, err := repository.CreateUserWithEmailIdentityAndSession(context.Background(), CreateUserWithEmailIdentityAndSessionInput{
		EmailNormalized:     "fan@example.com",
		SessionTokenHash:    "session-hash",
		VerifiedAt:          now,
		LastAuthenticatedAt: now,
		ExpiresAt:           now.Add(24 * time.Hour),
	}); !errors.Is(err, ErrIdentityAlreadyExists) {
		t.Fatalf("CreateUserWithEmailIdentityAndSession() error got %v want %v", err, ErrIdentityAlreadyExists)
	}
	if !tx.rolledBack {
		t.Fatal("CreateUserWithEmailIdentityAndSession() did not roll back failed transaction")
	}
}

func TestGetCurrentViewerBySessionTokenHash(t *testing.T) {
	t.Parallel()

	expectedID := uuid.New()
	repository := newRepository(stubQueries{
		getCurrentViewerBySessionTokenHash: func(context.Context, string) (sqlc.GetCurrentViewerBySessionTokenHashRow, error) {
			return sqlc.GetCurrentViewerBySessionTokenHashRow{
				UserID:               pgUUID(expectedID),
				ActiveMode:           "creator",
				CanAccessCreatorMode: true,
			}, nil
		},
	})

	got, err := repository.GetCurrentViewerBySessionTokenHash(context.Background(), "session-token-hash")
	if err != nil {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() error = %v, want nil", err)
	}
	if got.ID != expectedID {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() id got %s want %s", got.ID, expectedID)
	}
	if got.ActiveMode != ActiveModeCreator {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() active mode got %q want %q", got.ActiveMode, ActiveModeCreator)
	}
	if !got.CanAccessCreatorMode {
		t.Fatal("GetCurrentViewerBySessionTokenHash() can access creator mode = false, want true")
	}
}

func TestGetCurrentViewerBySessionTokenHashNotFound(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{
		getCurrentViewerBySessionTokenHash: func(context.Context, string) (sqlc.GetCurrentViewerBySessionTokenHashRow, error) {
			return sqlc.GetCurrentViewerBySessionTokenHashRow{}, pgx.ErrNoRows
		},
	})

	if _, err := repository.GetCurrentViewerBySessionTokenHash(context.Background(), "session-token-hash"); !errors.Is(err, ErrCurrentViewerNotFound) {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() error got %v want %v", err, ErrCurrentViewerNotFound)
	}
}

func TestGetCurrentViewerBySessionTokenHashRejectsInvalidUUID(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{
		getCurrentViewerBySessionTokenHash: func(context.Context, string) (sqlc.GetCurrentViewerBySessionTokenHashRow, error) {
			return sqlc.GetCurrentViewerBySessionTokenHashRow{
				UserID:               pgtype.UUID{},
				ActiveMode:           "fan",
				CanAccessCreatorMode: false,
			}, nil
		},
	})

	if _, err := repository.GetCurrentViewerBySessionTokenHash(context.Background(), "session-token-hash"); err == nil {
		t.Fatal("GetCurrentViewerBySessionTokenHash() error = nil, want conversion error")
	}
}

func TestTouchSessionLastSeenByTokenHash(t *testing.T) {
	t.Parallel()

	expectedID := uuid.New()
	now := time.Unix(1710000000, 0).UTC()
	repository := newRepository(stubQueries{
		touchAuthSessionLastSeenByTokenHash: func(context.Context, sqlc.TouchAuthSessionLastSeenByTokenHashParams) (sqlc.AppAuthSession, error) {
			return sqlc.AppAuthSession{
				ID:               pgUUID(expectedID),
				UserID:           pgUUID(uuid.New()),
				ActiveMode:       "fan",
				SessionTokenHash: "session-token-hash",
				ExpiresAt:        pgTime(now.Add(time.Hour)),
				LastSeenAt:       pgTime(now),
				CreatedAt:        pgTime(now),
				UpdatedAt:        pgTime(now),
			}, nil
		},
	})

	got, err := repository.TouchSessionLastSeenByTokenHash(context.Background(), "session-token-hash", now)
	if err != nil {
		t.Fatalf("TouchSessionLastSeenByTokenHash() error = %v, want nil", err)
	}
	if got.ID != expectedID {
		t.Fatalf("TouchSessionLastSeenByTokenHash() id got %s want %s", got.ID, expectedID)
	}
}

func TestTouchSessionLastSeenByTokenHashNotFound(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{
		touchAuthSessionLastSeenByTokenHash: func(context.Context, sqlc.TouchAuthSessionLastSeenByTokenHashParams) (sqlc.AppAuthSession, error) {
			return sqlc.AppAuthSession{}, pgx.ErrNoRows
		},
	})

	if _, err := repository.TouchSessionLastSeenByTokenHash(context.Background(), "session-token-hash", time.Now().UTC()); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("TouchSessionLastSeenByTokenHash() error got %v want %v", err, ErrSessionNotFound)
	}
}

func TestUpdateActiveModeByTokenHash(t *testing.T) {
	t.Parallel()

	expectedID := uuid.New()
	userID := uuid.New()
	now := time.Unix(1710000000, 0).UTC()
	var gotArgs []any

	repository := &Repository{
		db: dbtxStub{
			queryRow: func(_ context.Context, _ string, args ...any) pgx.Row {
				gotArgs = args
				return rowWithValues(
					pgUUID(expectedID),
					pgUUID(userID),
					string(ActiveModeCreator),
					"session-token-hash",
					pgTime(now.Add(time.Hour)),
					pgTime(now),
					pgtype.Timestamptz{},
					pgTime(now),
					pgTime(now.Add(time.Minute)),
				)
			},
		},
	}

	got, err := repository.UpdateActiveModeByTokenHash(context.Background(), "session-token-hash", ActiveModeCreator)
	if err != nil {
		t.Fatalf("UpdateActiveModeByTokenHash() error = %v, want nil", err)
	}
	if got.ID != expectedID {
		t.Fatalf("UpdateActiveModeByTokenHash() id got %s want %s", got.ID, expectedID)
	}
	if got.ActiveMode != ActiveModeCreator {
		t.Fatalf("UpdateActiveModeByTokenHash() active mode got %q want %q", got.ActiveMode, ActiveModeCreator)
	}
	if len(gotArgs) != 2 {
		t.Fatalf("UpdateActiveModeByTokenHash() arg count got %d want %d", len(gotArgs), 2)
	}
	if gotArgs[0] != string(ActiveModeCreator) {
		t.Fatalf("UpdateActiveModeByTokenHash() active mode arg got %#v want %q", gotArgs[0], ActiveModeCreator)
	}
	if gotArgs[1] != "session-token-hash" {
		t.Fatalf("UpdateActiveModeByTokenHash() token arg got %#v want %q", gotArgs[1], "session-token-hash")
	}
}

func TestUpdateActiveModeByTokenHashMissingSession(t *testing.T) {
	t.Parallel()

	repository := &Repository{
		db: dbtxStub{
			queryRow: func(context.Context, string, ...any) pgx.Row {
				return rowErr(pgx.ErrNoRows)
			},
		},
	}

	if _, err := repository.UpdateActiveModeByTokenHash(context.Background(), "session-token-hash", ActiveModeCreator); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("UpdateActiveModeByTokenHash() error got %v want %v", err, ErrSessionNotFound)
	}
}

func TestRevokeActiveSessionByTokenHash(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000900, 0).UTC()
	sessionID := uuid.New()
	userID := uuid.New()
	repository := &Repository{
		db: dbtxStub{
			queryRow: func(context.Context, string, ...any) pgx.Row {
				return rowWithValues(
					pgUUID(sessionID),
					pgUUID(userID),
					string(ActiveModeFan),
					"session-hash",
					pgTime(now.Add(24*time.Hour)),
					pgTime(now),
					pgTime(now.Add(time.Minute)),
					pgTime(now),
					pgTime(now),
				)
			},
		},
	}

	got, err := repository.RevokeActiveSessionByTokenHash(context.Background(), "session-hash", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("RevokeActiveSessionByTokenHash() error = %v, want nil", err)
	}
	if got.ID != sessionID {
		t.Fatalf("RevokeActiveSessionByTokenHash() id got %s want %s", got.ID, sessionID)
	}
	if got.RevokedAt == nil || !got.RevokedAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("RevokeActiveSessionByTokenHash() revoked at got %v want %s", got.RevokedAt, now.Add(time.Minute))
	}
}

func TestRevokeActiveSessionByTokenHashNotFound(t *testing.T) {
	t.Parallel()

	repository := &Repository{
		db: dbtxStub{
			queryRow: func(context.Context, string, ...any) pgx.Row {
				return rowErr(pgx.ErrNoRows)
			},
		},
	}

	if _, err := repository.RevokeActiveSessionByTokenHash(context.Background(), "session-hash", time.Now().UTC()); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("RevokeActiveSessionByTokenHash() error got %v want %v", err, ErrSessionNotFound)
	}
}

func TestDBQueriesRequiresDB(t *testing.T) {
	t.Parallel()

	repository := &Repository{}
	if _, err := repository.dbQueries(); err == nil {
		t.Fatal("dbQueries() error = nil, want initialization error")
	}
}

func TestBootstrapQueriesFallsBackToDB(t *testing.T) {
	t.Parallel()

	expected := stubQueries{}
	repository := &Repository{queries: expected}

	got, err := repository.bootstrapQueries()
	if err != nil {
		t.Fatalf("bootstrapQueries() error = %v, want nil", err)
	}
	if _, ok := got.(stubQueries); !ok {
		t.Fatal("bootstrapQueries() did not return explicit query implementation")
	}

	repository = &Repository{db: dbtxStub{queryRow: func(context.Context, string, ...any) pgx.Row {
		return rowErr(fmt.Errorf("not used"))
	}}}
	got, err = repository.bootstrapQueries()
	if err != nil {
		t.Fatalf("bootstrapQueries() fallback error = %v, want nil", err)
	}
	if got == nil {
		t.Fatal("bootstrapQueries() fallback got nil, want sqlc queries")
	}
}

func TestMapIdentityWriteError(t *testing.T) {
	t.Parallel()

	duplicateErr := &pgconn.PgError{
		Code:           "23505",
		ConstraintName: identityUniqueConstraint,
	}
	if !errors.Is(mapIdentityWriteError(duplicateErr), ErrIdentityAlreadyExists) {
		t.Fatal("mapIdentityWriteError() did not map duplicate constraint to ErrIdentityAlreadyExists")
	}

	expectedErr := errors.New("write failed")
	if got := mapIdentityWriteError(expectedErr); !errors.Is(got, expectedErr) {
		t.Fatalf("mapIdentityWriteError() got %v want %v", got, expectedErr)
	}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func pgText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
}

func pgTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}
