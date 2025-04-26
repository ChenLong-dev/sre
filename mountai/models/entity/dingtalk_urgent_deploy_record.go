package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"time"
)

const (
	UrgentDeployTmpName = "urgentDeployMsg"
	UrgentDeployMsg     = `
### {{.Title}}
>- [项目名] {{.ProjectName}}
>- [应用名] {{.AppName}}
>- [环境名] {{.Env}}
>- [操作人] {{.UserName}}
>- [操作] {{.Action}}
>- [时间] {{.OpTime}}
#### [进入AMS, 查看详情]({{.DetailURL}})
`
)

// dingtalk_urgent_deploy_record entity.
type UrgentDeploymentDingTalkMsgRecord struct {
	ID             primitive.ObjectID `bson:"_id" json:"_id"`
	Env            AppEnvName         `bson:"env" json:"env"`
	ProjectID      string             `bson:"project_id" json:"project_id"`
	IsP0Level      bool               `bson:"is_p0_level" json:"is_p0_level"`
	AppID          primitive.ObjectID `bson:"app_id" json:"app_id"`
	CropMsgTaskID  int64              `bson:"crop_msg_task_id" json:"crop_msg_task_id"`
	TaskActionType string             `bson:"task_action_type" json:"task_action_type"`
	Operator       string             `bson:"operator" json:"operator"`
	CreateTime     *time.Time         `bson:"create_time" json:"create_time"`
	UpdateTime     *time.Time         `bson:"update_time" json:"update_time"`
}

func (*UrgentDeploymentDingTalkMsgRecord) TableName() string {
	return "dingtalk_urgent_deploy_record"
}

func (item *UrgentDeploymentDingTalkMsgRecord) GenerateObjectIDString(args map[string]interface{}) string {
	return item.ID.Hex()
}
