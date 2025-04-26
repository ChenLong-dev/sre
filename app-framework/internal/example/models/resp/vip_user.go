package resp

// ==========================
// 响应模型
//
// deepcopy标记使用deepcopy包: https://gitlab.shanhai.int/sre/library/tree/master/base/deepcopy.v2
// ==========================
type VipUserDetail struct {
	ID         string `json:"id" deepcopy:"method:GenerateIDString"`
	UserID     string `json:"user_id"`
	Type       string `json:"type"`
	ExpireTime string `json:"expire_time" deepcopy:"timeformat:2006/01/02 15:04:05"`
}
