package handlers

import (
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/service"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func GetGitProjectDetail(c *gin.Context) {
	gitID := c.Param("id")

	repoInfo, err := service.SVC.GetGitlabSingleProject(c, gitID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, repoInfo, nil)
}

func GetGitProjectBranches(c *gin.Context) {
	gitID := c.Param("id")
	getReq := new(req.GetRepoBranchesReq)

	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	var branches []*resp.GitBranchResp

	if getReq.Keyword == "" && getReq.Page == 0 && getReq.PageSize == 0 {
		branches, err = service.SVC.GetGitlabSingleProjectAllBranches(c, gitID)
	} else {
		branches, err = service.SVC.GetGitlabSingleProjectBranches(c, gitID, getReq)
	}

	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, branches, nil)
}
