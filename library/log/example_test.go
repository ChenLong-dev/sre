package log

import (
	"context"

	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func ExampleInfo() {
	// 常规
	Init(&Config{
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] [level: %L] %S %J{m}",
		},
	})
	// 简单日志方法，不建议使用该方法，应优先使用基础日志方法
	Info("hello %s", "world")
	Warn("hello %s", "world")
	Error("hello %s", "world")
	// 基础日志方法
	Infoc(context.Background(), "keys: %s %s...", "key1", "key2")
	// 自定义日志方法
	Infov(context.Background(), map[string]interface{}{
		"key":   2222222,
		"test2": "test",
	})

	// 最低日志等级
	Init(&Config{
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] [level: %L] %S %J{m}",
		},
		V: int(_errorLevel),
	})
	Info("hello %s", "world")
	Warn("hello %s", "world")

	// 日志过滤
	Init(&Config{
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] [level: %L] %S %J{m}",
		},
		Filter: []string{
			"key",
		},
	})
	Infov(context.Background(), map[string]interface{}{
		"key":   2222222,
		"test2": "test",
	})
}
