package shorts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrMainNotFound は対象の main が存在しないことを表します。
var ErrMainNotFound = errors.New("main が見つかりません")

// ErrUnlockableMainNotFound は対象の unlockable main が存在しないことを表します。
var ErrUnlockableMainNotFound = errors.New("unlockable main が見つかりません")

// ErrShortNotFound は対象の short が存在しないことを表します。
var ErrShortNotFound = errors.New("short が見つかりません")

// ErrLinkedShortsRequired は linked short が 1 件以上必要なことを表します。
var ErrLinkedShortsRequired = errors.New("linked short が 1 件以上必要です")

type queries interface {
	CreateMain(ctx context.Context, arg sqlc.CreateMainParams) (sqlc.AppMain, error)
	GetMainByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMain, error)
	ListMainsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMain, error)
	UpdateMainState(ctx context.Context, arg sqlc.UpdateMainStateParams) (sqlc.AppMain, error)
	GetUnlockableMainByID(ctx context.Context, id pgtype.UUID) (sqlc.AppUnlockableMain, error)
	CreateShort(ctx context.Context, arg sqlc.CreateShortParams) (sqlc.AppShort, error)
	GetShortByID(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error)
	ListShortsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppShort, error)
	UpdateShortState(ctx context.Context, arg sqlc.UpdateShortStateParams) (sqlc.AppShort, error)
	PublishShort(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error)
	ListPublicShortsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppPublicShort, error)
	GetPublicShortByID(ctx context.Context, id pgtype.UUID) (sqlc.AppPublicShort, error)
	ListShortsByCanonicalMainID(ctx context.Context, canonicalMainID pgtype.UUID) ([]sqlc.AppShort, error)
	GetCanonicalMainIDByShortID(ctx context.Context, id pgtype.UUID) (pgtype.UUID, error)
}

// Repository は main / short 関連の永続化操作を包みます。
type Repository struct {
	beginner   postgres.TxBeginner
	queries    queries
	newQueries func(sqlc.DBTX) queries
}

// Main は domain 向けの main レコードです。
type Main struct {
	ID                  uuid.UUID
	CreatorUserID       uuid.UUID
	MediaAssetID        uuid.UUID
	State               string
	ReviewReasonCode    *string
	PostReportState     *string
	PriceMinor          *int64
	CurrencyCode        *string
	OwnershipConfirmed  bool
	ConsentConfirmed    bool
	ApprovedForUnlockAt *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Short は domain 向けの short レコードです。
type Short struct {
	ID                   uuid.UUID
	CreatorUserID        uuid.UUID
	CanonicalMainID      uuid.UUID
	MediaAssetID         uuid.UUID
	State                string
	ReviewReasonCode     *string
	PostReportState      *string
	ApprovedForPublishAt *time.Time
	PublishedAt          *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// CreateMainInput は CreateMain の入力です。
type CreateMainInput struct {
	CreatorUserID       uuid.UUID
	MediaAssetID        uuid.UUID
	State               string
	ReviewReasonCode    *string
	PostReportState     *string
	PriceMinor          *int64
	CurrencyCode        *string
	OwnershipConfirmed  bool
	ConsentConfirmed    bool
	ApprovedForUnlockAt *time.Time
}

// UpdateMainInput は UpdateMain の入力です。
type UpdateMainInput struct {
	ID                  uuid.UUID
	State               string
	ReviewReasonCode    *string
	PostReportState     *string
	PriceMinor          *int64
	CurrencyCode        *string
	OwnershipConfirmed  bool
	ConsentConfirmed    bool
	ApprovedForUnlockAt *time.Time
}

// CreateShortInput は CreateShort の入力です。
type CreateShortInput struct {
	CreatorUserID        uuid.UUID
	CanonicalMainID      uuid.UUID
	MediaAssetID         uuid.UUID
	State                string
	ReviewReasonCode     *string
	PostReportState      *string
	ApprovedForPublishAt *time.Time
	PublishedAt          *time.Time
}

// CreateLinkedShortInput は新規 main に紐づく short 作成入力です。
type CreateLinkedShortInput struct {
	MediaAssetID         uuid.UUID
	State                string
	ReviewReasonCode     *string
	PostReportState      *string
	ApprovedForPublishAt *time.Time
	PublishedAt          *time.Time
}

// UpdateShortInput は UpdateShort の入力です。
type UpdateShortInput struct {
	ID                   uuid.UUID
	State                string
	ReviewReasonCode     *string
	PostReportState      *string
	ApprovedForPublishAt *time.Time
	PublishedAt          *time.Time
}

// CreateMainWithShortsInput は CreateMainWithShorts の入力です。
type CreateMainWithShortsInput struct {
	Main   CreateMainInput
	Shorts []CreateLinkedShortInput
}

// MainWithShorts は CreateMainWithShorts の結果です。
type MainWithShorts struct {
	Main   Main
	Shorts []Short
}

// NewRepository は pgxpool ベースの shorts repository を構築します。
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

// CreateMain は main を作成します。
func (r *Repository) CreateMain(ctx context.Context, input CreateMainInput) (Main, error) {
	row, err := r.queries.CreateMain(ctx, buildCreateMainParams(input))
	if err != nil {
		return Main{}, fmt.Errorf("main 作成: %w", err)
	}

	main, err := mapMain(row)
	if err != nil {
		return Main{}, fmt.Errorf("main 作成結果の変換: %w", err)
	}

	return main, nil
}

// GetMain は ID から main を取得します。
func (r *Repository) GetMain(ctx context.Context, id uuid.UUID) (Main, error) {
	row, err := r.queries.GetMainByID(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Main{}, fmt.Errorf("main 取得 id=%s: %w", id, ErrMainNotFound)
		}

		return Main{}, fmt.Errorf("main 取得 id=%s: %w", id, err)
	}

	main, err := mapMain(row)
	if err != nil {
		return Main{}, fmt.Errorf("main 取得結果の変換 id=%s: %w", id, err)
	}

	return main, nil
}

// GetUnlockableMain は ID から unlock 可能な main を取得します。
func (r *Repository) GetUnlockableMain(ctx context.Context, id uuid.UUID) (Main, error) {
	row, err := r.queries.GetUnlockableMainByID(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Main{}, fmt.Errorf("unlockable main 取得 id=%s: %w", id, ErrUnlockableMainNotFound)
		}

		return Main{}, fmt.Errorf("unlockable main 取得 id=%s: %w", id, err)
	}

	main, err := mapUnlockableMain(row)
	if err != nil {
		return Main{}, fmt.Errorf("unlockable main 取得結果の変換 id=%s: %w", id, err)
	}

	return main, nil
}

// ListMainsByCreator は creator が所有する main 一覧を返します。
func (r *Repository) ListMainsByCreator(ctx context.Context, creatorUserID uuid.UUID) ([]Main, error) {
	rows, err := r.queries.ListMainsByCreatorUserID(ctx, postgres.UUIDToPG(creatorUserID))
	if err != nil {
		return nil, fmt.Errorf("main 一覧取得 creator=%s: %w", creatorUserID, err)
	}

	mains := make([]Main, 0, len(rows))
	for _, row := range rows {
		main, err := mapMain(row)
		if err != nil {
			return nil, fmt.Errorf("main 一覧取得結果の変換 creator=%s: %w", creatorUserID, err)
		}

		mains = append(mains, main)
	}

	return mains, nil
}

// UpdateMain は main の状態関連フィールドを更新します。
func (r *Repository) UpdateMain(ctx context.Context, input UpdateMainInput) (Main, error) {
	row, err := r.queries.UpdateMainState(ctx, sqlc.UpdateMainStateParams{
		State:               input.State,
		ReviewReasonCode:    postgres.TextToPG(input.ReviewReasonCode),
		PostReportState:     postgres.TextToPG(input.PostReportState),
		PriceMinor:          postgres.Int64ToPG(input.PriceMinor),
		CurrencyCode:        postgres.TextToPG(input.CurrencyCode),
		OwnershipConfirmed:  input.OwnershipConfirmed,
		ConsentConfirmed:    input.ConsentConfirmed,
		ApprovedForUnlockAt: postgres.TimeToPG(input.ApprovedForUnlockAt),
		ID:                  postgres.UUIDToPG(input.ID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Main{}, fmt.Errorf("main 更新 id=%s: %w", input.ID, ErrMainNotFound)
		}

		return Main{}, fmt.Errorf("main 更新 id=%s: %w", input.ID, err)
	}

	main, err := mapMain(row)
	if err != nil {
		return Main{}, fmt.Errorf("main 更新結果の変換 id=%s: %w", input.ID, err)
	}

	return main, nil
}

// CreateShort は short を作成します。
func (r *Repository) CreateShort(ctx context.Context, input CreateShortInput) (Short, error) {
	row, err := r.queries.CreateShort(ctx, buildCreateShortParams(input))
	if err != nil {
		return Short{}, fmt.Errorf("short 作成: %w", err)
	}

	short, err := mapShort(row)
	if err != nil {
		return Short{}, fmt.Errorf("short 作成結果の変換: %w", err)
	}

	return short, nil
}

// GetShort は ID から short を取得します。
func (r *Repository) GetShort(ctx context.Context, id uuid.UUID) (Short, error) {
	row, err := r.queries.GetShortByID(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Short{}, fmt.Errorf("short 取得 id=%s: %w", id, ErrShortNotFound)
		}

		return Short{}, fmt.Errorf("short 取得 id=%s: %w", id, err)
	}

	short, err := mapShort(row)
	if err != nil {
		return Short{}, fmt.Errorf("short 取得結果の変換 id=%s: %w", id, err)
	}

	return short, nil
}

// GetPublicShort は ID から公開 short を取得します。
func (r *Repository) GetPublicShort(ctx context.Context, id uuid.UUID) (Short, error) {
	row, err := r.queries.GetPublicShortByID(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Short{}, fmt.Errorf("公開 short 取得 id=%s: %w", id, ErrShortNotFound)
		}

		return Short{}, fmt.Errorf("公開 short 取得 id=%s: %w", id, err)
	}

	short, err := mapPublicShort(row)
	if err != nil {
		return Short{}, fmt.Errorf("公開 short 取得結果の変換 id=%s: %w", id, err)
	}

	return short, nil
}

// ListShortsByCreator は creator が所有する short 一覧を返します。
func (r *Repository) ListShortsByCreator(ctx context.Context, creatorUserID uuid.UUID) ([]Short, error) {
	rows, err := r.queries.ListShortsByCreatorUserID(ctx, postgres.UUIDToPG(creatorUserID))
	if err != nil {
		return nil, fmt.Errorf("short 一覧取得 creator=%s: %w", creatorUserID, err)
	}

	shorts := make([]Short, 0, len(rows))
	for _, row := range rows {
		short, err := mapShort(row)
		if err != nil {
			return nil, fmt.Errorf("short 一覧取得結果の変換 creator=%s: %w", creatorUserID, err)
		}

		shorts = append(shorts, short)
	}

	return shorts, nil
}

// ListPublicShortsByCreator は creator の公開 short 一覧を返します。
func (r *Repository) ListPublicShortsByCreator(ctx context.Context, creatorUserID uuid.UUID) ([]Short, error) {
	rows, err := r.queries.ListPublicShortsByCreatorUserID(ctx, postgres.UUIDToPG(creatorUserID))
	if err != nil {
		return nil, fmt.Errorf("公開 short 一覧取得 creator=%s: %w", creatorUserID, err)
	}

	shorts := make([]Short, 0, len(rows))
	for _, row := range rows {
		short, err := mapPublicShort(row)
		if err != nil {
			return nil, fmt.Errorf("公開 short 一覧取得結果の変換 creator=%s: %w", creatorUserID, err)
		}

		shorts = append(shorts, short)
	}

	return shorts, nil
}

// ListShortsByCanonicalMain は canonical main に紐づく short 一覧を返します。
func (r *Repository) ListShortsByCanonicalMain(ctx context.Context, mainID uuid.UUID) ([]Short, error) {
	rows, err := r.queries.ListShortsByCanonicalMainID(ctx, postgres.UUIDToPG(mainID))
	if err != nil {
		return nil, fmt.Errorf("canonical main 配下の short 一覧取得 main=%s: %w", mainID, err)
	}

	shorts := make([]Short, 0, len(rows))
	for _, row := range rows {
		short, err := mapShort(row)
		if err != nil {
			return nil, fmt.Errorf("canonical main 配下の short 一覧取得結果の変換 main=%s: %w", mainID, err)
		}

		shorts = append(shorts, short)
	}

	return shorts, nil
}

// GetCanonicalMainIDByShort は short に対応する canonical main ID を返します。
func (r *Repository) GetCanonicalMainIDByShort(ctx context.Context, shortID uuid.UUID) (uuid.UUID, error) {
	row, err := r.queries.GetCanonicalMainIDByShortID(ctx, postgres.UUIDToPG(shortID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("short の canonical main 取得 short=%s: %w", shortID, ErrShortNotFound)
		}

		return uuid.Nil, fmt.Errorf("short の canonical main 取得 short=%s: %w", shortID, err)
	}

	mainID, err := postgres.UUIDFromPG(row)
	if err != nil {
		return uuid.Nil, fmt.Errorf("short の canonical main 取得結果の変換 short=%s: %w", shortID, err)
	}

	return mainID, nil
}

// UpdateShort は short の状態関連フィールドを更新します。
func (r *Repository) UpdateShort(ctx context.Context, input UpdateShortInput) (Short, error) {
	row, err := r.queries.UpdateShortState(ctx, sqlc.UpdateShortStateParams{
		State:                input.State,
		ReviewReasonCode:     postgres.TextToPG(input.ReviewReasonCode),
		PostReportState:      postgres.TextToPG(input.PostReportState),
		ApprovedForPublishAt: postgres.TimeToPG(input.ApprovedForPublishAt),
		PublishedAt:          postgres.TimeToPG(input.PublishedAt),
		ID:                   postgres.UUIDToPG(input.ID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Short{}, fmt.Errorf("short 更新 id=%s: %w", input.ID, ErrShortNotFound)
		}

		return Short{}, fmt.Errorf("short 更新 id=%s: %w", input.ID, err)
	}

	short, err := mapShort(row)
	if err != nil {
		return Short{}, fmt.Errorf("short 更新結果の変換 id=%s: %w", input.ID, err)
	}

	return short, nil
}

// PublishShort は short を公開状態にします。
func (r *Repository) PublishShort(ctx context.Context, id uuid.UUID) (Short, error) {
	row, err := r.queries.PublishShort(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Short{}, fmt.Errorf("short 公開 id=%s: %w", id, ErrShortNotFound)
		}

		return Short{}, fmt.Errorf("short 公開 id=%s: %w", id, err)
	}

	short, err := mapShort(row)
	if err != nil {
		return Short{}, fmt.Errorf("short 公開結果の変換 id=%s: %w", id, err)
	}

	return short, nil
}

// CreateMainWithShorts は main と 1 件以上の linked short を同一 transaction で作成します。
func (r *Repository) CreateMainWithShorts(ctx context.Context, input CreateMainWithShortsInput) (MainWithShorts, error) {
	if len(input.Shorts) == 0 {
		return MainWithShorts{}, fmt.Errorf("main と short の一括作成: %w", ErrLinkedShortsRequired)
	}

	var result MainWithShorts
	err := postgres.RunInTx(ctx, r.beginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)

		mainRow, err := q.CreateMain(ctx, buildCreateMainParams(input.Main))
		if err != nil {
			return fmt.Errorf("main 作成: %w", err)
		}

		main, err := mapMain(mainRow)
		if err != nil {
			return fmt.Errorf("main 作成結果の変換: %w", err)
		}

		shorts := make([]Short, 0, len(input.Shorts))
		for index, shortInput := range input.Shorts {
			shortRow, err := q.CreateShort(ctx, sqlc.CreateShortParams{
				CreatorUserID:        postgres.UUIDToPG(input.Main.CreatorUserID),
				CanonicalMainID:      mainRow.ID,
				MediaAssetID:         postgres.UUIDToPG(shortInput.MediaAssetID),
				State:                shortInput.State,
				ReviewReasonCode:     postgres.TextToPG(shortInput.ReviewReasonCode),
				PostReportState:      postgres.TextToPG(shortInput.PostReportState),
				ApprovedForPublishAt: postgres.TimeToPG(shortInput.ApprovedForPublishAt),
				PublishedAt:          postgres.TimeToPG(shortInput.PublishedAt),
			})
			if err != nil {
				return fmt.Errorf("linked short 作成 index=%d: %w", index, err)
			}

			short, err := mapShort(shortRow)
			if err != nil {
				return fmt.Errorf("linked short 作成結果の変換 index=%d: %w", index, err)
			}

			shorts = append(shorts, short)
		}

		result = MainWithShorts{
			Main:   main,
			Shorts: shorts,
		}
		return nil
	})
	if err != nil {
		return MainWithShorts{}, fmt.Errorf("main と short の一括作成: %w", err)
	}

	return result, nil
}

func buildCreateMainParams(input CreateMainInput) sqlc.CreateMainParams {
	return sqlc.CreateMainParams{
		CreatorUserID:       postgres.UUIDToPG(input.CreatorUserID),
		MediaAssetID:        postgres.UUIDToPG(input.MediaAssetID),
		State:               input.State,
		ReviewReasonCode:    postgres.TextToPG(input.ReviewReasonCode),
		PostReportState:     postgres.TextToPG(input.PostReportState),
		PriceMinor:          postgres.Int64ToPG(input.PriceMinor),
		CurrencyCode:        postgres.TextToPG(input.CurrencyCode),
		OwnershipConfirmed:  input.OwnershipConfirmed,
		ConsentConfirmed:    input.ConsentConfirmed,
		ApprovedForUnlockAt: postgres.TimeToPG(input.ApprovedForUnlockAt),
	}
}

func buildCreateShortParams(input CreateShortInput) sqlc.CreateShortParams {
	return sqlc.CreateShortParams{
		CreatorUserID:        postgres.UUIDToPG(input.CreatorUserID),
		CanonicalMainID:      postgres.UUIDToPG(input.CanonicalMainID),
		MediaAssetID:         postgres.UUIDToPG(input.MediaAssetID),
		State:                input.State,
		ReviewReasonCode:     postgres.TextToPG(input.ReviewReasonCode),
		PostReportState:      postgres.TextToPG(input.PostReportState),
		ApprovedForPublishAt: postgres.TimeToPG(input.ApprovedForPublishAt),
		PublishedAt:          postgres.TimeToPG(input.PublishedAt),
	}
}

func mapMain(row sqlc.AppMain) (Main, error) {
	id, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return Main{}, fmt.Errorf("main の id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return Main{}, fmt.Errorf("main の creator user id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return Main{}, fmt.Errorf("main の media asset id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Main{}, fmt.Errorf("main の created_at 変換: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Main{}, fmt.Errorf("main の updated_at 変換: %w", err)
	}

	return Main{
		ID:                  id,
		CreatorUserID:       creatorUserID,
		MediaAssetID:        mediaAssetID,
		State:               row.State,
		ReviewReasonCode:    postgres.OptionalTextFromPG(row.ReviewReasonCode),
		PostReportState:     postgres.OptionalTextFromPG(row.PostReportState),
		PriceMinor:          postgres.OptionalInt64FromPG(row.PriceMinor),
		CurrencyCode:        postgres.OptionalTextFromPG(row.CurrencyCode),
		OwnershipConfirmed:  row.OwnershipConfirmed,
		ConsentConfirmed:    row.ConsentConfirmed,
		ApprovedForUnlockAt: postgres.OptionalTimeFromPG(row.ApprovedForUnlockAt),
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}, nil
}

func mapUnlockableMain(row sqlc.AppUnlockableMain) (Main, error) {
	id, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return Main{}, fmt.Errorf("unlockable main の id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return Main{}, fmt.Errorf("unlockable main の creator user id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return Main{}, fmt.Errorf("unlockable main の media asset id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Main{}, fmt.Errorf("unlockable main の created_at 変換: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Main{}, fmt.Errorf("unlockable main の updated_at 変換: %w", err)
	}

	return Main{
		ID:                  id,
		CreatorUserID:       creatorUserID,
		MediaAssetID:        mediaAssetID,
		State:               row.State,
		ReviewReasonCode:    postgres.OptionalTextFromPG(row.ReviewReasonCode),
		PostReportState:     postgres.OptionalTextFromPG(row.PostReportState),
		PriceMinor:          postgres.OptionalInt64FromPG(row.PriceMinor),
		CurrencyCode:        postgres.OptionalTextFromPG(row.CurrencyCode),
		OwnershipConfirmed:  row.OwnershipConfirmed,
		ConsentConfirmed:    row.ConsentConfirmed,
		ApprovedForUnlockAt: postgres.OptionalTimeFromPG(row.ApprovedForUnlockAt),
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}, nil
}

func mapShort(row sqlc.AppShort) (Short, error) {
	id, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return Short{}, fmt.Errorf("short の id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return Short{}, fmt.Errorf("short の creator user id 変換: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return Short{}, fmt.Errorf("short の canonical main id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return Short{}, fmt.Errorf("short の media asset id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Short{}, fmt.Errorf("short の created_at 変換: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Short{}, fmt.Errorf("short の updated_at 変換: %w", err)
	}

	return Short{
		ID:                   id,
		CreatorUserID:        creatorUserID,
		CanonicalMainID:      canonicalMainID,
		MediaAssetID:         mediaAssetID,
		State:                row.State,
		ReviewReasonCode:     postgres.OptionalTextFromPG(row.ReviewReasonCode),
		PostReportState:      postgres.OptionalTextFromPG(row.PostReportState),
		ApprovedForPublishAt: postgres.OptionalTimeFromPG(row.ApprovedForPublishAt),
		PublishedAt:          postgres.OptionalTimeFromPG(row.PublishedAt),
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
	}, nil
}

func mapPublicShort(row sqlc.AppPublicShort) (Short, error) {
	id, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return Short{}, fmt.Errorf("公開 short の id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return Short{}, fmt.Errorf("公開 short の creator user id 変換: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return Short{}, fmt.Errorf("公開 short の canonical main id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return Short{}, fmt.Errorf("公開 short の media asset id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Short{}, fmt.Errorf("公開 short の created_at 変換: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Short{}, fmt.Errorf("公開 short の updated_at 変換: %w", err)
	}

	return Short{
		ID:                   id,
		CreatorUserID:        creatorUserID,
		CanonicalMainID:      canonicalMainID,
		MediaAssetID:         mediaAssetID,
		State:                row.State,
		ReviewReasonCode:     postgres.OptionalTextFromPG(row.ReviewReasonCode),
		PostReportState:      postgres.OptionalTextFromPG(row.PostReportState),
		ApprovedForPublishAt: postgres.OptionalTimeFromPG(row.ApprovedForPublishAt),
		PublishedAt:          postgres.OptionalTimeFromPG(row.PublishedAt),
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
	}, nil
}
