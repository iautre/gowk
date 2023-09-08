package log

import (
	"context"
	"fmt"
	"os"

	"log/slog"
)

func Logger() *slog.Logger {
	options := &slog.HandlerOptions{
		AddSource: true,
	}
	return slog.New(slog.NewTextHandler(os.Stderr, options))
}

type TextHandler struct {
	Level slog.Level
}

func (h *TextHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return l <= h.Level
}

func (h *TextHandler) Handle(ctx context.Context, r slog.Record) error {
	fmt.Printf("")
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
