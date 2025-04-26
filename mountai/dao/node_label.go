package dao

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"

	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/entity"
)

// FindNodeLabelLists : 查询所有支持的节点标签列表
func (d *Dao) FindNodeLabelLists(ctx context.Context, filter bson.M) ([]*entity.NodeLabelList, error) {
	labels := make([]*entity.NodeLabelList, 0)
	err := d.Mongo.ReadOnlyCollection(new(entity.NodeLabelList).TableName()).
		Find(ctx, filter, MongoFindOptionWithSortByIDAsc).
		Decode(&labels)
	if err != nil {
		return nil, errors.Wrap(errcode.MongoError, err.Error())
	}

	return labels, nil
}
