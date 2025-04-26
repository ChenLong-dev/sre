package redis

import (
	"github.com/gomodule/redigo/redis"
	"gitlab.shanhai.int/sre/library/base/hook"
)

// 连接池
type Pool struct {
	// 连接池
	*redis.Pool
	// 配置文件
	config *Config
	// 钩子管理器
	manager *hook.Manager
}

// 获取连接
func (p *Pool) Get() *Conn {
	con := p.Pool.Get()

	return &Conn{
		Conn:    con,
		manager: p.manager,
		pool:    p,
	}
}

// 获取连接，执行命令，并关闭连接
func (p *Pool) WrapDo(doFunction func(con *Conn) error) error {
	con := p.Get()
	defer con.Close()

	return doFunction(con)
}

func (p *Pool) Close() (err error) {
	p.manager.Close()
	return
}
