package httpclient

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/metric"
	"gitlab.shanhai.int/sre/library/net/sentry"
	"gitlab.shanhai.int/sre/library/net/tracing"
)

// 处理方法
type HandlerFunc func(*Span)

// 原始Http处理方法
func GetOriginHttpHandler() HandlerFunc {
	return func(b *Span) {
		b.startTime = time.Now()
		resp, err := b.originClient.Do(b.Request)
		b.endTime = time.Now()
		b.duration = b.endTime.Sub(b.startTime)
		b.Response = resp
		b.SetError(err)
	}
}

// k8s客户端负载均衡处理方法
func GetK8sLoadBalancerHandler() HandlerFunc {
	return func(b *Span) {
		if b.GetConfig().EnableLoadBalancer {
			host := b.Request.URL.Host
			ctx := b.GetContext()
			lb := b.client.loadBalancer

			if lb.IsAdded(ctx, host) {
				_ = lb.Watch(ctx, host)
			} else {
				_ = lb.Add(ctx, host)
			}

			b.Request.URL.Host = lb.MustGetEndpoint(ctx, host)
		}

		b.Next()
	}
}

// 过滤处理方法
func GetFilterHandler() HandlerFunc {
	return func(b *Span) {
		b.Next()

		filterFunc := b.GetFilterFunc()
		if filterFunc != nil {
			err := filterFunc(b.Request, b.Response)
			b.SetError(err)
		} else {
			GetAccessStatusCodeHandler()(b)
		}
	}
}

// 状态码过滤处理方法
func GetAccessStatusCodeHandler() HandlerFunc {
	return func(b *Span) {
		b.Next()

		resp := b.Response
		if resp == nil {
			return
		}

		isAccess := len(b.GetAccessStatusCode()) == 0
		for _, code := range b.GetAccessStatusCode() {
			if resp.StatusCode == code {
				isAccess = true
				break
			}
		}
		if !isAccess {
			b.SetError(errors.New(fmt.Sprintf("status code is %d , can't access", resp.StatusCode)))
		}
	}
}

// 断路器处理方法
func GetBreakerHandler() HandlerFunc {
	return func(b *Span) {
		if b.GetConfig().DisableBreaker {
			b.Next()
			return
		}

		breaker, err := b.GetUrlBreaker(b.url)
		if err != nil {
			b.SetError(err)
			return
		}

		err = breaker.Call(b.GetContext(), func() error {
			b.Next()
			return b.GetError()
		}, 0)
		if errcode.EqualError(errcode.BreakerOpenError, err) && b.GetDegradedResponse() != nil {
			b.Response = b.GetDegradedResponse()
			b.SetError(errors.Wrap(errcode.BreakerDegradedError, err.Error()))
		} else if err != nil {
			b.SetError(err)
		}
	}
}

// 钩子处理方法
func GetHookHandler() HandlerFunc {
	return func(b *Span) {
		hk := b.manager.CreateHook(b.GetContext()).
			AddArg(render.SourceArgKey, b.source).
			AddArg("method_name", b.method).
			AddArg("url", b.url).
			AddArg("headers", b.headers.Header).
			AddArg("host", b.host)
		if b.body != nil && b.conf.RequestBodyOut {
			hk.AddArg("request_body", string(b.body))
		}
		hk.ProcessPreHook()
		b.ctx = hk.Context()

		b.Next()

		hk.AddArg(render.StartTimeArgKey, b.startTime).
			AddArg(render.EndTimeArgKey, b.endTime).
			AddArg(render.DurationArgKey, b.duration).
			AddArg("endpoint", b.endpoint).
			AddArg(render.ErrorArgKey, b.GetError())
		if originResponse := b.Response; originResponse != nil {
			hk.AddArg("status_code", originResponse.StatusCode)
			if !b.conf.ResponseBodyOut {
				goto SkipResponse
			}
			bodyBytes, err := ioutil.ReadAll(originResponse.Body)
			if err != nil {
				goto SkipResponse
			}
			err = originResponse.Body.Close()
			if err != nil {
				goto SkipResponse
			}
			originResponse.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			hk.AddArg("response_body", string(bodyBytes))
		}
	SkipResponse:
		hk.ProcessAfterHook()
	}
}

// 统计处理方法
func GetMetricsHandler() HandlerFunc {
	return func(b *Span) {
		b.manager.RegisterHook(func(hk *hook.Hook) {
			args := hk.Args()

			metric.HttpRequestTotal.With(
				prometheus.Labels{
					"web_url":     render.PatternWebUrl(args).StringValue(),
					"web_method":  render.PatternWebMethod(args).StringValue(),
					"host":        host(args).StringValue(),
					"method_name": methodName(args).StringValue(),
				},
			).Inc()
		}, func(hk *hook.Hook) {
			args := hk.Args()

			metric.HttpRequestDurationSummary.With(
				prometheus.Labels{
					"host":        host(args).StringValue(),
					"method_name": methodName(args).StringValue(),
				},
			).Observe(render.PatternDuration(args).Float64Value())

			metric.HttpResponseTotal.With(
				prometheus.Labels{
					"web_url":     render.PatternWebUrl(args).StringValue(),
					"web_method":  render.PatternWebMethod(args).StringValue(),
					"host":        host(args).StringValue(),
					"method_name": methodName(args).StringValue(),
					"status_code": strconv.Itoa(statusCode(args).IntValue()),
				},
			).Inc()
		})

		b.Next()
	}
}

// 链路跟踪处理方法
func GetTracingHandler() HandlerFunc {
	return func(b *Span) {
		if b.GetConfig().DisableTracing {
			b.Next()
			return
		}

		b.manager.RegisterTracingHook(func(hk *hook.Hook) string {
			return fmt.Sprintf("%s%s", tracing.SpanPrefixHttpClient, host(hk.Args()).StringValue())
		}, func(hk *hook.Hook, span opentracing.Span) {
			args := hk.Args()

			header, ok := headersFuc(args).Value.(http.Header)
			if !ok {
				return
			}

			err := span.Tracer().Inject(
				span.Context(), opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(header),
			)
			if err != nil {
				return
			}

			ext.HTTPMethod.Set(span, methodName(args).StringValue())
			ext.HTTPUrl.Set(span, urlFuc(args).StringValue())
		}, func(hk *hook.Hook, span opentracing.Span) {
			args := hk.Args()

			if err := render.PatternError(args).StringValue(); err != "" {
				span.SetTag("http.error", err)
				ext.Error.Set(span, true)
			}

			statusCode := statusCode(args).IntValue()
			ext.HTTPStatusCode.Set(span, uint16(statusCode))
			if statusCode != http.StatusOK {
				ext.Error.Set(span, true)
			}
		})

		b.Next()
	}
}

// sentry处理方法
func GetSentryHandler() HandlerFunc {
	return func(b *Span) {
		if b.GetConfig().DisableSentry {
			b.Next()
			return
		}

		b.manager.RegisterSentryBreadCrumbHook(func(hk *hook.Hook) *sentry.Breadcrumb {
			args := hk.Args()
			return &sentry.Breadcrumb{
				Category: title(args).StringValue(),
				Data: render.NewPatternResultMap().
					Add(httpClientExtra(args)).
					Add(render.PatternSource(args)).
					Add(render.PatternStartTime(args)).
					Add(render.PatternEndTime(args)),
			}
		})

		b.Next()
	}
}
