package resp

type DescribeVirtualServiceResp struct {
	Events []Event `json:"events" deepcopy:"from:Items"`
	Spec   *Spec   `json:"status"`
}

type Spec struct {
	Gateways []string     `json:"gateways"`
	Hosts    []string     `json:"hosts"`
	HTTP     []*HTTPRoute `json:"http"`
}

type HTTPRoute struct {
	Route []*Route `json:"route"`
}

type Route struct {
	Destination *DestinationRule `json:"destination"`
}

type DestinationRule struct {
	Host   string `json:"host"`
	Port   int    `json:"port,omitempty"`
	Subset string `json:"subset,omitempty"`
}

// SyncIstioVirtualServiceResp 同步更新 vs 响应结果
type SyncIstioVirtualServiceResp struct {
	Hosts              []string `json:"hosts"`
	VirtualServiceName string   `json:"virtualServiceName"`
}

// DetermineUpstreamNameResp 添加外网域名返回结果
type DetermineUpstreamNameResp map[string]*UpstreamInfo

type UpstreamInfo struct {
	BackendHostName          string `json:"backend_host_name"`           // 请求的 backend host name
	EnableIstio              bool   `json:"enable_istio"`                // 是否启用 istio
	RecommendBackendHostName string `json:"recommend_backend_host_name"` // 推荐的 backend host name
	Tag                      string `json:"tag"`
}
