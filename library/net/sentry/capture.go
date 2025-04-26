package sentry

import (
	"context"
	"github.com/getsentry/sentry-go"
	"time"
)

const (
	// errCode标签
	TagErrorCode = "errCode"
	// 内部ip标签
	TagInternalIP = "internal_ip"

	// 额外信息msg
	ExtraMessage = "message"
	// 额外信息errMessage
	ExtraErrorMsg = "errMessage"

	// 默认刷新超时时间
	DefaultFlushTimeout = time.Second * 10

	// context中hub的key
	ContextKey = "sentry"
)

// 标签
type Tag struct {
	Key   string
	Value string
}

// 捕获并打标签
func CaptureWithTags(ctx context.Context, err error, tags ...Tag) {
	if !IsInit() {
		return
	}

	hub := getOrCloneHubFromCtx(ctx)
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
		tagsMap := map[string]string{}
		for _, t := range tags {
			tagsMap[t.Key] = t.Value
		}
		scope.SetTags(tagsMap)
	})
	hub.CaptureException(err)
}

// 面包屑额外信息
type Breadcrumb struct {
	// 分类
	Category string
	// 额外信息map
	Data map[string]interface{}
}

// 捕获并加面包屑和标签
func CaptureWithBreadAndTags(ctx context.Context, err error, breadcrumb *Breadcrumb, tags ...Tag) {
	if !IsInit() {
		return
	}

	hub := getOrCloneHubFromCtx(ctx)
	hub.ConfigureScope(func(scope *sentry.Scope) {
		tagsMap := map[string]string{}
		for _, t := range tags {
			tagsMap[t.Key] = t.Value
		}
		scope.SetTags(tagsMap)

		scope.AddBreadcrumb(&sentry.Breadcrumb{
			Category: breadcrumb.Category,
			Data:     breadcrumb.Data,
		}, MaxBreadcrumbs)

		scope.SetLevel(sentry.LevelError)
	})

	hub.CaptureException(err)
}

// 增加面包屑
func AddBreadcrumb(ctx context.Context, breadcrumb *Breadcrumb) (newCtx context.Context) {
	if !IsInit() {
		return ctx
	}

	hub := GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
		newCtx = SetHubOnContext(ctx, hub)
	} else {
		newCtx = ctx
	}

	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Category: breadcrumb.Category,
		Data:     breadcrumb.Data,
	}, nil)

	return
}

// 捕获字符串
func CaptureMessage(ctx context.Context, msg string) {
	if !IsInit() {
		return
	}

	hub := getOrCloneHubFromCtx(ctx)
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
		scope.SetExtra(ExtraMessage, msg)
	})

	hub.CaptureMessage(msg)
}

// 异常恢复
func CapturePanic(ctx context.Context, f func()) {
	hub := getOrCloneHubFromCtx(ctx)
	hub.Scope().SetLevel(sentry.LevelFatal)
	defer func() {
		if err := recover(); err != nil {
			eventID := hub.RecoverWithContext(ctx, err)
			if eventID != nil {
				hub.Flush(DefaultFlushTimeout)
			}

			panic(err)
		}
	}()

	f()
}

// 根据context获取hub
func getOrCloneHubFromCtx(ctx context.Context) (hub *sentry.Hub) {
	hub = GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}

	return
}

// context中获取hub
func GetHubFromContext(ctx context.Context) *sentry.Hub {
	if hub, ok := ctx.Value(ContextKey).(*sentry.Hub); ok {
		return hub
	}
	return nil
}

// 设置hub到context中
func SetHubOnContext(ctx context.Context, hub *sentry.Hub) context.Context {
	return context.WithValue(ctx, ContextKey, hub)
}

// 从context中克隆hub并生成子context
func CloneHubOnContext(ctx context.Context) context.Context {
	hub := GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}
	return context.WithValue(ctx, ContextKey, hub.Clone())
}
