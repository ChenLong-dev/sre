package dao

import (
	"context"
	"time"

	"rulai/models/entity"
	_errcode "rulai/utils/errcode"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d *Dao) CreateSingleTask(ctx context.Context, task *entity.Task) (primitive.ObjectID, error) {
	res, err := d.Mongo.Collection(new(entity.Task).TableName()).
		InsertOne(ctx, task)
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
func (d *Dao) DeleteSingleTask(ctx context.Context, filter bson.M) error {
	_, err := d.Mongo.Collection(new(entity.Task).TableName()).
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

// 软删除
func (d *Dao) DeleteTasks(ctx context.Context, filter bson.M) error {
	_, err := d.Mongo.Collection(new(entity.Task).TableName()).
		UpdateMany(ctx, filter, bson.M{
			"$set": bson.M{
				"delete_time": time.Now(),
			},
		})
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) FindSingleTask(ctx context.Context, filter bson.M) (*entity.Task, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}
	task := new(entity.Task)

	err := d.Mongo.ReadOnlyCollection(task.TableName()).
		FindOne(ctx, filter).
		Decode(task)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return task, nil
}

func (d *Dao) FindTasks(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*entity.Task, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}
	res := make([]*entity.Task, 0)

	err := d.Mongo.ReadOnlyCollection(new(entity.Task).TableName()).
		Find(ctx, filter, opts...).
		Decode(&res)
	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return res, nil
}

func (d *Dao) CountTasks(ctx context.Context, filter bson.M, opts ...*options.CountOptions) (int, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}
	res, err := d.Mongo.ReadOnlyCollection(new(entity.Task).TableName()).
		CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return int(res), nil
}

func (d *Dao) UpdateSingleTask(ctx context.Context, id string, change interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
	}

	_, err = d.Mongo.Collection(new(entity.Task).TableName()).
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

// 根据app id及环境名来分组查找任务
func (d *Dao) FindTasksGroupByAppIDAndEnvName(ctx context.Context, filter bson.M) ([]*entity.Task, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}

	res := make([]*entity.Task, 0)
	err := d.Mongo.Collection(new(entity.Task).TableName()).
		Aggregate(ctx, bson.A{
			bson.M{
				"$match": filter,
			},
			bson.M{
				"$group": bson.M{
					"_id": bson.M{
						"app_id":   "$app_id",
						"env_name": "$env_name",
					},
					"doc": bson.M{
						"$first": "$$ROOT",
					},
				},
			},
			bson.M{
				"$replaceRoot": bson.M{
					"newRoot": "$doc",
				},
			},
		}).
		Decode(&res)
	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return res, nil
}
