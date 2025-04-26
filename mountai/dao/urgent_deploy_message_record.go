package dao

import (
	"rulai/models/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateUrgentDeployDingTalkMsgRecord creates urgent deployment message record.
func (d *Dao) CreateUrgentDeployDingTalkMsgRecord(ctx context.Context,
	msg *entity.UrgentDeploymentDingTalkMsgRecord) (primitive.ObjectID, error) {
	res, err := d.Mongo.Collection(new(entity.UrgentDeploymentDingTalkMsgRecord).TableName()).
		InsertOne(ctx, msg)
	if err != nil {
		return primitive.NilObjectID, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return res.InsertedID.(primitive.ObjectID), nil
}

// ListUrgentDeployDingTalkMsgRecords lists urgent deployment ding talk records.
func (d *Dao) ListUrgentDeployDingTalkMsgRecords(ctx context.Context, filter bson.M,
	opts ...*options.FindOptions) ([]*entity.UrgentDeploymentDingTalkMsgRecord, error) {
	records := make([]*entity.UrgentDeploymentDingTalkMsgRecord, 0)
	err := d.Mongo.ReadOnlyCollection(new(entity.UrgentDeploymentDingTalkMsgRecord).TableName()).
		Find(ctx, filter, opts...).
		Decode(&records)
	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return records, nil
}
