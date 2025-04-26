package dao

import (
	"rulai/models/entity"
	_errcode "rulai/utils/errcode"

	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d *Dao) CreateCIJob(ctx context.Context, ciJob *entity.CIJob) error {
	_, err := d.Mongo.Collection(ciJob.TableName()).
		InsertOne(ctx, ciJob)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) UpdateCIJob(ctx context.Context, id string, change bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
	}

	_, err = d.Mongo.Collection(new(entity.CIJob).TableName()).
		UpdateOne(ctx, bson.M{
			"_id": objectID,
		}, change)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) GetCIJob(ctx context.Context, filter bson.M) (*entity.CIJob, error) {
	ciJob := new(entity.CIJob)
	err := d.Mongo.ReadOnlyCollection(ciJob.TableName()).
		FindOne(ctx, filter).
		Decode(ciJob)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return ciJob, nil
}

func (d *Dao) DeleteCIJob(ctx context.Context, filter bson.M) error {
	_, err := d.Mongo.Collection(new(entity.CIJob).TableName()).
		DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) GetCIJobs(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*entity.CIJob, error) {
	res := make([]*entity.CIJob, 0)
	err := d.Mongo.ReadOnlyCollection(new(entity.CIJob).TableName()).
		Find(ctx, filter, opts...).
		Decode(&res)
	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return res, nil
}
