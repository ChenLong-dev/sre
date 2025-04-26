package service

import (
	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/httpclient"
)

const (
	taskDeployAlterationMsg = `
### {{.Title}}
>- [项目名] {{.ProjectName}}
>- [应用名] {{.AppName}}
>- [环境名] {{.Env}}
>- [版本] {{.Version}}
>- [操作人] {{.UserName}}
>- [原部署方式] {{.OldDeploy}}
{{if .OldDeployTime}}
>- [原部署时间] {{.OldDeployTime}}
{{end}}
>- [现部署方式] {{.CurDeploy}}
{{if .CurDeployTime}}
>- [现部署时间] {{.CurDeployTime}}
{{end}}
#### [进入AMS, 查看详情]({{.DetailURL}})
`
	taskDeployAlterationTmpName = "taskDeployUpdate"
)

// SendCropMessage 发送工作通知
func (s *Service) SendCropMessage(ctx context.Context, msgReq *req.SendCropMessageReq) (data *resp.DingTalkCropMessageData, err error) {
	res := new(resp.DingTalkCropMessageResp)
	err = s.httpClient.Builder().
		URL(fmt.Sprintf("%s/v1/messages", config.Conf.DingTalk.Host)).
		QueryParams(
			httpclient.NewUrlValue().
				Add("app", config.Conf.DingTalk.App).
				Add("token", config.Conf.DingTalk.Token),
		).
		JsonBody(msgReq).
		Method(http.MethodPost).
		Fetch(ctx).
		DecodeJSON(&res)

	if err != nil {
		return nil, errors.Wrap(_errcode.DingTalkInternalError, err.Error())
	} else if res.ErrCode != 0 {
		return nil, errors.Wrap(_errcode.DingTalkInternalError, res.ErrMsg)
	}

	log.Infoc(ctx, "send crop message success: taskid[%d], requestid[%s]", res.Data.TaskID, res.Data.RequestID)

	return res.Data, nil
}

// SendRobotMsgToDingTalk send dingDing message
func (s *Service) SendRobotMsgToDingTalk(ctx context.Context, msgReq *req.SendRobotMessageReq) (*resp.DingTalkCommonResp, error) {
	res := new(resp.DingTalkCommonResp)

	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/v1/messages/robot", config.Conf.DingTalk.Host)).
		QueryParams(
			httpclient.NewUrlValue().
				Add("app", config.Conf.DingTalk.App).
				Add("token", config.Conf.DingTalk.Token),
		).
		Method(http.MethodPost).
		JsonBody(msgReq).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.DingTalkInternalError, err.Error())
	} else if res.ErrCode != 0 {
		return nil, errors.Wrap(_errcode.DingTalkInternalError, res.ErrMsg)
	}

	return res, nil
}

// SendRobotMsgToFeishu send feishu message
func (s *Service) SendRobotMsgToFeishu(ctx context.Context, msgReq *req.FeishuInteractiveMessageReq) error {
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/message/send", config.Conf.Feishu.Host)).
		Method(http.MethodPost).
		JsonBody(msgReq).
		Fetch(ctx).Error()
	if err != nil {
		return errors.Wrap(_errcode.FeishuInternalError, err.Error())
	}

	return nil
}

// generateRobotMessageReq generate robot message request
func (s *Service) generateRobotMessageReq(appOpMessage *req.AppOpMessage) (msg *req.FeishuInteractiveMessageReq, err error) {
	text, err := s.RenderTemplate(context.TODO(), "./template/feishu/InteractiveContent.template", appOpMessage)
	if err != nil {
		return nil, err
	}

	return &req.FeishuInteractiveMessageReq{
		Msgtype:   "interactive",
		ReceiveId: config.Conf.Feishu.DeployNotiChatID,
		Content:   text,
	}, nil
}

// generateCropMessageReq generate crop message request
func (s *Service) generateCropMessageReq(appOpMsgReq *req.AppOpMessage, emails []string,
	tmpName, parseText string) (msg *req.SendCropMessageReq, err error) {
	text, err := s.RenderTemplateFromText(appOpMsgReq, tmpName, parseText)
	if err != nil {
		return nil, err
	}

	return &req.SendCropMessageReq{
		EmailList: emails,
		Msg: req.DingTalkMarkdownMessageReq{
			Msgtype: req.DingTalkMsgTypeMarkdown,
			Markdown: req.DingTalkMarkdownMessageBody{
				Title: appOpMsgReq.Title,
				Text:  text,
			},
		},
	}, nil
}

func (s *Service) CreateDingTalkMsgRecord(ctx context.Context, dingRecord *entity.DingTalkMessageRecord) error {
	_, err := s.dao.CreateDingTalkMessageRecord(ctx, dingRecord)

	return err
}

// CreateDingApprovalInstance creates ding approval process instance.
func (s *Service) CreateDingApprovalInstance(ctx context.Context,
	createReq *req.CreateDingApprovalInstanceReq) (*resp.CreateDingApprovalInstanceResp, error) {
	res := new(resp.CreateDingApprovalInstanceResp)
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/v1/process_instances", config.Conf.DingTalk.Host)).
		Method(http.MethodPost).
		JsonBody(createReq).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.DingTalkInternalError, err.Error())
	}
	if res.ErrCode != 0 {
		return nil, errors.Wrap(_errcode.DingTalkInternalError, res.ErrMsg)
	}
	if res.Data.ErrCode != 0 {
		return nil, errors.Wrap(_errcode.DingTalkInternalError, res.Data.ErrMsg)
	}

	return res, nil
}

// TerminateDingApprovalInstance terminates ding approval process instance.
func (s *Service) TerminateDingApprovalInstance(ctx context.Context, instanceID string,
	terminateReq *req.TerminateDingApprovalInstanceReq) error {
	res := new(resp.DingTalkCommonResp)
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/v1/process_instances/%s/terminate", config.Conf.DingTalk.Host, instanceID)).
		Method(http.MethodPost).
		JsonBody(terminateReq).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return errors.Wrap(_errcode.DingTalkInternalError, err.Error())
	}
	if res.ErrCode != 0 {
		return errors.Wrap(_errcode.DingTalkInternalError, res.ErrMsg)
	}

	return nil
}

// GetDingApprovalInstance gets ding approval process instance detail.
func (s *Service) GetDingApprovalInstance(ctx context.Context, instanceID string) (*resp.DingApprovalInstanceRespData, error) {
	res := new(resp.GetDingApprovalInstanceResp)
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/v1/process_instances/%s", config.Conf.DingTalk.Host, instanceID)).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.DingTalkInternalError, err.Error())
	}
	if res.ErrCode != 0 {
		return nil, errors.Wrap(_errcode.DingTalkInternalError, res.ErrMsg)
	}

	return res.Data, nil
}

func (s *Service) generateApproversIDs(projectApprovers, taskApprovers []*entity.DingDingUserDetail) []string {
	approvers := projectApprovers
	if len(taskApprovers) > 0 {
		approvers = taskApprovers
	}
	res := make([]string, 0)
	for _, approver := range approvers {
		res = append(res, approver.ID)
	}

	return res
}

func (s *Service) getDingApproverReq(userIDs []string) *req.DingApprover {
	taskAction := entity.NONEDingApprovalTaskActionType
	if len(userIDs) > 1 {
		taskAction = entity.ORDingApprovalTaskActionType
	}
	return &req.DingApprover{
		UserIDs:        userIDs,
		TaskActionType: taskAction,
	}
}

func (s *Service) generateDingApprovalReq(ctx context.Context, task *entity.Task,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) (*req.CreateDingApprovalInstanceReq, error) {
	operator := ctx.Value(utils.ContextInternalUserKey).(entity.InternalUser)

	approvers := make([]*req.DingApprover, 0)
	// Adds project owners to approvers.
	ownerIDs := make([]string, 0)
	for _, owner := range project.Owners {
		internalUser, e := s.GetInternalSingleUser(ctx, &req.GetInternalUsersReq{
			Email: owner.Email,
		})
		if e != nil {
			return nil, e
		}
		ownerIDs = append(ownerIDs, internalUser.DingTalkUserID)
	}

	// Add approvers.
	approvers = append(approvers, s.getDingApproverReq(ownerIDs),
		s.getDingApproverReq(s.generateApproversIDs(project.OperationEngineers, task.Approval.OperationEngineers)),
		s.getDingApproverReq(s.generateApproversIDs(project.QAEngineers, task.Approval.QAEngineers)))

	return &req.CreateDingApprovalInstanceReq{
		ProcessCode: config.Conf.DingTalk.Approval.ProcessCode,
		CCList: strings.Join(s.generateApproversIDs(project.ProductManagers,
			task.Approval.ProductManagers), ","),
		DeptID:              strconv.Itoa(operator.Departments[0].ID),
		Approvers:           approvers,
		OriginatorUserID:    operator.DingTalkUserID,
		FormComponentValues: s.generateDingApprovalFormValues(ctx, task, project, app),
	}, nil
}

// generateDingApprovalFormValues generates ding approval form values.
func (s *Service) generateDingApprovalFormValues(_ context.Context, task *entity.Task, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp) []*req.ApprovalFormComponentValue {
	formValues := make([]*req.ApprovalFormComponentValue, 0)

	formValuesMap := make(map[string]string)
	formValuesMap["项目名称"] = project.Name
	formValuesMap["项目ID"] = project.ID

	ownersBuilder := strings.Builder{}
	for _, owner := range project.Owners {
		ownersBuilder.WriteString(fmt.Sprintf("%s ", owner.Name))
	}
	formValuesMap["项目负责人"] = ownersBuilder.String()

	formValuesMap["团队"] = project.Team.Name
	formValuesMap["环境"] = string(task.EnvName)
	formValuesMap["项目地址"] = s.GetAmsFrontendProjectURL(project.ID, task.EnvName)
	formValuesMap["应用名称"] = app.Name
	formValuesMap["应用类型"] = string(app.Type)
	formValuesMap["部署方式"] = entity.GetTaskDeployTypeDisplay(task.DeployType)

	if task.ScheduleTime != nil {
		formValuesMap["计划部署时间"] = task.ScheduleTime.Format(utils.DefaultTimeFormatLayout)
	}

	formValuesMap["发布描述"] = task.Description
	formValuesMap["镜像版本"] = task.Param.ImageVersion
	formValuesMap["配置commit ID"] = task.Param.ConfigCommitID
	formValuesMap["taskID"] = task.ID.Hex()
	formValuesMap["最小实例数"] = strconv.Itoa(task.Param.MinPodCount)
	formValuesMap["最大实例数"] = strconv.Itoa(task.Param.MaxPodCount)
	formValuesMap["cpu规格"] = fmt.Sprintf("%s 核", task.Param.CPURequest)
	formValuesMap["cpu限制"] = fmt.Sprintf("%s 核", task.Param.CPULimit)
	formValuesMap["内存规格"] = string(task.Param.MemRequest)
	formValuesMap["内存限制"] = string(task.Param.MemLimit)

	for name, value := range formValuesMap {
		formValues = append(formValues, &req.ApprovalFormComponentValue{
			Name:  name,
			Value: value,
		})
	}

	return formValues
}

// GetInternalUser gets internal user with user id.
func (s *Service) GetInternalUser(ctx context.Context, userID string) (*entity.InternalUser, error) {
	userInfo, err := s.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}
	internalUser, err := s.GetInternalSingleUser(ctx, &req.GetInternalUsersReq{
		Email: userInfo.Email,
	})
	if err != nil {
		return nil, err
	}

	return internalUser, nil
}

func (s *Service) SendUrgentDeployDingCropMessage(ctx context.Context, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, createReq *req.CreateTaskReq) (*resp.DingTalkCropMessageData, error) {
	operator := ctx.Value(utils.ContextInternalUserKey).(entity.InternalUser)

	emails := make([]string, 0)
	for _, owner := range project.Owners {
		emails = append(emails, owner.Email)
	}

	ret, err := s.SendDingCropMessage(ctx, &req.AppOpMessage{
		Title:       fmt.Sprintf("%s 被 %s 紧急部署", project.Name, operator.Nickname),
		ProjectName: project.Name,
		AppName:     app.Name,
		Env:         string(createReq.EnvName),
		Action:      entity.GetTaskActionDisplay(createReq.Action),
		OpTime:      time.Now().Format(utils.DefaultTimeFormatLayout),
		UserName:    operator.Nickname,
		DetailURL:   s.GetAmsFrontendProjectURL(project.ID, createReq.EnvName),
	}, emails, entity.UrgentDeployTmpName, entity.UrgentDeployMsg)
	return ret, err
}

// CreateDingTalkUrgentDeployMsgRecord creates ding urgent deployment message record.
func (s *Service) CreateDingTalkUrgentDeployMsgRecord(ctx context.Context, dingRecord *entity.UrgentDeploymentDingTalkMsgRecord) error {
	_, err := s.dao.CreateUrgentDeployDingTalkMsgRecord(ctx, dingRecord)

	return err
}

// SendDingCropMessage sends ding crop message.
func (s *Service) SendDingCropMessage(ctx context.Context, msg *req.AppOpMessage, emails []string,
	tmpName, parseText string) (*resp.DingTalkCropMessageData, error) {
	cropMsgReq, err := s.generateCropMessageReq(msg, emails, tmpName, parseText)
	if err != nil {
		return nil, err
	}

	res, err := s.SendCropMessage(ctx, cropMsgReq)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) SendTaskDeployAlterationDingCropMessage(ctx context.Context, msg *req.TaskDeployAlterationMessage,
	emails []string, tmpName, parseText string) (*resp.DingTalkCropMessageData, error) {
	text, err := s.RenderTemplateFromText(msg, tmpName, parseText)
	if err != nil {
		return nil, err
	}

	res, err := s.SendCropMessage(ctx, &req.SendCropMessageReq{
		EmailList: emails,
		Msg: req.DingTalkMarkdownMessageReq{
			Msgtype: req.DingTalkMsgTypeMarkdown,
			Markdown: req.DingTalkMarkdownMessageBody{
				Title: msg.Title,
				Text:  text,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
