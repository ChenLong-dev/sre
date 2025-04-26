package req

type GetReplicaSetDetailReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type GetReplicaSetsReq struct {
	Namespace   string `json:"namespace"`
	ProjectName string `json:"project_name"`
	AppName     string `json:"app_name"`
	Version     string `json:"version"`
}
