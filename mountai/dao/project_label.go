package dao

import (
	"rulai/models/entity"

	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d *Dao) CreateProjectLabels(ctx context.Context, label *entity.ProjectLabel) error {
	_, err := d.Mongo.Collection(new(entity.ProjectLabel).TableName()).
		InsertOne(ctx, label)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) DeleteProjectLabel(ctx context.Context, filter bson.M) error {
	_, err := d.Mongo.Collection(new(entity.ProjectLabel).TableName()).
		DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) GetProjectLabels(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*entity.ProjectLabel, error) {
	res := make([]*entity.ProjectLabel, 0)
	err := d.Mongo.ReadOnlyCollection(new(entity.ProjectLabel).TableName()).
		Find(ctx, filter, opts...).
		Decode(&res)
	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return res, nil
}
