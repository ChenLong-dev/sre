package dao

import (
	"context"
	"testing"

	"rulai/models/entity"

	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDao_FavProject(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		_, err := d.CreateFavProject(context.Background(), &entity.FavProject{
			ID:        primitive.NewObjectID(),
			UserID:    "123",
			ProjectID: "1107",
		})
		assert.Nil(t, err)

		favProject, err := d.GetFavProject(context.Background(), bson.M{
			"user_id": "123",
		})
		assert.Nil(t, err)
		assert.Equal(t, "123", favProject.UserID)
		assert.Equal(t, "1107", favProject.ProjectID)

		count, err := d.CountFavProjects(context.Background(), bson.M{
			"user_id": "123",
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, count)

		favProjects, err := d.GetFavProjectList(context.Background(), bson.M{
			"user_id": "123",
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(favProjects))
		assert.Equal(t, "123", favProjects[0].UserID)

		err = d.DeleteFavProject(context.Background(), bson.M{
			"user_id": "123",
		})
		assert.Nil(t, err)

		_, err = d.GetFavProject(context.Background(), bson.M{
			"user_id": "123",
		})
		assert.True(t, errcode.EqualError(errcode.NoRowsFoundError, err))
	})
}
