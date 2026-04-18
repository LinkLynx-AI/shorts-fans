package payment

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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// ProviderCCBill は CCBill provider を表します。
	ProviderCCBill = "ccbill"

	// PaymentMethodModeSavedCard は保存済み card を使う purchase mode です。
	PaymentMethodModeSavedCard = "saved_card"
	// PaymentMethodModeNewCard は新規 card token を使う purchase mode です。
	PaymentMethodModeNewCard = "new_card"

	// PurchaseAttemptStatusProcessing は provider request 実行中です。
	PurchaseAttemptStatusProcessing = "processing"
	// PurchaseAttemptStatusSucceeded は purchase success です。
	PurchaseAttemptStatusSucceeded = "succeeded"
	// PurchaseAttemptStatusPending は provider completion 待ちです。
	PurchaseAttemptStatusPending = "pending"
	// PurchaseAttemptStatusFailed は purchase failure です。
	PurchaseAttemptStatusFailed = "failed"

	// FailureReasonCardBrandUnsupported は card brand 非対応を表します。
	FailureReasonCardBrandUnsupported = "card_brand_unsupported"
	// FailureReasonPurchaseDeclined は一般的な decline を表します。
	FailureReasonPurchaseDeclined = "purchase_declined"
	// FailureReasonAuthenticationFailed は 3DS / SCA failure を表します。
	FailureReasonAuthenticationFailed = "authentication_failed"

	// PendingReasonProviderProcessing は provider 処理継続中を表します。
	PendingReasonProviderProcessing = "provider_processing"

	// CardBrandVisa は Visa を表します。
	CardBrandVisa = "visa"
	// CardBrandMastercard は Mastercard を表します。
	CardBrandMastercard = "mastercard"
	// CardBrandJCB は JCB を表します。
	CardBrandJCB = "jcb"
	// CardBrandAmericanExpress は American Express を表します。
	CardBrandAmericanExpress = "american_express"

	mainPurchaseAttemptIdempotencyUniqueConstraint = "main_purchase_attempts_idempotency_key_key"
	mainPurchaseAttemptInflightUniqueConstraint    = "idx_main_purchase_attempts_user_main_inflight"
)

var (
	// ErrSavedPaymentMethodNotFound は saved payment method が見つからないことを表します。
	ErrSavedPaymentMethodNotFound = errors.New("saved payment method が見つかりません")
	// ErrMainPurchaseAttemptNotFound は purchase attempt が見つからないことを表します。
	ErrMainPurchaseAttemptNotFound = errors.New("main purchase attempt が見つかりません")
	// ErrMainPurchaseAttemptConflict は inflight/idempotency 境界により既存 attempt と衝突したことを表します。
	ErrMainPurchaseAttemptConflict = errors.New("main purchase attempt が既存 request と衝突しました")
)

type queries interface {
	AcquireMainPurchaseLock(ctx context.Context, arg sqlc.AcquireMainPurchaseLockParams) error
	CreateMainPurchaseAttempt(ctx context.Context, arg sqlc.CreateMainPurchaseAttemptParams) (sqlc.AppMainPurchaseAttempt, error)
	GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdate(ctx context.Context, arg sqlc.GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error)
	GetLatestSucceededMainPurchaseAttemptByUserIDAndMainIDForUpdate(ctx context.Context, arg sqlc.GetLatestSucceededMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error)
	GetMainPurchaseAttemptByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error)
	GetMainPurchaseAttemptByIDForUpdate(ctx context.Context, id pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error)
	GetMainPurchaseAttemptByIdempotencyKeyForUpdate(ctx context.Context, idempotencyKey string) (sqlc.AppMainPurchaseAttempt, error)
	GetMainPurchaseAttemptByProviderPurchaseRefForUpdate(ctx context.Context, providerPurchaseRef pgtype.Text) (sqlc.AppMainPurchaseAttempt, error)
	GetMainPurchaseAttemptByProviderTransactionRefForUpdate(ctx context.Context, providerTransactionRef pgtype.Text) (sqlc.AppMainPurchaseAttempt, error)
	GetUserPaymentMethodByIDAndUserID(ctx context.Context, arg sqlc.GetUserPaymentMethodByIDAndUserIDParams) (sqlc.AppUserPaymentMethod, error)
	ListUserPaymentMethodsByUserID(ctx context.Context, userID pgtype.UUID) ([]sqlc.AppUserPaymentMethod, error)
	TouchUserPaymentMethodLastUsedAt(ctx context.Context, arg sqlc.TouchUserPaymentMethodLastUsedAtParams) (sqlc.AppUserPaymentMethod, error)
	UpdateMainPurchaseAttemptOutcome(ctx context.Context, arg sqlc.UpdateMainPurchaseAttemptOutcomeParams) (sqlc.AppMainPurchaseAttempt, error)
	UpsertUserPaymentMethod(ctx context.Context, arg sqlc.UpsertUserPaymentMethodParams) (sqlc.AppUserPaymentMethod, error)
}

// TxRepository は purchase 処理が transaction scope で使う payment repository 契約です。
type TxRepository interface {
	AcquireMainPurchaseLock(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) error
	CreateMainPurchaseAttempt(ctx context.Context, input CreateMainPurchaseAttemptInput) (MainPurchaseAttempt, error)
	GetLatestInflightMainPurchaseAttemptForUpdate(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) (MainPurchaseAttempt, error)
	GetLatestSucceededMainPurchaseAttemptForUpdate(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) (MainPurchaseAttempt, error)
	GetMainPurchaseAttemptByIdempotencyKeyForUpdate(ctx context.Context, idempotencyKey string) (MainPurchaseAttempt, error)
	GetSavedPaymentMethod(ctx context.Context, userID uuid.UUID, paymentMethodID string) (SavedPaymentMethod, error)
	ListSavedPaymentMethods(ctx context.Context, userID uuid.UUID) ([]SavedPaymentMethod, error)
	TouchSavedPaymentMethodLastUsedAt(ctx context.Context, userID uuid.UUID, paymentMethodID string, lastUsedAt *time.Time) (SavedPaymentMethod, error)
	UpdateMainPurchaseAttemptOutcome(ctx context.Context, input UpdateMainPurchaseAttemptOutcomeInput) (MainPurchaseAttempt, error)
}

// Repository は payment persistence を包みます。
type Repository struct {
	txBeginner postgres.TxBeginner
	queries    queries
	newQueries func(sqlc.DBTX) queries
}

// SavedPaymentMethod は viewer が再利用できる saved card summary です。
type SavedPaymentMethod struct {
	Brand                     string
	CreatedAt                 time.Time
	ID                        uuid.UUID
	Last4                     string
	LastUsedAt                time.Time
	PaymentMethodID           string
	Provider                  string
	ProviderPaymentAccountRef string
	ProviderPaymentTokenRef   string
	UpdatedAt                 time.Time
	UserID                    uuid.UUID
}

// MainPurchaseAttempt は main purchase attempt の永続化表現です。
type MainPurchaseAttempt struct {
	AcceptedAge              bool
	AcceptedTerms            bool
	CreatedAt                time.Time
	FailureReason            *string
	FromShortID              uuid.UUID
	ID                       uuid.UUID
	IdempotencyKey           string
	MainID                   uuid.UUID
	PaymentMethodMode        string
	PendingReason            *string
	Provider                 string
	ProviderDeclineCode      *int32
	ProviderDeclineText      *string
	ProviderPaymentTokenRef  string
	ProviderPaymentUniqueRef *string
	ProviderProcessedAt      *time.Time
	ProviderPurchaseRef      *string
	ProviderSessionRef       *string
	ProviderTransactionRef   *string
	RequestedCurrencyCode    int32
	RequestedPriceJPY        int64
	Status                   string
	UpdatedAt                time.Time
	UserID                   uuid.UUID
	UserPaymentMethodID      *uuid.UUID
}

// CreateMainPurchaseAttemptInput は purchase attempt 新規作成入力です。
type CreateMainPurchaseAttemptInput struct {
	AcceptedAge             bool
	AcceptedTerms           bool
	FromShortID             uuid.UUID
	IdempotencyKey          string
	MainID                  uuid.UUID
	PaymentMethodMode       string
	Provider                string
	ProviderPaymentTokenRef string
	RequestedCurrencyCode   int32
	RequestedPriceJPY       int64
	Status                  string
	UserID                  uuid.UUID
	UserPaymentMethodID     *uuid.UUID
}

// UpdateMainPurchaseAttemptOutcomeInput は purchase attempt outcome 更新入力です。
type UpdateMainPurchaseAttemptOutcomeInput struct {
	FailureReason            *string
	ID                       uuid.UUID
	PendingReason            *string
	ProviderDeclineCode      *int32
	ProviderDeclineText      *string
	ProviderPaymentTokenRef  *string
	ProviderPaymentUniqueRef *string
	ProviderProcessedAt      *time.Time
	ProviderPurchaseRef      *string
	ProviderSessionRef       *string
	ProviderTransactionRef   *string
	Status                   string
}

// UpsertSavedPaymentMethodInput は saved card upsert 入力です。
type UpsertSavedPaymentMethodInput struct {
	Brand                     string
	Last4                     string
	LastUsedAt                *time.Time
	Provider                  string
	ProviderPaymentAccountRef string
	ProviderPaymentTokenRef   string
	UserID                    uuid.UUID
}

// NewRepository は pgxpool ベースの payment repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}

	return &Repository{
		txBeginner: pool,
		queries:    sqlc.New(pool),
		newQueries: func(db sqlc.DBTX) queries {
			return sqlc.New(db)
		},
	}
}

func newRepository(q queries) *Repository {
	return &Repository{
		queries: q,
		newQueries: func(db sqlc.DBTX) queries {
			return sqlc.New(db)
		},
	}
}

func (r *Repository) withQueries(q queries) *Repository {
	return &Repository{
		txBeginner: r.txBeginner,
		queries:    q,
		newQueries: r.newQueries,
	}
}

// RunInTx は tx-scoped repository を callback に渡して実行します。
func (r *Repository) RunInTx(ctx context.Context, fn func(TxRepository) error) error {
	if r == nil || r.txBeginner == nil || r.newQueries == nil {
		return fmt.Errorf("payment repository が初期化されていません")
	}
	if fn == nil {
		return fmt.Errorf("payment transaction callback が nil です")
	}

	return postgres.RunInTx(ctx, r.txBeginner, func(tx pgx.Tx) error {
		return fn(r.withQueries(r.newQueries(tx)))
	})
}

// AcquireMainPurchaseLock は viewer/main 単位の purchase 処理を transaction scope で直列化します。
func (r *Repository) AcquireMainPurchaseLock(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) error {
	if r == nil || r.queries == nil {
		return fmt.Errorf("payment repository が初期化されていません")
	}

	if err := r.queries.AcquireMainPurchaseLock(ctx, sqlc.AcquireMainPurchaseLockParams{
		UserKey: userID.String(),
		MainKey: mainID.String(),
	}); err != nil {
		return fmt.Errorf("main purchase lock 取得 user=%s main=%s: %w", userID, mainID, err)
	}

	return nil
}

// ListSavedPaymentMethods は viewer の saved card 一覧を返します。
func (r *Repository) ListSavedPaymentMethods(ctx context.Context, userID uuid.UUID) ([]SavedPaymentMethod, error) {
	if r == nil || r.queries == nil {
		return nil, fmt.Errorf("payment repository が初期化されていません")
	}

	rows, err := r.queries.ListUserPaymentMethodsByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		return nil, fmt.Errorf("saved payment methods 取得 user=%s: %w", userID, err)
	}

	items := make([]SavedPaymentMethod, 0, len(rows))
	for _, row := range rows {
		item, err := mapSavedPaymentMethod(row)
		if err != nil {
			return nil, fmt.Errorf("saved payment methods 取得結果の変換 user=%s: %w", userID, err)
		}

		items = append(items, item)
	}

	return items, nil
}

// GetSavedPaymentMethod は public payment method id から 1 件取得します。
func (r *Repository) GetSavedPaymentMethod(ctx context.Context, userID uuid.UUID, paymentMethodID string) (SavedPaymentMethod, error) {
	if r == nil || r.queries == nil {
		return SavedPaymentMethod{}, fmt.Errorf("payment repository が初期化されていません")
	}

	id, err := ParsePublicPaymentMethodID(paymentMethodID)
	if err != nil {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method id 解析 user=%s: %w", userID, ErrSavedPaymentMethodNotFound)
	}

	row, err := r.queries.GetUserPaymentMethodByIDAndUserID(ctx, sqlc.GetUserPaymentMethodByIDAndUserIDParams{
		ID:     postgres.UUIDToPG(id),
		UserID: postgres.UUIDToPG(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SavedPaymentMethod{}, fmt.Errorf("saved payment method 取得 user=%s payment_method=%s: %w", userID, paymentMethodID, ErrSavedPaymentMethodNotFound)
		}

		return SavedPaymentMethod{}, fmt.Errorf("saved payment method 取得 user=%s payment_method=%s: %w", userID, paymentMethodID, err)
	}

	method, err := mapSavedPaymentMethod(row)
	if err != nil {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method 取得結果の変換 user=%s payment_method=%s: %w", userID, paymentMethodID, err)
	}

	return method, nil
}

// UpsertSavedPaymentMethod は provider account ref を軸に saved card を upsert します。
func (r *Repository) UpsertSavedPaymentMethod(ctx context.Context, input UpsertSavedPaymentMethodInput) (SavedPaymentMethod, error) {
	if r == nil || r.queries == nil {
		return SavedPaymentMethod{}, fmt.Errorf("payment repository が初期化されていません")
	}

	row, err := r.queries.UpsertUserPaymentMethod(ctx, sqlc.UpsertUserPaymentMethodParams{
		UserID:                    postgres.UUIDToPG(input.UserID),
		Provider:                  input.Provider,
		ProviderPaymentTokenRef:   strings.TrimSpace(input.ProviderPaymentTokenRef),
		ProviderPaymentAccountRef: strings.TrimSpace(input.ProviderPaymentAccountRef),
		Brand:                     strings.TrimSpace(input.Brand),
		Last4:                     strings.TrimSpace(input.Last4),
		LastUsedAt:                postgres.TimeToPG(input.LastUsedAt),
	})
	if err != nil {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method upsert user=%s provider=%s: %w", input.UserID, input.Provider, err)
	}

	method, err := mapSavedPaymentMethod(row)
	if err != nil {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method upsert 結果の変換 user=%s provider=%s: %w", input.UserID, input.Provider, err)
	}

	return method, nil
}

// TouchSavedPaymentMethodLastUsedAt は saved card の last_used_at を更新します。
func (r *Repository) TouchSavedPaymentMethodLastUsedAt(ctx context.Context, userID uuid.UUID, paymentMethodID string, lastUsedAt *time.Time) (SavedPaymentMethod, error) {
	if r == nil || r.queries == nil {
		return SavedPaymentMethod{}, fmt.Errorf("payment repository が初期化されていません")
	}

	id, err := ParsePublicPaymentMethodID(paymentMethodID)
	if err != nil {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method last_used_at 更新 user=%s: %w", userID, ErrSavedPaymentMethodNotFound)
	}

	row, err := r.queries.TouchUserPaymentMethodLastUsedAt(ctx, sqlc.TouchUserPaymentMethodLastUsedAtParams{
		LastUsedAt: postgres.TimeToPG(lastUsedAt),
		ID:         postgres.UUIDToPG(id),
		UserID:     postgres.UUIDToPG(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SavedPaymentMethod{}, fmt.Errorf("saved payment method last_used_at 更新 user=%s payment_method=%s: %w", userID, paymentMethodID, ErrSavedPaymentMethodNotFound)
		}

		return SavedPaymentMethod{}, fmt.Errorf("saved payment method last_used_at 更新 user=%s payment_method=%s: %w", userID, paymentMethodID, err)
	}

	method, err := mapSavedPaymentMethod(row)
	if err != nil {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method last_used_at 更新結果の変換 user=%s payment_method=%s: %w", userID, paymentMethodID, err)
	}

	return method, nil
}

// CreateMainPurchaseAttempt は purchase attempt を新規作成します。
func (r *Repository) CreateMainPurchaseAttempt(ctx context.Context, input CreateMainPurchaseAttemptInput) (MainPurchaseAttempt, error) {
	if r == nil || r.queries == nil {
		return MainPurchaseAttempt{}, fmt.Errorf("payment repository が初期化されていません")
	}

	row, err := r.queries.CreateMainPurchaseAttempt(ctx, sqlc.CreateMainPurchaseAttemptParams{
		UserID:                  postgres.UUIDToPG(input.UserID),
		MainID:                  postgres.UUIDToPG(input.MainID),
		FromShortID:             postgres.UUIDToPG(input.FromShortID),
		Provider:                strings.TrimSpace(input.Provider),
		PaymentMethodMode:       strings.TrimSpace(input.PaymentMethodMode),
		UserPaymentMethodID:     optionalUUIDToPG(input.UserPaymentMethodID),
		ProviderPaymentTokenRef: strings.TrimSpace(input.ProviderPaymentTokenRef),
		IdempotencyKey:          strings.TrimSpace(input.IdempotencyKey),
		Status:                  strings.TrimSpace(input.Status),
		RequestedPriceJpy:       input.RequestedPriceJPY,
		RequestedCurrencyCode:   input.RequestedCurrencyCode,
		AcceptedAge:             input.AcceptedAge,
		AcceptedTerms:           input.AcceptedTerms,
	})
	if err != nil {
		if isMainPurchaseAttemptConflictError(err) {
			return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt 作成 user=%s main=%s: %w", input.UserID, input.MainID, ErrMainPurchaseAttemptConflict)
		}

		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt 作成 user=%s main=%s: %w", input.UserID, input.MainID, err)
	}

	attempt, err := mapMainPurchaseAttempt(row)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt 作成結果の変換 user=%s main=%s: %w", input.UserID, input.MainID, err)
	}

	return attempt, nil
}

// GetMainPurchaseAttempt は ID から purchase attempt を取得します。
func (r *Repository) GetMainPurchaseAttempt(ctx context.Context, id uuid.UUID) (MainPurchaseAttempt, error) {
	if r == nil || r.queries == nil {
		return MainPurchaseAttempt{}, fmt.Errorf("payment repository が初期化されていません")
	}

	row, err := r.queries.GetMainPurchaseAttemptByID(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt 取得 id=%s: %w", id, ErrMainPurchaseAttemptNotFound)
		}

		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt 取得 id=%s: %w", id, err)
	}

	attempt, err := mapMainPurchaseAttempt(row)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt 取得結果の変換 id=%s: %w", id, err)
	}

	return attempt, nil
}

// GetMainPurchaseAttemptForUpdate は ID から purchase attempt を lock して取得します。
func (r *Repository) GetMainPurchaseAttemptForUpdate(ctx context.Context, id uuid.UUID) (MainPurchaseAttempt, error) {
	if r == nil || r.queries == nil {
		return MainPurchaseAttempt{}, fmt.Errorf("payment repository が初期化されていません")
	}

	row, err := r.queries.GetMainPurchaseAttemptByIDForUpdate(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt lock 取得 id=%s: %w", id, ErrMainPurchaseAttemptNotFound)
		}

		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt lock 取得 id=%s: %w", id, err)
	}

	attempt, err := mapMainPurchaseAttempt(row)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt lock 取得結果の変換 id=%s: %w", id, err)
	}

	return attempt, nil
}

// GetMainPurchaseAttemptByIdempotencyKeyForUpdate は idempotency key で purchase attempt を lock して取得します。
func (r *Repository) GetMainPurchaseAttemptByIdempotencyKeyForUpdate(ctx context.Context, idempotencyKey string) (MainPurchaseAttempt, error) {
	if r == nil || r.queries == nil {
		return MainPurchaseAttempt{}, fmt.Errorf("payment repository が初期化されていません")
	}

	row, err := r.queries.GetMainPurchaseAttemptByIdempotencyKeyForUpdate(ctx, strings.TrimSpace(idempotencyKey))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt idempotency 取得 key=%s: %w", idempotencyKey, ErrMainPurchaseAttemptNotFound)
		}

		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt idempotency 取得 key=%s: %w", idempotencyKey, err)
	}

	attempt, err := mapMainPurchaseAttempt(row)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt idempotency 取得結果の変換 key=%s: %w", idempotencyKey, err)
	}

	return attempt, nil
}

// GetLatestInflightMainPurchaseAttemptForUpdate は viewer/main の inflight attempt を lock して取得します。
func (r *Repository) GetLatestInflightMainPurchaseAttemptForUpdate(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) (MainPurchaseAttempt, error) {
	if r == nil || r.queries == nil {
		return MainPurchaseAttempt{}, fmt.Errorf("payment repository が初期化されていません")
	}

	row, err := r.queries.GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdate(ctx, sqlc.GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdateParams{
		UserID: postgres.UUIDToPG(userID),
		MainID: postgres.UUIDToPG(mainID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MainPurchaseAttempt{}, fmt.Errorf("main purchase inflight 取得 user=%s main=%s: %w", userID, mainID, ErrMainPurchaseAttemptNotFound)
		}

		return MainPurchaseAttempt{}, fmt.Errorf("main purchase inflight 取得 user=%s main=%s: %w", userID, mainID, err)
	}

	attempt, err := mapMainPurchaseAttempt(row)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase inflight 取得結果の変換 user=%s main=%s: %w", userID, mainID, err)
	}

	return attempt, nil
}

// GetLatestSucceededMainPurchaseAttemptForUpdate は viewer/main の最新 successful attempt を lock して取得します。
func (r *Repository) GetLatestSucceededMainPurchaseAttemptForUpdate(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) (MainPurchaseAttempt, error) {
	if r == nil || r.queries == nil {
		return MainPurchaseAttempt{}, fmt.Errorf("payment repository が初期化されていません")
	}

	row, err := r.queries.GetLatestSucceededMainPurchaseAttemptByUserIDAndMainIDForUpdate(ctx, sqlc.GetLatestSucceededMainPurchaseAttemptByUserIDAndMainIDForUpdateParams{
		UserID: postgres.UUIDToPG(userID),
		MainID: postgres.UUIDToPG(mainID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MainPurchaseAttempt{}, fmt.Errorf("main purchase success 取得 user=%s main=%s: %w", userID, mainID, ErrMainPurchaseAttemptNotFound)
		}

		return MainPurchaseAttempt{}, fmt.Errorf("main purchase success 取得 user=%s main=%s: %w", userID, mainID, err)
	}

	attempt, err := mapMainPurchaseAttempt(row)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase success 取得結果の変換 user=%s main=%s: %w", userID, mainID, err)
	}

	return attempt, nil
}

// GetMainPurchaseAttemptByProviderPurchaseRefForUpdate は provider purchase ref で attempt を lock して取得します。
func (r *Repository) GetMainPurchaseAttemptByProviderPurchaseRefForUpdate(ctx context.Context, providerPurchaseRef string) (MainPurchaseAttempt, error) {
	if r == nil || r.queries == nil {
		return MainPurchaseAttempt{}, fmt.Errorf("payment repository が初期化されていません")
	}

	row, err := r.queries.GetMainPurchaseAttemptByProviderPurchaseRefForUpdate(ctx, postgres.TextToPG(stringPtr(strings.TrimSpace(providerPurchaseRef))))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MainPurchaseAttempt{}, fmt.Errorf("main purchase provider ref 取得 ref=%s: %w", providerPurchaseRef, ErrMainPurchaseAttemptNotFound)
		}

		return MainPurchaseAttempt{}, fmt.Errorf("main purchase provider ref 取得 ref=%s: %w", providerPurchaseRef, err)
	}

	attempt, err := mapMainPurchaseAttempt(row)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase provider ref 取得結果の変換 ref=%s: %w", providerPurchaseRef, err)
	}

	return attempt, nil
}

// GetMainPurchaseAttemptByProviderTransactionRefForUpdate は provider transaction ref で attempt を lock して取得します。
func (r *Repository) GetMainPurchaseAttemptByProviderTransactionRefForUpdate(ctx context.Context, providerTransactionRef string) (MainPurchaseAttempt, error) {
	if r == nil || r.queries == nil {
		return MainPurchaseAttempt{}, fmt.Errorf("payment repository が初期化されていません")
	}

	row, err := r.queries.GetMainPurchaseAttemptByProviderTransactionRefForUpdate(ctx, postgres.TextToPG(stringPtr(strings.TrimSpace(providerTransactionRef))))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MainPurchaseAttempt{}, fmt.Errorf("main purchase provider transaction ref 取得 ref=%s: %w", providerTransactionRef, ErrMainPurchaseAttemptNotFound)
		}

		return MainPurchaseAttempt{}, fmt.Errorf("main purchase provider transaction ref 取得 ref=%s: %w", providerTransactionRef, err)
	}

	attempt, err := mapMainPurchaseAttempt(row)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase provider transaction ref 取得結果の変換 ref=%s: %w", providerTransactionRef, err)
	}

	return attempt, nil
}

// UpdateMainPurchaseAttemptOutcome は purchase attempt の status/outcome を更新します。
func (r *Repository) UpdateMainPurchaseAttemptOutcome(ctx context.Context, input UpdateMainPurchaseAttemptOutcomeInput) (MainPurchaseAttempt, error) {
	if r == nil || r.queries == nil {
		return MainPurchaseAttempt{}, fmt.Errorf("payment repository が初期化されていません")
	}

	row, err := r.queries.UpdateMainPurchaseAttemptOutcome(ctx, sqlc.UpdateMainPurchaseAttemptOutcomeParams{
		Status:                   strings.TrimSpace(input.Status),
		FailureReason:            postgres.TextToPG(input.FailureReason),
		PendingReason:            postgres.TextToPG(input.PendingReason),
		ProviderPaymentTokenRef:  postgres.TextToPG(input.ProviderPaymentTokenRef),
		ProviderPurchaseRef:      postgres.TextToPG(input.ProviderPurchaseRef),
		ProviderTransactionRef:   postgres.TextToPG(input.ProviderTransactionRef),
		ProviderSessionRef:       postgres.TextToPG(input.ProviderSessionRef),
		ProviderPaymentUniqueRef: postgres.TextToPG(input.ProviderPaymentUniqueRef),
		ProviderDeclineCode:      optionalInt32ToPG(input.ProviderDeclineCode),
		ProviderDeclineText:      postgres.TextToPG(input.ProviderDeclineText),
		ProviderProcessedAt:      postgres.TimeToPG(input.ProviderProcessedAt),
		ID:                       postgres.UUIDToPG(input.ID),
	})
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase outcome 更新 id=%s: %w", input.ID, err)
	}

	attempt, err := mapMainPurchaseAttempt(row)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase outcome 更新結果の変換 id=%s: %w", input.ID, err)
	}

	return attempt, nil
}

func mapSavedPaymentMethod(row sqlc.AppUserPaymentMethod) (SavedPaymentMethod, error) {
	id, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method id 変換: %w", err)
	}
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method user id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method created_at 変換: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method updated_at 変換: %w", err)
	}
	lastUsedAt, err := postgres.RequiredTimeFromPG(row.LastUsedAt)
	if err != nil {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method last_used_at 変換: %w", err)
	}

	if strings.TrimSpace(row.Brand) == "" {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method brand がありません")
	}
	if strings.TrimSpace(row.Last4) == "" {
		return SavedPaymentMethod{}, fmt.Errorf("saved payment method last4 がありません")
	}

	return SavedPaymentMethod{
		Brand:                     strings.TrimSpace(row.Brand),
		CreatedAt:                 createdAt,
		ID:                        id,
		Last4:                     strings.TrimSpace(row.Last4),
		LastUsedAt:                lastUsedAt,
		PaymentMethodID:           FormatPublicPaymentMethodID(id),
		Provider:                  strings.TrimSpace(row.Provider),
		ProviderPaymentAccountRef: strings.TrimSpace(row.ProviderPaymentAccountRef),
		ProviderPaymentTokenRef:   strings.TrimSpace(row.ProviderPaymentTokenRef),
		UpdatedAt:                 updatedAt,
		UserID:                    userID,
	}, nil
}

func mapMainPurchaseAttempt(row sqlc.AppMainPurchaseAttempt) (MainPurchaseAttempt, error) {
	id, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt id 変換: %w", err)
	}
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt user id 変換: %w", err)
	}
	mainID, err := postgres.UUIDFromPG(row.MainID)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt main id 変換: %w", err)
	}
	fromShortID, err := postgres.UUIDFromPG(row.FromShortID)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt from short id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt created_at 変換: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt updated_at 変換: %w", err)
	}

	userPaymentMethodID, err := optionalUUIDFromPG(row.UserPaymentMethodID)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt user payment method id 変換: %w", err)
	}
	providerProcessedAt, err := optionalTimeFromPG(row.ProviderProcessedAt)
	if err != nil {
		return MainPurchaseAttempt{}, fmt.Errorf("main purchase attempt provider processed at 変換: %w", err)
	}

	return MainPurchaseAttempt{
		AcceptedAge:              row.AcceptedAge,
		AcceptedTerms:            row.AcceptedTerms,
		CreatedAt:                createdAt,
		FailureReason:            postgres.OptionalTextFromPG(row.FailureReason),
		FromShortID:              fromShortID,
		ID:                       id,
		IdempotencyKey:           strings.TrimSpace(row.IdempotencyKey),
		MainID:                   mainID,
		PaymentMethodMode:        strings.TrimSpace(row.PaymentMethodMode),
		PendingReason:            postgres.OptionalTextFromPG(row.PendingReason),
		Provider:                 strings.TrimSpace(row.Provider),
		ProviderDeclineCode:      optionalInt32FromPG(row.ProviderDeclineCode),
		ProviderDeclineText:      postgres.OptionalTextFromPG(row.ProviderDeclineText),
		ProviderPaymentTokenRef:  strings.TrimSpace(row.ProviderPaymentTokenRef),
		ProviderPaymentUniqueRef: postgres.OptionalTextFromPG(row.ProviderPaymentUniqueRef),
		ProviderProcessedAt:      providerProcessedAt,
		ProviderPurchaseRef:      postgres.OptionalTextFromPG(row.ProviderPurchaseRef),
		ProviderSessionRef:       postgres.OptionalTextFromPG(row.ProviderSessionRef),
		ProviderTransactionRef:   postgres.OptionalTextFromPG(row.ProviderTransactionRef),
		RequestedCurrencyCode:    row.RequestedCurrencyCode,
		RequestedPriceJPY:        row.RequestedPriceJpy,
		Status:                   strings.TrimSpace(row.Status),
		UpdatedAt:                updatedAt,
		UserID:                   userID,
		UserPaymentMethodID:      userPaymentMethodID,
	}, nil
}

// FormatPublicPaymentMethodID は saved payment method の public id を返します。
func FormatPublicPaymentMethodID(id uuid.UUID) string {
	return "pm_" + strings.ReplaceAll(id.String(), "-", "")
}

// ParsePublicPaymentMethodID は public payment method id を UUID へ戻します。
func ParsePublicPaymentMethodID(value string) (uuid.UUID, error) {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "pm_") {
		return uuid.Nil, fmt.Errorf("payment method id prefix が不正です")
	}

	raw := strings.TrimPrefix(trimmed, "pm_")
	if len(raw) != 32 {
		return uuid.Nil, fmt.Errorf("payment method id length が不正です")
	}

	normalized := fmt.Sprintf("%s-%s-%s-%s-%s", raw[0:8], raw[8:12], raw[12:16], raw[16:20], raw[20:32])
	id, err := uuid.Parse(normalized)
	if err != nil {
		return uuid.Nil, fmt.Errorf("payment method id UUID 変換: %w", err)
	}

	return id, nil
}

func optionalUUIDToPG(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}

	return postgres.UUIDToPG(*value)
}

func optionalUUIDFromPG(value pgtype.UUID) (*uuid.UUID, error) {
	if !value.Valid {
		return nil, nil
	}

	id, err := postgres.UUIDFromPG(value)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func optionalTimeFromPG(value pgtype.Timestamptz) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}

	timestamp, err := postgres.RequiredTimeFromPG(value)
	if err != nil {
		return nil, err
	}

	return &timestamp, nil
}

func optionalInt32ToPG(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}

	return pgtype.Int4{Int32: *value, Valid: true}
}

func optionalInt32FromPG(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}

	v := value.Int32
	return &v
}

func stringPtr(value string) *string {
	return &value
}

func isMainPurchaseAttemptConflictError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == "23505" && (pgErr.ConstraintName == mainPurchaseAttemptIdempotencyUniqueConstraint || pgErr.ConstraintName == mainPurchaseAttemptInflightUniqueConstraint)
}
