package log

import (
	"context"
	"fmt"
	"os"

	render "gitlab.shanhai.int/sre/library/base/logrender"
)

var (
	// 默认日志处理器
	h Handler
	// 配置文件
	c *Config
)

// 默认初始化
func init() {
	host, _ := os.Hostname()
	c = &Config{
		Host: host,
	}
	h = newDefaultBatchHandler(c, []string{}, NewStdout(""))
}

// 初始化日志库
func Init(conf *Config) {
	var isNil bool
	if conf == nil {
		isNil = true
		conf = &Config{
			Config: &render.Config{
				Stdout: true,
			},
		}
	}
	if len(conf.Host) == 0 {
		host, _ := os.Hostname()
		conf.Host = host
	}
	if conf.Config == nil {
		conf.Config = &render.Config{}
	}

	// 增加日志处理器
	var hs []Handler
	if conf.Stdout || isNil {
		hs = append(hs, NewStdout(conf.StdoutPattern))
	}
	if conf.OutDir != "" {
		hs = append(hs, NewFile(conf.OutPattern,
			conf.OutDir,
			conf.FileBufferSize,
			conf.RotateSize,
			conf.MaxLogFile,
		))
	}

	c = conf
	h = newDefaultBatchHandler(c, conf.Filter, hs...)
}

// 简单Info日志，应优先使用Infoc方法
func Info(format string, args ...interface{}) {
	h.Log(context.Background(), _infoLevel, map[string]interface{}{
		_log: fmt.Sprintf(format, args...),
	})
}

// 简单Warn日志，应优先使用Warnc方法
func Warn(format string, args ...interface{}) {
	h.Log(context.Background(), _warnLevel, map[string]interface{}{
		_log: fmt.Sprintf(format, args...),
	})
}

// 简单Error日志，应优先使用Errorc方法
func Error(format string, args ...interface{}) {
	h.Log(context.Background(), _errorLevel, map[string]interface{}{
		_log: fmt.Sprintf(format, args...),
	})
}

// 基础Info日志
func Infoc(ctx context.Context, format string, args ...interface{}) {
	h.Log(ctx, _infoLevel, map[string]interface{}{
		_log: fmt.Sprintf(format, args...),
	})
}

// 基础Error日志
func Errorc(ctx context.Context, format string, args ...interface{}) {
	h.Log(ctx, _errorLevel, map[string]interface{}{
		_log: fmt.Sprintf(format, args...),
	})
}

// 基础Warn日志
func Warnc(ctx context.Context, format string, args ...interface{}) {
	h.Log(ctx, _warnLevel, map[string]interface{}{
		_log: fmt.Sprintf(format, args...),
	})
}

// 自定义Info日志
func Infov(ctx context.Context, args map[string]interface{}) {
	h.Log(ctx, _infoLevel, args)
}

// 自定义Warn日志
func Warnv(ctx context.Context, args map[string]interface{}) {
	h.Log(ctx, _warnLevel, args)
}

// 自定义Error日志
func Errorv(ctx context.Context, args map[string]interface{}) {
	h.Log(ctx, _errorLevel, args)
}

// 设置渲染格式
func SetFormat(format string) {
	h.SetFormat(format)
}

// 关闭日志
func Close() (err error) {
	err = h.Close()
	h = _defaultStdout
	return
}
