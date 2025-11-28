package gowk

import (
	"bytes"
	"context"
	"github.com/jackc/pgx/v5/tracelog"
	"os"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		slog.Any(PSPAN_ID, ctx.Value(SPAN_ID)),
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

const (
	START_TIME string = "startTime"

	TRACE_ID string = "trace_id"
	SPAN_ID  string = "span_id"
	PSPAN_ID string = "pspan_id"
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

	slog.InfoContext(ctx, "start", arrts...)
}

// 请求输出日志
func (r *requestLog) RequestOutLog(ctx *gin.Context, body *bytes.Buffer) {
	// after request
	endTime := time.Now()
	startTime, _ := ctx.Get(START_TIME)
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

type PostgresLogger struct {
}

func (p *PostgresLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	// 提取 SQL 和参数
	if sql, ok := data["sql"]; ok {
		args := data["args"]
		// 打印格式化日志
		slog.InfoContext(ctx, "[SQL]", sql)
		slog.InfoContext(ctx, "[SQL]", args)
	}

}
