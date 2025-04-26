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

func (d *Dao) CreateSingleUserAuth(ctx context.Context, user *entity.UserAuth) error {
	_, err := d.Mongo.Collection(new(entity.UserAuth).TableName()).
		InsertOne(ctx, user)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

// 硬删除
// 由于通过git的用户id作为主键，无法软删除
func (d *Dao) DeleteSingleUserAuth(ctx context.Context, filter bson.M) error {
	_, err := d.Mongo.Collection(new(entity.UserAuth).TableName()).
		DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) FindSingleUserAuth(ctx context.Context, filter bson.M) (*entity.UserAuth, error) {
	auth := new(entity.UserAuth)

	err := d.Mongo.ReadOnlyCollection(auth.TableName()).
		FindOne(ctx, filter).
		Decode(auth)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return auth, nil
}

func (d *Dao) UpdateSingleUserAuth(ctx context.Context, userID string, change bson.M) error {
	_, err := d.Mongo.Collection(new(entity.UserAuth).TableName()).
		UpdateOne(ctx, bson.M{"_id": userID}, change)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) FindUserAuth(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*entity.UserAuth, error) {
	auth := make([]*entity.UserAuth, 0)

	err := d.Mongo.ReadOnlyCollection(new(entity.UserAuth).TableName()).
		Find(ctx, filter, opts...).
		Decode(&auth)
	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return auth, nil
}

func (d *Dao) CountUser(ctx context.Context, filter bson.M, opts ...*options.CountOptions) (int, error) {
	res, err := d.Mongo.ReadOnlyCollection(new(entity.UserAuth).TableName()).
		CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return int(res), nil
}
