package media

import (
	"context"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestMarkSucceededDoesNotAdvanceReviewState(t *testing.T) {
	t.Parallel()

	mediaAssetID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	jobID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	mainID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	playbackURL := "https://cdn.example.com/ready.m3u8"
	durationMS := int64(33000)
	updateCalls := 0
	successCalls := 0

	queries := processorQueriesStub{
		getMediaAssetByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			if id != pgUUID(mediaAssetID) {
				t.Fatalf("GetMediaAssetByID() id got %v want %v", id, pgUUID(mediaAssetID))
			}

			externalRef := "upload-1"
			return sqlc.AppMediaAsset{
				ID:                id,
				ExternalUploadRef: postgres.TextToPG(&externalRef),
			}, nil
		},
		updateMediaAssetProcessingState: func(_ context.Context, arg sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
			updateCalls++
			if arg.ID != pgUUID(mediaAssetID) {
				t.Fatalf("UpdateMediaAssetProcessingState() id got %v want %v", arg.ID, pgUUID(mediaAssetID))
			}
			if arg.ProcessingState != assetStateReady {
				t.Fatalf("UpdateMediaAssetProcessingState() state got %q want %q", arg.ProcessingState, assetStateReady)
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
		canonicalMainID: mainID,
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
