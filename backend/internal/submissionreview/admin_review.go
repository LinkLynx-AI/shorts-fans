package submissionreview

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrAdminReviewCaseNotFound は admin submission review detail 対象が見つからないことを表します。
var ErrAdminReviewCaseNotFound = errors.New("admin submission review case was not found")

type adminReviewQueries interface {
	GetAdminSubmissionReviewCaseSummaryByIntakeID(ctx context.Context, id pgtype.UUID) (sqlc.GetAdminSubmissionReviewCaseSummaryByIntakeIDRow, error)
	GetAdminSubmissionReviewMainByIntakeID(ctx context.Context, id pgtype.UUID) (sqlc.GetAdminSubmissionReviewMainByIntakeIDRow, error)
	ListAdminSubmissionReviewQueue(ctx context.Context) ([]sqlc.ListAdminSubmissionReviewQueueRow, error)
	ListAdminSubmissionReviewShortsByIntakeID(ctx context.Context, submissionReviewIntakeID pgtype.UUID) ([]sqlc.ListAdminSubmissionReviewShortsByIntakeIDRow, error)
}

// AdminReviewServiceConfig は admin submission review service の設定です。
type AdminReviewServiceConfig struct {
	MediaAccessTTL time.Duration
}

// AdminReviewCreator は admin queue/detail に表示する creator summary です。
type AdminReviewCreator struct {
	AvatarURL   *string
	Bio         string
	DisplayName string
	Handle      string
	UserID      uuid.UUID
}

// ReviewProvenance は current object state に残っている review provenance です。
type ReviewProvenance struct {
	DecisionSource *string
	DecisionedAt   *time.Time
	ReasonCode     *string
	ReviewNote     *string
}

// IntakeDecisionLog は対象 intake に紐づく decision log です。
type IntakeDecisionLog struct {
	DecisionSource string
	DecisionedAt   time.Time
	ReasonCode     *string
	ReviewNote     *string
	TargetState    string
}

// AdminReviewQueueItem は pending intake queue の summary です。
type AdminReviewQueueItem struct {
	Creator              AdminReviewCreator
	IntakeID             uuid.UUID
	MainDecisionRequired bool
	PendingShortCount    int64
	ShortCount           int64
	SubmitKind           string
	SubmittedAt          time.Time
}

// AdminReviewIntake は detail に表示する intake summary です。
type AdminReviewIntake struct {
	CanonicalMainID    uuid.UUID
	ConsentConfirmed   bool
	CreatorUserID      uuid.UUID
	ID                 uuid.UUID
	MainMediaAssetID   uuid.UUID
	MainPriceJpy       int64
	OwnershipConfirmed bool
	PreviousIntakeID   *uuid.UUID
	Status             string
	SubmitKind         string
	SubmittedAt        time.Time
}

// AdminReviewMain は admin review detail の canonical main です。
type AdminReviewMain struct {
	CurrencyCode      string
	DecisionRequired  bool
	ID                uuid.UUID
	IntakeDecisionLog *IntakeDecisionLog
	Media             media.VideoDisplayAsset
	PriceJpy          int64
	Review            ReviewProvenance
	State             string
}

// AdminReviewShort は admin review detail の linked short です。
type AdminReviewShort struct {
	Caption           *string
	DecisionRequired  bool
	ID                uuid.UUID
	IntakeDecisionLog *IntakeDecisionLog
	Media             media.VideoDisplayAsset
	Review            ReviewProvenance
	State             string
}

// AdminReviewCase は admin review detail 全体です。
type AdminReviewCase struct {
	Creator AdminReviewCreator
	Intake  AdminReviewIntake
	Main    AdminReviewMain
	Shorts  []AdminReviewShort
}

// AdminReviewService は admin submission review queue/detail/decision を扱います。
type AdminReviewService struct {
	applyDecision   func(ctx context.Context, input ReviewDecisionInput) error
	decisionService *Service
	mediaAccessTTL  time.Duration
	queries         adminReviewQueries
	resolveMain     func(ctx context.Context, source media.MainDisplaySource, boundary media.AccessBoundary, ttl time.Duration) (media.VideoDisplayAsset, error)
	resolveShort    func(source media.ShortDisplaySource, boundary media.AccessBoundary) (media.VideoDisplayAsset, error)
}

// NewAdminReviewService は admin submission review service を構築します。
func NewAdminReviewService(
	cfg AdminReviewServiceConfig,
	pool *pgxpool.Pool,
	delivery *media.Delivery,
) (*AdminReviewService, error) {
	if pool == nil {
		return nil, fmt.Errorf("submission review admin pool is required")
	}
	if delivery == nil {
		return nil, fmt.Errorf("submission review admin delivery is required")
	}
	if cfg.MediaAccessTTL <= 0 {
		cfg.MediaAccessTTL = media.DefaultSignedURLTTL
	}

	decisionService := NewService(pool)

	return &AdminReviewService{
		applyDecision:   decisionService.ApplyDecision,
		decisionService: decisionService,
		mediaAccessTTL:  cfg.MediaAccessTTL,
		queries:         sqlc.New(pool),
		resolveMain:     delivery.ResolveMainDisplayAsset,
		resolveShort:    delivery.ResolveShortDisplayAsset,
	}, nil
}

// ListCases は pending submission review intake queue を返します。
func (s *AdminReviewService) ListCases(ctx context.Context) ([]AdminReviewQueueItem, error) {
	if s == nil || s.queries == nil {
		return nil, fmt.Errorf("submission review admin service is not initialized")
	}

	rows, err := s.queries.ListAdminSubmissionReviewQueue(ctx)
	if err != nil {
		return nil, fmt.Errorf("submission review admin queue load: %w", err)
	}

	items := make([]AdminReviewQueueItem, 0, len(rows))
	for _, row := range rows {
		item, mapErr := mapAdminReviewQueueItem(row)
		if mapErr != nil {
			return nil, mapErr
		}
		items = append(items, item)
	}

	return items, nil
}

// GetCase は intake ごとの admin review detail を返します。
func (s *AdminReviewService) GetCase(ctx context.Context, intakeID uuid.UUID) (AdminReviewCase, error) {
	if s == nil || s.queries == nil {
		return AdminReviewCase{}, fmt.Errorf("submission review admin service is not initialized")
	}
	if intakeID == uuid.Nil {
		return AdminReviewCase{}, ErrAdminReviewCaseNotFound
	}

	summary, err := s.queries.GetAdminSubmissionReviewCaseSummaryByIntakeID(ctx, postgres.UUIDToPG(intakeID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminReviewCase{}, ErrAdminReviewCaseNotFound
		}
		return AdminReviewCase{}, fmt.Errorf("submission review admin case summary load intake=%s: %w", intakeID, err)
	}

	mainRow, err := s.queries.GetAdminSubmissionReviewMainByIntakeID(ctx, postgres.UUIDToPG(intakeID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminReviewCase{}, ErrAdminReviewCaseNotFound
		}
		return AdminReviewCase{}, fmt.Errorf("submission review admin main load intake=%s: %w", intakeID, err)
	}

	shortRows, err := s.queries.ListAdminSubmissionReviewShortsByIntakeID(ctx, postgres.UUIDToPG(intakeID))
	if err != nil {
		return AdminReviewCase{}, fmt.Errorf("submission review admin shorts load intake=%s: %w", intakeID, err)
	}

	creatorUserID, err := postgres.UUIDFromPG(summary.CreatorUserID)
	if err != nil {
		return AdminReviewCase{}, fmt.Errorf("submission review admin creator id parse intake=%s: %w", intakeID, err)
	}
	intakeMainID, err := postgres.UUIDFromPG(summary.CanonicalMainID)
	if err != nil {
		return AdminReviewCase{}, fmt.Errorf("submission review admin intake main id parse intake=%s: %w", intakeID, err)
	}
	mainMediaAssetID, err := postgres.UUIDFromPG(summary.MainMediaAssetID)
	if err != nil {
		return AdminReviewCase{}, fmt.Errorf("submission review admin intake main media parse intake=%s: %w", intakeID, err)
	}
	submittedAt, err := postgres.RequiredTimeFromPG(summary.SubmittedAt)
	if err != nil {
		return AdminReviewCase{}, fmt.Errorf("submission review admin submitted at parse intake=%s: %w", intakeID, err)
	}

	intake := AdminReviewIntake{
		CanonicalMainID:    intakeMainID,
		ConsentConfirmed:   summary.ConsentConfirmed,
		CreatorUserID:      creatorUserID,
		ID:                 intakeID,
		MainMediaAssetID:   mainMediaAssetID,
		MainPriceJpy:       summary.MainPriceMinor,
		OwnershipConfirmed: summary.OwnershipConfirmed,
		PreviousIntakeID:   optionalUUID(summary.PreviousIntakeID),
		Status:             summary.Status,
		SubmitKind:         summary.SubmitKind,
		SubmittedAt:        submittedAt,
	}

	creator := AdminReviewCreator{
		AvatarURL:   postgres.OptionalTextFromPG(summary.AvatarUrl),
		Bio:         summary.CreatorBio,
		DisplayName: summary.DisplayName,
		Handle:      summary.Handle,
		UserID:      creatorUserID,
	}

	main, err := s.buildAdminReviewMain(ctx, intakeID, mainRow)
	if err != nil {
		return AdminReviewCase{}, err
	}

	shorts := make([]AdminReviewShort, 0, len(shortRows))
	for _, row := range shortRows {
		shortItem, buildErr := s.buildAdminReviewShort(intakeID, row)
		if buildErr != nil {
			return AdminReviewCase{}, buildErr
		}
		shorts = append(shorts, shortItem)
	}

	if intake.Status != intakeStatusPendingReview {
		main.DecisionRequired = false
		for index := range shorts {
			shorts[index].DecisionRequired = false
		}
	}

	return AdminReviewCase{
		Creator: creator,
		Intake:  intake,
		Main:    main,
		Shorts:  shorts,
	}, nil
}

// ApplyDecision は intake に decision を適用します。
func (s *AdminReviewService) ApplyDecision(ctx context.Context, input ReviewDecisionInput) error {
	if s == nil || s.applyDecision == nil {
		return fmt.Errorf("submission review admin service is not initialized")
	}

	if err := s.applyDecision(ctx, input); err != nil {
		return err
	}

	return nil
}

func mapAdminReviewQueueItem(row sqlc.ListAdminSubmissionReviewQueueRow) (AdminReviewQueueItem, error) {
	intakeID, err := postgres.UUIDFromPG(row.IntakeID)
	if err != nil {
		return AdminReviewQueueItem{}, fmt.Errorf("submission review admin queue intake id parse: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return AdminReviewQueueItem{}, fmt.Errorf("submission review admin queue creator id parse intake=%s: %w", intakeID, err)
	}
	submittedAt, err := postgres.RequiredTimeFromPG(row.SubmittedAt)
	if err != nil {
		return AdminReviewQueueItem{}, fmt.Errorf("submission review admin queue submitted at parse intake=%s: %w", intakeID, err)
	}

	return AdminReviewQueueItem{
		Creator: AdminReviewCreator{
			AvatarURL:   postgres.OptionalTextFromPG(row.AvatarUrl),
			Bio:         row.CreatorBio,
			DisplayName: row.DisplayName,
			Handle:      row.Handle,
			UserID:      creatorUserID,
		},
		IntakeID:             intakeID,
		MainDecisionRequired: row.MainDecisionRequired,
		PendingShortCount:    row.PendingShortCount,
		ShortCount:           row.ShortCount,
		SubmitKind:           row.SubmitKind,
		SubmittedAt:          submittedAt,
	}, nil
}

func (s *AdminReviewService) buildAdminReviewMain(
	ctx context.Context,
	intakeID uuid.UUID,
	row sqlc.GetAdminSubmissionReviewMainByIntakeIDRow,
) (AdminReviewMain, error) {
	mainID, err := postgres.UUIDFromPG(row.MainID)
	if err != nil {
		return AdminReviewMain{}, fmt.Errorf("submission review admin main id parse intake=%s: %w", intakeID, err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return AdminReviewMain{}, fmt.Errorf("submission review admin main media asset parse intake=%s: %w", intakeID, err)
	}
	durationMS, err := requiredDurationMS(row.DurationMs)
	if err != nil {
		return AdminReviewMain{}, fmt.Errorf("submission review admin main duration parse intake=%s main=%s: %w", intakeID, mainID, err)
	}

	mediaAsset, err := s.resolveMain(
		ctx,
		media.MainDisplaySource{
			AssetID:    mediaAssetID,
			MainID:     mainID,
			DurationMS: durationMS,
		},
		media.AccessBoundaryOwner,
		s.mediaAccessTTL,
	)
	if err != nil {
		return AdminReviewMain{}, fmt.Errorf("submission review admin main media resolve intake=%s main=%s: %w", intakeID, mainID, err)
	}

	return AdminReviewMain{
		CurrencyCode:      row.CurrencyCode,
		DecisionRequired:  row.State == mainStatePendingReview,
		ID:                mainID,
		IntakeDecisionLog: buildIntakeDecisionLog(row.IntakeTargetState, row.IntakeReasonCode, row.IntakeReviewNote, row.IntakeDecisionSource, row.IntakeDecisionedAt),
		Media:             mediaAsset,
		PriceJpy:          row.PriceMinor,
		Review: ReviewProvenance{
			DecisionSource: postgres.OptionalTextFromPG(row.ReviewDecisionSource),
			DecisionedAt:   postgres.OptionalTimeFromPG(row.ReviewDecisionedAt),
			ReasonCode:     postgres.OptionalTextFromPG(row.ReviewReasonCode),
			ReviewNote:     postgres.OptionalTextFromPG(row.ReviewNote),
		},
		State: row.State,
	}, nil
}

func (s *AdminReviewService) buildAdminReviewShort(
	intakeID uuid.UUID,
	row sqlc.ListAdminSubmissionReviewShortsByIntakeIDRow,
) (AdminReviewShort, error) {
	shortID, err := postgres.UUIDFromPG(row.ShortID)
	if err != nil {
		return AdminReviewShort{}, fmt.Errorf("submission review admin short id parse intake=%s: %w", intakeID, err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return AdminReviewShort{}, fmt.Errorf("submission review admin short media asset parse intake=%s short=%s: %w", intakeID, shortID, err)
	}
	durationMS, err := requiredDurationMS(row.DurationMs)
	if err != nil {
		return AdminReviewShort{}, fmt.Errorf("submission review admin short duration parse intake=%s short=%s: %w", intakeID, shortID, err)
	}

	mediaAsset, err := s.resolveShort(
		media.ShortDisplaySource{
			AssetID:    mediaAssetID,
			ShortID:    shortID,
			DurationMS: durationMS,
		},
		media.AccessBoundaryOwner,
	)
	if err != nil {
		return AdminReviewShort{}, fmt.Errorf("submission review admin short media resolve intake=%s short=%s: %w", intakeID, shortID, err)
	}

	return AdminReviewShort{
		Caption:           postgres.OptionalTextFromPG(row.IntakeCaption),
		DecisionRequired:  row.State == shortStatePendingReview,
		ID:                shortID,
		IntakeDecisionLog: buildIntakeDecisionLog(row.IntakeTargetState, row.IntakeReasonCode, row.IntakeReviewNote, row.IntakeDecisionSource, row.IntakeDecisionedAt),
		Media:             mediaAsset,
		Review: ReviewProvenance{
			DecisionSource: postgres.OptionalTextFromPG(row.ReviewDecisionSource),
			DecisionedAt:   postgres.OptionalTimeFromPG(row.ReviewDecisionedAt),
			ReasonCode:     postgres.OptionalTextFromPG(row.ReviewReasonCode),
			ReviewNote:     postgres.OptionalTextFromPG(row.ReviewNote),
		},
		State: row.State,
	}, nil
}

func buildIntakeDecisionLog(
	targetState pgtype.Text,
	reasonCode pgtype.Text,
	reviewNote pgtype.Text,
	decisionSource pgtype.Text,
	decisionedAt pgtype.Timestamptz,
) *IntakeDecisionLog {
	targetStateValue := postgres.OptionalTextFromPG(targetState)
	decisionSourceValue := postgres.OptionalTextFromPG(decisionSource)
	decisionedAtValue := postgres.OptionalTimeFromPG(decisionedAt)
	if targetStateValue == nil || decisionSourceValue == nil || decisionedAtValue == nil {
		return nil
	}

	return &IntakeDecisionLog{
		DecisionSource: *decisionSourceValue,
		DecisionedAt:   *decisionedAtValue,
		ReasonCode:     postgres.OptionalTextFromPG(reasonCode),
		ReviewNote:     postgres.OptionalTextFromPG(reviewNote),
		TargetState:    *targetStateValue,
	}
}

func optionalUUID(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}

	result := uuid.UUID(value.Bytes)
	return &result
}

func requiredDurationMS(value pgtype.Int8) (int64, error) {
	duration := postgres.OptionalInt64FromPG(value)
	if duration == nil || *duration <= 0 {
		return 0, fmt.Errorf("duration ms is required")
	}

	return *duration, nil
}
