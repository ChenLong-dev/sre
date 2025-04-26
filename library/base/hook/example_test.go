package hook

import (
	"context"
	"time"

	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/base/runtime"
)

var patternMap = map[string]render.PatternFunc{
	"S": render.PatternStartTime,
	"E": render.PatternEndTime,
	"s": render.PatternSource,
}

func ExampleNewManager() {
	// 创建hook
	hk := NewManager().
		RegisterLogHook(&render.Config{}, patternMap).
		CreateHook(context.Background()).
		AddArg(render.StartTimeArgKey, time.Now()).
		AddArg(render.EndTimeArgKey, time.Now()).
		AddArg(render.SourceArgKey, runtime.GetDefaultFilterCallers())

	// 执行前置钩子
	hk.ProcessPreHook()

	// 执行后置钩子
	hk.ProcessAfterHook()
}
