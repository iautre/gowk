package log

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/exp/slog"

	gormLog "gorm.io/gorm/logger"
)

var (
	std   = New()
	level = slog.LevelError
)

type H map[string]any

const skip = 3
const (
	TraceId string = "traceId"
	SpanId  string = "spanId"
	PspanId string = "pspanId"
)

func New() *slog.Logger {
	handler := slog.HandlerOptions{
		AddSource: true,
	}
	textHandler := handler.NewTextHandler(os.Stdout)
	return slog.New(textHandler)
}

func GetContextAttrs(ctx context.Context) []slog.Attr {
	var traceId, pspanId, spanId string
	if ctx.Value(TraceId) != nil {
		traceId = ctx.Value(TraceId).(string)
	}
	if ctx.Value(PspanId) != nil {
		pspanId = ctx.Value(PspanId).(string)
	}
	if ctx.Value(SpanId) != nil {
		spanId = ctx.Value(SpanId).(string)
	}
	return []slog.Attr{
		{Key: "traceId", Value: slog.StringValue(traceId)},
		{Key: "pspanId", Value: slog.StringValue(pspanId)},
		{Key: "spanId", Value: slog.StringValue(spanId)},
	}
}

type GromLogger struct{}

func (gl *GromLogger) LogMode(logLevel gormLog.LogLevel) gormLog.Interface {
	return &GromLogger{}
}
func (gl *GromLogger) Info(ctx context.Context, format string, args ...interface{}) {
	std.LogAttrsDepth(defaultDepth, slog.LevelInfo, fmt.Sprintf(format, args...), GetContextAttrs(ctx)...)
}
func (gl *GromLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	std.LogAttrsDepth(defaultDepth, slog.LevelWarn, fmt.Sprintf(format, args...), GetContextAttrs(ctx)...)
}
func (gl *GromLogger) Error(ctx context.Context, format string, args ...interface{}) {
	std.LogAttrsDepth(defaultDepth, slog.LevelError, fmt.Sprintf(format, args...), GetContextAttrs(ctx)...)
}
func (gl *GromLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	nownow := time.Now()
	usedTime := nownow.Sub(begin).Milliseconds()
	sql, rowsAffected := fc()
	msg := fmt.Sprintf("sql:[%s] rows:[%d] %dms", sql, rowsAffected, usedTime)
	std.LogAttrsDepth(defaultDepth, slog.LevelInfo, msg, GetContextAttrs(ctx)...)
}

type GinLogger struct{}
