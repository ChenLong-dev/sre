package dao

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/utils"
	"gitlab.shanhai.int/sre/library/net/redlock"
)

const (
	OwnerGithubAggregationLockKey = "owner_github_aggregation_lock_"
)

// 获取redlock分布式锁
func (d *Dao) GetOwnerGithubAggregationLock(ctx context.Context, name string) *redlock.Mutex {
	return d.RedLock.NewMutex(fmt.Sprintf("%s%s", OwnerGithubAggregationLockKey, name))
}

// 加redlock分布式锁
func (d *Dao) LockOwnerGithub(ctx context.Context, mutex *redlock.Mutex) error {
	err := mutex.Lock(ctx)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return errors.Wrapf(utils.OwnerGithubAggregationLockError, "%s", err)
	}

	return nil
}

// 解redlock分布式锁
func (d *Dao) UnlockOwnerGithub(ctx context.Context, mutex *redlock.Mutex) error {
	result := mutex.Unlock(ctx)
	if !result {
		// ==========================
		// 在首次生成error时
		// 如无需任何额外消息，应当立即使用errors.WithStack包裹
		//
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return errors.WithStack(utils.OwnerGithubAggregationUnlockError)
	}

	return nil
}
