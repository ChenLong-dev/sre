package req

import "gitlab.shanhai.int/sre/library/base/null"

// 请求模型
type GetVipItemListReq struct {
	Include   null.String `form:"include" json:"include"`
	PhoneType null.String `form:"phonetype" json:"phonetype"`
	ItemType  null.String `form:"itemtype" json:"itemtype"`
}
