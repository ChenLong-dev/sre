package redlock

import (
	"github.com/gomodule/redigo/redis"
	redisUtil "gitlab.shanhai.int/sre/library/database/redis"
)

// 封装的连接池
type Pool struct {
	*redisUtil.Pool
}

// 实现redsync.Pool接口
func (p *Pool) Get() redis.Conn {
	return p.Pool.Get().GetOriginConnect()
}
