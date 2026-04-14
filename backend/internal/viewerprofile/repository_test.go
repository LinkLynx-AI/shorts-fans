package viewerprofile

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type stubQueries struct {
	createUserProfile       func(context.Context, sqlc.CreateUserProfileParams) (sqlc.AppUserProfile, error)
	getCreatorProfileByUser func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error)
	getUserProfileByUser    func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error)
	updateCreatorProfile    func(context.Context, sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	updateUserProfile       func(context.Context, sqlc.UpdateUserProfileParams) (sqlc.AppUserProfile, error)
}

func (s stubQueries) CreateUserProfile(ctx context.Context, arg sqlc.CreateUserProfileParams) (sqlc.AppUserProfile, error) {
	if s.createUserProfile == nil {
		return sqlc.AppUserProfile{}, fmt.Errorf("unexpected CreateUserProfile call")
	}

	return s.createUserProfile(ctx, arg)
}

func (s stubQueries) GetCreatorProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorProfile, error) {
	if s.getCreatorProfileByUser == nil {
		return sqlc.AppCreatorProfile{}, fmt.Errorf("unexpected GetCreatorProfileByUserID call")
	}

	return s.getCreatorProfileByUser(ctx, userID)
}

func (s stubQueries) GetUserProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppUserProfile, error) {
	if s.getUserProfileByUser == nil {
		return sqlc.AppUserProfile{}, fmt.Errorf("unexpected GetUserProfileByUserID call")
	}

	return s.getUserProfileByUser(ctx, userID)
}

func (s stubQueries) UpdateCreatorProfile(ctx context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
	if s.updateCreatorProfile == nil {
		return sqlc.AppCreatorProfile{}, fmt.Errorf("unexpected UpdateCreatorProfile call")
	}

	return s.updateCreatorProfile(ctx, arg)
}

func (s stubQueries) UpdateUserProfile(ctx context.Context, arg sqlc.UpdateUserProfileParams) (sqlc.AppUserProfile, error) {
	if s.updateUserProfile == nil {
		return sqlc.AppUserProfile{}, fmt.Errorf("unexpected UpdateUserProfile call")
	}

	return s.updateUserProfile(ctx, arg)
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

func (s *txStub) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

func (s *txStub) Conn() *pgx.Conn {
	return nil
}

func TestNewRepositoryInitializesDependencies(t *testing.T) {
	t.Parallel()

	repository := NewRepository(&pgxpool.Pool{})
	if repository.txBeginner == nil {
		t.Fatal("NewRepository() txBeginner = nil, want initialized")
	}
	if repository.queries == nil {
		t.Fatal("NewRepository() queries = nil, want initialized")
	}
	if repository.newQueries == nil {
		t.Fatal("NewRepository() newQueries = nil, want initialized")
	}
	if repository.newQueries(&pgxpool.Pool{}) == nil {
		t.Fatal("NewRepository() newQueries() = nil, want initialized queries")
	}
}

func TestNewRepositoryWithNilPoolReturnsEmptyRepository(t *testing.T) {
	t.Parallel()

	repository := NewRepository(nil)
	if repository.txBeginner != nil {
		t.Fatal("NewRepository(nil) txBeginner != nil, want nil")
	}
	if repository.queries != nil {
		t.Fatal("NewRepository(nil) queries != nil, want nil")
	}
	if repository.newQueries != nil {
		t.Fatal("NewRepository(nil) newQueries != nil, want nil")
	}
}

func TestNewRepositoryHelperInitializesQueryFactory(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{})
	if repository.queries == nil {
		t.Fatal("newRepository() queries = nil, want initialized")
	}
	if repository.newQueries == nil {
		t.Fatal("newRepository() newQueries = nil, want initialized")
	}
	if repository.newQueries(&pgxpool.Pool{}) == nil {
		t.Fatal("newRepository() newQueries() = nil, want initialized queries")
	}
}

func TestCreateProfileNormalizesInputAndReturnsProfile(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.New()
	avatarURL := "https://cdn.example.com/viewer/avatar.png"

	repository := newRepository(stubQueries{
		createUserProfile: func(_ context.Context, arg sqlc.CreateUserProfileParams) (sqlc.AppUserProfile, error) {
			if arg.UserID != postgres.UUIDToPG(userID) {
				t.Fatalf("CreateUserProfile() user id got %#v want %#v", arg.UserID, postgres.UUIDToPG(userID))
			}
			if arg.DisplayName != "Mina" {
				t.Fatalf("CreateUserProfile() display name got %q want %q", arg.DisplayName, "Mina")
			}
			if arg.Handle != "mina.01" {
				t.Fatalf("CreateUserProfile() handle got %q want %q", arg.Handle, "mina.01")
			}
			if arg.AvatarUrl != postgres.TextToPG(&avatarURL) {
				t.Fatalf("CreateUserProfile() avatar got %#v want %#v", arg.AvatarUrl, postgres.TextToPG(&avatarURL))
			}

			return validUserProfileRow(userID, "Mina", "mina.01", &avatarURL, now), nil
		},
	})

	got, err := repository.CreateProfile(context.Background(), CreateProfileInput{
		UserID:      userID,
		DisplayName: " Mina ",
		Handle:      "@MiNa.01",
		AvatarURL:   &avatarURL,
	})
	if err != nil {
		t.Fatalf("CreateProfile() error = %v, want nil", err)
	}
	if got.UserID != userID {
		t.Fatalf("CreateProfile() user id got %s want %s", got.UserID, userID)
	}
	if got.DisplayName != "Mina" {
		t.Fatalf("CreateProfile() display name got %q want %q", got.DisplayName, "Mina")
	}
	if got.Handle != "mina.01" {
		t.Fatalf("CreateProfile() handle got %q want %q", got.Handle, "mina.01")
	}
	if got.AvatarURL == nil || *got.AvatarURL != avatarURL {
		t.Fatalf("CreateProfile() avatar url got %v want %q", got.AvatarURL, avatarURL)
	}
	if got.CreatedAt != now {
		t.Fatalf("CreateProfile() created at got %s want %s", got.CreatedAt, now)
	}
}

func TestCreateProfileRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{})

	if _, err := repository.CreateProfile(context.Background(), CreateProfileInput{
		UserID:      uuid.New(),
		DisplayName: "   ",
		Handle:      "mina",
	}); !errors.Is(err, ErrInvalidDisplayName) {
		t.Fatalf("CreateProfile() display name error got %v want %v", err, ErrInvalidDisplayName)
	}

	if _, err := repository.CreateProfile(context.Background(), CreateProfileInput{
		UserID:      uuid.New(),
		DisplayName: "Mina",
		Handle:      "bad-handle",
	}); !errors.Is(err, ErrInvalidHandle) {
		t.Fatalf("CreateProfile() handle error got %v want %v", err, ErrInvalidHandle)
	}
}

func TestCreateProfileMapsHandleConflict(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{
		createUserProfile: func(context.Context, sqlc.CreateUserProfileParams) (sqlc.AppUserProfile, error) {
			return sqlc.AppUserProfile{}, &pgconn.PgError{
				Code:           "23505",
				ConstraintName: userProfilesHandleUniqueConstraint,
			}
		},
	})

	if _, err := repository.CreateProfile(context.Background(), CreateProfileInput{
		UserID:      uuid.New(),
		DisplayName: "Mina",
		Handle:      "mina",
	}); !errors.Is(err, ErrHandleAlreadyTaken) {
		t.Fatalf("CreateProfile() error got %v want %v", err, ErrHandleAlreadyTaken)
	}
}

func TestCreateProfileRequiresInitializedRepository(t *testing.T) {
	t.Parallel()

	var repository *Repository
	if _, err := repository.CreateProfile(context.Background(), CreateProfileInput{}); err == nil {
		t.Fatal("CreateProfile() error = nil, want initialization error")
	}
}

func TestCreateProfileRejectsInvalidPersistedRow(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{
		createUserProfile: func(context.Context, sqlc.CreateUserProfileParams) (sqlc.AppUserProfile, error) {
			return sqlc.AppUserProfile{
				UserID:      postgres.UUIDToPG(uuid.New()),
				DisplayName: "Mina",
				Handle:      "mina",
				UpdatedAt:   pgTime(time.Unix(1710000050, 0).UTC()),
			}, nil
		},
	})

	if _, err := repository.CreateProfile(context.Background(), CreateProfileInput{
		UserID:      uuid.New(),
		DisplayName: "Mina",
		Handle:      "mina",
	}); err == nil {
		t.Fatal("CreateProfile() error = nil, want row mapping error")
	}
}

func TestGetProfileReturnsSharedProfile(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000100, 0).UTC()
	userID := uuid.New()
	avatarURL := "https://cdn.example.com/viewer/avatar.png"

	repository := newRepository(stubQueries{
		getUserProfileByUser: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppUserProfile, error) {
			if gotUserID != postgres.UUIDToPG(userID) {
				t.Fatalf("GetUserProfileByUserID() user id got %#v want %#v", gotUserID, postgres.UUIDToPG(userID))
			}

			return validUserProfileRow(userID, "Mina", "mina", &avatarURL, now), nil
		},
	})

	got, err := repository.GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetProfile() error = %v, want nil", err)
	}
	if got.UserID != userID {
		t.Fatalf("GetProfile() user id got %s want %s", got.UserID, userID)
	}
	if got.Handle != "mina" {
		t.Fatalf("GetProfile() handle got %q want %q", got.Handle, "mina")
	}
}

func TestGetProfileNotFound(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{
		getUserProfileByUser: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
			return sqlc.AppUserProfile{}, pgx.ErrNoRows
		},
	})

	if _, err := repository.GetProfile(context.Background(), uuid.New()); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("GetProfile() error got %v want %v", err, ErrProfileNotFound)
	}
}

func TestGetProfileRequiresInitializedRepository(t *testing.T) {
	t.Parallel()

	var repository *Repository
	if _, err := repository.GetProfile(context.Background(), uuid.New()); err == nil {
		t.Fatal("GetProfile() error = nil, want initialization error")
	}
}

func TestGetProfileReturnsUnexpectedError(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{
		getUserProfileByUser: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
			return sqlc.AppUserProfile{}, fmt.Errorf("db unavailable")
		},
	})

	if _, err := repository.GetProfile(context.Background(), uuid.New()); err == nil {
		t.Fatal("GetProfile() error = nil, want unexpected error")
	}
}

func TestGetProfileRejectsInvalidPersistedRow(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{
		getUserProfileByUser: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
			return sqlc.AppUserProfile{
				UserID:      postgres.UUIDToPG(uuid.New()),
				DisplayName: "Mina",
				Handle:      "mina",
				UpdatedAt:   pgTime(time.Unix(1710000150, 0).UTC()),
			}, nil
		},
	})

	if _, err := repository.GetProfile(context.Background(), uuid.New()); err == nil {
		t.Fatal("GetProfile() error = nil, want row mapping error")
	}
}

func TestUpdateProfileWithoutCreatorMirrorKeepsSharedProfileOnly(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000200, 0).UTC()
	userID := uuid.New()
	avatarURL := "https://cdn.example.com/viewer/avatar.png"
	tx := &txStub{}
	beginner := &txBeginnerStub{tx: tx}

	repository := &Repository{
		txBeginner: beginner,
		newQueries: func(db sqlc.DBTX) queries {
			if db != tx {
				t.Fatalf("newQueries() db got %T want txStub", db)
			}

			return stubQueries{
				updateUserProfile: func(_ context.Context, arg sqlc.UpdateUserProfileParams) (sqlc.AppUserProfile, error) {
					if arg.DisplayName != "Mina" {
						t.Fatalf("UpdateUserProfile() display name got %q want %q", arg.DisplayName, "Mina")
					}
					if arg.Handle != "mina" {
						t.Fatalf("UpdateUserProfile() handle got %q want %q", arg.Handle, "mina")
					}
					if arg.AvatarUrl != postgres.TextToPG(&avatarURL) {
						t.Fatalf("UpdateUserProfile() avatar got %#v want %#v", arg.AvatarUrl, postgres.TextToPG(&avatarURL))
					}

					return validUserProfileRow(userID, "Mina", "mina", &avatarURL, now), nil
				},
				getCreatorProfileByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
				},
			}
		},
	}

	got, err := repository.UpdateProfile(context.Background(), UpdateProfileInput{
		UserID:      userID,
		DisplayName: " Mina ",
		Handle:      "@MiNa",
		AvatarURL:   &avatarURL,
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v, want nil", err)
	}
	if got.UserID != userID {
		t.Fatalf("UpdateProfile() user id got %s want %s", got.UserID, userID)
	}
	if !beginner.began {
		t.Fatal("UpdateProfile() did not begin transaction")
	}
	if !tx.committed {
		t.Fatal("UpdateProfile() did not commit transaction")
	}
	if tx.rolledBack {
		t.Fatal("UpdateProfile() rolled back successful transaction")
	}
}

func TestUpdateProfileRequiresInitializedRepository(t *testing.T) {
	t.Parallel()

	var repository *Repository
	if _, err := repository.UpdateProfile(context.Background(), UpdateProfileInput{}); err == nil {
		t.Fatal("UpdateProfile() error = nil, want initialization error")
	}
}

func TestUpdateCreatorProfileSyncUpdatesMirrorAndBio(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000300, 0).UTC()
	userID := uuid.New()
	avatarURL := "https://cdn.example.com/viewer/avatar.png"
	tx := &txStub{}

	repository := &Repository{
		txBeginner: &txBeginnerStub{tx: tx},
		newQueries: func(sqlc.DBTX) queries {
			return stubQueries{
				updateUserProfile: func(context.Context, sqlc.UpdateUserProfileParams) (sqlc.AppUserProfile, error) {
					return validUserProfileRow(userID, "Mina", "mina", &avatarURL, now), nil
				},
				getCreatorProfileByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return validCreatorProfileRow(userID, "Current", "current", nil, "current bio", now), nil
				},
				updateCreatorProfile: func(_ context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					if arg.DisplayName != postgres.TextToPG(stringPtr("Mina")) {
						t.Fatalf("UpdateCreatorProfile() display name got %#v want %#v", arg.DisplayName, postgres.TextToPG(stringPtr("Mina")))
					}
					if arg.Handle != "mina" {
						t.Fatalf("UpdateCreatorProfile() handle got %q want %q", arg.Handle, "mina")
					}
					if arg.AvatarUrl != postgres.TextToPG(&avatarURL) {
						t.Fatalf("UpdateCreatorProfile() avatar got %#v want %#v", arg.AvatarUrl, postgres.TextToPG(&avatarURL))
					}
					if arg.Bio != "next bio" {
						t.Fatalf("UpdateCreatorProfile() bio got %q want %q", arg.Bio, "next bio")
					}

					return validCreatorProfileRow(userID, "Mina", "mina", &avatarURL, "next bio", now), nil
				},
			}
		},
	}

	got, err := repository.UpdateCreatorProfileSync(context.Background(), UpdateProfileInput{
		UserID:      userID,
		DisplayName: "Mina",
		Handle:      "mina",
		AvatarURL:   &avatarURL,
	}, "next bio")
	if err != nil {
		t.Fatalf("UpdateCreatorProfileSync() error = %v, want nil", err)
	}
	if got.Handle != "mina" {
		t.Fatalf("UpdateCreatorProfileSync() handle got %q want %q", got.Handle, "mina")
	}
	if !tx.committed {
		t.Fatal("UpdateCreatorProfileSync() did not commit transaction")
	}
}

func TestUpdateProfileReturnsProfileNotFound(t *testing.T) {
	t.Parallel()

	tx := &txStub{}
	repository := &Repository{
		txBeginner: &txBeginnerStub{tx: tx},
		newQueries: func(sqlc.DBTX) queries {
			return stubQueries{
				updateUserProfile: func(context.Context, sqlc.UpdateUserProfileParams) (sqlc.AppUserProfile, error) {
					return sqlc.AppUserProfile{}, pgx.ErrNoRows
				},
			}
		},
	}

	if _, err := repository.UpdateProfile(context.Background(), UpdateProfileInput{
		UserID:      uuid.New(),
		DisplayName: "Mina",
		Handle:      "mina",
	}); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("UpdateProfile() error got %v want %v", err, ErrProfileNotFound)
	}
	if !tx.rolledBack {
		t.Fatal("UpdateProfile() did not roll back failed transaction")
	}
}

func TestUpdateProfileMapsCreatorMirrorHandleConflict(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000350, 0).UTC()
	userID := uuid.New()
	tx := &txStub{}

	repository := &Repository{
		txBeginner: &txBeginnerStub{tx: tx},
		newQueries: func(sqlc.DBTX) queries {
			return stubQueries{
				updateUserProfile: func(context.Context, sqlc.UpdateUserProfileParams) (sqlc.AppUserProfile, error) {
					return validUserProfileRow(userID, "Mina", "mina", nil, now), nil
				},
				getCreatorProfileByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return validCreatorProfileRow(userID, "Mina", "mina", nil, "current bio", now), nil
				},
				updateCreatorProfile: func(context.Context, sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, &pgconn.PgError{
						Code:           "23505",
						ConstraintName: creatorProfilesHandleUniqueConstraint,
					}
				},
			}
		},
	}

	if _, err := repository.UpdateProfile(context.Background(), UpdateProfileInput{
		UserID:      userID,
		DisplayName: "Mina",
		Handle:      "mina",
	}); !errors.Is(err, ErrHandleAlreadyTaken) {
		t.Fatalf("UpdateProfile() error got %v want %v", err, ErrHandleAlreadyTaken)
	}
	if !tx.rolledBack {
		t.Fatal("UpdateProfile() did not roll back failed transaction")
	}
}

func TestUpdateProfileRejectsInvalidPersistedRow(t *testing.T) {
	t.Parallel()

	tx := &txStub{}
	repository := &Repository{
		txBeginner: &txBeginnerStub{tx: tx},
		newQueries: func(sqlc.DBTX) queries {
			return stubQueries{
				updateUserProfile: func(context.Context, sqlc.UpdateUserProfileParams) (sqlc.AppUserProfile, error) {
					return sqlc.AppUserProfile{
						UserID:      postgres.UUIDToPG(uuid.New()),
						DisplayName: "Mina",
						Handle:      "mina",
						CreatedAt:   pgTime(time.Unix(1710000375, 0).UTC()),
					}, nil
				},
			}
		},
	}

	if _, err := repository.UpdateProfile(context.Background(), UpdateProfileInput{
		UserID:      uuid.New(),
		DisplayName: "Mina",
		Handle:      "mina",
	}); err == nil {
		t.Fatal("UpdateProfile() error = nil, want row mapping error")
	}
	if !tx.rolledBack {
		t.Fatal("UpdateProfile() did not roll back failed transaction")
	}
}

func TestUpdateProfileReturnsCreatorLookupError(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000380, 0).UTC()
	tx := &txStub{}

	repository := &Repository{
		txBeginner: &txBeginnerStub{tx: tx},
		newQueries: func(sqlc.DBTX) queries {
			return stubQueries{
				updateUserProfile: func(context.Context, sqlc.UpdateUserProfileParams) (sqlc.AppUserProfile, error) {
					return validUserProfileRow(uuid.New(), "Mina", "mina", nil, now), nil
				},
				getCreatorProfileByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, fmt.Errorf("creator lookup failed")
				},
			}
		},
	}

	if _, err := repository.UpdateProfile(context.Background(), UpdateProfileInput{
		UserID:      uuid.New(),
		DisplayName: "Mina",
		Handle:      "mina",
	}); err == nil {
		t.Fatal("UpdateProfile() error = nil, want creator lookup error")
	}
	if !tx.rolledBack {
		t.Fatal("UpdateProfile() did not roll back failed transaction")
	}
}

func TestUpdateCreatorProfileSyncReturnsCreatorProfileNotFoundWhenMirrorUpdateMisses(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000390, 0).UTC()
	userID := uuid.New()
	tx := &txStub{}

	repository := &Repository{
		txBeginner: &txBeginnerStub{tx: tx},
		newQueries: func(sqlc.DBTX) queries {
			return stubQueries{
				updateUserProfile: func(context.Context, sqlc.UpdateUserProfileParams) (sqlc.AppUserProfile, error) {
					return validUserProfileRow(userID, "Mina", "mina", nil, now), nil
				},
				getCreatorProfileByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return validCreatorProfileRow(userID, "Mina", "mina", nil, "current bio", now), nil
				},
				updateCreatorProfile: func(context.Context, sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
				},
			}
		},
	}

	if _, err := repository.UpdateCreatorProfileSync(context.Background(), UpdateProfileInput{
		UserID:      userID,
		DisplayName: "Mina",
		Handle:      "mina",
	}, "next bio"); !errors.Is(err, ErrCreatorProfileNotFound) {
		t.Fatalf("UpdateCreatorProfileSync() error got %v want %v", err, ErrCreatorProfileNotFound)
	}
	if !tx.rolledBack {
		t.Fatal("UpdateCreatorProfileSync() did not roll back failed transaction")
	}
}

func TestUpdateCreatorProfileSyncReturnsCreatorProfileNotFound(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000400, 0).UTC()
	userID := uuid.New()
	tx := &txStub{}

	repository := &Repository{
		txBeginner: &txBeginnerStub{tx: tx},
		newQueries: func(sqlc.DBTX) queries {
			return stubQueries{
				updateUserProfile: func(context.Context, sqlc.UpdateUserProfileParams) (sqlc.AppUserProfile, error) {
					return validUserProfileRow(userID, "Mina", "mina", nil, now), nil
				},
				getCreatorProfileByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
				},
			}
		},
	}

	if _, err := repository.UpdateCreatorProfileSync(context.Background(), UpdateProfileInput{
		UserID:      userID,
		DisplayName: "Mina",
		Handle:      "mina",
	}, "next bio"); !errors.Is(err, ErrCreatorProfileNotFound) {
		t.Fatalf("UpdateCreatorProfileSync() error got %v want %v", err, ErrCreatorProfileNotFound)
	}
	if tx.committed {
		t.Fatal("UpdateCreatorProfileSync() committed failed transaction")
	}
	if !tx.rolledBack {
		t.Fatal("UpdateCreatorProfileSync() did not roll back failed transaction")
	}
}

func TestMapProfileWriteErrorMapsCreatorHandleConflict(t *testing.T) {
	t.Parallel()

	got := mapProfileWriteError(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: creatorProfilesHandleUniqueConstraint,
	})
	if !errors.Is(got, ErrHandleAlreadyTaken) {
		t.Fatalf("mapProfileWriteError() got %v want %v", got, ErrHandleAlreadyTaken)
	}
}

func TestMapProfileWriteErrorReturnsOriginalError(t *testing.T) {
	t.Parallel()

	sourceErr := fmt.Errorf("write failed")
	if got := mapProfileWriteError(sourceErr); !errors.Is(got, sourceErr) {
		t.Fatalf("mapProfileWriteError() got %v want %v", got, sourceErr)
	}
}

func TestMapProfileRejectsNullUpdatedAt(t *testing.T) {
	t.Parallel()

	_, err := mapProfile(sqlc.AppUserProfile{
		UserID:      postgres.UUIDToPG(uuid.New()),
		DisplayName: "Mina",
		Handle:      "mina",
		CreatedAt:   pgTime(time.Unix(1710000500, 0).UTC()),
	})
	if err == nil {
		t.Fatal("mapProfile() error = nil, want error")
	}
}

func TestNormalizeProfileInputRejectsEmptyHandle(t *testing.T) {
	t.Parallel()

	if _, _, err := normalizeProfileInput("Mina", "@   "); !errors.Is(err, ErrInvalidHandle) {
		t.Fatalf("normalizeProfileInput() error got %v want %v", err, ErrInvalidHandle)
	}
}

func validUserProfileRow(userID uuid.UUID, displayName string, handle string, avatarURL *string, now time.Time) sqlc.AppUserProfile {
	return sqlc.AppUserProfile{
		UserID:      postgres.UUIDToPG(userID),
		DisplayName: displayName,
		Handle:      handle,
		AvatarUrl:   postgres.TextToPG(avatarURL),
		CreatedAt:   pgTime(now),
		UpdatedAt:   pgTime(now),
	}
}

func validCreatorProfileRow(userID uuid.UUID, displayName string, handle string, avatarURL *string, bio string, now time.Time) sqlc.AppCreatorProfile {
	return sqlc.AppCreatorProfile{
		UserID:      postgres.UUIDToPG(userID),
		DisplayName: postgres.TextToPG(stringPtr(displayName)),
		AvatarUrl:   postgres.TextToPG(avatarURL),
		Bio:         bio,
		PublishedAt: pgTime(now),
		CreatedAt:   pgTime(now),
		UpdatedAt:   pgTime(now),
		Handle:      handle,
	}
}

func pgTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}

func stringPtr(value string) *string {
	return &value
}
