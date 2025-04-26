package handler

import (
	"context"
	v1 "gitlab.shanhai.int/sre/app-framework/internal/example/api/common/v1"
)

// 入口处理器
type EntryHandler struct {
}

// 示例接口
func (h *EntryHandler) HelloWorld(context.Context, *v1.HelloWorldRequest) (*v1.HelloWorldResponse, error) {
	return &v1.HelloWorldResponse{
		Country: "China",
	}, nil
}
