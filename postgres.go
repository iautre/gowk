package gowk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

var (
	defaultPostgres atomic.Pointer[pgxpool.Pool]
	pgInitOnce      sync.Once
	pgInitErr       error
)

// initPostgres 同步建池并 Ping，只有真正连上才返回 nil。
// 行为由 DATABASE_ENABLED 决定：
//   - false：跳过初始化，Postgres(ctx) 恒为 nil；已配置的 DATABASE_DSN 被忽略并记一条 warn；
//   - true（默认）：DATABASE_DSN 必须非空且在 migrationConnectTimeout 内连接成功。
//
// 重试窗口与 connectForMigration 一致，让 DB 比应用晚就绪的编排（compose / K8s）
// 能自愈；超时仍连不上则返回错误，由 Run 决定是否退出进程。
func initPostgres() error {
	if !pgEnabled {
		if databaseDsn != "" {
			slog.Warn("DATABASE_ENABLED=false，已忽略配置的 DATABASE_DSN")
		}
		return nil
	}
	if databaseDsn == "" {
		return errors.New("PostgreSQL 已启用（DATABASE_ENABLED 默认为 true）但 DATABASE_DSN 为空：" +
			"请配置 DATABASE_DSN，或显式设置 DATABASE_ENABLED=false 关闭 PostgreSQL")
	}
	pgxConfig, err := pgxpool.ParseConfig(databaseDsn)
	if err != nil {
		return fmt.Errorf("解析 DATABASE_DSN 失败: %w", err)
	}
	pgxConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &PostgresLogger{},
		LogLevel: tracelog.LogLevelDebug,
	}

	ctx, cancel := context.WithTimeout(context.Background(), migrationConnectTimeout)
	defer cancel()

	start := time.Now()
	backoff := pgRetryBaseInterval
	var lastErr error
	for attempt := 1; ; attempt++ {
		pool, err := connectPool(ctx, pgxConfig)
		if err == nil {
			defaultPostgres.Store(pool)
			slog.Info("PostgreSQL 连接成功",
				"attempts", attempt,
				"elapsed", time.Since(start).Round(time.Millisecond))
			return nil
		}
		lastErr = err

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("连接 PostgreSQL 超时（%s，共 %d 次尝试）: %w（最后一次错误: %v）",
				migrationConnectTimeout, attempt, ctx.Err(), lastErr)
		case <-timer.C:
		}

		slog.Warn("PostgreSQL 连接失败，稍后重试",
			"attempt", attempt,
			"backoff", backoff.Round(time.Millisecond),
			"err", err)
		backoff *= 2
		if backoff > pgRetryMaxInterval {
			backoff = pgRetryMaxInterval
		}
	}
}

// connectPool 建一个连接池并用 Ping 验证可用；失败时释放池，避免重试累积连接。
func connectPool(ctx context.Context, cfg *pgxpool.Config) (*pgxpool.Pool, error) {
	pingCtx, cancel := context.WithTimeout(ctx, pgPingTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(pingCtx, cfg)
	if err != nil {
		// 防御未来 pgx 实现变化：同时返回 pool 与 err 时也要释放。
		if pool != nil {
			pool.Close()
		}
		return nil, err
	}
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
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

// Postgres 返回连接池。首次调用触发同步初始化（结果已被 sync.Once 缓存），
// 因此未调用过 InitPostgres 的代码首次进入这里可能阻塞至多 migrationConnectTimeout。
// 返回值语义：
//   - DATABASE_ENABLED=false → nil
//   - 启用且初始化失败 → nil，但 Run 早已在启动时退出进程，运行期不会走到
//   - 初始化成功 → 非 nil；之后池的生命周期与进程一致，不会回到 nil
//
// 调用方仍应检查 nil（例如 PostgresTx 的 "postgres unavailable"），不要对 nil 池调用方法。
func Postgres(ctx context.Context) *pgxpool.Pool {
	pgInitOnce.Do(func() { pgInitErr = initPostgres() })
	return defaultPostgres.Load()
}
