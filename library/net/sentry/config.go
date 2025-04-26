package sentry

import render "gitlab.shanhai.int/sre/library/base/logrender"

// sentry初始化配置
type Config struct {
	// sentry DSN
	DSN string `yaml:"dsn"`
	// 标签
	Tags map[string]string `yaml:"tags"`
	// 项目环境
	Environment string `yaml:"env"`
	// 过滤器
	ErrCodeFilter []string `yaml:"errCodeFilter"`
	// 日志配置
	*render.Config `yaml:",inline"`
}
