package creatorupload

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
)

type queries interface {
	CreateMain(ctx context.Context, arg sqlc.CreateMainParams) (sqlc.AppMain, error)
	CreateMediaAsset(ctx context.Context, arg sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error)
	CreateMediaProcessingJob(ctx context.Context, arg sqlc.CreateMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error)
	CreateShort(ctx context.Context, arg sqlc.CreateShortParams) (sqlc.AppShort, error)
}

type storedEntry struct {
	FileName      string `json:"fileName"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	MimeType      string `json:"mimeType"`
	Role          string `json:"role"`
	StorageKey    string `json:"storageKey"`
	UploadEntryID string `json:"uploadEntryId"`
}

type storedPackage struct {
	CreatorUserID string        `json:"creatorUserId"`
	ConsumedAt    *time.Time    `json:"consumedAt,omitempty"`
	ExpiresAt     time.Time     `json:"expiresAt"`
	Main          storedEntry   `json:"main"`
	Shorts        []storedEntry `json:"shorts"`
}

type createDraftPackageInput struct {
	CreatorUserID uuid.UUID
	Main          storedEntry
	MainConsent   bool
	MainOwnership bool
	MainPriceJpy  int64
	RawBucketName string
	Shorts        []createDraftShortInput
}

type createDraftShortInput struct {
	Caption *string
	Entry   storedEntry
}

// Repository は creator upload completion の DB 永続化を扱います。
type Repository struct {
	beginner   postgres.TxBeginner
	newQueries func(db sqlc.DBTX) queries
}

// NewRepository は pgxpool ベースの creator upload repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		beginner: pool,
		newQueries: func(db sqlc.DBTX) queries {
			return sqlc.New(db)
		},
	}
}

// CreateDraftPackage は upload 済み main / shorts を同一 transaction で draft row に確定します。
func (r *Repository) CreateDraftPackage(ctx context.Context, input createDraftPackageInput) (CompletePackageResult, error) {
	if r == nil || r.beginner == nil || r.newQueries == nil {
		return CompletePackageResult{}, fmt.Errorf("creator upload repository is not initialized")
	}

	var result CompletePackageResult
	err := postgres.RunInTx(ctx, r.beginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)

		mainAssetRow, err := createMediaAsset(ctx, q, input.CreatorUserID, input.RawBucketName, input.Main)
		if err != nil {
			return err
		}

		mainRow, err := q.CreateMain(ctx, sqlc.CreateMainParams{
			CreatorUserID:       postgres.UUIDToPG(input.CreatorUserID),
			MediaAssetID:        mainAssetRow.ID,
			State:               stateDraft,
			ReviewReasonCode:    postgres.TextToPG(nil),
			PostReportState:     postgres.TextToPG(nil),
			PriceMinor:          postgres.Int64ToPG(&input.MainPriceJpy),
			CurrencyCode:        postgres.TextToPG(pointerTo(currencyJPY)),
			OwnershipConfirmed:  input.MainOwnership,
			ConsentConfirmed:    input.MainConsent,
			ApprovedForUnlockAt: postgres.TimeToPG(nil),
		})
		if err != nil {
			return fmt.Errorf("create draft main: %w", err)
		}
		if err := createMediaProcessingJob(ctx, q, input.CreatorUserID, mainAssetRow.ID, roleMain); err != nil {
			return fmt.Errorf("create main processing job: %w", err)
		}

		shorts := make([]CreatedShort, 0, len(input.Shorts))
		for _, shortInput := range input.Shorts {
			shortAssetRow, createErr := createMediaAsset(ctx, q, input.CreatorUserID, input.RawBucketName, shortInput.Entry)
			if createErr != nil {
				return createErr
			}

			shortRow, createErr := q.CreateShort(ctx, sqlc.CreateShortParams{
				CreatorUserID:        postgres.UUIDToPG(input.CreatorUserID),
				CanonicalMainID:      mainRow.ID,
				MediaAssetID:         shortAssetRow.ID,
				Caption:              postgres.TextToPG(shortInput.Caption),
				State:                stateDraft,
				ReviewReasonCode:     postgres.TextToPG(nil),
				PostReportState:      postgres.TextToPG(nil),
				ApprovedForPublishAt: postgres.TimeToPG(nil),
				PublishedAt:          postgres.TimeToPG(nil),
			})
			if createErr != nil {
				return fmt.Errorf("create draft short upload_entry_id=%s: %w", shortInput.Entry.UploadEntryID, createErr)
			}
			if err := createMediaProcessingJob(ctx, q, input.CreatorUserID, shortAssetRow.ID, roleShort); err != nil {
				return fmt.Errorf("create short processing job upload_entry_id=%s: %w", shortInput.Entry.UploadEntryID, err)
			}

			short, mapErr := mapCreatedShort(shortRow, shortAssetRow)
			if mapErr != nil {
				return fmt.Errorf("map draft short upload_entry_id=%s: %w", shortInput.Entry.UploadEntryID, mapErr)
			}
			shorts = append(shorts, short)
		}

		main, err := mapCreatedMain(mainRow, mainAssetRow)
		if err != nil {
			return fmt.Errorf("map draft main: %w", err)
		}

		result = CompletePackageResult{
			Main:   main,
			Shorts: shorts,
		}
		return nil
	})
	if err != nil {
		return CompletePackageResult{}, fmt.Errorf("create draft package: %w", err)
	}

	return result, nil
}

func createMediaProcessingJob(
	ctx context.Context,
	q queries,
	creatorUserID uuid.UUID,
	mediaAssetID pgtype.UUID,
	assetRole string,
) error {
	if _, err := q.CreateMediaProcessingJob(ctx, sqlc.CreateMediaProcessingJobParams{
		CreatorUserID:    postgres.UUIDToPG(creatorUserID),
		MediaAssetID:     mediaAssetID,
		AssetRole:        assetRole,
		Status:           processingJobStatusQueued,
		AttemptCount:     0,
		LastErrorCode:    postgres.TextToPG(nil),
		LastErrorMessage: postgres.TextToPG(nil),
		StartedAt:        postgres.TimeToPG(nil),
		CompletedAt:      postgres.TimeToPG(nil),
		FailedAt:         postgres.TimeToPG(nil),
	}); err != nil {
		return err
	}

	return nil
}

func createMediaAsset(ctx context.Context, q queries, creatorUserID uuid.UUID, rawBucketName string, entry storedEntry) (sqlc.AppMediaAsset, error) {
	row, err := q.CreateMediaAsset(ctx, sqlc.CreateMediaAssetParams{
		CreatorUserID:     postgres.UUIDToPG(creatorUserID),
		ProcessingState:   stateUploaded,
		StorageProvider:   storageProviderS3,
		StorageBucket:     rawBucketName,
		StorageKey:        entry.StorageKey,
		PlaybackUrl:       postgres.TextToPG(nil),
		MimeType:          entry.MimeType,
		DurationMs:        postgres.Int64ToPG(nil),
		ExternalUploadRef: postgres.TextToPG(&entry.UploadEntryID),
	})
	if err != nil {
		return sqlc.AppMediaAsset{}, fmt.Errorf("create media asset upload_entry_id=%s: %w", entry.UploadEntryID, err)
	}

	return row, nil
}

func mapCreatedMain(mainRow sqlc.AppMain, assetRow sqlc.AppMediaAsset) (CreatedMain, error) {
	mainID, err := postgres.UUIDFromPG(mainRow.ID)
	if err != nil {
		return CreatedMain{}, fmt.Errorf("parse main id: %w", err)
	}
	asset, err := mapCreatedMediaAsset(assetRow)
	if err != nil {
		return CreatedMain{}, err
	}

	return CreatedMain{
		ID:         mainID,
		MediaAsset: asset,
		State:      mainRow.State,
	}, nil
}

func mapCreatedShort(shortRow sqlc.AppShort, assetRow sqlc.AppMediaAsset) (CreatedShort, error) {
	shortID, err := postgres.UUIDFromPG(shortRow.ID)
	if err != nil {
		return CreatedShort{}, fmt.Errorf("parse short id: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(shortRow.CanonicalMainID)
	if err != nil {
		return CreatedShort{}, fmt.Errorf("parse canonical main id: %w", err)
	}
	asset, err := mapCreatedMediaAsset(assetRow)
	if err != nil {
		return CreatedShort{}, err
	}

	return CreatedShort{
		CanonicalMainID: canonicalMainID,
		ID:              shortID,
		MediaAsset:      asset,
		State:           shortRow.State,
	}, nil
}

func mapCreatedMediaAsset(assetRow sqlc.AppMediaAsset) (CreatedMediaAsset, error) {
	assetID, err := postgres.UUIDFromPG(assetRow.ID)
	if err != nil {
		return CreatedMediaAsset{}, fmt.Errorf("parse media asset id: %w", err)
	}

	return CreatedMediaAsset{
		ID:              assetID,
		MimeType:        assetRow.MimeType,
		ProcessingState: assetRow.ProcessingState,
	}, nil
}

func pointerTo(value string) *string {
	return &value
}
