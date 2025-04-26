package handlers

import (
	"rulai/models"
	"rulai/models/req"
	"rulai/service"
	"rulai/utils"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// CreateFavProject 创建收藏
func CreateFavProject(c *gin.Context) {
	createReq := new(req.CreateFavProjectReq)
	err := c.ShouldBindJSON(createReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	_, err = service.SVC.GetProjectDetail(c, createReq.ProjectID)
	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 获取操作人id
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	isFav, err := service.SVC.IsProjectFav(c, createReq.ProjectID, operatorID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if isFav {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "project already favorite"))
		return
	}

	res, err := service.SVC.CreateFavProject(c, createReq, operatorID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, res, nil)
}

// DeleteFavProject 删除收藏
func DeleteFavProject(c *gin.Context) {
	id := c.Param("projectID")

	// 获取操作人id
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	err := service.SVC.DeleteFavProject(c, operatorID, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

// GetFavProjects 获取收藏项目列表
func GetFavProjects(c *gin.Context) {
	getReq := new(req.GetFavProjectReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	if getReq.UserID == "" {
		// 获取操作人id
		operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
		if !ok {
			response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
			return
		}

		getReq.UserID = operatorID
	}

	res, err := service.SVC.GetFavProjects(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	count, err := service.SVC.CountFavProject(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
	}

	response.JSON(c, models.BaseListResponse{
		List:  res,
		Limit: getReq.Limit,
		Page:  getReq.Page,
		Count: count,
	}, nil)
}
