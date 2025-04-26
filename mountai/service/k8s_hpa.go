package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	v2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"

	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// 默认值
const (
	// DefaultHPACpuTarget 默认 HPA CPU 控制目标值
	DefaultHPACpuTarget int32 = 80
	// DefaultHPAMemTarget 默认 HPA 内存控制目标值
	DefaultHPAMemTarget int32 = 80
)

var (
	// HPAWaitMetricsSyncLabelLevel 需要等待hpa度量数据的项目的等级
	HPAWaitMetricsSyncLabelLevel = entity.ProjectLabelP0
	// HPAWaitMetricsSyncEnvName 需要等待hap度量数据的项目的环境
	HPAWaitMetricsSyncEnvName = entity.AppEnvPrd
)

func (s *Service) initHPATemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, _ *resp.TeamDetailResp) (*entity.HPATemplate, error) {
	_, ok := app.Env[task.EnvName]
	if !ok {
		return nil, errors.Wrapf(errcode.InternalError, "couldn't find env:%s", task.EnvName)
	}

	// 当前只有 Deployment 需要 HPA
	scaleTargetRefKind := entity.K8sObjectKindDeployment
	scaleTargetRefGroupVersion, err := s.getK8sResourceGroupVersion(ctx, task.ClusterName, task.EnvName, scaleTargetRefKind)
	if err != nil {
		return nil, errors.Wrapf(errcode.InternalError, "get apiVersion for %s(cluster=%s, env=%s) failed: %s",
			scaleTargetRefKind, task.ClusterName, task.EnvName, err)
	}

	template := &entity.HPATemplate{
		Name:        task.Version,
		Namespace:   task.Namespace,
		ProjectName: project.Name,
		AppName:     app.Name,
		ScaleTargetRef: entity.ScaleTargetRefTemplate{
			Kind:       scaleTargetRefKind,
			Name:       task.Version,
			APIVersion: scaleTargetRefGroupVersion.String(),
		},
		MinReplicas: int32(task.Param.MinPodCount),
		MaxReplicas: int32(task.Param.MaxPodCount),
		CPUTarget:   DefaultHPACpuTarget,
		MemTarget:   DefaultHPAMemTarget,
	}

	return template, nil
}

func (s *Service) decodeHPAYamlData(_ context.Context, yamlData []byte) (*v2.HorizontalPodAutoscaler, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(yamlData, nil, nil)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}
	hpa, ok := obj.(*v2.HorizontalPodAutoscaler)
	if !ok {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "object is not v2.HorizontalPodAutoscaler")
	}
	return hpa, nil
}

// RenderHPATemplate 渲染 HPA 模板
func (s *Service) RenderHPATemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) ([]byte, error) {
	tpl, err := s.initHPATemplate(ctx, project, app, task, team)
	if err != nil {
		return nil, err
	}

	data, err := s.RenderK8sTemplate(ctx, entity.DefaultTemplateFileDir, task.ClusterName, task.EnvName, tpl)
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

// ApplyHPA 声明式创建/更新 HPA
func (s *Service) ApplyHPA(ctx context.Context, clusterName entity.ClusterName,
	yamlData []byte, env string) (*v2.HorizontalPodAutoscaler, error) {
	applyHPA, err := s.decodeHPAYamlData(ctx, yamlData)
	if err != nil {
		return nil, err
	}

	_, err = s.GetHPADetail(ctx, clusterName,
		&req.GetHPADetailReq{
			Namespace: applyHPA.GetNamespace(),
			Name:      applyHPA.GetName(),
			Env:       env,
		})
	if err != nil {
		if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil, err
		}

		hpa, e := s.CreateHPA(ctx, clusterName, applyHPA, env)
		if e != nil {
			return nil, e
		}
		return hpa, nil
	}

	hpa, err := s.PatchHPA(ctx, clusterName, applyHPA, env)
	if err != nil {
		return nil, err
	}
	return hpa, nil
}

// CreateHPA 创建 HPA
func (s *Service) CreateHPA(ctx context.Context, clusterName entity.ClusterName,
	hpa *v2.HorizontalPodAutoscaler, env string) (*v2.HorizontalPodAutoscaler, error) {
	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.AutoscalingV2().
		HorizontalPodAutoscalers(hpa.GetNamespace()).
		Create(ctx, hpa, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// GetHPADetail 获取 HPA 详情
func (s *Service) GetHPADetail(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetHPADetailReq) (*v2.HorizontalPodAutoscaler, error) {
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.AutoscalingV2().
		HorizontalPodAutoscalers(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// PatchHPA 更新 HPA
func (s *Service) PatchHPA(ctx context.Context, clusterName entity.ClusterName,
	hpa *v2.HorizontalPodAutoscaler, env string) (*v2.HorizontalPodAutoscaler, error) {
	patchData, err := json.Marshal(hpa)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.AutoscalingV2().
		HorizontalPodAutoscalers(hpa.GetNamespace()).
		Patch(ctx, hpa.GetName(), types.MergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

func (s *Service) getHPAsLabelSelector(getReq *req.GetHPAsReq) string {
	labels := make([]string, 0)
	if getReq.ProjectName != "" {
		labels = append(labels, fmt.Sprintf("project=%s", getReq.ProjectName))
	}
	if getReq.Version != "" {
		labels = append(labels, fmt.Sprintf("version=%s", getReq.Version))
	}
	if getReq.AppName != "" {
		labels = append(labels, fmt.Sprintf("app=%s", getReq.AppName))
	}
	return strings.Join(labels, ",")
}

// GetHPAs 获取 HPA 列表
func (s *Service) GetHPAs(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetHPAsReq) ([]v2.HorizontalPodAutoscaler, error) {
	selector := s.getHPAsLabelSelector(getReq)

	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	list, err := c.AutoscalingV2().
		HorizontalPodAutoscalers("").
		List(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return list.Items, nil
}

// DeleteHPAs 批量删除 HPA
func (s *Service) DeleteHPAs(ctx context.Context, clusterName entity.ClusterName,
	deleteReq *req.DeleteHPAsReq) error {
	res, err := s.GetHPAs(ctx, clusterName,
		&req.GetHPAsReq{
			Namespace:   deleteReq.Namespace,
			ProjectName: deleteReq.ProjectName,
			AppName:     deleteReq.AppName,
			Env:         deleteReq.Env,
		})
	if err != nil {
		return err
	}

	for i := range res {
		hpa := res[i]

		if deleteReq.InverseVersion != "" && hpa.GetName() == deleteReq.InverseVersion {
			continue
		}

		err = s.DeleteHPA(ctx, clusterName,
			&req.DeleteHPAReq{
				Namespace: hpa.GetNamespace(),
				Name:      hpa.GetName(),
				Env:       deleteReq.Env,
			})
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteHPA 删除 HPA
func (s *Service) DeleteHPA(ctx context.Context,
	clusterName entity.ClusterName, deleteReq *req.DeleteHPAReq) error {
	policy := metav1.DeletePropagationForeground
	c, err := s.GetK8sTypedClient(clusterName, deleteReq.Env)
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	err = c.AutoscalingV2().
		HorizontalPodAutoscalers(deleteReq.Namespace).
		Delete(ctx, deleteReq.Name, metav1.DeleteOptions{
			PropagationPolicy: &policy,
		})
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return nil
}

// DescribeHPA 获取 HPA describe 信息
func (s *Service) DescribeHPA(ctx context.Context, clusterName entity.ClusterName,
	descReq *req.DescribeHPAReq) (*resp.DescribeHPAResp, error) {
	res := new(resp.DescribeHPAResp)
	hpa, err := s.GetHPADetail(ctx, clusterName,
		&req.GetHPADetailReq{
			Name:      descReq.Name,
			Namespace: descReq.Namespace,
			Env:       descReq.Env,
		})
	if err != nil {
		return nil, err
	}

	err = deepcopy.Copy(hpa).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	res.Status.LastScaleTime = utils.FormatK8sTime(hpa.Status.LastScaleTime)

	conditions := make([]resp.HorizontalPodAutoscalerCondition, len(hpa.Status.Conditions))
	for i, condition := range res.Status.Conditions {
		condition.LastTransitionTime = utils.FormatK8sTime(&hpa.Status.Conditions[i].LastTransitionTime)
		conditions[i] = condition
	}
	res.Status.Conditions = conditions

	hpaEvents, err := s.GetK8sResourceEvents(ctx, clusterName,
		&req.GetK8sResourceEventsReq{
			Namespace: descReq.Namespace,
			Resource:  hpa,
			Env:       descReq.Env,
		})
	if err != nil {
		return nil, err
	}
	res.Events = hpaEvents.Events

	return res, nil
}

// GetHPAReadyTime 获取hpa初始化完毕时间
func (s *Service) GetHPAReadyTime(ctx context.Context,
	clusterName entity.ClusterName, getReq *req.GetHPADetailReq) (time.Time, error) {
	hpa, err := s.GetHPADetail(ctx, clusterName, getReq)
	if err != nil {
		return time.Time{}, err
	}

	// 需判断hpa的 ScalingActive 已经激活时才认为初始化完毕
	for _, condition := range hpa.Status.Conditions {
		if condition.Type != v2.ScalingActive || condition.Status != v1.ConditionTrue {
			continue
		}

		return condition.LastTransitionTime.Time, nil
	}

	return time.Time{}, errors.Wrapf(_errcode.K8sResourceNotReadyError, "hpa not ready: %s", hpa.GetName())
}

func needWaitMetricsSync(labels []string, env entity.AppEnvName) bool {
	needSyncLabel := false
	for _, pl := range labels {
		if pl == string(HPAWaitMetricsSyncLabelLevel) {
			needSyncLabel = true
			break
		}
	}

	return needSyncLabel && (env == HPAWaitMetricsSyncEnvName)
}
