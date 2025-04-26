package req

import "rulai/models/entity"

type VirtualServiceReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type DescribeVirtualServiceReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

// SyncIstioVirtualServiceReq 同步更新 virtual service 请求
type SyncIstioVirtualServiceReq struct {
	Tags            []string               `json:"tags"`              // tag name
	KongClusterName entity.KongClusterName `json:"kong_cluster_name"` // kong cluster name
	Env             entity.AppEnvName      `json:"env"`               // entity.AppEnvName
}

// DetermineUpstreamNameReq 确认后端服务地址是否需要更换
type DetermineUpstreamNameReq struct {
	BackendHost []string `json:"backend_host" binding:"required,min=1"` // 后端服务地址
}
