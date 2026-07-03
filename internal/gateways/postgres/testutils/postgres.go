package testutils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muriiloandrade/finsplitter/internal/config"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres/migrations"
	tcpgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	testDBName     = "finsplitter"
	testDBUser     = "finsplitter"
	testDBPassword = "finsplitter"
)

// TestDB holds the testcontainer and connection pool.
type TestDB struct {
	ConnPool  *pgxpool.Pool
	Container *tcpgres.PostgresContainer
}

// NewTestDB creates a new test database with automatic cleanup.
// This is the preferred function for most tests.
//
// Usage:
//
//	func TestMyRepo(t *testing.T) {
//	    db := testutils.NewTestDB(t)
//	    repo := myrepo.New(db.Pool())
//	    // ... test code
//	}
//
// The container is automatically terminated when the test completes.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping testcontainers test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := StartTestDB(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if closeErr := container.Close(cleanupCtx); closeErr != nil {
			t.Logf("Failed to close test container: %v", closeErr)
		}
	})

	return container
}

// Pool returns the connection pool (for convenience).
func (tdb *TestDB) Pool() *pgxpool.Pool {
	return tdb.ConnPool
}

// Close terminates the container and closes the pool.
func (tdb *TestDB) Close(ctx context.Context) error {
	if tdb.ConnPool != nil {
		tdb.ConnPool.Close()
	}
	if tdb.Container != nil {
		return tdb.Container.Terminate(ctx)
	}
	return nil
}

// StartTestDB starts a PostgreSQL container using Testcontainers.
// This is a lower-level function for advanced use cases where you need manual control.
// For most tests, use NewTestDB instead.
func StartTestDB(ctx context.Context) (*TestDB, error) {
	// Start PostgreSQL container with proper wait strategy
	pgContainer, err := tcpgres.Run(ctx,
		"postgres:18.4-trixie",
		tcpgres.WithDatabase(testDBName),
		tcpgres.WithUsername(testDBUser),
		tcpgres.WithPassword(testDBPassword),
		tcpgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Get connection string from container
	connStr, connErr := pgContainer.ConnectionString(ctx)
	if connErr != nil {
		if tErr := pgContainer.Terminate(ctx); tErr != nil {
			connErr = errors.Join(connErr, tErr)
		}
		return nil, fmt.Errorf("failed to get connection string: %w", connErr)
	}

	// Create pool config using existing factory
	dbConfig, poolCfgErr := pgxpool.ParseConfig(connStr)
	if poolCfgErr != nil {
		if tErr := pgContainer.Terminate(ctx); tErr != nil {
			poolCfgErr = errors.Join(poolCfgErr, tErr)
		}
		return nil, fmt.Errorf("failed to parse database config: %w", poolCfgErr)
	}

	// Connect to the database
	pool, poolErr := pgxpool.NewWithConfig(ctx, dbConfig)
	if poolErr != nil {
		if tErr := pgContainer.Terminate(ctx); tErr != nil {
			poolErr = errors.Join(poolErr, tErr)
		}
		return nil, fmt.Errorf("failed to create connection pool: %w", poolErr)
	}

	// Run migrations
	err = runMigrationsWithPool(pool)
	if err != nil {
		pool.Close()
		if tErr := pgContainer.Terminate(ctx); tErr != nil {
			err = errors.Join(err, tErr)
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &TestDB{
		ConnPool:  pool,
		Container: pgContainer,
	}, nil
}

// runMigrationsWithPool runs migrations using the existing RunMigrations function.
func runMigrationsWithPool(pool *pgxpool.Pool) error {
	// Use the existing migrations package
	// Build the database config for migrations
	dbCfg := config.Database{
		Schema: "public",
	}

	// Find project root by looking for go.mod
	projectRoot := findProjectRoot()
	migrationsPath := filepath.Join(projectRoot, "internal", "gateways", "postgres", "migrations")

	err := migrations.RunMigrations(context.Background(), migrations.MigrationOptions{
		DBInstance:     pool,
		DBCfg:          dbCfg,
		MigrationsPath: migrationsPath,
	})
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// findProjectRoot finds the project root directory by looking for go.mod.
func findProjectRoot() string {
	dir, dirErr := os.Getwd()
	if dirErr != nil {
		return "."
	}

	// Walk up the directory tree looking for go.mod
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}

		// Go up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
