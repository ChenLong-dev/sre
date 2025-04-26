package dao

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func (d *Dao) getJenkinsBuildCacheKey(projectName string) string {
	return fmt.Sprintf("jenkins:job:%s", projectName)
}

func (d *Dao) GetProjectBuildAllValueFromCache(ctx context.Context, projectName string) ([]string, error) {
	con := d.Redis.Get()
	defer con.Close()

	vals, err := redis.Strings(
		con.Do(ctx, "hvals", d.getJenkinsBuildCacheKey(projectName)))
	if err == redis.ErrNil {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrapf(errcode.RedisError, err.Error())
	}

	return vals, nil
}
