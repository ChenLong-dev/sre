package dao

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"

	_redis "gitlab.shanhai.int/sre/library/database/redis"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/entity"
)

const multiClusterSupportedProjectCacheKeyPrefix = "multi_cluster_supported_projects::"

// SetMultiClusterSupportForProject 为项目添加多集群支持
func (d *Dao) SetMultiClusterSupportForProject(ctx context.Context, projectID string) error {
	stgKey := formatMultiClusterSupportedProjectCacheKey(entity.AppEnvStg)
	prdKey := formatMultiClusterSupportedProjectCacheKey(entity.AppEnvPrd)
	return d.Redis.WrapDo(func(con *_redis.Conn) error {
		err := con.Send(ctx, "hset", stgKey, projectID, "")
		if err != nil {
			return errors.Wrapf(errcode.RedisError,
				"SetMultiClusterSupportForProject(%s) in env(%s) failed: %s", projectID, entity.AppEnvStg, err)
		}

		err = con.Send(ctx, "hset", prdKey, projectID, "")
		if err != nil {
			return errors.Wrapf(errcode.RedisError,
				"SetMultiClusterSupportForProject(%s) in env(%s) failed: %s", projectID, entity.AppEnvPrd, err)
		}

		err = con.Flush(ctx)
		if err != nil {
			return errors.Wrapf(errcode.RedisError,
				"SetMultiClusterSupportForProject(%s) flush failed: %s", projectID, err)
		}

		return nil
	})
}

// CheckMultiClusterSupportedProject 校验项目是否已支持多集群
func (d *Dao) CheckMultiClusterSupportedProject(ctx context.Context, envName entity.AppEnvName, projectID string) (bool, error) {
	con := d.Redis.Get()
	defer con.Close()

	_, err := redis.String(con.Do(ctx, "hget", formatMultiClusterSupportedProjectCacheKey(envName), projectID))

	if err == nil {
		return true, nil
	}

	if err == redis.ErrNil {
		return false, nil
	}

	return false, err
}

func formatMultiClusterSupportedProjectCacheKey(envName entity.AppEnvName) string {
	return multiClusterSupportedProjectCacheKeyPrefix + string(envName)
}
