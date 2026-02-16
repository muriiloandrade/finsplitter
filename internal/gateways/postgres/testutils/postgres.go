package testutils

import (
	"context"
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

// TestDB holds the testcontainer and connection pool
type TestDB struct {
	ConnPool  *pgxpool.Pool
	Container *tcpgres.PostgresContainer
}

// StartTestDB starts a PostgreSQL container using Testcontainers.
// This is a lower-level function for advanced use cases where you need manual control.
// For most tests, use NewTestDB instead.
func StartTestDB(ctx context.Context) (*TestDB, error) {
	// Start PostgreSQL container with proper wait strategy
	pgContainer, err := tcpgres.Run(ctx,
		"postgres:18.2-trixie",
		tcpgres.WithDatabase(testDBName),
		tcpgres.WithUsername(testDBUser),
		tcpgres.WithPassword(testDBPassword),
		tcpgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Get connection string from container
	connStr, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	// Create pool config using existing factory
	dbConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Connect to the database
	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Run migrations
	err = runMigrationsWithPool(pool)
	if err != nil {
		pool.Close()
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &TestDB{
		ConnPool:  pool,
		Container: pgContainer,
	}, nil
}

// runMigrationsWithPool runs migrations using the existing RunMigrations function
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

// findProjectRoot finds the project root directory by looking for go.mod
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Walk up the directory tree looking for go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
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

// Pool returns the connection pool (for convenience)
func (tdb *TestDB) Pool() *pgxpool.Pool {
	return tdb.ConnPool
}

// Close terminates the container and closes the pool
func (tdb *TestDB) Close(ctx context.Context) error {
	if tdb.ConnPool != nil {
		tdb.ConnPool.Close()
	}
	if tdb.Container != nil {
		return tdb.Container.Terminate(ctx)
	}
	return nil
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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := StartTestDB(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := container.Close(ctx); err != nil {
			t.Logf("Failed to close test container: %v", err)
		}
	})

	return container
}
