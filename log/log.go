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
func Info(ctx context.Context, format string, a ...any) {
	std.Info(ctx, format, a...)
}
func Warn(ctx context.Context, format string, a ...any) {
	std.Warn(ctx, format, a...)
}
func Error(ctx context.Context, format string, a ...any) {
	std.Error(ctx, format, a...)
}
func Trace(ctx context.Context, format string, a ...any) {
	std.Trace(ctx, format, a...)
}

//日志通道
func init() {

}
