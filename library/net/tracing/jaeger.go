package tracing

import (
	"github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
	"io"
	"strconv"
	"time"
)

// 新建jaeger跟踪器
func NewJaegerTracer(c *Config, logger *Logger) (opentracing.Tracer, io.Closer) {
	samplerParam, err := strconv.ParseFloat(c.Sampler.Param, 64)
	if err != nil {
		panic(err)
	}

	cfg := jaegercfg.Configuration{
		ServiceName: c.AppName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:                    c.Sampler.Type,
			Param:                   samplerParam,
			SamplingServerURL:       c.Sampler.SamplingServerURL,
			MaxOperations:           c.Sampler.MaxOperations,
			SamplingRefreshInterval: time.Duration(c.Sampler.SamplingRefreshInterval),
		},
		Reporter: &jaegercfg.ReporterConfig{
			QueueSize:           c.Reporter.QueueSize,
			BufferFlushInterval: time.Duration(c.Reporter.BufferFlushInterval),
			LogSpans:            c.Reporter.LogSpans,
			LocalAgentHostPort:  c.Reporter.LocalAgentHostPort,
			CollectorEndpoint:   c.Reporter.CollectorEndpoint,
			User:                c.Reporter.User,
			Password:            c.Reporter.Password,
		},
	}
	jMetricsFactory := metrics.NullFactory

	trace, closer, err := cfg.NewTracer(
		jaegercfg.Logger(logger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		panic(err)
	}

	return trace, closer
}
