package media

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type repositoryStubQueries struct {
	createAsset func(context.Context, sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error)
	getAsset    func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error)
	listAssets  func(context.Context, pgtype.UUID) ([]sqlc.AppMediaAsset, error)
	updateAsset func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error)
}

func (s repositoryStubQueries) CreateMediaAsset(ctx context.Context, arg sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error) {
	return s.createAsset(ctx, arg)
}

func (s repositoryStubQueries) GetMediaAssetByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
	return s.getAsset(ctx, id)
}

func (s repositoryStubQueries) ListMediaAssetsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMediaAsset, error) {
	return s.listAssets(ctx, creatorUserID)
}

func (s repositoryStubQueries) UpdateMediaAssetProcessingState(ctx context.Context, arg sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
	return s.updateAsset(ctx, arg)
}

func TestRepositorySuccessPaths(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	assetID := uuid.New()
	creatorID := uuid.New()
	playbackURL := stringPtr("https://cdn.example.com/asset.m3u8")
	durationMS := int64Ptr(42000)
	externalRef := stringPtr("upload-1")
	row := testAssetRow(assetID, creatorID, now, playbackURL, durationMS, externalRef)

	var createArg sqlc.CreateMediaAssetParams
	var updateArg sqlc.UpdateMediaAssetProcessingStateParams
	repo := newRepository(repositoryStubQueries{
		createAsset: func(_ context.Context, arg sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error) {
			createArg = arg
			return row, nil
		},
		getAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			if id != pgUUID(assetID) {
				t.Fatalf("GetMediaAssetByID() id got %v want %v", id, pgUUID(assetID))
			}
			return row, nil
		},
		listAssets: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMediaAsset, error) {
			if creatorUserID != pgUUID(creatorID) {
				t.Fatalf("ListMediaAssetsByCreatorUserID() creator got %v want %v", creatorUserID, pgUUID(creatorID))
			}
			return []sqlc.AppMediaAsset{row}, nil
		},
		updateAsset: func(_ context.Context, arg sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
			updateArg = arg
			return row, nil
		},
	})

	input := CreateAssetInput{
		CreatorUserID:     creatorID,
		ProcessingState:   "ready",
		StorageProvider:   "s3",
		StorageBucket:     "bucket",
		StorageKey:        "key",
		PlaybackURL:       playbackURL,
		MimeType:          "video/mp4",
		DurationMS:        durationMS,
		ExternalUploadRef: externalRef,
	}

	created, err := repo.CreateAsset(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateAsset() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(created, wantAsset(assetID, creatorID, now, playbackURL, durationMS, externalRef)) {
		t.Fatalf("CreateAsset() got %#v want %#v", created, wantAsset(assetID, creatorID, now, playbackURL, durationMS, externalRef))
	}
	if createArg.CreatorUserID != pgUUID(creatorID) {
		t.Fatalf("CreateAsset() creator arg got %v want %v", createArg.CreatorUserID, pgUUID(creatorID))
	}

	got, err := repo.GetAsset(context.Background(), assetID)
	if err != nil {
		t.Fatalf("GetAsset() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, created) {
		t.Fatalf("GetAsset() got %#v want %#v", got, created)
	}

	listed, err := repo.ListAssetsByCreator(context.Background(), creatorID)
	if err != nil {
		t.Fatalf("ListAssetsByCreator() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(listed, []Asset{created}) {
		t.Fatalf("ListAssetsByCreator() got %#v want %#v", listed, []Asset{created})
	}

	updated, err := repo.UpdateAssetProcessingState(context.Background(), UpdateAssetProcessingStateInput{
		ID:                assetID,
		ProcessingState:   "ready",
		PlaybackURL:       playbackURL,
		DurationMS:        durationMS,
		ExternalUploadRef: externalRef,
	})
	if err != nil {
		t.Fatalf("UpdateAssetProcessingState() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(updated, created) {
		t.Fatalf("UpdateAssetProcessingState() got %#v want %#v", updated, created)
	}
	if updateArg.ID != pgUUID(assetID) {
		t.Fatalf("UpdateAssetProcessingState() id arg got %v want %v", updateArg.ID, pgUUID(assetID))
	}
}

func TestRepositoryErrorPaths(t *testing.T) {
	t.Parallel()

	assetID := uuid.New()
	creatorID := uuid.New()
	genericErr := errors.New("query failed")
	repo := newRepository(repositoryStubQueries{
		createAsset: func(context.Context, sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error) {
			return sqlc.AppMediaAsset{}, genericErr
		},
		getAsset: func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error) {
			return sqlc.AppMediaAsset{}, genericErr
		},
		listAssets: func(context.Context, pgtype.UUID) ([]sqlc.AppMediaAsset, error) {
			return nil, genericErr
		},
		updateAsset: func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
			return sqlc.AppMediaAsset{}, pgx.ErrNoRows
		},
	})

	if _, err := repo.CreateAsset(context.Background(), CreateAssetInput{}); !errors.Is(err, genericErr) {
		t.Fatalf("CreateAsset() error got %v want wrapped %v", err, genericErr)
	}
	if _, err := repo.GetAsset(context.Background(), assetID); !errors.Is(err, genericErr) {
		t.Fatalf("GetAsset() error got %v want wrapped %v", err, genericErr)
	}
	if _, err := repo.ListAssetsByCreator(context.Background(), creatorID); !errors.Is(err, genericErr) {
		t.Fatalf("ListAssetsByCreator() error got %v want wrapped %v", err, genericErr)
	}
	if _, err := repo.UpdateAssetProcessingState(context.Background(), UpdateAssetProcessingStateInput{ID: assetID}); !errors.Is(err, ErrAssetNotFound) {
		t.Fatalf("UpdateAssetProcessingState() error got %v want %v", err, ErrAssetNotFound)
	}
}

func TestRepositoryConversionErrors(t *testing.T) {
	t.Parallel()

	assetID := uuid.New()
	creatorID := uuid.New()
	invalidRow := testAssetRow(assetID, creatorID, time.Unix(1710000000, 0).UTC(), nil, nil, nil)
	invalidRow.ID = pgtype.UUID{}

	repo := newRepository(repositoryStubQueries{
		createAsset: func(context.Context, sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error) {
			return invalidRow, nil
		},
		getAsset: func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error) {
			return invalidRow, nil
		},
		listAssets: func(context.Context, pgtype.UUID) ([]sqlc.AppMediaAsset, error) {
			return []sqlc.AppMediaAsset{invalidRow}, nil
		},
		updateAsset: func(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
			return invalidRow, nil
		},
	})

	if _, err := repo.CreateAsset(context.Background(), CreateAssetInput{}); err == nil {
		t.Fatal("CreateAsset() error = nil, want conversion error")
	}
	if _, err := repo.GetAsset(context.Background(), assetID); err == nil {
		t.Fatal("GetAsset() error = nil, want conversion error")
	}
	if _, err := repo.ListAssetsByCreator(context.Background(), creatorID); err == nil {
		t.Fatal("ListAssetsByCreator() error = nil, want conversion error")
	}
	if _, err := repo.UpdateAssetProcessingState(context.Background(), UpdateAssetProcessingStateInput{ID: assetID}); err == nil {
		t.Fatal("UpdateAssetProcessingState() error = nil, want conversion error")
	}
}

func testAssetRow(id uuid.UUID, creatorID uuid.UUID, now time.Time, playbackURL *string, durationMS *int64, externalRef *string) sqlc.AppMediaAsset {
	return sqlc.AppMediaAsset{
		ID:                pgUUID(id),
		CreatorUserID:     pgUUID(creatorID),
		ProcessingState:   "ready",
		StorageProvider:   "s3",
		StorageBucket:     "bucket",
		StorageKey:        "key",
		PlaybackUrl:       pgText(playbackURL),
		MimeType:          "video/mp4",
		DurationMs:        pgInt64(durationMS),
		ExternalUploadRef: pgText(externalRef),
		CreatedAt:         pgTime(now),
		UpdatedAt:         pgTime(now.Add(time.Minute)),
	}
}

func wantAsset(id uuid.UUID, creatorID uuid.UUID, now time.Time, playbackURL *string, durationMS *int64, externalRef *string) Asset {
	return Asset{
		ID:                id,
		CreatorUserID:     creatorID,
		ProcessingState:   "ready",
		StorageProvider:   "s3",
		StorageBucket:     "bucket",
		StorageKey:        "key",
		PlaybackURL:       playbackURL,
		MimeType:          "video/mp4",
		DurationMS:        durationMS,
		ExternalUploadRef: externalRef,
		CreatedAt:         now,
		UpdatedAt:         now.Add(time.Minute),
	}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func pgTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func pgText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func pgInt64(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}

func stringPtr(value string) *string {
	return &value
}

func int64Ptr(value int64) *int64 {
	return &value
}
