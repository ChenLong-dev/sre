package service

import (
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	// K8s ConfigMap label key "commit"
	ConfigMapLabelCommit = "commit"
	// K8s ConfigMap label key "hash"
	ConfigMapLabelHash = "hash"
)

// GetAppOldConfigMapName 获取应用 ConfigMap 名称
func (s *Service) GetAppOldConfigMapName(projectName, appName string) string {
	return fmt.Sprintf("%s-%s", projectName, appName)
}

// GetAppNewConfigMapName returns new configMap name
func (s *Service) GetAppNewConfigMapName(task *resp.TaskDetailResp) string {
	return task.Version
}

func (s *Service) initAppConfigMapTemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	team *resp.TeamDetailResp, task *resp.TaskDetailResp, rawData interface{}) (*entity.AppConfigMapTemplate, error) {
	multiConfigMap, err := s.GetConfigMapDataFromAppConfig(task, rawData)
	if err != nil {
		return nil, err
	}
	// 序列化
	configMapData := map[string]interface{}{
		"data": multiConfigMap,
	}

	configData, err := yaml.Marshal(configMapData)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	cmHash, err := s.GetConfigMapHash(ctx, multiConfigMap)
	if err != nil {
		return nil, err
	}

	name := s.GetAppNewConfigMapName(task)
	// We just only update ConfigMap when task action is `reload_config`,
	// so we get the compatible ConfigMap name and cover `name`.
	if task.Action == entity.TaskActionReloadConfig {
		cm, err := s.GetCompatibleConfigMapDetail(ctx, project, app, task)
		if err != nil {
			return nil, err
		}
		name = cm.Name
	}

	template := &entity.AppConfigMapTemplate{
		ProjectName: project.Name,
		AppName:     app.Name,
		Namespace:   task.Namespace,
		TeamLabel:   team.Label,
		Labels:      s.getCMLabels(task, cmHash),
		Name:        name,
		Data:        string(configData),
	}

	return template, nil
}

func (s *Service) decodeConfigMapYamlData(_ context.Context, yamlData []byte) (*v1.ConfigMap, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(yamlData, nil, nil)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "object is not v1.ConfigMap")
	}

	return cm, nil
}

// RenderConfigMapTemplate 渲染 ConfigMap 模板
func (s *Service) RenderConfigMapTemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp, rawData interface{}) ([]byte, error) {
	tpl, err := s.initAppConfigMapTemplate(ctx, project, app, team, task, rawData)
	if err != nil {
		return nil, err
	}

	data, err := s.RenderK8sTemplate(ctx, entity.DefaultTemplateFileDir, task.ClusterName, task.EnvName, tpl)
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

func (s *Service) GetConfigMapDataFromAppConfig(task *resp.TaskDetailResp, rawData interface{}) (map[string]string, error) {
	rawMap, ok := rawData.(map[string]interface{})
	if !ok {
		return nil, errors.Wrap(errcode.InternalError, "config raw data is not map")
	}

	// 多文件解析
	multiConfigMap := make(map[string]string)
	for key, value := range rawMap {
		if reflect.ValueOf(value).Kind() == reflect.String {
			multiConfigMap[key] = value.(string)
		} else {
			yamlValue, err := yaml.Marshal(value)
			if err != nil {
				return nil, errors.Wrap(errcode.InternalError, err.Error())
			}
			multiConfigMap[key] = string(yamlValue)
		}

		// 兼容老版本以环境变量命名的配置文件
		if key == "config.yaml" {
			multiConfigMap[fmt.Sprintf("%s.yaml", task.EnvName)] = multiConfigMap[key]
		}
	}

	return multiConfigMap, nil
}

// ApplyConfigMap 声明式创建/更新 ConfigMap
func (s *Service) ApplyConfigMap(ctx context.Context,
	clusterName entity.ClusterName, yamlData []byte, env string) (*v1.ConfigMap, error) {
	applyCM, err := s.decodeConfigMapYamlData(ctx, yamlData)
	if err != nil {
		return nil, err
	}

	_, err = s.GetConfigMapDetail(ctx, clusterName,
		&req.GetConfigMapDetailReq{
			Namespace: applyCM.GetNamespace(),
			Name:      applyCM.GetName(),
			Env:       env,
		})
	if err != nil {
		if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil, err
		}

		cm, e := s.CreateConfigMap(ctx, clusterName, applyCM, env)
		if e != nil {
			return nil, e
		}

		return cm, nil
	}
	cm, err := s.PatchConfigMap(ctx, clusterName, applyCM, env)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

// CreateConfigMap 创建 ConfigMap
func (s *Service) CreateConfigMap(ctx context.Context,
	clusterName entity.ClusterName, cm *v1.ConfigMap, env string) (*v1.ConfigMap, error) {
	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.CoreV1().ConfigMaps(cm.GetNamespace()).
		Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// GetConfigMapDetail 获取 ConfigMap 详情
func (s *Service) GetConfigMapDetail(ctx context.Context,
	clusterName entity.ClusterName, getReq *req.GetConfigMapDetailReq) (*v1.ConfigMap, error) {
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.CoreV1().
		ConfigMaps(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}

		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// GetCompatibleConfigMapDetail is just compatible with the configMap which name is old name.
func (s *Service) GetCompatibleConfigMapDetail(ctx context.Context, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, task *resp.TaskDetailResp) (*v1.ConfigMap, error) {
	res, err := s.GetConfigMapDetail(ctx, task.ClusterName, &req.GetConfigMapDetailReq{
		Namespace: task.Namespace,
		Name:      s.GetAppNewConfigMapName(task),
		Env:       string(task.EnvName),
	})
	if err == nil {
		return res, nil
	}
	if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
		return nil, err
	}

	res, err = s.GetConfigMapDetail(ctx, task.ClusterName, &req.GetConfigMapDetailReq{
		Namespace: task.Namespace,
		Name:      s.GetAppOldConfigMapName(project.Name, app.Name),
		Env:       string(task.EnvName),
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ListConfigMaps list all configMaps with options.
func (s *Service) ListConfigMaps(ctx context.Context, clusterName entity.ClusterName,
	listReq *req.ListConfigMapsReq) ([]v1.ConfigMap, error) {
	c, err := s.GetK8sTypedClient(clusterName, listReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	list, err := c.CoreV1().ConfigMaps(listReq.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: s.getConfigMapLabelSelector(listReq),
	})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	// sort configMap by time in descending order
	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].GetCreationTimestamp().After(
			list.Items[j].GetCreationTimestamp().Time)
	})

	return list.Items, nil
}

// PatchConfigMap 更新 ConfigMap
func (s *Service) PatchConfigMap(ctx context.Context,
	clusterName entity.ClusterName, cm *v1.ConfigMap, env string) (*v1.ConfigMap, error) {
	patchData, err := json.Marshal(cm)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.CoreV1().ConfigMaps(cm.GetNamespace()).
		Patch(ctx, cm.GetName(), types.MergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// DeleteConfigMap 删除 ConfigMap
func (s *Service) DeleteConfigMap(ctx context.Context,
	clusterName entity.ClusterName, deleteReq *req.DeleteConfigMapReq) error {
	policy := metav1.DeletePropagationForeground

	c, err := s.GetK8sTypedClient(clusterName, deleteReq.Env)
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	err = c.CoreV1().ConfigMaps(deleteReq.Namespace).
		Delete(ctx, deleteReq.Name, metav1.DeleteOptions{
			PropagationPolicy: &policy,
		})
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return nil
}

// DeleteConfigMaps delete k8s configMaps.
func (s *Service) DeleteConfigMaps(ctx context.Context, clusterName entity.ClusterName,
	deleteReq *req.DeleteConfigMapsReq) error {
	cms, err := s.ListConfigMaps(ctx, clusterName, &req.ListConfigMapsReq{
		Namespace:   deleteReq.Namespace,
		ProjectName: deleteReq.ProjectName,
		AppName:     deleteReq.AppName,
		Env:         deleteReq.Env,
	})
	if err != nil {
		return err
	}

	for i := range cms {
		if deleteReq.InverseVersion != "" && cms[i].GetName() == deleteReq.InverseVersion {
			continue
		}

		err = s.DeleteConfigMap(ctx, clusterName, &req.DeleteConfigMapReq{
			Namespace: cms[i].GetNamespace(),
			Name:      cms[i].GetName(),
			Env:       deleteReq.Env,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// getConfigMapLabelSelector joins label selector.
func (s *Service) getConfigMapLabelSelector(getReq *req.ListConfigMapsReq) string {
	labels := make([]string, 0)
	if getReq.ProjectName != "" {
		labels = append(labels, fmt.Sprintf("project=%s", getReq.ProjectName))
	}

	if getReq.AppName != "" {
		labels = append(labels, fmt.Sprintf("app=%s", getReq.AppName))
	}

	return strings.Join(labels, ",")
}

// GetConfigMapHash hashes cm data by HD5 algorithm.
func (s *Service) GetConfigMapHash(_ context.Context, cmData interface{}) (string, error) {
	bytes, err := yaml.Marshal(cmData)
	if err != nil {
		return "", errors.Wrap(errcode.InternalError, err.Error())
	}

	return fmt.Sprintf("%x", md5.Sum(bytes)), nil
}

// getCMLabels returns K8s configMap labels
func (s *Service) getCMLabels(task *resp.TaskDetailResp, cmHash string) map[string]string {
	labelsMap := make(map[string]string)
	labelsMap[ConfigMapLabelCommit] = task.Param.ConfigCommitID
	labelsMap[ConfigMapLabelHash] = cmHash

	return labelsMap
}
