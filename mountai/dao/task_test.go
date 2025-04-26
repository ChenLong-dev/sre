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

func TestDao_Task(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		id, err := d.CreateSingleTask(context.Background(), &entity.Task{
			ID:         primitive.NewObjectID(),
			Version:    "test-20190920152458",
			Action:     "start",
			Status:     "init",
			Detail:     "init",
			OperatorID: "1",
			AppID:      "5e81b1edcd3fb2a99a0f06f2",
			EnvName:    "stg",
		})
		assert.Nil(t, err)

		task, err := d.FindSingleTask(context.Background(), bson.M{
			"_id": id,
		})
		assert.Nil(t, err)
		assert.Equal(t, "test-20190920152458", task.Version)
		assert.Equal(t, "start", string(task.Action))
		assert.Equal(t, "init", task.Detail)

		err = d.UpdateSingleTask(context.Background(), id.Hex(), bson.A{
			bson.M{
				"$set": bson.M{
					"param": bson.M{
						"image_version":    "test:20190920152350",
						"health_check_url": "/health",
						"is_auto_scale":    true,
						"target_port":      80,
					},
					"detail": bson.M{
						"$concat": bson.A{
							"$detail", "\nstart",
						},
					},
				},
			},
		})
		assert.Nil(t, err)

		tasks, err := d.FindTasks(context.Background(), bson.M{
			"_id": id,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(tasks))
		assert.Equal(t, 80, tasks[0].Param.TargetPort)
		assert.True(t, tasks[0].Param.IsAutoScale)
		assert.Equal(t, "init\nstart", tasks[0].Detail)

		tasks, err = d.FindTasksGroupByAppIDAndEnvName(context.Background(), bson.M{
			"action": "start",
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(tasks))
		assert.Equal(t, 80, tasks[0].Param.TargetPort)

		err = d.DeleteSingleTask(context.Background(), bson.M{
			"_id": id,
		})
		assert.Nil(t, err)

		_, err = d.FindSingleTask(context.Background(), bson.M{
			"_id": id,
		})
		assert.True(t, errcode.EqualError(errcode.NoRowsFoundError, err))
	})
}
