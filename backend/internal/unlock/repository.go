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

// ErrMainUnlockNotFound は対象の main unlock が存在しないことを表します。
var ErrMainUnlockNotFound = errors.New("main unlock が見つかりません")

// ErrAlreadyUnlocked は対象の main がすでに unlock 済みであることを表します。
var ErrAlreadyUnlocked = errors.New("main はすでに unlock 済みです")

type queries interface {
	CreateMainUnlock(ctx context.Context, arg sqlc.CreateMainUnlockParams) (sqlc.AppMainUnlock, error)
	GetMainUnlockByUserIDAndMainID(ctx context.Context, arg sqlc.GetMainUnlockByUserIDAndMainIDParams) (sqlc.AppMainUnlock, error)
	ListUnlockedMainIDsByUserID(ctx context.Context, userID pgtype.UUID) ([]pgtype.UUID, error)
}

// Repository は unlock 関連の永続化操作を包みます。
type Repository struct {
	queries queries
}

// MainUnlock は domain 向けの main unlock レコードです。
type MainUnlock struct {
	UserID                     uuid.UUID
	MainID                     uuid.UUID
	PaymentProviderPurchaseRef *string
	PurchasedAt                time.Time
	CreatedAt                  time.Time
}

// RecordMainUnlockInput は RecordMainUnlock の入力です。
type RecordMainUnlockInput struct {
	UserID                     uuid.UUID
	MainID                     uuid.UUID
	PaymentProviderPurchaseRef *string
	PurchasedAt                *time.Time
}

// NewRepository は pgxpool ベースの unlock repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	return newRepository(sqlc.New(pool))
}

func newRepository(q queries) *Repository {
	return &Repository{queries: q}
}

// RecordMainUnlock は main の購入記録を保存します。
func (r *Repository) RecordMainUnlock(ctx context.Context, input RecordMainUnlockInput) (MainUnlock, error) {
	row, err := r.queries.CreateMainUnlock(ctx, sqlc.CreateMainUnlockParams{
		UserID:                     postgres.UUIDToPG(input.UserID),
		MainID:                     postgres.UUIDToPG(input.MainID),
		PaymentProviderPurchaseRef: postgres.TextToPG(input.PaymentProviderPurchaseRef),
		PurchasedAt:                postgres.TimeToPG(input.PurchasedAt),
	})
	if err != nil {
		if isAlreadyUnlockedError(err) {
			return MainUnlock{}, fmt.Errorf("main unlock 記録 user=%s main=%s: %w", input.UserID, input.MainID, ErrAlreadyUnlocked)
		}

		return MainUnlock{}, fmt.Errorf("main unlock 記録 user=%s main=%s: %w", input.UserID, input.MainID, err)
	}

	mainUnlock, err := mapMainUnlock(row)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("main unlock 記録結果の変換 user=%s main=%s: %w", input.UserID, input.MainID, err)
	}

	return mainUnlock, nil
}

// GetMainUnlock は user ID と main ID から main unlock を取得します。
func (r *Repository) GetMainUnlock(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) (MainUnlock, error) {
	row, err := r.queries.GetMainUnlockByUserIDAndMainID(ctx, sqlc.GetMainUnlockByUserIDAndMainIDParams{
		UserID: postgres.UUIDToPG(userID),
		MainID: postgres.UUIDToPG(mainID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MainUnlock{}, fmt.Errorf("main unlock 取得 user=%s main=%s: %w", userID, mainID, ErrMainUnlockNotFound)
		}

		return MainUnlock{}, fmt.Errorf("main unlock 取得 user=%s main=%s: %w", userID, mainID, err)
	}

	mainUnlock, err := mapMainUnlock(row)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("main unlock 取得結果の変換 user=%s main=%s: %w", userID, mainID, err)
	}

	return mainUnlock, nil
}

// ListUnlockedMainIDs は user の unlock 済み main ID 一覧を返します。
func (r *Repository) ListUnlockedMainIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.queries.ListUnlockedMainIDsByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		return nil, fmt.Errorf("unlock 済み main 一覧取得 user=%s: %w", userID, err)
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		id, err := postgres.UUIDFromPG(row)
		if err != nil {
			return nil, fmt.Errorf("unlock 済み main 一覧取得結果の変換 user=%s: %w", userID, err)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func mapMainUnlock(row sqlc.AppMainUnlock) (MainUnlock, error) {
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("main unlock の user id 変換: %w", err)
	}
	mainID, err := postgres.UUIDFromPG(row.MainID)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("main unlock の main id 変換: %w", err)
	}
	purchasedAt, err := postgres.RequiredTimeFromPG(row.PurchasedAt)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("main unlock の purchased_at 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return MainUnlock{}, fmt.Errorf("main unlock の created_at 変換: %w", err)
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
