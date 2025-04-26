package slice

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStrSlice_contains(t *testing.T) {
	strSli := []string{"name", "age", "gender"}

	t.Run("is_string_slice_contains", func(t *testing.T) {
		isContains := StrSliceContains(strSli, "gender")
		assert.Equal(t, true, isContains)
	})

	t.Run("not_string_slice_contains", func(t *testing.T) {
		notContains := StrSliceContains(strSli, "agee")
		assert.Equal(t, false, notContains)
	})
}
