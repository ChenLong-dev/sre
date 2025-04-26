package queue

import (
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	_infoFile = "queueInfo.log"

	defaultPattern = "%J{tTUA}"
)

// 上下文信息
type spanInfo struct {
	// 消息
	msg []byte
	// 消息类型
	contentType string
	// 额外信息
	extra string
	// 错误
	err error
	// 消费者名称
	consumerName string
}

// 新建钩子管理器
func NewHookManager(renderConfig *render.Config) *hook.Manager {
	return hook.NewManager().
		RegisterLogHook(renderConfig, patternMap)
}

// 渲染模版
var patternMap = map[string]render.PatternFunc{
	"T": render.PatternTime,
	"U": render.PatternUUID,
	"t": title,
	"A": amqpExtra,
	"N": queueName,
	"x": exchangeName,
	"r": routingKey,
	"C": contentType,
	"B": body,
	"e": extra,
	"n": consumerName,
	"i": sessionID,
	"E": render.PatternError,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "AMQP")
}

// 汇总的queue参数
func amqpExtra(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("amqp", []render.PatternFunc{
		queueName, exchangeName, render.PatternError,
		routingKey, contentType, body, extra,
		consumerName, sessionID,
	})(args)
}

// 队列名
func queueName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("queue_name", args.GetOrDefault("queue_name", "unknown"))
}

// 交换机名
func exchangeName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("exchange_name", args.GetOrDefault("exchange_name", "unknown"))
}

// 路由key
func routingKey(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("routing_key", args.GetOrDefault("routing_key", "unknown"))
}

// 消息类型
func contentType(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("content_type", args.GetOrDefault("content_type", ""))
}

// 消息体
func body(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("body", args.GetOrDefault("body", ""))
}

// 额外消息
func extra(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("extra", args.GetOrDefault("extra", ""))
}

// 消费者名称
func consumerName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("consumer_name", args.GetOrDefault("consumer_name", ""))
}

// 会话ID
func sessionID(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("session_id", args.GetOrDefault("session_id", ""))
}
