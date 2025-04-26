package dao

import (
	"rulai/models/entity"
	_errcode "rulai/utils/errcode"

	"context"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d *Dao) CreateSingleVariable(ctx context.Context, variable *entity.Variable) error {
	_, err := d.Mongo.Collection(new(entity.Variable).TableName()).
		InsertOne(ctx, variable)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) FindVariables(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*entity.Variable, error) {
	variables := make([]*entity.Variable, 0)
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}

	err := d.Mongo.ReadOnlyCollection(new(entity.Variable).TableName()).
		Find(ctx, filter, opts...).
		Decode(&variables)

	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return variables, nil
}

func (d *Dao) CountVariable(ctx context.Context, filter bson.M, opts ...*options.CountOptions) (int, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}

	res, err := d.Mongo.ReadOnlyCollection(new(entity.Variable).TableName()).
		CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return int(res), nil
}

func (d *Dao) UpdateSingleVariableByID(ctx context.Context, id string, change bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	_, err = d.Mongo.Collection(new(entity.Variable).TableName()).
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

func (d *Dao) DeleteSingleVariableByID(ctx context.Context, id, operatorID string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	_, err = d.Mongo.Collection(new(entity.Variable).TableName()).
		UpdateOne(ctx, bson.M{
			"_id": objectID,
		}, bson.M{
			"$set": bson.M{
				"editor_id":   operatorID,
				"delete_time": time.Now(),
			},
		})
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) FindSingleVariable(ctx context.Context, filter bson.M) (*entity.Variable, error) {
	filter["delete_time"] = bson.M{"$eq": primitive.Null{}}

	variable := new(entity.Variable)

	err := d.Mongo.ReadOnlyCollection(variable.TableName()).
		FindOne(ctx, filter).
		Decode(variable)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return variable, nil
}
