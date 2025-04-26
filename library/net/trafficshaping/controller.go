package trafficshaping

import (
	"math"
	"sync/atomic"
	"time"
)

// 控制器接口
type Controller interface {
	// 校验
	Check(p *Pipeline, r *Rule) *Result
}

// 拒绝控制器
type RejectController struct {
}

func (c *RejectController) Check(p *Pipeline, r *Rule) *Result {
	if r.ControlBehavior != Reject {
		return DefaultResult()
	} else if p == nil {
		return DefaultResult()
	}

	var curCount int64
	if r.Type == QPS {
		curCount = p.QPS()
	} else if r.Type == Concurrency {
		curCount = p.ConcurrentCount()
	}

	if float64(curCount+1) > r.Limit {
		return RejectResult()
	}
	return DefaultResult()
}

func NewRejectController() *RejectController {
	return &RejectController{}
}

// 等待控制器
type WaitingController struct {
	lastPassedTime int64
}

func (c *WaitingController) Check(p *Pipeline, r *Rule) *Result {
	if r.ControlBehavior != Waiting {
		return DefaultResult()
	} else if p == nil {
		return DefaultResult()
	}

	curTime := time.Now().UnixNano()
	interval := int64(math.Ceil(float64(time.Second) / r.Limit))

	expectedTime := atomic.LoadInt64(&c.lastPassedTime) + interval
	if expectedTime <= curTime {
		atomic.StoreInt64(&c.lastPassedTime, curTime)
		return DefaultResult()
	}

	waitingTime := atomic.LoadInt64(&c.lastPassedTime) + interval - curTime
	if waitingTime > r.MaxWaitingTime.Nanoseconds() {
		return RejectResult()
	}

	waitingTime = atomic.AddInt64(&c.lastPassedTime, interval) - curTime
	if waitingTime > r.MaxWaitingTime.Nanoseconds() {
		atomic.AddInt64(&c.lastPassedTime, -interval)
		return RejectResult()
	}

	if waitingTime > 0 {
		return WaitingResult(time.Duration(waitingTime))
	} else {
		return DefaultResult()
	}
}

func NewWaitingController() *WaitingController {
	return &WaitingController{}
}
