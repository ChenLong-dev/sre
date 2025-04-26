package log

import (
	"context"
	"os"

	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const defaultPattern = "%J{tTLUSm}"

// 默认控制台输出
var _defaultStdout = NewStdout("")

// 控制台处理器
type StdoutHandler struct {
	render render.Render
}

// 新建控制台处理器
func NewStdout(customPattern string) *StdoutHandler {
	if customPattern == "" {
		customPattern = defaultPattern
	}
	return &StdoutHandler{render: render.NewPatternRender(patternMap, customPattern)}
}

func (h *StdoutHandler) Log(ctx context.Context, lv Level, args map[string]interface{}) {
	// 增加额外参数
	addExtraField(ctx, args)

	if lv <= _infoLevel {
		h.render.Render(os.Stdout, args)
	} else {
		h.render.Render(os.Stderr, args)
	}
}

func (h *StdoutHandler) Close() error {
	h.render.Close()
	return nil
}

func (h *StdoutHandler) SetFormat(format string) {
	h.render = render.NewPatternRender(patternMap, format)
}
