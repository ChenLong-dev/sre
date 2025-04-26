package redlock

import (
	"time"

	"github.com/go-redsync/redsync"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/database/redis"
)

// 客户端
type RedLock struct {
	// 管理工具
	*redsync.Redsync
	// 配置文件
	config *Config
}

// 新建客户端
func New(c *Config, pools ...*redis.Pool) *RedLock {
	if c.Config == nil {
		c.Config = &render.Config{}
	}
	if c.Config.StdoutPattern == "" {
		c.Config.StdoutPattern = defaultPattern
	}
	if c.Config.OutPattern == "" {
		c.Config.OutPattern = defaultPattern
	}
	if c.Config.OutFile == "" {
		c.Config.OutFile = _infoFile
	}

	redPool := make([]redsync.Pool, 0)
	for _, r := range pools {
		redPool = append(redPool, &Pool{
			Pool: r,
		})
	}

	return &RedLock{
		Redsync: redsync.New(redPool),
		config:  c,
	}
}

// 新建分布式锁
func (r *RedLock) NewMutex(name string, options ...redsync.Option) *Mutex {
	options = append(r.getConfigOptions(), options...)

	return &Mutex{
		name:    name,
		Mutex:   r.Redsync.NewMutex(name, options...),
		manager: NewHookManager(r.config.Config),
	}
}

// 获取配置文件中的锁配置
func (r *RedLock) getConfigOptions() []redsync.Option {
	var options []redsync.Option

	if r.config.ExpiryTime != 0 {
		options = append(options, redsync.SetExpiry(time.Duration(r.config.ExpiryTime)))
	}

	if r.config.Tries != 0 {
		options = append(options, redsync.SetTries(r.config.Tries))
	}

	if r.config.RetryDelay != 0 {
		options = append(options, redsync.SetRetryDelay(time.Duration(r.config.RetryDelay)))
	}

	if r.config.DriftFactor != 0 {
		options = append(options, redsync.SetDriftFactor(r.config.DriftFactor))
	}

	return options
}
