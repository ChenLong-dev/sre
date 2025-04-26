package resp

import (
	bathV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

type DescribeJobResp struct {
	Status *JobStatus `json:"status"`
	Events []Event    `json:"events"`
}

type JobStatus struct {
	Conditions []JobCondition `json:"conditions,omitempty"`
	StartTime  string         `json:"start_time,omitempty"`
	Active     int32          `json:"active,omitempty"`
	Succeeded  int32          `json:"succeeded,omitempty"`
	Failed     int32          `json:"failed,omitempty"`
}

type JobCondition struct {
	Type               bathV1.JobConditionType `json:"type"`
	Status             v1.ConditionStatus      `json:"status"`
	LastProbeTime      string                  `json:"last_probe_time,omitempty"`
	LastTransitionTime string                  `json:"last_transition_time,omitempty"`
	Reason             string                  `json:"reason,omitempty"`
	Message            string                  `json:"message,omitempty"`
}
