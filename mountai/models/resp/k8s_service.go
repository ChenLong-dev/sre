package resp

type DescribeServiceResp struct {
	Status ServiceStatus `json:"status"`
	Events []Event       `json:"events"`
}

type ServiceStatus struct {
	Endpoints      []ServiceEndpoints `json:"endpoints"`
	LoadBalancerIP []string           `json:"lb_ip"`
	ClusterIP      string             `json:"cluster_ip"`
}

type ServiceEndpoints struct {
	IP       string `json:"ip"`
	NodeName string `json:"node_name"`
}
