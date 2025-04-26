package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"time"
)

type VariableType int

const (
	ProjectVariableType VariableType = iota + 1
)

// 变量 (目前用于渲染镜像参数模版)
type Variable struct {
	ID primitive.ObjectID `bson:"_id" json:"_id"`
	// 项目id
	ProjectID string `bson:"project_id" json:"project_id"`
	// 创建人id
	OwnerID string `bson:"owner_id" json:"owner_id"`
	// 修改人id
	EditorID string `bson:"editor_id" json:"editor_id"`

	Type VariableType `bson:"type" json:"type"`
	// 变量名
	Key string `bson:"key" json:"key"`
	// 变量值
	Value string `bson:"value" json:"value"`

	CreateTime *time.Time `bson:"create_time" json:"create_time"`
	UpdateTime *time.Time `bson:"update_time" json:"update_time"`
	// 软删除
	DeleteTime *time.Time `bson:"delete_time" json:"delete_time"`
}

func (*Variable) TableName() string {
	return "variable"
}
