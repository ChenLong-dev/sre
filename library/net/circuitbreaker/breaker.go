package circuitbreaker

import (
	"context"
	"sync/atomic"
	"time"

	"gitlab.shanhai.int/sre/library/base/sw"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// 断路器状态
type state = int64

const (
	// 关
	closed state = 0
	// 开
	open state = 1
)

// 断路器
type Breaker struct {
	// 触发断路函数
	ShouldTrip TripFunc
	// 关闭断路器，恢复连接函数
	ShouldResume ShouldResume
	// 断路重试接口
	ShouldRetry ShouldRetry

	// 连续失败次数
	consecutiveFailuresCount int64
	// 连续成功次数
	consecutiveSuccessCount int64

	// 上次请求时间
	lastRequestTime int64
	// 上次请求时间
	lastRequestSuccessTime int64
	// 上次请求时间
	lastRequestFailTime int64

	// 计数桶滑动窗口
	counts *sw.SlidingWindow
	// 断路器状态，分别 open\closed
	breakState int64
	// 是否终止
	broken int32

	// 修复599问题
	// https://github.com/golang/go/issues/599
	_ [4]byte
}

// 断路
func (cb *Breaker) Trip() {
	atomic.StoreInt64(&cb.breakState, open)
}

// 放行(关闭断路)
func (cb *Breaker) Resume() {
	atomic.StoreInt64(&cb.breakState, closed)
}

// 重置
func (cb *Breaker) Reset() {
	atomic.StoreInt64(&cb.breakState, closed)
	atomic.StoreInt32(&cb.broken, 0)
	atomic.StoreInt64(&cb.lastRequestFailTime, 0)
	atomic.StoreInt64(&cb.lastRequestSuccessTime, 0)
	atomic.StoreInt64(&cb.lastRequestTime, 0)
	cb.ShouldRetry.Reset(cb, true)
	cb.ResetCounters()
}

// 重置计数器
func (cb *Breaker) ResetCounters() {
	atomic.StoreInt64(&cb.consecutiveFailuresCount, 0)
	atomic.StoreInt64(&cb.consecutiveSuccessCount, 0)
	cb.counts.Reset()
}

// 终止断路器
func (cb *Breaker) Break() {
	atomic.StoreInt32(&cb.broken, 1)
	cb.Trip()
}

// 单次失败
func (cb *Breaker) Fail() {
	if atomic.LoadInt32(&cb.broken) == 1 {
		return
	}

	atomic.AddInt64(&cb.consecutiveFailuresCount, 1)
	atomic.StoreInt64(&cb.consecutiveSuccessCount, 0)
	cb.counts.Fail()

	currentState := cb.currentState()
	if currentState == closed {
		if cb.ShouldTrip != nil && cb.ShouldTrip(cb) {
			cb.Trip()
		}
	} else {
		defer cb.ShouldRetry.Reset(cb, false)
	}

	now := time.Now().UnixNano()
	atomic.StoreInt64(&cb.lastRequestTime, now)
	atomic.StoreInt64(&cb.lastRequestFailTime, now)
}

// 单次成功
func (cb *Breaker) Success() {
	if atomic.LoadInt32(&cb.broken) == 1 {
		return
	}

	atomic.AddInt64(&cb.consecutiveSuccessCount, 1)
	atomic.StoreInt64(&cb.consecutiveFailuresCount, 0)
	cb.counts.Success()

	currentState := cb.currentState()
	if currentState != closed {
		if cb.ShouldResume != nil && cb.ShouldResume(cb) {
			cb.Resume()
		}
		defer cb.ShouldRetry.Reset(cb, true)
	}

	now := time.Now().UnixNano()
	atomic.StoreInt64(&cb.lastRequestTime, now)
	atomic.StoreInt64(&cb.lastRequestSuccessTime, now)
}

// 是否准备好，即是否允许真实调用
func (cb *Breaker) Ready() bool {
	if atomic.LoadInt32(&cb.broken) == 1 {
		return true
	}

	currentState := cb.currentState()
	if currentState == closed {
		return true
	} else if cb.ShouldRetry.Retry(cb) {
		return true
	}

	return false
}

// 调用函数
// 当timeout为0时，表示无超时
func (cb *Breaker) Call(ctx context.Context, f func() error, timeout time.Duration) error {
	var err error

	if !cb.Ready() {
		return errcode.BreakerOpenError
	}

	if timeout == 0 {
		err = f()
	} else {
		c := make(chan error, 1)
		go func() {
			c <- f()
			close(c)
		}()

		select {
		case e := <-c:
			err = e
		case <-time.After(timeout):
			err = errcode.BreakerTimeoutError
		}
	}

	if err != nil {
		cb.Fail()
		return err
	}
	cb.Success()

	return nil
}

// 获取状态
func (cb *Breaker) currentState() state {
	return atomic.LoadInt64(&cb.breakState)
}

// 是否断路状态
func (cb *Breaker) IsTripped() bool {
	return cb.currentState() == open
}

// 失败次数
func (cb *Breaker) FailureCount() int64 {
	return cb.counts.FailureCount()
}

// 连续失败次数
func (cb *Breaker) ConsecutiveFailureCount() int64 {
	return atomic.LoadInt64(&cb.consecutiveFailuresCount)
}

// 连续失败次数
func (cb *Breaker) ConsecutiveSuccessCount() int64 {
	return atomic.LoadInt64(&cb.consecutiveSuccessCount)
}

// 成功次数
func (cb *Breaker) SuccessCount() int64 {
	return cb.counts.SuccessCount()
}

// 错误比例
func (cb *Breaker) ErrorRate() float64 {
	return cb.counts.ErrorRate()
}

// 上一次请求时间
func (cb *Breaker) LastRequestTime() time.Time {
	t := atomic.LoadInt64(&cb.lastRequestTime)
	if t == 0 {
		return time.Time{}
	} else {
		return time.Unix(0, t)
	}
}

// 上一次请求成功时间
func (cb *Breaker) LastRequestSuccessTime() time.Time {
	t := atomic.LoadInt64(&cb.lastRequestSuccessTime)
	if t == 0 {
		return time.Time{}
	} else {
		return time.Unix(0, t)
	}
}

// 上一次请求失败时间
func (cb *Breaker) LastRequestFailTime() time.Time {
	t := atomic.LoadInt64(&cb.lastRequestFailTime)
	if t == 0 {
		return time.Time{}
	} else {
		return time.Unix(0, t)
	}
}
