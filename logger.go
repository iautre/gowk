package gowk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gormLogger "gorm.io/gorm/logger"
)

func Logger(l slog.Level) *slog.Logger {
	options := &slog.HandlerOptions{
		AddSource: true,
		Level:     l,
	}
	return slog.New(NewTestHandler(os.Stderr, options))
}

type TextHandler struct {
	H *slog.TextHandler
}

func NewTestHandler(w io.Writer, opts *slog.HandlerOptions) *TextHandler {
	return &TextHandler{
		H: slog.NewTextHandler(w, opts),
	}
}

func (h *TextHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.H.Enabled(ctx, l)
}

func (h *TextHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(getTraceId(ctx)...)
	return h.H.Handle(ctx, r)
}

func (h *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.H.WithAttrs(attrs)
}

func (h *TextHandler) WithGroup(name string) slog.Handler {
	return h.H.WithGroup(name)
}
func getTraceId(ctx context.Context) []slog.Attr {
	return []slog.Attr{
		slog.Any(TraceId, ctx.Value(TraceId)),
		slog.Any(SpanId, ctx.Value(SpanId)),
		slog.Any(PspanId, ctx.Value(SpanId)),
	}
}

// func source(pc uintptr) *slog.Source {
// 	fs := runtime.CallersFrames([]uintptr{pc})
// 	f, _ := fs.Next()
// 	return &slog.Source{
// 		Function: f.Function,
// 		File:     f.File,
// 		Line:     f.Line,
// 	}
// }

type GromLogger struct {
}

func (g *GromLogger) LogMode(gormLogger.LogLevel) gormLogger.Interface {
	return g
}
func (g *GromLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	slog.InfoContext(ctx, msg)
}
func (g *GromLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	slog.WarnContext(ctx, msg)
}
func (g *GromLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	slog.ErrorContext(ctx, msg)
}
func (g *GromLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rowsAffected := fc()
	slog.InfoContext(ctx, fmt.Sprintf("%s sql: %s row:%d", begin, sql, rowsAffected))
}

const (
	StartTime string = "startTime"

	TraceId string = "traceId"
	SpanId  string = "spanId"
	PspanId string = "pspanId"
)

func RequestMiddleware() gin.HandlerFunc {
	r := &requestLog{}
	return func(c *gin.Context) {
		bw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		// c.Writer = bw
		r.RequestInLog(c)
		defer r.RequestOutLog(c, bw.body)
		c.Next()
	}
}

type requestLog struct{}

// 请求进入日志
func (r *requestLog) RequestInLog(ctx *gin.Context) {
	arrts := []any{
		// {Key: "type", Value: slog.StringValue("start")},
		"ip", slog.StringValue(ctx.ClientIP()),
		"method", slog.StringValue(ctx.Request.Method),
		"uri", slog.StringValue(ctx.Request.RequestURI),
		"header", slog.AnyValue(ctx.Request.Header),
		"body", slog.AnyValue(ctx.Request.Body),
	}
	startTime := time.Now()
	ctx.Set(StartTime, startTime)
	traceId := ctx.Request.Header.Get(TraceId)
	if traceId == "" {
		traceId = uuid.NewString()
	}
	ctx.Set(TraceId, traceId)
	ctx.Request.Header.Set(TraceId, traceId)

	pspanId := ctx.Request.Header.Get(SpanId)
	if pspanId != "" {
		ctx.Set(PspanId, pspanId)
		ctx.Request.Header.Set(PspanId, pspanId)
	}
	spanId := uuid.NewString()
	ctx.Set(SpanId, spanId)
	ctx.Request.Header.Set(SpanId, spanId)

	slog.InfoContext(ctx, "start", arrts...)
}

// 请求输出日志
func (r *requestLog) RequestOutLog(ctx *gin.Context, body *bytes.Buffer) {
	// after request
	endTime := time.Now()
	startTime, _ := ctx.Get(StartTime)
	usedTime := endTime.Sub(startTime.(time.Time)).Milliseconds()
	arrts := []any{
		// {Key: "type", Value: slog.StringValue("start")},
		"status", slog.IntValue(ctx.Writer.Status()),
		"usedTime", slog.Int64Value(usedTime),
		// "responeBody", slog.StringValue(body.String()),
	}
	slog.InfoContext(ctx, "end", arrts...)
}

type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w CustomResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w CustomResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
