package req

// CreateConfigRenamePrefixReq 创建特殊配置重命名前缀请求参数
type CreateConfigRenamePrefixReq struct {
	Prefix string `form:"prefix" json:"prefix" binding:"required"`
	Name   string `form:"name" json:"name" binding:"required"`
}
