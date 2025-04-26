package goroutine

import (
	"context"
	"fmt"
	"sync"
	"time"

	pkgErrors "github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/base/runtime"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

const (
	// 已启动
	stateStart State = "start"
	// 错误
	stateError State = "error"
	// 已结束
	stateEnd State = "end"
)

// 状态类型
type State string

// 协程组
type ErrGroup struct {
	// 协程waitgroup
	wg sync.WaitGroup
	// 模式
	mode Mode

	// 协程内部异常
	err error
	// 保证异常抛出不重复操作
	errOnce sync.Once

	// 当前context
	ctx context.Context
	// context取消函数
	cancel func()

	// 当前组名
	name string
	// 当前组uuid
	uuid string
	// 当前组内协程字典，负责存储每个协程相关信息
	goroutineSet sync.Map

	// 保证设置worker只执行一次
	workerOnce sync.Once
	// worker管道
	workerChan chan WorkerInfo
	// worker满时是否等待
	workerWait bool
}

// 协程组等待
func (g *ErrGroup) Wait() error {
	g.wg.Wait()
	if g.workerChan != nil {
		close(g.workerChan)
	}
	if g.mode == Cancel {
		g.CallCancel()
	}
	return g.Error()
}

// 调用cancel
func (g *ErrGroup) CallCancel() {
	if g.cancel != nil {
		g.cancel()
	}
}

// 获取error
func (g *ErrGroup) Error() error {
	return g.err
}

// 获取goroutine信息
func (g *ErrGroup) GetGoroutineInfo(name string) (SpanInfo, error) {
	value, ok := g.goroutineSet.Load(name)
	if !ok {
		return SpanInfo{}, pkgErrors.Errorf("name %s is not existed", name)
	}
	info, ok := value.(*SpanInfo)
	if !ok {
		return SpanInfo{}, pkgErrors.Errorf("span %s type error", name)
	}

	return *info, nil
}

// 获取context
func (g *ErrGroup) getContext(defaultContext context.Context) context.Context {
	if g.ctx != nil {
		return g.ctx
	}
	if defaultContext != nil {
		return defaultContext
	}

	return context.Background()
}

// 启动协程
func (g *ErrGroup) Go(ctx context.Context, name string, f func(ctx context.Context) error) {
	ctx, curSpan := g.before(g.getContext(ctx), name, uuid.NewV4().String())

	// 检查信息合法性
	beforeSpanValue, ok := g.goroutineSet.Load(name)
	if ok {
		beforeSpan, ok := beforeSpanValue.(*SpanInfo)
		if !ok {
			g.after(curSpan, stateError, pkgErrors.Errorf("span %s type error", name))
			return
		}
		if beforeSpan.State == stateStart {
			g.after(curSpan, stateError, pkgErrors.New("duplicate name"))
			return
		} else {
			g.goroutineSet.Delete(name)
		}
	}

	g.goroutineSet.Store(name, curSpan)
	g.wg.Add(1)
	if g.workerChan != nil {
		for {
			select {
			case g.workerChan <- WorkerInfo{
				f:    f,
				ctx:  ctx,
				span: curSpan,
			}:
				return
			default:
				if !g.workerWait {
					g.cleanUp(curSpan, pkgErrors.New("goroutine group exhausted"))
					return
				}
			}
		}
	}
	go g.do(ctx, curSpan, f)
}

func (g *ErrGroup) do(ctx context.Context, span *SpanInfo, f func(ctx context.Context) error) {
	var err error

	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("recover panic : %s", p)
		}
		g.cleanUp(span, err)
	}()

	err = f(ctx)
}

func (g *ErrGroup) cleanUp(span *SpanInfo, err error) {
	state := stateEnd
	if err != nil {
		state = stateError
		// 保证只要一个抛出异常即可
		g.errOnce.Do(func() {
			g.err = err
			if g.mode == Cancel {
				g.CallCancel()
			}
		})
	}

	g.after(span, state, pkgErrors.WithStack(err))
	g.wg.Done()
}

type WorkerInfo struct {
	f    func(ctx context.Context) error
	ctx  context.Context
	span *SpanInfo
}

type SpanInfo struct {
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	Source        string
	GroupName     string
	GroupID       string
	GoroutineName string
	GoroutineID   string
	Mode          Mode
	State         State
	Error         error

	Hook *hook.Hook
}

// 获取基本信息
func (g *ErrGroup) getSpanInfo(ctx context.Context, goroutineName, goroutineID string) *SpanInfo {
	msg := new(SpanInfo)
	msg.StartTime = time.Now()
	msg.Source = runtime.GetDefaultFilterCallers()
	msg.GroupName = g.name
	msg.GroupID = g.uuid
	msg.GoroutineName = goroutineName
	msg.GoroutineID = goroutineID
	msg.Mode = g.mode

	return msg
}

// 操作前注入
func (g *ErrGroup) before(ctx context.Context, goroutineName, goroutineID string) (context.Context, *SpanInfo) {
	msg := g.getSpanInfo(ctx, goroutineName, goroutineID)
	msg.State = stateStart

	hk := _manager.CreateHook(ctx).
		AddArg(render.StartTimeArgKey, msg.StartTime).
		AddArg(render.SourceArgKey, msg.Source).
		AddArg("state", msg.State).
		AddArg("mode", msg.Mode).
		AddArg("goroutine_name", msg.GoroutineName).
		AddArg("goroutine_id", msg.GoroutineID).
		AddArg("group_name", msg.GroupName).
		AddArg("group_id", msg.GroupID).
		ProcessPreHook()
	msg.Hook = hk

	return hk.Context(), msg
}

// 操作后注入
func (g *ErrGroup) after(msg *SpanInfo, state State, err error) {
	msg.EndTime = time.Now()
	msg.Duration = msg.EndTime.Sub(msg.StartTime)
	msg.Error = err
	msg.State = state

	hk := msg.Hook.AddArg(render.EndTimeArgKey, msg.EndTime).
		AddArg(render.DurationArgKey, msg.Duration).
		AddArg(render.ErrorArgKey, err).
		AddArg("state", msg.State)
	if err != nil {
		hk.AddArg("extra", errcode.GetErrorMessageMap(err))
	}
	hk.ProcessAfterHook()
	msg.Hook = hk

	return
}
