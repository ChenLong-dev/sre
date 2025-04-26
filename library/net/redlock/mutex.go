package redlock

import (
	"context"
	"time"

	"github.com/go-redsync/redsync"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/base/runtime"
)

const (
	StateSuccess = "success"
	StateFail    = "fail"
)

// 分布式锁
type Mutex struct {
	// 分布式锁
	*redsync.Mutex
	// 钩子管理器
	manager *hook.Manager
	// key名
	name string
}

// 锁
func (m *Mutex) Lock(ctx context.Context) error {
	ctx, hk := m.before(ctx, "LOCK")
	e := m.Mutex.Lock()
	m.after(hk, e)

	return e
}

// 解锁
func (m *Mutex) Unlock(ctx context.Context) bool {
	ctx, hk := m.before(ctx, "UNLOCK")
	result := m.Mutex.Unlock()
	var err error
	if !result {
		err = errors.New("Unlock Fail")
	}
	m.after(hk, err)

	return result
}

// 延长锁的时间
func (m *Mutex) Extend(ctx context.Context) bool {
	ctx, hk := m.before(ctx, "EXTEND")
	result := m.Mutex.Extend()
	var err error
	if !result {
		err = errors.New("Extend Fail")
	}
	m.after(hk, err)

	return result
}

// 操作前注入
func (m *Mutex) before(ctx context.Context, commandName string) (context.Context, *hook.Hook) {
	hk := m.manager.CreateHook(ctx).
		AddArg(render.StartTimeArgKey, time.Now()).
		AddArg(render.SourceArgKey, runtime.GetDefaultFilterCallers()).
		AddArg("mutex_name", m.name).
		AddArg("command_name", commandName).
		ProcessPreHook()

	return hk.Context(), hk
}

// 操作后注入
func (m *Mutex) after(hk *hook.Hook, err error) {
	var state string
	if err != nil {
		state = StateFail
	} else {
		state = StateSuccess
	}
	endTime := time.Now()
	duration := endTime.Sub(hk.Arg(render.StartTimeArgKey).(time.Time))

	hk = hk.AddArg(render.EndTimeArgKey, endTime).
		AddArg(render.DurationArgKey, duration).
		AddArg("state", state).
		AddArg(render.ErrorArgKey, err).
		ProcessAfterHook()
}
