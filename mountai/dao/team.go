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

func (d *Dao) CreateSingleTeam(ctx context.Context, team *entity.Team) (primitive.ObjectID, error) {
	res, err := d.Mongo.Collection(new(entity.Team).TableName()).
		InsertOne(ctx, team)
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
func (d *Dao) DeleteSingleTeam(ctx context.Context, filter bson.M) error {
	_, err := d.Mongo.Collection(new(entity.Team).TableName()).
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

func (d *Dao) FindSingleTeam(ctx context.Context, filter bson.M) (*entity.Team, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}
	team := new(entity.Team)

	err := d.Mongo.ReadOnlyCollection(team.TableName()).
		FindOne(ctx, filter).
		Decode(team)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return team, nil
}

func (d *Dao) FindTeams(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*entity.Team, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}
	res := make([]*entity.Team, 0)

	err := d.Mongo.ReadOnlyCollection(new(entity.Team).TableName()).
		Find(ctx, filter, opts...).
		Decode(&res)
	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return res, nil
}

func (d *Dao) CountTeam(ctx context.Context, filter bson.M, opts ...*options.CountOptions) (int, error) {
	filter["delete_time"] = bson.M{
		"$eq": primitive.Null{},
	}
	res, err := d.Mongo.ReadOnlyCollection(new(entity.Team).TableName()).
		CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return int(res), nil
}

func (d *Dao) UpdateSingleTeam(ctx context.Context, id string, change bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
	}

	_, err = d.Mongo.Collection(new(entity.Team).TableName()).
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
