package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/service"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/response"
)

// 获取vip推荐商品的handler
func GetVipRecommendItemDetail(c *gin.Context) {
	// 获取id
	itemID := c.Param("item_id")
	if itemID == "" {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		response.StandardJSON(c, nil, errors.Wrap(errcode.InvalidParams, fmt.Sprintf("参数不合法:item_id")))
		return
	}

	// 获取详情
	result, err := service.SVC.GetVipRecommendItemCacheDetail(c, itemID)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		response.StandardJSON(c, nil, err)
		return
	}

	response.StandardJSON(c, result, nil)
}
