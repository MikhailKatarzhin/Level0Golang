package postgre

import (
	"context"
	"fmt"

	postgres "github.com/MikhailKatarzhin/Level0Golang/pkg/database"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	dbAddr = "localhost:5433"
	dbName = "wbl0db"
	dbPass = "Useriuswbl0db"
	dbUser = "Userius"
)

func DefaultCredConfig() (*pgxpool.Pool, error) {
	pgConnPool, err := postgres.PgxConnPool(context.Background(), dbAddr, dbUser, dbPass, dbName)

	if err != nil {
		return nil, fmt.Errorf("failed connection: %w", err)
	}

	return pgConnPool, nil
}
