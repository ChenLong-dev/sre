package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/response"
)

// 捕获全局panic
// 使用默认response及日志打印
func CatchPanicMiddleware() gin.HandlerFunc {
	return CustomCatchPanicMiddleware(func(c *gin.Context, err interface{}) {
		if e, ok := err.(error); ok {
			response.StandardJSON(c, nil, e)
		} else {
			e := errors.Wrapf(errcode.InternalError, "%s", err)
			response.StandardJSON(c, nil, e)
		}
		c.Abort()
	})
}

// 自定义捕获全局panic
// recoverFunc 为恢复异常后调用的函数
func CustomCatchPanicMiddleware(recoverFunc func(c *gin.Context, err interface{})) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					recoverFunc(c, r)
				} else {
					errorWithStack := errors.WithStack(err)
					recoverFunc(c, errorWithStack)
				}
			}
		}()

		c.Next()
	}
}
