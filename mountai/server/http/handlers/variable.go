package handlers

import (
	"rulai/models"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/service"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
	"rulai/utils/response"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func GetVariables(c *gin.Context) {
	getReq := new(req.GetVariablesReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	hasPermission := false

	if getReq.Type == entity.ProjectVariableType {
		if getReq.ProjectID == "" {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "project id 不能为空"))
			return
		}

		intErr := service.SVC.ValidateHasPermission(c, &req.ValidateHasPermissionReq{
			OperateType: entity.OperateTypeReadVariableValue,
			ProjectID:   getReq.ProjectID,
			OperatorID:  operatorID,
		})

		hasPermission = intErr == nil
	} else {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "不支持type为%d的变量", getReq.Type))
		return
	}

	variables, count, err := service.SVC.GetVariables(c, getReq, !hasPermission)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, models.BaseListResponse{
		List:  variables,
		Limit: getReq.Limit,
		Page:  getReq.Page,
		Count: count,
	}, nil)
}

func UpdateVariable(c *gin.Context) {
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	variableID := c.Param("variable_id")

	variable, err := service.SVC.FindSingleVariableByID(c, variableID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if variable.Type == entity.ProjectVariableType {
		intErr := service.SVC.ValidateHasPermission(c, &req.ValidateHasPermissionReq{
			OperateType: entity.OperateTypeUpdateVariableValue,
			ProjectID:   variable.ProjectID,
			OperatorID:  operatorID,
		})
		if intErr != nil {
			response.JSON(c, nil, intErr)
			return
		}
	} else {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "不支持type为%d的变量", variable.Type))
		return
	}

	updateReq := new(req.UpdateVariableReq)
	err = c.ShouldBindJSON(updateReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	if updateReq.Key != "" && variable.Key != updateReq.Key {
		unique, intErr := service.SVC.CheckVariableKeyUnique(c, &req.GetVariablesReq{
			Key:       updateReq.Key,
			Type:      variable.Type,
			ProjectID: variable.ProjectID,
		})
		if intErr != nil {
			response.JSON(c, nil, intErr)
			return
		}

		if !unique {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "key 不唯一"))
			return
		}
	}

	updateReq.OperatorID = operatorID

	err = service.SVC.UpdateSingleVariable(c, variableID, updateReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func CreateVariable(c *gin.Context) {
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	createReq := new(req.CreateVariableReq)
	err := c.ShouldBindJSON(createReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	if createReq.Type == entity.ProjectVariableType {
		projectID := createReq.ProjectID
		if projectID == "" {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "project id 不能为空"))
			return
		}

		intErr := service.SVC.ValidateHasPermission(c, &req.ValidateHasPermissionReq{
			OperateType: entity.OperateTypeCreateVariableValue,
			ProjectID:   projectID,
			OperatorID:  operatorID,
		})
		if intErr != nil {
			response.JSON(c, nil, intErr)
			return
		}
	} else {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "不支持type为%d的变量", createReq.Type))
		return
	}

	unique, err := service.SVC.CheckVariableKeyUnique(c, &req.GetVariablesReq{
		Key:       createReq.Key,
		Type:      createReq.Type,
		ProjectID: createReq.ProjectID,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if !unique {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "key 不唯一"))
		return
	}

	createReq.OperatorID = operatorID

	variable, err := service.SVC.CreateSingleVariable(c, createReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, variable, nil)
}

func DeleteVariable(c *gin.Context) {
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	variableID := c.Param("variable_id")

	variable, err := service.SVC.FindSingleVariableByID(c, variableID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if variable.Type == entity.ProjectVariableType {
		intErr := service.SVC.ValidateHasPermission(c, &req.ValidateHasPermissionReq{
			OperateType: entity.OperateTypeDeleteVariableValue,
			ProjectID:   variable.ProjectID,
			OperatorID:  operatorID,
		})
		if intErr != nil {
			response.JSON(c, nil, intErr)
			return
		}
	} else {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "不支持type为%d的变量", variable.Type))
		return
	}

	err = service.SVC.DeleteSingleVariableByID(c, variableID, operatorID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func CheckVariable(ctx *gin.Context) {
	id := ctx.Param("variable_id")

	if id == "" {
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		ctx.Abort()
		response.JSON(ctx, nil, errors.Wrap(_errcode.NotFoundError, err.Error()))

		return
	}

	_, err = service.SVC.GetVariableByObjectID(ctx, objectID)

	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		ctx.Abort()
		response.JSON(ctx, nil, errors.Wrap(_errcode.NotFoundError, "variable id 不存在"))

		return
	}
	if err != nil {
		ctx.Abort()
		response.JSON(ctx, nil, err)

		return
	}
}
