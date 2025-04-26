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

// UpdateUserSubscription 更新订阅
func (d *Dao) UpdateUserSubscription(ctx context.Context, filter, change bson.M, opts ...*options.UpdateOptions) error {
	_, err := d.Mongo.Collection(new(entity.UserSubscription).TableName()).
		UpdateOne(ctx, filter, change, opts...)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

// FindUserSubscription 获取订阅信息
func (d *Dao) FindUserSubscription(ctx context.Context, filter bson.M) (sub *entity.UserSubscription, err error) {
	sub = new(entity.UserSubscription)

	err = d.Mongo.ReadOnlyCollection(sub.TableName()).
		FindOne(ctx, filter).
		Decode(sub)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return sub, nil
}

// FindUserSubscriptionList 获取订阅列表
func (d *Dao) FindUserSubscriptionList(ctx context.Context, filter bson.M,
	opts ...*options.FindOptions) ([]*entity.UserSubscription, error) {
	subs := make([]*entity.UserSubscription, 0)

	err := d.Mongo.ReadOnlyCollection(new(entity.UserSubscription).TableName()).
		Find(ctx, filter, opts...).
		Decode(&subs)
	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return subs, nil
}

// DeleteUserSubscription 删除订阅
func (d *Dao) DeleteUserSubscription(ctx context.Context, filter bson.M) error {
	_, err := d.Mongo.Collection(new(entity.UserSubscription).TableName()).
		DeleteMany(ctx, filter)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}
