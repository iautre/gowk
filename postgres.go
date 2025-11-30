package gowk

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/tracelog"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var defaultPostgres atomic.Pointer[pgxpool.Pool]
var pgInitOnce sync.Once

func initPostgres() {
	if serverAddr == "" {
		return
	}
	pgxConfig, err := pgxpool.ParseConfig(serverAddr)
	if err != nil {
		slog.Error("PostgreSQL配置解析异常: %s  err:%v\n", serverAddr, err)
		return
	}
	pgxConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &PostgresLogger{},
		LogLevel: tracelog.LogLevelDebug, // 设置为 Debug 会打印所有 SQL
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		slog.Error("PostgreSQL连接池异常: %v\n", err)
		return
	}
	defaultPostgres.Store(pool)
}
func closePostgres() {
	if pool := defaultPostgres.Load(); pool != nil {
		pool.Close()
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
	pgInitOnce.Do(initPostgres)
	return defaultPostgres.Load()
}
