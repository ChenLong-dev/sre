package goroutine

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
	_infoFile      = "goroutineInfo.log"
	defaultPattern = "%J{tTUSG}"
)

// 新建钩子管理器
func NewHookManager(renderConfig *render.Config) *hook.Manager {
	// 创建日志记录器
	logger := hook.GetDefaultLogger(renderConfig, patternMap)
	return hook.NewManager().
		SetLogger(logger).
		// goroutine包需要注入前后都打印日志
		RegisterHook(func(hk *hook.Hook) {
			hk.GetLogger().Print(hk.Args())
		}, func(hk *hook.Hook) {
			hk.GetLogger().Print(hk.Args())
		}).
		// goroutine包需要注入前后都增加面包屑
		RegisterHook(func(hk *hook.Hook) {
			args := hk.Args()
			hk.SetContext(
				sentry.AddBreadcrumb(
					hk.Context(),
					&sentry.Breadcrumb{
						Category: title(args).StringValue(),
						Data: render.NewPatternResultMap().
							Add(goroutine(args)).
							Add(render.PatternSource(args)).
							Add(render.PatternTime(args)),
					},
				),
			)
		}, func(hk *hook.Hook) {
			args := hk.Args()
			hk.SetContext(
				sentry.AddBreadcrumb(
					hk.Context(),
					&sentry.Breadcrumb{
						Category: title(args).StringValue(),
						Data: render.NewPatternResultMap().
							Add(goroutine(args)).
							Add(render.PatternSource(args)).
							Add(render.PatternTime(args)),
					},
				),
			)
		}).
		RegisterHook(func(hk *hook.Hook) {
			args := hk.Args()

			metric.GoroutineRequestTotal.With(
				prometheus.Labels{
					"web_url":    render.PatternWebUrl(args).StringValue(),
					"web_method": render.PatternWebMethod(args).StringValue(),
					"group_name": groupName(args).StringValue(),
				},
			).Inc()
		}, func(hk *hook.Hook) {
			args := hk.Args()

			metric.GoroutineRequestDurationSummary.With(
				prometheus.Labels{
					"group_name": groupName(args).StringValue(),
					"state":      state(args).StringValue(),
				},
			).Observe(render.PatternDuration(args).Float64Value())

			metric.GoroutineResponseTotal.With(
				prometheus.Labels{
					"web_url":    render.PatternWebUrl(args).StringValue(),
					"web_method": render.PatternWebMethod(args).StringValue(),
					"group_name": groupName(args).StringValue(),
					"state":      state(args).StringValue(),
				},
			).Inc()
		}).
		RegisterTracingHook(func(hk *hook.Hook) string {
			return fmt.Sprintf("%s%s", tracing.SpanPrefixGoroutine, goroutineName(hk.Args()).StringValue())
		}, func(hk *hook.Hook, span opentracing.Span) {
			args := hk.Args()

			span.SetTag("uuid", render.PatternUUID(args).StringValue())
			span.SetTag("go.group", groupName(args).StringValue())
			span.SetTag("go.goroutine", goroutineName(args).StringValue())
			span.SetTag("go.mode", mode(args).StringValue())
		}, func(hk *hook.Hook, span opentracing.Span) {
			args := hk.Args()

			if err := render.PatternError(args).StringValue(); err != "" {
				ext.Error.Set(span, true)
				span.SetTag("go.error", err)
			}
		})
}

// 渲染模版
var patternMap = map[string]render.PatternFunc{
	"T": render.PatternTime,
	"S": render.PatternSource,
	"U": render.PatternUUID,
	"t": title,
	"G": goroutine,
	"s": state,
	"N": groupName,
	"n": goroutineName,
	"I": groupID,
	"i": goroutineID,
	"E": extra,
	"m": mode,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "GOROUTINE")
}

// 汇总的goroutine参数
func goroutine(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("goroutine", []render.PatternFunc{
		state, groupName, groupID, goroutineID, goroutineName, extra, mode,
	})(args)
}

// 协程状态
func state(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("state", args["state"])
}

// 协程组名
func groupName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("group_name", args["group_name"])
}

// 协程组id
func groupID(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("group_id", args["group_id"])
}

// 协程名
func goroutineName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("goroutine_name", args["goroutine_name"])
}

// 协程id
func goroutineID(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("goroutine_id", args["goroutine_id"])
}

// 额外参数，如报错信息
func extra(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("extra", args.GetOrDefault("extra", ""))
}

// 协程组模式
func mode(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("mode", args["mode"])
}
