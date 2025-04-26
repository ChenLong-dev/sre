package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/null"
	"gitlab.shanhai.int/sre/library/goroutine"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"

	"rulai/config"
	"rulai/models"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
)

// 获取pod的web shell跳转url
func (s *Service) getPodShellURL(namespace, podName string) string {
	clusterID := config.Conf.Other.K8sStgClusterID
	if namespace == string(entity.AppEnvPrd) || namespace == string(entity.AppEnvPre) {
		clusterID = config.Conf.Other.K8sPrdClusterID
	}
	return fmt.Sprintf(
		"%s/%s/#!/shell/%s/%s/?namespace=%s",
		config.Conf.Other.AliK8sConsoleURL, clusterID, namespace, podName, namespace,
	)
}

func (s *Service) getGrafanaPodPath(namespace string, clusterName entity.ClusterName) (string, error) {
	clusters := config.Conf.K8sClusters[namespace]

	for _, cluster := range clusters {
		if cluster.Name == string(clusterName) {
			return fmt.Sprintf("%s%s", cluster.GrafanaHost, cluster.GrafanaPodPath), nil
		}
	}

	return "", errors.Wrapf(errcode.InvalidParams, "no grafana pod path found for namespace[%s], cluster[%s]", namespace, clusterName)
}

// 获取pod的监控url
func (s *Service) getPodMonitorURL(namespace string, appType entity.AppType, version string,
	clusterName entity.ClusterName) (string, error) {
	workerType := "deployment"
	if appType == entity.AppTypeCronJob {
		workerType = "cronjob"
	} else if appType == entity.AppTypeOneTimeJob {
		workerType = "job"
	}

	dataSource := config.Conf.K8s.PrdContextName
	if namespace == string(entity.AppEnvStg) || namespace == string(entity.AppEnvFat) {
		dataSource = config.Conf.K8s.StgContextName
	}

	dataSource = fmt.Sprintf("%s-%s", clusterName, dataSource)

	grafanaPath, err := s.getGrafanaPodPath(strings.ReplaceAll(namespace, entity.IstioNamespacePrefix, ""), clusterName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s?var-datasource=%s&var-namespace=%s&var-workload=%s&var-type=%s",
		grafanaPath, dataSource, namespace, version, workerType), nil
}

// GetRunningStatusDeploymentList 获取部署过程中的 Deployment 列表
func (s *Service) GetRunningStatusDeploymentList(ctx context.Context,
	clusterName entity.ClusterName, getReq *req.GetRunningStatusListReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) ([]*resp.RunningStatusListResp, error) {
	// 获取所有尚未完成的部署任务
	deployTasks, err := s.GetTasks(ctx, &req.GetTasksReq{
		AppID:             app.ID,
		EnvName:           getReq.EnvName,
		ClusterName:       clusterName,
		ActionList:        entity.TaskActionInitDeployList,
		StatusInverseList: entity.TaskStatusFinalStateList,
		Suspend:           null.BoolFrom(false),
	})
	if err != nil {
		return nil, err
	}
	// 默认每个任务未创建deploy
	taskCheckMap := make(map[string]bool)
	for _, task := range deployTasks {
		taskCheckMap[task.Version] = false
	}

	// 获取已经创建的deploy
	deployments, err := s.GetDeployments(ctx, clusterName, getReq.EnvName, &req.GetDeploymentsReq{
		Namespace:   getReq.Namespace,
		ProjectName: project.Name,
		AppName:     app.Name,
		Env:         string(getReq.EnvName),
	})
	if err != nil {
		return nil, err
	}

	// 获取已经创建的istio deploy
	list, e := s.GetDeployments(ctx, clusterName, getReq.EnvName, &req.GetDeploymentsReq{
		Namespace:   string(entity.IstioNamespacePrefix + getReq.EnvName),
		ProjectName: project.Name,
		AppName:     app.Name,
		Env:         string(getReq.EnvName),
	})
	if e != nil {
		return nil, e
	}
	// 拼接已部署的task
	deployments = append(deployments, list...)

	createdDeployList := make([]*resp.RunningStatusListResp, len(deployments))
	for i := range deployments {
		deploy := deployments[i]

		// 获取每个版本的上一次任务
		task, e := s.GetSingleTask(ctx, &req.GetTasksReq{
			AppID:             app.ID,
			EnvName:           getReq.EnvName,
			ClusterName:       clusterName,
			Version:           deploy.GetName(),
			ActionInverseList: entity.TaskActionSystemList,
		})
		if e != nil {
			return nil, e
		}
		// 获取上次部署的任务
		lastDeployTask, e := s.GetLatestDeploySuccessTaskFinalVersion(ctx, &req.GetLatestTaskReq{
			AppID:        getReq.AppID,
			EnvName:      getReq.EnvName,
			ClusterName:  clusterName,
			Version:      deploy.GetName(),
			IgnoreStatus: true,
		})
		if e != nil {
			return nil, e
		}

		monitorURL, e := s.getPodMonitorURL(deploy.GetNamespace(), app.Type, deploy.GetName(), clusterName)
		if e != nil {
			return nil, e
		}

		curResp := &resp.RunningStatusListResp{
			Version:           deploy.Name,
			CreateTime:        lastDeployTask.CreateTime,
			TaskID:            task.ID,
			TaskStatus:        string(task.Status),
			TaskStatusDisplay: task.StatusDisplay,
			TaskDisplayIcon:   task.DisplayIcon,
			TaskRetryCount:    task.RetryCount,
			TaskSuspend:       task.Suspend,
			PodMonitorURL:     monitorURL,
			ConfigURL:         lastDeployTask.Param.ConfigURL,
			ReadyPodCount:     int(deploy.Status.ReadyReplicas),
			TotalPodCount:     int(deploy.Status.Replicas),
			Namespace:         deploy.Namespace,
		}

		if len(deploy.Spec.Template.Spec.Containers) > 0 {
			curResp.ImageVersion = deploy.Spec.Template.Spec.Containers[0].Image
		}
		// 修改校验map
		taskCheckMap[deploy.Name] = true

		createdDeployList[i] = curResp
	}

	// 仍未创建deploy的部署任务，添加至结果集
	res := make([]*resp.RunningStatusListResp, 0)
	for _, lastDeployTask := range deployTasks {
		if taskCheckMap[lastDeployTask.Version] {
			continue
		}

		// 获取每个版本的上一次任务
		task, e := s.GetSingleTask(ctx, &req.GetTasksReq{
			AppID:             app.ID,
			EnvName:           getReq.EnvName,
			ClusterName:       clusterName,
			Version:           lastDeployTask.Version,
			ActionInverseList: entity.TaskActionSystemList,
		})
		if e != nil {
			return nil, e
		}

		res = append(res, &resp.RunningStatusListResp{
			Version:           lastDeployTask.Version,
			CreateTime:        lastDeployTask.CreateTime,
			TaskID:            task.ID,
			TaskStatus:        string(task.Status),
			TaskStatusDisplay: task.StatusDisplay,
			TaskDisplayIcon:   task.DisplayIcon,
			TaskRetryCount:    task.RetryCount,
			TaskSuspend:       task.Suspend,
			ConfigURL:         lastDeployTask.Param.ConfigURL,
			ReadyPodCount:     0,
			TotalPodCount:     lastDeployTask.Param.MaxPodCount,
			ImageVersion:      lastDeployTask.Param.ImageVersion,
			Namespace:         task.Namespace,
			NetworkTraffic:    false,
		})
	}
	res = append(res, createdDeployList...)

	failedStatus, err := s.getLatestFailedRunningStatus(ctx, clusterName, getReq.EnvName, app.ID, app.Type, string(getReq.EnvName))
	if err != nil {
		return nil, err
	}
	if failedStatus != nil {
		res = append([]*resp.RunningStatusListResp{failedStatus}, res...)
	}

	return res, nil
}

// GetRunningStatusCronJobList 获取部署过程中的 CronJob 列表
func (s *Service) GetRunningStatusCronJobList(ctx context.Context,
	clusterName entity.ClusterName, getReq *req.GetRunningStatusListReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) ([]*resp.RunningStatusListResp, error) {
	// 获取所有尚未完成的部署任务
	deployTasks, err := s.GetTasks(ctx, &req.GetTasksReq{
		AppID:             app.ID,
		EnvName:           getReq.EnvName,
		ClusterName:       clusterName,
		ActionList:        entity.TaskActionInitDeployList,
		StatusInverseList: entity.TaskStatusFinalStateList,
		Suspend:           null.BoolFrom(false),
	})
	if err != nil {
		return nil, err
	}
	// 默认每个任务未创建cronjob
	taskCheckMap := make(map[string]bool)
	for _, task := range deployTasks {
		taskCheckMap[task.Version] = false
	}

	// 获取已经创建的cronjob
	cronjobs, err := s.GetCronJobs(ctx, clusterName,
		&req.GetCronJobsReq{
			Namespace:   string(getReq.EnvName),
			ProjectName: project.Name,
			AppName:     app.Name,
			Env:         string(getReq.EnvName),
		})
	if err != nil {
		return nil, err
	}
	createdCronJobList := make([]*resp.RunningStatusListResp, len(cronjobs))
	for i := range cronjobs {
		cronjob := cronjobs[i]

		// 获取每个版本的上一次任务
		task, e := s.GetSingleTask(ctx, &req.GetTasksReq{
			AppID:             app.ID,
			EnvName:           getReq.EnvName,
			ClusterName:       clusterName,
			Version:           cronjob.GetName(),
			ActionInverseList: entity.TaskActionSystemList,
		})
		if e != nil {
			return nil, e
		}
		// 获取上次部署的任务
		lastDeployTask, e := s.GetLatestDeploySuccessTaskFinalVersion(ctx, &req.GetLatestTaskReq{
			AppID:        getReq.AppID,
			EnvName:      getReq.EnvName,
			ClusterName:  clusterName,
			Version:      cronjob.GetName(),
			IgnoreStatus: true,
		})
		if e != nil {
			return nil, e
		}

		curResp := &resp.RunningStatusListResp{
			Version:           cronjob.Name,
			CreateTime:        lastDeployTask.CreateTime,
			TaskID:            task.ID,
			TaskStatus:        string(task.Status),
			TaskStatusDisplay: task.StatusDisplay,
			TaskDisplayIcon:   task.DisplayIcon,
			TaskRetryCount:    task.RetryCount,
			TaskSuspend:       task.Suspend,
			ConfigURL:         lastDeployTask.Param.ConfigURL,
			CronParam:         cronjob.Spec.Schedule,
			IsSuspend:         *cronjob.Spec.Suspend,
			Namespace:         task.Namespace,
		}
		if len(cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers) > 0 {
			curResp.ImageVersion = cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image
		}
		if cronjob.Status.LastScheduleTime != nil {
			curResp.LastScheduleTime = cronjob.Status.LastScheduleTime.Format(utils.DefaultTimeFormatLayout)
		}

		earliestTime := cronjob.GetCreationTimestamp().Time
		if cronjob.Status.LastScheduleTime != nil {
			earliestTime = cronjob.Status.LastScheduleTime.Time
		}
		nextSched, e := s.getCronjobNextSchedTime(cronjob.Spec.Schedule, earliestTime)
		if e != nil {
			return nil, e
		}
		curResp.NextScheduleTime = nextSched

		// 修改校验map
		taskCheckMap[cronjob.Name] = true

		createdCronJobList[i] = curResp
	}

	// 仍未创建cronjob的部署任务，添加至结果集
	res := make([]*resp.RunningStatusListResp, 0)
	for _, lastDeployTask := range deployTasks {
		if taskCheckMap[lastDeployTask.Version] {
			continue
		}

		// 获取每个版本的上一次任务
		task, e := s.GetSingleTask(ctx, &req.GetTasksReq{
			AppID:             app.ID,
			EnvName:           getReq.EnvName,
			ClusterName:       clusterName,
			Version:           lastDeployTask.Version,
			ActionInverseList: entity.TaskActionSystemList,
		})
		if e != nil {
			return nil, e
		}

		res = append(res, &resp.RunningStatusListResp{
			Version:           lastDeployTask.Version,
			CreateTime:        lastDeployTask.CreateTime,
			TaskID:            task.ID,
			TaskStatus:        string(task.Status),
			TaskStatusDisplay: task.StatusDisplay,
			TaskDisplayIcon:   task.DisplayIcon,
			TaskRetryCount:    task.RetryCount,
			TaskSuspend:       task.Suspend,
			ConfigURL:         lastDeployTask.Param.ConfigURL,
			CronParam:         lastDeployTask.Param.CronParam,
			ImageVersion:      lastDeployTask.Param.ImageVersion,
			IsSuspend:         false,
			Namespace:         task.Namespace,
		})
	}
	res = append(res, createdCronJobList...)
	namespace := s.GetNamespaceBase(s.GetApplicationIstioState(ctx, getReq.EnvName, getReq.ClusterName, app), getReq.EnvName)
	failedStatus, err := s.getLatestFailedRunningStatus(ctx, clusterName, getReq.EnvName, app.ID, app.Type, namespace)
	if err != nil {
		return nil, err
	}
	if failedStatus != nil {
		res = append([]*resp.RunningStatusListResp{failedStatus}, res...)
	}

	return res, nil
}

// GetRunningStatusJobList 获取部署过程中的 Job 列表
func (s *Service) GetRunningStatusJobList(ctx context.Context,
	clusterName entity.ClusterName, getReq *req.GetRunningStatusListReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) ([]*resp.RunningStatusListResp, error) {
	// 获取所有尚未完成的部署任务
	deployTasks, err := s.GetTasks(ctx, &req.GetTasksReq{
		AppID:             app.ID,
		EnvName:           getReq.EnvName,
		ClusterName:       clusterName,
		ActionList:        entity.TaskActionInitDeployList,
		StatusInverseList: entity.TaskStatusFinalStateList,
		Suspend:           null.BoolFrom(false),
	})
	if err != nil {
		return nil, err
	}
	// 默认每个任务未创建job
	taskCheckMap := make(map[string]bool)
	for _, task := range deployTasks {
		taskCheckMap[task.Version] = false
	}

	// 获取已经创建的job
	jobs, err := s.GetJobs(ctx, clusterName,
		&req.GetJobsReq{
			Namespace:   string(getReq.EnvName),
			ProjectName: project.Name,
			AppName:     app.Name,
			Env:         string(getReq.EnvName),
		})
	if err != nil {
		return nil, err
	}
	createdJobList := make([]*resp.RunningStatusListResp, len(jobs))
	for i := range jobs {
		job := jobs[i]

		// 获取每个版本的上一次任务
		task, e := s.GetSingleTask(ctx, &req.GetTasksReq{
			AppID:             app.ID,
			EnvName:           getReq.EnvName,
			ClusterName:       clusterName,
			Version:           job.GetName(),
			ActionInverseList: entity.TaskActionSystemList,
		})
		if e != nil {
			return nil, e
		}
		// 获取上次部署的任务
		lastDeployTask, e := s.GetSingleTask(ctx, &req.GetTasksReq{
			BaseListRequest: models.BaseListRequest{
				Page:  1,
				Limit: 1,
			},
			AppID:             getReq.AppID,
			EnvName:           getReq.EnvName,
			ClusterName:       clusterName,
			Version:           job.GetName(),
			ActionList:        entity.TaskActionInitDeployList,
			ActionInverseList: entity.TaskActionSystemList,
		})
		if e != nil {
			return nil, e
		}

		monitorURL, e := s.getPodMonitorURL(job.GetNamespace(), app.Type, job.GetName(), clusterName)
		if e != nil {
			return nil, e
		}

		curResp := &resp.RunningStatusListResp{
			Version:           job.Name,
			CreateTime:        lastDeployTask.CreateTime,
			TaskID:            task.ID,
			TaskStatus:        string(task.Status),
			TaskStatusDisplay: task.StatusDisplay,
			TaskDisplayIcon:   task.DisplayIcon,
			TaskRetryCount:    task.RetryCount,
			TaskSuspend:       task.Suspend,
			PodMonitorURL:     monitorURL,
			ConfigURL:         lastDeployTask.Param.ConfigURL,
			Namespace:         task.Namespace,
		}
		if len(job.Spec.Template.Spec.Containers) > 0 {
			curResp.ImageVersion = job.Spec.Template.Spec.Containers[0].Image
		}

		// 修改校验map
		taskCheckMap[job.Name] = true

		createdJobList[i] = curResp
	}

	// 仍未创建job的部署任务，添加至结果集
	res := make([]*resp.RunningStatusListResp, 0)
	for _, lastDeployTask := range deployTasks {
		if taskCheckMap[lastDeployTask.Version] {
			continue
		}

		// 获取每个版本的上一次任务
		task, e := s.GetSingleTask(ctx, &req.GetTasksReq{
			AppID:             app.ID,
			EnvName:           getReq.EnvName,
			ClusterName:       clusterName,
			Version:           lastDeployTask.Version,
			ActionInverseList: entity.TaskActionSystemList,
		})
		if e != nil {
			return nil, e
		}

		res = append(res, &resp.RunningStatusListResp{
			Version:           lastDeployTask.Version,
			CreateTime:        lastDeployTask.CreateTime,
			TaskID:            task.ID,
			TaskStatus:        string(task.Status),
			TaskStatusDisplay: task.StatusDisplay,
			TaskDisplayIcon:   task.DisplayIcon,
			TaskRetryCount:    task.RetryCount,
			TaskSuspend:       task.Suspend,
			ConfigURL:         lastDeployTask.Param.ConfigURL,
			ImageVersion:      lastDeployTask.Param.ImageVersion,
			Namespace:         task.Namespace,
		})
	}
	res = append(res, createdJobList...)
	namespace := s.GetNamespaceBase(s.GetApplicationIstioState(ctx, getReq.EnvName, getReq.ClusterName, app), getReq.EnvName)
	failedStatus, err := s.getLatestFailedRunningStatus(ctx, clusterName, getReq.EnvName, app.ID, app.Type, namespace)
	if err != nil {
		return nil, err
	}
	if failedStatus != nil {
		res = append([]*resp.RunningStatusListResp{failedStatus}, res...)
	}

	return res, nil
}

// GetRunningStatusList 获取运行状态列表
func (s *Service) GetRunningStatusList(ctx context.Context,
	clusterName entity.ClusterName, getReq *req.GetRunningStatusListReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) ([]*resp.RunningStatusListResp, error) {
	if app.Type == entity.AppTypeWorker || app.Type == entity.AppTypeService {
		return s.GetRunningStatusDeploymentList(ctx, clusterName, getReq, project, app)
	} else if app.Type == entity.AppTypeCronJob {
		return s.GetRunningStatusCronJobList(ctx, clusterName, getReq, project, app)
	} else {
		return s.GetRunningStatusJobList(ctx, clusterName, getReq, project, app)
	}
}

// GetRunningStatusDeploymentDetail 获取部署过程中的 Deployment 详情
func (s *Service) GetRunningStatusDeploymentDetail(ctx context.Context, getReq *req.GetRunningStatusDetailReq,
	app *resp.AppDetailResp, lastTask, lastDeployTask *resp.TaskDetailResp) (*resp.RunningStatusDetailResp, error) {
	res := &resp.RunningStatusDetailResp{
		TaskID:            lastTask.ID,
		TaskStatus:        string(lastTask.Status),
		TaskStatusDisplay: lastTask.StatusDisplay,
		TaskDisplayIcon:   lastTask.DisplayIcon,
		TaskDetail:        lastTask.Detail,
		TaskRetryCount:    lastTask.RetryCount,
		TaskSuspend:       lastTask.Suspend,
		ConfigURL:         lastDeployTask.Param.ConfigURL,
		DeploymentPods:    make([]*resp.RunningStatusPodDetailResp, 0),
		DeployType:        lastTask.DeployType,
		ScheduleTime:      lastTask.ScheduleTime,
		Approval:          lastTask.Approval,
	}

	// 兼容istio 命名空间 查询两个命名空间 env  和 istio-env
	namespace := s.GetNamespaceByApp(app, getReq.EnvName, getReq.ClusterName)
	deploy, err := s.GetDeploymentDetail(ctx, getReq.ClusterName, lastTask.EnvName, &req.GetDeploymentDetailReq{
		Env:       string(getReq.EnvName),
		Name:      getReq.Version,
		Namespace: getReq.Namespace,
	})

	if err != nil {
		// 仍未创建deploy时，使用部署任务参数
		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) &&
			lastDeployTask.Status != entity.TaskStatusSuccess &&
			lastDeployTask.Status != entity.TaskStatusFail {
			res.Version = lastDeployTask.Version
			res.CreateTime = lastDeployTask.CreateTime
			res.ReadyPodCount = 0
			res.TotalPodCount = lastDeployTask.Param.MaxPodCount
			return res, nil
		}
		return nil, err
	}

	res.Version = deploy.Name
	res.CreateTime = deploy.GetCreationTimestamp().Format(utils.DefaultTimeFormatLayout)
	res.PodMonitorURL, err = s.getPodMonitorURL(deploy.GetNamespace(), app.Type, deploy.GetName(), getReq.ClusterName)
	if err != nil {
		return nil, err
	}

	res.ReadyPodCount = int(deploy.Status.ReadyReplicas)
	res.TotalPodCount = int(deploy.Status.Replicas)
	if len(deploy.Spec.Template.Spec.Containers) > 0 {
		res.ImageVersion = deploy.Spec.Template.Spec.Containers[0].Image
	}

	pods, err := s.GetPods(ctx, getReq.ClusterName,
		&req.GetPodsReq{
			Namespace: "", // 使用全局 namespace 检索
			Env:       string(getReq.EnvName),
			Version:   getReq.Version,
		})
	if err != nil {
		return nil, err
	}
	for i := range pods {
		pod := pods[i]

		restart := 0
		if len(pod.Status.ContainerStatuses) > 0 {
			restart = int(pod.Status.ContainerStatuses[0].RestartCount)
		}

		podResp := &resp.RunningStatusPodDetailResp{
			Name:         pod.GetName(),
			CreateTime:   pod.GetCreationTimestamp().Format(utils.DefaultTimeFormatLayout),
			RestartCount: restart,
			ShellURL:     s.getPodShellURL(pod.GetNamespace(), pod.GetName()),
			Phase:        pod.Status.Phase,
			NodeIP:       pod.Status.HostIP,
			PodIP:        pod.Status.PodIP,
			Namespace:    pod.GetNamespace(),
		}
		if pod.Status.StartTime != nil {
			podResp.Age = time.Since(pod.Status.StartTime.Time).Round(time.Second).String()
		}
		res.DeploymentPods = append(res.DeploymentPods, podResp)
	}

	if len(lastTask.Param.CronScaleJobGroups) > 0 {
		// 获取cronHPA详情
		cronHPA, e := s.GetCronHPADetail(ctx, getReq.ClusterName, &req.GetCronHPADetailReq{
			Namespace: namespace,
			Name:      getReq.Version,
		})
		if e != nil && !errcode.EqualError(_errcode.K8sResourceNotFoundError, e) {
			return nil, e
		}
		if cronHPA != nil {
			res.CronAutoScaleJobs = cronHPA.Status.Conditions
		}
	}

	return res, nil
}

// GetVersionAllowedActions 用于返回给前端某个具体 version 有哪些允许的操作，三个部署操作是否允许暂不考虑
func (s *Service) GetVersionAllowedActions(ctx context.Context, app *resp.AppDetailResp, lastTask *resp.TaskDetailResp) (
	map[entity.TaskAction]bool, error) {
	allowedActions := s.getDefaultAllowedActions(ctx)

	// 检查是否有其他未完成任务，如果有，所有操作都不被允许
	// FIXME: 当前页面不按照多集群区分显示，故无法区分未完成任务是否在目标集群中
	// 实际上非本集群的任务并不会影响到本集群的操作
	// 一旦页面按照多集群区分显示之后，GetTasksReq条件中应加入 ClusterName
	unFinishedCount, err := s.GetTasksCount(ctx, &req.GetTasksReq{
		AppID:             app.ID,
		EnvName:           lastTask.EnvName,
		StatusInverseList: entity.TaskStatusFinalStateList,
		Suspend:           null.BoolFrom(false),
	})
	if err != nil {
		return nil, errors.Wrapf(errcode.InvalidParams, "%s", err)
	}

	if unFinishedCount > 0 {
		return allowedActions, nil
	}
	// 根据应用类型，允许一些操作
	switch app.Type {
	case entity.AppTypeService, entity.AppTypeWorker:
		if lastTask.Action == entity.TaskActionStop {
			allowedActions[entity.TaskActionResume] = true
		} else {
			allowedActions[entity.TaskActionStop] = true
		}
		allowedActions[entity.TaskActionRestart] = true
		allowedActions[entity.TaskActionDelete] = true
		allowedActions[entity.TaskActionUpdateHPA] = true
		allowedActions[entity.TaskActionReloadConfig] = true
	case entity.AppTypeCronJob:
		if lastTask.Action == entity.TaskActionStop {
			allowedActions[entity.TaskActionResume] = true
		} else {
			allowedActions[entity.TaskActionStop] = true
		}
		allowedActions[entity.TaskActionDelete] = true
		allowedActions[entity.TaskActionManualLaunch] = true
	case entity.AppTypeOneTimeJob:
		allowedActions[entity.TaskActionDelete] = true
	}

	// 根据当前的 task 状态，不允许一些操作
	switch lastTask.Action {
	case entity.TaskActionFullDeploy, entity.TaskActionCanaryDeploy, entity.TaskActionFullCanaryDeploy,
		entity.TaskActionRestart, entity.TaskActionResume, entity.TaskActionManualLaunch, entity.TaskActionUpdateHPA,
		entity.TaskActionReloadConfig:
		if lastTask.Suspend {
			allowedActions[entity.TaskActionStop] = false
			allowedActions[entity.TaskActionRestart] = false
			allowedActions[entity.TaskActionResume] = false
			allowedActions[entity.TaskActionManualLaunch] = false
			allowedActions[entity.TaskActionUpdateHPA] = false
			allowedActions[entity.TaskActionReloadConfig] = false
		}
	case entity.TaskActionStop:
		allowedActions[entity.TaskActionStop] = false
		allowedActions[entity.TaskActionUpdateHPA] = false
		allowedActions[entity.TaskActionReloadConfig] = false

	case entity.TaskActionDelete:
		allowedActions = s.getDefaultAllowedActions(ctx)

	default:
		allowedActions = s.getDefaultAllowedActions(ctx)
		log.Errorc(ctx, "unknown action: %s", lastTask.Action)
	}
	return allowedActions, nil
}

// GetRunningStatusCronJobDetail 获取部署过程中的 CronJob 详情
func (s *Service) GetRunningStatusCronJobDetail(ctx context.Context, getReq *req.GetRunningStatusDetailReq,
	app *resp.AppDetailResp, lastTask, lastDeployTask *resp.TaskDetailResp) (*resp.RunningStatusDetailResp, error) {
	res := &resp.RunningStatusDetailResp{
		TaskID:            lastTask.ID,
		TaskStatus:        string(lastTask.Status),
		TaskStatusDisplay: lastTask.StatusDisplay,
		TaskDisplayIcon:   lastTask.DisplayIcon,
		TaskDetail:        lastTask.Detail,
		TaskRetryCount:    lastTask.RetryCount,
		TaskSuspend:       lastTask.Suspend,
		ConfigURL:         lastDeployTask.Param.ConfigURL,
		Jobs:              make([]*resp.RunningStatusJobDetailResp, 0),
		DeployType:        lastTask.DeployType,
		ScheduleTime:      lastTask.ScheduleTime,
		Approval:          lastTask.Approval,
	}

	cronjob, err := s.GetCronJobDetail(ctx, getReq.ClusterName,
		&req.GetCronJobDetailReq{
			Namespace: getReq.Namespace,
			Name:      getReq.Version,
			Env:       string(getReq.EnvName),
		})

	if err != nil {
		// 仍未创建cronjob时，使用部署任务参数
		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) &&
			lastDeployTask.Status != entity.TaskStatusSuccess &&
			lastDeployTask.Status != entity.TaskStatusFail {
			res.Version = lastDeployTask.Version
			res.CreateTime = lastDeployTask.CreateTime
			res.CronParam = lastDeployTask.Param.CronParam
			return res, nil
		}
		return nil, err
	}

	res.Version = cronjob.Name
	res.CreateTime = cronjob.GetCreationTimestamp().Format(utils.DefaultTimeFormatLayout)
	res.PodMonitorURL, err = s.getPodMonitorURL(cronjob.GetNamespace(), app.Type, cronjob.GetName(), getReq.ClusterName)
	if err != nil {
		return nil, err
	}

	res.CronParam = cronjob.Spec.Schedule
	res.IsSuspend = *cronjob.Spec.Suspend
	if cronjob.Status.LastScheduleTime != nil {
		res.LastScheduleTime = cronjob.Status.LastScheduleTime.Format(utils.DefaultTimeFormatLayout)
	}
	if len(cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers) > 0 {
		res.ImageVersion = cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image
	}

	earliestTime := cronjob.GetCreationTimestamp().Time
	if cronjob.Status.LastScheduleTime != nil {
		earliestTime = cronjob.Status.LastScheduleTime.Time
	}
	nextSched, err := s.getCronjobNextSchedTime(cronjob.Spec.Schedule, earliestTime)
	if err != nil {
		return nil, err
	}
	res.NextScheduleTime = nextSched

	var (
		jobs     []batch.Job
		job2Pods = make(map[string][]v1.Pod)
	)
	eg := goroutine.WithContext(ctx, "GetJobAndPods")
	eg.Go(ctx, "get jobs", func(ctx context.Context) error {
		var err error
		jobs, err = s.GetJobs(ctx, getReq.ClusterName,
			&req.GetJobsReq{
				Namespace: string(getReq.EnvName),
				Version:   cronjob.GetName(),
				Env:       string(getReq.EnvName),
			})
		return err
	})

	eg.Go(ctx, "get pods", func(ctx context.Context) error {
		pods, err := s.GetPods(ctx, getReq.ClusterName,
			&req.GetPodsReq{
				Namespace: getReq.Namespace,
				Version:   getReq.Version,
				Env:       string(getReq.EnvName),
			})
		if err != nil {
			return err
		}

		for _, v := range pods {
			if len(v.OwnerReferences) > 0 {
				job2Pods[v.OwnerReferences[0].Name] = append(job2Pods[v.OwnerReferences[0].Name], v)
			}
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	for i := range jobs {
		job := jobs[i]
		pods := job2Pods[job.Name]
		podRes := make([]resp.RunningStatusPodDetailResp, len(pods))
		for j := range pods {
			pod := pods[j]

			restart := 0
			if len(pod.Status.ContainerStatuses) > 0 {
				restart = int(pod.Status.ContainerStatuses[0].RestartCount)
			}

			curResp := resp.RunningStatusPodDetailResp{
				Name:         pod.GetName(),
				CreateTime:   pod.GetCreationTimestamp().Format(utils.DefaultTimeFormatLayout),
				RestartCount: restart,
				ShellURL:     s.getPodShellURL(pod.GetNamespace(), pod.GetName()),
				Phase:        pod.Status.Phase,
				NodeIP:       pod.Status.HostIP,
				PodIP:        pod.Status.PodIP,
				Namespace:    pod.GetNamespace(),
			}
			if pod.Status.StartTime != nil {
				curResp.Age = time.Since(pod.Status.StartTime.Time).Round(time.Second).String()
			}
			podRes[j] = curResp
		}

		jobResp := &resp.RunningStatusJobDetailResp{
			Name:              job.GetName(),
			SucceededCount:    int(job.Status.Succeeded),
			FailedCount:       int(job.Status.Failed),
			NeedCompleteCount: int(*job.Spec.Completions),
			LaunchType:        entity.LaunchType(job.GetLabels()[entity.LabelKeyLaunchType]),
			Pods:              podRes,
		}
		// 1.12集群在job执行超时的情况下由于race会出现job.status.failed被覆盖的情况
		// 目前NeedCompleteCount总会是1 之后可以考虑增加字段来直接标示一个job的状态running/failed/completed
		if jobResp.FailedCount == 0 && s.isJobFailed(&job) {
			jobResp.FailedCount = jobResp.NeedCompleteCount - jobResp.SucceededCount
		}
		if job.Status.StartTime != nil {
			jobResp.StartTime = job.Status.StartTime.Format(utils.DefaultTimeFormatLayout)
		}
		if job.Status.CompletionTime != nil {
			jobResp.CompletionTime = job.Status.CompletionTime.Format(utils.DefaultTimeFormatLayout)
		}
		res.Jobs = append(res.Jobs, jobResp)
	}
	return res, nil
}

// GetRunningStatusJobDetail 获取部署过程中的 Job 详情
func (s *Service) GetRunningStatusJobDetail(ctx context.Context, getReq *req.GetRunningStatusDetailReq,
	app *resp.AppDetailResp, lastTask, lastDeployTask *resp.TaskDetailResp) (*resp.RunningStatusDetailResp, error) {
	res := &resp.RunningStatusDetailResp{
		TaskID:            lastTask.ID,
		TaskStatus:        string(lastTask.Status),
		TaskStatusDisplay: lastTask.StatusDisplay,
		TaskDisplayIcon:   lastTask.DisplayIcon,
		TaskDetail:        lastTask.Detail,
		TaskRetryCount:    lastTask.RetryCount,
		TaskSuspend:       lastTask.Suspend,
		ConfigURL:         lastDeployTask.Param.ConfigURL,
		Jobs:              make([]*resp.RunningStatusJobDetailResp, 0),
		DeployType:        lastTask.DeployType,
		ScheduleTime:      lastTask.ScheduleTime,
		Approval:          lastTask.Approval,
	}

	job, err := s.GetJobDetail(ctx, getReq.ClusterName,
		&req.GetJobDetailReq{
			Namespace: string(getReq.EnvName),
			Name:      getReq.Version,
			Env:       string(getReq.EnvName),
		})
	if err != nil {
		// 仍未创建job时，使用部署任务参数
		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) &&
			lastDeployTask.Status != entity.TaskStatusSuccess &&
			lastDeployTask.Status != entity.TaskStatusFail {
			res.Version = lastDeployTask.Version
			res.CreateTime = lastDeployTask.CreateTime
			return res, nil
		}
		return nil, err
	}

	res.Version = job.Name
	res.CreateTime = job.GetCreationTimestamp().Format(utils.DefaultTimeFormatLayout)
	res.PodMonitorURL, err = s.getPodMonitorURL(job.GetNamespace(), app.Type, job.GetName(), getReq.ClusterName)
	if err != nil {
		return nil, err
	}

	pods, err := s.GetPods(ctx, getReq.ClusterName,
		&req.GetPodsReq{
			Namespace: getReq.Namespace,
			Version:   getReq.Version,
			JobName:   job.GetName(),
			Env:       string(getReq.EnvName),
		})
	if err != nil {
		return nil, err
	}
	podRes := make([]resp.RunningStatusPodDetailResp, len(pods))
	for j := range pods {
		pod := pods[j]

		restart := 0
		if len(pod.Status.ContainerStatuses) > 0 {
			restart = int(pod.Status.ContainerStatuses[0].RestartCount)
		}

		curResp := resp.RunningStatusPodDetailResp{
			Name:         pod.GetName(),
			CreateTime:   pod.GetCreationTimestamp().Format(utils.DefaultTimeFormatLayout),
			RestartCount: restart,
			ShellURL:     s.getPodShellURL(pod.GetNamespace(), pod.GetName()),
			Phase:        pod.Status.Phase,
			NodeIP:       pod.Status.HostIP,
			PodIP:        pod.Status.PodIP,
			Namespace:    pod.GetNamespace(),
		}
		if pod.Status.StartTime != nil {
			curResp.Age = time.Since(pod.Status.StartTime.Time).Round(time.Second).String()
		}
		podRes[j] = curResp
	}
	jobResp := &resp.RunningStatusJobDetailResp{
		Name:              job.GetName(),
		SucceededCount:    int(job.Status.Succeeded),
		FailedCount:       int(job.Status.Failed),
		NeedCompleteCount: int(*job.Spec.Completions),
		Pods:              podRes,
	}
	if jobResp.FailedCount == 0 && s.isJobFailed(job) {
		jobResp.FailedCount = jobResp.NeedCompleteCount - jobResp.SucceededCount
	}
	if job.Status.StartTime != nil {
		jobResp.StartTime = job.Status.StartTime.Format(utils.DefaultTimeFormatLayout)
	}
	if job.Status.CompletionTime != nil {
		jobResp.CompletionTime = job.Status.CompletionTime.Format(utils.DefaultTimeFormatLayout)
	}
	res.Jobs = append(res.Jobs, jobResp)
	return res, nil
}

// GetRunningStatusDetail 获取运行状态详情
func (s *Service) GetRunningStatusDetail(ctx context.Context, getReq *req.GetRunningStatusDetailReq,
	app *resp.AppDetailResp) (*resp.RunningStatusDetailResp, error) {
	var (
		// 获取上次的任务
		lastTask *resp.TaskDetailResp
		// 获取上次部署的任务
		lastDeployTask *resp.TaskDetailResp
		// 允许的操作
		allowedActions map[entity.TaskAction]bool
		// 运行状态
		runningStatusRes *resp.RunningStatusDetailResp
		err              error
	)

	eg := goroutine.WithContext(ctx, "GetRunningStatusDetail:"+getReq.Version)

	eg.Go(ctx, "last task", func(ctx context.Context) error {
		task, e := s.GetSingleTask(ctx, &req.GetTasksReq{
			AppID:             getReq.AppID,
			EnvName:           getReq.EnvName,
			ClusterName:       getReq.ClusterName,
			Version:           getReq.Version,
			ActionInverseList: entity.TaskActionSystemList,
		})
		if e != nil {
			return e
		}
		lastTask = task
		return nil
	})

	eg.Go(ctx, "last deploy task", func(ctx context.Context) error {
		task, e := s.GetLatestDeploySuccessTaskFinalVersion(ctx, &req.GetLatestTaskReq{
			AppID:        getReq.AppID,
			EnvName:      getReq.EnvName,
			ClusterName:  getReq.ClusterName,
			Version:      getReq.Version,
			IgnoreStatus: true,
		})
		if e != nil {
			return e
		}
		lastDeployTask = task
		return nil
	})

	err = eg.Wait()
	if err != nil {
		return nil, err
	}

	if lastTask == nil || lastDeployTask == nil {
		return nil, errors.Wrap(errcode.InvalidParams, "retrieved lastTask or lastDeployTask is empty")
	}

	if app.Type == entity.AppTypeService || app.Type == entity.AppTypeWorker {
		runningStatusRes, err = s.GetRunningStatusDeploymentDetail(ctx, getReq, app, lastTask, lastDeployTask)
	} else if app.Type == entity.AppTypeCronJob {
		runningStatusRes, err = s.GetRunningStatusCronJobDetail(ctx, getReq, app, lastTask, lastDeployTask)
	} else {
		runningStatusRes, err = s.GetRunningStatusJobDetail(ctx, getReq, app, lastTask, lastDeployTask)
	}

	if err != nil {
		// If task failed but Pod is not running, we must display it.
		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) &&
			lastDeployTask.Status == entity.TaskStatusFail {
			return &resp.RunningStatusDetailResp{
				TaskID:            lastDeployTask.ID,
				Version:           lastDeployTask.Version,
				CreateTime:        lastDeployTask.CreateTime,
				TaskStatus:        string(lastDeployTask.Status),
				TaskStatusDisplay: lastDeployTask.StatusDisplay,
				TaskDisplayIcon:   lastDeployTask.DisplayIcon,
				TaskDetail:        lastDeployTask.Detail,
				TaskRetryCount:    lastDeployTask.RetryCount,
				TaskSuspend:       lastDeployTask.Suspend,
				ConfigURL:         lastDeployTask.Param.ConfigURL,
				DeploymentPods:    make([]*resp.RunningStatusPodDetailResp, 0),
			}, nil
		}
		return nil, err
	}

	allowedActions, err = s.GetVersionAllowedActions(ctx, app, lastTask)
	if err != nil {
		return nil, err
	}

	runningStatusRes.AllowedActions = allowedActions

	return runningStatusRes, nil
}

// CheckRunningStatusExists 校验是否存在正在运行的部署
func (s *Service) CheckRunningStatusExists(ctx context.Context, cluster *resp.ClusterDetailResp,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) error {
	for envName := range app.Env {
		if cluster.Env != envName {
			continue
		}

		if app.Type == entity.AppTypeCronJob {
			jobs, err := s.GetCronJobs(ctx, cluster.Name,
				&req.GetCronJobsReq{
					Namespace:   string(envName),
					ProjectName: project.Name,
					AppName:     app.Name,
					Env:         string(envName),
				})
			if err != nil {
				return err
			}
			if len(jobs) > 0 {
				return errors.Wrap(errcode.InvalidParams, "存在未删除的任务")
			}
		} else if app.Type == entity.AppTypeOneTimeJob {
			jobs, err := s.GetJobs(ctx, cluster.Name,
				&req.GetJobsReq{
					Namespace:   string(envName),
					ProjectName: project.Name,
					AppName:     app.Name,
					Env:         string(envName),
				})
			if err != nil {
				return err
			}
			if len(jobs) > 0 {
				return errors.Wrap(errcode.InvalidParams, "存在未删除的任务")
			}
		} else {
			exists, err := s.CheckDeploymentsExistance(ctx, cluster.Name, envName, &req.GetDeploymentsReq{
				Namespace:   string(envName),
				ProjectName: project.Name,
				AppName:     app.Name,
				Env:         string(envName),
			})
			if err != nil {
				return err
			}
			if exists {
				return errors.Wrap(errcode.InvalidParams, "存在未删除的部署")
			}
		}
	}
	return nil
}

func (s *Service) getDefaultAllowedActions(_ context.Context) map[entity.TaskAction]bool {
	defaultAllowedActions := map[entity.TaskAction]bool{
		entity.TaskActionStop:         false,
		entity.TaskActionRestart:      false,
		entity.TaskActionResume:       false,
		entity.TaskActionDelete:       false,
		entity.TaskActionManualLaunch: false,
		entity.TaskActionUpdateHPA:    false,
		entity.TaskActionReloadConfig: false,
	}

	return defaultAllowedActions
}

// getLatestFailedRunningStatus find latest failed running status which pod is not running.
func (s *Service) getLatestFailedRunningStatus(ctx context.Context, clusterName entity.ClusterName,
	envName entity.AppEnvName, appID string, appType entity.AppType, ns string) (*resp.RunningStatusListResp, error) {
	task, err := s.GetSingleTask(ctx, &req.GetTasksReq{
		AppID:       appID,
		EnvName:     envName,
		ClusterName: clusterName,
		StatusList:  entity.TaskStatusFinalStateList,
		Suspend:     null.BoolFrom(false),
	})
	if err != nil {
		// No running status returns when no history task has been created.
		if errcode.EqualError(_errcode.NoRequiredTaskError, err) {
			return nil, nil
		}
		return nil, err
	}
	if task.Status != entity.TaskStatusFail {
		return nil, nil
	}
	// 要求部署类操作
	isDeployAction := false
	for _, deployAction := range entity.TaskActionInitDeployList {
		if task.Action == deployAction {
			isDeployAction = true
			break
		}
	}
	if !isDeployAction {
		return nil, nil
	}

	exist, err := s.IsAppK8sPrimaryResourceExist(ctx, clusterName, envName, appType, task.Version, ns)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, nil
	}

	return &resp.RunningStatusListResp{
		Version:           task.Version,
		CreateTime:        task.CreateTime,
		TaskID:            task.ID,
		TaskStatus:        string(task.Status),
		TaskStatusDisplay: task.StatusDisplay,
		TaskDisplayIcon:   task.DisplayIcon,
		TaskRetryCount:    task.RetryCount,
		TaskSuspend:       task.Suspend,
		ConfigURL:         task.Param.ConfigURL,
		ReadyPodCount:     0,
		TotalPodCount:     task.Param.MaxPodCount,
		ImageVersion:      task.Param.ImageVersion,
	}, nil
}

// GetAppInClusterDNSStatus return the current status of in cluster dns
func (s *Service) GetAppInClusterDNSStatus(ctx context.Context, clusterName entity.ClusterName,
	envName entity.AppEnvName, app *resp.AppDetailResp) (entity.AppInClusterDNSStatus, error) {
	if app.ServiceType != entity.AppServiceTypeRestful || app.ServiceExposeType != entity.AppServiceExposeTypeIngress {
		return entity.AppInClusterDNSStatusUnSupported, nil
	}

	task, err := s.GetSingleTask(ctx, &req.GetTasksReq{
		AppID:             app.ID,
		EnvName:           envName,
		ClusterName:       clusterName,
		ActionList:        *entity.TaskActionInClusterDNSList,
		StatusInverseList: entity.TaskStatusFinalStateList,
		Suspend:           null.BoolFrom(false),
	})

	if err != nil && !errcode.EqualError(_errcode.NoRequiredTaskError, err) {
		return "", err
	}

	if task != nil {
		return entity.AppInClusterDNSStatusUpdating, nil
	}

	serviceName, err := s.GetCurrentServiceName(ctx, clusterName, envName, app)
	if err != nil {
		return "", err
	}

	if app.ServiceName != serviceName {
		return entity.AppInClusterDNSStatusDisabled, nil
	}

	return entity.AppInClusterDNSStatusEnabled, nil
}

// IsAppK8sPrimaryResourceExist verify if k8s primary resource exists
// the resource is determined by app type.
func (s *Service) IsAppK8sPrimaryResourceExist(ctx context.Context, clusterName entity.ClusterName,
	envName entity.AppEnvName, appType entity.AppType, version, namespace string) (exist bool, err error) {
	switch appType {
	case entity.AppTypeService, entity.AppTypeWorker:
		_, err = s.GetDeploymentDetail(ctx, clusterName, envName, &req.GetDeploymentDetailReq{
			Namespace: namespace,
			Name:      version,
			Env:       string(envName),
		})
	case entity.AppTypeCronJob:
		_, err = s.GetCronJobDetail(ctx, clusterName, &req.GetCronJobDetailReq{
			Namespace: namespace,
			Name:      version,
			Env:       string(envName),
		})
	case entity.AppTypeOneTimeJob:
		_, err = s.GetJobDetail(ctx, clusterName, &req.GetJobDetailReq{
			Namespace: namespace,
			Name:      version,
			Env:       string(envName),
		})
	}
	if err != nil {
		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *Service) isJobFailed(j *batch.Job) bool {
	for _, c := range j.Status.Conditions {
		if c.Type == batch.JobFailed && c.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}
