package resp

// ==========================
// 响应模型
//
// deepcopy标记使用deepcopy包: https://gitlab.shanhai.int/sre/library/tree/master/base/deepcopy.v2
// ==========================
type VerifyUserResp struct {
	QingTingID string `json:"qingting_id"`
	Verify     bool   `json:"verify" deepcopy:"method:GetVerifyResult"`
}

const (
	UserAuthVerifyPass = "pass"
	UserAuthVerifyDeny = "deny"
)

// ==========================
// 用户鉴权实际响应模型
// ==========================
type UserAuthVerifyResp struct {
	Code    int                       `json:"errcode"`
	Message string                    `json:"errmsg"`
	Data    *UserAuthVerifyDetailResp `json:"data"`
}
type UserAuthVerifyDetailResp struct {
	QingTingID  string `json:"qingting_id"`
	Verify      string `json:"verify"`
	AccessToken string `json:"access_token"`
}

func (r *UserAuthVerifyDetailResp) GetVerifyResult(args map[string]interface{}) bool {
	return r.Verify == UserAuthVerifyPass
}
