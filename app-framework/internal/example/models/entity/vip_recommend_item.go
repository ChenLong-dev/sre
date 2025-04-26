package entity

import (
	"strconv"
	"time"
)

// ==========================
// 数据库映射实体
//
// deepcopy标记使用deepcopy包: https://gitlab.shanhai.int/sre/library/tree/master/base/deepcopy.v2
// ==========================
type VipRecommendItem struct {
	ID                int        `json:"id" gorm:"primary_key;column:id"`
	CreatedAt         *time.Time `json:"create_time" gorm:"column:create_time"`
	UpdatedAt         *time.Time `json:"update_time" gorm:"column:update_time"`
	DeletedAt         *time.Time `json:"delete_time" gorm:"column:delete_time"`
	VipItemID         string     `json:"vip_item_id" gorm:"column:vip_item_id"`
	Title             string     `json:"title" gorm:"column:title"`
	Desc              string     `json:"desc" gorm:"column:desc"`
	PurchasePresent   bool       `json:"purchase_present" gorm:"column:purchase_present"`
	PresentType       string     `json:"present_type" gorm:"column:present_type"`
	PresentActivityId string     `json:"present_activity_id" gorm:"column:present_activity_id"`
	PresentCount      int        `json:"present_count" gorm:"column:present_count"`
	PresentDuration   string     `json:"present_duration" gorm:"column:present_duration"`
}

// ==========================
// 表名方法，该方法名不可更改
// ==========================
func (*VipRecommendItem) TableName() string {
	return "vip_recommend_item"
}

func (item *VipRecommendItem) GenerateIDString(args map[string]interface{}) string {
	return strconv.Itoa(item.ID)
}

func (item *VipRecommendItem) GenerateCreateTimeFormatString(args map[string]interface{}) (string, error) {
	if item.CreatedAt == nil {
		return "", nil
	}

	formatString, ok := args["createTimeFormat"].(string)
	if ok {
		return item.CreatedAt.Format(formatString), nil
	}

	return item.CreatedAt.Format("2006-01-02 15:04:05"), nil
}

func (item *VipRecommendItem) GenerateUpdateTimeFormatString(args map[string]interface{}) (string, error) {
	if item.UpdatedAt == nil {
		return "", nil
	}

	formatString, ok := args["updateTimeFormat"].(string)
	if ok {
		return item.UpdatedAt.Format(formatString), nil
	}

	return item.UpdatedAt.Format("2006-01-02 15:04:05"), nil
}
