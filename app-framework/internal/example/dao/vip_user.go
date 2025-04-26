package dao

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/entity"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// 查找单个vip用户信息
func (d *Dao) FindSingleVipUser(ctx context.Context, filter bson.M) (*entity.VipUser, error) {
	vipUser := new(entity.VipUser)

	// ==========================
	// 查询mongo只读数据库
	// ==========================
	err := d.PayMongo.ReadOnlyCollection(vipUser.TableName()).
		FindOne(ctx, filter).
		Decode(vipUser)
	// ==========================
	// 在首次生成error时，应当立即使用errors.Wrapf包裹
	// 外层只需直接返回error，无需再次包裹
	// ==========================
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return vipUser, nil
}
