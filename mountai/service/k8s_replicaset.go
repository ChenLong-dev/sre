package service

import (
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetReplicaSetDetail 获取 ReplicaSet 详情
func (s *Service) GetReplicaSetDetail(ctx context.Context, clusterName entity.ClusterName, envName string,
	getReq *req.GetReplicaSetDetailReq) (*resp.ReplicaSet, error) {
	c, err := s.GetK8sTypedClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, entity.AppEnvName(getReq.Namespace), entity.K8sObjectKindReplicaSet)
	if err != nil {
		return nil, err
	}

	var resource runtime.Object
	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		resource, err = c.ExtensionsV1beta1().ReplicaSets(getReq.Namespace).Get(ctx, getReq.Name, metav1.GetOptions{})
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		resource, err = c.AppsV1().ReplicaSets(getReq.Namespace).Get(ctx, getReq.Name, metav1.GetOptions{})
	} else {
		return nil, errors.Wrapf(_errcode.K8sInternalError,
			"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindReplicaSet)
	}

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res := new(resp.ReplicaSet)
	err = deepcopy.Copy(resource).To(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}
	return res, nil
}

func (s *Service) getReplicaSetLabelSelector(getReq *req.GetReplicaSetsReq) string {
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

// GetReplicaSets 获取 ReplicaSet 列表
func (s *Service) GetReplicaSets(ctx context.Context, clusterName entity.ClusterName, envName string,
	getReq *req.GetReplicaSetsReq) ([]resp.ReplicaSet, error) {
	c, err := s.GetK8sTypedClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	selector := s.getReplicaSetLabelSelector(getReq)

	groupVersion, err := s.getK8sResourceGroupVersion(ctx, clusterName, entity.AppEnvName(getReq.Namespace), entity.K8sObjectKindReplicaSet)
	if err != nil {
		return nil, err
	}

	var resource runtime.Object
	// TODO: 未来期望使用 Dynamic Client 通过 schema.GroupVersionResource 动态创建, 避免通过版本或集群进行 Hack
	if equalK8sGroupVersion(groupVersion, groupVersionExtensionsV1beta1) {
		resource, err = c.ExtensionsV1beta1().ReplicaSets(getReq.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	} else if equalK8sGroupVersion(groupVersion, groupVersionAppsV1) {
		resource, err = c.AppsV1().ReplicaSets(getReq.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	} else {
		return nil, errors.Wrapf(_errcode.K8sInternalError,
			"unknown apiVersion(%s) of %s", groupVersion.String(), entity.K8sObjectKindReplicaSet)
	}

	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res := new(resp.ReplicaSetList)
	err = deepcopy.Copy(resource).To(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	// 时间倒序
	sort.Slice(res.Items, func(i, j int) bool {
		return res.Items[i].GetCreationTimestamp().After(res.Items[j].GetCreationTimestamp().Time)
	})
	return res.Items, nil
}
