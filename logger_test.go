package gowk

import (
	"testing"

	"log/slog"
)

func TestLogger(t *testing.T) {
	testing.Init()
	slog.SetDefault(Logger(slog.LevelInfo))
	slog.Info("22222")
}
