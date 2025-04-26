package req

import "rulai/models/entity"

type DingTalkMsgType string

const DingTalkMsgTypeMarkdown DingTalkMsgType = "markdown"

type DingTalkMarkdownMessageReq struct {
	Msgtype  DingTalkMsgType             `json:"msgtype"`
	Markdown DingTalkMarkdownMessageBody `json:"markdown"`
}

type RobotMarkdownMessageReq struct {
	DingTalkMarkdownMessageReq
	At RobotMarkdownMessageAt `json:"at"`
}

type DingTalkMarkdownMessageBody struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type RobotMarkdownMessageAt struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll"`
}

type BranchChangeRobotMessage struct {
	Title            string `json:"title"`
	ProjectName      string `json:"project_name"`
	AppName          string `json:"app_name"`
	Env              string `json:"env"`
	OriginBranch     string `json:"origin_branch"`
	CurrentBranch    string `json:"current_branch"`
	OperatorName     string `json:"operator_name"`
	SecurityKeywords string `json:"security_keywords"`
	DetailURL        string `json:"detail_url"`
}

// SendCropMessageReq 工作通知
type SendCropMessageReq struct {
	EmailList []string                   `json:"email_list"`
	Msg       DingTalkMarkdownMessageReq `json:"msg"`
}

// SendRobotMessageReq 机器人消息
type SendRobotMessageReq struct {
	Token   string                  `json:"token"`
	Message RobotMarkdownMessageReq `json:"message"`
}

// AppOpMessage app操作钉钉通知消息
type AppOpMessage struct {
	Title           string `json:"title"`
	ProjectName     string `json:"project_name"`
	AppName         string `json:"app_name"`
	UserName        string `json:"user_name"`
	Env             string `json:"env"`
	Action          string `json:"action"`
	OpTime          string `json:"op_time"`
	DetailURL       string `json:"detail_url"`
	ProjectLanguage string `json:"project_language"`
	TeamName        string `json:"team_name"`
	Branch          string `json:"branch"`
}

type CreateDingApprovalInstanceReq struct {
	ProcessCode         string                        `form:"processCode" json:"processCode"`
	CCList              string                        `form:"ccList" json:"ccList"`
	DeptID              string                        `form:"deptId" json:"deptId"`
	Approvers           []*DingApprover               `form:"approver" json:"approver"`
	OriginatorUserID    string                        `form:"originatorUserId" json:"originatorUserId"`
	FormComponentValues []*ApprovalFormComponentValue `form:"formComponentValues" json:"formComponentValues"`
}

type DingApprover struct {
	UserIDs        []string                          `form:"user_ids" json:"user_ids"`
	TaskActionType entity.DingApprovalTaskActionType `form:"task_action_type" json:"task_action_type"`
}

type ApprovalFormComponentValue struct {
	Name  string `form:"name" json:"name"`
	Value string `form:"value" json:"value"`
}

type TerminateDingApprovalInstanceReq struct {
	OperatingUserID string `form:"operating_userid" json:"operating_userid"`
}

type TaskDeployAlterationMessage struct {
	Title         string `json:"title"`
	ProjectName   string `json:"project_name"`
	AppName       string `json:"app_name"`
	UserName      string `json:"user_name"`
	Env           string `json:"env"`
	Version       string `json:"version"`
	OldDeploy     string `json:"old_deploy"`
	OldDeployTime string `json:"old_deploy_time"`
	CurDeploy     string `json:"cur_deploy"`
	CurDeployTime string `json:"cur_deploy_time"`
	DetailURL     string `json:"detail_url"`
}
