package dao

import (
	"context"
	"testing"

	"rulai/models/entity"
	"rulai/models/resp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDao_App(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		id, err := d.CreateSingleApp(context.Background(), &entity.App{
			ID:           primitive.NewObjectID(),
			Name:         "Web服务",
			Type:         "service",
			ServiceType:  entity.AppServiceTypeRestful,
			ProjectID:    "1",
			LogTTLInDays: 7,
		})
		assert.Nil(t, err)

		app, err := d.FindSingleApp(context.Background(), bson.M{
			"_id": id,
		})
		assert.Nil(t, err)
		assert.Equal(t, "Web服务", app.Name)

		err = d.UpdateSingleApp(context.Background(), id.Hex(), bson.M{
			"$set": bson.M{
				"env": bson.M{
					"stg": bson.M{
						"desired_branch": "staging",
					},
					"prd": bson.M{
						"desired_branch": "master",
					},
				},
			},
		})
		assert.Nil(t, err)

		count, err := d.CountApps(context.Background(), bson.M{
			"_id": id,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, count)

		apps, err := d.FindApps(context.Background(), bson.M{
			"_id": id,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(apps))
		assert.Equal(t, 2, len(apps[0].Env))

		err = d.DeleteSingleApp(context.Background(), bson.M{
			"_id": id,
		})
		assert.Nil(t, err)

		_, err = d.FindSingleApp(context.Background(), bson.M{
			"_id": id,
		})
		assert.True(t, errcode.EqualError(errcode.NoRowsFoundError, err))
	})
}

func Test_Dao_AppRunningTasksOperations(t *testing.T) {
	t.Run("app_running_tasks_operations", func(t *testing.T) {
		ctx := context.TODO()
		appID := "fake_app_id_001" // NOTE: 勿改成真实的, 目前会写入 stg 的 redis TODO: 调研 github.com/alicebob/miniredis 库是否可以用来做单元测试
		envName := entity.AppEnvStg
		clusterJuncle := entity.ClusterJuncle
		emptyTasks := make([]*resp.TaskDetailResp, 0)
		tasks := []*resp.TaskDetailResp{
			{Version: "1", ClusterName: entity.ClusterZeus},
			{Version: "2", ClusterName: entity.ClusterJuncle},
			{Version: "3", ClusterName: entity.ClusterJuncle},
			{Version: "4", ClusterName: entity.ClusterJuncle},
			{Version: "5", ClusterName: entity.ClusterZeus},
		}

		// 未同步数据时 Set 应当无效
		err := d.SetAppRunningTasks(ctx, appID, envName, nil, false) // 空值
		require.NoError(t, err)

		getTasks, err := d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		assert.Empty(t, getTasks)

		err = d.SetAppRunningTasks(ctx, appID, envName, emptyTasks, false) // 空数组
		require.NoError(t, err)

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		assert.Empty(t, getTasks)

		err = d.SetAppRunningTasks(ctx, appID, envName, tasks, false) // 正常值
		require.NoError(t, err)

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		assert.Empty(t, getTasks)

		// 进行数据同步
		err = d.SetAppRunningTasks(ctx, appID, envName, nil, false) // 空值
		require.NoError(t, err)

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		assert.Empty(t, getTasks)

		err = d.SetAppRunningTasks(ctx, appID, envName, emptyTasks, false) // 空数组
		require.NoError(t, err)

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		assert.Empty(t, getTasks)

		err = d.SetAppRunningTasks(ctx, appID, envName, tasks, true) // 正常值
		require.NoError(t, err)

		expectedRemainingTasks := make(map[string]*resp.TaskDetailResp, len(tasks))
		for _, task := range tasks {
			expectedRemainingTasks[d.getTaskClusterAndVersionKey(ctx, task.ClusterName, task.Version)] = task
		}

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		checkRemainingTasks(t, expectedRemainingTasks, getTasks)

		// 测试删除指定值
		removedTask := tasks[2]
		err = d.RemoveAppClusterRunningTasks(ctx, appID, envName, removedTask.ClusterName, removedTask.Version)
		assert.NoError(t, err)

		delete(expectedRemainingTasks, d.getTaskClusterAndVersionKey(ctx, removedTask.ClusterName, removedTask.Version))

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		checkRemainingTasks(t, expectedRemainingTasks, getTasks)

		// 测试清理参数不正确
		inversedTask := tasks[1]
		err = d.CleanAppRunningTasks(ctx, appID, envName, entity.EmptyClusterName, inversedTask.Version)
		assert.True(t, errcode.EqualError(errcode.InternalError, err))

		err = d.CleanAppRunningTasks(ctx, appID, envName, inversedTask.ClusterName, "")
		assert.True(t, errcode.EqualError(errcode.InternalError, err))

		// 测试清理指定集群并保留指定值
		err = d.CleanAppRunningTasks(ctx, appID, envName, inversedTask.ClusterName, inversedTask.Version)
		assert.NoError(t, err)

		for _, task := range tasks {
			if task.ClusterName == inversedTask.ClusterName && task.Version != inversedTask.Version {
				delete(expectedRemainingTasks, d.getTaskClusterAndVersionKey(ctx, task.ClusterName, task.Version))
			}
		}

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		checkRemainingTasks(t, expectedRemainingTasks, getTasks)

		// 清理指定集群并保留指定值特殊情形: 指定集群下除保留值外没有需要清理的值
		err = d.CleanAppRunningTasks(ctx, appID, envName, inversedTask.ClusterName, inversedTask.Version)
		assert.NoError(t, err)

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		checkRemainingTasks(t, expectedRemainingTasks, getTasks)

		// 测试所有删除情形, 包括删除不存在的值以及最后一个值
		for _, task := range tasks {
			err = d.RemoveAppClusterRunningTasks(ctx, appID, envName, task.ClusterName, task.Version)
			assert.NoError(t, err)

			delete(expectedRemainingTasks, d.getTaskClusterAndVersionKey(ctx, task.ClusterName, task.Version))

			getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
			assert.NoError(t, err)
			checkRemainingTasks(t, expectedRemainingTasks, getTasks)
		}

		// 清理并保留指定值特殊情形: 没有任何值
		err = d.CleanAppRunningTasks(ctx, appID, envName, clusterJuncle, inversedTask.Version)
		assert.NoError(t, err)

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		assert.Empty(t, getTasks)

		// 恢复初始数据量
		err = d.SetAppRunningTasks(ctx, appID, envName, tasks, false)
		require.NoError(t, err)

		for _, task := range tasks {
			expectedRemainingTasks[d.getTaskClusterAndVersionKey(ctx, task.ClusterName, task.Version)] = task
		}

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		checkRemainingTasks(t, expectedRemainingTasks, getTasks)

		// 测试完全清理
		err = d.CleanAppRunningTasks(ctx, appID, envName, entity.EmptyClusterName, NoInverseVersion)
		assert.NoError(t, err)

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		assert.Empty(t, getTasks)

		// 测试完全清理不存在的值
		err = d.CleanAppRunningTasks(ctx, appID, envName, entity.EmptyClusterName, NoInverseVersion)
		assert.NoError(t, err)

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		assert.Empty(t, getTasks)

		// 完全清理后测试 Set 无效的情形
		err = d.SetAppRunningTasks(ctx, appID, envName, tasks, false)
		require.NoError(t, err)

		getTasks, err = d.GetAppRunningTasks(ctx, envName, appID)
		assert.NoError(t, err)
		assert.Empty(t, getTasks)
	})
}

func checkRemainingTasks(t *testing.T, expectedRemainingTasks map[string]*resp.TaskDetailResp, act []*resp.TaskDetailResp) {
	if len(expectedRemainingTasks) == 0 {
		assert.Empty(t, act)
		return
	}

	ctx := context.TODO()
	if assert.Len(t, act, len(expectedRemainingTasks)) {
		for _, task := range act {
			expTask, ok := expectedRemainingTasks[d.getTaskClusterAndVersionKey(ctx, task.ClusterName, task.Version)]
			assert.True(t, ok)
			assert.EqualValues(t, expTask, task)
		}
	}
}
