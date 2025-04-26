package service

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/req"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/resp"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/base/null"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

// 根据请求生成不同的查询bson
func getAvailableVipItemListSearchBson(getReq *req.GetVipItemListReq) (search bson.M, err error) {
	search = bson.M{}
	search["disabled"] = bson.M{
		"$ne": true,
	}

	if getReq.ItemType.ValueOrZero() == "vip" {
		search["vip_id"] = "super_vip"
	}

	if getReq.Include.ValueOrZero() != "autorenew" {
		search["autorenew"] = bson.M{
			"$ne": true,
		}
	}

	getReq.PhoneType = null.StringFrom(strings.ToLower(getReq.PhoneType.ValueOrZero()))
	if getReq.PhoneType.ValueOrZero() == "ios" {
		search["phonetype"] = "ios"
	} else {
		search["phonetype"] = bson.M{
			"$ne": "ios",
		}
	}

	return search, nil
}

// 获取商品列表
func (s *Service) GetAvailableVipItemList(ctx context.Context, getReq *req.GetVipItemListReq) ([]*resp.VipItemListResponse, error) {
	// 生成查询bson
	search, err := getAvailableVipItemListSearchBson(getReq)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return nil, err
	}

	// ==========================
	// 查询mongo数据库
	// ==========================
	res, err := s.dao.FindVipItemsList(ctx, search, &options.FindOptions{
		Sort: bson.M{
			"fee": 1,
		},
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
	list := make([]*resp.VipItemListResponse, 0)
	for _, item := range res {
		listResp := new(resp.VipItemListResponse)
		err := deepcopy.Copy(item).To(listResp)
		if err != nil {
			// ==========================
			// 在首次生成error时，应当立即使用errors.Wrapf包裹
			// 外层只需直接返回error，无需再次包裹
			// ==========================
			return nil, errors.Wrapf(errcode.InternalError, "%s", err)
		}

		list = append(list, listResp)
	}

	return list, nil
}
