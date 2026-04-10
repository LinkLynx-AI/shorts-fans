package media

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type processorTxBeginnerStub struct {
	tx pgx.Tx
}

func (s processorTxBeginnerStub) Begin(context.Context) (pgx.Tx, error) {
	return s.tx, nil
}

type processorTxStub struct {
	committed  bool
	rolledBack bool
}

func (tx *processorTxStub) Begin(context.Context) (pgx.Tx, error) { return tx, nil }
func (tx *processorTxStub) Commit(context.Context) error {
	tx.committed = true
	return nil
}
func (tx *processorTxStub) Rollback(context.Context) error {
	tx.rolledBack = true
	return nil
}
func (tx *processorTxStub) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (tx *processorTxStub) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (tx *processorTxStub) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (tx *processorTxStub) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (tx *processorTxStub) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (tx *processorTxStub) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (tx *processorTxStub) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (tx *processorTxStub) Conn() *pgx.Conn                                         { return nil }

type processorQueriesStub struct {
	getMediaAssetByID                 func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error)
	getMediaProcessingJobByMediaAsset func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error)
	claimMediaProcessingJobByAsset    func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error)
	claimNextQueuedMediaProcessingJob func(context.Context) (sqlc.AppMediaProcessingJob, error)
	markMediaProcessingJobSucceeded   func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error)
	requeueMediaProcessingJob         func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error)
	markMediaProcessingJobFailed      func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error)
	updateMediaAssetProcessingState   func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error)
	getMainByID                       func(context.Context, pgtype.UUID) (sqlc.AppMain, error)
	getMainByMediaAssetID             func(context.Context, pgtype.UUID) (sqlc.AppMain, error)
	getShortByMediaAssetID            func(context.Context, pgtype.UUID) (sqlc.AppShort, error)
	listShortsByCanonicalMainID       func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error)
	updateMainState                   func(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error)
	publishShort                      func(context.Context, pgtype.UUID) (sqlc.AppShort, error)
}

func (s processorQueriesStub) GetMediaAssetByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
	return s.getMediaAssetByID(ctx, id)
}
func (s processorQueriesStub) UpdateMediaAssetProcessingState(ctx context.Context, arg sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
	return s.updateMediaAssetProcessingState(ctx, arg)
}
func (s processorQueriesStub) GetMediaProcessingJobByMediaAssetID(ctx context.Context, mediaAssetID pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
	return s.getMediaProcessingJobByMediaAsset(ctx, mediaAssetID)
}
func (s processorQueriesStub) ClaimMediaProcessingJobByAssetID(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
	return s.claimMediaProcessingJobByAsset(ctx, id)
}
func (s processorQueriesStub) ClaimNextQueuedMediaProcessingJob(ctx context.Context) (sqlc.AppMediaProcessingJob, error) {
	return s.claimNextQueuedMediaProcessingJob(ctx)
}
func (s processorQueriesStub) MarkMediaProcessingJobSucceeded(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
	return s.markMediaProcessingJobSucceeded(ctx, id)
}
func (s processorQueriesStub) RequeueMediaProcessingJob(ctx context.Context, arg sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
	return s.requeueMediaProcessingJob(ctx, arg)
}
func (s processorQueriesStub) MarkMediaProcessingJobFailed(ctx context.Context, arg sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
	return s.markMediaProcessingJobFailed(ctx, arg)
}
func (s processorQueriesStub) GetMainByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
	return s.getMainByID(ctx, id)
}
func (s processorQueriesStub) GetMainByMediaAssetID(ctx context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
	return s.getMainByMediaAssetID(ctx, id)
}
func (s processorQueriesStub) GetShortByMediaAssetID(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
	return s.getShortByMediaAssetID(ctx, id)
}
func (s processorQueriesStub) ListShortsByCanonicalMainID(ctx context.Context, id pgtype.UUID) ([]sqlc.AppShort, error) {
	return s.listShortsByCanonicalMainID(ctx, id)
}
func (s processorQueriesStub) UpdateMainState(ctx context.Context, arg sqlc.UpdateMainStateParams) (sqlc.AppMain, error) {
	return s.updateMainState(ctx, arg)
}
func (s processorQueriesStub) PublishShort(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
	return s.publishShort(ctx, id)
}

type stubAssetMaterializer struct {
	err error
}

func (s stubAssetMaterializer) Materialize(context.Context, MaterializeRequest) (MaterializeResult, error) {
	return MaterializeResult{}, s.err
}

func TestProcessClaimedJobRequeuesGenericMaterializationError(t *testing.T) {
	t.Parallel()

	mediaAssetID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	jobID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	requeueCalls := 0

	queries := processorQueriesStub{
		getMediaProcessingJobByMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
			if id != pgUUID(mediaAssetID) {
				t.Fatalf("GetMediaProcessingJobByMediaAssetID() id got %v want %v", id, pgUUID(mediaAssetID))
			}
			return sqlc.AppMediaProcessingJob{
				ID:           pgUUID(jobID),
				MediaAssetID: pgUUID(mediaAssetID),
				AttemptCount: 1,
			}, nil
		},
		getMediaAssetByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			if id != pgUUID(mediaAssetID) {
				t.Fatalf("GetMediaAssetByID() id got %v want %v", id, pgUUID(mediaAssetID))
			}
			return sqlc.AppMediaAsset{
				ID:                pgUUID(mediaAssetID),
				ExternalUploadRef: pgText(nil),
			}, nil
		},
		requeueMediaProcessingJob: func(_ context.Context, arg sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
			requeueCalls++
			if arg.ID != pgUUID(jobID) {
				t.Fatalf("RequeueMediaProcessingJob() id got %v want %v", arg.ID, pgUUID(jobID))
			}
			if got := postgres.OptionalTextFromPG(arg.LastErrorCode); got == nil || *got != materializationFailedCode {
				t.Fatalf("RequeueMediaProcessingJob() code got %v want %q", got, materializationFailedCode)
			}
			if got := postgres.OptionalTextFromPG(arg.LastErrorMessage); got == nil || !strings.Contains(*got, "copy poster failed") {
				t.Fatalf("RequeueMediaProcessingJob() message got %v want contains %q", got, "copy poster failed")
			}
			return sqlc.AppMediaProcessingJob{}, nil
		},
		markMediaProcessingJobFailed: func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
			t.Fatal("MarkMediaProcessingJobFailed() should not be called")
			return sqlc.AppMediaProcessingJob{}, nil
		},
		updateMediaAssetProcessingState: func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
			t.Fatal("UpdateMediaAssetProcessingState() should not be called for retryable error")
			return sqlc.AppMediaAsset{}, nil
		},
	}

	tx := &processorTxStub{}
	processor, err := newProcessor(
		processorTxBeginnerStub{tx: tx},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{err: errors.New("copy poster failed")},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}

	err = processor.processClaimedJob(context.Background(), claimedJob{
		job:   sqlc.AppMediaProcessingJob{ID: pgUUID(jobID), MediaAssetID: pgUUID(mediaAssetID)},
		asset: Asset{ID: mediaAssetID, StorageBucket: "raw-bucket", StorageKey: "raw/input.mp4"},
		role:  roleShort,
	})
	if err != nil {
		t.Fatalf("processClaimedJob() error = %v, want nil", err)
	}
	if requeueCalls != 1 {
		t.Fatalf("processClaimedJob() requeueCalls got %d want 1", requeueCalls)
	}
	if !tx.committed {
		t.Fatal("processClaimedJob() committed = false, want true")
	}
}

func TestDetachProcessingContextPreservesDeadline(t *testing.T) {
	t.Parallel()

	parentDeadline := time.Now().Add(2 * time.Minute)
	parent, parentCancel := context.WithDeadline(context.Background(), parentDeadline)
	defer parentCancel()

	detached, cancel := detachProcessingContext(parent)
	defer cancel()

	if err := detached.Err(); err != nil {
		t.Fatalf("detachProcessingContext() err = %v, want nil", err)
	}
	deadline, ok := detached.Deadline()
	if !ok {
		t.Fatal("detachProcessingContext() deadline missing, want preserved deadline")
	}
	if !deadline.Equal(parentDeadline) {
		t.Fatalf("detachProcessingContext() deadline got %v want %v", deadline, parentDeadline)
	}

	parentCancel()
	if err := detached.Err(); err != nil {
		t.Fatalf("detachProcessingContext() err after parent cancel = %v, want nil", err)
	}
}

func TestDetachProcessingContextAddsTimeoutWhenParentHasNoDeadline(t *testing.T) {
	t.Parallel()

	detached, cancel := detachProcessingContext(context.Background())
	defer cancel()

	deadline, ok := detached.Deadline()
	if !ok {
		t.Fatal("detachProcessingContext() deadline missing, want cleanup timeout")
	}

	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > cleanupPersistTimeout {
		t.Fatalf("detachProcessingContext() remaining timeout got %s want within (0,%s]", remaining, cleanupPersistTimeout)
	}
}

func TestProcessAssetReturnsNilWhenNoQueuedJob(t *testing.T) {
	t.Parallel()

	mediaAssetID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	queries := processorQueriesStub{
		claimMediaProcessingJobByAsset: func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
			return sqlc.AppMediaProcessingJob{}, pgx.ErrNoRows
		},
		getMediaProcessingJobByMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
			if id != pgUUID(mediaAssetID) {
				t.Fatalf("GetMediaProcessingJobByMediaAssetID() id got %v want %v", id, pgUUID(mediaAssetID))
			}
			return sqlc.AppMediaProcessingJob{Status: jobStatusQueued}, nil
		},
		getMediaAssetByID: func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error) {
			t.Fatal("GetMediaAssetByID() should not be called")
			return sqlc.AppMediaAsset{}, nil
		},
		requeueMediaProcessingJob: func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
			t.Fatal("RequeueMediaProcessingJob() should not be called")
			return sqlc.AppMediaProcessingJob{}, nil
		},
		markMediaProcessingJobFailed: func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
			t.Fatal("MarkMediaProcessingJobFailed() should not be called")
			return sqlc.AppMediaProcessingJob{}, nil
		},
		updateMediaAssetProcessingState: func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
			t.Fatal("UpdateMediaAssetProcessingState() should not be called")
			return sqlc.AppMediaAsset{}, nil
		},
		claimNextQueuedMediaProcessingJob: func(context.Context) (sqlc.AppMediaProcessingJob, error) {
			t.Fatal("ClaimNextQueuedMediaProcessingJob() should not be called")
			return sqlc.AppMediaProcessingJob{}, nil
		},
		markMediaProcessingJobSucceeded: func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
			t.Fatal("MarkMediaProcessingJobSucceeded() should not be called")
			return sqlc.AppMediaProcessingJob{}, nil
		},
		getMainByID:                 func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") },
		getMainByMediaAssetID:       func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") },
		getShortByMediaAssetID:      func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") },
		listShortsByCanonicalMainID: func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) { panic("unexpected call") },
		updateMainState:             func(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error) { panic("unexpected call") },
		publishShort:                func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") },
	}

	processor, err := newProcessor(
		processorTxBeginnerStub{tx: &processorTxStub{}},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}

	if err := processor.ProcessAsset(context.Background(), mediaAssetID); err != nil {
		t.Fatalf("ProcessAsset() error = %v, want nil", err)
	}
}

func TestClaimNextQueuedLoadsShortContext(t *testing.T) {
	t.Parallel()

	mediaAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	shortID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	mainID := uuid.MustParse("66666666-6666-6666-6666-666666666666")

	queries := processorQueriesStub{
		claimNextQueuedMediaProcessingJob: func(context.Context) (sqlc.AppMediaProcessingJob, error) {
			return sqlc.AppMediaProcessingJob{
				ID:           pgUUID(uuid.MustParse("77777777-7777-7777-7777-777777777777")),
				MediaAssetID: pgUUID(mediaAssetID),
				AssetRole:    roleShort,
			}, nil
		},
		getMediaAssetByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			if id != pgUUID(mediaAssetID) {
				t.Fatalf("GetMediaAssetByID() id got %v want %v", id, pgUUID(mediaAssetID))
			}
			return sqlc.AppMediaAsset{
				ID:                id,
				CreatorUserID:     pgUUID(uuid.MustParse("88888888-8888-8888-8888-888888888888")),
				ProcessingState:   assetStateUploaded,
				StorageProvider:   "s3",
				StorageBucket:     "raw-bucket",
				StorageKey:        "raw/input.mp4",
				MimeType:          "video/mp4",
				ExternalUploadRef: pgText(nil),
				CreatedAt:         pgTime(time.Now().UTC()),
				UpdatedAt:         pgTime(time.Now().UTC()),
			}, nil
		},
		updateMediaAssetProcessingState: func(_ context.Context, arg sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
			if got, want := arg.ProcessingState, assetStateWorking; got != want {
				t.Fatalf("UpdateMediaAssetProcessingState() state got %q want %q", got, want)
			}
			return sqlc.AppMediaAsset{
				ID:                arg.ID,
				CreatorUserID:     pgUUID(uuid.MustParse("88888888-8888-8888-8888-888888888888")),
				ProcessingState:   assetStateWorking,
				StorageProvider:   "s3",
				StorageBucket:     "raw-bucket",
				StorageKey:        "raw/input.mp4",
				MimeType:          "video/mp4",
				ExternalUploadRef: pgText(nil),
				CreatedAt:         pgTime(time.Now().UTC()),
				UpdatedAt:         pgTime(time.Now().UTC()),
			}, nil
		},
		getShortByMediaAssetID: func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
			if id != pgUUID(mediaAssetID) {
				t.Fatalf("GetShortByMediaAssetID() id got %v want %v", id, pgUUID(mediaAssetID))
			}
			return sqlc.AppShort{
				ID:              pgUUID(shortID),
				CanonicalMainID: pgUUID(mainID),
			}, nil
		},
		claimMediaProcessingJobByAsset:    func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		getMediaProcessingJobByMediaAsset: func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		requeueMediaProcessingJob: func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
			panic("unexpected call")
		},
		markMediaProcessingJobFailed: func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
			panic("unexpected call")
		},
		markMediaProcessingJobSucceeded: func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		getMainByID:                     func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") },
		getMainByMediaAssetID:           func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") },
		listShortsByCanonicalMainID:     func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) { panic("unexpected call") },
		updateMainState:                 func(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error) { panic("unexpected call") },
		publishShort:                    func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") },
	}

	processor, err := newProcessor(
		processorTxBeginnerStub{tx: &processorTxStub{}},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}

	claimed, err := processor.claimNextQueued(context.Background())
	if err != nil {
		t.Fatalf("claimNextQueued() error = %v, want nil", err)
	}
	if claimed.shortID != shortID {
		t.Fatalf("claimNextQueued() short id got %s want %s", claimed.shortID, shortID)
	}
	if claimed.canonicalMainID != mainID {
		t.Fatalf("claimNextQueued() main id got %s want %s", claimed.canonicalMainID, mainID)
	}
	if got, want := claimed.asset.ProcessingState, assetStateWorking; got != want {
		t.Fatalf("claimNextQueued() asset state got %q want %q", got, want)
	}
}

func TestMarkSucceededUpdatesAssetAndJob(t *testing.T) {
	t.Parallel()

	mediaAssetID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	jobID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000000")
	updateCalls := 0
	successCalls := 0
	playbackURL := "https://cdn.example.com/shorts/ready.mp4"
	durationMS := int64(42000)

	queries := processorQueriesStub{
		getMediaAssetByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			return sqlc.AppMediaAsset{ID: id, ExternalUploadRef: pgText(nil)}, nil
		},
		updateMediaAssetProcessingState: func(_ context.Context, arg sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
			updateCalls++
			if got, want := arg.ProcessingState, assetStateReady; got != want {
				t.Fatalf("UpdateMediaAssetProcessingState() state got %q want %q", got, want)
			}
			if got := postgres.OptionalTextFromPG(arg.PlaybackUrl); got == nil || *got != playbackURL {
				t.Fatalf("UpdateMediaAssetProcessingState() playback got %v want %q", got, playbackURL)
			}
			if got := postgres.OptionalInt64FromPG(arg.DurationMs); got == nil || *got != durationMS {
				t.Fatalf("UpdateMediaAssetProcessingState() duration got %v want %d", got, durationMS)
			}
			return sqlc.AppMediaAsset{}, nil
		},
		markMediaProcessingJobSucceeded: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
			successCalls++
			if id != pgUUID(jobID) {
				t.Fatalf("MarkMediaProcessingJobSucceeded() id got %v want %v", id, pgUUID(jobID))
			}
			return sqlc.AppMediaProcessingJob{}, nil
		},
		claimMediaProcessingJobByAsset:    func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		claimNextQueuedMediaProcessingJob: func(context.Context) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		getMediaProcessingJobByMediaAsset: func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		requeueMediaProcessingJob: func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
			panic("unexpected call")
		},
		markMediaProcessingJobFailed: func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
			panic("unexpected call")
		},
		getMainByID:                 func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") },
		getMainByMediaAssetID:       func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") },
		getShortByMediaAssetID:      func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") },
		listShortsByCanonicalMainID: func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) { panic("unexpected call") },
		updateMainState:             func(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error) { panic("unexpected call") },
		publishShort:                func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") },
	}

	processor, err := newProcessor(
		processorTxBeginnerStub{tx: &processorTxStub{}},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}

	err = processor.markSucceeded(context.Background(), claimedJob{
		job:             sqlc.AppMediaProcessingJob{ID: pgUUID(jobID)},
		asset:           Asset{ID: mediaAssetID},
		canonicalMainID: uuid.Nil,
	}, MaterializeResult{
		PlaybackURL: playbackURL,
		DurationMS:  durationMS,
	})
	if err != nil {
		t.Fatalf("markSucceeded() error = %v, want nil", err)
	}
	if updateCalls != 1 || successCalls != 1 {
		t.Fatalf("markSucceeded() calls got update=%d success=%d want 1/1", updateCalls, successCalls)
	}
}

func TestRequeueFailedAssetResetsState(t *testing.T) {
	t.Parallel()

	mediaAssetID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000000")
	jobID := uuid.MustParse("cccccccc-0000-0000-0000-000000000000")
	requeueCalls := 0
	updateCalls := 0

	queries := processorQueriesStub{
		getMediaProcessingJobByMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
			return sqlc.AppMediaProcessingJob{ID: pgUUID(jobID), MediaAssetID: id}, nil
		},
		getMediaAssetByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			ref := "entry-1"
			return sqlc.AppMediaAsset{ID: id, ExternalUploadRef: pgText(&ref)}, nil
		},
		updateMediaAssetProcessingState: func(_ context.Context, arg sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
			updateCalls++
			if got, want := arg.ProcessingState, assetStateUploaded; got != want {
				t.Fatalf("UpdateMediaAssetProcessingState() state got %q want %q", got, want)
			}
			if got := postgres.OptionalTextFromPG(arg.PlaybackUrl); got != nil {
				t.Fatalf("UpdateMediaAssetProcessingState() playback got %v want nil", got)
			}
			if got := postgres.OptionalInt64FromPG(arg.DurationMs); got != nil {
				t.Fatalf("UpdateMediaAssetProcessingState() duration got %v want nil", got)
			}
			return sqlc.AppMediaAsset{}, nil
		},
		requeueMediaProcessingJob: func(_ context.Context, arg sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
			requeueCalls++
			if arg.ID != pgUUID(jobID) {
				t.Fatalf("RequeueMediaProcessingJob() id got %v want %v", arg.ID, pgUUID(jobID))
			}
			return sqlc.AppMediaProcessingJob{}, nil
		},
		claimMediaProcessingJobByAsset:    func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		claimNextQueuedMediaProcessingJob: func(context.Context) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		markMediaProcessingJobSucceeded:   func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		markMediaProcessingJobFailed: func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
			panic("unexpected call")
		},
		getMainByID:                 func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") },
		getMainByMediaAssetID:       func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") },
		getShortByMediaAssetID:      func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") },
		listShortsByCanonicalMainID: func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) { panic("unexpected call") },
		updateMainState:             func(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error) { panic("unexpected call") },
		publishShort:                func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") },
	}

	processor, err := newProcessor(
		processorTxBeginnerStub{tx: &processorTxStub{}},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}

	if err := processor.RequeueFailedAsset(context.Background(), mediaAssetID); err != nil {
		t.Fatalf("RequeueFailedAsset() error = %v, want nil", err)
	}
	if updateCalls != 1 || requeueCalls != 1 {
		t.Fatalf("RequeueFailedAsset() calls got update=%d requeue=%d want 1/1", updateCalls, requeueCalls)
	}
}

func TestProcessNextQueuedReturnsFalseWhenQueueIsEmpty(t *testing.T) {
	t.Parallel()

	queries := processorQueriesStub{
		claimNextQueuedMediaProcessingJob: func(context.Context) (sqlc.AppMediaProcessingJob, error) {
			return sqlc.AppMediaProcessingJob{}, pgx.ErrNoRows
		},
		claimMediaProcessingJobByAsset:    func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		getMediaAssetByID:                 func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error) { panic("unexpected call") },
		getMediaProcessingJobByMediaAsset: func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		requeueMediaProcessingJob: func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
			panic("unexpected call")
		},
		markMediaProcessingJobFailed: func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
			panic("unexpected call")
		},
		updateMediaAssetProcessingState: func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
			panic("unexpected call")
		},
		markMediaProcessingJobSucceeded: func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") },
		getMainByID:                     func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") },
		getMainByMediaAssetID:           func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") },
		getShortByMediaAssetID:          func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") },
		listShortsByCanonicalMainID:     func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) { panic("unexpected call") },
		updateMainState:                 func(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error) { panic("unexpected call") },
		publishShort:                    func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") },
	}

	processor, err := newProcessor(
		processorTxBeginnerStub{tx: &processorTxStub{}},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}

	processed, err := processor.ProcessNextQueued(context.Background())
	if err != nil {
		t.Fatalf("ProcessNextQueued() error = %v, want nil", err)
	}
	if processed {
		t.Fatal("ProcessNextQueued() processed = true, want false")
	}
}
