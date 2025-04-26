package httpclient

import (
	"time"

	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	DefaultMaxIdleConns        = 100
	DefaultMaxIdleConnsPerHost = 2
	DefaultIdleConnTimeout     = time.Second * 90
)

type Config struct {
	// 请求超时时间
	RequestTimeout ctime.Duration `yaml:"requestTimeout"`
	// 是否打印响应body
	ResponseBodyOut bool `yaml:"responseBodyOut"`
	// 是否打印请求body
	RequestBodyOut bool `yaml:"requestBodyOut"`

	// 是否关闭断路器
	DisableBreaker bool `yaml:"disableBreaker"`
	// 断路器断路最小错误比例 [0,1]
	BreakerRate float64 `yaml:"breakerRate"`
	// 断路器断路最小采样数
	BreakerMinSample int `yaml:"breakerMinSample"`

	// 是否关闭链路跟踪
	DisableTracing bool `yaml:"disableTracing"`
	// 是否关闭sentry
	DisableSentry bool `yaml:"disableSentry"`
	// 开启负载均衡
	EnableLoadBalancer bool `yaml:"enableLoadBalancer"`

	// 最大空闲连接
	MaxIdleConns int `yaml:"maxIdleConns"`
	// 每个Host的最大空闲连接
	MaxIdleConnsPerHost int `yaml:"maxIdleConnsPerHost"`
	// 每个Host的最大连接
	MaxConnsPerHost int `yaml:"maxConnsPerHost"`
	// 空闲连接超时时间
	IdleConnTimeout ctime.Duration `yaml:"idleConnTimeout"`
	// 是否保持长连接
	DisableKeepAlives bool `yaml:"disableKeepAlives"`

	// 日志配置
	*render.Config `yaml:",inline"`
}
