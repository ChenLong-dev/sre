package req

// GetCronHPADetailReq 获取单个cronHPA详情请求体
type GetCronHPADetailReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

// GetCronHPAsReq 获取多个cronHAP请求体
type GetCronHPAsReq struct {
	Namespace   string `json:"namespace"`
	ProjectName string `json:"project_name"`
	AppName     string `json:"app_name"`
	Version     string `json:"version"`
	Env         string `json:"env"`
}

// DeleteCronHPAsReq 删除多个cronHPA请求体
type DeleteCronHPAsReq struct {
	Namespace      string `json:"namespace"`
	ProjectName    string `json:"project_name"`
	AppName        string `json:"app_name"`
	InverseVersion string `json:"inverse_version"`
	Env            string `json:"env"`
}

// DeleteCronHPAReq 删除单个cronHPA请求体
type DeleteCronHPAReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}
