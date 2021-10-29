package gowk

import (
	"context"
	"fmt"
	"path/filepath"
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
			logrus.SetFormatter(&logFormatter{})
			logs.gromLogger = &gromLogger{}
		})
	}
	return logs
}

type logFormatter struct{}

//格式详情
func (s *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Local().Format("0102-150405.000")
	var file string
	var len int
	if entry.HasCaller() {
		file = filepath.Base(entry.Caller.File)
		len = entry.Caller.Line
	}
	ctx := entry.Context
	bbb := ctx.Value("traceid")
	//fmt.Println(entry.Data)
	msg := fmt.Sprintf("%s [%s:%d][traceid:%s][%s] %s\n", timestamp, file, len, bbb, strings.ToUpper(entry.Level.String()), entry.Message)
	return []byte(msg), nil
}

func (l *logger) SetLevel(level uint32) {
	logrus.SetLevel((logrus.Level)(level))
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
func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	logrus.WithContext(ctx).Trace("trace msg")
}

func (l *logger) GromLogger() *gromLogger {
	return l.gromLogger
}

type gromLogger struct {
}

func (gl *gromLogger) LogMode(logLevel gormLog.LogLevel) gormLog.Interface {
	//alog.SetLevel(logrus.TraceLevel)
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
	Log().Trace(ctx, begin, fc, err)
}
