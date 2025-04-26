package resp

type DingTalkCommonResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// DingTalkCropMessageResp 工作通知
type DingTalkCropMessageResp struct {
	DingTalkCommonResp
	Data *DingTalkCropMessageData `json:"data"`
}

// DingTalkCropMessageData 工作通知数据
type DingTalkCropMessageData struct {
	TaskID    int64  `json:"task_id"`
	RequestID string `json:"request_id"`
}

type CreateDingApprovalInstanceResp struct {
	DingTalkCommonResp
	Data *DingApprovalInstanceData `json:"data"`
}

type DingApprovalInstanceData struct {
	DingTalkCommonResp
	ProcessInstanceID string `json:"process_instance_id"`
}

type GetDingApprovalInstanceResp struct {
	DingTalkCommonResp
	Data *DingApprovalInstanceRespData `json:"data"`
}

type DingApprovalInstanceRespData struct {
	OriginatorUserID    string                `json:"originator_userid"`
	Status              string                `json:"status"`
	Result              string                `json:"result"`
	FormComponentValues []*FormComponentValue `json:"form_component_values"`
}

type FormComponentValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
