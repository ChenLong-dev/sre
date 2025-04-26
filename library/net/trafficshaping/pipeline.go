package trafficshaping

import (
	"sync"
	"sync/atomic"
	"time"

	"gitlab.shanhai.int/sre/library/base/sw"
)

// 流量管道
type Pipeline struct {
	// 规则
	rules []*Rule
	// 读写锁
	rwMutex *sync.RWMutex
	// qps滑动窗口
	window *sw.SlidingWindow
	// 当前并发数量
	concurrentCount int64
	// 控制器数组
	ctls []Controller
}

// 新建管道
func NewPipeline(rules []*Rule) (*Pipeline, error) {
	for _, rule := range rules {
		if err := rule.IsValid(); err != nil {
			return nil, err
		}
	}

	ctls := []Controller{NewRejectController(), NewWaitingController()}

	return &Pipeline{
		rules:   rules,
		rwMutex: new(sync.RWMutex),
		window:  sw.NewSlidingWindow(time.Millisecond*100, 10),
		ctls:    ctls,
	}, nil
}

func (p *Pipeline) QPS() int64 {
	p.window.Slide()
	return p.window.Count()
}

func (p *Pipeline) ConcurrentCount() int64 {
	return atomic.LoadInt64(&p.concurrentCount)
}

func (p *Pipeline) beforeDo() {
	atomic.AddInt64(&p.concurrentCount, 1)
	p.window.Increase()
}

func (p *Pipeline) afterDo() {
	atomic.AddInt64(&p.concurrentCount, -1)
}

func (p *Pipeline) check(ctl Controller, rules []*Rule) error {
	for _, rule := range rules {
		r := ctl.Check(p, rule)
		if r.IsRejected() {
			return r.Error()
		} else if r.IsWaiting() {
			if wt := r.WaitingTime(); wt > 0 {
				time.Sleep(wt)
			}
			continue
		}
	}
	return nil
}

func (p *Pipeline) Do(f func()) error {
	for _, ctl := range p.ctls {
		err := p.check(ctl, p.rules)
		if err != nil {
			return err
		}
	}

	// todo:存在并发缺陷，无法严格控制
	p.beforeDo()
	defer p.afterDo()
	f()
	return nil
}
