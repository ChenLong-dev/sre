package redlock

import (
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

// 配置文件
type Config struct {
	// 锁过期时间
	ExpiryTime ctime.Duration `yaml:"expiryTime"`
	// 获取锁的尝试次数
	Tries int `yaml:"tries"`
	// 获取锁的重试延迟
	RetryDelay ctime.Duration `yaml:"retryDelay"`
	// 时钟漂移因子
	DriftFactor float64 `yaml:"driftFactor"`

	// 日志配置
	*render.Config `yaml:",inline"`
}
