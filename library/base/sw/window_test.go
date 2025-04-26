package sw

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSlidingWindowCounts(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		w := NewSlidingWindow(time.Second*10, 2)
		w.Fail()
		w.Fail()
		w.Success()
		w.Success()

		assert.Equal(t, int64(2), w.FailureCount())
		assert.Equal(t, int64(2), w.SuccessCount())
		assert.Equal(t, 0.5, w.ErrorRate())
	})

	t.Run("reset", func(t *testing.T) {
		w := NewSlidingWindow(time.Second*10, 2)
		w.Fail()
		w.Success()
		w.Reset()

		assert.Equal(t, int64(0), w.FailureCount())
		assert.Equal(t, int64(0), w.SuccessCount())
	})
}

func TestSlidingWindowSlide(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		w := NewSlidingWindow(time.Second, 5)
		w.Fail()
		assert.Equal(t, int64(1), w.FailureCount())

		time.Sleep(time.Second * 5)
		w.Fail()
		assert.Equal(t, int64(1), w.FailureCount())
	})
}
