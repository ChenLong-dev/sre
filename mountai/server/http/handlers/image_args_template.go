package handlers

import (
	"rulai/models"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"rulai/models/entity"
	"rulai/models/req"

	"context"

	"rulai/service"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func GetImageArgsTemplates(c *gin.Context) {
	getReq := new(req.GetImageArgsTemplateReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	res, count, err := service.SVC.GetImageArgsTemplates(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, models.BaseListResponse{
		List:  res,
		Limit: getReq.Limit,
		Page:  getReq.Page,
		Count: count,
	}, nil)
}

func UpdateImageArgsTemplate(c *gin.Context) {
	imageArgsTemplateID := c.Param("image_args_template_id")

	imageArgsTemplate, err := getAndValidateImageArgsTemplatePermission(c, imageArgsTemplateID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	updateReq := new(req.UpdateImageArgsTemplateReq)
	err = c.ShouldBindJSON(updateReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	if updateReq.Name != "" && updateReq.Name != imageArgsTemplate.Name {
		unique, intErr := service.SVC.CheckImageArgsTemplateNameUnique(c, &req.GetImageArgsTemplateReq{
			TeamID: imageArgsTemplate.TeamID,
			Name:   updateReq.Name,
		})
		if intErr != nil {
			response.JSON(c, nil, intErr)
			return
		}

		if !unique {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "名称不唯一"))
			return
		}
	}

	err = service.SVC.UpdateSingleImageArgsTemplate(c, imageArgsTemplateID, updateReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func CreateImageArgsTemplate(c *gin.Context) {
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	createReq := new(req.CreateImageArgsTemplateReq)
	err := c.ShouldBindJSON(createReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	unique, err := service.SVC.CheckImageArgsTemplateNameUnique(c, &req.GetImageArgsTemplateReq{
		TeamID: createReq.TeamID,
		Name:   createReq.Name,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if !unique {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "名称不唯一"))
		return
	}

	createReq.OperatorID = operatorID

	err = service.SVC.CreateSingleImageArgsTemplate(c, createReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func DeleteImageArgsTemplate(c *gin.Context) {
	imageArgsTemplateID := c.Param("image_args_template_id")

	_, err := getAndValidateImageArgsTemplatePermission(c, imageArgsTemplateID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	err = service.SVC.DeleteSingleImageArgsTemplateByID(c, imageArgsTemplateID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func getAndValidateImageArgsTemplatePermission(c context.Context, imageArgsTemplateID string) (*entity.ImageArgsTemplate, error) {
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		return nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid")
	}

	imageArgsTemplate, err := service.SVC.FindSingleImageArgsTemplateByID(c, imageArgsTemplateID)
	if err != nil {
		return nil, err
	}

	if imageArgsTemplate == nil {
		return nil, errors.Wrap(errcode.InvalidParams, "该模版不存在")
	}

	if operatorID != imageArgsTemplate.OwnerID {
		return nil, errors.Wrapf(_errcode.GitlabUserNoPermissionError, "没有权限操作不属于自己的模版")
	}

	return imageArgsTemplate, nil
}

func CheckImageArgsTemplate(ctx *gin.Context) {
	id := ctx.Param("image_args_template_id")

	if id == "" {
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		ctx.Abort()
		response.JSON(ctx, nil, errors.Wrap(_errcode.NotFoundError, err.Error()))

		return
	}

	_, err = service.SVC.GetImageArgsTemplateByObjectID(ctx, objectID)

	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		ctx.Abort()
		response.JSON(ctx, nil, errors.Wrap(_errcode.NotFoundError, "image_args_template id 不存在"))

		return
	}
	if err != nil {
		ctx.Abort()
		response.JSON(ctx, nil, err)

		return
	}
}
