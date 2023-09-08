package log

import (
	"context"

	"log/slog"
)

func Error(ctx context.Context, msg string, err error, arr ...any) {
	arr = append(arr, getTraceId(ctx)...)
	if err != nil {
		arr = append(arr, err.Error())
	}
	slog.Error(msg, arr...)
}
func Info(ctx context.Context, msg string, arr ...any) {
	arr = append(arr, getTraceId(ctx)...)
	slog.Info(msg, arr...)
}
func Trace(ctx context.Context, msg string, arr ...any) {
	arr = append(arr, getTraceId(ctx)...)
	slog.Info(msg, arr...)
}
func Debug(ctx context.Context, msg string, arr ...any) {
	arr = append(arr, getTraceId(ctx)...)
	slog.Debug(msg, arr...)
}
