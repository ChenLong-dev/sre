package redis

import (
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/net/metric"
	"gitlab.shanhai.int/sre/library/net/sentry"
	"gitlab.shanhai.int/sre/library/net/tracing"
)

const (
	_infoFile = "redisInfo.log"

	defaultPattern = "%J{tsTUSR}"
)

// 新建钩子管理器
func NewHookManager(renderConfig *render.Config) *hook.Manager {
	return hook.NewManager().
		RegisterLogHook(renderConfig, patternMap).
		RegisterHook(func(hk *hook.Hook) {
			args := hk.Args()
			if commandName(args).StringValue() == "" {
				return
			}

			metric.RedisRequestTotal.With(
				prometheus.Labels{
					"web_url":    render.PatternWebUrl(args).StringValue(),
					"web_method": render.PatternWebMethod(args).StringValue(),
					"endpoint":   endpoint(args).StringValue(),
				},
			).Inc()
		}, func(hk *hook.Hook) {
			args := hk.Args()
			if commandName(args).StringValue() == "" {
				return
			}

			metric.RedisRequestDurationSummary.With(
				prometheus.Labels{
					"endpoint": endpoint(args).StringValue(),
				},
			).Observe(render.PatternDuration(args).Float64Value())
		}).
		RegisterTracingHook(func(hk *hook.Hook) string {
			args := hk.Args()
			if commandName(args).StringValue() == "" {
				return ""
			}

			return fmt.Sprintf("%s%s", tracing.SpanPrefixRedis, commandName(args).StringValue())
		}, func(hk *hook.Hook, span opentracing.Span) {
			args := hk.Args()

			span.SetTag("uuid", render.PatternUUID(args).StringValue())
			ext.DBType.Set(span, "redis")
			ext.DBInstance.Set(span, endpoint(args).StringValue())
			ext.DBStatement.Set(span, commandName(args).StringValue())
		}, func(hk *hook.Hook, span opentracing.Span) {
			args := hk.Args()

			if err := render.PatternError(args).StringValue(); err != "" {
				ext.Error.Set(span, true)
				span.SetTag("db.error", err)
			}
		}).
		RegisterSentryBreadCrumbHook(func(hk *hook.Hook) *sentry.Breadcrumb {
			args := hk.Args()
			return &sentry.Breadcrumb{
				Category: title(args).StringValue(),
				Data: render.NewPatternResultMap().
					Add(redisExtra(args)).
					Add(render.PatternSource(args)).
					Add(render.PatternStartTime(args)).
					Add(render.PatternEndTime(args)),
			}
		})
}

// 渲染模版
var patternMap = map[string]render.PatternFunc{
	"E": endpoint,
	"T": render.PatternEndTime,
	"S": render.PatternSource,
	"s": render.PatternStartTime,
	"U": render.PatternUUID,
	"t": title,
	"R": redisExtra,
	"D": render.PatternDuration,
	"N": funcName,
	"n": commandName,
	"a": commandArgs,
	"r": reply,
}

// Redis 连接地址
func endpoint(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("endpoint", args["endpoint"])
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "REDIS")
}

// 汇总的redis参数
func redisExtra(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("redis", []render.PatternFunc{
		endpoint, render.PatternDuration,
		funcName, commandName, commandArgs, reply,
	})(args)
}

// 调用函数名
func funcName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("func_name", args["func_name"])
}

// 操作指令名
func commandName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("command_name", args["command_name"])
}

// 操作指令参数
func commandArgs(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("command_args", args["command_args"])
}

// 响应
func reply(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("reply", args["reply"])
}
