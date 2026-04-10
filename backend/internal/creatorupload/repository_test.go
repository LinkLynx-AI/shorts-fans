package creatorupload

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
)

type repositoryQueriesStub struct {
	createMain               func(context.Context, sqlc.CreateMainParams) (sqlc.AppMain, error)
	createMediaAsset         func(context.Context, sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error)
	createMediaProcessingJob func(context.Context, sqlc.CreateMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error)
	createShort              func(context.Context, sqlc.CreateShortParams) (sqlc.AppShort, error)
}

func (s repositoryQueriesStub) CreateMain(ctx context.Context, arg sqlc.CreateMainParams) (sqlc.AppMain, error) {
	return s.createMain(ctx, arg)
}

func (s repositoryQueriesStub) CreateMediaAsset(ctx context.Context, arg sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error) {
	return s.createMediaAsset(ctx, arg)
}

func (s repositoryQueriesStub) CreateMediaProcessingJob(ctx context.Context, arg sqlc.CreateMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
	return s.createMediaProcessingJob(ctx, arg)
}

func (s repositoryQueriesStub) CreateShort(ctx context.Context, arg sqlc.CreateShortParams) (sqlc.AppShort, error) {
	return s.createShort(ctx, arg)
}

type repositoryTxBeginnerStub struct {
	begin func(context.Context) (pgx.Tx, error)
}

func (s repositoryTxBeginnerStub) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.begin(ctx)
}

type repositoryTxStub struct {
	commitErr   error
	rollbackErr error
	committed   bool
	rolledBack  bool
}

func (tx *repositoryTxStub) Begin(context.Context) (pgx.Tx, error) { return tx, nil }
func (tx *repositoryTxStub) Commit(context.Context) error {
	tx.committed = true
	return tx.commitErr
}
func (tx *repositoryTxStub) Rollback(context.Context) error {
	tx.rolledBack = true
	return tx.rollbackErr
}
func (tx *repositoryTxStub) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (tx *repositoryTxStub) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (tx *repositoryTxStub) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (tx *repositoryTxStub) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (tx *repositoryTxStub) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (tx *repositoryTxStub) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (tx *repositoryTxStub) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (tx *repositoryTxStub) Conn() *pgx.Conn                                         { return nil }

func TestNewRepository(t *testing.T) {
	t.Parallel()

	repository := NewRepository(nil)
	if repository == nil {
		t.Fatal("NewRepository() = nil, want non-nil")
	}
	if repository.newQueries == nil {
		t.Fatal("NewRepository() newQueries = nil, want non-nil")
	}
}

func TestRepositoryCreateDraftPackageSuccess(t *testing.T) {
	t.Parallel()

	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainAssetID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	shortID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	shortCaption := "preview caption"

	tx := &repositoryTxStub{}
	createMediaAssetCalls := 0
	createMediaProcessingJobCalls := 0
	repository := &Repository{
		beginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) {
				return tx, nil
			},
		},
		newQueries: func(db sqlc.DBTX) queries {
			if db != tx {
				t.Fatalf("newQueries() db got %v want %v", db, tx)
			}
			return repositoryQueriesStub{
				createMediaAsset: func(_ context.Context, arg sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error) {
					createMediaAssetCalls++
					if arg.CreatorUserID != postgres.UUIDToPG(creatorID) {
						t.Fatalf("CreateMediaAsset() creator id got %v want %v", arg.CreatorUserID, postgres.UUIDToPG(creatorID))
					}
					if arg.ProcessingState != stateUploaded {
						t.Fatalf("CreateMediaAsset() processing state got %q want %q", arg.ProcessingState, stateUploaded)
					}
					if arg.StorageProvider != storageProviderS3 {
						t.Fatalf("CreateMediaAsset() storage provider got %q want %q", arg.StorageProvider, storageProviderS3)
					}
					if arg.StorageBucket != "raw-bucket" {
						t.Fatalf("CreateMediaAsset() storage bucket got %q want %q", arg.StorageBucket, "raw-bucket")
					}
					switch createMediaAssetCalls {
					case 1:
						if arg.StorageKey != "main-key" {
							t.Fatalf("CreateMediaAsset() main storage key got %q want %q", arg.StorageKey, "main-key")
						}
						return sqlc.AppMediaAsset{
							ID:              postgres.UUIDToPG(mainAssetID),
							MimeType:        "video/mp4",
							ProcessingState: stateUploaded,
						}, nil
					case 2:
						if arg.StorageKey != "short-key" {
							t.Fatalf("CreateMediaAsset() short storage key got %q want %q", arg.StorageKey, "short-key")
						}
						return sqlc.AppMediaAsset{
							ID:              postgres.UUIDToPG(shortAssetID),
							MimeType:        "video/mp4",
							ProcessingState: stateUploaded,
						}, nil
					default:
						t.Fatalf("CreateMediaAsset() unexpected call count %d", createMediaAssetCalls)
						return sqlc.AppMediaAsset{}, nil
					}
				},
				createMain: func(_ context.Context, arg sqlc.CreateMainParams) (sqlc.AppMain, error) {
					if arg.MediaAssetID != postgres.UUIDToPG(mainAssetID) {
						t.Fatalf("CreateMain() media asset id got %v want %v", arg.MediaAssetID, postgres.UUIDToPG(mainAssetID))
					}
					if arg.State != stateDraft {
						t.Fatalf("CreateMain() state got %q want %q", arg.State, stateDraft)
					}
					if arg.PriceMinor != 1800 {
						t.Fatalf("CreateMain() price minor got %d want %d", arg.PriceMinor, 1800)
					}
					if arg.CurrencyCode != currencyJPY {
						t.Fatalf("CreateMain() currency code got %q want %q", arg.CurrencyCode, currencyJPY)
					}
					if !arg.OwnershipConfirmed {
						t.Fatal("CreateMain() ownership confirmed = false, want true")
					}
					if !arg.ConsentConfirmed {
						t.Fatal("CreateMain() consent confirmed = false, want true")
					}
					return sqlc.AppMain{
						ID:    postgres.UUIDToPG(mainID),
						State: stateDraft,
					}, nil
				},
				createMediaProcessingJob: func(_ context.Context, arg sqlc.CreateMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
					createMediaProcessingJobCalls++
					if arg.CreatorUserID != postgres.UUIDToPG(creatorID) {
						t.Fatalf("CreateMediaProcessingJob() creator id got %v want %v", arg.CreatorUserID, postgres.UUIDToPG(creatorID))
					}
					if arg.Status != processingJobStatusQueued {
						t.Fatalf("CreateMediaProcessingJob() status got %q want %q", arg.Status, processingJobStatusQueued)
					}
					if arg.AttemptCount != 0 {
						t.Fatalf("CreateMediaProcessingJob() attempt count got %d want 0", arg.AttemptCount)
					}
					switch createMediaProcessingJobCalls {
					case 1:
						if arg.MediaAssetID != postgres.UUIDToPG(mainAssetID) {
							t.Fatalf("CreateMediaProcessingJob() main asset id got %v want %v", arg.MediaAssetID, postgres.UUIDToPG(mainAssetID))
						}
						if arg.AssetRole != roleMain {
							t.Fatalf("CreateMediaProcessingJob() main asset role got %q want %q", arg.AssetRole, roleMain)
						}
					case 2:
						if arg.MediaAssetID != postgres.UUIDToPG(shortAssetID) {
							t.Fatalf("CreateMediaProcessingJob() short asset id got %v want %v", arg.MediaAssetID, postgres.UUIDToPG(shortAssetID))
						}
						if arg.AssetRole != roleShort {
							t.Fatalf("CreateMediaProcessingJob() short asset role got %q want %q", arg.AssetRole, roleShort)
						}
					default:
						t.Fatalf("CreateMediaProcessingJob() unexpected call count %d", createMediaProcessingJobCalls)
					}

					return sqlc.AppMediaProcessingJob{}, nil
				},
				createShort: func(_ context.Context, arg sqlc.CreateShortParams) (sqlc.AppShort, error) {
					if arg.CanonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("CreateShort() canonical main id got %v want %v", arg.CanonicalMainID, postgres.UUIDToPG(mainID))
					}
					if arg.MediaAssetID != postgres.UUIDToPG(shortAssetID) {
						t.Fatalf("CreateShort() media asset id got %v want %v", arg.MediaAssetID, postgres.UUIDToPG(shortAssetID))
					}
					if arg.Caption != postgres.TextToPG(&shortCaption) {
						t.Fatalf("CreateShort() caption got %#v want %#v", arg.Caption, postgres.TextToPG(&shortCaption))
					}
					if arg.State != stateDraft {
						t.Fatalf("CreateShort() state got %q want %q", arg.State, stateDraft)
					}
					return sqlc.AppShort{
						ID:              postgres.UUIDToPG(shortID),
						CanonicalMainID: postgres.UUIDToPG(mainID),
						State:           stateDraft,
					}, nil
				},
			}
		},
	}

	result, err := repository.CreateDraftPackage(context.Background(), createDraftPackageInput{
		CreatorUserID: creatorID,
		RawBucketName: "raw-bucket",
		Main: storedEntry{
			StorageKey:    "main-key",
			MimeType:      "video/mp4",
			UploadEntryID: "main-entry",
		},
		MainConsent:   true,
		MainOwnership: true,
		MainPriceJpy:  1800,
		Shorts: []createDraftShortInput{{
			Caption: &shortCaption,
			Entry: storedEntry{
				StorageKey:    "short-key",
				MimeType:      "video/mp4",
				UploadEntryID: "short-entry",
			},
		}},
	})
	if err != nil {
		t.Fatalf("CreateDraftPackage() error = %v, want nil", err)
	}
	if result.Main.ID != mainID {
		t.Fatalf("CreateDraftPackage() main id got %s want %s", result.Main.ID, mainID)
	}
	if len(result.Shorts) != 1 || result.Shorts[0].ID != shortID {
		t.Fatalf("CreateDraftPackage() shorts got %#v want short id %s", result.Shorts, shortID)
	}
	if !tx.committed {
		t.Fatal("CreateDraftPackage() committed = false, want true")
	}
	if tx.rolledBack {
		t.Fatal("CreateDraftPackage() rolledBack = true, want false")
	}
	if createMediaProcessingJobCalls != 2 {
		t.Fatalf("CreateDraftPackage() processing job calls got %d want 2", createMediaProcessingJobCalls)
	}
}

func TestRepositoryCreateDraftPackageErrors(t *testing.T) {
	t.Parallel()

	if _, err := (*Repository)(nil).CreateDraftPackage(context.Background(), createDraftPackageInput{}); err == nil {
		t.Fatal("CreateDraftPackage() error = nil, want error for nil repository")
	}

	createErr := errors.New("insert failed")
	tx := &repositoryTxStub{}
	repository := &Repository{
		beginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) {
				return tx, nil
			},
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				createMediaAsset: func(context.Context, sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error) {
					return sqlc.AppMediaAsset{}, createErr
				},
				createMain: func(context.Context, sqlc.CreateMainParams) (sqlc.AppMain, error) {
					return sqlc.AppMain{}, nil
				},
				createMediaProcessingJob: func(context.Context, sqlc.CreateMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
					return sqlc.AppMediaProcessingJob{}, nil
				},
				createShort: func(context.Context, sqlc.CreateShortParams) (sqlc.AppShort, error) {
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	_, err := repository.CreateDraftPackage(context.Background(), createDraftPackageInput{
		CreatorUserID: uuid.New(),
		RawBucketName: "raw-bucket",
		Main: storedEntry{
			UploadEntryID: "main-entry",
		},
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("CreateDraftPackage() error got %v want wrapped %v", err, createErr)
	}
	if !tx.rolledBack {
		t.Fatal("CreateDraftPackage() rolledBack = false, want true")
	}
}

func TestRepositoryCreateDraftPackageRollsBackWhenProcessingJobInsertFails(t *testing.T) {
	t.Parallel()

	createErr := errors.New("processing job insert failed")
	tx := &repositoryTxStub{}
	repository := &Repository{
		beginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) {
				return tx, nil
			},
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				createMediaAsset: func(_ context.Context, arg sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error) {
					return sqlc.AppMediaAsset{
						ID:              postgres.UUIDToPG(uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")),
						MimeType:        arg.MimeType,
						ProcessingState: arg.ProcessingState,
					}, nil
				},
				createMain: func(context.Context, sqlc.CreateMainParams) (sqlc.AppMain, error) {
					return sqlc.AppMain{
						ID:    postgres.UUIDToPG(uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")),
						State: stateDraft,
					}, nil
				},
				createMediaProcessingJob: func(context.Context, sqlc.CreateMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
					return sqlc.AppMediaProcessingJob{}, createErr
				},
				createShort: func(context.Context, sqlc.CreateShortParams) (sqlc.AppShort, error) {
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	_, err := repository.CreateDraftPackage(context.Background(), createDraftPackageInput{
		CreatorUserID: uuid.New(),
		RawBucketName: "raw-bucket",
		Main: storedEntry{
			UploadEntryID: "main-entry",
		},
		MainConsent:   true,
		MainOwnership: true,
		MainPriceJpy:  1800,
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("CreateDraftPackage() error got %v want wrapped %v", err, createErr)
	}
	if !tx.rolledBack {
		t.Fatal("CreateDraftPackage() rolledBack = false, want true")
	}
}

func TestRepositoryMappersRejectInvalidUUIDs(t *testing.T) {
	t.Parallel()

	assetID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	assetRow := sqlc.AppMediaAsset{
		ID:              postgres.UUIDToPG(assetID),
		MimeType:        "video/mp4",
		ProcessingState: stateUploaded,
	}

	if _, err := mapCreatedMediaAsset(sqlc.AppMediaAsset{}); err == nil {
		t.Fatal("mapCreatedMediaAsset() error = nil, want error")
	}
	if _, err := mapCreatedMain(sqlc.AppMain{}, assetRow); err == nil {
		t.Fatal("mapCreatedMain() error = nil, want error")
	}
	if _, err := mapCreatedShort(sqlc.AppShort{ID: pgtype.UUID{}}, assetRow); err == nil {
		t.Fatal("mapCreatedShort() error = nil, want error")
	}
}
