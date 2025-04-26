package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/null"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/config"
	"rulai/models"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/service"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
	"rulai/utils/response"
)

func CreateProject(c *gin.Context) {
	createReq := new(req.CreateProjectReq)
	err := c.ShouldBindJSON(createReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 校验名称唯一
	err = service.SVC.CheckProjectNameLegal(c, createReq.Name)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if err = service.SVC.CheckProjectLabelsLegal(createReq.Labels); err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 创建jenkins ci流程job
	// if !createReq.DisableCI {
	// 	err = service.SVC.CreateProjectCIJob(c, createReq.GitID, createReq.Name, &req.CreateProjectCIJobReq{
	// 		MessageNotification: []entity.NotificationType{entity.NotificationTypeDingDing},
	// 		PipelineStages:      entity.DefaultPipelineStages,
	// 	})
	// 	if err != nil {
	// 		response.JSON(c, nil, err)
	// 		return
	// 	}
	// }

	project, err := service.SVC.CreateProject(c, createReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// HACK: 临时措施, 新项目默认添加多集群白名单
	err = service.SVC.SetMultiClusterSupportForProject(c, createReq.GitID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, project, nil)
}

func UpdateProject(c *gin.Context) {
	id := c.Param("project_id")
	updateReq := new(req.UpdateProjectReq)
	err := c.ShouldBindJSON(updateReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 校验团队
	if updateReq.TeamID != "" {
		_, err = service.SVC.GetTeamDetail(c, updateReq.TeamID)
		if errcode.EqualError(_errcode.InvalidHexStringError, err) || errcode.EqualError(errcode.NoRowsFoundError, err) {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
			return
		}
		if err != nil {
			response.JSON(c, nil, err)
			return
		}
	}

	// 校验负责人
	if len(updateReq.OwnerIDs) != 0 {
		ownersCount, e := service.SVC.GetUsersCount(c, &req.GetUsersReq{UserIDs: updateReq.OwnerIDs})
		if e != nil {
			response.JSON(c, nil, errors.Wrap(errcode.InternalError, e.Error()))
			return
		}

		if len(updateReq.OwnerIDs) != ownersCount {
			response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "contains invalid owner id"))
			return
		}
	}

	if err = service.SVC.CheckProjectLabelsLegal(updateReq.Labels); err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 新增进行中任务,不允许修改项目信息
	// 检查是否有其他未完成任务, 所有集群同时只能有一个任务在执行, 否则域名解析可能会有问题
	apps, err := service.SVC.GetApps(c, &req.GetAppsReq{
		ProjectID: id,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	appIDs := make([]string, 0, len(apps))
	for _, app := range apps {
		appIDs = append(appIDs, app.ID)
	}

	unFinishCount, err := service.SVC.GetTasksCount(c, &req.GetTasksReq{
		EnvName:           entity.AppEnvName(config.Conf.Env),
		StatusInverseList: entity.TaskStatusFinalStateList,
		Suspend:           null.BoolFrom(false),
		AppIDList:         appIDs,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if unFinishCount > 0 {
		response.JSON(c, nil, errors.Wrap(_errcode.OtherRunningTaskExistsError, " may not change project"))
		return
	}

	oldProject, err := service.SVC.GetProjectDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if oldProject.Team.ID == updateReq.TeamID {
		updateReq.TeamID = ""
	}

	// 更新mongoDB
	err = service.SVC.UpdateProject(c, id, updateReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 更新teamID则更新相关内容
	if updateReq.TeamID != "" {
		projectDetail, err := service.SVC.GetProjectDetail(c, id)
		if err != nil {
			response.JSON(c, nil, err)
			return
		}

		err = updateProjectTeam(c, projectDetail, apps, oldProject.Team)
		if err != nil {
			response.JSON(c, nil, err)
			return
		}
	}

	response.JSON(c, nil, nil)
}

func updateProjectTeam(ctx context.Context,
	projectDetail *resp.ProjectDetailResp, apps []*resp.AppListResp, oldTeam *resp.TeamDetailResp) error {
	for _, app := range apps {
		appDetail, err := service.SVC.GetAppDetail(ctx, app.ID)
		if err != nil {
			return err
		}

		if appDetail.SentryProjectSlug != "" {
			// 更新sentry钉钉
			err = service.SVC.CheckAndUpdateProjectSentryDingDing(ctx, projectDetail, appDetail)
			if err != nil {
				return err
			}

			// 更新sentry团队
			err = service.SVC.CheckAndUpdateProjectSentryTeam(ctx, projectDetail, appDetail, oldTeam.SentrySlug)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func DeleteProject(c *gin.Context) {
	id := c.Param("project_id")

	// 获取操作人id
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	// 检查是否有权限删除项目
	err := service.SVC.ValidateHasPermission(c, &req.ValidateHasPermissionReq{
		OperateType: entity.OperateTypeDeleteProject,
		ProjectID:   id,
		OperatorID:  operatorID,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	appCount, err := service.SVC.GetAppsCount(c, &req.GetAppsReq{
		ProjectID: id,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	} else if appCount > 0 {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "该项目存在未删除的应用"))
		return
	}

	project, err := service.SVC.GetProjectDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	err = service.SVC.DeleteProjectCIJob(c, id, project.Name)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	err = service.SVC.DeleteLogStoresByProject(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 删除项目收藏
	err = service.SVC.DeleteFavProjectByProjectID(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	err = service.SVC.DeleteSingleProject(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func GetProjectDetail(c *gin.Context) {
	id := c.Param("project_id")

	project, err := service.SVC.GetProjectDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	project.IsFav, err = service.SVC.IsProjectFav(c, project.ID, operatorID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, project, nil)
}

func GetProjects(c *gin.Context) {
	getReq := new(req.GetProjectsReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	if getReq.ProjectIDs != "" {
		ids := strings.Split(getReq.ProjectIDs, ",")
		getReq.IDs = ids
		getReq.ProjectIDs = ""
	}

	projects, err := service.SVC.GetProjects(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	count, err := service.SVC.GetProjectsCount(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	for _, project := range projects {
		project.IsFav, err = service.SVC.IsProjectFav(c, project.ID, operatorID)
		if err != nil {
			response.JSON(c, nil, err)
			return
		}
	}

	response.JSON(c, models.BaseListResponse{
		List:  projects,
		Limit: getReq.Limit,
		Page:  getReq.Page,
		Count: count,
	}, nil)
}

func GetProjectConfig(c *gin.Context) {
	id := c.Param("project_id")
	getReq := new(req.GetProjectConfigReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	project, err := service.SVC.GetProjectDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	isDecrypt := false
	if getReq.EnvName == entity.AppEnvStg || getReq.EnvName == entity.AppEnvFat {
		isDecrypt = true
	}

	formatType := req.ConfigManagerFormatTypeYaml
	if getReq.FormatType != "" {
		formatType = getReq.FormatType
	}

	// 项目配置应当全量返回, 不限定使用的特殊替换前缀
	getConfigReq := &req.GetConfigManagerFileReq{
		ProjectID:   project.ID,
		ProjectName: project.Name,
		EnvName:     getReq.EnvName,
		CommitID:    "",
		IsDecrypt:   isDecrypt,
		FormatType:  formatType,
	}

	appConfig, err := service.SVC.GetAppConfig(c, getConfigReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if formatType == req.ConfigManagerFormatTypeJSON {
		response.JSON(c, appConfig.Config, nil)
		return
	}
	c.String(http.StatusOK, "%s", appConfig.Config)
}

func GetProjectResource(c *gin.Context) {
	id := c.Param("project_id")
	getReq := new(req.GetProjectResourceReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	project, err := service.SVC.GetProjectDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	res := new(resp.ProjectResourceResp)
	// 从配置文件中获取资源信息
	resourceResp, err := service.SVC.GetProjectResourceFromConfig(c, &req.GetProjectResourceFromConfigReq{
		EnvName: getReq.EnvName,
	}, project)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	res.GetProjectResourceFromConfigResp = resourceResp

	// 获取版本信息
	versionResp, err := service.SVC.GetQTFrameworkVersion(c, id, "master")
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	res.Library = versionResp

	response.JSON(c, res, nil)
}

// GetProjectUserRole 获取项目下用户的权限信息
func GetProjectUserRole(c *gin.Context) {
	projectID := c.Param("project_id")

	_, err := service.SVC.GetProjectDetail(c, projectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	userID := c.Param("user_id")

	_, err = service.SVC.GetUserInfo(c, userID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	res, err := service.SVC.GetProjectUserRole(c, projectID, userID)

	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, res, nil)
}

// GetProjectAppsClustersWithWorkload 批量获取项目下多个应用各自在指定环境下有工作负载的集群列表
func GetProjectAppsClustersWithWorkload(c *gin.Context) {
	projectID := c.Param("project_id")
	getReq := new(req.GetProjectAppsClustersWithWorkloadReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}

	getAppsReq := &req.GetAppsReq{IDs: strings.Split(getReq.AppIDs, ",")}
	if len(getAppsReq.IDs) > 50 {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "app_ids' length is more than 50"))
		return
	}

	apps, err := service.SVC.GetApps(c, getAppsReq) // query 参数 app_ids 为逗号分隔
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 校验 project_id
	for i := range apps {
		if apps[i].ProjectID != projectID {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "app(%s) is not under project(%s)", apps[i].ID, projectID))
			return
		}
	}

	project, err := service.SVC.GetProjectDetail(c, projectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	appClusters, err := service.SVC.GetProjectAppsClustersWithWorkload(c, getReq.EnvName, project.ID, getAppsReq.IDs)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 没有分页逻辑, 分页参数只需要在返回前处理即可
	res := &models.BaseListResponse{
		Limit: 50,
		Page:  1,
		Count: len(appClusters),
		List:  appClusters,
	}

	response.JSON(c, res, nil)
}

func CheckProject(ctx *gin.Context) {
	id := ctx.Param("project_id")

	if id == "" {
		return
	}

	_, err := service.SVC.GetProjectByID(ctx, id)

	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		ctx.Abort()
		response.JSON(ctx, nil, errors.Wrap(_errcode.NotFoundError, "project id 不存在"))

		return
	}
	if err != nil {
		ctx.Abort()
		response.JSON(ctx, nil, err)

		return
	}
}
