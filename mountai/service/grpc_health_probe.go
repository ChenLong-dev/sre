package service

import (
	"context"
	"encoding/json"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"

	_redis "gitlab.shanhai.int/sre/library/database/redis"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

type gRPCHealthProbeConfig struct {
	Port string `json:"port"`
	TLS  bool   `json:"tls"`
}

// GRPCHealthProbeWhiteListKey 使用 grpc-health-probe 健康检查方式的白名单缓存键
const GRPCHealthProbeWhiteListKey = "grpc_health_probe_white_list"

// getGRPCHealthProbeConfig 检查应用是否在使用 grpc-health-probe 健康检查方式的白名单中并获取配置
// NOTE: 白名单目前在 redis 内手动添加，不提供 API(临时需求不做太复杂)
// TODO: 推动所有 GRPC 服务使用该方式进行健康检查后该功能及时下线, 改为固定的 GRPC 健康检查通用配置
func (s *Service) getGRPCHealthProbeConfig(ctx context.Context, appID string) (*gRPCHealthProbeConfig, error) {
	var cfg *gRPCHealthProbeConfig
	err := s.dao.Redis.WrapDo(func(con *_redis.Conn) error {
		value, e := redis.Bytes(con.Do(ctx, "hget", GRPCHealthProbeWhiteListKey, appID))
		if e == redis.ErrNil {
			return nil
		}

		if e != nil {
			return errors.Wrap(errcode.RedisError, e.Error())
		}

		e = json.Unmarshal(value, &cfg)
		if e != nil {
			return errors.Wrapf(errcode.InternalError, "invalid GRPC health probe config: %s", value)
		}

		return nil
	})

	return cfg, err
}
