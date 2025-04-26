package test

import (
	"github.com/stretchr/testify/assert"

	"path/filepath"
	"testing"
)

func TestDecodeUnitConfigFromLocal(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		assert.NotPanics(t, func() {
			unitConfig := DecodeUnitConfigFromLocal(filepath.Join("..", "..", UnitConfigPath))
			assert.NotNil(t, unitConfig)
		})
	})

	t.Run("invalid path", func(t *testing.T) {
		assert.Panics(t, func() {
			unitConfig := DecodeUnitConfigFromLocal("somepath")
			assert.Nil(t, unitConfig)
		})
	})
}
