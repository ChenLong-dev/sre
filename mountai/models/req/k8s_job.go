package req

type GetJobDetailReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type GetJobsReq struct {
	Namespace   string `json:"namespace"`
	ProjectName string `json:"project_name"`
	AppName     string `json:"app_name"`
	Version     string `json:"version"`
	Env         string `json:"env"`
}

type DeleteJobsReq struct {
	Namespace   string `json:"namespace"`
	ProjectName string `json:"project_name"`
	AppName     string `json:"app_name"`
	InverseName string `json:"inverse_name"`
	Env         string `json:"env"`
}

type DeleteJobReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type DescribeJobReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}
