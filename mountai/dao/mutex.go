package dao

import (
	"context"
	"fmt"

	"gitlab.shanhai.int/sre/library/net/redlock"
)

const (
	TaskStatusWorkerLockKey = "task_status_worker:"
)

// 获取redlock分布式锁
func (d *Dao) GetTaskStatusWorkerLock(ctx context.Context, id string) *redlock.Mutex {
	return d.Redlock.NewMutex(fmt.Sprintf("%s%s", TaskStatusWorkerLockKey, id))
}
