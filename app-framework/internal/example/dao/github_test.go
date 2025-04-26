package dao

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDao_OwnerGithubAggregationLock(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		mutex := d.GetOwnerGithubAggregationLock(context.Background(), "test")
		assert.NotNil(t, mutex)

		err := d.LockOwnerGithub(context.Background(), mutex)
		assert.Nil(t, err)

		err = d.UnlockOwnerGithub(context.Background(), mutex)
		assert.Nil(t, err)
	})
}
