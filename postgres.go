package gowk

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

var (
	defaultPostgres atomic.Pointer[pgxpool.Pool]
	pgInitOnce      sync.Once
	pgRetryCancel   context.CancelFunc
)

func initPostgres() {
	// DSN 未配置属于"没启用"，保持老行为直接返回，
	// 业务侧靠 Postgres(ctx) == nil / PostgresTx 的错误判断降级。
	if databaseDsn == "" {
		return
	}
	pgxConfig, err := pgxpool.ParseConfig(databaseDsn)
	if err != nil {
		// DSN 语法错误后台再怎么重试也是同一个错，直接降级并记一条错误，避免刷屏。
		slog.Error("PostgreSQL 配置解析异常，保持降级", "err", err)
		return
	}
	pgxConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &PostgresLogger{},
		LogLevel: tracelog.LogLevelDebug,
	}

	ctx, cancel := context.WithCancel(context.Background())
	pgRetryCancel = cancel
	go retryBackground(ctx, "PostgreSQL", pgRetryBaseInterval, pgRetryMaxInterval, func(c context.Context) error {
		pingCtx, cancelPing := context.WithTimeout(c, pgPingTimeout)
		defer cancelPing()
		p, err := pgxpool.NewWithConfig(pingCtx, pgxConfig)
		if err != nil {
			// 防御未来 pgx 实现变化：同时返回 pool 与 err 时也要释放。
			if p != nil {
				p.Close()
			}
			return err
		}
		if err := p.Ping(pingCtx); err != nil {
			p.Close()
			return err
		}
		defaultPostgres.Store(p)
		return nil
	})
}

func closePostgres() {
	if pgRetryCancel != nil {
		pgRetryCancel()
	}
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
			pool := Postgres(ctx)
			if pool == nil {
				return nil, errors.New("postgres unavailable")
			}
			pgTx, err := pool.Begin(ctx)
			if err != nil {
				return nil, err
			}
			tx.Tx = pgTx
			return tx.Tx, nil
		}
	}
	return nil, errors.New("no tx")
}

// Postgres 返回连接池，通过 sync.Once 保证 initPostgres 只触发一次后台重试。
// 返回值语义：
//   - DATABASE_DSN 未配置 / 配置错误 / 后台尚未连上 → nil
//   - 后台连接成功后 → 非 nil 连接池
//
// 调用方应在业务层检查 nil 并返回错误（例如 PostgresTx 的 "postgres unavailable"），
// 不要对 nil 池直接调用方法。
func Postgres(ctx context.Context) *pgxpool.Pool {
	pgInitOnce.Do(initPostgres)
	return defaultPostgres.Load()
}
