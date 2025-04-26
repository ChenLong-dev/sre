package entity

import (
	"time"
)

// ==========================
// 消息映射实体
// ==========================
type VipRefundNotice struct {
	OrderID   string     `json:"order_id"`
	ItemID    string     `json:"item_id"`
	Type      string     `json:"type"`
	Timestamp *time.Time `json:"timestamp"`
}
