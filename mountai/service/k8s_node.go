package service

import (
	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	"rulai/models/entity"
	_errcode "rulai/utils/errcode"
)

// WatchNodes returns a nodes watch channel to a cluster
func (s *Service) WatchNodes(ctx context.Context, clusterName entity.ClusterName, env string) (c kubernetes.Interface,
	ch watch.Interface, err error) {
	c, err = s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		log.Errorc(ctx, err.Error())
		return nil, nil, err
	}

	ch, err = c.CoreV1().Nodes().Watch(ctx, v1.ListOptions{
		LabelSelector:       "",
		Watch:               true,
		AllowWatchBookmarks: false,
		TimeoutSeconds:      nil,
	})

	if err != nil {
		log.Errorc(ctx, err.Error())
		return nil, nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}
	return c, ch, nil
}
