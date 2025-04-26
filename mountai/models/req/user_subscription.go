package req

import "rulai/models/entity"

// UserSubscribeReq 订阅
type UserSubscribeReq struct {
	EnvName entity.AppEnvName      `json:"env_name" binding:"required"`
	AppID   string                 `json:"app_id" binding:"required"`
	Action  entity.SubscribeAction `json:"action" binding:"required"`
}

// UserUnsubscribeReq 取消订阅
type UserUnsubscribeReq struct {
	EnvName entity.AppEnvName      `json:"env_name" binding:"required"`
	AppID   string                 `json:"app_id" binding:"required"`
	Action  entity.SubscribeAction `json:"action" binding:"required"`
}
