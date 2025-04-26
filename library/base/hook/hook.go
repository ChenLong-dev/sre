package hook

import (
	"context"

	_context "gitlab.shanhai.int/sre/library/base/context"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

// 处理方法
type HandlerFunc func(hk *Hook)

// 钩子
//
// 注意：不允许多线程操作同一hook
type Hook struct {
	// 前置方法链
	preChain []HandlerFunc
	// 后置方法链
	afterChain []HandlerFunc
	// 日志记录器
	logger Logger
	// 参数字典
	args map[string]interface{}
	// 上下文
	ctx context.Context
}

// 增加参数
func (h *Hook) AddArg(key string, value interface{}) *Hook {
	h.args[key] = value
	return h
}

// 获取全部参数
func (h *Hook) Args() map[string]interface{} {
	return h.args
}

// 获取参数
func (h *Hook) Arg(key string) interface{} {
	return h.args[key]
}

// 获取日志记录器
func (h *Hook) GetLogger() Logger {
	return h.logger
}

// 设置上下文
func (h *Hook) SetContext(ctx context.Context) *Hook {
	h.ctx = ctx
	return h
}

// 获取上下文
func (h *Hook) Context() context.Context {
	return h.ctx
}

// 准备系统参数
func (h *Hook) prepareSystemArgs() *Hook {
	// 增加必要参数
	h.AddArg(render.UUIDArgKey, _context.GetStringOrDefault(h.ctx, _context.ContextUUIDKey, "unknown")).
		AddArg(render.WebUrlArgKey, _context.GetString(h.ctx, _context.ContextRequestPathKey)).
		AddArg(render.WebMethodArgKey, _context.GetString(h.ctx, _context.ContextRequestMethodKey))

	return h
}

// 执行前置钩子
func (h *Hook) ProcessPreHook() *Hook {
	h.prepareSystemArgs()
	for _, f := range h.preChain {
		f(h)
	}
	return h
}

// 执行后置钩子
func (h *Hook) ProcessAfterHook() *Hook {
	h.prepareSystemArgs()
	for _, f := range h.afterChain {
		f(h)
	}
	return h
}

// 执行方法
func (h *Hook) Do(f func()) {
	h.ProcessPreHook()
	f()
	h.ProcessAfterHook()
}
