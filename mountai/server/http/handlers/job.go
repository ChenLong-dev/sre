package handlers

import (
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/service"
	"rulai/utils"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func DeleteJob(c *gin.Context) {
	request := new(req.DeleteJob)
	if err := c.ShouldBindJSON(request); err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 获取操作人id
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	app, err := service.SVC.GetAppDetail(c, request.AppId)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 检查是否有权限删除应用
	err = service.SVC.ValidateHasPermission(c, &req.ValidateHasPermissionReq{
		OperateType: entity.OperateTypeDeleteJob,
		ProjectID:   app.ProjectID,
		OperatorID:  operatorID,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	pod, err := service.SVC.GetPodDetail(c, entity.ClusterName(request.Clusterame), &req.GetPodDetailReq{
		Namespace: request.Namespace,
		Env:       request.EnvName,
		Name:      request.Podname,
	})

	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if len(pod.OwnerReferences) == 0 || pod.OwnerReferences[0].Kind != "Job" {
		response.JSON(c, nil, errcode.InternalError)
		return
	}

	service.SVC.DeleteJob(c, entity.ClusterName(request.Clusterame), &req.DeleteJobReq{
		Namespace: request.Namespace,
		Env:       request.EnvName,
		Name:      pod.OwnerReferences[0].Name,
	})

	response.JSON(c, nil, nil)
}
