package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 团队
type Team struct {
	// 团队id
	ID primitive.ObjectID `bson:"_id" json:"_id"`
	// 团队名
	Name string `bson:"name" json:"name"`
	// 默认钉钉通知地址
	DingHook string `bson:"ding_hook" json:"ding_hook"`
	// 团队标签
	Label string `bson:"label" json:"label"`
	// 阿里云告警名
	AliAlarmName string `bson:"ali_alarm_name" json:"ali_alarm_name"`
	// sentry team
	SentrySlug string `bson:"sentry_slug" json:"sentry_slug"`
	// 额外钉钉hooks
	ExtraDingHooks map[string]string `bson:"extra_ding_hooks" json:"extra_ding_hooks"`

	CreateTime *time.Time `bson:"create_time" json:"create_time"`
	UpdateTime *time.Time `bson:"update_time" json:"update_time"`
	// 软删除
	DeleteTime *time.Time `bson:"delete_time" json:"delete_time"`
}

func (*Team) TableName() string {
	return "team"
}

func (t *Team) GenerateObjectIDString(args map[string]interface{}) string {
	return t.ID.Hex()
}
