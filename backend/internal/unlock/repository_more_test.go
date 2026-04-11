package unlock

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
)

type repositoryStubQueries struct {
	createUnlock func(context.Context, sqlc.CreateMainUnlockParams) (sqlc.AppMainUnlock, error)
	ensureUnlock func(context.Context, sqlc.EnsureMainUnlockParams) (sqlc.EnsureMainUnlockRow, error)
	getUnlock    func(context.Context, sqlc.GetMainUnlockByUserIDAndMainIDParams) (sqlc.AppMainUnlock, error)
	listUnlocks  func(context.Context, pgtype.UUID) ([]pgtype.UUID, error)
}

func (s repositoryStubQueries) CreateMainUnlock(ctx context.Context, arg sqlc.CreateMainUnlockParams) (sqlc.AppMainUnlock, error) {
	return s.createUnlock(ctx, arg)
}

func (s repositoryStubQueries) EnsureMainUnlock(ctx context.Context, arg sqlc.EnsureMainUnlockParams) (sqlc.EnsureMainUnlockRow, error) {
	return s.ensureUnlock(ctx, arg)
}

func (s repositoryStubQueries) GetMainUnlockByUserIDAndMainID(ctx context.Context, arg sqlc.GetMainUnlockByUserIDAndMainIDParams) (sqlc.AppMainUnlock, error) {
	return s.getUnlock(ctx, arg)
}

func (s repositoryStubQueries) ListUnlockedMainIDsByUserID(ctx context.Context, userID pgtype.UUID) ([]pgtype.UUID, error) {
	return s.listUnlocks(ctx, userID)
}

func TestRepositorySuccessPaths(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.New()
	mainID := uuid.New()
	purchaseRef := stringPtr("purchase-1")
	row := testMainUnlockRow(userID, mainID, now, purchaseRef)

	var createArg sqlc.CreateMainUnlockParams
	var ensureArg sqlc.EnsureMainUnlockParams
	repo := newRepository(repositoryStubQueries{
		createUnlock: func(_ context.Context, arg sqlc.CreateMainUnlockParams) (sqlc.AppMainUnlock, error) {
			createArg = arg
			return row, nil
		},
		ensureUnlock: func(_ context.Context, arg sqlc.EnsureMainUnlockParams) (sqlc.EnsureMainUnlockRow, error) {
			ensureArg = arg
			return sqlc.EnsureMainUnlockRow(row), nil
		},
		getUnlock: func(_ context.Context, arg sqlc.GetMainUnlockByUserIDAndMainIDParams) (sqlc.AppMainUnlock, error) {
			if arg.UserID != pgUUID(userID) || arg.MainID != pgUUID(mainID) {
				t.Fatalf("GetMainUnlockByUserIDAndMainID() args got %#v", arg)
			}
			return row, nil
		},
		listUnlocks: func(_ context.Context, gotUserID pgtype.UUID) ([]pgtype.UUID, error) {
			if gotUserID != pgUUID(userID) {
				t.Fatalf("ListUnlockedMainIDsByUserID() user got %v want %v", gotUserID, pgUUID(userID))
			}
			return []pgtype.UUID{pgUUID(mainID)}, nil
		},
	})

	recorded, err := repo.RecordMainUnlock(context.Background(), RecordMainUnlockInput{
		UserID:                     userID,
		MainID:                     mainID,
		PaymentProviderPurchaseRef: purchaseRef,
		PurchasedAt:                timePtr(now),
	})
	if err != nil {
		t.Fatalf("RecordMainUnlock() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(recorded, wantMainUnlock(userID, mainID, now, purchaseRef)) {
		t.Fatalf("RecordMainUnlock() got %#v want %#v", recorded, wantMainUnlock(userID, mainID, now, purchaseRef))
	}
	if createArg.UserID != pgUUID(userID) || createArg.MainID != pgUUID(mainID) {
		t.Fatalf("RecordMainUnlock() args got %#v", createArg)
	}
	ensured, err := repo.EnsureMainUnlock(context.Background(), RecordMainUnlockInput{
		UserID:                     userID,
		MainID:                     mainID,
		PaymentProviderPurchaseRef: purchaseRef,
		PurchasedAt:                timePtr(now),
	})
	if err != nil {
		t.Fatalf("EnsureMainUnlock() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(ensured, wantMainUnlock(userID, mainID, now, purchaseRef)) {
		t.Fatalf("EnsureMainUnlock() got %#v want %#v", ensured, wantMainUnlock(userID, mainID, now, purchaseRef))
	}
	if ensureArg.UserID != pgUUID(userID) || ensureArg.MainID != pgUUID(mainID) {
		t.Fatalf("EnsureMainUnlock() args got %#v", ensureArg)
	}

	got, err := repo.GetMainUnlock(context.Background(), userID, mainID)
	if err != nil {
		t.Fatalf("GetMainUnlock() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, recorded) {
		t.Fatalf("GetMainUnlock() got %#v want %#v", got, recorded)
	}

	ids, err := repo.ListUnlockedMainIDs(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListUnlockedMainIDs() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(ids, []uuid.UUID{mainID}) {
		t.Fatalf("ListUnlockedMainIDs() got %#v want %#v", ids, []uuid.UUID{mainID})
	}
}

func TestRepositoryErrorPaths(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	mainID := uuid.New()
	genericErr := errors.New("query failed")
	repo := newRepository(repositoryStubQueries{
		createUnlock: func(context.Context, sqlc.CreateMainUnlockParams) (sqlc.AppMainUnlock, error) {
			return sqlc.AppMainUnlock{}, genericErr
		},
		ensureUnlock: func(context.Context, sqlc.EnsureMainUnlockParams) (sqlc.EnsureMainUnlockRow, error) {
			return sqlc.EnsureMainUnlockRow{}, genericErr
		},
		getUnlock: func(context.Context, sqlc.GetMainUnlockByUserIDAndMainIDParams) (sqlc.AppMainUnlock, error) {
			return sqlc.AppMainUnlock{}, pgx.ErrNoRows
		},
		listUnlocks: func(context.Context, pgtype.UUID) ([]pgtype.UUID, error) {
			return nil, genericErr
		},
	})

	if _, err := repo.RecordMainUnlock(context.Background(), RecordMainUnlockInput{UserID: userID, MainID: mainID}); !errors.Is(err, genericErr) {
		t.Fatalf("RecordMainUnlock() error got %v want wrapped %v", err, genericErr)
	}
	if _, err := repo.EnsureMainUnlock(context.Background(), RecordMainUnlockInput{UserID: userID, MainID: mainID}); !errors.Is(err, genericErr) {
		t.Fatalf("EnsureMainUnlock() error got %v want wrapped %v", err, genericErr)
	}
	if _, err := repo.GetMainUnlock(context.Background(), userID, mainID); !errors.Is(err, ErrMainUnlockNotFound) {
		t.Fatalf("GetMainUnlock() error got %v want %v", err, ErrMainUnlockNotFound)
	}
	if _, err := repo.ListUnlockedMainIDs(context.Background(), userID); !errors.Is(err, genericErr) {
		t.Fatalf("ListUnlockedMainIDs() error got %v want wrapped %v", err, genericErr)
	}
}

func TestRepositoryConversionErrors(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.New()
	mainID := uuid.New()
	invalidRow := testMainUnlockRow(userID, mainID, now, nil)
	invalidRow.UserID = pgtype.UUID{}

	repo := newRepository(repositoryStubQueries{
		createUnlock: func(context.Context, sqlc.CreateMainUnlockParams) (sqlc.AppMainUnlock, error) {
			return invalidRow, nil
		},
		ensureUnlock: func(context.Context, sqlc.EnsureMainUnlockParams) (sqlc.EnsureMainUnlockRow, error) {
			return sqlc.EnsureMainUnlockRow(invalidRow), nil
		},
		getUnlock: func(context.Context, sqlc.GetMainUnlockByUserIDAndMainIDParams) (sqlc.AppMainUnlock, error) {
			return invalidRow, nil
		},
		listUnlocks: func(context.Context, pgtype.UUID) ([]pgtype.UUID, error) {
			return []pgtype.UUID{{}}, nil
		},
	})

	if _, err := repo.RecordMainUnlock(context.Background(), RecordMainUnlockInput{UserID: userID, MainID: mainID}); err == nil {
		t.Fatal("RecordMainUnlock() error = nil, want conversion error")
	}
	if _, err := repo.EnsureMainUnlock(context.Background(), RecordMainUnlockInput{UserID: userID, MainID: mainID}); err == nil {
		t.Fatal("EnsureMainUnlock() error = nil, want conversion error")
	}
	if _, err := repo.GetMainUnlock(context.Background(), userID, mainID); err == nil {
		t.Fatal("GetMainUnlock() error = nil, want conversion error")
	}
	if _, err := repo.ListUnlockedMainIDs(context.Background(), userID); err == nil {
		t.Fatal("ListUnlockedMainIDs() error = nil, want conversion error")
	}
}

func TestIsAlreadyUnlockedErrorRejectsOtherErrors(t *testing.T) {
	t.Parallel()

	if isAlreadyUnlockedError(errors.New("plain")) {
		t.Fatal("isAlreadyUnlockedError() got true want false for plain error")
	}
	if isAlreadyUnlockedError(&pgconn.PgError{Code: "23505", ConstraintName: "other_constraint"}) {
		t.Fatal("isAlreadyUnlockedError() got true want false for other constraint")
	}
}

func testMainUnlockRow(userID uuid.UUID, mainID uuid.UUID, now time.Time, purchaseRef *string) sqlc.AppMainUnlock {
	return sqlc.AppMainUnlock{
		UserID:                     pgUUID(userID),
		MainID:                     pgUUID(mainID),
		PaymentProviderPurchaseRef: pgText(purchaseRef),
		PurchasedAt:                pgTime(now),
		CreatedAt:                  pgTime(now.Add(time.Minute)),
	}
}

func wantMainUnlock(userID uuid.UUID, mainID uuid.UUID, now time.Time, purchaseRef *string) MainUnlock {
	return MainUnlock{
		UserID:                     userID,
		MainID:                     mainID,
		PaymentProviderPurchaseRef: purchaseRef,
		PurchasedAt:                now,
		CreatedAt:                  now.Add(time.Minute),
	}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func pgTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func pgText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func stringPtr(value string) *string {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
