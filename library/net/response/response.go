package response

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	_errors "github.com/pkg/errors"
	_context "gitlab.shanhai.int/sre/library/base/context"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/sentry"
)

// 带错误码的响应接口
type ErrCodeResponse interface {
	PrintError(ctx context.Context, err error)

	WithErrCode(code errcode.Codes) interface{}

	GetStatusCode(code errcode.Codes) int
}

// 自定义返回json
// 自定义结构体需要实现ErrCodeResponse接口
func InjectJson(ctx *gin.Context, data ErrCodeResponse, err error) {
	if err != nil {
		sentry.CaptureWithTags(ctx, err)
		data.PrintError(ctx, err)
	}

	// 防止超时请求
	select {
	case <-ctx.Request.Context().Done():
		return
	default:
	}

	// 适配fmt.Errorf及errors.Wrapf方法
	var code errcode.Codes
	if !errors.As(err, &code) {
		code = errcode.Cause(err)
	}
	ctx.Set(_context.ContextErrCode, code.Code())

	statusCode := data.GetStatusCode(code)

	// 渲染
	ctx.Render(statusCode, render.JSON{
		Data: data.WithErrCode(code),
	})
}

// 返回json
// Deprecated
// 方法已废弃，推荐使用 StandardJSON
func JSON(ctx *gin.Context, data interface{}, err error) {
	InjectJson(
		ctx,
		&V1Response{
			Data: data,
		},
		err,
	)
}

// 返回非法请求参数的json
// Deprecated
// 方法已废弃，推荐使用 StandardInvalidParamsJSON
func InvalidParamsJSON(ctx *gin.Context, data interface{}, err error) {
	JSON(ctx, data, _errors.Wrap(errcode.InvalidParams, err.Error()))
}

// 返回json
func StandardJSON(ctx *gin.Context, data interface{}, err error) {
	InjectJson(
		ctx,
		&V2Response{
			Data: data,
		},
		err,
	)
}

// 返回非法请求参数的json
func StandardInvalidParamsJSON(ctx *gin.Context, data interface{}, err error) {
	StandardJSON(ctx, data, _errors.Wrap(errcode.InvalidParams, err.Error()))
}

// 返回分页的json
func StandardPaginationJSON(ctx *gin.Context, total, page, size int, items interface{}, err error) {
	StandardJSON(ctx, &V2PaginationResponse{
		Total:    total,
		Page:     page,
		PageSize: size,
		Items:    items,
	}, err)
}
