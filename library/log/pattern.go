package log

import (
	"fmt"
	"strings"
	"sync"

	render "gitlab.shanhai.int/sre/library/base/logrender"
)

var patternMap = map[string]render.PatternFunc{
	"T": render.PatternTime,
	"L": keyFactory("level", _level),
	"S": keyFactory("source", _source),
	"M": textMessage,
	"m": jsonMessage,
	"U": render.PatternUUID,
	"t": title,
}

// 基础key加工
func keyFactory(field, key string) render.PatternFunc {
	return func(args render.PatternArgs) render.PatternResult {
		if v, ok := args[key]; ok {
			if s, ok := v.(string); ok {
				return render.NewPatternResult(field, s)
			}
			return render.NewPatternResult(field, fmt.Sprint(v))
		}
		return render.NewPatternResult(field, "")
	}
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "LOG")
}

// 是否是内部键，内部键不打印到消息主体中
func isInternalKey(k string) bool {
	switch k {
	case _level, _levelValue, _time, _source, _appID, _uuid:
		return true
	}
	return false
}

// 消息主体,用于json形式
func jsonMessage(args render.PatternArgs) render.PatternResult {
	var m string
	msgMap := make(map[string]interface{})

	for k, v := range args {
		if k == _log {
			m = fmt.Sprint(v)
			continue
		}

		// 添加非内部键值对
		if isInternalKey(k) {
			continue
		}
		msgMap[k] = v
	}

	if m != "" {
		// 追加消息主体
		msgMap["log"] = m
	}

	return render.NewPatternResult("message", msgMap)
}

// 消息切片对象池
var messageSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0)
	},
}

// 消息主体,用于text形式
func textMessage(args render.PatternArgs) render.PatternResult {
	var m string
	s := messageSlicePool.Get().([]string)

	for k, v := range args {
		if k == _log {
			m = fmt.Sprint(v)
			continue
		}

		// 添加非内部键值对
		if isInternalKey(k) {
			continue
		}
		s = append(s, fmt.Sprintf("%s=%v", k, v))
	}
	// 追加消息主体
	msg := strings.Join(append(s, m), " ")

	// reset 对象池
	messageSlicePool.Put(s[0:0])

	return render.NewPatternResult("message", msg)
}
