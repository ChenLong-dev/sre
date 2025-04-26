package dao

import (
	"context"
	"testing"

	"rulai/models/entity"

	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
)

func TestDao_Project(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		err := d.CreateSingleProject(context.Background(), &entity.Project{
			ID:       "1",
			Name:     "test",
			Language: "golang",
			Desc:     "Framework示例",
			TeamID:   "1",
		})
		assert.Nil(t, err)

		project, err := d.FindSingleProject(context.Background(), bson.M{
			"_id": "1",
		})
		assert.Nil(t, err)
		assert.Equal(t, "test", project.Name)
		assert.Equal(t, "golang", project.Language)

		err = d.UpdateSingleProject(context.Background(), "1", bson.M{
			"$set": bson.M{
				"name": "app-framework",
			},
		})
		assert.Nil(t, err)

		count, err := d.CountProject(context.Background(), bson.M{
			"_id": "1",
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, count)

		projects, err := d.FindProjects(context.Background(), bson.M{
			"_id": "1",
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(projects))
		assert.Equal(t, "app-framework", projects[0].Name)

		err = d.DeleteSingleProject(context.Background(), bson.M{
			"_id": "1",
		})
		assert.Nil(t, err)

		_, err = d.FindSingleProject(context.Background(), bson.M{
			"_id": "1",
		})
		assert.True(t, errcode.EqualError(errcode.NoRowsFoundError, err))
	})
}
