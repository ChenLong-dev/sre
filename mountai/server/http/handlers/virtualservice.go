package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/req"
	"rulai/service"
	"rulai/utils/response"
)

// DetermineBackendHost 确定是否需要更新后端服务地址
func DetermineBackendHost(c *gin.Context) {
	request := new(req.DetermineUpstreamNameReq)
	if err := c.ShouldBindJSON(request); err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	resp, err := service.SVC.DetermineUpstreamName(c, request)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	response.JSON(c, resp, nil)
}
