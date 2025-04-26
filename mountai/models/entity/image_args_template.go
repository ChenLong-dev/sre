package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"time"
)

// 镜像参数模版
type ImageArgsTemplate struct {
	ID primitive.ObjectID `bson:"_id" json:"_id"`
	// 团队id
	TeamID string `bson:"team_id" json:"team_id"`
	// 创建人id
	OwnerID string `bson:"owner_id" json:"owner_id"`

	Name    string `bson:"name" json:"name"`
	Content string `bson:"content" json:"content"`

	CreateTime *time.Time `bson:"create_time" json:"create_time"`
	UpdateTime *time.Time `bson:"update_time" json:"update_time"`
	// 软删除
	DeleteTime *time.Time `bson:"delete_time" json:"delete_time"`
}

func (*ImageArgsTemplate) TableName() string {
	return "image_args_template"
}
