package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/req"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/resp"
	"gitlab.shanhai.int/sre/app-framework/internal/example/service"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/response"
)

// 用户鉴权的handler
func VerifyUser(c *gin.Context) {
	// 绑定请求
	var verifyReq req.VerifyUserReq
	err := c.ShouldBindJSON(&verifyReq)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		response.StandardJSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 将请求模型复制到http请求模型中，通过Library中的deepcopy包实现
	authReq := new(req.UserAuthVerifyReq)
	err = deepcopy.Copy(verifyReq).To(authReq)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		response.StandardJSON(c, nil, errors.Wrapf(errcode.InternalError, "%s", err))
		return
	}

	// 请求鉴权
	authResp, err := service.SVC.PostUserAuthVerify(c, authReq)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		response.StandardJSON(c, nil, err)
		return
	}

	// 将http响应模型复制到响应模型中，通过Library中的deepcopy包实现
	res := new(resp.VerifyUserResp)
	err = deepcopy.Copy(authResp).To(res)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		response.StandardJSON(c, nil, errors.Wrapf(errcode.InternalError, "%s", err))
		return
	}

	// 响应
	response.StandardJSON(c, res, nil)
}
