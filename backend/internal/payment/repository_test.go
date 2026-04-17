package payment

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repositoryStubQueries struct {
	acquireMainPurchaseLock func(context.Context, sqlc.AcquireMainPurchaseLockParams) error
	createAttempt           func(context.Context, sqlc.CreateMainPurchaseAttemptParams) (sqlc.AppMainPurchaseAttempt, error)
	getAttempt              func(context.Context, pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error)
	getAttemptForUpdate     func(context.Context, pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error)
	getAttemptByKey         func(context.Context, string) (sqlc.AppMainPurchaseAttempt, error)
	getInflightAttempt      func(context.Context, sqlc.GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error)
	getSucceededAttempt     func(context.Context, sqlc.GetLatestSucceededMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error)
	getAttemptByProvider    func(context.Context, pgtype.Text) (sqlc.AppMainPurchaseAttempt, error)
	getAttemptByTxn         func(context.Context, pgtype.Text) (sqlc.AppMainPurchaseAttempt, error)
	updateAttemptOutcome    func(context.Context, sqlc.UpdateMainPurchaseAttemptOutcomeParams) (sqlc.AppMainPurchaseAttempt, error)
	getPaymentMethod        func(context.Context, sqlc.GetUserPaymentMethodByIDAndUserIDParams) (sqlc.AppUserPaymentMethod, error)
	listPaymentMethods      func(context.Context, pgtype.UUID) ([]sqlc.AppUserPaymentMethod, error)
	touchPaymentMethod      func(context.Context, sqlc.TouchUserPaymentMethodLastUsedAtParams) (sqlc.AppUserPaymentMethod, error)
	upsertPaymentMethod     func(context.Context, sqlc.UpsertUserPaymentMethodParams) (sqlc.AppUserPaymentMethod, error)
}

func (s repositoryStubQueries) AcquireMainPurchaseLock(ctx context.Context, arg sqlc.AcquireMainPurchaseLockParams) error {
	if s.acquireMainPurchaseLock == nil {
		return nil
	}

	return s.acquireMainPurchaseLock(ctx, arg)
}

func (s repositoryStubQueries) CreateMainPurchaseAttempt(ctx context.Context, arg sqlc.CreateMainPurchaseAttemptParams) (sqlc.AppMainPurchaseAttempt, error) {
	return s.createAttempt(ctx, arg)
}

func (s repositoryStubQueries) GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdate(ctx context.Context, arg sqlc.GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error) {
	return s.getInflightAttempt(ctx, arg)
}

func (s repositoryStubQueries) GetLatestSucceededMainPurchaseAttemptByUserIDAndMainIDForUpdate(ctx context.Context, arg sqlc.GetLatestSucceededMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error) {
	return s.getSucceededAttempt(ctx, arg)
}

func (s repositoryStubQueries) GetMainPurchaseAttemptByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
	return s.getAttempt(ctx, id)
}

func (s repositoryStubQueries) GetMainPurchaseAttemptByIDForUpdate(ctx context.Context, id pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
	return s.getAttemptForUpdate(ctx, id)
}

func (s repositoryStubQueries) GetMainPurchaseAttemptByIdempotencyKeyForUpdate(ctx context.Context, idempotencyKey string) (sqlc.AppMainPurchaseAttempt, error) {
	return s.getAttemptByKey(ctx, idempotencyKey)
}

func (s repositoryStubQueries) GetMainPurchaseAttemptByProviderPurchaseRefForUpdate(ctx context.Context, providerPurchaseRef pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
	return s.getAttemptByProvider(ctx, providerPurchaseRef)
}

func (s repositoryStubQueries) GetMainPurchaseAttemptByProviderTransactionRefForUpdate(ctx context.Context, providerTransactionRef pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
	return s.getAttemptByTxn(ctx, providerTransactionRef)
}

func (s repositoryStubQueries) GetUserPaymentMethodByIDAndUserID(ctx context.Context, arg sqlc.GetUserPaymentMethodByIDAndUserIDParams) (sqlc.AppUserPaymentMethod, error) {
	return s.getPaymentMethod(ctx, arg)
}

func (s repositoryStubQueries) ListUserPaymentMethodsByUserID(ctx context.Context, userID pgtype.UUID) ([]sqlc.AppUserPaymentMethod, error) {
	return s.listPaymentMethods(ctx, userID)
}

func (s repositoryStubQueries) TouchUserPaymentMethodLastUsedAt(ctx context.Context, arg sqlc.TouchUserPaymentMethodLastUsedAtParams) (sqlc.AppUserPaymentMethod, error) {
	return s.touchPaymentMethod(ctx, arg)
}

func (s repositoryStubQueries) UpdateMainPurchaseAttemptOutcome(ctx context.Context, arg sqlc.UpdateMainPurchaseAttemptOutcomeParams) (sqlc.AppMainPurchaseAttempt, error) {
	return s.updateAttemptOutcome(ctx, arg)
}

func (s repositoryStubQueries) UpsertUserPaymentMethod(ctx context.Context, arg sqlc.UpsertUserPaymentMethodParams) (sqlc.AppUserPaymentMethod, error) {
	return s.upsertPaymentMethod(ctx, arg)
}

type txBeginnerStub struct {
	begin func(context.Context) (pgx.Tx, error)
}

func (s txBeginnerStub) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.begin(ctx)
}

type txStub struct {
	commitErr   error
	rollbackErr error
	committed   bool
	rolledBack  bool
}

func (tx *txStub) Begin(context.Context) (pgx.Tx, error) { return tx, nil }
func (tx *txStub) Commit(context.Context) error {
	tx.committed = true
	return tx.commitErr
}
func (tx *txStub) Rollback(context.Context) error {
	tx.rolledBack = true
	return tx.rollbackErr
}
func (tx *txStub) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (tx *txStub) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (tx *txStub) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (tx *txStub) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (tx *txStub) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (tx *txStub) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (tx *txStub) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (tx *txStub) Conn() *pgx.Conn                                         { return nil }

func TestRepositorySavedPaymentMethods(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_710_000_000, 0).UTC()
	userID := uuid.New()
	methodID := uuid.New()
	row := testPaymentMethodRow(methodID, userID, now)

	var getArg sqlc.GetUserPaymentMethodByIDAndUserIDParams
	var touchArg sqlc.TouchUserPaymentMethodLastUsedAtParams
	var upsertArg sqlc.UpsertUserPaymentMethodParams

	repo := newRepository(repositoryStubQueries{
		getPaymentMethod: func(_ context.Context, arg sqlc.GetUserPaymentMethodByIDAndUserIDParams) (sqlc.AppUserPaymentMethod, error) {
			getArg = arg
			return row, nil
		},
		listPaymentMethods: func(_ context.Context, gotUserID pgtype.UUID) ([]sqlc.AppUserPaymentMethod, error) {
			if gotUserID != pgUUID(userID) {
				t.Fatalf("ListUserPaymentMethodsByUserID() user got %v want %v", gotUserID, pgUUID(userID))
			}
			return []sqlc.AppUserPaymentMethod{row}, nil
		},
		touchPaymentMethod: func(_ context.Context, arg sqlc.TouchUserPaymentMethodLastUsedAtParams) (sqlc.AppUserPaymentMethod, error) {
			touchArg = arg
			return row, nil
		},
		upsertPaymentMethod: func(_ context.Context, arg sqlc.UpsertUserPaymentMethodParams) (sqlc.AppUserPaymentMethod, error) {
			upsertArg = arg
			return row, nil
		},
		createAttempt: func(context.Context, sqlc.CreateMainPurchaseAttemptParams) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, nil
		},
		getAttempt: func(context.Context, pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, nil
		},
		getAttemptForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, nil
		},
		getAttemptByKey: func(context.Context, string) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, nil
		},
		getInflightAttempt: func(context.Context, sqlc.GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, nil
		},
		getAttemptByProvider: func(context.Context, pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, nil
		},
		updateAttemptOutcome: func(context.Context, sqlc.UpdateMainPurchaseAttemptOutcomeParams) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, nil
		},
	})

	methods, err := repo.ListSavedPaymentMethods(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListSavedPaymentMethods() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(methods, []SavedPaymentMethod{wantPaymentMethod(methodID, userID, now)}) {
		t.Fatalf("ListSavedPaymentMethods() got %#v want %#v", methods, []SavedPaymentMethod{wantPaymentMethod(methodID, userID, now)})
	}

	method, err := repo.GetSavedPaymentMethod(context.Background(), userID, FormatPublicPaymentMethodID(methodID))
	if err != nil {
		t.Fatalf("GetSavedPaymentMethod() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(method, wantPaymentMethod(methodID, userID, now)) {
		t.Fatalf("GetSavedPaymentMethod() got %#v want %#v", method, wantPaymentMethod(methodID, userID, now))
	}
	if getArg.ID != pgUUID(methodID) || getArg.UserID != pgUUID(userID) {
		t.Fatalf("GetSavedPaymentMethod() args got %#v", getArg)
	}

	method, err = repo.UpsertSavedPaymentMethod(context.Background(), UpsertSavedPaymentMethodInput{
		Brand:                     CardBrandVisa,
		Last4:                     "4242",
		LastUsedAt:                timePtr(now),
		Provider:                  ProviderCCBill,
		ProviderPaymentAccountRef: "acct-1",
		ProviderPaymentTokenRef:   "token-1",
		UserID:                    userID,
	})
	if err != nil {
		t.Fatalf("UpsertSavedPaymentMethod() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(method, wantPaymentMethod(methodID, userID, now)) {
		t.Fatalf("UpsertSavedPaymentMethod() got %#v want %#v", method, wantPaymentMethod(methodID, userID, now))
	}
	if upsertArg.UserID != pgUUID(userID) || upsertArg.Provider != ProviderCCBill || upsertArg.Brand != CardBrandVisa {
		t.Fatalf("UpsertSavedPaymentMethod() args got %#v", upsertArg)
	}

	method, err = repo.TouchSavedPaymentMethodLastUsedAt(context.Background(), userID, FormatPublicPaymentMethodID(methodID), timePtr(now))
	if err != nil {
		t.Fatalf("TouchSavedPaymentMethodLastUsedAt() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(method, wantPaymentMethod(methodID, userID, now)) {
		t.Fatalf("TouchSavedPaymentMethodLastUsedAt() got %#v want %#v", method, wantPaymentMethod(methodID, userID, now))
	}
	if touchArg.ID != pgUUID(methodID) || touchArg.UserID != pgUUID(userID) {
		t.Fatalf("TouchSavedPaymentMethodLastUsedAt() args got %#v", touchArg)
	}
}

func TestNewRepositoryAndRunInTx(t *testing.T) {
	t.Parallel()

	t.Run("new repository nil pool", func(t *testing.T) {
		t.Parallel()

		repo := NewRepository((*pgxpool.Pool)(nil))
		if repo == nil {
			t.Fatal("NewRepository(nil) = nil, want zero repository")
		}
		if repo.txBeginner != nil || repo.queries != nil || repo.newQueries != nil {
			t.Fatalf("NewRepository(nil) got %#v want zero fields", repo)
		}
	})

	t.Run("run in tx validates inputs", func(t *testing.T) {
		t.Parallel()

		repo := &Repository{}
		if err := repo.RunInTx(context.Background(), func(TxRepository) error { return nil }); err == nil {
			t.Fatal("RunInTx() error = nil, want uninitialized repository")
		}

		repo = &Repository{
			txBeginner: txBeginnerStub{
				begin: func(context.Context) (pgx.Tx, error) { return &txStub{}, nil },
			},
			newQueries: func(sqlc.DBTX) queries { return repositoryStubQueries{} },
		}
		if err := repo.RunInTx(context.Background(), nil); err == nil {
			t.Fatal("RunInTx(nil callback) error = nil, want validation error")
		}
	})

	t.Run("run in tx delegates to tx scoped repository", func(t *testing.T) {
		t.Parallel()

		tx := &txStub{}
		userID := uuid.New()
		mainID := uuid.New()
		capturedTx := sqlc.DBTX(nil)
		lockCalls := 0
		txQueries := repositoryStubQueries{
			acquireMainPurchaseLock: func(_ context.Context, arg sqlc.AcquireMainPurchaseLockParams) error {
				lockCalls++
				if arg.UserKey != userID.String() || arg.MainKey != mainID.String() {
					t.Fatalf("AcquireMainPurchaseLock() args got %#v", arg)
				}
				return nil
			},
		}

		repo := &Repository{
			txBeginner: txBeginnerStub{
				begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
			},
			newQueries: func(db sqlc.DBTX) queries {
				capturedTx = db
				return txQueries
			},
		}

		err := repo.RunInTx(context.Background(), func(txRepo TxRepository) error {
			return txRepo.AcquireMainPurchaseLock(context.Background(), userID, mainID)
		})
		if err != nil {
			t.Fatalf("RunInTx() error = %v, want nil", err)
		}
		if capturedTx != tx {
			t.Fatalf("RunInTx() tx got %v want %v", capturedTx, tx)
		}
		if lockCalls != 1 {
			t.Fatalf("AcquireMainPurchaseLock() calls got %d want %d", lockCalls, 1)
		}
		if !tx.committed || tx.rolledBack {
			t.Fatalf("RunInTx() commit state got committed=%t rolledBack=%t", tx.committed, tx.rolledBack)
		}
	})
}

func TestRepositoryPurchaseAttempts(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_710_000_000, 0).UTC()
	userID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()
	attemptID := uuid.New()
	methodID := uuid.New()
	row := testMainPurchaseAttemptRow(attemptID, userID, mainID, shortID, &methodID, now)

	var createArg sqlc.CreateMainPurchaseAttemptParams
	var lockArg sqlc.AcquireMainPurchaseLockParams
	var updateArg sqlc.UpdateMainPurchaseAttemptOutcomeParams

	repo := newRepository(repositoryStubQueries{
		acquireMainPurchaseLock: func(_ context.Context, arg sqlc.AcquireMainPurchaseLockParams) error {
			lockArg = arg
			return nil
		},
		createAttempt: func(_ context.Context, arg sqlc.CreateMainPurchaseAttemptParams) (sqlc.AppMainPurchaseAttempt, error) {
			createArg = arg
			return row, nil
		},
		getAttempt: func(_ context.Context, gotID pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
			if gotID != pgUUID(attemptID) {
				t.Fatalf("GetMainPurchaseAttemptByID() id got %v want %v", gotID, pgUUID(attemptID))
			}
			return row, nil
		},
		getAttemptForUpdate: func(_ context.Context, gotID pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
			if gotID != pgUUID(attemptID) {
				t.Fatalf("GetMainPurchaseAttemptByIDForUpdate() id got %v want %v", gotID, pgUUID(attemptID))
			}
			return row, nil
		},
		getAttemptByKey: func(_ context.Context, gotKey string) (sqlc.AppMainPurchaseAttempt, error) {
			if gotKey != "request-key" {
				t.Fatalf("GetMainPurchaseAttemptByIdempotencyKeyForUpdate() key got %q want %q", gotKey, "request-key")
			}
			return row, nil
		},
		getInflightAttempt: func(_ context.Context, arg sqlc.GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error) {
			if arg.UserID != pgUUID(userID) || arg.MainID != pgUUID(mainID) {
				t.Fatalf("GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdate() args got %#v", arg)
			}
			return row, nil
		},
		getSucceededAttempt: func(_ context.Context, arg sqlc.GetLatestSucceededMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error) {
			if arg.UserID != pgUUID(userID) || arg.MainID != pgUUID(mainID) {
				t.Fatalf("GetLatestSucceededMainPurchaseAttemptByUserIDAndMainIDForUpdate() args got %#v", arg)
			}
			return row, nil
		},
		getAttemptByProvider: func(_ context.Context, gotRef pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
			if gotRef != pgText("purchase-ref") {
				t.Fatalf("GetMainPurchaseAttemptByProviderPurchaseRefForUpdate() ref got %#v want %#v", gotRef, pgText("purchase-ref"))
			}
			return row, nil
		},
		getAttemptByTxn: func(_ context.Context, gotRef pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
			if gotRef != pgText("txn-ref") {
				t.Fatalf("GetMainPurchaseAttemptByProviderTransactionRefForUpdate() ref got %#v want %#v", gotRef, pgText("txn-ref"))
			}
			return row, nil
		},
		updateAttemptOutcome: func(_ context.Context, arg sqlc.UpdateMainPurchaseAttemptOutcomeParams) (sqlc.AppMainPurchaseAttempt, error) {
			updateArg = arg
			return row, nil
		},
		getPaymentMethod: func(context.Context, sqlc.GetUserPaymentMethodByIDAndUserIDParams) (sqlc.AppUserPaymentMethod, error) {
			return sqlc.AppUserPaymentMethod{}, nil
		},
		listPaymentMethods: func(context.Context, pgtype.UUID) ([]sqlc.AppUserPaymentMethod, error) {
			return nil, nil
		},
		touchPaymentMethod: func(context.Context, sqlc.TouchUserPaymentMethodLastUsedAtParams) (sqlc.AppUserPaymentMethod, error) {
			return sqlc.AppUserPaymentMethod{}, nil
		},
		upsertPaymentMethod: func(context.Context, sqlc.UpsertUserPaymentMethodParams) (sqlc.AppUserPaymentMethod, error) {
			return sqlc.AppUserPaymentMethod{}, nil
		},
	})

	if err := repo.AcquireMainPurchaseLock(context.Background(), userID, mainID); err != nil {
		t.Fatalf("AcquireMainPurchaseLock() error = %v, want nil", err)
	}
	if lockArg.UserKey != userID.String() || lockArg.MainKey != mainID.String() {
		t.Fatalf("AcquireMainPurchaseLock() args got %#v", lockArg)
	}

	attempt, err := repo.CreateMainPurchaseAttempt(context.Background(), CreateMainPurchaseAttemptInput{
		AcceptedAge:             true,
		AcceptedTerms:           true,
		FromShortID:             shortID,
		IdempotencyKey:          "request-key",
		MainID:                  mainID,
		PaymentMethodMode:       PaymentMethodModeSavedCard,
		Provider:                ProviderCCBill,
		ProviderPaymentTokenRef: "token-1",
		RequestedCurrencyCode:   392,
		RequestedPriceJPY:       1800,
		Status:                  PurchaseAttemptStatusProcessing,
		UserID:                  userID,
		UserPaymentMethodID:     &methodID,
	})
	if err != nil {
		t.Fatalf("CreateMainPurchaseAttempt() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(attempt, wantMainPurchaseAttempt(attemptID, userID, mainID, shortID, &methodID, now)) {
		t.Fatalf("CreateMainPurchaseAttempt() got %#v want %#v", attempt, wantMainPurchaseAttempt(attemptID, userID, mainID, shortID, &methodID, now))
	}
	if createArg.UserID != pgUUID(userID) || createArg.MainID != pgUUID(mainID) || createArg.UserPaymentMethodID != pgUUID(methodID) {
		t.Fatalf("CreateMainPurchaseAttempt() args got %#v", createArg)
	}

	for _, tc := range []struct {
		name string
		run  func(context.Context) (MainPurchaseAttempt, error)
	}{
		{name: "get", run: func(ctx context.Context) (MainPurchaseAttempt, error) {
			return repo.GetMainPurchaseAttempt(ctx, attemptID)
		}},
		{name: "get for update", run: func(ctx context.Context) (MainPurchaseAttempt, error) {
			return repo.GetMainPurchaseAttemptForUpdate(ctx, attemptID)
		}},
		{name: "get by key", run: func(ctx context.Context) (MainPurchaseAttempt, error) {
			return repo.GetMainPurchaseAttemptByIdempotencyKeyForUpdate(ctx, "request-key")
		}},
		{name: "get inflight", run: func(ctx context.Context) (MainPurchaseAttempt, error) {
			return repo.GetLatestInflightMainPurchaseAttemptForUpdate(ctx, userID, mainID)
		}},
		{name: "get succeeded", run: func(ctx context.Context) (MainPurchaseAttempt, error) {
			return repo.GetLatestSucceededMainPurchaseAttemptForUpdate(ctx, userID, mainID)
		}},
		{name: "get by provider ref", run: func(ctx context.Context) (MainPurchaseAttempt, error) {
			return repo.GetMainPurchaseAttemptByProviderPurchaseRefForUpdate(ctx, "purchase-ref")
		}},
		{name: "get by provider txn ref", run: func(ctx context.Context) (MainPurchaseAttempt, error) {
			return repo.GetMainPurchaseAttemptByProviderTransactionRefForUpdate(ctx, "txn-ref")
		}},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.run(context.Background())
			if err != nil {
				t.Fatalf("%s error = %v, want nil", tc.name, err)
			}
			if !reflect.DeepEqual(got, wantMainPurchaseAttempt(attemptID, userID, mainID, shortID, &methodID, now)) {
				t.Fatalf("%s got %#v want %#v", tc.name, got, wantMainPurchaseAttempt(attemptID, userID, mainID, shortID, &methodID, now))
			}
		})
	}

	attempt, err = repo.UpdateMainPurchaseAttemptOutcome(context.Background(), UpdateMainPurchaseAttemptOutcomeInput{
		FailureReason:            stringPtr(FailureReasonPurchaseDeclined),
		ID:                       attemptID,
		ProviderDeclineCode:      int32Ptr(11),
		ProviderDeclineText:      stringPtr("declined"),
		ProviderPaymentTokenRef:  stringPtr("token-2"),
		ProviderPaymentUniqueRef: stringPtr("unique-ref"),
		ProviderProcessedAt:      timePtr(now),
		ProviderPurchaseRef:      stringPtr("purchase-ref"),
		ProviderSessionRef:       stringPtr("session-ref"),
		ProviderTransactionRef:   stringPtr("txn-ref"),
		Status:                   PurchaseAttemptStatusFailed,
	})
	if err != nil {
		t.Fatalf("UpdateMainPurchaseAttemptOutcome() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(attempt, wantMainPurchaseAttempt(attemptID, userID, mainID, shortID, &methodID, now)) {
		t.Fatalf("UpdateMainPurchaseAttemptOutcome() got %#v want %#v", attempt, wantMainPurchaseAttempt(attemptID, userID, mainID, shortID, &methodID, now))
	}
	if updateArg.ID != pgUUID(attemptID) || updateArg.Status != PurchaseAttemptStatusFailed {
		t.Fatalf("UpdateMainPurchaseAttemptOutcome() args got %#v", updateArg)
	}
	if updateArg.ProviderPaymentTokenRef != pgText("token-2") {
		t.Fatalf("UpdateMainPurchaseAttemptOutcome() provider payment token ref got %#v want %#v", updateArg.ProviderPaymentTokenRef, pgText("token-2"))
	}
}

func TestRepositoryNotFoundAndConversionErrors(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	mainID := uuid.New()
	now := time.Unix(1_710_000_000, 0).UTC()
	attemptID := uuid.New()
	shortID := uuid.New()

	repo := newRepository(repositoryStubQueries{
		createAttempt: func(context.Context, sqlc.CreateMainPurchaseAttemptParams) (sqlc.AppMainPurchaseAttempt, error) {
			row := testMainPurchaseAttemptRow(attemptID, userID, mainID, shortID, nil, now)
			row.ID = pgtype.UUID{}
			return row, nil
		},
		getAttempt: func(context.Context, pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, pgx.ErrNoRows
		},
		getAttemptForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, pgx.ErrNoRows
		},
		getAttemptByKey: func(context.Context, string) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, pgx.ErrNoRows
		},
		getInflightAttempt: func(context.Context, sqlc.GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, pgx.ErrNoRows
		},
		getSucceededAttempt: func(context.Context, sqlc.GetLatestSucceededMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, pgx.ErrNoRows
		},
		getAttemptByProvider: func(context.Context, pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, pgx.ErrNoRows
		},
		getAttemptByTxn: func(context.Context, pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, pgx.ErrNoRows
		},
		updateAttemptOutcome: func(context.Context, sqlc.UpdateMainPurchaseAttemptOutcomeParams) (sqlc.AppMainPurchaseAttempt, error) {
			return sqlc.AppMainPurchaseAttempt{}, pgx.ErrNoRows
		},
		getPaymentMethod: func(context.Context, sqlc.GetUserPaymentMethodByIDAndUserIDParams) (sqlc.AppUserPaymentMethod, error) {
			return sqlc.AppUserPaymentMethod{}, pgx.ErrNoRows
		},
		listPaymentMethods: func(context.Context, pgtype.UUID) ([]sqlc.AppUserPaymentMethod, error) {
			row := testPaymentMethodRow(uuid.New(), userID, now)
			row.ID = pgtype.UUID{}
			return []sqlc.AppUserPaymentMethod{row}, nil
		},
		touchPaymentMethod: func(context.Context, sqlc.TouchUserPaymentMethodLastUsedAtParams) (sqlc.AppUserPaymentMethod, error) {
			return sqlc.AppUserPaymentMethod{}, pgx.ErrNoRows
		},
		upsertPaymentMethod: func(context.Context, sqlc.UpsertUserPaymentMethodParams) (sqlc.AppUserPaymentMethod, error) {
			row := testPaymentMethodRow(uuid.New(), userID, now)
			row.ID = pgtype.UUID{}
			return row, nil
		},
	})

	if _, err := repo.GetSavedPaymentMethod(context.Background(), userID, "bad-id"); !errors.Is(err, ErrSavedPaymentMethodNotFound) {
		t.Fatalf("GetSavedPaymentMethod(invalid id) error got %v want %v", err, ErrSavedPaymentMethodNotFound)
	}
	if _, err := repo.GetMainPurchaseAttempt(context.Background(), attemptID); !errors.Is(err, ErrMainPurchaseAttemptNotFound) {
		t.Fatalf("GetMainPurchaseAttempt() error got %v want %v", err, ErrMainPurchaseAttemptNotFound)
	}
	if _, err := repo.GetMainPurchaseAttemptForUpdate(context.Background(), attemptID); !errors.Is(err, ErrMainPurchaseAttemptNotFound) {
		t.Fatalf("GetMainPurchaseAttemptForUpdate() error got %v want %v", err, ErrMainPurchaseAttemptNotFound)
	}
	if _, err := repo.GetMainPurchaseAttemptByIdempotencyKeyForUpdate(context.Background(), "k"); !errors.Is(err, ErrMainPurchaseAttemptNotFound) {
		t.Fatalf("GetMainPurchaseAttemptByIdempotencyKeyForUpdate() error got %v want %v", err, ErrMainPurchaseAttemptNotFound)
	}
	if _, err := repo.GetLatestInflightMainPurchaseAttemptForUpdate(context.Background(), userID, mainID); !errors.Is(err, ErrMainPurchaseAttemptNotFound) {
		t.Fatalf("GetLatestInflightMainPurchaseAttemptForUpdate() error got %v want %v", err, ErrMainPurchaseAttemptNotFound)
	}
	if _, err := repo.GetLatestSucceededMainPurchaseAttemptForUpdate(context.Background(), userID, mainID); !errors.Is(err, ErrMainPurchaseAttemptNotFound) {
		t.Fatalf("GetLatestSucceededMainPurchaseAttemptForUpdate() error got %v want %v", err, ErrMainPurchaseAttemptNotFound)
	}
	if _, err := repo.GetMainPurchaseAttemptByProviderPurchaseRefForUpdate(context.Background(), "ref"); !errors.Is(err, ErrMainPurchaseAttemptNotFound) {
		t.Fatalf("GetMainPurchaseAttemptByProviderPurchaseRefForUpdate() error got %v want %v", err, ErrMainPurchaseAttemptNotFound)
	}
	if _, err := repo.GetMainPurchaseAttemptByProviderTransactionRefForUpdate(context.Background(), "txn"); !errors.Is(err, ErrMainPurchaseAttemptNotFound) {
		t.Fatalf("GetMainPurchaseAttemptByProviderTransactionRefForUpdate() error got %v want %v", err, ErrMainPurchaseAttemptNotFound)
	}
	if _, err := repo.TouchSavedPaymentMethodLastUsedAt(context.Background(), userID, FormatPublicPaymentMethodID(uuid.New()), timePtr(now)); !errors.Is(err, ErrSavedPaymentMethodNotFound) {
		t.Fatalf("TouchSavedPaymentMethodLastUsedAt() error got %v want %v", err, ErrSavedPaymentMethodNotFound)
	}
	if _, err := repo.ListSavedPaymentMethods(context.Background(), userID); err == nil {
		t.Fatal("ListSavedPaymentMethods() error = nil, want conversion error")
	}
	if _, err := repo.UpsertSavedPaymentMethod(context.Background(), UpsertSavedPaymentMethodInput{
		Brand:                     CardBrandVisa,
		Last4:                     "4242",
		Provider:                  ProviderCCBill,
		ProviderPaymentAccountRef: "acct-1",
		ProviderPaymentTokenRef:   "token-1",
		UserID:                    userID,
	}); err == nil {
		t.Fatal("UpsertSavedPaymentMethod() error = nil, want conversion error")
	}
	if _, err := repo.CreateMainPurchaseAttempt(context.Background(), CreateMainPurchaseAttemptInput{
		FromShortID:             shortID,
		IdempotencyKey:          "k",
		MainID:                  mainID,
		PaymentMethodMode:       PaymentMethodModeNewCard,
		Provider:                ProviderCCBill,
		ProviderPaymentTokenRef: "token-1",
		RequestedCurrencyCode:   392,
		RequestedPriceJPY:       1800,
		Status:                  PurchaseAttemptStatusProcessing,
		UserID:                  userID,
	}); err == nil {
		t.Fatal("CreateMainPurchaseAttempt() error = nil, want conversion error")
	}
}

func TestRepositoryCreateMainPurchaseAttemptMapsConflictError(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()

	for _, constraintName := range []string{
		mainPurchaseAttemptIdempotencyUniqueConstraint,
		mainPurchaseAttemptInflightUniqueConstraint,
	} {
		constraintName := constraintName
		t.Run(constraintName, func(t *testing.T) {
			t.Parallel()

			repo := newRepository(repositoryStubQueries{
				createAttempt: func(context.Context, sqlc.CreateMainPurchaseAttemptParams) (sqlc.AppMainPurchaseAttempt, error) {
					return sqlc.AppMainPurchaseAttempt{}, &pgconn.PgError{
						Code:           "23505",
						ConstraintName: constraintName,
					}
				},
				getAttempt: func(context.Context, pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
					return sqlc.AppMainPurchaseAttempt{}, nil
				},
				getAttemptForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppMainPurchaseAttempt, error) {
					return sqlc.AppMainPurchaseAttempt{}, nil
				},
				getAttemptByKey: func(context.Context, string) (sqlc.AppMainPurchaseAttempt, error) {
					return sqlc.AppMainPurchaseAttempt{}, nil
				},
				getInflightAttempt: func(context.Context, sqlc.GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdateParams) (sqlc.AppMainPurchaseAttempt, error) {
					return sqlc.AppMainPurchaseAttempt{}, nil
				},
				getAttemptByProvider: func(context.Context, pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
					return sqlc.AppMainPurchaseAttempt{}, nil
				},
				getAttemptByTxn: func(context.Context, pgtype.Text) (sqlc.AppMainPurchaseAttempt, error) {
					return sqlc.AppMainPurchaseAttempt{}, nil
				},
				updateAttemptOutcome: func(context.Context, sqlc.UpdateMainPurchaseAttemptOutcomeParams) (sqlc.AppMainPurchaseAttempt, error) {
					return sqlc.AppMainPurchaseAttempt{}, nil
				},
				getPaymentMethod: func(context.Context, sqlc.GetUserPaymentMethodByIDAndUserIDParams) (sqlc.AppUserPaymentMethod, error) {
					return sqlc.AppUserPaymentMethod{}, nil
				},
				listPaymentMethods: func(context.Context, pgtype.UUID) ([]sqlc.AppUserPaymentMethod, error) {
					return nil, nil
				},
				touchPaymentMethod: func(context.Context, sqlc.TouchUserPaymentMethodLastUsedAtParams) (sqlc.AppUserPaymentMethod, error) {
					return sqlc.AppUserPaymentMethod{}, nil
				},
				upsertPaymentMethod: func(context.Context, sqlc.UpsertUserPaymentMethodParams) (sqlc.AppUserPaymentMethod, error) {
					return sqlc.AppUserPaymentMethod{}, nil
				},
			})

			_, err := repo.CreateMainPurchaseAttempt(context.Background(), CreateMainPurchaseAttemptInput{
				FromShortID:             shortID,
				IdempotencyKey:          "request-key",
				MainID:                  mainID,
				PaymentMethodMode:       PaymentMethodModeNewCard,
				Provider:                ProviderCCBill,
				ProviderPaymentTokenRef: "token-1",
				RequestedCurrencyCode:   392,
				RequestedPriceJPY:       1800,
				Status:                  PurchaseAttemptStatusProcessing,
				UserID:                  userID,
			})
			if !errors.Is(err, ErrMainPurchaseAttemptConflict) {
				t.Fatalf("CreateMainPurchaseAttempt() error got %v want %v", err, ErrMainPurchaseAttemptConflict)
			}
		})
	}
}

func testPaymentMethodRow(id uuid.UUID, userID uuid.UUID, now time.Time) sqlc.AppUserPaymentMethod {
	return sqlc.AppUserPaymentMethod{
		ID:                        pgUUID(id),
		UserID:                    pgUUID(userID),
		Provider:                  ProviderCCBill,
		ProviderPaymentTokenRef:   "token-1",
		ProviderPaymentAccountRef: "acct-1",
		Brand:                     CardBrandVisa,
		Last4:                     "4242",
		CreatedAt:                 pgTime(now),
		UpdatedAt:                 pgTime(now.Add(time.Minute)),
		LastUsedAt:                pgTime(now.Add(2 * time.Minute)),
	}
}

func testMainPurchaseAttemptRow(id uuid.UUID, userID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID, methodID *uuid.UUID, now time.Time) sqlc.AppMainPurchaseAttempt {
	return sqlc.AppMainPurchaseAttempt{
		ID:                       pgUUID(id),
		UserID:                   pgUUID(userID),
		MainID:                   pgUUID(mainID),
		FromShortID:              pgUUID(shortID),
		Provider:                 ProviderCCBill,
		PaymentMethodMode:        PaymentMethodModeSavedCard,
		UserPaymentMethodID:      optionalUUIDToPG(methodID),
		ProviderPaymentTokenRef:  "token-1",
		IdempotencyKey:           "request-key",
		Status:                   PurchaseAttemptStatusFailed,
		FailureReason:            pgText("purchase_declined"),
		PendingReason:            pgtype.Text{},
		ProviderPurchaseRef:      pgText("purchase-ref"),
		ProviderTransactionRef:   pgText("txn-ref"),
		ProviderSessionRef:       pgText("session-ref"),
		ProviderPaymentUniqueRef: pgText("unique-ref"),
		ProviderDeclineCode:      pgtype.Int4{Int32: 11, Valid: true},
		ProviderDeclineText:      pgText("declined"),
		RequestedPriceJpy:        1800,
		RequestedCurrencyCode:    392,
		AcceptedAge:              true,
		AcceptedTerms:            true,
		ProviderProcessedAt:      pgTime(now),
		CreatedAt:                pgTime(now),
		UpdatedAt:                pgTime(now.Add(time.Minute)),
	}
}

func wantPaymentMethod(id uuid.UUID, userID uuid.UUID, now time.Time) SavedPaymentMethod {
	return SavedPaymentMethod{
		Brand:                     CardBrandVisa,
		CreatedAt:                 now,
		ID:                        id,
		Last4:                     "4242",
		LastUsedAt:                now.Add(2 * time.Minute),
		PaymentMethodID:           FormatPublicPaymentMethodID(id),
		Provider:                  ProviderCCBill,
		ProviderPaymentAccountRef: "acct-1",
		ProviderPaymentTokenRef:   "token-1",
		UpdatedAt:                 now.Add(time.Minute),
		UserID:                    userID,
	}
}

func wantMainPurchaseAttempt(id uuid.UUID, userID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID, methodID *uuid.UUID, now time.Time) MainPurchaseAttempt {
	return MainPurchaseAttempt{
		AcceptedAge:              true,
		AcceptedTerms:            true,
		CreatedAt:                now,
		FailureReason:            stringPtr("purchase_declined"),
		FromShortID:              shortID,
		ID:                       id,
		IdempotencyKey:           "request-key",
		MainID:                   mainID,
		PaymentMethodMode:        PaymentMethodModeSavedCard,
		PendingReason:            nil,
		Provider:                 ProviderCCBill,
		ProviderDeclineCode:      int32Ptr(11),
		ProviderDeclineText:      stringPtr("declined"),
		ProviderPaymentTokenRef:  "token-1",
		ProviderPaymentUniqueRef: stringPtr("unique-ref"),
		ProviderProcessedAt:      timePtr(now),
		ProviderPurchaseRef:      stringPtr("purchase-ref"),
		ProviderSessionRef:       stringPtr("session-ref"),
		ProviderTransactionRef:   stringPtr("txn-ref"),
		RequestedCurrencyCode:    392,
		RequestedPriceJPY:        1800,
		Status:                   PurchaseAttemptStatusFailed,
		UpdatedAt:                now.Add(time.Minute),
		UserID:                   userID,
		UserPaymentMethodID:      methodID,
	}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func pgTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func pgText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
}

func int32Ptr(value int32) *int32 {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
