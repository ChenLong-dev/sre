package tracing

import (
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

type Config struct {
	// 服务名
	AppName string `yaml:"appName"`
	// 采样器配置
	Sampler *SamplerConfig `yaml:"sampler"`
	// 报告器配置
	Reporter *ReporterConfig `yaml:"reporter"`

	// 日志配置
	*render.Config `yaml:",inline"`
}

// 采样器配置
type SamplerConfig struct {
	// 采样器类型
	// jaeger支持的类型:const,probabilistic,rateLimiting,remote
	// zipkin支持的类型:modulo,boundary,counting
	Type string `yaml:"type"`
	// 采样器参数
	// jaeger支持的参数:
	//	const:0或1代表是否开关
	//	probabilistic:在0-1之间
	// 	rateLimiting:每秒span的数量
	//	remote:与probabilistic类型相同
	// zipkin支持的参数:
	//	modulo:取模的值
	//	boundary:包含两个参数，使用逗号 ',' 分割，第一个参数代表比例，在0-1之间，第二个参数代表id盐，为int类型
	//	counting:比例，在0-1之间
	Param string `yaml:"param"`

	// 仅jaeger支持
	// 采样服务地址
	SamplingServerURL string `yaml:"samplingServerURL"`
	// 仅jaeger支持
	// 为采样器将跟踪的最大操作数，仅在remote类型下可用
	MaxOperations int `yaml:"maxOperations"`
	// 仅jaeger支持
	// 控制远程采样器的轮询时间，仅在remote类型下可用
	SamplingRefreshInterval ctime.Duration `yaml:"samplingRefreshInterval"`
}

// 报告器配置
type ReporterConfig struct {
	// 采样器地址
	CollectorEndpoint string `yaml:"collectorEndpoint"`

	// 仅zipkin支持
	// 本地服务节点
	LocalEndpoint string `yaml:"localEndpoint"`

	// 仅jaeger支持
	// 内存span最大数量
	QueueSize int `yaml:"queueSize"`
	// 仅jaeger支持
	// 内存刷新时间
	BufferFlushInterval ctime.Duration `yaml:"bufferFlushInterval"`
	// 仅jaeger支持
	// 是否记录所有的span
	LogSpans bool `yaml:"logSpans"`
	// 仅jaeger支持
	// 本地接受span数据代理，用于收集后批量发送
	LocalAgentHostPort string `yaml:"localAgentHostPort"`
	// 仅jaeger支持
	// 收集器验证用户
	User string `yaml:"user"`
	// 仅jaeger支持
	// 收集器验证密码
	Password string `yaml:"password"`
}
