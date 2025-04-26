package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rulai/models/entity"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AppRunningTasksCacheKeyFormatter 处于运行状态(包括失败)的部署缓存键格式
const AppRunningTasksCacheKeyFormatter = "app_running_tasks::%s::%s"

// TaskClusterAndVersionKeyFormatter AMS 任务集群-版本缓存键格式
const TaskClusterAndVersionKeyFormatter = "%s::%s"

// AppRunningTasksSyncKeyFormatter 处于运行状态(包括失败)的部署缓存是否同步的记录键格式
const AppRunningTasksSyncKeyFormatter = "app_running_tasks_sync::%s::%s"

// NoInverseVersion 不指定需要保留的 task 版本
const NoInverseVersion = ""

func (d *Dao) CreateSingleApp(ctx context.Context, app *entity.App) (primitive.ObjectID, error) {
	res, err := d.Mongo.Collection(new(entity.App).TableName()).
		InsertOne(ctx, app)
	if err != nil {
		return primitive.NilObjectID, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, errors.Wrap(errcode.InternalError, "inserted id is not object id")
	}

	return id, nil
}

// 软删除
func (d *Dao) DeleteSingleApp(ctx context.Context, filter bson.M) error {
	_, err := d.Mongo.Collection(new(entity.App).TableName()).
		UpdateOne(ctx, filter, bson.M{
			"$set": bson.M{
				"delete_time": time.Now(),
			},
		})
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) UpdateSingleApp(ctx context.Context, id string, change bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
	}

	_, err = d.Mongo.Collection(new(entity.App).TableName()).
		UpdateOne(ctx, bson.M{
			"_id": objectID,
			"delete_time": bson.M{
				"$eq": primitive.Null{},
			},
		}, change)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) FindSingleApp(ctx context.Context, filter bson.M) (*entity.App, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}
	app := new(entity.App)

	err := d.Mongo.ReadOnlyCollection(app.TableName()).
		FindOne(ctx, filter).
		Decode(app)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return app, nil
}

// FindSingleAppByObjectID 通过主键查找应用(忽略删除状态)
func (d *Dao) FindSingleAppByObjectID(ctx context.Context, objectID primitive.ObjectID) (*entity.App, error) {
	app := new(entity.App)
	err := d.Mongo.ReadOnlyCollection(app.TableName()).
		FindOne(ctx, bson.M{"_id": objectID}).
		Decode(app)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return app, nil
}

func (d *Dao) FindApps(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*entity.App, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}
	res := make([]*entity.App, 0)

	err := d.Mongo.ReadOnlyCollection(new(entity.App).TableName()).
		Find(ctx, filter, opts...).
		Decode(&res)
	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return res, nil
}

func (d *Dao) CountApps(ctx context.Context, filter bson.M, opts ...*options.CountOptions) (int, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}
	res, err := d.Mongo.ReadOnlyCollection(new(entity.App).TableName()).
		CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return int(res), nil
}

// SetAppRunningTasks 设置应用在指定环境下处于运行状态(包括失败)的所有部署
// 设置时需要保证数据的有效性, 只有同步过数据的情况下才能进行设置(isSync=false的情形)
// 指定 isSync=true 代表强制进行数据同步(必须设置所有部署信息)
func (d *Dao) SetAppRunningTasks(ctx context.Context,
	appID string, envName entity.AppEnvName, tasks []*resp.TaskDetailResp, isSync bool) error {
	con := d.Redis.Get()
	defer con.Close()

	if !isSync {
		_, err := redis.Bytes(con.Do(ctx, "get", d.getAppRunningTasksSyncKey(ctx, envName, appID)))
		if err == redis.ErrNil {
			log.Warnc(ctx,
				"skipped setting %d running tasks for app(%s) in env(%s)", len(tasks), appID, envName)
			return nil
		}

		if err != nil {
			return errors.Wrap(errcode.RedisError, err.Error())
		}
	}

	// 待设置列表如果为空, 则不需要设置
	if len(tasks) > 0 {
		args := new(redis.Args).Add(d.getAppRunningTasksCacheKey(ctx, envName, appID))
		for _, task := range tasks {
			taskData, e := json.Marshal(task)
			if e != nil {
				return errors.Wrap(errcode.InternalError, e.Error())
			}

			args = args.Add(d.getTaskClusterAndVersionKey(ctx, task.ClusterName, task.Version))
			args = args.Add(taskData)
		}

		_, err := con.Do(ctx, "hmset", args...)
		if err != nil {
			return errors.Wrap(errcode.RedisError, err.Error())
		}
	}

	if isSync {
		// 无论待设置列表是否为空, 同步记录都需要设置
		_, err := con.Do(ctx, "set", d.getAppRunningTasksSyncKey(ctx, envName, appID), "")
		if err != nil {
			return errors.Wrap(errcode.RedisError, err.Error())
		}
	}

	return nil
}

// RemoveAppClusterRunningTasks 移除应用在指定环境指定集群下的指定部署
// 此时不需要判断是否同步过数据(未同步数据时 redis 即使有存储的数据也对查询无效, 并且 hdel 也不会出现错误)
func (d *Dao) RemoveAppClusterRunningTasks(ctx context.Context,
	appID string, envName entity.AppEnvName, clusterName entity.ClusterName, targetVersion string) error {
	con := d.Redis.Get()
	defer con.Close()

	_, err := con.Do(ctx, "hdel",
		d.getAppRunningTasksCacheKey(ctx, envName, appID), d.getTaskClusterAndVersionKey(ctx, clusterName, targetVersion))
	if err != nil {
		return errors.Wrap(errcode.RedisError, err.Error())
	}

	return nil
}

// CleanAppRunningTasks 清理应用在指定环境下处于运行状态(包括失败)的所有部署, 支持以下两种方式:
//  1. 指定集群名称和忽略版本, 清理指定环境同集群下其他所有部署记录(对应全量部署和全量金丝雀部署清理其他部署的情况)
//  2. 不指定集群名称且不指定忽略版本, 清理指定环境下所有部署记录(对应删除应用的情况)
func (d *Dao) CleanAppRunningTasks(ctx context.Context,
	appID string, envName entity.AppEnvName, clusterName entity.ClusterName, inverseVersion string) error {
	if (clusterName == entity.EmptyClusterName && inverseVersion != NoInverseVersion) ||
		(clusterName != entity.EmptyClusterName && inverseVersion == NoInverseVersion) {
		return errors.Wrapf(errcode.InternalError, "unsupported cluster_name(%s) with inverse_version(%s)", clusterName, inverseVersion)
	}

	con := d.Redis.Get()
	defer con.Close()

	if inverseVersion == NoInverseVersion {
		_, err := con.Do(ctx, "del", d.getAppRunningTasksCacheKey(ctx, envName, appID))
		if err != nil {
			return errors.Wrap(errcode.RedisError, err.Error())
		}

		_, err = con.Do(ctx, "del", d.getAppRunningTasksSyncKey(ctx, envName, appID))
		if err != nil {
			return errors.Wrap(errcode.RedisError, err.Error())
		}

		return nil
	}

	// 部署理论上应该不会多, 暂时不考虑 hkeys 的阻塞问题
	values, err := redis.ByteSlices(redis.Values(con.Do(ctx, "hkeys", d.getAppRunningTasksCacheKey(ctx, envName, appID))))
	if err != nil {
		return errors.Wrapf(errcode.RedisError, err.Error())
	}

	clusterNameBytes := []byte(clusterName)
	inverseVersionBytes := []byte(inverseVersion)
	args := new(redis.Args).Add(d.getAppRunningTasksCacheKey(ctx, envName, appID))
	for _, value := range values {
		// key 格式见 getTaskClusterAndVersionKey 方法
		if !bytes.HasPrefix(value, clusterNameBytes) || bytes.HasSuffix(value, inverseVersionBytes) {
			continue
		}

		args = args.Add(value)
	}

	// 有可能并没有需要清除的历史部署(例如首次发布)
	if len(args) > 1 {
		_, err = con.Do(ctx, "hdel", args...)
		if err != nil {
			return errors.Wrap(errcode.RedisError, err.Error())
		}
	}

	return nil
}

// GetAppsRunningTasks 批量查询多个应用各自在指定环境下处于运行状态(包括失败)的所有部署
func (d *Dao) GetAppsRunningTasks(ctx context.Context,
	envName entity.AppEnvName, appIDs []string) (map[string][]*resp.TaskDetailResp, error) {
	con := d.Redis.Get()
	defer con.Close()

	args := new(redis.Args).AddFlat(d.getAppsRunningTasksSyncKeys(ctx, envName, appIDs))
	syncStatuses, err := redis.ByteSlices(con.Do(ctx, "mget", args...))
	if err != nil {
		return nil, errors.Wrapf(errcode.RedisError, err.Error())
	}

	for i := range appIDs {
		// 部署理论上应该不会多, 暂时不考虑 hgetall 的阻塞问题
		err = con.Send(ctx, "hgetall", d.getAppRunningTasksCacheKey(ctx, envName, appIDs[i]))
		if err != nil {
			return nil, errors.Wrapf(errcode.RedisError, err.Error())
		}
	}

	err = con.Flush(ctx)
	if err != nil {
		return nil, errors.Wrapf(errcode.RedisError, err.Error())
	}

	res := make(map[string][]*resp.TaskDetailResp, len(appIDs))
	for i := range appIDs {
		values, e := redis.ByteSlices(con.Receive(ctx))
		if e != nil {
			return nil, errors.Wrapf(errcode.RedisError, e.Error())
		}

		if syncStatuses[i] == nil {
			// 没有同步的应用不返回
			continue
		}

		tasksLength := len(values) / 2 // 每个 task 有两条 value
		res[appIDs[i]] = make([]*resp.TaskDetailResp, tasksLength)
		for index := range values {
			if index&1 == 0 { // 跳过 key
				continue
			}

			res[appIDs[i]][index/2] = new(resp.TaskDetailResp)
			err = json.Unmarshal(values[index], res[appIDs[i]][index/2])
			if err != nil {
				return nil, errors.Wrapf(errcode.InternalError, err.Error())
			}
		}
	}

	return res, nil
}

// GetAppRunningTasks 查询应用在指定环境下处于运行状态(包括失败)的所有部署
// 如果数据还没有同步过, 固定返回 nil 作为标志, 以便强制触发同步
func (d *Dao) GetAppRunningTasks(ctx context.Context, envName entity.AppEnvName, appID string) ([]*resp.TaskDetailResp, error) {
	con := d.Redis.Get()
	defer con.Close()

	_, err := redis.Bytes(con.Do(ctx, "get", d.getAppRunningTasksSyncKey(ctx, envName, appID)))
	if err == redis.ErrNil {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrapf(errcode.RedisError, err.Error())
	}

	// 部署理论上应该不会多, 暂时不考虑 hgetall 的阻塞问题
	values, err := redis.ByteSlices(con.Do(ctx, "hgetall", d.getAppRunningTasksCacheKey(ctx, envName, appID)))
	if err != nil {
		return nil, errors.Wrapf(errcode.RedisError, err.Error())
	}

	tasksLength := len(values) / 2 // 每个 task 有两条 value
	tasks := make([]*resp.TaskDetailResp, tasksLength)
	for index := range values {
		if index&1 == 0 { // 跳过 key
			continue
		}

		tasks[index/2] = new(resp.TaskDetailResp)
		err = json.Unmarshal(values[index], tasks[index/2])
		if err != nil {
			return nil, errors.Wrapf(errcode.InternalError, err.Error())
		}
	}

	return tasks, nil
}

func (d *Dao) getAppRunningTasksCacheKey(_ context.Context, envName entity.AppEnvName, appID string) string {
	return fmt.Sprintf(AppRunningTasksCacheKeyFormatter, appID, envName)
}

func (d *Dao) getTaskClusterAndVersionKey(_ context.Context, clusterName entity.ClusterName, version string) string {
	return fmt.Sprintf(TaskClusterAndVersionKeyFormatter, clusterName, version)
}

func (d *Dao) getAppRunningTasksSyncKey(_ context.Context, envName entity.AppEnvName, appID string) string {
	return fmt.Sprintf(AppRunningTasksSyncKeyFormatter, appID, envName)
}

func (d *Dao) getAppsRunningTasksSyncKeys(ctx context.Context, envName entity.AppEnvName, appIDs []string) []string {
	keys := make([]string, len(appIDs))
	for i := range keys {
		keys[i] = d.getAppRunningTasksSyncKey(ctx, envName, appIDs[i])
	}
	return keys
}
