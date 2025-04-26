package trafficshaping

import (
	"github.com/pkg/errors"
	"time"
)

var (
	// 拒绝错误
	RejectError = errors.New("traffic shaping reject")
	// 需要等待错误
	NeedToWaitError = errors.New("current request need to wait")
)

// 控制结果
type Result struct {
	// 错误
	err error
	// 等待时间
	waitingTime time.Duration
}

// 错误
func (r *Result) Error() error {
	return r.err
}

// 是否被拒绝
func (r *Result) IsRejected() bool {
	return r.err == RejectError
}

// 是否在等待
func (r *Result) IsWaiting() bool {
	return r.err == NeedToWaitError
}

// 等待时间
func (r *Result) WaitingTime() time.Duration {
	return r.waitingTime
}

// 默认结果
func DefaultResult() *Result {
	return &Result{}
}

// 拒绝结果
func RejectResult() *Result {
	return &Result{
		err: RejectError,
	}
}

// 等待结果
func WaitingResult(waitingTime time.Duration) *Result {
	return &Result{
		err:         NeedToWaitError,
		waitingTime: waitingTime,
	}
}
