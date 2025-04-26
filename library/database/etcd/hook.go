package etcd

import (
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	_infoFile = "etcdInfo.log"

	defaultPattern = "%J{tTUSE}"
)

// 新建钩子管理器
func NewHookManager(renderConfig *render.Config) *hook.Manager {
	return hook.NewManager().
		RegisterLogHook(renderConfig, patternMap)
}

// 渲染模版
var patternMap = map[string]render.PatternFunc{
	"T": render.PatternTime,
	"S": render.PatternSource,
	"U": render.PatternUUID,
	"t": title,
	"E": etcdExtra,
	"P": prefix,
	"K": key,
	"V": value,
	"e": extra,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "ETCD")
}

// 汇总的etcd参数
func etcdExtra(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("etcd", []render.PatternFunc{prefix, key, value, extra})(args)
}

// 前缀
func prefix(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("prefix", args["prefix"])
}

// 键
func key(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("key", args["key"])
}

// 值
func value(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("value", args["value"])
}

// 额外信息
func extra(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("extra", args["extra"])
}
