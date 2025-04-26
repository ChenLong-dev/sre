package tracing

import (
	"context"
	"errors"
	"io"

	"github.com/opentracing/opentracing-go"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	// context中当前span存放的key
	CurrentSpanContextKey = "context_span_current"
	// context中父span存放的key
	ParentSpanContextKey = "context_span_parent"

	// context中存放trace id的key
	TraceIDContextKey = "context_trace_id"
)

const (
	SpanPrefixWeb        = "[WEB] "
	SpanPrefixHttpClient = "[HTTPCLIENT] "
	SpanPrefixGorm       = "[GORM] "
	SpanPrefixRedis      = "[REDIS] "
	SpanPrefixMongo      = "[MONGO] "
	SpanPrefixRedlock    = "[REDLOCK] "
	SpanPrefixGoroutine  = "[GOROUTINE] "
)

var (
	// 全局跟踪器
	_tracer Tracer
	// 全局logger
	_logger = getDefaultWriter(&Config{
		Config: &render.Config{
			Stdout: false,
		},
	})
)

type Tracer struct {
	// 跟踪器
	opentracing.Tracer
	// 报告器的关闭接口
	closer io.Closer
	// 日志logger
	logger *Logger
	// 配置文件
	conf *Config
}

// 新建跟踪器
func New(c *Config) {
	if c == nil {
		panic("tracing config is nil")
	}
	if c.Config == nil {
		c.Config = &render.Config{}
	}

	_logger = getDefaultWriter(c)

	trace, closer := NewZipkinTracer(c, _logger)

	opentracing.SetGlobalTracer(trace)

	_tracer = Tracer{
		Tracer: trace,
		closer: closer,
		logger: _logger,
		conf:   c,
	}
}

// 关闭跟踪器
func Close() {
	if _tracer.closer == nil {
		panic("closer is nil")
	}

	err := _tracer.closer.Close()
	if err != nil {
		panic(err)
	}
}

// 获取context中的当前Span
func GetCurrentSpanFromContext(ctx context.Context) (opentracing.Span, error) {
	span, ok := (ctx).Value(CurrentSpanContextKey).(opentracing.Span)
	if !ok || span == nil {
		return nil, errors.New("context value is null or not span")
	}

	return span, nil
}

// 设置context中的当前Span
func SetCurrentSpanToContext(ctx context.Context, span opentracing.Span) context.Context {
	ctx = context.WithValue(ctx, CurrentSpanContextKey, span)
	return ctx
}

// 获取context中的父级Span
func GetParentSpanFromContext(ctx context.Context) (opentracing.Span, error) {
	span, ok := (ctx).Value(ParentSpanContextKey).(opentracing.Span)
	if !ok || span == nil {
		return nil, errors.New("context value is null or not span")
	}

	return span, nil
}

// 设置context中的父级Span
func SetParentSpanToContext(ctx context.Context, span opentracing.Span) context.Context {
	ctx = context.WithValue(ctx, ParentSpanContextKey, span)
	return ctx
}
