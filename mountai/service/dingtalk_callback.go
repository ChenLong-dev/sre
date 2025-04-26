package service

import (
	"rulai/models/entity"
	"rulai/models/req"
	_errcode "rulai/utils/errcode"

	"context"
	"encoding/json"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// HandleDingCallbackMsgEvent handles ding callback message event.
func (s *Service) HandleDingCallbackMsgEvent(ctx context.Context, msg *sarama.ConsumerMessage) error {
	message := new(entity.DingCallbackMsg)
	err := json.Unmarshal(msg.Value, &message)
	if err != nil {
		return errors.Wrap(errcode.InternalError, err.Error())
	}

	// Message is effective when it satisfies the following conditions:
	// 1.callback event type is equal to "bpms_instance_change".
	// 2.the approval has already finished.
	// 3.the approval was agreed.
	if message.Data.EventType != entity.BPMSInstanceChange || message.Data.Type != entity.FinishDingCallbackType {
		return nil
	}

	approvingTasks, err := s.GetTasks(ctx, &req.GetTasksReq{
		ApprovalInstanceID: message.Data.ProcessInstanceID,
		ApprovalStatusList: entity.TaskApprovalStatusApprovingList,
	})
	if err != nil {
		return err
	}
	// We Discard this message when approval process instance id is not belong to deployment approval task.
	if len(approvingTasks) == 0 {
		return nil
	}
	if len(approvingTasks) > 1 {
		return errors.Wrapf(_errcode.DingTalkCallbackError, "approving task is not unique, instance id: %s",
			message.Data.ProcessInstanceID)
	}

	taskID := ""
	processInst, err := s.GetDingApprovalInstance(ctx, message.Data.ProcessInstanceID)
	if err != nil {
		return err
	}

	for _, form := range processInst.FormComponentValues {
		if form.Name == entity.BPMSInstanceChangeCallbackKeyTaskID && form.Value != "" {
			taskID = strings.TrimSpace(form.Value)
			break
		}
	}
	if taskID == "" {
		return errors.Wrap(_errcode.DingTalkCallbackError, "callback message taskID does not exist")
	}

	approvalStatus := entity.ApprovedTaskApprovalStatus
	if message.Data.Result == entity.RefuseDingCallbackResult {
		approvalStatus = entity.RefusedTaskApprovalStatus
	}

	task, err := s.GetTaskDetail(ctx, taskID)
	if err != nil {
		return err
	}
	app, err := s.GetAppDetail(ctx, task.AppID)
	if err != nil {
		return err
	}
	project, err := s.GetProjectDetail(ctx, app.ProjectID)
	if err != nil {
		return err
	}

	return s.UpdateTask(ctx, project, app, task, &req.UpdateTaskReq{
		ApprovalStatus: approvalStatus,
	})
}
