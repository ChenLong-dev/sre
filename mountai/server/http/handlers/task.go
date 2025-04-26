package handlers

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	batchV1 "k8s.io/api/batch/v1"

	"rulai/models"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/service"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
	"rulai/utils/response"

	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/base/null"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// CreateTask 创建任务
func CreateTask(c *gin.Context) {
	createReq := new(req.CreateTaskReq)
	err := c.ShouldBindJSON(createReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 校验 config
	err = checkAndUnifyConfigParams(createReq.Param)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}

	// 获取操作人id
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	if createReq.OperatorID != "" {
		operatorID = createReq.OperatorID
		err = service.SVC.CheckAndSyncUser(c, operatorID)
		if err != nil {
			response.JSON(c, nil, err)
			return
		}
	}

	// 校验集群名
	createReq.ClusterName, err = service.SVC.CheckAndUnifyClusterName(c, createReq.ClusterName, createReq.EnvName)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 校验应用
	app, err := service.SVC.GetAppDetail(c, createReq.AppID)
	if errcode.EqualError(_errcode.InvalidHexStringError, err) || errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 校验 LB，注意，LB 信息创建后不支持更改
	if app.Type == entity.AppTypeService && app.ServiceExposeType == entity.AppServiceExposeTypeLB {
		lbs := service.SVC.GetAppLoadBalancersByEnvAndCluster(c, app, createReq.ClusterName, createReq.EnvName)
		for _, lb := range lbs {
			if lb.LoadBalancerID == "" {
				response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "service load balancer instance id is empty"))
				return
			}
		}
	}

	// 校验项目
	project, err := service.SVC.GetProjectDetail(c, app.ProjectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	createReq.Namespace = service.SVC.GetNamespaceBase(service.SVC.GetApplicationIstioState(
		context.Background(), createReq.EnvName, createReq.ClusterName, app), createReq.EnvName)

	// 校验cpu/memory参数
	if createReq.Action == entity.TaskActionFullDeploy || createReq.Action == entity.TaskActionCanaryDeploy {
		err = service.SVC.ValidateResourceRequirements(&v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(string(createReq.Param.CPURequest)),
				v1.ResourceMemory: resource.MustParse(string(createReq.Param.MemRequest)),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(string(createReq.Param.CPULimit)),
				v1.ResourceMemory: resource.MustParse(string(createReq.Param.MemLimit)),
			},
		})
		if err != nil {
			response.JSON(c, nil, err)
			return
		}
	}

	// 校验是否允许创建任务
	err = validateIsAllowCreateTask(c, createReq, project, app, operatorID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if createReq.Approval == nil {
		createReq.Approval = new(req.ApprovalReq)
		// Default skip approval type.
		createReq.Approval.Type = entity.SkipTaskApprovalType
	}

	// validate approval params.
	if err = validateApprovalParams(c, createReq, project, app, operatorID); err != nil {
		response.JSON(c, nil, err)
		return
	}

	// Create task.
	res, err := service.SVC.CreateTask(c, project, app, createReq, operatorID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	response.JSON(c, res, nil)
}

// validateIsExpectBranch 验证当前部署的分支是否是预期分支
func validateIsExpectBranch(ctx context.Context, createReq *req.CreateTaskReq) error {
	// Step: 以下情况跳过验证
	if createReq.IgnoreExpectedBranch {
		return nil
	}

	// Step: 获取上次分支
	lastTaskBranch, err := service.SVC.GetLatestDeploySuccessTaskBranch(ctx, createReq)
	if err != nil || lastTaskBranch == "" {
		return err
	}

	// Step: 获取当前部署分支
	_, _, currTaskBranch, err := service.SVC.ExtraInfoFromImageVersion(createReq.Param.ImageVersion)
	if err != nil {
		return err
	}

	// Step: 比对分支
	if lastTaskBranch != currTaskBranch {
		return _errcode.CurrBranchNotMatchExpectBranchError
	}

	return nil
}

func validateCluster(ctx context.Context, createReq *req.CreateTaskReq, project *resp.ProjectDetailResp) error {
	clusters, err := service.SVC.GetProjectSupportedClusters(ctx, project.ID, createReq.EnvName)
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		if cluster.Env == createReq.EnvName && cluster.Name == createReq.ClusterName {
			return nil
		}
	}

	return errors.Wrapf(errcode.InvalidParams, "%s is not support", createReq.ClusterName)
}

// 校验是否允许创建任务
func validateIsAllowCreateTask(ctx context.Context, createReq *req.CreateTaskReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, operatorID string) error {
	// Validate whether task is system action task
	for _, action := range entity.TaskActionSystemList {
		if action == createReq.Action {
			return errors.Wrap(errcode.InvalidParams, "系统行为任务不允许创建")
		}
	}

	err := validateCluster(ctx, createReq, project)
	if err != nil {
		return err
	}

	// 检查是否有其他未完成任务, 所有集群同时只能有一个任务在执行, 否则域名解析可能会有问题
	unFinishCount, err := service.SVC.GetTasksCount(ctx, &req.GetTasksReq{
		AppID:             app.ID,
		EnvName:           createReq.EnvName,
		StatusInverseList: entity.TaskStatusFinalStateList,
		Suspend:           null.BoolFrom(false),
	})
	if err != nil {
		return errors.Wrapf(errcode.InvalidParams, "%s", err)
	}
	if unFinishCount > 0 {
		return errors.WithStack(_errcode.OtherRunningTaskExistsError)
	}

	err = service.SVC.ValidateHasPermission(ctx, &req.ValidateHasPermissionReq{
		OperateType:       entity.OperateTypeCreateTask,
		CreateTaskEnvName: createReq.EnvName,
		CreateTaskAction:  createReq.Action,
		ProjectID:         project.ID,
		OperatorID:        operatorID,
	})
	if err != nil {
		return err
	}

	action := createReq.Action
	// 校验部署类操作
	for _, deployAction := range entity.TaskActionInitDeployList {
		if action != deployAction {
			continue
		}
		if (app.Type == entity.AppTypeCronJob || app.Type == entity.AppTypeOneTimeJob) &&
			action != entity.TaskActionFullDeploy {
			return errors.Wrap(errcode.InvalidParams, "action should be full_deploy")
		}
		err = validateDeployTaskParams(ctx, createReq, app)
		if err != nil {
			return err
		}

		return nil
	}

	// DNS以外的非部署类操作要求必须指定版本
	if !entity.TaskActionInClusterDNSList.Contains(createReq.Action) && createReq.Version == "" {
		return errors.Wrap(errcode.InvalidParams, "version is empty")
	}

	err = validateByAction(ctx, createReq, project, app)
	if err != nil {
		return err
	}

	// DNS以外的非部署类操作要求必须存在资源
	if !entity.TaskActionInClusterDNSList.Contains(createReq.Action) {
		// 这里检测的时候,考虑 istio 单独部署空间,因此,需要进行全局查找
		exist, err := service.SVC.IsAppK8sPrimaryResourceExist(ctx, createReq.ClusterName,
			createReq.EnvName, app.Type, createReq.Version, createReq.Namespace)
		if err != nil {
			return err
		}
		if !exist {
			return errors.Wrapf(_errcode.TaskPrimaryResourceNotExistsError, "can't %s task", action)
		}
	}

	return nil
}

func validateByAction(ctx context.Context, createReq *req.CreateTaskReq, project *resp.ProjectDetailResp, app *resp.AppDetailResp) error {
	// 校验非部署类操作参数
	switch createReq.Action {
	case entity.TaskActionManualLaunch:
		if app.Type != entity.AppTypeCronJob {
			return errors.Wrap(errcode.InvalidParams, "action is only supported for cronjob")
		}
		err := validateManualLaunchCronJob(ctx, createReq)
		if err != nil {
			return err
		}
	case entity.TaskActionUpdateHPA:
		if !(app.Type == entity.AppTypeService || app.Type == entity.AppTypeWorker) {
			return errors.Wrap(errcode.InvalidParams, "action is not supported")
		}
		err := validateUpdateHPA(ctx, createReq)
		if err != nil {
			return err
		}
	case entity.TaskActionReloadConfig:
		// TODO: Cronjob 暂时不支持 reload 功能
		if app.Type == entity.AppTypeOneTimeJob || app.Type == entity.AppTypeCronJob ||
			createReq.Param.ConfigCommitID == "" {
			return errors.Wrap(errcode.InvalidParams, "action is not supported")
		}
		err := validateReloadConfig(ctx, createReq, project, app)
		if err != nil {
			return err
		}
	case entity.TaskActionDisableInClusterDNS:
		err := validateDisableInClusterDNSCreateTask(ctx, createReq, project, app)
		if err != nil {
			return err
		}
	case entity.TaskActionEnableInClusterDNS:
		err := validateEnableInClusterDNSCreateTask(ctx, app, createReq)
		if err != nil {
			return err
		}
	case entity.TaskActionStop:
		if app.Type != entity.AppTypeCronJob && createReq.Param.MinPodCount != 0 {
			return errors.Wrap(errcode.InvalidParams, "min_pod_count should be 0")
		}
		if app.Type == entity.AppTypeOneTimeJob {
			return errors.Wrap(errcode.InvalidParams, "action is unsupported for one time job")
		}
		return validateAllowStopOrDeleteActionCreateTask(ctx, app, createReq)
	case entity.TaskActionDelete:
		return validateAllowStopOrDeleteActionCreateTask(ctx, app, createReq)
	default:
		err := validateIsAllowCreateCommonTask(ctx, createReq, app.Type, createReq.Action)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateAllowStopOrDeleteActionCreateTask(ctx context.Context, app *resp.AppDetailResp, createReq *req.CreateTaskReq) (err error) {
	if app.ServiceExposeType != entity.AppServiceExposeTypeIngress || app.ServiceType != entity.AppServiceTypeRestful {
		return nil
	}

	multiClusterSupported, err := service.SVC.CheckMultiClusterSupport(ctx, createReq.EnvName, app.ProjectID)
	if err != nil {
		return err
	}

	if !multiClusterSupported {
		return nil
	}

	serviceName, err := service.SVC.GetCurrentServiceName(ctx, createReq.ClusterName, createReq.EnvName, app)
	if err != nil {
		return err
	}

	if serviceName != app.ServiceName {
		return nil
	}

	deploy, err := service.SVC.GetDeploymentDetail(ctx, createReq.ClusterName, createReq.EnvName, &req.GetDeploymentDetailReq{
		Namespace: createReq.Namespace,
		Name:      createReq.Version,
		Env:       string(createReq.EnvName),
	})

	if err != nil {
		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil
		}

		return err
	}

	if deploy.Status.ReadyReplicas == 0 {
		return nil
	}

	// FIXME: 多集群已经迁移完成, 临时先注销删除的 dns 限制, 先修复域名解析问题后再处理这里
	// exists, err := service.SVC.CheckHealthyDeploymentExistance(ctx, createReq.ClusterName, createReq.Namespace, &req.GetDeploymentsReq{
	// 	Namespace:   string(createReq.Namespace),
	// 	ProjectName: project.Name,
	// 	AppName:     app.Name,
	// }, createReq.Version)

	// if err != nil {
	// 	return err
	// }

	// if !exists {
	// 	return errors.Wrapf(errcode.InvalidParams, "you must disable cluster dns")
	// }

	return nil
}

func validateEnableInClusterDNSCreateTask(ctx context.Context, app *resp.AppDetailResp, createReq *req.CreateTaskReq) (err error) {
	if app.ServiceExposeType != entity.AppServiceExposeTypeIngress || app.ServiceType != entity.AppServiceTypeRestful {
		return errors.Wrapf(errcode.InvalidParams, "only restful app can disable/enable incluster dns")
	}

	serviceName, err := service.SVC.GetCurrentServiceName(ctx, createReq.ClusterName, createReq.EnvName, app)
	if err != nil {
		return err
	}

	if serviceName == app.ServiceName {
		return errors.Wrapf(errcode.InvalidParams, "incluster dns has been enabled")
	}

	return nil
}

// 1. LB或者GRPC方式的svc 不允许禁用
// 2. 复用了多集群的检查, 需要保证完成了kong配置的切换和解析的修改
// 3. 当时app关联的service是否被禁用
func validateDisableInClusterDNSCreateTask(ctx context.Context, createReq *req.CreateTaskReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) (err error) {
	if app.ServiceExposeType != entity.AppServiceExposeTypeIngress || app.ServiceType != entity.AppServiceTypeRestful {
		return errors.Wrapf(errcode.InvalidParams, "only restful app can disable/enable incluster dns")
	}

	exist, err := service.SVC.IsAppExistsARecord(ctx, app, createReq.EnvName)
	if err != nil {
		return err
	}

	if exist {
		return errors.Wrap(errcode.InvalidParams, "task private zone type should be cname")
	}

	serviceName, err := service.SVC.GetCurrentServiceName(ctx, createReq.ClusterName, createReq.EnvName, app)
	if err != nil {
		return err
	}

	if serviceName != app.ServiceName {
		return errors.Wrapf(errcode.InvalidParams, "incluster dns has been disabled")
	}

	deployments, err := service.SVC.GetDeployments(ctx, createReq.ClusterName, createReq.EnvName, &req.GetDeploymentsReq{
		Namespace:   string(createReq.EnvName),
		ProjectName: project.Name,
		AppName:     app.Name,
		Env:         string(createReq.EnvName),
	})

	if err != nil {
		return err
	}

	for i := range deployments {
		if deployments[i].Status.UnavailableReplicas > 0 {
			return errors.Wrapf(errcode.InvalidParams,
				"%s contains unavailable pods, count=%d ", deployments[i].Name, deployments[i].Status.UnavailableReplicas)
		}
	}

	// FIXME: 临时不允许 禁用DNS 功能
	return errors.Wrapf(errcode.Forbidden, "temporarily not allowed ... may cause problems")
}

// validateUpdateHPA 检查更新HPA&cronHPA的参数
func validateUpdateHPA(_ context.Context, createReq *req.CreateTaskReq) (err error) {
	if createReq.Param.MinPodCount == 0 {
		return errors.Wrap(errcode.InvalidParams, "min_pod_count is 0")
	}
	if createReq.Param.MaxPodCount == 0 {
		return errors.Wrap(errcode.InvalidParams, "max_pod_count is 0")
	}
	if createReq.Param.MaxPodCount < createReq.Param.MinPodCount {
		return errors.Wrap(errcode.InvalidParams, "max_pod_count < min_pod_count")
	}

	e := validateCronAutoScaleJobs(createReq.Param)
	if e != nil {
		return errors.Wrap(errcode.InvalidParams, e.Error())
	}

	return nil
}

// validateManualLaunchCronJob 检测手动启动的参数
func validateManualLaunchCronJob(ctx context.Context, createReq *req.CreateTaskReq) (err error) {
	// 检测cronjob存在
	cronJob, err := service.SVC.GetCronJobDetail(ctx, createReq.ClusterName,
		&req.GetCronJobDetailReq{
			Namespace: createReq.Namespace,
			Name:      createReq.Version,
			Env:       string(createReq.EnvName),
		})
	if err != nil {
		return errors.Wrapf(errcode.InvalidParams, "%s", err)
	}

	// 检测ConcurrencyPolicy是否为“Forbid”
	if cronJob.Spec.ConcurrencyPolicy == batchV1.ForbidConcurrent && len(cronJob.Status.Active) > 0 {
		return errors.Wrap(_errcode.ExistingActiveJobsError, "current cronjob concurrencyPolicy is forbid")
	}
	return nil
}

// 校验部署任务的参数
func validateDeployTaskParams(ctx context.Context, createReq *req.CreateTaskReq,
	app *resp.AppDetailResp) error {
	param := createReq.Param

	if param.ImageVersion == "" {
		return errors.Wrap(errcode.InvalidParams, "image version is empty")
	}
	if param.CPURequest == "" || param.CPULimit == "" {
		return errors.Wrap(errcode.InvalidParams, "cpu resource is empty")
	}
	if param.MemRequest == "" || param.MemLimit == "" {
		return errors.Wrap(errcode.InvalidParams, "memory resource is empty")
	}

	// 校验宽限终止时长
	if param.TerminationGracePeriodSeconds < 0 {
		return errors.Wrap(errcode.InvalidParams, "terminationGracePeriodSeconds is invalid")
	}

	switch tp := app.Type; tp {
	case entity.AppTypeOneTimeJob:
		if param.BackoffLimit < 0 {
			return errors.Wrap(errcode.InvalidParams, "backoff limit >= 0")
		}
	case entity.AppTypeCronJob:
		if param.CronParam == "" {
			return errors.Wrap(errcode.InvalidParams, "cron param is empty")
		}

		if param.BackoffLimit < 0 || param.BackoffLimit > entity.MaxBackOffLimit {
			return errors.Wrapf(errcode.InvalidParams, "backoff limit should between 0 and %d", entity.MaxBackOffLimit)
		}
		if param.ConcurrencyPolicy == "" {
			return errors.Wrap(errcode.InvalidParams, "concurrency policy is empty")
		}
		if param.RestartPolicy == "" {
			return errors.Wrap(errcode.InvalidParams, "restart policy is empty")
		}
		if param.SuccessfulHistoryLimit == 0 {
			return errors.Wrap(errcode.InvalidParams, "successful history limit is empty")
		}
		if param.FailedHistoryLimit == 0 {
			return errors.Wrap(errcode.InvalidParams, "failed history limit is empty")
		}
	case entity.AppTypeService, entity.AppTypeWorker:
		if tp == entity.AppTypeService {
			err := validateServiceDeployTaskParams(ctx, createReq, app)
			if err != nil {
				return err
			}
		}

		if param.MinPodCount == 0 {
			return errors.Wrap(errcode.InvalidParams, "min_pod_count is 0")
		}

		if param.IsAutoScale {
			if param.MaxPodCount == 0 {
				return errors.Wrap(errcode.InvalidParams, "max_pod_count is 0")
			}
			if param.MaxPodCount < param.MinPodCount {
				return errors.Wrap(errcode.InvalidParams, "max_pod_count < min_pod_count")
			}

			err := validateCronAutoScaleJobs(param)
			if err != nil {
				return err
			}
		}
	}

	// 校验分支是否是预期分支
	if err := validateIsExpectBranch(ctx, createReq); err != nil {
		return err
	}

	return nil
}

// validateCronAutoScaleJobs 校验cronHPA扩缩容任务
func validateCronAutoScaleJobs(param *req.CreateTaskParamReq) error {
	if len(param.CronScaleJobGroups) == 0 && len(param.CronScaleJobExcludeDates) != 0 {
		return errors.Wrap(errcode.InvalidParams,
			"CronScaleJobExcludeDates with empty CronScaleJobGroups")
	}
	if len(param.CronScaleJobGroups) == 0 {
		return nil
	}

	err := service.SVC.ValidateCronHPA(param, service.MinInterval, service.MaxInterval)
	if err != nil {
		return err
	}

	return nil
}

// 校验服务部署任务的参数
func validateServiceDeployTaskParams(ctx context.Context, createReq *req.CreateTaskReq,
	app *resp.AppDetailResp) error {
	param := createReq.Param

	if param.HealthCheckURL == "" {
		return errors.Wrap(errcode.InvalidParams, "health check url is empty")
	}
	if param.TargetPort == 0 {
		return errors.Wrap(errcode.InvalidParams, "target port is 0")
	}

	// TODO: 直接从阿里云取
	// err := service.SVC.CheckMultiClusterCreateTask(ctx, app, createReq)
	// if err != nil {
	// 	return err
	// }

	// grpc服务ingress默认对外暴露端口为443，并且使用headless service
	// 若pod暴露非443端口，将会导致集群外访问端口与集群内访问端口不一致的情况
	// 故强制要求grpc服务实例暴露443端口
	if app.ServiceType == entity.AppServiceTypeGRPC && param.TargetPort != int(entity.ServiceGRPCDefaultInternalPort) {
		return errors.Wrapf(errcode.InvalidParams, "grpc service target port must be %d", entity.ServiceGRPCDefaultInternalPort)
	}

	if len(param.ExposedPorts) > 0 {
		var (
			ok            bool
			protectedName string
			protectedPort int32
		)
		switch app.ServiceType {
		case entity.AppServiceTypeRestful:
			ok, err := service.SVC.CheckExposedPortsWhiteList(ctx, app.ID)
			if err != nil {
				return err
			}

			if !ok {
				return errors.Wrapf(errcode.InvalidParams, "unspport exposed ports")
			}

			protectedName = entity.ServiceDefaultHTTPName
			protectedPort = entity.ServiceDefaultInternalPort

		case entity.AppServiceTypeGRPC:
			protectedName = entity.ServiceDefaultHTTPSName
			protectedPort = entity.ServiceGRPCDefaultInternalPort

		default:
			return errors.Wrapf(errcode.InvalidParams, "unspport exposed ports")
		}

		var portName string
		portMap := make(map[int]string)
		for name, port := range param.ExposedPorts {
			// 防止重复
			if portName, ok = portMap[port]; ok {
				return errors.Wrapf(errcode.InvalidParams, "duplicate exposed port(%d) with name(%s and %s)", port, portName, name)
			}

			// 保护默认端口
			if name == protectedName {
				return errors.Wrapf(errcode.InvalidParams, "exposed port name shouldn't be default(%s)", entity.ServiceDefaultHTTPName)
			}

			if port == int(protectedPort) {
				return errors.Wrapf(errcode.InvalidParams, "exposed port shouldn't be default(%d)", port)
			}

			if port == param.TargetPort {
				return errors.Wrapf(errcode.InvalidParams, "exposed port shouldn't be equal to target_port(%d)", port)
			}

			portMap[port] = name
		}
	}

	return nil
}

// GetTaskDetail : 获取任务详情
func GetTaskDetail(c *gin.Context) {
	id := c.Param("id")
	getReq := new(req.GetTaskDetailReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	task, err := service.SVC.GetTaskDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, task, nil)
}

// GetTasks : 获取任务列表
func GetTasks(c *gin.Context) {
	getReq := new(req.GetTasksReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}
	getReq.ActionInverseList = entity.TaskActionSystemList

	tasks, err := service.SVC.GetTasks(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	count, err := service.SVC.GetTasksCount(c, getReq)
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

// GetLatestDeployTaskDetail : 获取最新部署的任务详情
func GetLatestDeployTaskDetail(c *gin.Context) {
	getReq := new(req.GetLatestTaskReq)
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

	task, err := service.SVC.GetLatestDeploySuccessTaskFinalVersion(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, task, nil)
}

// UpdateTask : 更新任务
func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	updateReq := new(req.UpdateTaskReq)
	err := c.ShouldBindJSON(updateReq)
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
	updateReq.OperatorID = operatorID

	// 获取task详情
	task, err := service.SVC.GetTaskDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	if entity.TaskActionInClusterDNSList.Contains(task.Action) {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "task(action=%s) can not be updated", task.Action))
		return
	}

	// 终止态或者init状态的任务无法修改
	if updateReq.DeployType == "" {
		for _, status := range entity.TaskStatusFinalAndInitList {
			if task.Status == status {
				response.JSON(c, nil, errors.Wrapf(_errcode.FinalStateTaskUpdateError, "current status:%s", task.Status))
				return
			}
		}
	}

	if updateReq.DeployType != "" {
		if task.Approval == nil {
			response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "task approval is nil"))
			return
		}
		if task.Approval.Status != entity.ApprovedTaskApprovalStatus {
			response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "can't update approving task's deployment type"))
			return
		}
		if updateReq.DeployType != entity.ImmediateTaskDeployType {
			response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "can't update prd app's task's deployment type"))
			return
		}
		if updateReq.DeployType == entity.ScheduledTaskDeployType {
			if updateReq.ScheduleTime.IsZero() {
				response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "schedule time is empty"))
				return
			}
			if time.Unix(updateReq.ScheduleTime.ValueOrZero(), 0).Before(time.Now()) {
				response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "schedule time is before now"))
				return
			}
		}
	}

	// 获取应用详情
	app, err := service.SVC.GetAppDetail(c, task.AppID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	// 获取项目详情
	project, err := service.SVC.GetProjectDetail(c, app.ProjectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// Update task.
	err = service.SVC.UpdateTask(c, project, app, task, updateReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	response.JSON(c, nil, nil)
}

func DeleteTask(c *gin.Context) {
	id := c.Param("id")

	task, err := service.SVC.GetTaskDetail(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	if task.Status != entity.TaskStatusInit {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams,
			"can't delete init status task"))
		return
	}

	err = service.SVC.DeleteSingleTask(c, task, operatorID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	response.JSON(c, nil, nil)
}

// BatchCreateTask : 批量创建任务
func BatchCreateTask(c *gin.Context) {
	batchReq := new(req.BatchCreateTaskReq)
	err := c.ShouldBindJSON(batchReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 校验 config
	err = checkAndUnifyConfigParams(batchReq.Param)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}

	// 获取操作人id
	operatorID, ok := c.Value(utils.ContextUserIDKey).(string)
	if !ok {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator id is invalid"))
		return
	}

	// 校验集群名
	batchReq.ClusterName, err = service.SVC.CheckAndUnifyClusterName(c, batchReq.ClusterName, batchReq.EnvName)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 校验项目
	project, err := service.SVC.GetProjectDetail(c, batchReq.ProjectID)
	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 获取应用
	apps, err := service.SVC.GetApps(c, &req.GetAppsReq{
		ProjectID: batchReq.ProjectID,
		EnvName:   batchReq.EnvName,
		IDs:       batchReq.AppIDs,
	})
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 校验 action
	err = checkTaskAction(batchReq.EnvName, batchReq.Action, apps)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}

	// P0 level project ignores approval process.
	if !batchReq.IgnoreApprovalProcess.IsZero() && !batchReq.IgnoreApprovalProcess.ValueOrZero() &&
		service.SVC.IsPrdP0LevelApp(batchReq.EnvName, project) {
		response.JSON(c, nil, errors.Wrap(_errcode.NoApprovalProcessError, "no approval process error"))
		return
	}

	createTaskReqMap, err := transBatchCreateTaskToTasksMapping(c, operatorID, batchReq, project, apps)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 批量创建任务
	for _, app := range apps {
		createReq, ok := createTaskReqMap[app.ID]
		if !ok {
			log.Errorc(c, "create task req not found: id:%s name:%s", app.ID, app.Name)
			continue
		}

		_, err = service.SVC.CreateTask(c, project, &resp.AppDetailResp{
			ID:          app.ID,
			Name:        app.Name,
			Type:        app.Type,
			ServiceType: app.ServiceType,
			ProjectID:   app.ProjectID,
			CreateTime:  app.CreateTime,
			UpdateTime:  app.UpdateTime,
		}, createReq, operatorID)
		if err != nil {
			log.Errorc(c, "app:%s err:%s", app.Name, err.Error())
			continue
		}
	}

	response.JSON(c, nil, nil)
}

func transBatchCreateTaskToTasksMapping(ctx context.Context, operatorID string,
	batchReq *req.BatchCreateTaskReq, project *resp.ProjectDetailResp, apps []*resp.AppListResp) (map[string]*req.CreateTaskReq, error) {
	// 遍历检查合法性
	// 批量发布时如果遇到参数校验错误，需要进行提示
	// 每个错误都需要记录并在循环结束后进行处理
	createTaskReqMap := make(map[string]*req.CreateTaskReq)
	errGroup := errcode.NewGroup(_errcode.BatchCreateTaskWarningsError)
	for _, app := range apps {
		// 获取上次成功任务(同集群)
		latestSuccessTask, err := service.SVC.GetLatestDeploySuccessTaskFinalVersion(ctx, &req.GetLatestTaskReq{
			AppID:       app.ID,
			EnvName:     batchReq.EnvName,
			ClusterName: batchReq.ClusterName,
		})
		if err != nil {
			return nil, err
		}

		// 以上次成功任务作为基本参数
		createReq := new(req.CreateTaskReq)
		err = deepcopy.Copy(latestSuccessTask).To(createReq)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}

		// 额外赋值
		createReq.EnvName = batchReq.EnvName
		createReq.ClusterName = batchReq.ClusterName
		createReq.Action = batchReq.Action
		createReq.AppID = app.ID
		createReq.OperatorID = operatorID
		createReq.Param.ConfigRenameMode = batchReq.Param.ConfigRenameMode
		createReq.Param.ConfigRenamePrefix = batchReq.Param.ConfigRenamePrefix
		switch createReq.Action {
		case entity.TaskActionCanaryDeploy, entity.TaskActionFullDeploy:
			createReq.Version = ""
			createReq.Param.ImageVersion = batchReq.Param.ImageVersion
		case entity.TaskActionStop:
			createReq.Param.MinPodCount = batchReq.Param.MinPodCount
		}
		if batchReq.Param.ConfigCommitID != "" {
			createReq.Param.ConfigCommitID = batchReq.Param.ConfigCommitID
		}
		// Batch create task should ignore expected branch
		createReq.IgnoreExpectedBranch = true

		appDetail := &resp.AppDetailResp{
			ID:                app.ID,
			Name:              app.Name,
			Type:              app.Type,
			ServiceType:       app.ServiceType,
			ProjectID:         app.ProjectID,
			CreateTime:        app.CreateTime,
			UpdateTime:        app.UpdateTime,
			ServiceName:       app.ServiceName,
			ServiceExposeType: app.ServiceExposeType,
		}

		// 兼容历史数据
		// TODO: 后续需要刷一下数据，减少代码复杂度
		if app.Type == entity.AppTypeService && app.ServiceExposeType == "" {
			appDetail.ServiceExposeType = entity.AppServiceExposeTypeIngress
		}

		// 校验是否允许创建任务
		err = validateIsAllowCreateTask(ctx, createReq, project, appDetail, operatorID)
		if err != nil {
			errGroup = errGroup.AddChildren(errors.Wrap(err, "app_name="+app.Name))
			continue
		}

		// Set empty.
		createReq.Approval, createReq.DeployType, createReq.ScheduleTime = new(req.ApprovalReq), "", 0

		createTaskReqMap[app.ID] = createReq
	}

	if len(errGroup.Details()) > 0 {
		return nil, errGroup
	}

	return createTaskReqMap, nil
}

func validateIsAllowCreateCommonTask(_ context.Context, createReq *req.CreateTaskReq,
	appType entity.AppType, action entity.TaskAction) error {
	if appType == entity.AppTypeOneTimeJob &&
		(action == entity.TaskActionRestart || action == entity.TaskActionResume) {
		return errors.Wrap(errcode.InvalidParams, "action is unsupported for one time job")
	}

	if action == entity.TaskActionFullCanaryDeploy && createReq.Param.MinPodCount == 0 {
		return errors.Wrap(errcode.InvalidParams, "min_pod_count shouldn't be 0")
	}

	if action == entity.TaskActionResume &&
		appType != entity.AppTypeCronJob && createReq.Param.MinPodCount == 0 {
		return errors.Wrap(errcode.InvalidParams, "min_pod_count shouldn't be 0")
	}

	return nil
}

// validateReloadConfig validate reload config
func validateReloadConfig(ctx context.Context, createReq *req.CreateTaskReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) error {
	_, err := service.SVC.GetCompatibleConfigMapDetail(ctx, project, app, &resp.TaskDetailResp{
		ClusterName: createReq.ClusterName,
		EnvName:     createReq.EnvName,
		Version:     createReq.Version,
	})
	if err != nil {
		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return errors.Wrap(errcode.InvalidParams, "action is not supported because configMap does not exist")
		}
		return err
	}

	return nil
}

// validateApprovalParams validates approval params.
func validateApprovalParams(ctx context.Context, createReq *req.CreateTaskReq, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, operatorID string) error {
	if createReq.DeployType == "" {
		return nil
	}

	// Validate whether user sets scheduleTime and scheduleTime is before now when deploy type is scheduled.
	if createReq.DeployType == entity.ScheduledTaskDeployType {
		if createReq.ScheduleTime <= time.Now().Unix() {
			return errors.Wrap(errcode.InvalidParams, "schedule time is before now")
		}
	}

	if createReq.Approval.Type != entity.SkipTaskApprovalType {
		if createReq.Action != entity.TaskActionFullDeploy && createReq.Action != entity.TaskActionCanaryDeploy {
			return errors.Wrap(errcode.InvalidParams, "task cant not create approval")
		}
		if operatorID == entity.K8sSystemUserID {
			return errors.Wrapf(errcode.InvalidParams, "system user can not create approval")
		}
		if createReq.Description == "" {
			return errors.Wrap(errcode.InvalidParams, "description is empty")
		}
		// Validate approval users.
		if len(project.Owners) == 0 {
			return errors.Wrap(_errcode.ApproversNotExistsError, "approval user owners don't exist")
		}
		if len(project.QAEngineers)+len(createReq.Approval.QAEngineers) == 0 {
			return errors.Wrap(_errcode.ApproversNotExistsError, "approval user QA engineers don't exist")
		}
		if len(project.ProductManagers)+len(createReq.Approval.ProductManagers) == 0 {
			return errors.Wrap(_errcode.ApproversNotExistsError, "approval user product managers don't exist")
		}
		if len(project.OperationEngineers)+len(createReq.Approval.OperationEngineers) == 0 {
			return errors.Wrap(_errcode.ApproversNotExistsError, "approval user operation engineers don't exist")
		}

		approvingTasks, err := service.SVC.GetTasks(ctx, &req.GetTasksReq{
			AppID:              app.ID,
			EnvName:            createReq.EnvName,
			ClusterName:        createReq.ClusterName,
			StatusInverseList:  entity.TaskStatusFinalStateList,
			Suspend:            null.BoolFrom(false),
			ApprovalStatusList: entity.TaskApprovalStatusApprovingList,
		})
		if err != nil {
			return err
		}
		// Validate if only one task is being approving now.
		if len(approvingTasks) > 0 {
			return errors.Wrapf(_errcode.OtherApprovalProcessExistError, "other task is being approving")
		}
	}

	if createReq.EnvName != entity.AppEnvPrd {
		return nil
	}

	// Prd task supports only two deployment type:
	// 1.urgent deployment(deployType=immediate && approval.type=skip), and only project owners have permission.
	// 2.manual deployment with approval.
	switch createReq.DeployType {
	case entity.ImmediateTaskDeployType:
		// User can only create urgent deployment task when env is prd and type is immediate.
		if createReq.Approval.Type != entity.SkipTaskApprovalType {
			return errors.Wrap(errcode.InvalidParams, "task does not need approval process")
		}
		// Only owners can create urgent task without approval process.
		if service.SVC.IsP0LevelProject(project) {
			isOwner := operatorID == entity.K8sSystemUserID
			for _, owner := range project.Owners {
				if owner.ID == operatorID {
					isOwner = true
					break
				}
			}
			if !isOwner {
				return errors.Wrap(_errcode.CreateTaskNoPermissionError, "no permission to create urgently deploy task")
			}
		}
	case entity.ManualTaskDeployType:
		// User can't create manual task without approval process.
		if createReq.Approval.Type == entity.SkipTaskApprovalType {
			return errors.Wrap(errcode.InvalidParams, "task needs approval process")
		}
	case entity.ScheduledTaskDeployType:
		return errors.Wrapf(errcode.InvalidParams, "can't create %s task when env is prd", createReq.DeployType)
	default:
		return errors.Wrap(errcode.InvalidParams, "invalid deploy type")
	}

	return nil
}

func CheckTask(ctx *gin.Context) {
	id := ctx.Param("id")

	if id == "" || id == "latest" { // 过滤特殊通配路由
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		ctx.Abort()
		response.JSON(ctx, nil, errors.Wrap(_errcode.NotFoundError, err.Error()))

		return
	}

	_, err = service.SVC.GetTaskByObjectID(ctx, objectID)

	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		ctx.Abort()
		response.JSON(ctx, nil, errors.Wrap(_errcode.NotFoundError, "task id 不存在"))

		return
	}
	if err != nil {
		ctx.Abort()
		response.JSON(ctx, nil, err)

		return
	}
}

func checkAndUnifyConfigParams(param *req.CreateTaskParamReq) error {
	if param.ConfigRenamePrefix == "" {
		if param.ConfigRenameMode == 0 {
			return nil
		}

		return errors.Wrapf(errcode.InvalidParams, "config_rename_mode(%d) without config_rename_prefix", param.ConfigRenameMode)
	}

	if param.ConfigRenameMode == 0 {
		// 当前只有一种模式, 设为默认值
		param.ConfigRenameMode = entity.ConfigRenameModeExact
	}

	valid := false
	for _, m := range entity.SupportedConfigRenameModes {
		if param.ConfigRenameMode == m {
			valid = true
			break
		}
	}

	if !valid {
		return errors.Wrapf(errcode.InvalidParams, "unsupported config_rename_mode(%d)", param.ConfigRenameMode)
	}

	return nil
}

func checkTaskAction(env entity.AppEnvName, action entity.TaskAction, apps []*resp.AppListResp) error {
	var actionMap = map[entity.AppType]entity.TaskActionList{
		entity.AppTypeService:    entity.TaskActionServiceBatchList,
		entity.AppTypeWorker:     entity.TaskActionWorkerBatchList,
		entity.AppTypeCronJob:    entity.TaskActionCronJobBatchList,
		entity.AppTypeOneTimeJob: entity.TaskActionOneTimeJobBatchList,
	}

	for _, app := range apps {
		valid := false
		for _, v := range actionMap[app.Type] {
			if action == v {
				valid = true
				break
			}
		}

		if !valid {
			return errors.Wrapf(errcode.InvalidParams, "%s not support action: %s", app.Type, action)
		}

		if app.Type == entity.AppTypeWorker || app.Type == entity.AppTypeService {
			switch env {
			case entity.AppEnvStg, entity.AppEnvFat:
				if action == entity.TaskActionCanaryDeploy {
					valid = false
				}

			case entity.AppEnvPre, entity.AppEnvPrd:
				if action == entity.TaskActionFullDeploy {
					valid = false
				}
			}

			if !valid {
				return errors.Wrapf(errcode.InvalidParams, "%s env %s not support action: %s", env, app.Type, action)
			}
		}
	}

	return nil
}
