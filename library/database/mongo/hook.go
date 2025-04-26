package mongo

import (
	"fmt"
	"reflect"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	reflectUtil "gitlab.shanhai.int/sre/library/base/reflect"
	"gitlab.shanhai.int/sre/library/net/metric"
	"gitlab.shanhai.int/sre/library/net/sentry"
	"gitlab.shanhai.int/sre/library/net/tracing"
)

const (
	_infoFile = "mongoInfo.log"

	defaultPattern = "%J{tsTUSM}"
)

// 新建钩子管理器
func NewHookManager(renderConfig *render.Config, dsn string) *hook.Manager {
	return hook.NewManager().
		AddArg("dsn", dsn).
		RegisterLogHook(renderConfig, patternMap).
		RegisterHook(func(hk *hook.Hook) {
			args := hk.Args()

			metric.MongoRequestTotal.With(
				prometheus.Labels{
					"web_url":         render.PatternWebUrl(args).StringValue(),
					"web_method":      render.PatternWebMethod(args).StringValue(),
					"func_name":       funcName(args).StringValue(),
					"db_name":         dbName(args).StringValue(),
					"collection_name": collectionName(args).StringValue(),
				},
			).Inc()
		}, func(hk *hook.Hook) {
			args := hk.Args()

			metric.MongoRequestDurationSummary.With(
				prometheus.Labels{
					"func_name":       funcName(args).StringValue(),
					"db_name":         dbName(args).StringValue(),
					"collection_name": collectionName(args).StringValue(),
				},
			).Observe(render.PatternDuration(args).Float64Value())
		}).
		RegisterTracingHook(func(hk *hook.Hook) string {
			return fmt.Sprintf("%s%s", tracing.SpanPrefixMongo, funcName(hk.Args()).StringValue())
		}, func(hk *hook.Hook, span opentracing.Span) {
			args := hk.Args()

			span.SetTag("uuid", render.PatternUUID(args).StringValue())
			ext.DBType.Set(span, "mongo")
			ext.DBStatement.Set(span, funcName(args).StringValue())
			span.SetTag("db.collection", collectionName(args).StringValue())
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
					Add(mongoExtra(args)).
					Add(render.PatternSource(args)).
					Add(render.PatternStartTime(args)).
					Add(render.PatternEndTime(args)),
			}
		})
}

// 渲染模版
var patternMap = map[string]render.PatternFunc{
	"T": render.PatternEndTime,
	"s": render.PatternStartTime,
	"S": render.PatternSource,
	"U": render.PatternUUID,
	"t": title,
	"M": mongoExtra,
	"D": dsn,
	"d": render.PatternDuration,
	"N": dbName,
	"n": collectionName,
	"F": funcName,
	"f": filterField,
	"C": changeField,
	"E": extraField,
	"O": optionField,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "MONGO")
}

// 汇总的mongo参数
func mongoExtra(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("mongo", []render.PatternFunc{
		dsn, render.PatternDuration, dbName, collectionName, funcName,
		filterField, changeField, extraField, optionField,
	})(args)
}

// 数据源
func dsn(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("dsn", args["dsn"])
}

// db名称
func dbName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("db_name", args["db_name"])
}

// 集合名称
func collectionName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("collection_name", args["collection_name"])
}

// 调用函数名称
func funcName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("func_name", args["func_name"])
}

// 过滤的字段
func filterField(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("filter_field", args["filter_field"])
}

// 改变的字段
func changeField(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("change_field", args["change_field"])
}

// 额外的字段，如聚合管道
func extraField(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("extra_field", args["extra_field"])
}

// 参数字段
func optionField(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("option_field", getOptionsMap(args["option_field"]))
}

// 获取option字段的map
func getOptionsMap(value interface{}) (m map[string]string) {
	slice, err := reflectUtil.InterfaceToSlice(value)
	if err != nil {
		return nil
	}

	m = make(map[string]string)
	for _, item := range slice {
		itemValue := reflect.ValueOf(item).Elem()
		itemType := reflect.TypeOf(item).Elem()
		for i := 0; i < itemValue.NumField(); i++ {
			if !itemValue.Field(i).IsNil() {
				m[itemType.Field(i).Name] = fmt.Sprintf("%#v", itemValue.Field(i).Elem().Interface())
			}
		}
	}
	return
}
