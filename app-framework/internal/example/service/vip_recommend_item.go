package service

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/dao"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/entity"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/resp"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/database/sql"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// 查找推荐商品缓存详情
func (s *Service) GetVipRecommendItemCacheDetail(ctx context.Context, id string) (*resp.VipRecommendItemDetail, error) {
	// ==========================
	// 获取缓存
	// ==========================
	cacheResp, err := s.dao.GetVipRecommendItemDetailFromCache(ctx, id)
	if err == nil {
		return cacheResp, nil
	}
	// ==========================
	// 如果没有命中缓存，应直接回源，不应报错
	// ==========================
	if !errcode.EqualError(errcode.RedisEmptyKeyError, err) {
		// ==========================
		// 打印error时，使用errcode.GetErrorMessageMap辅助方法
		// 可以打印包含调用栈的详细信息
		// ==========================
		log.Errorv(ctx, errcode.GetErrorMessageMap(err))
	}

	// ==========================
	// 回源
	// ==========================
	realResp, err := s.GetVipRecommendItemDetail(ctx, id)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return nil, err
	}

	// ==========================
	// 存入缓存
	// ==========================
	err = s.dao.SaveVipRecommendItemDetailToCache(ctx, id, realResp)
	// ==========================
	// 如果没有存入缓存，不应报错，打印错误日志即可
	// ==========================
	if err != nil {
		// ==========================
		// 打印error时，使用errcode.GetErrorMessageMap辅助方法
		// 可以打印包含调用栈的详细信息
		// ==========================
		log.Errorv(ctx, errcode.GetErrorMessageMap(err))
	}

	return realResp, nil
}

// 查找推荐商品详情
func (s *Service) GetVipRecommendItemDetail(ctx context.Context, id string) (*resp.VipRecommendItemDetail, error) {
	// ==========================
	// 开启事务并查询mysql数据库
	// ==========================
	res := new(entity.VipRecommendItem)
	err := s.dao.MySQL.Transaction(ctx, func(ctx context.Context, tx *sql.OrmDB) error {
		// ==========================
		// 使用当前事务拷贝至dao结构体以便调用
		// ==========================
		txDao, txErr := s.dao.Clone(
			dao.CloneOption{
				Key:   dao.CKMySQL,
				Value: tx,
			})
		if txErr != nil {
			// ==========================
			// 在首次生成error时，应当立即使用errors.Wrapf包裹
			// 外层只需直接返回error，无需再次包裹
			// ==========================
			return errors.Wrap(errcode.MysqlError, txErr.Error())
		}

		cur, err := txDao.GetVipRecommendItemByID(ctx, id)
		if err != nil {
			// ==========================
			// 在首次生成error时，应当立即使用errors.Wrapf包裹
			// 外层只需直接返回error，无需再次包裹
			// ==========================
			return err
		}
		// ==========================
		// 闭包方式赋值
		// ==========================
		res = cur

		return nil
	})

	// 将实体复制到响应模型中，通过Library中的deepcopy包实现
	detailResponse := new(resp.VipRecommendItemDetail)
	err = deepcopy.Copy(res).To(detailResponse)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return nil, errors.Wrapf(errcode.InternalError, "%s", err)
	}

	return detailResponse, err
}
