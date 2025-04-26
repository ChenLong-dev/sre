package trafficshaping

import (
	"github.com/pkg/errors"
	"time"
)

// 规则类型
type RuleType int

const (
	// 入口qps限制
	QPS RuleType = iota
	// 并发限制
	Concurrency
)

// 控制行为
type ControlBehavior int

const (
	// 直接拒绝
	Reject ControlBehavior = iota
	// 排队等待
	// 此种控制方式采用漏桶算法，与规则类型本身无关
	Waiting
)

// 规则
type Rule struct {
	// 规则类型
	Type RuleType
	// 规则控制行为
	ControlBehavior ControlBehavior
	// 最大等待时间，为0表示不等待
	// 要求ControlBehavior为Waiting
	MaxWaitingTime time.Duration
	// 限制个数，为0表示不限制
	// 当类型为QPS模式时，该参数为qps限制个数
	// 当类型为并发模式时，该参数为最大并发数量
	Limit float64
}

// 是否合法
func (r *Rule) IsValid() error {
	if r.MaxWaitingTime != 0 && r.ControlBehavior != Waiting {
		return errors.Errorf("MaxWaitingTime isn't empty, but type isn't Waiting")
	}
	return nil
}
