package resp

type ImageArgsTemplateDetail struct {
	ID         string           `json:"id" deepcopy:"objectid"`
	Name       string           `json:"name"`
	Content    string           `json:"content"`
	Team       *TeamDetailResp  `json:"team"`
	Owner      *UserProfileResp `json:"owner"`
	CreateTime string           `json:"create_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	UpdateTime string           `json:"update_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
}

type ImageArgsTemplate struct {
	ID         string `json:"id" deepcopy:"objectid"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	CreateTime string `json:"create_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	UpdateTime string `json:"update_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
}
