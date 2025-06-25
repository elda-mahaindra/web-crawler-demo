package store

import (
	"context"
	"errors"
	"fmt"

	"web-crawler/store/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TxOptions represents available transaction options
type TxOptions struct {
	Isolation  pgx.TxIsoLevel
	AccessMode pgx.TxAccessMode
	Deferrable bool
	ReadOnly   bool
}

// DefaultTxOptions returns default transaction options (Serializable, ReadWrite, NotDeferrable)
func DefaultTxOptions() TxOptions {
	return TxOptions{
		Isolation:  pgx.Serializable,
		AccessMode: pgx.ReadWrite,
		Deferrable: false,
		ReadOnly:   false,
	}
}

// ReadOnlyTxOptions returns options optimized for read-only transactions
func ReadOnlyTxOptions() TxOptions {
	return TxOptions{
		Isolation:  pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
		Deferrable: true,
		ReadOnly:   true,
	}
}

// // WithTxOptions executes a function within a database transaction with custom options
func (store *Store) WithTxOptions(ctx context.Context, opts TxOptions, fn func(*sqlc.Queries) error) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	deferrable := pgx.NotDeferrable
	if opts.Deferrable {
		deferrable = pgx.Deferrable
	}

	txOptions := &pgx.TxOptions{
		IsoLevel:       opts.Isolation,
		AccessMode:     opts.AccessMode,
		DeferrableMode: deferrable,
	}

	tx, err := store.pool.BeginTx(ctx, *txOptions)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	q := sqlc.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			if errors.Is(rbErr, pgx.ErrTxClosed) {
				return err
			}

			var pgErr *pgconn.PgError
			if errors.As(rbErr, &pgErr) {
				return fmt.Errorf("PG Error Code: %s, PG Error Message: %s, RollbackTx() error = %v: Original error = %s", pgErr.Code, pgErr.Message, rbErr, err.Error())
			}

			return fmt.Errorf("RollbackTx() error = %v: Original error = %s", rbErr, err.Error())
		}

		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return tx.Commit(ctx)
}

// // WithTx executes a function within a database transaction with default options
func (store *Store) WithTx(ctx context.Context, fn func(*sqlc.Queries) error) error {
	return store.WithTxOptions(ctx, DefaultTxOptions(), fn)
}
