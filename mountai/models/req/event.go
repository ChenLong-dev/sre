package req

import "k8s.io/apimachinery/pkg/runtime"

type GetK8sResourceEventsReq struct {
	Namespace       string         `json:"namespace"`
	Resource        runtime.Object `json:"resource"`
	MaxEventsLength int            `json:"max_events_length"`
	Env             string         `json:"env"`
}
