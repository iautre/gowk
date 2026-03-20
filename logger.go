package gowk

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/tracelog"
)

func Logger(l slog.Level) *slog.Logger {
	options := &slog.HandlerOptions{
		AddSource: true,
		Level:     l,
	}
	return slog.New(&TextHandler{slog.NewTextHandler(os.Stderr, options)})
}

type TextHandler struct {
	slog.Handler
}

func (h *TextHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(getTraceId(ctx)...)
	return h.Handler.Handle(ctx, r)
}

func getTraceId(ctx context.Context) []slog.Attr {
	return []slog.Attr{
		slog.Any(TRACE_ID, ctx.Value(TRACE_ID)),
		slog.Any(SPAN_ID, ctx.Value(SPAN_ID)),
		slog.Any(PSPAN_ID, ctx.Value(PSPAN_ID)),
	}
}

const (
	START_TIME string = "startTime"
	TRACE_ID   string = "trace_id"
	SPAN_ID    string = "span_id"
	PSPAN_ID   string = "pspan_id"
)

func RequestMiddleware() gin.HandlerFunc {
	r := &requestLog{}
	return func(c *gin.Context) {
		bw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bw
		r.RequestInLog(c)
		defer r.RequestOutLog(c, bw.body)
		c.Next()
	}
}

type requestLog struct{}

func (r *requestLog) RequestInLog(ctx *gin.Context) {
	startTime := time.Now()
	ctx.Set(START_TIME, startTime)

	traceId := ctx.Request.Header.Get(TRACE_ID)
	if traceId == "" {
		traceId = uuid.NewString()
	}
	ctx.Set(TRACE_ID, traceId)
	ctx.Request.Header.Set(TRACE_ID, traceId)

	pspanId := ctx.Request.Header.Get(SPAN_ID)
	if pspanId != "" {
		ctx.Set(PSPAN_ID, pspanId)
		ctx.Request.Header.Set(PSPAN_ID, pspanId)
	}
	spanId := uuid.NewString()
	ctx.Set(SPAN_ID, spanId)
	ctx.Request.Header.Set(SPAN_ID, spanId)

	// 只记录非敏感信息，不记录 Header（含 Authorization）和 Body
	slog.InfoContext(ctx, "start",
		"ip", ctx.ClientIP(),
		"method", ctx.Request.Method,
		"uri", ctx.Request.RequestURI,
	)
}

func (r *requestLog) RequestOutLog(ctx *gin.Context, body *bytes.Buffer) {
	endTime := time.Now()
	startTime, _ := ctx.Get(START_TIME)
	usedTime := endTime.Sub(startTime.(time.Time)).Milliseconds()
	slog.InfoContext(ctx, "end",
		"status", ctx.Writer.Status(),
		"usedTime", usedTime,
	)
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

// PostgresLogger 实现 pgx tracelog.Logger 接口，按实际日志级别输出 SQL。
type PostgresLogger struct{}

func (p *PostgresLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	sql, _ := data["sql"]
	args := data["args"]

	logFn := slog.InfoContext
	switch level {
	case tracelog.LogLevelDebug, tracelog.LogLevelTrace:
		logFn = slog.DebugContext
	case tracelog.LogLevelWarn:
		logFn = slog.WarnContext
	case tracelog.LogLevelError:
		logFn = slog.ErrorContext
	}
	logFn(ctx, "[SQL] "+msg, "sql", sql, "args", args)
}
