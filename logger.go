package gowk

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	"log/slog"
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
	fmt.Printf(source.File, source.Function, source.Line)
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
