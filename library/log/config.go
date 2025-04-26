package log

import render "gitlab.shanhai.int/sre/library/base/logrender"

// Config log config.
type Config struct {
	// App id
	AppID string `yaml:"appID"`
	// 服务运行的host
	Host string `yaml:"host"`
	// 打印的最低日志级别
	V int `yaml:"v"`
	// 过滤日志中指定的key，并用***代替
	Filter []string `yaml:"filter"`

	// 日志配置
	*render.Config `yaml:",inline"`
}
