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
	requeueMediaProcessingJob         func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error)
	markMediaProcessingJobFailed      func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error)
	updateMediaAssetProcessingState   func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error)
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
func (s processorQueriesStub) ClaimMediaProcessingJobByAssetID(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
	panic("unexpected call")
}
func (s processorQueriesStub) ClaimNextQueuedMediaProcessingJob(context.Context) (sqlc.AppMediaProcessingJob, error) {
	panic("unexpected call")
}
func (s processorQueriesStub) MarkMediaProcessingJobSucceeded(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
	panic("unexpected call")
}
func (s processorQueriesStub) RequeueMediaProcessingJob(ctx context.Context, arg sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
	return s.requeueMediaProcessingJob(ctx, arg)
}
func (s processorQueriesStub) MarkMediaProcessingJobFailed(ctx context.Context, arg sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
	return s.markMediaProcessingJobFailed(ctx, arg)
}
func (s processorQueriesStub) GetMainByID(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
	panic("unexpected call")
}
func (s processorQueriesStub) GetMainByMediaAssetID(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
	panic("unexpected call")
}
func (s processorQueriesStub) GetShortByMediaAssetID(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
	panic("unexpected call")
}
func (s processorQueriesStub) ListShortsByCanonicalMainID(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
	panic("unexpected call")
}
func (s processorQueriesStub) UpdateMainState(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error) {
	panic("unexpected call")
}
func (s processorQueriesStub) PublishShort(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
	panic("unexpected call")
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
