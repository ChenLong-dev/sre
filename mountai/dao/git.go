package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/resp"
)

const projectMembersTTL = 60

// GetGitlabProjectActiveMembersFromCache 从缓存中获取项目下活跃的成员信息
func (d *Dao) GetGitlabProjectActiveMembersFromCache(ctx context.Context, projectID string) ([]*resp.GitProjectMemberResp, error) {
	con := d.Redis.Get()
	defer con.Close()

	key := d.getGitlabProjectMembersCacheKey(ctx, projectID)

	value, err := redis.Bytes(con.Do(ctx, "get", key))
	if err == redis.ErrNil {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrap(errcode.RedisError, err.Error())
	}

	res := make([]*resp.GitProjectMemberResp, 0)
	err = json.Unmarshal(value, &res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// SetGitlabProjectActiveMembersToCache 缓存项目下活跃的成员信息
func (d *Dao) SetGitlabProjectActiveMembersToCache(ctx context.Context, projectID string, members []*resp.GitProjectMemberResp) error {
	con := d.Redis.Get()
	defer con.Close()

	key := d.getGitlabProjectMembersCacheKey(ctx, projectID)

	value, err := json.Marshal(members)

	if err != nil {
		return errors.Wrap(errcode.InternalError, err.Error())
	}

	err = con.Send(ctx, "setex", redis.Args{}.Add(key).Add(projectMembersTTL).Add(value)...)

	if err != nil {
		return errors.Wrap(errcode.RedisError, err.Error())
	}

	return nil
}

func (d *Dao) getGitlabProjectMembersCacheKey(_ context.Context, projectID string) string {
	return fmt.Sprintf("project:%s:members", projectID)
}
