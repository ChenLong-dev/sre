package handlers

import (
	"fmt"
	"net/http"
	"time"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/service"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// GetRunningStatusList 获取正在执行中的状态列表
func GetRunningStatusList(c *gin.Context) {
	getReq := new(req.GetRunningStatusListReq)
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

	// 获取详情
	app, err := service.SVC.GetAppDetail(c, getReq.AppID)
	if errcode.EqualError(_errcode.InvalidHexStringError, err) || errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	project, err := service.SVC.GetProjectDetail(c, app.ProjectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	res, err := service.SVC.GetRunningStatusList(c, getReq.ClusterName, getReq, project, app)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, res, nil)
}

// GetRunningStatusDetail 获取正在执行中的状态详情
func GetRunningStatusDetail(c *gin.Context) {
	version := c.Param("version")
	getReq := new(req.GetRunningStatusDetailReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	getReq.Version = version

	// 校验集群名
	getReq.ClusterName, err = service.SVC.CheckAndUnifyClusterName(c, getReq.ClusterName, getReq.EnvName)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 兼容旧版本 ams portal
	if getReq.Namespace == "" {
		getReq.Namespace = string(getReq.EnvName)
	}

	app, err := service.SVC.GetAppDetail(c, getReq.AppID)
	if errcode.EqualError(_errcode.InvalidHexStringError, err) || errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	res, err := service.SVC.GetRunningStatusDetail(c, getReq, app)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, res, nil)
}

// GetRunningPodLogs 获取正在运行中的 pod 日志
func GetRunningPodLogs(c *gin.Context) {
	podName := c.Param("pod_name")
	getReq := new(req.GetRunningPodLogsReq)
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

	// 兼容老版本 ams portal
	if getReq.Namespace == "" {
		getReq.Namespace = string(getReq.EnvName)
	}

	l, err := service.SVC.GetPodLog(c, getReq.ClusterName,
		&req.GetPodLogReq{
			Namespace:     getReq.Namespace,
			Name:          podName,
			Env:           string(getReq.EnvName),
			ContainerName: getReq.ContainerName,
		})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	c.String(http.StatusOK, l)
}

// CreateRunningPodPProf 创建正在运行中的 pod 的 pprof 监控信息
func CreateRunningPodPProf(c *gin.Context) {
	podName := c.Param("pod_name")
	createReq := new(req.CreatePProfReq)
	err := c.ShouldBindJSON(createReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	// 校验集群名
	createReq.ClusterName, err = service.SVC.CheckAndUnifyClusterName(c, createReq.ClusterName, createReq.EnvName)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 兼容老版本 ams portal
	if createReq.Namespace == "" {
		createReq.Namespace = string(createReq.EnvName)
	}

	// 检查语言
	app, err := service.SVC.GetAppDetail(c, createReq.AppID)
	if errcode.EqualError(_errcode.InvalidHexStringError, err) || errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	project, err := service.SVC.GetProjectDetail(c, app.ProjectID)

	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	if project.Language != entity.LanguageGo {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "project "+
			"language is %s not %s", project.Language, entity.LanguageGo))
		return
	}

	createReq.Container = utils.GetPodContainerName(project.Name, app.Name)
	// 初始化参数
	createReq.PodName = podName
	if createReq.PodPort == 0 {
		createReq.PodPort = 8089
	}
	unix := time.Now().Unix()
	createReq.SourceFilePath = fmt.Sprintf(
		"./pprof/source/%s-%s-%s-%d",
		createReq.PodName, createReq.Type, createReq.Action, unix,
	)
	var generateFileType string
	if createReq.Action == entity.PProfActionSvg {
		generateFileType = ".svg"
	}
	createReq.GenerateFilePath = fmt.Sprintf(
		"./pprof/generate/%s-%s-%s-%s-%d%s",
		createReq.ClusterName, createReq.PodName, createReq.Type, createReq.Action,
		unix, generateFileType,
	)

	log.Infoc(c, "start pprof: %#v", createReq)

	// 创建pprof原始文件
	err = service.SVC.GenerateOriginPProfFile(c, createReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 直接下载
	if createReq.Action == entity.PProfActionDownload || createReq.Type == entity.PProfTypeTrace {
		c.File(createReq.SourceFilePath)
		return
	}

	// 生成pprof分析文件
	err = service.SVC.GenerateAnalyzedPProfFile(c, createReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	c.File(createReq.GenerateFilePath)
}

// GetRunningPodDescription 获取正在运行中的 pod 的 k8s describe 信息
func GetRunningPodDescription(c *gin.Context) {
	podName := c.Param("pod_name")
	getReq := new(req.GetRunningPodDescriptionReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err.Error()))
		return
	}

	// 校验集群名
	getReq.ClusterName, err = service.SVC.CheckAndUnifyClusterName(c, getReq.ClusterName, entity.AppEnvName(getReq.EnvName))
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// 兼容老版本 ams portal
	if getReq.Namespace == "" {
		getReq.Namespace = getReq.EnvName
	}

	podDesc, err := service.SVC.DescribePod(c, getReq.ClusterName,
		&req.GetRunningPodDescriptionReq{
			Name:      podName,
			EnvName:   getReq.EnvName,
			Env:       getReq.EnvName,
			Namespace: getReq.Namespace,
			Container: getReq.ContainerName,
		})
	if err != nil {
		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			response.JSON(c, new(resp.DescribePodResp), nil)
			return
		}
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, podDesc, nil)
}

// GetRunningStatusDescription 获取正在运行中的资源状态机的 k8s describe 信息
func GetRunningStatusDescription(c *gin.Context) {
	version := c.Param("version")
	getReq := new(req.GetRunningStatusDescriptionReq)
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

	app, err := service.SVC.GetAppDetail(c, getReq.AppID)
	if errcode.EqualError(_errcode.InvalidHexStringError, err) || errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	res := new(resp.GetRunningStatusDescriptionResp)
	namespace := service.SVC.GetNamespaceBase(service.SVC.GetApplicationIstioState(c, getReq.EnvName,
		getReq.ClusterName, app), getReq.EnvName)
	switch app.Type {
	case entity.AppTypeService:
		serviceName, err := service.SVC.GetCurrentServiceName(c, getReq.ClusterName, getReq.EnvName, app)
		if err != nil {
			log.Errorc(c, "get serviceName failed app: %s, err: %s", app.Name, err.Error())
		} else {
			svcDesc, intErr := service.SVC.DescribeService(c, getReq.ClusterName,
				&req.DescribeServiceReq{
					Namespace: namespace,
					Name:      serviceName,
					Env:       string(getReq.EnvName),
				})
			if intErr != nil {
				log.Errorc(c, "describe service failed app: %s, err: %s", app.Name, intErr.Error())
			}
			res.ServiceDesc = svcDesc
		}

		deployDesc, err := service.SVC.DescribeDeployment(c, getReq.ClusterName, getReq.EnvName,
			&req.DescribeDeploymentReq{
				Namespace: namespace,
				Name:      version,
				Env:       string(getReq.EnvName),
			})
		if err != nil {
			log.Errorc(c, "describe deployment failed app: %s, err: %s", app.Name, err.Error())
		}
		res.DeploymentDesc = deployDesc

		hpaDesc, err := service.SVC.DescribeHPA(c, getReq.ClusterName,
			&req.DescribeHPAReq{
				Namespace: namespace,
				Name:      version,
				Env:       string(getReq.EnvName),
			})
		if err != nil && !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			log.Errorc(c, "describe hpa failed app: %s, err: %s", app.Name, err.Error())
		}
		res.HPADesc = hpaDesc

		if service.SVC.GetApplicationIstioState(c, getReq.EnvName, getReq.ClusterName, app) {
			project, err := service.SVC.GetProjectDetail(c, app.ProjectID)
			if err != nil {
				response.JSON(c, nil, err)
				return
			}

			vsDesc, err := service.SVC.DescribeVirtualService(c, getReq.ClusterName, getReq.EnvName,
				&req.DescribeVirtualServiceReq{
					Namespace: namespace,
					Name:      utils.GetPodContainerName(project.Name, app.Name),
					Env:       string(getReq.EnvName),
				})
			if err != nil {
				log.Errorc(c, "describe vs failed app: %s, err: %s", app.Name, err.Error())
			}

			res.VirtualServiceDesc = vsDesc
		} else {
			ingressDesc, err := service.SVC.DescribeIngress(c, getReq.ClusterName,
				&req.DescribeIngressReq{
					Namespace: namespace,
					Name:      app.ServiceName,
					Env:       string(getReq.EnvName),
				})
			if err != nil {
				log.Errorc(c, "describe ingress failed app: %s, err: %s", app.Name, err.Error())
			}

			res.IngressDesc = ingressDesc
		}

	case entity.AppTypeCronJob:
		cronjobDesc, err := service.SVC.DescribeCronJob(c, getReq.ClusterName,
			&req.DescribeCronJobReq{
				Namespace: namespace,
				Name:      version,
				Env:       string(getReq.EnvName),
			})
		if err != nil {
			log.Errorc(c, "describe cronjob failed app: %s, err: %s", app.Name, err.Error())
		}
		res.CronJobDesc = cronjobDesc
	case entity.AppTypeOneTimeJob:
		jobDesc, err := service.SVC.DescribeJob(c, getReq.ClusterName,
			&req.DescribeJobReq{
				Namespace: namespace,
				Name:      version,
				Env:       string(getReq.EnvName),
			})
		if err != nil {
			log.Errorc(c, "describe oneTimeJob failed app: %s, err: %s", app.Name, err.Error())
		}
		res.JobDesc = jobDesc
	case entity.AppTypeWorker:
		deploymentDesc, err := service.SVC.DescribeDeployment(c, getReq.ClusterName, getReq.EnvName,
			&req.DescribeDeploymentReq{
				Namespace: namespace,
				Name:      version,
				Env:       string(getReq.EnvName),
			})
		if err != nil {
			log.Errorc(c, "describe deployment failed app: %s, err: %s", app.Name, err.Error())
		}
		res.DeploymentDesc = deploymentDesc

		hpaDesc, err := service.SVC.DescribeHPA(c, getReq.ClusterName,
			&req.DescribeHPAReq{
				Namespace: namespace,
				Name:      version,
				Env:       string(getReq.EnvName),
			})
		if err != nil && !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			log.Errorc(c, "describe hpa failed app: %s, err: %s", app.Name, err.Error())
		}
		res.HPADesc = hpaDesc
	default:
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "app type: %s is invalid", app.Type))
		return
	}

	response.JSON(c, res, nil)
}
