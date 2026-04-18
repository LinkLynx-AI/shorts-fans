package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/unlock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type webhookUnlockRecorderStub struct {
	recordMainUnlock func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error)
}

func (s webhookUnlockRecorderStub) RecordMainUnlock(ctx context.Context, input unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
	return s.recordMainUnlock(ctx, input)
}

func TestNewCCBillWebhookHandler(t *testing.T) {
	t.Parallel()

	client := newWebhookTestClient(t)
	repo := newRepository(repositoryStubQueries{})
	recorder := webhookUnlockRecorderStub{
		recordMainUnlock: func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
			return unlock.MainUnlock{}, nil
		},
	}

	if handler := NewCCBillWebhookHandler(nil, recorder, client); handler != nil {
		t.Fatalf("NewCCBillWebhookHandler(nil repo) = %#v, want nil", handler)
	}
	if handler := NewCCBillWebhookHandler(repo, nil, client); handler != nil {
		t.Fatalf("NewCCBillWebhookHandler(nil recorder) = %#v, want nil", handler)
	}
	if handler := NewCCBillWebhookHandler(repo, recorder, nil); handler != nil {
		t.Fatalf("NewCCBillWebhookHandler(nil client) = %#v, want nil", handler)
	}
	if handler := NewCCBillWebhookHandler(repo, recorder, client); handler == nil {
		t.Fatal("NewCCBillWebhookHandler() = nil, want value")
	}
}

func TestCCBillWebhookHandlerHandleWebhookSuccess(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_710_000_000, 0).UTC()
	userID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()
	attemptID := uuid.New()
	methodID := uuid.New()
	attemptRow := testMainPurchaseAttemptRow(attemptID, userID, mainID, shortID, &methodID, now)
	attemptRow.ProviderPaymentTokenRef = "token-2"
	attemptRow.Status = PurchaseAttemptStatusPending
	paymentMethodRow := testPaymentMethodRow(methodID, userID, now)
	upserts := 0
	updates := 0
	unlocks := 0

	repo := newRepository(repositoryStubQueries{
		getAttemptForUpdate: func(_ context.Context, gotID pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
			if gotID != pgUUID(attemptID) {
				t.Fatalf("GetMainPurchaseAttemptForUpdate() got %v want %v", gotID, pgUUID(attemptID))
			}
			return attemptRow, nil
		},
		upsertPaymentMethod: func(_ context.Context, arg sqlc.UpsertUserPaymentMethodParams) (sqlc.AppUserPaymentMethod, error) {
			upserts++
			if arg.ProviderPaymentTokenRef != "token-2" || arg.ProviderPaymentAccountRef != "acct-2" {
				t.Fatalf("UpsertUserPaymentMethod() arg got %#v", arg)
			}
			if arg.Brand != CardBrandVisa || arg.Last4 != "4242" {
				t.Fatalf("UpsertUserPaymentMethod() arg got %#v", arg)
			}
			return paymentMethodRow, nil
		},
		updateAttemptOutcome: func(_ context.Context, arg sqlc.UpdateMainPurchaseAttemptOutcomeParams) (sqlc.AppMainPurchaseAttempt, error) {
			updates++
			if arg.ID != pgUUID(attemptID) || arg.Status != PurchaseAttemptStatusSucceeded {
				t.Fatalf("UpdateMainPurchaseAttemptOutcome() arg got %#v", arg)
			}
			if arg.ProviderPurchaseRef != pgText("txn-1") || arg.ProviderTransactionRef != pgText("txn-1") {
				t.Fatalf("UpdateMainPurchaseAttemptOutcome() refs got %#v", arg)
			}

			updated := attemptRow
			updated.Status = PurchaseAttemptStatusSucceeded
			updated.ProviderPurchaseRef = pgText("txn-1")
			updated.ProviderTransactionRef = pgText("txn-1")
			updated.ProviderProcessedAt = arg.ProviderProcessedAt
			return updated, nil
		},
	})
	handler := NewCCBillWebhookHandler(repo, webhookUnlockRecorderStub{
		recordMainUnlock: func(_ context.Context, input unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
			unlocks++
			if input.UserID != userID || input.MainID != mainID {
				t.Fatalf("RecordMainUnlock() input got %+v", input)
			}
			if input.PaymentProviderPurchaseRef == nil || *input.PaymentProviderPurchaseRef != "txn-1" {
				t.Fatalf("RecordMainUnlock() purchase ref got %#v want txn-1", input.PaymentProviderPurchaseRef)
			}

			return unlock.MainUnlock{
				UserID:      userID,
				MainID:      mainID,
				PurchasedAt: now,
				CreatedAt:   now,
			}, nil
		},
	}, newWebhookTestClient(t))
	handler.now = func() time.Time { return now }

	err := handler.HandleWebhook(
		context.Background(),
		"203.0.113.10:443",
		url.Values{"eventType": []string{ccbillWebhookEventUpSaleSuccess}},
		"application/json",
		[]byte(fmt.Sprintf(`{"transactionId":"txn-1","paymentAccount":"acct-2","last4":"4242","cardType":"VISA","X-attemptId":"%s","timestamp":"2024-03-09 16:00:00"}`, attemptID)),
	)
	if err != nil {
		t.Fatalf("HandleWebhook() error = %v, want nil", err)
	}
	if upserts != 1 || updates != 1 || unlocks != 1 {
		t.Fatalf("HandleWebhook() side effects got upserts=%d updates=%d unlocks=%d", upserts, updates, unlocks)
	}
}

func TestCCBillWebhookHandlerHandleWebhookFailure(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_710_000_000, 0).UTC()
	userID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()
	attemptID := uuid.New()
	attemptRow := testMainPurchaseAttemptRow(attemptID, userID, mainID, shortID, nil, now)
	attemptRow.Status = PurchaseAttemptStatusPending
	updates := 0

	repo := newRepository(repositoryStubQueries{
		getAttemptByProvider: func(_ context.Context, providerPurchaseRef pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
			if providerPurchaseRef != pgText("purchase-ref-2") {
				t.Fatalf("GetMainPurchaseAttemptByProviderPurchaseRefForUpdate() got %v want %v", providerPurchaseRef, pgText("purchase-ref-2"))
			}
			return attemptRow, nil
		},
		getAttemptByTxn: func(context.Context, pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, fmt.Errorf("lookup by txn: %w", ErrMainPurchaseAttemptNotFound)
		},
		updateAttemptOutcome: func(_ context.Context, arg sqlc.UpdateMainPurchaseAttemptOutcomeParams) (sqlc.AppMainPurchaseAttempt, error) {
			updates++
			if arg.Status != PurchaseAttemptStatusFailed {
				t.Fatalf("UpdateMainPurchaseAttemptOutcome() status got %q want %q", arg.Status, PurchaseAttemptStatusFailed)
			}
			if arg.FailureReason != pgText(FailureReasonCardBrandUnsupported) {
				t.Fatalf("UpdateMainPurchaseAttemptOutcome() failure reason got %#v want %#v", arg.FailureReason, pgText(FailureReasonCardBrandUnsupported))
			}
			return attemptRow, nil
		},
	})
	handler := NewCCBillWebhookHandler(repo, webhookUnlockRecorderStub{
		recordMainUnlock: func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
			t.Fatal("RecordMainUnlock() was called unexpectedly")
			return unlock.MainUnlock{}, nil
		},
	}, newWebhookTestClient(t))
	handler.now = func() time.Time { return now }

	err := handler.HandleWebhook(
		context.Background(),
		"203.0.113.10:443",
		nil,
		"application/x-www-form-urlencoded",
		[]byte("eventType=NewSaleFailure&transactionId=purchase-ref-2&failureCode=40&failureReason=declined"),
	)
	if err != nil {
		t.Fatalf("HandleWebhook() error = %v, want nil", err)
	}
	if updates != 1 {
		t.Fatalf("HandleWebhook() updates got %d want %d", updates, 1)
	}
}

func TestCCBillWebhookHandlerIgnoresUnknownOrMissingAttempts(t *testing.T) {
	t.Parallel()

	handler := NewCCBillWebhookHandler(
		newRepository(repositoryStubQueries{
			getAttemptByTxn: func(context.Context, pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
				return sqlc.AppMainPurchaseAttempt{}, fmt.Errorf("not found: %w", ErrMainPurchaseAttemptNotFound)
			},
			getAttemptByProvider: func(context.Context, pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
				return sqlc.AppMainPurchaseAttempt{}, fmt.Errorf("not found: %w", ErrMainPurchaseAttemptNotFound)
			},
		}),
		webhookUnlockRecorderStub{
			recordMainUnlock: func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
				t.Fatal("RecordMainUnlock() was called unexpectedly")
				return unlock.MainUnlock{}, nil
			},
		},
		newWebhookTestClient(t),
	)

	err := handler.HandleWebhook(
		context.Background(),
		"203.0.113.10:443",
		url.Values{"eventType": []string{"OtherEvent"}},
		"application/json",
		[]byte(`{"transactionId":"txn-1"}`),
	)
	if err != nil {
		t.Fatalf("HandleWebhook() unknown event error = %v, want nil", err)
	}

	err = handler.HandleWebhook(
		context.Background(),
		"203.0.113.10:443",
		url.Values{"eventType": []string{ccbillWebhookEventNewSaleSuccess}},
		"application/json",
		[]byte(`{"transactionId":"txn-1"}`),
	)
	if err != nil {
		t.Fatalf("HandleWebhook() missing attempt error = %v, want nil", err)
	}
}

func TestDecodeCCBillWebhookEventAndHelpers(t *testing.T) {
	t.Parallel()

	event, err := decodeCCBillWebhookEvent(
		url.Values{"eventType": []string{ccbillWebhookEventNewSaleSuccess}},
		"application/json",
		[]byte(`{"transactionId":"txn-1","paymentAccount":"acct-1","last4":"4242","cardType":"VISA","X-attemptId":"11111111-1111-1111-1111-111111111111","timestamp":"2024-03-09 16:00:00"}`),
	)
	if err != nil {
		t.Fatalf("decodeCCBillWebhookEvent(json) error = %v, want nil", err)
	}
	if event.TransactionID != "txn-1" || event.PassThrough["X-attemptId"] == "" {
		t.Fatalf("decodeCCBillWebhookEvent(json) got %#v", event)
	}
	if event.Timestamp == nil {
		t.Fatalf("decodeCCBillWebhookEvent(json) timestamp = nil, want value")
	}

	fields, err := decodeCCBillWebhookFields("application/x-www-form-urlencoded", []byte("eventType=NewSaleSuccess&transactionId=txn-1"))
	if err != nil {
		t.Fatalf("decodeCCBillWebhookFields(form) error = %v, want nil", err)
	}
	if fields["transactionId"] != "txn-1" {
		t.Fatalf("decodeCCBillWebhookFields(form) got %#v", fields)
	}

	if _, err := decodeCCBillWebhookEvent(nil, "application/json", []byte(`{}`)); !errors.Is(err, ErrCCBillWebhookInvalid) {
		t.Fatalf("decodeCCBillWebhookEvent(missing type) error got %v want wrapped %v", err, ErrCCBillWebhookInvalid)
	}
	if _, err := decodeCCBillWebhookFields("application/json", []byte(" ")); !errors.Is(err, ErrCCBillWebhookInvalid) {
		t.Fatalf("decodeCCBillWebhookFields(empty) error got %v want wrapped %v", err, ErrCCBillWebhookInvalid)
	}

	if !shouldPersistWebhookPaymentMethod(MainPurchaseAttempt{ProviderPaymentTokenRef: "token-1"}, ccbillWebhookEvent{PaymentAccount: "acct-1", Last4: "4242", CardType: "VISA"}) {
		t.Fatal("shouldPersistWebhookPaymentMethod() = false, want true")
	}
	if shouldPersistWebhookPaymentMethod(MainPurchaseAttempt{}, ccbillWebhookEvent{}) {
		t.Fatal("shouldPersistWebhookPaymentMethod() = true, want false")
	}
	if mapCCBillCardType("AMEX") != CardBrandAmericanExpress {
		t.Fatalf("mapCCBillCardType(AMEX) got %q want %q", mapCCBillCardType("AMEX"), CardBrandAmericanExpress)
	}
	if mapWebhookFailureReason("40", nil) != FailureReasonCardBrandUnsupported {
		t.Fatalf("mapWebhookFailureReason(40) got %q want %q", mapWebhookFailureReason("40", nil), FailureReasonCardBrandUnsupported)
	}
	if string(bytesTrimSpace([]byte("  value \n"))) != "value" {
		t.Fatalf("bytesTrimSpace() got %q want %q", string(bytesTrimSpace([]byte("  value \n"))), "value")
	}
}

func TestCCBillWebhookHandlerFindAttemptByPassThroughAndSucceededFailureNoop(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_710_000_000, 0).UTC()
	userID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()
	attemptID := uuid.New()
	row := testMainPurchaseAttemptRow(attemptID, userID, mainID, shortID, nil, now)
	row.Status = PurchaseAttemptStatusSucceeded
	repo := newRepository(repositoryStubQueries{
		getAttemptForUpdate: func(_ context.Context, id pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
			if id != pgUUID(attemptID) {
				t.Fatalf("GetMainPurchaseAttemptByIDForUpdate() id got %v want %v", id, pgUUID(attemptID))
			}
			return row, nil
		},
	})
	handler := NewCCBillWebhookHandler(repo, webhookUnlockRecorderStub{
		recordMainUnlock: func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
			t.Fatal("RecordMainUnlock() was called unexpectedly")
			return unlock.MainUnlock{}, nil
		},
	}, newWebhookTestClient(t))

	err := handler.HandleWebhook(
		context.Background(),
		"203.0.113.10:443",
		nil,
		"application/json",
		mustJSON(t, map[string]string{
			"eventType":     ccbillWebhookEventNewSaleFailure,
			"transactionId": "txn-1",
			"X-attemptId":   attemptID.String(),
		}),
	)
	if err != nil {
		t.Fatalf("HandleWebhook() error = %v, want nil", err)
	}
}

func TestCCBillWebhookHandlerSucceededSuccessNoopWhenRefsAlreadyMatch(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_710_000_000, 0).UTC()
	userID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()
	attemptID := uuid.New()
	row := testMainPurchaseAttemptRow(attemptID, userID, mainID, shortID, nil, now)
	row.Status = PurchaseAttemptStatusSucceeded
	row.ProviderPurchaseRef = pgText("txn-1")
	row.ProviderTransactionRef = pgText("txn-1")

	repo := newRepository(repositoryStubQueries{
		getAttemptForUpdate: func(_ context.Context, id pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
			if id != pgUUID(attemptID) {
				t.Fatalf("GetMainPurchaseAttemptByIDForUpdate() id got %v want %v", id, pgUUID(attemptID))
			}
			return row, nil
		},
		updateAttemptOutcome: func(context.Context, sqlc.UpdateMainPurchaseAttemptOutcomeParams) (sqlc.AppMainPurchaseAttempt, error) {
			t.Fatal("UpdateMainPurchaseAttemptOutcome() was called unexpectedly")
			return sqlc.AppMainPurchaseAttempt{}, nil
		},
		upsertPaymentMethod: func(context.Context, sqlc.UpsertUserPaymentMethodParams) (sqlc.AppUserPaymentMethod, error) {
			t.Fatal("UpsertUserPaymentMethod() was called unexpectedly")
			return sqlc.AppUserPaymentMethod{}, nil
		},
	})
	handler := NewCCBillWebhookHandler(repo, webhookUnlockRecorderStub{
		recordMainUnlock: func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
			t.Fatal("RecordMainUnlock() was called unexpectedly")
			return unlock.MainUnlock{}, nil
		},
	}, newWebhookTestClient(t))

	err := handler.HandleWebhook(
		context.Background(),
		"203.0.113.10:443",
		nil,
		"application/json",
		mustJSON(t, map[string]string{
			"eventType":     ccbillWebhookEventNewSaleSuccess,
			"transactionId": "txn-1",
			"X-attemptId":   attemptID.String(),
		}),
	)
	if err != nil {
		t.Fatalf("HandleWebhook() error = %v, want nil", err)
	}
}

func TestCCBillWebhookHelperBranches(t *testing.T) {
	t.Parallel()

	handler := &CCBillWebhookHandler{
		now: func() time.Time { return time.Unix(1_710_000_000, 0).UTC() },
	}

	if got := handler.eventProcessedAt(ccbillWebhookEvent{}); !got.Equal(time.Unix(1_710_000_000, 0).UTC()) {
		t.Fatalf("eventProcessedAt() got %s want fallback now", got)
	}
	if mapCCBillCardType("MASTERCARD") != CardBrandMastercard {
		t.Fatalf("mapCCBillCardType(MASTERCARD) got %q want %q", mapCCBillCardType("MASTERCARD"), CardBrandMastercard)
	}
	if mapCCBillCardType("JCB") != CardBrandJCB {
		t.Fatalf("mapCCBillCardType(JCB) got %q want %q", mapCCBillCardType("JCB"), CardBrandJCB)
	}
	if mapCCBillCardType("other") != "" {
		t.Fatalf("mapCCBillCardType(other) got %q want empty", mapCCBillCardType("other"))
	}
	if mapWebhookFailureReason("", stringPtr("plain decline")) != FailureReasonPurchaseDeclined {
		t.Fatalf("mapWebhookFailureReason(default) got %q want %q", mapWebhookFailureReason("", stringPtr("plain decline")), FailureReasonPurchaseDeclined)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v, want nil", err)
	}

	return raw
}

func newWebhookTestClient(t *testing.T) *CCBillClient {
	t.Helper()

	client, err := NewCCBillClient(CCBillConfig{
		BackendClientID:     "backend-id",
		BackendClientSecret: "backend-secret",
		ClientAccountNumber: 900100,
		ClientSubAccount:    1,
		CurrencyCode:        392,
		InitialPeriodDays:   30,
		WebhookAllowedCIDRs: []string{"203.0.113.0/24"},
	}, nil)
	if err != nil {
		t.Fatalf("NewCCBillClient() error = %v, want nil", err)
	}

	return client
}
