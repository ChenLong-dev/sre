package service

import (
	"context"

	"github.com/pkg/errors"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"rulai/models/entity"
	"rulai/models/req"
	_errcode "rulai/utils/errcode"
)

func (s *Service) GetGatewayDetail(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName,
	getReq *req.GetGatewayReq) (*v1beta1.Gateway, error) {
	info, err := s.getClusterInfo(clusterName, string(envName))
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	ic, err := versionedclient.NewForConfig(info.config)
	if err != nil {
		return nil, errors.Wrapf(_errcode.K8sInternalError, err.Error())
	}

	return ic.NetworkingV1beta1().Gateways(getReq.Namespace).Get(ctx, getReq.Name, metav1.GetOptions{})
}
