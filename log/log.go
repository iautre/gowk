package log

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

func Init(m *mongo.Database) {
	SetMongo(m)
}

func SetLevel(level Level) {
	std.Level = level
}
func SetMongo(m *mongo.Database) {
	std.Mongo = m
}

//日志使用
func Fatalf(ctx context.Context, format string, a ...any) {
	std.Fatalf(ctx, format, a...)
}
func Errorf(ctx context.Context, format string, a ...any) {
	std.Errorf(ctx, format, a...)
}
func Warnf(ctx context.Context, format string, a ...any) {
	std.Warnf(ctx, format, a...)
}
func Infof(ctx context.Context, format string, a ...any) {
	std.Infof(ctx, format, a...)
}
func Debugf(ctx context.Context, format string, a ...any) {
	std.Debugf(ctx, format, a...)
}
func Tracef(ctx context.Context, format string, a ...any) {
	std.Tracef(ctx, format, a...)
}
func Fatal(ctx context.Context, a any) {
	std.Fatal(ctx, a)
}
func Error(ctx context.Context, a any) {
	std.Error(ctx, a)
}
func Warn(ctx context.Context, a any) {
	std.Warn(ctx, a)
}
func Info(ctx context.Context, a any) {
	std.Info(ctx, a)
}
func Debug(ctx context.Context, a any) {
	std.Debug(ctx, a)
}
func Trace(ctx context.Context, a any) {
	std.Trace(ctx, a)
}
