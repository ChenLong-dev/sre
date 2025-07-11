package sql

import (
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

type EndpointConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

// DSN配置
type DSNConfig struct {
	UserName string          `yaml:"userName"`
	Password string          `yaml:"password"`
	Endpoint *EndpointConfig `yaml:"endpoint"`
	DBName   string          `yaml:"dbName"`
	Options  []string        `yaml:"options"`
}

// 配置文件
type Config struct {
	// 主dsn配置
	DSN *DSNConfig `yaml:"dsn"`
	// 只读dsn配置
	ReadDSN []*DSNConfig `yaml:"readDSN"`
	// 最大可用数量
	Active int `yaml:"active"`
	// 最大闲置数量
	Idle int `yaml:"idle"`
	// 闲置超时时间
	IdleTimeout ctime.Duration `yaml:"idleTimeout"`
	// 查询超时时间
	QueryTimeout ctime.Duration `yaml:"queryTimeout"`
	// 执行超时时间
	ExecTimeout ctime.Duration `yaml:"execTimeout"`
	// 事务超时时间
	TranTimeout ctime.Duration `yaml:"tranTimeout"`

	// 日志配置
	*render.Config `yaml:",inline"`
}
