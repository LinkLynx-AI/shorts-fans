package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// TxBeginner starts a pgx transaction.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// RunInTx executes fn inside a transaction and commits on success.
func RunInTx(ctx context.Context, beginner TxBeginner, fn func(pgx.Tx) error) (err error) {
	if beginner == nil {
		return fmt.Errorf("tx beginner is nil")
	}
	if fn == nil {
		return fmt.Errorf("tx function is nil")
	}

	tx, err := beginner.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	committed := false
	defer func() {
		if committed {
			return
		}

		rollbackErr := tx.Rollback(ctx)
		if rollbackErr == nil || errors.Is(rollbackErr, pgx.ErrTxClosed) {
			return
		}

		if err == nil {
			err = fmt.Errorf("rollback tx: %w", rollbackErr)
			return
		}

		err = fmt.Errorf("%w: rollback tx: %v", err, rollbackErr)
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	committed = true
	return nil
}
