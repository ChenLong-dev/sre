package redlock

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
	_infoFile = "redlockInfo.log"

	defaultPattern = "%J{tbTUSR}"
)

// 新建钩子管理器
func NewHookManager(renderConfig *render.Config) *hook.Manager {
	return hook.NewManager().
		RegisterLogHook(renderConfig, patternMap).
		RegisterPreHook(func(hk *hook.Hook) {
			args := hk.Args()

			metric.RedlockRequestTotal.With(
				prometheus.Labels{
					"web_url":    render.PatternWebUrl(args).StringValue(),
					"web_method": render.PatternWebMethod(args).StringValue(),
				},
			).Inc()
		}).
		RegisterTracingHook(func(hk *hook.Hook) string {
			return fmt.Sprintf("%s%s", tracing.SpanPrefixRedlock, mutexName(hk.Args()).StringValue())
		}, func(hk *hook.Hook, span opentracing.Span) {
			args := hk.Args()

			span.SetTag("uuid", render.PatternUUID(args).StringValue())
			ext.DBType.Set(span, "redlock")
			ext.DBStatement.Set(span, commandName(args).StringValue())
			ext.DBInstance.Set(span, mutexName(args).StringValue())
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
					Add(redlockExtra(args)).
					Add(render.PatternSource(args)).
					Add(render.PatternStartTime(args)).
					Add(render.PatternEndTime(args)),
			}
		})
}

// 渲染模版
var patternMap = map[string]render.PatternFunc{
	"T": render.PatternEndTime,
	"S": render.PatternSource,
	"b": render.PatternStartTime,
	"U": render.PatternUUID,
	"t": title,
	"R": redlockExtra,
	"s": state,
	"D": render.PatternDuration,
	"N": mutexName,
	"n": commandName,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "REDLOCK")
}

// 汇总的redlock参数
func redlockExtra(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("redlock", []render.PatternFunc{
		mutexName, commandName, state, render.PatternDuration,
	})(args)
}

// 锁名称
func mutexName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("mutex_name", args["mutex_name"])
}

// 操作名称
func commandName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("command_name", args["command_name"])
}

// 状态
func state(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("state", args["state"])
}
