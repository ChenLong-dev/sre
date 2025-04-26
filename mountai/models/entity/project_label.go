package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type ProjectLabelValue string

// 已支持的项目 Label 的枚举值
const (
	// BFF
	ProjectLabelBff ProjectLabelValue = "bff"
	//	前端
	ProjectLabelFrontend ProjectLabelValue = "frontend"
	//	在线
	ProjectLabelOnline ProjectLabelValue = "online"
	//	离线
	ProjectLabelOffline ProjectLabelValue = "offline"
	//	后端
	ProjectLabelBackend ProjectLabelValue = "backend"
	//	内部产品
	ProjectLabelInternal ProjectLabelValue = "internal"
	//	外部产品
	ProjectLabelExternal ProjectLabelValue = "external"
	//	P2级别
	ProjectLabelP2 ProjectLabelValue = "P2"
	//	P1级别
	ProjectLabelP1 ProjectLabelValue = "P1"
	//	P0级别
	ProjectLabelP0 ProjectLabelValue = "P0"
)

// IsProjectImportanceLevelLabel 判断标签是否是项目重要等级标签(P + 非负整数)
func (l ProjectLabelValue) IsProjectImportanceLevelLabel() bool {
	length := len(l)
	if length < 2 || l[0] != 'P' {
		return false
	}

	for i := 1; i < length; i++ {
		if l[i] < '0' || l[i] > '9' {
			return false
		}
	}

	return true
}

// 项目标签
type ProjectLabel struct {
	ID    primitive.ObjectID `bson:"_id" json:"_id"`
	Label string             `bson:"label" json:"label"`
	Name  string             `bson:"name" json:"name"`
}

func (*ProjectLabel) TableName() string {
	return "project_label"
}
