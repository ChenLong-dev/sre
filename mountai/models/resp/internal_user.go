package resp

import "rulai/models/entity"

type GetInternalUsersResp struct {
	ErrCode int               `json:"errcode"`
	ErrMsg  string            `json:"errmsg"`
	Data    *InternalUserData `json:"data"`
}

type InternalUserData struct {
	InternalUsersItems []*entity.InternalUser `json:"items"`
}
