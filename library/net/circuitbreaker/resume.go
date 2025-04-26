package circuitbreaker

import (
	"time"
)

// 关闭断路器的判断函数
type ShouldResume func(*Breaker) bool

// 距上一次请求间隔
func IntervalResumeFunc(interval time.Duration) ShouldResume {
	return func(cb *Breaker) bool {
		return time.Now().Sub(cb.LastRequestTime()) >= interval
	}
}

// 连续成功次数达到阀值
func ConsecutiveResumeFunc(threshold int64) ShouldResume {
	return func(cb *Breaker) bool {
		return cb.ConsecutiveSuccessCount() >= threshold
	}
}

// 连续成功次数达到阀值或者距离上次成功间隔
func ConsecutiveOrIntervalResumeFunc(interval time.Duration, threshold int64) ShouldResume {
	return func(cb *Breaker) bool {
		return cb.ConsecutiveSuccessCount() >= threshold ||
			time.Now().Sub(cb.LastRequestTime()) >= interval
	}
}

// 成功次数比例达到指定阀值且大于最小样品值
func RateResumeFunc(rate float64, minSamples int64) ShouldResume {
	return func(cb *Breaker) bool {
		samples := cb.FailureCount() + cb.SuccessCount()
		return samples >= minSamples && cb.ErrorRate() < (1-rate)
	}
}
