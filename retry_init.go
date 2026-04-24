package gowk

import (
	"context"
	"log/slog"
	"time"
)

// retryBackground 以指数退避无限重试 attempt，直到成功或 ctx 被取消。
// 用于 PostgreSQL / Redis 等外部依赖的后台自愈：启动时不阻塞，
// 连接失败不退出进程，后台持续重试到恢复；期间外部访问看到 nil，
// 由使用侧（例如 PostgresTx）翻译成业务错误。
//
// backoff 语义：初始 sleep = base，每轮失败后 backoff *= 2，最终被 max 封顶；
// max < base 时内部会把 max 抬到 base，退化为固定间隔 base。
// 进程退出时调用方 cancel(ctx)，当前 sleep 与后续 attempt 立刻结束。
func retryBackground(ctx context.Context, name string, base, max time.Duration, attempt func(context.Context) error) {
	if base <= 0 {
		base = 2 * time.Second
	}
	if max <= 0 {
		max = base
	}
	if max < base {
		max = base
	}
	start := time.Now()
	backoff := base
	for attemptN := 1; ; attemptN++ {
		if err := ctx.Err(); err != nil {
			slog.Info(name+" 后台重试已取消",
				"attempts", attemptN-1,
				"elapsed", time.Since(start).Round(time.Millisecond))
			return
		}
		err := attempt(ctx)
		if err == nil {
			slog.Info(name+" 连接成功",
				"attempts", attemptN,
				"elapsed", time.Since(start).Round(time.Millisecond))
			return
		}
		slog.Warn(name+" 连接失败，稍后重试",
			"attempt", attemptN,
			"elapsed", time.Since(start).Round(time.Millisecond),
			"backoff", backoff.Round(time.Millisecond),
			"err", err)
		timer := time.NewTimer(backoff)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			slog.Info(name+" 后台重试已取消",
				"attempts", attemptN,
				"elapsed", time.Since(start).Round(time.Millisecond))
			return
		}
		backoff *= 2
		if backoff > max {
			backoff = max
		}
	}
}
