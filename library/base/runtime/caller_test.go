package runtime

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestGetFullCallers(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		callers := GetFullCallers()
		assert.Equal(t, true, len(callers) >= 3)
		assert.Equal(t, true, strings.Contains(callers[2], "base/runtime/caller_test.go"))
	})
}

func TestGetCaller(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		result := GetCaller(1)
		assert.Equal(t, true, strings.Contains(result, "base/runtime/caller_test.go"))
	})
}

func TestGetFilterCallers(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		result := GetFilterCallers(DefaultFilterCallerRegexp)
		assert.Equal(t, true, strings.Contains(result, "base/runtime/caller_test.go"))
	})
}
