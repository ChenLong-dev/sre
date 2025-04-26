package errcode

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

var (
	// register codes.
	_codes = map[int]ErrCode{}
)

// 新建业务错误码
// 业务错误码必须大于等于2000000
func New(e int, msg string) ErrCode {
	if e < 2000000 {
		panic("business ecode must greater than zero")
	}
	return add(http.StatusInternalServerError, e, msg)
}

// 从状态码中获取错误码
func FromStatusCode(statusCode int) ErrCode {
	errMsg := http.StatusText(statusCode)
	if errMsg == "" {
		panic(fmt.Sprintf("status code: %d not exist", statusCode))
	}

	errCode := 1000000 + (statusCode/100)*10000 + statusCode
	if _, ok := _codes[errCode]; ok {
		panic(fmt.Sprintf("ecode: %d already exist", errCode))
	}

	return add(statusCode, errCode, errMsg)
}

// 添加错误码
func add(statusCode int, errCode int, errMsg string) ErrCode {
	if _, ok := _codes[errCode]; ok {
		panic(fmt.Sprintf("ecode: %d already exist", errCode))
	}

	e := ErrCode{
		statusCode:   statusCode,
		code:         errCode,
		frontendCode: errCode,
		message:      errMsg,
	}
	_codes[errCode] = e
	return e
}

// 错误码
type ErrCode struct {
	// 状态码
	statusCode int
	// 基础错误码，要求唯一
	code int
	// 返回给前端的错误码
	frontendCode int
	// 错误信息
	message string
}

func (e ErrCode) Clone() ErrCode {
	return ErrCode{
		code:         e.code,
		frontendCode: e.frontendCode,
		statusCode:   e.statusCode,
		message:      e.message,
	}
}

func (e ErrCode) Error() string {
	return fmt.Sprintf("%d:%s", e.code, e.message)
}

func (e ErrCode) Code() int { return e.code }

func (e ErrCode) Message() string {
	return e.message
}

func (e ErrCode) FrontendCode() int {
	return e.frontendCode
}

func (e ErrCode) Details() []interface{} {
	return []interface{}{e.Error()}
}

func (e ErrCode) StatusCode() int {
	if e.statusCode == 0 {
		return http.StatusOK
	}
	return e.statusCode
}

// 设置前端错误码
func (e ErrCode) WithFrontCode(frontendCode int) ErrCode {
	c := e.Clone()
	c.frontendCode = frontendCode
	return c
}

// 设置错误信息
func (e ErrCode) WithMessage(message string) ErrCode {
	c := e.Clone()
	c.message = message
	return c
}

// 设置状态码
func (e ErrCode) WithStatusCode(code int) ErrCode {
	c := e.Clone()
	c.statusCode = code
	return c
}

// 实现errors的Is接口
func (e ErrCode) Is(target error) bool {
	return EqualError(e, target)
}

// 实现errors的As接口
func (e ErrCode) As(target interface{}) bool {
	targetVal := reflect.ValueOf(target)
	if reflect.TypeOf(e).AssignableTo(targetVal.Type().Elem()) {
		targetVal.Elem().Set(reflect.ValueOf(e))
		return true
	}

	return false
}

// 判断是否是公共错误码
func IsCommonErrCode(code int) bool {
	if code >= 2000000 {
		return false
	}
	return true
}

// 通过code及message获取错误码
// 若存在，则返回
// 若不存在，则新建
func GetOrNewErrCode(code int, msg string) ErrCode {
	if c, ok := _codes[code]; ok {
		return c
	}
	return ErrCode{
		code:         code,
		frontendCode: code,
		message:      msg,
		statusCode:   http.StatusInternalServerError,
	}
}

// 从错误字符串中解析错误码
func String(e string) ErrCode {
	if e == "" {
		return OK
	}

	data := strings.Split(e, ":")
	if len(data) == 0 {
		return InternalError
	}

	code, err := strconv.Atoi(data[0])
	if err != nil {
		return InternalError
	}

	if _, ok := _codes[code]; !ok {
		return InternalError
	}

	return _codes[code]
}
