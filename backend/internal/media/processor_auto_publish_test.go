package media

import (
	"context"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type autoPublishQueriesStub struct {
	getMainByID                 func(context.Context, pgtype.UUID) (sqlc.AppMain, error)
	getMediaAssetByID           func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error)
	listShortsByCanonicalMainID func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error)
	updateMainState             func(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error)
	publishShort                func(context.Context, pgtype.UUID) (sqlc.AppShort, error)
}

func (s autoPublishQueriesStub) GetMediaAssetByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
	return s.getMediaAssetByID(ctx, id)
}
func (s autoPublishQueriesStub) UpdateMediaAssetProcessingState(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
	panic("unexpected call")
}
func (s autoPublishQueriesStub) GetMediaProcessingJobByMediaAssetID(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
	panic("unexpected call")
}
func (s autoPublishQueriesStub) ClaimMediaProcessingJobByAssetID(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
	panic("unexpected call")
}
func (s autoPublishQueriesStub) ClaimNextQueuedMediaProcessingJob(context.Context) (sqlc.AppMediaProcessingJob, error) {
	panic("unexpected call")
}
func (s autoPublishQueriesStub) MarkMediaProcessingJobSucceeded(context.Context, pgtype.UUID) (sqlc.AppMediaProcessingJob, error) {
	panic("unexpected call")
}
func (s autoPublishQueriesStub) RequeueMediaProcessingJob(context.Context, sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error) {
	panic("unexpected call")
}
func (s autoPublishQueriesStub) MarkMediaProcessingJobFailed(context.Context, sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error) {
	panic("unexpected call")
}
func (s autoPublishQueriesStub) GetMainByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
	return s.getMainByID(ctx, id)
}
func (s autoPublishQueriesStub) GetMainByMediaAssetID(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
	panic("unexpected call")
}
func (s autoPublishQueriesStub) GetShortByMediaAssetID(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
	panic("unexpected call")
}
func (s autoPublishQueriesStub) ListShortsByCanonicalMainID(ctx context.Context, canonicalMainID pgtype.UUID) ([]sqlc.AppShort, error) {
	return s.listShortsByCanonicalMainID(ctx, canonicalMainID)
}
func (s autoPublishQueriesStub) UpdateMainState(ctx context.Context, arg sqlc.UpdateMainStateParams) (sqlc.AppMain, error) {
	return s.updateMainState(ctx, arg)
}
func (s autoPublishQueriesStub) PublishShort(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
	return s.publishShort(ctx, id)
}

func TestAutoPublishIfReady(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	mainID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	mainAssetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	shortID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	shortAssetID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	updateCalls := 0
	publishCalls := 0

	processor := &Processor{now: func() time.Time { return now }}
	err := processor.autoPublishIfReady(context.Background(), autoPublishQueriesStub{
		getMainByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
			if id != pgUUID(mainID) {
				t.Fatalf("GetMainByID() id got %v want %v", id, pgUUID(mainID))
			}
			return sqlc.AppMain{
				ID:                  pgUUID(mainID),
				MediaAssetID:        pgUUID(mainAssetID),
				State:               mainStateDraft,
				OwnershipConfirmed:  true,
				ConsentConfirmed:    true,
				ApprovedForUnlockAt: pgtype.Timestamptz{},
			}, nil
		},
		getMediaAssetByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			switch id {
			case pgUUID(mainAssetID):
				return sqlc.AppMediaAsset{ID: id, ProcessingState: assetStateReady}, nil
			case pgUUID(shortAssetID):
				return sqlc.AppMediaAsset{ID: id, ProcessingState: assetStateReady}, nil
			default:
				t.Fatalf("GetMediaAssetByID() unexpected id %v", id)
				return sqlc.AppMediaAsset{}, nil
			}
		},
		listShortsByCanonicalMainID: func(_ context.Context, id pgtype.UUID) ([]sqlc.AppShort, error) {
			if id != pgUUID(mainID) {
				t.Fatalf("ListShortsByCanonicalMainID() id got %v want %v", id, pgUUID(mainID))
			}
			return []sqlc.AppShort{{
				ID:              pgUUID(shortID),
				CanonicalMainID: pgUUID(mainID),
				MediaAssetID:    pgUUID(shortAssetID),
				State:           shortStateDraft,
			}}, nil
		},
		updateMainState: func(_ context.Context, arg sqlc.UpdateMainStateParams) (sqlc.AppMain, error) {
			updateCalls++
			if arg.State != mainStateApprovedForUnlock {
				t.Fatalf("UpdateMainState() state got %q want %q", arg.State, mainStateApprovedForUnlock)
			}
			if approvedAt := postgres.OptionalTimeFromPG(arg.ApprovedForUnlockAt); approvedAt == nil || !approvedAt.Equal(now) {
				t.Fatalf("UpdateMainState() approved at got %v want %v", approvedAt, now)
			}
			return sqlc.AppMain{}, nil
		},
		publishShort: func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
			publishCalls++
			if id != pgUUID(shortID) {
				t.Fatalf("PublishShort() id got %v want %v", id, pgUUID(shortID))
			}
			return sqlc.AppShort{}, nil
		},
	}, mainID)
	if err != nil {
		t.Fatalf("autoPublishIfReady() error = %v, want nil", err)
	}
	if updateCalls != 1 {
		t.Fatalf("autoPublishIfReady() updateCalls got %d want 1", updateCalls)
	}
	if publishCalls != 1 {
		t.Fatalf("autoPublishIfReady() publishCalls got %d want 1", publishCalls)
	}
}

func TestAutoPublishIfReadySkipsWhenAssetNotReady(t *testing.T) {
	t.Parallel()

	mainID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	mainAssetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	updateCalls := 0

	processor := &Processor{now: time.Now}
	err := processor.autoPublishIfReady(context.Background(), autoPublishQueriesStub{
		getMainByID: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
			return sqlc.AppMain{
				ID:           pgUUID(mainID),
				MediaAssetID: pgUUID(mainAssetID),
				State:        mainStateDraft,
			}, nil
		},
		getMediaAssetByID: func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error) {
			return sqlc.AppMediaAsset{ProcessingState: assetStateWorking}, nil
		},
		listShortsByCanonicalMainID: func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
			t.Fatal("ListShortsByCanonicalMainID() should not be called")
			return nil, nil
		},
		updateMainState: func(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error) {
			updateCalls++
			return sqlc.AppMain{}, nil
		},
		publishShort: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
			t.Fatal("PublishShort() should not be called")
			return sqlc.AppShort{}, nil
		},
	}, mainID)
	if err != nil {
		t.Fatalf("autoPublishIfReady() error = %v, want nil", err)
	}
	if updateCalls != 0 {
		t.Fatalf("autoPublishIfReady() updateCalls got %d want 0", updateCalls)
	}
}
