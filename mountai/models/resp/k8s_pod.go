package resp

import v1 "k8s.io/api/core/v1"

type DescribePodResp struct {
	Status PodStatus `json:"status"`
	Events []Event   `json:"events"`
}

type PodStatus struct {
	Phase      v1.PodPhase    `json:"phase"`
	Conditions []PodCondition `json:"conditions"`
	HostIP     string         `json:"host_ip"`
	PodIP      string         `json:"pod_ip"`
}

type PodCondition struct {
	Type               v1.PodConditionType `json:"type"`
	Status             v1.ConditionStatus  `json:"status"`
	LastProbeTime      string              `json:"last_probe_time"`
	LastTransitionTime string              `json:"last_transition_time"`
	Reason             string              `json:"reason,omitempty"`
	Message            string              `json:"message,omitempty"`
}
