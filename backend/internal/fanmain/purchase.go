package fanmain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/payment"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/unlock"
	"github.com/google/uuid"
)

var supportedCardBrands = []string{
	payment.CardBrandVisa,
	payment.CardBrandMastercard,
	payment.CardBrandJCB,
	payment.CardBrandAmericanExpress,
}

// ErrInvalidPurchaseRequest は purchase request が contract を満たさないことを表します。
var ErrInvalidPurchaseRequest = errors.New("purchase request が不正です")

// ErrPurchaseNotFound は purchase 対象の short/main が見つからないことを表します。
var ErrPurchaseNotFound = errors.New("main or short が見つかりません")

type paymentRepository interface {
	CreateMainPurchaseAttempt(ctx context.Context, input payment.CreateMainPurchaseAttemptInput) (payment.MainPurchaseAttempt, error)
	GetLatestInflightMainPurchaseAttemptForUpdate(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) (payment.MainPurchaseAttempt, error)
	GetLatestSucceededMainPurchaseAttemptForUpdate(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) (payment.MainPurchaseAttempt, error)
	GetMainPurchaseAttemptByIdempotencyKeyForUpdate(ctx context.Context, idempotencyKey string) (payment.MainPurchaseAttempt, error)
	GetSavedPaymentMethod(ctx context.Context, userID uuid.UUID, paymentMethodID string) (payment.SavedPaymentMethod, error)
	ListSavedPaymentMethods(ctx context.Context, userID uuid.UUID) ([]payment.SavedPaymentMethod, error)
	TouchSavedPaymentMethodLastUsedAt(ctx context.Context, userID uuid.UUID, paymentMethodID string, lastUsedAt *time.Time) (payment.SavedPaymentMethod, error)
	UpdateMainPurchaseAttemptOutcome(ctx context.Context, input payment.UpdateMainPurchaseAttemptOutcomeInput) (payment.MainPurchaseAttempt, error)
}

type transactionalPaymentRepository interface {
	paymentRepository
	RunInTx(ctx context.Context, fn func(payment.TxRepository) error) error
}

type purchaseGateway interface {
	Charge(ctx context.Context, input payment.ChargeInput) (payment.ChargeResult, error)
}

// SavedPaymentMethodSummary は unlock surface に返す saved card summary です。
type SavedPaymentMethodSummary struct {
	Brand           string
	Last4           string
	PaymentMethodID string
}

// PurchaseSetupState は purchase 前 setup 状態です。
type PurchaseSetupState struct {
	Required                bool
	RequiresAgeConfirmation bool
	RequiresCardSetup       bool
	RequiresTermsAcceptance bool
}

// UnlockPurchaseState は paywall 上の purchase state です。
type UnlockPurchaseState struct {
	PendingReason       *string
	SavedPaymentMethods []SavedPaymentMethodSummary
	Setup               PurchaseSetupState
	State               string
	SupportedCardBrands []string
}

// PurchasePaymentMethodInput は purchase に使う payment method 入力です。
type PurchasePaymentMethodInput struct {
	CardSetupToken  string
	Mode            string
	PaymentMethodID string
}

// PurchaseInput は main purchase 実行入力です。
type PurchaseInput struct {
	AcceptedAge   bool
	AcceptedTerms bool
	EntryToken    string
	FromShortID   uuid.UUID
	MainID        uuid.UUID
	PaymentMethod PurchasePaymentMethodInput
	ViewerID      uuid.UUID
	ViewerIP      string
}

// PurchaseOutcome は public purchase status を表します。
type PurchaseOutcome struct {
	CanRetry      bool
	FailureReason *string
	Status        string
}

// PurchaseResult は purchase endpoint 用の結果です。
type PurchaseResult struct {
	Access     MainAccessState
	EntryToken *string
	Purchase   PurchaseOutcome
}

// PurchaseMain は provider purchase を実行し、success 時だけ durable unlock を記録します。
func (s *Service) PurchaseMain(ctx context.Context, sessionBinding string, input PurchaseInput) (PurchaseResult, error) {
	detail, main, err := s.loadLinkedSurface(ctx, input.ViewerID, input.FromShortID)
	if err != nil {
		switch {
		case errors.Is(err, feed.ErrPublicShortNotFound):
			return PurchaseResult{}, ErrPurchaseNotFound
		case errors.Is(err, shorts.ErrUnlockableMainNotFound):
			return PurchaseResult{}, ErrMainLocked
		default:
			return PurchaseResult{}, err
		}
	}

	if detail.Item.Short.CanonicalMainID != input.MainID || main.ID != input.MainID {
		return PurchaseResult{}, ErrPurchaseNotFound
	}

	entryToken, err := readSignedToken(sessionBinding, s.now().UTC(), input.EntryToken)
	if err != nil {
		return PurchaseResult{}, ErrMainLocked
	}
	if entryToken.Kind != entryTokenKind ||
		entryToken.MainID != input.MainID ||
		entryToken.FromShortID != input.FromShortID ||
		entryToken.ViewerID != input.ViewerID {
		return PurchaseResult{}, ErrMainLocked
	}

	savedMethods, err := s.listSavedPaymentMethods(ctx, input.ViewerID)
	if err != nil {
		return PurchaseResult{}, err
	}

	switch {
	case detail.Item.Unlock.IsOwner:
		return buildOwnerPurchaseResult(main.ID, input.EntryToken), nil
	case detail.Item.Unlock.IsUnlocked:
		return buildAlreadyPurchasedResult(main.ID, input.EntryToken), nil
	}

	if txRepo, ok := s.paymentRepository.(transactionalPaymentRepository); ok {
		var result PurchaseResult
		if err := txRepo.RunInTx(ctx, func(repo payment.TxRepository) error {
			if err := repo.AcquireMainPurchaseLock(ctx, input.ViewerID, input.MainID); err != nil {
				return err
			}

			txService := *s
			txService.paymentRepository = repo

			txResult, err := txService.purchaseMainWithLockedPaymentState(ctx, sessionBinding, main, savedMethods, input)
			if err != nil {
				return err
			}

			result = txResult
			return nil
		}); err != nil {
			return PurchaseResult{}, err
		}

		return result, nil
	}

	return s.purchaseMainWithLockedPaymentState(ctx, sessionBinding, main, savedMethods, input)
}

func (s *Service) purchaseMainWithLockedPaymentState(
	ctx context.Context,
	sessionBinding string,
	main shorts.Main,
	savedMethods []payment.SavedPaymentMethod,
	input PurchaseInput,
) (PurchaseResult, error) {
	if _, err := s.getInflightAttempt(ctx, input.ViewerID, input.MainID); err == nil {
		return buildPendingPurchaseResult(main.ID), nil
	} else if !errors.Is(err, payment.ErrMainPurchaseAttemptNotFound) {
		return PurchaseResult{}, err
	}
	if completedAttempt, err := s.getLatestSucceededAttempt(ctx, input.ViewerID, input.MainID); err == nil {
		return buildPurchaseResultFromAttempt(main.ID, completedAttempt, input.EntryToken), nil
	} else if !errors.Is(err, payment.ErrMainPurchaseAttemptNotFound) {
		return PurchaseResult{}, err
	}

	if err := validatePurchaseInput(input, savedMethods); err != nil {
		return PurchaseResult{}, err
	}

	requestedCurrencyCode, err := currencyNumericCode(main.CurrencyCode)
	if err != nil {
		return PurchaseResult{}, err
	}

	idempotencyKey := buildPurchaseIdempotencyKey(input)
	if existingAttempt, err := s.paymentRepository.GetMainPurchaseAttemptByIdempotencyKeyForUpdate(ctx, idempotencyKey); err == nil {
		return buildPurchaseResultFromAttempt(main.ID, existingAttempt, input.EntryToken), nil
	} else if !errors.Is(err, payment.ErrMainPurchaseAttemptNotFound) {
		return PurchaseResult{}, fmt.Errorf("purchase attempt idempotency 取得 viewer=%s main=%s: %w", input.ViewerID, input.MainID, err)
	}

	paymentMode := strings.TrimSpace(input.PaymentMethod.Mode)
	providerPaymentTokenRef := ""
	var savedMethod *payment.SavedPaymentMethod
	var userPaymentMethodID *uuid.UUID

	switch paymentMode {
	case payment.PaymentMethodModeSavedCard:
		method, err := s.paymentRepository.GetSavedPaymentMethod(ctx, input.ViewerID, input.PaymentMethod.PaymentMethodID)
		if err != nil {
			if errors.Is(err, payment.ErrSavedPaymentMethodNotFound) {
				return PurchaseResult{}, ErrInvalidPurchaseRequest
			}

			return PurchaseResult{}, fmt.Errorf("saved payment method 取得 viewer=%s: %w", input.ViewerID, err)
		}

		savedMethod = &method
		userPaymentMethodID = &method.ID
		providerPaymentTokenRef = method.ProviderPaymentTokenRef
	case payment.PaymentMethodModeNewCard:
		providerPaymentTokenRef, err = resolveCardSetupPaymentTokenRef(sessionBinding, s.now().UTC(), input.ViewerID, input.PaymentMethod.CardSetupToken)
		if err != nil {
			return PurchaseResult{}, ErrInvalidPurchaseRequest
		}
	default:
		return PurchaseResult{}, ErrInvalidPurchaseRequest
	}

	attempt, err := s.paymentRepository.CreateMainPurchaseAttempt(ctx, payment.CreateMainPurchaseAttemptInput{
		AcceptedAge:             input.AcceptedAge,
		AcceptedTerms:           input.AcceptedTerms,
		FromShortID:             input.FromShortID,
		IdempotencyKey:          idempotencyKey,
		MainID:                  input.MainID,
		PaymentMethodMode:       paymentMode,
		Provider:                payment.ProviderCCBill,
		ProviderPaymentTokenRef: providerPaymentTokenRef,
		RequestedCurrencyCode:   requestedCurrencyCode,
		RequestedPriceJPY:       main.PriceMinor,
		Status:                  payment.PurchaseAttemptStatusProcessing,
		UserID:                  input.ViewerID,
		UserPaymentMethodID:     userPaymentMethodID,
	})
	if err != nil {
		if errors.Is(err, payment.ErrMainPurchaseAttemptConflict) {
			existingAttempt, resolveErr := s.resolveConflictingPurchaseAttempt(ctx, input.ViewerID, input.MainID, idempotencyKey)
			if resolveErr != nil {
				return PurchaseResult{}, resolveErr
			}

			return buildPurchaseResultFromAttempt(main.ID, existingAttempt, input.EntryToken), nil
		}

		return PurchaseResult{}, err
	}

	chargeResult, err := s.purchaseGateway.Charge(ctx, payment.ChargeInput{
		AttemptID:       attempt.ID,
		IPAddress:       input.ViewerIP,
		PaymentTokenRef: providerPaymentTokenRef,
		PriceJPY:        main.PriceMinor,
	})
	if err != nil {
		if errors.Is(err, payment.ErrChargeOutcomeUnknown) {
			return s.markUnknownChargePending(ctx, attempt)
		}

		if updateErr := s.markInternalChargeFailure(ctx, attempt); updateErr != nil {
			return PurchaseResult{}, updateErr
		}

		return PurchaseResult{}, err
	}

	updateInput := payment.UpdateMainPurchaseAttemptOutcomeInput{
		FailureReason:            chargeResult.FailureReason,
		ID:                       attempt.ID,
		PendingReason:            chargeResult.PendingReason,
		ProviderDeclineCode:      chargeResult.ProviderDeclineCode,
		ProviderDeclineText:      chargeResult.ProviderDeclineText,
		ProviderPaymentTokenRef:  chargeResult.NewPaymentTokenRef,
		ProviderPaymentUniqueRef: chargeResult.ProviderPaymentUniqueRef,
		ProviderProcessedAt:      &chargeResult.ProviderProcessedAt,
		ProviderPurchaseRef:      chargeResult.ProviderPurchaseRef,
		ProviderSessionRef:       chargeResult.ProviderSessionRef,
		ProviderTransactionRef:   chargeResult.ProviderTransactionRef,
		Status:                   chargeResult.Status,
	}

	switch chargeResult.Status {
	case payment.PurchaseAttemptStatusSucceeded:
		if err := s.recordPurchaseUnlock(ctx, input.ViewerID, input.MainID, chargeResult.ProviderProcessedAt, chargeResult.ProviderPurchaseRef); err != nil {
			return PurchaseResult{}, err
		}
		if _, err := s.paymentRepository.UpdateMainPurchaseAttemptOutcome(ctx, updateInput); err != nil {
			return PurchaseResult{}, fmt.Errorf("purchase success outcome 更新 attempt=%s: %w", attempt.ID, err)
		}
		if savedMethod != nil {
			if _, err := s.paymentRepository.TouchSavedPaymentMethodLastUsedAt(ctx, input.ViewerID, savedMethod.PaymentMethodID, &chargeResult.ProviderProcessedAt); err != nil {
				return PurchaseResult{}, fmt.Errorf("saved payment method last_used_at 更新 attempt=%s: %w", attempt.ID, err)
			}
		}

		return PurchaseResult{
			Access:     buildPurchasedAccessState(input.MainID),
			EntryToken: stringPtr(input.EntryToken),
			Purchase: PurchaseOutcome{
				CanRetry: false,
				Status:   "succeeded",
			},
		}, nil
	case payment.PurchaseAttemptStatusPending:
		if _, err := s.paymentRepository.UpdateMainPurchaseAttemptOutcome(ctx, updateInput); err != nil {
			return PurchaseResult{}, fmt.Errorf("purchase pending outcome 更新 attempt=%s: %w", attempt.ID, err)
		}

		return buildPendingPurchaseResult(input.MainID), nil
	case payment.PurchaseAttemptStatusFailed:
		if _, err := s.paymentRepository.UpdateMainPurchaseAttemptOutcome(ctx, updateInput); err != nil {
			return PurchaseResult{}, fmt.Errorf("purchase failure outcome 更新 attempt=%s: %w", attempt.ID, err)
		}

		return PurchaseResult{
			Access: buildLockedAccessState(input.MainID),
			Purchase: PurchaseOutcome{
				CanRetry:      chargeResult.CanRetry,
				FailureReason: chargeResult.FailureReason,
				Status:        "failed",
			},
		}, nil
	default:
		return PurchaseResult{}, fmt.Errorf("unsupported charge status %q", chargeResult.Status)
	}
}

func currencyNumericCode(currencyCode string) (int32, error) {
	switch strings.ToUpper(strings.TrimSpace(currencyCode)) {
	case "JPY":
		return 392, nil
	default:
		return 0, fmt.Errorf("unsupported currency code %q", currencyCode)
	}
}

func (s *Service) listSavedPaymentMethods(ctx context.Context, viewerID uuid.UUID) ([]payment.SavedPaymentMethod, error) {
	if s == nil || s.paymentRepository == nil {
		return nil, fmt.Errorf("fan main payment repository が初期化されていません")
	}

	methods, err := s.paymentRepository.ListSavedPaymentMethods(ctx, viewerID)
	if err != nil {
		return nil, fmt.Errorf("saved payment methods 取得 viewer=%s: %w", viewerID, err)
	}

	return methods, nil
}

func (s *Service) getInflightAttempt(ctx context.Context, viewerID uuid.UUID, mainID uuid.UUID) (payment.MainPurchaseAttempt, error) {
	if s == nil || s.paymentRepository == nil {
		return payment.MainPurchaseAttempt{}, fmt.Errorf("fan main payment repository が初期化されていません")
	}

	return s.paymentRepository.GetLatestInflightMainPurchaseAttemptForUpdate(ctx, viewerID, mainID)
}

func (s *Service) getLatestSucceededAttempt(ctx context.Context, viewerID uuid.UUID, mainID uuid.UUID) (payment.MainPurchaseAttempt, error) {
	if s == nil || s.paymentRepository == nil {
		return payment.MainPurchaseAttempt{}, fmt.Errorf("fan main payment repository が初期化されていません")
	}

	return s.paymentRepository.GetLatestSucceededMainPurchaseAttemptForUpdate(ctx, viewerID, mainID)
}

func (s *Service) markUnknownChargePending(ctx context.Context, attempt payment.MainPurchaseAttempt) (PurchaseResult, error) {
	processedAt := s.now().UTC()
	if _, updateErr := s.paymentRepository.UpdateMainPurchaseAttemptOutcome(ctx, payment.UpdateMainPurchaseAttemptOutcomeInput{
		ID:                  attempt.ID,
		PendingReason:       stringPtr(payment.PendingReasonProviderProcessing),
		ProviderProcessedAt: &processedAt,
		Status:              payment.PurchaseAttemptStatusPending,
	}); updateErr != nil {
		return PurchaseResult{}, fmt.Errorf("purchase pending outcome 更新 attempt=%s: %w", attempt.ID, updateErr)
	}

	return buildPendingPurchaseResult(attempt.MainID), nil
}

func (s *Service) markInternalChargeFailure(ctx context.Context, attempt payment.MainPurchaseAttempt) error {
	processedAt := s.now().UTC()
	if _, err := s.paymentRepository.UpdateMainPurchaseAttemptOutcome(ctx, payment.UpdateMainPurchaseAttemptOutcomeInput{
		FailureReason:       stringPtr(payment.FailureReasonPurchaseDeclined),
		ID:                  attempt.ID,
		ProviderProcessedAt: &processedAt,
		Status:              payment.PurchaseAttemptStatusFailed,
	}); err != nil {
		return fmt.Errorf("purchase failure outcome 更新 attempt=%s: %w", attempt.ID, err)
	}

	return nil
}

func (s *Service) resolveConflictingPurchaseAttempt(
	ctx context.Context,
	viewerID uuid.UUID,
	mainID uuid.UUID,
	idempotencyKey string,
) (payment.MainPurchaseAttempt, error) {
	attempt, err := s.paymentRepository.GetMainPurchaseAttemptByIdempotencyKeyForUpdate(ctx, idempotencyKey)
	if err == nil {
		return attempt, nil
	}
	if !errors.Is(err, payment.ErrMainPurchaseAttemptNotFound) {
		return payment.MainPurchaseAttempt{}, fmt.Errorf("purchase attempt idempotency 再取得 viewer=%s main=%s: %w", viewerID, mainID, err)
	}

	attempt, err = s.getInflightAttempt(ctx, viewerID, mainID)
	if err == nil {
		return attempt, nil
	}

	return payment.MainPurchaseAttempt{}, fmt.Errorf("purchase attempt conflict 解決 viewer=%s main=%s: %w", viewerID, mainID, err)
}

func (s *Service) recordPurchaseUnlock(
	ctx context.Context,
	viewerID uuid.UUID,
	mainID uuid.UUID,
	purchasedAt time.Time,
	providerPurchaseRef *string,
) error {
	if s == nil || s.unlockRecorder == nil {
		return fmt.Errorf("fan main unlock recorder が初期化されていません")
	}

	_, err := s.unlockRecorder.RecordMainUnlock(ctx, unlock.RecordMainUnlockInput{
		UserID:                     viewerID,
		MainID:                     mainID,
		PaymentProviderPurchaseRef: providerPurchaseRef,
		PurchasedAt:                &purchasedAt,
	})
	if err == nil || errors.Is(err, unlock.ErrAlreadyUnlocked) {
		return nil
	}

	return fmt.Errorf("main unlock 記録 viewer=%s main=%s: %w", viewerID, mainID, err)
}

func validatePurchaseInput(input PurchaseInput, savedMethods []payment.SavedPaymentMethod) error {
	switch strings.TrimSpace(input.PaymentMethod.Mode) {
	case payment.PaymentMethodModeSavedCard:
		if strings.TrimSpace(input.PaymentMethod.PaymentMethodID) == "" {
			return ErrInvalidPurchaseRequest
		}
	case payment.PaymentMethodModeNewCard:
		if strings.TrimSpace(input.PaymentMethod.CardSetupToken) == "" {
			return ErrInvalidPurchaseRequest
		}
		if len(savedMethods) == 0 && (!input.AcceptedAge || !input.AcceptedTerms) {
			return ErrInvalidPurchaseRequest
		}
	default:
		return ErrInvalidPurchaseRequest
	}

	return nil
}

func buildPurchaseIdempotencyKey(input PurchaseInput) string {
	builder := strings.Builder{}
	builder.WriteString(input.ViewerID.String())
	builder.WriteString(":")
	builder.WriteString(input.MainID.String())
	builder.WriteString(":")
	builder.WriteString(input.FromShortID.String())
	builder.WriteString(":")
	builder.WriteString(strings.TrimSpace(input.EntryToken))
	builder.WriteString(":")
	builder.WriteString(strings.TrimSpace(input.PaymentMethod.Mode))
	builder.WriteString(":")
	builder.WriteString(strings.TrimSpace(input.PaymentMethod.PaymentMethodID))
	builder.WriteString(":")
	builder.WriteString(strings.TrimSpace(input.PaymentMethod.CardSetupToken))

	sum := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}

func buildUnlockPurchaseState(
	source feed.UnlockPreview,
	savedMethods []payment.SavedPaymentMethod,
	inflight *payment.MainPurchaseAttempt,
) UnlockPurchaseState {
	state := "purchase_ready"
	setup := PurchaseSetupState{}
	pendingReason := (*string)(nil)

	switch {
	case source.IsOwner:
		state = "owner_preview"
	case source.IsUnlocked:
		state = "already_purchased"
	case inflight != nil:
		state = "purchase_pending"
		pendingReason = inflight.PendingReason
		if pendingReason == nil {
			pendingReason = stringPtr(payment.PendingReasonProviderProcessing)
		}
	case len(savedMethods) == 0:
		state = "setup_required"
		setup = PurchaseSetupState{
			Required:                true,
			RequiresAgeConfirmation: true,
			RequiresCardSetup:       true,
			RequiresTermsAcceptance: true,
		}
	}

	summaries := make([]SavedPaymentMethodSummary, 0, len(savedMethods))
	for _, method := range savedMethods {
		summaries = append(summaries, SavedPaymentMethodSummary{
			Brand:           method.Brand,
			Last4:           method.Last4,
			PaymentMethodID: method.PaymentMethodID,
		})
	}

	return UnlockPurchaseState{
		PendingReason:       pendingReason,
		SavedPaymentMethods: summaries,
		Setup:               setup,
		State:               state,
		SupportedCardBrands: append([]string(nil), supportedCardBrands...),
	}
}

func buildPurchaseResultFromAttempt(mainID uuid.UUID, attempt payment.MainPurchaseAttempt, entryToken string) PurchaseResult {
	switch attempt.Status {
	case payment.PurchaseAttemptStatusSucceeded:
		return PurchaseResult{
			Access:     buildPurchasedAccessState(mainID),
			EntryToken: stringPtr(entryToken),
			Purchase: PurchaseOutcome{
				CanRetry: false,
				Status:   "succeeded",
			},
		}
	case payment.PurchaseAttemptStatusPending, payment.PurchaseAttemptStatusProcessing:
		return buildPendingPurchaseResult(mainID)
	default:
		return PurchaseResult{
			Access: buildLockedAccessState(mainID),
			Purchase: PurchaseOutcome{
				CanRetry:      true,
				FailureReason: attempt.FailureReason,
				Status:        "failed",
			},
		}
	}
}

func buildPendingPurchaseResult(mainID uuid.UUID) PurchaseResult {
	return PurchaseResult{
		Access: buildLockedAccessState(mainID),
		Purchase: PurchaseOutcome{
			CanRetry: false,
			Status:   "pending",
		},
	}
}

func buildAlreadyPurchasedResult(mainID uuid.UUID, entryToken string) PurchaseResult {
	return PurchaseResult{
		Access:     buildPurchasedAccessState(mainID),
		EntryToken: stringPtr(entryToken),
		Purchase: PurchaseOutcome{
			CanRetry: false,
			Status:   "already_purchased",
		},
	}
}

func buildOwnerPurchaseResult(mainID uuid.UUID, entryToken string) PurchaseResult {
	return PurchaseResult{
		Access:     buildOwnerAccessState(mainID),
		EntryToken: stringPtr(entryToken),
		Purchase: PurchaseOutcome{
			CanRetry: false,
			Status:   "owner_preview",
		},
	}
}
