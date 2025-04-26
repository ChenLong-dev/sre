package resp

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentConditionType [多集群临时方案] 统一k8s各个版本的DeploymentConditionType
type DeploymentConditionType string

// DescribeDeploymentResp 获取Deployment的Describe信息响应参数
type DescribeDeploymentResp struct {
	Status DeploymentStatus `json:"status"`
	Events []Event          `json:"events"`
}

// DeploymentStatus 获取Deployment运行状态信息(非全部信息)响应参数
type DeploymentStatus struct {
	Replicas            int32                 `json:"replicas"`
	ReadyReplicas       int32                 `json:"readyReplicas"`
	AvailableReplicas   int32                 `json:"availableReplicas"`
	UpdatedReplicas     int32                 `json:"updatedReplicas"`
	UnavailableReplicas int32                 `json:"unavailableReplicas"`
	Conditions          []DeploymentCondition `json:"conditions"`
}

// DeploymentCondition [多集群临时方案] 统一k8s各个版本的DeploymentCondition
type DeploymentCondition struct {
	Type               DeploymentConditionType `json:"type"`
	Status             v1.ConditionStatus      `json:"status"`
	LastUpdateTime     string                  `json:"last_update_time"`
	LastTransitionTime string                  `json:"last_transition_time"`
	Reason             string                  `json:"reason,omitempty"`
	Message            string                  `json:"message,omitempty"`
}

// Deployment [多集群临时方案] 统一k8s各个版本的Deployment
type Deployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DeploymentSpec   `json:"spec"`
	Status            DeploymentStatus `json:"status"`
}

// DeploymentSpec [多集群临时方案] 统一k8s各个版本的DeploymentSpec
type DeploymentSpec struct {
	Replicas *int32             `json:"replicas"`
	Template v1.PodTemplateSpec `json:"template"`
	Paused   bool               `json:"paused,omitempty"`
}

// DeploymentList [多集群临时方案] 统一k8s各个版本的 DeploymentList (只取需要的字段 items)
type DeploymentList struct {
	Items []Deployment `json:"items"`
}
