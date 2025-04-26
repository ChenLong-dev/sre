package handlers

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/goroutine"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson/primitive"

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

// CreateApp 创建应用
func CreateApp(c *gin.Context) {
	createReq := new(req.CreateAppReq)
	err := c.ShouldBindJSON(createReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 校验应用类型
	if createReq.Type == entity.AppTypeService && createReq.ServiceType == "" {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "service type is empty"))
		return
	}
	if createReq.Type != entity.AppTypeService && createReq.ServiceType != "" {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "service type should empty"))
		return
	}

	if createReq.Type == entity.AppTypeService {
		if createReq.ServiceExposeType == "" {
			// 默认通过 Ingress 方式暴露服务
			createReq.ServiceExposeType = entity.AppServiceExposeTypeIngress
		}
		// grpc服务不允许使用非ingress类型暴露方式
		if createReq.ServiceType == entity.AppServiceTypeGRPC &&
			createReq.ServiceExposeType != entity.AppServiceExposeTypeIngress {
			response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "service expose type should be Ingress"))
			return
		}
	}

	// 校验应用描述长度
	if len(createReq.Description) > req.MaxDescriptionLength {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "description is too long"))
		return
	}

	// 校验应用
	project, err := service.SVC.GetProjectDetail(c, createReq.ProjectID)
	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 校验名称合法性
	err = service.SVC.CheckAppNameLegal(c, project, createReq.Type, createReq.Name)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 限制单个project下最大脚本应用数为20
	if createReq.Type == entity.AppTypeOneTimeJob {
		jobCount, e := service.SVC.GetAppsCount(c, &req.GetAppsReq{
			ProjectID: createReq.ProjectID,
			Type:      entity.AppTypeOneTimeJob,
		})
		if e != nil {
			response.JSON(c, nil, e)
			return
		}
		if jobCount >= config.Conf.OneTimeJobMaxCount {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "job count is over the limit. Please reuse or delete finished job"))
			return
		}
	}

	app, err := service.SVC.CreateApp(c, createReq, project)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, app, nil)
}

// UpdateApp 更新应用
func UpdateApp(c *gin.Context) {
	id := c.Param("id")
	updateReq := new(req.UpdateAppReq)
	err := c.ShouldBindJSON(updateReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 校验应用描述长度
	if len(updateReq.Description.ValueOrZero()) > req.MaxDescriptionLength {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "description is too long"))
		return
	}

	_, err = service.SVC.GetAppDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	err = service.SVC.UpdateApp(c, id, updateReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

// DeleteApp 删除应用
func DeleteApp(c *gin.Context) {
	id := c.Param("id")
	deleteReq := new(req.DeleteAppReq)
	// FIXME: remove when front support delete sentry check
	notEmpty, err := utils.IsRequestBodyNotEmpty(c)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	if notEmpty {
		if err = c.ShouldBindJSON(deleteReq); err != nil {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
			return
		}
	}

	// 获取操作人id
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	app, err := service.SVC.GetAppDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 检查是否有权限删除应用
	err = service.SVC.ValidateHasPermission(c, &req.ValidateHasPermissionReq{
		OperateType: entity.OperateTypeDeleteApp,
		ProjectID:   app.ProjectID,
		OperatorID:  operatorID,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	project, err := service.SVC.GetProjectDetail(c, app.ProjectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 获取当前项目支持的集群，不支持则不需清理
	clusters, err := service.SVC.GetProjectSupportedClusters(c, project.ID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	for _, cluster := range clusters {
		err = service.SVC.CheckRunningStatusExists(c, cluster, project, app)
		if err != nil {
			response.JSON(c, nil, err)
			return
		}
	}

	// 先删除
	err = service.SVC.DeleteAppTasks(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	err = service.SVC.DeleteSingleApp(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	for envName, env := range app.Env {
		for _, cluster := range clusters {
			// 集群区分环境
			if cluster.Env != envName {
				continue
			}

			serviceName := app.ServiceName

			if app.ServiceType == entity.AppServiceTypeRestful && app.ServiceExposeType == entity.AppServiceExposeTypeIngress {
				serviceName, err = service.SVC.GetCurrentServiceName(c, cluster.Name, envName, app)
				if err != nil {
					log.Errorv(c, errcode.GetErrorMessageMap(err))
					continue
				}
			}

			_, err = service.SVC.CreateTask(c, project, app, &req.CreateTaskReq{
				AppID:       app.ID,
				ClusterName: cluster.Name,
				EnvName:     envName,
				Action:      entity.TaskActionClean,
				Param: &req.CreateTaskParamReq{
					CleanedProjectName:          project.Name,
					CleanedAppName:              app.Name,
					CleanedAppType:              app.Type,
					CleanedAppServiceType:       app.ServiceType,
					CleanedAppServiceExposeType: app.ServiceExposeType,
					CleanedServiceName:          serviceName,
					CleanedAliAlarmName:         env.AliAlarmName,
					CleanedAliLogConfigName:     app.AliLogConfigName,
				},
			}, operatorID)
			if err != nil {
				log.Errorv(c, errcode.GetErrorMessageMap(err))
				continue
			}
		}
	}

	// 删除Sentry项目
	if deleteReq.DeleteSentry && app.SentryProjectSlug != "" {
		err = service.SVC.DeleteSentryProject(c, &req.DeleteSentryProjectReq{
			ProjectSlug: app.SentryProjectSlug,
		})
		if err != nil {
			log.Errorv(c, errcode.GetErrorMessageMap(err))
		}
	}

	response.JSON(c, nil, nil)
}

// GetAppDetail 获取应用详情
func GetAppDetail(c *gin.Context) {
	id := c.Param("id")
	getReq := new(req.GetAppDetailReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 各调用方添加必传 cluster_name 的代码
	// 已知调用方:
	//   1. misc/web-bj-cicd-bash: 已解决
	//   2. 公共库解析 config: // TODO // NOTE: 注意这个地方需要等所有依赖公共库的代码都升级才算完
	//   3. 监控报警中心: // TODO: 尚未确认是否存在
	//   4. jenkins: 原本都有传递
	//   5. 配置中心: // TODO: 尚未确认是否存在
	getReq.ClusterName, err = service.SVC.CheckAndUnifyClusterName(c, getReq.ClusterName, getReq.EnvName)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	app, err := service.SVC.GetAppDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	res := new(resp.AppDisplayDetailResp)
	err = deepcopy.Copy(app).To(res)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InternalError, err.Error()))
		return
	}

	res.ClusterName = getReq.ClusterName

	project, err := service.SVC.GetProjectDetail(c, app.ProjectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	res.ProjectName = project.Name

	if getReq.DataType == req.AppDetailDataTypeGeneral {
		response.JSON(c, res, nil)
		return
	}

	// 获取新版project的LogStoreName
	res.LogStoreNameBasedProject = project.LogStoreName

	// 加工当前环境数据
	currentEnv, ok := app.Env[getReq.EnvName]
	if ok {
		res.AliAlarmName = currentEnv.AliAlarmName
		res.LogStoreName = currentEnv.LogStoreName
		res.ServiceProtocol = currentEnv.ServiceProtocol
		res.EnableBranchChangeNotification = currentEnv.EnableBranchChangeNotification
		res.EnableHotReload = currentEnv.EnableHotReload

		extra, e := service.SVC.GetAppExtraInfo(c, getReq.ClusterName, project, app, getReq.EnvName)
		if e != nil {
			response.JSON(c, nil, e)
			return
		}
		res.AppEnvExtraDetailResp = *extra

		kongFrontendInfo, e := service.SVC.GetQDNSFrontendInfo(c, app.Type, app.ServiceType, extra.AccessHosts)
		if e != nil {
			response.JSON(c, nil, e)
			return
		}
		res.KongFrontendInfo = kongFrontendInfo
		status, e := service.SVC.GetAppInClusterDNSStatus(c, getReq.ClusterName, getReq.EnvName, app)
		if e != nil {
			response.JSON(c, nil, e)
			return
		}
		res.InClusterDNSStatus = status
	}

	res.RunningStatus, err = service.SVC.GetRunningStatusList(c, getReq.ClusterName, &req.GetRunningStatusListReq{
		AppID:       app.ID,
		EnvName:     getReq.EnvName,
		Namespace:   string(getReq.EnvName),
		ClusterName: getReq.ClusterName,
	}, project, app)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	res.Subscriptions, err = service.SVC.GetUserSubscribeInfo(c, getReq.EnvName, app.ID, operatorID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, res, nil)
}

// GetApps 获取应用列表
func GetApps(c *gin.Context) {
	getReq := new(req.GetAppsReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	if getReq.AppIDs != "" {
		ids := strings.Split(getReq.AppIDs, ",")
		getReq.IDs = ids
		getReq.AppIDs = ""
	}

	apps, err := service.SVC.GetApps(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	count, err := service.SVC.GetAppsCount(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, models.BaseListResponse{
		List:  apps,
		Limit: getReq.Limit,
		Page:  getReq.Page,
		Count: count,
	}, nil)
}

func checkJobHasRunningPod(ctx context.Context, app *resp.AppDetailResp, project *resp.ProjectDetailResp) error {
	clusterList, err := service.SVC.GetClusters(ctx, new(req.GetClustersReq))
	if err != nil {
		return err
	}

	for envName := range app.Env {
		for _, cluster := range clusterList {
			if cluster.Env != envName {
				continue
			}

			jobs, err := service.SVC.GetJobs(ctx, cluster.Name,
				&req.GetJobsReq{
					Namespace:   string(envName),
					ProjectName: project.Name,
					AppName:     app.Name,
					Env:         string(envName),
				})
			if err != nil {
				return err
			}

			for i := range jobs {
				if jobs[i].Status.Active != 0 {
					return errors.Wrap(errcode.InvalidParams, "一次性任务存在正在运行的pod")
				}
			}
		}
	}

	return nil
}

// CorrectAppName 规范应用名称
func CorrectAppName(c *gin.Context) {
	id := c.Param("id")
	correctReq := new(req.CorrectAppNameReq)
	err := c.ShouldBindJSON(correctReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 获取操作人id
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	app, err := service.SVC.GetAppDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	oldAppName := app.Name

	err = service.SVC.ValidateHasPermission(c, &req.ValidateHasPermissionReq{
		OperateType: entity.OperateTypeCorrectAppName,
		ProjectID:   app.ProjectID,
		OperatorID:  operatorID,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 校验名称合法性
	project, err := service.SVC.GetProjectDetail(c, app.ProjectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	err = service.SVC.CheckAppNameLegal(c, project, app.Type, correctReq.Name)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if app.Type == entity.AppTypeOneTimeJob {
		if err = checkJobHasRunningPod(c, app, project); err != nil {
			response.JSON(c, nil, err)
			return
		}
	}

	// 获取所有集群
	clusterList, err := service.SVC.GetClusters(c, &req.GetClustersReq{})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 修改名称
	err = service.SVC.UpdateApp(c, app.ID, &req.UpdateAppReq{
		Name: correctReq.Name,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	app.Name = correctReq.Name

	// 调整正在运行的部署
	wg := goroutine.New(fmt.Sprintf("%s-%s", project.Name, app.Name))
	for envName := range app.Env {
		// 判断当前环境是否有正在运行的部署(所有集群)
		// 若存在，创建上次部署的任务
		curEnvName := envName

		for _, cluster := range clusterList {
			if cluster.Env != curEnvName {
				continue
			}

			clusterName := cluster.Name
			// 检查原有应用名是否仍存在部署
			isRunning, e := checkRunningStatus(c, clusterName, envName, app.Type, project.Name, oldAppName)
			if e != nil {
				log.Errorv(c, errcode.GetErrorMessageMap(e))
				continue
			}

			// 若已存在部署，需要迁移。则以新应用名重新创建部署。
			if isRunning {
				e = service.SVC.CreateLatestDeployTask(c, clusterName, project, app, envName, operatorID)
				if e != nil {
					log.Errorv(c, errcode.GetErrorMessageMap(e))
					continue
				}
			}

			// 执行清理任务
			wg.Go(c, fmt.Sprintf("env:%s-cluster:%s", curEnvName, clusterName),
				func(ctx context.Context) error {
					// 循环检查上次部署的状态
					for {
						time.Sleep(time.Second)

						// 获取上次部署任务
						task, err := service.SVC.GetSingleTask(ctx, &req.GetTasksReq{
							AppID:       app.ID,
							EnvName:     curEnvName,
							ClusterName: clusterName,
							ActionList:  entity.TaskActionInitDeployList,
						})
						if err != nil {
							// 避免找不到任何任务
							if errcode.EqualError(_errcode.NoRequiredTaskError, err) {
								return nil
							}
							log.Errorv(ctx, errcode.GetErrorMessageMap(err))
							continue
						}
						// 失败则不清理
						if task.Status == entity.TaskStatusFail {
							log.Errorv(ctx, errcode.GetErrorMessageMap(err))
							break
						}
						// 其他状态继续等待
						if task.Status != entity.TaskStatusSuccess {
							continue
						}

						// 部署成功，创建清理任务
						_, err = service.SVC.CreateTask(ctx, project, app, &req.CreateTaskReq{
							AppID:       app.ID,
							ClusterName: clusterName,
							EnvName:     curEnvName,
							Action:      entity.TaskActionClean,
							Param: &req.CreateTaskParamReq{
								CleanedProjectName:          project.Name,
								CleanedAppName:              oldAppName,
								CleanedAppType:              app.Type,
								CleanedAppServiceType:       app.ServiceType,
								CleanedAppServiceExposeType: app.ServiceExposeType,
							},
						}, operatorID)
						if err != nil {
							log.Errorv(ctx, errcode.GetErrorMessageMap(err))
						}

						return nil
					}
					return nil
				})
		}
	}

	response.JSON(c, nil, nil)
}

// GetAppTips 获取应用资源配置提示
func GetAppTips(c *gin.Context) {
	getReq := new(req.GetAppTipsReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}
	id := c.Param("id")

	// 获取详情
	app, err := service.SVC.GetAppDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	project, err := service.SVC.GetProjectDetail(c, app.ProjectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	res := &resp.AppTipsResp{}

	var (
		dailyMinTotalCPU, weeklyMaxTotalCPU, dailyMinTotalMem, weeklyMaxTotalMem float64
		wastedMaxCPUUsageRate, wastedMaxMemUsageRate                             float64
		isLowestCPURequested, isLowestMemoRequested                              bool
	)
	// 获取基础数据
	basicGroup := goroutine.WithContext(c, "basic tips data")
	basicGroup.Go(c, "dailyMinTotalCPU", func(ctx context.Context) error {
		dailyMinTotalCPU, err = service.SVC.GetMinTotalCPUTime(ctx, &req.GetMinTotalCPUReq{
			EnvName:   getReq.EnvName,
			CountTime: "1d",
		}, project, app)
		if err != nil && !errcode.EqualError(_errcode.PrometheusQueryEmptyError, err) {
			return err
		}
		return nil
	})
	basicGroup.Go(c, "weeklyMaxTotalCPU", func(ctx context.Context) error {
		weeklyMaxTotalCPU, err = service.SVC.GetMaxTotalCPUTime(ctx, &req.GetMaxTotalCPUReq{
			EnvName:   getReq.EnvName,
			CountTime: "7d",
		}, project, app)
		if err != nil && !errcode.EqualError(_errcode.PrometheusQueryEmptyError, err) {
			return err
		}
		return nil
	})
	basicGroup.Go(c, "dailyMaxTotalMem", func(ctx context.Context) error {
		dailyMinTotalMem, err = service.SVC.GetMinTotalMemBytes(ctx, &req.GetMinTotalMemReq{
			EnvName:   getReq.EnvName,
			CountTime: "1d",
		}, project, app)
		if err != nil && !errcode.EqualError(_errcode.PrometheusQueryEmptyError, err) {
			return err
		}
		return nil
	})
	basicGroup.Go(c, "weeklyMaxTotalMem", func(ctx context.Context) error {
		weeklyMaxTotalMem, err = service.SVC.GetMaxTotalMemBytes(ctx, &req.GetMaxTotalMemReq{
			EnvName:   getReq.EnvName,
			CountTime: "7d",
		}, project, app)
		if err != nil && !errcode.EqualError(_errcode.PrometheusQueryEmptyError, err) {
			return err
		}
		return nil
	})
	basicGroup.Go(c, "wastedMaxCPUUsageRate", func(ctx context.Context) error {
		wastedMaxCPUUsageRate, err = service.SVC.GetWastedMaxCPUUsageRate(ctx, &req.GetWastedMaxCPUUsageRateReq{
			EnvName:   getReq.EnvName,
			CountTime: "7d",
		}, project, app)
		if err != nil && !errcode.EqualError(_errcode.PrometheusQueryEmptyError, err) {
			return err
		}
		return nil
	})
	basicGroup.Go(c, "wastedMaxMemUsageRate", func(ctx context.Context) error {
		wastedMaxMemUsageRate, err = service.SVC.GetWastedMaxMemUsageRate(ctx, &req.GetWastedMaxMemUsageRateReq{
			EnvName:   getReq.EnvName,
			CountTime: "7d",
		}, project, app)
		if err != nil && !errcode.EqualError(_errcode.PrometheusQueryEmptyError, err) {
			return err
		}
		return nil
	})
	basicGroup.Go(c, "latestTask", func(ctx context.Context) error {
		task, err := service.SVC.GetLatestDeploySuccessTaskFinalVersion(ctx, &req.GetLatestTaskReq{
			AppID:       id,
			EnvName:     getReq.EnvName,
			ClusterName: getReq.ClusterName,
		})
		if err != nil {
			return err
		}

		isLowestCPURequested = (getReq.EnvName != entity.AppEnvPrd) || (task.Param.MinPodCount == 1 && task.Param.CPURequest == entity.CPUResourceNano)
		isLowestMemoRequested = (getReq.EnvName != entity.AppEnvPrd) || (task.Param.MinPodCount == 1 && task.Param.MemRequest == entity.MemResourceNano)
		return nil

	})
	err = basicGroup.Wait()
	if err != nil {
		log.Errorc(c, "get basic tips data error:%s", err.Error())
	}

	// 推荐合适的资源
	if dailyMinTotalCPU != 0 && weeklyMaxTotalCPU != 0 &&
		dailyMinTotalMem != 0 && weeklyMaxTotalMem != 0 {
		resourceResp, e := service.SVC.CalculateAppRecommendResource(c, &req.CalculateAppRecommendReq{
			EnvName:           getReq.EnvName,
			DailyMinTotalCPU:  dailyMinTotalCPU,
			WeeklyMaxTotalCPU: weeklyMaxTotalCPU,
			DailyMinTotalMem:  dailyMinTotalMem,
			WeeklyMaxTotalMem: weeklyMaxTotalMem,
		})
		if e != nil {
			response.JSON(c, nil, e)
			return
		}
		res.RecommendMemRequest = resourceResp.RecommendMemRequest
		res.RecommendCPURequest = resourceResp.RecommendCPURequest
		res.RecommendMaxPodCount = resourceResp.RecommendMaxPodCount
		res.RecommendMinPodCount = resourceResp.RecommendMinPodCount
	}

	percentFloat := 100.0
	if !isLowestCPURequested && wastedMaxCPUUsageRate != 0 {
		res.WastedMaxCPUUsageRate = fmt.Sprintf("%.2f%%", wastedMaxCPUUsageRate*percentFloat)
	}

	if !isLowestMemoRequested && wastedMaxMemUsageRate != 0 {
		res.WastedMaxMemUsageRate = fmt.Sprintf("%.2f%%", wastedMaxMemUsageRate*percentFloat)
	}

	response.JSON(c, res, nil)
}

// CreateAppSentry 创建应用 sentry 报警
func CreateAppSentry(c *gin.Context) {
	appID := c.Param("id")
	app, err := service.SVC.GetAppDetail(c, appID)

	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	project, err := service.SVC.GetProjectDetail(c, app.ProjectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	createResp, err := service.SVC.CreateAppSentry(c, project.Team, project, app)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	err = service.SVC.UpdateApp(c, appID, &req.UpdateAppReq{
		SentryProjectSlug:      createResp.SentryProjectSlug,
		SentryProjectPublicDsn: createResp.SentryProjectPublicDsn,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

// SetAppClusterWeights 设置应用环境集群的权重
func SetAppClusterWeights(c *gin.Context) {
	id := c.Param("id")

	setReq := new(req.SetAppClusterQDNSWeightsReq)
	err := c.ShouldBindJSON(setReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}

	// 获取操作人id
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "operator id(%+v) is invalid", c.Value(utils.ContextUserIDKey)))
		return
	}

	app, err := service.SVC.GetAppDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	_, ok = app.Env[setReq.Env]
	if !ok {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "env(%s) not found in app(%s)", setReq.Env, app.ID))
		return
	}

	// 只有 restful 服务支持配置权重
	if app.Type != entity.AppTypeService || app.ServiceType != entity.AppServiceTypeRestful {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams,
			"app(%s) of type(%s) and service_type(%s) does not support multicluster weight setting", app.ID, app.Type, app.ServiceType))
		return
	}

	multiClusterSupported, err := service.SVC.CheckMultiClusterSupport(c, setReq.Env, app.ProjectID)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(_errcode.K8sInternalError, err.Error()))
		return
	}

	// 非多集群白名单内的项目应用不支持配置
	if !multiClusterSupported {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams,
			"app(%s) of project(%s) is not in multicluster white list", app.ID, app.ProjectID))
		return
	}

	// 配置集群权重的操作需要权限控制
	err = service.SVC.ValidateHasPermission(c, &req.ValidateHasPermissionReq{
		OperateType:       entity.OperateTypeSetAppClusterKongWeights,
		CreateTaskEnvName: setReq.Env,
		ProjectID:         app.ProjectID,
		OperatorID:        operatorID,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	project, err := service.SVC.GetProjectDetail(c, app.ProjectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	err = validateClusterWeights(setReq.ClusterWeights)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 请求设置权重的所有集群下所有已存在部署的健康检查路径必须一致
	setReq.HealthCheckPath, err = service.SVC.CheckAppHealthCheckURLDifference(c, project, app, setReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	setReq.OperatorID = operatorID
	setReq.DomainController = service.SVC.GetDomainControllerFromProjectOwners(project.Owners)

	err = service.SVC.SetAppClusterQDNSWeights(c, app, setReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

// GetAppClusterWeights 获取应用环境所有集群的权重
func GetAppClusterWeights(c *gin.Context) {
	id := c.Param("id")

	app, err := service.SVC.GetAppDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	envName := c.Query("env")
	_, ok := app.Env[entity.AppEnvName(envName)]
	if !ok {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "env(%s) not found in app(%s)", envName, app.ID))
		return
	}

	res, err := service.SVC.GetAppClusterQDNSWeights(c, app.ServiceName, entity.AppEnvName(envName))
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 目前没有分页需求
	response.JSON(c, models.BaseListResponse{
		List: res,
	}, nil)
}

// GetAppClustersWithWorkload 获取应用在指定环境下有工作负载的所有集群
func GetAppClustersWithWorkload(c *gin.Context) {
	id := c.Param("id")
	getReq := new(req.GetAppClustersWithWorkloadReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}

	app, err := service.SVC.GetAppDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	project, err := service.SVC.GetProjectDetail(c, app.ProjectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 集群数量有限, 获取全部再按照分页返回即可
	clusters, err := service.SVC.GetAppClustersWithWorkload(c, getReq.EnvName, project, app)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	sort.SliceStable(clusters, func(i, j int) bool {
		return clusters[i].Name < clusters[j].Name
	})

	getReq.Unify()
	res := &models.BaseListResponse{
		Limit: getReq.Limit,
		Page:  getReq.Page,
		Count: len(clusters),
	}

	start := (getReq.Page - 1) * getReq.Limit
	if start >= res.Count {
		res.List = resp.EmptyClusterDetailRespList
	} else if end := start + getReq.Limit; end <= res.Count {
		res.List = clusters[start:end]
	} else {
		res.List = clusters[start:]
	}

	response.JSON(c, res, nil)
}

func checkRunningStatus(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName,
	appType entity.AppType, projectName, appName string) (bool, error) {
	switch appType {
	case entity.AppTypeCronJob:
		cronjobs, err := service.SVC.GetCronJobs(ctx, clusterName, &req.GetCronJobsReq{
			Namespace:   string(envName),
			ProjectName: projectName,
			AppName:     appName,
		})
		if err != nil {
			return false, err
		}

		return len(cronjobs) > 0, nil

	case entity.AppTypeOneTimeJob:

	default:
		exists, err := service.SVC.CheckDeploymentsExistance(ctx, clusterName, envName,
			&req.GetDeploymentsReq{
				Namespace:   string(envName),
				ProjectName: projectName,
				AppName:     appName,
				Env:         string(envName),
			})
		if err != nil {
			return false, err
		}

		return exists, nil
	}

	return false, nil
}

// validateClusterWeights 校验集群权重设置是否合法
func validateClusterWeights(weights []*req.AppClusterKongWeight) error {
	totalWeight := 0
	clusterNameMapping := make(map[entity.ClusterName]struct{})
	for _, weight := range weights {
		if !(entity.ValidateClusterName(weight.ClusterName)) {
			return errors.Wrapf(errcode.InvalidParams, "invalid cluster name(%s)", weight.ClusterName)
		}

		if weight.Weight < 0 || weight.Weight > 100 {
			return errors.Wrapf(errcode.InvalidParams, "invalid weight(%d)", weight.Weight)
		}

		if _, ok := clusterNameMapping[weight.ClusterName]; ok {
			return errors.Wrapf(errcode.InvalidParams, "duplicate cluster(%s)", weight.ClusterName)
		}

		totalWeight += weight.Weight
		clusterNameMapping[weight.ClusterName] = struct{}{}
	}

	if totalWeight != 100 {
		return errors.Wrapf(errcode.InvalidParams, "total weight(%d) not equal to 100", totalWeight)
	}

	return nil
}

// CheckApp check app info as middleware
func CheckApp(ctx *gin.Context) {
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

	_, err = service.SVC.GetAppByObjectID(ctx, objectID)

	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		ctx.Abort()
		response.JSON(ctx, nil, errors.Wrap(_errcode.NotFoundError, "app id 不存在"))

		return
	}
	if err != nil {
		ctx.Abort()
		response.JSON(ctx, nil, err)

		return
	}
}
