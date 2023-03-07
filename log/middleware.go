package log

import (
	"bytes"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

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

	Trace(ctx, "start", arrts...)
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
	Trace(ctx, "end", arrts...)
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
