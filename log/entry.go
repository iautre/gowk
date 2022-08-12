package log

import (
	"context"
	"runtime"
)

type Entry struct {
	Logger  *Logger
	Level   Level
	Message string
	Context context.Context
	Caller  *runtime.Frame
}
