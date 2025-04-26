package v

import (
	"rulai/models/req"
	reqV2 "rulai/models/req/v2"
	"rulai/models/resp"
	respV2 "rulai/models/resp/v2"

	"rulai/service"
	"rulai/utils/response"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func GetRunningStatusList(c *gin.Context) {
	getReq := new(reqV2.GetRunningStatusListReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 校验集群名
	getReq.ClusterName, err = service.SVC.CheckAndUnifyClusterName(c, getReq.ClusterName, getReq.EnvName)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	appIDs := strings.Split(getReq.AppIDs, ",")

	project, err := service.SVC.GetProjectDetail(c, getReq.ProjectID)
	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	apps, err := service.SVC.GetAppsDetails(c, &req.GetAppsReq{
		IDs:       appIDs,
		ProjectID: getReq.ProjectID,
	})

	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	res := make([]*respV2.AppRunningStatusListResp, len(apps))

	for idx, app := range apps {
		var ret []*resp.RunningStatusListResp
		v1GetReq := &req.GetRunningStatusListReq{
			AppID:       app.ID,
			EnvName:     getReq.EnvName,
			ClusterName: getReq.ClusterName,
			Namespace:   getReq.Namespace,
		}
		ret, err = service.SVC.GetRunningStatusJobList(c, getReq.ClusterName, v1GetReq, project, app)
		if err != nil {
			response.JSON(c, nil, err)
			return
		}

		res[idx] = &respV2.AppRunningStatusListResp{
			AppID:         app.ID,
			RunningStatus: make([]*resp.RunningStatusDetailResp, len(ret)),
		}

		for i, runningStatus := range ret {
			detail, e := service.SVC.GetRunningStatusDetail(c, &req.GetRunningStatusDetailReq{
				AppID:     app.ID,
				EnvName:   getReq.EnvName,
				Version:   runningStatus.Version,
				Namespace: getReq.Namespace,
			}, app)

			if e != nil {
				response.JSON(c, nil, e)
				return
			}

			res[idx].RunningStatus[i] = detail
		}
	}

	response.JSON(c, res, nil)
}
