package repository

import (
	"context"
	"fmt"

	"errors"

	"github.com/jmoiron/sqlx"
)

type txFunc func(tx *sqlx.Tx) error

// Обертка для транзакций, txFunc - это функция внутри которой должны быть вызовы атомарных методов репозитория
func sqlxTransaction(ctx context.Context, db *sqlx.DB, f txFunc) (txErr error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error creating transaction: %w", err)
	}

	defer func() {
		rollbackFunc := func() error {
			if rbErr := tx.Rollback(); rbErr != nil {
				return errors.Join(txErr, fmt.Errorf("rollback failed: %w", rbErr))
			}
			return nil
		}

		var panicErr error
		if r := recover(); r != nil {
			panicErr = fmt.Errorf("panic in txFunc: %v", r)
			txErr = errors.Join(txErr, panicErr)
			return
		}

		if txErr != nil {
			rbErr := rollbackFunc()
			if rbErr != nil {
				txErr = errors.Join(txErr, rbErr)
				return
			}
		}
	}()

	err = f(tx)
	if err != nil {
		txErr = tx.Rollback()
		if txErr != nil {
			return errors.Join(err, txErr)
		}

		return fmt.Errorf("error during transaction: %w", err)
	}

	txErr = tx.Commit()
	if txErr != nil {
		return fmt.Errorf("error committing transaction: %w", txErr)
	}

	return nil
}
