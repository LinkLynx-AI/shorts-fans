package devseed

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type stubBeginner struct {
	tx  pgx.Tx
	err error
}

func (s stubBeginner) Begin(context.Context) (pgx.Tx, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.tx, nil
}

type execCall struct {
	query string
	args  []any
}

type stubTx struct {
	execCalls   []execCall
	execErrAt   int
	commitErr   error
	rollbackErr error
	committed   bool
	rolledBack  bool
}

func (tx *stubTx) Begin(context.Context) (pgx.Tx, error) { return tx, nil }

func (tx *stubTx) Commit(context.Context) error {
	tx.committed = true
	return tx.commitErr
}

func (tx *stubTx) Rollback(context.Context) error {
	tx.rolledBack = true
	return tx.rollbackErr
}

func (tx *stubTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (tx *stubTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (tx *stubTx) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }

func (tx *stubTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (tx *stubTx) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	tx.execCalls = append(tx.execCalls, execCall{
		query: query,
		args:  args,
	})
	if tx.execErrAt > 0 && len(tx.execCalls) == tx.execErrAt {
		return pgconn.CommandTag{}, errors.New("exec failed")
	}

	return pgconn.CommandTag{}, nil
}

func (tx *stubTx) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (tx *stubTx) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (tx *stubTx) Conn() *pgx.Conn                                         { return nil }

func TestRunRejectsNilBeginner(t *testing.T) {
	t.Parallel()

	_, err := Run(context.Background(), nil)
	if err == nil {
		t.Fatal("Run() error = nil, want nil beginner error")
	}
	if !strings.Contains(err.Error(), "tx beginner が nil") {
		t.Fatalf("Run() error got %q want tx beginner message", err.Error())
	}
}

func TestRunSeedsAllStatementsInOneTransaction(t *testing.T) {
	t.Parallel()

	tx := &stubTx{}
	summary, err := Run(context.Background(), stubBeginner{tx: tx})
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if !tx.committed {
		t.Fatal("Run() committed = false, want true")
	}
	if tx.rolledBack {
		t.Fatal("Run() rolledBack = true, want false")
	}
	if len(tx.execCalls) != 15 {
		t.Fatalf("Run() exec call count got %d want 15", len(tx.execCalls))
	}
	if summary.CreatorUserID != creatorUserID {
		t.Fatalf("Run() creator user id got %s want %s", summary.CreatorUserID, creatorUserID)
	}
	if summary.FanUserID != fanUserID {
		t.Fatalf("Run() fan user id got %s want %s", summary.FanUserID, fanUserID)
	}
	if summary.MainID != mainID {
		t.Fatalf("Run() main id got %s want %s", summary.MainID, mainID)
	}
	if len(summary.ShortIDs) != 2 || summary.ShortIDs[0] != shortAID || summary.ShortIDs[1] != shortBID {
		t.Fatalf("Run() short ids got %#v want [%s %s]", summary.ShortIDs, shortAID, shortBID)
	}
	if summary.FanSessionToken != fanSessionToken {
		t.Fatalf("Run() fan session token got %q want %q", summary.FanSessionToken, fanSessionToken)
	}
	if summary.CreatorSessionToken != creatorSessionToken {
		t.Fatalf("Run() creator session token got %q want %q", summary.CreatorSessionToken, creatorSessionToken)
	}

	if got := tx.execCalls[0].args[0]; got != creatorUserID {
		t.Fatalf("Run() first upsert user arg got %v want %v", got, creatorUserID)
	}
	if got := tx.execCalls[1].args[0]; got != fanUserID {
		t.Fatalf("Run() second upsert user arg got %v want %v", got, fanUserID)
	}
	if got := tx.execCalls[3].args[2]; got != creatorHandle {
		t.Fatalf("Run() creator profile handle arg got %v want %v", got, creatorHandle)
	}
	if got := tx.execCalls[7].args[0]; got != mainID {
		t.Fatalf("Run() main upsert id arg got %v want %v", got, mainID)
	}
	if got := tx.execCalls[10].args[0]; got != fanUserID {
		t.Fatalf("Run() main unlock user arg got %v want %v", got, fanUserID)
	}
	if got := tx.execCalls[12].args[1]; got != shortAID {
		t.Fatalf("Run() pinned short id arg got %v want %v", got, shortAID)
	}
	if got := tx.execCalls[13].args[0]; got != fanUserID {
		t.Fatalf("Run() fan session user arg got %v want %v", got, fanUserID)
	}
	if got := tx.execCalls[14].args[0]; got != creatorUserID {
		t.Fatalf("Run() creator session user arg got %v want %v", got, creatorUserID)
	}
}

func TestRunRollsBackWhenExecFails(t *testing.T) {
	t.Parallel()

	tx := &stubTx{execErrAt: 8}

	_, err := Run(context.Background(), stubBeginner{tx: tx})
	if err == nil {
		t.Fatal("Run() error = nil, want exec failure")
	}
	if !tx.rolledBack {
		t.Fatal("Run() rolledBack = false, want true")
	}
	if tx.committed {
		t.Fatal("Run() committed = true, want false")
	}
	if !strings.Contains(err.Error(), "dev seed 適用") {
		t.Fatalf("Run() error got %q want dev seed context", err.Error())
	}
	if !strings.Contains(err.Error(), "mains upsert") {
		t.Fatalf("Run() error got %q want main upsert context", err.Error())
	}
}

func TestRunReturnsBeginError(t *testing.T) {
	t.Parallel()

	_, err := Run(context.Background(), stubBeginner{err: errors.New("begin failed")})
	if err == nil {
		t.Fatal("Run() error = nil, want begin failure")
	}
	if !strings.Contains(err.Error(), "transaction 開始") {
		t.Fatalf("Run() error got %q want transaction start context", err.Error())
	}
}
