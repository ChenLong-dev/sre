package test

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestRunSyncCases(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		var num int
		RunSyncCases(t, []SyncUnitCase{
			{
				Name: "first",
				Func: func(t *testing.T) {
					num++
					assert.Equal(t, 1, num)
				},
			},
			{
				Name: "second",
				Func: func(t *testing.T) {
					num++
					assert.Equal(t, 2, num)
				},
			},
		}...)
	})
}
