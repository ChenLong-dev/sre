package job

import (
	"context"
	"gitlab.shanhai.int/sre/app-framework"
	"gitlab.shanhai.int/sre/app-framework/internal/example/service"
	"time"
)

// ====================
// >>>请勿删除<<<
//
// 获取任务服务器
// ====================
func GetServer() framework.ServerInterface {
	svr := new(framework.JobServer)

	// ====================
	// >>>请勿删除<<<
	//
	// 根据实际情况修改
	// ====================
	// 设置任务函数
	svr.SetJob("random-number", func(ctx context.Context) error {
		for i := 0; i < 30; i++ {
			num := i
			// 生成随机数
			err := service.SVC.PrintRandomNumber(ctx, num)
			// 若返回错误，则任务结束
			if err != nil {
				return err
			}
			time.Sleep(time.Second)
		}
		return nil
	})

	return svr
}
