package dao

import (
	"context"
	"testing"

	"rulai/models/entity"

	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
)

func TestDao_UserAuth(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		err := d.CreateSingleUserAuth(context.Background(), &entity.UserAuth{
			ID:    "1",
			Token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9",
		})
		assert.Nil(t, err)

		auth, err := d.FindSingleUserAuth(context.Background(), bson.M{
			"_id": "1",
		})
		assert.Nil(t, err)
		assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9", auth.Token)

		err = d.UpdateSingleUserAuth(context.Background(), "1", bson.M{
			"$set": bson.M{
				"token": "eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6ImRhaWhlbmciLCJleHAiOjE1NzczODk0ODYsImVtYWlsIjoiIn0",
			},
		})
		assert.Nil(t, err)

		auth, err = d.FindSingleUserAuth(context.Background(), bson.M{
			"_id": "1",
		})
		assert.Nil(t, err)
		assert.Equal(t, "eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6ImRhaWhlbmciLCJleHAiOjE1NzczODk0ODYsImVtYWlsIjoiIn0", auth.Token)

		err = d.DeleteSingleUserAuth(context.Background(), bson.M{
			"_id": "1",
		})
		assert.Nil(t, err)

		_, err = d.FindSingleUserAuth(context.Background(), bson.M{
			"_id": "1",
		})
		assert.True(t, errcode.EqualError(errcode.NoRowsFoundError, err))
	})
}
