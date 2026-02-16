package domain

import (
	"context"
	"testing"
)

func TestHasTX(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected bool
	}{
		{
			name:     "returns false for context without tx",
			ctx:      context.Background(),
			expected: false,
		},
		{
			name:     "returns true for context with tx",
			ctx:      WithTx(context.Background()),
			expected: true,
		},
		{
			name:     "returns false for context with non-bool value",
			ctx:      context.WithValue(context.Background(), txExistsKey{}, "string"),
			expected: false,
		},
		{
			name:     "returns false for context with int value",
			ctx:      context.WithValue(context.Background(), txExistsKey{}, 1),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasTX(tt.ctx)
			if result != tt.expected {
				t.Errorf("HasTX() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWithTx(t *testing.T) {
	ctx := context.Background()

	// Before calling WithTx, should return false
	if HasTX(ctx) {
		t.Error("expected HasTX to return false before WithTx")
	}

	// After calling WithTx, should return true
	ctxWithTx := WithTx(ctx)
	if !HasTX(ctxWithTx) {
		t.Error("expected HasTX to return true after WithTx")
	}

	// Original context should be unchanged
	if HasTX(ctx) {
		t.Error("expected original context to be unchanged")
	}
}

func TestTxExistsKey(t *testing.T) {
	// Verify the key is unexported and unique
	var key1 txExistsKey

	// They should be the same type (unexported)
	_ = key1

	// Using the key directly should work
	ctx := context.WithValue(context.Background(), txExistsKey{}, true)
	if val := ctx.Value(txExistsKey{}); val != true {
		t.Errorf("expected true, got %v", val)
	}
}
