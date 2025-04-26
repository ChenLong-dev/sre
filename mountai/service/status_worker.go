package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AliyunContainerService/kubernetes-cronhpa-controller/pkg/apis/autoscaling/v1beta1"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/sentry"

	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
)

// TransformTaskStatus 转换任务状态
func (s *Service) TransformTaskStatus(ctx context.Context, task *resp.TaskDetailResp) (err error) {
	if task.ClusterName == "" {
		return errors.Wrapf(errcode.InternalError, "no cluster_name found in task(%s)", task.ID)
	}

	// 加分布式锁
	mutex := s.dao.GetTaskStatusWorkerLock(ctx, task.ID)
	err = mutex.Lock(ctx)
	if err != nil {
		return errors.Wrapf(errcode.RedLockLockError, "%s", err)
	}

	defer func() {
		result := mutex.Unlock(ctx)
		if !result {
			err = errors.WithStack(errcode.RedLockUnLockError)
		}
	}()

	project, app := new(resp.ProjectDetailResp), new(resp.AppDetailResp)
	// 更新状态
	var nextStatus entity.TaskStatus
	defer func() {
		if err == nil && nextStatus != "" && nextStatus != task.Status {
			curError := s.UpdateTask(ctx, project, app, task, &req.UpdateTaskReq{
				Status: nextStatus,
			})
			if curError != nil {
				err = curError
				log.Errorc(ctx, "An error occurred during updating task %s: %s.", task.ID, curError.Error())
			}

			// Push event to message queue when task succeeded or failed
			for _, action := range entity.TaskStatusFinalStateList {
				if nextStatus != action {
					continue
				}
				actionType := entity.TransformSubscribeAction(task.Action)
				pubErr := s.PublishAppOpEvent(ctx, &entity.SubscribeEventMsg{
					ActionType: actionType,
					TaskID:     task.ID,
					OperatorID: task.OperatorID,
					AppID:      task.AppID,
					Env:        task.EnvName,
					OpTime:     utils.FormatK8sTime(time.Now()),
				})
				if pubErr != nil {
					err = pubErr
					log.Errorc(ctx, "An error occurred during publishing task %s op event [action=%s]"+
						" to message queue: %s.", actionType, task.ID, pubErr.Error())
				}
			}
		}
		log.Infoc(ctx, "Successfully updated task %s.", task.ID)
	}()

	// 最大重试后则认为失败
	if task.RetryCount > entity.TaskMaxRetryCount {
		nextStatus = entity.TaskStatusFail
		return nil
	}

	// 任务某一阶段最多10分钟，否则认为失败
	updateTime, err := time.ParseInLocation(utils.DefaultTimeFormatLayout, task.UpdateTime, time.Local)
	if err != nil {
		return err
	}
	if time.Since(updateTime) > entity.TaskExecuteTimeout && task.Status != entity.TaskStatusInit {
		nextStatus = entity.TaskStatusFail
		return nil
	}

	// 清理任务较特殊，无需获取项目详情
	if task.Action == entity.TaskActionClean {
		nextStatus, err = s.transformCleanTaskStatus(ctx, task)
		if err != nil {
			sentry.CaptureWithBreadAndTags(ctx, err, &sentry.Breadcrumb{
				Category: "CleanTask",
				Data: map[string]interface{}{
					"TaskID": task.ID,
				},
			})
			return err
		}
		return nil
	}

	// 由于删除应用时创建 clean task，需要保证获取 app 在 transformCleanTaskStatus 之后。
	app, err = s.GetAppDetail(ctx, task.AppID)
	if err != nil {
		return err
	}

	// TODO: 支持istio
	// enable multi env istio control for a better user experience
	// app.EnableIstio = s.GetApplicationIstioState(ctx, task.EnvName, task.ClusterName, app)
	app.EnableIstio = false

	project, err = s.GetProjectDetail(ctx, app.ProjectID)
	if err != nil {
		return err
	}

	// 如果 task 未设置,则默认的命名空间为其 EnvName
	if task.Namespace == "" {
		task.Namespace = string(task.EnvName)
	}

	switch task.Action {
	case entity.TaskActionManualLaunch:
		nextStatus, err = s.transformManualLaunchTaskStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	case entity.TaskActionUpdateHPA:
		nextStatus, err = s.transformUpdateHPATaskStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	case entity.TaskActionStop:
		nextStatus, err = s.transformStopTaskStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	case entity.TaskActionResume:
		nextStatus, err = s.transformResumeTaskStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	case entity.TaskActionRestart:
		nextStatus, err = s.transformRestartTaskStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	case entity.TaskActionDelete:
		nextStatus, err = s.transformDeleteTaskStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	case entity.TaskActionFullDeploy:
		nextStatus, err = s.transformFullDeployTaskStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	case entity.TaskActionFullCanaryDeploy:
		nextStatus, err = s.transformFullCanaryDeployTaskStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	case entity.TaskActionCanaryDeploy:
		nextStatus, err = s.transformCanaryDeployTaskStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	case entity.TaskActionReloadConfig:
		nextStatus, err = s.transformReloadConfigStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	case entity.TaskActionDisableInClusterDNS:
		nextStatus, err = s.transformInClusterDNSStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	case entity.TaskActionEnableInClusterDNS:
		nextStatus, err = s.transformInClusterDNSStatus(ctx, project, app, task, project.Team)
		if err != nil {
			return err
		}
	default:
		return _errcode.UnknownTaskActionError
	}
	return nil
}

// TODO: CreateK8sServiceUnderway/CleanupK8sServiceUnderway 这类操作似乎都没有意义，应该重构状态机中相关的操作？
func (s *Service) transformInClusterDNSStatus(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) (entity.TaskStatus, error) {
	currentServiceName := s.GetShadowServiceName(app.ServiceName)
	newServiceName := app.ServiceName
	if task.Action == entity.TaskActionDisableInClusterDNS {
		currentServiceName, newServiceName = newServiceName, currentServiceName
	}

	switch task.Status {
	case entity.TaskStatusInit:
		latestSuccessTask, err := s.GetLatestSuccessTask(ctx, &req.GetLatestTaskReq{
			AppID:       app.ID,
			EnvName:     task.EnvName,
			ClusterName: task.ClusterName,
			ActionList:  entity.TaskActionInitDeployList,
		})
		if err != nil && !errcode.EqualError(_errcode.NoRequiredTaskError, err) {
			return "", err
		}

		if latestSuccessTask != nil {
			task.Param.TargetPort = latestSuccessTask.Param.TargetPort
			task.Param.ExposedPorts = latestSuccessTask.Param.ExposedPorts
		}

		err = s.ApplyServiceFromTpl(ctx, project, app, task, team, newServiceName)
		if err != nil {
			return "", err
		}

		return entity.TaskStatusCreateK8sServiceUnderway, nil
	case entity.TaskStatusCreateK8sServiceUnderway:
		return entity.TaskStatusCreateK8sServiceFinish, nil
	case entity.TaskStatusCreateK8sServiceFinish:
		latestSuccessTask, err := s.GetLatestSuccessTask(ctx, &req.GetLatestTaskReq{
			AppID:       app.ID,
			EnvName:     task.EnvName,
			ClusterName: task.ClusterName,
			ActionList:  entity.TaskActionInitDeployList,
		})
		if err != nil && !errcode.EqualError(_errcode.NoRequiredTaskError, err) {
			return "", err
		}

		if latestSuccessTask != nil {
			task.Param.IsSupportStickySession = latestSuccessTask.Param.IsSupportStickySession
			task.Param.SessionCookieMaxAge = latestSuccessTask.Param.SessionCookieMaxAge
		}

		err = s.ApplyIngressFromTpl(ctx, project, app, task, newServiceName)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCreateK8sIngressUnderway, nil
	case entity.TaskStatusCreateK8sIngressUnderway:
		s.WaitIngressChangeReady(ctx)
		return entity.TaskStatusCreateK8sIngressFinish, nil
	case entity.TaskStatusCreateK8sIngressFinish:
		if task.Action == entity.TaskActionDisableInClusterDNS {
			svc, err := s.GetServiceDetail(ctx, task.ClusterName,
				&req.GetServiceDetailReq{
					Namespace: task.Namespace,
					Name:      currentServiceName,
					Env:       string(task.EnvName),
				})

			if err != nil {
				if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
					return entity.TaskStatusCleanK8sServiceFinish, nil
				}
				return "", err
			}

			svc.Labels[entity.ServiceInClusterDNSLable] = entity.ServiceInClusterValueDisabled
			_, err = s.PatchService(ctx, task.ClusterName, svc, string(task.EnvName))
			if err != nil {
				return "", err
			}
			return entity.TaskStatusUpdateInClusterDNSUnderway, nil
		}
		return entity.TaskStatusUpdateInClusterDNSUnderway, nil
	case entity.TaskStatusUpdateInClusterDNSUnderway:
		s.WaitInClusterDNSChangeReady(ctx)
		if task.Action == entity.TaskActionEnableInClusterDNS {
			return entity.TaskStatusRestartDeploymentFinish, nil
		}
		return entity.TaskStatusUpdateInClusterDNSFinish, nil
	case entity.TaskStatusUpdateInClusterDNSFinish:
		deployments, err := s.GetDeployments(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentsReq{
			Namespace:   task.Namespace,
			ProjectName: project.Name,
			AppName:     app.Name,
			Env:         string(task.EnvName),
		})

		if err != nil {
			return "", err
		}

		for i := range deployments {
			if deployments[i].Spec.Replicas == nil || *deployments[i].Spec.Replicas == 0 {
				continue
			}

			err = s.RestartDeploymentAndIgnoreResponse(ctx, task.ClusterName, task.EnvName, &req.RestartDeploymentReq{
				Namespace: task.Namespace,
				Name:      deployments[i].Name,
				Env:       string(task.EnvName),
			})
			if err != nil {
				return "", err
			}
		}

		if len(deployments) > 0 {
			return entity.TaskStatusRestartDeploymentUnderway, nil
		}
		return entity.TaskStatusRestartDeploymentFinish, nil
	case entity.TaskStatusRestartDeploymentUnderway:
		pods, err := s.GetPods(ctx, task.ClusterName, &req.GetPodsReq{
			Namespace:   task.Namespace,
			ProjectName: project.Name,
			AppName:     app.Name,
			Env:         string(task.EnvName),
		})
		if err != nil {
			return "", err
		}

		createTime, err := time.ParseInLocation(utils.DefaultTimeFormatLayout, task.CreateTime, time.Local)
		if err != nil {
			return "", err
		}

		for i := range pods {
			if pods[i].Annotations[K8sAnnotationRestart] == "" {
				return "", nil
			}

			restartTime, err := time.Parse(RestartAnnotationLayout, pods[i].Annotations[K8sAnnotationRestart])
			if err != nil {
				return "", err
			}

			if restartTime.Before(createTime) {
				return "", nil
			}

			if !IsPodReadyConditionTrue(&pods[i].Status) {
				return "", nil
			}
		}
		return entity.TaskStatusRestartDeploymentFinish, nil
	case entity.TaskStatusRestartDeploymentFinish:
		err := s.DeleteService(ctx, task.ClusterName, &req.DeleteServiceReq{
			Namespace: task.Namespace,
			Name:      currentServiceName,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return "", err
		}

		return entity.TaskStatusCleanK8sServiceUnderway, nil
	case entity.TaskStatusCleanK8sServiceUnderway:
		_, err := s.GetEndpointsDetail(ctx, task.ClusterName, &req.GetEndpointsReq{
			Namespace: task.Namespace,
			Name:      currentServiceName,
			Env:       string(task.EnvName),
		})
		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return entity.TaskStatusCleanK8sServiceFinish, nil
		}
		return "", nil
	case entity.TaskStatusCleanK8sServiceFinish:
		if task.Action == entity.TaskActionDisableInClusterDNS {
			err := s.CreateKongGatewayEndpoints(ctx, project, app, task)
			if err != nil {
				return "", err
			}
		}
		return entity.TaskStatusSuccess, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}
}

func (s *Service) transformUpdateHPATaskStatus(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) (entity.TaskStatus, error) {
	switch task.Status {
	case entity.TaskStatusInit:
		// 渲染模版
		data, err := s.RenderHPATemplate(ctx, project, app, task, team)
		if err != nil {
			return "", err
		}
		// 生成HPA
		_, err = s.ApplyHPA(ctx, task.ClusterName, data, string(task.EnvName))
		if err != nil {
			return "", err
		}

		return entity.TaskStatusUpdateHPAUnderway, nil
	case entity.TaskStatusUpdateHPAUnderway:
		return entity.TaskStatusUpdateHPAFinish, nil
		_, err := s.GetHPAReadyTime(ctx, task.ClusterName,
			&req.GetHPADetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return "", nil
			}
			if errcode.EqualError(_errcode.K8sResourceNotReadyError, err) {
				return "", nil
			}
			return "", err
		}

		return entity.TaskStatusUpdateHPAFinish, nil
	case entity.TaskStatusUpdateHPAFinish:
		if len(task.Param.CronScaleJobGroups) == 0 {
			return entity.TaskStatusCreateCronHPAFinish, nil
		}

		data, err := s.RenderCronHPATemplate(ctx, project, app, task, team)
		if err != nil {
			return "", err
		}
		_, err = s.ApplyCronHPA(ctx, task.ClusterName, data, string(task.EnvName))
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCreateCronHPAUnderway, nil
	case entity.TaskStatusCreateCronHPAUnderway:
		// 校验cronHAP任务注册状态
		cronHPA, err := s.GetCronHPADetail(ctx, task.ClusterName, &req.GetCronHPADetailReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		for _, j := range cronHPA.Status.Conditions {
			if j.State == v1beta1.Failed {
				return "", _errcode.CronHPAJobRegisterFailed
			}
		}
		return entity.TaskStatusCreateCronHPAFinish, nil
	case entity.TaskStatusCreateCronHPAFinish:
		return entity.TaskStatusSuccess, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}
}

func (s *Service) transformManualLaunchTaskStatus(ctx context.Context, _ *resp.ProjectDetailResp, _ *resp.AppDetailResp,
	task *resp.TaskDetailResp, _ *resp.TeamDetailResp) (entity.TaskStatus, error) {
	switch task.Status {
	case entity.TaskStatusInit:
		// 获取CronJob的spec
		cronJob, err := s.GetCronJobDetail(ctx, task.ClusterName,
			&req.GetCronJobDetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			return "", err
		}

		// 获取Job的spec
		job, err := s.GetJobSpecFromCronJob(task, cronJob)
		if err != nil {
			return "", err
		}

		// 创建Job
		_, err = s.CreateJob(ctx, task.ClusterName, job, string(task.EnvName))
		if err != nil {
			return "", err
		}

		return entity.TaskStatusCreateJobUnderway, nil
	case entity.TaskStatusCreateJobUnderway:
		// 获取Job详情
		job, err := s.GetJobDetail(ctx, task.ClusterName,
			&req.GetJobDetailReq{
				Namespace: task.Namespace,
				Name:      task.Param.ManualJobName,
				Env:       string(task.EnvName),
			})

		if err == nil {
			if job.GetLabels()[entity.LabelKeyLaunchType] != string(entity.LaunchTypeManual) {
				return "", errors.Wrapf(_errcode.LabelValueError, entity.LabelKeyLaunchType+"is not manual")
			}
			return entity.TaskStatusCreateJobFinish, nil
		} else if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return "", err
		} else {
			return "", errors.Wrapf(_errcode.K8sResourceNotFoundError, "create manual job failed")
		}
	case entity.TaskStatusCreateJobFinish:
		return entity.TaskStatusSuccess, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}
}

// transformFullDeployTaskStatus 全量部署
func (s *Service) transformFullDeployTaskStatus(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) (entity.TaskStatus, error) {
	switch task.Status {
	// 初始状态
	case entity.TaskStatusInit:
		err := s.ApplyLogConfig(ctx, project, app, task, team)
		if err != nil {
			if errcode.EqualError(_errcode.LogConfigDisabled, err) {
				// 禁用日志接入时跳过阶段
				return entity.TaskStatusCreateAliLogConfigFinish, nil
			}

			return "", err
		}

		return entity.TaskStatusCreateAliLogConfigUnderway, nil
	// K8s ConfigMap 创建阶段完成
	// case entity.TaskStatusCreateConfigMapFinish:
	// 云日志配置创建中
	case entity.TaskStatusCreateAliLogConfigUnderway:
		err := s.LogConfigExistanceCheck(ctx, app, task)
		if err == nil {
			return entity.TaskStatusCreateAliLogConfigFinish, nil
		}

		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return "", nil
		}

		return "", err
	// 云日志配置创建完成
	// case entity.TaskStatusCreateAliLogConfigFinish:
	// return entity.TaskStatusCreateLogStoreIndexFinish, nil
	// err := s.EnsureLogIndex(ctx, app, task)
	// if err == nil {
	// 	return entity.TaskStatusCreateLogStoreIndexFinish, nil
	// }

	// if errcode.EqualError(_errcode.AliResourceNotFoundError, err) {
	// 	return "", nil
	// }

	// return "", err
	// 云日志索引创建完成
	case entity.TaskStatusCreateLogStoreIndexFinish:
		// TODO: 日志转储配置
		// err := s.ApplyLogDump(ctx, project, app, task)
		// if err != nil {
		// 	return "", err
		// }

		return entity.TaskStatusSyncColdStorageDeliverTaskFinish, nil
	// 投递任务同步成功
	case entity.TaskStatusCreateAliLogConfigFinish:
		if app.Type == entity.AppTypeCronJob {
			// 渲染模版
			data, err := s.RenderCronJobTemplate(ctx, project, app, task, team)
			if err != nil {
				return "", err
			}
			// 生成CronJob
			_, err = s.ApplyCronJob(ctx, task.ClusterName, data, string(task.EnvName))
			if err != nil {
				return "", err
			}
			return entity.TaskStatusCreateFullCronJobUnderway, nil
		} else if app.Type == entity.AppTypeOneTimeJob {
			// 渲染模版
			data, err := s.RenderJobTemplate(ctx, project, app, task, team)
			if err != nil {
				return "", err
			}
			// 生成Job
			_, err = s.ApplyJob(ctx, task.ClusterName, data, string(task.EnvName))
			if err != nil {
				return "", err
			}
			return entity.TaskStatusCreateJobUnderway, nil
		} else {
			// 渲染模版
			data, err := s.RenderDeploymentTemplate(ctx, project, app, task, team)
			if err != nil {
				return "", err
			}

			// 生成Deployment
			err = s.ApplyDeploymentAndIgnoreResponse(ctx, task.ClusterName, task.EnvName, data)
			if err != nil {
				return "", err
			}
			return entity.TaskStatusCreateFullDeploymentUnderway, nil
		}
	// K8s CronJob 创建中
	case entity.TaskStatusCreateFullCronJobUnderway:
		_, err := s.GetCronJobDetail(ctx, task.ClusterName,
			&req.GetCronJobDetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err == nil {
			return entity.TaskStatusCreateFullCronJobFinish, nil
		} else if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return "", err
		}
	// K8s Job 创建中
	case entity.TaskStatusCreateJobUnderway:
		_, err := s.GetJobDetail(ctx, task.ClusterName,
			&req.GetJobDetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err == nil {
			return entity.TaskStatusCreateJobFinish, nil
		} else if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return "", err
		}
	// K8s CronJob 创建完成
	case entity.TaskStatusCreateFullCronJobFinish, entity.TaskStatusCreateJobFinish:
		return entity.TaskStatusCreateHPAFinish, nil
	// K8s Deployment 创建中
	case entity.TaskStatusCreateFullDeploymentUnderway:
		status, err := s.GetDeploymentStatus(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentDetailReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err == nil {
			if int(status.UpdatedReplicas) >= task.Param.MinPodCount &&
				status.AvailableReplicas == status.UpdatedReplicas &&
				int(status.UnavailableReplicas) == 0 {
				return entity.TaskStatusCreateFullDeploymentFinish, nil
			}
		} else if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return "", err
		}
	// K8s Deployment 创建完成
	case entity.TaskStatusCreateFullDeploymentFinish:
		// todo:: 暂时不移除 K8sAnnotationHPASkipped 注解
		//err := s.EnableDeploymentHPAAndIgnoreResponse(ctx, task.ClusterName, task.EnvName, &req.EnableDeploymentHPAReq{
		//	Env:       string(task.EnvName),
		//	Name:      task.Version,
		//	Namespace: task.Namespace,
		//})
		//if err != nil {
		//	return "", err
		//}

		// 不自动扩缩容
		if !task.Param.IsAutoScale {
			return entity.TaskStatusCreateHPAFinish, nil
		}

		// 为防止prometheus中没有数据(TODO: 使用svc服务发现，可以避免？)
		// 需要确保pod初始化完毕一定时间后再初始化hpa
		//latestTime, err := s.GetPodsLatestReadyTime(ctx, task.ClusterName,
		//	&req.GetPodsReq{
		//		Namespace:   task.Namespace,
		//		ProjectName: project.Name,
		//		AppName:     app.Name,
		//		Version:     task.Version,
		//		Env:         string(task.EnvName),
		//	})
		//if err != nil {
		//	if errcode.EqualError(_errcode.K8sResourceNotReadyError, err) {
		//		return "", nil
		//	}
		//	return "", err
		//}
		//
		//if needWaitMetricsSync(project.Labels, task.EnvName) && time.Since(latestTime) <= entity.HPAInitDelayDuration {
		//	return "", nil
		//}

		// 渲染模版
		data, err := s.RenderHPATemplate(ctx, project, app, task, team)
		if err != nil {
			return "", err
		}
		// 生成HPA
		_, err = s.ApplyHPA(ctx, task.ClusterName, data, string(task.EnvName))
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCreateHPAUnderway, nil
	// K8s HPA  创建中
	case entity.TaskStatusCreateHPAUnderway:
		return entity.TaskStatusCreateHPAFinish, nil
		readyTime, err := s.GetHPAReadyTime(ctx, task.ClusterName,
			&req.GetHPADetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return "", nil
			}
			if errcode.EqualError(_errcode.K8sResourceNotReadyError, err) {
				return "", nil
			}
			return "", err
		}

		taskCreateTime, err := time.ParseInLocation(utils.DefaultTimeFormatLayout, task.CreateTime, time.Local)
		if err != nil {
			return "", err
		}

		// 上一次同步时间要比任务创建时间晚，确保hpa已经同步
		if readyTime.After(taskCreateTime) {
			return entity.TaskStatusCreateHPAFinish, nil
		}
	// K8s HPA 创建完成
	case entity.TaskStatusCreateHPAFinish:
		if len(task.Param.CronScaleJobGroups) == 0 {
			return entity.TaskStatusCreateCronHPAFinish, nil
		}

		data, err := s.RenderCronHPATemplate(ctx, project, app, task, team)
		if err != nil {
			return "", err
		}
		_, err = s.ApplyCronHPA(ctx, task.ClusterName, data, string(task.EnvName))
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCreateCronHPAUnderway, nil
	// K8s CronHPA 创建中
	case entity.TaskStatusCreateCronHPAUnderway:
		// 校验cronHAP任务注册状态
		cronHPA, err := s.GetCronHPADetail(ctx, task.ClusterName, &req.GetCronHPADetailReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		for _, j := range cronHPA.Status.Conditions {
			if j.State == v1beta1.Failed {
				return "", _errcode.CronHPAJobRegisterFailed
			}
		}
		return entity.TaskStatusCreateCronHPAFinish, nil
	// K8s CronHPA 创建完成
	case entity.TaskStatusCreateCronHPAFinish:
		if app.Type != entity.AppTypeService {
			return entity.TaskStatusAllCreationPhasesFinish, nil
		}

		serviceName, err := s.GetCurrentServiceName(ctx, task.ClusterName, task.EnvName, app)
		if err != nil {
			return "", err
		}

		// 生成Service
		err = s.ApplyServiceFromTpl(ctx, project, app, task, team, serviceName)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCreateK8sServiceFinish, nil
	// K8s Service 创建完成
	case entity.TaskStatusCreateK8sServiceFinish:
		// 非Ingress方式暴露服务，不需要创建 Ingress 对象
		if app.ServiceExposeType != entity.AppServiceExposeTypeIngress {
			return entity.TaskStatusCreateK8sIngressFinish, nil
		}

		serviceName, err := s.GetCurrentServiceName(ctx, task.ClusterName, task.EnvName, app)
		if err != nil {
			return "", err
		}

		if app.ServiceExposeType == entity.AppServiceExposeTypeIngress &&
			s.GetApplicationIstioState(ctx, task.EnvName, task.ClusterName, app) {
			_, err = s.ApplyVirtualServiceFromTpl(ctx, project, app, task, serviceName)
			if err != nil {
				return "", err
			}
			return entity.TaskStatusCreateVirtualServiceUnderway, nil
		}

		err = s.ApplyIngressFromTpl(ctx, project, app, task, serviceName)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCreateK8sIngressUnderway, nil
	// K8s Ingress 创建中
	case entity.TaskStatusCreateK8sIngressUnderway:
		ingress, err := s.GetIngressDetail(ctx, task.ClusterName,
			&req.GetIngressDetailReq{
				Namespace: task.Namespace,
				Name:      app.ServiceName,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return "", nil
			}
			return "", err
		}
		if len(ingress.Status.LoadBalancer.Ingress) == 0 || ingress.Status.LoadBalancer.Ingress[0].IP == "" { // ingress ip 未更新
			return "", nil
		}
		return entity.TaskStatusCreateK8sIngressFinish, nil
	// K8s Ingress 创建完成
	case entity.TaskStatusCreateK8sIngressFinish:
		// 目前不需要网关来管理路由，ingress已足够
		return entity.TaskStatusAllCreationPhasesFinish, nil
		multiClusterSupported, err := s.CheckMultiClusterSupport(ctx, task.EnvName, project.ID)
		if err != nil {
			return "", err
		}

		// 因为 kong target 依赖于应用集群域名, 所以将这步提前到 ensureQDNSBusinessWithCalculation 之前
		err = s.createPrivateZoneRecordEntryForClusterDomain(ctx, project, app, task)
		if err != nil {
			return "", err
		}

		// 多集群白名单内的应用使用 Kong 配置通配域名路由规则, 如果尚未配置需要按照已部署集群数自动计算
		if multiClusterSupported {
			err = s.ensureQDNSBusinessWithCalculation(ctx, project, app, task, entity.TaskActionFullDeploy)
			if err != nil {
				return "", err
			}

			return entity.TaskStatusCreateKongObjectsFinish, nil
		}

		// todo 这里永远走不到了, 因为 multiClusterSupported 恒为 true
		// 非多集群白名单内的应用按照老方式处理域名解析
		err = s.createPrivateZoneRecordEntry(ctx, project, app, task)
		if err != nil {
			return "", err
		}

		return entity.TaskStatusAllCreationPhasesFinish, nil
	// Kong Object 创建完成
	case entity.TaskStatusCreateKongObjectsFinish:
		// k8s 通配域名的解析只有在不存在时才自动创建, 否则保持不动(由运维人工维护或通过运维系统维护)
		exist, err := s.checkPrivateZoneRecordExistanceForK8sDomain(ctx, task.EnvName, app.ServiceName)
		if err != nil {
			return "", err
		}

		if !exist {
			domainController := s.GetDomainControllerFromProjectOwners(project.Owners)
			err = s.createKongPrivateZoneRecordForK8sDomain(ctx, task.EnvName, app, task, domainController)
			if err != nil {
				return "", err
			}
		}

		return entity.TaskStatusAllCreationPhasesFinish, nil
	// 所有创建阶段完成
	case entity.TaskStatusAllCreationPhasesFinish:
		// CronJob组件
		if app.Type == entity.AppTypeCronJob {
			// 删除除当前版本以外的所有cron job
			err := s.DeleteCronJobs(ctx, task.ClusterName,
				&req.DeleteCronJobsReq{
					Namespace:   task.Namespace,
					ProjectName: project.Name,
					AppName:     app.Name,
					InverseName: task.Version,
					Env:         string(task.EnvName),
				})
			if err != nil {
				return "", err
			}
			return entity.TaskStatusCleanCronJobUnderway, nil
		} else if app.Type == entity.AppTypeOneTimeJob {
			// 删除除当前版本以外的所有job
			err := s.DeleteJobs(ctx, task.ClusterName,
				&req.DeleteJobsReq{
					Namespace:   task.Namespace,
					ProjectName: project.Name,
					AppName:     app.Name,
					Env:         string(task.EnvName),
					InverseName: task.Version,
				})
			if err != nil {
				return "", err
			}
			return entity.TaskStatusCleanJobUnderway, nil
		} else {
			err := s.DeleteCronHPAs(ctx, task.ClusterName,
				&req.DeleteCronHPAsReq{
					Namespace:      task.Namespace,
					ProjectName:    project.Name,
					AppName:        app.Name,
					InverseVersion: task.Version,
					Env:            string(task.EnvName),
				})

			if err != nil {
				return "", err
			}
			return entity.TaskStatusCleanCronHPAUnderway, nil
		}
	// Cron HPA 清理中
	case entity.TaskStatusCleanCronHPAUnderway:
		cronHPAlist, err := s.GetCronHPAs(ctx, task.ClusterName,
			&req.GetCronHPAsReq{
				Namespace:   task.Namespace,
				ProjectName: project.Name,
				AppName:     app.Name,
				Env:         string(task.EnvName),
			})
		if err != nil {
			return "", err
		}

		// 检查是否除当前版本外都已删除
		isDeleted := true
		for i := range cronHPAlist {
			if cronHPAlist[i].GetName() != task.Version {
				isDeleted = false
				break
			}
		}
		if isDeleted {
			return entity.TaskStatusCleanCronHPAFinish, nil
		}
	// Cron HPA 清理完成
	case entity.TaskStatusCleanCronHPAFinish:
		inverseVersion := task.Version
		if !task.Param.IsAutoScale {
			inverseVersion = ""
		}

		// 删除除当前版本以外的所有hpa
		err := s.DeleteHPAs(ctx, task.ClusterName,
			&req.DeleteHPAsReq{
				Namespace:      task.Namespace,
				ProjectName:    project.Name,
				AppName:        app.Name,
				Env:            string(task.EnvName),
				InverseVersion: inverseVersion,
			})
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCleanHPAUnderway, nil
	// HPA 清理中
	case entity.TaskStatusCleanHPAUnderway:
		hpaList, err := s.GetHPAs(ctx, task.ClusterName,
			&req.GetHPAsReq{
				Namespace:   task.Namespace,
				ProjectName: project.Name,
				AppName:     app.Name,
				Env:         string(task.EnvName),
			})
		if err != nil {
			return "", err
		}

		// 检查是否除当前版本外都已删除
		isDeleted := true
		for i := range hpaList {
			if hpaList[i].GetName() != task.Version {
				isDeleted = false
				break
			}
		}
		if isDeleted {
			return entity.TaskStatusCleanHPAFinish, nil
		}
	// HPA 清理完成
	case entity.TaskStatusCleanHPAFinish:
		// 删除除当前版本以外的所有deployment
		err := s.DeleteDeployments(ctx, task.ClusterName, task.EnvName, &req.DeleteDeploymentsReq{
			Namespace:      task.Namespace,
			ProjectName:    project.Name,
			AppName:        app.Name,
			InverseVersion: task.Version,
			Env:            string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		// 清理 对立面部署: 用于ingress 切换至 isito 或者 istio 切换至 ingress 进行的第一次全量发布(包括金丝雀发布)
		// TODO: istio
		// err = s.DeleteDeployments(ctx, task.ClusterName, task.EnvName, &req.DeleteDeploymentsReq{
		// 	Namespace:      string(s.getOppositeEnvName(task.Namespace)),
		// 	ProjectName:    project.Name,
		// 	AppName:        app.Name,
		// 	InverseVersion: task.Version,
		// 	Env:            string(task.EnvName),
		// })
		// if err != nil {
		// 	return "", err
		// }
		return entity.TaskStatusCleanDeploymentUnderway, nil
	// Deployment 清理中
	case entity.TaskStatusCleanDeploymentUnderway:
		deployments, err := s.GetDeployments(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentsReq{
			Namespace:   task.Namespace,
			ProjectName: project.Name,
			AppName:     app.Name,
			Env:         string(task.EnvName),
		})
		if err != nil {
			return "", err
		}

		// 检查是否除当前版本外都已删除
		isDeleted := true
		for i := range deployments {
			if deployments[i].GetName() != task.Version {
				isDeleted = false
				break
			}
		}
		if isDeleted {
			return entity.TaskStatusCleanDeploymentFinish, nil
		}
	// Cronjob 清理中
	case entity.TaskStatusCleanCronJobUnderway:
		jobs, err := s.GetCronJobs(ctx, task.ClusterName,
			&req.GetCronJobsReq{
				Namespace:   task.Namespace,
				ProjectName: project.Name,
				AppName:     app.Name,
				Env:         string(task.EnvName),
			})
		if err != nil {
			return "", err
		}

		// 检查是否除当前版本外都已删除
		isDeleted := true
		for i := range jobs {
			if jobs[i].GetName() != task.Version {
				isDeleted = false
				break
			}
		}
		if isDeleted {
			return entity.TaskStatusCleanCronJobFinish, nil
		}
	// Job 清理中
	case entity.TaskStatusCleanJobUnderway:
		jobs, err := s.GetJobs(ctx, task.ClusterName,
			&req.GetJobsReq{
				Namespace:   task.Namespace,
				ProjectName: project.Name,
				AppName:     app.Name,
				Env:         string(task.EnvName),
			})
		if err != nil {
			return "", err
		}

		// 检查是否除当前版本外都已删除
		isDeleted := true
		for i := range jobs {
			if jobs[i].GetName() != task.Version {
				isDeleted = false
				break
			}
		}
		if isDeleted {
			return entity.TaskStatusCleanJobFinish, nil
		}
	// Cronjob, Deployment, Job 清理完成
	case entity.TaskStatusCleanDeploymentFinish, entity.TaskStatusCleanCronJobFinish, entity.TaskStatusCleanJobFinish:
		// 清理 redis 中老版本部署记录
		err := s.dao.CleanAppRunningTasks(ctx, task.AppID, task.EnvName, task.ClusterName, task.Version)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusSuccess, nil
	// K8s VirtualService 创建中
	case entity.TaskStatusCreateVirtualServiceUnderway:
		// 检查 virtual service 状态
		_, err := s.GetVirtualServiceDetail(ctx, task.ClusterName, task.EnvName, &req.VirtualServiceReq{
			Namespace: task.Namespace,
			// todo 确认 getServiceName 方法存在的必要性
			Name: fmt.Sprintf("%s-%s", project.Name, app.Name),
			Env:  string(task.EnvName),
		})
		if err != nil {
			return entity.TaskStatusFail, errors.Wrap(_errcode.K8sInternalError, err.Error())
		}

		multiClusterSupported, err := s.CheckMultiClusterSupport(ctx, task.GetNamespaceAppEnv(
			s.GetApplicationIstioState(ctx, task.EnvName, task.ClusterName, app)), project.ID)
		if err != nil {
			return entity.TaskStatusFail, err
		}

		// 因为 kong target 依赖于应用集群域名, 所以将这步提前到 ensureQDNSBusinessWithCalculation 之前
		err = s.createPrivateZoneRecordEntryForClusterDomain(ctx, project, app, task)
		if err != nil {
			return "", err
		}

		// 多集群白名单内的应用使用 Kong 配置通配域名路由规则, 如果尚未配置需要按照已部署集群数自动计算
		if multiClusterSupported {
			err = s.ensureQDNSBusinessWithCalculation(ctx, project, app, task, entity.TaskActionFullDeploy)
			if err != nil {
				return entity.TaskStatusFail, err
			}

			return entity.TaskStatusCreateKongObjectsFinish, nil
		}

		// 非多集群白名单内的应用按照老方式处理域名解析
		err = s.createPrivateZoneRecordEntry(ctx, project, app, task)
		if err != nil {
			return "", err
		}

		return entity.TaskStatusCreateKongObjectsFinish, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}
	return "", nil
}

func (s *Service) transformCanaryDeployTaskStatus(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) (entity.TaskStatus, error) {
	switch task.Status {
	// 初始状态
	case entity.TaskStatusInit:
		err := s.ApplyLogConfig(ctx, project, app, task, team)
		if err != nil {
			if errcode.EqualError(_errcode.LogConfigDisabled, err) {
				// 禁用日志接入时跳过阶段
				return entity.TaskStatusCreateAliLogConfigFinish, nil
			}

			return "", err
		}

		return entity.TaskStatusCreateAliLogConfigUnderway, nil
	// k8s configmap 创建完成
	// case entity.TaskStatusCreateConfigMapFinish:
	// 云日志创建中
	case entity.TaskStatusCreateAliLogConfigUnderway:
		err := s.LogConfigExistanceCheck(ctx, app, task)
		if err == nil {
			return entity.TaskStatusCreateAliLogConfigFinish, nil
		}

		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return "", nil
		}

		return "", err
	// 云日志创建完成
	// case entity.TaskStatusCreateAliLogConfigFinish:
	// 	err := s.EnsureLogIndex(ctx, app, task)
	// 	if err == nil {
	// 		return entity.TaskStatusCreateLogStoreIndexFinish, nil
	// 	}

	// 	if errcode.EqualError(_errcode.AliResourceNotFoundError, err) {
	// 		return "", nil
	// 	}

	// 	return "", err
	// 云日志索引创建完成
	case entity.TaskStatusCreateLogStoreIndexFinish:
		err := s.ApplyLogDump(ctx, project, app, task)
		if err != nil {
			return "", err
		}

		return entity.TaskStatusSyncColdStorageDeliverTaskFinish, nil
	// 日志冷投创建完成
	case entity.TaskStatusCreateAliLogConfigFinish:
		// 金丝雀发布只发布一个
		task.Param.MinPodCount = 1
		// 渲染模版
		data, err := s.RenderDeploymentTemplate(ctx, project, app, task, team)
		if err != nil {
			return "", err
		}
		// 生成Deployment
		err = s.ApplyDeploymentAndIgnoreResponse(ctx, task.ClusterName, task.EnvName, data)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCreateCanaryDeploymentUnderway, nil
	// 灰度发布
	case entity.TaskStatusCreateCanaryDeploymentUnderway:
		status, err := s.GetDeploymentStatus(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentDetailReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err == nil {
			// 只有一个实例
			if int(status.ReadyReplicas) == 1 && int(status.UnavailableReplicas) == 0 {
				return entity.TaskStatusCreateCanaryDeploymentFinish, nil
			}
		} else if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return "", err
		}
	// 灰度发布完成
	case entity.TaskStatusCreateCanaryDeploymentFinish:
		if app.Type != entity.AppTypeService {
			return entity.TaskStatusAllCreationPhasesFinish, nil
		}

		serviceName, err := s.GetCurrentServiceName(ctx, task.ClusterName, task.EnvName, app)
		if err != nil {
			return "", err
		}

		// 生成Service
		err = s.ApplyServiceFromTpl(ctx, project, app, task, team, serviceName)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCreateK8sServiceFinish, nil
	// k8s service 创建完成
	case entity.TaskStatusCreateK8sServiceFinish:
		// 非Ingress方式暴露服务，不需要创建 Ingress 对象
		if app.ServiceExposeType != entity.AppServiceExposeTypeIngress {
			return entity.TaskStatusCreateK8sIngressFinish, nil
		}

		serviceName, err := s.GetCurrentServiceName(ctx, task.ClusterName, task.EnvName, app)
		if err != nil {
			return "", err
		}

		// istio 服务部署,只创建 vs, 域名解析
		if app.ServiceExposeType == entity.AppServiceExposeTypeIngress &&
			s.GetApplicationIstioState(ctx, task.EnvName, task.ClusterName, app) {
			_, err = s.ApplyVirtualServiceFromTpl(ctx, project, app, task, serviceName)
			if err != nil {
				return "", err
			}
			return entity.TaskStatusCreateVirtualServiceUnderway, nil
		}

		err = s.ApplyIngressFromTpl(ctx, project, app, task, serviceName)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCreateK8sIngressUnderway, nil
	// k8s ingress 创建中
	case entity.TaskStatusCreateK8sIngressUnderway:
		ingress, err := s.GetIngressDetail(ctx, task.ClusterName,
			&req.GetIngressDetailReq{
				Namespace: task.Namespace,
				Name:      app.ServiceName,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return "", nil
			}
			return "", err
		}
		if len(ingress.Status.LoadBalancer.Ingress) == 0 || ingress.Status.LoadBalancer.Ingress[0].IP == "" { // ingress ip 未更新
			return "", nil
		}
		return entity.TaskStatusCreateK8sIngressFinish, nil
	// k8s ingress 创建完成
	case entity.TaskStatusCreateK8sIngressFinish:
		return entity.TaskStatusAllCreationPhasesFinish, nil
		multiClusterSupported, err := s.CheckMultiClusterSupport(ctx, task.EnvName, project.ID)
		if err != nil {
			return "", err
		}

		// 因为 kong target 依赖于应用集群域名, 所以将这步提前到 ensureQDNSBusinessWithCalculation 之前
		err = s.createPrivateZoneRecordEntryForClusterDomain(ctx, project, app, task)
		if err != nil {
			return "", err
		}

		// 多集群白名单内的应用使用 Kong 配置通配域名路由规则, 如果尚未配置需要按照已部署集群数自动计算
		if multiClusterSupported {
			err = s.ensureQDNSBusinessWithCalculation(ctx, project, app, task, entity.TaskActionCanaryDeploy)
			if err != nil {
				return "", err
			}

			return entity.TaskStatusCreateKongObjectsFinish, nil
		}

		// 非多集群白名单内的应用按照老方式处理域名解析
		err = s.createPrivateZoneRecordEntry(ctx, project, app, task)
		if err != nil {
			return "", err
		}

		return entity.TaskStatusAllCreationPhasesFinish, nil
	// kong object 创建完成, 创建域名
	case entity.TaskStatusCreateKongObjectsFinish:
		err := s.createPrivateZoneRecordEntryForClusterDomain(ctx, project, app, task)
		if err != nil {
			return "", err
		}

		// k8s 通配域名的解析只有在不存在时才自动创建, 否则保持不动(由运维人工维护或通过运维系统维护)
		exist, err := s.checkPrivateZoneRecordExistanceForK8sDomain(ctx, task.EnvName, app.ServiceName)
		if err != nil {
			return "", err
		}

		if !exist {
			domainController := s.GetDomainControllerFromProjectOwners(project.Owners)
			err = s.createKongPrivateZoneRecordForK8sDomain(ctx, task.EnvName, app, task, domainController)
			if err != nil {
				return "", err
			}
		}

		return entity.TaskStatusAllCreationPhasesFinish, nil
	// 服务相关创建完成
	case entity.TaskStatusAllCreationPhasesFinish:
		return entity.TaskStatusSuccess, nil
	// istio virtualservice 创建中
	case entity.TaskStatusCreateVirtualServiceUnderway:
		// 检查 virtual service 状态
		_, err := s.GetVirtualServiceDetail(ctx, task.ClusterName, task.EnvName, &req.VirtualServiceReq{
			Namespace: task.Namespace,
			Name:      fmt.Sprintf("%s-%s", project.Name, app.Name),
			Env:       string(task.EnvName),
		})
		if err != nil {
			return "", errors.Wrap(_errcode.K8sInternalError, err.Error())
		}

		return entity.TaskStatusCreateK8sIngressFinish, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}
	return "", nil
}

func (s *Service) transformFullCanaryDeployTaskStatus(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) (entity.TaskStatus, error) {
	switch task.Status {
	case entity.TaskStatusInit:
		err := s.UpdateDeploymentScaleAndIgnoreResponse(ctx,
			task.ClusterName, task.EnvName, task.Namespace, task.Version, task.Param.MinPodCount)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusUpdateDeploymentScaleUnderway, nil
	case entity.TaskStatusUpdateDeploymentScaleUnderway:
		status, err := s.GetDeploymentStatus(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentDetailReq{
			Namespace: task.GetNamespace(s.GetApplicationIstioState(ctx, task.EnvName, task.ClusterName, app)),
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		// 实例数符合
		if int(status.UpdatedReplicas) == task.Param.MinPodCount &&
			status.AvailableReplicas == status.UnavailableReplicas &&
			int(status.UnavailableReplicas) == 0 {
			return entity.TaskStatusUpdateDeploymentScaleFinish, nil
		}
	case entity.TaskStatusUpdateDeploymentScaleFinish:
		// todo:: 暂时不移除 K8sAnnotationHPASkipped 注解
		//err := s.EnableDeploymentHPAAndIgnoreResponse(ctx, task.ClusterName, task.EnvName, &req.EnableDeploymentHPAReq{
		//	Env:       string(task.EnvName),
		//	Name:      task.Version,
		//	Namespace: task.Namespace,
		//})
		//if err != nil {
		//	return "", err
		//}

		// 不自动扩缩容
		if !task.Param.IsAutoScale {
			return entity.TaskStatusCreateHPAFinish, nil
		}

		// 为防止prometheus中没有数据
		// 需要确保pod初始化完毕一定时间后再初始化hpa
		//latestTime, err := s.GetPodsLatestReadyTime(ctx, task.ClusterName,
		//	&req.GetPodsReq{
		//		Namespace:   task.GetNamespace(s.GetApplicationIstioState(ctx, task.EnvName, task.ClusterName, app)),
		//		ProjectName: project.Name,
		//		AppName:     app.Name,
		//		Version:     task.Version,
		//		Env:         string(task.EnvName),
		//	})
		//if err != nil {
		//	if errcode.EqualError(_errcode.K8sResourceNotReadyError, err) {
		//		return "", nil
		//	}
		//	return "", err
		//}
		//
		//if needWaitMetricsSync(project.Labels, task.EnvName) && time.Since(latestTime) <= entity.HPAInitDelayDuration {
		//	return "", nil
		//}

		// 渲染模版
		data, err := s.RenderHPATemplate(ctx, project, app, task, team)
		if err != nil {
			return "", err
		}
		// 生成HPA
		_, err = s.ApplyHPA(ctx, task.ClusterName, data, string(task.EnvName))
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCreateHPAUnderway, nil
	case entity.TaskStatusCreateHPAUnderway:
		return entity.TaskStatusCreateHPAFinish, nil
		readyTime, err := s.GetHPAReadyTime(ctx, task.ClusterName,
			&req.GetHPADetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return "", nil
			}
			if errcode.EqualError(_errcode.K8sResourceNotReadyError, err) {
				return "", nil
			}
			return "", err
		}

		taskCreateTime, err := time.ParseInLocation(utils.DefaultTimeFormatLayout, task.CreateTime, time.Local)
		if err != nil {
			return "", err
		}

		// 上一次同步时间要比任务创建时间晚，确保hpa已经同步
		if readyTime.After(taskCreateTime) {
			return entity.TaskStatusCreateHPAFinish, nil
		}
	case entity.TaskStatusCreateHPAFinish:
		if len(task.Param.CronScaleJobGroups) == 0 {
			return entity.TaskStatusCreateCronHPAFinish, nil
		}

		data, err := s.RenderCronHPATemplate(ctx, project, app, task, team)
		if err != nil {
			return "", err
		}
		_, err = s.ApplyCronHPA(ctx, task.ClusterName, data, string(task.EnvName))
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCreateCronHPAUnderway, nil
	case entity.TaskStatusCreateCronHPAUnderway:
		// 校验cronHAP任务注册状态
		cronHPA, err := s.GetCronHPADetail(ctx, task.ClusterName, &req.GetCronHPADetailReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		for _, j := range cronHPA.Status.Conditions {
			if j.State == v1beta1.Failed {
				return "", _errcode.CronHPAJobRegisterFailed
			}
		}
		return entity.TaskStatusCreateCronHPAFinish, nil
	case entity.TaskStatusCreateVirtualServiceFinish:
		return entity.TaskStatusCreateCronHPAFinish, nil
	case entity.TaskStatusCreateCronHPAFinish:
		err := s.DeleteCronHPAs(ctx, task.ClusterName,
			&req.DeleteCronHPAsReq{
				Namespace:      task.Namespace,
				ProjectName:    project.Name,
				AppName:        app.Name,
				InverseVersion: task.Version,
				Env:            string(task.EnvName),
			})

		if err != nil {
			return "", err
		}
		return entity.TaskStatusCleanCronHPAUnderway, nil
	case entity.TaskStatusCleanCronHPAUnderway:
		cronHPAlist, err := s.GetCronHPAs(ctx, task.ClusterName,
			&req.GetCronHPAsReq{
				Namespace:   task.Namespace,
				ProjectName: project.Name,
				AppName:     app.Name,
				Env:         string(task.EnvName),
			})
		if err != nil {
			return "", err
		}

		// 检查是否除当前版本外都已删除
		isDeleted := true
		for i := range cronHPAlist {
			if cronHPAlist[i].GetName() != task.Version {
				isDeleted = false
				break
			}
		}
		if isDeleted {
			return entity.TaskStatusCleanCronHPAFinish, nil
		}
	case entity.TaskStatusCleanCronHPAFinish:
		inverseVersion := task.Version
		if !task.Param.IsAutoScale {
			inverseVersion = ""
		}

		// 删除除当前版本以外的所有hpa
		err := s.DeleteHPAs(ctx, task.ClusterName,
			&req.DeleteHPAsReq{
				Namespace:      task.Namespace,
				ProjectName:    project.Name,
				AppName:        app.Name,
				InverseVersion: inverseVersion,
				Env:            string(task.EnvName),
			})
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCleanHPAUnderway, nil
	case entity.TaskStatusCleanHPAUnderway:
		hpaList, err := s.GetHPAs(ctx, task.ClusterName,
			&req.GetHPAsReq{
				Namespace:   task.Namespace,
				ProjectName: project.Name,
				AppName:     app.Name,
				Env:         string(task.EnvName),
			})
		if err != nil {
			return "", err
		}

		// 检查是否除当前版本外都已删除
		isDeleted := true
		for i := range hpaList {
			if hpaList[i].GetName() != task.Version {
				isDeleted = false
				break
			}
		}
		if isDeleted {
			return entity.TaskStatusCleanHPAFinish, nil
		}
	case entity.TaskStatusCleanHPAFinish:
		// 删除除当前版本以外的所有deployment
		err := s.DeleteDeployments(ctx, task.ClusterName, task.EnvName, &req.DeleteDeploymentsReq{
			Namespace:      task.Namespace,
			ProjectName:    project.Name,
			AppName:        app.Name,
			InverseVersion: task.Version,
			Env:            string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		// 清理 对立面部署: 用于ingress 切换至 isito 或者 istio 切换至 ingress 进行的第一次全量发布(包括金丝雀发布)
		// err = s.DeleteDeployments(ctx, task.ClusterName, task.EnvName, &req.DeleteDeploymentsReq{
		// 	Namespace:      string(s.getOppositeEnvName(task.Namespace)),
		// 	ProjectName:    project.Name,
		// 	AppName:        app.Name,
		// 	InverseVersion: task.Version,
		// 	Env:            string(task.EnvName),
		// })
		// if err != nil {
		// 	return "", err
		// }
		return entity.TaskStatusCleanDeploymentUnderway, nil
	case entity.TaskStatusCleanDeploymentUnderway:
		deployments, err := s.GetDeployments(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentsReq{
			Namespace:   task.Namespace,
			ProjectName: project.Name,
			AppName:     app.Name,
			Env:         string(task.EnvName),
		})
		if err != nil {
			return "", err
		}

		// 检查是否除当前版本外都已删除
		isDeleted := true
		for i := range deployments {
			if deployments[i].GetName() != task.Version {
				isDeleted = false
				break
			}
		}
		if isDeleted {
			return entity.TaskStatusCleanDeploymentFinish, nil
		}
	case entity.TaskStatusCleanDeploymentFinish:
		// 清理 redis 中老版本部署记录
		err := s.dao.CleanAppRunningTasks(ctx, task.AppID, task.EnvName, task.ClusterName, task.Version)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusSuccess, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}
	return "", nil
}

func (s *Service) transformStopTaskStatus(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, _ *resp.TeamDetailResp) (entity.TaskStatus, error) {
	switch task.Status {
	case entity.TaskStatusInit:
		if app.Type == entity.AppTypeCronJob {
			cronjob, err := s.GetCronJobDetail(ctx, task.ClusterName,
				&req.GetCronJobDetailReq{
					Namespace: task.Namespace,
					Name:      task.Version,
					Env:       string(task.EnvName),
				})
			if err != nil {
				return "", err
			}
			_, err = s.SuspendCronJob(ctx, task.ClusterName, cronjob, string(task.EnvName))
			if err != nil {
				return "", err
			}
			return entity.TaskStatusUpdateCronJobSuspendUnderway, nil
		}

		if app.Type == entity.AppTypeService {
			// 服务需要检查同集群内是否有其他成功的部署以确定是否需要删除通配域名的阿里云解析
			exists, err := s.CheckHealthyDeploymentExistance(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentsReq{
				Namespace:   task.Namespace,
				ProjectName: project.Name,
				AppName:     app.Name,
				Env:         string(task.EnvName),
			}, task.Version)
			if err != nil {
				return "", err
			}

			if !exists {
				// 直接跳到阿里云解析结束
				return entity.TaskStatusDeleteAliRecordFinish, nil
			}

			// 同集群内有其他成功部署时跳过云解析处理
		}

		return entity.TaskStatusDeleteAliRecordFinish, nil
	case entity.TaskStatusDeleteAliRecordUnderway:
		// 如果老服务在支持多集群后尚未重新部署且未调整过权重, 先保持按照非多集群方式删除两个域名解析
		weights, err := s.GetAppClusterQDNSWeights(ctx, app.ServiceName, task.EnvName)
		if err != nil {
			return "", err
		}

		if len(weights) == 0 {
			err = s.deletePrivateZoneRecordEntry(ctx, app, task)
			if err != nil {
				return "", err
			}
			return entity.TaskStatusDeleteAliRecordFinish, nil
		}

		// 通配域名解析改至 QDNS 统一接入后, 停止 Service 的策略是不操作 Kong 路由规则, 以保留 target 对应的权重信息
		// 停止部署时不操作域名, 集群专用域名后端无 pod, 此时健康检查必定不通过
		// 如果通配域名在 Kong 网关有不止一个 Target, 则通过健康检查会移除不健康路由
		// 如果通配域名在 Kong 网关只有这一个 Target, 则本身就无法提供服务, 不需要关注域名不可用状态
		return entity.TaskStatusDeleteAliRecordFinish, nil
	case entity.TaskStatusDeleteAliRecordFinish:
		err := s.UpdateDeploymentScaleAndIgnoreResponse(ctx,
			task.ClusterName, task.EnvName, task.Namespace, task.Version, task.Param.MinPodCount)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusUpdateDeploymentScaleUnderway, nil
	case entity.TaskStatusUpdateDeploymentScaleUnderway:
		status, err := s.GetDeploymentStatus(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentDetailReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		// 实例数符合
		if int(status.ReadyReplicas) == task.Param.MinPodCount {
			return entity.TaskStatusUpdateDeploymentScaleFinish, nil
		}
	case entity.TaskStatusUpdateCronJobSuspendUnderway:
		job, err := s.GetCronJobDetail(ctx, task.ClusterName,
			&req.GetCronJobDetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			return "", err
		}
		// 停止任务时暂停
		if task.Action == entity.TaskActionStop && *job.Spec.Suspend {
			return entity.TaskStatusUpdateCronJobSuspendFinish, nil
		}
	case entity.TaskStatusUpdateDeploymentScaleFinish, entity.TaskStatusUpdateCronJobSuspendFinish:
		return entity.TaskStatusSuccess, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}
	return "", nil
}

func (s *Service) transformResumeTaskStatus(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, _ *resp.TeamDetailResp) (entity.TaskStatus, error) {
	switch task.Status {
	case entity.TaskStatusInit:
		if app.Type == entity.AppTypeCronJob {
			// 获取当前CronJob
			job, err := s.GetCronJobDetail(ctx, task.ClusterName,
				&req.GetCronJobDetailReq{
					Namespace: task.Namespace,
					Name:      task.Version,
					Env:       string(task.EnvName),
				})
			if err != nil {
				return "", err
			}
			// 恢复当前CronJob
			_, err = s.ResumeCronJob(ctx, task.ClusterName, job, string(task.EnvName))
			if err != nil {
				return "", err
			}
			return entity.TaskStatusUpdateCronJobSuspendUnderway, nil
		}
		err := s.UpdateDeploymentScaleAndIgnoreResponse(ctx,
			task.ClusterName, task.EnvName, task.Namespace, task.Version, task.Param.MinPodCount)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusUpdateDeploymentScaleUnderway, nil
	case entity.TaskStatusUpdateDeploymentScaleUnderway:
		status, err := s.GetDeploymentStatus(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentDetailReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		// 实例数符合
		if int(status.ReadyReplicas) >= task.Param.MinPodCount && int(status.UnavailableReplicas) == 0 {
			return entity.TaskStatusUpdateDeploymentScaleFinish, nil
		}
	case entity.TaskStatusUpdateCronJobSuspendUnderway:
		job, err := s.GetCronJobDetail(ctx, task.ClusterName,
			&req.GetCronJobDetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			return "", err
		}
		// 恢复任务时恢复
		if task.Action == entity.TaskActionResume && !*job.Spec.Suspend {
			return entity.TaskStatusUpdateCronJobSuspendFinish, nil
		}
	case entity.TaskStatusUpdateDeploymentScaleFinish:
		if app.Type != entity.AppTypeService {
			return entity.TaskStatusSuccess, nil
		}

		// return entity.TaskStatusCreateAliRecordUnderway, nil
		return entity.TaskStatusCreateAliRecordFinish, nil
	case entity.TaskStatusCreateAliRecordUnderway:
		// 如果老服务在支持多集群后尚未重新部署且未调整过权重, 先保持按照非多集群方式重新创建两个域名解析
		weights, err := s.GetAppClusterQDNSWeights(ctx, app.ServiceName, task.EnvName)
		if err != nil {
			return "", err
		}

		if len(weights) == 0 {
			err = s.createPrivateZoneRecordEntry(ctx, project, app, task)
			if err != nil {
				return "", err
			}
			return entity.TaskStatusCreateAliRecordFinish, nil
		}

		// 通配域名解析改至 Kong 后, 停止 Service 的策略是不操作 Kong 路由规则, 以保留 Target 对应的权重信息
		// 由于停止部署时并未操作域名, 此时集群专用域名自动恢复, 通配域名在 Kong 网关通过健康检查也自动恢复对应路由
		return entity.TaskStatusCreateAliRecordFinish, nil
	case entity.TaskStatusCreateAliRecordFinish, entity.TaskStatusUpdateCronJobSuspendFinish:
		return entity.TaskStatusSuccess, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}

	return "", nil
}

// 重启任务并指定配置文件的task会走reload_config转换任务
// 该转换任务方法仅仅适用于不指定配置文件的重启任务
func (s *Service) transformRestartTaskStatus(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, _ *resp.TeamDetailResp) (entity.TaskStatus, error) {
	switch task.Status {
	case entity.TaskStatusInit:
		// 重启该Deployment
		err := s.RestartDeploymentAndIgnoreResponse(ctx, task.ClusterName, task.EnvName, &req.RestartDeploymentReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		return entity.TaskStatusRestartDeploymentUnderway, nil

	case entity.TaskStatusRestartDeploymentUnderway:
		rs, err := s.GetReplicaSets(context.Background(), task.ClusterName, string(task.EnvName),
			&req.GetReplicaSetsReq{
				Namespace:   task.Namespace,
				ProjectName: project.Name,
				AppName:     app.Name,
				Version:     task.Version,
			})
		if err != nil {
			return "", err
		}
		// 最新的rs的实例数符合条件且上一次rs的实例数为0
		if len(rs) >= 2 &&
			int(rs[0].Status.ReadyReplicas) >= task.Param.MinPodCount &&
			int(rs[1].Status.ReadyReplicas) == 0 {
			return entity.TaskStatusRestartDeploymentFinish, nil
		}
	case entity.TaskStatusRestartDeploymentFinish:
		return entity.TaskStatusSuccess, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}
	return "", nil
}

func (s *Service) transformDeleteTaskStatus(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, _ *resp.TeamDetailResp) (entity.TaskStatus, error) {
	switch task.Status {
	case entity.TaskStatusInit:
		if app.Type == entity.AppTypeCronJob {
			// 删除cronjob
			err := s.DeleteCronJob(ctx, task.ClusterName,
				&req.DeleteCronJobReq{
					Namespace: task.Namespace,
					Name:      task.Version,
					Env:       string(task.EnvName),
				})
			if err != nil {
				return "", err
			}
			return entity.TaskStatusCleanCronJobUnderway, nil
		}

		if app.Type == entity.AppTypeOneTimeJob {
			err := s.DeleteJob(ctx, task.ClusterName,
				&req.DeleteJobReq{
					Namespace: task.Namespace,
					Name:      task.Version,
					Env:       string(task.EnvName),
				})
			if err != nil {
				return "", err
			}
			return entity.TaskStatusCleanJobUnderway, nil
		}

		if app.Type == entity.AppTypeService {
			// 服务需要检查同集群内是否有其他成功的部署以确定是否需要删除通配域名的阿里云解析
			exists, err := s.CheckHealthyDeploymentExistance(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentsReq{
				Namespace:   task.Namespace,
				ProjectName: project.Name,
				AppName:     app.Name,
				Env:         string(task.EnvName),
			}, task.Version)
			if err != nil {
				return "", err
			}

			if !exists {
				return entity.TaskStatusDeleteAliRecordFinish, nil
			}

			// 同集群同命名空间内有其他成功部署时跳过云解析处理,完成 kong upstream target 的权重更新
		}

		return entity.TaskStatusDeleteAliRecordFinish, nil
	case entity.TaskStatusDeleteAliRecordFinish:
		// 确保cronHPA存在
		_, err := s.GetCronHPADetail(ctx, task.ClusterName,
			&req.GetCronHPADetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanCronHPAFinish, nil
			}
			return "", err
		}

		err = s.DeleteCronHPA(ctx, task.ClusterName, string(task.EnvName),
			&req.DeleteCronHPAReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			return "", err
		}

		return entity.TaskStatusCleanCronHPAUnderway, nil
	case entity.TaskStatusCleanCronHPAUnderway:
		_, err := s.GetCronHPADetail(ctx, task.ClusterName,
			&req.GetCronHPADetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanCronHPAFinish, nil
			}
			return "", err
		}
		return entity.TaskStatusCleanCronHPAFinish, nil
	case entity.TaskStatusCleanCronHPAFinish:
		// 确保hpa存在
		_, err := s.GetHPADetail(ctx, task.ClusterName,
			&req.GetHPADetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanHPAFinish, nil
			}
			return "", err
		}

		// 删除当前版本的HPA
		err = s.DeleteHPA(ctx, task.ClusterName,
			&req.DeleteHPAReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCleanHPAUnderway, nil
	case entity.TaskStatusCleanHPAUnderway:
		_, err := s.GetHPADetail(ctx, task.ClusterName,
			&req.GetHPADetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanHPAFinish, nil
			}
			return "", err
		}
		return entity.TaskStatusCleanHPAFinish, nil
	case entity.TaskStatusCleanHPAFinish:
		err := s.DeleteDeployment(ctx, task.ClusterName, task.EnvName, &req.DeleteDeploymentReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCleanDeploymentUnderway, nil
	case entity.TaskStatusCleanDeploymentUnderway:
		err := s.CheckDeploymentExistance(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentDetailReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanDeploymentFinish, nil
			}
			return "", err
		}
	case entity.TaskStatusCleanCronJobUnderway:
		_, err := s.GetCronJobDetail(ctx, task.ClusterName,
			&req.GetCronJobDetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanCronJobFinish, nil
			}
			return "", err
		}
	case entity.TaskStatusCleanCronJobFinish:
		return entity.TaskStatusCleanDeploymentFinish, nil
	case entity.TaskStatusCleanJobUnderway:
		_, err := s.GetJobDetail(ctx, task.ClusterName,
			&req.GetJobDetailReq{
				Namespace: task.Namespace,
				Name:      task.Version,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanJobFinish, nil
			}
			return "", err
		}
	case entity.TaskStatusCleanJobFinish:
		return entity.TaskStatusCleanDeploymentFinish, nil
	case entity.TaskStatusCleanDeploymentFinish:
		return entity.TaskStatusCleanConfigMapFinish, nil
	case entity.TaskStatusCleanConfigMapFinish:
		// 需要确保是否还有其他部署
		if app.Type == entity.AppTypeCronJob {
			jobs, err := s.GetCronJobs(ctx, task.ClusterName,
				&req.GetCronJobsReq{
					Namespace:   task.Namespace,
					ProjectName: project.Name,
					AppName:     app.Name,
					Env:         string(task.EnvName),
				})
			if err != nil {
				return "", err
			}
			// 包含其他部署，不能删除日志配置
			if len(jobs) > 0 {
				return entity.TaskStatusCleanAliLogConfigFinish, nil
			}
		} else if app.Type == entity.AppTypeOneTimeJob {
			jobs, err := s.GetJobs(ctx, task.ClusterName,
				&req.GetJobsReq{
					Namespace:   task.Namespace,
					ProjectName: project.Name,
					AppName:     app.Name,
					Env:         string(task.EnvName),
				})
			if err != nil {
				return "", err
			}
			// 包含其他部署，不能删除日志配置
			if len(jobs) > 0 {
				return entity.TaskStatusCleanAliLogConfigFinish, nil
			}
		} else {
			exists, err := s.CheckDeploymentsExistance(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentsReq{
				Namespace:   task.Namespace,
				ProjectName: project.Name,
				AppName:     app.Name,
				Env:         string(task.EnvName),
			})
			if err != nil {
				return "", err
			}
			// 包含其他部署，不能删除日志配置
			if exists {
				return entity.TaskStatusCleanAliLogConfigFinish, nil
			}
		}

		err := s.DeleteLogConfig(ctx, task.ClusterName, task.EnvName, app)
		if err != nil {
			if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return "", err
			}

			return entity.TaskStatusCleanAliLogConfigFinish, nil
		}

		return entity.TaskStatusCleanAliLogConfigUnderway, nil
	case entity.TaskStatusCleanAliLogConfigUnderway:
		_, err := s.GetAliLogConfigDetail(ctx, task.ClusterName,
			&req.GetAliLogConfigDetailReq{
				Namespace: task.Namespace,
				Name:      app.AliLogConfigName,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanAliLogConfigFinish, nil
			}
			return "", err
		}
	case entity.TaskStatusCleanAliLogConfigFinish:
		// 移除 redis 中的部署记录
		err := s.dao.RemoveAppClusterRunningTasks(ctx, task.AppID, task.EnvName, task.ClusterName, task.Version)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusSuccess, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}
	return "", nil
}

// transformCleanTaskStatus 清理应用信息
func (s *Service) transformCleanTaskStatus(ctx context.Context, task *resp.TaskDetailResp) (
	entity.TaskStatus, error) {
	switch task.Status {
	case entity.TaskStatusInit:
		if task.Param.CleanedAppType == entity.AppTypeCronJob {
			// 需要清理CronJobs
			if task.Param.CleanedProjectName != "" && task.Param.CleanedAppName != "" {
				err := s.DeleteCronJobs(ctx, task.ClusterName,
					&req.DeleteCronJobsReq{
						Namespace:   task.Namespace,
						ProjectName: task.Param.CleanedProjectName,
						AppName:     task.Param.CleanedAppName,
						Env:         string(task.EnvName),
					})
				if err != nil {
					return "", err
				}
				return entity.TaskStatusCleanCronJobUnderway, nil
			}
			return entity.TaskStatusCleanCronJobFinish, nil
		} else if task.Param.CleanedAppType == entity.AppTypeOneTimeJob {
			// 需要清理Jobs
			if task.Param.CleanedProjectName != "" && task.Param.CleanedAppName != "" {
				err := s.DeleteJobs(ctx, task.ClusterName,
					&req.DeleteJobsReq{
						Namespace:   task.Namespace,
						ProjectName: task.Param.CleanedProjectName,
						AppName:     task.Param.CleanedAppName,
						Env:         string(task.EnvName),
					})
				if err != nil {
					return "", err
				}
				return entity.TaskStatusCleanJobUnderway, nil
			}
			return entity.TaskStatusCleanJobFinish, nil
		} else {
			// 需要清理cronHPA
			if task.Param.CleanedProjectName != "" && task.Param.CleanedAppName != "" {
				err := s.DeleteCronHPAs(ctx, task.ClusterName,
					&req.DeleteCronHPAsReq{
						Namespace:   task.Namespace,
						ProjectName: task.Param.CleanedProjectName,
						Env:         string(task.EnvName),
						AppName:     task.Param.CleanedAppName,
					})
				if err != nil {
					return "", err
				}
				return entity.TaskStatusCleanCronHPAUnderway, nil
			}
			return entity.TaskStatusCleanDeploymentFinish, nil
		}
	case entity.TaskStatusCleanCronHPAUnderway:
		cronHPAlist, err := s.GetCronHPAs(ctx, task.ClusterName,
			&req.GetCronHPAsReq{
				Namespace:   task.Namespace,
				ProjectName: task.Param.CleanedProjectName,
				AppName:     task.Param.CleanedAppName,
				Env:         string(task.EnvName),
			})
		if err != nil {
			return "", err
		}

		if len(cronHPAlist) == 0 {
			return entity.TaskStatusCleanCronHPAFinish, nil
		}
	case entity.TaskStatusCleanCronHPAFinish:
		err := s.DeleteHPAs(ctx, task.ClusterName,
			&req.DeleteHPAsReq{
				Namespace:   task.Namespace,
				ProjectName: task.Param.CleanedProjectName,
				AppName:     task.Param.CleanedAppName,
				Env:         string(task.EnvName),
			})
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCleanHPAUnderway, nil
	case entity.TaskStatusCleanHPAUnderway:
		hpaList, err := s.GetHPAs(ctx, task.ClusterName,
			&req.GetHPAsReq{
				Namespace:   task.Namespace,
				ProjectName: task.Param.CleanedProjectName,
				AppName:     task.Param.CleanedAppName,
				Env:         string(task.EnvName),
			})
		if err != nil {
			return "", err
		}
		if len(hpaList) == 0 {
			return entity.TaskStatusCleanHPAFinish, nil
		}
	case entity.TaskStatusCleanHPAFinish:
		err := s.DeleteDeployments(ctx, task.ClusterName, task.EnvName, &req.DeleteDeploymentsReq{
			Namespace:   task.Namespace,
			ProjectName: task.Param.CleanedProjectName,
			AppName:     task.Param.CleanedAppName,
			Env:         string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCleanDeploymentUnderway, nil
	case entity.TaskStatusCleanDeploymentUnderway:
		exists, err := s.CheckDeploymentsExistance(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentsReq{
			Namespace:   task.Namespace,
			ProjectName: task.Param.CleanedProjectName,
			AppName:     task.Param.CleanedAppName,
			Env:         string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		if !exists {
			return entity.TaskStatusCleanDeploymentFinish, nil
		}
	case entity.TaskStatusCleanDeploymentFinish:
		if task.Param.CleanedAppType != entity.AppTypeService || task.Param.CleanedServiceName == "" {
			return entity.TaskStatusCleanK8sServiceFinish, nil
		}

		return entity.TaskStatusCleanAliServiceFinish, nil

		err := s.deletePrivateZoneRecordEntry(ctx, &resp.AppDetailResp{
			Name:              task.Param.CleanedAppName,
			Type:              task.Param.CleanedAppType,
			ServiceType:       task.Param.CleanedAppServiceType,
			ServiceExposeType: task.Param.CleanedAppServiceExposeType,
			ServiceName:       task.Param.CleanedServiceName,
		}, task)
		if err != nil && !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) { // 并非每个环境都有
			return "", err
		}

		// 支持多集群后通配域名可能是 Kong 的解析记录
		err = s.deleteKongPrivateZoneRecordForK8sDomain(ctx, task.Param.CleanedAppServiceExposeType,
			task.Param.CleanedServiceName, task)
		if err != nil && !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) { // 并非每个环境都有
			return "", err
		}

		return entity.TaskStatusCleanAliServiceFinish, nil

	case entity.TaskStatusCleanAliServiceFinish:
		ns := task.Namespace
		if strings.HasPrefix(ns, entity.IstioNamespacePrefix) {
			ns = strings.ReplaceAll(ns, entity.IstioNamespacePrefix, "")
		}
		_, err := s.GetIngressDetail(ctx, task.ClusterName,
			&req.GetIngressDetailReq{
				Namespace: ns,
				Name:      task.Param.CleanedServiceName,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanK8sIngressFinish, nil
			}
			return "", err
		}

		err = s.DeleteIngress(ctx, task.ClusterName,
			&req.DeleteIngressReq{
				Namespace: ns,
				Name:      task.Param.CleanedServiceName,
				Env:       string(task.EnvName),
			})
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCleanK8sIngressUnderway, nil
	case entity.TaskStatusCleanK8sIngressUnderway:
		ns := task.Namespace
		if strings.HasPrefix(ns, entity.IstioNamespacePrefix) {
			ns = strings.ReplaceAll(ns, entity.IstioNamespacePrefix, "")
		}
		_, err := s.GetIngressDetail(ctx, task.ClusterName,
			&req.GetIngressDetailReq{
				Namespace: ns,
				Name:      task.Param.CleanedServiceName,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanK8sIngressFinish, nil
			}
			return "", err
		}

		return entity.TaskStatusCleanK8sIngressFinish, nil
	// 清理 k8s ingress 完成
	case entity.TaskStatusCleanK8sIngressFinish:
		return entity.TaskStatusCleanVirtualServiceUnderWay, nil
	// 开始清理 istio virtual service 资源
	case entity.TaskStatusCleanVirtualServiceUnderWay:
		ns := task.Namespace
		if !strings.HasPrefix(task.Namespace, entity.IstioNamespacePrefix) {
			ns = entity.IstioNamespacePrefix + task.Namespace
		}
		_, err := s.GetVirtualServiceDetail(ctx, task.ClusterName, task.EnvName, &req.VirtualServiceReq{
			Namespace: ns,
			Name:      task.Param.CleanedServiceName,
			Env:       string(task.EnvName),
		})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanVirtualServiceFinish, nil
			}

			return "", err
		}

		err = s.DeleteVirtualService(ctx, task.ClusterName, task, task.Param.CleanedServiceName, ns)
		if err != nil {
			return "", err
		}

		return entity.TaskStatusCleanVirtualServiceFinish, nil
	// 清理 istio virtual service 资源中
	case entity.TaskStatusCleanVirtualServiceFinish:
		app, err := s.GetAppDetailByIDAndIgnoreDeleteStatus(ctx, task.AppID)
		if err != nil {
			return "", err
		}

		err = s.cleanQDNSBusiness(ctx, app.ServiceName, task)
		if err != nil {
			return "", err
		}

		return entity.TaskStatusCleanKongRecordFinish, nil
	case entity.TaskStatusCleanKongRecordFinish:
		_, err := s.GetServiceDetail(ctx, task.ClusterName,
			&req.GetServiceDetailReq{
				Namespace: task.Namespace,
				Name:      task.Param.CleanedServiceName,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanK8sServiceFinish, nil
			}
			return "", err
		}

		err = s.DeleteService(ctx, task.ClusterName,
			&req.DeleteServiceReq{
				Namespace: task.Namespace,
				Name:      task.Param.CleanedServiceName,
				Env:       string(task.EnvName),
			})
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCleanK8sServiceUnderway, nil
	case entity.TaskStatusCleanK8sServiceUnderway:
		_, err := s.GetServiceDetail(ctx, task.ClusterName,
			&req.GetServiceDetailReq{
				Namespace: task.Namespace,
				Name:      task.Param.CleanedServiceName,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanK8sServiceFinish, nil
			}
			return "", err
		}
	case entity.TaskStatusCleanCronJobUnderway:
		jobs, err := s.GetCronJobs(ctx, task.ClusterName,
			&req.GetCronJobsReq{
				Namespace:   task.Namespace,
				ProjectName: task.Param.CleanedProjectName,
				AppName:     task.Param.CleanedAppName,
				Env:         string(task.EnvName),
			})
		if err != nil {
			return "", err
		}
		if len(jobs) == 0 {
			return entity.TaskStatusCleanCronJobFinish, nil
		}
	case entity.TaskStatusCleanCronJobFinish:
		return entity.TaskStatusCleanK8sServiceFinish, nil
	case entity.TaskStatusCleanJobUnderway:
		jobs, err := s.GetJobs(ctx, task.ClusterName,
			&req.GetJobsReq{
				Namespace:   task.Namespace,
				ProjectName: task.Param.CleanedProjectName,
				AppName:     task.Param.CleanedAppName,
				Env:         string(task.EnvName),
			})
		if err != nil {
			return "", err
		}
		if len(jobs) == 0 {
			return entity.TaskStatusCleanJobFinish, nil
		}
	case entity.TaskStatusCleanJobFinish:
		return entity.TaskStatusCleanK8sServiceFinish, nil
	case entity.TaskStatusCleanK8sServiceFinish:
		if task.Param.CleanedAliLogConfigName == "" {
			return entity.TaskStatusCleanAliLogConfigFinish, nil
		}

		// 禁用日志接入时跳过清理
		clusterInfo, err := s.getClusterInfo(task.ClusterName, task.Namespace)
		if err != nil {
			return "", err
		}

		if clusterInfo.disableLogConfig {
			return entity.TaskStatusCleanAliLogConfigFinish, nil
		}

		_, err = s.GetAliLogConfigDetail(ctx, task.ClusterName,
			&req.GetAliLogConfigDetailReq{
				Namespace: task.Namespace,
				Name:      task.Param.CleanedAliLogConfigName,
				Env:       string(task.EnvName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanAliLogConfigFinish, nil
			}
			return "", err
		}

		err = s.DeleteAliLogConfig(ctx, task.ClusterName, string(task.EnvName),
			&req.DeleteAliLogConfigReq{
				Namespace: task.Namespace,
				Name:      task.Param.CleanedAliLogConfigName,
			})
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCleanAliLogConfigUnderway, nil
	case entity.TaskStatusCleanAliLogConfigUnderway:
		_, err := s.GetAliLogConfigDetail(ctx, task.ClusterName,
			&req.GetAliLogConfigDetailReq{
				Namespace: task.Namespace,
				Name:      task.Param.CleanedAliLogConfigName,
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return entity.TaskStatusCleanAliLogConfigFinish, nil
			}
			return "", err
		}
	case entity.TaskStatusCleanAliLogConfigFinish:
		err := s.DeleteLogStore(ctx, task)
		if err != nil {
			return "", err
		}

		return entity.TaskStatusCleanAliLogStoreFinish, nil
	case entity.TaskStatusCleanAliLogStoreFinish:
		err := s.DeleteConfigMaps(ctx, task.ClusterName, &req.DeleteConfigMapsReq{
			Namespace:   task.Namespace,
			ProjectName: task.Param.CleanedProjectName,
			AppName:     task.Param.CleanedAppName,
			Env:         string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		return entity.TaskStatusCleanConfigMapUnderway, nil
	case entity.TaskStatusCleanConfigMapUnderway:
		cms, err := s.ListConfigMaps(ctx, task.ClusterName, &req.ListConfigMapsReq{
			Namespace:   task.Namespace,
			ProjectName: task.Param.CleanedProjectName,
			AppName:     task.Param.CleanedAppName,
			Env:         string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		if len(cms) == 0 {
			return entity.TaskStatusCleanConfigMapFinish, nil
		}
	case entity.TaskStatusCleanConfigMapFinish:
		// 清理 redis 中的部署记录
		err := s.dao.CleanAppRunningTasks(ctx, task.AppID, task.EnvName, entity.EmptyClusterName, dao.NoInverseVersion)
		if err != nil {
			return "", err
		}
		return entity.TaskStatusSuccess, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}
	return "", nil
}

// transformReloadConfigStatus transform reload config task status.
func (s *Service) transformReloadConfigStatus(ctx context.Context, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, task *resp.TaskDetailResp, team *resp.TeamDetailResp) (entity.TaskStatus, error) {
	switch task.Status {
	case entity.TaskStatusInit:
		getReq := &req.GetConfigManagerFileReq{
			ProjectID:   project.ID,
			ProjectName: project.Name,
			EnvName:     task.EnvName,
			CommitID:    task.Param.ConfigCommitID,
			IsDecrypt:   true,
			FormatType:  req.ConfigManagerFormatTypeJSON,
		}

		if task.Param.ConfigRenamePrefix != "" {
			getReq.ConfigRenamePrefix = task.Param.ConfigRenamePrefix
			getReq.ConfigRenameMode = task.Param.ConfigRenameMode
		}

		configData, err := s.GetAppConfig(ctx, getReq)
		if err != nil {
			return "", err
		}
		// Render template, update configMap's commit id and config data hash.
		data, err := s.RenderConfigMapTemplate(ctx, project, app, task, team, configData.Config)
		if err != nil {
			return "", err
		}
		// Apply ConfigMap.
		_, err = s.ApplyConfigMap(ctx, task.ClusterName, data, string(task.EnvName))
		if err != nil {
			return "", err
		}
		return entity.TaskStatusUpdateConfigMapUnderway, nil
	case entity.TaskStatusUpdateConfigMapUnderway:
		cm, err := s.GetCompatibleConfigMapDetail(ctx, project, app, task)
		if err == nil {
			if cm.GetLabels()[ConfigMapLabelCommit] == task.Param.ConfigCommitID {
				return entity.TaskStatusUpdateConfigMapFinish, nil
			}
		} else if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return "", err
		}
	case entity.TaskStatusUpdateConfigMapFinish:
		// Update or add deployment's label "configHash",
		// pod annotation "shanhai.int/config-map-hash".
		cm, err := s.GetCompatibleConfigMapDetail(ctx, project, app, task)
		if err != nil {
			return "", err
		}

		err = s.ReloadDeployment(ctx, task, cm.GetLabels()[ConfigMapLabelHash])
		if err != nil {
			return "", err
		}
		return entity.TaskStatusUpdateDeploymentAnnotationUnderway, nil
	case entity.TaskStatusUpdateDeploymentAnnotationUnderway:
		cm, err := s.GetCompatibleConfigMapDetail(ctx, project, app, task)
		if err != nil {
			return "", err
		}
		hash := cm.GetLabels()[ConfigMapLabelHash]

		deployment, err := s.GetDeploymentDetail(ctx, task.ClusterName, task.EnvName, &req.GetDeploymentDetailReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return "", err
		}
		if deployment.GetLabels()[LabelConfigHash] != hash {
			return entity.TaskStatusUpdateDeploymentAnnotationUnderway, nil
		}

		pods, err := s.GetPods(ctx, task.ClusterName, &req.GetPodsReq{
			Namespace:   task.Namespace,
			ProjectName: project.Name,
			AppName:     app.Name,
			Env:         string(task.EnvName),
		})
		if err != nil {
			return "", err
		}

		for i := range pods {
			if pods[i].Annotations[AnnotationConfigHash] != hash {
				return entity.TaskStatusUpdateDeploymentAnnotationUnderway, nil
			}
		}
		return entity.TaskStatusUpdateDeploymentAnnotationFinish, nil
	case entity.TaskStatusUpdateDeploymentAnnotationFinish:
		return entity.TaskStatusSuccess, nil
	default:
		return "", _errcode.InvalidTaskStatusError
	}
	return "", nil
}
