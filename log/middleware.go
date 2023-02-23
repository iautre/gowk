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
)

func RequestMiddleware() gin.HandlerFunc {
	r := &requestLog{}
	return func(c *gin.Context) {
		r.RequestInLog(c)
		bw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		defer r.RequestOutLog(c, bw.body)
		c.Writer = bw
		c.Next()
	}
}

type requestLog struct{}

// 请求进入日志
func (r *requestLog) RequestInLog(c *gin.Context) {
	startTime := time.Now()
	c.Set(StartTime, startTime)
	if traceId := c.Request.Header.Get(TraceId); traceId == "" {
		traceId = uuid.NewString()
		c.Set(TraceId, traceId)
		c.Request.Header.Set(TraceId, traceId)
	}
	if spanId := c.Request.Header.Get(SpanId); spanId != "" {
		c.Set(PspanId, spanId)
		c.Request.Header.Set(PspanId, spanId)
	}
	spanId := uuid.NewString()
	c.Set(SpanId, spanId)
	c.Request.Header.Set(SpanId, spanId)
	// msg := fmt.Sprintf("%s %s %s Header: %v Body: %v start", c.ClientIP(), c.Request.Method, c.Request.RequestURI, c.Request.Header, c.Request.Body)
	arrts := []slog.Attr{
		// {Key: "type", Value: slog.StringValue("start")},
		{Key: "ip", Value: slog.StringValue(c.ClientIP())},
		{Key: "method", Value: slog.StringValue(c.Request.Method)},
		{Key: "uri", Value: slog.StringValue(c.Request.RequestURI)},
		{Key: "header", Value: slog.AnyValue(c.Request.Header)},
		{Key: "body", Value: slog.AnyValue(c.Request.Body)},
	}
	Trace(c, "start", arrts...)
}

// 请求输出日志
func (r *requestLog) RequestOutLog(c *gin.Context, body *bytes.Buffer) {
	// after request
	endTime := time.Now()
	startTime, _ := c.Get(StartTime)
	usedTime := endTime.Sub(startTime.(time.Time)).Milliseconds()
	// msg := fmt.Sprintf("%d end %dms", c.Writer.Status(), usedTime)
	arrts := []slog.Attr{
		// {Key: "type", Value: slog.StringValue("start")},
		{Key: "status", Value: slog.IntValue(c.Writer.Status())},
		{Key: "usedTime", Value: slog.Int64Value(usedTime)},
		{Key: "responeBody", Value: slog.StringValue(body.String())},
	}

	Trace(c, "end", arrts...)
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
