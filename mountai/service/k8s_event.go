package service

import (
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"

	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
)

// 默认值配置
const (
	// 默认 k8s 事件数量
	DefaultK8sResourceEventsLength = 10
)

// GetK8sResourceEvents 获取 k8s 资源事件
func (s *Service) GetK8sResourceEvents(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetK8sResourceEventsReq) (*resp.GetK8sResourceEventsResp, error) {
	if getReq.MaxEventsLength == 0 {
		getReq.MaxEventsLength = DefaultK8sResourceEventsLength
	}

	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	events, err := c.CoreV1().Events(getReq.Namespace).
		Search(scheme.Scheme, getReq.Resource)
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
		}

		return nil, nil
	}

	if events != nil && len(events.Items) > getReq.MaxEventsLength {
		events.Items = events.Items[:getReq.MaxEventsLength]
	}

	res := new(resp.GetK8sResourceEventsResp)
	err = deepcopy.Copy(events).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	eventList := make([]resp.Event, len(res.Events))
	for i := range res.Events {
		item := res.Events[i]
		item.LastTimestamp = utils.FormatK8sTime(&events.Items[i].LastTimestamp)
		item.FirstTimestamp = utils.FormatK8sTime(&events.Items[i].FirstTimestamp)
		item.EventTime = utils.FormatK8sTime(&events.Items[i].EventTime)
		eventList[i] = item
	}
	res.Events = eventList

	return res, nil
}
