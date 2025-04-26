package req

type GetIngressDetailReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type DeleteIngressReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type DescribeIngressReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}
