package log

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	gormLog "gorm.io/gorm/logger"
)

var (
	std = New()
)

const (
	TraceId string = "traceId"
	SpanId  string = "spanId"
	PspanId string = "pspanId"
)

type Level int

const (
	OffLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type H map[string]string

func (l Level) ToString() string {
	switch l {
	case 0:
		return "OFF"
	case 1:
		return "FATAL"
	case 2:
		return "ERROR"
	case 3:
		return "WARN"
	case 4:
		return "INFO"
	case 5:
		return "DEBUG"
	case 6:
		return "TRACE"
	default:
		return ""
	}
}

type Logger struct {
	Formatter Formatter
	Mongo     *mongo.Database
	Level     Level
	ch        chan *H
}

func New() *Logger {
	lo := &Logger{
		Level:     ErrorLevel,
		Formatter: &DefaultFormatter{},
	}
	lo.createLogRoutinue()
	return lo
}
func (lo *Logger) Fatal(ctx context.Context, format string, a ...any) {
	if lo.Level >= 1 {
		lo.write(ctx, format, a...)
	}
}
func (lo *Logger) Error(ctx context.Context, format string, a ...any) {
	if lo.Level >= 2 {
		lo.write(ctx, format, a...)
	}
}
func (lo *Logger) Warn(ctx context.Context, format string, a ...any) {
	if lo.Level >= 3 {
		lo.write(ctx, format, a...)
	}
}
func (lo *Logger) Info(ctx context.Context, format string, a ...any) {
	if lo.Level >= 4 {
		lo.write(ctx, format, a...)
	}
}
func (lo *Logger) Debug(ctx context.Context, format string, a ...any) {
	if lo.Level >= 5 {
		lo.write(ctx, format, a...)
	}
}
func (lo *Logger) Trace(ctx context.Context, format string, a ...any) {
	if lo.Level >= 6 {
		lo.write(ctx, format, a...)
	}
}
func (lo *Logger) write(ctx context.Context, format string, a ...any) {
	msg, err := lo.Formatter.Format(&Entry{
		Context: ctx,
		Message: fmt.Sprintf(format, a...),
	})
	if err != nil {
		panic("Format错误")
	}
	lo.ch <- msg
}

//创建日志协程，并消化日志
func (lo *Logger) createLogRoutinue() {
	go func() {
		for {
			h := <-lo.ch
			_, err := lo.Mongo.Collection("log").InsertOne(context.TODO(), h)
			if err != nil {
				_ = fmt.Errorf("%s", err.Error())
			}
		}
	}()
}

type Formatter interface {
	Format(*Entry) (*H, error)
}
type DefaultFormatter struct{}

func (df *DefaultFormatter) Format(entry *Entry) (*H, error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	//var file string
	//var len int
	if entry.Caller != nil {
		//file = filepath.Base(entry.Caller.File)
		//len = entry.Caller.Line
	}
	ctx := entry.Context
	//fmt.Println(entry.Data)
	//msg := fmt.Sprintf("%s [%s:%d][traceid:%s][%s] %s\n", timestamp, file, len, bbb, strings.ToUpper(entry.Level.String()), entry.Message)
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
	// msg := fmt.Sprintf("%s [%s][%s][%s][%s] %s\n",
	// 	timestamp,
	// 	traceId,
	// 	pspanId,
	// 	spanId,
	// 	strings.ToUpper(entry.Level.ToString()),
	// 	entry.Message)
	msg := &H{
		"timestamp": timestamp,
		"traceId":   traceId,
		"pspanId":   pspanId,
		"spanId":    spanId,
		"level":     entry.Level.ToString(),
		"message":   entry.Message,
	}
	return msg, nil
}

type gromLogger struct{}

func (gl *gromLogger) LogMode(logLevel gormLog.LogLevel) gormLog.Interface {
	return &gromLogger{}
}
func (gl *gromLogger) Info(ctx context.Context, format string, args ...interface{}) {
	std.Info(ctx, format, args...)
}
func (gl *gromLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	std.Warn(ctx, format, args...)
}
func (gl *gromLogger) Error(ctx context.Context, format string, args ...interface{}) {
	std.Error(ctx, format, args...)
}
func (gl *gromLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	nownow := time.Now()
	usedTime := nownow.Sub(begin).Milliseconds()
	sql, rowsAffected := fc()
	msg := fmt.Sprintf("sql:[%s] rows:[%d] %dms", sql, rowsAffected, usedTime)
	std.Trace(ctx, msg)
}
