package service

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/resp"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
)

// 获取vip用户信息
func (s *Service) GetVipUserDetail(ctx context.Context, userID string) (*resp.VipUserDetail, error) {
	// ==========================
	// 查询mongo数据库
	// ==========================
	user, err := s.dao.FindSingleVipUser(ctx, bson.M{
		"user_id": userID,
	})
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return nil, err
	}

	// ==========================
	// 将实体复制到响应模型中，通过Library中的deepcopy包实现
	// ==========================
	result := new(resp.VipUserDetail)
	err = deepcopy.Copy(user).To(result)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return result, nil
}
