package log

import (
	"context"
	"runtime"
)

type Entry struct {
	Logger  *Logger
	Level   Level
	Message any
	Context context.Context
	Caller  *runtime.Frame
}
