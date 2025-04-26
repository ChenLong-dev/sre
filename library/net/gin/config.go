package gin

import (
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

type EndpointConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type Config struct {
	// 服务监听地址
	Endpoint *EndpointConfig `yaml:"endpoint"`
	// 服务请求超时时间
	Timeout ctime.Duration `yaml:"timeout"`

	// 日志配置
	*render.Config `yaml:",inline"`
	// 路由日志文件渲染模版
	OutRouterPattern string `yaml:"outRouterPattern"`
	// 控制台路由日志渲染模版
	StdoutRouterPattern string `yaml:"stdoutRouterPattern"`
	// 是否打印请求body
	RequestBodyOut bool `yaml:"requestBodyOut"`
}
