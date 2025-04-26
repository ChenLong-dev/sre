package resp

type DescribeIngressResp struct {
	Status IngressStatus `json:"status"`
	Events []Event       `json:"events"`
}

type IngressStatus struct {
	LoadBalancer LoadBalancerStatus `json:"load_balancer"`
}

type LoadBalancerStatus struct {
	Ingress []LoadBalancerIngress `json:"ingress"`
}

type LoadBalancerIngress struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
}
