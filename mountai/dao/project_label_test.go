package dao

import (
	"rulai/models/entity"

	"context"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"testing"
)

func TestDao_projectLabels(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		labels, err := d.GetProjectLabels(context.Background(), bson.M{})
		lenLabels := len(labels)
		assert.Nil(t, err)

		err = d.CreateProjectLabels(context.Background(), &entity.ProjectLabel{
			Label: "test_label",
			Name:  "测试",
		})
		assert.Nil(t, err)

		labels, err = d.GetProjectLabels(context.Background(), bson.M{})
		assert.Nil(t, err)
		assert.Equal(t, lenLabels+1, len(labels))

		err = d.DeleteProjectLabel(context.Background(), bson.M{
			"label": "test_label",
		})
		assert.Nil(t, err)

		labels, err = d.GetProjectLabels(context.Background(), bson.M{})
		assert.Nil(t, err)
		assert.Equal(t, lenLabels, len(labels))
	})
}
