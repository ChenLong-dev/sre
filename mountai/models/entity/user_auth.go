package entity

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// 操作类型
type OperateType string

const (
	// 创建task操作类型
	OperateTypeCreateTask OperateType = "createTask"
	// 删除项目操作类型
	OperateTypeDeleteProject OperateType = "deleteProject"
	// 删除应用操作类型
	OperateTypeDeleteApp OperateType = "deleteApp"
	// 更改应用名称操作类型
	OperateTypeCorrectAppName OperateType = "correctAppName"
	// 查看项目变量值类型
	OperateTypeReadVariableValue OperateType = "readVariableValue"
	// 修改项目变量值类型
	OperateTypeUpdateVariableValue OperateType = "updateVariableValue"
	// 创建项目变量值类型
	OperateTypeCreateVariableValue OperateType = "createVariableValue"
	// 删除项目变量值类型
	OperateTypeDeleteVariableValue OperateType = "deleteVariableValue"
	// 设置应用环境所有集群在 Kong 转发规则中的权重类型
	OperateTypeSetAppClusterKongWeights OperateType = "setAppClusterKongWeights"
	// 删除job操作类型
	OperateTypeDeleteJob OperateType = "deleteJob"
)

const (
	// k8s system id
	K8sSystemUserID = "-1"
)

// 用户声明
// 用于JWT
type UserClaims struct {
	// 标准信息
	jwt.StandardClaims
	// 用户名
	Name string `json:"name"`
	// Email
	Email string `json:"email"`
}

// 用户
type UserAuth struct {
	// git的用户id
	ID string `bson:"_id" json:"_id"`
	// 用户名
	Name string `bson:"name" json:"name"`
	// 头像
	AvatarURL string `bson:"avatar_url" json:"avatar_url"`
	// email
	Email string `bson:"email" json:"email"`
	// jwt的token
	Token string `bson:"token" json:"token"`

	CreateTime *time.Time `bson:"create_time" json:"create_time"`
	UpdateTime *time.Time `bson:"update_time" json:"update_time"`
}

func (*UserAuth) TableName() string {
	return "user_auth"
}
