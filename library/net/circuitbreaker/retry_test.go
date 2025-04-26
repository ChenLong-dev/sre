package circuitbreaker

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewSingleRequestRetry(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: ConsecutiveResumeFunc(3),
			ShouldRetry: NewSingleRequestRetry(
				NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500),
			),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())

		// 等待静默期后，进入半开状态
		time.Sleep(time.Millisecond * 500)
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, true, cb.IsTripped())

		// 只会允许一个请求
		assert.Equal(t, false, cb.Ready())
		assert.Equal(t, true, cb.IsTripped())

		cb.Success()
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, true, cb.IsTripped())
	})
}

func TestNewFailBackoffRequestRetry(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: ConsecutiveResumeFunc(3),
			ShouldRetry: NewFailBackoffRequestRetry(
				NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500),
			),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())

		// 等待静默期后，进入半开状态
		time.Sleep(time.Millisecond * 500)
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, true, cb.IsTripped())

		// 允许多个请求
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, true, cb.IsTripped())

		// 失败后退避
		cb.Fail()
		assert.Equal(t, false, cb.Ready())
		assert.Equal(t, true, cb.IsTripped())

		// 退避后，进入半开状态
		time.Sleep(time.Millisecond * 500)
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, true, cb.IsTripped())
	})
}

func TestNewThresholdRequestRetry(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: ConsecutiveResumeFunc(3),
			ShouldRetry: NewThresholdRequestRetry(
				3,
				NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500),
			),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())

		// 等待静默期后，进入半开状态
		time.Sleep(time.Millisecond * 500)
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, true, cb.IsTripped())

		// 只会允许3个请求
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, false, cb.Ready())
		assert.Equal(t, true, cb.IsTripped())

		cb.Success()
		assert.Equal(t, true, cb.Ready())
		assert.Equal(t, true, cb.IsTripped())
	})
}

func TestNewRateRequestRetry(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewBreakerWithOptions(&Options{
			ShouldTrip:   ThresholdTripFunc(1),
			ShouldResume: ConsecutiveResumeFunc(3),
			ShouldRetry: NewRateRequestRetry(
				0.3,
				NewExponentialBackOff(time.Millisecond*100, time.Millisecond*500),
			),
		})

		// 断路
		cb.Fail()
		assert.Equal(t, true, cb.IsTripped())

		// 等待静默期后，进入半开状态
		time.Sleep(time.Millisecond * 500)

		// 保证样本集数量
		var failCount, successCount float64
		for i := 0; i < 1000; i++ {
			if cb.Ready() {
				successCount++
			} else {
				failCount++
			}
			time.Sleep(time.Millisecond * 10)
		}

		assert.Equal(t, "0.3", fmt.Sprintf("%.1f", successCount/(failCount+successCount)))
	})
}
