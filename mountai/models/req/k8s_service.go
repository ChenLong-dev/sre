package req

type GetServiceDetailReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type GetServicesReq struct {
	Namespace   string `json:"namespace"`
	ProjectName string `json:"project_name"`
	AppName     string `json:"app_name"`
	Env         string `json:"env"`
}

type DeleteServiceReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type DescribeServiceReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type GetEndpointsReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}
