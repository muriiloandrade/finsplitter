package logctx

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestLogger returns a distinct logger that writes to io.Discard.
func newTestLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestFromCtx_ReturnsDefaultWhenNothingStored(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context
	}{
		{name: "empty context", ctx: context.Background()},
		{name: "nil context", ctx: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := FromCtx(tc.ctx)

			assert.Same(t, slog.Default(), logger)
		})
	}
}

func TestNewCtx_StoresLoggerAndReturnsCopy(t *testing.T) {
	testCases := []struct {
		name   string
		parent context.Context
	}{
		{name: "with parent", parent: context.Background()},
		{name: "nil parent falls back to background", parent: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := newTestLogger(t)

			ctx := NewCtx(tc.parent, logger)

			require.NotEqual(t, tc.parent, ctx, "NewCtx must return a derived context")
			assert.Same(t, logger, FromCtx(ctx))
		})
	}
}

func TestNewCtx_DoesNotMutateParent(t *testing.T) {
	parent := context.Background()
	_ = NewCtx(parent, newTestLogger(t))

	assert.Same(t, slog.Default(), FromCtx(parent))
}

func TestFromCtx_PropagatesThroughDerivedContexts(t *testing.T) {
	logger := newTestLogger(t)
	ctx := NewCtx(context.Background(), logger)

	// The logger must survive further context derivation and coexist
	// with other values stored under different keys.
	derived := context.WithValue(ctx, someOtherKey{}, "req-123")

	assert.Same(t, logger, FromCtx(derived))
}

func TestFromCtx_IgnoresNonLoggerValues(t *testing.T) {
	// A value stored under a different key must not satisfy the lookup.
	ctx := context.WithValue(context.Background(), someOtherKey{}, "req-123")

	assert.Same(t, slog.Default(), FromCtx(ctx))
}

// someOtherKey is a typed key used to prove loggers survive derivation
// alongside unrelated context values.
type someOtherKey struct{}
