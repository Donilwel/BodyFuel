package postgres

import (
	"backend/pkg/logging"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

type dbClient interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type dbClientGetter struct {
	db *sqlx.DB
}

func (g dbClientGetter) Get(ctx context.Context) dbClient { //nolint:ireturn // it's necessary to return abstraction for this case
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}

	return g.db
}

type TransactionManager struct {
	db *sqlx.DB
}

func NewTransactionManager(db *sqlx.DB) *TransactionManager {
	return &TransactionManager{
		db: db,
	}
}

func (txm *TransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) (happenedErr error) {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before beginning transaction: %w", err)
	}

	tx, happenedErr := txm.db.BeginTxx(ctx, nil)
	if happenedErr != nil {
		return fmt.Errorf("begin transaction: %w", happenedErr)
	}
	defer func() {
		if happenedErr != nil {
			if err := tx.Rollback(); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				logging.GetLoggerFromContext(ctx).Errorf("rollback transaction: %v", err)
			}
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if happenedErr = fn(txCtx); happenedErr != nil {
		return fmt.Errorf("call inner func: %w", happenedErr)
	}

	if happenedErr = tx.Commit(); happenedErr != nil {
		return fmt.Errorf("commit transaction: %w", happenedErr)
	}

	return nil
}
