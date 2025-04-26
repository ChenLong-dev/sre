package entity

// Ding callback event type.
type DingCallbackEventType string

const (
	// Ding approval callback event.
	BPMSInstanceChange DingCallbackEventType = "bpms_instance_change"
)

// Ding callback type.
type DingApprovalCallbackPhaseType string

const (
	// Ding start approval callback phase.
	StartDingCallbackType DingApprovalCallbackPhaseType = "start"
	// Ding finish approval callback phase.
	FinishDingCallbackType DingApprovalCallbackPhaseType = "finish"
	// Ding cancel approval callback phase.
	CancelDingCallbackType DingApprovalCallbackPhaseType = "cancel"
)

// Ding approval callback result.
type DingApprovalCallbackResult string

const (
	// Agree approval callback result.
	AgreeDingCallbackResult DingApprovalCallbackResult = "agree"
	// Refuse approval callback result.
	RefuseDingCallbackResult DingApprovalCallbackResult = "refuse"
)

const (
	// Ding deployment approval process form key "taskID"
	BPMSInstanceChangeCallbackKeyTaskID = "taskID"
)

// Ding approval creation param "task_action_type".
type DingApprovalTaskActionType string

const (
	// Ding approval task_action_type `AND`
	ANDDingApprovalTaskActionType DingApprovalTaskActionType = "AND"
	// Ding approval task_action_type `OR`
	ORDingApprovalTaskActionType DingApprovalTaskActionType = "OR"
	// Ding approval task_action_type `NONE`
	NONEDingApprovalTaskActionType DingApprovalTaskActionType = "NONE"
)

// Ding callback message.
type DingCallbackMsg struct {
	Type  string               `json:"type"`
	Data  *DingCallbackMsgData `json:"data"`
	MsgID string               `json:"msg_id"`
	MsgTS int                  `json:"msg_ts"`
}

// Ding callback message data.
type DingCallbackMsgData struct {
	ProcessInstanceID string                        `json:"processInstanceId"`
	CorpID            string                        `json:"corpId"`
	EventType         DingCallbackEventType         `json:"EventType"`
	BusinessID        string                        `json:"businessId"`
	Title             string                        `json:"title"`
	Type              DingApprovalCallbackPhaseType `json:"type"`
	Result            DingApprovalCallbackResult    `json:"result"`
	URL               string                        `json:"url"`
	CreateTime        int                           `json:"createTime"`
	ProcessCode       string                        `json:"processCode"`
	BizCategoryID     string                        `json:"bizCategoryId"`
	StaffID           string                        `json:"staffId"`
}
