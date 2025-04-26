package req

import (
	"gitlab.shanhai.int/sre/library/base/null"
)

// ==========================
// 请求模型
// ==========================
type CreateVipRefundReq struct {
	OrderID   null.String `form:"order_id" json:"order_id"`
	ItemID    null.String `form:"item_id" json:"item_id"`
	Type      null.String `form:"type" json:"type"`
	Timestamp null.Time   `form:"timestamp" json:"timestamp"`
}
