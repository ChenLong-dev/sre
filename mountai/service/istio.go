package service

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/slice"
	_redis "gitlab.shanhai.int/sre/library/database/redis"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/config"
	"rulai/models/entity"
	"rulai/models/resp"
)

// GetApplicationIstioState returns istio switch state
func (s *Service) GetApplicationIstioState(ctx context.Context, env entity.AppEnvName, cluster entity.ClusterName,
	app *resp.AppDetailResp) bool {
	// 目前仅 zeus 集群支持 istio 发布
	if cluster != entity.DefaultClusterName || string(env) == "" || app.ID == "" {
		log.Errorc(ctx,
			"GetApplicationIstioStateError: [%s] cluster:%s, env:%s, service:%s, appName: %s, empty env"+
				" or empty appID or unsupported cluster",
			app.ID, cluster, env, app.ServiceName, app.Name)
		return false
	}

	// 判断 app 在对应的环境中是否支持 istio 部署
	istioSwitchOn := false

	switch env {
	// 测试环境只判断 ams 配置选项
	case entity.AppEnvStg:
		istioSwitchOn = slice.StrSliceContains(config.Conf.IstioOnEnv, string(env))
	// 其他的环境参考 redis 中的配置
	default:
		key := getApplicationIstioStateKey(env)
		con := s.dao.Redis.Get()
		defer func(con *_redis.Conn) {
			err := con.Close()
			if err != nil {
				log.Errorc(ctx, "%s", "redis close failed, ignored.")
			}
		}(con)

		idExist, err := redis.Bool(con.Do(ctx, "hget", key, app.ID))
		if err != nil {
			log.Errorc(ctx, "GetApplicationIstioStateError: %s", errors.Wrap(errcode.RedisError, err.Error()))
			return false
		}

		istioSwitchOn = idExist
	}

	return istioSwitchOn && app.EnableIstio
}

func getApplicationIstioStateKey(env entity.AppEnvName) string {
	return fmt.Sprintf("ams:config:istio:switchhash:%s", string(env))
}
