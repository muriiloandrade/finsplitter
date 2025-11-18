package domain

import (
	"context"
)

type txExistsKey struct{}

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
