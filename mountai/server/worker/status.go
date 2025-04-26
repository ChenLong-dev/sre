package worker

import (
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/service"

	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	zipkintracer "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/uber/jaeger-client-go"
	framework "gitlab.shanhai.int/sre/app-framework"
	_context "gitlab.shanhai.int/sre/library/base/context"
	"gitlab.shanhai.int/sre/library/base/null"
	"gitlab.shanhai.int/sre/library/goroutine"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/tracing"
)

const (
	// Default schedule timeout 5s.
	defaultScheduleTimeout = time.Second * 5
)

func GetServer() framework.ServerInterface {
	svr := new(framework.JobServer)

	// ====================
	// >>>请勿删除<<<
	//
	// 根据实际情况修改
	// ====================
	// 设置任务函数
	svr.SetJob("status", func(ctx context.Context) error {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			// 根据应用id及环境获取未完成的任务
			// 两个集群的任务都需要处理，不指定集群名
			/**
			There are three situations:
			    1.Status is init: Task which status is not in "TaskStatusFinalStateList".
				2.Status isn't init:
					1): approved immediate tasks.
					2): Scheduled tasks and there schedules are just smaller than the current time.
			*/
			notFinalAndInitTasks, err := service.SVC.GetTasksGroupByAppIDAndEnvName(ctx, &req.GetTasksReq{
				StatusInverseList: entity.TaskStatusFinalAndInitList,
				Suspend:           null.BoolFrom(false),
			})
			if err != nil {
				log.Errorc(ctx, "get task error:%s", err)
				continue
			}
			immediateTasks, err := service.SVC.GetTasksGroupByAppIDAndEnvName(ctx, &req.GetTasksReq{
				StatusList:         []entity.TaskStatus{entity.TaskStatusInit},
				Suspend:            null.BoolFrom(false),
				DeployTypeList:     entity.TaskDeployTypeImmediateList,
				ApprovalStatusList: entity.TaskApprovalStatusApprovedList,
			})
			if err != nil {
				log.Errorc(ctx, "get task error:%s", err)
				continue
			}
			now := time.Now()
			scheduledTasks, err := service.SVC.GetTasksGroupByAppIDAndEnvName(ctx, &req.GetTasksReq{
				StatusList:         []entity.TaskStatus{entity.TaskStatusInit},
				Suspend:            null.BoolFrom(false),
				DeployTypeList:     entity.TaskDeployTypeScheduledList,
				MaxScheduleTime:    now.Unix(),
				MinScheduleTime:    now.Add(-defaultScheduleTimeout).Unix(),
				ApprovalStatusList: entity.TaskApprovalStatusApprovedList,
			})
			if err != nil {
				log.Errorc(ctx, "get task error:%s", err)
				continue
			}

			tasks := make([]*resp.TaskDetailResp, 0)
			tasks = append(tasks, notFinalAndInitTasks...)
			tasks = append(tasks, immediateTasks...)
			tasks = append(tasks, scheduledTasks...)

			wg := goroutine.New("status-worker")
			for _, task := range tasks {
				curTask := task
				curCtx, span := getTracingContext(ctx, fmt.Sprintf("[Worker] %s", curTask.ID))

				wg.Go(curCtx, curTask.ID, func(ctx context.Context) error {
					defer func() {
						span.Finish()
					}()

					e := service.SVC.TransformTaskStatus(ctx, curTask)
					if e != nil {
						log.Errorc(ctx, "transform task error: id:%s error:%+v", curTask.ID, e)
						// 出错则更新任务重试次数及日志
						_ = service.SVC.IncreaseTaskRetryCount(ctx, curTask.ID, e.Error())
					}
					return nil
				})
			}
			err = wg.Wait()
			if err != nil {
				log.Errorc(ctx, "transform task error: %+v", err)
			}
		}
		return nil
	})

	return svr
}

func getTracingContext(ctx context.Context, spanName string) (context.Context, opentracing.Span) {
	span := opentracing.GlobalTracer().StartSpan(spanName)
	curCtx := tracing.SetCurrentSpanToContext(ctx, span)

	var traceID string
	switch sc := span.Context().(type) {
	case zipkintracer.SpanContext:
		traceID = sc.TraceID.String()
	case jaeger.SpanContext:
		traceID = sc.TraceID().String()
	}
	curCtx = context.WithValue(
		curCtx,
		_context.ContextUUIDKey,
		traceID,
	)

	return curCtx, span
}
