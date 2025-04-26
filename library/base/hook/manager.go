package hook

import (
	"context"
	"sync"

	"github.com/opentracing/opentracing-go"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/net/sentry"
	"gitlab.shanhai.int/sre/library/net/tracing"
)

// 钩子管理器
type Manager struct {
	// 前置方法链
	preChain []HandlerFunc
	// 后置方法链
	afterChain []HandlerFunc
	// 日志记录器
	logger Logger
	// 参数字典
	args *sync.Map
}

// 新建管理器
func NewManager() *Manager {
	manager := &Manager{
		preChain:   []HandlerFunc{},
		afterChain: []HandlerFunc{},
		args:       new(sync.Map),
	}
	return manager
}

// 增加参数
func (m *Manager) AddArg(key string, value interface{}) *Manager {
	m.args.Store(key, value)
	return m
}

// 注册前置钩子
func (m *Manager) RegisterPreHook(preHook HandlerFunc) *Manager {
	m.preChain = append(m.preChain, preHook)
	return m
}

// 注册后置钩子
func (m *Manager) RegisterAfterHook(afterHook HandlerFunc) *Manager {
	m.afterChain = append(m.afterChain, afterHook)
	return m
}

// 注册sentry面包屑钩子
func (m *Manager) RegisterSentryBreadCrumbHook(getBreadcrumb func(hk *Hook) *sentry.Breadcrumb) *Manager {
	m.RegisterAfterHook(func(hk *Hook) {
		hk.SetContext(sentry.AddBreadcrumb(hk.Context(), getBreadcrumb(hk)))
	})
	return m
}

// 注册钩子
func (m *Manager) RegisterHook(preHook HandlerFunc, afterHook HandlerFunc) *Manager {
	m.RegisterPreHook(preHook)
	m.RegisterAfterHook(afterHook)
	return m
}

// 注册链路跟踪钩子
func (m *Manager) RegisterTracingHook(
	getSpanNameFunc func(hk *Hook) string,
	preFunc func(hk *Hook, span opentracing.Span),
	afterFunc func(hk *Hook, span opentracing.Span),
) *Manager {
	m.RegisterHook(func(hk *Hook) {
		name := getSpanNameFunc(hk)
		if name == "" {
			return
		}

		parentSpan, err := tracing.GetCurrentSpanFromContext(hk.Context())
		if err != nil {
			return
		}
		span := parentSpan.Tracer().StartSpan(name, opentracing.ChildOf(parentSpan.Context()))
		preFunc(hk, span)
		hk.SetContext(tracing.SetCurrentSpanToContext(hk.Context(), span))
	}, func(hk *Hook) {
		span, err := tracing.GetCurrentSpanFromContext(hk.Context())
		if err != nil {
			return
		}
		defer span.Finish()
		afterFunc(hk, span)
	})
	return m
}

// 注册日志钩子
func (m *Manager) RegisterLogHook(logConfig *render.Config, patternMap map[string]render.PatternFunc) *Manager {
	// 避免重复注册
	if m.logger != nil {
		panic("already register log handler")
	}

	m.SetLogger(GetDefaultLogger(logConfig, patternMap))
	m.RegisterAfterHook(func(hook *Hook) {
		hook.logger.Print(hook.args)
	})

	return m
}

// 设置日志记录器
func (m *Manager) SetLogger(logger Logger) *Manager {
	m.logger = logger
	return m
}

// 获取日志记录器
func (m *Manager) GetLogger() Logger {
	return m.logger
}

// 关闭
func (m *Manager) Close() {
	if m.logger != nil {
		m.logger.Close()
	}
}

// 创建钩子
//
// 注意：不允许多线程操作同一hook
func (m *Manager) CreateHook(ctx context.Context) *Hook {
	// 拷贝
	args := make(map[string]interface{})
	m.args.Range(func(key, value interface{}) bool {
		args[key.(string)] = value
		return true
	})
	return &Hook{
		preChain:   m.preChain,
		afterChain: m.afterChain,
		logger:     m.logger,
		args:       args,
		ctx:        ctx,
	}
}
