package gowk

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// migrationsTable 记录已应用的迁移版本。
const migrationsTable = "gowk_schema_migrations"

// migrationLockKey 是 pg_advisory_lock 的固定 key，保证多实例并发启动时迁移串行执行。
const migrationLockKey int64 = 1734023551

// migrationSource 是一个迁移来源：fsys 中 dir 目录下的 NNNN_name.sql 文件集合。
// 业务方通过 AddMigrations 注册（通常是 embed.FS）。
type migrationSource struct {
	fsys fs.FS
	dir  string
}

var migrationSources []migrationSource

// AddMigrations 注册一个迁移来源。fsys 通常是业务项目用 //go:embed 嵌入的 embed.FS，
// dir 是其中存放 NNNN_name.sql 的目录（embed 当前目录传 "."）。
//
// 文件名格式：版本号在前、下划线、描述，扩展名 .sql，例如 0001_baseline.sql、0002_add_avatar.sql。
// 版本号按数值升序执行；版本号在所有来源内必须唯一。
//
// 注册后：gowk.Run 启动时会在连接 DB 成功后、对外提供服务前自动执行未应用的迁移，
// 失败则进程退出（fail-fast）。未调用 AddMigrations 的服务，迁移功能完全不启用，Run 行为不变。
func AddMigrations(fsys fs.FS, dir string) {
	migrationSources = append(migrationSources, migrationSource{fsys: fsys, dir: dir})
}

// hasMigrations 报告是否注册过迁移来源。
func hasMigrations() bool { return len(migrationSources) > 0 }

type migration struct {
	version int64
	name    string
	sql     string
}

// loadMigrations 读取所有已注册来源里的迁移文件，按版本号升序返回，版本号重复时报错。
func loadMigrations() ([]migration, error) {
	var all []migration
	seen := make(map[int64]string)
	for _, src := range migrationSources {
		entries, err := fs.ReadDir(src.fsys, src.dir)
		if err != nil {
			return nil, fmt.Errorf("read migrations dir %q: %w", src.dir, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
				continue
			}
			version, name, err := parseMigrationName(e.Name())
			if err != nil {
				return nil, err
			}
			if prev, ok := seen[version]; ok {
				return nil, fmt.Errorf("duplicate migration version %d: %q and %q", version, prev, e.Name())
			}
			seen[version] = e.Name()
			path := e.Name()
			if src.dir != "" && src.dir != "." {
				path = src.dir + "/" + e.Name()
			}
			content, err := fs.ReadFile(src.fsys, path)
			if err != nil {
				return nil, fmt.Errorf("read migration %q: %w", e.Name(), err)
			}
			all = append(all, migration{version: version, name: name, sql: string(content)})
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].version < all[j].version })
	return all, nil
}

// parseMigrationName 从 NNNN_name.sql 解析版本号与描述。描述允许为空（如 0001.sql）。
func parseMigrationName(filename string) (int64, string, error) {
	base := strings.TrimSuffix(filename, ".sql")
	numStr := base
	name := ""
	if idx := strings.IndexByte(base, '_'); idx >= 0 {
		numStr = base[:idx]
		name = base[idx+1:]
	}
	v, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid migration filename %q (expect NNNN_name.sql): %w", filename, err)
	}
	if v <= 0 {
		return 0, "", fmt.Errorf("invalid migration version in %q: version must be > 0", filename)
	}
	return v, name, nil
}

// runMigrations 执行所有未应用的迁移。流程：建独立连接 → advisory lock → 建版本表 →
// 逐个在事务内执行未应用迁移并记录版本。无注册来源时直接返回 nil。
func runMigrations(ctx context.Context) error {
	if !hasMigrations() {
		return nil
	}
	if databaseDsn == "" {
		return errors.New("migrations registered but DATABASE_DSN is empty")
	}
	migs, err := loadMigrations()
	if err != nil {
		return err
	}
	if len(migs) == 0 {
		return nil
	}

	conn, err := connectForMigration(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		if _, err := conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", migrationLockKey); err != nil {
			slog.Warn("释放迁移锁失败", "err", err)
		}
	}()

	if _, err := conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS "+migrationsTable+
		" (version bigint PRIMARY KEY, name text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())"); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	applied, err := loadAppliedVersions(ctx, conn)
	if err != nil {
		return err
	}

	pending := 0
	for _, m := range migs {
		if applied[m.version] {
			continue
		}
		if err := applyMigration(ctx, conn, m); err != nil {
			return fmt.Errorf("apply migration %04d_%s: %w", m.version, m.name, err)
		}
		pending++
		slog.Info("迁移已应用", "version", m.version, "name", m.name)
	}
	if pending == 0 {
		slog.Info("数据库迁移：无待应用项", "total", len(migs))
	} else {
		slog.Info("数据库迁移完成", "applied", pending, "total", len(migs))
	}
	return nil
}

// applyMigration 在单个事务里执行迁移 SQL 并记录版本，任一步失败回滚。
func applyMigration(ctx context.Context, conn *pgx.Conn, m migration) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, m.sql); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		"INSERT INTO "+migrationsTable+" (version, name) VALUES ($1, $2)", m.version, m.name); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// loadAppliedVersions 读取已应用的版本集合。
func loadAppliedVersions(ctx context.Context, conn *pgx.Conn) (map[int64]bool, error) {
	rows, err := conn.Query(ctx, "SELECT version FROM "+migrationsTable)
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()
	applied := make(map[int64]bool)
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// connectForMigration 建立用于迁移的独立连接，在 migrationConnectTimeout 内带重试，连不上则返回错误。
//
// 用 pgxpool.ParseConfig 解析 DSN（与 gowk 主连接池一致），它能识别并剥离 pool_max_conns 等
// pgxpool 专属参数；取其 ConnConfig 建单连接。若直接用 pgx.Connect(dsn)，pool_* 参数会被当作
// PG 服务器配置发送，触发 FATAL: unrecognized configuration parameter。
func connectForMigration(ctx context.Context) (*pgx.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, migrationConnectTimeout)
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(databaseDsn)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_DSN for migration: %w", err)
	}
	connCfg := poolCfg.ConnConfig

	var lastErr error
	for {
		conn, err := pgx.ConnectConfig(ctx, connCfg)
		if err == nil {
			if perr := conn.Ping(ctx); perr == nil {
				return conn, nil
			} else {
				lastErr = perr
				conn.Close(ctx)
			}
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("connect for migration: %w (last error: %v)", ctx.Err(), lastErr)
		case <-time.After(pgRetryBaseInterval):
		}
	}
}

// migrationStatus 打印各迁移的应用状态，供 `server migrate-status` 子命令使用。返回退出码。
func migrationStatus(ctx context.Context) int {
	if !hasMigrations() {
		slog.Info("未注册任何迁移来源")
		return 0
	}
	migs, err := loadMigrations()
	if err != nil {
		slog.Error("加载迁移失败", "err", err)
		return 1
	}
	conn, err := connectForMigration(ctx)
	if err != nil {
		slog.Error("连接数据库失败", "err", err)
		return 1
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS "+migrationsTable+
		" (version bigint PRIMARY KEY, name text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())"); err != nil {
		slog.Error("创建迁移表失败", "err", err)
		return 1
	}
	applied, err := loadAppliedVersions(ctx, conn)
	if err != nil {
		slog.Error("读取已应用版本失败", "err", err)
		return 1
	}
	for _, m := range migs {
		state := "pending"
		if applied[m.version] {
			state = "applied"
		}
		slog.Info("迁移状态", "version", m.version, "name", m.name, "state", state)
	}
	return 0
}
