package req

type FeishuMsgType string

const FeishuMsgTypeInteractive FeishuMsgType = "interactive"

type FeishuInteractiveMessageReq struct {
	Msgtype   FeishuMsgType `json:"msg_type"`
	ReceiveId string        `json:"receive_id"`
	Content   string        `jssn:"content"`
}
