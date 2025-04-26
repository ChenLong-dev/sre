package circuitbreaker

import (
	"github.com/cenkalti/backoff"
	"math/rand"
	"sync/atomic"
	"time"
)

// 断路重试接口
type ShouldRetry interface {
	// 判断是否应重试
	Retry(cb *Breaker) bool
	// 重试后重置函数
	Reset(cb *Breaker, isSuccess bool)
}

// 新建指数退避
func NewExponentialBackOff(minInterval, maxInterval time.Duration) backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = minInterval
	b.MaxInterval = maxInterval
	b.MaxElapsedTime = 0
	b.Reset()
	return b
}

// 单一请求重试结构体
// 存在请求失败退避，尝试时只会放行1个请求
type SingleRequestRetry struct {
	// 退避相关函数
	backOff backoff.BackOff
	// 下次退避时间
	nextBackOff time.Duration
	// 是否正在请求中（控制只能发起一个请求）
	isRequesting int64
}

// 新建单一请求重试结构体
func NewSingleRequestRetry(backOff backoff.BackOff) ShouldRetry {
	return &SingleRequestRetry{
		backOff:     backOff,
		nextBackOff: backOff.NextBackOff(),
	}
}

func (r *SingleRequestRetry) Retry(cb *Breaker) bool {
	if time.Now().Sub(cb.LastRequestFailTime()) >= r.nextBackOff &&
		atomic.CompareAndSwapInt64(&r.isRequesting, 0, 1) {
		r.nextBackOff = r.backOff.NextBackOff()
		return true
	}
	return false
}

func (r *SingleRequestRetry) Reset(cb *Breaker, isSuccess bool) {
	if isSuccess {
		r.backOff.Reset()
		r.nextBackOff = r.backOff.NextBackOff()
	}
	atomic.StoreInt64(&r.isRequesting, 0)
}

// 阀值重试结构体
// 存在请求失败退避，尝试时会放行指定数量请求
type ThresholdRequestRetry struct {
	// 退避相关函数
	backOff backoff.BackOff
	// 下次退避时间
	nextBackOff time.Duration
	// 当前请求数量
	cur int64
	// 阀值
	threshold int64
}

// 新建阀值重试结构体
func NewThresholdRequestRetry(threshold int64, backOff backoff.BackOff) ShouldRetry {
	return &ThresholdRequestRetry{
		threshold:   threshold,
		cur:         0,
		backOff:     backOff,
		nextBackOff: backOff.NextBackOff(),
	}
}

func (r *ThresholdRequestRetry) Retry(cb *Breaker) bool {
	if time.Now().Sub(cb.LastRequestFailTime()) >= r.nextBackOff &&
		atomic.LoadInt64(&r.cur) < r.threshold {
		r.nextBackOff = r.backOff.NextBackOff()
		atomic.AddInt64(&r.cur, 1)
		return true
	}
	return false
}

func (r *ThresholdRequestRetry) Reset(cb *Breaker, isSuccess bool) {
	if isSuccess {
		r.backOff.Reset()
		r.nextBackOff = r.backOff.NextBackOff()
	}
	atomic.AddInt64(&r.cur, -1)
}

// 比例重试结构体
// 存在请求失败退避，尝试时会放行指定数量请求
type RateRequestRetry struct {
	// 退避相关函数
	backOff backoff.BackOff
	// 下次退避时间
	nextBackOff time.Duration
	// 放行比例
	rate float64
	// 随机种子
	rd *rand.Rand
}

// 新建比例重试结构体
func NewRateRequestRetry(rate float64, backOff backoff.BackOff) ShouldRetry {
	return &RateRequestRetry{
		rate:        rate,
		rd:          rand.New(rand.NewSource(time.Now().Unix())),
		backOff:     backOff,
		nextBackOff: backOff.NextBackOff(),
	}
}

func (r *RateRequestRetry) Retry(cb *Breaker) bool {
	if time.Now().Sub(cb.LastRequestFailTime()) >= r.nextBackOff &&
		r.rd.Float64() <= r.rate {
		r.nextBackOff = r.backOff.NextBackOff()
		return true
	}
	return false
}

func (r *RateRequestRetry) Reset(cb *Breaker, isSuccess bool) {
	if isSuccess {
		r.backOff.Reset()
		r.nextBackOff = r.backOff.NextBackOff()
	}
}

// 失败退避重试结构体
// 存在请求失败退避，尝试时会放行所有请求
type FailBackoffRequestRetry struct {
	// 退避相关函数
	backOff backoff.BackOff
	// 下次退避时间
	nextBackOff time.Duration
}

// 新建失败退避重试结构体
func NewFailBackoffRequestRetry(backOff backoff.BackOff) ShouldRetry {
	return &FailBackoffRequestRetry{
		backOff:     backOff,
		nextBackOff: backOff.NextBackOff(),
	}
}

func (r *FailBackoffRequestRetry) Retry(cb *Breaker) bool {
	if time.Now().Sub(cb.LastRequestFailTime()) >= r.nextBackOff {
		r.nextBackOff = r.backOff.NextBackOff()
		return true
	}
	return false
}

func (r *FailBackoffRequestRetry) Reset(cb *Breaker, isSuccess bool) {
	if isSuccess {
		r.backOff.Reset()
		r.nextBackOff = r.backOff.NextBackOff()
	}
}
