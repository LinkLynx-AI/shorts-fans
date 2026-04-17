package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/unlock"
	"github.com/google/uuid"
)

const (
	ccbillWebhookEventNewSaleSuccess = "NewSaleSuccess"
	ccbillWebhookEventNewSaleFailure = "NewSaleFailure"
	ccbillWebhookEventUpSaleSuccess  = "UpSaleSuccess"
	ccbillWebhookEventUpSaleFailure  = "UpSaleFailure"
	ccbillWebhookTimestampLayout     = "2006-01-02 15:04:05"
)

// ErrCCBillWebhookInvalid は webhook payload が不正なことを表します。
var ErrCCBillWebhookInvalid = errors.New("ccbill webhook payload is invalid")

type ccbillWebhookEvent struct {
	CardType       string
	EventType      string
	FailureCode    string
	FailureReason  string
	Last4          string
	PassThrough    map[string]string
	PaymentAccount string
	Timestamp      *time.Time
	TransactionID  string
}

type webhookUnlockRecorder interface {
	RecordMainUnlock(ctx context.Context, input unlock.RecordMainUnlockInput) (unlock.MainUnlock, error)
}

// CCBillWebhookHandler は CCBill webhook を解釈して purchase state を更新します。
type CCBillWebhookHandler struct {
	client         *CCBillClient
	now            func() time.Time
	repository     *Repository
	unlockRecorder webhookUnlockRecorder
}

// NewCCBillWebhookHandler は webhook reconciliation handler を構築します。
func NewCCBillWebhookHandler(repository *Repository, unlockRecorder webhookUnlockRecorder, client *CCBillClient) *CCBillWebhookHandler {
	if repository == nil || unlockRecorder == nil || client == nil {
		return nil
	}

	return &CCBillWebhookHandler{
		client:         client,
		now:            time.Now,
		repository:     repository,
		unlockRecorder: unlockRecorder,
	}
}

// HandleWebhook は 1 件の CCBill webhook を処理します。
func (h *CCBillWebhookHandler) HandleWebhook(
	ctx context.Context,
	remoteIP string,
	query url.Values,
	contentType string,
	body []byte,
) error {
	if h == nil || h.client == nil || h.repository == nil || h.unlockRecorder == nil {
		return fmt.Errorf("ccbill webhook handler が初期化されていません")
	}

	if err := h.client.ValidateWebhookOrigin(remoteIP); err != nil {
		return err
	}

	event, err := decodeCCBillWebhookEvent(query, contentType, body)
	if err != nil {
		return err
	}
	if !isSupportedCCBillWebhookEvent(event.EventType) {
		return nil
	}

	attempt, err := h.findAttemptForWebhook(ctx, event)
	if err != nil {
		if errors.Is(err, ErrMainPurchaseAttemptNotFound) {
			return nil
		}

		return err
	}

	switch event.EventType {
	case ccbillWebhookEventNewSaleSuccess, ccbillWebhookEventUpSaleSuccess:
		return h.handleSuccess(ctx, attempt, event)
	case ccbillWebhookEventNewSaleFailure, ccbillWebhookEventUpSaleFailure:
		return h.handleFailure(ctx, attempt, event)
	default:
		return nil
	}
}

func (h *CCBillWebhookHandler) findAttemptForWebhook(ctx context.Context, event ccbillWebhookEvent) (MainPurchaseAttempt, error) {
	if attemptID := strings.TrimSpace(event.PassThrough["X-attemptId"]); attemptID != "" {
		parsedID, err := uuid.Parse(attemptID)
		if err == nil {
			return h.repository.GetMainPurchaseAttemptForUpdate(ctx, parsedID)
		}
	}

	if transactionID := strings.TrimSpace(event.TransactionID); transactionID != "" {
		attempt, err := h.repository.GetMainPurchaseAttemptByProviderTransactionRefForUpdate(ctx, transactionID)
		if err == nil {
			return attempt, nil
		}
		if !errors.Is(err, ErrMainPurchaseAttemptNotFound) {
			return MainPurchaseAttempt{}, err
		}

		return h.repository.GetMainPurchaseAttemptByProviderPurchaseRefForUpdate(ctx, transactionID)
	}

	return MainPurchaseAttempt{}, ErrMainPurchaseAttemptNotFound
}

func (h *CCBillWebhookHandler) handleSuccess(ctx context.Context, attempt MainPurchaseAttempt, event ccbillWebhookEvent) error {
	processedAt := h.eventProcessedAt(event)
	providerRef := nonEmptyStringPtr(strings.TrimSpace(event.TransactionID))
	if providerRef == nil {
		providerRef = attempt.ProviderPurchaseRef
	}

	if shouldPersistWebhookPaymentMethod(attempt, event) {
		if _, err := h.repository.UpsertSavedPaymentMethod(ctx, UpsertSavedPaymentMethodInput{
			Brand:                     mapCCBillCardType(event.CardType),
			Last4:                     strings.TrimSpace(event.Last4),
			LastUsedAt:                &processedAt,
			Provider:                  attempt.Provider,
			ProviderPaymentAccountRef: strings.TrimSpace(event.PaymentAccount),
			ProviderPaymentTokenRef:   attempt.ProviderPaymentTokenRef,
			UserID:                    attempt.UserID,
		}); err != nil {
			return fmt.Errorf("ccbill webhook saved payment method upsert attempt=%s: %w", attempt.ID, err)
		}
	}

	if attempt.Status != PurchaseAttemptStatusSucceeded {
		if _, err := h.unlockRecorder.RecordMainUnlock(ctx, unlock.RecordMainUnlockInput{
			UserID:                     attempt.UserID,
			MainID:                     attempt.MainID,
			PaymentProviderPurchaseRef: providerRef,
			PurchasedAt:                &processedAt,
		}); err != nil && !errors.Is(err, unlock.ErrAlreadyUnlocked) {
			return fmt.Errorf("ccbill webhook main unlock record attempt=%s: %w", attempt.ID, err)
		}
	}

	if attempt.Status == PurchaseAttemptStatusSucceeded &&
		valueOrEmpty(attempt.ProviderTransactionRef) == valueOrEmpty(providerRef) &&
		valueOrEmpty(attempt.ProviderPurchaseRef) == valueOrEmpty(providerRef) {
		return nil
	}

	_, err := h.repository.UpdateMainPurchaseAttemptOutcome(ctx, UpdateMainPurchaseAttemptOutcomeInput{
		ID:                     attempt.ID,
		ProviderProcessedAt:    &processedAt,
		ProviderPurchaseRef:    providerRef,
		ProviderTransactionRef: providerRef,
		Status:                 PurchaseAttemptStatusSucceeded,
	})
	if err != nil {
		return fmt.Errorf("ccbill webhook main purchase success attempt=%s: %w", attempt.ID, err)
	}

	return nil
}

func (h *CCBillWebhookHandler) handleFailure(ctx context.Context, attempt MainPurchaseAttempt, event ccbillWebhookEvent) error {
	if attempt.Status == PurchaseAttemptStatusSucceeded {
		return nil
	}

	processedAt := h.eventProcessedAt(event)
	providerRef := nonEmptyStringPtr(strings.TrimSpace(event.TransactionID))
	declineText := nonEmptyStringPtr(strings.TrimSpace(event.FailureReason))
	failureReason := stringPtr(mapWebhookFailureReason(strings.TrimSpace(event.FailureCode), declineText))

	_, err := h.repository.UpdateMainPurchaseAttemptOutcome(ctx, UpdateMainPurchaseAttemptOutcomeInput{
		FailureReason:          failureReason,
		ID:                     attempt.ID,
		ProviderDeclineText:    declineText,
		ProviderProcessedAt:    &processedAt,
		ProviderPurchaseRef:    providerRef,
		ProviderTransactionRef: providerRef,
		Status:                 PurchaseAttemptStatusFailed,
	})
	if err != nil {
		return fmt.Errorf("ccbill webhook main purchase failure attempt=%s: %w", attempt.ID, err)
	}

	return nil
}

func (h *CCBillWebhookHandler) eventProcessedAt(event ccbillWebhookEvent) time.Time {
	if event.Timestamp != nil {
		return event.Timestamp.UTC()
	}

	return h.now().UTC()
}

func decodeCCBillWebhookEvent(query url.Values, contentType string, body []byte) (ccbillWebhookEvent, error) {
	fields, err := decodeCCBillWebhookFields(contentType, body)
	if err != nil {
		return ccbillWebhookEvent{}, err
	}

	eventType := strings.TrimSpace(query.Get("eventType"))
	if eventType == "" {
		eventType = strings.TrimSpace(fields["eventType"])
	}
	if eventType == "" {
		return ccbillWebhookEvent{}, fmt.Errorf("%w: missing eventType", ErrCCBillWebhookInvalid)
	}

	event := ccbillWebhookEvent{
		CardType:       strings.TrimSpace(fields["cardType"]),
		EventType:      eventType,
		FailureCode:    strings.TrimSpace(fields["failureCode"]),
		FailureReason:  strings.TrimSpace(fields["failureReason"]),
		Last4:          strings.TrimSpace(fields["last4"]),
		PassThrough:    make(map[string]string),
		PaymentAccount: strings.TrimSpace(fields["paymentAccount"]),
		TransactionID:  strings.TrimSpace(fields["transactionId"]),
	}

	for key, value := range fields {
		if strings.HasPrefix(key, "X-") {
			event.PassThrough[key] = value
		}
	}

	if rawTimestamp := strings.TrimSpace(fields["timestamp"]); rawTimestamp != "" {
		parsedTimestamp, err := time.ParseInLocation(ccbillWebhookTimestampLayout, rawTimestamp, time.UTC)
		if err != nil {
			return ccbillWebhookEvent{}, fmt.Errorf("%w: parse timestamp: %v", ErrCCBillWebhookInvalid, err)
		}
		event.Timestamp = &parsedTimestamp
	}

	return event, nil
}

func isSupportedCCBillWebhookEvent(eventType string) bool {
	switch strings.TrimSpace(eventType) {
	case ccbillWebhookEventNewSaleSuccess, ccbillWebhookEventNewSaleFailure, ccbillWebhookEventUpSaleSuccess, ccbillWebhookEventUpSaleFailure:
		return true
	default:
		return false
	}
}

func decodeCCBillWebhookFields(contentType string, body []byte) (map[string]string, error) {
	trimmedBody := bytesTrimSpace(body)
	if len(trimmedBody) == 0 {
		return nil, fmt.Errorf("%w: empty body", ErrCCBillWebhookInvalid)
	}

	switch {
	case strings.Contains(strings.ToLower(contentType), "application/json"):
		var decoded map[string]any
		if err := json.Unmarshal(trimmedBody, &decoded); err != nil {
			return nil, fmt.Errorf("%w: decode json: %v", ErrCCBillWebhookInvalid, err)
		}

		fields := make(map[string]string, len(decoded))
		for key, value := range decoded {
			fields[key] = stringFromAny(value)
		}
		return fields, nil
	default:
		values, err := url.ParseQuery(string(trimmedBody))
		if err != nil {
			return nil, fmt.Errorf("%w: decode form: %v", ErrCCBillWebhookInvalid, err)
		}

		fields := make(map[string]string, len(values))
		for key, value := range values {
			fields[key] = strings.TrimSpace(strings.Join(value, ","))
		}
		return fields, nil
	}
}

func shouldPersistWebhookPaymentMethod(attempt MainPurchaseAttempt, event ccbillWebhookEvent) bool {
	if strings.TrimSpace(attempt.ProviderPaymentTokenRef) == "" {
		return false
	}
	if strings.TrimSpace(event.PaymentAccount) == "" || strings.TrimSpace(event.Last4) == "" {
		return false
	}
	return mapCCBillCardType(event.CardType) != ""
}

func mapCCBillCardType(cardType string) string {
	switch strings.ToUpper(strings.TrimSpace(cardType)) {
	case "VISA":
		return CardBrandVisa
	case "MASTERCARD":
		return CardBrandMastercard
	case "JCB":
		return CardBrandJCB
	case "AMEX":
		return CardBrandAmericanExpress
	default:
		return ""
	}
}

func mapWebhookFailureReason(failureCode string, failureText *string) string {
	switch strings.TrimSpace(failureCode) {
	case "3", "40":
		return FailureReasonCardBrandUnsupported
	}

	return mapChargeFailureReason(nil, failureText)
}

func bytesTrimSpace(value []byte) []byte {
	return []byte(strings.TrimSpace(string(value)))
}
