package goroutine

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"gitlab.shanhai.int/sre/library/base/hook"
)

var (
	// 钩子管理器
	_manager *hook.Manager
)

// 新建协程组
func newWithOptions(name string, options ...Option) *ErrGroup {
	eg := &ErrGroup{
		uuid: uuid.NewV4().String(),
		name: name,
	}

	for _, option := range options {
		option.Apply(eg)
	}

	return eg
}

// 新建协程组
//
//	当某个协程报错时，不会取消其他协程
func New(name string, options ...Option) *ErrGroup {
	options = append([]Option{SetNormalMode()}, options...)
	return newWithOptions(name, options...)
}

// 通过context创建协程组
//
//	当某个协程报错时，会调用cancel，取消其他协程
func WithContext(ctx context.Context, name string) *ErrGroup {
	return newWithOptions(name, SetCancelMode(ctx))
}
