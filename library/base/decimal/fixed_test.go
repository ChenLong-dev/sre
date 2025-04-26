package decimal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToFixed(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		n := float64(1.123)
		n, err := ToFixed(1, n)
		assert.Nil(t, err)
		assert.Equal(t, 1.1, n)
	})

	t.Run("3", func(t *testing.T) {
		n := float64(1.1234567)
		n, err := ToFixed(3, n)
		assert.Nil(t, err)
		assert.Equal(t, 1.123, n)
	})
}
