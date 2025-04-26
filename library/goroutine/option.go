package goroutine

import (
	"context"
)

type Option interface {
	Apply(*ErrGroup)
}

type OptionFunc func(*ErrGroup)

func (f OptionFunc) Apply(eg *ErrGroup) {
	f(eg)
}

func SetNormalMode() Option {
	return OptionFunc(func(eg *ErrGroup) {
		eg.mode = Normal
	})
}

func SetCancelMode(ctx context.Context) Option {
	cancelCtx, cancel := context.WithCancel(ctx)

	return OptionFunc(func(eg *ErrGroup) {
		eg.mode = Cancel
		eg.ctx = cancelCtx
		eg.cancel = cancel
	})
}

// 设置最大协程数
// 只允许设置一次
func SetMaxWorker(max int, wait bool) Option {
	return OptionFunc(func(eg *ErrGroup) {
		eg.workerOnce.Do(func() {
			eg.workerChan = make(chan WorkerInfo, max)
			eg.workerWait = wait
			for i := 0; i < max; i++ {
				go func() {
					for info := range eg.workerChan {
						eg.do(info.ctx, info.span, info.f)
					}
				}()
			}
		})
	})
}
