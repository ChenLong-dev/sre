package circuitbreaker

// 触发断路的判断函数
type TripFunc func(*Breaker) bool

// 时间间隔内的失败次数达到阀值
func ThresholdTripFunc(threshold int64) TripFunc {
	return func(cb *Breaker) bool {
		return cb.FailureCount() == threshold
	}
}

// 连续失败次数达到阀值
func ConsecutiveTripFunc(threshold int64) TripFunc {
	return func(cb *Breaker) bool {
		return cb.ConsecutiveFailureCount() == threshold
	}
}

// 失败次数比例达到指定阀值且大于最小样品值
func RateTripFunc(rate float64, minSamples int64) TripFunc {
	return func(cb *Breaker) bool {
		samples := cb.FailureCount() + cb.SuccessCount()
		return samples >= minSamples && cb.ErrorRate() >= rate
	}
}
