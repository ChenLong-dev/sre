package middleware

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	_context "gitlab.shanhai.int/sre/library/base/context"
	_gin "gitlab.shanhai.int/sre/library/net/gin"
	"gitlab.shanhai.int/sre/library/net/tracing"
)

// 从请求头或查询参数中获取用户id并装入context中
// userIDKey为请求头/查询参数的键名
func GetUUIDMiddleware(userIDKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var id string

		if userIDKey != "" {
			id = c.GetHeader(userIDKey)
			if id == "" {
				id = c.Query(userIDKey)
			}
		}

		if id != "" {
			c.Set(_context.ContextUUIDKey, id)
		}

		c.Next()
	}
}

// 设置context中默认值中间件
func SetDefaultContextValueMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if id := c.GetString(_context.ContextUUIDKey); id == "" {
			id = c.GetString(tracing.TraceIDContextKey)
			if id == "" {
				id = uuid.NewV4().String()
			}
			c.Set(_context.ContextUUIDKey, id)
		}

		c.Set(_context.ContextRequestPathKey, _gin.GetGinRelativePath(c))
		c.Set(_context.ContextRequestMethodKey, c.Request.Method)

		c.Next()
	}
}
