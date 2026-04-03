package unlock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrMainUnlockNotFound indicates that the requested main unlock does not exist.
var ErrMainUnlockNotFound = errors.New("main unlock not found")

// ErrAlreadyUnlocked indicates that the user already unlocked the target main.
var ErrAlreadyUnlocked = errors.New("main already unlocked")

type queries interface {
	CreateMainUnlock(ctx context.Context, arg sqlc.CreateMainUnlockParams) (sqlc.AppMainUnlock, error)
	GetMainUnlockByUserIDAndMainID(ctx context.Context, arg sqlc.GetMainUnlockByUserIDAndMainIDParams) (sqlc.AppMainUnlock, error)
	ListUnlockedMainIDsByUserID(ctx context.Context, userID pgtype.UUID) ([]pgtype.UUID, error)
}

// Repository wraps unlock-related persistence operations.
type Repository struct {
	queries queries
}

// MainUnlock is the domain-facing main unlock record.
type MainUnlock struct {
	UserID                     uuid.UUID
	MainID                     uuid.UUID
	PaymentProviderPurchaseRef *string
	PurchasedAt                time.Time
	CreatedAt                  time.Time
}

// RecordMainUnlockInput is the input for RecordMainUnlock.
type RecordMainUnlockInput struct {
	UserID                     uuid.UUID
	MainID                     uuid.UUID
	PaymentProviderPurchaseRef *string
	PurchasedAt                *time.Time
}

// NewRepository constructs an unlock repository backed by pgxpool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return newRepository(sqlc.New(pool))
}

func newRepository(q queries) *Repository {
	return &Repository{queries: q}
}

// RecordMainUnlock records a successful purchase for a main.
func (r *Repository) RecordMainUnlock(ctx context.Context, input RecordMainUnlockInput) (MainUnlock, error) {
	row, err := r.queries.CreateMainUnlock(ctx, sqlc.CreateMainUnlockParams{
		UserID:                     postgres.UUIDToPG(input.UserID),
		MainID:                     postgres.UUIDToPG(input.MainID),
		PaymentProviderPurchaseRef: postgres.TextToPG(input.PaymentProviderPurchaseRef),
		PurchasedAt:                postgres.TimeToPG(input.PurchasedAt),
	})
	if err != nil {
		if isAlreadyUnlockedError(err) {
			return MainUnlock{}, fmt.Errorf("record main unlock user=%s main=%s: %w", input.UserID, input.MainID, ErrAlreadyUnlocked)
		}

		return MainUnlock{}, fmt.Errorf("record main unlock user=%s main=%s: %w", input.UserID, input.MainID, err)
	}

	mainUnlock, err := mapMainUnlock(row)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("record main unlock user=%s main=%s: %w", input.UserID, input.MainID, err)
	}

	return mainUnlock, nil
}

// GetMainUnlock returns a main unlock by user ID and main ID.
func (r *Repository) GetMainUnlock(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) (MainUnlock, error) {
	row, err := r.queries.GetMainUnlockByUserIDAndMainID(ctx, sqlc.GetMainUnlockByUserIDAndMainIDParams{
		UserID: postgres.UUIDToPG(userID),
		MainID: postgres.UUIDToPG(mainID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MainUnlock{}, fmt.Errorf("get main unlock user=%s main=%s: %w", userID, mainID, ErrMainUnlockNotFound)
		}

		return MainUnlock{}, fmt.Errorf("get main unlock user=%s main=%s: %w", userID, mainID, err)
	}

	mainUnlock, err := mapMainUnlock(row)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("get main unlock user=%s main=%s: %w", userID, mainID, err)
	}

	return mainUnlock, nil
}

// ListUnlockedMainIDs returns unlocked main IDs for a user.
func (r *Repository) ListUnlockedMainIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.queries.ListUnlockedMainIDsByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		return nil, fmt.Errorf("list unlocked main ids for user %s: %w", userID, err)
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		id, err := postgres.UUIDFromPG(row)
		if err != nil {
			return nil, fmt.Errorf("list unlocked main ids for user %s: %w", userID, err)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func mapMainUnlock(row sqlc.AppMainUnlock) (MainUnlock, error) {
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("map main unlock user id: %w", err)
	}
	mainID, err := postgres.UUIDFromPG(row.MainID)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("map main unlock main id: %w", err)
	}
	purchasedAt, err := postgres.RequiredTimeFromPG(row.PurchasedAt)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("map main unlock purchased at: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("map main unlock created at: %w", err)
	}

	return MainUnlock{
		UserID:                     userID,
		MainID:                     mainID,
		PaymentProviderPurchaseRef: postgres.OptionalTextFromPG(row.PaymentProviderPurchaseRef),
		PurchasedAt:                purchasedAt,
		CreatedAt:                  createdAt,
	}, nil
}

func isAlreadyUnlockedError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == "23505" && pgErr.ConstraintName == "main_unlocks_pkey"
}
