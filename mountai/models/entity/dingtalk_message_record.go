package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DingTalkMessageRecord 钉钉消息记录
type DingTalkMessageRecord struct {
	ID         primitive.ObjectID     `bson:"_id" json:"_id"`
	Content    DingTalkMessageContent `bson:"content" json:"content"`
	TaskID     int64                  `bson:"task_id" json:"task_id"`
	UserIDs    []string               `bson:"user_ids" json:"user_ids"`
	CreateTime *time.Time             `bson:"create_time" json:"create_time"`
	UpdateTime *time.Time             `bson:"update_time" json:"update_time"`
}

// DingTalkMessageContent 消息内容
type DingTalkMessageContent struct {
	ActionType SubscribeAction    `bson:"action_type" json:"action_type"`
	AppID      primitive.ObjectID `bson:"app_id" json:"app_id"`
	ProjectID  string             `bson:"project_id" json:"project_id"`
	Env        AppEnvName         `bson:"env" json:"env"`
}

// TableName 表名称
func (*DingTalkMessageRecord) TableName() string {
	return "dintalk_message_record"
}

// GenerateObjectIDString 生成应用 MongoID
func (item *DingTalkMessageRecord) GenerateObjectIDString(args map[string]interface{}) string {
	return item.ID.Hex()
}
