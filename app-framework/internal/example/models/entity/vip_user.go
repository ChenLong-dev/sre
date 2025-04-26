package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// ==========================
// 数据库映射实体
// ==========================
type VipUser struct {
	ID         primitive.ObjectID `bson:"_id" json:"_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Type       string             `bson:"type" json:"type"`
	ExpireTime *time.Time         `bson:"expire_time" json:"expire_time"`
}

// ==========================
// 表名方法，该方法名不可更改
// ==========================
func (*VipUser) TableName() string {
	return "vipusers"
}

func (user *VipUser) GenerateIDString(args map[string]interface{}) string {
	return user.ID.Hex()
}
