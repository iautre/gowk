package log

import (
	"os"

	"golang.org/x/exp/slog"
)

func Logger() *slog.Logger {
	options := slog.HandlerOptions{
		AddSource: true,
	}
	return slog.New(options.NewJSONHandler(os.Stderr))
}
