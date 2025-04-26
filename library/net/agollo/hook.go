package agollo

import (
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	_infoFile      = "apollo.log"
	defaultPattern = "%J{tTSP}"
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
	"t": title,
	"E": extra,
	"A": appID,
	"C": cluster,
	"N": namespace,
	"I": ip,
	"c": changes,
	"W": watcherType,
	"P": apolloEvent,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "Apollo")
}

// 额外信息
func extra(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("extra", args["extra"])
}

// Apollo的appID
func appID(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("appID", args["appID"])
}

// Apollo的cluster
func cluster(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("cluster", args["cluster"])
}

// Apollo的namespace
func namespace(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("namespace", args["namespace"])
}

// Apollo的serveHost
func ip(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("serveHost", args["ip"])
}

// Apollo配置变化
func changes(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("changes", args["changes"])
}

// watcher标识
func watcherType(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("watcherType", args["watcherType"])
}

// Apollo配置汇总
func apolloEvent(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("apolloEvent",
		[]render.PatternFunc{
			appID, cluster, namespace, ip, changes, watcherType, extra,
		},
	)(args)
}
