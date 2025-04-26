package circuitbreaker

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewIntervalResumeBreaker(t *testing.T) {
	t.Run("interval resume", func(t *testing.T) {
		var (
			interval = time.Second * 2
		)
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: IntervalResumeFunc(interval),
			ShouldRetry:  NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())

		// 等待静默期后，进入半开状态
		time.Sleep(time.Millisecond * 500)
		assert.Equal(t, true, cb.Ready())

		// 间隔未达到阈值，断路器未关闭
		time.Sleep(time.Second)
		cb.Success()
		assert.Equal(t, true, cb.IsTripped())

		// 达到阈值，断路器关闭
		time.Sleep(interval)
		cb.Success()
		assert.Equal(t, false, cb.IsTripped())
	})
}

func TestNewConsecutiveResumeBreaker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var (
			count int64 = 3
		)
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: ConsecutiveResumeFunc(count),
			ShouldRetry:  NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())
		assert.Equal(t, int64(0), cb.ConsecutiveSuccessCount())

		// 等待静默期后，进入半开状态
		time.Sleep(time.Millisecond * 500)
		assert.Equal(t, true, cb.Ready())

		// 连续成功数未达到阈值
		cb.Success()
		cb.Success()
		assert.Equal(t, int64(2), cb.ConsecutiveSuccessCount())
		assert.Equal(t, true, cb.IsTripped())

		// 连续成功数达到阈值
		cb.Success()
		assert.Equal(t, int64(3), cb.ConsecutiveSuccessCount())
		assert.Equal(t, false, cb.IsTripped())
	})

	t.Run("fail", func(t *testing.T) {
		var (
			count int64 = 3
		)
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: ConsecutiveResumeFunc(count),
			ShouldRetry:  NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())
		assert.Equal(t, int64(0), cb.ConsecutiveSuccessCount())

		// 等待静默期后，进入半开状态
		time.Sleep(time.Millisecond * 500)
		assert.Equal(t, true, cb.Ready())

		// 连续成功数未达到阈值
		cb.Success()
		cb.Success()
		assert.Equal(t, int64(2), cb.ConsecutiveSuccessCount())
		assert.Equal(t, true, cb.IsTripped())

		// 失败后，连续成功数从0开始算，未达到阈值
		cb.Fail()
		cb.Success()
		assert.Equal(t, int64(1), cb.ConsecutiveSuccessCount())
		assert.Equal(t, true, cb.IsTripped())

		// 等待静默期后，进入半开状态
		time.Sleep(time.Millisecond * 500)
		assert.Equal(t, true, cb.Ready())

		// 连续成功数到阈值，断路器关闭
		cb.Success()
		cb.Success()
		assert.Equal(t, int64(3), cb.ConsecutiveSuccessCount())
		assert.Equal(t, false, cb.IsTripped())
	})
}

func TestNewConsecutiveOrIntervalResumeBreaker(t *testing.T) {
	t.Run("only consecutive count trigger resume", func(t *testing.T) {
		var (
			count    int64 = 3
			interval       = time.Second * 2
		)
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: ConsecutiveOrIntervalResumeFunc(interval, count),
			ShouldRetry:  NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())

		// 连续成功数和时间未达到阈值,断路器未关闭
		cb.Success()
		cb.Success()
		assert.Equal(t, true, cb.IsTripped())

		// 时间阈值未达到
		time.Sleep(time.Second)
		assert.Equal(t, true, cb.IsTripped())

		// 连续成功数到阈值，断路器关闭
		cb.Success()
		assert.Equal(t, false, cb.IsTripped())
	})

	t.Run("only interval trigger resume", func(t *testing.T) {
		var (
			count    int64 = 3
			interval       = time.Second * 2
		)
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: ConsecutiveOrIntervalResumeFunc(interval, count),
			ShouldRetry:  NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())

		// 连续成功数和时间都未达到阈值,断路器未关闭
		cb.Success()
		time.Sleep(time.Second)
		assert.Equal(t, true, cb.IsTripped())

		// 间隔达到阈值，断路器关闭
		time.Sleep(interval + 1)
		cb.Success()
		assert.Equal(t, false, cb.IsTripped())
	})

	t.Run("interval and consecutive trigger resume at same time", func(t *testing.T) {
		var (
			count    int64 = 3
			interval       = time.Second * 2
		)
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: ConsecutiveOrIntervalResumeFunc(interval, count),
			ShouldRetry:  NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())

		// 连续成功数和时间未达到阈值,断路器未关闭
		cb.Success()
		cb.Success()
		time.Sleep(time.Second)
		assert.Equal(t, true, cb.IsTripped())

		// 连续成功数和时间都达到阈值，断路器关闭
		time.Sleep(interval + 1)
		cb.Success()
		assert.Equal(t, false, cb.IsTripped())
	})
}

func TestRateResumeFunc(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: RateResumeFunc(0.4, 3),
			ShouldRetry:  NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())

		// 请求总数未达到阈值
		cb.Success()
		assert.Equal(t, true, cb.IsTripped())

		// 请求数达到阈值，并且满足比例
		cb.Success()
		assert.Equal(t, false, cb.IsTripped())
	})

	t.Run("fail", func(t *testing.T) {
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: RateResumeFunc(0.4, 3),
			ShouldRetry:  NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())

		// 请求总数未达到阈值
		cb.Success()
		assert.Equal(t, true, cb.IsTripped())

		// 请求数达到阈值，但未满足比例
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())

		// 满足比例
		cb.Success()
		assert.Equal(t, false, cb.IsTripped())
	})
}
