package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// TxBeginner は pgx transaction を開始します。
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// RunInTx は transaction 内で fn を実行し、成功時に commit します。
func RunInTx(ctx context.Context, beginner TxBeginner, fn func(pgx.Tx) error) (err error) {
	if beginner == nil {
		return fmt.Errorf("tx beginner が nil です")
	}
	if fn == nil {
		return fmt.Errorf("tx function が nil です")
	}

	tx, err := beginner.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction 開始: %w", err)
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
			err = fmt.Errorf("transaction rollback: %w", rollbackErr)
			return
		}

		err = fmt.Errorf("%w: transaction rollback: %v", err, rollbackErr)
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit: %w", err)
	}

	committed = true
	return nil
}
