package httpclient

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// 响应
type Response struct {
	*http.Response
	// 错误
	err error
}

// 新建Response
func NewResponse(resp *http.Response, err error) *Response {
	return &Response{
		Response: resp,
		err:      err,
	}
}

// 返回Error
func (resp *Response) Error() (err error) {
	if resp.err != nil && !errcode.EqualError(errcode.BreakerDegradedError, resp.err) {
		return resp.err
	} else {
		return nil
	}
}

// 是否已降级
func (resp *Response) IsDegraded() bool {
	if resp.err != nil && errcode.EqualError(errcode.BreakerDegradedError, resp.err) {
		return true
	}
	return false
}

// 解码json
func (resp *Response) DecodeJSON(value interface{}) (err error) {
	if err = resp.Error(); err != nil {
		return err
	}

	if value == nil {
		return errors.New("value is nil")
	}
	if reflect.ValueOf(value).IsNil() {
		return errors.New("reflect value is nil")
	}
	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		return errors.New("value is not ptr")
	}

	// 注意，此处不可使用resp.Body方法
	// 若使用，会在string与byte数组转换中产生大量内存消耗
	defer resp.Response.Body.Close()
	body, err := ioutil.ReadAll(resp.Response.Body)
	if err != nil {
		return
	}

	return json.Unmarshal(body, value)
}

func (resp *Response) Body() (body string, err error) {
	if err = resp.Error(); err != nil {
		return "", err
	}

	defer resp.Response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Response.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
