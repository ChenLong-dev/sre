package circuitbreaker

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestThresholdTripFunc(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewThresholdBreaker(2)
		assert.Equal(t, false, cb.IsTripped())

		cb.Fail()
		assert.Equal(t, false, cb.IsTripped())

		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())
	})
}

func TestConsecutiveTripFunc(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewConsecutiveBreaker(3)
		assert.Equal(t, false, cb.IsTripped())

		cb.Fail()
		cb.Success()
		cb.Fail()
		cb.Fail()
		// 未连续失败3次，不断路
		assert.Equal(t, false, cb.IsTripped())

		// 连续失败3次，断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())
	})
}

func TestRateTripFunc(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewBreakerWithOptions(&Options{
			ShouldRetry: NewSingleRequestRetry(NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500)),
			ShouldTrip:  RateTripFunc(0.5, 5),
		})

		cb.Success()
		cb.Success()
		cb.Fail()
		cb.Fail()
		// 采样数未大于5，不断路
		assert.Equal(t, false, cb.IsTripped())
		assert.Equal(t, 0.5, cb.ErrorRate())
		cb.Fail()
		// 采样数等于5，断路
		assert.Equal(t, true, cb.IsTripped())
		// 断路后，不重置计数
		assert.Equal(t, 0.6, cb.ErrorRate())
	})
}
