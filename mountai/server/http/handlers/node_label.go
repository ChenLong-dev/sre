package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/req"
	"rulai/service"
	"rulai/utils/response"
)

// GetNodeLabelLists : 获取支持的节点标签列表(用于ams前端选择框显示)
func GetNodeLabelLists(c *gin.Context) {
	getReq := new(req.GetNodeLabelListReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	res, err := service.SVC.GetNodeLabelLists(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, res, nil)
}
