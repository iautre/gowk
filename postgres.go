package gowk

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

var defaultPostgres atomic.Pointer[pgxpool.Pool]
var pgInitOnce sync.Once

func initPostgres() {
	if databaseDsn == "" {
		return
	}
	pgxConfig, err := pgxpool.ParseConfig(databaseDsn)
	if err != nil {
		slog.Error("PostgreSQL配置解析异常", "dsn", databaseDsn, "err", err)
		return
	}
	pgxConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &PostgresLogger{},
		LogLevel: tracelog.LogLevelDebug,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		slog.Error("PostgreSQL连接池创建失败", "err", err)
		return
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("PostgreSQL Ping 失败", "err", err)
		pool.Close()
		return
	}
	defaultPostgres.Store(pool)
	slog.Info("PostgreSQL 连接成功")
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
			if err != nil {
				return nil, err
			}
			tx.Tx = pgTx
			return tx.Tx, nil
		}
	}
	return nil, errors.New("no tx")
}

// Postgres 返回连接池，通过 sync.Once 保证只初始化一次。
func Postgres(ctx context.Context) *pgxpool.Pool {
	pgInitOnce.Do(initPostgres)
	return defaultPostgres.Load()
}
