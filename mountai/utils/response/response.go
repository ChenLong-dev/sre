package response

import (
	"context"
	"fmt"

	_errcode "rulai/utils/errcode"

	"github.com/gin-gonic/gin"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/response"
)

// 基础响应
type CustomResponse struct {
	Code    int         `json:"errcode"`
	Message string      `json:"errmsg"`
	Data    interface{} `json:"data,omitempty"`
	Detail  interface{} `json:"errdetail,omitempty"`

	// 用于返回实际错误
	err error
}

func (r *CustomResponse) WithErrCode(code errcode.Codes) interface{} {
	var message string
	switch code.Code() {
	case errcode.InvalidParams.Code():
		message = fmt.Sprintf("%#v", r.err)
	case _errcode.NeedCorrectAppNameError.Code():
		message = fmt.Sprintf("%#v", r.err)
	case _errcode.OtherRunningTaskExistsError.Code():
		message = fmt.Sprintf("%#v", r.err)
	default:
		message = code.Message()
	}

	var detail interface{}
	if errGroup, ok := code.(errcode.Group); ok {
		detail = errGroup.Details()
	}

	return CustomResponse{
		Code:    code.FrontendCode(),
		Message: message,
		Data:    r.Data,
		Detail:  detail,
	}
}

func (r *CustomResponse) PrintError(ctx context.Context, err error) {
	log.Errorv(ctx, errcode.GetErrorMessageMap(err))
	// 设置真实错误
	r.err = err
}

func (r *CustomResponse) GetStatusCode(code errcode.Codes) int {
	return code.StatusCode()
}

// 返回json
func JSON(ctx *gin.Context, data interface{}, err error) {
	response.InjectJson(
		ctx,
		&CustomResponse{
			Data: data,
		},
		err,
	)
}
