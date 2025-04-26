package tracing

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	zipkintracer "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/uber/jaeger-client-go"
	_context "gitlab.shanhai.int/sre/library/base/context"
	_gin "gitlab.shanhai.int/sre/library/net/gin"
)

// 从上游解析中间件
func ExtractFromUpstream(advancedOpts ...opentracing.StartSpanOption) gin.HandlerFunc {
	if _tracer.Tracer == nil {
		panic("Tracer is nil")
	}

	return func(ctx *gin.Context) {
		spanContext, err := _tracer.Extract(opentracing.TextMap, opentracing.HTTPHeadersCarrier(ctx.Request.Header))
		if err != nil {
			_logger.Error(fmt.Sprintf("%s", err))
		}

		opts := append([]opentracing.StartSpanOption{opentracing.ChildOf(spanContext)}, advancedOpts...)
		relativePath := _gin.GetGinRelativePath(ctx)
		span := _tracer.StartSpan(fmt.Sprintf("%s%s", SpanPrefixWeb, relativePath), opts...)
		defer span.Finish()
		ctx.Set(CurrentSpanContextKey, span)

		var traceID string
		switch sc := span.Context().(type) {
		case zipkintracer.SpanContext:
			traceID = sc.TraceID.String()
		case jaeger.SpanContext:
			traceID = sc.TraceID().String()
		}
		ctx.Set(TraceIDContextKey, traceID)

		ctx.Next()

		span.SetTag("uuid", _context.GetStringOrDefault(ctx, _context.ContextUUIDKey, "unknown"))
		statusCode := ctx.Writer.Status()
		ext.HTTPStatusCode.Set(span, uint16(statusCode))
		ext.HTTPMethod.Set(span, ctx.Request.Method)
		ext.HTTPUrl.Set(span, ctx.Request.URL.String())
		if statusCode != http.StatusOK {
			ext.Error.Set(span, true)
		}
	}
}

// 注入到下游中间件
func InjectToDownstream() gin.HandlerFunc {
	if _tracer.Tracer == nil {
		panic("Tracer is nil")
	}

	return func(ctx *gin.Context) {
		{
			var spanContext opentracing.SpanContext
			upstreamSpan, exist := ctx.Get(CurrentSpanContextKey)
			if !exist {
				_logger.Error("context haven't span key")
				goto NEXT
			}

			span, ok := upstreamSpan.(opentracing.Span)
			if !ok || span == nil {
				_logger.Error("upstream span is not opentracing span")
				goto NEXT
			}

			spanContext = span.Context()
			err := _tracer.Inject(spanContext, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(ctx.Request.Header))
			if err != nil {
				_logger.Error(fmt.Sprintf("%s", err))
				goto NEXT
			}
		}

	NEXT:
		ctx.Next()
	}
}
