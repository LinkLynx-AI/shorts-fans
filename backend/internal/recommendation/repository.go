package recommendation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrEventKindInvalid は recommendation event kind が未対応なことを表します。
	ErrEventKindInvalid = errors.New("recommendation event kind が不正です")
	// ErrViewerUserIDRequired は recommendation event に viewer user id が必要なことを表します。
	ErrViewerUserIDRequired = errors.New("recommendation event の viewer user id が必要です")
	// ErrIdempotencyKeyRequired は recommendation event に idempotency key が必要なことを表します。
	ErrIdempotencyKeyRequired = errors.New("recommendation event の idempotency key が必要です")
	// ErrIdempotencyConflict は recommendation event の idempotency key が別 payload と衝突したことを表します。
	ErrIdempotencyConflict = errors.New("recommendation event の idempotency key が別 payload と衝突しています")
	// ErrAggregateIdentityConflict は recommendation aggregate の identity が既存 row と衝突したことを表します。
	ErrAggregateIdentityConflict = errors.New("recommendation aggregate の identity が既存 row と衝突しています")
	// ErrCreatorUserIDRequired は recommendation event に creator user id が必要なことを表します。
	ErrCreatorUserIDRequired = errors.New("recommendation event の creator user id が必要です")
	// ErrCanonicalMainIDRequired は recommendation event に canonical main id が必要なことを表します。
	ErrCanonicalMainIDRequired = errors.New("recommendation event の canonical main id が必要です")
	// ErrCanonicalMainIDForbidden は recommendation event に canonical main id を含められないことを表します。
	ErrCanonicalMainIDForbidden = errors.New("recommendation event の canonical main id はこの event kind では指定できません")
	// ErrShortIDRequired は recommendation event に short id が必要なことを表します。
	ErrShortIDRequired = errors.New("recommendation event の short id が必要です")
	// ErrShortIDForbidden は recommendation event に short id を含められないことを表します。
	ErrShortIDForbidden = errors.New("recommendation event の short id はこの event kind では指定できません")
	// ErrShortIDInvalid は recommendation event の short id が不正なことを表します。
	ErrShortIDInvalid = errors.New("recommendation event の short id が不正です")
)

const postgresTimestampPrecision = time.Microsecond

// EventKind は recommendation aggregate に記録する signal 種別です。
type EventKind string

const (
	// EventKindImpression records that a public short was shown to the viewer.
	EventKindImpression EventKind = "impression"
	// EventKindViewStart records that the viewer started watching a short.
	EventKindViewStart EventKind = "view_start"
	// EventKindViewCompletion records that the viewer completed a short.
	EventKindViewCompletion EventKind = "view_completion"
	// EventKindRewatchLoop records that the viewer looped or rewatched a short.
	EventKindRewatchLoop EventKind = "rewatch_loop"
	// EventKindProfileClick records that the viewer opened a creator profile.
	EventKindProfileClick EventKind = "profile_click"
	// EventKindMainClick records that the viewer opened a canonical main CTA.
	EventKindMainClick EventKind = "main_click"
	// EventKindUnlockConversion records that the viewer converted into a main unlock.
	EventKindUnlockConversion EventKind = "unlock_conversion"
)

type queries interface {
	GetRecommendationEventByViewerAndIdempotencyKey(ctx context.Context, arg sqlc.GetRecommendationEventByViewerAndIdempotencyKeyParams) (sqlc.AppRecommendationEvent, error)
	InsertRecommendationEvent(ctx context.Context, arg sqlc.InsertRecommendationEventParams) (sqlc.AppRecommendationEvent, error)
	ListRecommendationFollowedCreatorIDsByViewerAndCreatorIDs(ctx context.Context, arg sqlc.ListRecommendationFollowedCreatorIDsByViewerAndCreatorIDsParams) ([]pgtype.UUID, error)
	ListRecommendationPinnedShortIDsByViewerAndShortIDs(ctx context.Context, arg sqlc.ListRecommendationPinnedShortIDsByViewerAndShortIDsParams) ([]pgtype.UUID, error)
	ListRecommendationShortGlobalFeaturesByShortIDs(ctx context.Context, shortIds []pgtype.UUID) ([]sqlc.AppRecommendationShortGlobalFeature, error)
	ListRecommendationUnlockedMainIDsByViewerAndMainIDs(ctx context.Context, arg sqlc.ListRecommendationUnlockedMainIDsByViewerAndMainIDsParams) ([]pgtype.UUID, error)
	ListRecommendationViewerCreatorFeaturesByViewerAndCreatorIDs(ctx context.Context, arg sqlc.ListRecommendationViewerCreatorFeaturesByViewerAndCreatorIDsParams) ([]sqlc.AppRecommendationViewerCreatorFeature, error)
	ListRecommendationViewerMainFeaturesByViewerAndMainIDs(ctx context.Context, arg sqlc.ListRecommendationViewerMainFeaturesByViewerAndMainIDsParams) ([]sqlc.AppRecommendationViewerMainFeature, error)
	ListRecommendationViewerShortFeaturesByViewerAndShortIDs(ctx context.Context, arg sqlc.ListRecommendationViewerShortFeaturesByViewerAndShortIDsParams) ([]sqlc.AppRecommendationViewerShortFeature, error)
	UpsertRecommendationShortGlobalFeatures(ctx context.Context, arg sqlc.UpsertRecommendationShortGlobalFeaturesParams) (int64, error)
	UpsertRecommendationViewerCreatorFeatures(ctx context.Context, arg sqlc.UpsertRecommendationViewerCreatorFeaturesParams) error
	UpsertRecommendationViewerMainFeatures(ctx context.Context, arg sqlc.UpsertRecommendationViewerMainFeaturesParams) (int64, error)
	UpsertRecommendationViewerShortFeatures(ctx context.Context, arg sqlc.UpsertRecommendationViewerShortFeaturesParams) (int64, error)
}

// Repository は recommendation 用の永続化操作を包みます。
type Repository struct {
	beginner   postgres.TxBeginner
	queries    queries
	newQueries func(sqlc.DBTX) queries
}

// RecordEventInput は RecordEvent の入力です。
type RecordEventInput struct {
	ViewerUserID    uuid.UUID
	EventKind       EventKind
	CreatorUserID   *uuid.UUID
	CanonicalMainID *uuid.UUID
	ShortID         *uuid.UUID
	OccurredAt      *time.Time
	IdempotencyKey  string
}

// RecordEventResult は RecordEvent の結果です。
type RecordEventResult struct {
	EventID     uuid.UUID
	EventKind   EventKind
	OccurredAt  time.Time
	Idempotency string
	Recorded    bool
}

// ViewerShortFeatures は viewer x short aggregate row です。
type ViewerShortFeatures struct {
	ShortID                uuid.UUID
	CreatorUserID          uuid.UUID
	CanonicalMainID        uuid.UUID
	ImpressionCount        int64
	LastImpressionAt       *time.Time
	ViewStartCount         int64
	LastViewStartAt        *time.Time
	ViewCompletionCount    int64
	LastViewCompletionAt   *time.Time
	RewatchLoopCount       int64
	LastRewatchLoopAt      *time.Time
	MainClickCount         int64
	LastMainClickAt        *time.Time
	UnlockConversionCount  int64
	LastUnlockConversionAt *time.Time
}

// ViewerCreatorFeatures は viewer x creator aggregate row です。
type ViewerCreatorFeatures struct {
	CreatorUserID          uuid.UUID
	ImpressionCount        int64
	LastImpressionAt       *time.Time
	ViewStartCount         int64
	LastViewStartAt        *time.Time
	ViewCompletionCount    int64
	LastViewCompletionAt   *time.Time
	RewatchLoopCount       int64
	LastRewatchLoopAt      *time.Time
	ProfileClickCount      int64
	LastProfileClickAt     *time.Time
	MainClickCount         int64
	LastMainClickAt        *time.Time
	UnlockConversionCount  int64
	LastUnlockConversionAt *time.Time
}

// ViewerMainFeatures は viewer x canonical main aggregate row です。
type ViewerMainFeatures struct {
	CanonicalMainID        uuid.UUID
	CreatorUserID          uuid.UUID
	ImpressionCount        int64
	LastImpressionAt       *time.Time
	ViewStartCount         int64
	LastViewStartAt        *time.Time
	ViewCompletionCount    int64
	LastViewCompletionAt   *time.Time
	RewatchLoopCount       int64
	LastRewatchLoopAt      *time.Time
	MainClickCount         int64
	LastMainClickAt        *time.Time
	UnlockConversionCount  int64
	LastUnlockConversionAt *time.Time
}

// ShortGlobalFeatures は short global aggregate row です。
type ShortGlobalFeatures struct {
	ShortID                uuid.UUID
	CreatorUserID          uuid.UUID
	CanonicalMainID        uuid.UUID
	ImpressionCount        int64
	LastImpressionAt       *time.Time
	ViewStartCount         int64
	LastViewStartAt        *time.Time
	ViewCompletionCount    int64
	LastViewCompletionAt   *time.Time
	RewatchLoopCount       int64
	LastRewatchLoopAt      *time.Time
	MainClickCount         int64
	LastMainClickAt        *time.Time
	UnlockConversionCount  int64
	LastUnlockConversionAt *time.Time
}

// NewRepository は recommendation repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	return newRepository(
		pool,
		sqlc.New(pool),
		func(db sqlc.DBTX) queries { return sqlc.New(db) },
	)
}

func newRepository(beginner postgres.TxBeginner, q queries, newQueries func(sqlc.DBTX) queries) *Repository {
	return &Repository{
		beginner:   beginner,
		queries:    q,
		newQueries: newQueries,
	}
}

// RecordEvent は recommendation signal を idempotent に保存し、必要な aggregate を更新します。
func (r *Repository) RecordEvent(ctx context.Context, input RecordEventInput) (RecordEventResult, error) {
	if err := validateRecordEventInput(input); err != nil {
		return RecordEventResult{}, fmt.Errorf("recommendation event 記録: %w", err)
	}
	if r == nil || r.beginner == nil {
		return RecordEventResult{}, fmt.Errorf("recommendation repository transaction が初期化されていません")
	}
	if r.newQueries == nil {
		return RecordEventResult{}, fmt.Errorf("recommendation repository query factory が初期化されていません")
	}

	occurredAt := time.Now().UTC()
	if input.OccurredAt != nil {
		occurredAt = input.OccurredAt.UTC()
	}
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)

	var result RecordEventResult
	err := postgres.RunInTx(ctx, r.beginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)

		insertRow, err := q.InsertRecommendationEvent(ctx, sqlc.InsertRecommendationEventParams{
			ViewerUserID:    postgres.UUIDToPG(input.ViewerUserID),
			EventKind:       string(input.EventKind),
			CreatorUserID:   optionalUUIDToPG(input.CreatorUserID),
			CanonicalMainID: optionalUUIDToPG(input.CanonicalMainID),
			ShortID:         optionalUUIDToPG(input.ShortID),
			OccurredAt:      postgres.TimeToPG(&occurredAt),
			IdempotencyKey:  idempotencyKey,
		})
		switch {
		case err == nil:
			result, err = mapRecordEventResult(storedRecommendationEventFromInsertRow(insertRow), true)
			if err != nil {
				return fmt.Errorf("recommendation event 保存結果の変換 viewer=%s key=%q: %w", input.ViewerUserID, idempotencyKey, err)
			}
		case errors.Is(err, pgx.ErrNoRows):
			row, getErr := q.GetRecommendationEventByViewerAndIdempotencyKey(ctx, sqlc.GetRecommendationEventByViewerAndIdempotencyKeyParams{
				ViewerUserID:   postgres.UUIDToPG(input.ViewerUserID),
				IdempotencyKey: idempotencyKey,
			})
			if getErr != nil {
				return fmt.Errorf("recommendation event duplicate 既存行取得 viewer=%s key=%q: %w", input.ViewerUserID, idempotencyKey, getErr)
			}

			storedRow := storedRecommendationEventFromGetRow(row)
			if err := validateDuplicateEventIdentity(storedRow, input); err != nil {
				return fmt.Errorf("recommendation event duplicate 判定 viewer=%s key=%q: %w", input.ViewerUserID, idempotencyKey, err)
			}

			result, err = mapRecordEventResult(storedRow, false)
			if err != nil {
				return fmt.Errorf("recommendation event duplicate 結果の変換 viewer=%s key=%q: %w", input.ViewerUserID, idempotencyKey, err)
			}
			return nil
		default:
			return fmt.Errorf("recommendation event 保存 viewer=%s key=%q: %w", input.ViewerUserID, idempotencyKey, err)
		}

		eventOccurredAt := result.OccurredAt
		delta := buildFeatureDelta(input.EventKind, eventOccurredAt)

		if input.ShortID != nil && delta.appliesToShort() {
			rowsAffected, err := q.UpsertRecommendationViewerShortFeatures(ctx, sqlc.UpsertRecommendationViewerShortFeaturesParams{
				ViewerUserID:           postgres.UUIDToPG(input.ViewerUserID),
				ShortID:                postgres.UUIDToPG(*input.ShortID),
				CreatorUserID:          postgres.UUIDToPG(*input.CreatorUserID),
				CanonicalMainID:        postgres.UUIDToPG(*input.CanonicalMainID),
				ImpressionCount:        delta.ImpressionCount,
				LastImpressionAt:       postgres.TimeToPG(delta.LastImpressionAt),
				ViewStartCount:         delta.ViewStartCount,
				LastViewStartAt:        postgres.TimeToPG(delta.LastViewStartAt),
				ViewCompletionCount:    delta.ViewCompletionCount,
				LastViewCompletionAt:   postgres.TimeToPG(delta.LastViewCompletionAt),
				RewatchLoopCount:       delta.RewatchLoopCount,
				LastRewatchLoopAt:      postgres.TimeToPG(delta.LastRewatchLoopAt),
				MainClickCount:         delta.MainClickCount,
				LastMainClickAt:        postgres.TimeToPG(delta.LastMainClickAt),
				UnlockConversionCount:  delta.UnlockConversionCount,
				LastUnlockConversionAt: postgres.TimeToPG(delta.LastUnlockConversionAt),
			})
			if err != nil {
				return fmt.Errorf("recommendation viewer x short aggregate 更新 viewer=%s short=%s: %w", input.ViewerUserID, *input.ShortID, err)
			}
			if err := ensureAggregateMutation(rowsAffected, "recommendation viewer x short aggregate", input.ViewerUserID, *input.ShortID); err != nil {
				return err
			}

			rowsAffected, err = q.UpsertRecommendationShortGlobalFeatures(ctx, sqlc.UpsertRecommendationShortGlobalFeaturesParams{
				ShortID:                postgres.UUIDToPG(*input.ShortID),
				CreatorUserID:          postgres.UUIDToPG(*input.CreatorUserID),
				CanonicalMainID:        postgres.UUIDToPG(*input.CanonicalMainID),
				ImpressionCount:        delta.ImpressionCount,
				LastImpressionAt:       postgres.TimeToPG(delta.LastImpressionAt),
				ViewStartCount:         delta.ViewStartCount,
				LastViewStartAt:        postgres.TimeToPG(delta.LastViewStartAt),
				ViewCompletionCount:    delta.ViewCompletionCount,
				LastViewCompletionAt:   postgres.TimeToPG(delta.LastViewCompletionAt),
				RewatchLoopCount:       delta.RewatchLoopCount,
				LastRewatchLoopAt:      postgres.TimeToPG(delta.LastRewatchLoopAt),
				MainClickCount:         delta.MainClickCount,
				LastMainClickAt:        postgres.TimeToPG(delta.LastMainClickAt),
				UnlockConversionCount:  delta.UnlockConversionCount,
				LastUnlockConversionAt: postgres.TimeToPG(delta.LastUnlockConversionAt),
			})
			if err != nil {
				return fmt.Errorf("recommendation short global aggregate 更新 short=%s: %w", *input.ShortID, err)
			}
			if err := ensureAggregateMutation(rowsAffected, "recommendation short global aggregate", uuid.Nil, *input.ShortID); err != nil {
				return err
			}
		}
		if delta.appliesToCreator() {
			if err := q.UpsertRecommendationViewerCreatorFeatures(ctx, sqlc.UpsertRecommendationViewerCreatorFeaturesParams{
				ViewerUserID:           postgres.UUIDToPG(input.ViewerUserID),
				CreatorUserID:          postgres.UUIDToPG(*input.CreatorUserID),
				ImpressionCount:        delta.ImpressionCount,
				LastImpressionAt:       postgres.TimeToPG(delta.LastImpressionAt),
				ViewStartCount:         delta.ViewStartCount,
				LastViewStartAt:        postgres.TimeToPG(delta.LastViewStartAt),
				ViewCompletionCount:    delta.ViewCompletionCount,
				LastViewCompletionAt:   postgres.TimeToPG(delta.LastViewCompletionAt),
				RewatchLoopCount:       delta.RewatchLoopCount,
				LastRewatchLoopAt:      postgres.TimeToPG(delta.LastRewatchLoopAt),
				ProfileClickCount:      delta.ProfileClickCount,
				LastProfileClickAt:     postgres.TimeToPG(delta.LastProfileClickAt),
				MainClickCount:         delta.MainClickCount,
				LastMainClickAt:        postgres.TimeToPG(delta.LastMainClickAt),
				UnlockConversionCount:  delta.UnlockConversionCount,
				LastUnlockConversionAt: postgres.TimeToPG(delta.LastUnlockConversionAt),
			}); err != nil {
				return fmt.Errorf("recommendation viewer x creator aggregate 更新 viewer=%s creator=%s: %w", input.ViewerUserID, *input.CreatorUserID, err)
			}
		}
		if delta.appliesToMain() {
			rowsAffected, err := q.UpsertRecommendationViewerMainFeatures(ctx, sqlc.UpsertRecommendationViewerMainFeaturesParams{
				ViewerUserID:           postgres.UUIDToPG(input.ViewerUserID),
				CanonicalMainID:        postgres.UUIDToPG(*input.CanonicalMainID),
				CreatorUserID:          postgres.UUIDToPG(*input.CreatorUserID),
				ImpressionCount:        delta.ImpressionCount,
				LastImpressionAt:       postgres.TimeToPG(delta.LastImpressionAt),
				ViewStartCount:         delta.ViewStartCount,
				LastViewStartAt:        postgres.TimeToPG(delta.LastViewStartAt),
				ViewCompletionCount:    delta.ViewCompletionCount,
				LastViewCompletionAt:   postgres.TimeToPG(delta.LastViewCompletionAt),
				RewatchLoopCount:       delta.RewatchLoopCount,
				LastRewatchLoopAt:      postgres.TimeToPG(delta.LastRewatchLoopAt),
				MainClickCount:         delta.MainClickCount,
				LastMainClickAt:        postgres.TimeToPG(delta.LastMainClickAt),
				UnlockConversionCount:  delta.UnlockConversionCount,
				LastUnlockConversionAt: postgres.TimeToPG(delta.LastUnlockConversionAt),
			})
			if err != nil {
				return fmt.Errorf("recommendation viewer x main aggregate 更新 viewer=%s main=%s: %w", input.ViewerUserID, *input.CanonicalMainID, err)
			}
			if err := ensureAggregateMutation(rowsAffected, "recommendation viewer x main aggregate", input.ViewerUserID, *input.CanonicalMainID); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return RecordEventResult{}, err
	}

	return result, nil
}

// ListViewerShortFeatures は viewer x short aggregate を候補 short 群に対して返します。
func (r *Repository) ListViewerShortFeatures(ctx context.Context, viewerUserID uuid.UUID, shortIDs []uuid.UUID) ([]ViewerShortFeatures, error) {
	if len(shortIDs) == 0 {
		return []ViewerShortFeatures{}, nil
	}

	rows, err := r.queries.ListRecommendationViewerShortFeaturesByViewerAndShortIDs(ctx, sqlc.ListRecommendationViewerShortFeaturesByViewerAndShortIDsParams{
		ViewerUserID: postgres.UUIDToPG(viewerUserID),
		ShortIds:     uuidSliceToPG(shortIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("recommendation viewer x short feature 取得 viewer=%s: %w", viewerUserID, err)
	}

	return mapViewerShortFeatures(rows)
}

// ListViewerCreatorFeatures は viewer x creator aggregate を候補 creator 群に対して返します。
func (r *Repository) ListViewerCreatorFeatures(ctx context.Context, viewerUserID uuid.UUID, creatorUserIDs []uuid.UUID) ([]ViewerCreatorFeatures, error) {
	if len(creatorUserIDs) == 0 {
		return []ViewerCreatorFeatures{}, nil
	}

	rows, err := r.queries.ListRecommendationViewerCreatorFeaturesByViewerAndCreatorIDs(ctx, sqlc.ListRecommendationViewerCreatorFeaturesByViewerAndCreatorIDsParams{
		ViewerUserID:   postgres.UUIDToPG(viewerUserID),
		CreatorUserIds: uuidSliceToPG(creatorUserIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("recommendation viewer x creator feature 取得 viewer=%s: %w", viewerUserID, err)
	}

	return mapViewerCreatorFeatures(rows)
}

// ListViewerMainFeatures は viewer x canonical main aggregate を候補 main 群に対して返します。
func (r *Repository) ListViewerMainFeatures(ctx context.Context, viewerUserID uuid.UUID, canonicalMainIDs []uuid.UUID) ([]ViewerMainFeatures, error) {
	if len(canonicalMainIDs) == 0 {
		return []ViewerMainFeatures{}, nil
	}

	rows, err := r.queries.ListRecommendationViewerMainFeaturesByViewerAndMainIDs(ctx, sqlc.ListRecommendationViewerMainFeaturesByViewerAndMainIDsParams{
		ViewerUserID:     postgres.UUIDToPG(viewerUserID),
		CanonicalMainIds: uuidSliceToPG(canonicalMainIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("recommendation viewer x main feature 取得 viewer=%s: %w", viewerUserID, err)
	}

	return mapViewerMainFeatures(rows)
}

// ListShortGlobalFeatures は候補 short 群に対する global aggregate を返します。
func (r *Repository) ListShortGlobalFeatures(ctx context.Context, shortIDs []uuid.UUID) ([]ShortGlobalFeatures, error) {
	if len(shortIDs) == 0 {
		return []ShortGlobalFeatures{}, nil
	}

	rows, err := r.queries.ListRecommendationShortGlobalFeaturesByShortIDs(ctx, uuidSliceToPG(shortIDs))
	if err != nil {
		return nil, fmt.Errorf("recommendation short global feature 取得: %w", err)
	}

	return mapShortGlobalFeatures(rows)
}

// ListViewerPinnedShortIDs は viewer が pin 済みの候補 short ID を返します。
func (r *Repository) ListViewerPinnedShortIDs(ctx context.Context, viewerUserID uuid.UUID, shortIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(shortIDs) == 0 {
		return []uuid.UUID{}, nil
	}

	rows, err := r.queries.ListRecommendationPinnedShortIDsByViewerAndShortIDs(ctx, sqlc.ListRecommendationPinnedShortIDsByViewerAndShortIDsParams{
		ViewerUserID: postgres.UUIDToPG(viewerUserID),
		ShortIds:     uuidSliceToPG(shortIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("recommendation pinned short state 取得 viewer=%s: %w", viewerUserID, err)
	}

	return uuidSliceFromPG(rows, "recommendation pinned short id")
}

// ListViewerFollowedCreatorIDs は viewer が follow 済みの候補 creator ID を返します。
func (r *Repository) ListViewerFollowedCreatorIDs(ctx context.Context, viewerUserID uuid.UUID, creatorUserIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(creatorUserIDs) == 0 {
		return []uuid.UUID{}, nil
	}

	rows, err := r.queries.ListRecommendationFollowedCreatorIDsByViewerAndCreatorIDs(ctx, sqlc.ListRecommendationFollowedCreatorIDsByViewerAndCreatorIDsParams{
		ViewerUserID:   postgres.UUIDToPG(viewerUserID),
		CreatorUserIds: uuidSliceToPG(creatorUserIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("recommendation followed creator state 取得 viewer=%s: %w", viewerUserID, err)
	}

	return uuidSliceFromPG(rows, "recommendation followed creator id")
}

// ListViewerUnlockedMainIDs は viewer が unlock 済みの候補 main ID を返します。
func (r *Repository) ListViewerUnlockedMainIDs(ctx context.Context, viewerUserID uuid.UUID, canonicalMainIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(canonicalMainIDs) == 0 {
		return []uuid.UUID{}, nil
	}

	rows, err := r.queries.ListRecommendationUnlockedMainIDsByViewerAndMainIDs(ctx, sqlc.ListRecommendationUnlockedMainIDsByViewerAndMainIDsParams{
		ViewerUserID:     postgres.UUIDToPG(viewerUserID),
		CanonicalMainIds: uuidSliceToPG(canonicalMainIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("recommendation unlocked main state 取得 viewer=%s: %w", viewerUserID, err)
	}

	return uuidSliceFromPG(rows, "recommendation unlocked main id")
}

type featureDelta struct {
	ImpressionCount        int64
	LastImpressionAt       *time.Time
	ViewStartCount         int64
	LastViewStartAt        *time.Time
	ViewCompletionCount    int64
	LastViewCompletionAt   *time.Time
	RewatchLoopCount       int64
	LastRewatchLoopAt      *time.Time
	ProfileClickCount      int64
	LastProfileClickAt     *time.Time
	MainClickCount         int64
	LastMainClickAt        *time.Time
	UnlockConversionCount  int64
	LastUnlockConversionAt *time.Time
}

type storedRecommendationEvent struct {
	ID              pgtype.UUID
	EventKind       string
	CreatorUserID   pgtype.UUID
	CanonicalMainID pgtype.UUID
	ShortID         pgtype.UUID
	OccurredAt      pgtype.Timestamptz
	IdempotencyKey  string
}

func buildFeatureDelta(kind EventKind, occurredAt time.Time) featureDelta {
	occurredAt = occurredAt.UTC()

	switch kind {
	case EventKindImpression:
		return featureDelta{ImpressionCount: 1, LastImpressionAt: &occurredAt}
	case EventKindViewStart:
		return featureDelta{ViewStartCount: 1, LastViewStartAt: &occurredAt}
	case EventKindViewCompletion:
		return featureDelta{ViewCompletionCount: 1, LastViewCompletionAt: &occurredAt}
	case EventKindRewatchLoop:
		return featureDelta{RewatchLoopCount: 1, LastRewatchLoopAt: &occurredAt}
	case EventKindProfileClick:
		return featureDelta{ProfileClickCount: 1, LastProfileClickAt: &occurredAt}
	case EventKindMainClick:
		return featureDelta{MainClickCount: 1, LastMainClickAt: &occurredAt}
	case EventKindUnlockConversion:
		return featureDelta{UnlockConversionCount: 1, LastUnlockConversionAt: &occurredAt}
	default:
		return featureDelta{}
	}
}

func (d featureDelta) appliesToShort() bool {
	return d.ImpressionCount > 0 ||
		d.ViewStartCount > 0 ||
		d.ViewCompletionCount > 0 ||
		d.RewatchLoopCount > 0 ||
		d.MainClickCount > 0 ||
		d.UnlockConversionCount > 0
}

func (d featureDelta) appliesToCreator() bool {
	return d.appliesToShort() || d.ProfileClickCount > 0
}

func (d featureDelta) appliesToMain() bool {
	return d.ImpressionCount > 0 ||
		d.ViewStartCount > 0 ||
		d.ViewCompletionCount > 0 ||
		d.RewatchLoopCount > 0 ||
		d.MainClickCount > 0 ||
		d.UnlockConversionCount > 0
}

func validateRecordEventInput(input RecordEventInput) error {
	if input.ViewerUserID == uuid.Nil {
		return ErrViewerUserIDRequired
	}
	if strings.TrimSpace(input.IdempotencyKey) == "" {
		return ErrIdempotencyKeyRequired
	}

	switch input.EventKind {
	case EventKindImpression, EventKindViewStart, EventKindViewCompletion, EventKindRewatchLoop:
		if input.CreatorUserID == nil || *input.CreatorUserID == uuid.Nil {
			return ErrCreatorUserIDRequired
		}
		if input.CanonicalMainID == nil || *input.CanonicalMainID == uuid.Nil {
			return ErrCanonicalMainIDRequired
		}
		if input.ShortID == nil || *input.ShortID == uuid.Nil {
			return ErrShortIDRequired
		}
	case EventKindProfileClick:
		if input.CreatorUserID == nil || *input.CreatorUserID == uuid.Nil {
			return ErrCreatorUserIDRequired
		}
		if input.CanonicalMainID != nil {
			return ErrCanonicalMainIDForbidden
		}
		if input.ShortID != nil {
			return ErrShortIDForbidden
		}
	case EventKindMainClick, EventKindUnlockConversion:
		if input.CreatorUserID == nil || *input.CreatorUserID == uuid.Nil {
			return ErrCreatorUserIDRequired
		}
		if input.CanonicalMainID == nil || *input.CanonicalMainID == uuid.Nil {
			return ErrCanonicalMainIDRequired
		}
		if input.ShortID != nil && *input.ShortID == uuid.Nil {
			return ErrShortIDInvalid
		}
	default:
		return ErrEventKindInvalid
	}

	return nil
}

func mapRecordEventResult(row storedRecommendationEvent, recorded bool) (RecordEventResult, error) {
	eventID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return RecordEventResult{}, fmt.Errorf("event id 変換: %w", err)
	}
	occurredAt, err := postgres.RequiredTimeFromPG(row.OccurredAt)
	if err != nil {
		return RecordEventResult{}, fmt.Errorf("event occurred_at 変換: %w", err)
	}

	return RecordEventResult{
		EventID:     eventID,
		EventKind:   EventKind(row.EventKind),
		OccurredAt:  occurredAt,
		Idempotency: row.IdempotencyKey,
		Recorded:    recorded,
	}, nil
}

func storedRecommendationEventFromInsertRow(row sqlc.AppRecommendationEvent) storedRecommendationEvent {
	return storedRecommendationEvent{
		ID:              row.ID,
		EventKind:       row.EventKind,
		CreatorUserID:   row.CreatorUserID,
		CanonicalMainID: row.CanonicalMainID,
		ShortID:         row.ShortID,
		OccurredAt:      row.OccurredAt,
		IdempotencyKey:  row.IdempotencyKey,
	}
}

func storedRecommendationEventFromGetRow(row sqlc.AppRecommendationEvent) storedRecommendationEvent {
	return storedRecommendationEvent{
		ID:              row.ID,
		EventKind:       row.EventKind,
		CreatorUserID:   row.CreatorUserID,
		CanonicalMainID: row.CanonicalMainID,
		ShortID:         row.ShortID,
		OccurredAt:      row.OccurredAt,
		IdempotencyKey:  row.IdempotencyKey,
	}
}

func validateDuplicateEventIdentity(row storedRecommendationEvent, input RecordEventInput) error {
	if row.EventKind != string(input.EventKind) {
		return fmt.Errorf("%w: event_kind got=%q want=%q", ErrIdempotencyConflict, row.EventKind, input.EventKind)
	}
	if !optionalUUIDMatches(row.CreatorUserID, input.CreatorUserID) {
		return fmt.Errorf("%w: creator_user_id が一致しません", ErrIdempotencyConflict)
	}
	if !optionalUUIDMatches(row.CanonicalMainID, input.CanonicalMainID) {
		return fmt.Errorf("%w: canonical_main_id が一致しません", ErrIdempotencyConflict)
	}
	if !optionalUUIDMatches(row.ShortID, input.ShortID) {
		return fmt.Errorf("%w: short_id が一致しません", ErrIdempotencyConflict)
	}
	if input.OccurredAt != nil {
		occurredAt, err := postgres.RequiredTimeFromPG(row.OccurredAt)
		if err != nil {
			return fmt.Errorf("event occurred_at 変換: %w", err)
		}
		storedOccurredAt := normalizePostgresTimestampPrecision(occurredAt)
		inputOccurredAt := normalizePostgresTimestampPrecision(input.OccurredAt.UTC())
		if !storedOccurredAt.Equal(inputOccurredAt) {
			return fmt.Errorf("%w: occurred_at got=%s want=%s", ErrIdempotencyConflict, storedOccurredAt, inputOccurredAt)
		}
	}

	return nil
}

func mapViewerShortFeatures(rows []sqlc.AppRecommendationViewerShortFeature) ([]ViewerShortFeatures, error) {
	items := make([]ViewerShortFeatures, 0, len(rows))
	for _, row := range rows {
		item, err := mapViewerShortFeature(row)
		if err != nil {
			return nil, fmt.Errorf("recommendation viewer x short feature 変換: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func mapViewerShortFeature(row sqlc.AppRecommendationViewerShortFeature) (ViewerShortFeatures, error) {
	shortID, err := postgres.UUIDFromPG(row.ShortID)
	if err != nil {
		return ViewerShortFeatures{}, fmt.Errorf("short id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return ViewerShortFeatures{}, fmt.Errorf("creator user id 変換: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return ViewerShortFeatures{}, fmt.Errorf("canonical main id 変換: %w", err)
	}

	return ViewerShortFeatures{
		ShortID:                shortID,
		CreatorUserID:          creatorUserID,
		CanonicalMainID:        canonicalMainID,
		ImpressionCount:        row.ImpressionCount,
		LastImpressionAt:       postgres.OptionalTimeFromPG(row.LastImpressionAt),
		ViewStartCount:         row.ViewStartCount,
		LastViewStartAt:        postgres.OptionalTimeFromPG(row.LastViewStartAt),
		ViewCompletionCount:    row.ViewCompletionCount,
		LastViewCompletionAt:   postgres.OptionalTimeFromPG(row.LastViewCompletionAt),
		RewatchLoopCount:       row.RewatchLoopCount,
		LastRewatchLoopAt:      postgres.OptionalTimeFromPG(row.LastRewatchLoopAt),
		MainClickCount:         row.MainClickCount,
		LastMainClickAt:        postgres.OptionalTimeFromPG(row.LastMainClickAt),
		UnlockConversionCount:  row.UnlockConversionCount,
		LastUnlockConversionAt: postgres.OptionalTimeFromPG(row.LastUnlockConversionAt),
	}, nil
}

func mapViewerCreatorFeatures(rows []sqlc.AppRecommendationViewerCreatorFeature) ([]ViewerCreatorFeatures, error) {
	items := make([]ViewerCreatorFeatures, 0, len(rows))
	for _, row := range rows {
		item, err := mapViewerCreatorFeature(row)
		if err != nil {
			return nil, fmt.Errorf("recommendation viewer x creator feature 変換: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func mapViewerCreatorFeature(row sqlc.AppRecommendationViewerCreatorFeature) (ViewerCreatorFeatures, error) {
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return ViewerCreatorFeatures{}, fmt.Errorf("creator user id 変換: %w", err)
	}

	return ViewerCreatorFeatures{
		CreatorUserID:          creatorUserID,
		ImpressionCount:        row.ImpressionCount,
		LastImpressionAt:       postgres.OptionalTimeFromPG(row.LastImpressionAt),
		ViewStartCount:         row.ViewStartCount,
		LastViewStartAt:        postgres.OptionalTimeFromPG(row.LastViewStartAt),
		ViewCompletionCount:    row.ViewCompletionCount,
		LastViewCompletionAt:   postgres.OptionalTimeFromPG(row.LastViewCompletionAt),
		RewatchLoopCount:       row.RewatchLoopCount,
		LastRewatchLoopAt:      postgres.OptionalTimeFromPG(row.LastRewatchLoopAt),
		ProfileClickCount:      row.ProfileClickCount,
		LastProfileClickAt:     postgres.OptionalTimeFromPG(row.LastProfileClickAt),
		MainClickCount:         row.MainClickCount,
		LastMainClickAt:        postgres.OptionalTimeFromPG(row.LastMainClickAt),
		UnlockConversionCount:  row.UnlockConversionCount,
		LastUnlockConversionAt: postgres.OptionalTimeFromPG(row.LastUnlockConversionAt),
	}, nil
}

func mapViewerMainFeatures(rows []sqlc.AppRecommendationViewerMainFeature) ([]ViewerMainFeatures, error) {
	items := make([]ViewerMainFeatures, 0, len(rows))
	for _, row := range rows {
		item, err := mapViewerMainFeature(row)
		if err != nil {
			return nil, fmt.Errorf("recommendation viewer x main feature 変換: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func mapViewerMainFeature(row sqlc.AppRecommendationViewerMainFeature) (ViewerMainFeatures, error) {
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return ViewerMainFeatures{}, fmt.Errorf("canonical main id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return ViewerMainFeatures{}, fmt.Errorf("creator user id 変換: %w", err)
	}

	return ViewerMainFeatures{
		CanonicalMainID:        canonicalMainID,
		CreatorUserID:          creatorUserID,
		ImpressionCount:        row.ImpressionCount,
		LastImpressionAt:       postgres.OptionalTimeFromPG(row.LastImpressionAt),
		ViewStartCount:         row.ViewStartCount,
		LastViewStartAt:        postgres.OptionalTimeFromPG(row.LastViewStartAt),
		ViewCompletionCount:    row.ViewCompletionCount,
		LastViewCompletionAt:   postgres.OptionalTimeFromPG(row.LastViewCompletionAt),
		RewatchLoopCount:       row.RewatchLoopCount,
		LastRewatchLoopAt:      postgres.OptionalTimeFromPG(row.LastRewatchLoopAt),
		MainClickCount:         row.MainClickCount,
		LastMainClickAt:        postgres.OptionalTimeFromPG(row.LastMainClickAt),
		UnlockConversionCount:  row.UnlockConversionCount,
		LastUnlockConversionAt: postgres.OptionalTimeFromPG(row.LastUnlockConversionAt),
	}, nil
}

func mapShortGlobalFeatures(rows []sqlc.AppRecommendationShortGlobalFeature) ([]ShortGlobalFeatures, error) {
	items := make([]ShortGlobalFeatures, 0, len(rows))
	for _, row := range rows {
		item, err := mapShortGlobalFeature(row)
		if err != nil {
			return nil, fmt.Errorf("recommendation short global feature 変換: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func mapShortGlobalFeature(row sqlc.AppRecommendationShortGlobalFeature) (ShortGlobalFeatures, error) {
	shortID, err := postgres.UUIDFromPG(row.ShortID)
	if err != nil {
		return ShortGlobalFeatures{}, fmt.Errorf("short id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return ShortGlobalFeatures{}, fmt.Errorf("creator user id 変換: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return ShortGlobalFeatures{}, fmt.Errorf("canonical main id 変換: %w", err)
	}

	return ShortGlobalFeatures{
		ShortID:                shortID,
		CreatorUserID:          creatorUserID,
		CanonicalMainID:        canonicalMainID,
		ImpressionCount:        row.ImpressionCount,
		LastImpressionAt:       postgres.OptionalTimeFromPG(row.LastImpressionAt),
		ViewStartCount:         row.ViewStartCount,
		LastViewStartAt:        postgres.OptionalTimeFromPG(row.LastViewStartAt),
		ViewCompletionCount:    row.ViewCompletionCount,
		LastViewCompletionAt:   postgres.OptionalTimeFromPG(row.LastViewCompletionAt),
		RewatchLoopCount:       row.RewatchLoopCount,
		LastRewatchLoopAt:      postgres.OptionalTimeFromPG(row.LastRewatchLoopAt),
		MainClickCount:         row.MainClickCount,
		LastMainClickAt:        postgres.OptionalTimeFromPG(row.LastMainClickAt),
		UnlockConversionCount:  row.UnlockConversionCount,
		LastUnlockConversionAt: postgres.OptionalTimeFromPG(row.LastUnlockConversionAt),
	}, nil
}

func optionalUUIDToPG(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}

	return postgres.UUIDToPG(*value)
}

func optionalUUIDMatches(got pgtype.UUID, want *uuid.UUID) bool {
	if want == nil {
		return !got.Valid
	}
	if !got.Valid {
		return false
	}

	return got == postgres.UUIDToPG(*want)
}

func ensureAggregateMutation(rowsAffected int64, label string, viewerID uuid.UUID, targetID uuid.UUID) error {
	if rowsAffected == 1 {
		return nil
	}

	if viewerID == uuid.Nil {
		return fmt.Errorf("%s target=%s: %w", label, targetID, ErrAggregateIdentityConflict)
	}

	return fmt.Errorf("%s viewer=%s target=%s: %w", label, viewerID, targetID, ErrAggregateIdentityConflict)
}

func normalizePostgresTimestampPrecision(value time.Time) time.Time {
	return value.UTC().Round(postgresTimestampPrecision)
}

func uuidSliceToPG(values []uuid.UUID) []pgtype.UUID {
	result := make([]pgtype.UUID, 0, len(values))
	for _, value := range values {
		result = append(result, postgres.UUIDToPG(value))
	}

	return result
}

func uuidSliceFromPG(values []pgtype.UUID, label string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		id, err := postgres.UUIDFromPG(value)
		if err != nil {
			return nil, fmt.Errorf("%s 変換: %w", label, err)
		}
		result = append(result, id)
	}

	return result, nil
}
