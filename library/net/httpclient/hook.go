package httpclient

import (
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	_infoFile = "httpClientInfo.log"

	defaultPattern = "%J{taTUSH}"
)

// 新建钩子管理器
func NewHookManager() *hook.Manager {
	return hook.NewManager()
}

// 渲染模版
var patternMap = map[string]render.PatternFunc{
	"T": render.PatternEndTime,
	"S": render.PatternSource,
	"a": render.PatternStartTime,
	"U": render.PatternUUID,
	"t": title,
	"H": httpClientExtra,
	"s": statusCode,
	"B": responseBody,
	"D": render.PatternDuration,
	"e": endpoint,
	"u": urlFuc,
	"h": headersFuc,
	"b": requestBodyFuc,
	"M": methodName,
	"E": extraMessage,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "HTTPCLIENT")
}

// 汇总的http client参数
func httpClientExtra(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("httpclient", []render.PatternFunc{
		statusCode, responseBody, render.PatternDuration, urlFuc,
		headersFuc, requestBodyFuc, methodName,
		endpoint, extraMessage,
	})(args)
}

// 额外信息
func extraMessage(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("extra_message", args.GetOrDefault("extra_message", ""))
}

// 状态码
func statusCode(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("status_code", args.GetOrDefault("status_code", -1))
}

// 响应body
func responseBody(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("response_body", args.GetOrDefault("response_body", "unknown"))
}

// 请求url
func endpoint(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("endpoint", args["endpoint"])
}

// 请求host
func host(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("host", args["host"])
}

// 请求url
func urlFuc(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("url", args["url"])
}

// 请求头
func headersFuc(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("headers", args["headers"])
}

// 请求body
func requestBodyFuc(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("request_body", args.GetOrDefault("request_body", "unknown"))
}

// 请求方法
func methodName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("method_name", args["method_name"])
}
