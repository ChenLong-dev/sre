package dao

import (
	"rulai/models/entity"

	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d *Dao) CreateSingleProject(ctx context.Context, project *entity.Project) error {
	_, err := d.Mongo.Collection(new(entity.Project).TableName()).
		InsertOne(ctx, project)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

// 硬删除
// 由于通过git的项目id作为主键，无法软删除
func (d *Dao) DeleteSingleProject(ctx context.Context, filter bson.M) error {
	_, err := d.Mongo.Collection(new(entity.Project).TableName()).
		DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) FindSingleProject(ctx context.Context, filter bson.M) (*entity.Project, error) {
	project := new(entity.Project)

	err := d.Mongo.ReadOnlyCollection(project.TableName()).
		FindOne(ctx, filter).
		Decode(project)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return project, nil
}

func (d *Dao) FindProjects(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*entity.Project, error) {
	res := make([]*entity.Project, 0)

	err := d.Mongo.ReadOnlyCollection(new(entity.Project).TableName()).
		Find(ctx, filter, opts...).
		Decode(&res)
	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return res, nil
}

func (d *Dao) CountProject(ctx context.Context, filter bson.M, opts ...*options.CountOptions) (int, error) {
	res, err := d.Mongo.ReadOnlyCollection(new(entity.Project).TableName()).
		CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return int(res), nil
}

func (d *Dao) UpdateSingleProject(ctx context.Context, id string, change bson.M) error {
	_, err := d.Mongo.Collection(new(entity.Project).TableName()).
		UpdateOne(ctx, bson.M{
			"_id": id,
		}, change)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}
