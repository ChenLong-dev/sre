package service

import (
	"context"

	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"github.com/pkg/errors"
)

// CheckAppHealthCheckURLDifference 检验应用下各部署健康检查路径是否有差别
// 顺便检查是否待设置权重的集群已有部署
func (s *Service) CheckAppHealthCheckURLDifference(ctx context.Context, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, setReq *req.SetAppClusterQDNSWeightsReq) (healthCheckURL string, err error) {
	for _, weight := range setReq.ClusterWeights {
		var deployments []resp.Deployment
		deployments, err = s.GetDeployments(ctx, weight.ClusterName, setReq.Env, &req.GetDeploymentsReq{
			Namespace:   string(setReq.Env),
			ProjectName: project.Name,
			AppName:     app.Name,
			Env:         string(setReq.Env),
		})
		if err != nil {
			return "", err
		}

		if len(deployments) == 0 && weight.Weight > 0 {
			// 如果待配置权重的集群没有部署, 则不允许设置(正常流程先部署后切流量)
			return "", errors.Wrapf(_errcode.NoSuccessfulDeploymentError, "cluster=%s", weight.ClusterName)
		}

		for i := range deployments {
			// NOTE: 当前 Deployment 模板只有一个 container, 未来如果出现多个, 这里的逻辑需要留意一下
			for j := range deployments[i].Spec.Template.Spec.Containers {
				if healthCheckURL == "" {
					healthCheckURL = deployments[i].Spec.Template.Spec.Containers[j].ReadinessProbe.HTTPGet.Path
				} else if healthCheckURL != deployments[i].Spec.Template.Spec.Containers[j].ReadinessProbe.HTTPGet.Path {
					return "", errors.Wrapf(_errcode.DifferentHealthCheckURLError, "app=%s", app.ID)
				}
			}
		}
	}

	return healthCheckURL, nil
}
