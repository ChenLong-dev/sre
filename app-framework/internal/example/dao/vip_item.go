package dao

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/entity"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 查找vip商品列表
func (d *Dao) FindVipItemsList(ctx context.Context, search bson.M, opts ...*options.FindOptions) ([]*entity.VipItem, error) {
	var items []*entity.VipItem

	// ==========================
	// 查询mongo只读数据库
	// ==========================
	err := d.VipMongo.ReadOnlyCollection(new(entity.VipItem).TableName()).
		Find(ctx, search, opts...).Decode(&items)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return items, nil
}
