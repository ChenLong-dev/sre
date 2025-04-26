package mongo

import (
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

type DSNConfig struct {
	UserName  string            `yaml:"userName"`
	Password  string            `yaml:"password"`
	Endpoints []*EndpointConfig `yaml:"endpoints"`
	DBName    string            `yaml:"dbName"`
	Options   []string          `yaml:"options"`
}

type EndpointConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type Config struct {
	// 主dsn配置
	DSN *DSNConfig `yaml:"dsn"`
	// 只读dsn配置
	ReadDSN []*DSNConfig `yaml:"readDSN"`
	// 执行超时时间
	ExecTimeout ctime.Duration `yaml:"execTimeout"`
	// 查询超时时间
	QueryTimeout ctime.Duration `yaml:"queryTimeout"`
	// 连接闲置超时时间
	IdleTimeout ctime.Duration `yaml:"idleTimeout"`
	// 连接池最大数量
	MaxPoolSize int `yaml:"maxPoolSize"`
	// 连接池最小数量
	MinPoolSize int `yaml:"minPoolSize"`

	// 日志配置
	*render.Config `yaml:",inline"`
}
