package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

// 新建连接池
func NewPool(cfg *Config) *Pool {
	if cfg == nil {
		panic("redis config is nil")
	}
	if cfg.Proto == "" || cfg.Endpoint == nil {
		panic("redis must be set proto/addr")
	}

	if cfg.Config == nil {
		cfg.Config = &render.Config{}
	}
	if cfg.Config.StdoutPattern == "" {
		cfg.Config.StdoutPattern = defaultPattern
	}
	if cfg.Config.OutPattern == "" {
		cfg.Config.OutPattern = defaultPattern
	}
	if cfg.Config.OutFile == "" {
		cfg.Config.OutFile = _infoFile
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = ctime.Duration(DefaultConnectTimeout)
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = ctime.Duration(DefaultReadTimeout)
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = ctime.Duration(DefaultWriteTimeout)
	}

	pool := &Pool{
		Pool: &redis.Pool{
			MaxIdle:         cfg.PoolConfig.Idle,
			IdleTimeout:     time.Duration(cfg.PoolConfig.IdleTimeout),
			MaxActive:       cfg.PoolConfig.Active,
			Wait:            cfg.PoolConfig.Wait,
			MaxConnLifetime: time.Duration(cfg.MaxConnLifetime),
			Dial: func() (redis.Conn, error) {
				endpoint := fmt.Sprintf("%s:%d", cfg.Endpoint.Address, cfg.Endpoint.Port)
				c, err := redis.Dial(
					cfg.Proto,
					endpoint,
					redis.DialConnectTimeout(time.Duration(cfg.ConnectTimeout)),
					redis.DialReadTimeout(time.Duration(cfg.ReadTimeout)),
					redis.DialWriteTimeout(time.Duration(cfg.WriteTimeout)),
				)
				if err != nil {
					return nil, err
				}

				if cfg.Auth != "" {
					if _, err := c.Do("AUTH", cfg.Auth); err != nil {
						c.Close()
						return nil, err
					}
				}

				if cfg.DB != 0 {
					if _, err := c.Do("SELECT", cfg.DB); err != nil {
						c.Close()
						return nil, err
					}
					return c, nil
				}

				return c, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Duration(cfg.PoolConfig.CheckTime) {
					return nil
				}
				_, err := c.Do("PING")
				return err
			},
		},
		config:  cfg,
		manager: NewHookManager(cfg.Config),
	}

	err := pool.WrapDo(func(con *Conn) error {
		_, err := con.ping(context.Background())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(errors.Wrap(err, "redis health check error"))
	}

	return pool
}
