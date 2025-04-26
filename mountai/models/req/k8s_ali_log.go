package req

import "gitlab.shanhai.int/sre/library/base/null"

type GetAliLogConfigDetailReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type UpdateAliLogConfigReq struct {
	Namespace string `json:"namespace"`
}

type DeleteAliLogConfigReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type ListAliLogConfigReq struct {
	Namespace       string    `json:"namespace"`
	ProjectName     string    `json:"project_name"`
	OpenColdStorage null.Bool `json:"open_cold_storage"`
}
