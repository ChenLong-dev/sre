package service

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"rulai/config"
	"rulai/models/entity"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"
)

// isClusterSupportedK8sObjectKind 校验 k8s 资源类型是否被集群支持
func (s *Service) isClusterSupportedK8sObjectKind(clusterName entity.ClusterName, namespace, kind string) error {
	clusterInfo, err := s.getClusterInfo(clusterName, namespace)
	if err != nil {
		return err
	}

	_, ok := entity.AMSSupportedK8sObjectKindsByAllVendors[kind]
	if ok {
		return nil
	}

	vendorSupportedKinds, ok := entity.AMSSupportedK8sObjectKindsByVendors[clusterInfo.vendor]
	if !ok {
		return errors.Wrap(_errcode.UnknownVendorError, string(clusterInfo.vendor))
	}

	_, ok = vendorSupportedKinds[kind]
	if ok {
		return nil
	}

	return errors.Wrap(_errcode.UnsupportedK8sObjectKind, kind)
}

// getWorkloadAnnotationsByVendor 添加运营商独有的 spec.template.metadata.annotations
func (s *Service) getWorkloadAnnotationsByVendor(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, team *resp.TeamDetailResp) (map[string]string, error) {
	annotations := make(map[string]string)
	clusterInfo, err := s.getClusterInfo(clusterName, string(envName))
	if err != nil {
		return nil, err
	}

	switch clusterInfo.vendor {
	case entity.VendorAli:
		// 滚动更新期间暂停 HPA
		annotations[K8sAnnotationHPASkipped] = "true"
		return annotations, nil

	case entity.VendorHuawei:
		// 当前 AMS 没有 pod 内多容器区分收集日志的场景, 可以不填写 AnnotationsAOMLogStdout
		// 注意 entity.vendor.go 中定义的各种限制
		relabels := map[string]string{
			"team":        team.Label,
			"project":     project.Name,
			"app":         app.Name,
			"cluster":     string(clusterName),
			"env":         string(envName),
			"coldStorage": "false", // TODO: 支持华为云日志转储后字段能够变更
		}

		// 华为云日志相关 annotations relabel 标签限制 今后可能会发生变化, 加 WARNING 日志提醒, 如果部署失败时有依据可查
		if len(relabels) > entity.AnnotationsAOMLogRelabelLimits {
			log.Warnc(ctx, "%s's annotations relabels key count limit(%d) exceeded",
				clusterInfo.vendor, entity.AnnotationsAOMLogRelabelLimits)
		}

		for k, v := range relabels {
			if _, ok := entity.AnnotationsAOMLogRelabelKeyReverseSet[k]; ok {
				log.Warnc(ctx, "%s key is in %s's annotations relabels reverse set", k, clusterInfo.vendor)
			}

			if len(v) > entity.AnnotationsAOMLogRelabelKeyValueLengthLimit {
				// 此时值会被强制截断
				log.Warnc(ctx, "%s field %s exceeds %s's annotations relabels limit(%d)",
					k, v, clusterInfo.vendor, entity.AnnotationsAOMLogRelabelKeyValueLengthLimit)
			}
		}

		relabelsBytes, e := json.Marshal(relabels)
		if e != nil {
			return nil, errors.Wrapf(_errcode.K8sInternalError, "marshal relabels error: %s", err)
		}

		annotations[entity.AnnotationsAOMLogRelabel] = string(relabelsBytes)

	default:
		return nil, errors.Wrap(_errcode.UnknownVendorError, string(clusterInfo.vendor))
	}

	return annotations, nil
}

// 获取通用系统环境变量
func (s *Service) getSystemEnv(project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, _ *resp.TeamDetailResp) ([]*entity.EnvTemplate, error) {
	envVars, err := s.getVendorSystemEnv(task.ClusterName, task.EnvName)
	if err != nil {
		return nil, err
	}

	envVars = append(envVars, &entity.EnvTemplate{
		Name:  SystemEnvKeyAppID,
		Value: app.ID,
	}, &entity.EnvTemplate{
		Name:  SystemEnvKeyAppName,
		Value: app.Name,
	}, &entity.EnvTemplate{
		Name:  SystemEnvKeyProjectName,
		Value: project.Name,
	}, &entity.EnvTemplate{
		Name:  SystemEnvKeyProjectID,
		Value: project.ID,
	}, &entity.EnvTemplate{
		Name:  SystemEnvKeyAppEnv,
		Value: string(task.EnvName),
	}, &entity.EnvTemplate{
		Name:  SystemEnvKeyHost,
		Value: config.Conf.Other.AMSHost,
	}, &entity.EnvTemplate{
		Name:  "TZ",
		Value: "Asia/Shanghai",
	})

	if len(config.Conf.JWT.K8sSystemUserTokens) != 0 {
		envVars = append(envVars, &entity.EnvTemplate{
			Name:  SystemEnvKeyAuthorizationToken,
			Value: config.Conf.JWT.K8sSystemUserTokens[len(config.Conf.JWT.K8sSystemUserTokens)-1],
		})
	}

	return envVars, nil
}

// getCommonVendorEnv 获取通用运营商相关环境变量
func (s *Service) getVendorSystemEnv(clusterName entity.ClusterName, envName entity.AppEnvName) ([]*entity.EnvTemplate, error) {
	c, err := s.getVendorControllerFromClusterAndNamespace(clusterName, string(envName))
	if err != nil {
		return nil, err
	}

	return []*entity.EnvTemplate{
		{Name: SystemEnvKeyVendor, Value: string(c.Name())},
		{Name: SystemEnvKeyRegion, Value: c.Region()},
	}, nil
}

// equalK8sGroupVersion 比较两个 k8s schema.GroupVersion 是否相等
func equalK8sGroupVersion(ver1, ver2 *schema.GroupVersion) bool {
	return ver1 != nil && ver2 != nil && ver1.Group == ver2.Group && ver1.Version == ver2.Version
}

// GetNamespaceByApp 根据 app 返回对应的命名空间
func (s *Service) GetNamespaceByApp(app *resp.AppDetailResp, env entity.AppEnvName, cluster entity.ClusterName) string {
	return s.GetNamespaceBase(s.GetApplicationIstioState(context.Background(), env, cluster, app), env)
}

// GetNamespaceBase 根据是否启用 istio 状态返回对应的命名空间
func (s *Service) GetNamespaceBase(enableIstio bool, envName entity.AppEnvName) string {
	if enableIstio {
		return entity.IstioNamespacePrefix + string(envName)
	}
	return string(envName)
}
