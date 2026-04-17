package fanmain

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/payment"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/unlock"
	"github.com/google/uuid"
)

type stubFeedReader struct {
	getDetail func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error)
}

func (s stubFeedReader) GetDetail(ctx context.Context, shortID uuid.UUID, viewerUserID *uuid.UUID) (feed.Detail, error) {
	return s.getDetail(ctx, shortID, viewerUserID)
}

type stubMainReader struct {
	getUnlockableMain func(context.Context, uuid.UUID) (shorts.Main, error)
}

func (s stubMainReader) GetUnlockableMain(ctx context.Context, id uuid.UUID) (shorts.Main, error) {
	return s.getUnlockableMain(ctx, id)
}

type stubUnlockRecorder struct {
	recordMainUnlock func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error)
}

func (s stubUnlockRecorder) RecordMainUnlock(ctx context.Context, input unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
	if s.recordMainUnlock == nil {
		return unlock.MainUnlock{}, nil
	}

	return s.recordMainUnlock(ctx, input)
}

type stubPaymentRepository struct {
	createMainPurchaseAttempt              func(context.Context, payment.CreateMainPurchaseAttemptInput) (payment.MainPurchaseAttempt, error)
	getLatestInflightMainPurchaseAttempt   func(context.Context, uuid.UUID, uuid.UUID) (payment.MainPurchaseAttempt, error)
	getLatestSucceededMainPurchaseAttempt  func(context.Context, uuid.UUID, uuid.UUID) (payment.MainPurchaseAttempt, error)
	getMainPurchaseAttemptByIdempotencyKey func(context.Context, string) (payment.MainPurchaseAttempt, error)
	getSavedPaymentMethod                  func(context.Context, uuid.UUID, string) (payment.SavedPaymentMethod, error)
	listSavedPaymentMethods                func(context.Context, uuid.UUID) ([]payment.SavedPaymentMethod, error)
	touchSavedPaymentMethodLastUsedAt      func(context.Context, uuid.UUID, string, *time.Time) (payment.SavedPaymentMethod, error)
	updateMainPurchaseAttemptOutcome       func(context.Context, payment.UpdateMainPurchaseAttemptOutcomeInput) (payment.MainPurchaseAttempt, error)
}

type stubTransactionalPaymentRepository struct {
	stubPaymentRepository
	acquireMainPurchaseLock func(context.Context, uuid.UUID, uuid.UUID) error
	runInTx                 func(context.Context, func(payment.TxRepository) error) error
}

func (s stubPaymentRepository) CreateMainPurchaseAttempt(ctx context.Context, input payment.CreateMainPurchaseAttemptInput) (payment.MainPurchaseAttempt, error) {
	if s.createMainPurchaseAttempt == nil {
		return payment.MainPurchaseAttempt{}, nil
	}

	return s.createMainPurchaseAttempt(ctx, input)
}

func (s stubPaymentRepository) GetLatestInflightMainPurchaseAttemptForUpdate(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) (payment.MainPurchaseAttempt, error) {
	if s.getLatestInflightMainPurchaseAttempt == nil {
		return payment.MainPurchaseAttempt{}, payment.ErrMainPurchaseAttemptNotFound
	}

	return s.getLatestInflightMainPurchaseAttempt(ctx, userID, mainID)
}

func (s stubPaymentRepository) GetLatestSucceededMainPurchaseAttemptForUpdate(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) (payment.MainPurchaseAttempt, error) {
	if s.getLatestSucceededMainPurchaseAttempt == nil {
		return payment.MainPurchaseAttempt{}, payment.ErrMainPurchaseAttemptNotFound
	}

	return s.getLatestSucceededMainPurchaseAttempt(ctx, userID, mainID)
}

func (s stubPaymentRepository) GetMainPurchaseAttemptByIdempotencyKeyForUpdate(ctx context.Context, idempotencyKey string) (payment.MainPurchaseAttempt, error) {
	if s.getMainPurchaseAttemptByIdempotencyKey == nil {
		return payment.MainPurchaseAttempt{}, payment.ErrMainPurchaseAttemptNotFound
	}

	return s.getMainPurchaseAttemptByIdempotencyKey(ctx, idempotencyKey)
}

func (s stubPaymentRepository) GetSavedPaymentMethod(ctx context.Context, userID uuid.UUID, paymentMethodID string) (payment.SavedPaymentMethod, error) {
	if s.getSavedPaymentMethod == nil {
		return payment.SavedPaymentMethod{}, payment.ErrSavedPaymentMethodNotFound
	}

	return s.getSavedPaymentMethod(ctx, userID, paymentMethodID)
}

func (s stubPaymentRepository) ListSavedPaymentMethods(ctx context.Context, userID uuid.UUID) ([]payment.SavedPaymentMethod, error) {
	if s.listSavedPaymentMethods == nil {
		return []payment.SavedPaymentMethod{}, nil
	}

	return s.listSavedPaymentMethods(ctx, userID)
}

func (s stubPaymentRepository) TouchSavedPaymentMethodLastUsedAt(ctx context.Context, userID uuid.UUID, paymentMethodID string, lastUsedAt *time.Time) (payment.SavedPaymentMethod, error) {
	if s.touchSavedPaymentMethodLastUsedAt == nil {
		return payment.SavedPaymentMethod{}, nil
	}

	return s.touchSavedPaymentMethodLastUsedAt(ctx, userID, paymentMethodID, lastUsedAt)
}

func (s stubPaymentRepository) UpdateMainPurchaseAttemptOutcome(ctx context.Context, input payment.UpdateMainPurchaseAttemptOutcomeInput) (payment.MainPurchaseAttempt, error) {
	if s.updateMainPurchaseAttemptOutcome == nil {
		return payment.MainPurchaseAttempt{}, nil
	}

	return s.updateMainPurchaseAttemptOutcome(ctx, input)
}

func (s stubTransactionalPaymentRepository) AcquireMainPurchaseLock(ctx context.Context, userID uuid.UUID, mainID uuid.UUID) error {
	if s.acquireMainPurchaseLock == nil {
		return nil
	}

	return s.acquireMainPurchaseLock(ctx, userID, mainID)
}

func (s stubTransactionalPaymentRepository) RunInTx(ctx context.Context, fn func(payment.TxRepository) error) error {
	if s.runInTx == nil {
		return fn(s)
	}

	return s.runInTx(ctx, fn)
}

type stubPurchaseGateway struct {
	charge func(context.Context, payment.ChargeInput) (payment.ChargeResult, error)
}

func (s stubPurchaseGateway) Charge(ctx context.Context, input payment.ChargeInput) (payment.ChargeResult, error) {
	if s.charge == nil {
		return payment.ChargeResult{}, nil
	}

	return s.charge(ctx, input)
}

func TestGetUnlockSurfaceReturnsSetupRequiredPurchaseState(t *testing.T) {
	t.Parallel()

	fixture := newServiceFixture()
	service := fixture.newService(stubPaymentRepository{
		listSavedPaymentMethods: func(context.Context, uuid.UUID) ([]payment.SavedPaymentMethod, error) {
			return []payment.SavedPaymentMethod{}, nil
		},
	})

	surface, err := service.GetUnlockSurface(context.Background(), fixture.viewerID, fixture.sessionBinding, fixture.shortID)
	if err != nil {
		t.Fatalf("GetUnlockSurface() error = %v, want nil", err)
	}
	if surface.Access.Reason != "unlock_required" || surface.Access.Status != "locked" {
		t.Fatalf("GetUnlockSurface() access got %#v want locked/unlock_required", surface.Access)
	}
	if surface.Purchase.State != "setup_required" {
		t.Fatalf("GetUnlockSurface() purchase state got %q want %q", surface.Purchase.State, "setup_required")
	}
	if !surface.Purchase.Setup.Required || !surface.Purchase.Setup.RequiresCardSetup {
		t.Fatalf("GetUnlockSurface() setup got %#v want required card setup", surface.Purchase.Setup)
	}
	if surface.UnlockCta.State != "setup_required" {
		t.Fatalf("GetUnlockSurface() unlock cta got %q want %q", surface.UnlockCta.State, "setup_required")
	}
	if surface.MainAccessToken == "" {
		t.Fatal("GetUnlockSurface() main access token = empty, want value")
	}
}

func TestPurchaseMainSuccessRecordsUnlockAndTouchesSavedCard(t *testing.T) {
	t.Parallel()

	fixture := newServiceFixture()
	entryToken := fixture.entryToken(t)
	attemptID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	recordedUnlocks := 0
	touchedMethods := 0
	outcomes := 0
	savedMethodID := "paymeth_88888888888888888888888888888888"
	processedAt := fixture.now.Add(2 * time.Minute)

	service := fixture.newService(
		stubPaymentRepository{
			getSavedPaymentMethod: func(_ context.Context, gotUserID uuid.UUID, paymentMethodID string) (payment.SavedPaymentMethod, error) {
				if gotUserID != fixture.viewerID || paymentMethodID != savedMethodID {
					t.Fatalf("GetSavedPaymentMethod() got user=%s paymentMethod=%s", gotUserID, paymentMethodID)
				}

				return payment.SavedPaymentMethod{
					ID:                      uuid.MustParse("88888888-8888-8888-8888-888888888888"),
					PaymentMethodID:         savedMethodID,
					ProviderPaymentTokenRef: "saved-token-1",
					Brand:                   payment.CardBrandVisa,
					Last4:                   "4242",
				}, nil
			},
			listSavedPaymentMethods: func(context.Context, uuid.UUID) ([]payment.SavedPaymentMethod, error) {
				return []payment.SavedPaymentMethod{
					{
						PaymentMethodID: savedMethodID,
						Brand:           payment.CardBrandVisa,
						Last4:           "4242",
					},
				}, nil
			},
			createMainPurchaseAttempt: func(_ context.Context, input payment.CreateMainPurchaseAttemptInput) (payment.MainPurchaseAttempt, error) {
				if input.MainID != fixture.mainID || input.UserID != fixture.viewerID {
					t.Fatalf("CreateMainPurchaseAttempt() input got %+v", input)
				}
				if input.PaymentMethodMode != payment.PaymentMethodModeSavedCard || input.ProviderPaymentTokenRef != "saved-token-1" {
					t.Fatalf("CreateMainPurchaseAttempt() payment input got %+v", input)
				}

				return payment.MainPurchaseAttempt{
					ID:     attemptID,
					MainID: fixture.mainID,
					Status: payment.PurchaseAttemptStatusProcessing,
					UserID: fixture.viewerID,
				}, nil
			},
			updateMainPurchaseAttemptOutcome: func(_ context.Context, input payment.UpdateMainPurchaseAttemptOutcomeInput) (payment.MainPurchaseAttempt, error) {
				outcomes++
				if input.ID != attemptID || input.Status != payment.PurchaseAttemptStatusSucceeded {
					t.Fatalf("UpdateMainPurchaseAttemptOutcome() input got %+v", input)
				}

				return payment.MainPurchaseAttempt{
					ID:     attemptID,
					MainID: fixture.mainID,
					Status: input.Status,
				}, nil
			},
			touchSavedPaymentMethodLastUsedAt: func(_ context.Context, gotUserID uuid.UUID, paymentMethodID string, lastUsedAt *time.Time) (payment.SavedPaymentMethod, error) {
				touchedMethods++
				if gotUserID != fixture.viewerID || paymentMethodID != savedMethodID {
					t.Fatalf("TouchSavedPaymentMethodLastUsedAt() got user=%s paymentMethod=%s", gotUserID, paymentMethodID)
				}
				if lastUsedAt == nil || !lastUsedAt.Equal(processedAt) {
					t.Fatalf("TouchSavedPaymentMethodLastUsedAt() lastUsedAt got %#v want %s", lastUsedAt, processedAt)
				}

				return payment.SavedPaymentMethod{PaymentMethodID: paymentMethodID}, nil
			},
		},
		stubPurchaseGateway{
			charge: func(_ context.Context, input payment.ChargeInput) (payment.ChargeResult, error) {
				if input.PaymentTokenRef != "saved-token-1" || input.PriceJPY != fixture.main.PriceMinor || input.IPAddress != "203.0.113.10" {
					t.Fatalf("Charge() input got %+v", input)
				}

				return payment.ChargeResult{
					NewPaymentTokenRef:     stringPtr("saved-token-2"),
					ProviderProcessedAt:    processedAt,
					ProviderPurchaseRef:    stringPtr("purchase-ref-1"),
					ProviderTransactionRef: stringPtr("transaction-ref-1"),
					Status:                 payment.PurchaseAttemptStatusSucceeded,
				}, nil
			},
		},
		stubUnlockRecorder{
			recordMainUnlock: func(_ context.Context, input unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
				recordedUnlocks++
				if input.UserID != fixture.viewerID || input.MainID != fixture.mainID {
					t.Fatalf("RecordMainUnlock() input got %+v", input)
				}
				if input.PurchasedAt == nil || !input.PurchasedAt.Equal(processedAt) {
					t.Fatalf("RecordMainUnlock() purchasedAt got %#v want %s", input.PurchasedAt, processedAt)
				}
				if input.PaymentProviderPurchaseRef == nil || *input.PaymentProviderPurchaseRef != "purchase-ref-1" {
					t.Fatalf("RecordMainUnlock() purchase ref got %#v want purchase-ref-1", input.PaymentProviderPurchaseRef)
				}

				return unlock.MainUnlock{
					UserID:      fixture.viewerID,
					MainID:      fixture.mainID,
					PurchasedAt: processedAt,
					CreatedAt:   processedAt,
				}, nil
			},
		},
	)

	result, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
		EntryToken:  entryToken,
		FromShortID: fixture.shortID,
		MainID:      fixture.mainID,
		PaymentMethod: PurchasePaymentMethodInput{
			Mode:            payment.PaymentMethodModeSavedCard,
			PaymentMethodID: savedMethodID,
		},
		ViewerID: fixture.viewerID,
		ViewerIP: "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("PurchaseMain() error = %v, want nil", err)
	}
	if result.Purchase.Status != "succeeded" {
		t.Fatalf("PurchaseMain() status got %q want %q", result.Purchase.Status, "succeeded")
	}
	if result.Access.Reason != "purchased" || result.Access.Status != "unlocked" {
		t.Fatalf("PurchaseMain() access got %#v want purchased/unlocked", result.Access)
	}
	if result.EntryToken == nil || *result.EntryToken != entryToken {
		t.Fatalf("PurchaseMain() entry token got %#v want %q", result.EntryToken, entryToken)
	}
	if recordedUnlocks != 1 || touchedMethods != 1 || outcomes != 1 {
		t.Fatalf("PurchaseMain() side effects got unlocks=%d touches=%d outcomes=%d", recordedUnlocks, touchedMethods, outcomes)
	}
}

func TestPurchaseMainProviderErrorBecomesPending(t *testing.T) {
	t.Parallel()

	fixture := newServiceFixture()
	entryToken := fixture.entryToken(t)
	cardSetupToken := fixture.cardSetupToken(t)
	attemptID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	outcomes := 0
	recordedUnlocks := 0

	service := fixture.newService(
		stubPaymentRepository{
			createMainPurchaseAttempt: func(context.Context, payment.CreateMainPurchaseAttemptInput) (payment.MainPurchaseAttempt, error) {
				return payment.MainPurchaseAttempt{
					ID:     attemptID,
					MainID: fixture.mainID,
					Status: payment.PurchaseAttemptStatusProcessing,
					UserID: fixture.viewerID,
				}, nil
			},
			updateMainPurchaseAttemptOutcome: func(_ context.Context, input payment.UpdateMainPurchaseAttemptOutcomeInput) (payment.MainPurchaseAttempt, error) {
				outcomes++
				if input.ID != attemptID || input.Status != payment.PurchaseAttemptStatusPending {
					t.Fatalf("UpdateMainPurchaseAttemptOutcome() input got %+v", input)
				}

				return payment.MainPurchaseAttempt{
					ID:     attemptID,
					MainID: fixture.mainID,
					Status: input.Status,
				}, nil
			},
		},
		stubPurchaseGateway{
			charge: func(context.Context, payment.ChargeInput) (payment.ChargeResult, error) {
				return payment.ChargeResult{}, fmt.Errorf("%w: temporary provider failure", payment.ErrChargeOutcomeUnknown)
			},
		},
		stubUnlockRecorder{
			recordMainUnlock: func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
				recordedUnlocks++
				return unlock.MainUnlock{}, nil
			},
		},
	)

	result, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
		AcceptedAge:   true,
		AcceptedTerms: true,
		EntryToken:    entryToken,
		FromShortID:   fixture.shortID,
		MainID:        fixture.mainID,
		PaymentMethod: PurchasePaymentMethodInput{
			Mode:           payment.PaymentMethodModeNewCard,
			CardSetupToken: cardSetupToken,
		},
		ViewerID: fixture.viewerID,
	})
	if err != nil {
		t.Fatalf("PurchaseMain() error = %v, want nil", err)
	}
	if result.Purchase.Status != "pending" {
		t.Fatalf("PurchaseMain() status got %q want %q", result.Purchase.Status, "pending")
	}
	if result.Access.Reason != "unlock_required" || result.Access.Status != "locked" {
		t.Fatalf("PurchaseMain() access got %#v want unlock_required/locked", result.Access)
	}
	if outcomes != 1 || recordedUnlocks != 0 {
		t.Fatalf("PurchaseMain() side effects got outcomes=%d unlocks=%d", outcomes, recordedUnlocks)
	}
}

func TestPurchaseMainChargeFailureReturnsContractFailure(t *testing.T) {
	t.Parallel()

	fixture := newServiceFixture()
	entryToken := fixture.entryToken(t)
	cardSetupToken := fixture.cardSetupToken(t)
	attemptID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	outcomes := 0

	service := fixture.newService(
		stubPaymentRepository{
			createMainPurchaseAttempt: func(context.Context, payment.CreateMainPurchaseAttemptInput) (payment.MainPurchaseAttempt, error) {
				return payment.MainPurchaseAttempt{
					ID:     attemptID,
					MainID: fixture.mainID,
					Status: payment.PurchaseAttemptStatusProcessing,
					UserID: fixture.viewerID,
				}, nil
			},
			updateMainPurchaseAttemptOutcome: func(_ context.Context, input payment.UpdateMainPurchaseAttemptOutcomeInput) (payment.MainPurchaseAttempt, error) {
				outcomes++
				if input.ID != attemptID || input.Status != payment.PurchaseAttemptStatusFailed {
					t.Fatalf("UpdateMainPurchaseAttemptOutcome() input got %+v", input)
				}

				return payment.MainPurchaseAttempt{
					ID:            attemptID,
					MainID:        fixture.mainID,
					Status:        input.Status,
					FailureReason: input.FailureReason,
				}, nil
			},
		},
		stubPurchaseGateway{
			charge: func(context.Context, payment.ChargeInput) (payment.ChargeResult, error) {
				processedAt := fixture.now.Add(2 * time.Minute)
				return payment.ChargeResult{
					CanRetry:            true,
					FailureReason:       stringPtr(payment.FailureReasonPurchaseDeclined),
					ProviderProcessedAt: processedAt,
					Status:              payment.PurchaseAttemptStatusFailed,
				}, nil
			},
		},
	)

	result, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
		AcceptedAge:   true,
		AcceptedTerms: true,
		EntryToken:    entryToken,
		FromShortID:   fixture.shortID,
		MainID:        fixture.mainID,
		PaymentMethod: PurchasePaymentMethodInput{
			Mode:           payment.PaymentMethodModeNewCard,
			CardSetupToken: cardSetupToken,
		},
		ViewerID: fixture.viewerID,
	})
	if err != nil {
		t.Fatalf("PurchaseMain() error = %v, want nil", err)
	}
	if result.Purchase.Status != "failed" || result.Purchase.FailureReason == nil || *result.Purchase.FailureReason != payment.FailureReasonPurchaseDeclined {
		t.Fatalf("PurchaseMain() got %#v want failed/purchase_declined", result)
	}
	if outcomes != 1 {
		t.Fatalf("PurchaseMain() outcomes got %d want %d", outcomes, 1)
	}
}

func TestIssueAccessEntryRejectsViewerWithoutPurchase(t *testing.T) {
	t.Parallel()

	fixture := newServiceFixture()
	service := fixture.newService(stubPaymentRepository{})

	_, err := service.IssueAccessEntry(context.Background(), fixture.sessionBinding, AccessEntryInput{
		EntryToken:  fixture.entryToken(t),
		FromShortID: fixture.shortID,
		MainID:      fixture.mainID,
		ViewerID:    fixture.viewerID,
	})
	if !errors.Is(err, ErrMainLocked) {
		t.Fatalf("IssueAccessEntry() error got %v want %v", err, ErrMainLocked)
	}
}

func TestIssueAccessEntryReturnsPurchasedGrantForUnlockedViewer(t *testing.T) {
	t.Parallel()

	fixture := newServiceFixture()
	fixture.detail.Item.Unlock.IsUnlocked = true
	recordedUnlocks := 0
	service := fixture.newService(
		stubPaymentRepository{
			listSavedPaymentMethods: func(context.Context, uuid.UUID) ([]payment.SavedPaymentMethod, error) {
				return []payment.SavedPaymentMethod{{PaymentMethodID: "paymeth_1", Brand: payment.CardBrandVisa, Last4: "4242"}}, nil
			},
		},
		stubPurchaseGateway{},
		stubUnlockRecorder{
			recordMainUnlock: func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
				recordedUnlocks++
				return unlock.MainUnlock{}, nil
			},
		},
	)

	issued, err := service.IssueAccessEntry(context.Background(), fixture.sessionBinding, AccessEntryInput{
		EntryToken:  fixture.entryToken(t),
		FromShortID: fixture.shortID,
		MainID:      fixture.mainID,
		ViewerID:    fixture.viewerID,
	})
	if err != nil {
		t.Fatalf("IssueAccessEntry() error = %v, want nil", err)
	}
	if issued.GrantKind != MainPlaybackGrantKindPurchased {
		t.Fatalf("IssueAccessEntry() grant kind got %q want %q", issued.GrantKind, MainPlaybackGrantKindPurchased)
	}
	if issued.GrantToken == "" {
		t.Fatal("IssueAccessEntry() grant token = empty, want value")
	}
	if recordedUnlocks != 0 {
		t.Fatalf("IssueAccessEntry() unlock writes got %d want 0", recordedUnlocks)
	}
}

func TestIssueAccessEntryReturnsPurchasedGrantForSucceededAttemptBeforeProjectionCatchUp(t *testing.T) {
	t.Parallel()

	fixture := newServiceFixture()
	service := fixture.newService(stubPaymentRepository{
		getLatestSucceededMainPurchaseAttempt: func(context.Context, uuid.UUID, uuid.UUID) (payment.MainPurchaseAttempt, error) {
			return payment.MainPurchaseAttempt{
				MainID: fixture.mainID,
				Status: payment.PurchaseAttemptStatusSucceeded,
				UserID: fixture.viewerID,
			}, nil
		},
	})

	issued, err := service.IssueAccessEntry(context.Background(), fixture.sessionBinding, AccessEntryInput{
		EntryToken:  fixture.entryToken(t),
		FromShortID: fixture.shortID,
		MainID:      fixture.mainID,
		ViewerID:    fixture.viewerID,
	})
	if err != nil {
		t.Fatalf("IssueAccessEntry() error = %v, want nil", err)
	}
	if issued.GrantKind != MainPlaybackGrantKindPurchased {
		t.Fatalf("IssueAccessEntry() grant kind got %q want %q", issued.GrantKind, MainPlaybackGrantKindPurchased)
	}
	if issued.GrantToken == "" {
		t.Fatal("IssueAccessEntry() grant token = empty, want value")
	}
}

func TestGetPlaybackSurfaceReturnsPurchasedReason(t *testing.T) {
	t.Parallel()

	fixture := newServiceFixture()
	fixture.detail.Item.Unlock.IsUnlocked = true
	service := fixture.newService(stubPaymentRepository{})
	grantToken, err := issueSignedToken(fixture.sessionBinding, fixture.now, defaultGrantTTL, signedTokenPayload{
		GrantKind:   MainPlaybackGrantKindPurchased,
		Kind:        playbackTokenKind,
		MainID:      fixture.mainID,
		FromShortID: fixture.shortID,
		ViewerID:    fixture.viewerID,
	})
	if err != nil {
		t.Fatalf("issueSignedToken() error = %v, want nil", err)
	}

	playback, err := service.GetPlaybackSurface(context.Background(), fixture.viewerID, fixture.sessionBinding, fixture.mainID, fixture.shortID, grantToken)
	if err != nil {
		t.Fatalf("GetPlaybackSurface() error = %v, want nil", err)
	}
	if playback.Access.Reason != "purchased" || playback.Access.Status != "unlocked" {
		t.Fatalf("GetPlaybackSurface() access got %#v want purchased/unlocked", playback.Access)
	}
}

func TestPurchaseMainShortCircuitsForExistingPurchaseAndOwner(t *testing.T) {
	t.Parallel()

	t.Run("already purchased", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		fixture.detail.Item.Unlock.IsUnlocked = true
		service := fixture.newService(
			stubPaymentRepository{
				listSavedPaymentMethods: func(context.Context, uuid.UUID) ([]payment.SavedPaymentMethod, error) {
					return []payment.SavedPaymentMethod{
						{PaymentMethodID: "paymeth_1", Brand: payment.CardBrandVisa, Last4: "4242"},
					}, nil
				},
			},
			stubPurchaseGateway{
				charge: func(context.Context, payment.ChargeInput) (payment.ChargeResult, error) {
					t.Fatal("Charge() was called unexpectedly")
					return payment.ChargeResult{}, nil
				},
			},
		)

		result, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
			EntryToken:  fixture.entryToken(t),
			FromShortID: fixture.shortID,
			MainID:      fixture.mainID,
			PaymentMethod: PurchasePaymentMethodInput{
				Mode:            payment.PaymentMethodModeSavedCard,
				PaymentMethodID: "paymeth_1",
			},
			ViewerID: fixture.viewerID,
		})
		if err != nil {
			t.Fatalf("PurchaseMain() error = %v, want nil", err)
		}
		if result.Purchase.Status != "already_purchased" || result.Access.Reason != "purchased" {
			t.Fatalf("PurchaseMain() got %#v want already_purchased/purchased", result)
		}
	})

	t.Run("owner preview", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		fixture.detail.Item.Unlock.IsOwner = true
		service := fixture.newService()

		result, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
			EntryToken:  fixture.entryToken(t),
			FromShortID: fixture.shortID,
			MainID:      fixture.mainID,
			PaymentMethod: PurchasePaymentMethodInput{
				Mode:           payment.PaymentMethodModeNewCard,
				CardSetupToken: fixture.cardSetupToken(t),
			},
			ViewerID: fixture.viewerID,
		})
		if err != nil {
			t.Fatalf("PurchaseMain() error = %v, want nil", err)
		}
		if result.Purchase.Status != "owner_preview" || result.Access.Reason != "owner_preview" {
			t.Fatalf("PurchaseMain() got %#v want owner_preview", result)
		}
	})
}

func TestPurchaseHelpers(t *testing.T) {
	t.Parallel()

	mainID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	succeeded := buildPurchaseResultFromAttempt(mainID, payment.MainPurchaseAttempt{
		Status: payment.PurchaseAttemptStatusSucceeded,
	}, "entry-token")
	if succeeded.Purchase.Status != "succeeded" || succeeded.Access.Reason != "purchased" {
		t.Fatalf("buildPurchaseResultFromAttempt(succeeded) got %#v", succeeded)
	}

	pending := buildPurchaseResultFromAttempt(mainID, payment.MainPurchaseAttempt{
		Status: payment.PurchaseAttemptStatusPending,
	}, "entry-token")
	if pending.Purchase.Status != "pending" || pending.Access.Reason != "unlock_required" {
		t.Fatalf("buildPurchaseResultFromAttempt(pending) got %#v", pending)
	}

	failed := buildPurchaseResultFromAttempt(mainID, payment.MainPurchaseAttempt{
		Status:        payment.PurchaseAttemptStatusFailed,
		FailureReason: stringPtr(payment.FailureReasonPurchaseDeclined),
	}, "entry-token")
	if failed.Purchase.Status != "failed" || failed.Purchase.FailureReason == nil || *failed.Purchase.FailureReason != payment.FailureReasonPurchaseDeclined {
		t.Fatalf("buildPurchaseResultFromAttempt(failed) got %#v", failed)
	}

	owner := buildOwnerPurchaseResult(mainID, "entry-token")
	if owner.Purchase.Status != "owner_preview" || owner.Access.Reason != "owner_preview" {
		t.Fatalf("buildOwnerPurchaseResult() got %#v", owner)
	}

	alreadyPurchased := buildAlreadyPurchasedResult(mainID, "entry-token")
	if alreadyPurchased.Purchase.Status != "already_purchased" || alreadyPurchased.Access.Reason != "purchased" {
		t.Fatalf("buildAlreadyPurchasedResult() got %#v", alreadyPurchased)
	}
}

func TestUnlockPurchaseStateAndCTAHelpers(t *testing.T) {
	t.Parallel()

	mainID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	basePreview := feed.UnlockPreview{
		MainDurationSeconds: 480,
		PriceJPY:            1800,
	}
	savedMethods := []payment.SavedPaymentMethod{
		{PaymentMethodID: "paymeth_1", Brand: payment.CardBrandVisa, Last4: "4242"},
	}
	pendingReason := payment.PendingReasonProviderProcessing

	tests := []struct {
		name          string
		preview       feed.UnlockPreview
		inflight      *payment.MainPurchaseAttempt
		savedMethods  []payment.SavedPaymentMethod
		wantPurchase  string
		wantUnlockCTA string
		wantAccess    string
		wantGrantKind MainPlaybackGrantKind
	}{
		{
			name:          "owner",
			preview:       feed.UnlockPreview{IsOwner: true},
			savedMethods:  savedMethods,
			wantPurchase:  "owner_preview",
			wantUnlockCTA: "owner_preview",
			wantAccess:    "owner_preview",
			wantGrantKind: MainPlaybackGrantKindOwner,
		},
		{
			name:          "already purchased",
			preview:       feed.UnlockPreview{IsUnlocked: true},
			savedMethods:  savedMethods,
			wantPurchase:  "already_purchased",
			wantUnlockCTA: "continue_main",
			wantAccess:    "purchased",
			wantGrantKind: MainPlaybackGrantKindPurchased,
		},
		{
			name:          "pending",
			preview:       basePreview,
			savedMethods:  savedMethods,
			inflight:      &payment.MainPurchaseAttempt{Status: payment.PurchaseAttemptStatusPending, PendingReason: &pendingReason},
			wantPurchase:  "purchase_pending",
			wantUnlockCTA: "unlock_available",
			wantAccess:    "unlock_required",
			wantGrantKind: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			purchaseState := buildUnlockPurchaseState(tt.preview, tt.savedMethods, tt.inflight)
			if purchaseState.State != tt.wantPurchase {
				t.Fatalf("buildUnlockPurchaseState() state got %q want %q", purchaseState.State, tt.wantPurchase)
			}

			cta := buildUnlockCtaState(tt.preview, purchaseState)
			if cta.State != tt.wantUnlockCTA {
				t.Fatalf("buildUnlockCtaState() state got %q want %q", cta.State, tt.wantUnlockCTA)
			}

			access := buildMainAccessState(tt.preview, mainID)
			if access.Reason != tt.wantAccess {
				t.Fatalf("buildMainAccessState() reason got %q want %q", access.Reason, tt.wantAccess)
			}

			if got := resolveGrantKind(tt.preview); got != tt.wantGrantKind {
				t.Fatalf("resolveGrantKind() got %q want %q", got, tt.wantGrantKind)
			}
		})
	}
}

func TestGetPlaybackSurfaceReturnsOwnerReason(t *testing.T) {
	t.Parallel()

	fixture := newServiceFixture()
	fixture.detail.Item.Unlock.IsOwner = true
	service := fixture.newService()
	grantToken, err := issueSignedToken(fixture.sessionBinding, fixture.now, defaultGrantTTL, signedTokenPayload{
		GrantKind:   MainPlaybackGrantKindOwner,
		Kind:        playbackTokenKind,
		MainID:      fixture.mainID,
		FromShortID: fixture.shortID,
		ViewerID:    fixture.viewerID,
	})
	if err != nil {
		t.Fatalf("issueSignedToken() error = %v, want nil", err)
	}

	playback, err := service.GetPlaybackSurface(context.Background(), fixture.viewerID, fixture.sessionBinding, fixture.mainID, fixture.shortID, grantToken)
	if err != nil {
		t.Fatalf("GetPlaybackSurface() error = %v, want nil", err)
	}
	if playback.Access.Reason != "owner_preview" || playback.Access.Status != "owner" {
		t.Fatalf("GetPlaybackSurface() access got %#v want owner_preview/owner", playback.Access)
	}
}

func TestServiceErrorMappings(t *testing.T) {
	t.Parallel()

	t.Run("unlock surface short not found", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		service := NewService(
			stubFeedReader{
				getDetail: func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error) {
					return feed.Detail{}, feed.ErrPublicShortNotFound
				},
			},
			stubMainReader{
				getUnlockableMain: func(context.Context, uuid.UUID) (shorts.Main, error) {
					t.Fatal("GetUnlockableMain() was called unexpectedly")
					return shorts.Main{}, nil
				},
			},
			stubUnlockRecorder{},
			stubPaymentRepository{},
			stubPurchaseGateway{},
		)

		_, err := service.GetUnlockSurface(context.Background(), fixture.viewerID, fixture.sessionBinding, fixture.shortID)
		if !errors.Is(err, ErrShortUnlockNotFound) {
			t.Fatalf("GetUnlockSurface() error got %v want %v", err, ErrShortUnlockNotFound)
		}
	})

	t.Run("unlock surface main locked", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		service := NewService(
			stubFeedReader{
				getDetail: func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error) {
					return fixture.detail, nil
				},
			},
			stubMainReader{
				getUnlockableMain: func(context.Context, uuid.UUID) (shorts.Main, error) {
					return shorts.Main{}, shorts.ErrUnlockableMainNotFound
				},
			},
			stubUnlockRecorder{},
			stubPaymentRepository{},
			stubPurchaseGateway{},
		)

		_, err := service.GetUnlockSurface(context.Background(), fixture.viewerID, fixture.sessionBinding, fixture.shortID)
		if !errors.Is(err, ErrMainLocked) {
			t.Fatalf("GetUnlockSurface() error got %v want %v", err, ErrMainLocked)
		}
	})

	t.Run("issue access entry main mismatch", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		fixture.detail.Item.Unlock.IsUnlocked = true
		service := fixture.newService()

		_, err := service.IssueAccessEntry(context.Background(), fixture.sessionBinding, AccessEntryInput{
			EntryToken:  fixture.entryToken(t),
			FromShortID: fixture.shortID,
			MainID:      uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
			ViewerID:    fixture.viewerID,
		})
		if !errors.Is(err, ErrAccessEntryNotFound) {
			t.Fatalf("IssueAccessEntry() error got %v want %v", err, ErrAccessEntryNotFound)
		}
	})

	t.Run("playback invalid grant", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		service := fixture.newService()

		_, err := service.GetPlaybackSurface(context.Background(), fixture.viewerID, fixture.sessionBinding, fixture.mainID, fixture.shortID, "invalid-grant")
		if !errors.Is(err, ErrMainLocked) {
			t.Fatalf("GetPlaybackSurface() error got %v want %v", err, ErrMainLocked)
		}
	})

	t.Run("purchase invalid request", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		service := fixture.newService()

		_, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
			EntryToken:  fixture.entryToken(t),
			FromShortID: fixture.shortID,
			MainID:      fixture.mainID,
			PaymentMethod: PurchasePaymentMethodInput{
				Mode: payment.PaymentMethodModeSavedCard,
			},
			ViewerID: fixture.viewerID,
		})
		if !errors.Is(err, ErrInvalidPurchaseRequest) {
			t.Fatalf("PurchaseMain() error got %v want %v", err, ErrInvalidPurchaseRequest)
		}
	})

	t.Run("purchase rejects raw card setup token", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		service := fixture.newService()

		_, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
			AcceptedAge:   true,
			AcceptedTerms: true,
			EntryToken:    fixture.entryToken(t),
			FromShortID:   fixture.shortID,
			MainID:        fixture.mainID,
			PaymentMethod: PurchasePaymentMethodInput{
				Mode:           payment.PaymentMethodModeNewCard,
				CardSetupToken: "new-card-token",
			},
			ViewerID: fixture.viewerID,
		})
		if !errors.Is(err, ErrInvalidPurchaseRequest) {
			t.Fatalf("PurchaseMain() error got %v want %v", err, ErrInvalidPurchaseRequest)
		}
	})

	t.Run("purchase create conflict resolves existing attempt", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		service := fixture.newService(stubPaymentRepository{
			createMainPurchaseAttempt: func(context.Context, payment.CreateMainPurchaseAttemptInput) (payment.MainPurchaseAttempt, error) {
				return payment.MainPurchaseAttempt{}, payment.ErrMainPurchaseAttemptConflict
			},
			getMainPurchaseAttemptByIdempotencyKey: func(context.Context, string) (payment.MainPurchaseAttempt, error) {
				return payment.MainPurchaseAttempt{
					MainID: fixture.mainID,
					Status: payment.PurchaseAttemptStatusPending,
				}, nil
			},
		})

		result, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
			AcceptedAge:   true,
			AcceptedTerms: true,
			EntryToken:    fixture.entryToken(t),
			FromShortID:   fixture.shortID,
			MainID:        fixture.mainID,
			PaymentMethod: PurchasePaymentMethodInput{
				Mode:           payment.PaymentMethodModeNewCard,
				CardSetupToken: fixture.cardSetupToken(t),
			},
			ViewerID: fixture.viewerID,
		})
		if err != nil {
			t.Fatalf("PurchaseMain() error = %v, want nil", err)
		}
		if result.Purchase.Status != "pending" {
			t.Fatalf("PurchaseMain() status got %q want %q", result.Purchase.Status, "pending")
		}
	})

	t.Run("purchase known charge error clears inflight attempt", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		attemptID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
		outcomes := 0
		service := fixture.newService(
			stubPaymentRepository{
				createMainPurchaseAttempt: func(context.Context, payment.CreateMainPurchaseAttemptInput) (payment.MainPurchaseAttempt, error) {
					return payment.MainPurchaseAttempt{
						ID:     attemptID,
						MainID: fixture.mainID,
						Status: payment.PurchaseAttemptStatusProcessing,
						UserID: fixture.viewerID,
					}, nil
				},
				updateMainPurchaseAttemptOutcome: func(_ context.Context, input payment.UpdateMainPurchaseAttemptOutcomeInput) (payment.MainPurchaseAttempt, error) {
					outcomes++
					if input.ID != attemptID || input.Status != payment.PurchaseAttemptStatusFailed {
						t.Fatalf("UpdateMainPurchaseAttemptOutcome() input got %+v", input)
					}
					if input.FailureReason == nil || *input.FailureReason != payment.FailureReasonPurchaseDeclined {
						t.Fatalf("UpdateMainPurchaseAttemptOutcome() failure reason got %#v want purchase_declined", input.FailureReason)
					}

					return payment.MainPurchaseAttempt{
						ID:     attemptID,
						MainID: fixture.mainID,
						Status: input.Status,
					}, nil
				},
			},
			stubPurchaseGateway{
				charge: func(context.Context, payment.ChargeInput) (payment.ChargeResult, error) {
					return payment.ChargeResult{}, errors.New("oauth unavailable")
				},
			},
		)

		_, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
			AcceptedAge:   true,
			AcceptedTerms: true,
			EntryToken:    fixture.entryToken(t),
			FromShortID:   fixture.shortID,
			MainID:        fixture.mainID,
			PaymentMethod: PurchasePaymentMethodInput{
				Mode:           payment.PaymentMethodModeNewCard,
				CardSetupToken: fixture.cardSetupToken(t),
			},
			ViewerID: fixture.viewerID,
		})
		if err == nil {
			t.Fatal("PurchaseMain() error = nil, want propagated internal error")
		}
		if outcomes != 1 {
			t.Fatalf("PurchaseMain() outcomes got %d want %d", outcomes, 1)
		}
	})
}

func TestPurchaseMainIdempotencyAndInflightShortCircuits(t *testing.T) {
	t.Parallel()

	t.Run("existing idempotent attempt", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		cardSetupToken := fixture.cardSetupToken(t)
		service := fixture.newService(stubPaymentRepository{
			getMainPurchaseAttemptByIdempotencyKey: func(context.Context, string) (payment.MainPurchaseAttempt, error) {
				return payment.MainPurchaseAttempt{
					MainID: fixture.mainID,
					Status: payment.PurchaseAttemptStatusSucceeded,
				}, nil
			},
		})

		result, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
			AcceptedAge:   true,
			AcceptedTerms: true,
			EntryToken:    fixture.entryToken(t),
			FromShortID:   fixture.shortID,
			MainID:        fixture.mainID,
			PaymentMethod: PurchasePaymentMethodInput{
				Mode:           payment.PaymentMethodModeNewCard,
				CardSetupToken: cardSetupToken,
			},
			ViewerID: fixture.viewerID,
		})
		if err != nil {
			t.Fatalf("PurchaseMain() error = %v, want nil", err)
		}
		if result.Purchase.Status != "succeeded" || result.Access.Reason != "purchased" {
			t.Fatalf("PurchaseMain() got %#v want succeeded/purchased", result)
		}
	})

	t.Run("inflight attempt", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		cardSetupToken := fixture.cardSetupToken(t)
		pendingReason := payment.PendingReasonProviderProcessing
		service := fixture.newService(stubPaymentRepository{
			getLatestInflightMainPurchaseAttempt: func(context.Context, uuid.UUID, uuid.UUID) (payment.MainPurchaseAttempt, error) {
				return payment.MainPurchaseAttempt{
					MainID:        fixture.mainID,
					Status:        payment.PurchaseAttemptStatusPending,
					PendingReason: &pendingReason,
				}, nil
			},
		})

		result, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
			AcceptedAge:   true,
			AcceptedTerms: true,
			EntryToken:    fixture.entryToken(t),
			FromShortID:   fixture.shortID,
			MainID:        fixture.mainID,
			PaymentMethod: PurchasePaymentMethodInput{
				Mode:           payment.PaymentMethodModeNewCard,
				CardSetupToken: cardSetupToken,
			},
			ViewerID: fixture.viewerID,
		})
		if err != nil {
			t.Fatalf("PurchaseMain() error = %v, want nil", err)
		}
		if result.Purchase.Status != "pending" || result.Access.Reason != "unlock_required" {
			t.Fatalf("PurchaseMain() got %#v want pending/locked", result)
		}
	})

	t.Run("completed success attempt short-circuits stale unlock reads", func(t *testing.T) {
		t.Parallel()

		fixture := newServiceFixture()
		cardSetupToken := fixture.cardSetupToken(t)
		service := fixture.newService(stubPaymentRepository{
			getLatestSucceededMainPurchaseAttempt: func(context.Context, uuid.UUID, uuid.UUID) (payment.MainPurchaseAttempt, error) {
				return payment.MainPurchaseAttempt{
					MainID: fixture.mainID,
					Status: payment.PurchaseAttemptStatusSucceeded,
				}, nil
			},
		})

		result, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
			AcceptedAge:   true,
			AcceptedTerms: true,
			EntryToken:    fixture.entryToken(t),
			FromShortID:   fixture.shortID,
			MainID:        fixture.mainID,
			PaymentMethod: PurchasePaymentMethodInput{
				Mode:           payment.PaymentMethodModeNewCard,
				CardSetupToken: cardSetupToken,
			},
			ViewerID: fixture.viewerID,
		})
		if err != nil {
			t.Fatalf("PurchaseMain() error = %v, want nil", err)
		}
		if result.Purchase.Status != "succeeded" || result.Access.Reason != "purchased" {
			t.Fatalf("PurchaseMain() got %#v want succeeded/purchased", result)
		}
	})
}

func TestPurchaseMainTransactionalRepositoryAcquiresLockBeforeAttemptChecks(t *testing.T) {
	t.Parallel()

	fixture := newServiceFixture()
	cardSetupToken := fixture.cardSetupToken(t)
	attemptID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	processedAt := fixture.now.Add(time.Minute)
	order := make([]string, 0, 8)

	var repo stubTransactionalPaymentRepository
	repo = stubTransactionalPaymentRepository{
		stubPaymentRepository: stubPaymentRepository{
			getLatestInflightMainPurchaseAttempt: func(context.Context, uuid.UUID, uuid.UUID) (payment.MainPurchaseAttempt, error) {
				order = append(order, "inflight")
				return payment.MainPurchaseAttempt{}, payment.ErrMainPurchaseAttemptNotFound
			},
			getLatestSucceededMainPurchaseAttempt: func(context.Context, uuid.UUID, uuid.UUID) (payment.MainPurchaseAttempt, error) {
				order = append(order, "succeeded")
				return payment.MainPurchaseAttempt{}, payment.ErrMainPurchaseAttemptNotFound
			},
			getMainPurchaseAttemptByIdempotencyKey: func(context.Context, string) (payment.MainPurchaseAttempt, error) {
				order = append(order, "idempotency")
				return payment.MainPurchaseAttempt{}, payment.ErrMainPurchaseAttemptNotFound
			},
			createMainPurchaseAttempt: func(_ context.Context, input payment.CreateMainPurchaseAttemptInput) (payment.MainPurchaseAttempt, error) {
				order = append(order, "create")
				if input.MainID != fixture.mainID || input.UserID != fixture.viewerID {
					t.Fatalf("CreateMainPurchaseAttempt() input got %+v", input)
				}

				return payment.MainPurchaseAttempt{
					ID:     attemptID,
					MainID: fixture.mainID,
					Status: payment.PurchaseAttemptStatusProcessing,
					UserID: fixture.viewerID,
				}, nil
			},
			updateMainPurchaseAttemptOutcome: func(_ context.Context, input payment.UpdateMainPurchaseAttemptOutcomeInput) (payment.MainPurchaseAttempt, error) {
				order = append(order, "update")
				if input.ID != attemptID || input.Status != payment.PurchaseAttemptStatusSucceeded {
					t.Fatalf("UpdateMainPurchaseAttemptOutcome() input got %+v", input)
				}

				return payment.MainPurchaseAttempt{
					ID:     attemptID,
					MainID: fixture.mainID,
					Status: input.Status,
				}, nil
			},
		},
		acquireMainPurchaseLock: func(_ context.Context, userID uuid.UUID, mainID uuid.UUID) error {
			order = append(order, "lock")
			if userID != fixture.viewerID || mainID != fixture.mainID {
				t.Fatalf("AcquireMainPurchaseLock() got user=%s main=%s", userID, mainID)
			}

			return nil
		},
		runInTx: func(ctx context.Context, fn func(payment.TxRepository) error) error {
			order = append(order, "tx")
			return fn(repo)
		},
	}

	service := fixture.newService(
		repo,
		stubPurchaseGateway{
			charge: func(_ context.Context, input payment.ChargeInput) (payment.ChargeResult, error) {
				order = append(order, "charge")
				if input.AttemptID != attemptID {
					t.Fatalf("Charge() attempt id got %s want %s", input.AttemptID, attemptID)
				}

				return payment.ChargeResult{
					ProviderProcessedAt: processedAt,
					Status:              payment.PurchaseAttemptStatusSucceeded,
				}, nil
			},
		},
	)

	result, err := service.PurchaseMain(context.Background(), fixture.sessionBinding, PurchaseInput{
		AcceptedAge:   true,
		AcceptedTerms: true,
		EntryToken:    fixture.entryToken(t),
		FromShortID:   fixture.shortID,
		MainID:        fixture.mainID,
		PaymentMethod: PurchasePaymentMethodInput{
			Mode:           payment.PaymentMethodModeNewCard,
			CardSetupToken: cardSetupToken,
		},
		ViewerID: fixture.viewerID,
	})
	if err != nil {
		t.Fatalf("PurchaseMain() error = %v, want nil", err)
	}
	if result.Purchase.Status != "succeeded" || result.Access.Reason != "purchased" {
		t.Fatalf("PurchaseMain() got %#v want succeeded/purchased", result)
	}

	wantOrder := []string{"tx", "lock", "inflight", "succeeded", "idempotency", "create", "charge", "update"}
	if fmt.Sprint(order) != fmt.Sprint(wantOrder) {
		t.Fatalf("PurchaseMain() order got %v want %v", order, wantOrder)
	}
}

func TestCurrencyNumericCode(t *testing.T) {
	t.Parallel()

	if code, err := currencyNumericCode("JPY"); err != nil || code != 392 {
		t.Fatalf("currencyNumericCode(JPY) got code=%d err=%v want 392 nil", code, err)
	}
	if _, err := currencyNumericCode("USD"); err == nil {
		t.Fatal("currencyNumericCode(USD) error = nil, want error")
	}
}

type serviceFixture struct {
	detail         feed.Detail
	main           shorts.Main
	mainID         uuid.UUID
	now            time.Time
	sessionBinding string
	shortID        uuid.UUID
	viewerID       uuid.UUID
}

func newServiceFixture() serviceFixture {
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	mainAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	now := time.Unix(1_710_000_000, 0).UTC()

	return serviceFixture{
		detail: feed.Detail{
			Item: feed.Item{
				Creator: feed.CreatorSummary{
					Bio:         "quiet rooftop specialist",
					DisplayName: "Mina Rei",
					Handle:      "minarei",
					ID:          viewerID,
				},
				Short: feed.ShortSummary{
					Caption:                "quiet rooftop preview",
					CanonicalMainID:        mainID,
					CreatorUserID:          viewerID,
					ID:                     shortID,
					MediaAssetID:           shortAssetID,
					PreviewDurationSeconds: 16,
				},
				Unlock: feed.UnlockPreview{
					IsOwner:             false,
					IsUnlocked:          false,
					MainDurationSeconds: 480,
					PriceJPY:            1800,
				},
			},
		},
		main: shorts.Main{
			ID:            mainID,
			CreatorUserID: viewerID,
			MediaAssetID:  mainAssetID,
			PriceMinor:    1800,
			CurrencyCode:  "JPY",
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		mainID:         mainID,
		now:            now,
		sessionBinding: "session-hash",
		shortID:        shortID,
		viewerID:       viewerID,
	}
}

func (f serviceFixture) newService(args ...any) *Service {
	var paymentRepository paymentRepository = stubPaymentRepository{}
	purchaseGateway := stubPurchaseGateway{}
	unlockRecorder := stubUnlockRecorder{}

	for _, arg := range args {
		switch value := arg.(type) {
		case stubPaymentRepository:
			paymentRepository = value
		case stubTransactionalPaymentRepository:
			paymentRepository = value
		case stubPurchaseGateway:
			purchaseGateway = value
		case stubUnlockRecorder:
			unlockRecorder = value
		}
	}

	service := NewService(
		stubFeedReader{
			getDetail: func(_ context.Context, gotShortID uuid.UUID, gotViewerID *uuid.UUID) (feed.Detail, error) {
				if gotShortID != f.shortID {
					return feed.Detail{}, errors.New("unexpected short id")
				}
				if gotViewerID == nil || *gotViewerID != f.viewerID {
					return feed.Detail{}, errors.New("unexpected viewer id")
				}

				return f.detail, nil
			},
		},
		stubMainReader{
			getUnlockableMain: func(_ context.Context, gotMainID uuid.UUID) (shorts.Main, error) {
				if gotMainID != f.mainID {
					return shorts.Main{}, errors.New("unexpected main id")
				}

				return f.main, nil
			},
		},
		unlockRecorder,
		paymentRepository,
		purchaseGateway,
	)
	service.now = func() time.Time { return f.now }

	return service
}

func (f serviceFixture) entryToken(t *testing.T) string {
	t.Helper()

	token, err := issueSignedToken(f.sessionBinding, f.now, defaultTokenTTL, signedTokenPayload{
		Kind:        entryTokenKind,
		MainID:      f.mainID,
		FromShortID: f.shortID,
		ViewerID:    f.viewerID,
	})
	if err != nil {
		t.Fatalf("issueSignedToken() error = %v, want nil", err)
	}

	return token
}

func (f serviceFixture) cardSetupToken(t *testing.T) string {
	t.Helper()

	token, err := issueSignedCardSetupToken(
		f.sessionBinding,
		f.now,
		defaultTokenTTL,
		f.viewerID,
		payment.ProviderCCBill,
		"new-card-token",
	)
	if err != nil {
		t.Fatalf("issueSignedCardSetupToken() error = %v, want nil", err)
	}

	return token
}
