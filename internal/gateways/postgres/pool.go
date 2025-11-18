package postgres

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/exaring/otelpgx"
	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muriiloandrade/finsplitter/internal/config"
)

func NewPoolConfig(cfg config.Database) (*pgxpool.Config, error) {
	hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	dbConfig, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s&application_name=finsplitter",
		cfg.User, cfg.Password, hostPort, cfg.DBName, cfg.SSLMode,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	dbConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}
	dbConfig.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithIncludeQueryParameters())
	dbConfig.MaxConns = cfg.Pool.MaxConns
	dbConfig.MinConns = cfg.Pool.MinConns
	dbConfig.MaxConnLifetime = cfg.Pool.MaxConnLifetime
	dbConfig.MaxConnIdleTime = cfg.Pool.MaxConnIdleTime
	dbConfig.HealthCheckPeriod = cfg.Pool.HealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = cfg.Pool.ConnectTimeout

	return dbConfig, nil
}
