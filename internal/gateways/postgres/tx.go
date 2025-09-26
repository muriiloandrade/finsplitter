package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muriiloandrade/finsplitter/internal/domain"
)

type txKey struct{}

type TxManager struct {
	ConnPool *pgxpool.Pool
}

func (b TxManager) WithTx(ctx context.Context, f domain.TransactionFunc) error {
	if domain.HasTX(ctx) {
		return &domain.TransactionError{Cause: errors.New("already in transaction")}
	}

	tx, err := b.ConnPool.Begin(ctx)
	if err != nil {
		return &domain.TransactionError{Cause: fmt.Errorf("cannot begin a transaction: %w", err)}
	}

	ctxWithTx := context.WithValue(domain.WithTx(ctx), txKey{}, tx)

	defer func() {
		if p := recover(); p != nil {
			// ensure a rollback attempt and panic again
			_ = tx.Rollback(ctx)

			panic(p)
		}
	}()

	if err = f(ctxWithTx); err != nil {
		if rollBackErr := tx.Rollback(ctx); rollBackErr != nil {
			return &domain.TransactionError{
				Cause: fmt.Errorf("rollback failed after transaction error: %w (original: %w)", rollBackErr, err),
			}
		}
		return &domain.TransactionError{Cause: err}
	}

	if err = tx.Commit(ctx); err != nil {
		return &domain.TransactionError{Cause: fmt.Errorf("cannot commit transaction: %w", err)}
	}

	return nil
}

// querier should be used when either a transaction or a common connection could be used.
type querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

func (b TxManager) querier(ctx context.Context) querier {
	if btx, ok := ctx.Value(txKey{}).(*pgxpool.Tx); ok {
		return btx
	}

	return b.ConnPool
}

func (b TxManager) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (pgconn.CommandTag, error) {
	return b.querier(ctx).Exec(ctx, query, args...)
}

func (b TxManager) Query(
	ctx context.Context,
	query string,
	args ...any,
) (pgx.Rows, error) {
	return b.querier(ctx).Query(ctx, query, args...)
}

func (b TxManager) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return b.querier(ctx).QueryRow(ctx, query, args...)
}

func (b TxManager) SendBatch(ctx context.Context, batch *pgx.Batch) pgx.BatchResults {
	return b.querier(ctx).SendBatch(ctx, batch)
}

func (b TxManager) Close() {
	b.ConnPool.Close()
}
