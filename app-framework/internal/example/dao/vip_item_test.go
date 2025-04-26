package dao

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

func TestDao_FindVipItemsList(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list, err := d.FindVipItemsList(context.Background(), bson.M{})
		assert.Nil(t, err)
		assert.Equal(t, true, len(list) > 0)
	})

	t.Run("filter", func(t *testing.T) {
		list, err := d.FindVipItemsList(context.Background(), bson.M{
			"disabled": bson.M{
				"$ne": true,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, true, len(list) > 0)
	})

	t.Run("sort", func(t *testing.T) {
		list, err := d.FindVipItemsList(context.Background(), bson.M{}, &options.FindOptions{
			Sort: bson.M{
				"fee": 1,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, true, len(list) >= 2)
		assert.Equal(t, true, list[0].Fee <= list[1].Fee)
	})
}
