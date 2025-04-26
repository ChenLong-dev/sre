package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FavProject 收藏的项目
type FavProject struct {
	// id
	ID primitive.ObjectID `bson:"_id" json:"_id"`
	// 用户id
	UserID string `bson:"user_id" json:"user_id"`
	// 项目id
	ProjectID string `bson:"project_id" json:"project_id"`

	CreateTime *time.Time `bson:"create_time" json:"create_time"`
	UpdateTime *time.Time `bson:"update_time" json:"update_time"`
}

// TableName 表名称
func (*FavProject) TableName() string {
	return "fav_project"
}

// GenerateObjectIDString 生成应用 MongoID
func (item *FavProject) GenerateObjectIDString(args map[string]interface{}) string {
	return item.ID.Hex()
}
