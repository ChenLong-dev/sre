package dao

import (
	"rulai/models/entity"

	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateDingTalkMessageRecord 创建消息记录
func (d *Dao) CreateDingTalkMessageRecord(ctx context.Context, msg *entity.DingTalkMessageRecord) (primitive.ObjectID, error) {
	res, err := d.Mongo.Collection(new(entity.DingTalkMessageRecord).TableName()).
		InsertOne(ctx, msg)
	if err != nil {
		return primitive.NilObjectID, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return res.InsertedID.(primitive.ObjectID), nil
}
