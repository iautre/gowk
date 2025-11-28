package gowk

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func DB(ctx context.Context) *pgxpool.Pool {
	return Postgres(ctx)
}

func initDB() {
	initPostgres()
}
