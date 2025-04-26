package service

import (
	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// GetEndpointsDetail 获取 k8s Endpoints 详情
func (s *Service) GetEndpointsDetail(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetEndpointsReq) (*v1.Endpoints, error) {
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	endpoints, err := c.CoreV1().Endpoints(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return endpoints, nil
}

// CreateKongGatewayEndpoints create ep for kong
func (s *Service) CreateKongGatewayEndpoints(ctx context.Context, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	c, err := s.GetK8sTypedClient(task.ClusterName, string(task.EnvName))
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	ep := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.ServiceName,
			Namespace: string(task.EnvName),
			Annotations: map[string]string{
				resourcelock.LeaderElectionRecordAnnotationKey: "",
			},
			Labels: map[string]string{
				"project": project.Name,
				"app":     app.Name,
			},
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: config.Conf.Kong.Envs[string(task.EnvName)].KongLB,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     "http",
						Protocol: entity.ServiceTCPProtocol,
						Port:     80,
					},
				},
			},
		},
	}

	_, err = c.CoreV1().Endpoints(string(task.EnvName)).Create(ctx, ep, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return nil
}
