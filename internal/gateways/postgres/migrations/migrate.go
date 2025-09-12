package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	migratePgx "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Import file source driver
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muriiloandrade/finsplitter/internal/config"
	slogctx "github.com/veqryn/slog-context"
)

type MigrationOptions struct {
	DBInstance     *pgxpool.Pool
	DBCfg          config.Database
	MigrationsPath string
}

// RunMigrations applies all available up migrations.
func RunMigrations(ctx context.Context, opts MigrationOptions) error {
	logger := slogctx.FromCtx(ctx)

	logger.Info("Starting database migrations...")
	logger.Debug(
		"Database connection string",
		slog.String("conn_string", opts.DBInstance.Config().ConnString()),
	)

	db, err := sql.Open("pgx", opts.DBInstance.Config().ConnString())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Setup database driver instance
	dbDriver, err := migratePgx.WithInstance(db, &migratePgx.Config{
		SchemaName: opts.DBCfg.Schema,
	})
	if err != nil {
		return fmt.Errorf("failed to create migrate pgx database driver instance: %w", err)
	}
	defer dbDriver.Close()

	// Setup source driver instance
	// The URL format for the file source is file://<path>
	sourceURL := fmt.Sprintf("file://%s", opts.MigrationsPath)

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"pgx", // Specify the database driver name
		dbDriver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Apply migrations
	logger.Info(
		"Applying migrations from migrations path",
		slog.String("migrations_path", opts.MigrationsPath),
	)
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Check for migration errors (e.g., dirty state)
	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		return fmt.Errorf("migration source error: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("migration database error: %w", dbErr)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		logger.Info("Database is up to date, no new migrations to apply.")
	} else {
		logger.Info("Database migrations applied successfully.")
	}

	return nil
}
