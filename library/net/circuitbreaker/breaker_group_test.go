package circuitbreaker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBreakerGroup_Get(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cb := NewBreaker()
		bg := NewBreakerGroup()
		bg.Add("a", cb)

		b := bg.Get("a")
		assert.Equal(t, cb, b)

		b = bg.Get("b")
		assert.Nil(t, b)
	})
}
