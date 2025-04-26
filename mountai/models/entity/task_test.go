package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntity_TaskActionList(t *testing.T) {
	t.Run("list is nil", func(t *testing.T) {
		list := &TaskActionList{}
		ret := list.Contains(TaskActionStop)
		assert.False(t, ret)
	})

	t.Run("normal", func(t *testing.T) {
		list := &TaskActionList{
			TaskActionCanaryDeploy,
			TaskActionUpdateHPA,
		}

		assert.False(t, list.Contains(TaskActionStop))
		assert.True(t, list.Contains(TaskActionCanaryDeploy))
	})
}
