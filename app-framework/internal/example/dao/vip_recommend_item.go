package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/entity"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/resp"
	"gitlab.shanhai.int/sre/library/database/redis"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// vip推荐商品缓存时间
const VipRecommendItemCacheSeconds = 10 * 60

// 获取vip推荐商品详情
func (d *Dao) GetVipRecommendItemByID(ctx context.Context, id string) (item *entity.VipRecommendItem, err error) {
	item = new(entity.VipRecommendItem)

	// ==========================
	// 查询mysql只读数据库
	// ==========================
	err = d.MySQL.ReadOnlyTable(ctx, item.TableName()).
		Where("id = ?", id).
		Find(item).
		Error
	// ==========================
	// 在首次生成error时，应当立即使用errors.Wrapf包裹
	// 外层只需直接返回error，无需再次包裹
	// ==========================
	if err == gorm.ErrRecordNotFound {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MysqlError, "%s", err)
	}

	return
}

// 获取vip推荐商品缓存key
func (d *Dao) getVipRecommendItemCacheKey(id string) string {
	return fmt.Sprintf("vip_recommend_item:%s", id)
}

// 获取vip推荐商品缓存
func (d *Dao) GetVipRecommendItemDetailFromCache(ctx context.Context, id string) (*resp.VipRecommendItemDetail, error) {
	cacheResp := new(resp.VipRecommendItemDetail)

	// ==========================
	// 查询redis缓存，使用pipeline方式
	// ==========================
	err := d.Redis.WrapDo(func(con *redis.Conn) error {
		e := con.Send(ctx, "get", d.getVipRecommendItemCacheKey(id))
		if e != nil {
			// ==========================
			// 在首次生成error时，应当立即使用errors.Wrapf包裹
			// 外层只需直接返回error，无需再次包裹
			// ==========================
			e = errors.Wrapf(errcode.RedisError, "%s", e)
			return e
		}
		e = con.Flush(ctx)
		if e != nil {
			// ==========================
			// 在首次生成error时，应当立即使用errors.Wrapf包裹
			// 外层只需直接返回error，无需再次包裹
			// ==========================
			e = errors.Wrapf(errcode.RedisError, "%s", e)
			return e
		}
		reply, e := con.Receive(ctx)
		if e != nil {
			// ==========================
			// 在首次生成error时，应当立即使用errors.Wrapf包裹
			// 外层只需直接返回error，无需再次包裹
			// ==========================
			e = errors.Wrapf(errcode.RedisError, "%s", e)
			return e
		}

		// 没有找到缓存
		if reply == nil {
			// ==========================
			// 在首次生成error时，应当立即使用errors.Wrapf包裹
			// 外层只需直接返回error，无需再次包裹
			// ==========================
			e = errors.Wrapf(errcode.RedisEmptyKeyError, "%s", e)
			return e
		}

		// 解码
		e = json.Unmarshal(reply.([]byte), cacheResp)
		if e != nil {
			// ==========================
			// 在首次生成error时，应当立即使用errors.Wrapf包裹
			// 外层只需直接返回error，无需再次包裹
			// ==========================
			e = errors.Wrapf(errcode.InternalError, "%s", e)
			return e
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cacheResp, nil
}

// 保存vip推荐商品缓存
func (d *Dao) SaveVipRecommendItemDetailToCache(ctx context.Context, id string, resultResp *resp.VipRecommendItemDetail) error {
	err := d.Redis.WrapDo(func(con *redis.Conn) error {
		// 编码
		data, e := json.Marshal(resultResp)
		if e != nil {
			// ==========================
			// 在首次生成error时，应当立即使用errors.Wrapf包裹
			// 外层只需直接返回error，无需再次包裹
			// ==========================
			e = errors.Wrapf(errcode.InternalError, "%s", e)
			return e
		}

		// ==========================
		// 使用redis缓存
		// ==========================
		_, e = con.Do(ctx, "setex", d.getVipRecommendItemCacheKey(id), VipRecommendItemCacheSeconds, data)
		if e != nil {
			// ==========================
			// 在首次生成error时，应当立即使用errors.Wrapf包裹
			// 外层只需直接返回error，无需再次包裹
			// ==========================
			e = errors.Wrapf(errcode.RedisError, "%s", e)
			return e
		}

		return nil
	})

	return err
}
