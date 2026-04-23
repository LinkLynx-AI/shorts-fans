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

	defaultMaxProcessingAttempts = int32(3)
	cleanupPersistTimeout        = 15 * time.Second
	materializationFailedCode    = "materialization_failed"
	materializationInterrupted   = "materialization_interrupted"
	reviewSubmitFailedCode       = "review_submit_failed"

	reviewMainStateDraft  = "draft"
	reviewShortStateDraft = "draft"
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
	GetInitialReviewReadyPackageByMainID(ctx context.Context, id pgtype.UUID) (sqlc.GetInitialReviewReadyPackageByMainIDRow, error)
	MarkMediaProcessingJobSucceeded(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaProcessingJob, error)
	MarkMediaProcessingJobReviewSubmitFailed(ctx context.Context, arg sqlc.MarkMediaProcessingJobReviewSubmitFailedParams) (sqlc.AppMediaProcessingJob, error)
	ClearMediaProcessingJobReviewSubmitFailure(ctx context.Context, id pgtype.UUID) error
	GetNextSucceededInitialReviewMediaProcessingJob(ctx context.Context) (sqlc.AppMediaProcessingJob, error)
	RequeueMediaProcessingJob(ctx context.Context, arg sqlc.RequeueMediaProcessingJobParams) (sqlc.AppMediaProcessingJob, error)
	MarkMediaProcessingJobFailed(ctx context.Context, arg sqlc.MarkMediaProcessingJobFailedParams) (sqlc.AppMediaProcessingJob, error)
	GetMainByMediaAssetID(ctx context.Context, id pgtype.UUID) (sqlc.AppMain, error)
	GetShortByMediaAssetID(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error)
}

type assetMaterializer interface {
	Materialize(ctx context.Context, req MaterializeRequest) (MaterializeResult, error)
}

// ReviewSubmitter は delivery-ready になった submission package を review intake へ投入します。
type ReviewSubmitter interface {
	SubmitInitialPackageIfReady(ctx context.Context, viewerUserID uuid.UUID, mainID uuid.UUID) error
}

type claimedJob struct {
	job             sqlc.AppMediaProcessingJob
	asset           Asset
	role            string
	mainID          uuid.UUID
	shortID         uuid.UUID
	canonicalMainID uuid.UUID
}

// Processor は media processing job の claim / materialize / retry を統括します。
type Processor struct {
	beginner     postgres.TxBeginner
	queries      processorQueries
	newQueries   func(sqlc.DBTX) processorQueries
	materializer assetMaterializer
	submitter    ReviewSubmitter
	now          func() time.Time
	maxAttempts  int32
}

// NewProcessor は pgxpool ベースの media processor を構築します。
func NewProcessor(pool *pgxpool.Pool, materializer assetMaterializer) (*Processor, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is required")
	}

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

// SetReviewSubmitter は media asset ready 後の review intake 自動投入先を設定します。
func (p *Processor) SetReviewSubmitter(submitter ReviewSubmitter) {
	if p == nil {
		return
	}

	p.submitter = submitter
}

// ProcessAsset は指定 media asset の queued job を 1 件 claim して処理します。
func (p *Processor) ProcessAsset(ctx context.Context, mediaAssetID uuid.UUID) error {
	claimed, err := p.claimByAssetID(ctx, mediaAssetID)
	if err != nil {
		if errors.Is(err, ErrNoQueuedProcessingJob) {
			return p.retryReviewSubmitForSucceededAsset(ctx, mediaAssetID)
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
			return p.retryNextSucceededInitialReviewSubmit(ctx)
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
		if err := p.markSucceeded(resultCtx, claimed, result); err != nil {
			return err
		}
		if err := p.submitReadyPackageForReview(resultCtx, claimed); err != nil {
			if markerErr := p.markReviewSubmitFailed(resultCtx, claimed, err); markerErr != nil {
				return fmt.Errorf(
					"submit ready package for review media_asset_id=%s main_id=%s: %w",
					claimed.asset.ID,
					claimed.canonicalMainID,
					errors.Join(err, markerErr),
				)
			}
			return nil
		}
		return nil
	}

	jobErr := materializationJobError(err)
	if err := p.handleJobError(resultCtx, claimed, jobErr); err != nil {
		return fmt.Errorf("persist materialization failure media_asset_id=%s: %w", claimed.asset.ID, err)
	}

	return nil
}

func (p *Processor) submitReadyPackageForReview(ctx context.Context, claimed claimedJob) error {
	if p == nil || p.submitter == nil || claimed.asset.CreatorUserID == uuid.Nil || claimed.canonicalMainID == uuid.Nil {
		return nil
	}

	ready, err := p.isInitialPackageReadyForReview(ctx, claimed.asset.CreatorUserID, claimed.canonicalMainID)
	if err != nil {
		return err
	}
	if !ready {
		return nil
	}

	return p.submitter.SubmitInitialPackageIfReady(ctx, claimed.asset.CreatorUserID, claimed.canonicalMainID)
}

func (p *Processor) retryReviewSubmitForSucceededAsset(ctx context.Context, mediaAssetID uuid.UUID) error {
	if p == nil || p.submitter == nil {
		return nil
	}

	claimed, ok, err := p.loadSucceededClaimForReview(ctx, mediaAssetID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if !hasReviewSubmitFailedMarker(claimed.job) {
		return nil
	}

	if err := p.submitReadyPackageForReview(ctx, claimed); err != nil {
		return fmt.Errorf("retry ready package review submit media_asset_id=%s main_id=%s: %w", mediaAssetID, claimed.canonicalMainID, err)
	}
	if err := p.clearReviewSubmitFailure(ctx, claimed); err != nil {
		return fmt.Errorf("clear review submit failure media_asset_id=%s main_id=%s: %w", mediaAssetID, claimed.canonicalMainID, err)
	}

	return nil
}

func (p *Processor) retryNextSucceededInitialReviewSubmit(ctx context.Context) (bool, error) {
	if p == nil || p.submitter == nil {
		return false, nil
	}

	job, err := p.queries.GetNextSucceededInitialReviewMediaProcessingJob(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("load succeeded initial review retry candidate: %w", err)
	}

	mediaAssetID, err := postgres.UUIDFromPG(job.MediaAssetID)
	if err != nil {
		return false, fmt.Errorf("parse review retry media asset id: %w", err)
	}
	claimed, ok, err := p.loadSucceededClaimForReview(ctx, mediaAssetID)
	if err != nil {
		return true, err
	}
	if !ok {
		return false, nil
	}
	if !hasReviewSubmitFailedMarker(claimed.job) {
		return false, nil
	}

	if err := p.submitReadyPackageForReview(ctx, claimed); err != nil {
		return true, fmt.Errorf("retry ready package review submit media_asset_id=%s main_id=%s: %w", mediaAssetID, claimed.canonicalMainID, err)
	}
	if err := p.clearReviewSubmitFailure(ctx, claimed); err != nil {
		return true, fmt.Errorf("clear review submit failure media_asset_id=%s main_id=%s: %w", mediaAssetID, claimed.canonicalMainID, err)
	}

	return true, nil
}

func (p *Processor) loadSucceededClaimForReview(ctx context.Context, mediaAssetID uuid.UUID) (claimedJob, bool, error) {
	var claimed claimedJob
	found := false
	err := postgres.RunInTx(ctx, p.beginner, func(tx pgx.Tx) error {
		q := p.newQueries(tx)

		job, err := q.GetMediaProcessingJobByMediaAssetID(ctx, postgres.UUIDToPG(mediaAssetID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrProcessingJobNotFound
			}
			return fmt.Errorf("load processing job media_asset_id=%s: %w", mediaAssetID, err)
		}
		if job.Status != jobStatusSucceeded {
			return nil
		}

		assetRow, err := q.GetMediaAssetByID(ctx, postgres.UUIDToPG(mediaAssetID))
		if err != nil {
			return fmt.Errorf("load succeeded media asset media_asset_id=%s: %w", mediaAssetID, err)
		}
		asset, err := mapAsset(assetRow)
		if err != nil {
			return fmt.Errorf("map succeeded media asset media_asset_id=%s: %w", mediaAssetID, err)
		}

		target, err := resolveClaimedTarget(ctx, q, job)
		if err != nil {
			return err
		}
		target.job = job
		target.asset = asset
		claimed = target
		found = true
		return nil
	})
	if err != nil {
		return claimedJob{}, false, err
	}

	return claimed, found, nil
}

func (p *Processor) isInitialPackageReadyForReview(ctx context.Context, creatorUserID uuid.UUID, mainID uuid.UUID) (bool, error) {
	row, err := p.queries.GetInitialReviewReadyPackageByMainID(ctx, postgres.UUIDToPG(mainID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	readyCreatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return false, fmt.Errorf("parse ready package creator id main_id=%s: %w", mainID, err)
	}

	return readyCreatorUserID == creatorUserID, nil
}

func hasReviewSubmitFailedMarker(job sqlc.AppMediaProcessingJob) bool {
	return job.LastErrorCode.Valid && job.LastErrorCode.String == reviewSubmitFailedCode
}

func (p *Processor) markReviewSubmitFailed(ctx context.Context, claimed claimedJob, cause error) error {
	if p == nil || p.queries == nil || cause == nil {
		return nil
	}

	message := cause.Error()
	if _, err := p.queries.MarkMediaProcessingJobReviewSubmitFailed(ctx, sqlc.MarkMediaProcessingJobReviewSubmitFailedParams{
		ID:               claimed.job.ID,
		LastErrorMessage: pgTextPtr(&message),
	}); err != nil {
		return fmt.Errorf("mark media processing job review submit failed id=%s: %w", claimed.job.ID, err)
	}

	return nil
}

func (p *Processor) clearReviewSubmitFailure(ctx context.Context, claimed claimedJob) error {
	if p == nil || p.queries == nil || !hasReviewSubmitFailedMarker(claimed.job) {
		return nil
	}

	return p.queries.ClearMediaProcessingJobReviewSubmitFailure(ctx, claimed.job.ID)
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
