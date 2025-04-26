package render

import (
	"fmt"
	"strconv"
	"time"
)

// 常用pattern的key
const (
	WebUrlArgKey    = "web_url"
	WebMethodArgKey = "web_method"
	UUIDArgKey      = "uuid"
	SourceArgKey    = "source"
	StartTimeArgKey = "start_time"
	EndTimeArgKey   = "end_time"
	DurationArgKey  = "duration"
	ErrorArgKey     = "error"
)

// web url
func PatternWebUrl(args PatternArgs) PatternResult {
	return NewPatternResult("web_url", args.GetOrDefault(WebUrlArgKey, ""))
}

// web url
func PatternWebMethod(args PatternArgs) PatternResult {
	return NewPatternResult("web_method", args.GetOrDefault(WebMethodArgKey, ""))
}

// 操作持续时间
// 毫秒级别
func PatternDuration(args PatternArgs) PatternResult {
	originDuration, ok := args[DurationArgKey]
	if !ok {
		return NewPatternResult("duration", -1)
	}

	originValue := float64(originDuration.(time.Duration).Nanoseconds()/1e4) / 100.0
	duration, err := strconv.ParseFloat(fmt.Sprintf("%.2f", originValue), 64)
	if err != nil {
		duration = originValue
	}

	return NewPatternResult("duration", duration)
}

// 当前时间
func PatternTime(args PatternArgs) PatternResult {
	return NewPatternResult("time", time.Now().Format("2006/01/02 15:04:05.000"))
}

// 结束时间
func PatternEndTime(args PatternArgs) PatternResult {
	t, ok := args[EndTimeArgKey].(time.Time)
	if !ok {
		return NewPatternResult("time", "")
	}
	return NewPatternResult("time", t.Format("2006/01/02 15:04:05.000"))
}

// 开始时间
func PatternStartTime(args PatternArgs) PatternResult {
	t, ok := args[StartTimeArgKey].(time.Time)
	if !ok {
		return NewPatternResult("start_time", "")
	}
	return NewPatternResult("start_time", t.Format("2006/01/02 15:04:05.000"))
}

// 日志打印的调用源
func PatternSource(args PatternArgs) PatternResult {
	return NewPatternResult("source", args.GetOrDefault(SourceArgKey, ""))
}

// context中的uuid
func PatternUUID(args PatternArgs) PatternResult {
	return NewPatternResult("uuid", args.GetOrDefault(UUIDArgKey, "unknown"))
}

// error
func PatternError(args PatternArgs) PatternResult {
	err := args.GetOrDefault(ErrorArgKey, nil)
	if err == nil {
		return DefaultPatternResult()
	}
	return NewPatternResult("error", fmt.Sprintf("%v", err))
}
