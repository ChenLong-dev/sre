package handlers

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"rulai/models"
	"rulai/models/req"
	"rulai/service"
	_errcode "rulai/utils/errcode"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func CreateTeam(c *gin.Context) {
	createReq := new(req.CreateTeamReq)
	err := c.ShouldBindJSON(createReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	count, err := service.SVC.GetTeamsCount(c, &req.GetTeamsReq{
		Label: createReq.Label,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	if count > 0 {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "label:%s already exists", createReq.Label))
		return
	}

	team, err := service.SVC.CreateTeam(c, createReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, team, nil)
}

func UpdateTeam(c *gin.Context) {
	id := c.Param("id")
	updateReq := new(req.UpdateTeamReq)
	err := c.ShouldBindJSON(updateReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	if updateReq.Label != "" {
		list, e := service.SVC.GetTeams(c, &req.GetTeamsReq{
			Label: updateReq.Label,
		})
		if e != nil {
			response.JSON(c, nil, e)
			return
		}
		if len(list) > 0 && list[0].ID != id {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "label:%s already exists", updateReq.Label))
			return
		}
	}

	err = service.SVC.UpdateTeam(c, id, updateReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func DeleteTeam(c *gin.Context) {
	id := c.Param("id")

	projectCount, err := service.SVC.GetProjectsCount(c, &req.GetProjectsReq{
		TeamID: id,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	} else if projectCount > 0 {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "该组存在未删除的项目"))
		return
	}

	err = service.SVC.DeleteSingleTeam(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func GetTeamDetail(c *gin.Context) {
	id := c.Param("id")

	team, err := service.SVC.GetTeamDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, team, nil)
}

func GetTeams(c *gin.Context) {
	getReq := new(req.GetTeamsReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	if getReq.TeamIDs != "" {
		ids := strings.Split(getReq.TeamIDs, ",")
		getReq.IDs = ids
		getReq.TeamIDs = ""
	}

	teams, err := service.SVC.GetTeams(c, getReq)
	if errcode.EqualError(_errcode.InvalidHexStringError, err) {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	count, err := service.SVC.GetTeamsCount(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, models.BaseListResponse{
		List:  teams,
		Limit: getReq.Limit,
		Page:  getReq.Page,
		Count: count,
	}, nil)
}

func CheckTeam(ctx *gin.Context) {
	id := ctx.Param("id")

	if id == "" {
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		ctx.Abort()
		response.JSON(ctx, nil, errors.Wrap(_errcode.NotFoundError, err.Error()))

		return
	}

	_, err = service.SVC.GetTeamByObjectID(ctx, objectID)

	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		ctx.Abort()
		response.JSON(ctx, nil, errors.Wrap(_errcode.NotFoundError, "team id 不存在"))

		return
	}
	if err != nil {
		ctx.Abort()
		response.JSON(ctx, nil, err)

		return
	}
}
