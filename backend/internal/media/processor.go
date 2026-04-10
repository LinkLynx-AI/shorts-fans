package media

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/mediaconvert"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	assetStateUploaded = "uploaded"
	assetStateFailed   = "failed"
	assetStateReady    = "ready"
	assetStateWorking  = "processing"

	jobStatusQueued     = "queued"
	jobStatusProcessing = "processing"
	jobStatusFailed     = "failed"
	jobStatusSucceeded  = "succeeded"

	mainStateDraft               = "draft"
	mainStateApprovedForUnlock   = "approved_for_unlock"
	shortStateDraft              = "draft"
	shortStateApprovedForPublish = "approved_for_publish"

	defaultMaxProcessingAttempts = int32(3)
	cleanupPersistTimeout        = 15 * time.Second
	materializationFailedCode    = "materialization_failed"
	materializationInterrupted   = "materialization_interrupted"
)

var (
	// ErrProcessingJobNotFound は media asset に processing job が存在しないことを表します。
	ErrProcessingJobNotFound = errors.New("media processing job が見つかりません")
	// ErrNoQueuedProcessingJob は claim 可能な queued job がないことを表します。
	ErrNoQueuedProcessingJob = errors.New("claim 可能な queued media processing job がありません")
)

type processorQueries interface {
	GetMediaAssetByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error)
	UpdateMediaAssetProcessingState(ctx context.Context, arg sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error)
	GetMediaProcessingJobByMediaAssetID(ctx context.Context, mediaAssetID pgtype.UUID) (sqlc.AppMediaProcessingJob, error)
	ClaimMediaProcessingJobByAssetID(ctx context.Context, mediaAssetID pgtype.UUID) (sqlc.AppMediaProcessingJob, error)
	ClaimNextQueuedMediaProcessingJob(ctx context.Context) (sqlc.AppMediaProcessingJob, error)
	MarkMediaProcessingJobSucceeded(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error)
	RequeueMediaProcessingJob(ctx context.Context, arg sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error)
	MarkMediaProcessingJobFailed(ctx context.Context, arg sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error)
	GetMainByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMain, error)
	GetMainByMediaAssetID(ctx context.Context, id pgtype.UUID) (sqlc.AppMain, error)
	GetShortByMediaAssetID(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error)
	ListShortsByCanonicalMainID(ctx context.Context, canonicalMainID pgtype.UUID) ([]sqlc.AppShort, error)
	UpdateMainState(ctx context.Context, arg sqlc.UpdateMainStateParams) (sqlc.AppMain, error)
	PublishShort(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error)
}

type assetMaterializer interface {
	Materialize(ctx context.Context, req MaterializeRequest) (MaterializeResult, error)
}

type claimedJob struct {
	job             sqlc.AppMediaProcessingJob
	asset           Asset
	role            string
	mainID          uuid.UUID
	shortID         uuid.UUID
	canonicalMainID uuid.UUID
}

// Processor は media processing job の claim / materialize / retry / auto-publish を統括します。
type Processor struct {
	beginner     postgres.TxBeginner
	queries      processorQueries
	newQueries   func(sqlc.DBTX) processorQueries
	materializer assetMaterializer
	now          func() time.Time
	maxAttempts  int32
}

// NewProcessor は pgxpool ベースの media processor を構築します。
func NewProcessor(pool *pgxpool.Pool, materializer assetMaterializer) (*Processor, error) {
	return newProcessor(
		pool,
		sqlc.New(pool),
		func(db sqlc.DBTX) processorQueries { return sqlc.New(db) },
		materializer,
	)
}

func newProcessor(beginner postgres.TxBeginner, q processorQueries, newQueries func(sqlc.DBTX) processorQueries, materializer assetMaterializer) (*Processor, error) {
	switch {
	case beginner == nil:
		return nil, fmt.Errorf("tx beginner is required")
	case q == nil:
		return nil, fmt.Errorf("processor queries are required")
	case newQueries == nil:
		return nil, fmt.Errorf("processor transaction queries factory is required")
	case materializer == nil:
		return nil, fmt.Errorf("asset materializer is required")
	}

	return &Processor{
		beginner:     beginner,
		queries:      q,
		newQueries:   newQueries,
		materializer: materializer,
		now:          time.Now,
		maxAttempts:  defaultMaxProcessingAttempts,
	}, nil
}

// ProcessAsset は指定 media asset の queued job を 1 件 claim して処理します。
func (p *Processor) ProcessAsset(ctx context.Context, mediaAssetID uuid.UUID) error {
	claimed, err := p.claimByAssetID(ctx, mediaAssetID)
	if err != nil {
		if errors.Is(err, ErrNoQueuedProcessingJob) {
			return nil
		}
		return err
	}

	return p.processClaimedJob(ctx, claimed)
}

// ProcessNextQueued は queue message がなくても stranded queued job を進めます。
func (p *Processor) ProcessNextQueued(ctx context.Context) (bool, error) {
	claimed, err := p.claimNextQueued(ctx)
	if err != nil {
		if errors.Is(err, ErrNoQueuedProcessingJob) {
			return false, nil
		}
		return false, err
	}

	if err := p.processClaimedJob(ctx, claimed); err != nil {
		return true, err
	}

	return true, nil
}

// RequeueFailedAsset は internal recovery 用に failed job を queued へ戻します。
func (p *Processor) RequeueFailedAsset(ctx context.Context, mediaAssetID uuid.UUID) error {
	if p == nil {
		return fmt.Errorf("media processor is nil")
	}
	if mediaAssetID == uuid.Nil {
		return fmt.Errorf("media asset id is required")
	}

	return postgres.RunInTx(ctx, p.beginner, func(tx pgx.Tx) error {
		q := p.newQueries(tx)

		job, err := q.GetMediaProcessingJobByMediaAssetID(ctx, postgres.UUIDToPG(mediaAssetID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrProcessingJobNotFound
			}
			return fmt.Errorf("load processing job media_asset_id=%s: %w", mediaAssetID, err)
		}
		assetRow, err := q.GetMediaAssetByID(ctx, postgres.UUIDToPG(mediaAssetID))
		if err != nil {
			return fmt.Errorf("load media asset media_asset_id=%s: %w", mediaAssetID, err)
		}

		if _, err := q.UpdateMediaAssetProcessingState(ctx, sqlc.UpdateMediaAssetProcessingStateParams{
			ID:                assetRow.ID,
			ProcessingState:   assetStateUploaded,
			PlaybackUrl:       pgTextPtr(nil),
			DurationMs:        pgInt64Ptr(nil),
			ExternalUploadRef: assetRow.ExternalUploadRef,
		}); err != nil {
			return fmt.Errorf("reset media asset processing state media_asset_id=%s: %w", mediaAssetID, err)
		}
		if _, err := q.RequeueMediaProcessingJob(ctx, sqlc.RequeueMediaProcessingJobParams{
			ID:               job.ID,
			LastErrorCode:    pgTextPtr(nil),
			LastErrorMessage: pgTextPtr(nil),
		}); err != nil {
			return fmt.Errorf("requeue media processing job media_asset_id=%s: %w", mediaAssetID, err)
		}

		return nil
	})
}

func (p *Processor) processClaimedJob(ctx context.Context, claimed claimedJob) error {
	result, err := p.materializer.Materialize(ctx, MaterializeRequest{
		Role:         claimed.role,
		SourceBucket: claimed.asset.StorageBucket,
		SourceKey:    claimed.asset.StorageKey,
		MainID:       claimed.mainID,
		ShortID:      claimed.shortID,
	})
	resultCtx, cancel := detachProcessingContext(ctx)
	defer cancel()
	if err == nil {
		return p.markSucceeded(resultCtx, claimed, result)
	}

	jobErr := materializationJobError(err)
	if err := p.handleJobError(resultCtx, claimed, jobErr); err != nil {
		return fmt.Errorf("persist materialization failure media_asset_id=%s: %w", claimed.asset.ID, err)
	}

	return nil
}

func (p *Processor) claimByAssetID(ctx context.Context, mediaAssetID uuid.UUID) (claimedJob, error) {
	if p == nil {
		return claimedJob{}, fmt.Errorf("media processor is nil")
	}
	if mediaAssetID == uuid.Nil {
		return claimedJob{}, fmt.Errorf("media asset id is required")
	}

	var claimed claimedJob
	err := postgres.RunInTx(ctx, p.beginner, func(tx pgx.Tx) error {
		q := p.newQueries(tx)

		job, err := q.ClaimMediaProcessingJobByAssetID(ctx, postgres.UUIDToPG(mediaAssetID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return translateClaimMiss(ctx, q, mediaAssetID)
			}
			return fmt.Errorf("claim media processing job media_asset_id=%s: %w", mediaAssetID, err)
		}

		asset, target, err := p.loadClaimContext(ctx, q, job)
		if err != nil {
			return err
		}

		claimed = target
		claimed.asset = asset
		return nil
	})
	if err != nil {
		return claimedJob{}, err
	}

	return claimed, nil
}

func (p *Processor) claimNextQueued(ctx context.Context) (claimedJob, error) {
	if p == nil {
		return claimedJob{}, fmt.Errorf("media processor is nil")
	}

	var claimed claimedJob
	err := postgres.RunInTx(ctx, p.beginner, func(tx pgx.Tx) error {
		q := p.newQueries(tx)

		job, err := q.ClaimNextQueuedMediaProcessingJob(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNoQueuedProcessingJob
			}
			return fmt.Errorf("claim next queued media processing job: %w", err)
		}

		asset, target, err := p.loadClaimContext(ctx, q, job)
		if err != nil {
			return err
		}

		claimed = target
		claimed.asset = asset
		return nil
	})
	if err != nil {
		return claimedJob{}, err
	}

	return claimed, nil
}

func (p *Processor) loadClaimContext(ctx context.Context, q processorQueries, job sqlc.AppMediaProcessingJob) (Asset, claimedJob, error) {
	mediaAssetID, err := postgres.UUIDFromPG(job.MediaAssetID)
	if err != nil {
		return Asset{}, claimedJob{}, fmt.Errorf("parse job media asset id: %w", err)
	}

	assetRow, err := q.GetMediaAssetByID(ctx, job.MediaAssetID)
	if err != nil {
		return Asset{}, claimedJob{}, fmt.Errorf("load media asset media_asset_id=%s: %w", mediaAssetID, err)
	}
	updatedAssetRow, err := q.UpdateMediaAssetProcessingState(ctx, sqlc.UpdateMediaAssetProcessingStateParams{
		ID:                assetRow.ID,
		ProcessingState:   assetStateWorking,
		PlaybackUrl:       assetRow.PlaybackUrl,
		DurationMs:        assetRow.DurationMs,
		ExternalUploadRef: assetRow.ExternalUploadRef,
	})
	if err != nil {
		return Asset{}, claimedJob{}, fmt.Errorf("mark media asset processing media_asset_id=%s: %w", mediaAssetID, err)
	}
	updatedAsset, err := mapAsset(updatedAssetRow)
	if err != nil {
		return Asset{}, claimedJob{}, fmt.Errorf("map processing media asset media_asset_id=%s: %w", mediaAssetID, err)
	}

	target, err := resolveClaimedTarget(ctx, q, job)
	if err != nil {
		return Asset{}, claimedJob{}, err
	}
	target.job = job

	return updatedAsset, target, nil
}

func resolveClaimedTarget(ctx context.Context, q processorQueries, job sqlc.AppMediaProcessingJob) (claimedJob, error) {
	mediaAssetID, err := postgres.UUIDFromPG(job.MediaAssetID)
	if err != nil {
		return claimedJob{}, fmt.Errorf("parse processing job media asset id: %w", err)
	}

	switch job.AssetRole {
	case roleMain:
		mainRow, err := q.GetMainByMediaAssetID(ctx, job.MediaAssetID)
		if err != nil {
			return claimedJob{}, fmt.Errorf("load main by media asset id=%s: %w", mediaAssetID, err)
		}
		mainID, err := postgres.UUIDFromPG(mainRow.ID)
		if err != nil {
			return claimedJob{}, fmt.Errorf("parse main id for media asset id=%s: %w", mediaAssetID, err)
		}
		return claimedJob{
			role:            roleMain,
			mainID:          mainID,
			canonicalMainID: mainID,
		}, nil
	case roleShort:
		shortRow, err := q.GetShortByMediaAssetID(ctx, job.MediaAssetID)
		if err != nil {
			return claimedJob{}, fmt.Errorf("load short by media asset id=%s: %w", mediaAssetID, err)
		}
		shortID, err := postgres.UUIDFromPG(shortRow.ID)
		if err != nil {
			return claimedJob{}, fmt.Errorf("parse short id for media asset id=%s: %w", mediaAssetID, err)
		}
		mainID, err := postgres.UUIDFromPG(shortRow.CanonicalMainID)
		if err != nil {
			return claimedJob{}, fmt.Errorf("parse canonical main id for media asset id=%s: %w", mediaAssetID, err)
		}
		return claimedJob{
			role:            roleShort,
			shortID:         shortID,
			canonicalMainID: mainID,
		}, nil
	default:
		return claimedJob{}, fmt.Errorf("unsupported media processing job asset role: %s", job.AssetRole)
	}
}

func (p *Processor) markSucceeded(ctx context.Context, claimed claimedJob, result MaterializeResult) error {
	return postgres.RunInTx(ctx, p.beginner, func(tx pgx.Tx) error {
		q := p.newQueries(tx)

		assetRow, err := q.GetMediaAssetByID(ctx, postgres.UUIDToPG(claimed.asset.ID))
		if err != nil {
			return fmt.Errorf("reload media asset id=%s: %w", claimed.asset.ID, err)
		}
		if _, err := q.UpdateMediaAssetProcessingState(ctx, sqlc.UpdateMediaAssetProcessingStateParams{
			ID:                assetRow.ID,
			ProcessingState:   assetStateReady,
			PlaybackUrl:       pgTextPtr(&result.PlaybackURL),
			DurationMs:        pgInt64Ptr(&result.DurationMS),
			ExternalUploadRef: assetRow.ExternalUploadRef,
		}); err != nil {
			return fmt.Errorf("mark media asset ready id=%s: %w", claimed.asset.ID, err)
		}
		if _, err := q.MarkMediaProcessingJobSucceeded(ctx, claimed.job.ID); err != nil {
			return fmt.Errorf("mark media processing job succeeded id=%s: %w", claimed.asset.ID, err)
		}

		if err := p.autoPublishIfReady(ctx, q, claimed.canonicalMainID); err != nil {
			return err
		}

		return nil
	})
}

func (p *Processor) handleJobError(ctx context.Context, claimed claimedJob, jobErr *mediaconvert.JobError) error {
	return postgres.RunInTx(ctx, p.beginner, func(tx pgx.Tx) error {
		q := p.newQueries(tx)

		job, err := q.GetMediaProcessingJobByMediaAssetID(ctx, postgres.UUIDToPG(claimed.asset.ID))
		if err != nil {
			return fmt.Errorf("reload processing job media_asset_id=%s: %w", claimed.asset.ID, err)
		}
		assetRow, err := q.GetMediaAssetByID(ctx, postgres.UUIDToPG(claimed.asset.ID))
		if err != nil {
			return fmt.Errorf("reload media asset media_asset_id=%s: %w", claimed.asset.ID, err)
		}

		if !jobErr.Retryable || job.AttemptCount >= p.maxAttempts {
			if _, err := q.UpdateMediaAssetProcessingState(ctx, sqlc.UpdateMediaAssetProcessingStateParams{
				ID:                assetRow.ID,
				ProcessingState:   assetStateFailed,
				PlaybackUrl:       pgTextPtr(nil),
				DurationMs:        pgInt64Ptr(nil),
				ExternalUploadRef: assetRow.ExternalUploadRef,
			}); err != nil {
				return fmt.Errorf("mark media asset failed media_asset_id=%s: %w", claimed.asset.ID, err)
			}
			if _, err := q.MarkMediaProcessingJobFailed(ctx, sqlc.MarkMediaProcessingJobFailedParams{
				ID:               job.ID,
				LastErrorCode:    pgTextPtr(&jobErr.Code),
				LastErrorMessage: pgTextPtr(&jobErr.Message),
			}); err != nil {
				return fmt.Errorf("mark media processing job failed media_asset_id=%s: %w", claimed.asset.ID, err)
			}

			return nil
		}

		if _, err := q.RequeueMediaProcessingJob(ctx, sqlc.RequeueMediaProcessingJobParams{
			ID:               job.ID,
			LastErrorCode:    pgTextPtr(&jobErr.Code),
			LastErrorMessage: pgTextPtr(&jobErr.Message),
		}); err != nil {
			return fmt.Errorf("requeue media processing job media_asset_id=%s: %w", claimed.asset.ID, err)
		}

		return nil
	})
}

func (p *Processor) autoPublishIfReady(ctx context.Context, q processorQueries, mainID uuid.UUID) error {
	if mainID == uuid.Nil {
		return nil
	}

	mainRow, err := q.GetMainByID(ctx, postgres.UUIDToPG(mainID))
	if err != nil {
		return fmt.Errorf("load canonical main id=%s: %w", mainID, err)
	}
	if mainRow.State != mainStateDraft && mainRow.State != mainStateApprovedForUnlock {
		return nil
	}

	mainAssetRow, err := q.GetMediaAssetByID(ctx, mainRow.MediaAssetID)
	if err != nil {
		return fmt.Errorf("load canonical main asset id=%s: %w", mainID, err)
	}
	if mainAssetRow.ProcessingState != assetStateReady {
		return nil
	}

	shortRows, err := q.ListShortsByCanonicalMainID(ctx, postgres.UUIDToPG(mainID))
	if err != nil {
		return fmt.Errorf("list canonical shorts main_id=%s: %w", mainID, err)
	}
	if len(shortRows) == 0 {
		return nil
	}

	for _, shortRow := range shortRows {
		if shortRow.State != shortStateDraft && shortRow.State != shortStateApprovedForPublish {
			return nil
		}

		assetRow, err := q.GetMediaAssetByID(ctx, shortRow.MediaAssetID)
		if err != nil {
			return fmt.Errorf("load short asset main_id=%s: %w", mainID, err)
		}
		if assetRow.ProcessingState != assetStateReady {
			return nil
		}
	}

	if mainRow.State != mainStateApprovedForUnlock {
		approvedAt := postgres.OptionalTimeFromPG(mainRow.ApprovedForUnlockAt)
		if approvedAt == nil {
			now := p.now().UTC()
			approvedAt = &now
		}
		if _, err := q.UpdateMainState(ctx, sqlc.UpdateMainStateParams{
			ID:                  mainRow.ID,
			State:               mainStateApprovedForUnlock,
			ReviewReasonCode:    mainRow.ReviewReasonCode,
			PostReportState:     mainRow.PostReportState,
			PriceMinor:          mainRow.PriceMinor,
			CurrencyCode:        mainRow.CurrencyCode,
			OwnershipConfirmed:  mainRow.OwnershipConfirmed,
			ConsentConfirmed:    mainRow.ConsentConfirmed,
			ApprovedForUnlockAt: postgres.TimeToPG(approvedAt),
		}); err != nil {
			return fmt.Errorf("approve main for unlock id=%s: %w", mainID, err)
		}
	}

	for _, shortRow := range shortRows {
		if _, err := q.PublishShort(ctx, shortRow.ID); err != nil {
			shortID, parseErr := postgres.UUIDFromPG(shortRow.ID)
			if parseErr != nil {
				return fmt.Errorf("parse short id while publishing main=%s: %w", mainID, parseErr)
			}
			return fmt.Errorf("approve short for publish id=%s: %w", shortID, err)
		}
	}

	return nil
}

func translateClaimMiss(ctx context.Context, q processorQueries, mediaAssetID uuid.UUID) error {
	job, err := q.GetMediaProcessingJobByMediaAssetID(ctx, postgres.UUIDToPG(mediaAssetID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProcessingJobNotFound
		}
		return fmt.Errorf("load media processing job media_asset_id=%s: %w", mediaAssetID, err)
	}
	if job.Status == jobStatusQueued || job.Status == jobStatusProcessing || job.Status == jobStatusSucceeded || job.Status == jobStatusFailed {
		return ErrNoQueuedProcessingJob
	}

	return fmt.Errorf("unsupported media processing job status for media_asset_id=%s: %s", mediaAssetID, job.Status)
}

func detachProcessingContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), cleanupPersistTimeout)
	}

	base := context.WithoutCancel(ctx)
	if deadline, ok := ctx.Deadline(); ok {
		return context.WithDeadline(base, deadline)
	}

	return context.WithTimeout(base, cleanupPersistTimeout)
}

func materializationJobError(err error) *mediaconvert.JobError {
	if err == nil {
		return nil
	}

	var jobErr *mediaconvert.JobError
	if errors.As(err, &jobErr) {
		return jobErr
	}

	code := materializationFailedCode
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		code = materializationInterrupted
	}

	return &mediaconvert.JobError{
		Code:      code,
		Message:   err.Error(),
		Retryable: true,
	}
}

func pgTextPtr(value *string) pgtype.Text {
	return postgres.TextToPG(value)
}

func pgInt64Ptr(value *int64) pgtype.Int8 {
	return postgres.Int64ToPG(value)
}
