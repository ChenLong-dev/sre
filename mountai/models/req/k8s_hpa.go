package req

type GetHPADetailReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type UpdateHPAReq struct {
	Namespace string `json:"namespace"`
	Env       string `json:"env"`
}

type DeleteHPAReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type GetHPAsReq struct {
	Namespace   string `json:"namespace"`
	ProjectName string `json:"project_name"`
	AppName     string `json:"app_name"`
	Version     string `json:"version"`
	Env         string `json:"env"`
}

type DeleteHPAsReq struct {
	Namespace      string `json:"namespace"`
	ProjectName    string `json:"project_name"`
	AppName        string `json:"app_name"`
	InverseVersion string `json:"inverse_version"`
	Env            string `json:"env"`
}

type DescribeHPAReq struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Env       string `json:"env"`
}
