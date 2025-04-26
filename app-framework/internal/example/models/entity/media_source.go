package entity

import "time"

// ==========================
// 消息映射实体
// ==========================
type MediaSourceNotice struct {
	Type    string                   `json:"type"`
	Action  string                   `json:"action"`
	Current *MediaSourceNoticeDetail `json:"current"`
}

type MediaSourceNoticeDetail struct {
	ID          int        `json:"id"`
	WorkID      int        `json:"work_id"`
	IsStock     bool       `json:"is_stock"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      int        `json:"status"`
	TransStatus int        `json:"trans_status"`
	AppID       int        `json:"app_id"`
	Duration    float64    `json:"duration"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdateAt    *time.Time `json:"updated_at"`
}
