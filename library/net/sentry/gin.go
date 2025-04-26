package sentry

import (
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	_context "gitlab.shanhai.int/sre/library/base/context"
)

const (
	// gin默认刷新超时时间
	GinDefaultFlushTimeout = time.Second * 2
)

// gin中间件配置
type GinOption struct {
	// 是否捕获
	Recover bool
	// 事件发送的超时时间
	Timeout time.Duration
}

func GinMiddleware(option *GinOption) gin.HandlerFunc {
	if option == nil {
		option = &GinOption{}
	}
	if option.Timeout == 0 {
		option.Timeout = GinDefaultFlushTimeout
	}

	return sentrygin.New(sentrygin.Options{
		Repanic: !option.Recover,
		Timeout: option.Timeout,
	})
}

func GlobalTagsMiddleware(globalTags map[string]string) gin.HandlerFunc {
	if globalTags == nil {
		globalTags = make(map[string]string)
	}

	return func(ctx *gin.Context) {
		if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
			// 新建map，避免并发冲突
			tags := make(map[string]string)
			for k, v := range globalTags {
				tags[k] = v
			}
			// 增加默认tag
			tags[_context.ContextRequestPathKey] = _context.GetString(ctx, _context.ContextRequestPathKey)
			tags[_context.ContextRequestMethodKey] = _context.GetString(ctx, _context.ContextRequestMethodKey)
			hub.Scope().SetTags(tags)
			// 增加uuid
			hub.Scope().SetExtra(_context.ContextUUIDKey, _context.GetStringOrDefault(ctx, _context.ContextUUIDKey, "unknown"))

			hub.Scope().SetLevel(sentry.LevelFatal)
		}
		ctx.Next()
	}
}
