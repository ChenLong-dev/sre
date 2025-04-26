package resp

import (
	"k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type DescribeHPAResp struct {
	Status HorizontalPodAutoscalerStatus `json:"status"`
	Events []Event                       `json:"events"`
}

type HorizontalPodAutoscalerStatus struct {
	LastScaleTime  string                             `json:"last_scale_time,omitempty"`
	CurrentMetrics []MetricStatus                     `json:"current_metrics"`
	Conditions     []HorizontalPodAutoscalerCondition `json:"conditions"`
}

type MetricStatus struct {
	Type     v2.MetricSourceType   `json:"type"`
	Resource *ResourceMetricStatus `json:"resource,omitempty"`
}

type ResourceMetricStatus struct {
	Name    v1.ResourceName   `json:"name"`
	Current MetricValueStatus `json:"current"`
}

type MetricValueStatus struct {
	Value              *resource.Quantity `json:"value,omitempty"`
	AverageValue       *resource.Quantity `json:"average_value,omitempty"`
	AverageUtilization *int32             `json:"average_utilization,omitempty"`
}

type HorizontalPodAutoscalerCondition struct {
	Type               v2.HorizontalPodAutoscalerConditionType `json:"type"`
	Status             v1.ConditionStatus                      `json:"status"`
	LastTransitionTime string                                  `json:"last_transition_time,omitempty"`
	Reason             string                                  `json:"reason,omitempty"`
	Message            string                                  `json:"message,omitempty"`
}
