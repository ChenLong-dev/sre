package handlers

import (
	"rulai/models/req"
	"rulai/service"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/response"
)

// 给项目创建Jenkins跑CI流程的job，并创建gitlab钩子触发器
func CreateProjectCIJob(c *gin.Context) {
	createReq := new(req.CreateProjectCIJobReq)
	err := c.ShouldBindJSON(createReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	id := c.Param("project_id")
	project, err := service.SVC.GetProjectDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	for _, branchName := range createReq.DeployBranchName {
		if branchName == "" {
			continue
		}
		_, err = service.SVC.GetGitlabSingleProjectBranch(c, id, branchName)
		if err != nil {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
			return
		}
	}

	err = service.SVC.CreateProjectCIJob(c, id, project.Name, createReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func GetProjectCIJob(c *gin.Context) {
	id := c.Param("project_id")
	_, err := service.SVC.GetProjectDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	getResp, err := service.SVC.GetCIJobDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, getResp, nil)
}

func UpdateProjectCIJob(c *gin.Context) {
	id := c.Param("project_id")
	updateReq := new(req.UpdateProjectCIJobReq)
	err := c.ShouldBindJSON(updateReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	for _, branchName := range updateReq.DeployBranchName {
		if branchName == "" {
			continue
		}
		_, err = service.SVC.GetGitlabSingleProjectBranch(c, id, branchName)
		if err != nil {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
			return
		}
	}

	ciJob, err := service.SVC.GetCIJobDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	project, err := service.SVC.GetProjectDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	err = service.SVC.UpdateProjectCIJob(c, id, project.Name, ciJob.ID, updateReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}
