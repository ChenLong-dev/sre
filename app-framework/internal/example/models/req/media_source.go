package req

import (
	"gitlab.shanhai.int/sre/library/base/null"
)

// ==========================
// 请求模型
// ==========================
type CreateMediaSourceReq struct {
	ID          null.Int    `form:"id" json:"id"`
	WorkID      null.Int    `form:"work_id" json:"work_id"`
	IsStock     null.Bool   `form:"is_stock" json:"is_stock"`
	Title       null.String `form:"title" json:"title"`
	Description null.String `form:"description" json:"description"`
	Status      null.Int    `form:"status" json:"status"`
	TransStatus null.Int    `form:"trans_status" json:"trans_status"`
	AppID       null.Int    `form:"app_id" json:"app_id"`
	Duration    null.Float  `form:"duration" json:"duration"`
	CreatedAt   null.Time   `form:"created_at" json:"created_at"`
	UpdateAt    null.Time   `form:"updated_at" json:"updated_at"`
}
