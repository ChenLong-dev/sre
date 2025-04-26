package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"time"
)

type NotificationType string

const (
	// CI通知方式email
	NotificationTypeEmail NotificationType = "email"
	// CI通知方式DingDing
	NotificationTypeDingDing NotificationType = "dingding"
)

type PipelineStage string

const (
	// 代码扫描
	PipelineStageCodeScan PipelineStage = "CodeScan"
	// 单元测试
	PipelineStageUnitTest PipelineStage = "UnitTest"
	// 构建
	PipelineStageBuild PipelineStage = "Build"
	// 部署fat
	PipelineStageDeployFat PipelineStage = "DeployFat"
	// api测试
	PipelineStageAPITest PipelineStage = "APITest"
	// 部署stg
	PipelineStageDeployStg PipelineStage = "DeployStg"
)

var DefaultPipelineStages = []PipelineStage{PipelineStageCodeScan, PipelineStageUnitTest,
	PipelineStageBuild, PipelineStageDeployFat, PipelineStageAPITest, PipelineStageDeployStg}

type CIJob struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	Name      string             `bson:"name" json:"name"`
	ProjectID string             `bson:"project_id" json:"project_id"`
	ViewURL   string             `bson:"view_url" json:"view_url"`
	HookURL   string             `bson:"hook_url" json:"hook_url"`
	// CI消息通知方式列表
	MessageNotification []NotificationType `bson:"message_notification" json:"message_notification"`
	// 流水线stage
	PipelineStages []PipelineStage `bson:"pipeline_stages" json:"pipeline_stages"`
	// 分支名
	// 工作流需要发指定分支，如果为空则发当前分支
	DeployBranchName map[AppEnvName]string `bson:"deploy_branch_name" json:"deploy_branch_name"`

	CreateTime *time.Time `bson:"create_time" json:"create_time"`
	UpdateTime *time.Time `bson:"update_time" json:"update_time"`
}

func (*CIJob) TableName() string {
	return "ci_job"
}

func (item *CIJob) GenerateObjectIDString(args map[string]interface{}) string {
	return item.ID.Hex()
}
