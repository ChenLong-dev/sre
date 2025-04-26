package req

import "gitlab.shanhai.int/sre/library/base/null"

// ==========================
// 请求模型
//
// binding标记使用v9校验器: https://godoc.org/gopkg.in/go-playground/validator.v9
// ==========================
type GetGithubRepositoryListReq struct {
	Page  null.Int `form:"page" json:"page" binding:"gt=0,numeric"`
	Limit null.Int `form:"limit" json:"limit" binding:"gt=0,numeric"`

	Owner string `json:"-"`
}
