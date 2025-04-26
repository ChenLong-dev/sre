package request

// 分页请求
type V1PaginationRequest struct {
	// 页码
	// 从0开始
	Page int `json:"page" form:"page" binding:"gte=0"`
	// 每页数量
	PageSize int `json:"pagesize" form:"pagesize" binding:"gte=0"`
}
