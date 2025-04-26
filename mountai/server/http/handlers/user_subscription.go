package handlers

import (
	"rulai/models/req"
	"rulai/service"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// UserSubscribe 订阅
// nolint: dupl
func UserSubscribe(c *gin.Context) {
	subReq := new(req.UserSubscribeReq)
	err := c.ShouldBindJSON(subReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	_, err = service.SVC.GetAppDetail(c, subReq.AppID)
	if errcode.EqualError(_errcode.InvalidHexStringError, err) || errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	err = service.SVC.UserSubscribe(c, subReq, operatorID)
	if errcode.EqualError(_errcode.InvalidHexStringError, err) {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, map[string]string{
		"status": "ok",
	}, nil)
}

// UserUnsubscribe 取消订阅
// nolint: dupl
func UserUnsubscribe(c *gin.Context) {
	unsubReq := new(req.UserUnsubscribeReq)
	err := c.ShouldBindJSON(unsubReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	_, err = service.SVC.GetAppDetail(c, unsubReq.AppID)
	if errcode.EqualError(_errcode.InvalidHexStringError, err) || errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	err = service.SVC.UserUnsubscribe(c, unsubReq, operatorID)
	if errcode.EqualError(_errcode.InvalidHexStringError, err) {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, map[string]string{
		"status": "ok",
	}, nil)
}
