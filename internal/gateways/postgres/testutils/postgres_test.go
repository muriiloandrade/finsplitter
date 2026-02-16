package testutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartTestDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// This test verifies that Testcontainers can spin up a PostgreSQL instance
	// and run migrations successfully
	ctx := context.Background()

	db, err := StartTestDB(ctx)
	require.NoError(t, err, "Failed to start test DB")
	require.NotNil(t, db, "DB should not be nil")
	require.NotNil(t, db.ConnPool, "Connection pool should not be nil")
	require.NotNil(t, db.Container, "Container should not be nil")

	// Verify we can ping the database
	err = db.ConnPool.Ping(ctx)
	require.NoError(t, err, "Failed to ping database")

	// Verify we can query the database
	var result int
	err = db.ConnPool.QueryRow(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err, "Failed to execute query")
	assert.Equal(t, 1, result, "Query should return 1")

	// Verify migrations ran successfully by checking for a known table.
	// This confirms the migration suite executed without errors.
	// Note: card_brand is the first table created in the migration sequence.
	err = db.ConnPool.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'card_brand'").Scan(&result)
	require.NoError(t, err, "Failed to check for card_brand table")
	assert.Equal(t, 1, result, "card_brand table should exist")

	// Clean up
	err = db.Close(ctx)
	assert.NoError(t, err, "Failed to close test DB")
}
