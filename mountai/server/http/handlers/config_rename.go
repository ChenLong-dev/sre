package handlers

import (
	"rulai/models/req"
	"rulai/service"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// CreateConfigRenamePrefix 创建特殊配置重命名前缀
func CreateConfigRenamePrefix(c *gin.Context) {
	createReq := new(req.CreateConfigRenamePrefixReq)
	err := c.ShouldBindJSON(createReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}

	err = service.SVC.CreateConfigRenamePrefix(c, createReq)
	response.JSON(c, nil, err)
}

// DeleteConfigRenamePrefix 删除特殊配置重命名前缀
func DeleteConfigRenamePrefix(c *gin.Context) {
	prefix := c.Param("prefix")
	err := service.SVC.DeleteConfigRenamePrefix(c, prefix)
	response.JSON(c, nil, err)
}
