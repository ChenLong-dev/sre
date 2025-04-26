package req

type GetRepoTagsReq struct {
	// 项目名
	ProjectName string `json:"project_name"`
	Page        int    `json:"page"`
	Size        int    `json:"size"`
}
