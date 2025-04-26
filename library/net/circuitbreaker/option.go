package circuitbreaker

import (
	"time"

	"gitlab.shanhai.int/sre/library/base/sw"
)

// 配置
type Options struct {
	// 触发断路函数
	ShouldTrip TripFunc
	// 关闭断路函数
	ShouldResume ShouldResume
	// 断路重试接口
	ShouldRetry ShouldRetry
	// 滑动窗口滑动间隔
	WindowSlideInterval time.Duration
	// 滑动窗口覆盖的桶数量
	WindowBucketCount int
}

// 通过配置新建断路器
func NewBreakerWithOptions(options *Options) *Breaker {
	if options == nil {
		options = &Options{}
	}

	if options.ShouldResume == nil {
		options.ShouldResume = RateResumeFunc(0.4, 10)
	}

	if options.ShouldTrip == nil {
		options.ShouldTrip = RateTripFunc(0.5, 10)
	}

	if options.ShouldRetry == nil {
		options.ShouldRetry = NewThresholdRequestRetry(
			5,
			NewExponentialBackOff(time.Millisecond*100, time.Second*5),
		)
	}

	if options.WindowSlideInterval == 0 {
		options.WindowSlideInterval = time.Second
	}

	if options.WindowBucketCount == 0 {
		options.WindowBucketCount = 10
	}

	return &Breaker{
		ShouldTrip:               options.ShouldTrip,
		ShouldResume:             options.ShouldResume,
		ShouldRetry:              options.ShouldRetry,
		consecutiveFailuresCount: 0,
		consecutiveSuccessCount:  0,
		lastRequestTime:          0,
		lastRequestSuccessTime:   0,
		lastRequestFailTime:      0,
		counts:                   sw.NewSlidingWindow(options.WindowSlideInterval, options.WindowBucketCount),
	}
}

// 新建断路器
func NewBreaker() *Breaker {
	return NewBreakerWithOptions(nil)
}

// 新建阀值断路器
func NewThresholdBreaker(threshold int64) *Breaker {
	return NewBreakerWithOptions(&Options{
		ShouldTrip: ThresholdTripFunc(threshold),
	})
}

// 新建连续断路器
func NewConsecutiveBreaker(threshold int64) *Breaker {
	return NewBreakerWithOptions(&Options{
		ShouldTrip: ConsecutiveTripFunc(threshold),
	})
}

// 新建比例断路器
func NewRateBreaker(rate float64, minSamples int64) *Breaker {
	return NewBreakerWithOptions(&Options{
		ShouldTrip: RateTripFunc(rate, minSamples),
	})
}
