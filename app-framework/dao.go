package framework

import "context"

// 数据层接口
type DaoInterface interface {
	// 关闭数据层
	Close(c context.Context)
}
