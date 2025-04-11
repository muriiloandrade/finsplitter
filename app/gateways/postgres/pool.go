package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muriiloandrade/finsplitter/app/config"
)

func NewPoolConfig(cfg config.Database) (*pgxpool.Config, error) {
	dbConfig, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&application_name=finsplitter",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DbName, cfg.SSLMode,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	dbConfig.MaxConns = cfg.Pool.MaxConns
	dbConfig.MinConns = cfg.Pool.MinConns
	dbConfig.MaxConnLifetime = cfg.Pool.MaxConnLifetime
	dbConfig.MaxConnIdleTime = cfg.Pool.MaxConnIdleTime
	dbConfig.HealthCheckPeriod = cfg.Pool.HealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = cfg.Pool.ConnectTimeout

	return dbConfig, nil
}
