package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
)

// 特殊标签值
const (
	K8sAnnotationRestart    = "kubectl.kubernetes.io/restartedAt"
	K8sAnnotationHPASkipped = "HPARollingUpdateSkipped"

	QingTingAnnotationMetricsPort = "shanhai.int/metrics-port"
	RestartAnnotationLayout       = "2006-01-02T15:04:05-07:00"
)

// 系统环境变量
const (
	SystemEnvKeyAppID              = "RULAI_APP_ID"
	SystemEnvKeyAppName            = "RULAI_APP_NAME"
	SystemEnvKeyAppEnv             = "RULAI_APP_ENV"
	SystemEnvKeyProjectName        = "RULAI_PROJECT_NAME"
	SystemEnvKeyProjectID          = "RULAI_PROJECT_ID"
	SystemEnvKeyHost               = "RULAI_HOST"
	SystemEnvKeyAuthorizationToken = "RULAI_AUTHORIZATION_TOKEN"
	// 云服务商相关环境变量(需要动态获取)
	SystemEnvKeyVendor = "RULAI_VENDOR"
	SystemEnvKeyRegion = "RULAI_REGION" // 云服务商定义的区域
)

const (
	// MetricsLabelEnable 监控标签，表示该标签有效
	MetricsLabelEnable = "true"
	// AnnotationConfigHash Annotation reload configHash
	AnnotationConfigHash = "shanhai.int/config-map-hash"
	// LabelConfigHash Label configHash
	LabelConfigHash = "configHash"
	// LabelCanonicalName Istio CanonicalName
	LabelCanonicalName = "service.istio.io/canonical-name"
)

const (
	// 通过环境变量管理阿里云日志配置
	AliyunLogConfigLogstorePattern    = "aliyun_logs_%s"
	AliyunLogConfigTagsPattern        = "aliyun_logs_%s_tags"
	AliyunLogConfigLogstoreTTLPattern = "aliyun_logs_%s_ttl"

	AliyunLogConfigLogstoreTTL = "7"
)

// SortDeploymentReverseSlice Deployment的创建时间倒序数组
type SortDeploymentReverseSlice []v1beta1.Deployment

func (s SortDeploymentReverseSlice) Len() int      { return len(s) }
func (s SortDeploymentReverseSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortDeploymentReverseSlice) Less(i, j int) bool {
	return s[i].GetCreationTimestamp().After(s[j].GetCreationTimestamp().Time)
}

func (s *Service) initDeploymentTemplate(ctx context.Context, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, task *resp.TaskDetailResp, team *resp.TeamDetailResp) (*entity.DeploymentTemplate, error) {
	podAnnotations, err := s.getWorkloadAnnotationsByVendor(ctx, task.ClusterName, task.EnvName, project, app, team)
	if err != nil {
		return nil, err
	}

	if task.Param.MetricsPort != 0 {
		podAnnotations[QingTingAnnotationMetricsPort] = strconv.Itoa(task.Param.MetricsPort)
	}

	var configName string
	cm, e := s.GetCompatibleConfigMapDetail(ctx, project, app, task)
	if e != nil {
		log.Warnc(ctx, "get k8s configmap:%s error:%s", s.GetAppNewConfigMapName(task), e)
	} else {
		configName = cm.Name
	}

	envVars, err := s.getSystemEnv(project, app, task, team)
	if err != nil {
		return nil, err
	}

	for _, key := range utils.SortedMapKeys(task.Param.Vars) {
		envVars = append(envVars, &entity.EnvTemplate{
			Name:  key,
			Value: task.Param.Vars[key],
		})
	}

	cluster, err := s.getCluster(task)
	if err != nil {
		return nil, err
	}

	template := &entity.DeploymentTemplate{
		ProjectName:                   project.Name,
		AppName:                       app.Name,
		AppServiceType:                app.ServiceType,
		DeploymentVersion:             task.Version,
		Namespace:                     task.Namespace,
		Replicas:                      int32(task.Param.MinPodCount),
		Labels:                        s.getLabels(project, app, task, team),
		PodAnnotations:                podAnnotations,
		LogStoreName:                  project.LogStoreName,
		ImageName:                     entity.ImageName(task.Param.ImageVersion),
		ContainerName:                 utils.GetPodContainerName(project.Name, app.Name),
		Env:                           envVars,
		CoverCommand:                  task.Param.CoverCommand,
		PreStopCommand:                task.Param.PreStopCommand,
		TerminationGracePeriodSeconds: task.Param.TerminationGracePeriodSeconds,
		EnableHealth:                  app.Type == entity.AppTypeService,
		HealthCheckURL:                task.Param.HealthCheckURL,
		TargetPort:                    int32(task.Param.TargetPort),
		MetricsPort:                   int32(task.Param.MetricsPort),
		ConfigName:                    configName,
		ConfigMountPath:               task.Param.ConfigMountPath,
		CPULimit:                      task.Param.CPULimit,
		MemoryLimit:                   task.Param.MemLimit,
		CPURequest:                    task.Param.CPURequest,
		MemoryRequest:                 task.Param.MemRequest,
		NodeAffinity: entity.GenerateNodeAffinityTemplate(
			app.Type,
			task.Param.NodeAffinityLabelConfig,
		),
		DisableHighAvailability:           task.Param.DisableHighAvailability,
		LivenessProbeInitialDelaySeconds:  task.Param.LivenessProbeInitialDelaySeconds,
		ReadinessProbeInitialDelaySeconds: task.Param.ReadinessProbeInitialDelaySeconds,
		LocalDNS:                          cluster.localDNS,
		// apps/v1 和 extensions/v1beta1 默认值不一致，保持一致
		ProgressDeadlineSeconds: math.MaxInt32,
	}

	if app.Type == entity.AppTypeService && app.ServiceType == entity.AppServiceTypeGRPC {
		cfg, e := s.getGRPCHealthProbeConfig(ctx, app.ID)
		if e != nil {
			return nil, e
		}

		if cfg != nil {
			template.GRPCHealthProbePort = cfg.Port
			template.GRPCHealthProbeUseTLS = cfg.TLS
		}
	}

	return template, nil
}

func (s *Service) getCluster(task *resp.TaskDetailResp) (*k8sCluster, error) {
	envClusters, ok := s.k8sClusters[task.EnvName]
	if !ok {
		return nil, errors.Wrap(_errcode.ClusterNotExistsError, fmt.Sprintf("环境为[%s]集群不存在", task.EnvName))
	}

	cluster, ok := envClusters[task.ClusterName]
	if !ok {
		return nil, errors.Wrap(_errcode.ClusterNotExistsError, fmt.Sprintf("名字为[%s]集群不存在", task.ClusterName))
	}

	return cluster, nil
}

func (s *Service) decodeExtensionsV1Beta1DeploymentYamlData(_ context.Context, yamlData []byte) (*v1beta1.Deployment, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(yamlData, nil, nil)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	deployment, ok := obj.(*v1beta1.Deployment)
	if !ok {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "object is not v1beta1.Deployment")
	}

	return deployment, nil
}

func (s *Service) decodeAppsV1DeploymentYamlData(_ context.Context, yamlData []byte) (*v1.Deployment, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(yamlData, nil, nil)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	deployment, ok := obj.(*v1.Deployment)
	if !ok {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "object is not v1.Deployment")
	}
	return deployment, nil
}

// RenderDeploymentTemplate 渲染Deployment模板
func (s *Service) RenderDeploymentTemplate(ctx context.Context,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, task *resp.TaskDetailResp, team *resp.TeamDetailResp) ([]byte, error) {
	tpl, err := s.initDeploymentTemplate(ctx, project, app, task, team)
	if err != nil {
		return nil, err
	}

	data, err := s.RenderK8sTemplate(ctx, entity.DefaultTemplateFileDir, task.ClusterName, task.EnvName, tpl)
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

// ApplyDeploymentAndIgnoreResponse 声明式创建/更新Deployment，忽略返回值
func (s *Service) ApplyDeploymentAndIgnoreResponse(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, yamlData []byte) error {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return err
	}

	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		_, err = s.ApplyExtensionsV1Beta1Deployment(ctx, clusterName, yamlData, string(envName))
		return err
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		_, err = s.ApplyAppsV1Deployment(ctx, clusterName, yamlData, string(envName))
		return err
	}

	return errors.Wrapf(_errcode.K8sInternalError,
		"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
}

// DescribeDeployment 获取 Deployment 的 Describe 信息
func (s *Service) DescribeDeployment(ctx context.Context, clusterName entity.ClusterName,
	envName entity.AppEnvName, descReq *req.DescribeDeploymentReq) (*resp.DescribeDeploymentResp, error) {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return nil, err
	}

	var resource runtime.Object
	res := new(resp.DescribeDeploymentResp)
	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		deployment, e := s.GetExtensionsV1Beta1DeploymentDetail(ctx,
			clusterName,
			&req.GetDeploymentDetailReq{
				Namespace: descReq.Namespace,
				Name:      descReq.Name,
				Env:       descReq.Env,
			})
		if e != nil {
			return nil, e
		}

		e = deepcopy.Copy(deployment).To(res)
		if e != nil {
			return nil, errors.Wrap(errcode.InternalError, e.Error())
		}

		resource = deployment
		res.Status.Conditions = make([]resp.DeploymentCondition, len(deployment.Status.Conditions))
		for i := range res.Status.Conditions {
			res.Status.Conditions[i].LastUpdateTime = utils.FormatK8sTime(
				&deployment.Status.Conditions[i].LastUpdateTime)
			res.Status.Conditions[i].LastTransitionTime = utils.FormatK8sTime(
				&deployment.Status.Conditions[i].LastTransitionTime)
		}
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		deployment, e := s.GetAppsV1DeploymentDetail(ctx,
			clusterName,
			&req.GetDeploymentDetailReq{
				Namespace: descReq.Namespace,
				Name:      descReq.Name,
				Env:       descReq.Env,
			})
		if e != nil {
			return nil, e
		}

		e = deepcopy.Copy(deployment).To(res)
		if e != nil {
			return nil, errors.Wrap(errcode.InternalError, e.Error())
		}

		resource = deployment
		res.Status.Conditions = make([]resp.DeploymentCondition, len(deployment.Status.Conditions))
		for i := range res.Status.Conditions {
			res.Status.Conditions[i].LastUpdateTime = utils.FormatK8sTime(
				&deployment.Status.Conditions[i].LastUpdateTime)
			res.Status.Conditions[i].LastTransitionTime = utils.FormatK8sTime(
				&deployment.Status.Conditions[i].LastTransitionTime)
		}
	} else {
		return nil, errors.Wrapf(_errcode.K8sInternalError,
			"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
	}

	if err != nil {
		return nil, err
	}

	deployEvents, err := s.GetK8sResourceEvents(ctx, clusterName,
		&req.GetK8sResourceEventsReq{
			Namespace: descReq.Namespace,
			Resource:  resource,
			Env:       string(envName),
		})
	if err != nil {
		return nil, err
	}
	res.Events = deployEvents.Events

	return res, nil
}

// CheckDeploymentExistance 校验Deployment是否存在
func (s *Service) CheckDeploymentExistance(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, getReq *req.GetDeploymentDetailReq) error {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return err
	}

	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		_, err = s.GetExtensionsV1Beta1DeploymentDetail(ctx, clusterName, getReq)
		return err
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		_, err = s.GetAppsV1DeploymentDetail(ctx, clusterName, getReq)
		return err
	}

	return errors.Wrapf(_errcode.K8sInternalError,
		"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
}

// CheckDeploymentsExistance 校验Deployments列表中是否有存在项
func (s *Service) CheckDeploymentsExistance(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, getReq *req.GetDeploymentsReq) (bool, error) {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return false, err
	}

	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		deploymentList, e := s.GetExtensionsV1Beta1Deployments(ctx, clusterName, getReq)
		if e != nil {
			return false, e
		}
		return len(deploymentList.Items) > 0, nil
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		deploymentList, e := s.GetAppsV1Deployments(ctx, clusterName, getReq)
		if e != nil {
			return false, e
		}
		return len(deploymentList.Items) > 0, nil
	}

	return false, errors.Wrapf(_errcode.K8sInternalError,
		"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
}

// CheckHealthyDeploymentExistance 校验Deployments列表中是否有除了 inverseVersion 外成功且有replica的存在项
func (s *Service) CheckHealthyDeploymentExistance(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, getReq *req.GetDeploymentsReq, inverseVersion string) (bool, error) {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return false, err
	}

	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		deploymentList, err := s.GetExtensionsV1Beta1Deployments(ctx, clusterName, getReq)
		if err != nil {
			return false, err
		}

		for i := range deploymentList.Items {
			if deploymentList.Items[i].ObjectMeta.Name == inverseVersion {
				continue
			}

			// Deployment仍在发生变化的情况下返回特殊错误
			progressing := false
			for _, condition := range deploymentList.Items[i].Status.Conditions {
				if condition.Type == v1beta1.DeploymentAvailable && deploymentList.Items[i].Status.ReadyReplicas > 0 {
					// 此时 condition.Status 有可能为 false, 但确实有生效的 pod
					return true, nil
				}

				if condition.Type == v1beta1.DeploymentProgressing && condition.Status == corev1.ConditionTrue {
					progressing = true
				}
			}

			if progressing {
				return false, _errcode.DeploymentInChangeExistsError
			}
		}

		return false, nil
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		deploymentList, err := s.GetAppsV1Deployments(ctx, clusterName, getReq)
		if err != nil {
			return false, err
		}

		for i := range deploymentList.Items {
			if deploymentList.Items[i].ObjectMeta.Name == inverseVersion {
				continue
			}

			// Deployment仍在发生变化的情况下返回特殊错误
			progressing := false
			for _, condition := range deploymentList.Items[i].Status.Conditions {
				if condition.Type == v1.DeploymentAvailable && deploymentList.Items[i].Status.ReadyReplicas > 0 {
					// 此时 condition.Status 有可能为 false, 但确实有生效的 pod
					return true, nil
				}

				if condition.Type == v1.DeploymentProgressing && condition.Status == corev1.ConditionTrue {
					progressing = true
				}
			}

			if progressing {
				return false, _errcode.DeploymentInChangeExistsError
			}
		}

		return false, nil
	}

	return false, errors.Wrapf(_errcode.K8sInternalError,
		"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
}

// GetDeploymentStatus 获取Deployment运行状态信息(非全部信息)
func (s *Service) GetDeploymentStatus(ctx context.Context, clusterName entity.ClusterName,
	envName entity.AppEnvName, getReq *req.GetDeploymentDetailReq) (*resp.DeploymentStatus, error) {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return nil, err
	}

	var resource runtime.Object
	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		resource, err = s.GetExtensionsV1Beta1DeploymentDetail(ctx, clusterName, getReq)
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		resource, err = s.GetAppsV1DeploymentDetail(ctx, clusterName, getReq)
	} else {
		return nil, errors.Wrapf(_errcode.K8sInternalError,
			"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
	}

	if err != nil {
		return nil, err
	}

	res := new(resp.DescribeDeploymentResp)
	err = deepcopy.Copy(resource).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}
	return &res.Status, nil
}

// RestartDeploymentAndIgnoreResponse 声明式创建/更新Deployment，忽略返回值
func (s *Service) RestartDeploymentAndIgnoreResponse(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, restartReq *req.RestartDeploymentReq) error {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return err
	}

	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		deployment, e := s.GetExtensionsV1Beta1DeploymentDetail(ctx,
			clusterName,
			&req.GetDeploymentDetailReq{
				Namespace: restartReq.Namespace,
				Name:      restartReq.Name,
				Env:       string(envName),
			})
		if e != nil {
			return e
		}

		_, err = s.RestartExtensionsV1Beta1Deployment(ctx, clusterName, deployment, string(envName))
		return err
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		deployment, e := s.GetAppsV1DeploymentDetail(ctx, clusterName,
			&req.GetDeploymentDetailReq{
				Namespace: restartReq.Namespace,
				Name:      restartReq.Name,
				Env:       string(envName),
			})
		if e != nil {
			return e
		}

		_, err = s.RestartAppsV1Deployment(ctx, clusterName, deployment, string(envName))
		return err
	}

	return errors.Wrapf(_errcode.K8sInternalError,
		"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
}

// UpdateDeploymentScaleAndIgnoreResponse 伸缩Deployment，忽略返回值
func (s *Service) UpdateDeploymentScaleAndIgnoreResponse(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, namespace, deploymentName string, scaleCount int) error {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return err
	}

	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		_, err = s.UpdateExtensionsV1Beta1DeploymentScale(ctx, clusterName, namespace, deploymentName, string(envName), scaleCount)
		return err
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		_, err = s.UpdateAppsV1DeploymentScale(ctx, clusterName, namespace, deploymentName, string(envName), scaleCount)
		return err
	}

	return errors.Wrapf(_errcode.K8sInternalError,
		"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
}

// GetDeployments 批量获取Deployment列表
func (s *Service) GetDeployments(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, getReq *req.GetDeploymentsReq) ([]resp.Deployment, error) {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return nil, err
	}

	var resource runtime.Object
	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		resource, err = s.GetExtensionsV1Beta1Deployments(ctx, clusterName, getReq)
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		resource, err = s.GetAppsV1Deployments(ctx, clusterName, getReq)
	} else {
		return nil, errors.Wrapf(_errcode.K8sInternalError,
			"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
	}

	if err != nil {
		return nil, err
	}

	deploymentList := new(resp.DeploymentList)
	err = deepcopy.Copy(resource).To(deploymentList)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}
	return deploymentList.Items, nil
}

// GetDeploymentDetail 获取Deployment详情
func (s *Service) GetDeploymentDetail(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, getReq *req.GetDeploymentDetailReq) (*resp.Deployment, error) {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return nil, err
	}

	var resource runtime.Object
	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		resource, err = s.GetExtensionsV1Beta1DeploymentDetail(ctx, clusterName, getReq)
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		resource, err = s.GetAppsV1DeploymentDetail(ctx, clusterName, getReq)
	} else {
		return nil, errors.Wrapf(_errcode.K8sInternalError,
			"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
	}

	if err != nil {
		return nil, err
	}

	res := new(resp.Deployment)
	err = deepcopy.Copy(resource).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// DeleteDeployments 批量删除Deployment
func (s *Service) DeleteDeployments(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, deleteReq *req.DeleteDeploymentsReq) error {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return err
	}

	var resource runtime.Object
	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		resource, err = s.GetExtensionsV1Beta1Deployments(ctx, clusterName, &req.GetDeploymentsReq{
			Namespace:   deleteReq.Namespace,
			ProjectName: deleteReq.ProjectName,
			AppName:     deleteReq.AppName,
			Env:         deleteReq.Env,
		})
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		resource, err = s.GetAppsV1Deployments(ctx, clusterName, &req.GetDeploymentsReq{
			Namespace:   deleteReq.Namespace,
			ProjectName: deleteReq.ProjectName,
			AppName:     deleteReq.AppName,
			Env:         deleteReq.Env,
		})
	} else {
		return errors.Wrapf(_errcode.K8sInternalError,
			"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
	}

	if err != nil {
		return err
	}

	deploymentList := new(resp.DeploymentList)
	err = deepcopy.Copy(resource).To(deploymentList)
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	for i := range deploymentList.Items {
		if deleteReq.InverseVersion != "" && deploymentList.Items[i].GetName() == deleteReq.InverseVersion {
			continue
		}

		err = s.DeleteDeployment(ctx, clusterName, envName, &req.DeleteDeploymentReq{
			Namespace: deploymentList.Items[i].GetNamespace(),
			Name:      deploymentList.Items[i].GetName(),
			Policy:    deleteReq.Policy,
			Env:       deleteReq.Env,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteDeployment 删除Deployment
func (s *Service) DeleteDeployment(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, deleteReq *req.DeleteDeploymentReq) error {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return err
	}

	if deleteReq.Policy == "" {
		deleteReq.Policy = metav1.DeletePropagationForeground
	}

	c, err := s.GetK8sTypedClient(clusterName, deleteReq.Env)
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		err = c.ExtensionsV1beta1().Deployments(deleteReq.Namespace).
			Delete(ctx, deleteReq.Name, metav1.DeleteOptions{
				PropagationPolicy: &deleteReq.Policy,
			})
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		err = c.AppsV1().Deployments(deleteReq.Namespace).
			Delete(ctx, deleteReq.Name, metav1.DeleteOptions{
				PropagationPolicy: &deleteReq.Policy,
			})
	} else {
		return errors.Wrapf(_errcode.K8sInternalError,
			"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
	}

	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return nil
}

// ApplyExtensionsV1Beta1Deployment 声明式创建/更新extensions/v1beta1版本的Deployment
// 该版本的Deployment在v1.18取消支持，详见release note
func (s *Service) ApplyExtensionsV1Beta1Deployment(ctx context.Context,
	clusterName entity.ClusterName, yamlData []byte, env string) (*v1beta1.Deployment, error) {
	applyDeployment, err := s.decodeExtensionsV1Beta1DeploymentYamlData(ctx, yamlData)
	if err != nil {
		return nil, err
	}

	_, err = s.GetExtensionsV1Beta1DeploymentDetail(ctx, clusterName,
		&req.GetDeploymentDetailReq{
			Namespace: applyDeployment.GetNamespace(),
			Name:      applyDeployment.GetName(),
			Env:       env,
		})
	if err != nil {
		if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil, err
		}

		deployment, e := s.CreateExtensionsV1Beta1Deployment(ctx, clusterName, applyDeployment, env)
		if e != nil {
			return nil, e
		}
		return deployment, nil
	}

	deployment, err := s.PatchExtensionsV1Beta1Deployment(ctx, clusterName, applyDeployment, env)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

// ApplyAppsV1Deployment 声明式创建/更新apps/v1版本的Deployment
func (s *Service) ApplyAppsV1Deployment(ctx context.Context,
	clusterName entity.ClusterName, yamlData []byte, env string) (*v1.Deployment, error) {
	applyDeployment, err := s.decodeAppsV1DeploymentYamlData(ctx, yamlData)
	if err != nil {
		return nil, err
	}

	_, err = s.GetAppsV1DeploymentDetail(ctx, clusterName,
		&req.GetDeploymentDetailReq{
			Namespace: applyDeployment.GetNamespace(),
			Name:      applyDeployment.GetName(),
			Env:       env,
		})
	if err != nil {
		if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil, err
		}

		deployment, e := s.CreateAppsV1Deployment(ctx, clusterName, applyDeployment, env)
		if e != nil {
			return nil, e
		}
		return deployment, nil
	}

	deployment, err := s.PatchAppsV1Deployment(ctx, clusterName, applyDeployment, yamlData, env)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

// CreateExtensionsV1Beta1Deployment 创建extensions/v1beta1版本的Deployment
// 该版本的Deployment在v1.18取消支持，详见release note
func (s *Service) CreateExtensionsV1Beta1Deployment(ctx context.Context, clusterName entity.ClusterName,
	deployment *v1beta1.Deployment, env string) (*v1beta1.Deployment, error) {
	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.ExtensionsV1beta1().Deployments(deployment.GetNamespace()).
		Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// CreateAppsV1Deployment 创建apps/v1版本的Deployment
func (s *Service) CreateAppsV1Deployment(ctx context.Context, clusterName entity.ClusterName,
	deployment *v1.Deployment, env string) (*v1.Deployment, error) {
	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.AppsV1().Deployments(deployment.GetNamespace()).
		Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// GetExtensionsV1Beta1DeploymentDetail 获取extensions/v1beta1版本的Deployment详情
// 该版本的Deployment在v1.18取消支持，详见release note
func (s *Service) GetExtensionsV1Beta1DeploymentDetail(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetDeploymentDetailReq) (*v1beta1.Deployment, error) {
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.ExtensionsV1beta1().Deployments(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// GetAppsV1DeploymentDetail 获取apps/v1版本的Deployment详情
func (s *Service) GetAppsV1DeploymentDetail(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetDeploymentDetailReq) (*v1.Deployment, error) {
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.AppsV1().Deployments(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}

		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

func (s *Service) getDeploymentsLabelSelector(getReq *req.GetDeploymentsReq) string {
	labels := make([]string, 0)
	if getReq.ProjectName != "" {
		labels = append(labels, fmt.Sprintf("project=%s", getReq.ProjectName))
	}

	if getReq.AppName != "" {
		labels = append(labels, fmt.Sprintf("app=%s", getReq.AppName))
	}

	if getReq.Version != "" {
		labels = append(labels, fmt.Sprintf("version=%s", getReq.Version))
	}
	return strings.Join(labels, ",")
}

// GetExtensionsV1Beta1Deployments 批量获取extensions/v1beta1版本的Deployment列表
// 该版本的Deployment在v1.18取消支持，详见release note
func (s *Service) GetExtensionsV1Beta1Deployments(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetDeploymentsReq) (*v1beta1.DeploymentList, error) {
	selector := s.getDeploymentsLabelSelector(getReq)

	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	list, err := c.ExtensionsV1beta1().Deployments(getReq.Namespace).
		List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	// 时间倒序
	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].GetCreationTimestamp().After(
			list.Items[j].GetCreationTimestamp().Time)
	})
	return list, nil
}

// GetAppsV1Deployments 批量获取apps/v1版本的Deployment列表
func (s *Service) GetAppsV1Deployments(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetDeploymentsReq) (*v1.DeploymentList, error) {
	selector := s.getDeploymentsLabelSelector(getReq)

	if getReq.Env == "" {
		getReq.Env = getReq.Namespace
	}

	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	list, err := c.AppsV1().Deployments(getReq.Namespace).
		List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	// 时间倒序
	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].GetCreationTimestamp().After(
			list.Items[j].GetCreationTimestamp().Time)
	})
	return list, nil
}

// PatchExtensionsV1Beta1Deployment 更新extensions/v1beta1版本的Deployment
// 该版本的Deployment在v1.18取消支持，详见release note
func (s *Service) PatchExtensionsV1Beta1Deployment(ctx context.Context, clusterName entity.ClusterName,
	deployment *v1beta1.Deployment, env string) (*v1beta1.Deployment, error) {
	patchData, err := json.Marshal(deployment)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.ExtensionsV1beta1().Deployments(deployment.GetNamespace()).
		Patch(ctx, deployment.GetName(), types.MergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// PatchAppsV1Deployment 更新apps/v1版本的Deployment
func (s *Service) PatchAppsV1Deployment(ctx context.Context, clusterName entity.ClusterName,
	deployment *v1.Deployment, yamlData []byte, env string) (*v1.Deployment, error) {
	var patchData []byte
	if yamlData != nil {
		var jsonObj map[string]interface{}
		err := yaml.Unmarshal(yamlData, &jsonObj)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}

		jsonData, err := json.Marshal(jsonObj)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}

		patchData = jsonData
	} else {
		jsonData, err := json.Marshal(deployment)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}

		patchData = jsonData
	}

	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.AppsV1().Deployments(deployment.GetNamespace()).
		Patch(ctx, deployment.GetName(), types.MergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// RestartExtensionsV1Beta1Deployment 重启extensions/v1beta1版本的Deployment
// 该版本的Deployment在v1.18取消支持，详见release note
func (s *Service) RestartExtensionsV1Beta1Deployment(ctx context.Context, clusterName entity.ClusterName,
	deployment *v1beta1.Deployment, env string) (*v1beta1.Deployment, error) {
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	deployment.Spec.Template.ObjectMeta.Annotations[K8sAnnotationRestart] =
		time.Now().Format(RestartAnnotationLayout)

	return s.PatchExtensionsV1Beta1Deployment(ctx, clusterName, deployment, env)
}

// RestartAppsV1Deployment 重启apps/v1版本的Deployment
func (s *Service) RestartAppsV1Deployment(ctx context.Context, clusterName entity.ClusterName,
	deployment *v1.Deployment, env string) (*v1.Deployment, error) {
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	deployment.Spec.Template.ObjectMeta.Annotations[K8sAnnotationRestart] =
		time.Now().Format(RestartAnnotationLayout)

	return s.PatchAppsV1Deployment(ctx, clusterName, deployment, nil, env)
}

// UpdateExtensionsV1Beta1DeploymentScale 伸缩extensions/v1beta1版本的Deployment
// 该版本的Deployment在v1.18取消支持，详见release note
func (s *Service) UpdateExtensionsV1Beta1DeploymentScale(ctx context.Context, clusterName entity.ClusterName,
	namespace, deploymentName, envName string, scaleCount int) (*v1beta1.Scale, error) {
	scale := &v1beta1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		Spec: v1beta1.ScaleSpec{
			Replicas: int32(scaleCount),
		},
	}

	c, err := s.GetK8sTypedClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.ExtensionsV1beta1().Deployments(namespace).
		UpdateScale(ctx, deploymentName, scale, metav1.UpdateOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// UpdateAppsV1DeploymentScale 伸缩apps/v1版本的Deployment
func (s *Service) UpdateAppsV1DeploymentScale(ctx context.Context, clusterName entity.ClusterName,
	namespace, deploymentName, envName string, scaleCount int) (*autoscalingv1.Scale, error) {
	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: int32(scaleCount),
		},
	}

	c, err := s.GetK8sTypedClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.AppsV1().Deployments(namespace).
		UpdateScale(ctx, deploymentName, scale, metav1.UpdateOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

func (s *Service) getLabels(project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) map[string]string {
	labelsMap := make(map[string]string)
	labelsMap["team"] = team.Label
	labelsMap["app-id"] = app.ID
	labelsMap["metrics"] = entity.DeploymentMetricsLabelDisable
	if task.Param.IsSupportMetrics {
		labelsMap["metrics"] = entity.DeploymentMetricsLabelEnable
	}
	labelsMap["appType"] = string(app.Type)
	if app.Type == entity.AppTypeService {
		labelsMap["appServiceType"] = string(app.ServiceType)

		if app.EnableIstio {
			labelsMap[LabelCanonicalName] = app.ServiceName + "-" + task.Namespace
		}
	}

	for _, label := range project.Labels {
		labelsMap[s.getMetricsLabelKeyName(label)] = MetricsLabelEnable
	}

	return labelsMap
}

func (s *Service) getMetricsLabelKeyName(label string) string {
	return fmt.Sprintf("shanhai.int/project-label-%s", label)
}

// ReloadDeployment updates deployment's ConfigMap hash data
func (s *Service) ReloadDeployment(ctx context.Context, task *resp.TaskDetailResp, cmHash string) error {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, task.ClusterName, task.EnvName, entity.K8sObjectKindDeployment)
	if err != nil {
		return err
	}

	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		deployment, err := s.GetExtensionsV1Beta1DeploymentDetail(ctx, task.ClusterName, &req.GetDeploymentDetailReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return err
		}

		_, err = s.ReloadExtensionsV1Beta1Deployment(ctx, task.ClusterName, cmHash, deployment, string(task.EnvName))
		return err
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		deployment, err := s.GetAppsV1DeploymentDetail(ctx, task.ClusterName, &req.GetDeploymentDetailReq{
			Namespace: task.Namespace,
			Name:      task.Version,
			Env:       string(task.EnvName),
		})
		if err != nil {
			return err
		}

		_, err = s.ReloadAppsV1Deployment(ctx, task.ClusterName, cmHash, deployment, string(task.EnvName))
		return err
	}

	return errors.Wrapf(_errcode.K8sInternalError,
		"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
}

// ReloadExtensionsV1Beta1Deployment update extensionsV1Beta1 deployment annotation & labels of cm hash
func (s *Service) ReloadExtensionsV1Beta1Deployment(ctx context.Context, clusterName entity.ClusterName,
	cmHash string, deployment *v1beta1.Deployment, env string) (*v1beta1.Deployment, error) {
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.GetLabels()[LabelConfigHash] = cmHash
	deployment.Spec.Template.ObjectMeta.Annotations[AnnotationConfigHash] = cmHash
	deployment.Spec.Template.ObjectMeta.Labels[LabelConfigHash] = cmHash

	return s.PatchExtensionsV1Beta1Deployment(ctx, clusterName, deployment, env)
}

// ReloadAppsV1Deployment update appV1 deployment annotation & labels of cm hash
func (s *Service) ReloadAppsV1Deployment(ctx context.Context,
	clusterName entity.ClusterName, cmHash string, deployment *v1.Deployment, env string) (*v1.Deployment, error) {
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.GetLabels()[LabelConfigHash] = cmHash
	deployment.Spec.Template.ObjectMeta.Annotations[AnnotationConfigHash] = cmHash
	deployment.Spec.Template.ObjectMeta.Labels[LabelConfigHash] = cmHash

	return s.PatchAppsV1Deployment(ctx, clusterName, deployment, nil, env)
}

// EnableDeploymentHPAAndIgnoreResponse 启用Deployment HPA，忽略返回值
func (s *Service) EnableDeploymentHPAAndIgnoreResponse(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, enableReq *req.EnableDeploymentHPAReq) error {
	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, envName, entity.K8sObjectKindDeployment)
	if err != nil {
		return err
	}

	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		deployment, e := s.GetExtensionsV1Beta1DeploymentDetail(ctx,
			clusterName,
			&req.GetDeploymentDetailReq{
				Namespace: enableReq.Namespace,
				Name:      enableReq.Name,
				Env:       string(envName),
			})
		if e != nil {
			return e
		}

		_, err = s.EnableExtensionsV1Beta1DeploymentHPA(ctx, clusterName, deployment, string(envName))
		return err
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		deployment, e := s.GetAppsV1DeploymentDetail(ctx, clusterName,
			&req.GetDeploymentDetailReq{
				Namespace: enableReq.Namespace,
				Name:      enableReq.Name,
				Env:       string(envName),
			})
		if e != nil {
			return e
		}

		_, err = s.EnableAppsV1DeploymentHPA(ctx, clusterName, deployment, string(envName))
		return err
	}

	return errors.Wrapf(_errcode.K8sInternalError,
		"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindDeployment)
}

// EnableExtensionsV1Beta1DeploymentHPA 启用extensions/v1beta1版本的Deployment HPA
// 该版本的Deployment在v1.18取消支持，详见release note
func (s *Service) EnableExtensionsV1Beta1DeploymentHPA(ctx context.Context, clusterName entity.ClusterName,
	deployment *v1beta1.Deployment, env string) (*v1beta1.Deployment, error) {
	if deployment.Spec.Template.ObjectMeta.Annotations != nil {
		annotations := deployment.ObjectMeta.Annotations
		delete(annotations, K8sAnnotationHPASkipped)
		deployment.Annotations = annotations
	}

	return s.PatchExtensionsV1Beta1Deployment(ctx, clusterName, deployment, env)
}

// EnableAppsV1DeploymentHPA 启用apps/v1版本的Deployment HPA
func (s *Service) EnableAppsV1DeploymentHPA(ctx context.Context, clusterName entity.ClusterName,
	deployment *v1.Deployment, env string) (*v1.Deployment, error) {
	if deployment.Spec.Template.ObjectMeta.Annotations != nil {
		annotations := deployment.ObjectMeta.Annotations
		delete(annotations, K8sAnnotationHPASkipped)
		deployment.Annotations = annotations
	}

	return s.PatchAppsV1Deployment(ctx, clusterName, deployment, nil, env)
}
