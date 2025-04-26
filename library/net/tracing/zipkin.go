package tracing

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	zipkintracer "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter/http"
	"gitlab.shanhai.int/sre/library/base/net"
)

const (
	SamplerTypeModulo   = "modulo"
	SamplerTypeBoundary = "boundary"
	SamplerTypeCounting = "counting"
)

// 新建zipkin跟踪器
func NewZipkinTracer(c *Config, logger *Logger) (opentracing.Tracer, io.Closer) {
	// todo:openzipkin内部logger居然使用结构体，无法自定义，好蠢，可以考虑自定义

	if c.Reporter.LocalEndpoint == "" {
		ip, err := net.GetInternalIP()
		if err == nil {
			c.Reporter.LocalEndpoint = ip
		} else {
			c.Reporter.LocalEndpoint = "0.0.0.0"
		}
	}

	reporter := http.NewReporter(c.Reporter.CollectorEndpoint)

	endpoint, err := zipkin.NewEndpoint(c.AppName, c.Reporter.LocalEndpoint)
	if err != nil {
		panic(err)
	}

	nativeTracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(endpoint),
		zipkin.WithSampler(getSampler(c)),
	)
	if err != nil {
		panic(err)
	}

	tracer := zipkintracer.Wrap(nativeTracer)
	return tracer, reporter
}

// 根据配置获取采样器
func getSampler(c *Config) zipkin.Sampler {
	var sampler zipkin.Sampler

	if c.Sampler == nil {
		c.Sampler = &SamplerConfig{}
	}
	if c.Sampler.Type == "" {
		c.Sampler.Type = SamplerTypeBoundary
	}
	if c.Sampler.Param == "" {
		c.Sampler.Param = "0.01"
	}

	switch c.Sampler.Type {
	case SamplerTypeBoundary:
		params := strings.Split(c.Sampler.Param, ",")
		if len(params) == 0 {
			panic(errors.New("boundary sampler param is invalid"))
		}

		rate, err := strconv.ParseFloat(params[0], 64)
		if err != nil {
			panic(err)
		}

		var salt int64
		if len(params) == 2 {
			customSalt, err := strconv.ParseInt(params[1], 10, 64)
			if err != nil {
				panic(err)
			}
			salt = customSalt
		} else {
			salt = time.Now().UnixNano()
		}

		sampler, err = zipkin.NewBoundarySampler(rate, salt)
		if err != nil {
			panic(err)
		}
	case SamplerTypeCounting:
		rate, err := strconv.ParseFloat(c.Sampler.Param, 64)
		if err != nil {
			panic(err)
		}
		sampler, err = zipkin.NewCountingSampler(rate)
		if err != nil {
			panic(err)
		}
	case SamplerTypeModulo:
		fallthrough
	default:
		modValue, err := strconv.Atoi(c.Sampler.Param)
		if err != nil {
			panic(err)
		}
		sampler = zipkin.NewModuloSampler(uint64(modValue))
	}

	return sampler
}
