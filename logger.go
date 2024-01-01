package gowk

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"log/slog"

	gormLogger "gorm.io/gorm/logger"
)

func Logger(l slog.Level) *slog.Logger {
	// options := &slog.HandlerOptions{
	// 	AddSource: true,
	// }
	// return slog.New(slog.NewTextHandler(os.Stderr, options))
	return slog.New(NewTestHandler(os.Stderr))
}

type TextHandler struct {
	Level slog.Level
	w     io.Writer
}

func NewTestHandler(w io.Writer) *TextHandler {
	return &TextHandler{
		w: w,
	}
}

func (h *TextHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return l <= h.Level
}

func (h *TextHandler) Handle(ctx context.Context, r slog.Record) error {

	source := source(r.PC)
	fmt.Printf(source.File, source.Function, source.Line, r.Message)
	fmt.Println("")
	return nil
}

func (h *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *TextHandler) WithGroup(name string) slog.Handler {
	return h
}

func source(pc uintptr) *slog.Source {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()
	return &slog.Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}

type GromLogger struct {
}

func (g *GromLogger) LogMode(gormLogger.LogLevel) gormLogger.Interface {
	return g
}
func (g *GromLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	slog.InfoContext(ctx, msg)
}
func (g *GromLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	slog.WarnContext(ctx, msg)
}
func (g *GromLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	slog.ErrorContext(ctx, msg)
}
func (g *GromLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rowsAffected := fc()
	slog.InfoContext(ctx, fmt.Sprintf("%s sql: %s row:%d", begin, sql, rowsAffected))
}
