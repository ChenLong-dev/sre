package models

type BaseListRequest struct {
	Limit int `form:"limit" json:"limit" binding:"required,max=50"`
	Page  int `form:"page" json:"page" binding:"required,min=1"`
}

type BaseListResponse struct {
	List  interface{} `json:"list"`
	Limit int         `json:"limit"`
	Page  int         `json:"page"`
	Count int         `json:"count"`
}

// BaseListRequestWithUnifier 非必填带默认值标准化方法的通用列表请求参数, 需要调用 Unify 方法填充默认值
type BaseListRequestWithUnifier struct {
	Limit int `form:"limit" json:"limit" binding:"omitempty,min=1,max=50"`
	Page  int `form:"page" json:"page" binding:"omitempty,min=1"`
}

func (req *BaseListRequestWithUnifier) Unify() {
	if req.Limit == 0 {
		req.Limit = 20
	}

	if req.Page == 0 {
		req.Page = 1
	}
}
