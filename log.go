package gowk

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	gormLog "gorm.io/gorm/logger"
)

type logger struct {
	gromLogger *gromLogger
}

var (
	logs    *logger
	logOnce sync.Once
)

func Log() *logger {
	if logs == nil {
		logOnce.Do(func() {
			logs = &logger{}
			logs.SetLevel("trace")
			logrus.SetFormatter(&logFormatter{})
			logs.gromLogger = &gromLogger{}
		})
	}
	return logs
}

type logFormatter struct{}

//格式详情
func (s *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := Now().Format("2006-01-02 15:04:05.000")
	//var file string
	//var len int
	if entry.HasCaller() {
		//file = filepath.Base(entry.Caller.File)
		//len = entry.Caller.Line
	}
	ctx := entry.Context
	//fmt.Println(entry.Data)
	//msg := fmt.Sprintf("%s [%s:%d][traceid:%s][%s] %s\n", timestamp, file, len, bbb, strings.ToUpper(entry.Level.String()), entry.Message)

	traceid := ctx.Value("traceid")
	pspanid := ctx.Value("pspanid")
	if pspanid == nil {
		pspanid = ""
	}
	spanId := ctx.Value("spanId")
	msg := fmt.Sprintf("%s [%s][%s][%s][%s] %s\n",
		timestamp,
		traceid,
		pspanid,
		spanId,
		strings.ToUpper(entry.Level.String()),
		entry.Message)

	return []byte(msg), nil
}

func (l *logger) SetLevel(level string) {
	f := func(level string) logrus.Level {
		switch level {
		case "trace":
			return logrus.TraceLevel
		case "debug":
			return logrus.DebugLevel
		case "info":
			return logrus.InfoLevel
		case "warning":
			return logrus.WarnLevel
		case "error":
			return logrus.ErrorLevel
		case "fatal":
			return logrus.FatalLevel
		case "panic":
			return logrus.PanicLevel
		}
		return logrus.InfoLevel
	}
	logrus.SetLevel(f(strings.ToLower(level)))
}
func (l *logger) Info(ctx context.Context, msg string, args ...interface{}) {
	logrus.WithContext(ctx).Info(msg)
}
func (l *logger) Warn(ctx context.Context, msg string, args ...interface{}) {
	logrus.WithContext(ctx).Warn(msg)
}
func (l *logger) Error(ctx context.Context, msg string, args ...interface{}) {
	logrus.WithContext(ctx).Error(msg)
}
func (l *logger) Trace(ctx context.Context, msg string, args ...interface{}) {
	logrus.WithContext(ctx).Trace(msg)
}

func (l *logger) GromLogger() *gromLogger {
	return l.gromLogger
}

type gromLogger struct {
}

func (gl *gromLogger) LogMode(logLevel gormLog.LogLevel) gormLog.Interface {

	return &gromLogger{}
}
func (gl *gromLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	Log().Info(ctx, msg)
}
func (gl *gromLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	Log().Warn(ctx, msg)
}
func (gl *gromLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	Log().Error(ctx, msg)
}
func (gl *gromLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	nownow := time.Now()
	usedTime := nownow.Sub(begin)
	sql, rowsAffected := fc()
	msg := fmt.Sprintf("sql:[%s] rows:[%d] %dms", sql, rowsAffected, usedTime)
	Log().Trace(ctx, msg)
}
