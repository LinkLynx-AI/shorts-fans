package postgres

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

func TestUUIDConversion(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	got, err := UUIDFromPG(UUIDToPG(id))
	if err != nil {
		t.Fatalf("UUIDFromPG() error = %v, want nil", err)
	}
	if got != id {
		t.Fatalf("UUIDFromPG() got %s want %s", got, id)
	}
}

func TestUUIDFromPGRejectsNull(t *testing.T) {
	t.Parallel()

	_, err := UUIDFromPG(pgtype.UUID{})
	if err == nil {
		t.Fatal("UUIDFromPG() error = nil, want error")
	}
}

func TestTimeConversions(t *testing.T) {
	t.Parallel()

	value := time.Unix(1710000000, 0).UTC()

	got, err := RequiredTimeFromPG(TimeToPG(&value))
	if err != nil {
		t.Fatalf("RequiredTimeFromPG() error = %v, want nil", err)
	}
	if !got.Equal(value) {
		t.Fatalf("RequiredTimeFromPG() got %s want %s", got, value)
	}

	if _, err := RequiredTimeFromPG(pgtype.Timestamptz{}); err == nil {
		t.Fatal("RequiredTimeFromPG() error = nil, want error")
	}

	if got := OptionalTimeFromPG(pgtype.Timestamptz{}); got != nil {
		t.Fatalf("OptionalTimeFromPG() got %v want nil", got)
	}

	optional := OptionalTimeFromPG(TimeToPG(&value))
	if optional == nil || !optional.Equal(value) {
		t.Fatalf("OptionalTimeFromPG() got %v want %s", optional, value)
	}

	if got := TimeToPG(nil); got.Valid {
		t.Fatalf("TimeToPG(nil) got valid=%t want false", got.Valid)
	}
}

func TestTextAndIntConversions(t *testing.T) {
	t.Parallel()

	textValue := "playback"
	intValue := int64(42)

	if got := OptionalTextFromPG(pgtype.Text{}); got != nil {
		t.Fatalf("OptionalTextFromPG() got %v want nil", got)
	}
	if got := OptionalTextFromPG(TextToPG(&textValue)); got == nil || *got != textValue {
		t.Fatalf("OptionalTextFromPG() got %v want %q", got, textValue)
	}
	if got := TextToPG(nil); got.Valid {
		t.Fatalf("TextToPG(nil) got valid=%t want false", got.Valid)
	}

	if got := OptionalInt64FromPG(pgtype.Int8{}); got != nil {
		t.Fatalf("OptionalInt64FromPG() got %v want nil", got)
	}
	if got := OptionalInt64FromPG(Int64ToPG(&intValue)); got == nil || *got != intValue {
		t.Fatalf("OptionalInt64FromPG() got %v want %d", got, intValue)
	}
	if got := Int64ToPG(nil); got.Valid {
		t.Fatalf("Int64ToPG(nil) got valid=%t want false", got.Valid)
	}
}

func TestRunInTxValidatesArguments(t *testing.T) {
	t.Parallel()

	if err := RunInTx(context.Background(), nil, func(pgx.Tx) error { return nil }); err == nil {
		t.Fatal("RunInTx() error = nil, want error for nil beginner")
	}

	if err := RunInTx(context.Background(), txBeginnerStub{
		begin: func(context.Context) (pgx.Tx, error) { return &txStub{}, nil },
	}, nil); err == nil {
		t.Fatal("RunInTx() error = nil, want error for nil function")
	}
}

func TestRunInTxSuccessCommits(t *testing.T) {
	t.Parallel()

	tx := &txStub{}
	executed := false

	err := RunInTx(context.Background(), txBeginnerStub{
		begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
	}, func(got pgx.Tx) error {
		executed = true
		if got != tx {
			t.Fatalf("RunInTx() tx got %v want %v", got, tx)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunInTx() error = %v, want nil", err)
	}
	if !executed {
		t.Fatal("RunInTx() executed = false, want true")
	}
	if !tx.committed {
		t.Fatal("RunInTx() committed = false, want true")
	}
	if tx.rolledBack {
		t.Fatal("RunInTx() rolledBack = true, want false")
	}
}

func TestRunInTxWrapsBeginError(t *testing.T) {
	t.Parallel()

	beginErr := errors.New("dial tcp")
	err := RunInTx(context.Background(), txBeginnerStub{
		begin: func(context.Context) (pgx.Tx, error) { return nil, beginErr },
	}, func(pgx.Tx) error { return nil })
	if !errors.Is(err, beginErr) {
		t.Fatalf("RunInTx() error got %v want wrapped %v", err, beginErr)
	}
}

func TestRunInTxWrapsRollbackError(t *testing.T) {
	t.Parallel()

	fnErr := errors.New("write failed")
	rollbackErr := errors.New("rollback failed")
	tx := &txStub{rollbackErr: rollbackErr}

	err := RunInTx(context.Background(), txBeginnerStub{
		begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
	}, func(pgx.Tx) error {
		return fnErr
	})
	if !errors.Is(err, fnErr) {
		t.Fatalf("RunInTx() error got %v want wrapped %v", err, fnErr)
	}
	if !tx.rolledBack {
		t.Fatal("RunInTx() rolledBack = false, want true")
	}
}

func TestRunInTxIgnoresClosedRollback(t *testing.T) {
	t.Parallel()

	fnErr := errors.New("fn failed")
	tx := &txStub{rollbackErr: pgx.ErrTxClosed}

	err := RunInTx(context.Background(), txBeginnerStub{
		begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
	}, func(pgx.Tx) error {
		return fnErr
	})
	if !errors.Is(err, fnErr) {
		t.Fatalf("RunInTx() error got %v want wrapped %v", err, fnErr)
	}
}

func TestRunInTxWrapsCommitError(t *testing.T) {
	t.Parallel()

	commitErr := errors.New("commit failed")
	tx := &txStub{commitErr: commitErr}

	err := RunInTx(context.Background(), txBeginnerStub{
		begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
	}, func(pgx.Tx) error {
		return nil
	})
	if !errors.Is(err, commitErr) {
		t.Fatalf("RunInTx() error got %v want wrapped %v", err, commitErr)
	}
	if !tx.rolledBack {
		t.Fatal("RunInTx() rolledBack = false, want true after commit error")
	}
}

func TestNewReadinessCheckerAndNilCheck(t *testing.T) {
	t.Parallel()

	checker := NewReadinessChecker((*pgxpool.Pool)(nil))
	if !reflect.DeepEqual(checker, ReadinessChecker{}) {
		t.Fatalf("NewReadinessChecker(nil) got %#v want zero checker", checker)
	}
	if err := checker.CheckReadiness(context.Background()); err == nil {
		t.Fatal("CheckReadiness() error = nil, want error")
	}
}

func TestNewPoolRejectsInvalidDSN(t *testing.T) {
	t.Parallel()

	pool, err := NewPool(context.Background(), "://bad dsn")
	if pool != nil {
		t.Fatalf("NewPool() pool got %v want nil", pool)
	}
	if err == nil {
		t.Fatal("NewPool() error = nil, want error")
	}
}
