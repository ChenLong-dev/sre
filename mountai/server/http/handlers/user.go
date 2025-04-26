package handlers

import (
	"rulai/models"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/service"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func Login(c *gin.Context) {
	postReq := new(req.UserAuthLoginReq)
	err := c.ShouldBindJSON(postReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	authResp, err := service.SVC.AuthGitlabToken(c, postReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	profile, err := service.SVC.GetGitlabUserProfile(c, authResp.AccessToken)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InternalError, err.Error()))
		return
	} else if profile.State != "active" {
		response.JSON(c, nil, _errcode.GitlabUserStateError)
		return
	}

	token, err := utils.GenerateJWTToken(c, profile)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	user, err := service.SVC.SaveUserInfo(c, profile, token)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, user, nil)
}

func GetRecentlyProjects(c *gin.Context) {
	// 获取操作人id
	userID := c.Param("user_id")

	currentUserID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	getTaskReq, err := service.SVC.ConvertToActiveTasksReq(c, &req.GetActivitiesReq{
		BaseListRequest: models.BaseListRequest{
			Limit: 20,
			Page:  1,
		},
		OperatorID: userID,
	}, currentUserID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	tasks, err := service.SVC.GetRecentlyActiveTasks(c, getTaskReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 转换结构体
	projectSet := make(map[string]bool)
	res := make([]*resp.ActiveProjectResp, 0)
	for _, task := range tasks {
		if _, ok := projectSet[task.ProjectID]; ok {
			continue
		}
		projectSet[task.ProjectID] = true
		res = append(res, &resp.ActiveProjectResp{
			ID:             task.ProjectID,
			Name:           task.ProjectName,
			Desc:           task.ProjectDesc,
			TeamID:         task.TeamID,
			TeamName:       task.TeamName,
			TaskCreateTime: task.CreateTime,
		})
	}

	response.JSON(c, res, nil)
}

func GetActivities(c *gin.Context) {
	getReq := new(req.GetActivitiesReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	currentUserID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	getTaskReq, err := service.SVC.ConvertToActiveTasksReq(c, getReq, currentUserID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if getTaskReq.NoNeedQuery {
		response.JSON(c, models.BaseListResponse{
			List:  make([]*resp.ActiveTaskResp, 0),
			Limit: getReq.Limit,
			Page:  getReq.Page,
			Count: 0,
		}, nil)
		return
	}

	tasks, err := service.SVC.GetRecentlyActiveTasks(c, getTaskReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	count, err := service.SVC.GetRecentlyActiveTasksCount(c, getTaskReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, models.BaseListResponse{
		List:  tasks,
		Limit: getReq.Limit,
		Page:  getReq.Page,
		Count: count,
	}, nil)
}

func GetUsers(c *gin.Context) {
	getReq := new(req.GetUsersReq)
	err := c.ShouldBindQuery(getReq)

	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	users, err := service.SVC.GetUsers(c, getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InternalError, err.Error()))
		return
	}

	count, err := service.SVC.GetUsersCount(c, getReq)

	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, models.BaseListResponse{
		List:  users,
		Limit: getReq.Limit,
		Page:  getReq.Page,
		Count: count,
	}, nil)
}

func CheckUser(ctx *gin.Context) {
	id := ctx.Param("user_id")

	if id == "" {
		return
	}

	_, err := service.SVC.GetUserByID(ctx, id)

	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		ctx.Abort()
		response.JSON(ctx, nil, errors.Wrap(_errcode.NotFoundError, "user id 不存在"))

		return
	}
	if err != nil {
		ctx.Abort()
		response.JSON(ctx, nil, err)

		return
	}
}
