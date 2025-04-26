package service

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"

	_redis "gitlab.shanhai.int/sre/library/database/redis"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// ExposedPortsWhiteListKey 支持额外端口的白名单缓存键
const ExposedPortsWhiteListKey = "exposed_ports_white_list"

// CheckExposedPortsWhiteList 校验应用是否在支持额外端口的白名单内
func (s *Service) CheckExposedPortsWhiteList(ctx context.Context, appID string) (bool, error) {
	var ok bool
	err := s.dao.Redis.WrapDo(func(con *_redis.Conn) error {
		var e error
		ok, e = redis.Bool(con.Do(ctx, "hexists", ExposedPortsWhiteListKey, appID))
		if e != nil && e != redis.ErrNil {
			return errors.Wrap(errcode.RedisError, e.Error())
		}

		return nil
	})

	return ok, err
}
