package circuitbreaker

import (
	"sync"
)

// 断路器组
type BreakerGroup struct {
	// 断路器字典
	BreakerMap map[string]*Breaker
	// 组锁
	groupRWMutex sync.RWMutex
}

// 添加断路器
func (p *BreakerGroup) Add(name string, cb *Breaker) {
	p.groupRWMutex.Lock()
	p.BreakerMap[name] = cb
	p.groupRWMutex.Unlock()
}

// 获取断路器
func (p *BreakerGroup) Get(name string) *Breaker {
	p.groupRWMutex.RLock()
	cb, ok := p.BreakerMap[name]
	p.groupRWMutex.RUnlock()

	if !ok {
		return nil
	}

	return cb
}

// 新建断路器组
func NewBreakerGroup() *BreakerGroup {
	return &BreakerGroup{
		BreakerMap: make(map[string]*Breaker),
	}
}
