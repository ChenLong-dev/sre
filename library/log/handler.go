package log

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/runtime"
)

const (
	// 时间格式化
	_timeFormat = "2006-01-02T15:04:05.999999"
	// 当前时间
	_time = "time"
	// 日志级别值
	_levelValue = "level_value"
	//  日志级别名
	_level = "level"
	// 日志源
	_source = "source"
	// 日志消息
	_log = "log"
	// App ID
	_appID = "app_id"
	// UUID
	_uuid = "uuid"
)

// 日志处理接口
type Handler interface {
	// 打印日志
	Log(context.Context, Level, map[string]interface{})

	// 设置渲染格式
	SetFormat(string)

	// 关闭
	Close() error
}

// 默认日志批量处理器
type DefaultBatchHandler struct {
	// 过滤的字段
	filters map[string]struct{}
	// 所包含的日志处理器
	handlers []Handler
	// 配置文件
	config *Config
}

func (hs DefaultBatchHandler) Log(ctx context.Context, lv Level, args map[string]interface{}) {
	// 过滤日志级别
	if hs.config.V > int(lv) {
		return
	}

	// 过滤字段
	hasSource := false
	for k := range args {
		if _, ok := hs.filters[k]; ok {
			args[k] = "***"
		}
		if k == _source {
			hasSource = true
		}
	}

	// 如果没有日志源则增加
	if !hasSource {
		fn := runtime.GetDefaultFilterCallers()
		args[_source] = fn
	}

	// 增加必要信息
	args[_time] = time.Now()
	args[_levelValue] = lv
	args[_level] = lv.String()

	for _, h := range hs.handlers {
		h.Log(ctx, lv, args)
	}
}

func (hs DefaultBatchHandler) Close() (err error) {
	for _, h := range hs.handlers {
		if e := h.Close(); e != nil {
			err = errors.WithStack(e)
		}
	}
	return
}

func (hs DefaultBatchHandler) SetFormat(format string) {
	for _, h := range hs.handlers {
		h.SetFormat(format)
	}
}

// 新建默认日志批量处理器
func newDefaultBatchHandler(config *Config, filters []string, handlers ...Handler) *DefaultBatchHandler {
	set := make(map[string]struct{})
	for _, k := range filters {
		set[k] = struct{}{}
	}
	return &DefaultBatchHandler{
		config:   config,
		filters:  set,
		handlers: handlers,
	}
}
