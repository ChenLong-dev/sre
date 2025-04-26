package redis

import (
	"fmt"
	"time"

	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	DefaultReadTimeout    = time.Second * 3
	DefaultWriteTimeout   = time.Second * 3
	DefaultConnectTimeout = time.Second * 10
)

// 连接池配置
type PoolConfig struct {
	// 最大可用数量
	Active int `yaml:"active"`
	// 最大闲置数量
	Idle int `yaml:"idle"`
	// 闲置超时时间
	IdleTimeout ctime.Duration `yaml:"idleTimeout"`
	// 检查可用时间
	CheckTime ctime.Duration `yaml:"checkTime"`
	// 当连接数满时是否等待连接
	Wait bool `yaml:"wait"`
	// 读取命令超时时间
	ReadTimeout ctime.Duration `yaml:"readTimeout"`
	// 写入命令超时时间
	WriteTimeout ctime.Duration `yaml:"writeTimeout"`
	// 连接超时时间
	ConnectTimeout ctime.Duration `yaml:"connectTimeout"`
}

type EndpointConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

// 配置文件
type Config struct {
	// 连接池配置
	*PoolConfig `yaml:",inline"`

	// 连接协议
	Proto string `yaml:"proto"`
	// 数据库名
	DB int `yaml:"db"`
	// 连接地址
	Endpoint *EndpointConfig `yaml:"endpoint"`
	// 校验密码
	Auth string `yaml:"auth"`
	// 连接完整生命周期时间
	MaxConnLifetime ctime.Duration `yaml:"maxConnLifetime"`

	// 日志配置
	*render.Config `yaml:",inline"`
}

// 获取 Redis 连接地址
func (c *Config) GetEndpoint() string {
	return fmt.Sprintf("%s:%d", c.Endpoint.Address, c.Endpoint.Port)
}
