package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/req"
	"gitlab.shanhai.int/sre/app-framework/internal/example/service"
	"gitlab.shanhai.int/sre/library/net/errcode"
	ginUtil "gitlab.shanhai.int/sre/library/net/gin"
	"gitlab.shanhai.int/sre/library/net/response"
)

// 获取github用户仓库handler
func GetOwnerGithubRepos(c *gin.Context) {
	// ==========================
	// 绑定请求
	// 使用v9校验器校验请求：https://godoc.org/gopkg.in/go-playground/validator.v9
	// ==========================
	var getReq req.GetGithubRepositoryListReq
	getReq.Owner = c.Param("owner")
	err := c.ShouldBindWith(&getReq, ginUtil.Query)
	if err != nil {
		response.StandardJSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 获取列表
	res, err := service.SVC.GetOwnerGithubReposList(c, &getReq)
	if err != nil {
		response.StandardJSON(c, nil, err)
		return
	}

	// 响应
	response.StandardJSON(c, res, nil)
}

// 获取github用户聚合数据handler
func GetOwnerGithubAggregation(c *gin.Context) {
	// 获取owner
	owner := c.Param("owner")
	if owner == "" {
		response.StandardJSON(c, nil, errors.Wrap(errcode.InvalidParams, fmt.Sprintf("参数不合法:owner")))
		return
	}

	// 获取详情
	res, err := service.SVC.GetOwnerGithubAggregationResp(c, owner)
	if err != nil {
		response.StandardJSON(c, nil, err)
		return
	}

	// 响应
	response.StandardJSON(c, res, nil)
}
