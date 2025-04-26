package service

import (
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gopkg.in/yaml.v3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var (
	// 阿里云日志采集器crd资源
	aliyunLogConfigResource = schema.GroupVersionResource{
		Group:    "log.alibabacloud.com",
		Version:  "v1alpha1",
		Resource: "aliyunlogconfigs",
	}
)

// initAliLogConfigTemplate 初始化阿里云日志采集器渲染模版
func (s *Service) initAliLogConfigTemplate(_ context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) (*entity.AliyunLogConfigTemplate, error) {
	if _, ok := app.Env[task.EnvName]; !ok {
		return nil, errors.Wrapf(errcode.InvalidParams, "project: %v app: %v env: %v", project.Name, app.Name, task.EnvName)
	}

	// 模板结构体构建
	template := &entity.AliyunLogConfigTemplate{
		Namespace:       task.Namespace,
		Name:            app.AliLogConfigName,
		ClusterName:     task.ClusterName,
		ProjectName:     project.Name,
		AppName:         app.Name,
		TeamLabel:       team.Label,
		LogStoreName:    project.LogStoreName,
		LogTailName:     app.Env[task.EnvName].LogTailName,
		ContainerName:   utils.GetPodContainerName(project.Name, app.Name),
		ExcludeLabel:    s.getAliyunLogConfigExcludeLabel(string(task.EnvName)),
		OpenColdStorage: strconv.FormatBool(task.Param.OpenColdStorage),
	}

	// 检测仓库日志名
	if template.LogStoreName == "" {
		return nil, errors.Wrapf(errcode.InvalidParams, "project: %v app: %v StoreName is empty", project.Name, app.Name)
	}

	return template, nil
}

// decodeAliLogConfigYamlData 解析阿里云日志采集器的 yaml 配置
func (s *Service) decodeAliLogConfigYamlData(_ context.Context, yamlData []byte) (*unstructured.Unstructured, error) {
	aliLogConfig := new(unstructured.Unstructured)

	err := yaml.Unmarshal(yamlData, &aliLogConfig.Object)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return aliLogConfig, nil
}

// RenderAliLogConfigTemplate 渲染阿里云日志采集器模板文件
func (s *Service) RenderAliLogConfigTemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) ([]byte, error) {
	tpl, err := s.initAliLogConfigTemplate(ctx, project, app, task, team)
	if err != nil {
		return nil, err
	}

	data, err := s.RenderK8sTemplate(ctx, entity.DefaultTemplateFileDir, task.ClusterName, task.EnvName, tpl)
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

// ApplyAliLogConfig 声明式创建/更新阿里云日志采集器
func (s *Service) ApplyAliLogConfig(ctx context.Context,
	clusterName entity.ClusterName, envName string, yamlData []byte) (*unstructured.Unstructured, error) {
	applyAliLogConfig, err := s.decodeAliLogConfigYamlData(ctx, yamlData)
	if err != nil {
		return nil, err
	}

	_, err = s.GetAliLogConfigDetail(ctx, clusterName, &req.GetAliLogConfigDetailReq{
		Namespace: applyAliLogConfig.GetNamespace(),
		Name:      applyAliLogConfig.GetName(),
		Env:       envName,
	})
	if err != nil {
		if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil, err
		}

		return s.CreateAliLogConfig(ctx, clusterName, envName, applyAliLogConfig)
	}

	return s.PatchAliLogConfig(ctx, clusterName, envName, applyAliLogConfig)
}

// CreateAliLogConfig 创建阿里云日志采集器
func (s *Service) CreateAliLogConfig(ctx context.Context,
	clusterName entity.ClusterName, envName string, aliLogConfig *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	c, err := s.GetK8sDynamicClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.Resource(aliyunLogConfigResource).
		Namespace(aliLogConfig.GetNamespace()).
		Create(ctx, aliLogConfig, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// GetAliLogConfigDetail 获取阿里云日志采集器详情
func (s *Service) GetAliLogConfigDetail(ctx context.Context,
	clusterName entity.ClusterName, getReq *req.GetAliLogConfigDetailReq) (*unstructured.Unstructured, error) {
	c, err := s.GetK8sDynamicClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.Resource(aliyunLogConfigResource).
		Namespace(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}

		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// PatchAliLogConfig 更新阿里云日志采集器
func (s *Service) PatchAliLogConfig(ctx context.Context,
	clusterName entity.ClusterName, envName string, aliLogConfig *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	patchData, err := aliLogConfig.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	c, err := s.GetK8sDynamicClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.Resource(aliyunLogConfigResource).
		Namespace(aliLogConfig.GetNamespace()).
		Patch(ctx, aliLogConfig.GetName(), types.MergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}

		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// DeleteAliLogConfig 删除阿里云日志采集器
func (s *Service) DeleteAliLogConfig(ctx context.Context, clusterName entity.ClusterName, envName string,
	deleteReq *req.DeleteAliLogConfigReq) error {
	policy := metav1.DeletePropagationForeground

	c, err := s.GetK8sDynamicClient(clusterName, envName)
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	err = c.Resource(aliyunLogConfigResource).
		Namespace(deleteReq.Namespace).
		Delete(ctx, deleteReq.Name, metav1.DeleteOptions{
			PropagationPolicy: &policy,
		})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}

		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return nil
}

// GetLogTailName 获取logtail名称
func (s *Service) GetLogTailName(projectName, appName, envName string) string {
	return fmt.Sprintf("%s-%s-%s", projectName, appName, envName)
}

// getAliyunLogConfigExcludeLabel 获取过滤环境标签
func (s *Service) getAliyunLogConfigExcludeLabel(envName string) string {
	var patterns []string
	for i := range envName {
		patterns = append(patterns, fmt.Sprintf("%s[^%s]", envName[:i], envName[i:i+1]))
	}
	return fmt.Sprintf("^%s$", strings.Join(patterns, "|"))
}

// 获取阿里云日志采集器 k8s 选择器标签(labelSelector)
func (s *Service) getAliLogConfigLabelSelector(getReq *req.ListAliLogConfigReq) string {
	labels := make([]string, 0)
	if getReq.ProjectName != "" {
		labels = append(labels, fmt.Sprintf("project=%s", getReq.ProjectName))
	}
	if !getReq.OpenColdStorage.IsZero() {
		labels = append(labels, fmt.Sprintf("coldStorage=%s", strconv.FormatBool(getReq.OpenColdStorage.ValueOrZero())))
	}

	return strings.Join(labels, ",")
}

// ListAliLogConfig 按照 k8s 选择器标签(labelSelector) 获取阿里云日志采集器列表
func (s *Service) ListAliLogConfig(ctx context.Context, clusterName entity.ClusterName, envName string,
	listReq *req.ListAliLogConfigReq) (*unstructured.UnstructuredList, error) {
	labelSelector := s.getAliLogConfigLabelSelector(listReq)
	c, err := s.GetK8sDynamicClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	list, err := c.Resource(aliyunLogConfigResource).
		Namespace(listReq.Namespace).
		List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return list, nil
}
