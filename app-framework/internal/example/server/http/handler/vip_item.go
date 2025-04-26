package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/req"
	"gitlab.shanhai.int/sre/app-framework/internal/example/service"
	"gitlab.shanhai.int/sre/library/net/errcode"
	ginUtil "gitlab.shanhai.int/sre/library/net/gin"
	"gitlab.shanhai.int/sre/library/net/response"
)

// 获取vip商品的handler
func GetVipItemsList(c *gin.Context) {
	// 绑定请求
	var getReq req.GetVipItemListReq
	err := c.ShouldBindWith(&getReq, ginUtil.Query)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		response.StandardJSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 获取列表
	res, err := service.SVC.GetAvailableVipItemList(c, &getReq)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		response.StandardJSON(c, nil, err)
		return
	}

	// 响应
	response.StandardJSON(c, res, nil)
}
