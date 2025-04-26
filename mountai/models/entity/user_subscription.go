package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubscribeAction 订阅的操作类型
type SubscribeAction string

const (
	// SubscribeActionDeploy 发布
	SubscribeActionDeploy SubscribeAction = "deploy"

	// SubscribeActionStop 停止
	SubscribeActionStop SubscribeAction = "stop"

	// SubscribeActionRestart 重启
	SubscribeActionRestart SubscribeAction = "restart"

	// SubscribeActionResume 恢复
	SubscribeActionResume SubscribeAction = "resume"

	// SubscribeActionDelete 删除
	SubscribeActionDelete SubscribeAction = "delete"

	// SubscribeActionClean 清理
	SubscribeActionClean SubscribeAction = "clean"

	// SubscribeActionManualLaunch 启动
	SubscribeActionManualLaunch SubscribeAction = "manual_launch"

	// SubscribeActionUpdateHPA
	SubscribeActionUpdateHPA SubscribeAction = "update_hpa"

	// SubscribeActionReloadConfig
	SubscribeActionReloadConfig SubscribeAction = "reload_config"
)

// SubscribeEventMsg kafka消息
type SubscribeEventMsg struct {
	ActionType SubscribeAction `json:"action_type"`
	AppID      string          `json:"app_id"`
	Env        AppEnvName      `json:"env"`
	OpTime     string          `json:"op_time"`
	TaskID     string          `json:"task_id"`
	OperatorID string          `json:"operator_id"`
}

// UserSubscription 用户订阅信息
type UserSubscription struct {
	ID         primitive.ObjectID `bson:"_id" json:"_id"`
	AppID      primitive.ObjectID `bson:"app_id" json:"app_id"`
	ActionType SubscribeAction    `bson:"action_type" json:"action_type"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Env        AppEnvName         `bson:"env" json:"env"`
	CreateTime *time.Time         `bson:"create_time" json:"create_time"`
	UpdateTime *time.Time         `bson:"update_time" json:"update_time"`
}

// TableName 表名称
func (*UserSubscription) TableName() string {
	return "user_subscription"
}

// GenerateObjectIDString 生成应用 MongoID
func (item *UserSubscription) GenerateObjectIDString(args map[string]interface{}) string {
	return item.ID.Hex()
}

// GetSubscribeActionDisplay get subscription action display
func GetSubscribeActionDisplay(action SubscribeAction) string {
	switch action {
	case SubscribeActionDeploy:
		return "部署"
	case SubscribeActionStop:
		return "停止"
	case SubscribeActionRestart:
		return "重启"
	case SubscribeActionResume:
		return "恢复"
	case SubscribeActionDelete:
		return "删除"
	case SubscribeActionClean:
		return "清理"
	case SubscribeActionManualLaunch:
		return "手动启动"
	case SubscribeActionUpdateHPA:
		return "弹性伸缩"
	case SubscribeActionReloadConfig:
		return "热加载配置"
	default:
		return "未知"
	}
}

// TransformSubscribeAction transform subscription action
func TransformSubscribeAction(taskAction TaskAction) SubscribeAction {
	switch taskAction {
	case TaskActionFullDeploy, TaskActionCanaryDeploy, TaskActionFullCanaryDeploy:
		return SubscribeActionDeploy
	case TaskActionStop:
		return SubscribeActionStop
	case TaskActionRestart:
		return SubscribeActionRestart
	case TaskActionResume:
		return SubscribeActionResume
	case TaskActionDelete:
		return SubscribeActionDelete
	case TaskActionManualLaunch:
		return SubscribeActionManualLaunch
	case TaskActionUpdateHPA:
		return SubscribeActionUpdateHPA
	case TaskActionReloadConfig:
		return SubscribeActionReloadConfig
	default:
		return "unknown"
	}
}
