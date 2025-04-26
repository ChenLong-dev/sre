package resp

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicaSetConditionType [多集群临时方案] 统一k8s各个版本的 ReplicaSetConditionType
type ReplicaSetConditionType string

// ReplicaSet [多集群临时方案] 统一k8s各个版本的 ReplicaSet
type ReplicaSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ReplicaSetSpec   `json:"spec"`
	Status            ReplicaSetStatus `json:"status"`
}

// ReplicaSetSpec [多集群临时方案] 统一k8s各个版本的 ReplicaSetSpec
type ReplicaSetSpec struct {
	Replicas        *int32                `json:"replicas"`
	MinReadySeconds int32                 `json:"minReadySeconds"`
	Selector        *metav1.LabelSelector `json:"selector"`
	Template        v1.PodTemplateSpec    `json:"template"`
}

// ReplicaSetStatus [多集群临时方案] 统一k8s各个版本的 ReplicaSetStatus
type ReplicaSetStatus struct {
	Replicas             int32                 `json:"replicas"`
	FullyLabeledReplicas int32                 `json:"fullyLabeledReplicas"`
	ReadyReplicas        int32                 `json:"readyReplicas"`
	AvailableReplicas    int32                 `json:"availableReplicas"`
	ObservedGeneration   int64                 `json:"observedGeneration"`
	Conditions           []ReplicaSetCondition `json:"conditions"`
}

// ReplicaSetCondition [多集群临时方案] 统一k8s各个版本的 ReplicaSetCondition
type ReplicaSetCondition struct {
	Type               ReplicaSetConditionType `json:"type"`
	Status             v1.ConditionStatus      `json:"status"`
	LastTransitionTime metav1.Time             `json:"lastTransitionTime"`
	Reason             string                  `json:"reason"`
	Message            string                  `json:"message"`
}

// ReplicaSetList [多集群临时方案] 统一k8s各个版本的 ReplicaSetList (只取需要的字段 items)
type ReplicaSetList struct {
	Items []ReplicaSet `json:"items"`
}
