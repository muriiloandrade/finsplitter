package logctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRequestID_ReturnsEmptyWhenNothingStored(t *testing.T) {
	assert.Empty(t, GetRequestID(context.Background()))
}

func TestGetRequestID_ReturnsEmptyForNilContext(t *testing.T) {
	//nolint:staticcheck // deliberately exercising the nil-context guard
	assert.Empty(t, GetRequestID(nil))
}

func TestWithRequestID_StoresAndRetrieves(t *testing.T) {
	parent := context.Background()

	ctx := WithRequestID(parent, "req-123")

	require.NotEqual(t, parent, ctx, "WithRequestID must return a derived context")
	assert.Equal(t, "req-123", GetRequestID(ctx))
}

func TestWithRequestID_DoesNotMutateParent(t *testing.T) {
	parent := context.Background()
	_ = WithRequestID(parent, "req-123")

	assert.Empty(t, GetRequestID(parent))
}

func TestGetRequestID_PropagatesThroughDerivedContexts(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-123")

	// The request ID must survive further context derivation.
	derived := context.WithValue(ctx, someOtherKey{}, "value")

	assert.Equal(t, "req-123", GetRequestID(derived))
}

func TestGetRequestID_IgnoresNonStringValues(t *testing.T) {
	ctx := context.WithValue(context.Background(), RequestIDKey, 42)

	assert.Empty(t, GetRequestID(ctx))
}
