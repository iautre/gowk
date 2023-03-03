package log

import (
	"context"
	"os"

	"golang.org/x/exp/slog"
)

func Logger() *slog.Logger {
	options := slog.HandlerOptions{
		AddSource: true,
	}
	return slog.New(options.NewJSONHandler(os.Stderr))
}

type TextHandler struct{}

func (h *TextHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *TextHandler) Handle(ctx context.Context, r slog.Record) error {
	return nil
}

func (h *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *TextHandler) WithGroup(name string) slog.Handler {
	return h
}

func getTraceId(ctx context.Context) []any {
	return []any{
		TraceId, ctx.Value(TraceId).(string),
		SpanId, ctx.Value(SpanId).(string),
		PspanId, ctx.Value(SpanId).(string),
	}
}
