package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	DefaultMaxConnLifetime   = 10 * time.Minute
	DefaultMaxConnIdleTime   = time.Minute
	DefaultHealthCheckPeriod = 30 * time.Second
	DefaultMaxConnCount      = 100
	DefaultMinConnCount      = 10
)

func PgxConnPool(
	ctx context.Context,
	address string,
	username string, password string,
	databaseName string,
) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(
		buildConnectionString(address, username, password, databaseName),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	setUpConnectionConfig(config)

	pgConnPool, err := pgxpool.ConnectConfig(ctx, config)

	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return pgConnPool, nil
}

func setUpConnectionConfig(config *pgxpool.Config) {
	config.MaxConnLifetime = DefaultMaxConnLifetime
	config.MaxConnIdleTime = DefaultMaxConnIdleTime
	config.HealthCheckPeriod = DefaultHealthCheckPeriod
	config.MaxConns = DefaultMaxConnCount
	config.MinConns = DefaultMinConnCount
}

func buildConnectionString(
	address string, username string, password string, databaseName string,
) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		username, password, address, databaseName,
	)
}
