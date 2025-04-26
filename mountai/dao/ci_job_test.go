package dao

import (
	"context"
	"testing"
	"time"

	"rulai/models/entity"

	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDao_CIJob(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		now := time.Now()
		ctx := context.Background()
		err := d.CreateCIJob(ctx, &entity.CIJob{
			ID:         primitive.NewObjectID(),
			Name:       "test_ci_job",
			ProjectID:  "-1",
			HookURL:    "hook_url",
			ViewURL:    "ciJob",
			CreateTime: &now,
			UpdateTime: &now,
		})
		assert.Nil(t, err)

		ciJob, err := d.GetCIJob(ctx, bson.M{
			"project_id": "-1",
		})
		assert.Nil(t, err)
		assert.Equal(t, "ciJob", ciJob.ViewURL)

		err = d.UpdateCIJob(ctx, ciJob.ID.Hex(), bson.M{
			"$set": bson.M{
				"view_url":    "test_ciJob",
				"update_time": time.Now(),
			},
		})
		assert.Nil(t, err)

		ciJob, err = d.GetCIJob(ctx, bson.M{
			"project_id": "-1",
		})
		assert.Nil(t, err)
		assert.Equal(t, "test_ciJob", ciJob.ViewURL)

		err = d.DeleteCIJob(ctx, bson.M{
			"project_id": "-1",
		})
		assert.Nil(t, err)

		_, err = d.GetCIJob(ctx, bson.M{
			"project_id": "-1",
		})
		assert.True(t, errcode.EqualError(errcode.NoRowsFoundError, err))
	})
}
