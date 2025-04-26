package dao

import (
	"context"

	"rulai/models/entity"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateFavProject 创建项目收藏
func (d *Dao) CreateFavProject(ctx context.Context, favProject *entity.FavProject) (primitive.ObjectID, error) {
	res, err := d.Mongo.Collection(new(entity.FavProject).TableName()).
		InsertOne(ctx, favProject)
	if err != nil {
		return primitive.NilObjectID, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return res.InsertedID.(primitive.ObjectID), nil
}

// DeleteFavProject 删除项目收藏
func (d *Dao) DeleteFavProject(ctx context.Context, filter bson.M) error {
	_, err := d.Mongo.Collection(new(entity.FavProject).TableName()).
		DeleteMany(ctx, filter)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

// GetFavProject 获取收藏项目
func (d *Dao) GetFavProject(ctx context.Context, filter bson.M) (*entity.FavProject, error) {
	favProject := new(entity.FavProject)

	err := d.Mongo.ReadOnlyCollection(favProject.TableName()).
		FindOne(ctx, filter).
		Decode(favProject)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return favProject, nil
}

// GetFavProjectList 获取收藏项目列表
func (d *Dao) GetFavProjectList(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*entity.FavProject, error) {
	favProjects := make([]*entity.FavProject, 0)

	err := d.Mongo.ReadOnlyCollection(new(entity.FavProject).TableName()).
		Find(ctx, filter, opts...).
		Decode(&favProjects)
	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return favProjects, nil
}

// CountFavProjects 获取收藏数量
func (d *Dao) CountFavProjects(ctx context.Context, filter bson.M, opts ...*options.CountOptions) (int, error) {
	count, err := d.Mongo.ReadOnlyCollection(new(entity.FavProject).TableName()).
		CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return int(count), nil
}
