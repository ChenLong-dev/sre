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

func (d *Dao) CreateSingleImageArgsTemplate(ctx context.Context, imageArgsTemplate *entity.ImageArgsTemplate) error {
	_, err := d.Mongo.Collection(new(entity.ImageArgsTemplate).TableName()).
		InsertOne(ctx, imageArgsTemplate)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) FindImageArgsTemplates(ctx context.Context, filter bson.M, opts ...*options.FindOptions) (
	[]*entity.ImageArgsTemplate, error) {
	imageArgsTemplates := make([]*entity.ImageArgsTemplate, 0)
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}

	err := d.Mongo.ReadOnlyCollection(new(entity.ImageArgsTemplate).TableName()).
		Find(ctx, filter, opts...).
		Decode(&imageArgsTemplates)

	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return imageArgsTemplates, nil
}

func (d *Dao) CountImageArgsTemplate(ctx context.Context, filter bson.M, opts ...*options.CountOptions) (int, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}

	count, err := d.Mongo.ReadOnlyCollection(new(entity.ImageArgsTemplate).TableName()).
		CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return int(count), nil
}

func (d *Dao) FindSingleImageArgsTemplate(ctx context.Context, filter bson.M) (*entity.ImageArgsTemplate, error) {
	filter["delete_time"] = bson.M{"$eq": primitive.Null{}}

	imageArgsTemplate := new(entity.ImageArgsTemplate)

	err := d.Mongo.ReadOnlyCollection(imageArgsTemplate.TableName()).
		FindOne(ctx, filter).
		Decode(imageArgsTemplate)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return imageArgsTemplate, nil
}

func (d *Dao) UpdateSingleImageArgsTemplate(ctx context.Context, id string, change bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	_, err = d.Mongo.Collection(new(entity.ImageArgsTemplate).TableName()).
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

func (d *Dao) DeleteSingleImageArgsTemplateByID(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	_, err = d.Mongo.Collection(new(entity.ImageArgsTemplate).TableName()).
		UpdateOne(ctx, bson.M{
			"_id": objectID,
		}, bson.M{
			"$set": bson.M{
				"delete_time": time.Now(),
			},
		})
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}
