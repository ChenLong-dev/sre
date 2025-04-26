package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/req"
	"gitlab.shanhai.int/sre/app-framework/internal/example/service"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/response"
)

// 创建会员退款的handler
func CreateVipRefund(c *gin.Context) {
	// 绑定请求
	var createReq req.CreateVipRefundReq
	err := c.ShouldBindJSON(&createReq)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		response.StandardJSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 创建资源
	err = service.SVC.CreateRefundNotice(c, &createReq)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		response.StandardJSON(c, nil, err)
		return
	}

	// 响应
	response.StandardJSON(c, nil, nil)
}
