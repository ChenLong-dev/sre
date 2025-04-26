package service

import (
	"context"
	"fmt"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/req"
	"gitlab.shanhai.int/sre/library/goroutine"
	"gitlab.shanhai.int/sre/library/log"
	"strconv"
	"time"
)

// 调用restful
func (s *Service) InvokeRestful(ctx context.Context, invokeReq req.InvokeRestfulReq) (err error) {
	// ==========================
	// 声明协程组
	// ==========================
	wg := goroutine.New(fmt.Sprintf("%s-%s", invokeReq.Method.ValueOrZero(), invokeReq.Url.ValueOrZero()))
	// ==========================
	// 循环启动协程
	// ==========================
	for i := 0; i < invokeReq.GoroutineCount.ValueOrZero(); i++ {
		curNum := i
		wg.Go(ctx, strconv.Itoa(curNum), func(ctx context.Context) error {
			for j := 0; j < invokeReq.Times.ValueOrZero(); j++ {
				time.Sleep(time.Millisecond * time.Duration(invokeReq.SleepMilliseconds.ValueOrZero()))
				// 调用http客户端获取数据
				err := s.httpClient.Builder().
					Method(invokeReq.Method.ValueOrZero()).
					URL(invokeReq.Url.ValueOrZero()).
					Fetch(ctx).
					Error()
				if err != nil {
					log.Errorc(ctx, err.Error())
				}
			}
			return nil
		})
	}
	// ==========================
	// 该接口为异步操作，无需等待所有协程执行完毕
	// ==========================
	return nil
}
