package response

import (
	"context"

	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// 基础响应
type V2Response struct {
	// 错误码
	Code int `json:"errcode"`
	// 错误信息
	Message string `json:"errmsg"`
	// 错误详情
	Detail interface{} `json:"errdetail,omitempty"`
	// 数据
	Data interface{} `json:"data,omitempty"`
}

func (r *V2Response) WithErrCode(code errcode.Codes) interface{} {
	var detail interface{}
	if code.Code() != errcode.OK.Code() {
		detail = code.Details()
	}
	return V2Response{
		Code:    code.FrontendCode(),
		Message: code.Message(),
		Detail:  detail,
		Data:    r.Data,
	}
}

func (r *V2Response) PrintError(ctx context.Context, err error) {
	log.Errorv(ctx, errcode.GetErrorMessageMap(err))
}

func (r *V2Response) GetStatusCode(code errcode.Codes) int {
	return code.StatusCode()
}

// 分页响应
type V2PaginationResponse struct {
	// 总数
	Total int `json:"total"`
	// 页码
	// 从0开始
	Page int `json:"page"`
	// 每页数量
	PageSize int `json:"pagesize"`
	// 数据
	Items interface{} `json:"items"`
}
