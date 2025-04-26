package response

import (
	"context"
	"net/http"

	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// 基础响应
// Deprecated: 请使用v2版本
type V1Response struct {
	// 错误码
	Code int `json:"errcode"`
	// 错误信息
	Message string `json:"errmsg"`
	// 错误详情
	Detail interface{} `json:"errdetail,omitempty"`
	// 数据
	Data interface{} `json:"data,omitempty"`
}

func (r *V1Response) WithErrCode(code errcode.Codes) interface{} {
	var detail interface{}
	if code.Code() != errcode.OK.Code() {
		detail = code.Details()
	}
	return V1Response{
		Code:    code.FrontendCode(),
		Message: code.Message(),
		Detail:  detail,
		Data:    r.Data,
	}
}

func (r *V1Response) PrintError(ctx context.Context, err error) {
	log.Errorv(ctx, errcode.GetErrorMessageMap(err))
}

func (r *V1Response) GetStatusCode(code errcode.Codes) int {
	switch code.Code() {
	case errcode.NotFound.Code():
		return http.StatusNotFound
	case errcode.ServiceUnavailable.Code():
		return http.StatusServiceUnavailable
	default:
		return http.StatusOK
	}
}
