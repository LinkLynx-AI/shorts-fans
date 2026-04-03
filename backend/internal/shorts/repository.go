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

// ErrMainNotFound indicates that the requested main does not exist.
var ErrMainNotFound = errors.New("main not found")

// ErrUnlockableMainNotFound indicates that the requested unlockable main does not exist.
var ErrUnlockableMainNotFound = errors.New("unlockable main not found")

// ErrShortNotFound indicates that the requested short does not exist.
var ErrShortNotFound = errors.New("short not found")

// ErrLinkedShortsRequired indicates that at least one linked short must be created with the main.
var ErrLinkedShortsRequired = errors.New("linked shorts are required")

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

// Repository wraps main/short persistence operations.
type Repository struct {
	beginner   postgres.TxBeginner
	queries    queries
	newQueries func(sqlc.DBTX) queries
}

// Main is the domain-facing main record.
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

// Short is the domain-facing short record.
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

// CreateMainInput is the input for CreateMain.
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

// UpdateMainInput is the input for UpdateMain.
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

// CreateShortInput is the input for CreateShort.
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

// CreateLinkedShortInput is the input for creating a short linked to a newly-created main.
type CreateLinkedShortInput struct {
	MediaAssetID         uuid.UUID
	State                string
	ReviewReasonCode     *string
	PostReportState      *string
	ApprovedForPublishAt *time.Time
	PublishedAt          *time.Time
}

// UpdateShortInput is the input for UpdateShort.
type UpdateShortInput struct {
	ID                   uuid.UUID
	State                string
	ReviewReasonCode     *string
	PostReportState      *string
	ApprovedForPublishAt *time.Time
	PublishedAt          *time.Time
}

// CreateMainWithShortsInput is the input for CreateMainWithShorts.
type CreateMainWithShortsInput struct {
	Main   CreateMainInput
	Shorts []CreateLinkedShortInput
}

// MainWithShorts is the result of CreateMainWithShorts.
type MainWithShorts struct {
	Main   Main
	Shorts []Short
}

// NewRepository constructs a shorts repository backed by pgxpool.
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

// CreateMain creates a main row.
func (r *Repository) CreateMain(ctx context.Context, input CreateMainInput) (Main, error) {
	row, err := r.queries.CreateMain(ctx, buildCreateMainParams(input))
	if err != nil {
		return Main{}, fmt.Errorf("create main: %w", err)
	}

	main, err := mapMain(row)
	if err != nil {
		return Main{}, fmt.Errorf("create main: %w", err)
	}

	return main, nil
}

// GetMain returns a main by ID.
func (r *Repository) GetMain(ctx context.Context, id uuid.UUID) (Main, error) {
	row, err := r.queries.GetMainByID(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Main{}, fmt.Errorf("get main %s: %w", id, ErrMainNotFound)
		}

		return Main{}, fmt.Errorf("get main %s: %w", id, err)
	}

	main, err := mapMain(row)
	if err != nil {
		return Main{}, fmt.Errorf("get main %s: %w", id, err)
	}

	return main, nil
}

// GetUnlockableMain returns an unlockable main by ID.
func (r *Repository) GetUnlockableMain(ctx context.Context, id uuid.UUID) (Main, error) {
	row, err := r.queries.GetUnlockableMainByID(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Main{}, fmt.Errorf("get unlockable main %s: %w", id, ErrUnlockableMainNotFound)
		}

		return Main{}, fmt.Errorf("get unlockable main %s: %w", id, err)
	}

	main, err := mapUnlockableMain(row)
	if err != nil {
		return Main{}, fmt.Errorf("get unlockable main %s: %w", id, err)
	}

	return main, nil
}

// ListMainsByCreator returns mains owned by the creator.
func (r *Repository) ListMainsByCreator(ctx context.Context, creatorUserID uuid.UUID) ([]Main, error) {
	rows, err := r.queries.ListMainsByCreatorUserID(ctx, postgres.UUIDToPG(creatorUserID))
	if err != nil {
		return nil, fmt.Errorf("list mains for creator %s: %w", creatorUserID, err)
	}

	mains := make([]Main, 0, len(rows))
	for _, row := range rows {
		main, err := mapMain(row)
		if err != nil {
			return nil, fmt.Errorf("list mains for creator %s: %w", creatorUserID, err)
		}

		mains = append(mains, main)
	}

	return mains, nil
}

// UpdateMain updates main state fields.
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
			return Main{}, fmt.Errorf("update main %s: %w", input.ID, ErrMainNotFound)
		}

		return Main{}, fmt.Errorf("update main %s: %w", input.ID, err)
	}

	main, err := mapMain(row)
	if err != nil {
		return Main{}, fmt.Errorf("update main %s: %w", input.ID, err)
	}

	return main, nil
}

// CreateShort creates a short row.
func (r *Repository) CreateShort(ctx context.Context, input CreateShortInput) (Short, error) {
	row, err := r.queries.CreateShort(ctx, buildCreateShortParams(input))
	if err != nil {
		return Short{}, fmt.Errorf("create short: %w", err)
	}

	short, err := mapShort(row)
	if err != nil {
		return Short{}, fmt.Errorf("create short: %w", err)
	}

	return short, nil
}

// GetShort returns a short by ID.
func (r *Repository) GetShort(ctx context.Context, id uuid.UUID) (Short, error) {
	row, err := r.queries.GetShortByID(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Short{}, fmt.Errorf("get short %s: %w", id, ErrShortNotFound)
		}

		return Short{}, fmt.Errorf("get short %s: %w", id, err)
	}

	short, err := mapShort(row)
	if err != nil {
		return Short{}, fmt.Errorf("get short %s: %w", id, err)
	}

	return short, nil
}

// GetPublicShort returns a public short by ID.
func (r *Repository) GetPublicShort(ctx context.Context, id uuid.UUID) (Short, error) {
	row, err := r.queries.GetPublicShortByID(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Short{}, fmt.Errorf("get public short %s: %w", id, ErrShortNotFound)
		}

		return Short{}, fmt.Errorf("get public short %s: %w", id, err)
	}

	short, err := mapPublicShort(row)
	if err != nil {
		return Short{}, fmt.Errorf("get public short %s: %w", id, err)
	}

	return short, nil
}

// ListShortsByCreator returns shorts owned by the creator.
func (r *Repository) ListShortsByCreator(ctx context.Context, creatorUserID uuid.UUID) ([]Short, error) {
	rows, err := r.queries.ListShortsByCreatorUserID(ctx, postgres.UUIDToPG(creatorUserID))
	if err != nil {
		return nil, fmt.Errorf("list shorts for creator %s: %w", creatorUserID, err)
	}

	shorts := make([]Short, 0, len(rows))
	for _, row := range rows {
		short, err := mapShort(row)
		if err != nil {
			return nil, fmt.Errorf("list shorts for creator %s: %w", creatorUserID, err)
		}

		shorts = append(shorts, short)
	}

	return shorts, nil
}

// ListPublicShortsByCreator returns public shorts owned by the creator.
func (r *Repository) ListPublicShortsByCreator(ctx context.Context, creatorUserID uuid.UUID) ([]Short, error) {
	rows, err := r.queries.ListPublicShortsByCreatorUserID(ctx, postgres.UUIDToPG(creatorUserID))
	if err != nil {
		return nil, fmt.Errorf("list public shorts for creator %s: %w", creatorUserID, err)
	}

	shorts := make([]Short, 0, len(rows))
	for _, row := range rows {
		short, err := mapPublicShort(row)
		if err != nil {
			return nil, fmt.Errorf("list public shorts for creator %s: %w", creatorUserID, err)
		}

		shorts = append(shorts, short)
	}

	return shorts, nil
}

// ListShortsByCanonicalMain returns all shorts linked to a main.
func (r *Repository) ListShortsByCanonicalMain(ctx context.Context, mainID uuid.UUID) ([]Short, error) {
	rows, err := r.queries.ListShortsByCanonicalMainID(ctx, postgres.UUIDToPG(mainID))
	if err != nil {
		return nil, fmt.Errorf("list shorts by canonical main %s: %w", mainID, err)
	}

	shorts := make([]Short, 0, len(rows))
	for _, row := range rows {
		short, err := mapShort(row)
		if err != nil {
			return nil, fmt.Errorf("list shorts by canonical main %s: %w", mainID, err)
		}

		shorts = append(shorts, short)
	}

	return shorts, nil
}

// GetCanonicalMainIDByShort returns the canonical main ID for a short.
func (r *Repository) GetCanonicalMainIDByShort(ctx context.Context, shortID uuid.UUID) (uuid.UUID, error) {
	row, err := r.queries.GetCanonicalMainIDByShortID(ctx, postgres.UUIDToPG(shortID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("get canonical main id for short %s: %w", shortID, ErrShortNotFound)
		}

		return uuid.Nil, fmt.Errorf("get canonical main id for short %s: %w", shortID, err)
	}

	mainID, err := postgres.UUIDFromPG(row)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get canonical main id for short %s: %w", shortID, err)
	}

	return mainID, nil
}

// UpdateShort updates short state fields.
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
			return Short{}, fmt.Errorf("update short %s: %w", input.ID, ErrShortNotFound)
		}

		return Short{}, fmt.Errorf("update short %s: %w", input.ID, err)
	}

	short, err := mapShort(row)
	if err != nil {
		return Short{}, fmt.Errorf("update short %s: %w", input.ID, err)
	}

	return short, nil
}

// PublishShort marks a short as published.
func (r *Repository) PublishShort(ctx context.Context, id uuid.UUID) (Short, error) {
	row, err := r.queries.PublishShort(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Short{}, fmt.Errorf("publish short %s: %w", id, ErrShortNotFound)
		}

		return Short{}, fmt.Errorf("publish short %s: %w", id, err)
	}

	short, err := mapShort(row)
	if err != nil {
		return Short{}, fmt.Errorf("publish short %s: %w", id, err)
	}

	return short, nil
}

// CreateMainWithShorts creates a main and one or more linked shorts in a single transaction.
func (r *Repository) CreateMainWithShorts(ctx context.Context, input CreateMainWithShortsInput) (MainWithShorts, error) {
	if len(input.Shorts) == 0 {
		return MainWithShorts{}, fmt.Errorf("create main with shorts: %w", ErrLinkedShortsRequired)
	}

	var result MainWithShorts
	err := postgres.RunInTx(ctx, r.beginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)

		mainRow, err := q.CreateMain(ctx, buildCreateMainParams(input.Main))
		if err != nil {
			return fmt.Errorf("create main: %w", err)
		}

		main, err := mapMain(mainRow)
		if err != nil {
			return fmt.Errorf("map main: %w", err)
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
				return fmt.Errorf("create linked short %d: %w", index, err)
			}

			short, err := mapShort(shortRow)
			if err != nil {
				return fmt.Errorf("map linked short %d: %w", index, err)
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
		return MainWithShorts{}, fmt.Errorf("create main with shorts: %w", err)
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
		return Main{}, fmt.Errorf("map main id: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return Main{}, fmt.Errorf("map main creator user id: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return Main{}, fmt.Errorf("map main media asset id: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Main{}, fmt.Errorf("map main created at: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Main{}, fmt.Errorf("map main updated at: %w", err)
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
		return Main{}, fmt.Errorf("map unlockable main id: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return Main{}, fmt.Errorf("map unlockable main creator user id: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return Main{}, fmt.Errorf("map unlockable main media asset id: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Main{}, fmt.Errorf("map unlockable main created at: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Main{}, fmt.Errorf("map unlockable main updated at: %w", err)
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
		return Short{}, fmt.Errorf("map short id: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return Short{}, fmt.Errorf("map short creator user id: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return Short{}, fmt.Errorf("map short canonical main id: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return Short{}, fmt.Errorf("map short media asset id: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Short{}, fmt.Errorf("map short created at: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Short{}, fmt.Errorf("map short updated at: %w", err)
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
		return Short{}, fmt.Errorf("map public short id: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return Short{}, fmt.Errorf("map public short creator user id: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return Short{}, fmt.Errorf("map public short canonical main id: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return Short{}, fmt.Errorf("map public short media asset id: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Short{}, fmt.Errorf("map public short created at: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Short{}, fmt.Errorf("map public short updated at: %w", err)
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
