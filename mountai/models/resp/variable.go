package resp

type Variable struct {
	ID         string           `json:"id" deepcopy:"objectid"`
	Key        string           `json:"key"`
	Value      string           `json:"value"`
	Owner      *UserProfileResp `json:"owner"`
	Editor     *UserProfileResp `json:"editor"`
	CreateTime string           `json:"create_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	UpdateTime string           `json:"update_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
}
