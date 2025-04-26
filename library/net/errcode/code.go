package errcode

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

// 错误码接口
type Codes interface {
	// 实际错误
	Error() string
	// 基础错误码
	Code() int
	// 返回到前端的错误码
	FrontendCode() int
	// 错误信息
	Message() string
	// 错误详情
	Details() []interface{}
	// 状态码
	StatusCode() int
}

// 查询导致错误的错误码
func Cause(e error) Codes {
	if e == nil {
		return OK
	}
	ec, ok := errors.Cause(e).(Codes)
	if ok {
		return ec
	}
	return String(e.Error())
}

// 判断错误码是否相等
func Equal(a, b Codes) bool {
	if a == nil {
		a = OK
	}
	if b == nil {
		b = OK
	}
	return a.Code() == b.Code()
}

// 判断错误与错误码是否相等
func EqualError(code Codes, err error) bool {
	return Cause(err).Code() == code.Code()
}

type ErrorDetail struct {
	Line    string `json:"line"`
	Message string `json:"message"`
}

// 获取错误信息map
func GetErrorMessageMap(err error) map[string]interface{} {
	detail := make([]ErrorDetail, 0)
	errString := strings.Replace(fmt.Sprintf("%+v", err), "\n\t", " => ", -1)
	for _, item := range strings.Split(errString, "\n") {
		var line, funcName string
		detailArray := strings.SplitN(item, " => ", 2)
		if len(detailArray) == 0 {
			continue
		} else if len(detailArray) == 1 {
			line = ""
			funcName = detailArray[0]
		} else {
			funcName = detailArray[0]
			line = detailArray[1]
		}
		detail = append(detail, ErrorDetail{
			Line:    line,
			Message: funcName,
		})
	}

	code := Cause(err)

	result := map[string]interface{}{
		"code":       code.Code(),
		"log":        code.Message(),
		"front_code": code.FrontendCode(),
		"detail":     detail,
	}

	return result
}
