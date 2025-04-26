package sql

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
	_InfoFile = "gormInfo.log"

	defaultPattern = "%J{tsTUSG}"
)

// 拼接数据源名称
func concatDataSourceName(dsnConfig *DSNConfig) string {
	return fmt.Sprintf("%s@(%s:%d)/%s",
		dsnConfig.UserName,
		dsnConfig.Endpoint.Address, dsnConfig.Endpoint.Port,
		dsnConfig.DBName)
}

// 钩子管理器
func NewHookManager(renderConfig *render.Config, dsnConfig *DSNConfig) *hook.Manager {
	return hook.NewManager().
		AddArg("dsn", concatDataSourceName(dsnConfig)).
		RegisterLogHook(renderConfig, patternMap).
		RegisterHook(func(hk *hook.Hook) {
			args := hk.Args()

			metric.GormRequestTotal.With(
				prometheus.Labels{
					"web_url":    render.PatternWebUrl(args).StringValue(),
					"web_method": render.PatternWebMethod(args).StringValue(),
					"dsn":        dsn(args).StringValue(),
					"operation":  operation(args).StringValue(),
				},
			).Inc()
		}, func(hk *hook.Hook) {
			args := hk.Args()

			metric.GormRequestDurationSummary.With(
				prometheus.Labels{
					"dsn":       dsn(args).StringValue(),
					"operation": operation(args).StringValue(),
				},
			).Observe(render.PatternDuration(args).Float64Value())
		}).
		RegisterTracingHook(func(hk *hook.Hook) string {
			return fmt.Sprintf("%s%s", tracing.SpanPrefixGorm, operation(hk.Args()).StringValue())
		}, func(hk *hook.Hook, span opentracing.Span) {
			args := hk.Args()

			span.SetTag("uuid", render.PatternUUID(args).StringValue())
			ext.DBType.Set(span, "mysql")
			ext.DBInstance.Set(span, dsn(args).StringValue())
		}, func(hk *hook.Hook, span opentracing.Span) {
			args := hk.Args()

			span.SetTag("db.method", operation(args).StringValue())
			span.SetTag("db.table", table(args).StringValue())
			span.SetTag("db.rows", rows(args).IntValue())
			ext.DBStatement.Set(span, fullSQL(args).StringValue())
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
					Add(gormSQL(args)).
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
	"s": render.PatternStartTime,
	"U": render.PatternUUID,
	"t": title,
	"G": gormSQL,
	"D": dsn,
	"d": render.PatternDuration,
	"R": rows,
	"L": level,
	"F": fullSQL,
	"o": operation,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "GORM")
}

// 汇总的gorm参数
func gormSQL(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("gorm", []render.PatternFunc{
		dsn, render.PatternDuration, rows, level, fullSQL,
	})(args)
}

// 数据源名称
func dsn(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("dsn", args.GetOrDefault("dsn", ""))
}

// 操作行数
func rows(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("rows", args.GetOrDefault("rows", -1))
}

// gorm日志等级
func level(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("level", args.GetOrDefault("level", ""))
}

// gorm操作符
func operation(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("operation", args.GetOrDefault("operation", ""))
}

// gorm操作表名
func table(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("table", args.GetOrDefault("table", ""))
}

// 完整sql
func fullSQL(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("sql", args.GetOrDefault("sql", ""))
}
