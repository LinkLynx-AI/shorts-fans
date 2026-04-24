package media

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

type recordingReviewSubmitter struct {
	calls         int
	creatorUserID uuid.UUID
	mainID        uuid.UUID
	err           error
	errs          []error
}

func (s *recordingReviewSubmitter) SubmitInitialPackageIfReady(_ context.Context, viewerUserID uuid.UUID, mainID uuid.UUID) error {
	s.calls++
	s.creatorUserID = viewerUserID
	s.mainID = mainID
	if len(s.errs) > 0 {
		err := s.errs[0]
		s.errs = s.errs[1:]
		return err
	}
	return s.err
}

func readyInitialReviewGateQueries(
	t testing.TB,
	creatorUserID uuid.UUID,
	mainID uuid.UUID,
	mainAssetID uuid.UUID,
	shortID uuid.UUID,
	shortAssetID uuid.UUID,
	shortAssetState string,
) processorQueriesStub {
	t.Helper()

	return processorQueriesStub{
		getMediaAssetByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			switch id {
			case pgUUID(mainAssetID):
				return sqlc.AppMediaAsset{
					ID:                id,
					CreatorUserID:     pgUUID(creatorUserID),
					ProcessingState:   assetStateReady,
					ExternalUploadRef: pgText(nil),
					CreatedAt:         pgtype.Timestamptz{Valid: true},
					UpdatedAt:         pgtype.Timestamptz{Valid: true},
				}, nil
			case pgUUID(shortAssetID):
				return sqlc.AppMediaAsset{
					ID:              id,
					CreatorUserID:   pgUUID(creatorUserID),
					ProcessingState: shortAssetState,
					CreatedAt:       pgtype.Timestamptz{Valid: true},
					UpdatedAt:       pgtype.Timestamptz{Valid: true},
				}, nil
			default:
				t.Fatalf("GetMediaAssetByID() id got %v want main=%v or short=%v", id, pgUUID(mainAssetID), pgUUID(shortAssetID))
				return sqlc.AppMediaAsset{}, nil
			}
		},
		getInitialReviewReadyPackageByMainID: func(_ context.Context, id pgtype.UUID) (sqlc.GetInitialReviewReadyPackageByMainIDRow, error) {
			if id != pgUUID(mainID) {
				t.Fatalf("GetInitialReviewReadyPackageByMainID() id got %v want %v", id, pgUUID(mainID))
			}
			if shortAssetState != assetStateReady {
				return sqlc.GetInitialReviewReadyPackageByMainIDRow{}, pgx.ErrNoRows
			}
			return sqlc.GetInitialReviewReadyPackageByMainIDRow{
				ID:            id,
				CreatorUserID: pgUUID(creatorUserID),
			}, nil
		},
	}
}

func reviewSubmitFailedPGText() pgtype.Text {
	code := reviewSubmitFailedCode
	return postgres.TextToPG(&code)
}

func TestNewProcessorValidation(t *testing.T) {
	t.Parallel()

	queries := processorQueriesStub{}
	factory := func(sqlc.DBTX) processorQueries { return queries }
	beginner := processorTxBeginnerStub{tx: &processorTxStub{}}

	tests := []struct {
		name         string
		beginner     postgres.TxBeginner
		queries      processorQueries
		newQueries   func(sqlc.DBTX) processorQueries
		materializer assetMaterializer
		wantMessage  string
	}{
		{
			name:         "missing beginner",
			beginner:     nil,
			queries:      queries,
			newQueries:   factory,
			materializer: stubAssetMaterializer{},
			wantMessage:  "tx beginner is required",
		},
		{
			name:         "missing queries",
			beginner:     beginner,
			queries:      nil,
			newQueries:   factory,
			materializer: stubAssetMaterializer{},
			wantMessage:  "processor queries are required",
		},
		{
			name:         "missing transaction factory",
			beginner:     beginner,
			queries:      queries,
			newQueries:   nil,
			materializer: stubAssetMaterializer{},
			wantMessage:  "processor transaction queries factory is required",
		},
		{
			name:         "missing materializer",
			beginner:     beginner,
			queries:      queries,
			newQueries:   factory,
			materializer: nil,
			wantMessage:  "asset materializer is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := newProcessor(tt.beginner, tt.queries, tt.newQueries, tt.materializer)
			if err == nil {
				t.Fatalf("newProcessor() error = nil, want %q", tt.wantMessage)
			}
			if processor != nil {
				t.Fatalf("newProcessor() processor = %v, want nil", processor)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("newProcessor() error got %q want contains %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestNewProcessorRejectsNilPool(t *testing.T) {
	t.Parallel()

	processor, err := NewProcessor(nil, stubAssetMaterializer{})
	if err == nil {
		t.Fatal("NewProcessor() error = nil, want non-nil")
	}
	if processor != nil {
		t.Fatalf("NewProcessor() processor = %v, want nil", processor)
	}
	if !strings.Contains(err.Error(), "postgres pool is required") {
		t.Fatalf("NewProcessor() error got %q want postgres pool validation", err.Error())
	}
}

func TestSetReviewSubmitterAllowsNilReceiver(t *testing.T) {
	t.Parallel()

	var processor *Processor
	processor.SetReviewSubmitter(&recordingReviewSubmitter{})
}

func TestProcessClaimedJobSubmitsReadyPackageForReview(t *testing.T) {
	t.Parallel()

	creatorUserID := uuid.MustParse("11111111-aaaa-aaaa-aaaa-111111111111")
	mediaAssetID := uuid.MustParse("22222222-aaaa-aaaa-aaaa-222222222222")
	jobID := uuid.MustParse("33333333-aaaa-aaaa-aaaa-333333333333")
	mainID := uuid.MustParse("44444444-aaaa-aaaa-aaaa-444444444444")
	shortID := uuid.MustParse("55555555-aaaa-aaaa-aaaa-555555555555")
	shortAssetID := uuid.MustParse("66666666-aaaa-aaaa-aaaa-666666666666")
	submitter := &recordingReviewSubmitter{}

	queries := readyInitialReviewGateQueries(t, creatorUserID, mainID, mediaAssetID, shortID, shortAssetID, assetStateReady)
	queries.updateMediaAssetProcessingState = func(_ context.Context, arg sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
		if arg.ID != pgUUID(mediaAssetID) {
			t.Fatalf("UpdateMediaAssetProcessingState() id got %v want %v", arg.ID, pgUUID(mediaAssetID))
		}
		if arg.ProcessingState != assetStateReady {
			t.Fatalf("UpdateMediaAssetProcessingState() state got %q want %q", arg.ProcessingState, assetStateReady)
		}
		return sqlc.AppMediaAsset{}, nil
	}
	queries.markMediaProcessingJobSucceeded = func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
		if id != pgUUID(jobID) {
			t.Fatalf("MarkMediaProcessingJobSucceeded() id got %v want %v", id, pgUUID(jobID))
		}
		return sqlc.AppMediaProcessingJob{}, nil
	}
	queries.claimMediaProcessingJobByAsset = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.claimNextQueuedMediaProcessingJob = func(context.Context) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.getMediaProcessingJobByMediaAsset = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.requeueMediaProcessingJob = func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.markMediaProcessingJobFailed = func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.getMainByMediaAssetID = func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") }
	queries.getShortByMediaAssetID = func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") }

	processor, err := newProcessor(
		processorTxBeginnerStub{tx: &processorTxStub{}},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}
	processor.SetReviewSubmitter(submitter)

	err = processor.processClaimedJob(context.Background(), claimedJob{
		job: sqlc.AppMediaProcessingJob{
			ID:           pgUUID(jobID),
			MediaAssetID: pgUUID(mediaAssetID),
		},
		asset: Asset{
			ID:            mediaAssetID,
			CreatorUserID: creatorUserID,
			StorageBucket: "raw-bucket",
			StorageKey:    "raw/main.mp4",
		},
		role:            roleMain,
		mainID:          mainID,
		canonicalMainID: mainID,
	})
	if err != nil {
		t.Fatalf("processClaimedJob() error = %v, want nil", err)
	}
	if submitter.calls != 1 {
		t.Fatalf("SubmitInitialPackageIfReady() calls got %d want 1", submitter.calls)
	}
	if submitter.creatorUserID != creatorUserID {
		t.Fatalf("SubmitInitialPackageIfReady() creator got %s want %s", submitter.creatorUserID, creatorUserID)
	}
	if submitter.mainID != mainID {
		t.Fatalf("SubmitInitialPackageIfReady() main got %s want %s", submitter.mainID, mainID)
	}
}

func TestSubmitReadyPackageForReviewSkipsMissingInputs(t *testing.T) {
	t.Parallel()

	creatorUserID := uuid.MustParse("55555555-aaaa-aaaa-aaaa-555555555555")
	mainID := uuid.MustParse("66666666-aaaa-aaaa-aaaa-666666666666")
	submitter := &recordingReviewSubmitter{}

	tests := []struct {
		name      string
		processor *Processor
		claimed   claimedJob
	}{
		{
			name:      "nil processor",
			processor: nil,
			claimed: claimedJob{
				asset:           Asset{CreatorUserID: creatorUserID},
				canonicalMainID: mainID,
			},
		},
		{
			name:      "missing submitter",
			processor: &Processor{},
			claimed: claimedJob{
				asset:           Asset{CreatorUserID: creatorUserID},
				canonicalMainID: mainID,
			},
		},
		{
			name:      "missing creator user",
			processor: &Processor{submitter: submitter},
			claimed: claimedJob{
				asset:           Asset{},
				canonicalMainID: mainID,
			},
		},
		{
			name:      "missing canonical main",
			processor: &Processor{submitter: submitter},
			claimed: claimedJob{
				asset: Asset{CreatorUserID: creatorUserID},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.processor.submitReadyPackageForReview(context.Background(), tt.claimed); err != nil {
				t.Fatalf("submitReadyPackageForReview() error = %v, want nil", err)
			}
		})
	}

	if submitter.calls != 0 {
		t.Fatalf("SubmitInitialPackageIfReady() calls got %d want 0", submitter.calls)
	}
}

func TestProcessClaimedJobMarksReviewSubmitterErrorAsSoftFailure(t *testing.T) {
	t.Parallel()

	creatorUserID := uuid.MustParse("77777777-aaaa-aaaa-aaaa-777777777777")
	mediaAssetID := uuid.MustParse("88888888-aaaa-aaaa-aaaa-888888888888")
	jobID := uuid.MustParse("99999999-aaaa-aaaa-aaaa-999999999999")
	mainID := uuid.MustParse("aaaaaaaa-bbbb-bbbb-bbbb-aaaaaaaaaaaa")
	shortID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	shortAssetID := uuid.MustParse("cccccccc-bbbb-bbbb-bbbb-cccccccccccc")
	submitErr := errors.New("review submit unavailable")
	markCalls := 0

	queries := readyInitialReviewGateQueries(t, creatorUserID, mainID, mediaAssetID, shortID, shortAssetID, assetStateReady)
	queries.updateMediaAssetProcessingState = func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
		return sqlc.AppMediaAsset{}, nil
	}
	queries.markMediaProcessingJobSucceeded = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
		return sqlc.AppMediaProcessingJob{}, nil
	}
	queries.markMediaProcessingJobReviewSubmitFailed = func(_ context.Context, arg sqlc.MarkMediaProcessingJobReviewSubmitFailedParams) (sqlc.AppMediaProcessingJob, error) {
		markCalls++
		if arg.ID != pgUUID(jobID) {
			t.Fatalf("MarkMediaProcessingJobReviewSubmitFailed() id got %v want %v", arg.ID, pgUUID(jobID))
		}
		if got := postgres.OptionalTextFromPG(arg.LastErrorMessage); got == nil || !strings.Contains(*got, submitErr.Error()) {
			t.Fatalf("MarkMediaProcessingJobReviewSubmitFailed() message got %v want contains %q", got, submitErr.Error())
		}
		return sqlc.AppMediaProcessingJob{}, nil
	}
	queries.claimMediaProcessingJobByAsset = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.claimNextQueuedMediaProcessingJob = func(context.Context) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.getMediaProcessingJobByMediaAsset = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.requeueMediaProcessingJob = func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.markMediaProcessingJobFailed = func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.getMainByMediaAssetID = func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") }
	queries.getShortByMediaAssetID = func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") }

	processor, err := newProcessor(
		processorTxBeginnerStub{tx: &processorTxStub{}},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}
	processor.SetReviewSubmitter(&recordingReviewSubmitter{err: submitErr})

	err = processor.processClaimedJob(context.Background(), claimedJob{
		job: sqlc.AppMediaProcessingJob{
			ID:           pgUUID(jobID),
			MediaAssetID: pgUUID(mediaAssetID),
		},
		asset: Asset{
			ID:            mediaAssetID,
			CreatorUserID: creatorUserID,
			StorageBucket: "raw-bucket",
			StorageKey:    "raw/main.mp4",
		},
		role:            roleMain,
		mainID:          mainID,
		canonicalMainID: mainID,
	})
	if err != nil {
		t.Fatalf("processClaimedJob() error = %v, want nil", err)
	}
	if markCalls != 1 {
		t.Fatalf("MarkMediaProcessingJobReviewSubmitFailed() calls got %d want 1", markCalls)
	}
}

func TestProcessAssetRetriesReviewSubmitForSucceededJob(t *testing.T) {
	t.Parallel()

	creatorUserID := uuid.MustParse("11111111-cccc-cccc-cccc-111111111111")
	mediaAssetID := uuid.MustParse("22222222-cccc-cccc-cccc-222222222222")
	jobID := uuid.MustParse("33333333-cccc-cccc-cccc-333333333333")
	mainID := uuid.MustParse("44444444-cccc-cccc-cccc-444444444444")
	shortID := uuid.MustParse("55555555-cccc-cccc-cccc-555555555555")
	shortAssetID := uuid.MustParse("66666666-cccc-cccc-cccc-666666666666")
	submitErr := errors.New("review submit temporarily unavailable")
	submitter := &recordingReviewSubmitter{errs: []error{submitErr, nil}}
	markCalls := 0
	clearCalls := 0

	queries := readyInitialReviewGateQueries(t, creatorUserID, mainID, mediaAssetID, shortID, shortAssetID, assetStateReady)
	queries.updateMediaAssetProcessingState = func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
		return sqlc.AppMediaAsset{
			ID:              pgUUID(shortAssetID),
			CreatorUserID:   pgUUID(creatorUserID),
			ProcessingState: assetStateWorking,
			CreatedAt:       pgtype.Timestamptz{Valid: true},
			UpdatedAt:       pgtype.Timestamptz{Valid: true},
		}, nil
	}
	queries.markMediaProcessingJobSucceeded = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
		return sqlc.AppMediaProcessingJob{}, nil
	}
	queries.markMediaProcessingJobReviewSubmitFailed = func(_ context.Context, arg sqlc.MarkMediaProcessingJobReviewSubmitFailedParams) (sqlc.AppMediaProcessingJob, error) {
		markCalls++
		if arg.ID != pgUUID(jobID) {
			t.Fatalf("MarkMediaProcessingJobReviewSubmitFailed() id got %v want %v", arg.ID, pgUUID(jobID))
		}
		return sqlc.AppMediaProcessingJob{}, nil
	}
	queries.clearMediaProcessingJobReviewSubmitFailure = func(_ context.Context, id pgtype.UUID) error {
		clearCalls++
		if id != pgUUID(jobID) {
			t.Fatalf("ClearMediaProcessingJobReviewSubmitFailure() id got %v want %v", id, pgUUID(jobID))
		}
		return nil
	}
	queries.claimMediaProcessingJobByAsset = func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
		if id != pgUUID(mediaAssetID) {
			t.Fatalf("ClaimMediaProcessingJobByAssetID() id got %v want %v", id, pgUUID(mediaAssetID))
		}
		return sqlc.AppMediaProcessingJob{}, pgx.ErrNoRows
	}
	queries.getMediaProcessingJobByMediaAsset = func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
		if id != pgUUID(mediaAssetID) {
			t.Fatalf("GetMediaProcessingJobByMediaAssetID() id got %v want %v", id, pgUUID(mediaAssetID))
		}
		return sqlc.AppMediaProcessingJob{
			ID:            pgUUID(jobID),
			MediaAssetID:  pgUUID(mediaAssetID),
			AssetRole:     roleMain,
			Status:        jobStatusSucceeded,
			LastErrorCode: reviewSubmitFailedPGText(),
		}, nil
	}
	queries.getMainByMediaAssetID = func(_ context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
		if id != pgUUID(mediaAssetID) {
			t.Fatalf("GetMainByMediaAssetID() id got %v want %v", id, pgUUID(mediaAssetID))
		}
		return sqlc.AppMain{ID: pgUUID(mainID)}, nil
	}
	queries.claimNextQueuedMediaProcessingJob = func(context.Context) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.requeueMediaProcessingJob = func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.markMediaProcessingJobFailed = func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.getShortByMediaAssetID = func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") }

	processor, err := newProcessor(
		processorTxBeginnerStub{tx: &processorTxStub{}},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}
	processor.SetReviewSubmitter(submitter)

	err = processor.processClaimedJob(context.Background(), claimedJob{
		job: sqlc.AppMediaProcessingJob{
			ID:           pgUUID(jobID),
			MediaAssetID: pgUUID(mediaAssetID),
		},
		asset: Asset{
			ID:            mediaAssetID,
			CreatorUserID: creatorUserID,
			StorageBucket: "raw-bucket",
			StorageKey:    "raw/main.mp4",
		},
		role:            roleMain,
		mainID:          mainID,
		canonicalMainID: mainID,
	})
	if err != nil {
		t.Fatalf("processClaimedJob() error = %v, want nil", err)
	}
	if submitter.calls != 1 {
		t.Fatalf("SubmitInitialPackageIfReady() calls after first attempt got %d want 1", submitter.calls)
	}
	if markCalls != 1 {
		t.Fatalf("MarkMediaProcessingJobReviewSubmitFailed() calls got %d want 1", markCalls)
	}

	if err := processor.ProcessAsset(context.Background(), mediaAssetID); err != nil {
		t.Fatalf("ProcessAsset() retry error = %v, want nil", err)
	}
	if submitter.calls != 2 {
		t.Fatalf("SubmitInitialPackageIfReady() calls after retry got %d want 2", submitter.calls)
	}
	if submitter.creatorUserID != creatorUserID {
		t.Fatalf("SubmitInitialPackageIfReady() creator got %s want %s", submitter.creatorUserID, creatorUserID)
	}
	if submitter.mainID != mainID {
		t.Fatalf("SubmitInitialPackageIfReady() main got %s want %s", submitter.mainID, mainID)
	}
	if clearCalls != 1 {
		t.Fatalf("ClearMediaProcessingJobReviewSubmitFailure() calls got %d want 1", clearCalls)
	}
}

func TestProcessAssetRetriesReviewSubmitForSucceededShortJob(t *testing.T) {
	t.Parallel()

	creatorUserID := uuid.MustParse("77777777-dddd-dddd-dddd-777777777777")
	mainAssetID := uuid.MustParse("88888888-dddd-dddd-dddd-888888888888")
	shortAssetID := uuid.MustParse("99999999-dddd-dddd-dddd-999999999999")
	jobID := uuid.MustParse("aaaaaaaa-dddd-dddd-dddd-aaaaaaaaaaaa")
	mainID := uuid.MustParse("bbbbbbbb-dddd-dddd-dddd-bbbbbbbbbbbb")
	shortID := uuid.MustParse("cccccccc-dddd-dddd-dddd-cccccccccccc")
	submitter := &recordingReviewSubmitter{}
	clearCalls := 0

	queries := readyInitialReviewGateQueries(t, creatorUserID, mainID, mainAssetID, shortID, shortAssetID, assetStateReady)
	queries.claimMediaProcessingJobByAsset = func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
		if id != pgUUID(shortAssetID) {
			t.Fatalf("ClaimMediaProcessingJobByAssetID() id got %v want %v", id, pgUUID(shortAssetID))
		}
		return sqlc.AppMediaProcessingJob{}, pgx.ErrNoRows
	}
	queries.getMediaProcessingJobByMediaAsset = func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
		if id != pgUUID(shortAssetID) {
			t.Fatalf("GetMediaProcessingJobByMediaAssetID() id got %v want %v", id, pgUUID(shortAssetID))
		}
		return sqlc.AppMediaProcessingJob{
			ID:            pgUUID(jobID),
			CreatorUserID: pgUUID(creatorUserID),
			MediaAssetID:  pgUUID(shortAssetID),
			AssetRole:     roleShort,
			Status:        jobStatusSucceeded,
			LastErrorCode: reviewSubmitFailedPGText(),
		}, nil
	}
	queries.clearMediaProcessingJobReviewSubmitFailure = func(_ context.Context, id pgtype.UUID) error {
		clearCalls++
		if id != pgUUID(jobID) {
			t.Fatalf("ClearMediaProcessingJobReviewSubmitFailure() id got %v want %v", id, pgUUID(jobID))
		}
		return nil
	}
	queries.getShortByMediaAssetID = func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
		if id != pgUUID(shortAssetID) {
			t.Fatalf("GetShortByMediaAssetID() id got %v want %v", id, pgUUID(shortAssetID))
		}
		return sqlc.AppShort{
			ID:              pgUUID(shortID),
			CreatorUserID:   pgUUID(creatorUserID),
			CanonicalMainID: pgUUID(mainID),
			MediaAssetID:    pgUUID(shortAssetID),
		}, nil
	}
	queries.claimNextQueuedMediaProcessingJob = func(context.Context) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.getNextSucceededInitialReviewMediaProcessingJob = func(context.Context) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.updateMediaAssetProcessingState = func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
		panic("unexpected call")
	}
	queries.markMediaProcessingJobSucceeded = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.requeueMediaProcessingJob = func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.markMediaProcessingJobFailed = func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.getMainByMediaAssetID = func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") }

	processor, err := newProcessor(
		processorTxBeginnerStub{tx: &processorTxStub{}},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}
	processor.SetReviewSubmitter(submitter)

	if err := processor.ProcessAsset(context.Background(), shortAssetID); err != nil {
		t.Fatalf("ProcessAsset() error = %v, want nil", err)
	}
	if submitter.calls != 1 {
		t.Fatalf("SubmitInitialPackageIfReady() calls got %d want 1", submitter.calls)
	}
	if submitter.creatorUserID != creatorUserID {
		t.Fatalf("SubmitInitialPackageIfReady() creator got %s want %s", submitter.creatorUserID, creatorUserID)
	}
	if submitter.mainID != mainID {
		t.Fatalf("SubmitInitialPackageIfReady() main got %s want %s", submitter.mainID, mainID)
	}
	if clearCalls != 1 {
		t.Fatalf("ClearMediaProcessingJobReviewSubmitFailure() calls got %d want 1", clearCalls)
	}
}

func TestProcessNextQueuedRetriesSucceededShortReviewSubmitWithoutWake(t *testing.T) {
	t.Parallel()

	creatorUserID := uuid.MustParse("dddddddd-cccc-cccc-cccc-dddddddddddd")
	mainAssetID := uuid.MustParse("eeeeeeee-cccc-cccc-cccc-eeeeeeeeeeee")
	shortAssetID := uuid.MustParse("ffffffff-cccc-cccc-cccc-ffffffffffff")
	jobID := uuid.MustParse("11111111-dddd-dddd-dddd-111111111111")
	mainID := uuid.MustParse("22222222-dddd-dddd-dddd-222222222222")
	shortID := uuid.MustParse("33333333-dddd-dddd-dddd-333333333333")
	submitErr := errors.New("review submit temporarily unavailable")
	submitter := &recordingReviewSubmitter{errs: []error{submitErr, nil}}
	claimCalls := 0
	markCalls := 0
	clearCalls := 0

	queries := readyInitialReviewGateQueries(t, creatorUserID, mainID, mainAssetID, shortID, shortAssetID, assetStateReady)
	queries.claimNextQueuedMediaProcessingJob = func(context.Context) (sqlc.AppMediaProcessingJob, error) {
		claimCalls++
		if claimCalls == 1 {
			return sqlc.AppMediaProcessingJob{
				ID:            pgUUID(jobID),
				CreatorUserID: pgUUID(creatorUserID),
				MediaAssetID:  pgUUID(shortAssetID),
				AssetRole:     roleShort,
			}, nil
		}
		return sqlc.AppMediaProcessingJob{}, pgx.ErrNoRows
	}
	queries.getNextSucceededInitialReviewMediaProcessingJob = func(context.Context) (sqlc.AppMediaProcessingJob, error) {
		return sqlc.AppMediaProcessingJob{
			ID:            pgUUID(jobID),
			CreatorUserID: pgUUID(creatorUserID),
			MediaAssetID:  pgUUID(shortAssetID),
			AssetRole:     roleShort,
			Status:        jobStatusSucceeded,
			LastErrorCode: reviewSubmitFailedPGText(),
		}, nil
	}
	queries.getMediaProcessingJobByMediaAsset = func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
		if id != pgUUID(shortAssetID) {
			t.Fatalf("GetMediaProcessingJobByMediaAssetID() id got %v want %v", id, pgUUID(shortAssetID))
		}
		return sqlc.AppMediaProcessingJob{
			ID:            pgUUID(jobID),
			CreatorUserID: pgUUID(creatorUserID),
			MediaAssetID:  pgUUID(shortAssetID),
			AssetRole:     roleShort,
			Status:        jobStatusSucceeded,
			LastErrorCode: reviewSubmitFailedPGText(),
		}, nil
	}
	queries.getShortByMediaAssetID = func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
		if id != pgUUID(shortAssetID) {
			t.Fatalf("GetShortByMediaAssetID() id got %v want %v", id, pgUUID(shortAssetID))
		}
		return sqlc.AppShort{
			ID:              pgUUID(shortID),
			CreatorUserID:   pgUUID(creatorUserID),
			CanonicalMainID: pgUUID(mainID),
			MediaAssetID:    pgUUID(shortAssetID),
		}, nil
	}
	queries.updateMediaAssetProcessingState = func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
		return sqlc.AppMediaAsset{
			ID:              pgUUID(shortAssetID),
			CreatorUserID:   pgUUID(creatorUserID),
			ProcessingState: assetStateWorking,
			CreatedAt:       pgtype.Timestamptz{Valid: true},
			UpdatedAt:       pgtype.Timestamptz{Valid: true},
		}, nil
	}
	queries.markMediaProcessingJobSucceeded = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
		return sqlc.AppMediaProcessingJob{}, nil
	}
	queries.markMediaProcessingJobReviewSubmitFailed = func(_ context.Context, arg sqlc.MarkMediaProcessingJobReviewSubmitFailedParams) (sqlc.AppMediaProcessingJob, error) {
		markCalls++
		if arg.ID != pgUUID(jobID) {
			t.Fatalf("MarkMediaProcessingJobReviewSubmitFailed() id got %v want %v", arg.ID, pgUUID(jobID))
		}
		return sqlc.AppMediaProcessingJob{}, nil
	}
	queries.clearMediaProcessingJobReviewSubmitFailure = func(_ context.Context, id pgtype.UUID) error {
		clearCalls++
		if id != pgUUID(jobID) {
			t.Fatalf("ClearMediaProcessingJobReviewSubmitFailure() id got %v want %v", id, pgUUID(jobID))
		}
		return nil
	}
	queries.claimMediaProcessingJobByAsset = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.requeueMediaProcessingJob = func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.markMediaProcessingJobFailed = func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.getMainByMediaAssetID = func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") }

	processor, err := newProcessor(
		processorTxBeginnerStub{tx: &processorTxStub{}},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}
	processor.SetReviewSubmitter(submitter)

	processed, err := processor.ProcessNextQueued(context.Background())
	if err != nil {
		t.Fatalf("ProcessNextQueued() first error = %v, want nil", err)
	}
	if !processed {
		t.Fatal("ProcessNextQueued() processed = false, want true")
	}
	if submitter.calls != 1 {
		t.Fatalf("SubmitInitialPackageIfReady() calls after first attempt got %d want 1", submitter.calls)
	}
	if markCalls != 1 {
		t.Fatalf("MarkMediaProcessingJobReviewSubmitFailed() calls got %d want 1", markCalls)
	}

	processed, err = processor.ProcessNextQueued(context.Background())
	if err != nil {
		t.Fatalf("ProcessNextQueued() retry error = %v, want nil", err)
	}
	if !processed {
		t.Fatal("ProcessNextQueued() retry processed = false, want true")
	}
	if submitter.calls != 2 {
		t.Fatalf("SubmitInitialPackageIfReady() calls after retry got %d want 2", submitter.calls)
	}
	if submitter.creatorUserID != creatorUserID {
		t.Fatalf("SubmitInitialPackageIfReady() creator got %s want %s", submitter.creatorUserID, creatorUserID)
	}
	if submitter.mainID != mainID {
		t.Fatalf("SubmitInitialPackageIfReady() main got %s want %s", submitter.mainID, mainID)
	}
	if clearCalls != 1 {
		t.Fatalf("ClearMediaProcessingJobReviewSubmitFailure() calls got %d want 1", clearCalls)
	}
}

func TestProcessClaimedJobSkipsSubmitterUntilAllPackageAssetsReady(t *testing.T) {
	t.Parallel()

	creatorUserID := uuid.MustParse("77777777-cccc-cccc-cccc-777777777777")
	mediaAssetID := uuid.MustParse("88888888-cccc-cccc-cccc-888888888888")
	jobID := uuid.MustParse("99999999-cccc-cccc-cccc-999999999999")
	mainID := uuid.MustParse("aaaaaaaa-cccc-cccc-cccc-aaaaaaaaaaaa")
	shortID := uuid.MustParse("bbbbbbbb-cccc-cccc-cccc-bbbbbbbbbbbb")
	shortAssetID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	submitter := &recordingReviewSubmitter{}

	queries := readyInitialReviewGateQueries(t, creatorUserID, mainID, mediaAssetID, shortID, shortAssetID, assetStateWorking)
	queries.updateMediaAssetProcessingState = func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
		return sqlc.AppMediaAsset{}, nil
	}
	queries.markMediaProcessingJobSucceeded = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
		return sqlc.AppMediaProcessingJob{}, nil
	}
	queries.claimMediaProcessingJobByAsset = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.claimNextQueuedMediaProcessingJob = func(context.Context) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.getMediaProcessingJobByMediaAsset = func(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) { panic("unexpected call") }
	queries.requeueMediaProcessingJob = func(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.markMediaProcessingJobFailed = func(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
		panic("unexpected call")
	}
	queries.getMainByMediaAssetID = func(context.Context, pgtype.UUID) (sqlc.AppMain, error) { panic("unexpected call") }
	queries.getShortByMediaAssetID = func(context.Context, pgtype.UUID) (sqlc.AppShort, error) { panic("unexpected call") }

	processor, err := newProcessor(
		processorTxBeginnerStub{tx: &processorTxStub{}},
		queries,
		func(sqlc.DBTX) processorQueries { return queries },
		stubAssetMaterializer{},
	)
	if err != nil {
		t.Fatalf("newProcessor() error = %v, want nil", err)
	}
	processor.SetReviewSubmitter(submitter)

	err = processor.processClaimedJob(context.Background(), claimedJob{
		job: sqlc.AppMediaProcessingJob{
			ID:           pgUUID(jobID),
			MediaAssetID: pgUUID(mediaAssetID),
		},
		asset: Asset{
			ID:            mediaAssetID,
			CreatorUserID: creatorUserID,
			StorageBucket: "raw-bucket",
			StorageKey:    "raw/main.mp4",
		},
		role:            roleMain,
		mainID:          mainID,
		canonicalMainID: mainID,
	})
	if err != nil {
		t.Fatalf("processClaimedJob() error = %v, want nil", err)
	}
	if submitter.calls != 0 {
		t.Fatalf("SubmitInitialPackageIfReady() calls got %d want 0", submitter.calls)
	}
}
