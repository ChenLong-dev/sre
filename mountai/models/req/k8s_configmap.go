package req

type GetConfigMapDetailReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type UpdateConfigMapReq struct {
	Namespace string `json:"namespace"`
	Env       string `json:"env"`
}

type DeleteConfigMapReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type ListConfigMapsReq struct {
	Namespace   string `json:"namespace"`
	ProjectName string `json:"project_name"`
	AppName     string `json:"app_name"`
	Env         string `json:"env"`
}

type DeleteConfigMapsReq struct {
	Namespace      string `json:"namespace"`
	ProjectName    string `json:"project_name"`
	AppName        string `json:"app_name"`
	InverseVersion string `json:"inverse_version"`
	Env            string `json:"env"`
}
