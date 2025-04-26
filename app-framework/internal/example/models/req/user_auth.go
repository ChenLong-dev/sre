package req

import "gitlab.shanhai.int/sre/library/base/null"

// ==========================
// 请求模型
// ==========================
type VerifyUserReq struct {
	QingTingID  null.String `json:"qingting_id" form:"qingting_id"`
	DeviceID    null.String `json:"device_id" form:"device_id"`
	AccessToken null.String `json:"access_token" form:"access_token"`
	JWTToken    null.String `json:"jwt_token" form:"jwt_token"`
	TokenType   null.String `json:"token_type" form:"token_type"`
}

// ==========================
// 用户鉴权实际请求模型
// ==========================
type UserAuthVerifyReq struct {
	QingTingID  string `json:"qingting_id" form:"qingting_id"`
	DeviceID    string `json:"device_id" form:"device_id"`
	AccessToken string `json:"access_token" form:"access_token"`
	JWTToken    string `json:"jwt_token" form:"jwt_token"`
	TokenType   string `json:"token_type" form:"token_type"`
}
