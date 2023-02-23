package log

import (
	"context"
	"fmt"

	"golang.org/x/exp/slog"
)

const defaultDepth = 1

func SetLevel(level slog.Level) {
	// level = level
}

// 日志使用
func Errorf(ctx context.Context, format string, a ...any) {
	std.LogAttrsDepth(defaultDepth, slog.LevelError, fmt.Sprintf(format, a...), GetContextAttrs(ctx)...)
}
func Warnf(ctx context.Context, format string, a ...any) {
	std.LogAttrsDepth(defaultDepth, slog.LevelWarn, fmt.Sprintf(format, a...), GetContextAttrs(ctx)...)
}
func Infof(ctx context.Context, format string, a ...any) {
	std.LogAttrsDepth(defaultDepth, slog.LevelInfo, fmt.Sprintf(format, a...), GetContextAttrs(ctx)...)
}
func Debugf(ctx context.Context, format string, a ...any) {
	std.LogAttrsDepth(defaultDepth, slog.LevelDebug, fmt.Sprintf(format, a...), GetContextAttrs(ctx)...)
}
func Tracef(ctx context.Context, format string, a ...any) {
	std.LogAttrsDepth(defaultDepth, slog.LevelInfo, fmt.Sprintf(format, a...), GetContextAttrs(ctx)...)
}

func Error(ctx context.Context, a any) {
	std.LogAttrsDepth(defaultDepth, slog.LevelError, fmt.Sprintf("%v", a), GetContextAttrs(ctx)...)
}
func Warn(ctx context.Context, a any) {
	std.LogAttrsDepth(defaultDepth, slog.LevelWarn, fmt.Sprintf("%v", a), GetContextAttrs(ctx)...)
}
func Info(ctx context.Context, a any) {
	std.LogAttrsDepth(defaultDepth, slog.LevelInfo, fmt.Sprintf("%v", a), GetContextAttrs(ctx)...)
}
func Debug(ctx context.Context, a any) {
	std.LogAttrsDepth(defaultDepth, slog.LevelDebug, fmt.Sprintf("%v", a), GetContextAttrs(ctx)...)
}
func Trace(ctx context.Context, a any, arrt ...slog.Attr) {
	arrt = append(arrt, GetContextAttrs(ctx)...)
	std.LogAttrsDepth(2, slog.LevelInfo, fmt.Sprintf("%v", a), arrt...)
}
