package huawei

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/httpclient"

	"rulai/models/resp"
	_errcode "rulai/utils/errcode"
)

// 华为云 LTS API 需要处理的状态码(只取文档中有描述的部分)
var (
	StatusCodesFailures = []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusInternalServerError,
	}
	StatusCodesInMethodGetAndPut = append(StatusCodesFailures, http.StatusOK)
	StatusCodesInMethodPost      = append(StatusCodesFailures, http.StatusOK, http.StatusCreated)
	// 华为云 DELETE 方法没有统一成功时的状态码, 存在 200 和 204 两种情况, 200 时可能有返回值
	StatusCodesInMethodDelete = append(StatusCodesFailures, http.StatusOK, http.StatusNoContent)
)

// ltsErrorCodeWrapper 华为云日志服务(LTS)错误码包装器
// FIXME: 由于华为云目前的错误响应不统一, 并且与文档也很不一致, 暂时只能先采取将某些关键错误码包装成 AMS errcode 的方式处理返回值
// 等华为云日志服务完善后再考虑进行标准化
type ltsErrorCodeWrapper struct {
	code    resp.LTSErrorCode
	wrapper func(resp.HuaweiLTSResp) error
}

var (
	duplicateLogStreamNameWrapper = &ltsErrorCodeWrapper{
		code: resp.LTSDuplicateLogStreamName,
		wrapper: func(r resp.HuaweiLTSResp) error {
			return errors.Wrap(_errcode.LTSResourceAlreadyExists, toString(r))
		},
	}
	logStreamAssociatedByTransferWrapper = &ltsErrorCodeWrapper{
		code: resp.LTSLogStreamAssociatedByTransfer,
		wrapper: func(r resp.HuaweiLTSResp) error {
			return errors.Wrap(_errcode.LTSResourceAssociated, toString(r))
		},
	}
	logStreamNotFoundWrapper = &ltsErrorCodeWrapper{
		code: resp.LTSLogStreamNotExist,
		wrapper: func(r resp.HuaweiLTSResp) error {
			return errors.Wrap(_errcode.LTSResourceNotFound, toString(r))
		},
	}
	duplicateAOMMappingRuleNameWrapper = &ltsErrorCodeWrapper{
		code: resp.LTSDuplicateAOMMappingRuleName,
		wrapper: func(r resp.HuaweiLTSResp) error {
			return errors.Wrap(_errcode.LTSResourceAlreadyExists, toString(r))
		},
	}
	invalidAOMMappingRuleIDWrapper = &ltsErrorCodeWrapper{
		code: resp.LTSInvalidAOMMappingRuleID,
		wrapper: func(r resp.HuaweiLTSResp) error {
			return errors.Wrap(_errcode.LTSResourceNotFound, toString(r))
		},
	}
)

// doHuaweiAPIRequest 请求华为云 LTS API
// 华为云配置(目前有些功能 SDK 并不支持, 只能依靠配置参数进行 API 调用)
// NOTE: 当前使用 AK/SK 签名认证方式, 华为云限制消息体大小在 12M 以内, 未来如果有超过 12M 的请求, 必须改用 Token 认证方式
func (c *Controller) doHuaweiAPIRequest(ctx context.Context, method, url string,
	queryParams *httpclient.UrlValue, body, responseData interface{}, errResp resp.HuaweiLTSResp, wrappers ...*ltsErrorCodeWrapper) error {
	span := c.HTTPClient.Builder().URL(url).Method(method)

	switch method {
	case http.MethodGet, http.MethodPut:
		span = span.AccessStatusCode(StatusCodesInMethodGetAndPut...)

	case http.MethodPost:
		span = span.AccessStatusCode(StatusCodesInMethodPost...)

	case http.MethodDelete:
		span = span.AccessStatusCode(StatusCodesInMethodDelete...)

	default:
		return errors.Wrapf(errcode.InternalError, "unsupported LTS method(%s)", method)
	}

	if queryParams != nil {
		span = span.QueryParams(queryParams)
	}

	if body != nil {
		span = span.JsonBody(body)
	}

	// 添加华为云签名生成器
	span = span.AddHandler(func(sp *httpclient.Span) {
		err := c.APISigner.Sign(sp.Request)
		if err != nil {
			sp.SetError(err)
			return
		}
	})

	// 华为云必填 Header 配置
	headers := &httpclient.Header{Header: make(http.Header)}
	headers.Set("X-Sdk-Date", time.Now().String())
	// NOTE: 当前仅接入了 LTS 的 API, 不确定如果有其他 API 接入的话 Content-Type 设置是否会有区别
	headers.Set("Content-Type", "application/json;charset=utf8")
	span = span.Headers(headers)
	span.DisableTracing(true)

	res := span.Fetch(ctx)
	if res.Error() != nil {
		return errors.Wrapf(_errcode.HuaweiAPIInternalError, "request LTS API %s failed: %s", url, res.Error().Error())
	}

	switch res.StatusCode {
	case http.StatusCreated, http.StatusOK:
		// 某些 API 有标准的正确响应值, 但对应的错误码是非标准的 SVCSTG.ALS.200200
		// 该错误码未来会被替换, 故当前采取只按照 StatusCode 进行处理的方式
		if responseData == nil {
			return nil
		}

		err := res.DecodeJSON(responseData)
		if err != nil {
			return errors.Wrapf(_errcode.HuaweiAPIInternalError,
				"LTS API %s response: status(%d), parse JSON body error: %s", url, res.StatusCode, err.Error())
		}

	case http.StatusNoContent: // 没有 body 无需处理

	default:
		// 华为云的响应参数一方面没有统一, 另一方面也没有与文档做到一致, 暂时只能先以 body 的方式记录响应(有些请求有 request_id)
		defer res.Response.Body.Close()
		body, err := ioutil.ReadAll(res.Response.Body)
		if err != nil {
			return errors.Wrapf(errcode.InternalError,
				"LTS API %s response: unexpected status(%d), read body error(%s)", url, res.StatusCode, err)
		}

		err = json.Unmarshal(body, errResp)
		if err != nil {
			return errors.Wrapf(errcode.InternalError,
				"LTS API %s response: unexpected status(%d)", url, res.StatusCode)
		}

		errCode := errResp.GetErrorCode()
		for _, wrapper := range wrappers {
			if wrapper.code == errCode {
				return wrapper.wrapper(errResp)
			}
		}

		return errors.Wrapf(errcode.InternalError, "LTS API error response: request_id(%s), error_code(%s), error_msg(%s)",
			errResp.GetRequestID(), errCode, errResp.GetErrorMsg())
	}

	return nil
}
