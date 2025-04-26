package circuitbreaker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func TestBreaker_Trip(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewBreaker()
		assert.Equal(t, false, cb.IsTripped())

		cb.Success()
		assert.Equal(t, false, cb.IsTripped())

		cb.Trip()
		assert.Equal(t, true, cb.IsTripped())
	})
}

func TestBreaker_Resume(t *testing.T) {
	t.Run("resume break", func(t *testing.T) {
		cb := NewBreaker()
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, false, cb.IsTripped())

		cb.Success()
		assert.Equal(t, false, cb.IsTripped())

		cb.Trip()
		assert.Equal(t, true, cb.IsTripped())

		cb.Resume()
		assert.Equal(t, false, cb.IsTripped())
	})
}

func TestBreakerCounts(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewBreaker()

		cb.Fail()
		assert.Equal(t, int64(1), cb.FailureCount())

		cb.Fail()
		assert.Equal(t, int64(2), cb.ConsecutiveFailureCount())

		cb.Success()
		assert.Equal(t, int64(1), cb.SuccessCount())
		assert.Equal(t, int64(0), cb.ConsecutiveFailureCount())

		cb.Reset()
		assert.Equal(t, int64(0), cb.SuccessCount())
		assert.Equal(t, int64(0), cb.ConsecutiveFailureCount())
		assert.Equal(t, int64(0), cb.FailureCount())
	})
}

func TestBreaker_Ready(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewBreakerWithOptions(&Options{
			ShouldRetry: NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
			ShouldTrip:  ConsecutiveTripFunc(1),
		})
		assert.Equal(t, true, cb.Ready())

		cb.Success()
		assert.Equal(t, true, cb.Ready())

		cb.Fail()
		assert.Equal(t, false, cb.Ready())

		// 等待静默期后，进入半开状态
		time.Sleep(time.Millisecond * 500)
		assert.Equal(t, true, cb.Ready())
		// 半开状态下只能有一个请求通过
		assert.Equal(t, false, cb.Ready())

		// 断路器关闭后
		cb.Break()
		assert.Equal(t, true, cb.Ready())
	})
}

func TestBreaker_Fail(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewBreakerWithOptions(&Options{
			ShouldRetry: NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
			ShouldTrip:  ConsecutiveTripFunc(2),
		})
		assert.Equal(t, true, cb.Ready())

		cb.Fail()
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, int64(1), cb.ConsecutiveFailureCount())
		assert.Equal(t, int64(0), cb.ConsecutiveSuccessCount())

		// 失败次数达到阈值，进入到open
		cb.Fail()
		assert.Equal(t, false, cb.Ready())
		assert.Equal(t, int64(2), cb.ConsecutiveFailureCount())
		assert.Equal(t, int64(0), cb.ConsecutiveSuccessCount())

		// 等待静默期后，进入半开状态
		time.Sleep(time.Millisecond * 500)
		assert.Equal(t, true, cb.Ready())

		// 半开状态下失败再次进入open
		cb.Fail()
		assert.Equal(t, false, cb.Ready())
		assert.Equal(t, int64(3), cb.ConsecutiveFailureCount())
		assert.Equal(t, int64(0), cb.ConsecutiveSuccessCount())
	})

	t.Run("breaker broken", func(t *testing.T) {
		cb := NewBreakerWithOptions(&Options{
			ShouldRetry: NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
			ShouldTrip:  ConsecutiveTripFunc(2),
		})

		cb.Break()
		assert.Equal(t, true, cb.Ready())

		cb.Fail()
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, int64(0), cb.ConsecutiveFailureCount())
		assert.Equal(t, int64(0), cb.ConsecutiveSuccessCount())

		cb.Fail()
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, int64(0), cb.ConsecutiveFailureCount())
		assert.Equal(t, int64(0), cb.ConsecutiveSuccessCount())
	})
}

func TestBreaker_Break(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewThresholdBreaker(1)
		assert.Equal(t, true, cb.Ready())

		cb.Fail()
		assert.Equal(t, false, cb.Ready())

		cb.Break()
		assert.Equal(t, true, cb.Ready())

		cb.Fail()
		assert.Equal(t, true, cb.Ready())
	})
}

func TestBreaker_Call(t *testing.T) {
	t.Run("normal error", func(t *testing.T) {
		f := func() error {
			return errors.New("this is a test")
		}

		cb := NewThresholdBreaker(2)
		ctx := context.Background()

		err := cb.Call(ctx, f, 0)
		assert.NotNil(t, err)
		assert.NotEqual(t, errcode.BreakerOpenError, err)
		assert.Equal(t, false, cb.IsTripped())

		// 第二次失败，触发断路
		err = cb.Call(ctx, f, 0)
		assert.NotNil(t, err)
		assert.NotEqual(t, errcode.BreakerOpenError, err)
		assert.Equal(t, true, cb.IsTripped())

		err = cb.Call(ctx, f, 0)
		assert.Equal(t, errcode.BreakerOpenError, err)
	})
}

func TestBreaker_LastRequestTime(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewBreaker()
		assert.Empty(t, cb.LastRequestTime())
		assert.Empty(t, cb.LastRequestFailTime())
		assert.Empty(t, cb.LastRequestSuccessTime())

		cb.Fail()
		requestTime := cb.LastRequestTime()
		assert.NotEmpty(t, cb.LastRequestTime())
		assert.NotEmpty(t, cb.LastRequestFailTime())

		cb.Success()
		assert.NotEqual(t, requestTime, cb.LastRequestTime())
		assert.NotEmpty(t, cb.LastRequestSuccessTime())

		cb.Reset()
		assert.Empty(t, cb.LastRequestTime())
		assert.Empty(t, cb.LastRequestFailTime())
		assert.Empty(t, cb.LastRequestSuccessTime())
	})
}

func BenchmarkBreaker_Call(b *testing.B) {
	b.Run("normal", func(b *testing.B) {
		cb := NewBreaker()
		var index int64

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = cb.Call(context.Background(), func() error {
					atomic.AddInt64(&index, 1)
					if atomic.LoadInt64(&index)%10 > 3 {
						return errors.New("this is a error")
					}
					return nil
				}, 0)
			}
		})
	})
}
