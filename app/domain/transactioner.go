package domain

import (
	"context"
)

type txExistsKey struct{}

type TransactionError struct {
	Cause error
}

func (t *TransactionError) Error() string {
	return "transaction error: " + t.Cause.Error()
}

func (t *TransactionError) TransactionErrCause() string {
	return t.Cause.Error()
}

type TransactionFunc func(ctx context.Context) error

type Transactioner interface {
	WithTx(context.Context, TransactionFunc) error
}

func HasTX(ctx context.Context) bool {
	if b, ok := ctx.Value(txExistsKey{}).(bool); ok {
		return b
	}

	return false
}

func WithTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, txExistsKey{}, true)
}
