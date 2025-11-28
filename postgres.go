package gowk

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/tracelog"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var default_postgres *pgxpool.Pool

func initPostgres() {
	pgxConfig, err := pgxpool.ParseConfig(DATABASE_DSN)
	if err != nil {
		log.Fatalf("PostgreSQL配置解析异常: %s  err:%v\n", DATABASE_DSN, err)
	}
	pgxConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &PostgresLogger{},
		LogLevel: tracelog.LogLevelDebug, // 设置为 Debug 会打印所有 SQL
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		log.Fatalf("PostgreSQL连接池异常: %v\n", err)
	}
	default_postgres = pool
}
func closePostgres() {
	if default_postgres != nil {
		default_postgres.Close()
	}
}

func PostgresTx(ctx context.Context) (pgx.Tx, error) {
	if tx, ok := ctx.Value(TRANSACTION).(*Transaction); ok && tx != nil {
		if tx.Tx != nil {
			return tx.Tx, nil
		}
		if tx.Begin {
			pgTx, err := Postgres(ctx).Begin(ctx)
			tx.Tx = pgTx
			return tx.Tx, err
		}
	}
	return nil, errors.New("no tx")
}

func Postgres(ctx context.Context) *pgxpool.Pool {
	if default_postgres == nil {
		initPostgres()
	}
	return default_postgres
}
