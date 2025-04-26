package req

type GetCronJobDetailReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type GetCronJobsReq struct {
	Namespace   string `json:"namespace"`
	ProjectName string `json:"project_name"`
	AppName     string `json:"app_name"`
	Version     string `json:"version"`
	Env         string `json:"env"`
}

type UpdateCronJobReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Replicas  int    `json:"replicas"`
	Env       string `json:"env"`
}

type DeleteCronJobsReq struct {
	Namespace   string `json:"namespace"`
	ProjectName string `json:"project_name"`
	AppName     string `json:"app_name"`
	InverseName string `json:"inverse_name"`
	Env         string `json:"env"`
}

type DeleteCronJobReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type DescribeCronJobReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}
