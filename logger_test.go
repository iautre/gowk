package gowk

import (
	"context"
	"testing"

	"log/slog"
)

func TestLogger(t *testing.T) {
	testing.Init()
	slog.SetDefault(Logger(slog.LevelInfo))
	slog.InfoContext(context.WithValue(context.TODO(), "traceId", "safsf"), "22222")
}
