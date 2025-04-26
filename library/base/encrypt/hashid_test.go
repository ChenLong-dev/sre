package encrypt

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeID(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data, err := EncodeID("normal", 1)
		assert.Nil(t, err)
		assert.NotEqual(t, "1", data)
	})
}

func TestDecodeID(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data, err := EncodeID("normal", 1)
		assert.Nil(t, err)

		id, err := DecodeID("normal", data)
		assert.Nil(t, err)
		assert.Equal(t, 1, id)
	})
}
